package process

import (
	"fmt"
	"math"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("process", Install) }

type listener struct {
	fn   js.Value
	once bool
}

// Install builds the process object, installs it as a global, and wires exit
// handling into the environment.
func Install(e *env.Env) js.Object {
	proc := e.JS.Obj()

	// --- identity -----------------------------------------------------------
	argv := append([]string(nil), e.Argv()...)
	// Node rewrites argv[0] to the executable path before user code runs.
	if len(argv) > 0 && e.Opts.ExecPath != "" {
		argv[0] = e.Opts.ExecPath
	}
	argv0 := ""
	if len(argv) > 0 {
		argv0 = argv[0]
	}
	proc.Set("argv", strArray(e, argv).V())
	proc.SetStr("argv0", argv0)
	runtimeFlags := e.Opts.ExecArgv
	if runtimeFlags == nil {
		runtimeFlags = execArgv(e.Argv())
	}
	proc.Set("execArgv", strArray(e, runtimeFlags).V())
	proc.SetStr("execPath", e.Opts.ExecPath)
	proc.SetStr("title", "node")
	proc.SetStr("version", e.Opts.Version)

	// Node's own version table. Keys, values and descriptors are asserted by
	// test-process-versions, so keep the list exact and lock the values.
	versions := e.JS.Obj()
	versions.SetStr("node", strings.TrimPrefix(e.Opts.Version, "v"))
	versions.SetStr("acorn", "8.15.0")
	versions.SetStr("ada", "2.9.2")
	versions.SetStr("ares", "1.34.5")
	versions.SetStr("brotli", "1.1.0")
	versions.SetStr("cldr", "46.0")
	versions.SetStr("icu", "76.1")
	versions.SetStr("llhttp", "9.3.0")
	versions.SetStr("merve", "1.0.1")
	versions.SetStr("modules", "137")
	versions.SetStr("napi", "10")
	versions.SetStr("nbytes", "0.1.1")
	versions.SetStr("ncrypto", "0.0.0")
	versions.SetStr("nghttp2", "1.64.0")
	versions.SetStr("nghttp3", "1.5.0")
	versions.SetStr("ngtcp2", "1.7.0")
	versions.SetStr("openssl", "3.5.4")
	versions.SetStr("simdjson", "3.13.0")
	versions.SetStr("simdutf", "6.4.0")
	versions.SetStr("tz", "2025b")
	versions.SetStr("undici", "7.16.0")
	versions.SetStr("unicode", "16.0")
	versions.SetStr("uv", "1.51.0")
	versions.SetStr("uvwasi", "0.0.23")
	versions.SetStr("v8", "14.1.146.11-node.14")
	versions.SetStr("zlib", "1.3.1")
	versions.SetStr("zstd", "1.5.7")
	proc.SetObj("versions", versions)

	release := e.JS.Obj()
	release.SetStr("name", "node")
	release.SetStr("sourceUrl", "")
	release.SetStr("headersUrl", "")
	release.SetStr("libUrl", "")
	// Node 24.11+ is the Krypton LTS line (test-process-release).
	if strings.HasPrefix(strings.TrimPrefix(e.Opts.Version, "v"), "24.") {
		release.SetStr("lts", "Krypton")
	}
	proc.SetObj("release", release)

	config := e.JS.Obj()
	configVars := e.JS.Obj()
	configVars.SetStr("napi_build_version", "10")
	configVars.SetNum("node_module_version", 137)
	configVars.SetNum("node_shared_openssl", 0)
	configVars.SetNum("node_use_amaro", 0)
	configVars.SetNum("openssl", 1)
	configVars.SetNum("v8_enable_i18n_support", 1)
	configVars.SetNum("node_shared", 0)
	configVars.SetNum("asan", 0)
	configVars.Set("node_builtin_shareable_builtins", e.JS.Arr(nil).V())
	config.SetObj("variables", configVars)
	targetDefaults := e.JS.Obj()
	targetDefaults.SetStr("default_configuration", "Release")
	config.SetObj("target_defaults", targetDefaults)
	proc.SetObj("config", config)

	features := e.JS.Obj()
	features.SetBool("inspector", false)
	features.SetBool("debug", false)
	features.SetBool("uv", false)
	features.SetBool("ipv6", true)
	features.SetBool("openssl_is_boringssl", false)
	features.SetBool("quic", false)
	features.SetBool("tls_alpn", true)
	features.SetBool("tls_sni", true)
	features.SetBool("tls_ocsp", true)
	features.SetBool("tls", true)
	features.SetBool("cached_builtins", true)
	features.SetBool("require_module", true)
	features.SetStr("typescript", "strip")
	proc.SetObj("features", features)

	proc.SetStr("platform", "linux")
	proc.SetStr("arch", nodeArch())
	proc.SetNum("pid", float64(os.Getpid()))
	proc.SetNum("ppid", float64(os.Getppid()))
	proc.SetBool("_exiting", false)
	// process.type distinguishes Electron's renderer/browser processes; a plain
	// Node process reports 'process'.
	proc.SetStr("type", "process")
	installReport(e, proc)
	// The IPC entry points exist even when no channel was created. Node reports
	// ERR_IPC_CHANNEL_CLOSED / ERR_IPC_DISCONNECTED rather than hiding them.
	proc.SetFunc("send", func(this js.Object, args []js.Value) (js.Value, error) {
		return js.Value{}, js.ThrowCode("Error", "ERR_IPC_CHANNEL_CLOSED", "Channel closed")
	})
	proc.SetFunc("disconnect", func(this js.Object, args []js.Value) (js.Value, error) {
		return js.Value{}, js.ThrowCode("Error", "ERR_IPC_DISCONNECTED", "IPC channel is already disconnected")
	})
	// initgroups applies a user's supplementary groups plus the requested gid,
	// exactly as libuv's uv_set_process_* wrappers do.
	proc.SetFunc("initgroups", func(this js.Object, args []js.Value) (js.Value, error) {
		who := arg(args, 0)
		extra := arg(args, 1)
		if !who.IsString() && !who.IsNumber() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "user" argument must be one of type number or string.`+invalidArgSuffix(who))
		}
		if !extra.IsNumber() && !extra.IsString() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "extraGroup" argument must be one of type number or string.`+invalidArgSuffix(extra))
		}
		// Node resolves the extra group before the user name.
		extraGID := 0
		if extra.IsNumber() {
			extraGID = int(extra.Float())
		} else {
			grp, gerr := user.LookupGroup(extra.Str())
			if gerr != nil {
				return js.Value{}, js.ThrowCode("Error", "ERR_UNKNOWN_CREDENTIAL",
					"Group identifier does not exist: "+extra.Str())
			}
			extraGID, _ = strconv.Atoi(grp.Gid)
		}
		userName := who.Str()
		var u *user.User
		var err error
		if who.IsNumber() {
			userName = strconv.FormatInt(int64(who.Float()), 10)
			u, err = user.LookupId(userName)
		} else {
			u, err = user.Lookup(userName)
		}
		if err != nil {
			return js.Value{}, js.ThrowCode("Error", "ERR_UNKNOWN_CREDENTIAL",
				"User identifier does not exist: "+who.Str())
		}
		gids := []int{extraGID}
		if groups, gerr := u.GroupIds(); gerr == nil {
			for _, g := range groups {
				if n, cerr := strconv.Atoi(g); cerr == nil {
					gids = append(gids, n)
				}
			}
		}
		if serr := syscall.Setgroups(gids); serr != nil {
			return js.Value{}, js.ThrowCode("Error", "EPERM", "EPERM, "+serr.Error())
		}
		return e.JS.Undefined(), nil
	})
	proc.SetFunc("traceProcessWarnings", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 && args[0].Valid() && !args[0].IsUndefined() && !args[0].IsBool() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"The \"value\" argument must be of type boolean")
		}
		return e.JS.Undefined(), nil
	})

	// process.env is a proxy: reads/writes/deletes/own-keys go through the
	// host so child processes inherit mutations, and symbols are rejected the
	// way Node rejects them.
	var envMu sync.Mutex
	var envOrder []string
	envVals := map[string]string{}
	for _, kv := range e.Environ() {
		if i := strings.IndexByte(kv, '='); i > 0 {
			key := kv[:i]
			if _, seen := envVals[key]; !seen {
				envOrder = append(envOrder, key)
			}
			envVals[key] = kv[i+1:]
		}
	}
	syncEnv := func() {
		out := make([]string, 0, len(envOrder))
		for _, key := range envOrder {
			if v, ok := envVals[key]; ok {
				out = append(out, key+"="+v)
			}
		}
		e.Opts.Env = out
	}

	host := e.JS.Obj()
	host.SetFuncHidden("envGet", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Undefined(), nil
		}
		envMu.Lock()
		v, ok := envVals[args[0].Str()]
		envMu.Unlock()
		if !ok {
			return e.JS.Undefined(), nil
		}
		return e.JS.Str(v), nil
	})
	host.SetFuncHidden("envSet", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 {
			return e.JS.Undefined(), nil
		}
		key := args[0].Str()
		if key == "" {
			return e.JS.Undefined(), nil
		}
		var value string
		switch {
		case args[1].IsUndefined():
			value = "undefined"
		case args[1].IsNull():
			value = "null"
		default:
			value = args[1].Str()
		}
		envMu.Lock()
		if _, seen := envVals[key]; !seen {
			envOrder = append(envOrder, key)
		}
		envVals[key] = value
		syncEnv()
		envMu.Unlock()
		return e.JS.Undefined(), nil
	})
	host.SetFuncHidden("envDelete", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Undefined(), nil
		}
		key := args[0].Str()
		envMu.Lock()
		if _, ok := envVals[key]; ok {
			delete(envVals, key)
			out := envOrder[:0]
			for _, k := range envOrder {
				if k != key {
					out = append(out, k)
				}
			}
			envOrder = out
			syncEnv()
		}
		envMu.Unlock()
		return e.JS.Undefined(), nil
	})
	host.SetFuncHidden("envHas", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Bool(false), nil
		}
		envMu.Lock()
		_, ok := envVals[args[0].Str()]
		envMu.Unlock()
		return e.JS.Bool(ok), nil
	})
	host.SetFuncHidden("envKeys", func(this js.Object, args []js.Value) (js.Value, error) {
		envMu.Lock()
		keys := append([]string(nil), envOrder...)
		envMu.Unlock()
		vals := make([]js.Value, len(keys))
		for i, k := range keys {
			vals[i] = e.JS.Str(k)
		}
		return e.JS.Arr(vals).V(), nil
	})
	host.SetFuncHidden("getExitCode", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(float64(e.ExitCode)), nil
	})
	host.SetFuncHidden("setExitCode", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 && args[0].IsNumber() {
			e.ExitCode = args[0].Int()
		}
		return e.JS.Undefined(), nil
	})
	host.SetFuncHidden("cwd", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Str(e.Opts.Cwd), nil
	})

	// --- event surface ------------------------------------------------------
	var mu sync.Mutex
	listeners := map[string][]*listener{}

	emitErr := func(event string, args []js.Value) error {
		mu.Lock()
		entries := append([]*listener(nil), listeners[event]...)
		var keep []*listener
		for _, l := range entries {
			if l.once {
				continue
			}
			keep = append(keep, l)
		}
		listeners[event] = keep
		mu.Unlock()
		for _, l := range entries {
			if _, err := l.fn.Object().Call(proc, args...); err != nil {
				return err
			}
		}
		return nil
	}
	emit := func(event string, args []js.Value) { _ = emitErr(event, args) }

	addListener := func(event string, fn js.Value, once bool) {
		fn.Protect()
		mu.Lock()
		listeners[event] = append(listeners[event], &listener{fn: fn, once: once})
		mu.Unlock()
	}

	proc.SetFunc("on", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 || !args[1].IsFunction() {
			return js.Value{}, fmt.Errorf("The \"listener\" argument must be of type function")
		}
		addListener(args[0].Str(), args[1], false)
		return proc.V(), nil
	})
	proc.SetFunc("addListener", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 || !args[1].IsFunction() {
			return js.Value{}, fmt.Errorf("The \"listener\" argument must be of type function")
		}
		addListener(args[0].Str(), args[1], false)
		return proc.V(), nil
	})
	proc.SetFunc("once", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 || !args[1].IsFunction() {
			return js.Value{}, fmt.Errorf("The \"listener\" argument must be of type function")
		}
		addListener(args[0].Str(), args[1], true)
		return proc.V(), nil
	})
	remove := func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 {
			return proc.V(), nil
		}
		event := args[0].Str()
		mu.Lock()
		entries := listeners[event]
		out := entries[:0]
		for _, l := range entries {
			if sameValue(e, l.fn, args[1]) {
				l.fn.Unprotect()
				continue
			}
			out = append(out, l)
		}
		listeners[event] = out
		mu.Unlock()
		return proc.V(), nil
	}
	proc.SetFunc("off", remove)
	proc.SetFunc("removeListener", remove)
	proc.SetFunc("listenerCount", func(this js.Object, args []js.Value) (js.Value, error) {
		event := ""
		if len(args) > 0 {
			event = args[0].Str()
		}
		mu.Lock()
		n := len(listeners[event])
		mu.Unlock()
		return e.JS.Num(float64(n)), nil
	})
	listOf := func(event string) js.Value {
		mu.Lock()
		entries := append([]*listener(nil), listeners[event]...)
		mu.Unlock()
		vals := make([]js.Value, len(entries))
		for i, l := range entries {
			vals[i] = l.fn
		}
		return e.JS.Arr(vals).V()
	}
	proc.SetFunc("listeners", func(this js.Object, args []js.Value) (js.Value, error) {
		event := ""
		if len(args) > 0 {
			event = args[0].Str()
		}
		return listOf(event), nil
	})
	proc.SetFunc("rawListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		event := ""
		if len(args) > 0 {
			event = args[0].Str()
		}
		return listOf(event), nil
	})
	proc.SetFunc("eventNames", func(this js.Object, args []js.Value) (js.Value, error) {
		mu.Lock()
		names := make([]string, 0, len(listeners))
		for name, entries := range listeners {
			if len(entries) > 0 {
				names = append(names, name)
			}
		}
		mu.Unlock()
		vals := make([]js.Value, len(names))
		for i, name := range names {
			vals[i] = e.JS.Str(name)
		}
		return e.JS.Arr(vals).V(), nil
	})
	prepend := func(this js.Object, args []js.Value, once bool) (js.Value, error) {
		if len(args) < 2 || !args[1].IsFunction() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"listener\" argument must be of type function")
		}
		event := args[0].Str()
		args[1].Protect()
		mu.Lock()
		listeners[event] = append([]*listener{{fn: args[1], once: once}}, listeners[event]...)
		mu.Unlock()
		return proc.V(), nil
	}
	proc.SetFunc("prependListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return prepend(this, args, false)
	})
	proc.SetFunc("prependOnceListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return prepend(this, args, true)
	})
	proc.SetFunc("removeAllListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		mu.Lock()
		if len(args) == 0 || args[0].IsUndefined() {
			for name, entries := range listeners {
				for _, l := range entries {
					l.fn.Unprotect()
				}
				delete(listeners, name)
			}
		} else {
			event := args[0].Str()
			for _, l := range listeners[event] {
				l.fn.Unprotect()
			}
			delete(listeners, event)
		}
		mu.Unlock()
		return proc.V(), nil
	})
	proc.SetFunc("emit", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Bool(false), nil
		}
		event := args[0].Str()
		mu.Lock()
		n := len(listeners[event])
		mu.Unlock()
		emit(event, args[1:])
		return e.JS.Bool(n > 0), nil
	})

	// The capture callback the runtime delivers uncaught exceptions to. It is
	// declared here because the delivery hook below closes over it.
	var captureFn js.Value
	countOf := func(event string) int {
		mu.Lock()
		n := len(listeners[event])
		mu.Unlock()
		return n
	}

	// beforeExit fires each time the loop empties and the handler schedules
	// more work; process.exit() skips it. 'exit' fires once, synchronously.
	e.SetBeforeExitHook(func(code int) bool {
		emit("beforeExit", []js.Value{e.JS.Num(float64(code))})
		return true
	})
	e.OnExit(func(code int) {
		proc.SetBool("_exiting", true)
		if !e.Exited {
			if v := proc.Get("exitCode"); v.IsNumber() {
				e.ExitCode = v.Int()
			}
		}
		emit("exit", []js.Value{e.JS.Num(float64(e.ExitCode))})
	})

	// Uncaught exceptions reach the capture callback first, then the
	// 'uncaughtException' listeners; an unhandled rejection is an unhandled
	// promise rejection with no 'unhandledRejection' listener of its own.
	e.SetUncaughtHandler(func(err js.Value, origin string) bool {
		if captureFn.Valid() && captureFn.IsFunction() {
			if _, callErr := captureFn.Object().Call(proc, err); callErr != nil {
				e.ExitCode = 7
				return false
			}
			return true
		}
		if origin == "unhandledRejection" {
			if countOf(origin) == 0 {
				return false
			}
			return emitErr(origin, []js.Value{err, e.JS.Undefined()}) == nil
		}
		if countOf("uncaughtException") == 0 {
			return false
		}
		// A listener that throws leaves the new exception unhandled; Node then
		// terminates with exit code 7.
		if emitErr("uncaughtException", []js.Value{err, e.JS.Str("uncaughtException")}) != nil {
			e.ExitCode = 7
			return false
		}
		return true
	})

	// --- stdio --------------------------------------------------------------
	writeStream := func(fd int, out func() jsValueWriter) js.Object {
		stream := e.JS.Obj()
		stream.SetNum("fd", float64(fd))
		stream.SetBool("isTTY", false)
		stream.SetFunc("write", func(this js.Object, args []js.Value) (js.Value, error) {
			if len(args) == 0 {
				return js.Value{}, fmt.Errorf("The \"chunk\" argument must be of type string or an instance of Buffer or Uint8Array")
			}
			var cb js.Value
			for i := 1; i < len(args); i++ {
				if args[i].IsFunction() {
					cb = args[i]
				}
			}
			w := out()
			if b, ok := args[0].Bytes(); ok {
				if _, err := w.Write(b); err != nil {
					return js.Value{}, err
				}
			} else if args[0].IsString() {
				if _, err := w.Write([]byte(args[0].Str())); err != nil {
					return js.Value{}, err
				}
			} else {
				return js.Value{}, fmt.Errorf("The \"chunk\" argument must be of type string or an instance of Buffer or Uint8Array")
			}
			if cb.Valid() {
				e.NextTick(cb)
			}
			return e.JS.Bool(true), nil
		})
		stream.SetFunc("end", func(this js.Object, args []js.Value) (js.Value, error) {
			var cb js.Value
			for _, a := range args {
				if a.IsFunction() {
					cb = a
				}
			}
			if cb.Valid() {
				e.NextTick(cb)
			}
			return stream.V(), nil
		})
		return stream
	}
	proc.SetObj("stdout", writeStream(1, func() jsValueWriter { return e.Out() }))
	proc.SetObj("stderr", writeStream(2, func() jsValueWriter { return e.Err() }))

	// process.stdin is a real Readable fed by polled reads on fd 0. The
	// getter is lazy so installing process does not force stream in isolated
	// tests.
	installStdin(e, proc)

	// --- methods ------------------------------------------------------------
	// process.cwd() asks the kernel instead of trusting the cached string, so
	// a directory removed after chdir() surfaces as ENOENT with Node's uv_cwd
	// message (test-cwd-enoent-improved-message).
	proc.SetFunc("cwd", func(this js.Object, args []js.Value) (js.Value, error) {
		wd, err := os.Getwd()
		if err != nil {
			return js.Value{}, js.ThrowCode("Error", "ENOENT",
				"ENOENT: process.cwd failed with error no such file or directory, "+
					"the current working directory was likely removed without changing "+
					"the working directory, uv_cwd")
		}
		e.Opts.Cwd = wd
		return e.JS.Str(wd), nil
	})
	host.SetFuncHidden("chdir", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"directory\" argument must be of type string")
		}
		from := e.Opts.Cwd
		dir := args[0].Str()
		if err := os.Chdir(dir); err != nil {
			code, text := "ENOENT", "no such file or directory"
			if errno, ok := err.(syscall.Errno); ok {
				switch errno {
				case syscall.ENOTDIR:
					code, text = "ENOTDIR", "not a directory"
				case syscall.EACCES:
					code, text = "EACCES", "permission denied"
				case syscall.ENAMETOOLONG:
					code, text = "ENAMETOOLONG", "name too long"
				}
			}
			return js.Value{}, js.ThrowCode("Error", code,
				fmt.Sprintf("%s: %s, chdir %s -> '%s'", code, text, from, dir))
		}
		abs, err := os.Getwd()
		if err == nil {
			e.Opts.Cwd = abs
		}
		return e.JS.Undefined(), nil
	})
	host.SetFuncHidden("exit", func(this js.Object, args []js.Value) (js.Value, error) {
		proc.SetBool("_exiting", true)
		code := e.ExitCode
		if len(args) > 0 && args[0].IsNumber() {
			code = args[0].Int()
		}
		e.Exit(code)
		return js.Value{}, env.ErrProcessExit
	})
	proc.SetFunc("reallyExit", func(this js.Object, args []js.Value) (js.Value, error) {
		proc.SetBool("_exiting", true)
		code := e.ExitCode
		if len(args) > 0 && args[0].IsNumber() {
			code = args[0].Int()
		}
		e.Exit(code)
		return js.Value{}, env.ErrProcessExit
	})
	proc.SetFunc("abort", func(this js.Object, args []js.Value) (js.Value, error) {
		// Node aborts the process with SIGABRT, so a parent sees the signal.
		_ = syscall.Kill(os.Getpid(), syscall.SIGABRT)
		return js.Value{}, fmt.Errorf("The process was aborted")
	})
	// process.dlopen(module, filename) is the CommonJS native-addon entry
	// point. The Node-API loader owns dlopen and the exported C ABI; this
	// method just binds the loaded exports to the caller's module object.
	proc.SetFunc("dlopen", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 || !args[0].IsObject() || args[0].IsNull() || !args[1].IsString() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"process.dlopen requires a module object and a filename")
		}
		filename := args[1].Str()
		exp, loadErr := e.LoadNative(filename)
		if loadErr != nil {
			return js.Value{}, js.ThrowCode("Error", "ERR_DLOPEN_FAILED",
				"ERR_DLOPEN_FAILED: "+loadErr.Error())
		}
		module := args[0].Object()
		module.Set("exports", exp.V())
		return exp.V(), nil
	})
	proc.SetFunc("uptime", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(time.Since(e.StartTime).Seconds()), nil
	})
	host.SetFuncHidden("hrtime", func(this js.Object, args []js.Value) (js.Value, error) {
		elapsed := time.Since(e.StartTime).Nanoseconds()
		if len(args) >= 2 && args[0].IsNumber() && args[1].IsNumber() {
			elapsed -= int64(args[0].Float())*1e9 + int64(args[1].Float())
		}
		sec := elapsed / 1e9
		nsec := elapsed % 1e9
		if nsec < 0 {
			sec--
			nsec += 1e9
		}
		return e.JS.Arr([]js.Value{
			e.JS.Num(float64(sec)),
			e.JS.Num(float64(nsec)),
		}).V(), nil
	})
	host.SetFuncHidden("hrtimeBigint", func(this js.Object, args []js.Value) (js.Value, error) {
		elapsed := time.Since(e.StartTime).Nanoseconds()
		if len(args) > 0 {
			if args[0].IsNumber() {
				elapsed -= int64(args[0].Float())
			} else {
				prev, err := e.JS.Eval("(function (v) { return Number(v); })", "[gode:process]")
				if err == nil {
					if n, err := prev.Object().Call(js.Object{}, args[0]); err == nil {
						elapsed -= int64(n.Float())
					}
				}
			}
		}
		bigint, err := e.JS.Eval("(function (s) { return BigInt(s); })", "[gode:process]")
		if err != nil {
			return js.Value{}, err
		}
		return bigint.Object().Call(js.Object{}, e.JS.Str(strconv.FormatInt(elapsed, 10)))
	})
	host.SetFuncHidden("umask", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || args[0].IsUndefined() {
			old := syscall.Umask(0)
			syscall.Umask(old)
			return e.JS.Num(float64(old)), nil
		}
		// libuv only applies the permission bits, so the high bits of a mask
		// like 0o10664 are ignored both when set and when read back
		// (test-process-umask-mask).
		mask := args[0].Int() & 0o777
		return e.JS.Num(float64(syscall.Umask(mask))), nil
	})
	host.SetFuncHidden("kill", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"signal\" argument must be of type number or string")
		}
		pid := args[0].Int()
		sig, err := parseSignal(args[1])
		if err != nil {
			return js.Value{}, err
		}
		if err := syscall.Kill(pid, sig); err != nil {
			return js.Value{}, js.ThrowCode("Error", errCode(err), fmt.Sprintf("kill %s", errCode(err)))
		}
		return e.JS.Bool(true), nil
	})

	proc.SetFunc("memoryUsage", func(this js.Object, args []js.Value) (js.Value, error) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		rss := memRSS()
		out := e.JS.Obj()
		out.SetNum("rss", float64(rss))
		out.SetNum("heapTotal", float64(m.HeapSys))
		out.SetNum("heapUsed", float64(m.HeapAlloc))
		out.SetNum("external", float64(m.MSpanSys+m.MCacheSys))
		out.SetNum("arrayBuffers", float64(m.HeapInuse))
		return out.V(), nil
	})
	// Node exposes process.memoryUsage.rss() as well as process.rss().
	if memUsage := proc.Get("memoryUsage"); memUsage.IsFunction() {
		memUsage.Object().SetFuncHidden("rss", func(this js.Object, args []js.Value) (js.Value, error) {
			return e.JS.Num(float64(memRSS())), nil
		})
	}
	proc.SetFunc("cpuUsage", func(this js.Object, args []js.Value) (js.Value, error) {
		user, system := cpuUsage()
		if len(args) > 0 && args[0].IsObject() && !args[0].IsArray() {
			prev := args[0].Object()
			pu, ps, err := validatePrevUsage(prev)
			if err != nil {
				return js.Value{}, err
			}
			user -= pu
			system -= ps
		} else if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "prevValue" argument must be of type object.`+invalidArgSuffix(args[0]))
		}
		out := e.JS.Obj()
		out.SetNum("user", float64(user))
		out.SetNum("system", float64(system))
		return out.V(), nil
	})
	proc.SetFunc("nextTick", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsFunction() {
			received := "undefined"
			if len(args) > 0 {
				received = "type " + jsTypeName(args[0])
			}
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"The \"callback\" argument must be of type function. Received "+received)
		}
		e.NextTick(args[0], args[1:]...)
		return e.JS.Undefined(), nil
	})
	proc.SetFunc("getuid", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(float64(syscall.Getuid())), nil
	})
	proc.SetFunc("geteuid", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(float64(syscall.Geteuid())), nil
	})
	proc.SetFunc("getgid", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(float64(syscall.Getgid())), nil
	})
	proc.SetFunc("getegid", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(float64(syscall.Getegid())), nil
	})
	proc.SetFunc("setuid", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), setID("setuid", args, syscall.Setuid)
	})
	proc.SetFunc("seteuid", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), setID("seteuid", args, syscall.Seteuid)
	})
	proc.SetFunc("setgid", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), setID("setgid", args, syscall.Setgid)
	})
	proc.SetFunc("setegid", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), setID("setegid", args, syscall.Setegid)
	})
	// _rawDebug writes straight to fd 2, bypassing process.stderr and the
	// microtask queue, which is what makes it usable while everything else is
	// broken (test-process-raw-debug).
	rawDebug := func(this js.Object, args []js.Value) (js.Value, error) {
		line := formatArgs(e, args)
		fmt.Fprintln(e.Err(), line)
		return e.JS.Undefined(), nil
	}
	proc.SetFunc("_rawDebug", rawDebug)
	proc.SetFunc("rawDebug", rawDebug)
	proc.SetFunc("execve", func(this js.Object, args []js.Value) (js.Value, error) {
		fileArg := arg(args, 0)
		if !fileArg.IsString() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "execPath" argument must be of type string.`+invalidArgSuffix(fileArg))
		}
		argsArg := arg(args, 1)
		if !argsArg.IsArray() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "args" argument must be an instance of Array.`+invalidArgSuffix(argsArg))
		}
		file := fileArg.Str()
		arr := argsArg.Object()
		aargv := make([]string, arr.Length())
		for i := range aargv {
			v := arr.At(i)
			if !v.IsString() || strings.IndexByte(v.Str(), 0) >= 0 {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
					"The argument 'args["+strconv.Itoa(i)+"]' must be a string without null bytes. Received "+
						nodeInspectErrorValue(v))
			}
			aargv[i] = v.Str()
		}
		envList := append([]string(nil), e.Environ()...)
		envArg := arg(args, 2)
		if envArg.IsObject() && !envArg.IsNull() {
			obj := envArg.Object()
			keys := envKeys(e, obj)
			envList = envList[:0]
			for _, k := range keys {
				v := obj.Get(k)
				if !v.IsString() || strings.IndexByte(v.Str(), 0) >= 0 {
					return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
						"The argument 'env' must be an object with string keys and values without null bytes. Received "+
							inspectEnvObject(e, obj))
				}
				envList = append(envList, k+"="+v.Str())
			}
		} else if envArg.Valid() && !envArg.IsUndefined() && !envArg.IsNull() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "env" argument must be of type object.`+invalidArgSuffix(envArg))
		}
		if err := syscall.Exec(file, aargv, envList); err != nil {
			return js.Value{}, js.ThrowCode("Error", "ENOENT",
				fmt.Sprintf("ENOENT: no such file or directory, execve '%s'", file))
		}
		return e.JS.Undefined(), nil
	})

	proc.SetFunc("getgroups", func(this js.Object, args []js.Value) (js.Value, error) {
		gids, err := syscall.Getgroups()
		if err != nil {
			return js.Value{}, err
		}
		vals := make([]js.Value, len(gids))
		for i, g := range gids {
			vals[i] = e.JS.Num(float64(g))
		}
		return e.JS.Arr(vals).V(), nil
	})
	proc.SetFunc("setgroups", func(this js.Object, args []js.Value) (js.Value, error) {
		groups := arg(args, 0)
		if !groups.IsArray() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "groups" argument must be an instance of Array.`+invalidArgSuffix(groups))
		}
		arr := groups.Object()
		gids := make([]int, arr.Length())
		for i := range gids {
			item := arr.At(i)
			switch {
			case item.IsNumber():
				f := item.Float()
				if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 || f > 4294967295 || math.Floor(f) != f {
					return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
						`The value of "groups[`+strconv.Itoa(i)+`]" is out of range. It must be >= 0 && <= 4294967295. Received `+formatNumber(f))
				}
				gids[i] = int(f)
			case item.IsString():
				g, gerr := user.LookupGroup(item.Str())
				if gerr != nil {
					return js.Value{}, js.ThrowCode("Error", "ERR_UNKNOWN_CREDENTIAL",
						"Group identifier does not exist: "+item.Str())
				}
				n, _ := strconv.Atoi(g.Gid)
				gids[i] = n
			default:
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "groups[`+strconv.Itoa(i)+`]" argument must be one of type number or string.`+invalidArgSuffix(item))
			}
		}
		if err := syscall.Setgroups(gids); err != nil {
			return js.Value{}, js.ThrowCode("Error", errCode(err), "setgroups "+errCode(err))
		}
		return e.JS.Undefined(), nil
	})
	proc.SetFunc("threadCpuUsage", func(this js.Object, args []js.Value) (js.Value, error) {
		user, system := cpuUsage()
		if len(args) > 0 && args[0].IsObject() && !args[0].IsArray() {
			prev := args[0].Object()
			pu, ps, err := validatePrevUsage(prev)
			if err != nil {
				return js.Value{}, err
			}
			user -= pu
			system -= ps
		} else if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "prevValue" argument must be of type object.`+invalidArgSuffix(args[0]))
		}
		out := e.JS.Obj()
		out.SetNum("user", float64(user))
		out.SetNum("system", float64(system))
		return out.V(), nil
	})

	proc.SetFunc("setSourceMapsEnabled", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 && !args[0].IsBool() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"val\" argument must be of type boolean")
		}
		return e.JS.Undefined(), nil
	})
	proc.SetFunc("hasUncaughtExceptionCaptureCallback", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Bool(captureFn.Valid() && captureFn.IsFunction()), nil
	})
	proc.SetFunc("setUncaughtExceptionCaptureCallback", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || args[0].IsNull() {
			if captureFn.Valid() {
				captureFn.Unprotect()
			}
			captureFn = js.Value{}
			return e.JS.Undefined(), nil
		}
		if !args[0].IsFunction() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "fn" argument must be of type function or null.`+invalidArgSuffix(args[0]))
		}
		if captureFn.Valid() && captureFn.IsFunction() {
			return js.Value{}, js.ThrowCode("Error", "ERR_UNCAUGHT_EXCEPTION_CAPTURE_ALREADY_SET",
				"setupUncaughtExceptionCaptureCallback() was called while a capture callback was already set")
		}
		args[0].Protect()
		captureFn = args[0]
		return e.JS.Undefined(), nil
	})
	// loadEnvFile reports read failures as data so the prelude can raise a real
	// Error carrying code/path/syscall (test-process-load-env-file).
	host.SetFuncHidden("loadEnvFile", func(this js.Object, args []js.Value) (js.Value, error) {
		path := ".env"
		if len(args) > 0 && args[0].IsString() {
			path = args[0].Str()
		}
		data, err := os.ReadFile(path)
		if err != nil {
			info := e.JS.Obj()
			info.SetBool("error", true)
			info.SetStr("code", "ENOENT")
			info.SetStr("path", path)
			info.SetStr("syscall", "open")
			info.SetStr("message", "ENOENT: no such file or directory, open '"+path+"'")
			return info.V(), nil
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			value = strings.Trim(value, "\"'")
			if key == "" {
				continue
			}
			envMu.Lock()
			if _, seen := envVals[key]; !seen {
				envOrder = append(envOrder, key)
			}
			envVals[key] = value
			syncEnv()
			envMu.Unlock()
		}
		return e.JS.Undefined(), nil
	})
	proc.SetFunc("rss", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(float64(memRSS())), nil
	})
	proc.SetFunc("register", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})
	proc.SetFunc("unregister", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})
	proc.SetFunc("registerBeforeExit", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})
	proc.SetFunc("getBuiltinModule", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return js.Value{}, fmt.Errorf("The \"id\" argument must be of type string")
		}
		name := strings.TrimPrefix(args[0].Str(), "node:")
		if !env.Registered(name) {
			return js.Value{}, js.ThrowCode("Error", "ERR_UNKNOWN_BUILTIN_MODULE", "No such built-in module: "+name)
		}
		return e.Exports(name).V(), nil
	})
	// process.emitWarning follows lib/internal/process/warning.js:
	//   emitWarning(error)
	//   emitWarning(str[, type[, code]][, ctor])
	//   emitWarning(str[, options])
	// Warnings are delivered to 'warning' listeners and echoed to stderr with
	// Node's "(node:PID) [CODE] Name: message" shape.
	warningHelperShown := false
	var writeWarning func(js.Value)
	emitWarning := func(args []js.Value, sync bool) (js.Value, error) {
		invalid := func(name, want string, got js.Value) error {
			return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				fmt.Sprintf("The \"%s\" argument must be %s. Received %s", name, want, jsTypeName(got)))
		}
		warning := arg(args, 0)
		typ := arg(args, 1)
		code := arg(args, 2)
		ctor := arg(args, 3)
		typeName := "Warning"
		var detail js.Value
		var ctorFn js.Value

		switch {
		case isPlainOptions(typ):
			opts := typ.Object()
			if v := opts.Get("ctor"); v.IsFunction() {
				ctorFn = v
			}
			if v := opts.Get("code"); v.Valid() && !v.IsUndefined() {
				code = v
			}
			if v := opts.Get("detail"); v.IsString() {
				detail = v
			}
			if v := opts.Get("type"); v.IsString() && v.Str() != "" {
				typeName = v.Str()
			}
		case typ.Valid() && typ.IsFunction():
			ctorFn = typ
			code = js.Value{}
		case typ.Valid() && !typ.IsUndefined():
			if !typ.IsString() {
				return js.Value{}, invalid("type", "of type string", typ)
			}
			typeName = typ.Str()
		}
		if code.Valid() && !code.IsUndefined() {
			if code.IsFunction() {
				ctorFn = code
				code = js.Value{}
			} else if !code.IsString() {
				return js.Value{}, invalid("code", "of type string", code)
			}
		}
		_ = ctor
		_ = ctorFn

		var warn js.Object
		if warning.IsString() {
			msg := warning.Str()
			if ctorFn.Valid() && ctorFn.IsFunction() {
				// The `ctor` option lets callers shape the warning class; Node
				// constructs it with the message and keeps whatever name/code
				// the class sets.
				if construct := e.Global().Get("Reflect").Object().Get("construct"); construct.IsFunction() {
					if made, err := construct.Object().Call(js.Object{}, ctorFn, e.JS.Arr([]js.Value{e.JS.Str(msg)}).V()); err == nil && made.IsObject() {
						warn = made.Object()
					}
				}
			}
			if !warn.Valid() {
				warn = e.JS.Error(msg)
				warn.SetStr("name", typeName)
				if code.Valid() && code.IsString() {
					warn.SetStr("code", code.Str())
				}
			}
			if detail.Valid() && detail.IsString() {
				warn.SetStr("detail", detail.Str())
			}
			warning = warn.V()
		} else if !errorLikeValue(warning) {
			return js.Value{}, invalid("warning", "of type string or an instance of Error", warning)
		}

		// process.noDeprecation silences deprecation warnings entirely. Read it
		// off the object being built: e.Exports("process") is a placeholder while
		// this install is still running.
		if name := proc.Get("noDeprecation"); name.Valid() && name.IsBool() && name.Bool() {
			if warning.IsObject() {
				nm := warning.Object().Get("name")
				if nm.IsString() && nm.Str() == "DeprecationWarning" {
					return e.JS.Undefined(), nil
				}
			}
		}

		// Warnings are delivered on the next tick, so a 'warning' listener
		// registered synchronously after emitWarning still receives it
		// (lib/internal/process/warning.js schedules doEmitWarning).
		deliver := func() error {
			// --throw-deprecation (or process.throwDeprecation = true) throws the
			// warning instead of emitting or printing it.
			if warningName(warning) == "DeprecationWarning" {
				if td := proc.Get("throwDeprecation"); td.Valid() && td.IsBool() && td.Bool() {
					thrower, err := e.JS.Eval("(function (w) { throw w; })", "[gode:process]")
					if err != nil {
						return err
					}
					_, err = thrower.Object().Call(js.Object{}, warning)
					return err
				}
			}
			emit("warning", []js.Value{warning})
			writeWarning(warning)
			return nil
		}
		if sync {
			if err := deliver(); err != nil {
				return js.Value{}, err
			}
			return e.JS.Undefined(), nil
		}
		deferEmit(e, func() error { return deliver() })
		return e.JS.Undefined(), nil
	}
	// writeWarning echoes a warning to stderr with Node's shape.
	writeWarning = func(warning js.Value) {
		suppressed := false
		for _, a := range e.Argv() {
			if a == "--no-warnings" || strings.HasPrefix(a, "--no-warnings=") || a == "--no-deprecation" {
				suppressed = true
			}
		}
		if !suppressed && errorLikeValue(warning) {
			obj := warning.Object()
			name := "Warning"
			if v := obj.Get("name"); v.IsString() {
				name = v.Str()
			}
			msg := ""
			if v := obj.Get("message"); v.IsString() {
				msg = v.Str()
			}
			codeStr := ""
			if v := obj.Get("code"); v.IsString() {
				codeStr = v.Str()
			}
			prefix := fmt.Sprintf("(node:%d) ", os.Getpid())
			if codeStr != "" {
				prefix += "[" + codeStr + "] "
			}
			line := prefix + name + ": " + msg
			if v := obj.Get("detail"); v.IsString() {
				line += "\n" + v.Str()
			}
			if !warningHelperShown {
				flag := "--trace-warnings"
				if name == "DeprecationWarning" {
					flag = "--trace-deprecation"
				}
				line += "\n(Use `node " + flag + " ...` to show where the warning was created)"
				warningHelperShown = true
			}
			fmt.Fprintln(e.Err(), line)
		}
	}
	proc.SetFunc("emitWarning", func(this js.Object, args []js.Value) (js.Value, error) {
		return emitWarning(args, false)
	})
	proc.SetFunc("emitWarningSync", func(this js.Object, args []js.Value) (js.Value, error) {
		return emitWarning(args, true)
	})
	proc.SetFunc("availableMemory", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(float64(availableMemory())), nil
	})
	proc.SetFunc("constrainedMemory", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(0), nil
	})
	proc.SetFunc("resourceUsage", func(this js.Object, args []js.Value) (js.Value, error) {
		var ru syscall.Rusage
		if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
			return js.Value{}, err
		}
		out := e.JS.Obj()
		out.SetNum("userCPUTime", float64(ru.Utime.Sec*1e6+ru.Utime.Usec))
		out.SetNum("systemCPUTime", float64(ru.Stime.Sec*1e6+ru.Stime.Usec))
		out.SetNum("maxRSS", float64(ru.Maxrss))
		out.SetNum("sharedMemorySize", float64(ru.Ixrss))
		out.SetNum("unsharedDataSize", float64(ru.Idrss))
		out.SetNum("unsharedStackSize", float64(ru.Isrss))
		out.SetNum("minorPageFault", float64(ru.Minflt))
		out.SetNum("majorPageFault", float64(ru.Majflt))
		out.SetNum("swappedOut", float64(ru.Nswap))
		out.SetNum("fsRead", float64(ru.Inblock))
		out.SetNum("fsWrite", float64(ru.Oublock))
		out.SetNum("ipcSent", float64(ru.Msgsnd))
		out.SetNum("ipcReceived", float64(ru.Msgrcv))
		out.SetNum("signalsCount", float64(ru.Nsignals))
		out.SetNum("voluntaryContextSwitches", float64(ru.Nvcsw))
		out.SetNum("involuntaryContextSwitches", float64(ru.Nivcsw))
		return out.V(), nil
	})

	// process.binding is Node's legacy internal-binding table. Only the two
	// allow-listed modules are exposed; anything else reports "No such module".
	host.SetFuncHidden("binding", func(this js.Object, args []js.Value) (js.Value, error) {
		name := arg(args, 0)
		switch name.Str() {
		case "buffer":
			return e.JS.Obj().V(), nil
		case "util":
			out := e.JS.Obj()
			if types := e.Exports("util").Get("types"); types.IsObject() {
				for _, key := range []string{
					"isAnyArrayBuffer", "isArrayBuffer", "isArrayBufferView", "isAsyncFunction",
					"isDataView", "isDate", "isExternal", "isMap", "isMapIterator",
					"isNativeError", "isPromise", "isRegExp", "isSet", "isSetIterator",
					"isTypedArray", "isUint8Array",
				} {
					out.Set(key, types.Object().Get(key))
				}
			}
			return out.V(), nil
		}
		return js.Value{}, js.ThrowCode("Error", "ERR_UNKNOWN_BUILTIN_MODULE",
			"No such module: "+name.Str())
	})

	// Install timers before the prelude so its active-resource tracking wraps
	// the real setTimeout/setImmediate rather than a bootstrap stub.
	if env.Registered("timers") {
		e.Exports("timers")
	}

	// The parts of process that need JS-level validation and proxies are
	// installed together so the host functions above stay private.
	// process inherits from EventEmitter; load it so the prototype chain
	// matches Node (test-process-prototype). events does not depend on process.
	ee := e.JS.Undefined()
	if env.Registered("events") {
		ee = e.Exports("events").Get("EventEmitter")
	}
	if prelude, err := e.JS.Eval(`(function (proc, host, EventEmitter) {
  'use strict';
  function describe(v) {
    if (v === null) return 'null';
    var t = typeof v;
    if (t === 'string') return "type string ('" + v + "')";
    if (t === 'boolean') return 'type boolean (' + v + ')';
    if (t === 'bigint') return 'type bigint (' + String(v) + 'n)';
    if (t === 'undefined') return 'undefined';
    if (t === 'number') return 'type number (' + String(v) + ')';
    if (t === 'symbol') return 'type symbol';
    if (t === 'function') return 'type function';
    if (Array.isArray(v)) return 'an instance of Array';
    return 'an instance of Object';
  }
  function invalidArg(name, expected, v) {
    var e = new TypeError('The "' + name + '" argument must be ' + expected +
      '. Received ' + describe(v));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function rangeErr(msg) {
    var e = new RangeError(msg);
    e.code = 'ERR_OUT_OF_RANGE';
    return e;
  }
  function validateExitCode(code) {
    if (code === undefined || code === null) return 0;
    var t = typeof code;
    if (t === 'number') {
      if (code !== code) {
        throw rangeErr('The value of "code" is out of range. It must be an integer. Received NaN');
      }
      if (!isFinite(code) || Math.floor(code) !== code) {
        throw rangeErr('The value of "code" is out of range. It must be an integer. Received ' + String(code));
      }
      return code;
    }
    if (t === 'string') {
      if (/^[0-9]+$/.test(code)) return Number(code);
      throw invalidArg('code', 'of type number or a string that represents an integer', code);
    }
    throw invalidArg('code', 'of type number', code);
  }

  // The target inherits Object.prototype so process.env.hasOwnProperty is
  // the real method until an env key shadows it.
  var envTarget = {};
  var envProxy = new Proxy(envTarget, {
    get: function (t, k) {
      if (typeof k === 'symbol') return undefined;
      if (host.envHas(k)) return host.envGet(k);
      return t[k];
    },
    set: function (t, k, v) {
      if (typeof k === 'symbol') {
        throw new TypeError('Cannot convert a Symbol value to a string');
      }
      if (typeof v === 'symbol') {
        throw new TypeError('Cannot convert a Symbol value to a string');
      }
      host.envSet(k, v);
      return true;
    },
    has: function (t, k) {
      return typeof k === 'symbol' ? false : (host.envHas(k) || k in t);
    },
    deleteProperty: function (t, k) {
      if (typeof k !== 'symbol') host.envDelete(k);
      return true;
    },
    ownKeys: function () { return host.envKeys(); },
    getOwnPropertyDescriptor: function (t, k) {
      if (typeof k === 'symbol' || !host.envHas(k)) return undefined;
      return { value: host.envGet(k), writable: true, enumerable: true, configurable: true };
    },
    defineProperty: function (t, k, desc) {
      if (typeof k === 'symbol') return false;
      if (desc.get || desc.set) {
        var accessor = new TypeError("'process.env' does not accept an accessor(getter/setter) descriptor");
        accessor.code = 'ERR_INVALID_OBJECT_DEFINE_PROPERTY';
        throw accessor;
      }
      if (!('value' in desc) || desc.configurable !== true ||
          desc.writable !== true || desc.enumerable !== true) {
        var plain = new TypeError("'process.env' only accepts a configurable, writable, and enumerable data descriptor");
        plain.code = 'ERR_INVALID_OBJECT_DEFINE_PROPERTY';
        throw plain;
      }
      host.envSet(k, desc.value);
      return true;
    },
  });
  Object.defineProperty(proc, 'env', {
    value: envProxy, writable: true, enumerable: true, configurable: true,
  });

  Object.defineProperty(proc, 'exitCode', {
    enumerable: true,
    configurable: false,
    get: function () { return host.getExitCode(); },
    set: function (code) { host.setExitCode(validateExitCode(code)); },
  });

  var rawExit = host.exit;
  proc.exit = function exit(code) {
    if (code === undefined) return rawExit(host.getExitCode());
    return rawExit(validateExitCode(code));
  };

  var rawHrtime = host.hrtime;
  function hrtime(time) {
    if (time !== undefined) {
      if (!Array.isArray(time)) {
        throw invalidArg('time', 'an instance of Array', time);
      }
      if (time.length !== 2) {
        throw rangeErr('The value of "time" is out of range. It must be 2. Received ' + time.length);
      }
    }
    return time === undefined ? rawHrtime() : rawHrtime(time[0], time[1]);
  }
  hrtime.bigint = function bigint(time) { return host.hrtimeBigint(time); };
  proc.hrtime = hrtime;

  proc.umask = function umask(mask) {
    if (mask === undefined) return host.umask();
    if (typeof mask === 'number') {
      if (!isFinite(mask) || Math.floor(mask) !== mask || mask < 0 || mask > 0xFFFFFFFF) {
        throw invalidArg('mask', 'of type number', mask);
      }
      return host.umask(mask);
    }
    if (typeof mask === 'string') {
      if (!/^[0-7]+$/.test(mask)) {
        var e = new TypeError('The argument \'mask\' must be a 32-bit unsigned integer or an octal string. Received ' + JSON.stringify(mask));
        e.code = 'ERR_INVALID_ARG_VALUE';
        throw e;
      }
      return host.umask(parseInt(mask, 8));
    }
    throw invalidArg('mask', 'of type number or string', mask);
  };

  proc.chdir = function chdir(directory) {
    if (typeof directory !== 'string') {
      throw invalidArg('directory', 'of type string', directory);
    }
    try {
      return host.chdir(directory);
    } catch (e) {
      var m = /^([A-Z][A-Z0-9]+):/.exec(String(e.message));
      if (m) e.code = m[1];
      e.syscall = 'chdir';
      e.dest = directory;
      e.path = host.cwd();
      throw e;
    }
  };

  proc.kill = function kill(pid, sig) {
    if (typeof pid !== 'number' || Math.floor(pid) !== pid) {
      throw invalidArg('pid', 'of type number', pid);
    }
    var name = sig === undefined ? 'SIGTERM' : sig;
    try {
      return host.kill(pid, name);
    } catch (e) {
      var m = /^(E[A-Z0-9]+)/.exec(String(e.message));
      if (m) e.code = m[1];
      e.syscall = 'kill';
      throw e;
    }
  };	proc.ref = function ref(maybeRefable) {
    if (maybeRefable == null) return;
    var fn = maybeRefable[Symbol.for('nodejs.ref')] || maybeRefable.ref;
    if (typeof fn === 'function') fn.call(maybeRefable);
  };
  proc.unref = function unref(maybeRefable) {
    if (maybeRefable == null) return;
    var fn = maybeRefable[Symbol.for('nodejs.unref')] || maybeRefable.unref;
    if (typeof fn === 'function') fn.call(maybeRefable);
  };

  // --- active resources ---------------------------------------------------
  // Node answers getActiveResourcesInfo() from libuv, which we cannot see, so
  // the timer globals installed by the timers builtin are wrapped here. The
  // wrap is idempotent and keeps every argument/return value intact.
  var activeResources = new Map();
  function trackTimers() {
    var g = globalThis;
    ['setTimeout', 'setInterval', 'setImmediate'].forEach(function (name) {
      var raw = g[name];
      if (typeof raw !== 'function' || raw.__godeTracked) return;
      var isImmediate = name === 'setImmediate';
      var isInterval = name === 'setInterval';
      var wrapped = function (fn) {
        var args = Array.prototype.slice.call(arguments);
        if (typeof fn !== 'function') return raw.apply(g, args);
        var handle;
        args[0] = function () {
          if (isImmediate && handle) activeResources.delete(handle);
          try {
            return fn.apply(this, arguments);
          } finally {
            // A one-shot Timeout stays active while its own callback runs.
            if (!isImmediate && !isInterval && handle) activeResources.delete(handle);
          }
        };
        handle = raw.apply(g, args);
        if (handle && typeof handle === 'object') {
          activeResources.set(handle, isImmediate ? 'Immediate' : 'Timeout');
        }
        return handle;
      };
      wrapped.__godeTracked = true;
      g[name] = wrapped;
      ['clearTimeout', 'clearInterval', 'clearImmediate'].forEach(function (clearName) {
        var rawClear = g[clearName];
        if (typeof rawClear !== 'function' || rawClear.__godeTracked) return;
        var clearWrapped = function (handle) {
          if (handle && typeof handle === 'object') activeResources.delete(handle);
          return rawClear.apply(g, arguments);
        };
        clearWrapped.__godeTracked = true;
        g[clearName] = clearWrapped;
      });
    });
  }
  trackTimers();
  proc.getActiveResourcesInfo = function getActiveResourcesInfo() {
    trackTimers();
    var out = [];
    activeResources.forEach(function (name) { out.push(name); });
    return out;
  };
  proc.getActiveRequests = proc.getActiveResourcesInfo;

  // The NODE_OPTIONS allow-list. Node exposes a frozen, Set-like object whose
  // has() accepts every documented spelling and whose iteration only yields
  // canonical flags (test-process-env-allowed-flags).
  var allowedLong = [
    'perf-basic-prof', 'perf-basic-prof-only-functions', 'perf-prof',
    'perf-prof-unwinding-info', 'stack-trace-limit', 'inspect-brk',
    'no-warnings', 'trace-warnings', 'no-deprecation', 'throw-deprecation',
    'pending-deprecation', 'trace-deprecation', 'enable-source-maps',
    'expose-gc', 'use-strict', 'use-largepages', 'max-old-space-size',
    'max-semi-space-size', 'stack-size', 'title', 'icu-data-dir',
    'openssl-config', 'redirect-warnings', 'experimental-vm-modules',
    'experimental-worker', 'inspect', 'inspect-port', 'inspect-publish-uid',
    'conditions', 'require', 'import', 'loader', 'input-type', 'test',
    'test-only', 'test-reporter', 'watch', 'trace-event-categories',
    'trace-event-file-pattern', 'unhandled-rejections',
    'abort-on-uncaught-exception', 'disable-warning', 'permission',
    'allow-fs-read', 'allow-fs-write'
  ];
  var allowedShort = ['-r'];
  var longNames = new Set(allowedLong);
  var shortNames = new Set(allowedShort);
  var canonicalFlags = allowedLong.map(function (f) { return '--' + f; }).concat(allowedShort);
  function allowedFlag(flag) {
    if (typeof flag !== 'string') return false;
    var name = flag;
    var eq = name.indexOf('=');
    if (eq !== -1) name = name.slice(0, eq);
    if (name.slice(0, 2) === '--') return longNames.has(name.slice(2).replace(/_/g, '-'));
    if (name.charAt(0) === '-') return shortNames.has(name);
    return longNames.has(name.replace(/_/g, '-')) || shortNames.has('-' + name);
  }
  var allowedFlags = new Set(canonicalFlags);
  function defineAllowed(prop, value) {
    Object.defineProperty(allowedFlags, prop, {
      value: value, writable: true, enumerable: false, configurable: true
    });
  }
  defineAllowed('has', allowedFlag);
  defineAllowed('add', function () { return allowedFlags; });
  defineAllowed('delete', function () { return false; });
  defineAllowed('clear', function () {});
  Object.defineProperty(allowedFlags, 'size', {
    configurable: true, enumerable: false,
    get: function () { return canonicalFlags.length; }
  });
  defineAllowed('keys', function () { return new Set(canonicalFlags).keys(); });
  defineAllowed('values', function () { return new Set(canonicalFlags).values(); });
  defineAllowed('entries', function () { return new Set(canonicalFlags).entries(); });
  defineAllowed('forEach', function (fn, thisArg) {
    for (var i = 0; i < canonicalFlags.length; i++) {
      fn.call(thisArg, canonicalFlags[i], canonicalFlags[i], allowedFlags);
    }
  });
  if (typeof Symbol === 'function' && Symbol.iterator) {
    defineAllowed(Symbol.iterator, function () { return new Set(canonicalFlags).values(); });
  }
  Object.freeze(allowedFlags);
  Object.defineProperty(proc, 'allowedNodeEnvironmentFlags', {
    enumerable: false,
    configurable: true,
    writable: false,
    value: allowedFlags,
  });
  Object.defineProperty(proc, 'debugPort', {
    enumerable: true, configurable: true, writable: true, value: 9229,
  });

  // process.config values are read-only; assignment throws in strict mode.
  if (proc.config && proc.config.variables) {
    Object.defineProperty(proc.config, 'variables', {
      value: proc.config.variables, writable: false, enumerable: true, configurable: false
    });
  }

  // process.binding only exposes the legacy allow-list (test-process-binding).
  proc.binding = function binding(moduleName) {
    if (typeof moduleName !== 'string') {
      throw invalidArg('module', 'of type string', moduleName);
    }
    return host.binding(moduleName);
  };

  // loadEnvFile surfaces ENOENT with Node's code/path/syscall fields.
  proc.loadEnvFile = function loadEnvFile(path) {
    if (path === undefined) path = '.env';
    else if (typeof path !== 'string') throw invalidArg('path', 'of type string', path);
    var result = host.loadEnvFile(path);
    if (result && result.error) {
      var err = new Error(result.message);
      err.code = result.code;
      err.path = result.path;
      err.syscall = result.syscall;
      throw err;
    }
  };

  // process inherits from EventEmitter, with its own constructor on the
  // intermediate prototype (test-process-prototype).
  if (typeof EventEmitter === 'function' && EventEmitter.prototype) {
    var processProto = Object.create(EventEmitter.prototype);
    var Process = function Process() {
      throw new TypeError('Illegal constructor');
    };
    Object.defineProperty(processProto, 'constructor', {
      value: Process, writable: true, enumerable: false, configurable: true
    });
    Process.prototype = processProto;
    Object.setPrototypeOf(proc, processProto);
  }

  var versions = proc.versions;
	Object.keys(versions).forEach(function (key) {
    Object.defineProperty(versions, key, {
      value: versions[key], writable: false, enumerable: true, configurable: false,
    });
  });
})`, "[gode:process]"); err == nil && prelude.IsFunction() {
		if _, callErr := prelude.Object().Call(js.Object{}, proc.V(), host.V(), ee); callErr != nil {
			fmt.Fprintf(e.Err(), "process prelude: %v\n", callErr)
		}
	} else if err != nil {
		fmt.Fprintf(e.Err(), "process prelude: %v\n", err)
	}

	e.Global().Set("process", proc.V())
	// Node brands the process object: process.constructor.name is 'process'
	// and Object.prototype.toString reports '[object process]'.
	if init, err := e.JS.Eval(`(function () {
  Object.defineProperty(globalThis.process, 'constructor', {
    value: function process() {}, configurable: true, writable: true
  });
  Object.defineProperty(globalThis.process, Symbol.toStringTag, {
    value: 'process', configurable: true
  });
})`, "[gode:process identity]"); err == nil && init.Valid() && init.IsFunction() {
		_, _ = init.Object().Call(js.Object{})
	}
	return proc
}

