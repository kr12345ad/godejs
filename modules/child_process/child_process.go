// Package child_process implements Node's child_process builtin.
//
// spawn() starts the real process and returns immediately, matching Node: a
// small repeating timer (the module's pump) checks for completion, then emits
// the buffered output, 'exit' and 'close' in Node's order. The synchronous
// entry points (spawnSync, execSync, execFileSync) still run to completion
// before returning.
package child_process

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("child_process", Install) }

type module struct {
	e        *env.Env
	cpProto  js.Object
	streams  js.Object
	children map[int]*procState
	seq      int
	thrower  js.Value
	childIPC *childIPC // set in a forked child
	ipcChan  js.Object // process.channel in a forked child
}

type procState struct {
	id           int
	cp           js.Object
	stdout       js.Object
	stderr       js.Object
	stdin        js.Object
	outBytes     []byte
	errBytes     []byte
	status       int
	signal       string
	spawnErr     js.Object
	callback     js.Value
	encoding     string
	cmdline      string
	code         string
	killed       bool
	bufferStream string
	callbackDone bool
	maxBuf       int
	signalValue  js.Value
	async        *asyncState
	finished     bool
	timedOut     bool
	pid          int
	exitCode     int
	signalCode   string
	spawnErrFlag bool
	connected    bool
	disconnected bool
	aborted      bool
	unrefed      bool
	killTimer    int
}

// asyncState is the live process behind an asynchronous spawn.
//
// The child's output is copied by os/exec itself (cmd.Stdout/cmd.Stderr point
// at locked buffers), so by the time cmd.Wait returns every byte is buffered.
// A goroutine closes done once Wait returns; the pump timer then emits the
// events on the JavaScript goroutine, which is the only one that may touch JSC.
type asyncState struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	stdinW  io.WriteCloser
	stdinR  *os.File
	out     *lockedBuf
	err     *lockedBuf
	done    chan struct{}
	wait    error
	timer   int
	ipc     *ipcChannel
	ipcDone bool
	inbox   []string
}

// lockedBuf is an io.Writer the pump reads while os/exec copies into it.
type lockedBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuf) Bytes() []byte {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]byte(nil), l.b.Bytes()...)
}

type runOptions struct {
	cmd         string
	args        []string
	spawnArgs   []string // the real argv, for spawnargs/spawnfile
	spawnFile   string
	ipc         bool
	detached    bool
	shell       bool
	shellArg    string
	argv0       string
	cwd         string
	env         []string
	input       []byte
	timeout     time.Duration
	maxBuf      int
	signal      string
	aborted     bool
	signalValue js.Value
	inherit     [3]bool
	ignore      [3]bool
	encoding    string
	argsGiven   bool
	optErr      error
}

func (o runOptions) killSignal() string {
	if o.signal != "" {
		return o.signal
	}
	return "SIGTERM"
}

type runResult struct {
	pid          int
	stdout       []byte
	stderr       []byte
	status       int
	signal       string
	startErr     error
	code         string
	message      string
	killed       bool
	timedOut     bool
	bufferStream string
	errno        int
	syscall      string
	path         string
	spawnargs    []string
}

// Install builds the child_process module's exports object.
func Install(e *env.Env) js.Object {
	m := &module{e: e, children: map[int]*procState{}}
	m.streams = e.JS.Obj().Keep()
	m.cpProto = m.makeCPProto().Keep()

	exports := e.JS.Obj()

	host := e.JS.Obj()
	host.SetFuncHidden("sync", m.hostSync)
	host.SetFuncHidden("spawnSync", m.hostSpawnSync)
	host.SetFuncHidden("spawn", m.hostSpawn)
	host.SetFuncHidden("spawnShell", m.hostSpawnShell)
	host.SetFuncHidden("spawnFile", m.hostSpawnFile)
	host.SetFuncHidden("fork", m.hostFork)
	host.SetFuncHidden("spawnInto", m.hostSpawnInto)
	host.SetFuncHidden("abort", m.hostAbort)

	wrapper, err := e.JS.Eval(`(function (host, ChildProcessProto, exports) {
  'use strict';
  function ChildProcess() {
    if (!(this instanceof ChildProcess)) {
      throw new TypeError("Class constructor ChildProcess cannot be invoked without 'new'");
    }
    this.stdin = null;
    this.stdout = null;
    this.stderr = null;
    this.stdio = [null, null, null];
    this.spawnfile = undefined;
    this.spawnargs = [];
    this.pid = undefined;
    this.connected = false;
    this.channel = undefined;
    this.exitCode = null;
    this.signalCode = null;
    this.killed = false;
  }
  ChildProcess.prototype = ChildProcessProto;
  Object.defineProperty(ChildProcess.prototype, 'constructor', {
    value: ChildProcess, writable: true, configurable: true
  });
  function describe(v) {
    if (v === null) return 'null';
    if (v === undefined) return 'undefined';
    var t = typeof v;
    if (t === 'number' || t === 'boolean' || t === 'bigint') return 'type ' + t + ' (' + String(v) + ')';
    if (t === 'string') return 'type string (\'' + v + '\')';
    if (t === 'function') return 'function ' + (v.name || '<anonymous>');
    if (Array.isArray(v)) return 'an instance of Array';
    if (v.constructor && v.constructor.name) return 'an instance of ' + v.constructor.name;
    return 'an instance of Object';
  }
  function invalidArg(name, v) {
    var e = new TypeError('The "' + name + '" argument must be of type string. Received ' + describe(v));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  // Node's invalidArgTypeHelper(): ' Received type number (1)', ' Received
  // an instance of Array', and so on. describe() already renders the value,
  // so it only needs the leading clause.
  function received(v) { return ' Received ' + describe(v); }
  function argType(name, v) {
    var e = new TypeError('The "' + name + '" argument must be of type object.' + received(v));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function propType(name, expected, v) {
    var e = new TypeError('The "' + name + '" property must be ' + expected + '.' + received(v));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  // validateArgumentNullCheck(): ERR_INVALID_ARG_VALUE for embedded NULs.
  function checkNull(name, v) {
    if (typeof v === 'string' && v.indexOf('\u0000') !== -1) {
      var e = new TypeError("The argument '" + name + "' must be a string without null bytes. Received '" + v + "'");
      e.code = 'ERR_INVALID_ARG_VALUE';
      throw e;
    }
  }
  function checkCommonOptions(o) {
    if (!o || typeof o !== 'object') return;
    if (typeof o.cwd === 'string') checkNull('options.cwd', o.cwd);
    if (typeof o.argv0 === 'string') checkNull('options.argv0', o.argv0);
    if (typeof o.shell === 'string') checkNull('options.shell', o.shell);
    if (o.env && typeof o.env === 'object') {
      for (var k in o.env) {
        var v = o.env[k];
        if (v === undefined) continue;
        checkNull("options.env['" + k + "']", k);
        checkNull("options.env['" + k + "']", String(v));
      }
    }
  }
  function fail(r) {
    var err = r.code === 'ERR_CHILD_PROCESS_STDIO_MAXBUFFER' ?
      new RangeError(r.message) : new Error(r.message);
    if (r.code === 'ABORT_ERR') err.name = 'AbortError';
    if (r.code) err.code = r.code;
    if (r.errno !== undefined) err.errno = r.errno;
    if (r.syscall !== undefined) err.syscall = r.syscall;
    if (r.path !== undefined) err.path = r.path;
    if (r.spawnargs !== undefined) err.spawnargs = r.spawnargs;
    err.status = r.status;
    err.signal = r.signal;
    if (r.pid) err.pid = r.pid;
    err.cmd = r.cmd;
    err.killed = r.killed;
    err.stdout = r.stdout;
    err.stderr = r.stderr;
    err.output = [null, r.stdout, r.stderr];
    throw err;
  }
  function norm(args, options) {
    if (args === undefined || args === null) return [[], options || {}];
    if (Array.isArray(args)) {
      // Node stringifies every element with a template literal rather than
      // rejecting numbers or objects, then rejects embedded NUL bytes.
      var out = new Array(args.length);
      for (var i = 0; i < args.length; i++) {
        out[i] = '' + args[i];
        checkNull('args[' + i + ']', out[i]);
      }
      return [out, options || {}];
    }
    if (typeof args === 'object') return [[], args];
    throw invalidArg('args', args);
  }
  function invalidOptions(v) {
    var e = new TypeError('The "options" argument must be of type object. Received ' + describe(v));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function checkOptions(o) {
    if (o !== undefined && o !== null && typeof o !== 'object') throw invalidOptions(o);
    return o || {};
  }
  function checkPath(name, v) {
    if (typeof v === 'string' && v.indexOf('\u0000') !== -1) {
      var e = new TypeError("The argument '" + name + "' must be a string without null bytes. Received '" + v + "'");
      e.code = 'ERR_INVALID_ARG_VALUE';
      throw e;
    }
  }
  function abortError() {
    var e = new Error('The operation was aborted');
    e.name = 'AbortError';
    e.code = 'ABORT_ERR';
    return e;
  }
  function attachSignal(cp, options, isExec) {
    var s = options && options.signal;
    if (!s) return cp;
    var signal = (options.killSignal || 'SIGTERM');
    if (typeof signal !== 'string') signal = 'SIGTERM';
    if (s.aborted) {
      // The Go side starts the child and then aborts it, so 'exit' fires with
      // the kill signal exactly like Node.
      return cp;
    }
    s.addEventListener('abort', function () {
      if (!host.abort(cp, signal)) return;
      if (!isExec) {
        var err = abortError();
        if (s.reason !== undefined) err.cause = s.reason;
        cp.emit('error', err);
      }
    }, { once: true });
    return cp;
  }
  function execSync(command, options) {
    if (typeof command !== 'string') throw invalidArg('command', command);
    checkPath('command', command);
    var opts = checkOptions(options);
    checkCommonOptions(opts);
    // execSync defaults to utf8; an explicitly supplied encoding (even
    // undefined) follows exec's Buffer rules instead.
    if (!('encoding' in opts)) opts = Object.assign({}, opts, { encoding: 'utf8' });
    var r = host.sync(command, [], opts, true);
    if (r.error || r.status !== 0 || r.signal !== null) fail(r);
    return (r.encoding && r.encodingName !== 'buffer') ? r.stdoutText : r.stdout;
  }
  function execFileSync(file, args, options) {
    if (typeof file !== 'string') throw invalidArg('file', file);
    checkPath('file', file);
    var n = norm(args, options);
    checkOptions(n[1]);
    checkCommonOptions(n[1]);
    var r = host.sync(file, n[0], n[1], false);
    if (r.error || r.status !== 0 || r.signal !== null) fail(r);
    return (r.encoding && r.encodingName !== 'buffer') ? r.stdoutText : r.stdout;
  }
  function spawnSync(file, args, options) {
    if (typeof file !== 'string') throw invalidArg('file', file);
    checkPath('file', file);
    var command = file;
    var n = norm(args, options);
    checkOptions(n[1]);
    checkCommonOptions(n[1]);
    checkSerialization(n[1]);
    checkInput(n[1]);
    return host.spawnSync(command, n[0], n[1]);
  }
  // Node validates options.serialization in normalizeSpawnArguments, so every
  // spawn entry point rejects an unknown mode with ERR_INVALID_ARG_VALUE.
  function inspectOptionValue(v) {
    try {
      var u = require('util');
      if (u && typeof u.inspect === 'function') return u.inspect(v);
    } catch (e) {}
    return String(v);
  }
  function checkSerialization(options) {
    if (!options || typeof options !== 'object') return;
    if (options.serialization !== undefined &&
        options.serialization !== 'json' && options.serialization !== 'advanced') {
      var e = new TypeError("The property 'options.serialization' must be one of: " +
        "undefined, 'json', 'advanced'. Received " + inspectOptionValue(options.serialization));
      e.code = 'ERR_INVALID_ARG_VALUE';
      throw e;
    }
  }
  // spawnSync's input option only accepts string or binary data; anything else
  // fails before the process is created.
  function checkInput(options) {
    if (!options || options.input === undefined) return;
    var v = options.input;
    var ok = typeof v === 'string';
    if (!ok && v !== null && typeof v === 'object' && typeof v.byteLength === 'number') {
      var tag = Object.prototype.toString.call(v);
      ok = tag === '[object DataView]' || tag === '[object ArrayBuffer]' || v.buffer !== undefined;
    }
    if (!ok) {
      var e = new TypeError('The "options.input" property must be of type string or ' +
        'an instance of Buffer, TypedArray, or DataView. Received ' + describe(v));
      e.code = 'ERR_INVALID_ARG_TYPE';
      throw e;
    }
  }
  function checkSignal(options) {
    if (options && options.signal !== undefined) {
      var s = options.signal;
      if (typeof s !== 'object' || s === null || typeof s.aborted !== 'boolean') {
        var e = new TypeError('The "options.signal" property must be an instance of AbortSignal. Received ' + typeof s);
        e.code = 'ERR_INVALID_ARG_TYPE';
        throw e;
      }
    }
  }
  function spawn(file, args, options) {
    if (typeof file !== 'string') throw invalidArg('file', file);
    checkPath('file', file);
    var n = norm(args, options);
    checkOptions(n[1]);
    checkCommonOptions(n[1]);
    checkSerialization(n[1]);
    checkSignal(n[1]);
    return attachSignal(host.spawn(file, n[0], n[1]), n[1], false);
  }
  function fork(modulePath, args, options) {
    if (typeof modulePath !== 'string') throw invalidArg('modulePath', modulePath);
    checkNull('modulePath', modulePath);
    var n = norm(args, options);
    if (n[1] !== undefined && n[1] !== null && typeof n[1] !== 'object') throw invalidOptions(n[1]);
    checkCommonOptions(n[1]);
    if (n[1]) {
      if (typeof n[1].execPath === 'string') checkNull('options.execPath', n[1].execPath);
      if (Array.isArray(n[1].execArgv)) {
        for (var ei = 0; ei < n[1].execArgv.length; ei++) {
          checkNull('options.execArgv[' + ei + ']', '' + n[1].execArgv[ei]);
        }
      }
    }
    checkSerialization(n[1]);
    return attachSignal(host.fork(modulePath, n[0], n[1] || {}), n[1] || {}, true);
  }
  function exec(command, options, callback) {
    if (typeof command !== 'string') throw invalidArg('command', command);
    checkPath('command', command);
    if (typeof options === 'function') { callback = options; options = {}; }
    checkOptions(options);
    checkCommonOptions(options);
    checkSerialization(options);
    checkSignal(options);
    return attachSignal(host.spawnShell(command, options || {}, callback), options, true);
  }
  function execFile(file, args, options, callback) {
    if (typeof file !== 'string') throw invalidArg('file', file);
    checkPath('file', file);
    if (typeof args === 'function') { callback = args; args = []; options = {}; }
    else if (typeof options === 'function') { callback = options; options = {}; }
    else if (args !== undefined && args !== null && !Array.isArray(args)) {
      callback = typeof options === 'function' ? options : callback;
      options = args; args = [];
    }
    checkOptions(options);
    checkCommonOptions(options);
    checkSerialization(options);
    checkSignal(options);
    var fileArgs = norm(args === undefined || args === null ? [] : args, options)[0];
    return attachSignal(host.spawnFile(file, fileArgs, options || {}, callback), options, true);
  }
  // promisify(exec)/promisify(execFile): Node resolves with { stdout, stderr }
  // and exposes the live ChildProcess as promise.child.
  var kCustomPromisify = (typeof Symbol === 'function' && Symbol.for) ?
    Symbol.for('nodejs.util.promisify.custom') : null;
  // Node calls the original function outside the Promise executor, so invalid
  // options still throw synchronously instead of rejecting the promise.
  function promiseParts() {
    var parts = {};
    parts.promise = new Promise(function (resolve, reject) {
      parts.resolve = resolve;
      parts.reject = reject;
    });
    return parts;
  }
  function settle(parts) {
    return function (err, stdout, stderr) {
      if (err !== null && err !== undefined) {
        err.stdout = stdout;
        err.stderr = stderr;
        parts.reject(err);
      } else {
        parts.resolve({ stdout: stdout, stderr: stderr });
      }
    };
  }
  function promisifyExec(command, options) {
    var parts = promiseParts();
    parts.promise.child = exec(command, options, settle(parts));
    return parts.promise;
  }
  function promisifyExecFile(file, args, options) {
    var parts = promiseParts();
    parts.promise.child = execFile(file, args, options, settle(parts));
    return parts.promise;
  }
  if (kCustomPromisify) {
    exec[kCustomPromisify] = promisifyExec;
    execFile[kCustomPromisify] = promisifyExecFile;
  }
  // ChildProcess#spawn mirrors the argument checks in libuv's binding wrapper:
  // it is the only public entry point that takes the normalized options.
  ChildProcess.prototype.spawn = function (options) {
    if (options === null || typeof options !== 'object') throw argType('options', options);
    // Node validates stdio first, then serialization, then envPairs (only when
    // the stdio array carries an ipc entry), then file, then args. The order is
    // observable because every step throws a differently named error.
    function badStdio(v) {
      var e = new TypeError("The argument 'stdio' is invalid. Received " + describe(v));
      e.code = 'ERR_INVALID_ARG_VALUE';
      return e;
    }
    var stdio = options.stdio === undefined ? 'pipe' : options.stdio;
    var hasIPC = false;
    if (typeof stdio === 'string') {
      if (stdio !== 'pipe' && stdio !== 'ignore' && stdio !== 'inherit') throw badStdio(stdio);
    } else if (Array.isArray(stdio)) {
      for (var si = 0; si < stdio.length; si++) {
        var entry = stdio[si];
        if (typeof entry === 'string') {
          if (entry === 'ipc') hasIPC = true;
          else if (entry !== 'pipe' && entry !== 'ignore' && entry !== 'inherit') throw badStdio(entry);
        } else if (entry !== null && entry !== undefined &&
                   typeof entry !== 'number' && typeof entry !== 'object') {
          throw badStdio(entry);
        }
      }
    } else {
      throw badStdio(stdio);
    }
    if (options.serialization !== undefined &&
        options.serialization !== 'json' && options.serialization !== 'advanced') {
      var ser = new TypeError('The "options.serialization" property must be one of: ' +
        "'json', 'advanced'. Received " + describe(options.serialization));
      ser.code = 'ERR_INVALID_ARG_VALUE';
      throw ser;
    }
    if (hasIPC && options.envPairs !== undefined && !Array.isArray(options.envPairs)) {
      throw propType('options.envPairs', 'an instance of Array', options.envPairs);
    }
    if (typeof options.file !== 'string') throw propType('options.file', 'of type string', options.file);
    if (options.args !== undefined && !Array.isArray(options.args)) {
      throw propType('options.args', 'an instance of Array', options.args);
    }
    host.spawnInto(this, options);
    return undefined;
  };
  // Install onto the caller's exports object from JavaScript. Copying these
  // functions through the Go bridge would replace ChildProcess with a host
  // wrapper, which is callable but not constructible with 'new'.
  exports.exec = exec;
  exports.execFile = execFile;
  exports.execFileSync = execFileSync;
  exports.execSync = execSync;
  exports.fork = fork;
  exports.spawn = spawn;
  exports.spawnSync = spawnSync;
  exports.ChildProcess = ChildProcess;
})`, "child_process.js")
	if err != nil {
		panic("godejs child_process: " + err.Error())
	}
	// The wrapper installs its exports into `exports` itself, so the JS-side
	// functions (notably the ChildProcess constructor) keep their identity.
	if _, err := wrapper.Object().Call(js.Object{}, host.V(), m.cpProto.V(), exports.V()); err != nil {
		panic("godejs child_process: " + err.Error())
	}
	if t, err := e.JS.Eval(`(function (e) { throw e; })`, "[gode:child_process]"); err == nil {
		m.thrower = t
	}
	// In a forked child this gives process.send/process.disconnect/channel
	// before any user code runs.
	m.installChildIPC()
	if m.childIPC != nil {
		// Builtins are installed in registry order, so the process module may
		// still be installed after this one and replace process.send with its
		// channel-less stub. Re-apply the channel once installation settles.
		wire := e.JS.Obj()
		wire.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
			m.wireChildIPCProcess()
			return e.JS.Undefined(), nil
		})
		e.NextTick(wire.Get("f"))
	}
	return exports
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"child_process.ChildProcess",
		"child_process.ChildProcess#disconnect",
		"child_process.ChildProcess#kill",
		"child_process.ChildProcess#[Symbol.dispose]",
		"child_process.ChildProcess#send",
		"child_process.ChildProcess#ref",
		"child_process.ChildProcess#unref",
		"child_process.ChildProcess.channel",
		"child_process.ChildProcess.connected",
		"child_process.ChildProcess.exitCode",
		"child_process.ChildProcess.killed",
		"child_process.ChildProcess.pid",
		"child_process.ChildProcess.signalCode",
		"child_process.ChildProcess.spawnargs",
		"child_process.ChildProcess.spawnfile",
		"child_process.ChildProcess.stderr",
		"child_process.ChildProcess.stdin",
		"child_process.ChildProcess.stdio",
		"child_process.ChildProcess.stdout",
		"child_process.exec",
		"child_process.execFile",
		"child_process.execFileSync",
		"child_process.execSync",
		"child_process.fork",
		"child_process.spawn",
		"child_process.spawnSync",
	}
}

