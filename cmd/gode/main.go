// Command gode runs JavaScript on JavaScriptCore with Node-compatible core
// modules.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"godejs/internal/all"
	"godejs/internal/diag"
	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/internal/version"
)

func main() { os.Exit(run(os.Args)) }

type options struct {
	eval                 string
	hasEval              bool
	print                bool
	check                bool
	interactive          bool
	script               string
	args                 []string
	help                 bool
	version              bool
	completionBash       bool
	preserveSymlinks     bool
	preserveSymlinksMain bool
	entryURL             bool
	debug                bool
	watch                bool
	watchPaths           []string
	watchPreserve        bool
	watchKill            string
	envFiles             []string
	envFilesOptional     []string
	runScript            string
	inputType            string
	title                string
	stackTraceLimit      int
	exposeGC             bool
	frozenIntrinsics     bool
	disableProto         string
	traceExit            bool
	execArgv             []string
	preloads             []string
	cfg                  *cliConfig
}

func run(argv []string) int {
	// The process owns a single JavaScriptCore context. Pin the goroutine that
	// creates and drives it to this OS thread for the life of the process;
	// JavaScriptCore keeps thread-affine VM state.
	runtime.LockOSThread()
	opts, err := parseArgs(argv[1:])
	if err != nil {
		fmt.Fprint(os.Stderr, diag.Render(diag.Diagnostic{
			Program: "gode",
			Message: err.Error(),
			Help:    "run 'gode --help' for usage",
		}))
		return 9
	}
	if opts.help {
		usage(os.Stdout)
		return 0
	}
	if opts.version {
		fmt.Println(version.NodeVersion)
		return 0
	}
	if opts.completionBash {
		printCompletion(os.Stdout)
		return 0
	}
	if reason := opts.unsupportedReason(); reason != "" {
		fmt.Fprintf(os.Stderr, "gode: %s\n", reason)
		return 9
	}

	environment := os.Environ()
	if len(opts.envFiles)+len(opts.envFilesOptional) > 0 {
		environment, err = mergeEnvFiles(environment, opts.envFiles, opts.envFilesOptional)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gode: %s\n", err)
			return 9
		}
	}
	if opts.runScript != "" {
		return runPackageScript(opts, environment)
	}
	if opts.watch && os.Getenv("GODE_WATCH_CHILD") != "1" {
		return watchLoop(argv, opts)
	}
	if opts.entryURL && opts.script != "" {
		path, err := entryPath(opts.script)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gode: %s\n", err)
			return 9
		}
		opts.script = path
	}
	opts = applyScriptFlags(opts)
	if opts.debug && (opts.hasEval || opts.script == "") {
		fmt.Fprintln(os.Stderr, "gode: --debug requires a script file")
		return 9
	}

	evalMode := opts.hasEval || (opts.print && opts.script != "" && !fileExists(opts.script))

	if opts.check {
		source, file, err := checkSource(opts)
		if err != nil {
			fmt.Fprint(os.Stderr, diag.Render(diag.Diagnostic{Program: "gode", Message: err.Error()}))
			return 1
		}
		e := env.New(env.Options{
			Args:     []string{argv[0]},
			Env:      environment,
			Cwd:      mustGetwd(),
			ExecPath: argv[0],
			Version:  version.NodeVersion,
			ExecArgv: opts.execArgv,
			Stdout:   os.Stdout,
			Stderr:   os.Stderr,
			Stdin:    os.Stdin,
		})
		defer e.JS.Close()
		if err := e.CheckScript(source, file, formatForInputWithSource(opts.inputType, file, source)); err != nil {
			fmt.Fprint(os.Stderr, diag.Render(diag.FromError("gode", err, file, source)))
			return 1
		}
		return 0
	}

	argvForProcess := []string{argv[0]}
	if !evalMode && opts.script != "" {
		argvForProcess = append(argvForProcess, opts.script)
		argvForProcess = append(argvForProcess, opts.args...)
	}

	e := env.New(env.Options{
		Args:             argvForProcess,
		Env:              environment,
		Cwd:              mustGetwd(),
		ExecPath:         argv[0],
		Version:          version.NodeVersion,
		ExecArgv:         opts.execArgv,
		PreserveSymlinks: opts.preserveSymlinks || opts.preserveSymlinksMain,
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
		Stdin:            os.Stdin,
	})
	defer e.JS.Close()
	if opts.traceExit {
		// Register after Close so this deferred trace runs first (defers run
		// last-in, first-out).
		defer func() { printTraceExit(e) }()
	}
	e.InstallAll()
	all.InstallGlobals(e)
	if opts.exposeGC {
		e.Global().SetFunc("gc", func(this js.Object, args []js.Value) (js.Value, error) {
			runtime.GC()
			return e.JS.Undefined(), nil
		})
	}
	applyStartupFlags(e, opts)

	for _, pre := range opts.preloads {
		if _, err := e.RequireFrom(mustGetwd(), pre); err != nil {
			fmt.Fprint(os.Stderr, diag.Render(diag.FromError("gode", err, pre, "")))
			return 1
		}
	}

	var source, file string
	var runErr error
	replMode := false

	switch {
	case evalMode:
		source = opts.eval
		if !opts.hasEval {
			source = opts.script
		}
		file = "[eval]"
		// Node's -e context exposes builtins as globals; tests and snippets rely
		// on `os`, `path`, and friends without require().
		if osExports := e.Exports("os"); osExports.Valid() {
			e.Global().Set("os", osExports.V())
		}
		if vmExports := e.Exports("vm"); vmExports.Valid() {
			e.Global().Set("vm", vmExports.V())
		}
		val, err := e.RunScriptType(source, file, formatForInputWithSource(opts.inputType, file, source))
		runErr = err
		if err == nil && opts.print {
			printValue(e, val)
		}
	case opts.script != "":
		file = opts.script
		source, _ = readSource(file)
		if opts.debug {
			runErr = runDebug(e, file)
			break
		}
		if me := e.Exports("module"); me.Valid() {
			if rm := me.Get("runMain"); rm.IsObject() && !rm.IsNull() {
				_, runErr = rm.Object().Call(me, e.JS.Str(file))
			} else {
				_, runErr = e.RunFile(file)
			}
		} else {
			_, runErr = e.RunFile(file)
		}
	case opts.interactive && !isPipe(os.Stdin):
		replMode = true
		runErr = repl(e)
	case isPipe(os.Stdin):
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "gode:", err)
			return 1
		}
		source = string(data)
		file = "[stdin]"
		val, err := e.RunScriptType(source, file, formatForInputWithSource(opts.inputType, file, source))
		runErr = err
		if err == nil && opts.print {
			printValue(e, val)
		}
	default:
		replMode = true
		runErr = repl(e)
	}

	if runErr == nil && opts.interactive && !replMode {
		runErr = repl(e)
	}
	if runErr == nil {
		runErr = e.RunMain()
	}
	// An 'uncaughtException' listener (or the capture callback) may absorb the
	// exception; only an error nobody handles ends the program.
	for runErr != nil && !e.Exited && !errors.Is(runErr, env.ErrProcessExit) {
		if !e.DeliverUncaught(runErr, "uncaughtException") {
			break
		}
		runErr = e.RunMain()
	}
	e.FireExitHooks()
	if runErr != nil && !e.Exited && !errors.Is(runErr, env.ErrProcessExit) {
		fmt.Fprint(os.Stderr, diag.Render(diag.FromError("gode", runErr, file, source)))
		code := e.ExitCode
		if code == 0 {
			code = 1
		}
		e.ExitCode = code
		return code
	}
	// process.exitCode (and an exit hook that sets it) must decide the status
	// even when the program ended by draining the event loop.
	return e.ExitCode
}

