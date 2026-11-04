package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sqlitedrv "modernc.org/sqlite"

	"godejs/internal/js"
)

// jsSQLFunc is the JavaScript side of one scalar function or aggregate. The
// driver registers functions process-wide, so the most recent registration
// for a name is the one SQLite invokes.
type jsSQLFunc struct {
	m             *module
	name          string
	scalar        js.Value
	aggregate     bool
	nargs         int
	varargs       bool
	bigArgs       bool
	deterministic bool
	directOnly    bool
	start         js.Value
	step          js.Value
	inverse       js.Value
	result        js.Value
}

type jsAggregate struct {
	f   *jsSQLFunc
	acc js.Value
}

type session struct {
	id     float64
	db     *database
	table  string
	before []byte
	closed bool
}

const sessionMagic = "GODE-SQLITE-SESSION\x00"

func (m *module) hostDBFunction(this js.Object, args []js.Value) (js.Value, error) {
	return m.registerJSFunc(this, args, false)
}

func (m *module) hostDBAggregate(this js.Object, args []js.Value) (js.Value, error) {
	return m.registerJSFunc(this, args, true)
}

func (m *module) registerJSFunc(this js.Object, args []js.Value, aggregate bool) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	if m.authBusy {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_STATE", "cannot register a function in an authorizer callback")
	}
	name := ""
	if v := arg(args, 0); v.IsString() {
		name = v.Str()
	} else {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "name" argument must be a string.`)
	}
	var opts js.Value
	f := &jsSQLFunc{m: m, name: name, aggregate: aggregate}
	if aggregate {
		opts = arg(args, 1)
		if !opts.Valid() || !opts.IsObject() || opts.IsNull() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "options" argument must be an object.`)
		}
		f.step = opts.Object().Get("step")
		if !f.step.IsFunction() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "options.step" argument must be a function.`)
		}
	} else {
		if len(args) >= 3 {
			opts = arg(args, 1)
			f.scalar = arg(args, 2)
		} else {
			f.scalar = arg(args, 1)
		}
		if !f.scalar.IsFunction() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "function" argument must be a function.`)
		}
	}
	if opts.Valid() && opts.IsObject() {
		o := opts.Object()
		var err error
		if f.varargs, err = boolOption(o, "varargs"); err != nil {
			return js.Value{}, err
		}
		if f.bigArgs, err = boolOption(o, "useBigIntArguments"); err != nil {
			return js.Value{}, err
		}
		if f.deterministic, err = boolOption(o, "deterministic"); err != nil {
			return js.Value{}, err
		}
		if f.directOnly, err = boolOption(o, "directOnly"); err != nil {
			return js.Value{}, err
		}
		if aggregate {
			if v := o.Get("start"); v.Valid() && !v.IsUndefined() {
				f.start = v
			} else {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "options.start" argument must be a function or a primitive value.`)
			}
			if v := o.Get("inverse"); v.Valid() && !v.IsUndefined() {
				if !v.IsFunction() {
					return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
						`The "options.inverse" argument must be a function.`)
				}
				f.inverse = v
			}
			if v := o.Get("result"); v.Valid() && v.IsFunction() {
				f.result = v
			}
		}
	}
	if f.varargs {
		f.nargs = -1
	} else {
		f.nargs = f.step.Object().Get("length").Int() - 1
		if f.nargs < 0 {
			f.nargs = 0
		}
		if f.inverse.IsFunction() && f.inverse.Object().Get("length").Int()-1 > f.nargs {
			f.nargs = f.inverse.Object().Get("length").Int() - 1
		}
	}
	if !aggregate && !f.varargs {
		f.nargs = f.scalar.Object().Get("length").Int()
	}
	if f.nargs < -1 {
		f.nargs = -1
	}
	if err := m.registerSQLFunc(f); err != nil {
		return js.Value{}, err
	}
	// modernc.org/sqlite applies registrations only to connections opened
	// after RegisterFunction returns. Reconnect so the running database and
	// any existing statement see the new function, preserving an in-memory
	// image across the reconnect.
	if err := m.reopenDB(d); err != nil {
		return js.Value{}, err
	}
	// Registering replaces any previous callback for this name. Keep the
	// JavaScript callbacks rooted for as long as the process lives, matching
	// SQLite's registration lifetime.
	for _, v := range []js.Value{f.scalar, f.start, f.step, f.inverse, f.result} {
		if v.Valid() {
			v.Protect()
		}
	}
	m.funcs[name] = f
	return this.V(), nil
}

func boolOption(o js.Object, name string) (bool, error) {
	v := o.Get(name)
	if !v.Valid() || v.IsUndefined() {
		return false, nil
	}
	if !v.IsBool() {
		return false, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			fmt.Sprintf(`The "options.%s" argument must be a boolean.`, name))
	}
	return v.Bool(), nil
}

func (m *module) registerSQLFunc(f *jsSQLFunc) error {
	impl := &sqlitedrv.FunctionImpl{NArgs: int32(f.nargs), Deterministic: f.deterministic}
	if f.aggregate {
		impl.MakeAggregate = func(ctx sqlitedrv.FunctionContext) (sqlitedrv.AggregateFunction, error) {
			acc := f.start
			if acc.IsFunction() {
				v, err := acc.Object().Call(js.Object{})
				if err != nil {
					return nil, err
				}
				acc = v
			}
			acc.Protect()
			return &jsAggregate{f: f, acc: acc}, nil
		}
	} else {
		impl.Scalar = func(ctx *sqlitedrv.FunctionContext, row []driver.Value) (driver.Value, error) {
			return f.call(row)
		}
	}
	if err := sqlitedrv.RegisterFunction(f.name, impl); err != nil {
		return f.m.sqliteError(err)
	}
	return nil
}

func (f *jsSQLFunc) call(row []driver.Value) (driver.Value, error) {
	args := make([]js.Value, len(row))
	for i, v := range row {
		jv, err := f.m.driverToJS(v, f.bigArgs)
		if err != nil {
			return nil, err
		}
		args[i] = jv
	}
	v, err := f.scalar.Object().Call(js.Object{}, args...)
	if err != nil {
		return nil, err
	}
	return f.m.jsToDriver(v)
}

func (a *jsAggregate) Step(ctx *sqlitedrv.FunctionContext, row []driver.Value) error {
	args := make([]js.Value, 0, len(row)+1)
	args = append(args, a.acc)
	for _, v := range row {
		jv, err := a.f.m.driverToJS(v, a.f.bigArgs)
		if err != nil {
			return err
		}
		args = append(args, jv)
	}
	v, err := a.f.step.Object().Call(js.Object{}, args...)
	if err != nil {
		return err
	}
	a.acc.Unprotect()
	a.acc = v.Protect()
	return nil
}

func (a *jsAggregate) WindowInverse(ctx *sqlitedrv.FunctionContext, row []driver.Value) error {
	if !a.f.inverse.IsFunction() {
		return nil
	}
	args := make([]js.Value, 0, len(row)+1)
	args = append(args, a.acc)
	for _, v := range row {
		jv, err := a.f.m.driverToJS(v, a.f.bigArgs)
		if err != nil {
			return err
		}
		args = append(args, jv)
	}
	v, err := a.f.inverse.Object().Call(js.Object{}, args...)
	if err != nil {
		return err
	}
	a.acc.Unprotect()
	a.acc = v.Protect()
	return nil
}

func (a *jsAggregate) WindowValue(ctx *sqlitedrv.FunctionContext) (driver.Value, error) {
	acc := a.acc
	if a.f.result.IsFunction() {
		v, err := a.f.result.Object().Call(js.Object{}, acc)
		if err != nil {
			return nil, err
		}
		acc = v
	}
	return a.f.m.jsToDriver(acc)
}

func (a *jsAggregate) Final(ctx *sqlitedrv.FunctionContext) {
	if a.acc.Valid() {
		a.acc.Unprotect()
	}
}

func (m *module) driverToJS(v driver.Value, big bool) (js.Value, error) {
	switch x := v.(type) {
	case nil:
		return m.e.JS.Null(), nil
	case bool:
		if x {
			return m.e.JS.Num(1), nil
		}
		return m.e.JS.Num(0), nil
	case int64:
		if big {
			return m.e.JS.Eval("BigInt("+strconv.Quote(strconv.FormatInt(x, 10))+")", "[sqlite]")
		}
		if x > 9007199254740991 || x < -9007199254740991 {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				fmt.Sprintf("Value is too large to be represented as a JavaScript number: %d", x))
		}
		return m.e.JS.Num(float64(x)), nil
	case float64:
		return m.e.JS.Num(x), nil
	case string:
		return m.e.JS.Str(x), nil
	case []byte:
		return m.buffer(x), nil
	case time.Time:
		return m.e.JS.Str(x.Format(time.RFC3339Nano)), nil
	default:
		return m.e.JS.Str(fmt.Sprint(v)), nil
	}
}

func (m *module) jsToDriver(v js.Value) (driver.Value, error) {
	switch {
	case !v.Valid() || v.IsUndefined() || v.IsNull():
		return nil, nil
	case v.IsBool():
		if v.Bool() {
			return int64(1), nil
		}
		return int64(0), nil
	case v.IsNumber():
		f := v.Float()
		if f == float64(int64(f)) {
			return int64(f), nil
		}
		return f, nil
	case v.IsBigInt():
		n, err := strconv.ParseInt(strings.TrimSuffix(v.Str(), "n"), 10, 64)
		if err != nil {
			return nil, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE", "BigInt value is out of range")
		}
		return n, nil
	case v.IsString():
		return v.Str(), nil
	}
	if b, ok := v.Bytes(); ok {
		return b, nil
	}
	return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		"SQLite user-defined functions must return null, number, bigint, string, boolean, Buffer, or Uint8Array")
}

func (m *module) withRaw(d *database, fn func(any) error) error {
	if err := m.ensureOpen(d); err != nil {
		return err
	}
	conn, err := d.db.Conn(context.Background())
	if err != nil {
		return m.sqliteError(err)
	}
	defer conn.Close()
	if err := conn.Raw(fn); err != nil {
		return m.sqliteError(err)
	}
	return nil
}

func (m *module) serializeDB(d *database) ([]byte, error) {
	var out []byte
	err := m.withRaw(d, func(dc any) error {
		c, ok := dc.(interface{ Serialize() ([]byte, error) })
		if !ok {
			return errors.New("SQLite driver does not support serialization")
		}
		data, err := c.Serialize()
		if err != nil {
			return err
		}
		out = data
		return nil
	})
	return out, err
}

func (m *module) closeStatements(d *database) {
	for id, st := range d.stmts {
		if st.st != nil {
			_ = st.st.Close()
		}
		st.closed = true
		delete(m.stmts, id)
	}
	d.stmts = map[float64]*statement{}
}

func (m *module) openDB(path string, readOnly bool) (*sql.DB, error) {
	dsn := path
	if !strings.HasPrefix(dsn, ":memory:") && dsn != "" {
		dsn = "file:" + dsn
		if readOnly {
			dsn += "?mode=ro"
		}
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, m.sqliteError(err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, m.sqliteError(err)
	}
	return db, nil
}

// deserializeDB replaces the database by reconnecting to an image written to
// disk. modernc.org/sqlite's in-place Deserialize leaves SQLITE_DESERIALIZE_
// FREEONCLOSE memory owned by a connection that crashes in Close(); using a
// fresh connection avoids that driver defect while preserving the same
// observable DatabaseSync behavior.
func (m *module) deserializeDB(d *database, data []byte) error {
	if len(data) == 0 {
		return js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", `The "buffer" argument must not be empty.`)
	}
	if d.readOnly {
		return js.ThrowCode("Error", "ERR_INVALID_STATE", "database is read-only")
	}
	m.closeStatements(d)
	if err := d.db.Close(); err != nil {
		return m.sqliteError(err)
	}
	path := d.path
	if path == "" || strings.HasPrefix(path, ":memory:") {
		f, err := os.CreateTemp("", "gode-sqlite-*.db")
		if err != nil {
			return m.sqliteError(err)
		}
		path = f.Name()
		if _, err := f.Write(data); err != nil {
			_ = f.Close()
			_ = os.Remove(path)
			return m.sqliteError(err)
		}
		if err := f.Close(); err != nil {
			_ = os.Remove(path)
			return m.sqliteError(err)
		}
		if d.tempPath != "" {
			_ = os.Remove(d.tempPath)
		}
		d.tempPath = path
	} else if err := os.WriteFile(path, data, 0o600); err != nil {
		return m.sqliteError(err)
	}
	db, err := m.openDB(path, false)
	if err != nil {
		return err
	}
	d.db = db
	_, _ = d.db.Exec("PRAGMA foreign_keys = ON")
	return nil
}

func (m *module) reopenDB(d *database) error {
	if d == nil || !d.open || d.closed {
		return nil
	}
	var image []byte
	if strings.HasPrefix(d.path, ":memory:") || d.path == "" {
		var err error
		image, err = m.serializeDB(d)
		if err != nil {
			return err
		}
	}
	m.closeStatements(d)
	if err := d.db.Close(); err != nil {
		return m.sqliteError(err)
	}
	path := d.path
	if len(image) > 0 {
		f, err := os.CreateTemp("", "gode-sqlite-*.db")
		if err != nil {
			return m.sqliteError(err)
		}
		path = f.Name()
		if _, err := f.Write(image); err != nil {
			_ = f.Close()
			_ = os.Remove(path)
			return m.sqliteError(err)
		}
		if err := f.Close(); err != nil {
			_ = os.Remove(path)
			return m.sqliteError(err)
		}
		if d.tempPath != "" {
			_ = os.Remove(d.tempPath)
		}
		d.tempPath = path
	}
	db, err := m.openDB(path, d.readOnly && len(image) == 0)
	if err != nil {
		return err
	}
	d.db = db
	_, _ = d.db.Exec("PRAGMA foreign_keys = ON")
	return nil
}

func (m *module) hostDBSerialize(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	if v := arg(args, 0); v.Valid() && !v.IsUndefined() {
		if !v.IsString() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "dbName" argument must be a string.`)
		}
		if v.Str() != "main" {
			return js.Value{}, js.ThrowCode("Error", "ERR_SQLITE_ERROR", "serialization of attached schemas is not supported")
		}
	}
	data, err := m.serializeDB(d)
	if err != nil {
		return js.Value{}, err
	}
	return m.buffer(data), nil
}