// ------------------------------------------------------------------ host API --

func (m *module) hostSync(this js.Object, args []js.Value) (js.Value, error) {
	cmd := arg(args, 0).Str()
	argv := stringArgs(arg(args, 1))
	opts := m.parseOptions(arg(args, 2), true)
	if opts.optErr != nil {
		return js.Value{}, opts.optErr
	}
	opts.cmd, opts.args = cmd, argv
	opts.shell = opts.shell || arg(args, 3).Bool()

	// Node's execSync/execFileSync return the raw Buffer from spawnSync unless
	// the caller asked for an encoding; they never default to utf8.
	res := m.run(opts)
	return m.resultObject(res, opts).V(), nil
}

func (m *module) hostSpawnSync(this js.Object, args []js.Value) (js.Value, error) {
	cmd := arg(args, 0).Str()
	argv := stringArgs(arg(args, 1))
	opts := m.parseOptions(arg(args, 2), false)
	if opts.optErr != nil {
		return js.Value{}, opts.optErr
	}
	opts.cmd, opts.args = cmd, argv

	res := m.run(opts)
	return m.spawnSyncObject(res, opts).V(), nil
}

func (m *module) hostSpawn(this js.Object, args []js.Value) (js.Value, error) {
	cmd := arg(args, 0).Str()
	argv := stringArgs(arg(args, 1))
	opts := m.parseOptions(arg(args, 2), false)
	if opts.optErr != nil {
		return js.Value{}, opts.optErr
	}
	opts.cmd, opts.args = cmd, argv
	// Node's spawn() has no maxBuffer; keep the complete output.
	opts.maxBuf = 0
	return m.spawnChild(opts, js.Value{}, "", false).V(), nil
}

func (m *module) hostSpawnShell(this js.Object, args []js.Value) (js.Value, error) {
	cmd := arg(args, 0).Str()
	opts := m.parseOptions(arg(args, 1), true)
	// exec always captures stdout/stderr itself; a stdio option cannot redirect
	// them, so the default is a pipe regardless of what was passed.
	opts.inherit[1], opts.inherit[2], opts.ignore[1], opts.ignore[2] = false, false, false, false
	if opts.optErr != nil {
		return js.Value{}, opts.optErr
	}
	opts.cmd = cmd
	opts.shell = true
	callback := arg(args, 2)
	enc := m.encodingOf(arg(args, 1), "utf8")
	return m.spawnChild(opts, callback, enc, true).V(), nil
}

func (m *module) hostSpawnFile(this js.Object, args []js.Value) (js.Value, error) {
	cmd := arg(args, 0).Str()
	argv := stringArgs(arg(args, 1))
	opts := m.parseOptions(arg(args, 2), true)
	if opts.optErr != nil {
		return js.Value{}, opts.optErr
	}
	opts.cmd, opts.args = cmd, argv
	opts.argsGiven = len(argv) > 0
	callback := arg(args, 3)
	enc := m.encodingOf(arg(args, 2), "utf8")
	return m.spawnChild(opts, callback, enc, true).V(), nil
}

// hostAbort records an AbortSignal firing and kills the child with the
// requested signal. exec/execFile surface it as AbortError from the callback.
func (m *module) hostAbort(this js.Object, args []js.Value) (js.Value, error) {
	cp := arg(args, 0)
	if !cp.IsObject() {
		return m.e.JS.Undefined(), nil
	}
	sig := strings.ToUpper(arg(args, 1).Str())
	if sig == "" || !validSignal(sig) {
		sig = "SIGTERM"
	}
	id := cp.Object().Get("$godeCpId").Int()
	killed := false
	if st, ok := m.children[id]; ok {
		st.aborted = true
		if !st.finished {
			killed = m.signalChild(st, sig)
		}
	}
	// Node emits the AbortError only when kill() reported that it delivered
	// the signal; aborting after 'exit' must stay silent.
	return m.e.JS.Bool(killed), nil
}

func (m *module) hostFork(this js.Object, args []js.Value) (js.Value, error) {
	modulePath := arg(args, 0).Str()
	argv := stringArgs(arg(args, 1))
	opts := m.parseOptions(arg(args, 2), false)
	if opts.optErr != nil {
		return js.Value{}, opts.optErr
	}
	opts.cmd = m.e.Opts.ExecPath
	if opts.cmd == "" {
		opts.cmd = "gode"
	}
	var execArgv []string
	if o := arg(args, 2); o.IsObject() && o.Object().Get("execArgv").IsArray() {
		execArgv = stringArgs(o.Object().Get("execArgv"))
	} else {
		// Node defaults fork's execArgv to the parent's runtime flags.
		execArgv = append([]string(nil), m.e.Opts.ExecArgv...)
	}
	opts.args = append(append(execArgv, modulePath), argv...)
	opts.maxBuf = 0
	opts.ipc = true
	// fork() never runs the module through a shell, regardless of the option.
	opts.shell = false
	opts.shellArg = ""
	// fork() inherits stdout/stderr unless `silent` is set; stdin is a pipe.
	opts.inherit[1], opts.inherit[2] = true, true
	o := arg(args, 2)
	if o.IsObject() {
		obj := o.Object()
		if obj.Get("silent").Bool() {
			opts.inherit[1], opts.inherit[2] = false, false
		}
		opts.detached = obj.Get("detached").Bool()
		if v := obj.Get("stdio"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
			if err := m.forkStdio(v, &opts); err != nil {
				return js.Value{}, err
			}
		}
	}
	return m.spawnChild(opts, js.Value{}, "", false).V(), nil
}