// execArgv extracts the runtime flags that precede the entry script.
func execArgv(argv []string) []string {
	out := []string{}
	for _, a := range argv[1:] {
		if a == "--" {
			break
		}
		if !strings.HasPrefix(a, "-") {
			break
		}
		out = append(out, a)
	}
	return out
}

func jsTypeName(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return "undefined"
	}
	if v.IsNull() {
		return "null"
	}
	if v.IsString() {
		return "string ('" + v.Str() + "')"
	}
	if v.IsBool() {
		return "boolean (" + strconv.FormatBool(v.Bool()) + ")"
	}
	if v.IsNumber() {
		return "number (" + strconv.FormatFloat(v.Float(), 'g', -1, 64) + ")"
	}
	if v.IsFunction() {
		return "function"
	}
	if v.IsArray() {
		return "an instance of Array"
	}
	return "an instance of Object"
}

// validatePrevUsage checks process.cpuUsage(prev)/threadCpuUsage(prev).
func validatePrevUsage(prev js.Object) (int64, int64, error) {
	user, err := prevUsageValue(prev, "user")
	if err != nil {
		return 0, 0, err
	}
	system, err := prevUsageValue(prev, "system")
	if err != nil {
		return 0, 0, err
	}
	return user, system, nil
}

func prevUsageValue(prev js.Object, field string) (int64, error) {
	v := prev.Get(field)
	if !v.IsNumber() {
		return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "prevValue.`+field+`" property must be of type number.`+invalidArgSuffix(v))
	}
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 || f > 9007199254740991 {
		return 0, js.ThrowCode("RangeError", "ERR_INVALID_ARG_VALUE",
			"The property 'prevValue."+field+"' is invalid. Received "+formatNumber(f))
	}
	return int64(f), nil
}

func formatNumber(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func invalidArgSuffix(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return " Received undefined"
	case v.IsNull():
		return " Received null"
	case v.IsString():
		return " Received type string ('" + v.Str() + "')"
	case v.IsNumber():
		return " Received type number (" + formatNumber(v.Float()) + ")"
	case v.IsBool():
		return " Received type boolean (" + strconv.FormatBool(v.Bool()) + ")"
	case v.IsArray():
		return " Received an instance of Array"
	case v.IsFunction():
		fnName := ""
		if n := v.Object().Get("name"); n.IsString() {
			fnName = n.Str()
		}
		return " Received function " + fnName
	default:
		return " Received an instance of Object"
	}
}

// nodeInspectErrorValue renders a value the way Node's inspect does inside a
// "Received ..." clause (execve argument/env validation).
func nodeInspectErrorValue(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsString():
		return nodeInspectString(v.Str())
	case v.IsNumber():
		return formatNumber(v.Float())
	case v.IsBool():
		return strconv.FormatBool(v.Bool())
	case v.IsArray():
		return "[]"
	}
	return "{}"
}

// nodeInspectString quotes a string with Node's single-quote/escape style.
func nodeInspectString(s string) string {
	var b strings.Builder
	b.WriteByte('\'')
	for _, r := range s {
		switch r {
		case '\'':
			b.WriteString("\\'")
		case '\\':
			b.WriteString("\\\\")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		case 0:
			b.WriteString("\\x00")
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

// inspectEnvObject renders an env bag as { key: value, ... } for the execve
// validation message.
func inspectEnvObject(e *env.Env, obj js.Object) string {
	keys := envKeys(e, obj)
	if len(keys) == 0 {
		return "{}"
	}
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + ": " + nodeInspectErrorValue(obj.Get(k))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// deferEmit schedules fn as a nextTick callback. When fn returns an error the
// host bridge rethrows it, which is how --throw-deprecation surfaces as an
// uncaught exception.
func deferEmit(e *env.Env, fn func() error) {
	holder := e.JS.Obj()
	holder.SetFunc("run", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), fn()
	})
	e.NextTick(holder.Get("run"))
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

// jsValueWriter is the subset both stdio sinks share.
type jsValueWriter interface {
	Write([]byte) (int, error)
}

// setID implements process.setuid/seteuid/setgid/setegid validation and maps
// the syscall error onto Node's code/syscall shape.
func setID(name string, args []js.Value, apply func(int) error) error {
	id := arg(args, 0)
	var value int
	switch {
	case id.IsNumber():
		f := id.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 || f > 4294967295 || math.Floor(f) != f {
			return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				`The value of "id" is out of range. It must be >= 0 && <= 4294967295. Received `+formatNumber(f))
		}
		value = int(f)
	case id.IsString():
		isGroup := name == "setgid" || name == "setegid"
		var field string
		if isGroup {
			grp, err := user.LookupGroup(id.Str())
			if err != nil {
				return js.ThrowCode("Error", "ERR_UNKNOWN_CREDENTIAL",
					"Group identifier does not exist: "+id.Str())
			}
			field = grp.Gid
		} else {
			u, err := user.Lookup(id.Str())
			if err != nil {
				return js.ThrowCode("Error", "ERR_UNKNOWN_CREDENTIAL",
					"User identifier does not exist: "+id.Str())
			}
			field = u.Uid
		}
		n, convErr := strconv.Atoi(field)
		if convErr != nil {
			if isGroup {
				return js.ThrowCode("Error", "ERR_UNKNOWN_CREDENTIAL",
					"Group identifier does not exist: "+id.Str())
			}
			return js.ThrowCode("Error", "ERR_UNKNOWN_CREDENTIAL",
				"User identifier does not exist: "+id.Str())
		}
		value = n
	default:
		return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "id" argument must be one of type number or string.`+invalidArgSuffix(id))
	}
	if err := apply(value); err != nil {
		code := errCode(err)
		return js.ThrowCode("Error", code, code+", "+err.Error())
	}
	return nil
}