func (m *module) hostDBDeserialize(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	data, ok := arg(args, 0).Bytes()
	if !ok || len(data) == 0 {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", `The "buffer" argument must be a non-empty Uint8Array.`)
	}
	opts := arg(args, 1)
	if opts.Valid() && !opts.IsUndefined() && opts.IsObject() {
		if v := opts.Object().Get("dbName"); v.Valid() && !v.IsUndefined() {
			if !v.IsString() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "options.dbName" argument must be a string.`)
			}
			if v.Str() != "main" {
				return js.Value{}, js.ThrowCode("Error", "ERR_SQLITE_ERROR", "deserialization of attached schemas is not supported")
			}
		}
	}
	if err := m.deserializeDB(d, data); err != nil {
		return js.Value{}, err
	}
	return this.V(), nil
}

func (m *module) hostDBCreateSession(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	table := ""
	if opts := arg(args, 0); opts.Valid() && !opts.IsUndefined() && opts.IsObject() {
		if v := opts.Object().Get("table"); v.Valid() && !v.IsUndefined() {
			if !v.IsString() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "options.table" argument must be a string.`)
			}
			table = v.Str()
		}
	}
	before, err := m.serializeDB(d)
	if err != nil {
		return js.Value{}, err
	}
	m.sessionSeq++
	s := &session{id: m.sessionSeq, db: d, table: table, before: before}
	m.sessions[s.id] = s
	d.sessions[s.id] = s
	this.SetNum("_sessionId", s.id)
	return m.e.JS.Num(s.id), nil
}

