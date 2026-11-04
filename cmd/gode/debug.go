package main

// The debugger instruments every safe statement boundary in the entry script
// with a call to a host hook. The hook receives a closure created in the
// script's lexical scope, so `eval` runs against the real locals of the paused
// line instead of a detached sandbox.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"godejs/internal/env"
	"godejs/internal/js"
)

type debugger struct {
	e      *env.Env
	file   string
	source string
	in     *bufio.Reader

	breaks map[string]bool
	entry  bool
	step   bool
	quit   bool

	evalFn     js.Value
	curFile    string
	curLine    int
	lineMap    map[int]int
	watches    []string
	lastPrompt string
}

var (
	dbgDeclRE  = regexp.MustCompile(`\b(?:var|let|const|function|class)\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	dbgFuncRE  = regexp.MustCompile(`function\s*[A-Za-z0-9_$]*\s*\(([^)]*)\)`)
	dbgArrowRE = regexp.MustCompile(`\(([^()]*)\)\s*=>`)
	dbgCatchRE = regexp.MustCompile(`catch\s*\(([^)]*)\)`)
	dbgIdentRE = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)
)

func runDebug(e *env.Env, file string) error {
	abs, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		abs = real
	}
	raw, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	instrumented, hooks, lineMap := instrumentSource(abs, string(raw))
	if hooks == 0 {
		instrumented = "__godeDbg(" + strconv.Quote(abs) +
			", 1, function (__godeExpr) { return eval(__godeExpr); });\n" + instrumented
	}
	d := &debugger{
		e:       e,
		file:    abs,
		source:  string(raw),
		in:      bufio.NewReader(e.In()),
		breaks:  map[string]bool{},
		lineMap: lineMap,
		entry:   true,
	}
	e.SetSourceOverride(abs, instrumented)
	e.Global().SetFunc("__godeDbg", d.hook)
	fmt.Fprintf(e.Out(), "gode debugger: %s (%d breakpoint sites)\n", abs, maxInt(hooks, 1))
	fmt.Fprintln(e.Out(), "Paused at entry. Type 'help' for commands, 'continue' to run.")
	_, err = e.RunFile(abs)
	return err
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (d *debugger) hook(_ js.Object, args []js.Value) (js.Value, error) {
	file := dbgArg(args, 0).Str()
	line := int(dbgArg(args, 1).Int())
	d.curFile, d.curLine = file, line
	if fn := dbgArg(args, 2); fn.Valid() && fn.IsFunction() {
		d.evalFn = fn
	}
	key := file + ":" + strconv.Itoa(line)
	short := filepath.Base(file) + ":" + strconv.Itoa(line)
	lineKey := "line:" + strconv.Itoa(line)
	pause := d.entry || d.step || d.breaks[key] || d.breaks[short] || d.breaks[lineKey]
	if d.entry {
		d.entry = false
	}
	if !pause {
		return d.e.JS.Undefined(), nil
	}
	d.step = false
	d.pause()
	return d.e.JS.Undefined(), nil
}

func (d *debugger) pause() {
	out := d.e.Out()
	fmt.Fprintf(out, "\n--- paused at %s:%d\n", d.curFile, d.curLine)
	d.printCurrentLine(out)
	for _, w := range d.watches {
		if v, err := d.eval(w); err != nil {
			fmt.Fprintf(out, "watch %s = <error: %s>\n", w, err)
		} else {
			fmt.Fprintf(out, "watch %s = %s\n", w, d.format(v))
		}
	}
	for {
		fmt.Fprint(out, "(gode) ")
		text, err := d.in.ReadString('\n')
		if err != nil && text == "" {
			fmt.Fprintln(out)
			return
		}
		fields := strings.Fields(strings.TrimSpace(text))
		if len(fields) == 0 {
			continue
		}
		cmd := fields[0]
		arg := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), cmd))
		switch cmd {
		case "help", "h", "?":
			d.help(out)
		case "continue", "c", "run", "r":
			return
		case "step", "s", "next", "n":
			d.step = true
			return
		case "quit", "q", "exit":
			d.quit = true
			os.Exit(0)
		case "bt", "where", "backtrace":
			if v, err := d.eval("new Error('pause').stack"); err != nil {
				fmt.Fprintf(out, "stack unavailable: %s\n", err)
			} else {
				fmt.Fprintln(out, mapDebugStack(d.file, v.Str(), d.lineMap))
			}
		case "list", "l":
			d.list(out, arg)
		case "break", "b":
			d.addBreak(out, arg)
		case "delete", "d":
			d.delBreak(out, arg)
		case "info", "i":
			if arg == "breakpoints" || arg == "b" || arg == "" {
				d.listBreaks(out)
			} else {
				fmt.Fprintln(out, "usage: info breakpoints")
			}
		case "locals", "loc":
			d.locals(out)
		case "eval", "p", "print":
			if arg == "" {
				fmt.Fprintln(out, "usage: eval <expression>")
				continue
			}
			if v, err := d.eval(arg); err != nil {
				fmt.Fprintf(out, "error: %s\n", err)
			} else {
				fmt.Fprintln(out, d.format(v))
			}
		case "watch":
			if arg == "" {
				fmt.Fprintln(out, "usage: watch <expression>")
				continue
			}
			d.watches = append(d.watches, arg)
			fmt.Fprintf(out, "watching %s\n", arg)
		case "unwatch":
			d.watches = nil
			fmt.Fprintln(out, "watches cleared")
		case "clear":
			d.breaks = map[string]bool{}
			fmt.Fprintln(out, "breakpoints cleared")
		default:
			fmt.Fprintf(out, "unknown command %q (try 'help')\n", cmd)
		}
	}
}

func (d *debugger) help(out interface{ Write([]byte) (int, error) }) {
	fmt.Fprint(out, `commands:
  break <line|file:line>   set a breakpoint (default: current file)
  delete <line|file:line>  remove a breakpoint
  info breakpoints         list breakpoints
  continue                 run until the next breakpoint
  step                     pause at the next statement
  next                     alias for step
  locals                   evaluate declarations visible at this line
  eval <expr>              evaluate an expression in the current scope
  watch <expr>             print an expression at every pause
  bt                       print the JavaScript call stack
  list [line]              show source around the current line
  clear                    clear all breakpoints
  quit                     exit the process
