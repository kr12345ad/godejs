// Package sqlite implements Node's synchronous SQLite API with modernc.org's
// pure-Go SQLite driver. Calls are synchronous on the JavaScript thread, which
// matches DatabaseSync and avoids crossing a goroutine boundary with JS values.
package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"modernc.org/sqlite"

	"godejs/internal/env"
	"godejs/internal/js"

	_ "godejs/modules/buffer"
)

func init() { env.Register("sqlite", Install) }

type module struct {
	e       *env.Env
	exports js.Object
	host    js.Object

	dbs        map[float64]*database
	stmts      map[float64]*statement
	sessions   map[float64]*session
	funcs      map[string]*jsSQLFunc
	dbSeq      float64
	stmtSeq    float64
	sessionSeq float64
	authBusy   bool

	bufferCtor js.Object
	bufferFrom js.Object
}

type database struct {
	id         float64
	obj        js.Object
	path       string
	db         *sql.DB
	open       bool
	closed     bool
	readOnly   bool
	extEnabled bool
	tempPath   string
	authorizer js.Value
	inTx       bool
	stmts      map[float64]*statement
	sessions   map[float64]*session
}

type statement struct {
	id      float64
	db      *database
	st      *sql.Stmt
	sql     string
	closed  bool
	bigInts bool
	arrays  bool
	bare    bool
}

func Install(e *env.Env) js.Object {
	m := &module{
		e:        e,
		dbs:      map[float64]*database{},
		stmts:    map[float64]*statement{},
		sessions: map[float64]*session{},
		funcs:    map[string]*jsSQLFunc{},
	}
	m.exports = e.JS.Obj().Keep()
	m.host = e.JS.Obj().Keep()
	m.host.SetFuncHidden("dbInit", m.hostDBInit)
	m.host.SetFuncHidden("dbOpen", m.hostDBOpen)
	m.host.SetFuncHidden("dbIsOpen", m.hostDBIsOpen)
	m.host.SetFuncHidden("dbClose", m.hostDBClose)
	m.host.SetFuncHidden("dbLocation", m.hostDBLocation)
	m.host.SetFuncHidden("dbExec", m.hostDBExec)
	m.host.SetFuncHidden("dbPrepare", m.hostDBPrepare)
	m.host.SetFuncHidden("stmtRun", m.hostStmtRun)
	m.host.SetFuncHidden("stmtGet", m.hostStmtGet)
	m.host.SetFuncHidden("stmtAll", m.hostStmtAll)
	m.host.SetFuncHidden("stmtColumns", m.hostStmtColumns)
	m.host.SetFuncHidden("stmtClose", m.hostStmtClose)
	m.host.SetFuncHidden("stmtSource", m.hostStmtSource)
	m.host.SetFuncHidden("stmtSetReadBigInts", m.hostStmtSetReadBigInts)
	m.host.SetFuncHidden("stmtSetAllowBare", m.hostStmtSetAllowBare)
	m.host.SetFuncHidden("stmtSetReturnArrays", m.hostStmtSetReturnArrays)
	m.host.SetFuncHidden("stmtExpanded", m.hostStmtExpanded)
	m.host.SetFuncHidden("dbFunction", m.hostDBFunction)
	m.host.SetFuncHidden("dbAggregate", m.hostDBAggregate)
	m.host.SetFuncHidden("dbSerialize", m.hostDBSerialize)
	m.host.SetFuncHidden("dbDeserialize", m.hostDBDeserialize)
	m.host.SetFuncHidden("dbCreateSession", m.hostDBCreateSession)
	m.host.SetFuncHidden("dbSessionChangeset", m.hostDBSessionChangeset)
	m.host.SetFuncHidden("dbSessionClose", m.hostDBSessionClose)
	m.host.SetFuncHidden("dbApplyChangeset", m.hostDBApplyChangeset)
	m.host.SetFuncHidden("dbBackup", m.hostDBBackup)
	m.host.SetFuncHidden("dbEnableLoadExtension", m.hostDBEnableLoadExtension)
	m.host.SetFuncHidden("dbLoadExtension", m.hostDBLoadExtension)
	m.host.SetFuncHidden("dbSetAuthorizer", m.hostDBSetAuthorizer)
	m.host.SetFuncHidden("dbIsTransaction", m.hostDBIsTransaction)

	if buf := e.Exports("buffer").Get("Buffer"); buf.Valid() && buf.IsFunction() {
		m.bufferCtor = buf.Object().Keep()
		if from := m.bufferCtor.Get("from"); from.Valid() && from.IsFunction() {
			m.bufferFrom = from.Object().Keep()
		}
	}
	api, err := e.JS.Eval(prelude, "sqlite.js")
	if err != nil {
		panic("godejs sqlite: " + err.Error())
	}
	built, err := api.Object().Call(js.Object{}, m.host.V())
	if err != nil {
		panic("godejs sqlite: " + err.Error())
	}
	for _, key := range built.Object().OwnKeys() {
		m.exports.Set(key, built.Object().Get(key))
	}
	return m.exports
}