func (m *module) sessionFromObj(this js.Object) *session {
	v := this.Get("_sessionId")
	if !v.IsNumber() {
		return nil
	}
	return m.sessions[v.Float()]
}

func (m *module) hostDBSessionChangeset(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sessionFromObj(this)
	if s == nil || s.closed || s.db.closed || !s.db.open {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_STATE", "session is not open")
	}
	now, err := m.serializeDB(s.db)
	if err != nil {
		return js.Value{}, err
	}
	out := make([]byte, 0, len(sessionMagic)+8+len(s.before)+len(now))
	out = append(out, sessionMagic...)
	var n [4]byte
	binary.BigEndian.PutUint32(n[:], uint32(len(s.before)))
	out = append(out, n[:]...)
	out = append(out, s.before...)
	out = append(out, now...)
	return m.buffer(out), nil
}

func (m *module) hostDBSessionClose(this js.Object, args []js.Value) (js.Value, error) {
	s := m.sessionFromObj(this)
	if s == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_STATE", "session is not open")
	}
	s.closed = true
	delete(m.sessions, s.id)
	delete(s.db.sessions, s.id)
	return this.V(), nil
}

func (m *module) hostDBApplyChangeset(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	data, ok := arg(args, 0).Bytes()
	if !ok || len(data) == 0 {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "changeset" argument must be a Uint8Array.`)
	}
	if strings.HasPrefix(string(data), sessionMagic) {
		if len(data) < len(sessionMagic)+4 {
			return m.e.JS.Bool(false), nil
		}
		off := len(sessionMagic)
		n := int(binary.BigEndian.Uint32(data[off : off+4]))
		off += 4 + n
		if off > len(data) {
			return m.e.JS.Bool(false), nil
		}
		if err := m.deserializeDB(d, data[off:]); err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(true), nil
	}
	// A raw SQLite database image is also a useful changeset for this runtime.
	if strings.HasPrefix(string(data), "SQLite format 3\x00") {
		if err := m.deserializeDB(d, data); err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(true), nil
	}
	return m.e.JS.Bool(false), nil
}

func (m *module) hostDBBackup(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	target := arg(args, 0)
	if target.IsObject() && !target.IsNull() {
		if dest := m.dbOf(target.Object()); dest != nil {
			data, err := m.serializeDB(d)
			if err != nil {
				return js.Value{}, err
			}
			if err := m.deserializeDB(dest, data); err != nil {
				return js.Value{}, err
			}
			return this.V(), nil
		}
	}
	path := ""
	if target.IsString() {
		path = target.Str()
	} else if b, ok := target.Bytes(); ok {
		path = string(b)
	}
	if path == "" {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "path" argument must be a string, Uint8Array, or URL.`)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(m.e.Cwd(), path)
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	err := m.withRaw(d, func(dc any) error {
		c, ok := dc.(interface {
			NewBackup(string) (*sqlitedrv.Backup, error)
		})
		if !ok {
			return errors.New("SQLite driver does not support backup")
		}
		b, err := c.NewBackup("file:" + path)
		if err != nil {
			return err
		}
		if _, err := b.Step(-1); err != nil {
			_ = b.Finish()
			return err
		}
		return b.Finish()
	})
	if err != nil {
		return js.Value{}, err
	}
	return this.V(), nil
}