// formatArgs applies util.format's substitutions for _rawDebug, which must
// keep working even when util itself cannot be loaded.
func formatArgs(e *env.Env, args []js.Value) string {
	if len(args) == 0 {
		return ""
	}
	if !args[0].IsString() {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.Str()
		}
		return strings.Join(parts, " ")
	}
	format := args[0].Str()
	var b strings.Builder
	next := 1
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' || i+1 >= len(format) {
			b.WriteByte(c)
			continue
		}
		spec := format[i+1]
		i++
		if spec == '%' {
			b.WriteByte('%')
			continue
		}
		v := js.Value{}
		if next < len(args) {
			v = args[next]
		}
		next++
		switch spec {
		case 's':
			b.WriteString(v.Str())
		case 'd', 'i', 'f':
			b.WriteString(strconv.FormatFloat(v.Float(), 'f', -1, 64))
		case 'j', 'o', 'O':
			if s, err := v.JSON(0); err == nil {
				b.WriteString(s)
			} else {
				b.WriteString("[Circular]")
			}
		default:
			b.WriteByte('%')
			b.WriteByte(spec)
		}
	}
	for ; next < len(args); next++ {
		b.WriteByte(' ')
		b.WriteString(args[next].Str())
	}
	return b.String()
}

// envKeys lists an object's own enumerable string keys in insertion order.
func envKeys(e *env.Env, obj js.Object) []string {
	keys := e.Global().Get("Object").Object().Get("keys")
	if !keys.IsFunction() {
		return nil
	}
	res, err := keys.Object().Call(js.Object{}, obj.V())
	if err != nil || !res.IsArray() {
		return nil
	}
	arr := res.Object()
	out := make([]string, arr.Length())
	for i := range out {
		out[i] = arr.At(i).Str()
	}
	return out
}