func Implemented() []string {
	return []string{
		"sqlite.DatabaseSync",
		"sqlite.DatabaseSync#close",
		"sqlite.DatabaseSync#exec",
		"sqlite.DatabaseSync#isOpen",
		"sqlite.DatabaseSync#location",
		"sqlite.DatabaseSync#open",
		"sqlite.DatabaseSync#prepare",
		"sqlite.StatementSync",
		"sqlite.StatementSync#all",
		"sqlite.StatementSync#close",
		"sqlite.StatementSync#columns",
		"sqlite.StatementSync#get",
		"sqlite.StatementSync#iterate",
		"sqlite.StatementSync#run",
		"sqlite.StatementSync#setAllowBareNamedParameters",
		"sqlite.StatementSync#setReadBigInts",
		"sqlite.StatementSync#setReturnArrays",
		"sqlite.StatementSync#sourceSQL",
		"sqlite.StatementSync#expandedSQL",
		"sqlite.DatabaseSync#function",
		"sqlite.DatabaseSync#aggregate",
		"sqlite.DatabaseSync#serialize",
		"sqlite.DatabaseSync#deserialize",
		"sqlite.DatabaseSync#createSession",
		"sqlite.DatabaseSync#applyChangeset",
		"sqlite.DatabaseSync#enableLoadExtension",
		"sqlite.DatabaseSync#loadExtension",
		"sqlite.DatabaseSync#setAuthorizer",
		"sqlite.DatabaseSync#isTransaction",
		"sqlite.DatabaseSync#readOnly",
		"sqlite.backup",
	}
}

func (m *module) dbOf(this js.Object) *database {
	v := this.Get("_dbId")
	if !v.IsNumber() {
		return nil
	}
	return m.dbs[v.Float()]
}

func (m *module) stmtOf(this js.Object) *statement {
	v := this.Get("_id")
	if !v.IsNumber() {
		return nil
	}
	return m.stmts[v.Float()]
}

func (m *module) hostDBInit(this js.Object, args []js.Value) (js.Value, error) {
	path := ":memory:"
	if v := arg(args, 0); v.IsString() {
		path = v.Str()
	}
	opts := arg(args, 1)
	readOnly := false
	openNow := true
	fk := true
	timeout := 0
	if opts.Valid() && opts.IsObject() {
		o := opts.Object()
		if v := o.Get("readOnly"); v.IsBool() {
			readOnly = v.Bool()
		}
		if v := o.Get("open"); v.IsBool() {
			openNow = v.Bool()
		}
		if v := o.Get("enableForeignKeyConstraints"); v.IsBool() {
			fk = v.Bool()
		}
		if v := o.Get("timeout"); v.IsNumber() {
			timeout = int(v.Int())
		}
	}
	dsn := path
	if !strings.HasPrefix(path, ":memory:") && path != "" {
		dsn = "file:" + path
		if readOnly {
			dsn += "?mode=ro"
		}
		if timeout > 0 {
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			dsn += sep + "_pragma=busy_timeout(" + strconv.Itoa(timeout) + ")"
		}
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return js.Value{}, m.sqliteError(err)
	}
	// A single connection is required for :memory: databases and for the
	// raw connection operations (serialize/deserialize/backup), which must
	// address the same connection the statements use.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if openNow {
		if err := db.Ping(); err != nil {
			_ = db.Close()
			return js.Value{}, m.sqliteError(err)
		}
		if fk {
			_, _ = db.Exec("PRAGMA foreign_keys = ON")
		}
	}
	m.dbSeq++
	d := &database{id: m.dbSeq, obj: this, path: path, db: db, open: openNow, readOnly: readOnly, stmts: map[float64]*statement{}, sessions: map[float64]*session{}}
	m.dbs[d.id] = d
	this.SetHidden("$godeSQLiteDB", m.e.JS.Num(d.id))
	this.SetNum("_dbId", d.id)
	this.Protect()
	return m.e.JS.Num(d.id), nil
}