// hostSpawnInto implements ChildProcess#spawn. The JavaScript layer validates
// the shape of the normalized options object; this parses the same options
// spawn() accepts and starts the process on the existing instance.
func (m *module) hostSpawnInto(this js.Object, args []js.Value) (js.Value, error) {
	cp := arg(args, 0)
	optsV := arg(args, 1)
	if !cp.IsObject() || !optsV.IsObject() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "options" argument must be of type object.`)
	}
	opts := m.parseOptions(optsV, false)
	if opts.optErr != nil {
		return js.Value{}, opts.optErr
	}
	o := optsV.Object()
	opts.cmd = o.Get("file").Str()
	opts.args = stringArgs(o.Get("args"))
	opts.spawnFile = opts.cmd
	if ep := o.Get("envPairs"); ep.IsArray() {
		opts.env = stringArgs(ep)
	}
	m.spawnChildWith(cp.Object(), opts, js.Value{}, "", false)
	return m.e.JS.Undefined(), nil
}

// forkStdio applies a fork() stdio option. Node only accepts the three stdio
// shorthands for a string ('ipc' is not one of them), and an explicit array
// must carry an 'ipc' entry because fork always needs a channel.
func (m *module) forkStdio(v js.Value, opts *runOptions) error {
	if v.IsString() {
		switch v.Str() {
		case "ignore", "pipe", "inherit":
			for i := 0; i < 3; i++ {
				opts.inherit[i] = v.Str() == "inherit"
				opts.ignore[i] = v.Str() == "ignore"
			}
			return nil
		}
		return js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			"The argument 'stdio' is invalid. Received '"+v.Str()+"'")
	}
	if !v.IsArray() {
		return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "options.stdio" property must be of type string or an instance of Array. Received `+jsTypeName(v))
	}
	arr := v.Object()
	hasIPC := false
	for i := 0; i < arr.Length(); i++ {
		e := arr.At(i)
		if e.IsString() {
			switch e.Str() {
			case "ipc":
				hasIPC = true
				continue
			case "pipe", "overlapped", "ignore", "inherit":
				if i < 3 {
					opts.inherit[i] = e.Str() == "inherit"
					opts.ignore[i] = e.Str() == "ignore"
				}
			}
		} else if e.IsNumber() && i < 3 && (e.Int() == 1 || e.Int() == 2) {
			opts.inherit[i] = true
		}
	}
	if !hasIPC {
		return js.ThrowCode("Error", "ERR_CHILD_PROCESS_IPC_REQUIRED",
			"Forked processes must have an IPC channel, missing value 'ipc' in options.stdio")
	}
	return nil
}

// encodingOf returns the normalized output encoding for exec/execFile.
// Node defaults to utf8; null, 'buffer' and unknown encodings produce Buffers.
func (m *module) encodingOf(o js.Value, def string) string {
	if !o.IsObject() {
		return def
	}
	obj := o.Object()
	if !obj.Has("encoding") {
		return def
	}
	v := obj.Get("encoding")
	if !v.IsString() || !isEncoding(v.Str()) || strings.EqualFold(v.Str(), "buffer") {
		return "buffer"
	}
	return strings.ToLower(v.Str())
}

func isEncoding(s string) bool {
	switch strings.ToLower(s) {
	case "utf8", "utf-8", "hex", "base64", "latin1", "binary", "ascii",
		"ucs2", "ucs-2", "utf16le", "utf-16le", "base64url":
		return true
	}
	return false
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

func stringArgs(v js.Value) []string {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return nil
	}
	if !v.IsArray() {
		return nil
	}
	arr := v.Object()
	out := make([]string, 0, arr.Length())
	for i := 0; i < arr.Length(); i++ {
		out = append(out, arr.At(i).Str())
	}
	return out
}

func (m *module) parseOptions(v js.Value, defaults bool) runOptions {
	opts := runOptions{encoding: "", maxBuf: 1024 * 1024}
	if !v.Valid() || !v.IsObject() {
		return opts
	}
	o := v.Object()
	if cwd := o.Get("cwd"); cwd.IsString() {
		opts.cwd = cwd.Str()
	} else if cwd.IsObject() && cwd.Object().Get("protocol").IsString() {
		// Only URL-like objects are paths; any other object is invalid, which
		// getValidatedPath reports as ERR_INVALID_ARG_TYPE.
		if p, err := m.fileURLPath(cwd.Object()); err != nil {
			opts.optErr = err
		} else {
			opts.cwd = p
		}
	} else if cwd.Valid() && !cwd.IsUndefined() && !cwd.IsNull() {
		opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "options.cwd" property must be of type string or an instance of URL. Received `+jsTypeName(cwd))
		return opts
	}
	if envv := o.Get("env"); envv.IsObject() {
		opts.env = m.envFromObject(envv.Object())
	} else if envv.Valid() && !envv.IsUndefined() && !envv.IsNull() {
		opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "options.env" property must be of type object. Received `+jsTypeName(envv))
		return opts
	}
	if sh := o.Get("shell"); sh.Valid() && !sh.IsUndefined() && !sh.IsNull() && !sh.IsBool() && !sh.IsString() {
		opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "options.shell" property must be of type boolean or string. Received `+jsTypeName(sh))
		return opts
	}
	if s := o.Get("input"); s.IsString() {
		opts.input = []byte(s.Str())
	} else if b, ok := o.Get("input").Bytes(); ok {
		opts.input = b
	}
	if sh := o.Get("shell"); sh.IsString() {
		opts.shell = true
		opts.shellArg = sh.Str()
	} else if sh.IsBool() && sh.Bool() {
		opts.shell = true
	}
	if o.Has("encoding") {
		// Node distinguishes an absent encoding (module default) from an
		// explicitly falsy or unknown one, which selects Buffer output.
		if enc := o.Get("encoding"); enc.IsString() && isEncoding(enc.Str()) &&
			!strings.EqualFold(enc.Str(), "buffer") {
			opts.encoding = strings.ToLower(enc.Str())
		} else {
			opts.encoding = "buffer"
		}
	}
	if n := o.Get("timeout"); n.Valid() && !n.IsUndefined() && !n.IsNull() {
		if !n.IsNumber() || math.IsNaN(n.Float()) || n.Float() < 0 ||
			math.IsInf(n.Float(), 0) || n.Float() != math.Trunc(n.Float()) {
			opts.optErr = js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				`The value of "timeout" is out of range. It must be >= 0 && <= `+
					strconv.FormatInt(9007199254740991, 10)+`. Received `+numberText(n))
			return opts
		}
		opts.timeout = time.Duration(n.Float() * float64(time.Millisecond))
	}
	for _, name := range []string{"uid", "gid"} {
		if u := o.Get(name); u.Valid() && !u.IsUndefined() && !u.IsNull() {
			if !u.IsNumber() || u.Float() != math.Trunc(u.Float()) ||
				u.Float() < math.MinInt32 || u.Float() > math.MaxInt32 {
				opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "options.`+name+`" property must be of type int32. Received `+jsTypeName(u))
				return opts
			}
		}
	}
	if d := o.Get("detached"); d.Valid() && !d.IsUndefined() && !d.IsNull() {
		if !d.IsBool() {
			opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "options.detached" property must be of type boolean. Received `+jsTypeName(d))
			return opts
		}
		opts.detached = d.Bool()
	}
	for _, name := range []string{"windowsHide", "windowsVerbatimArguments"} {
		if b := o.Get(name); b.Valid() && !b.IsUndefined() && !b.IsNull() && !b.IsBool() {
			opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "options.`+name+`" property must be of type boolean. Received `+jsTypeName(b))
			return opts
		}
	}
	if n := o.Get("maxBuffer"); n.Valid() && !n.IsUndefined() && !n.IsNull() {
		if !n.IsNumber() || math.IsNaN(n.Float()) || n.Float() < 0 {
			opts.optErr = js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				`The value of "options.maxBuffer" is out of range. It must be a positive number. Received `+numberText(n))
			return opts
		}
		if math.IsInf(n.Float(), 1) {
			opts.maxBuf = 0
		} else {
			opts.maxBuf = n.Int()
		}
	}
	if s := o.Get("killSignal"); s.Valid() && !s.IsUndefined() && !s.IsNull() {
		switch {
		case s.IsString():
			sig := strings.ToUpper(s.Str())
			if !validSignal(sig) {
				opts.optErr = js.ThrowCode("TypeError", "ERR_UNKNOWN_SIGNAL", "Unknown signal: "+sig)
				return opts
			}
			opts.signal = sig
		case s.IsNumber():
			sig, ok := signalByNumber(int(s.Int()))
			if !ok {
				opts.optErr = js.ThrowCode("TypeError", "ERR_UNKNOWN_SIGNAL", "Unknown signal: "+numberText(s))
				return opts
			}
			opts.signal = sig
		default:
			opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "options.killSignal" property must be of type string or number. Received `+jsTypeName(s))
			return opts
		}
	}
	if a := o.Get("argv0"); a.Valid() && !a.IsUndefined() && !a.IsNull() {
		if !a.IsString() {
			opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "options.argv0" property must be of type string. Received `+jsTypeName(a))
			return opts
		}
		opts.argv0 = a.Str()
	}
	if sig := o.Get("signal"); sig.IsObject() {
		opts.signalValue = sig
		if sig.Object().Get("aborted").Bool() {
			opts.aborted = true
		}
	}
	m.parseStdio(o.Get("stdio"), &opts)
	return opts
}

// numberText renders a numeric argument the way Node's error messages do.
func numberText(v js.Value) string {
	if v.IsNumber() {
		f := v.Float()
		if f == math.Trunc(f) && !math.IsInf(f, 0) {
			return strconv.FormatInt(int64(f), 10)
		}
		return strconv.FormatFloat(f, 'g', -1, 64)
	}
	return jsTypeName(v)
}

// jsTypeName renders a value the way Node's invalidArgTypeHelper does:
// 'type number (1)', 'an instance of Array', 'undefined' and so on. The
// wording is observable because most validation tests assert the full message.
func jsTypeName(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsNumber():
		f := v.Float()
		if f == math.Trunc(f) && !math.IsInf(f, 0) {
			return "type number (" + strconv.FormatInt(int64(f), 10) + ")"
		}
		return "type number (" + strconv.FormatFloat(f, 'g', -1, 64) + ")"
	case v.IsString():
		return "type string ('" + v.Str() + "')"
	case v.IsBool():
		if v.Bool() {
			return "type boolean (true)"
		}
		return "type boolean (false)"
	case v.IsFunction():
		if v.IsObject() {
			if name := v.Object().Get("name"); name.IsString() && name.Str() != "" {
				return "function " + name.Str()
			}
		}
		return "function"
	case v.IsArray():
		return "an instance of Array"
	}
	return "an instance of Object"
}

// fileURLPath converts a file: URL object to a filesystem path, matching
// Node's errors for wrong schemes and non-local hosts.
func (m *module) fileURLPath(u js.Object) (string, error) {
	proto := u.Get("protocol")
	if !proto.IsString() || proto.Str() != "file:" {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_URL_SCHEME", "The URL must be of scheme file")
	}
	host := u.Get("host")
	if host.IsString() && host.Str() != "" && host.Str() != "localhost" {
		return "", js.Throw("TypeError", `File URL host must be "localhost" or empty on `+platformName())
	}
	p := u.Get("pathname")
	if !p.IsString() {
		return "", nil
	}
	out := p.Str()
	if decoded, err := urlPathUnescape(out); err == nil {
		out = decoded
	}
	return out, nil
}

func platformName() string {
	if v := os.Getenv("GOOS"); v != "" {
		return v
	}
	return "linux"
}

func (m *module) parseStdio(v js.Value, opts *runOptions) {
	if v.Valid() && v.IsString() {
		mode := v.Str()
		if mode != "pipe" && mode != "overlapped" && mode != "ignore" && mode != "inherit" {
			opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
				"The argument 'stdio' is invalid. Received '"+mode+"'")
			return
		}
		for i := 0; i < 3; i++ {
			opts.inherit[i] = mode == "inherit"
			opts.ignore[i] = mode == "ignore"
		}
		return
	}
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return
	}
	if !v.IsArray() {
		opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			"The argument 'stdio' is invalid. Received "+jsTypeName(v))
		return
	}
	arr := v.Object()
	ipcCount := 0
	for i := 0; i < arr.Length(); i++ {
		e := arr.At(i)
		if e.IsString() {
			mode := e.Str()
			switch mode {
			case "ipc":
				ipcCount++
				if i < 3 {
					opts.ignore[i] = true
				}
				continue
			case "pipe", "overlapped", "ignore", "inherit":
				if i < 3 {
					opts.inherit[i] = mode == "inherit"
					opts.ignore[i] = mode == "ignore"
				}
				continue
			}
			opts.optErr = js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
				"The argument 'stdio' is invalid. Received '"+mode+"'")
			return
		}
		if e.IsNumber() {
			if i < 3 && (e.Int() == 1 || e.Int() == 2) {
				// A parent fd: 1 and 2 mean the child shares that stream.
				opts.inherit[i] = true
			}
			continue
		}
	}
	if ipcCount > 1 {
		opts.optErr = js.ThrowCode("Error", "ERR_IPC_ONE_PIPE", "Found multiple IPC file descriptors")
		return
	}
	// An explicit 'ipc' stdio entry opens a channel for spawn() too, not just
	// fork(); the child sees fd 3/4 exactly as it does for a fork.
	if ipcCount == 1 {
		opts.ipc = true
	}
}

// envFromObject flattens an env object exactly like Node: inherited enumerable
// keys are included, undefined values are skipped and other values stringify.
func (m *module) envFromObject(o js.Object) []string {
	out := []string{}
	for _, k := range m.keysIn(o) {
		v := o.Get(k)
		if v.Valid() && !v.IsUndefined() {
			out = append(out, k+"="+v.Str())
		}
	}
	return out
}