`)
}

func (d *debugger) printCurrentLine(out interface{ Write([]byte) (int, error) }) {
	lines := strings.Split(d.source, "\n")
	idx := d.curLine - 1
	if idx < 0 || idx >= len(lines) {
		return
	}
	fmt.Fprintf(out, "%5d | %s\n", d.curLine, lines[idx])
}

func (d *debugger) list(out interface{ Write([]byte) (int, error) }, arg string) {
	center := d.curLine
	if arg != "" {
		if n, err := strconv.Atoi(arg); err == nil {
			center = n
		}
	}
	lines := strings.Split(d.source, "\n")
	start, end := center-5, center+5
	if start < 1 {
		start = 1
	}
	if end > len(lines) {
		end = len(lines)
	}
	for n := start; n <= end; n++ {
		mark := " "
		if n == d.curLine {
			mark = ">"
		}
		fmt.Fprintf(out, "%s%5d | %s\n", mark, n, lines[n-1])
	}
}

func (d *debugger) addBreak(out interface{ Write([]byte) (int, error) }, arg string) {
	if arg == "" {
		arg = strconv.Itoa(d.curLine)
	}
	if !strings.Contains(arg, ":") {
		if _, err := strconv.Atoi(arg); err != nil {
			fmt.Fprintln(out, "usage: break <line|file:line>")
			return
		}
		d.breaks["line:"+arg] = true
		fmt.Fprintf(out, "breakpoint at %s:%s\n", d.curFile, arg)
		return
	}
	d.breaks[arg] = true
	fmt.Fprintf(out, "breakpoint at %s\n", arg)
}

func (d *debugger) delBreak(out interface{ Write([]byte) (int, error) }, arg string) {
	if arg == "" {
		arg = strconv.Itoa(d.curLine)
	}
	if !strings.Contains(arg, ":") {
		delete(d.breaks, "line:"+arg)
		delete(d.breaks, d.curFile+":"+arg)
	} else {
		delete(d.breaks, arg)
	}
	fmt.Fprintf(out, "removed breakpoint at %s\n", arg)
}

func (d *debugger) listBreaks(out interface{ Write([]byte) (int, error) }) {
	if len(d.breaks) == 0 {
		fmt.Fprintln(out, "no breakpoints")
		return
	}
	keys := make([]string, 0, len(d.breaks))
	for k := range d.breaks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintln(out, k)
	}
}

func (d *debugger) locals(out interface{ Write([]byte) (int, error) }) {
	names := d.localNames()
	if len(names) == 0 {
		fmt.Fprintln(out, "no declarations visible yet")
		return
	}
	printed := 0
	for _, name := range names {
		v, err := d.eval(name)
		if err != nil {
			// Variables declared later, or in a different branch, are simply
			// not visible at this line.
			continue
		}
		fmt.Fprintf(out, "%s = %s\n", name, d.format(v))
		printed++
	}
	if printed == 0 {
		fmt.Fprintln(out, "no initialized locals visible at this line")
	}
}

func (d *debugger) localNames() []string {
	lines := strings.Split(d.source, "\n")
	end := d.curLine
	if end > len(lines) {
		end = len(lines)
	}
	if end < 1 {
		end = 1
	}
	text := strings.Join(lines[:end], "\n")
	seen := map[string]bool{}
	var names []string
	add := func(part string) {
		part = strings.TrimSpace(strings.SplitN(part, "=", 2)[0])
		part = strings.TrimSpace(strings.Trim(part, "{}[]()"))
		if !dbgIdentRE.MatchString(part) || seen[part] {
			return
		}
		seen[part] = true
		names = append(names, part)
	}
	for _, re := range []*regexp.Regexp{dbgDeclRE, dbgFuncRE, dbgArrowRE, dbgCatchRE} {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			for _, part := range strings.Split(m[1], ",") {
				add(part)
			}
		}
	}
	if len(names) > 40 {
		names = names[:40]
	}
	return names
}

func (d *debugger) eval(expr string) (js.Value, error) {
	if !d.evalFn.Valid() || !d.evalFn.IsFunction() {
		return js.Value{}, fmt.Errorf("no active JavaScript scope")
	}
	return d.evalFn.Object().Call(js.Object{}, d.e.JS.Str(expr))
}

func (d *debugger) format(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return "undefined"
	}
	if v.IsNull() {
		return "null"
	}
	if v.IsString() {
		return strconv.Quote(v.Str())
	}
	if util := d.e.Exports("util"); util.Valid() {
		if inspect := util.Get("inspect"); inspect.IsFunction() {
			opts := d.e.JS.Obj()
			opts.SetNum("depth", 2)
			opts.SetBool("colors", false)
			if out, err := inspect.Object().Call(util, v, opts.V()); err == nil && out.IsString() {
				return out.Str()
			}
		}
	}
	return v.Str()
}

// instrumentSource inserts __godeDbg hooks before statement boundaries. The
// scanner intentionally skips continuations, object/class members, labels and
// multi-line expressions so the instrumented script keeps the original
// semantics.
func instrumentSource(file, source string) (string, int, map[int]int) {
	lines := strings.Split(source, "\n")
	var out strings.Builder
	var paren, bracket, brace int
	var quote byte
	var block, escaped bool
	prev := ""
	classDepth := -1
	hooks := 0
	outLine := 1
	lineMap := map[int]int{}
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		clean := quote == 0 && !block && paren == 0 && bracket == 0
		inClass := classDepth >= 0 && brace >= classDepth
		if clean && !inClass && shouldHook(trim, prev, brace) {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			out.WriteString(indent)
			fmt.Fprintf(&out, "__godeDbg(%s, %d, function (__godeExpr) { return eval(__godeExpr); });\n", strconv.Quote(file), i+1)
			hooks++
			outLine++
		}
		out.WriteString(line)
		lineMap[outLine] = i + 1
		outLine++
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
		scanLine(line, &paren, &bracket, &brace, &quote, &block, &escaped)
		if classDepth >= 0 && brace < classDepth {
			classDepth = -1
		}
		if clean && strings.Contains(trim, "class ") && strings.HasSuffix(trim, "{") {
			classDepth = brace
		}
		if trim != "" && !strings.HasPrefix(trim, "//") && !strings.HasPrefix(trim, "*") {
			prev = trim
		}
	}
	return out.String(), hooks, lineMap
}

var dbgStackRE = regexp.MustCompile(`^(.*):(\d+):(\d+)$`)

// mapDebugStack rewrites instrumented line numbers in an Error.stack back to
// the original source lines the user sees in the debugger.
func mapDebugStack(file, stack string, lineMap map[int]int) string {
	lines := strings.Split(stack, "\n")
	for i, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		m := dbgStackRE.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		if m[1] != file && filepath.Base(m[1]) != filepath.Base(file) {
			continue
		}
		n, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		orig, ok := lineMap[n]
		if !ok {
			continue
		}
		lines[i] = strings.TrimSuffix(trimmed, ":"+m[2]+":"+m[3]) + ":" + strconv.Itoa(orig) + ":" + m[3]
	}
	return strings.Join(lines, "\n")
}

func shouldHook(trim, prev string, brace int) bool {
	if trim == "" || strings.HasPrefix(trim, "//") || strings.HasPrefix(trim, "/*") ||
		strings.HasPrefix(trim, "*") || strings.HasPrefix(trim, "#!") {
		return false
	}
	if strings.HasPrefix(trim, "\"use strict\"") || strings.HasPrefix(trim, "'use strict'") {
		return false
	}
	for _, p := range []string{"}", ")", "]", ".", ",", "?", ":", "+", "-", "*", "/", "%", "&", "|", "^", "=", "<", ">", "!"} {
		if strings.HasPrefix(trim, p) {
			return false
		}
	}
	for _, kw := range []string{"else", "catch", "finally", "case ", "default:", "import ", "export "} {
		if strings.HasPrefix(trim, kw) {
			return false
		}
	}
	if strings.HasPrefix(trim, "while (") {
		return false
	}
	if strings.HasSuffix(trim, ",") {
		return false
	}
	if prev != "" {
		if !strings.HasSuffix(prev, ";") && !strings.HasSuffix(prev, "{") &&
			!strings.HasSuffix(prev, "}") && !strings.HasSuffix(prev, ":") {
			return false
		}
		for _, kw := range []string{"if ", "for ", "while (", "with (", "else", "do"} {
			if strings.HasPrefix(prev, kw) && !strings.HasSuffix(prev, "{") && !strings.HasSuffix(prev, ";") {
				return false
			}
		}
	}
	if brace > 0 {
		if idx := strings.Index(trim, ":"); idx > 0 && dbgIdentRE.MatchString(strings.TrimSpace(trim[:idx])) {
			return false
		}
		if dbgIdentRE.MatchString(trim) {
			return false
		}
		if strings.HasSuffix(trim, "{") && strings.Contains(trim, "(") && !dbgStartsControl(trim) {
			return false
		}
	}
	return true
}

func dbgStartsControl(s string) bool {
	for _, kw := range []string{"if ", "for ", "while ", "switch ", "try", "function ", "class ", "async ", "const ", "let ", "var ", "return ", "throw ", "do ", "else ", "catch ", "finally"} {
		if strings.HasPrefix(s, kw) {
			return true
		}
	}
	return false
}

func scanLine(line string, paren, bracket, brace *int, quote *byte, block, escaped *bool) {
	for i := 0; i < len(line); i++ {
		c := line[i]
		if *block {
			if c == '*' && i+1 < len(line) && line[i+1] == '/' {
				*block = false
				i++
			}
			continue
		}
		if *quote != 0 {
			if *escaped {
				*escaped = false
				continue
			}
			if c == '\\' {
				*escaped = true
				continue
			}
			if c == *quote {
				*quote = 0
			}
			continue
		}
		if c == '/' && i+1 < len(line) {
			if line[i+1] == '/' {
				return
			}
			if line[i+1] == '*' {
				*block = true
				i++
				continue
			}
		}
		switch c {
		case '\'', '"', '`':
			*quote = c
		case '(':
			*paren++
		case ')':
			if *paren > 0 {
				*paren--
			}
		case '[':
			*bracket++
		case ']':
			if *bracket > 0 {
				*bracket--
			}
		case '{':
			*brace++
		case '}':
			if *brace > 0 {
				*brace--
			}
		}
	}
}

func dbgArg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}