func (m *module) hostDBOpen(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if d == nil || d.closed {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_STATE", "database is not open")
	}
	if !d.open {
		if err := d.db.Ping(); err != nil {
			return js.Value{}, m.sqliteError(err)
		}
		d.open = true
	}
	return this.V(), nil
}

func (m *module) hostDBIsOpen(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	return m.e.JS.Bool(d != nil && d.open && !d.closed), nil
}

func (m *module) hostDBClose(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if d == nil {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_STATE", "database is not open")
	}
	if d.closed {
		return js.Value{}, nil
	}
	if !d.open {
		d.closed = true
		return js.Value{}, nil
	}
	for _, st := range d.stmts {
		if st.st != nil {
			_ = st.st.Close()
		}
		st.closed = true
	}
	d.stmts = map[float64]*statement{}
	for _, sess := range d.sessions {
		sess.closed = true
	}
	d.sessions = map[float64]*session{}
	if d.authorizer.Valid() {
		d.authorizer.Unprotect()
		d.authorizer = js.Value{}
	}
	if err := d.db.Close(); err != nil {
		return js.Value{}, m.sqliteError(err)
	}
	if d.tempPath != "" {
		_ = os.Remove(d.tempPath)
		d.tempPath = ""
	}
	d.open = false
	d.closed = true
	return js.Value{}, nil
}

func (m *module) hostDBLocation(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if d == nil {
		return m.e.JS.Str(""), nil
	}
	return m.e.JS.Str(d.path), nil
}

func (m *module) hostDBExec(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	text := ""
	if v := arg(args, 0); v.IsString() {
		text = v.Str()
	} else {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"sql\" argument must be a string")
	}
	if err := m.authorize(d, text); err != nil {
		return js.Value{}, err
	}
	if _, err := d.db.Exec(text); err != nil {
		return js.Value{}, m.sqliteError(err)
	}
	m.trackTransaction(d, text)
	return this.V(), nil
}

func (m *module) hostDBPrepare(this js.Object, args []js.Value) (js.Value, error) {
	d := m.dbOf(this)
	if err := m.ensureOpen(d); err != nil {
		return js.Value{}, err
	}
	text := ""
	if v := arg(args, 0); v.IsString() {
		text = v.Str()
	} else {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"sql\" argument must be a string")
	}
	if err := m.authorize(d, text); err != nil {
		return js.Value{}, err
	}
	st, err := d.db.Prepare(text)
	if err != nil {
		return js.Value{}, m.sqliteError(err)
	}
	m.stmtSeq++
	statement := &statement{id: m.stmtSeq, db: d, st: st, sql: text, bare: true}
	m.stmts[statement.id] = statement
	d.stmts[statement.id] = statement
	return m.e.JS.Num(statement.id), nil
}

func (m *module) ensureOpen(d *database) error {
	if d == nil || d.closed {
		return js.ThrowCode("Error", "ERR_INVALID_STATE", "database is not open")
	}
	if !d.open {
		return js.ThrowCode("Error", "ERR_INVALID_STATE", "database is not open")
	}
	return nil
}

func (m *module) stmtArgs(args []js.Value) []js.Value {
	if len(args) == 1 && args[0].IsArray() {
		arr := args[0].Object()
		out := make([]js.Value, arr.Length())
		for i := range out {
			out[i] = arr.At(i)
		}
		return out
	}
	return args
}