// warningName reads the `name` of a warning object.
func warningName(v js.Value) string {
	if !v.Valid() || !v.IsObject() {
		return ""
	}
	if n := v.Object().Get("name"); n.IsString() {
		return n.Str()
	}
	return ""
}

func strArray(e *env.Env, items []string) js.Object {
	refs := make([]js.Value, len(items))
	for i, s := range items {
		refs[i] = e.JS.Str(s)
	}
	return e.JS.Arr(refs)
}

func sameValue(e *env.Env, a, b js.Value) bool {
	is := e.Global().Get("Object").Object().Get("is")
	if !is.IsFunction() {
		return false
	}
	out, err := is.Object().Call(js.Object{}, a, b)
	return err == nil && out.Bool()
}

func nodeArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "386":
		return "ia32"
	default:
		return runtime.GOARCH
	}
}

func memRSS() int64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0
	}
	return ru.Maxrss * 1024
}

func cpuUsage() (user, system int64) {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0, 0
	}
	return ru.Utime.Sec*1e6 + ru.Utime.Usec, ru.Stime.Sec*1e6 + ru.Stime.Usec
}

func availableMemory() int64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					return kb * 1024
				}
			}
		}
	}
	return 0
}

func parseSignal(v js.Value) (syscall.Signal, error) {
	if v.IsNumber() {
		return syscall.Signal(v.Int()), nil
	}
	if v.IsString() {
		name := strings.ToUpper(v.Str())
		if name == "SIGTERM" {
			return syscall.SIGTERM, nil
		}
		sigs := map[string]syscall.Signal{
			"SIGKILL": syscall.SIGKILL,
			"SIGINT":  syscall.SIGINT,
			"SIGQUIT": syscall.SIGQUIT,
			"SIGABRT": syscall.SIGABRT,
			"SIGHUP":  syscall.SIGHUP,
			"SIGUSR1": syscall.SIGUSR1,
			"SIGUSR2": syscall.SIGUSR2,
			"SIGPIPE": syscall.SIGPIPE,
			"SIGALRM": syscall.SIGALRM,
			"SIGCHLD": syscall.SIGCHLD,
		}
		if s, ok := sigs[name]; ok {
			return s, nil
		}
		return 0, js.ThrowCode("TypeError", "ERR_UNKNOWN_SIGNAL", "Unknown signal: "+v.Str())
	}
	return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", "The \"signal\" argument must be of type number or string")
}

