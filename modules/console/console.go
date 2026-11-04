package console

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/modules/util"
)

func init() { env.Register("console", Install) }

// Install builds the console module and replaces JavaScriptCore's native
// console global (which writes nowhere) with the GodeJS instance.
func Install(e *env.Env) js.Object {
	defaultOut, defaultErr := e.Out(), e.Err()
	// The global console binds process.stdout/process.stderr lazily, like
	// Node's initializeGlobalConsole: tests replace process.stdout.write and
	// expect console.* to observe the new function.
	con := buildInstance(e,
		&dynamicWriter{e: e, name: "stdout", fallback: defaultOut},
		&dynamicWriter{e: e, name: "stderr", fallback: defaultErr},
		consoleConfig{groupWidth: 2, lazyOut: "stdout", lazyErr: "stderr"})

	factory := e.JS.Obj()
	// consoleProto is filled once the class wrapper exists; factory instances
	// and the global console both use it so `instanceof Console` holds.
	var consoleProto js.Object
	factory.SetFunc("Console", func(this js.Object, args []js.Value) (js.Value, error) {
		out, errOut := defaultOut, defaultErr
		width := 2
		var outTarget js.Value
		switch {
		case len(args) == 0:
			// default streams
		case args[0].IsObject() && !hasWrite(args[0]):
			opts := args[0].Object()
			if g := opts.Get("groupIndentation"); g.Valid() && !g.IsUndefined() {
				n, err := groupIndentation(g)
				if err != nil {
					return js.Value{}, err
				}
				width = n
			}
			so := opts.Get("stdout")
			if !hasWrite(so) {
				return js.Value{}, consoleWritableErr(so, "stdout")
			}
			se := opts.Get("stderr")
			if se.Valid() && !se.IsUndefined() && !se.IsNull() && !hasWrite(se) {
				return js.Value{}, consoleWritableErr(se, "stderr")
			}
			out, outTarget = writerFor(so, defaultOut), so
			if se.Valid() && !se.IsUndefined() && !se.IsNull() {
				errOut = writerFor(se, defaultErr)
			} else {
				errOut = out
			}
		default:
			if !args[0].IsUndefined() {
				out, outTarget = writerFor(args[0], defaultOut), args[0]
			}
			if len(args) > 1 && !args[1].IsUndefined() {
				errOut = writerFor(args[1], defaultErr)
			}
		}
		inst := buildInstance(e, out, errOut, consoleConfig{groupWidth: width, outTarget: outTarget})
		if consoleProto.Valid() {
			inst.SetProto(consoleProto)
		}
		return inst.V(), nil
	})

	ctor := factory.Get("Console")
	g := e.Global()
	// A tiny JS wrapper supplies class/new semantics; all behavior stays in Go.
	g.SetHidden("__godeConsoleFactory", ctor)
	if v, err := e.JS.Eval(`(function(factory){
		function Console(stdout, stderr) { return factory(stdout, stderr); }
		return Console;
	})(globalThis.__godeConsoleFactory)`, "<console>"); err == nil && v.IsFunction() {
		ctor = v
	}
	g.Delete("__godeConsoleFactory")
	if proto := ctor.Object().Get("prototype"); proto.IsObject() {
		consoleProto = proto.Object()
		con.SetProto(consoleProto)
	}

	con.Set("Console", ctor)
	g.SetObj("console", con)
	return con
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"console.Console",
		"console.assert",
		"console.clear",
		"console.count",
		"console.countReset",
		"console.debug",
		"console.dir",
		"console.dirxml",
		"console.error",
		"console.group",
		"console.groupCollapsed",
		"console.groupEnd",
		"console.info",
		"console.log",
		"console.profile",
		"console.profileEnd",
		"console.table",
		"console.time",
		"console.timeEnd",
		"console.timeLog",
		"console.timeStamp",
		"console.trace",
		"console.warn",
	}
}

// consoleConfig carries the parts of Node's Console options that affect
// output: groupIndentation and the stream used for isTTY checks.
type consoleConfig struct {
	groupWidth int
	outTarget  js.Value
	// lazyOut/lazyErr name process fields to resolve on every write (the
	// global console); empty means the writer is bound for the instance life.
	lazyOut string
	lazyErr string
}

type consoleInstance struct {
	e      *env.Env
	out    io.Writer
	errOut io.Writer

	indent     int
	groupWidth int
	outTarget  js.Value
	lazyOut    string
	lazyErr    string
	counts     map[string]int
	times      map[string]time.Time
}