func (m *module) keysIn(o js.Object) []string {
	holder := m.e.JS.Obj()
	holder.SetFunc("keys", func(this js.Object, args []js.Value) (js.Value, error) {
		out := []js.Value{}
		if len(args) > 0 && args[0].IsObject() {
			obj := args[0].Object()
			if keys := obj.OwnEnumerableKeys(); len(keys) > 0 {
				for _, k := range keys {
					out = append(out, m.e.JS.Str(k))
				}
			}
			// Walk the prototype chain for `for..in` semantics.
			p := obj.Proto()
			for i := 0; i < 32 && p.Valid() && p.V().IsObject(); i++ {
				for _, k := range p.OwnEnumerableKeys() {
					dup := false
					for _, seen := range out {
						if seen.Str() == k {
							dup = true
							break
						}
					}
					if !dup {
						out = append(out, m.e.JS.Str(k))
					}
				}
				p = p.Proto()
			}
		}
		return m.e.JS.Arr(out).V(), nil
	})
	v, err := holder.Get("keys").Object().Call(js.Object{}, o.V())
	if err != nil || !v.IsArray() {
		return o.OwnEnumerableKeys()
	}
	arr := v.Object()
	keys := make([]string, 0, arr.Length())
	for i := 0; i < arr.Length(); i++ {
		keys = append(keys, arr.At(i).Str())
	}
	return keys
}

// ------------------------------------------------------------------ process --

// output captures a child's output without truncating it: Node reports the
// full buffer when maxBuffer is exceeded, and truncates only in the callback.
type output struct{ buf bytes.Buffer }

func (o *output) Write(p []byte) (int, error) { return o.buf.Write(p) }

func (o *output) Bytes() []byte { return o.buf.Bytes() }

func (m *module) run(opts runOptions) runResult {
	res := runResult{status: -1}
	if opts.aborted {
		// An already-aborted signal never starts the process.
		res.startErr = errors.New("The operation was aborted")
		res.code = "ABORT_ERR"
		res.message = "The operation was aborted"
		return res
	}
	command, argv := resolveCommand(opts)

	cmd := exec.Command(command, argv...)
	if opts.argv0 != "" && len(cmd.Args) > 0 {
		cmd.Args[0] = opts.argv0
	}
	cmd.Dir = opts.cwd
	if cmd.Dir == "" {
		cmd.Dir = m.e.Cwd()
	}
	if opts.env != nil {
		cmd.Env = opts.env
	} else {
		cmd.Env = m.e.Environ()
	}
	if opts.input != nil {
		cmd.Stdin = bytes.NewReader(opts.input)
	}

	outBuf := &output{}
	errBuf := &output{}
	if opts.inherit[1] {
		cmd.Stdout = m.e.Out()
	} else if !opts.ignore[1] {
		cmd.Stdout = outBuf
	}
	if opts.inherit[2] {
		cmd.Stderr = m.e.Err()
	} else if !opts.ignore[2] {
		cmd.Stderr = errBuf
	}
	if opts.inherit[0] {
		cmd.Stdin = m.e.In()
	}

	var err error
	if opts.timeout > 0 {
		if err = cmd.Start(); err == nil {
			if cmd.Process != nil {
				res.pid = cmd.Process.Pid
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case err = <-done:
			case <-time.After(opts.timeout):
				if cmd.Process != nil {
					_ = cmd.Process.Signal(signalFor(opts.killSignal()))
				}
				err = <-done
				res.timedOut = true
				res.killed = true
				res.signal = opts.killSignal()
			}
		}
	} else {
		err = cmd.Run()
		if cmd.Process != nil {
			res.pid = cmd.Process.Pid
		}
	}
	res.stdout = outBuf.Bytes()
	res.stderr = errBuf.Bytes()

	if !res.timedOut {
		res.status, res.signal = exitStatus(err)
	}
	if opts.maxBuf > 0 {
		switch {
		case len(res.stdout) > opts.maxBuf:
			res.bufferStream = "stdout"
		case len(res.stderr) > opts.maxBuf:
			res.bufferStream = "stderr"
		}
	}

	if err != nil && !res.timedOut {
		if _, ok := err.(*exec.ExitError); !ok {
			res.startErr = err
			res.code = spawnCode(err)
			res.status = -1
			res.signal = ""
		}
	}
	if res.startErr != nil {
		res.errno = errnoFor(res.code)
		// spawnSync reports its own syscall name, unlike the asynchronous
		// spawn() path ("spawn <file>").
		res.syscall = "spawnSync " + command
		res.path = command
		res.spawnargs = argv
		res.message = res.syscall + " " + res.code
	} else {
		res.message = "Command failed: " + commandLine(opts)
		if text := strings.TrimRight(string(res.stderr), "\n"); text != "" {
			res.message += "\n" + text
		}
	}
	if res.bufferStream != "" {
		res.message = res.bufferStream + " maxBuffer length exceeded"
		// libuv reports a full stdio buffer as ENOBUFS; spawnSync callers rely
		// on that errno name rather than the JS-only error code.
		res.code = "ENOBUFS"
		res.errno = -105
		res.status = -1
		res.startErr = errors.New(res.message)
	}
	if res.timedOut && res.startErr == nil {
		// A spawnSync timeout is a libuv ETIMEDOUT, reported through `error`
		// while the process's signal is preserved.
		res.code = "ETIMEDOUT"
		res.errno = -110
		res.syscall = "spawnSync " + command
		res.path = command
		res.spawnargs = argv
		res.message = res.syscall + " " + res.code
		res.startErr = errors.New(res.message)
	}
	return res
}

func signalFor(name string) syscall.Signal {
	switch strings.ToUpper(name) {
	case "SIGABRT", "SIGIOT":
		return syscall.SIGABRT
	case "SIGALRM":
		return syscall.SIGALRM
	case "SIGBUS":
		return syscall.SIGBUS
	case "SIGCHLD":
		return syscall.SIGCHLD
	case "SIGCONT":
		return syscall.SIGCONT
	case "SIGFPE":
		return syscall.SIGFPE
	case "SIGHUP":
		return syscall.SIGHUP
	case "SIGILL":
		return syscall.SIGILL
	case "SIGINT":
		return syscall.SIGINT
	case "SIGIO", "SIGPOLL":
		return syscall.SIGIO
	case "SIGKILL":
		return syscall.SIGKILL
	case "SIGPIPE":
		return syscall.SIGPIPE
	case "SIGPROF":
		return syscall.SIGPROF
	case "SIGPWR":
		return syscall.SIGPWR
	case "SIGQUIT":
		return syscall.SIGQUIT
	case "SIGSEGV":
		return syscall.SIGSEGV
	case "SIGSTKFLT":
		return syscall.SIGSTKFLT
	case "SIGSTOP":
		return syscall.SIGSTOP
	case "SIGSYS":
		return syscall.SIGSYS
	case "SIGTERM":
		return syscall.SIGTERM
	case "SIGTRAP":
		return syscall.SIGTRAP
	case "SIGTSTP":
		return syscall.SIGTSTP
	case "SIGTTIN":
		return syscall.SIGTTIN
	case "SIGTTOU":
		return syscall.SIGTTOU
	case "SIGURG":
		return syscall.SIGURG
	case "SIGUSR1":
		return syscall.SIGUSR1
	case "SIGUSR2":
		return syscall.SIGUSR2
	case "SIGVTALRM":
		return syscall.SIGVTALRM
	case "SIGWINCH":
		return syscall.SIGWINCH
	case "SIGXCPU":
		return syscall.SIGXCPU
	case "SIGXFSZ":
		return syscall.SIGXFSZ
	}
	return syscall.SIGTERM
}

func errnoFor(code string) int {
	switch code {
	case "ENOENT":
		return -2
	case "EACCES":
		return -13
	case "ENFILE":
		return -23
	case "EMFILE":
		return -24
	}
	return 0
}

func spawnCode(err error) string {
	switch {
	case errors.Is(err, exec.ErrNotFound), errors.Is(err, os.ErrNotExist):
		return "ENOENT"
	case errors.Is(err, syscall.EACCES), errors.Is(err, os.ErrPermission):
		return "EACCES"
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case syscall.ENOENT:
			return "ENOENT"
		case syscall.EACCES:
			return "EACCES"
		}
	}
	return ""
}

func exitStatus(err error) (int, string) {
	if err == nil {
		return 0, ""
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			return -1, signalName(ws.Signal())
		}
		return ee.ExitCode(), ""
	}
	return -1, ""
}

func signalName(s syscall.Signal) string {
	switch s {
	case syscall.SIGHUP:
		return "SIGHUP"
	case syscall.SIGINT:
		return "SIGINT"
	case syscall.SIGKILL:
		return "SIGKILL"
	case syscall.SIGTERM:
		return "SIGTERM"
	case syscall.SIGQUIT:
		return "SIGQUIT"
	case syscall.SIGPIPE:
		return "SIGPIPE"
	case syscall.SIGABRT:
		return "SIGABRT"
	case syscall.SIGALRM:
		return "SIGALRM"
	case syscall.SIGCHLD:
		return "SIGCHLD"
	case syscall.SIGSEGV:
		return "SIGSEGV"
	}
	return "SIG" + fmt.Sprint(int(s))
}

// capped truncates output the way Node's exec does. Buffer encodings truncate
// at max bytes; string encodings keep at most max characters.
func capped(b []byte, max int, asString bool) []byte {
	if max <= 0 || len(b) <= max {
		return b
	}
	if asString {
		r := []rune(string(b))
		if len(r) <= max {
			return b
		}
		return []byte(string(r[:max]))
	}
	return b[:max]
}

func (m *module) rangeError(msg, code string) js.Object {
	v, err := m.e.JS.Eval(`new RangeError(`+strconv.Quote(msg)+`)`, "[gode:child_process]")
	if err != nil || !v.IsObject() {
		o := m.e.JS.Error(msg)
		if code != "" {
			o.SetStr("code", code)
		}
		return o
	}
	o := v.Object()
	if code != "" {
		o.SetStr("code", code)
	}
	return o
}