func errCode(err error) string {
	if errno, ok := err.(syscall.Errno); ok {
		switch errno {
		case syscall.ESRCH:
			return "ESRCH"
		case syscall.EPERM:
			return "EPERM"
		case syscall.EINVAL:
			return "EINVAL"
		}
	}
	return "UNKNOWN"
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"process.abort",
		"process.availableMemory",
		"process.bigint",
		"process.chdir",
		"process.constrainedMemory",
		"process.cpuUsage",
		"process.cwd",
		"process.dlopen",
		"process.emitWarning",
		"process.exit",
		"process.getBuiltinModule",
		"process.getegid",
		"process.geteuid",
		"process.getgid",
		"process.getgroups",
		"process.getuid",
		"process.hasUncaughtExceptionCaptureCallback",
		"process.hrtime",
		"process.kill",
		"process.loadEnvFile",
		"process.memoryUsage",
		"process.nextTick",
		"process.ref",
		"process.register",
		"process.registerBeforeExit",
		"process.resourceUsage",
		"process.rss",
		"process.setSourceMapsEnabled",
		"process.setUncaughtExceptionCaptureCallback",
		"process.setgroups",
		"process.threadCpuUsage",
		"process.umask",
		"process.unref",
		"process.unregister",
		"process.uptime",
		"process.getActiveResourcesInfo",
		"process.execve",
		"process.Type",
		"process.send",
		"process.disconnect",
		"process.initgroups",
		"process.traceProcessWarnings",
		"process.setuid",
		"process.seteuid",
		"process.setgid",
		"process.setegid",
	}
}

// errorLikeValue reports whether v looks like an Error instance. Host-created
// error subclasses (CustomWarning, ...) are matched by their message string
// because the JSC bridge does not expose the [[ErrorData]] slot for them.
func errorLikeValue(v js.Value) bool {
	if !v.Valid() || !v.IsObject() || v.IsArray() || v.IsFunction() {
		return false
	}
	if v.IsError() {
		return true
	}
	if msg := v.Object().Get("message"); msg.IsString() {
		return true
	}
	return false
}

// isPlainOptions reports whether v is the options bag form of emitWarning.
func isPlainOptions(v js.Value) bool {
	return v.Valid() && !v.IsUndefined() && v.IsObject() && !v.IsArray() && !v.IsFunction()
}