func buildInstance(e *env.Env, out, errOut io.Writer, cfg consoleConfig) js.Object {
	width := cfg.groupWidth
	if width < 0 {
		width = 0
	}
	ci := &consoleInstance{
		e:          e,
		out:        out,
		errOut:     errOut,
		groupWidth: width,
		outTarget:  cfg.outTarget,
		lazyOut:    cfg.lazyOut,
		lazyErr:    cfg.lazyErr,
		counts:     map[string]int{},
		times:      map[string]time.Time{},
	}
	o := e.JS.Obj()
	o.SetFunc("log", ci.log)
	o.SetFunc("info", ci.log)
	o.SetFunc("debug", ci.log)
	o.SetFunc("dirxml", ci.log)
	o.SetFunc("warn", ci.warn)
	o.SetFunc("error", ci.warn)
	o.SetFunc("dir", ci.dir)
	o.SetFunc("assert", ci.assert)
	o.SetFunc("trace", ci.trace)
	o.SetFunc("count", ci.count)
	o.SetFunc("countReset", ci.countReset)
	o.SetFunc("time", ci.time)
	o.SetFunc("timeLog", ci.timeLog)
	o.SetFunc("timeEnd", ci.timeEnd)
	o.SetFunc("group", ci.group)
	o.SetFunc("groupCollapsed", ci.group)
	o.SetFunc("groupEnd", ci.groupEnd)
	o.SetFunc("table", ci.table)
	o.SetFunc("clear", ci.clear)
	// Node routes profile/profileEnd/timeStamp to the inspector. With no
	// inspector attached they are no-ops; keep the functions so feature
	// probes and libraries that call them keep running.
	noop := func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	}
	o.SetFunc("profile", noop)
	o.SetFunc("profileEnd", noop)
	o.SetFunc("timeStamp", noop)
	return o
}

func (ci *consoleInstance) log(this js.Object, args []js.Value) (js.Value, error) {
	ci.writeLine(ci.out, util.Format(args))
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) warn(this js.Object, args []js.Value) (js.Value, error) {
	ci.writeLine(ci.errOut, util.Format(args))
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) dir(this js.Object, args []js.Value) (js.Value, error) {
	target := ci.e.JS.Undefined()
	if len(args) > 0 {
		target = args[0]
	}
	ci.writeLine(ci.out, util.Inspect(target))
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) assert(this js.Object, args []js.Value) (js.Value, error) {
	ok := false
	if len(args) > 0 {
		ok = args[0].Bool()
	}
	if ok {
		return ci.e.JS.Undefined(), nil
	}
	msg := "Assertion failed"
	if len(args) > 1 {
		msg += ": " + util.Format(args[1:])
	}
	ci.writeLine(ci.errOut, msg)
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) trace(this js.Object, args []js.Value) (js.Value, error) {
	msg := "Trace"
	if len(args) > 0 {
		msg += ": " + util.Format(args)
	}
	ci.writeLine(ci.errOut, msg)
	errObj := ci.e.JS.Error(msg)
	stack := errObj.Get("stack").Str()
	lines := strings.Split(stack, "\n")
	for i, line := range lines {
		if i == 0 {
			continue // repeats the Error: <msg> headline
		}
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		ci.writeLine(ci.errOut, "    "+strings.TrimSpace(line))
	}
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) count(this js.Object, args []js.Value) (js.Value, error) {
	label, err := countLabel(args)
	if err != nil {
		return js.Value{}, err
	}
	ci.counts[label]++
	ci.writeLine(ci.out, fmt.Sprintf("%s: %d", label, ci.counts[label]))
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) countReset(this js.Object, args []js.Value) (js.Value, error) {
	label, err := countLabel(args)
	if err != nil {
		return js.Value{}, err
	}
	if _, ok := ci.counts[label]; !ok {
		ci.writeLine(ci.errOut, fmt.Sprintf("Warning: Count for '%s' does not exist", label))
		return ci.e.JS.Undefined(), nil
	}
	// Node deletes the entry so the next count() starts at one again.
	delete(ci.counts, label)
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) time(this js.Object, args []js.Value) (js.Value, error) {
	label := labelArg(args, "default")
	if _, ok := ci.times[label]; ok {
		ci.writeLine(ci.errOut, fmt.Sprintf("Warning: Label '%s' already exists for console.time()", label))
		return ci.e.JS.Undefined(), nil
	}
	ci.times[label] = time.Now()
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) timeLog(this js.Object, args []js.Value) (js.Value, error) {
	label := labelArg(args, "default")
	start, ok := ci.times[label]
	if !ok {
		ci.writeLine(ci.errOut, fmt.Sprintf("Warning: No such label '%s' for console.timeLog()", label))
		return ci.e.JS.Undefined(), nil
	}
	msg := fmt.Sprintf("%s: %.3fms", label, elapsed(start))
	if len(args) > 1 {
		msg += " " + util.Format(args[1:])
	}
	ci.writeLine(ci.out, msg)
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) timeEnd(this js.Object, args []js.Value) (js.Value, error) {
	label := labelArg(args, "default")
	start, ok := ci.times[label]
	if !ok {
		ci.writeLine(ci.errOut, fmt.Sprintf("Warning: No such label '%s' for console.timeEnd()", label))
		return ci.e.JS.Undefined(), nil
	}
	delete(ci.times, label)
	ci.writeLine(ci.out, fmt.Sprintf("%s: %.3fms", label, elapsed(start)))
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) group(this js.Object, args []js.Value) (js.Value, error) {
	if len(args) > 0 {
		ci.writeLine(ci.out, util.Format(args))
	}
	ci.indent += ci.groupWidth
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) groupEnd(this js.Object, args []js.Value) (js.Value, error) {
	if ci.indent >= ci.groupWidth {
		ci.indent -= ci.groupWidth
	} else {
		ci.indent = 0
	}
	return ci.e.JS.Undefined(), nil
}