func parseArgs(args []string) (options, error) {
	cfg, err := parseCLI(args)
	if err != nil {
		return options{}, err
	}
	o := options{cfg: cfg, execArgv: cfg.execArgv}
	o.help = cfg.has("help")
	o.version = cfg.has("version")
	o.check = cfg.has("check")
	o.interactive = cfg.has("interactive")
	o.debug = cfg.has("debug")
	o.completionBash = cfg.has("completion-bash")
	o.preserveSymlinks = cfg.has("preserve-symlinks")
	o.preserveSymlinksMain = cfg.has("preserve-symlinks-main")
	o.entryURL = cfg.has("entry-url")
	o.watch = cfg.has("watch")
	o.watchPreserve = cfg.has("watch-preserve-output")
	o.watchPaths = cfg.all("watch-path")
	o.watchKill = cfg.value("watch-kill-signal")
	if o.watchKill == "" {
		o.watchKill = "SIGTERM"
	}
	o.envFiles = cfg.all("env-file")
	o.envFilesOptional = cfg.all("env-file-if-exists")
	o.runScript = cfg.value("run")
	o.inputType = cfg.value("input-type")
	switch o.inputType {
	case "", "commonjs", "module":
	default:
		return o, fmt.Errorf("--input-type must be 'commonjs' or 'module'")
	}
	o.title = cfg.value("title")
	o.exposeGC = cfg.has("expose-gc")
	o.frozenIntrinsics = cfg.has("frozen-intrinsics")
	o.traceExit = cfg.has("trace-exit")
	if value, ok := cfg.last("stack-trace-limit"); ok {
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return o, fmt.Errorf("--stack-trace-limit must be a non-negative integer")
		}
		o.stackTraceLimit = n
	} else {
		o.stackTraceLimit = -1
	}
	if value, ok := cfg.last("disable-proto"); ok {
		switch value {
		case "delete", "throw":
			o.disableProto = value
		default:
			return o, fmt.Errorf("--disable-proto must be 'delete' or 'throw'")
		}
	}
	if value, ok := cfg.last("eval"); ok {
		o.eval = value
		o.hasEval = true
	}
	o.print = cfg.has("print")
	if value, ok := cfg.last("print"); ok {
		o.eval = value
		o.hasEval = true
	}
	o.preloads = append(cfg.all("require"), cfg.all("import")...)
	if len(cfg.positionals) > 0 {
		o.script = cfg.positionals[0]
		o.args = cfg.positionals[1:]
	}
	return o, nil
}