func urlPathUnescape(s string) (string, error) {
	if !strings.Contains(s, "%") {
		return s, nil
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			hi, ok1 := unhex(s[i+1])
			lo, ok2 := unhex(s[i+2])
			if ok1 && ok2 {
				b.WriteByte(hi<<4 | lo)
				i += 2
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String(), nil
}

func unhex(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// ------------------------------------------------------------------ results --

func (m *module) resultObject(r runResult, opts runOptions) js.Object {
	o := m.e.JS.Obj()
	asString := opts.encoding != "" && opts.encoding != "buffer"
	// Sync results report every captured byte: libuv reads in 64k chunks, so
	// the buffer may legitimately be larger than maxBuffer.
	out := r.stdout
	errOut := r.stderr
	o.Set("status", nullableNumber(m.e, r.status))
	o.Set("signal", nullableString(m.e, r.signal))
	o.SetNum("pid", float64(r.pid))
	o.SetStr("cmd", commandLine(opts))
	o.SetBool("killed", r.killed)
	o.Set("stdout", m.bytesValue(out))
	o.Set("stderr", m.bytesValue(errOut))
	o.SetStr("stdoutText", string(out))
	o.SetStr("stderrText", string(errOut))
	o.SetBool("encoding", asString)
	o.SetStr("encodingName", opts.encoding)
	o.SetStr("message", r.message)
	if r.code != "" {
		o.SetStr("code", r.code)
	} else if r.status > 0 {
		o.SetNum("code", float64(r.status))
	} else {
		o.SetStr("code", "")
	}
	o.SetBool("error", r.startErr != nil)
	if r.errno != 0 {
		o.SetNum("errno", float64(r.errno))
	}
	if r.syscall != "" {
		o.SetStr("syscall", r.syscall)
	}
	if r.path != "" {
		o.SetStr("path", r.path)
	}
	if len(r.spawnargs) > 0 {
		o.Set("spawnargs", m.argsArray(r.spawnargs))
	}
	return o
}

func (m *module) spawnSyncObject(r runResult, opts runOptions) js.Object {
	o := m.e.JS.Obj()
	asString := opts.encoding != "" && opts.encoding != "buffer"
	outBytes := r.stdout
	errBytes := r.stderr
	out := m.bytesValue(outBytes)
	errOut := m.bytesValue(errBytes)
	if asString {
		out = m.decodeBytes(outBytes, opts.encoding)
		errOut = m.decodeBytes(errBytes, opts.encoding)
	}
	// Inherited streams are not captured, so Node reports them as null.
	if opts.inherit[1] {
		out = m.e.JS.Null()
	}
	if opts.inherit[2] {
		errOut = m.e.JS.Null()
	}
	o.SetNum("pid", float64(r.pid))
	o.Set("output", m.e.JS.Arr([]js.Value{m.e.JS.Null(), out, errOut}).V())
	o.Set("stdout", out)
	o.Set("stderr", errOut)
	o.Set("status", nullableNumber(m.e, r.status))
	o.Set("signal", nullableString(m.e, r.signal))
	if r.bufferStream != "" {
		eobj := m.e.JS.Error(r.message)
		eobj.SetStr("code", "ENOBUFS")
		eobj.SetNum("errno", float64(-105))
		eobj.Set("status", nullableNumber(m.e, r.status))
		eobj.Set("signal", nullableString(m.e, r.signal))
		eobj.Set("stdout", out)
		eobj.Set("stderr", errOut)
		o.Set("error", eobj.V())
	} else if r.startErr != nil {
		eobj := m.e.JS.Error(r.message)
		if r.code == "ABORT_ERR" {
			eobj.SetStr("name", "AbortError")
			eobj.SetStr("code", "ABORT_ERR")
		} else if r.code != "" {
			eobj.SetStr("code", r.code)
			eobj.SetNum("errno", float64(r.errno))
			eobj.SetStr("syscall", r.syscall)
			eobj.SetStr("path", r.path)
			args := m.e.JS.Arr(nil)
			for i, a := range r.spawnargs {
				args.SetAt(i, m.e.JS.Str(a))
			}
			eobj.Set("spawnargs", args.V())
		}
		eobj.Set("status", nullableNumber(m.e, r.status))
		eobj.Set("signal", nullableString(m.e, r.signal))
		eobj.Set("stdout", out)
		eobj.Set("stderr", errOut)
		o.Set("error", eobj.V())
	} else {
		o.Set("error", m.e.JS.Undefined())
	}
	if r.startErr != nil && !r.timedOut && r.bufferStream == "" {
		// A process that never started produced no output at all; a process
		// that ran (timeout or maxBuffer) still reports what it wrote.
		z := m.e.JS.Null()
		o.Set("output", m.e.JS.Arr([]js.Value{z, z, z}).V())
		o.Set("stdout", z)
		o.Set("stderr", z)
	}
	return o
}

func (m *module) errorValue(r runResult, opts runOptions) js.Object {
	var eobj js.Object
	if r.bufferStream != "" {
		eobj = m.rangeError(r.message, "ERR_CHILD_PROCESS_STDIO_MAXBUFFER")
	} else {
		eobj = m.e.JS.Error(r.message)
	}
	if r.code != "" {
		eobj.SetStr("code", r.code)
	}
	if r.startErr != nil {
		eobj.SetNum("errno", float64(r.errno))
		eobj.SetStr("syscall", r.syscall)
		eobj.SetStr("path", r.path)
		args := m.e.JS.Arr(nil)
		for i, a := range r.spawnargs {
			args.SetAt(i, m.e.JS.Str(a))
		}
		eobj.Set("spawnargs", args.V())
	}
	if r.killed {
		eobj.SetBool("killed", true)
	}
	eobj.Set("status", nullableNumber(m.e, r.status))
	eobj.Set("signal", nullableString(m.e, r.signal))
	eobj.Set("stdout", m.bytesValue(r.stdout))
	eobj.Set("stderr", m.bytesValue(r.stderr))
	eobj.SetNum("pid", float64(r.pid))
	eobj.SetStr("cmd", commandLine(opts))
	return eobj
}

func commandLine(opts runOptions) string {
	if len(opts.args) == 0 {
		return opts.cmd
	}
	return opts.cmd + " " + strings.Join(opts.args, " ")
}

// resolveCommand applies the shell wrapping, returning the executable and argv
// the kernel actually sees. Node reports these as spawnfile/spawnargs, so the
// shell invocation (not the user's command line) is what errors name.
func resolveCommand(opts runOptions) (string, []string) {
	if !opts.shell {
		return opts.cmd, opts.args
	}
	shell := opts.shellArg
	if shell == "" {
		shell = "/bin/sh"
	}
	line := opts.cmd
	if len(opts.args) > 0 {
		line += " " + strings.Join(opts.args, " ")
	}
	return shell, []string{"-c", line}
}

func nullableNumber(e *env.Env, n int) js.Value {
	if n < 0 {
		return e.JS.Null()
	}
	return e.JS.Num(float64(n))
}

func nullableString(e *env.Env, s string) js.Value {
	if s == "" {
		return e.JS.Null()
	}
	return e.JS.Str(s)
}

func (m *module) bytesValue(b []byte) js.Value {
	ctor := m.e.Global().Get("Buffer")
	if ctor.Valid() && ctor.IsFunction() {
		from := ctor.Object().Get("from")
		if from.IsFunction() {
			if v, err := from.Object().Call(ctor.Object(), m.e.JS.Bytes(b).V()); err == nil {
				return v
			}
		}
	}
	return m.e.JS.Bytes(b).V()
}

// ------------------------------------------------------------------- spawn --

func (m *module) spawnChild(opts runOptions, callback js.Value, encoding string, async bool) js.Object {
	return m.spawnChildWith(js.Object{}, opts, callback, encoding, async)
}

// spawnChildWith starts a child and, when cp is already a live object, adopts
// it: ChildProcess#spawn must fill in the instance the caller constructed
// instead of returning a different one.
func (m *module) spawnChildWith(cp js.Object, opts runOptions, callback js.Value, encoding string, async bool) js.Object {
	if !cp.Valid() {
		cp = m.e.JS.Obj()
	}
	cp.SetProto(m.cpProto)

	m.seq++
	id := m.seq
	st := &procState{
		id: id, cp: cp, status: -1, encoding: encoding, cmdline: commandLine(opts),
		maxBuf: opts.maxBuf, signalValue: opts.signalValue,
	}
	// Go holds the process object and the exec callback across many event loop
	// turns while the child runs. Pin them: JavaScriptCore otherwise collects
	// the wrapper once user code drops it, and the next listener call crashes.
	cp.V().Protect()
	if callback.Valid() && callback.IsFunction() {
		callback.Protect()
	}
	cp.SetHidden("$godeCpId", m.e.JS.Num(float64(id)))

	cmd, argv := resolveCommand(opts)
	c := exec.Command(cmd, argv...)
	if opts.argv0 != "" && len(c.Args) > 0 {
		c.Args[0] = opts.argv0
	}
	c.Dir = opts.cwd
	if c.Dir == "" {
		c.Dir = m.e.Cwd()
	}
	if opts.env != nil {
		c.Env = opts.env
	} else {
		c.Env = m.e.Environ()
	}
	if opts.detached {
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	st.async = &asyncState{cmd: c, out: &lockedBuf{}, err: &lockedBuf{}, done: make(chan struct{})}

	// stdin is a real OS pipe unless the caller inherited or ignored it. A
	// generic io.Reader would make os/exec copy in a goroutine that blocks
	// Wait() until the write end closes: a child that never reads stdin would
	// deadlock the parent.
	if opts.inherit[0] {
		c.Stdin = m.e.In()
	} else if !opts.ignore[0] {
		if pr, pw, err := os.Pipe(); err == nil {
			c.Stdin = pr
			st.async.stdinW = pw
			st.async.stdinR = pr
		}
	}

	// The IPC channel rides on inherited fd 3 (child reads) and fd 4 (child
	// writes), so the child never needs to negotiate anything.
	var ipcChildFiles []*os.File
	if opts.ipc {
		if pw, childFiles, pr, err := makeIPCPipes(); err == nil {
			c.ExtraFiles = childFiles
			ipcChildFiles = childFiles
			st.async.ipc = &ipcChannel{w: pw, r: pr}
			c.Env = append(stripIPCEnv(c.Env),
				ipcRecvEnv+"=3", ipcSendEnv+"=4", "NODE_CHANNEL_FD=3")
		}
	}
	if opts.inherit[1] {
		c.Stdout = m.e.Out()
	} else if !opts.ignore[1] {
		c.Stdout = st.async.out
	}
	if opts.inherit[2] {
		c.Stderr = m.e.Err()
	} else if !opts.ignore[2] {
		c.Stderr = st.async.err
	}

	if err := c.Start(); err != nil {
		code := spawnCode(err)
		res := runResult{
			status: -1, code: code, startErr: err, errno: errnoFor(code),
			syscall: "spawn " + cmd, path: cmd, spawnargs: argv,
			message: "spawn " + cmd + " " + code, pid: 0,
		}
		if code == "" {
			res.message = err.Error()
		}
		if opts.env == nil && opts.cwd != "" {
			// A missing cwd surfaces as ENOENT on the shell, exactly as in
			// Node, but callers expect path/syscall to name what failed.
			if _, statErr := os.Stat(opts.cwd); statErr != nil {
				res.code = "ENOENT"
				res.errno = errnoFor(res.code)
				res.message = "spawn " + cmd + " ENOENT"
			}
		}
		m.closeSpawnPipes(st, ipcChildFiles)
		st.async.ipc = nil
		o := m.errorValue(res, opts)
		st.spawnErr = o
		st.spawnErrFlag = true
		st.callback = callback
		cp.SetStr("spawnfile", cmd)
		cp.Set("spawnargs", m.argsArray(argv))
		m.attachStreams(st, opts, nil, nil)
		cp.SetBool("connected", false)
		cp.Set("channel", m.e.JS.Undefined())
		cp.SetBool("killed", false)
		cp.Set("exitCode", m.e.JS.Null())
		cp.Set("signalCode", m.e.JS.Null())
		if callback.Valid() && callback.IsFunction() {
			m.addCloseCallback(st)
		}
		m.children[id] = st
		fail := m.e.JS.Obj()
		fail.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
			m.queueFail(st, o)
			return m.e.JS.Undefined(), nil
		})
		m.e.NextTick(fail.Get("f"))
		return cp
	}

	m.children[id] = st
	st.async.cmd = c
	// The parent must drop its copies of the child's ends or the pipe never
	// reaches EOF.
	if st.async.stdinR != nil {
		_ = st.async.stdinR.Close()
		st.async.stdinR = nil
	}
	for _, f := range ipcChildFiles {
		_ = f.Close()
	}
	if c.Process != nil {
		st.pid = c.Process.Pid
		cp.SetNum("pid", float64(st.pid))
	}
	cp.SetStr("spawnfile", cmd)
	cp.Set("spawnargs", m.argsArray(argv))
	if st.async.ipc != nil {
		st.connected = true
		cp.SetBool("connected", true)
		ch := m.channelObject()
		ch.SetHidden("$godeCpId", m.e.JS.Num(float64(id)))
		cp.Set("channel", ch.V())
		// Node exposes the same object as child._channel; official tests use
		// it to assert that the public getter is stable.
		cp.SetHidden("_channel", ch.V())
		readIPCLines(st.async.ipc.r, func(line string) {
			st.async.mu.Lock()
			st.async.inbox = append(st.async.inbox, line)
			st.async.mu.Unlock()
		}, func() {
			st.async.mu.Lock()
			st.async.ipcDone = true
			st.async.mu.Unlock()
		})
	} else {
		cp.SetBool("connected", false)
		cp.Set("channel", m.e.JS.Undefined())
	}
	cp.SetBool("killed", false)
	cp.Set("exitCode", m.e.JS.Null())
	cp.Set("signalCode", m.e.JS.Null())
	m.attachStreams(st, opts, nil, nil)
	if st.async.ipc != nil {
		// Node exposes the channel as stdio[3] for a forked child.
		cp.Set("stdio", m.e.JS.Arr([]js.Value{
			cp.Get("stdin"), cp.Get("stdout"), cp.Get("stderr"), cp.Get("channel"),
		}).V())
	}

	st.callback = callback
	if callback.Valid() && callback.IsFunction() {
		m.addCloseCallback(st)
	}

	go func() {
		// A misbehaving child (or a wait raced by process teardown) must never
		// take the whole runtime down with a Go panic: report it as the exit
		// status instead.
		defer func() {
			if r := recover(); r != nil {
				st.async.mu.Lock()
				st.async.wait = fmt.Errorf("wait: %v", r)
				st.async.mu.Unlock()
			}
			close(st.async.done)
		}()
		err := c.Wait()
		st.async.mu.Lock()
		st.async.wait = err
		st.async.mu.Unlock()
	}()

	// The pump keeps the loop alive while the child runs and emits the
	// completion events from the JavaScript goroutine.
	pump := m.e.JS.Obj()
	pump.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
		m.poll(st)
		return m.e.JS.Undefined(), nil
	})
	st.async.timer = m.e.SetTimeout(pump.Get("f"), 2, true, nil)

	if opts.timeout > 0 {
		kill := m.e.JS.Obj()
		kill.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
			if st.finished || st.timedOut {
				return m.e.JS.Undefined(), nil
			}
			st.timedOut = true
			m.signalChild(st, opts.killSignal())
			return m.e.JS.Undefined(), nil
		})
		st.killTimer = m.e.SetTimeout(kill.Get("f"), opts.timeout.Seconds()*1000, false, nil)
	}

	// An already-aborted signal still starts the process (Node kills it
	// immediately), so 'exit' reports the kill signal and exec callbacks get
	// AbortError instead of a never-spawned child.
	if opts.aborted {
		holder := m.e.JS.Obj()
		holder.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
			if st.finished {
				return m.e.JS.Undefined(), nil
			}
			st.aborted = true
			m.signalChild(st, opts.killSignal())
			if callback.Valid() && callback.IsFunction() {
				return m.e.JS.Undefined(), nil
			}
			err := m.abortError()
			if opts.signalValue.Valid() && opts.signalValue.IsObject() {
				if reason := opts.signalValue.Object().Get("reason"); reason.Valid() && !reason.IsUndefined() {
					err.Set("cause", reason)
				}
			}
			m.emitValue(cp, "error", []js.Value{err.V()})
			return m.e.JS.Undefined(), nil
		})
		m.e.NextTick(holder.Get("f"))
	}

	// 'spawn' fires once the process exists and before any output.
	spawn := m.e.JS.Obj()
	spawn.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
		m.emitValue(cp, "spawn", nil)
		return m.e.JS.Undefined(), nil
	})
	m.e.NextTick(spawn.Get("f"))
	return cp
}

