// Command gode runs JavaScript on JavaScriptCore with Node-compatible core
// modules.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"godejs/internal/diag"
	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/internal/version"

	"godejs/internal/all"
)

func main() { os.Exit(run(os.Args)) }

type options struct {
	eval             string
	hasEval          bool
	print            bool
	script           string
	args             []string
	help             bool
	version          bool
	preserveSymlinks bool
	debug            bool
	execArgv         []string
	preloads         []string
}

// valueFlags are the Node CLI options that consume the following argument.
// Everything else that starts with a dash is a boolean flag we record for
// process.execArgv and otherwise ignore: real programs pass flags like
// --no-warnings, --expose-internals or --experimental-vm-modules, and refusing
// them would stop working code from running.
var valueFlags = map[string]bool{
	"--require": true, "-r": true,
	"--import": true,
	"--loader": true, "--experimental-loader": true,
	"--inspect-port":        true,
	"--title":               true,
	"--stack-trace-limit":   true,
	"--max-old-space-size":  true,
	"--max-semi-space-size": true,
	"--icu-data-dir":        true,
	"--openssl-config":      true,
	"--redirect-warnings":   true,
	"--diagnostic-dir":      true,
	"--conditions":          true, "-C": true,
	"--input-type":          true,
	"--experimental-policy": true,
	"--heapsnapshot-signal": true,
	"--localstorage-file":   true,
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
	opts = applyScriptFlags(opts)
	if opts.help {
		usage(os.Stdout)
		return 0
	}
	if opts.version {
		fmt.Println(version.NodeVersion)
		return 0
	}
	if opts.debug && (opts.hasEval || opts.script == "") {
		fmt.Fprintln(os.Stderr, "gode: --debug requires a script file")
		return 9
	}

	evalMode := opts.hasEval || (opts.print && opts.script != "" && !fileExists(opts.script))
	argvForProcess := []string{argv[0]}
	if !evalMode && opts.script != "" {
		argvForProcess = append(argvForProcess, opts.script)
		argvForProcess = append(argvForProcess, opts.args...)
	}

	e := env.New(env.Options{
		Args:             argvForProcess,
		Env:              os.Environ(),
		Cwd:              mustGetwd(),
		ExecPath:         argv[0],
		Version:          version.NodeVersion,
		ExecArgv:         opts.execArgv,
		PreserveSymlinks: opts.preserveSymlinks,
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
		Stdin:            os.Stdin,
	})
	defer e.JS.Close()
	e.InstallAll()
	all.InstallGlobals(e)
	e.Global().SetFunc("gc", func(this js.Object, args []js.Value) (js.Value, error) {
		runtime.GC()
		return e.JS.Undefined(), nil
	})

	for _, pre := range opts.preloads {
		if _, err := e.RequireFrom(mustGetwd(), pre); err != nil {
			fmt.Fprint(os.Stderr, diag.Render(diag.FromError("gode", err, pre, "")))
			return 1
		}
	}

	var source, file string
	var runErr error

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
		val, err := e.RunScript(source, file)
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
	case isPipe(os.Stdin):
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "gode:", err)
			return 1
		}
		source = string(data)
		file = "[stdin]"
		_, runErr = e.RunScript(source, file)
	default:
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
		return code
	}
	// process.exitCode (and an exit hook that sets it) must decide the status
	// even when the program ended by draining the event loop.
	return e.ExitCode
}

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
	banner := "Welcome to GodeJS v0.1.0!"
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

func parseArgs(args []string) (options, error) {
	var o options
	i := 0
	for ; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			i++
			goto positional
		case a == "-h" || a == "--help":
			o.help = true
		case a == "-v" || a == "--version":
			o.version = true
		case a == "-p" || a == "--print":
			o.print = true
		case a == "--preserve-symlinks":
			o.preserveSymlinks = true
			o.execArgv = append(o.execArgv, a)
		case a == "--debug" || a == "-d":
			o.debug = true
		case a == "-e" || a == "--eval":
			if i+1 >= len(args) {
				return o, fmt.Errorf("bad option '%s': requires an argument", a)
			}
			i++
			o.eval = args[i]
			o.hasEval = true
		case strings.HasPrefix(a, "--eval="):
			o.eval = strings.TrimPrefix(a, "--eval=")
			o.hasEval = true
		case strings.HasPrefix(a, "--print="):
			o.print = true
			o.eval = strings.TrimPrefix(a, "--print=")
			o.hasEval = true
		case strings.HasPrefix(a, "-") && a != "-":
			o.execArgv = append(o.execArgv, a)
			flag := a
			if eq := strings.IndexByte(a, '='); eq >= 0 {
				flag = a[:eq]
			} else if valueFlags[flag] {
				if i+1 >= len(args) {
					return o, fmt.Errorf("bad option '%s': requires an argument", a)
				}
				i++
				o.execArgv = append(o.execArgv, args[i])
				if flag == "--require" || flag == "-r" || flag == "--import" {
					o.preloads = append(o.preloads, args[i])
				}
				continue
			}
			if flag == "--require" || flag == "-r" || flag == "--import" {
				if eq := strings.IndexByte(a, '='); eq >= 0 {
					o.preloads = append(o.preloads, a[eq+1:])
				}
			}
		default:
			goto positional
		}
	}
positional:
	if i < len(args) {
		o.script = args[i]
		o.args = args[i+1:]
	}
	return o, nil
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

Options:
  -e, --eval <source>   evaluate source as JavaScript
  -p, --print           print the result of --eval (or the positional source)
  -d, --debug           run a script under the line debugger
  -v, --version         print the tracked Node.js version
  -h, --help            print this help
`)
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