func applyStartupFlags(e *env.Env, o options) {
	set := func(src string) {
		if _, err := e.Eval(src); err != nil {
			fmt.Fprintf(e.Err(), "gode: startup option failed: %v\n", err)
		}
	}
	if o.title != "" {
		set("process.title = " + quoteJS(o.title))
	}
	if o.stackTraceLimit >= 0 {
		set(fmt.Sprintf("Error.stackTraceLimit = %d", o.stackTraceLimit))
	}
	if o.cfg.has("throw-deprecation") {
		set("process.throwDeprecation = true")
	}
	if o.cfg.has("no-deprecation") {
		set("process.noDeprecation = true")
	}
	if o.cfg.has("trace-deprecation") {
		set("process.traceDeprecation = true")
	}
	if o.cfg.has("pending-deprecation") {
		set("process.pendingDeprecation = true")
	}
	if o.frozenIntrinsics {
		set(frozenIntrinsicsJS)
	}
	switch o.disableProto {
	case "delete":
		set("delete Object.prototype.__proto__")
	case "throw":
		set(disableProtoThrowJS)
	}
	if props := reportProperties(o.cfg); len(props) > 0 {
		set(strings.Join(props, "; "))
	}
}

func reportProperties(cfg *cliConfig) []string {
	var props []string
	addString := func(flag, prop string) {
		if value, ok := cfg.last(flag); ok {
			props = append(props, prop+" = "+quoteJS(value))
		}
	}
	addBool := func(flag, prop string) {
		if cfg.has(flag) {
			props = append(props, prop+" = true")
		}
	}
	if _, ok := cfg.last("report-dir"); ok {
		addString("report-dir", "process.report.directory")
	} else {
		addString("diagnostic-dir", "process.report.directory")
	}
	addString("report-filename", "process.report.filename")
	addBool("report-compact", "process.report.compact")
	addBool("report-on-fatalerror", "process.report.reportOnFatalError")
	addBool("report-on-signal", "process.report.reportOnSignal")
	addBool("report-uncaught-exception", "process.report.reportOnUncaughtException")
	addString("report-signal", "process.report.signal")
	return props
}

func runPackageScript(o options, environment []string) int {
	name := o.runScript
	if name == "" {
		fmt.Fprintln(os.Stderr, "gode: --run requires a script name")
		return 9
	}
	path := "package.json"
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gode: --run: %v\n", err)
		return 1
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		fmt.Fprintf(os.Stderr, "gode: --run: %s: %v\n", path, err)
		return 1
	}
	command, ok := pkg.Scripts[name]
	if !ok || strings.TrimSpace(command) == "" {
		fmt.Fprintf(os.Stderr, "gode: --run: no script named %q in %s\n", name, path)
		return 9
	}
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	abs, _ := os.Getwd()
	cmd.Env = append(append([]string{}, environment...),
		"npm_lifecycle_event="+name,
		"npm_package_json="+abs+string(os.PathSeparator)+path,
	)
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "gode: --run: %v\n", err)
		return 1
	}
	return 0
}