func (m *module) argsArray(args []string) js.Value {
	arr := m.e.JS.Arr(nil)
	for i, a := range args {
		arr.SetAt(i, m.e.JS.Str(a))
	}
	return arr.V()
}

// attachStreams installs the stdin/stdout/stderr surface for a child: real
// nulls when inherited or ignored, emitters otherwise.
func (m *module) attachStreams(st *procState, opts runOptions, out, errOut []byte) {
	cp := st.cp
	stdin := m.e.JS.Null()
	stdout := m.e.JS.Null()
	stderr := m.e.JS.Null()
	if !opts.ignore[0] && !opts.inherit[0] {
		stdin = m.newStream("stdin", st.id, nil, "")
		st.stdin = stdin.Object()
	}
	if !opts.ignore[1] && !opts.inherit[1] {
		// A plain spawn leaves the stream in Buffer mode; exec/execFile apply
		// their encoding to stdout/stderr exactly like setEncoding() does.
		stdout = m.newStream("stdout", st.id, out, st.encoding)
		st.stdout = stdout.Object()
	}
	if !opts.ignore[2] && !opts.inherit[2] {
		stderr = m.newStream("stderr", st.id, errOut, st.encoding)
		st.stderr = stderr.Object()
	}
	cp.Set("stdin", stdin)
	cp.Set("stdout", stdout)
	cp.Set("stderr", stderr)
	cp.Set("stdio", m.e.JS.Arr([]js.Value{stdin, stdout, stderr}).V())
}

// queueFail reports a spawn failure: 'error', the exec callback, then 'close'.
func (m *module) queueFail(st *procState, spawnErr js.Object) {
	if st.finished {
		return
	}
	st.finished = true
	if eobj := spawnErr; eobj.Valid() {
		hasListener := len(m.listeners(st.cp, "error")) > 0
		hasCallback := st.callback.Valid() && st.callback.IsFunction()
		if hasListener {
			m.emitValue(st.cp, "error", []js.Value{eobj.V()})
		}
		if hasCallback && !st.callbackDone {
			// exec/execFile report a failed spawn through their callback (Node
			// attaches an internal error listener), so it must not go uncaught.
			st.callbackDone = true
			out, errOut := m.callbackOutput(st)
			m.e.NextTick(st.callback, eobj.V(), out, errOut)
		} else if !hasListener {
			m.throwValue(eobj.V())
		}
	}
	m.emitValue(st.cp, "close", []js.Value{m.e.JS.Null(), m.e.JS.Null()})
}

// throwValue raises a JavaScript value as an uncaught exception.
func (m *module) throwValue(v js.Value) {
	if m.thrower.Valid() {
		if _, err := m.thrower.Object().Call(js.Object{}, v); err != nil {
			panic(err)
		}
	}
	panic(js.Throw("Error", v.Object().Get("message").Str()))
}

// signalChild delivers sig to the live process and records the kill.
func (m *module) signalChild(st *procState, sig string) bool {
	if st.async == nil || st.async.cmd == nil || st.async.cmd.Process == nil || st.finished {
		return false
	}
	if err := st.async.cmd.Process.Signal(signalFor(sig)); err != nil {
		return false
	}
	st.killed = true
	st.cp.SetBool("killed", true)
	st.cp.SetHidden("$godeKillSignal", m.e.JS.Str(sig))
	return true
}

// poll is the pump: it checks whether the child has exited and, if so, emits
// the buffered output followed by 'exit' and 'close'.
func (m *module) poll(st *procState) {
	if st.finished || st.async == nil {
		return
	}
	// Messages arrive on the reader goroutine; hand them to JS here, on the
	// only goroutine allowed to touch JavaScriptCore.
	st.async.mu.Lock()
	live := st.async.inbox
	st.async.inbox = nil
	st.async.mu.Unlock()
	for _, line := range live {
		m.emitIPCFrame(st, line)
	}
	select {
	case <-st.async.done:
	default:
		return
	}
	if st.async.timer != 0 {
		m.e.ClearTimer(st.async.timer)
		st.async.timer = 0
	}
	if st.killTimer != 0 {
		m.e.ClearTimer(st.killTimer)
		st.killTimer = 0
	}
	st.finished = true
	if st.async.stdinW != nil {
		_ = st.async.stdinW.Close()
	}
	if st.timedOut {
		st.killed = true
		st.cp.SetBool("killed", true)
	}
	status, sig := exitStatus(st.async.wait)
	st.exitCode, st.signalCode = status, sig
	st.status, st.signal = status, sig
	st.outBytes = st.async.out.Bytes()
	st.errBytes = st.async.err.Bytes()
	if st.maxBuf > 0 {
		switch {
		case len(st.outBytes) > st.maxBuf:
			st.bufferStream = "stdout"
		case len(st.errBytes) > st.maxBuf:
			st.bufferStream = "stderr"
		}
	}
	// Deliver frames that raced with the exit before announcing it, then drop
	// the channel: Node emits 'disconnect' ahead of 'exit'.
	st.async.mu.Lock()
	pending := st.async.inbox
	st.async.inbox = nil
	st.async.mu.Unlock()
	for _, line := range pending {
		m.emitIPCFrame(st, line)
	}
	if st.connected {
		m.closeChildIPC(st)
	} else if st.async.ipc != nil {
		st.async.ipc.closeAll()
		st.async.ipc = nil
	}
	if st.stdout.Valid() && !st.stdout.IsNull() {
		m.emitStream(st.stdout, m.streamOutput(st, st.outBytes))
	}
	if st.stderr.Valid() && !st.stderr.IsNull() {
		m.emitStream(st.stderr, m.streamOutput(st, st.errBytes))
	}
	st.cp.Set("exitCode", nullableNumber(m.e, status))
	st.cp.Set("signalCode", nullableString(m.e, sig))
	m.emitValue(st.cp, "exit", []js.Value{nullableNumber(m.e, status), nullableString(m.e, sig)})
	m.emitValue(st.cp, "close", []js.Value{nullableNumber(m.e, status), nullableString(m.e, sig)})
}

// closeSpawnPipes releases every pipe created for a child that never started.
func (m *module) closeSpawnPipes(st *procState, childFiles []*os.File) {
	for _, f := range childFiles {
		_ = f.Close()
	}
	if st.async == nil {
		return
	}
	if st.async.stdinR != nil {
		_ = st.async.stdinR.Close()
		st.async.stdinR = nil
	}
	if st.async.stdinW != nil {
		_ = st.async.stdinW.Close()
		st.async.stdinW = nil
	}
	if st.async.ipc != nil {
		st.async.ipc.closeAll()
	}
}

// channelObject is the parent's view of the IPC channel: enough surface for
// `child.channel.ref()` and `close()`, which tests and libraries touch.
func (m *module) channelObject() js.Object {
	o := m.e.JS.Obj()
	o.SetFunc("ref", func(this js.Object, args []js.Value) (js.Value, error) { return this.V(), nil })
	o.SetFunc("unref", func(this js.Object, args []js.Value) (js.Value, error) { return this.V(), nil })
	o.SetFunc("close", func(this js.Object, args []js.Value) (js.Value, error) {
		id := this.Get("$godeCpId")
		if id.Valid() {
			if st, ok := m.children[id.Int()]; ok {
				m.closeChildIPC(st)
			}
		}
		return m.e.JS.Undefined(), nil
	})
	o.Set("_handle", m.e.JS.Null())
	return o
}

// emitIPCFrame parses one JSON frame and emits it as a 'message' event.
func (m *module) emitIPCFrame(st *procState, line string) {
	v, err := m.e.JS.Eval(`JSON.parse(`+strconv.Quote(line)+`)`, "[gode:ipc]")
	if err != nil {
		return
	}
	m.emitValue(st.cp, "message", []js.Value{v})
}

// closeChildIPC tears the channel down and announces it, matching Node's order
// ('disconnect' before 'exit'/'close').
func (m *module) closeChildIPC(st *procState) {
	if !st.connected {
		return
	}
	st.connected = false
	st.cp.SetBool("connected", false)
	if st.async != nil && st.async.ipc != nil {
		st.async.ipc.closeAll()
		st.async.ipc = nil
	}
	m.emitValue(st.cp, "disconnect", nil)
}

// abortedChild represents a process that was never spawned because its
// AbortSignal was already aborted, mirroring Node's exec/execFile behavior.
func (m *module) abortedChild(opts runOptions, callback js.Value) js.Object {
	cp := m.e.JS.Obj()
	cp.SetProto(m.cpProto)
	cp.SetStr("spawnfile", opts.cmd)
	cp.Set("spawnargs", m.e.JS.Arr(nil).V())
	cp.SetBool("connected", false)
	cp.Set("channel", m.e.JS.Undefined())
	cp.SetBool("killed", false)
	cp.Set("exitCode", m.e.JS.Null())
	cp.Set("signalCode", m.e.JS.Null())
	cp.Set("stdin", m.e.JS.Null())
	cp.Set("stdout", m.e.JS.Null())
	cp.Set("stderr", m.e.JS.Null())
	cp.Set("stdio", m.e.JS.Arr([]js.Value{m.e.JS.Null(), m.e.JS.Null(), m.e.JS.Null()}).V())
	err := m.abortError()
	if callback.Valid() && callback.IsFunction() {
		holder := m.e.JS.Obj()
		holder.SetFunc("f", func(this js.Object, args []js.Value) (js.Value, error) {
			if _, err := callback.Object().Call(js.Object{}, err.V(), m.e.JS.Undefined(), m.e.JS.Undefined()); err != nil {
				panic(err)
			}
			return m.e.JS.Undefined(), nil
		})
		m.e.NextTick(holder.Get("f"))
	}
	return cp
}

func (m *module) abortError() js.Object {
	v, err := m.e.JS.Eval(`(function () {
  var e = new Error('The operation was aborted');
  e.name = 'AbortError';
  e.code = 'ABORT_ERR';
  return e;
})()`, "[gode:child_process]")
	if err == nil && v.IsObject() {
		return v.Object()
	}
	o := m.e.JS.Error("The operation was aborted")
	o.SetStr("name", "AbortError")
	o.SetStr("code", "ABORT_ERR")
	return o
}

// addCloseCallback wires an exec/execFile callback to the child's close event,
// so a user-emitted close still resolves the callback like Node.
func (m *module) addCloseCallback(st *procState) {
	holder := m.e.JS.Obj()
	holder.SetFunc("f", func(this js.Object, args []js.Value) (js.Value, error) {
		m.invokeCallback(st, args)
		return m.e.JS.Undefined(), nil
	})
	rec := m.e.JS.Obj()
	rec.SetHidden("fn", holder.Get("f"))
	rec.SetHidden("original", holder.Get("f"))
	rec.SetBool("once", true)
	m.seq++
	rec.SetHidden("id", m.e.JS.Num(float64(m.seq)))
	if !st.cp.Get(listenerKey).IsObject() {
		st.cp.SetHidden(listenerKey, m.e.JS.Obj().V())
	}
	list := m.listeners(st.cp, "close")
	list = append(list, rec)
	m.storeListeners(st.cp, "close", list)
}

func (m *module) invokeCallback(st *procState, args []js.Value) {
	if st.callbackDone || !st.callback.Valid() || !st.callback.IsFunction() {
		return
	}
	st.callbackDone = true
	// The close listener can run before poll() recorded an overflow (a user
	// emitted close, or the spawn-failure path); recompute from the buffered
	// bytes so exec still reports ERR_CHILD_PROCESS_STDIO_MAXBUFFER.
	if st.bufferStream == "" && st.maxBuf > 0 {
		switch {
		case len(st.outBytes) > st.maxBuf:
			st.bufferStream = "stdout"
		case len(st.errBytes) > st.maxBuf:
			st.bufferStream = "stderr"
		}
	}
	st.killed = st.cp.Get("killed").Bool()
	if len(args) > 0 {
		if args[0].IsNumber() {
			st.status = args[0].Int()
		} else {
			st.status = -1
		}
	}
	if len(args) > 1 {
		if args[1].IsString() {
			st.signal = args[1].Str()
		} else {
			st.signal = ""
		}
	}
	out, errOut := m.callbackOutput(st)
	errVal := m.e.JS.Null()
	if st.aborted || (st.signalValue.Valid() && st.signalValue.IsObject() &&
		st.signalValue.Object().Get("aborted").Bool()) {
		eobj := m.abortError()
		eobj.SetStr("cmd", st.cmdline)
		eobj.SetBool("killed", st.killed)
		if st.signalValue.Valid() && st.signalValue.IsObject() {
			if reason := st.signalValue.Object().Get("reason"); reason.Valid() && !reason.IsUndefined() {
				eobj.Set("cause", reason)
			}
		}
		errVal = eobj.V()
		out = m.e.JS.Undefined()
		errOut = m.e.JS.Undefined()
	} else if st.bufferStream != "" {
		errVal = m.errorValue(runResult{
			status: st.status, signal: st.signal, stdout: st.outBytes, stderr: st.errBytes,
			message: st.bufferStream + " maxBuffer length exceeded", bufferStream: st.bufferStream,
			killed: st.killed,
		}, runOptions{cmd: st.cmdline, maxBuf: st.maxBuf}).V()
	} else if st.status != 0 || st.signal != "" || st.killed {
		msg := "Command failed: " + st.cmdline
		if text := strings.TrimRight(string(st.errBytes), "\n"); text != "" {
			msg += "\n" + text
		}
		errVal = m.callbackError(st, msg).V()
	}
	if _, err := st.callback.Object().Call(js.Object{}, errVal, out, errOut); err != nil {
		panic(err)
	}
}