func (m *module) hostStmtRun(this js.Object, args []js.Value) (js.Value, error) {
	st := m.stmtOf(this)
	if st == nil || st.closed {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_STATE", "statement is not open")
	}
	bound, err := m.bind(st, m.stmtArgs(args))
	if err != nil {
		return js.Value{}, err
	}
	res, err := st.st.Exec(bound...)
	if err != nil {
		return js.Value{}, m.sqliteError(err)
	}
	changes, _ := res.RowsAffected()
	last, _ := res.LastInsertId()
	o := m.e.JS.Obj()
	o.SetNum("changes", float64(changes))
	o.Set("lastInsertRowid", m.integerValue(last, st.bigInts))
	return o.V(), nil
}

func (m *module) hostStmtGet(this js.Object, args []js.Value) (js.Value, error) {
	rows, err := m.query(this, m.stmtArgs(args))
	if err != nil {
		return js.Value{}, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return js.Value{}, m.sqliteError(err)
		}
		return m.e.JS.Undefined(), nil
	}
	row, err := m.scanRow(rows, cols, this)
	if err != nil {
		return js.Value{}, err
	}
	return row.V(), nil
}

func (m *module) hostStmtAll(this js.Object, args []js.Value) (js.Value, error) {
	rows, err := m.query(this, m.stmtArgs(args))
	if err != nil {
		return js.Value{}, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	arr := m.e.JS.Arr(nil)
	i := 0
	for rows.Next() {
		row, scanErr := m.scanRow(rows, cols, this)
		if scanErr != nil {
			return js.Value{}, scanErr
		}
		arr.SetAt(i, row.V())
		i++
	}
	if err := rows.Err(); err != nil {
		return js.Value{}, m.sqliteError(err)
	}
	return arr.V(), nil
}

func (m *module) query(this js.Object, args []js.Value) (*sql.Rows, error) {
	st := m.stmtOf(this)
	if st == nil || st.closed {
		return nil, js.ThrowCode("Error", "ERR_INVALID_STATE", "statement is not open")
	}
	bound, err := m.bind(st, args)
	if err != nil {
		return nil, err
	}
	rows, err := st.st.Query(bound...)
	if err != nil {
		return nil, m.sqliteError(err)
	}
	return rows, nil
}

func (m *module) hostStmtColumns(this js.Object, args []js.Value) (js.Value, error) {
	st := m.stmtOf(this)
	if st == nil || st.closed {
		return js.Value{}, js.ThrowCode("Error", "ERR_INVALID_STATE", "statement is not open")
	}
	rows, err := st.st.Query()
	if err != nil {
		return m.e.JS.Arr(nil).V(), nil
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return m.e.JS.Arr(nil).V(), nil
	}
	types, _ := rows.ColumnTypes()
	out := m.e.JS.Arr(nil)
	for i, name := range cols {
		o := m.e.JS.Obj()
		o.SetStr("name", name)
		o.Set("column", m.e.JS.Null())
		o.Set("table", m.e.JS.Null())
		o.Set("database", m.e.JS.Null())
		if i < len(types) {
			o.SetStr("type", types[i].DatabaseTypeName())
		} else {
			o.SetStr("type", "")
		}
		out.SetAt(i, o.V())
	}
	return out.V(), nil
}

func (m *module) hostStmtClose(this js.Object, args []js.Value) (js.Value, error) {
	st := m.stmtOf(this)
	if st == nil || st.closed {
		return this.V(), nil
	}
	if st.st != nil {
		_ = st.st.Close()
	}
	st.closed = true
	delete(st.db.stmts, st.id)
	delete(m.stmts, st.id)
	return this.V(), nil
}

func (m *module) hostStmtSource(this js.Object, args []js.Value) (js.Value, error) {
	st := m.stmtOf(this)
	if st == nil {
		return m.e.JS.Str(""), nil
	}
	return m.e.JS.Str(st.sql), nil
}

func (m *module) hostStmtExpanded(this js.Object, args []js.Value) (js.Value, error) {
	st := m.stmtOf(this)
	if st == nil {
		return m.e.JS.Str(""), nil
	}
	return m.e.JS.Str(st.sql), nil
}

func (m *module) hostStmtSetReadBigInts(this js.Object, args []js.Value) (js.Value, error) {
	if st := m.stmtOf(this); st != nil {
		st.bigInts = arg(args, 0).Bool()
	}
	return this.V(), nil
}

func (m *module) hostStmtSetAllowBare(this js.Object, args []js.Value) (js.Value, error) {
	if st := m.stmtOf(this); st != nil {
		st.bare = arg(args, 0).Bool()
	}
	return this.V(), nil
}

func (m *module) hostStmtSetReturnArrays(this js.Object, args []js.Value) (js.Value, error) {
	if st := m.stmtOf(this); st != nil {
		st.arrays = arg(args, 0).Bool()
	}
	return this.V(), nil
}

func (m *module) bind(st *statement, args []js.Value) ([]any, error) {
	if len(args) == 1 && args[0].IsObject() && !isBinary(args[0]) {
		obj := args[0].Object()
		out := make([]any, 0)
		for _, key := range obj.OwnKeys() {
			name := key
			if st.bare && (strings.HasPrefix(name, ":") || strings.HasPrefix(name, "@") || strings.HasPrefix(name, "$")) {
				name = name[1:]
			}
			value, err := m.bindValue(obj.Get(key))
			if err != nil {
				return nil, err
			}
			out = append(out, sql.Named(name, value))
		}
		return out, nil
	}
	out := make([]any, 0, len(args))
	for _, v := range args {
		value, err := m.bindValue(v)
		if err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, nil
}

func (m *module) bindValue(v js.Value) (any, error) {
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
		"SQLite parameters must be null, number, bigint, string, Buffer, or Uint8Array")
}

func isBinary(v js.Value) bool {
	if _, ok := v.Bytes(); ok {
		return true
	}
	return false
}

func (m *module) scanRow(rows *sql.Rows, cols []string, this js.Object) (js.Object, error) {
	st := m.stmtOf(this)
	values := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return js.Object{}, m.sqliteError(err)
	}
	if st != nil && st.arrays {
		arr := m.e.JS.Arr(nil)
		for i, value := range values {
			arr.SetAt(i, m.columnValue(value, st.bigInts))
		}
		return arr.V().Object(), nil
	}
	o := m.e.JS.Obj()
	for i, name := range cols {
		o.Set(name, m.columnValue(values[i], st != nil && st.bigInts))
	}
	return o, nil
}