// clear mirrors internal/console/constructor.js: only a TTY (and not TERM=dumb)
// is worth clearing, and the sequence is cursorTo(0,0) + clearScreenDown.
func (ci *consoleInstance) clear(this js.Object, args []js.Value) (js.Value, error) {
	if os.Getenv("TERM") == "dumb" {
		return ci.e.JS.Undefined(), nil
	}
	if !ci.stdoutIsTTY() {
		return ci.e.JS.Undefined(), nil
	}
	_, _ = io.WriteString(ci.out, "\x1b[1;1H\x1b[0J")
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) stdoutIsTTY() bool {
	if ci.lazyOut != "" {
		if st := ci.processStream(ci.lazyOut); st.IsObject() {
			return st.Object().Get("isTTY").Bool()
		}
		return false
	}
	if ci.outTarget.Valid() && ci.outTarget.IsObject() {
		return ci.outTarget.Object().Get("isTTY").Bool()
	}
	if w, ok := ci.out.(*jsWriter); ok && w.target.Valid() && w.target.IsObject() {
		return w.target.Object().Get("isTTY").Bool()
	}
	if f, ok := ci.out.(*os.File); ok {
		if st, err := f.Stat(); err == nil {
			return st.Mode()&os.ModeCharDevice != 0
		}
	}
	return false
}

func (ci *consoleInstance) processStream(name string) js.Value {
	proc := ci.e.Global().Get("process")
	if !proc.IsObject() {
		return js.Value{}
	}
	return proc.Object().Get(name)
}

// tableArgType renders the received value for console.table's validation.
func tableArgType(e *env.Env, v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsString():
		return "type string ('" + v.Str() + "')"
	case v.IsNumber():
		return "type number (" + strconv.FormatFloat(v.Float(), 'f', -1, 64) + ")"
	case v.IsObject():
		return "an instance of " + v.Tag()
	}
	return "an unsupported value"
}