func (m *module) callbackError(st *procState, msg string) js.Object {
	eobj := m.e.JS.Error(msg)
	switch {
	case st.code != "":
		eobj.SetStr("code", st.code)
	case st.signal != "":
		eobj.Set("code", m.e.JS.Null())
	case st.status > 0:
		eobj.SetNum("code", float64(st.status))
	case st.status < 0:
		eobj.SetStr("code", "EPERM")
	default:
		eobj.Set("code", m.e.JS.Null())
	}
	eobj.SetBool("killed", st.killed)
	eobj.Set("signal", nullableString(m.e, st.signal))
	eobj.SetStr("cmd", st.cmdline)
	out, errOut := m.callbackOutput(st)
	eobj.Set("stdout", out)
	eobj.Set("stderr", errOut)
	return eobj
}

func (m *module) finish(st *procState) {
	if st.spawnErr.Valid() {
		if len(m.listeners(st.cp, "error")) > 0 {
			m.emitValue(st.cp, "error", []js.Value{st.spawnErr.V()})
		}
		if st.callback.Valid() && st.callback.IsFunction() && !st.callbackDone {
			st.callbackDone = true
			out, errOut := m.callbackOutput(st)
			m.e.NextTick(st.callback, st.spawnErr.V(), out, errOut)
		}
		if len(m.listeners(st.cp, "close")) > 0 {
			m.emitValue(st.cp, "close", []js.Value{m.e.JS.Null(), m.e.JS.Null()})
		}
		return
	}
	if st.stdout.Valid() && !st.stdout.IsNull() {
		m.emitStream(st.stdout, m.streamOutput(st, st.outBytes))
	}
	if st.stderr.Valid() && !st.stderr.IsNull() {
		m.emitStream(st.stderr, m.streamOutput(st, st.errBytes))
	}
	m.emitValue(st.cp, "exit", []js.Value{nullableNumber(m.e, st.status), nullableString(m.e, st.signal)})
	m.emitValue(st.cp, "close", []js.Value{nullableNumber(m.e, st.status), nullableString(m.e, st.signal)})
}

// streamOutput applies maxBuffer truncation for the 'data' event payload.
func (m *module) streamOutput(st *procState, b []byte) []byte {
	return capped(b, st.maxBuf, st.encoding != "" && st.encoding != "buffer")
}

func (m *module) callbackOutput(st *procState) (js.Value, js.Value) {
	asString := st.encoding != "" && st.encoding != "buffer"
	out := capped(st.outBytes, st.maxBuf, asString)
	errOut := capped(st.errBytes, st.maxBuf, asString)
	if asString {
		return m.decodeBytes(out, st.encoding), m.decodeBytes(errOut, st.encoding)
	}
	return m.bytesValue(out), m.bytesValue(errOut)
}

// decodeBytes turns raw output into the representation Node uses for encoding:
// a Buffer when no encoding was requested, otherwise the decoded string.
func (m *module) decodeBytes(b []byte, enc string) js.Value {
	if enc == "" || enc == "buffer" {
		return m.bytesValue(b)
	}
	if enc == "utf8" {
		return m.e.JS.Str(string(b))
	}
	buf := m.bytesValue(b)
	if fn := buf.Object().Get("toString"); fn.IsFunction() {
		if v, err := fn.Object().Call(buf.Object(), m.e.JS.Str(enc)); err == nil {
			return v
		}
	}
	return m.e.JS.Str(string(b))
}

func (m *module) emitStream(stream js.Object, data []byte) {
	enc := stream.Get("$godeEncoding").Str()
	flowing := stream.Get("readableFlowing").Bool() ||
		len(m.listeners(stream, "data")) > 0 || len(m.pipes(stream)) > 0
	if len(data) > 0 {
		payload := m.decodeBytes(data, enc)
		if flowing {
			stream.SetBool("readableFlowing", true)
			m.emitValue(stream, "data", []js.Value{payload})
			for _, dest := range m.pipes(stream) {
				if fn := dest.Get("write"); fn.IsFunction() {
					if _, err := fn.Object().Call(dest, payload); err != nil {
						panic(err)
					}
				}
			}
		} else {
			// Nobody is flowing: buffer the bytes for read() and announce the
			// readable event on the next tick, matching Node's paused mode.
			buf := data
			if prev, ok := stream.Get("$godeData").Bytes(); ok && len(prev) > 0 {
				buf = append(append([]byte(nil), prev...), data...)
			}
			stream.Set("$godeData", m.bytesValue(buf))
			stream.SetNum("_bytes", float64(len(buf)))
			holder := m.e.JS.Obj()
			holder.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
				if !stream.Get("destroyed").Bool() {
					m.emitValue(stream, "readable", nil)
				}
				return m.e.JS.Undefined(), nil
			})
			m.e.NextTick(holder.Get("f"))
		}
	}
	m.emitValue(stream, "end", nil)
	for _, dest := range m.pipes(stream) {
		if fn := dest.Get("end"); fn.IsFunction() {
			if _, err := fn.Object().Call(dest); err != nil {
				panic(err)
			}
		}
	}
	m.emitValue(stream, "close", nil)
}

func (m *module) pipes(stream js.Object) []js.Object {
	v := stream.Get("$godePipes")
	if !v.IsArray() {
		return nil
	}
	arr := v.Object()
	out := make([]js.Object, 0, arr.Length())
	for i := 0; i < arr.Length(); i++ {
		if arr.At(i).IsObject() {
			out = append(out, arr.At(i).Object())
		}
	}
	return out
}

// ------------------------------------------------------------------ emitters --

const listenerKey = "$godeListeners"

func (m *module) listeners(o js.Object, name string) []js.Object {
	st := o.Get(listenerKey)
	if !st.IsObject() {
		return nil
	}
	v := st.Object().Get(name)
	if !v.IsArray() {
		return nil
	}
	arr := v.Object()
	out := make([]js.Object, 0, arr.Length())
	for i := 0; i < arr.Length(); i++ {
		if arr.At(i).IsObject() {
			out = append(out, arr.At(i).Object())
		}
	}
	return out
}

func (m *module) addListener(o js.Object, args []js.Value, once bool) (js.Value, error) {
	if len(args) < 2 || !args[1].IsFunction() {
		return js.Value{}, js.Throw("TypeError", `The "listener" argument must be of type function. Received undefined`)
	}
	name := args[0].Str()
	st := o.Get(listenerKey)
	if !st.IsObject() {
		st = m.e.JS.Obj().V()
		o.SetHidden(listenerKey, st)
	}
	rec := m.e.JS.Obj()
	rec.SetHidden("fn", args[1])
	rec.SetHidden("original", args[1])
	rec.SetBool("once", once)
	m.seq++
	rec.SetHidden("id", m.e.JS.Num(float64(m.seq)))

	list := m.listeners(o, name)
	list = append(list, rec)
	m.storeListeners(o, name, list)
	return o.V(), nil
}

func (m *module) storeListeners(o js.Object, name string, recs []js.Object) {
	st := o.Get(listenerKey).Object()
	arr := m.e.JS.Arr(nil)
	for i, rec := range recs {
		arr.SetAt(i, rec.V())
	}
	st.Set(name, arr.V())
}

// removeAllListeners drops every listener (or just one event's listeners).
func (m *module) removeAllListeners(o js.Object, args []js.Value) (js.Value, error) {
	st := o.Get(listenerKey)
	if !st.IsObject() {
		return o.V(), nil
	}
	store := st.Object()
	if len(args) > 0 && args[0].Valid() && !args[0].IsUndefined() {
		store.Set(args[0].Str(), m.e.JS.Arr(nil).V())
		return o.V(), nil
	}
	o.SetHidden(listenerKey, m.e.JS.Obj().V())
	return o.V(), nil
}

func (m *module) removeListener(o js.Object, args []js.Value) (js.Value, error) {
	if len(args) < 2 {
		return o.V(), nil
	}
	name := args[0].Str()
	target := args[1]
	recs := m.listeners(o, name)
	for i := len(recs) - 1; i >= 0; i-- {
		if sameValue(m.e, recs[i].Get("original"), target) || sameValue(m.e, recs[i].Get("fn"), target) {
			recs = append(recs[:i], recs[i+1:]...)
			m.storeListeners(o, name, recs)
			return o.V(), nil
		}
	}
	return o.V(), nil
}

func sameValue(e *env.Env, a, b js.Value) bool {
	if !a.Valid() || !b.Valid() {
		return false
	}
	v, err := e.JS.Eval(`(function (a, b) { return a === b; })`, "eq.js")
	if err != nil {
		return false
	}
	r, err := v.Object().Call(js.Object{}, a, b)
	return err == nil && r.Bool()
}

func (m *module) emitValue(o js.Object, name string, payload []js.Value) bool {
	recs := m.listeners(o, name)
	if len(recs) == 0 {
		if name == "error" && len(payload) > 0 && payload[0].IsObject() {
			// Match Node: an unhandled error event rethrows.
			panic(js.Throw("Error", payload[0].Object().Get("message").Str()))
		}
		return false
	}
	var keep []js.Object
	for _, rec := range recs {
		fn := rec.Get("fn")
		if fn.Valid() && fn.IsFunction() {
			if _, err := fn.Object().Call(o, payload...); err != nil {
				panic(err)
			}
		}
		if !rec.Get("once").Bool() {
			keep = append(keep, rec)
		}
	}
	m.storeListeners(o, name, keep)
	return true
}