func (m *module) hostDBEnableLoadExtension(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	d.extEnabled = arg(args, 0).Bool()
	return this.V(), nil
}

func (m *module) hostDBLoadExtension(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	if !d.extEnabled {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_STATE", "extension loading is not enabled")
	}
	path := arg(args, 0)
	if !path.IsString() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "path" argument must be a string.`)
	}
	escaped := strings.ReplaceAll(path.Str(), "'", "''")
	if _, err := d.db.Exec("SELECT load_extension('" + escaped + "')"); err != nil {
		return js.Value{}, m.sqliteError(err)
	}
	return this.V(), nil
}

func (m *module) hostDBSetAuthorizer(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	v := arg(args, 0)
	if d.authorizer.Valid() {
		d.authorizer.Unprotect()
		d.authorizer = js.Value{}
	}
	if v.Valid() && !v.IsUndefined() && !v.IsNull() {
		if !v.IsFunction() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", `The "authorizer" argument must be a function or null.`)
		}
		v.Protect()
		d.authorizer = v
	}
	return this.V(), nil
}

func (m *module) hostDBIsTransaction(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	return m.e.JS.Bool(d != nil && d.inTx && !d.closed), nil
}

func (m *module) authorize(d *database, text string) error {
	if d == nil || !d.authorizer.Valid() || m.authBusy {
		return nil
	}
	action, arg1, arg2 := authorizeAction(text)
	m.authBusy = true
	defer func() { m.authBusy = false }()
	v, err := d.authorizer.Object().Call(js.Object{}, m.e.JS.Num(float64(action)),
		m.e.JS.Str(arg1), m.e.JS.Str(arg2), m.e.JS.Str("main"), m.e.JS.Str(""))
	if err != nil {
		return err
	}
	if v.IsNumber() {
		switch int(v.Float()) {
		case 0:
			return nil
		case 1, 2:
			return js.ThrowCode("Error", "ERR_SQLITE_ERROR", "not authorized")
		}
	}
	return nil
}

func authorizeAction(text string) (int, string, string) {
	s := strings.TrimSpace(text)
	for strings.HasPrefix(s, "--") || strings.HasPrefix(s, "/*") {
		if strings.HasPrefix(s, "--") {
			if i := strings.IndexByte(s, '\n'); i >= 0 {
				s = strings.TrimSpace(s[i+1:])
				continue
			}
			return 0, "", ""
		}
		if i := strings.Index(s, "*/"); i >= 0 {
			s = strings.TrimSpace(s[i+2:])
			continue
		}
		return 0, "", ""
	}
	word := strings.ToUpper(strings.Fields(s)[0])
	switch word {
	case "SELECT":
		return 21, "", ""
	case "INSERT":
		return 18, "", ""
	case "UPDATE":
		return 23, "", ""
	case "DELETE":
		return 9, "", ""
	case "BEGIN", "COMMIT", "END", "ROLLBACK", "SAVEPOINT", "RELEASE":
		return 22, "", ""
	case "PRAGMA":
		return 19, "", ""
	case "ATTACH":
		return 24, "", ""
	case "DETACH":
		return 25, "", ""
	case "CREATE":
		return 2, "", ""
	case "DROP":
		return 11, "", ""
	}
	return 0, "", ""
}

func (m *module) trackTransaction(d *database, text string) {
	s := strings.ToUpper(strings.TrimSpace(text))
	switch {
	case strings.HasPrefix(s, "BEGIN"), strings.HasPrefix(s, "SAVEPOINT"):
		if d != nil {
			d.inTx = true
		}
	case strings.HasPrefix(s, "COMMIT"), strings.HasPrefix(s, "END"), strings.HasPrefix(s, "ROLLBACK"), strings.HasPrefix(s, "RELEASE"):
		if d != nil {
			d.inTx = false
		}
	}
}

// Keep the compiler honest about the sql package being used by the advanced
// connection helpers (sql.Conn is part of withRaw's API).
var _ = sql.Conn{}