func (ci *consoleInstance) table(this js.Object, args []js.Value) (js.Value, error) {
	if len(args) == 0 {
		ci.writeLine(ci.out, "")
		return ci.e.JS.Undefined(), nil
	}
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsArray() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "properties" argument must be an instance of Array. Received `+tableArgType(ci.e, args[1]))
	}
	data := args[0]
	if !data.IsArray() && !data.IsObject() {
		// Node logs non-object tabular data through console.log, so strings
		// keep their quotes-free form.
		if data.IsString() {
			ci.writeLine(ci.out, data.Str())
		} else {
			ci.writeLine(ci.out, util.Inspect(data))
		}
		return ci.e.JS.Undefined(), nil
	}

	var columns []string
	if len(args) > 1 && args[1].IsArray() {
		arr := args[1].Object()
		for i := 0; i < arr.Length(); i++ {
			columns = append(columns, arr.At(i).Str())
		}
	}

	type tableRow struct {
		index  string
		values []js.Value
	}
	var rows []tableRow

	switch {
	case data.IsArray():
		arr := data.Object()
		n := arr.Length()
		objects := n > 0
		for i := 0; i < n; i++ {
			v := arr.At(i)
			if !v.IsObject() || v.IsArray() || v.IsFunction() || isBytes(v) {
				objects = false
				break
			}
		}
		if objects {
			if len(columns) == 0 {
				seen := map[string]bool{}
				for i := 0; i < n; i++ {
					for _, k := range arr.At(i).Object().OwnEnumerableKeys() {
						if !seen[k] {
							seen[k] = true
							columns = append(columns, k)
						}
					}
				}
			}
			for i := 0; i < n; i++ {
				obj := arr.At(i).Object()
				vals := make([]js.Value, len(columns))
				for j, c := range columns {
					vals[j] = obj.Get(c)
				}
				rows = append(rows, tableRow{strconv.Itoa(i), vals})
			}
		} else {
			columns = []string{"Values"}
			for i := 0; i < n; i++ {
				rows = append(rows, tableRow{strconv.Itoa(i), []js.Value{arr.At(i)}})
			}
		}
	default:
		obj := data.Object()
		keys := obj.OwnEnumerableKeys()
		if len(columns) == 0 {
			allObjects := len(keys) > 0
			for _, k := range keys {
				v := obj.Get(k)
				if !v.IsObject() || v.IsArray() || v.IsFunction() || isBytes(v) {
					allObjects = false
					break
				}
			}
			if allObjects {
				seen := map[string]bool{}
				for _, k := range keys {
					for _, c := range obj.Get(k).Object().OwnEnumerableKeys() {
						if !seen[c] {
							seen[c] = true
							columns = append(columns, c)
						}
					}
				}
			}
		}
		if len(columns) == 0 {
			columns = []string{"Values"}
		}
		for _, k := range keys {
			v := obj.Get(k)
			if len(columns) == 1 && columns[0] == "Values" {
				rows = append(rows, tableRow{k, []js.Value{v}})
				continue
			}
			vals := make([]js.Value, len(columns))
			for j, c := range columns {
				if v.IsObject() {
					vals[j] = v.Object().Get(c)
				}
			}
			rows = append(rows, tableRow{k, vals})
		}
	}

	if len(rows) == 0 {
		ci.writeLine(ci.out, util.Inspect(data))
		return ci.e.JS.Undefined(), nil
	}

	headers := append([]string{"(index)"}, columns...)
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = displayWidth(h)
	}
	rowCells := make([][]string, len(rows))
	for i, r := range rows {
		cells := make([]string, len(headers))
		cells[0] = r.index
		for j := range columns {
			if j < len(r.values) {
				cells[j+1] = util.Inspect(r.values[j])
			}
		}
		rowCells[i] = cells
		for j, c := range cells {
			if w := displayWidth(c); w > widths[j] {
				widths[j] = w
			}
		}
	}

	var b strings.Builder
	border := func(left, mid, right string) string {
		var parts []string
		for _, w := range widths {
			parts = append(parts, strings.Repeat("─", w+2))
		}
		return left + strings.Join(parts, mid) + right
	}
	line := func(cells []string) string {
		var parts []string
		for i, c := range cells {
			parts = append(parts, " "+c+strings.Repeat(" ", widths[i]-displayWidth(c))+" ")
		}
		return "│" + strings.Join(parts, "│") + "│"
	}

	b.WriteString(border("┌", "┬", "┐"))
	b.WriteByte('\n')
	b.WriteString(line(headers))
	b.WriteByte('\n')
	b.WriteString(border("├", "┼", "┤"))
	b.WriteByte('\n')
	for _, cells := range rowCells {
		b.WriteString(line(cells))
		b.WriteByte('\n')
	}
	b.WriteString(border("└", "┴", "┘"))
	// Node passes the whole table to console.log(), so the stream sees one
	// write per table rather than one per rendered line.
	ci.writeLine(ci.out, b.String())
	return ci.e.JS.Undefined(), nil
}

func (ci *consoleInstance) writeLine(w io.Writer, msg string) {
	var b strings.Builder
	for i, line := range strings.Split(msg, "\n") {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(strings.Repeat(" ", ci.indent))
		b.WriteString(line)
	}
	b.WriteByte('\n')
	_, _ = io.WriteString(w, b.String())
}

func (ci *consoleInstance) writeLineRaw(w io.Writer, msg string) {
	_, _ = io.WriteString(w, strings.Repeat(" ", ci.indent)+msg+"\n")
}

func labelArg(args []js.Value, fallback string) string {
	if len(args) == 0 || args[0].IsUndefined() {
		return fallback
	}
	if s, ok := args[0].StrOK(); ok && s != "" {
		return s
	}
	return fallback
}

func elapsed(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds()) / 1e6
}

// hasWrite reports whether v looks like a writable stream. Node treats any
// object with a write function as a stream, not as Console options.
func hasWrite(v js.Value) bool {
	return v.Valid() && v.IsObject() && v.Object().Get("write").IsFunction()
}

func consoleWritableErr(v js.Value, name string) error {
	received := "undefined"
	if v.Valid() && !v.IsUndefined() {
		received = "an instance of Object"
	}
	return js.ThrowCode("TypeError", "ERR_CONSOLE_WRITABLE_STREAM",
		`The "`+name+`" option must be a writable stream. Received `+received)
}

func groupIndentation(v js.Value) (int, error) {
	if !v.IsNumber() {
		return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "groupIndentation" argument must be of type number. Received `+describedValue(v))
	}
	if math.IsNaN(v.Float()) || math.Trunc(v.Float()) != v.Float() {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			`The value of "groupIndentation" is out of range. It must be an integer. Received `+
				strconv.FormatFloat(v.Float(), 'f', -1, 64))
	}
	n := v.Int()
	if n < 0 || n > 1000 {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			`The value of "groupIndentation" is out of range. It must be >= 0 && <= 1000. Received `+
				strconv.Itoa(n))
	}
	return n, nil
}