func (m *module) makeCPProto() js.Object {
	p := m.e.JS.Obj()
	p.SetFunc("on", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.addListener(this, args, false)
	})
	p.SetFunc("addListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.addListener(this, args, false)
	})
	p.SetFunc("once", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.addListener(this, args, true)
	})
	p.SetFunc("removeListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.removeListener(this, args)
	})
	p.SetFunc("off", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.removeListener(this, args)
	})
	p.SetFunc("removeAllListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.removeAllListeners(this, args)
	})
	p.SetFunc("setMaxListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return this.V(), nil
	})
	p.SetFunc("getMaxListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Num(1 << 31), nil
	})
	p.SetFunc("emit", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return m.e.JS.Bool(false), nil
		}
		ok := m.emitValue(this, args[0].Str(), args[1:])
		return m.e.JS.Bool(ok), nil
	})
	p.SetFunc("ref", func(this js.Object, args []js.Value) (js.Value, error) {
		return this.V(), nil
	})
	p.SetFunc("unref", func(this js.Object, args []js.Value) (js.Value, error) {
		id := this.Get("$godeCpId").Int()
		if st, ok := m.children[id]; ok {
			st.unrefed = true
			// With nothing observing the child, drop the pump so a detached
			// child does not keep the parent alive (env has no unref'd timers).
			if !st.finished && st.async != nil && len(m.listeners(st.cp, "exit")) == 0 &&
				len(m.listeners(st.cp, "close")) == 0 {
				if st.async.timer != 0 {
					m.e.ClearTimer(st.async.timer)
					st.async.timer = 0
				}
			}
		}
		return this.V(), nil
	})

	p.SetFunc("listenerCount", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return m.e.JS.Num(0), nil
		}
		return m.e.JS.Num(float64(len(m.listeners(this, args[0].Str())))), nil
	})
	p.SetFunc("send", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.sendToChild(this, args)
	})
	p.SetFunc("disconnect", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.disconnectChild(this)
	})
	p.SetFunc("kill", func(this js.Object, args []js.Value) (js.Value, error) {
		sig := "SIGTERM"
		if len(args) > 0 && args[0].Valid() && !args[0].IsUndefined() {
			switch {
			case args[0].IsString():
				sig = strings.ToUpper(args[0].Str())
			case args[0].IsNumber():
				name, ok := signalByNumber(int(args[0].Int()))
				if !ok {
					return js.Value{}, js.ThrowCode("TypeError", "ERR_UNKNOWN_SIGNAL",
						"Unknown signal: "+numberText(args[0]))
				}
				sig = name
			default:
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "signal" argument must be of type string or number. Received `+jsTypeName(args[0]))
			}
		}
		if !validSignal(sig) {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_UNKNOWN_SIGNAL", "Unknown signal: "+sig)
		}
		id := this.Get("$godeCpId").Int()
		st, ok := m.children[id]
		if !ok || st.finished {
			// Node returns false once the child is gone; it never throws ESRCH
			// from ChildProcess#kill.
			return m.e.JS.Bool(false), nil
		}
		if !m.signalChild(st, sig) {
			return m.e.JS.Bool(false), nil
		}
		return m.e.JS.Bool(true), nil
	})
	// Explicit resource management: `using child = spawn(...)` terminates the
	// child on scope exit, which Node implements as kill().
	if def, err := m.e.JS.Eval(`(function (proto) {
		if (typeof Symbol.dispose !== 'symbol') return false;
		Object.defineProperty(proto, Symbol.dispose, {
			value: function () { try { this.kill(); } catch (e) {} return undefined; },
			writable: true, configurable: true
		});
		return true;
	})`, "child_process.dispose"); err == nil {
		_, _ = def.Object().Call(js.Object{}, p.V())
	}
	return p
}

// sendToChild serializes a message onto the fork channel.
// signalByNumber maps a libuv signal number to the canonical name Node uses.
func signalByNumber(n int) (string, bool) {
	switch syscall.Signal(n) {
	case syscall.SIGABRT:
		return "SIGABRT", true
	case syscall.SIGALRM:
		return "SIGALRM", true
	case syscall.SIGBUS:
		return "SIGBUS", true
	case syscall.SIGCHLD:
		return "SIGCHLD", true
	case syscall.SIGCONT:
		return "SIGCONT", true
	case syscall.SIGFPE:
		return "SIGFPE", true
	case syscall.SIGHUP:
		return "SIGHUP", true
	case syscall.SIGILL:
		return "SIGILL", true
	case syscall.SIGINT:
		return "SIGINT", true
	case syscall.SIGIO:
		return "SIGIO", true
	case syscall.SIGKILL:
		return "SIGKILL", true
	case syscall.SIGPIPE:
		return "SIGPIPE", true
	case syscall.SIGPROF:
		return "SIGPROF", true
	case syscall.SIGPWR:
		return "SIGPWR", true
	case syscall.SIGQUIT:
		return "SIGQUIT", true
	case syscall.SIGSEGV:
		return "SIGSEGV", true
	case syscall.SIGSTKFLT:
		return "SIGSTKFLT", true
	case syscall.SIGSTOP:
		return "SIGSTOP", true
	case syscall.SIGSYS:
		return "SIGSYS", true
	case syscall.SIGTERM:
		return "SIGTERM", true
	case syscall.SIGTRAP:
		return "SIGTRAP", true
	case syscall.SIGTSTP:
		return "SIGTSTP", true
	case syscall.SIGTTIN:
		return "SIGTTIN", true
	case syscall.SIGTTOU:
		return "SIGTTOU", true
	case syscall.SIGURG:
		return "SIGURG", true
	case syscall.SIGUSR1:
		return "SIGUSR1", true
	case syscall.SIGUSR2:
		return "SIGUSR2", true
	case syscall.SIGVTALRM:
		return "SIGVTALRM", true
	case syscall.SIGWINCH:
		return "SIGWINCH", true
	case syscall.SIGXCPU:
		return "SIGXCPU", true
	case syscall.SIGXFSZ:
		return "SIGXFSZ", true
	}
	return "", false
}

func (m *module) sendToChild(cp js.Object, args []js.Value) (js.Value, error) {
	// ChildProcess#send(message[, handle[, options]][, callback]).
	message := arg(args, 0)
	handle := arg(args, 1)
	options := arg(args, 2)
	callback := arg(args, 3)
	if handle.Valid() && handle.IsFunction() {
		callback = handle
		handle = js.Value{}
		options = js.Value{}
	} else if options.Valid() && options.IsFunction() {
		callback = options
		options = js.Value{}
	} else if options.Valid() && !options.IsUndefined() {
		if !options.IsObject() || options.IsNull() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "options" argument must be of type object. Received `+jsTypeName(options))
		}
	}
	if !message.Valid() || message.IsUndefined() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_MISSING_ARGS",
			`The "message" argument must be specified`)
	}
	if !message.IsString() && !message.IsNumber() && !message.IsBool() && !message.IsObject() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "message" argument must be of type string, object, number, or boolean. Received `+jsTypeName(message))
	}
	if handle.Valid() && !handle.IsUndefined() && !handle.IsNull() && !handle.IsObject() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_HANDLE_TYPE",
			"This handle type cannot be sent")
	}
	id := cp.Get("$godeCpId").Int()
	st, ok := m.children[id]
	if !ok || !st.connected || st.async == nil || st.async.ipc == nil {
		// Node reports a closed channel asynchronously: to the callback when
		// one was supplied, otherwise as an 'error' event.
		err := m.e.JS.Error("Channel closed")
		err.SetStr("code", "ERR_IPC_CHANNEL_CLOSED")
		if callback.Valid() && callback.IsFunction() {
			m.e.NextTick(callback, err.V())
		} else {
			holder := m.e.JS.Obj()
			holder.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
				m.emitValue(cp, "error", []js.Value{err.V()})
				return m.e.JS.Undefined(), nil
			})
			m.e.NextTick(holder.Get("f"))
		}
		return m.e.JS.Bool(false), nil
	}
	text, err := m.stringify(message)
	if err != nil {
		return js.Value{}, err
	}
	if err := st.async.ipc.write(text); err != nil {
		m.closeChildIPC(st)
		return m.e.JS.Bool(false), nil
	}
	if cb := lastFunction(args); cb.Valid() {
		m.e.NextTick(cb, m.e.JS.Null())
	}
	return m.e.JS.Bool(true), nil
}

// disconnectChild closes the parent's write end and announces it, leaving the
// child free to flush what it already queued.
func (m *module) disconnectChild(cp js.Object) (js.Value, error) {
	id := cp.Get("$godeCpId").Int()
	st, ok := m.children[id]
	if !ok || !st.connected {
		return cp.V(), nil
	}
	if st.async != nil && st.async.ipc != nil {
		st.async.ipc.closeWrite()
	}
	st.connected = false
	cp.SetBool("connected", false)
	holder := m.e.JS.Obj()
	holder.SetFuncHidden("f", func(this js.Object, args []js.Value) (js.Value, error) {
		if !st.disconnected {
			st.disconnected = true
			m.emitValue(st.cp, "disconnect", nil)
		}
		return m.e.JS.Undefined(), nil
	})
	m.e.NextTick(holder.Get("f"))
	return cp.V(), nil
}

// validSignal reports whether name is a signal this platform can deliver.
func validSignal(name string) bool {
	switch name {
	case "SIGABRT", "SIGALRM", "SIGBUS", "SIGCHLD", "SIGCONT", "SIGFPE",
		"SIGHUP", "SIGILL", "SIGINT", "SIGIO", "SIGIOT", "SIGKILL", "SIGPIPE",
		"SIGPOLL", "SIGPROF", "SIGPWR", "SIGQUIT", "SIGSEGV", "SIGSTKFLT",
		"SIGSTOP", "SIGSYS", "SIGTERM", "SIGTRAP", "SIGTSTP", "SIGTTIN",
		"SIGTTOU", "SIGURG", "SIGUSR1", "SIGUSR2", "SIGVTALRM", "SIGWINCH",
		"SIGXCPU", "SIGXFSZ":
		return true
	}
	return false
}

func (m *module) constructor() js.Value {
	holder := m.e.JS.Obj()
	holder.SetFunc("ChildProcess", func(this js.Object, args []js.Value) (js.Value, error) {
		return js.Value{}, js.Throw("TypeError", "Class constructor ChildProcess cannot be invoked without 'new'")
	})
	ctor := holder.Get("ChildProcess")
	ctor.Object().Set("prototype", m.cpProto.V())
	m.cpProto.Set("constructor", ctor)
	return ctor
}

// ------------------------------------------------------------------- streams --

func (m *module) newStream(kind string, id int, data []byte, encoding string) js.Value {
	holder := m.e.JS.Obj()
	holder.SetFunc("stream", func(this js.Object, args []js.Value) (js.Value, error) {
		o := m.e.JS.Obj()
		o.SetProto(m.streamProto())
		o.SetStr("kind", kind)
		o.SetHidden("$godeCpId", m.e.JS.Num(float64(id)))
		o.Set("$godePipes", m.e.JS.Arr(nil).V())
		enc := ""
		if kind != "stdin" {
			enc = encoding
		}
		if enc == "utf-8" {
			enc = "utf8"
		}
		o.SetHidden("$godeEncoding", m.e.JS.Str(enc))
		o.SetBool("readable", kind != "stdin")
		o.SetBool("writable", kind == "stdin")
		o.SetBool("readableFlowing", false)
		o.SetBool("destroyed", false)
		if data != nil {
			o.SetNum("_bytes", float64(len(data)))
			o.Set("$godeData", m.bytesValue(data))
		}
		// Node's pipe streams carry a libuv handle; tests poke at readStart and
		// friends to prove data flows through the stream layer instead.
		handle := m.e.JS.Obj()
		handle.SetFunc("readStart", func(this js.Object, args []js.Value) (js.Value, error) {
			return m.e.JS.Undefined(), nil
		})
		handle.SetFunc("readStop", func(this js.Object, args []js.Value) (js.Value, error) {
			return m.e.JS.Undefined(), nil
		})
		handle.SetFunc("close", func(this js.Object, args []js.Value) (js.Value, error) {
			return m.e.JS.Undefined(), nil
		})
		o.SetHidden("_handle", handle.V())
		return o.V(), nil
	})
	v, err := holder.Get("stream").Object().Call(js.Object{})
	if err != nil {
		panic("godejs child_process stream: " + err.Error())
	}
	return v
}

func (m *module) streamProto() js.Object {
	if p := m.streams.Get("proto"); p.IsObject() {
		return p.Object()
	}
	p := m.e.JS.Obj()
	m.streams.Set("proto", p.V())
	p.SetFunc("on", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.addListener(this, args, false)
	})
	p.SetFunc("addListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.addListener(this, args, false)
	})
	p.SetFunc("once", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.addListener(this, args, true)
	})
	p.SetFunc("removeListener", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.removeListener(this, args)
	})
	p.SetFunc("off", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.removeListener(this, args)
	})
	p.SetFunc("removeAllListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.removeAllListeners(this, args)
	})
	p.SetFunc("setMaxListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return this.V(), nil
	})
	p.SetFunc("getMaxListeners", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Num(1 << 31), nil
	})
	p.SetFunc("emit", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return m.e.JS.Bool(false), nil
		}
		return m.e.JS.Bool(m.emitValue(this, args[0].Str(), args[1:])), nil
	})
	p.SetFunc("setEncoding", func(this js.Object, args []js.Value) (js.Value, error) {
		enc := "utf8"
		if len(args) > 0 {
			enc = args[0].Str()
		}
		this.SetHidden("$godeEncoding", m.e.JS.Str(enc))
		return this.V(), nil
	})
	p.SetFunc("write", func(this js.Object, args []js.Value) (js.Value, error) {
		if this.Get("kind").Str() != "stdin" || this.Get("writableEnded").Bool() {
			return js.Value{}, js.ThrowCode("Error", "ERR_STREAM_WRITE_AFTER_END", "write after end")
		}
		cb := js.Value{}
		if len(args) > 0 {
			var data []byte
			if b, ok := args[0].Bytes(); ok {
				data = b
			} else if args[0].IsString() {
				data = []byte(args[0].Str())
			} else {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "chunk" argument must be of type string or an instance of Buffer or Uint8Array. Received `+jsTypeName(args[0]))
			}
			if len(args) > 1 && args[1].IsFunction() {
				cb = args[1]
			}
			id := this.Get("$godeCpId").Int()
			if st, ok := m.children[id]; ok && st.async != nil && st.async.stdinW != nil {
				if _, err := st.async.stdinW.Write(data); err != nil {
					return m.e.JS.Bool(false), nil
				}
			}
		}
		if cb.Valid() {
			m.e.NextTick(cb)
		}
		return m.e.JS.Bool(true), nil
	})
	p.SetFunc("end", func(this js.Object, args []js.Value) (js.Value, error) {
		id := this.Get("$godeCpId").Int()
		if st, ok := m.children[id]; ok && st.async != nil && st.async.stdinW != nil {
			if len(args) > 0 {
				if b, ok := args[0].Bytes(); ok {
					_, _ = st.async.stdinW.Write(b)
				} else if args[0].IsString() {
					_, _ = st.async.stdinW.Write([]byte(args[0].Str()))
				}
			}
			_ = st.async.stdinW.Close()
		}
		this.SetBool("writableEnded", true)
		m.emitValue(this, "finish", nil)
		m.emitValue(this, "close", nil)
		return this.V(), nil
	})
	p.SetFunc("pipe", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsObject() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "destination" argument must be of type object. Received `+jsTypeName(arg(args, 0)))
		}
		dest := args[0].Object()
		list := m.pipes(this)
		list = append(list, dest)
		arr := m.e.JS.Arr(nil)
		for i, d := range list {
			arr.SetAt(i, d.V())
		}
		this.Set("$godePipes", arr.V())
		m.emitValue(this, "pipe", []js.Value{dest.V()})
		return dest.V(), nil
	})
	p.SetFunc("read", func(this js.Object, args []js.Value) (js.Value, error) {
		data := this.Get("$godeData")
		if !data.Valid() || data.IsUndefined() {
			return m.e.JS.Null(), nil
		}
		this.Set("$godeData", m.e.JS.Undefined())
		this.SetNum("_bytes", 0)
		enc := this.Get("$godeEncoding").Str()
		if b, ok := data.Bytes(); ok {
			return m.decodeBytes(b, enc), nil
		}
		return data, nil
	})
	p.SetFunc("resume", func(this js.Object, args []js.Value) (js.Value, error) {
		this.SetBool("readableFlowing", true)
		return this.V(), nil
	})
	p.SetFunc("pause", func(this js.Object, args []js.Value) (js.Value, error) {
		this.SetBool("readableFlowing", false)
		return this.V(), nil
	})
	p.SetFunc("isPaused", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.e.JS.Bool(!this.Get("readableFlowing").Bool()), nil
	})
	p.SetFunc("destroy", func(this js.Object, args []js.Value) (js.Value, error) {
		if this.Get("destroyed").Bool() {
			return this.V(), nil
		}
		this.SetBool("destroyed", true)
		id := this.Get("$godeCpId").Int()
		if st, ok := m.children[id]; ok && st.async != nil && st.async.stdinW != nil &&
			this.Get("kind").Str() == "stdin" {
			_ = st.async.stdinW.Close()
			st.async.stdinW = nil
		}
		m.emitValue(this, "close", nil)
		return this.V(), nil
	})
	p.SetFunc("ref", func(this js.Object, args []js.Value) (js.Value, error) { return this.V(), nil })
	p.SetFunc("unref", func(this js.Object, args []js.Value) (js.Value, error) { return this.V(), nil })
	return p
}