func checkSource(o options) (string, string, error) {
	if o.hasEval {
		return o.eval, "[eval]", nil
	}
	if o.script != "" {
		source, err := readSource(o.script)
		return source, o.script, err
	}
	if isPipe(os.Stdin) {
		data, err := io.ReadAll(os.Stdin)
		return string(data), "[stdin]", err
	}
	return "", "", fmt.Errorf("--check requires a script file, --eval, or piped input")
}

func entryPath(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("--entry-url: %v", err)
	}
	if u.Scheme != "file" {
		return "", fmt.Errorf("--entry-url expects a file: URL, got %q", raw)
	}
	return u.Path, nil
}

func printTraceExit(e *env.Env) {
	fmt.Fprintf(e.Err(), "WARNING: Exited the environment with code %d\n", e.ExitCode)
	if value, err := e.Eval("new Error().stack"); err == nil && value.IsString() && value.Str() != "" {
		fmt.Fprintln(e.Err(), value.Str())
	}
}

const frozenIntrinsicsJS = `(function () {
  const names = ["Object", "Function", "Array", "String", "Number", "Boolean",
    "Symbol", "Math", "JSON", "Reflect", "Promise", "Map", "Set", "WeakMap",
    "WeakSet", "WeakRef", "FinalizationRegistry", "Date", "RegExp", "Error",
    "EvalError", "RangeError", "ReferenceError", "SyntaxError", "TypeError",
    "URIError", "AggregateError", "ArrayBuffer", "SharedArrayBuffer", "DataView",
    "Int8Array", "Uint8Array", "Uint8ClampedArray", "Int16Array", "Uint16Array",
    "Int32Array", "Uint32Array", "Float32Array", "Float64Array",
    "BigInt64Array", "BigUint64Array", "Intl", "Temporal"];
  for (const name of names) {
    const value = globalThis[name];
    if (value === undefined) continue;
    if (typeof value === "function" && value.prototype) {
      try { Object.freeze(value.prototype); } catch {}
    }
    try { Object.freeze(value); } catch {}
  }
})();`

const disableProtoThrowJS = `Object.defineProperty(Object.prototype, "__proto__", {
  configurable: false,
  get() {
    const error = new TypeError("Cannot access Object.prototype.__proto__");
    error.code = "ERR_PROTO_ACCESS";
    throw error;
  },
  set() {
    const error = new TypeError("Cannot set Object.prototype.__proto__");
    error.code = "ERR_PROTO_ACCESS";
    throw error;
  }
});`

func repl(e *env.Env) error {
	in := bufio.NewReader(e.In())
	color := !isPipe(os.Stdout) && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
	const (
		green  = "\x1b[32m"
		yellow = "\x1b[33m"
		red    = "\x1b[31m"
		reset  = "\x1b[0m"
	)
	style := func(code, s string) string {
		if !color {
			return s
		}
		return code + s + reset
	}
	basePrompt := "> "
	contPrompt := "... "
	if color {
		basePrompt = green + "> " + reset
		contPrompt = yellow + "... " + reset
	}
	banner := "Welcome to GodeJS " + version.NodeVersion + "!"
	fmt.Fprintln(e.Out(), style(green, banner))
	fmt.Fprintln(e.Out(), "Type \".help\" for more information.")

	prompt := basePrompt
	var buffer string
	for {
		fmt.Fprint(e.Out(), prompt)
		line, err := in.ReadString('\n')
		if err != nil && line == "" {
			fmt.Fprintln(e.Out())
			return nil
		}
		line = strings.TrimRight(line, "\r\n")
		trimmed := strings.TrimSpace(line)
		if trimmed == ".exit" || trimmed == "exit" {
			return nil
		}
		if trimmed == ".clear" {
			buffer = ""
			prompt = basePrompt
			continue
		}
		if buffer == "" {
			switch trimmed {
			case "":
				continue
			case ".help":
				fmt.Fprintln(e.Out(), ".help   Show this help\n.clear  Clear the current input\n.exit   Exit the REPL")
				continue
			}
		}

		buffer += line + "\n"
		if replNeedsMore(buffer) {
			prompt = contPrompt
			continue
		}

		val, evalErr := e.RunScript(buffer, "[repl]")
		if evalErr != nil {
			if replIncomplete(evalErr) {
				prompt = contPrompt
				continue
			}
			fmt.Fprint(e.Err(), style(red, diag.Render(diag.FromError("gode", evalErr, "[repl]", buffer))))
			buffer = ""
			prompt = basePrompt
			continue
		}
		if val.Valid() && !val.IsUndefined() {
			e.Global().Set("_", val)
			printReplValue(e, val, color)
		}
		buffer = ""
		prompt = basePrompt
	}
}