func (m *module) columnValue(value any, bigInts bool) js.Value {
	switch v := value.(type) {
	case nil:
		return m.e.JS.Null()
	case int64:
		return m.integerValue(v, bigInts)
	case float64:
		return m.e.JS.Num(v)
	case string:
		return m.e.JS.Str(v)
	case []byte:
		return m.buffer(v)
	case bool:
		if v {
			return m.e.JS.Num(1)
		}
		return m.e.JS.Num(0)
	default:
		return m.e.JS.Str(fmt.Sprint(v))
	}
}

func (m *module) integerValue(n int64, big bool) js.Value {
	if big || n > 9007199254740991 || n < -9007199254740991 {
		v, err := m.e.JS.Eval("BigInt("+strconv.Quote(strconv.FormatInt(n, 10))+")", "[sqlite]")
		if err == nil {
			return v
		}
	}
	return m.e.JS.Num(float64(n))
}

func (m *module) buffer(b []byte) js.Value {
	if m.bufferFrom.Valid() && m.bufferCtor.Valid() {
		if v, err := m.bufferFrom.Call(m.bufferCtor, m.e.JS.Bytes(b).V()); err == nil {
			return v
		}
	}
	return m.e.JS.Bytes(b).V()
}

func (m *module) sqliteError(err error) error {
	code := "ERR_SQLITE_ERROR"
	msg := err.Error()
	if strings.Contains(strings.ToLower(msg), "constraint") {
		code = "ERR_SQLITE_CONSTRAINT"
	}
	var se *sqlite.Error
	if errors.As(err, &se) {
		msg = se.Error()
	}
	return js.ThrowCode("Error", code, msg)
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

const prelude = `(function (host) {
  'use strict';

  function invalidArg(name, expected, value) {
    var err = new TypeError('The "' + name + '" argument must be ' + expected + '. Received ' + String(value));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  class StatementSync {
    constructor(db, sql, id) {
      this._db = db;
      this._sql = sql;
      this._id = id;
      this._closed = false;
      this._big = false;
      this._bare = true;
      this._arrays = false;
    }
    run() { return host.stmtRun.call(this, Array.prototype.slice.call(arguments)); }
    get() { return host.stmtGet.call(this, Array.prototype.slice.call(arguments)); }
    all() { return host.stmtAll.call(this, Array.prototype.slice.call(arguments)); }
    iterate() {
      var rows = this.all.apply(this, arguments);
      var index = 0;
      var iterator = {
        next: function () {
          if (index >= rows.length) return { value: undefined, done: true };
          return { value: rows[index++], done: false };
        },
        return: function () { index = rows.length; return { value: undefined, done: true }; }
      };
      iterator[Symbol.iterator] = function () { return this; };
      return iterator;
    }
    columns() { return host.stmtColumns.call(this); }
    setReadBigInts(value) { this._big = !!value; host.stmtSetReadBigInts.call(this, this._big); }
    setAllowBareNamedParameters(value) { this._bare = !!value; host.stmtSetAllowBare.call(this, this._bare); }
    setReturnArrays(value) { this._arrays = !!value; host.stmtSetReturnArrays.call(this, this._arrays); }
    close() { this._closed = true; return host.stmtClose.call(this); }
    [Symbol.dispose]() { this.close(); }
    get sourceSQL() { return host.stmtSource.call(this); }
    get expandedSQL() { return host.stmtExpanded.call(this); }
  }

  class Session {
    constructor(db, id) {
      this._db = db;
      this._sessionId = id;
    }
    changeset() { return host.dbSessionChangeset.call(this); }
    patchset() { return host.dbSessionChangeset.call(this); }
    close() { return host.dbSessionClose.call(this); }
    [Symbol.dispose]() { this.close(); }
  }

  class DatabaseSync {
    constructor(path, options) {
      if (path === undefined) path = ':memory:';
      if (typeof path !== 'string') throw invalidArg('path', 'of type string', path);
      options = options || {};
      this._stmts = [];
      this._sessions = [];
      this._readOnly = !!options.readOnly;
      this._dbId = host.dbInit.call(this, path, options);
    }
    exec(sql) { return host.dbExec.call(this, String(sql)); }
    prepare(sql) {
      if (typeof sql !== 'string') throw invalidArg('sql', 'of type string', sql);
      var id = host.dbPrepare.call(this, sql);
      var stmt = new StatementSync(this, sql, id);
      this._stmts.push(stmt);
      return stmt;
    }
    close() {
      for (var i = 0; i < this._stmts.length; i++) this._stmts[i].close();
      this._stmts.length = 0;
      for (var j = 0; j < this._sessions.length; j++) {
        try { this._sessions[j].close(); } catch (e) {}
      }
      this._sessions.length = 0;
      return host.dbClose.call(this);
    }
    open() { return host.dbOpen.call(this); }
    isOpen() { return host.dbIsOpen.call(this); }
    location() { return host.dbLocation.call(this); }
    function(name, options, fn) {
      if (arguments.length < 3) { fn = options; options = undefined; }
      if (typeof name !== 'string') throw invalidArg('name', 'a string', name);
      if (options !== undefined && (typeof options !== 'object' || options === null || Array.isArray(options))) {
        throw invalidArg('options', 'an object', options);
      }
      if (typeof fn !== 'function') throw invalidArg('function', 'a function', fn);
      return host.dbFunction.call(this, name, options, fn);
    }
    aggregate(name, options) {
      if (typeof name !== 'string') throw invalidArg('name', 'a string', name);
      if (typeof options !== 'object' || options === null || Array.isArray(options)) {
        throw invalidArg('options', 'an object', options);
      }
      return host.dbAggregate.call(this, name, options);
    }
    serialize(dbName) {
      if (dbName !== undefined && typeof dbName !== 'string') {
        throw invalidArg('dbName', 'a string', dbName);
      }
      return host.dbSerialize.call(this, dbName);
    }
    deserialize(buffer, options) {
      if (!(buffer instanceof Uint8Array)) throw invalidArg('buffer', 'a Uint8Array', buffer);
      if (buffer.byteLength === 0) {
        var empty = new TypeError('The "buffer" argument must not be empty.');
        empty.code = 'ERR_INVALID_ARG_VALUE'; throw empty;
      }
      if (options !== undefined && (typeof options !== 'object' || options === null)) {
        throw invalidArg('options', 'an object', options);
      }
      return host.dbDeserialize.call(this, buffer, options);
    }
    createSession(options) {
      if (options !== undefined && (typeof options !== 'object' || options === null || Array.isArray(options))) {
        throw invalidArg('options', 'an object', options);
      }
      var id = host.dbCreateSession.call(this, options);
      var session = new Session(this, id);
      this._sessions.push(session);
      return session;
    }
    applyChangeset(changeset, options) {
      if (!(changeset instanceof Uint8Array)) throw invalidArg('changeset', 'a Uint8Array', changeset);
      if (options !== undefined && (typeof options !== 'object' || options === null)) {
        throw invalidArg('options', 'an object', options);
      }
      return host.dbApplyChangeset.call(this, changeset, options);
    }
    enableLoadExtension(enable) {
      if (typeof enable !== 'boolean') throw invalidArg('enable', 'a boolean', enable);
      return host.dbEnableLoadExtension.call(this, enable);
    }
    loadExtension(path) {
      if (typeof path !== 'string') throw invalidArg('path', 'a string', path);
      return host.dbLoadExtension.call(this, path);
    }
    setAuthorizer(authorizer) {
      if (authorizer !== null && authorizer !== undefined && typeof authorizer !== 'function') {
        throw invalidArg('authorizer', 'a function or null', authorizer);
      }
      return host.dbSetAuthorizer.call(this, authorizer === undefined ? null : authorizer);
    }
    get isTransaction() { return host.dbIsTransaction.call(this); }
    get readOnly() { return this._readOnly; }
    [Symbol.dispose]() { this.close(); }
  }

  function backup(sourceDb, path, options) {
    if (!sourceDb || typeof sourceDb !== 'object' || typeof sourceDb._dbId !== 'number') {
      throw invalidArg('sourceDb', 'an object', sourceDb);
    }
    var validPath = typeof path === 'string' || path instanceof Uint8Array ||
      (path && typeof path === 'object' && typeof path.href === 'string');
    if (!validPath) {
      var pe = new TypeError('The "path" argument must be a string, Uint8Array, or URL without null bytes.');
      pe.code = 'ERR_INVALID_ARG_TYPE'; throw pe;
    }
    if (typeof path === 'string' && path.indexOf('\0') !== -1) {
      var ne = new TypeError('The "path" argument must be a string, Uint8Array, or URL without null bytes.');
      ne.code = 'ERR_INVALID_ARG_TYPE'; throw ne;
    }
    if (options !== undefined && (typeof options !== 'object' || options === null)) {
      throw invalidArg('options', 'an object', options);
    }
    var target = path && path.href ? path.href : path;
    return Promise.resolve().then(function () {
      return host.dbBackup.call(sourceDb, target, options);
    });
  }

  var constants = {
    SQLITE_OK: 0, SQLITE_ERROR: 1, SQLITE_BUSY: 5, SQLITE_LOCKED: 6,
    SQLITE_READONLY: 8, SQLITE_CONSTRAINT: 19, SQLITE_MISMATCH: 20,
    SQLITE_CHANGESET_DATA: 1, SQLITE_CHANGESET_NOTFOUND: 2,
    SQLITE_CHANGESET_CONFLICT: 3, SQLITE_CHANGESET_CONSTRAINT: 4,
    SQLITE_CHANGESET_FOREIGN_KEY: 5, SQLITE_CHANGESET_OMIT: 0,
    SQLITE_CHANGESET_REPLACE: 1, SQLITE_CHANGESET_ABORT: 2,
    SQLITE_DENY: 1, SQLITE_IGNORE: 2, SQLITE_SELECT: 21, SQLITE_INSERT: 18,
    SQLITE_UPDATE: 23, SQLITE_DELETE: 9, SQLITE_CREATE_TABLE: 2,
    SQLITE_DROP_TABLE: 11, SQLITE_PRAGMA: 19, SQLITE_TRANSACTION: 22,
    SQLITE_ATTACH: 24, SQLITE_DETACH: 25
  };

  return { DatabaseSync: DatabaseSync, StatementSync: StatementSync, Session: Session, backup: backup, constants: constants };
})`