// describedValue matches Node's "Received ..." clause closely enough for the
// option validators here.
func describedValue(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		return "type boolean (" + strconv.FormatBool(v.Bool()) + ")"
	case v.IsString():
		return "type string ('" + v.Str() + "')"
	case v.IsNumber():
		return "type number (" + strconv.FormatFloat(v.Float(), 'f', -1, 64) + ")"
	case v.IsArray():
		return "an instance of Array"
	case v.IsFunction():
		return "function"
	}
	return "an instance of Object"
}

func countLabel(args []js.Value) (string, error) {
	if len(args) == 0 || args[0].IsUndefined() {
		return "default", nil
	}
	s, ok := args[0].StrOK()
	if !ok {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			"Cannot convert a Symbol value to a string")
	}
	return s, nil
}

func isBytes(v js.Value) bool {
	if !v.IsObject() {
		return false
	}
	_, ok := v.Bytes()
	return ok
}

func displayWidth(s string) int {
	return len([]rune(s))
}

func writerFor(v js.Value, fallback io.Writer) io.Writer {
	if v.Valid() && v.IsObject() && v.Object().Get("write").IsFunction() {
		return &jsWriter{target: v}
	}
	return fallback
}

// jsWriter adapts a Node-style writable (an object with write(chunk)) to
// io.Writer. write is resolved on every call so replacing stream.write later
// (as the test harnesses do) takes effect.
type jsWriter struct {
	target js.Value
}

func (w *jsWriter) Write(p []byte) (int, error) {
	if w.target.Ctx() == nil || !w.target.IsObject() {
		return len(p), nil
	}
	fn := w.target.Object().Get("write")
	if !fn.IsFunction() {
		return len(p), nil
	}
	_, err := fn.Object().Call(w.target.Object(), w.target.Ctx().Str(string(p)))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// dynamicWriter resolves process.<name>.write at write time, which is how the
// global console follows process.stdout being replaced or hijacked.
type dynamicWriter struct {
	e        *env.Env
	name     string
	fallback io.Writer
}

func (w *dynamicWriter) Write(p []byte) (int, error) {
	if proc := w.e.Global().Get("process"); proc.IsObject() {
		if st := proc.Object().Get(w.name); st.IsObject() {
			fn := st.Object().Get("write")
			if fn.IsFunction() {
				if _, err := fn.Object().Call(st.Object(), w.e.JS.Str(string(p))); err != nil {
					return 0, err
				}
				return len(p), nil
			}
		}
	}
	if w.fallback != nil {
		return w.fallback.Write(p)
	}
	return len(p), nil
}