func replNeedsMore(src string) bool {
	depth := 0
	var quote byte
	lineComment := false
	blockComment := false
	escaped := false
	for i := 0; i < len(src); i++ {
		c := src[i]
		if lineComment {
			if c == '\n' {
				lineComment = false
			}
			continue
		}
		if blockComment {
			if c == '*' && i+1 < len(src) && src[i+1] == '/' {
				blockComment = false
				i++
			}
			continue
		}
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"', '`':
			quote = c
		case '/':
			if i+1 < len(src) {
				if src[i+1] == '/' {
					lineComment = true
					i++
				} else if src[i+1] == '*' {
					blockComment = true
					i++
				}
			}
		case '{', '(', '[':
			depth++
		case '}', ')', ']':
			if depth > 0 {
				depth--
			}
		}
	}
	trimmed := strings.TrimRight(src, " \t\r\n")
	return depth > 0 || quote != 0 || blockComment || strings.HasSuffix(trimmed, "\\")
}

func replIncomplete(err error) bool {
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unexpected end") ||
		strings.Contains(s, "unexpected eof") ||
		strings.Contains(s, "unterminated") ||
		strings.Contains(s, "missing )") ||
		strings.Contains(s, "missing }") ||
		strings.Contains(s, "missing ]")
}

func printReplValue(e *env.Env, v js.Value, color bool) {
	if util := e.Exports("util"); util.Valid() {
		if inspect := util.Get("inspect"); inspect.IsFunction() {
			opts := e.JS.Obj()
			opts.SetBool("colors", color)
			opts.Set("depth", e.JS.Num(2))
			if out, err := inspect.Object().Call(util, v, opts.V()); err == nil && out.IsString() {
				fmt.Fprintln(e.Out(), out.Str())
				return
			}
		}
	}
	printValue(e, v)
}

func printValue(e *env.Env, v js.Value) {
	// Use the module's console rather than the native one JavaScriptCore
	// injects, which writes nowhere.
	if log := e.Exports("console").Get("log"); log.Valid() && log.IsFunction() {
		if _, err := log.Object().Call(e.JS.Obj(), v); err == nil {
			return
		}
	}
	fmt.Fprintln(e.Out(), v.Str())
}

func applyScriptFlags(o options) options {
	if o.script == "" || o.hasEval {
		return o
	}
	f, err := os.Open(o.script)
	if err != nil {
		return o
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	head := string(buf[:n])
	if strings.Contains(head, "// Flags:") && strings.Contains(head, "--preserve-symlinks") {
		o.preserveSymlinks = true
		o.execArgv = append(o.execArgv, "--preserve-symlinks")
	}
	return o
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `Usage: gode [options] [script.js] [arguments]

A Node.js-compatible runtime on JavaScriptCore, written in Go.

If no script is given, gode reads a script from stdin or starts the REPL.
Every Node.js CLI flag is recognized. Flags marked below are implemented;
engine-specific flags (V8, OpenSSL, snapshots, SEA, inspector) fail with a
clear message instead of being silently ignored.

Common options:
  -e, --eval=...        evaluate source as JavaScript
  -p, --print           evaluate and print the result
  -c, --check           syntax check without executing
  -i, --interactive     enter the REPL after the script
  -r, --require=...     preload a CommonJS module
      --import=...      preload an ES module
      --env-file=...    load environment variables from a file
      --run=...         run a script from package.json
      --watch           restart when files change
      --title=...       set process.title
      --input-type=...  module type for --eval and stdin
  -d, --debug           run under gode's line debugger
  -v, --version         print the Node.js compatibility version
  -h, --help            print this help

All recognized flags:
`)
	flagHelpText(w)
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func readSource(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func isPipe(f *os.File) bool {
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice == 0
}
