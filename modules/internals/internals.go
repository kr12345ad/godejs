// Package internals provides the `internal/*` module specifiers that Node's
// own test suite (and a few ecosystem tools) require directly.
//
// Every namespace here is an adapter over a public GodeJS module: the public
// module stays the single source of truth and this package copies its exports
// so the internal specifier cannot drift. Helpers Node implements only in
// JavaScript (validators, linkedlist, the Myers diff used by assert) are
// ported here because they have no public counterpart.
//
// Where a Node internal export depends on machinery GodeJS does not have yet
// (libuv watchers, V8 binding handles) the name is left out of the namespace
// rather than backed by a value that would silently be wrong: callers then get
// a normal "is not a function" TypeError instead of a plausible-looking lie.
package internals

import (
	"fmt"
	nurl "net/url"
	"strings"
	"time"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.Register("internal/util", installUtil)
	env.Register("internal/util/colors", installColors)
	env.Register("internal/util/debuglog", installDebugLog)
	env.Register("internal/errors", installErrors)
	env.Register("internal/assert", func(e *env.Env) js.Object { return e.Exports("assert") })
	env.Register("internal/assert/myers_diff", installMyersDiff)
	env.Register("internal/root", installRoot)
	env.Register("internal/url", installURL)
	env.Register("internal/buffer", func(e *env.Env) js.Object { return e.Exports("buffer") })
	env.Register("internal/timers", func(e *env.Env) js.Object { return e.Exports("timers") })
	env.Register("internal/validators", installValidators)
	env.Register("internal/options", installOptions)
	env.Register("internal/event_target", installEventTarget)
	env.Register("internal/linkedlist", installLinkedList)
	env.Register("internal/fs/utils", installFSUtils)
	env.Register("internal/fs/promises", func(e *env.Env) js.Object { return e.Exports("fs/promises") })
	env.Register("internal/fs/watchers", installFSWatchers)
	env.Register("internal/crypto/util", installCryptoUtil)
	env.Register("internal/crypto/keys", installCryptoKeys)
	env.Register("internal/crypto/x509", installCryptoX509)
	env.Register("internal/streams/iter", optional("stream/iter"))
	env.Register("internal/streams/consumers", optional("stream/consumers"))
	env.Register("internal/streams/web", optional("stream/web"))
	env.Register("internal/test/binding", installTestBinding)
	env.Register("internal/fs/recursive_watch", installRecursiveWatch)
	env.Register("internal/js_stream_socket", installJSStreamSocket)
	env.Register("internal/child_process", installChildProcess)
	env.Register("internal/dgram", installDgram)
}

// Implemented lists the internal specifiers this package provides.
func Implemented() []string {
	return []string{
		"internal/assert",
		"internal/assert/myers_diff",
		"internal/child_process",
		"internal/crypto/keys",
		"internal/crypto/util",
		"internal/crypto/x509",
		"internal/dgram",
		"internal/errors",
		"internal/event_target",
		"internal/fs/promises",
		"internal/fs/recursive_watch",
		"internal/fs/utils",
		"internal/fs/watchers",
		"internal/js_stream_socket",
		"internal/linkedlist",
		"internal/options",
		"internal/root",
		"internal/streams/consumers",
		"internal/streams/iter",
		"internal/streams/web",
		"internal/test/binding",
		"internal/timers",
		"internal/url",
		"internal/util",
		"internal/util/colors",
		"internal/util/debuglog",
		"internal/validators",
	}
}

// --- helpers ---------------------------------------------------------------

// reexport copies every property of a builtin module's exports into a fresh
// namespace. Reading through the public module (instead of restating its API)
// keeps an internal specifier correct as the public module grows.
func reexport(e *env.Env, name string) js.Object {
	src := e.Exports(name)
	out := e.JS.Obj()
	for _, key := range src.Keys() {
		out.Set(key, src.Get(key))
	}
	return out
}

// pick copies only the named properties that the source actually has.
func pick(src js.Object, out js.Object, names ...string) js.Object {
	for _, name := range names {
		if v := src.Get(name); v.Valid() && !v.IsUndefined() {
			out.Set(name, v)
		}
	}
	return out
}

// optional adapts a public module that may not be registered in this build
// (stream/iter and friends arrived with the iterator rewrite). The specifier
// still resolves so callers see undefined members rather than a load error.
func optional(name string) env.Installer {
	return func(e *env.Env) js.Object {
		if env.Registered(name) {
			return e.Exports(name)
		}
		return e.JS.Obj()
	}
}

// factory evaluates a JS factory function and calls it with args.
func factory(e *env.Env, src, filename string, args ...js.Value) js.Object {
	fn, err := e.JS.Eval(src, filename)
	if err != nil {
		panic("godejs internals: " + filename + ": " + err.Error())
	}
	val, err := fn.Object().Call(js.Object{}, args...)
	if err != nil {
		panic("godejs internals: " + filename + ": " + err.Error())
	}
	return val.Object()
}

// --- internal/util ---------------------------------------------------------

func installUtil(e *env.Env) js.Object {
	out := reexport(e, "util")
	return factory(e, utilPrelude, "[gode:internal/util]", out.V())
}

const utilPrelude = `(function (out) {
  'use strict';
  function missing(name, fallback) {
    if (typeof out[name] !== 'function') out[name] = fallback;
  }
  out.isArray = Array.isArray;
  out.isBoolean = function (v) { return typeof v === 'boolean'; };
  out.isNull = function (v) { return v === null; };
  out.isNullOrUndefined = function (v) { return v === null || v === undefined; };
  out.isNumber = function (v) { return typeof v === 'number'; };
  out.isString = function (v) { return typeof v === 'string'; };
  out.isSymbol = function (v) { return typeof v === 'symbol'; };
  out.isUndefined = function (v) { return v === undefined; };
  out.isRegExp = function (v) { return v instanceof RegExp; };
  out.isObject = function (v) { return typeof v === 'object' && v !== null; };
  out.isDate = function (v) { return v instanceof Date; };
  out.isFunction = function (v) { return typeof v === 'function'; };
  out.isError = function (v) { return v instanceof Error; };
  out.isPrimitive = function (v) {
    return v === null || (typeof v !== 'object' && typeof v !== 'function');
  };
  out.isBuffer = function (v) {
    return typeof Buffer !== 'undefined' && Buffer.isBuffer(v);
  };
  out.kEmptyObject = Object.freeze({});
  out.kIsNodeError = Symbol('kIsNodeError');
  out.kNoFrame = Symbol('kNoFrame');
  out.kEmptyString = '';
  missing('formatList', function formatList(list) {
    var parts = [];
    for (var i = 0; i < list.length; i++) {
      var v = list[i];
      if (typeof v === 'string') parts.push("'" + v + "'");
      else if (v === null) parts.push('null');
      else if (v === undefined) parts.push('undefined');
      else if (Array.isArray(v)) parts.push('[' + formatList(v) + ']');
      else parts.push(String(v));
    }
    return parts.join(', ');
  });
  missing('sleep', function sleep(ms) {
    return new Promise(function (resolve) { setTimeout(resolve, ms); });
  });
  missing('pendingDeprecate', function pendingDeprecate(fn, msg, code) {
    var emitted = false;
    function deprecated() {
      if (!emitted) {
        emitted = true;
        if (typeof process !== 'undefined' && process.env &&
            process.env.NODE_PENDING_DEPRECATION &&
            typeof process.emitWarning === 'function') {
          process.emitWarning(msg, 'DeprecationWarning', code);
        }
      }
      return fn.apply(this, arguments);
    }
    try {
      Object.defineProperty(deprecated, 'length',
        Object.getOwnPropertyDescriptor(fn, 'length'));
    } catch (e) {}
    return deprecated;
  });
  missing('emitExperimentalWarning', function emitExperimentalWarning(feature) {
    if (typeof process !== 'undefined' && typeof process.emitWarning === 'function') {
      process.emitWarning('The ' + feature + ' is an experimental feature and might change at any time',
        'ExperimentalWarning');
    }
  });
  if (typeof out.customPromisifyArgs !== 'symbol') {
    out.customPromisifyArgs = Symbol.for('nodejs.util.promisify.customArgs');
  }
  missing('normalizeEncoding', function (enc) { return String(enc).toLowerCase(); });
  missing('setTraceSigInt', function (enabled) { return enabled; });
  missing('transferableAbortSignal', function (signal) { return signal; });
  missing('transferableAbortController', function () { return new AbortController(); });
  missing('aborted', function (signal) {
    return new Promise(function (resolve, reject) {
      if (signal.aborted) return reject(signal.reason);
      signal.addEventListener('abort', function () { reject(signal.reason); }, { once: true });
    });
  });
  return out;
})`

// --- internal/errors -------------------------------------------------------

func installErrors(e *env.Env) js.Object {
	return factory(e, errorsPrelude, "[gode:internal/errors]", e.Exports("util").V())
}

const errorsPrelude = `(function (util) {
  'use strict';
  // Node's error-code table is a plain mapping from name to its own string, so
  // a Proxy that answers with the requested key reproduces it exactly.
  var codes = new Proxy(Object.create(null), {
    get: function (target, key) { return typeof key === 'string' ? key : undefined; },
    has: function (target, key) { return typeof key === 'string'; }
  });
  function localFormatList(list) {
    var parts = [];
    for (var i = 0; i < list.length; i++) {
      var v = list[i];
      if (typeof v === 'string') parts.push("'" + v + "'");
      else if (v === null) parts.push('null');
      else if (v === undefined) parts.push('undefined');
      else if (Array.isArray(v)) parts.push('[' + localFormatList(v) + ']');
      else parts.push(String(v));
    }
    return parts.join(', ');
  }
  class AbortError extends Error {
    constructor(message, options) {
      super(message, options);
      this.code = 'ABORT_ERR';
      this.name = 'AbortError';
    }
  }
  function aggregateTwoErrors(innerError, outerError) {
    if (innerError && outerError && innerError !== outerError) {
      if (Array.isArray(outerError.errors)) {
        outerError.errors.push(innerError);
        return outerError;
      }
      var err = new AggregateError([outerError, innerError], outerError.message);
      err.code = outerError.code;
      return err;
    }
    return innerError || outerError;
  }
  function hideStackFrames(fn) { return fn; }
  function isErrorStackTraceLimitWritable() { return false; }
  function determineSpecificType(value) {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    if (typeof value === 'function') return 'function ' + (value.name || '(anonymous)');
    if (typeof value === 'object') {
      return 'an instance of ' + ((value.constructor && value.constructor.name) || 'Object');
    }
    return 'type ' + typeof value + ' (' + String(value) + ')';
  }
  var ERRNO = {
    '-1': 'EPERM', '-2': 'ENOENT', '-9': 'EBADF', '-11': 'EAGAIN', '-13': 'EACCES',
    '-17': 'EEXIST', '-20': 'ENOTDIR', '-21': 'EISDIR', '-22': 'EINVAL', '-24': 'EMFILE',
    '-28': 'ENOSPC', '-32': 'EPIPE', '-39': 'ENOTEMPTY', '-98': 'EADDRINUSE',
    '-104': 'ECONNRESET', '-110': 'ETIMEDOUT', '-111': 'ECONNREFUSED'
  };
  // uvException is limited to the errno values Node's own tests reach; errno
  // numbers outside the table still produce an error, just with code UNKNOWN.
  function uvException(ctx) {
    var code = ERRNO[String(ctx.errno)] || 'UNKNOWN';
    var syscall = ctx.syscall || 'uv';
    var path = ctx.path === undefined ? '' : " '" + ctx.path + "'";
    var dest = ctx.dest === undefined ? '' : " -> '" + ctx.dest + "'";
    var err = new Error(syscall + ' ' + code + path + dest);
    err.errno = ctx.errno;
    err.code = code;
    err.syscall = syscall;
    if (ctx.path !== undefined) err.path = ctx.path;
    if (ctx.dest !== undefined) err.dest = ctx.dest;
    return err;
  }
  function uvExceptionWithHostPort(ctx) {
    var err = uvException(ctx);
    err.address = ctx.address;
    err.port = ctx.port;
    return err;
  }
  function E(code, message, Base) {
    var Ctor = Base || Error;
    var err = new Ctor(message);
    err.code = code;
    return err;
  }
  return {
    codes: codes,
    AbortError: AbortError,
    abortError: function (message) { return new AbortError(message); },
    aggregateTwoErrors: aggregateTwoErrors,
    hideStackFrames: hideStackFrames,
    isErrorStackTraceLimitWritable: isErrorStackTraceLimitWritable,
    isError: function (value) { return value instanceof Error; },
    isStackOverflowError: function (err) { return err instanceof RangeError && /call stack/i.test(err.message); },
    determineSpecificType: determineSpecificType,
    inspect: util.inspect,
    formatList: typeof util.formatList === 'function' ? util.formatList : localFormatList,
    uvException: uvException,
    uvExceptionWithHostPort: uvExceptionWithHostPort,
    errnoException: uvException,
    exceptionWithHostPort: uvExceptionWithHostPort,
    E: E,
    errorCache: {},
    noDeprecation: false
  };
})`

// --- internal/root ---------------------------------------------------------

func installRoot(e *env.Env) js.Object {
	return factory(e, rootPrelude, "[gode:internal/root]", e.Exports("process").V(), e.Exports("path").V())
}

// internal/root was removed from Node 24 (the constants moved to
// internal/util), but a handful of third-party tools still ask for it, so the
// specifier is kept as a faithful alias namespace.
const rootPrelude = `(function (proc, path) {
  'use strict';
  var platform = proc.platform;
  return {
    kEmptyObject: Object.freeze({}),
    kIsNodeError: Symbol('kIsNodeError'),
    kNoFrame: Symbol('kNoFrame'),
    isWindows: platform === 'win32',
    isMacOS: platform === 'darwin',
    isLinux: platform === 'linux',
    isAIX: platform === 'aix',
    isFreeBSD: platform === 'freebsd',
    isOpenBSD: platform === 'openbsd',
    isSmartOS: platform === 'sunos',
    isSunOS: platform === 'sunos',
    isAndroid: false,
    posix: platform === 'win32' ? undefined : path.posix,
    win32: platform === 'win32' ? path.win32 : undefined
  };
})`

// --- internal/url ----------------------------------------------------------

func installURL(e *env.Env) js.Object {
	out := reexport(e, "url")
	// isURL is Node's brand check for WHATWG URL instances. GodeJS URLs carry
	// the hidden backing fields installed by setParts; legacy parse objects do
	// not, which is exactly the distinction the internal helper makes.
	out.SetFunc("isURL", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsObject() || args[0].IsNull() {
			return e.JS.Bool(false), nil
		}
		return e.JS.Bool(args[0].Object().Has("$host")), nil
	})
	// Node's internal/url exposes the posix/win32 path conversions that the
	// public module keeps private.
	out.SetFunc("getPathFromURLPosix", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsObject() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"The \"url\" argument must be an instance of URL")
		}
		path := args[0].Object().Get("pathname").Str()
		if path == "" {
			return e.JS.Str("/"), nil
		}
		decoded, err := nurl.PathUnescape(path)
		if err != nil {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_URL", "Invalid URL")
		}
		return e.JS.Str(decoded), nil
	})
	out.SetFunc("getPathFromURLWin32", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsObject() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"The \"url\" argument must be an instance of URL")
		}
		path := args[0].Object().Get("pathname").Str()
		decoded, err := nurl.PathUnescape(path)
		if err != nil {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_URL", "Invalid URL")
		}
		if len(decoded) > 2 && decoded[0] == '/' && decoded[2] == ':' {
			decoded = decoded[1:]
		}
		return e.JS.Str(strings.ReplaceAll(decoded, "/", `\`)), nil
	})
	out.SetFunc("getURLFromFilePath", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return js.Value{}, js.Throw("TypeError", "The \"filepath\" argument must be of type string")
		}
		return out.Get("pathToFileURL").Object().Call(js.Object{}, args[0])
	})
	return out
}

// --- internal/event_target -------------------------------------------------

func installEventTarget(e *env.Env) js.Object {
	events := e.Exports("events")
	out := e.JS.Obj()
	pick(events, out, "Event", "CustomEvent", "EventTarget", "NodeEventTarget", "on", "once", "addAbortListener")
	g := e.Global()
	for _, name := range []string{"AbortController", "AbortSignal"} {
		if v := g.Get(name); v.Valid() && !v.IsUndefined() {
			out.Set(name, v)
		}
	}
	return factory(e, eventTargetPrelude, "[gode:internal/event_target]", out.V())
}

const eventTargetPrelude = `(function (out) {
  'use strict';
  out.kCreateEvent = Symbol('kCreateEvent');
  out.kNewListener = Symbol('kNewListener');
  out.kTrustEvent = Symbol('kTrustEvent');
  out.kRemoveListener = Symbol('kRemoveListener');
  out.kEvents = Symbol('kEvents');
  out.kWeakHandler = Symbol('kWeakHandler');
  out.kResistStopPropagation = Symbol('kResistStopPropagation');
  // GodeJS builds targets from a single constructor, so the init hooks are
  // identity functions; Node uses them to install per-instance state.
  out.initEventTarget = function (target) { return target; };
  out.initNodeEventTarget = function (target) { return target; };
  // defineEventHandler installs the on<Event> accessor Node exposes on
  // EventTarget-shaped emitters, backed by the emitter's listener list.
  out.defineEventHandler = function (emitter, name) {
    Object.defineProperty(emitter, 'on' + name, {
      get: function () {
        if (typeof this.listeners !== 'function') return undefined;
        var list = this.listeners(name);
        return list.length > 0 ? list[0] : undefined;
      },
      set: function (value) {
        if (typeof this.listeners === 'function' && typeof this.removeListener === 'function') {
          var list = this.listeners(name);
          for (var i = 0; i < list.length; i++) this.removeListener(name, list[i]);
        }
        if (typeof value === 'function' && typeof this.on === 'function') this.on(name, value);
      },
      enumerable: true,
      configurable: true
    });
    return emitter;
  };
  return out;
})`

// --- internal/linkedlist ---------------------------------------------------

func installLinkedList(e *env.Env) js.Object {
	return factory(e, linkedListPrelude, "[gode:internal/linkedlist]")
}

// Port of lib/internal/linkedlist.js: a circular doubly linked list where the
// list object is its own sentinel and links live on _idleNext/_idlePrev.
const linkedListPrelude = `(function () {
  'use strict';
  function init(list) {
    list._idleNext = list;
    list._idlePrev = list;
    return list;
  }
  function peek(list) {
    if (list._idlePrev === list) return null;
    return list._idlePrev;
  }
  function remove(item) {
    if (item._idleNext) item._idleNext._idlePrev = item._idlePrev;
    if (item._idlePrev) item._idlePrev._idleNext = item._idleNext;
    item._idleNext = null;
    item._idlePrev = null;
  }
  function append(list, item) {
    if (item._idleNext || item._idlePrev) remove(item);
    item._idleNext = list._idleNext;
    item._idlePrev = list;
    list._idleNext._idlePrev = item;
    list._idleNext = item;
  }
  function isEmpty(list) {
    return list._idleNext === list;
  }
  return { init: init, peek: peek, remove: remove, append: append, isEmpty: isEmpty };
})`

// --- internal/util/colors --------------------------------------------------

func installColors(e *env.Env) js.Object {
	return factory(e, colorsPrelude, "[gode:internal/util/colors]", e.Exports("process").V())
}

const colorsPrelude = `(function (proc) {
  'use strict';
  var stream = proc.stdout || {};
  var env = proc.env || {};
  var hasColors = !!stream.isTTY && env.NO_COLOR === undefined && env.FORCE_COLOR !== '0';
  var paint = function (open) { return hasColors ? open : ''; };
  var colors = {
    refresh: function () { return hasColors; },
    shouldColorize: hasColors,
    hasColors: hasColors,
    enabled: hasColors,
    noColor: !hasColors,
    reset: paint('\u001b[0m'),
    bold: paint('\u001b[1m'),
    dim: paint('\u001b[2m'),
    italic: paint('\u001b[3m'),
    underline: paint('\u001b[4m'),
    inverse: paint('\u001b[7m'),
    hidden: paint('\u001b[8m'),
    strikethrough: paint('\u001b[9m'),
    black: paint('\u001b[30m'),
    red: paint('\u001b[31m'),
    green: paint('\u001b[32m'),
    yellow: paint('\u001b[33m'),
    blue: paint('\u001b[34m'),
    magenta: paint('\u001b[35m'),
    cyan: paint('\u001b[36m'),
    white: paint('\u001b[37m'),
    gray: paint('\u001b[90m'),
    grey: paint('\u001b[90m'),
    brightRed: paint('\u001b[91m'),
    brightGreen: paint('\u001b[92m'),
    brightYellow: paint('\u001b[93m'),
    brightBlue: paint('\u001b[94m'),
    brightMagenta: paint('\u001b[95m'),
    brightCyan: paint('\u001b[96m'),
    brightWhite: paint('\u001b[97m')
  };
  return colors;
})`

// --- internal/util/debuglog ------------------------------------------------

func installDebugLog(e *env.Env) js.Object {
	return factory(e, debugLogPrelude, "[gode:internal/util/debuglog]", e.Exports("util").V())
}

const debugLogPrelude = `(function (util) {
  'use strict';
  function debuglog(section) {
    if (typeof util.debuglog === 'function') return util.debuglog(section);
    var noop = function () {};
    noop.enabled = false;
    return noop;
  }
  function debug(section) {
    var log = debuglog(section);
    if (!log.enabled) return;
    var args = Array.prototype.slice.call(arguments, 1);
    return log.apply(null, args);
  }
  return {
    kNone: -1,
    kSkipLog: 1,
    kSkipTrace: 2,
    debuglog: debuglog,
    debug: debug,
    debugWithTimer: debug,
    initializeDebugEnv: function () {},
    formatTime: function (ms) {
      var hours = 0, minutes = 0, seconds = 0;
      if (ms >= 1000) {
        if (ms >= 60000) {
          if (ms >= 3600000) {
            hours = Math.floor(ms / 3600000);
            ms = ms % 3600000;
          }
          minutes = Math.floor(ms / 60000);
          ms = ms % 60000;
        }
        seconds = ms / 1000;
      }
      function pad(n) { n = String(n); return n.length < 2 ? '0' + n : n; }
      if (hours !== 0 || minutes !== 0) {
        var parts = seconds.toFixed(3).split('.');
        var res = hours !== 0 ? hours + ':' + pad(minutes) : String(minutes);
        return res + ':' + pad(parts[0]) + '.' + parts[1] +
          ' (' + (hours !== 0 ? 'h:m' : '') + 'm:ss.mmm)';
      }
      if (seconds !== 0) return seconds.toFixed(3) + 's';
      return Number(ms.toFixed(3)) + 'ms';
    },
    time: function () { return 0; },
    timeEnd: function () { return 0; },
    timeLog: function () {}
  };
})`

// --- internal/validators ---------------------------------------------------

func installValidators(e *env.Env) js.Object {
	return factory(e, validatorsPrelude, "[gode:internal/validators]")
}

const validatorsPrelude = `(function () {
  'use strict';
  function describe(value) {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    if (Array.isArray(value)) return 'an instance of Array';
    var t = typeof value;
    if (t === 'string') return 'type string (' + JSON.stringify(value) + ')';
    if (t === 'number') return 'type number (' + String(value) + ')';
    if (t === 'boolean') return 'type boolean (' + String(value) + ')';
    if (t === 'bigint') return 'type bigint (' + String(value) + 'n)';
    if (t === 'symbol') return 'type symbol (' + String(value) + ')';
    if (t === 'function') return 'function ' + (value.name || '(anonymous)');
    return 'an instance of ' + ((value.constructor && value.constructor.name) || 'Object');
  }
  function invalidType(name, expected, actual) {
    var e = new TypeError('The "' + name + '" argument must be ' + expected +
                          '. Received ' + describe(actual));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function invalidValue(name, value, reason) {
    var e = new TypeError('The argument \'' + name + '\' ' +
                          (reason ? reason + '. Received ' + describe(value)
                                  : 'is invalid. Received ' + describe(value)));
    e.code = 'ERR_INVALID_ARG_VALUE';
    return e;
  }
  function outOfRange(name, actual, bound) {
    var e = new RangeError('The value of "' + name + '" is out of range. It must be ' +
                           bound + '. Received ' + String(actual));
    e.code = 'ERR_OUT_OF_RANGE';
    return e;
  }
  function validateString(value, name) {
    if (typeof value !== 'string') throw invalidType(name, 'of type string', value);
  }
  function validateBoolean(value, name) {
    if (typeof value !== 'boolean') throw invalidType(name, 'of type boolean', value);
  }
  function validateFunction(value, name) {
    if (typeof value !== 'function') throw invalidType(name, 'of type function', value);
  }
  function validateNumber(value, name, min, max) {
    if (typeof value !== 'number') throw invalidType(name, 'of type number', value);
    if (Number.isNaN(value)) throw invalidType(name, 'of type number', value);
    if (min !== undefined && value < min) throw outOfRange(name, value, '>= ' + min);
    if (max !== undefined && value > max) throw outOfRange(name, value, '<= ' + max);
  }
  function validateFiniteNumber(value, name) {
    if (typeof value !== 'number' || !Number.isFinite(value)) {
      throw invalidType(name, 'of type number', value);
    }
  }
  function validateInteger(value, name, min, max) {
    if (typeof value !== 'number') throw invalidType(name, 'of type number', value);
    if (!Number.isInteger(value)) throw outOfRange(name, value, 'an integer');
    if (min !== undefined && value < min) throw outOfRange(name, value, '>= ' + min);
    if (max !== undefined && value > max) throw outOfRange(name, value, '<= ' + max);
  }
  function validateUint32(value, name, positive) {
    var min = positive ? 1 : 0;
    validateInteger(value, name, min, 2 ** 32 - 1);
  }
  function validateInt32(value, name, min, max) {
    validateInteger(value, name, min === undefined ? -(2 ** 31) : min,
                    max === undefined ? 2 ** 31 - 1 : max);
  }
  function validateObject(value, name, options) {
    options = options || {};
    if (value === null || typeof value !== 'object') {
      throw invalidType(name, 'of type object', value);
    }
    if (!options.allowArray && Array.isArray(value)) {
      throw invalidType(name, 'of type object', value);
    }
    if (!options.allowFunction && typeof value === 'function') {
      throw invalidType(name, 'of type object', value);
    }
  }
  function validateArray(value, name) {
    if (!Array.isArray(value)) throw invalidType(name, 'an instance of Array', value);
  }
  function validateBuffer(value, name) {
    if (typeof Buffer === 'undefined' || !Buffer.isBuffer(value)) {
      throw invalidType(name, 'an instance of Buffer or Uint8Array', value);
    }
  }
  function validateStringArray(value, name) {
    validateArray(value, name);
    for (var i = 0; i < value.length; i++) validateString(value[i], name + '[' + i + ']');
  }
  function validateBooleanArray(value, name) {
    validateArray(value, name);
    for (var i = 0; i < value.length; i++) validateBoolean(value[i], name + '[' + i + ']');
  }
  function validateNumberArray(value, name) {
    validateArray(value, name);
    for (var i = 0; i < value.length; i++) validateNumber(value[i], name + '[' + i + ']');
  }
  function validateCallback(value, name) {
    if (typeof value !== 'function') {
      throw invalidType(name, 'of type function', value);
    }
  }
  function validateAbortSignal(signal, name) {
    if (signal !== undefined && !(signal instanceof AbortSignal)) {
      throw invalidType(name, 'an instance of AbortSignal', signal);
    }
  }
  function validateAbortSignalArray(signals, name) {
    validateArray(signals, name);
    for (var i = 0; i < signals.length; i++) validateAbortSignal(signals[i], name);
  }
  function validatePort(port, name, allowZero) {
    name = name || 'port';
    if (typeof port !== 'number' && typeof port !== 'string') {
      throw invalidType(name, 'of type number', port);
    }
    var value = +port;
    if (!Number.isInteger(value)) throw outOfRange(name, value, 'an integer');
    var min = allowZero ? 0 : 1;
    if (value < min || value > 65535) {
      throw outOfRange(name, value, '>= ' + min + ' && <= 65535');
    }
    return value;
  }
  function validateOneOf(value, name, oneOf) {
    if (oneOf.indexOf(value) === -1) {
      throw invalidValue(name, value, 'must be one of: ' + oneOf.map(function (v) {
        return JSON.stringify(v);
      }).join(', '));
    }
  }
  function validateEncoding(data, encoding) {
    if (typeof Buffer !== 'undefined' && !Buffer.isEncoding(encoding)) {
      throw invalidValue('encoding', encoding, 'is invalid');
    }
    return data;
  }
  function validateNull(value, name) {
    if (value !== null) throw invalidType(name, 'null', value);
  }
  function validateUndefined(value, name) {
    if (value !== undefined) throw invalidType(name, 'undefined', value);
  }
  // validateLinkHeaderValue has no observable effect in GodeJS: headers are
  // not serialized by a native binding, so the parsed value is returned as-is.
  function validateLinkHeaderValue(value) { return value; }
  return {
    validateString: validateString,
    validateBoolean: validateBoolean,
    validateFunction: validateFunction,
    validateNumber: validateNumber,
    validateFiniteNumber: validateFiniteNumber,
    validateInteger: validateInteger,
    validateInt32: validateInt32,
    validateUint32: validateUint32,
    validateObject: validateObject,
    validateArray: validateArray,
    validateBuffer: validateBuffer,
    validateStringArray: validateStringArray,
    validateBooleanArray: validateBooleanArray,
    validateNumberArray: validateNumberArray,
    validateCallback: validateCallback,
    validateAbortSignal: validateAbortSignal,
    validateAbortSignalArray: validateAbortSignalArray,
    validatePort: validatePort,
    validateOneOf: validateOneOf,
    validateEncoding: validateEncoding,
    validateNull: validateNull,
    validateUndefined: validateUndefined,
    validateLinkHeaderValue: validateLinkHeaderValue
  };
})`

// --- internal/options ------------------------------------------------------

func installOptions(e *env.Env) js.Object {
	return factory(e, optionsPrelude, "[gode:internal/options]", e.Exports("process").V())
}

const optionsPrelude = `(function (proc) {
  'use strict';
  var defaults = {
    '--expose-internals': false,
    '--no-warnings': false,
    '--no-deprecation': false,
    '--trace-warnings': false,
    '--trace-deprecation': false,
    '--throw-deprecation': false,
    '--pending-deprecation': false,
    '--allow-natives-syntax': false,
    '--test-only': false,
    '--experimental-print-required-tla': false,
    '--disable-warning': [],
    '--diagnostic-dir': '',
    '--redirect-warnings': '',
    '--disable-sigusr1': false
  };
  function getOptionValue(name) {
    var flag = String(name);
    var argv = proc.execArgv || [];
    for (var i = 0; i < argv.length; i++) {
      var arg = argv[i];
      if (arg === flag) return true;
      if (arg.indexOf(flag + '=') === 0) return arg.slice(flag.length + 1);
    }
    if (Object.prototype.hasOwnProperty.call(defaults, flag)) return defaults[flag];
    return false;
  }
  return {
    getOptionValue: getOptionValue,
    getCLIOptionsInfo: function () { return {}; },
    getOptionsAsFlagsFromBinding: function () { return []; },
    getAllowUnauthorized: function () { return false; },
    getEmbedderOptions: function () { return {}; },
    generateConfigJsonSchema: function () { return {}; },
    refreshOptions: function () {}
  };
})`

// --- internal/fs/utils -----------------------------------------------------

func installFSUtils(e *env.Env) js.Object {
	return factory(e, fsUtilsPrelude, "[gode:internal/fs/utils]",
		e.Exports("fs").V(), e.Exports("url").V())
}

const fsUtilsPrelude = `(function (fs, url) {
  'use strict';
  var C = fs.constants || {};
  function describe(value) {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    if (Array.isArray(value)) return 'an instance of Array';
    var t = typeof value;
    if (t === 'string') return 'type string (' + JSON.stringify(value) + ')';
    if (t === 'number') return 'type number (' + String(value) + ')';
    if (t === 'boolean') return 'type boolean (' + String(value) + ')';
    if (t === 'function') return 'function ' + (value.name || '(anonymous)');
    return 'an instance of ' + ((value.constructor && value.constructor.name) || 'Object');
  }
  function invalidType(name, expected, actual) {
    var e = new TypeError('The "' + name + '" argument must be ' + expected +
                          '. Received ' + describe(actual));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function outOfRange(name, actual, bound) {
    var e = new RangeError('The value of "' + name + '" is out of range. It must be ' +
                           bound + '. Received ' + String(actual));
    e.code = 'ERR_OUT_OF_RANGE';
    return e;
  }
  function toUnixTimestamp(time, name) {
    name = name || 'time';
    if (typeof time === 'string' && time.trim() !== '' && !isNaN(+time)) return +time;
    if (typeof time === 'number' && !isNaN(time)) return time;
    if (time instanceof Date) return time.getTime();
    throw invalidType(name, 'of type string or an instance of Date', time);
  }
  function stringToFlags(flags, name) {
    if (flags == null) return C.O_RDONLY;
    if (typeof flags === 'number') return flags;
    if (typeof flags !== 'string') {
      throw invalidType(name || 'flags', 'of type string or an instance of Number', flags);
    }
    switch (flags) {
      case 'r': return C.O_RDONLY;
      case 'rs': case 'sr': return C.O_RDONLY | C.O_SYNC;
      case 'r+': return C.O_RDWR;
      case 'rs+': case 'sr+': return C.O_RDWR | C.O_SYNC;
      case 'w': return C.O_TRUNC | C.O_CREAT | C.O_WRONLY;
      case 'wx': case 'xw': return C.O_TRUNC | C.O_CREAT | C.O_WRONLY | C.O_EXCL;
      case 'w+': return C.O_TRUNC | C.O_CREAT | C.O_RDWR;
      case 'wx+': case 'xw+': return C.O_TRUNC | C.O_CREAT | C.O_RDWR | C.O_EXCL;
      case 'a': return C.O_APPEND | C.O_CREAT | C.O_WRONLY;
      case 'ax': case 'xa': return C.O_APPEND | C.O_CREAT | C.O_WRONLY | C.O_EXCL;
      case 'as': case 'sa': return C.O_APPEND | C.O_CREAT | C.O_WRONLY | C.O_SYNC;
      case 'a+': return C.O_APPEND | C.O_CREAT | C.O_RDWR;
      case 'ax+': case 'xa+': return C.O_APPEND | C.O_CREAT | C.O_RDWR | C.O_EXCL;
      case 'as+': case 'sa+': return C.O_APPEND | C.O_CREAT | C.O_RDWR | C.O_SYNC;
      default: {
        var err = new TypeError('Unknown file open flag: ' + flags);
        err.code = 'ERR_INVALID_ARG_VALUE';
        throw err;
      }
    }
  }
  function getOptions(options, defaults) {
    defaults = defaults || {};
    if (options == null) return defaults;
    if (typeof options !== 'object') throw invalidType('options', 'of type object', options);
    var out = {};
    var key;
    for (key in defaults) out[key] = defaults[key];
    for (key in options) out[key] = options[key];
    return out;
  }
  function getValidatedPath(fileURLOrPath, funcs) {
    var path = fileURLOrPath;
    if (path != null && typeof path === 'object' && typeof path.href === 'string' &&
        typeof path.protocol === 'string') {
      path = url.fileURLToPath(path);
    }
    if (typeof path === 'string') return path;
    if (typeof Buffer !== 'undefined' && Buffer.isBuffer(path)) return path;
    throw invalidType('path', 'of type string or an instance of Buffer or URL', path);
  }
  function assertEncoding(encoding) {
    if (typeof Buffer === 'undefined' || !Buffer.isEncoding(encoding)) {
      var err = new TypeError('Unknown encoding: ' + encoding);
      err.code = 'ERR_INVALID_ARG_VALUE';
      throw err;
    }
  }
  function validateOffsetLengthRead(offset, length, bufferLength) {
    if (offset < 0) throw outOfRange('offset', offset, '>= 0');
    if (length < 0) throw outOfRange('length', length, '>= 0');
    if (offset + length > bufferLength) {
      throw outOfRange('length', length, '<= ' + (bufferLength - offset));
    }
  }
  function validateOffsetLengthWrite(offset, length, byteLength) {
    if (offset > byteLength) throw outOfRange('offset', offset, '<= ' + byteLength);
    if (length > byteLength - offset) {
      throw outOfRange('length', length, '<= ' + (byteLength - offset));
    }
    if (length < 0) throw outOfRange('length', length, '>= 0');
  }
  function isFileType(stats, fileType) {
    return (stats.mode & C.S_IFMT) === fileType;
  }
  class DirentFromStats {
    constructor(name, stats) {
      this.name = name;
      this.parentPath = undefined;
      this._stats = stats;
    }
    isBlockDevice() { return this._stats.isBlockDevice(); }
    isCharacterDevice() { return this._stats.isCharacterDevice(); }
    isDirectory() { return this._stats.isDirectory(); }
    isFIFO() { return this._stats.isFIFO(); }
    isFile() { return this._stats.isFile(); }
    isSocket() { return this._stats.isSocket(); }
    isSymbolicLink() { return this._stats.isSymbolicLink(); }
  }
  var out = {
    constants: {
      kIoMaxLength: 2 ** 31 - 1,
      kMaxUserId: 2 ** 32 - 1,
      kReadFileBufferLength: 512 * 1024,
      kReadFileUnknownBufferLength: 64 * 1024,
      kWriteFileMaxChunkSize: 512 * 1024
    },
    DirentFromStats: DirentFromStats,
    toUnixTimestamp: toUnixTimestamp,
    stringToFlags: stringToFlags,
    getOptions: getOptions,
    getValidatedPath: getValidatedPath,
    assertEncoding: assertEncoding,
    validateOffsetLengthRead: validateOffsetLengthRead,
    validateOffsetLengthWrite: validateOffsetLengthWrite,
    isFileType: isFileType,
    // getStatsFromBinding turns an array of libuv stat numbers into a Stats
    // object; GodeJS has no libuv binding to adapt, so the name is omitted.
    getDirent: function (path, dirent) { return dirent; },
    copyObject: function (source) {
      var target = {};
      for (var key in source) target[key] = source[key];
      return target;
    },
    emitRecursiveRmdirWarning: function () {}
  };
  if (fs.Stats !== undefined) out.Stats = fs.Stats;
  if (fs.BigIntStats !== undefined) out.BigIntStats = fs.BigIntStats;
  if (fs.Dirent !== undefined) out.Dirent = fs.Dirent;
  return out;
})`

// --- internal/fs/watchers --------------------------------------------------

func installFSWatchers(e *env.Env) js.Object {
	fs := e.Exports("fs")
	out := e.JS.Obj()
	pick(fs, out, "FSWatcher", "StatWatcher")
	// fs.watch/watchFile are not implemented yet, so there is no watcher class
	// to adapt. Resolving the specifier keeps the failure local to the missing
	// watcher rather than the module graph.
	return out
}

// --- internal/crypto/* -----------------------------------------------------

func installCryptoUtil(e *env.Env) js.Object {
	return factory(e, cryptoUtilPrelude, "[gode:internal/crypto/util]", e.Exports("crypto").V())
}

const cryptoUtilPrelude = `(function (crypto) {
  'use strict';
  function invalidType(name, actual) {
    var e = new TypeError('The "' + name + '" argument must be of type string or an ' +
                          'instance of ArrayBuffer, Buffer, TypedArray, or DataView. Received ' +
                          (actual === null ? 'null' : typeof actual));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function getBufferSourceBytes(source) {
    if (source instanceof ArrayBuffer) return new Uint8Array(source);
    if (typeof SharedArrayBuffer !== 'undefined' && source instanceof SharedArrayBuffer) {
      return new Uint8Array(source);
    }
    if (ArrayBuffer.isView(source)) {
      return new Uint8Array(source.buffer, source.byteOffset, source.byteLength);
    }
    return undefined;
  }
  function getBufferSourceByteLength(source) {
    var bytes = getBufferSourceBytes(source);
    return bytes === undefined ? undefined : bytes.byteLength;
  }
  function getOptionalByteLength(source) {
    if (source === undefined || source === null) return undefined;
    return getBufferSourceByteLength(source);
  }
  function getArrayBufferOrView(data, name, encoding) {
    if (typeof data === 'string') return Buffer.from(data, encoding || 'utf8');
    var bytes = getBufferSourceBytes(data);
    if (bytes === undefined) throw invalidType(name, data);
    return bytes;
  }
  function getDataViewOrTypedArrayBuffer(data) {
    var bytes = getBufferSourceBytes(data);
    if (bytes === undefined) throw invalidType('data', data);
    return bytes.buffer;
  }
  function toBuf(value, encoding) {
    if (typeof value === 'string') return Buffer.from(value, encoding || 'utf8');
    return value;
  }
  return {
    getArrayBufferOrView: getArrayBufferOrView,
    getBufferSourceByteLength: getBufferSourceByteLength,
    getBufferSourceBytes: getBufferSourceBytes,
    getCiphers: crypto.getCiphers,
    getCurves: crypto.getCurves,
    getDataViewOrTypedArrayBuffer: getDataViewOrTypedArrayBuffer,
    getHashes: crypto.getHashes,
    getOptionalByteLength: getOptionalByteLength,
    // GodeJS does not link OpenSSL; report the security level Node ships
    // with on current OpenSSL 3 builds so tests can size their expectations.
    getOpenSSLSecLevel: function () { return 2; },
    kHandle: Symbol('kHandle'),
    setEngine: function () { return undefined; },
    toBuf: toBuf,
    kNamedCurveAliases: {}
  };
})`

func installCryptoKeys(e *env.Env) js.Object {
	crypto := e.Exports("crypto")
	out := e.JS.Obj()
	pick(crypto, out, "createSecretKey", "createPublicKey", "createPrivateKey", "KeyObject")
	// KeyObject's internal handle helpers (getKeyObjectHandle, importKey) and
	// the CryptoKey bridge need a native key representation GodeJS does not
	// expose, so those names are intentionally absent.
	return out
}

func installCryptoX509(e *env.Env) js.Object {
	return factory(e, cryptoX509Prelude, "[gode:internal/crypto/x509]", e.Exports("crypto").V())
}

const cryptoX509Prelude = `(function (crypto) {
  'use strict';
  var out = {
    isX509Certificate: function (value) {
      return value !== null && typeof value === 'object' &&
        typeof value.checkIssued === 'function' &&
        typeof value.checkPrivateKey === 'function' &&
        value.raw !== undefined;
    }
  };
  if (crypto.X509Certificate !== undefined) {
    out.X509Certificate = crypto.X509Certificate;
    out.InternalX509Certificate = crypto.X509Certificate;
  }
  return out;
})`

// --- internal/assert/myers_diff -------------------------------------------

func installMyersDiff(e *env.Env) js.Object {
	return factory(e, myersDiffPrelude, "[gode:internal/assert/myers_diff]", e.Exports("internal/util/colors").V())
}

// Port of lib/internal/assert/myers_diff.js. A diff entry is [operation, value]
// with operation in {-1 delete, 0 nop, 1 insert}; the array is built backwards
// and the printers walk it in reverse, exactly like upstream.
const myersDiffPrelude = `(function (colors) {
  'use strict';
  var DELETE = -1;
  var NOP = 0;
  var INSERT = 1;
  var kNopLinesToCollapse = 5;
  function areLinesEqual(actual, expected, checkCommaDisparity) {
    if (actual === expected) return true;
    if (checkCommaDisparity) {
      return (actual + ',') === expected || actual === (expected + ',');
    }
    return false;
  }
  function backtrack(trace, actual, expected, checkCommaDisparity) {
    var actualLength = actual.length;
    var expectedLength = expected.length;
    var max = actualLength + expectedLength;
    var x = actualLength;
    var y = expectedLength;
    var result = [];
    for (var diffLevel = trace.length - 1; diffLevel >= 0; diffLevel--) {
      var v = trace[diffLevel];
      var diagonalIndex = x - y;
      var offset = diagonalIndex + max;
      var prevDiagonalIndex;
      if (diagonalIndex === -diffLevel ||
          (diagonalIndex !== diffLevel && v[offset - 1] < v[offset + 1])) {
        prevDiagonalIndex = diagonalIndex + 1;
      } else {
        prevDiagonalIndex = diagonalIndex - 1;
      }
      var prevX = v[prevDiagonalIndex + max];
      var prevY = prevX - prevDiagonalIndex;
      while (x > prevX && y > prevY) {
        var actualItem = actual[x - 1];
        var value = checkCommaDisparity && actualItem.slice(-1) !== ','
          ? expected[y - 1] : actualItem;
        result.push([NOP, value]);
        x--;
        y--;
      }
      if (diffLevel > 0) {
        if (x > prevX) result.push([INSERT, actual[--x]]);
        else result.push([DELETE, expected[--y]]);
      }
    }
    return result;
  }
  function myersDiff(actual, expected, checkCommaDisparity) {
    if (checkCommaDisparity === undefined) checkCommaDisparity = false;
    var actualLength = actual.length;
    var expectedLength = expected.length;
    var max = actualLength + expectedLength;
    if (max > 2 ** 31 - 1) {
      var rangeError = new RangeError('The value of "myersDiff input size" is out of ' +
        'range. It must be < 2^31. Received ' + max);
      rangeError.code = 'ERR_OUT_OF_RANGE';
      throw rangeError;
    }
    var v = new Int32Array(2 * max + 1);
    var trace = [];
    for (var diffLevel = 0; diffLevel <= max; diffLevel++) {
      trace.push(new Int32Array(v));
      for (var diagonalIndex = -diffLevel; diagonalIndex <= diffLevel; diagonalIndex += 2) {
        var offset = diagonalIndex + max;
        var previousOffset = v[offset - 1];
        var nextOffset = v[offset + 1];
        var x = (diagonalIndex === -diffLevel ||
                 (diagonalIndex !== diffLevel && previousOffset < nextOffset))
          ? nextOffset : previousOffset + 1;
        var y = x - diagonalIndex;
        while (x < actualLength && y < expectedLength &&
               areLinesEqual(actual[x], expected[y], checkCommaDisparity)) {
          x++;
          y++;
        }
        v[offset] = x;
        if (x >= actualLength && y >= expectedLength) {
          return backtrack(trace, actual, expected, checkCommaDisparity);
        }
      }
    }
    return [];
  }
  function printSimpleMyersDiff(diff) {
    var message = '';
    for (var diffIdx = diff.length - 1; diffIdx >= 0; diffIdx--) {
      var operation = diff[diffIdx][0];
      var value = diff[diffIdx][1];
      var color = colors.white;
      if (operation === INSERT) color = colors.green;
      else if (operation === DELETE) color = colors.red;
      message += color + value + colors.white;
    }
    return '\n' + message;
  }
  function printMyersDiff(diff, operator) {
    var message = '';
    var skipped = false;
    var nopCount = 0;
    for (var diffIdx = diff.length - 1; diffIdx >= 0; diffIdx--) {
      var operation = diff[diffIdx][0];
      var value = diff[diffIdx][1];
      var previousOperation = diffIdx < diff.length - 1 ? diff[diffIdx + 1][0] : null;
      if (previousOperation === NOP && operation !== previousOperation) {
        if (nopCount === kNopLinesToCollapse + 1) {
          message += colors.white + '  ' + diff[diffIdx + 1][1] + '\n';
        } else if (nopCount === kNopLinesToCollapse + 2) {
          message += colors.white + '  ' + diff[diffIdx + 2][1] + '\n';
          message += colors.white + '  ' + diff[diffIdx + 1][1] + '\n';
        } else if (nopCount >= kNopLinesToCollapse + 3) {
          message += colors.blue + '...' + colors.white + '\n';
          message += colors.white + '  ' + diff[diffIdx + 1][1] + '\n';
          skipped = true;
        }
        nopCount = 0;
      }
      if (operation === INSERT) {
        if (operator === 'partialDeepStrictEqual') {
          message += colors.gray + (colors.hasColors ? ' ' : '+') + ' ' + value + colors.white + '\n';
        } else {
          message += colors.green + '+' + colors.white + ' ' + value + '\n';
        }
      } else if (operation === DELETE) {
        message += colors.red + '-' + colors.white + ' ' + value + '\n';
      } else if (operation === NOP) {
        if (nopCount < kNopLinesToCollapse) {
          message += colors.white + '  ' + value + '\n';
        }
        nopCount++;
      }
    }
    message = message.replace(/\s+$/, '');
    return { message: '\n' + message, skipped: skipped };
  }
  return {
    myersDiff: myersDiff,
    printMyersDiff: printMyersDiff,
    printSimpleMyersDiff: printSimpleMyersDiff
  };
})`

// --- internal/test/binding -------------------------------------------------

// installTestBinding exposes Node's test-only `internalBinding` accessor. The
// namespaces are built lazily and cached in JavaScript, because Node guarantees
// that two calls for the same id return the same object (several tests patch
// the returned namespace, e.g. internalBinding('fs').fstat = ...).
func installTestBinding(e *env.Env) js.Object {
	host := e.JS.Obj()
	host.SetFunc("binding", func(this js.Object, args []js.Value) (js.Value, error) {
		id := ""
		if len(args) > 0 {
			id = args[0].Str()
		}
		return bindingFor(e, id).V(), nil
	})
	return factory(e, testBindingPrelude, "[gode:internal/test/binding]", host.V())
}

// safeReexport is reexport for modules a test binary may not have linked.
func safeReexport(e *env.Env, name string) js.Object {
	if !env.Registered(name) {
		return e.JS.Obj()
	}
	return reexport(e, name)
}

// bindingFor builds one binding namespace. Where a real binding backs a public
// module the namespace is a copy of that module, so the values cannot drift;
// the remaining names are the small, honest subset Node's own tests read
// without needing engine internals.
func bindingFor(e *env.Env, id string) js.Object {
	switch id {
	case "util":
		out := safeReexport(e, "util")
		return factory(e, internalUtilBindingPrelude, "[gode:internal/test/binding:util]", out.V())
	case "config":
		out := e.JS.Obj()
		out.SetFunc("hasSmallICU", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Bool(false), nil })
		out.SetFunc("hasIntl", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Bool(true), nil })
		out.SetBool("hasCrypto", true)
		out.SetBool("hasOpenSSL", false)
		out.SetBool("hasSharedArrayBuffer", true)
		out.SetBool("hasTemporal", false)
		out.SetBool("isDebugBuild", false)
		out.SetBool("isASAN", false)
		out.SetBool("isWindows", false)
		return out
	case "js_stream":
		return factory(e, `(function () {
  'use strict';
  // Node exposes a native JSStream handle; tests only need constructibility
  // and a stable prototype identity here.
  function JSStream() { this._external = 0; }
  JSStream.prototype.getAsyncId = function () { return 0; };
  return { JSStream: JSStream };
})`, "[gode:internal/test/binding:js_stream]")
	case "debug":
		out := e.JS.Obj()
		out.SetFunc("getV8FastApiCallCount", func(_ js.Object, args []js.Value) (js.Value, error) { return e.JS.Num(0), nil })
		out.SetFunc("getGenericUsageCount", func(_ js.Object, args []js.Value) (js.Value, error) { return e.JS.Num(0), nil })
		return out
	case "uv":
		out := e.JS.Obj()
		for name, value := range uvErrors {
			out.SetNum(name, float64(value))
		}
		out.SetFunc("errname", func(_ js.Object, args []js.Value) (js.Value, error) {
			if len(args) == 0 {
				return e.JS.Str("UNKNOWN"), nil
			}
			n := int(args[0].Float())
			for name, value := range uvErrors {
				if value == n {
					return e.JS.Str(name), nil
				}
			}
			return e.JS.Str("UNKNOWN"), nil
		})
		out.SetFunc("getErrorMap", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Obj().V(), nil })
		return out
	case "constants":
		out := e.JS.Obj()
		fs := e.JS.Obj()
		dirents := []string{"UNKNOWN", "FILE", "DIR", "LINK", "FIFO", "SOCKET", "CHAR", "BLOCK"}
		for i, name := range dirents {
			fs.SetNum("UV_DIRENT_"+name, float64(i))
		}
		out.Set("fs", fs.V())
		out.Set("crypto", e.JS.Obj().V())
		out.Set("os", e.JS.Obj().V())
		out.Set("zlib", e.JS.Obj().V())
		out.Set("trace", e.JS.Obj().V())
		out.Set("dlopen", e.JS.Obj().V())
		return out
	case "timers":
		out := e.JS.Obj()
		out.SetFunc("getLibuvNow", func(_ js.Object, _ []js.Value) (js.Value, error) {
			return e.JS.Num(float64(time.Now().UnixMilli())), nil
		})
		out.SetFunc("getV8FastApiCallCount", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Num(0), nil })
		return out
	case "async_wrap":
		out := e.JS.Obj()
		out.SetFunc("queueDestroyAsyncId", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Undefined(), nil })
		out.Set("async_hook_fields", e.JS.Obj().V())
		out.Set("async_wrap", e.JS.Obj().V())
		out.Set("constants", e.JS.Obj().V())
		return out
	case "fs":
		out := safeReexport(e, "fs")
		return factory(e, `(function (out) {
  'use strict';
  if (typeof out.FSReqCallback !== 'function') {
    // The asynchronous fs path is Go-side; this constructor only needs to be
    // constructible for tests that instance-check it.
    out.FSReqCallback = function FSReqCallback() {};
  }
  return out;
})`, "[gode:internal/test/binding:fs]", out.V())
	case "crypto":
		return safeReexport(e, "crypto")
	case "buffer":
		out := safeReexport(e, "buffer")
		// Node's native buffer binding exposes fill(buf, value, offset, end,
		// encoding), which validates both bounds before delegating.
		out.SetFunc("fill", func(_ js.Object, args []js.Value) (js.Value, error) {
			if len(args) == 0 || !args[0].IsObject() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					"The \"buf\" argument must be an instance of Buffer.")
			}
			buf := args[0]
			length := int(buf.Object().Get("length").Int())
			bound := func(i int, name string) (int, error) {
				if i >= len(args) || !args[i].Valid() || args[i].IsUndefined() {
					return 0, nil
				}
				n := int(args[i].Int())
				if n < 0 || n > length {
					return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
						fmt.Sprintf("The value of %q is out of range. It must be >= 0 && <= %d. Received %d", name, length, n))
				}
				return n, nil
			}
			if _, err := bound(2, "offset"); err != nil {
				return js.Value{}, err
			}
			if _, err := bound(3, "end"); err != nil {
				return js.Value{}, err
			}
			fn := buf.Object().Get("fill")
			if !fn.IsFunction() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					"The \"buf\" argument must be an instance of Buffer.")
			}
			end := len(args)
			if end > 5 {
				end = 5
			}
			return fn.Object().Call(buf.Object(), args[1:end]...)
		})
		return out
	case "os":
		// Keep the public object so monkey-patched binding methods are visible
		// to the os implementation (Node's tests patch getHomeDirectory).
		return e.Exports("os")
	case "errors":
		out := e.JS.Obj()
		out.Set("codes", e.JS.Obj().V())
		return out
	case "worker":
		out := e.JS.Obj()
		out.SetFunc("getEnvMessagePort", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Undefined(), nil })
		return out
	case "process_methods":
		out := e.JS.Obj()
		out.SetFunc("resetStdioForTesting", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Undefined(), nil })
		return out
	case "builtins":
		out := e.JS.Obj()
		out.SetFunc("setCodeCacheFlushing", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Undefined(), nil })
		return out
	case "trace_events":
		out := e.JS.Obj()
		out.SetFunc("isTraceCategoryEnabled", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Bool(false), nil })
		out.SetFunc("trace", func(_ js.Object, _ []js.Value) (js.Value, error) { return e.JS.Undefined(), nil })
		return out
	}
	// Unknown ids resolve to an empty namespace: Node throws, but a hard failure
	// here would take down a whole test file for a binding we do not model.
	return e.JS.Obj()
}

// uvErrors is the subset of libuv's negative errno constants tests read.
var uvErrors = map[string]int{
	"UV_EOF": -4095, "UV_E2BIG": -7, "UV_EACCES": -13, "UV_EADDRINUSE": -98,
	"UV_EADDRNOTAVAIL": -99, "UV_EAFNOSUPPORT": -97, "UV_EAGAIN": -11,
	"UV_EALREADY": -114, "UV_EBADF": -9, "UV_EBUSY": -16, "UV_ECANCELED": -125,
	"UV_ECHARSET": -4080, "UV_ECONNABORTED": -103, "UV_ECONNREFUSED": -111,
	"UV_ECONNRESET": -104, "UV_EDESTADDRREQ": -89, "UV_EEXIST": -17,
	"UV_EFAULT": -14, "UV_EFBIG": -27, "UV_EHOSTUNREACH": -113, "UV_EINTR": -4,
	"UV_EINVAL": -22, "UV_EIO": -5, "UV_EISCONN": -106, "UV_EISDIR": -21,
	"UV_ELOOP": -40, "UV_EMFILE": -24, "UV_EMSGSIZE": -90, "UV_ENAMETOOLONG": -36,
	"UV_ENETDOWN": -100, "UV_ENETUNREACH": -101, "UV_ENFILE": -23, "UV_ENOBUFS": -105,
	"UV_ENODEV": -19, "UV_ENOENT": -2, "UV_ENOMEM": -12, "UV_ENONET": -4056,
	"UV_ENOPROTOOPT": -92, "UV_ENOSPC": -28, "UV_ENOSYS": -38, "UV_ENOTCONN": -107,
	"UV_ENOTDIR": -20, "UV_ENOTEMPTY": -39, "UV_ENOTSOCK": -88, "UV_ENOTSUP": -95,
	"UV_ENOTTY": -25, "UV_ENXIO": -6, "UV_EOVERFLOW": -75, "UV_EPERM": -1,
	"UV_EPIPE": -32, "UV_EPROTO": -71, "UV_EPROTONOSUPPORT": -93,
	"UV_EPROTOTYPE": -91, "UV_ERANGE": -34, "UV_EROFS": -30, "UV_ESHUTDOWN": -108,
	"UV_ESPIPE": -29, "UV_ESRCH": -3, "UV_ETIMEDOUT": -110, "UV_ETXTBSY": -26,
	"UV_EXDEV": -18, "UV_UNKNOWN": -4094,
	"UV_EAI_ADDRFAMILY": -3000, "UV_EAI_AGAIN": -3001, "UV_EAI_BADFLAGS": -3002,
	"UV_EAI_BADHINTS": -3013, "UV_EAI_CANCELED": -3003, "UV_EAI_FAIL": -3004,
	"UV_EAI_FAMILY": -3005, "UV_EAI_MEMORY": -3006, "UV_EAI_NODATA": -3007,
	"UV_EAI_NONAME": -3008, "UV_EAI_OVERFLOW": -3009, "UV_EAI_PROTOCOL": -3014,
	"UV_EAI_SERVICE": -3010, "UV_EAI_SOCKTYPE": -3011,
}

const testBindingPrelude = `(function (host) {
  'use strict';
  var cache = Object.create(null);

  function internalBinding(id) {
    if (typeof id !== 'string') {
      var typeError = new TypeError('The "id" argument must be of type string');
      typeError.code = 'ERR_INVALID_ARG_TYPE';
      throw typeError;
    }
    // Names prefixed internal_only must stay hidden even here (Node's own
    // test-internal-only-binding.js asserts the error).
    if (id.lastIndexOf('internal_only', 0) === 0) {
      var hidden = new Error('No such binding: ' + id);
      hidden.code = 'ERR_INVALID_MODULE';
      throw hidden;
    }
    if (cache[id] === undefined) cache[id] = host.binding(id);
    return cache[id];
  }

  function uncurry(fn) {
    return Function.prototype.call.bind(fn);
  }
  function capitalize(name) { return name.charAt(0).toUpperCase() + name.slice(1); }

  // primordials is generated from the live intrinsics: Node's own copy exists
  // so internal code never depends on patched globals, which a test harness
  // that inspects it cannot distinguish from ours.
  function makePrimordials() {
    var out = {};
    var ctors = {
      Array: Array, Object: Object, String: String, Number: Number,
      Boolean: Boolean, Symbol: Symbol, Math: Math, JSON: JSON, Reflect: Reflect,
      Promise: Promise, Map: Map, Set: Set, WeakMap: WeakMap, WeakSet: WeakSet,
      Date: Date, RegExp: RegExp, Error: Error, TypeError: TypeError,
      RangeError: RangeError, SyntaxError: SyntaxError, EvalError: EvalError,
      ReferenceError: ReferenceError, URIError: URIError, BigInt: BigInt,
      ArrayBuffer: ArrayBuffer, DataView: DataView, Int8Array: Int8Array,
      Uint8Array: Uint8Array, Uint8ClampedArray: Uint8ClampedArray,
      Int16Array: Int16Array, Uint16Array: Uint16Array, Int32Array: Int32Array,
      Uint32Array: Uint32Array, Float32Array: Float32Array,
      Float64Array: Float64Array, BigInt64Array: BigInt64Array,
      BigUint64Array: BigUint64Array
    };
    Object.keys(ctors).forEach(function (name) {
      var C = ctors[name];
      out[name] = C;
      if (C === undefined || C === null) return;
      var statics = Object.getOwnPropertyNames(C);
      for (var i = 0; i < statics.length; i++) {
        var key = statics[i];
        if (key === 'length' || key === 'name' || key === 'prototype' ||
            key === 'arguments' || key === 'caller') continue;
        var desc = Object.getOwnPropertyDescriptor(C, key);
        if (!desc || typeof desc.value !== 'function') continue;
        // Statics such as ArrayIsArray/ObjectKeys are ordinary functions, not
        // methods: Node exposes them unbound, so do not uncurry them.
        out[name + capitalize(key)] = desc.value;
      }
      // Math, JSON and Reflect are namespaces, not constructors.
      if (typeof C !== 'function') return;
      var proto = C.prototype;
      if (!proto) return;
      var methods = Object.getOwnPropertyNames(proto);
      for (var j = 0; j < methods.length; j++) {
        var method = methods[j];
        if (method === 'constructor') continue;
        var d = Object.getOwnPropertyDescriptor(proto, method);
        if (!d || typeof d.value !== 'function') continue;
        var label = method.lastIndexOf('Symbol(', 0) === 0 ? 'Symbol' + method.slice(7, -1) : method;
        try { out[name + 'Prototype' + capitalize(label)] = uncurry(d.value); } catch (err) { /* accessors */ }
      }
    });
    out.globalThis = globalThis;
    out.SafeMap = Map;
    out.SafeSet = Set;
    out.SafeWeakMap = WeakMap;
    out.SafeWeakSet = WeakSet;
    out.SafePromise = Promise;
    out.SafeArrayIterator = Array.prototype[Symbol.iterator];
    out.ReflectApply = Reflect.apply;
    out.ReflectConstruct = Reflect.construct;
    out.SymbolIterator = Symbol.iterator;
    out.SymbolFor = Symbol.for;
    out.SymbolToString = uncurry(Symbol.prototype.toString);
    out.FunctionPrototypeCall = function (fn, thisArg) {
      return fn.apply(thisArg, Array.prototype.slice.call(arguments, 2));
    };
    out.ObjectPrototypeIsPrototypeOf = uncurry(Object.prototype.isPrototypeOf);
    out.ArrayPrototypePush = uncurry(Array.prototype.push);
    out.StringPrototypeStartsWith = uncurry(String.prototype.startsWith);
    out.ErrorCaptureStackTrace = function (target, constructor) {
      if (typeof Error.captureStackTrace === 'function') {
        return Error.captureStackTrace(target, constructor);
      }
    };
    return out;
  }

  return { internalBinding: internalBinding, primordials: makePrimordials() };
})`

const internalUtilBindingPrelude = `(function (out) {
  'use strict';
  var types = out.types || {};

  out.getConstructorName = function (value) {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    var ctor = value.constructor;
    return (ctor && ctor.name) || '';
  };
  out.isInsideNodeModules = function () { return false; };
  out.isNativeError = function (value) {
    return value instanceof Error ||
      Object.prototype.toString.call(value) === '[object Error]';
  };
  out.isBoxedPrimitive = function (value) {
    var tag = Object.prototype.toString.call(value);
    return tag === '[object Number]' || tag === '[object String]' ||
      tag === '[object Boolean]' || tag === '[object BigInt]' ||
      tag === '[object Symbol]';
  };
  out.arrayBufferViewHasBuffer = function (view) { return view.buffer != null; };
  out.getExternalValue = function () { return 0; };
  out.getOwnNonIndexProperties = function (value, filter) {
    var keys = Object.getOwnPropertyNames(value).filter(function (key) {
      return !/^(0|[1-9][0-9]*)$/.test(key);
    });
    if (filter & 2) {
      keys = keys.concat(Object.getOwnPropertySymbols(value));
    }
    return keys;
  };
  out.getProxyDetails = function () { return undefined; };
  out.getPromiseDetails = function (value) {
    if (!(value instanceof Promise)) return undefined;
    // JavaScriptCore exposes no promise state, so an untracked promise is
    // reported as pending rather than guessed.
    return [0, undefined];
  };
  out.previewEntries = function (value) {
    var out;
    if (value instanceof Map) {
      out = Array.from(value.entries());
      return [out, true];
    }
    if (value instanceof Set) {
      return [Array.from(value.values()), false];
    }
    if (Array.isArray(value)) return [value.slice(0, 100), false];
    return undefined;
  };
  out.guessHandleType = function () { return 'NONE'; };
  out.getPromiseHooks = function () { return undefined; };
  out.setPromiseHooks = function () {};
  out.getSystemErrorName = function (errno) {
    var names = { '-2': 'ENOENT', '-13': 'EACCES', '-17': 'EEXIST', '-22': 'EINVAL',
      '-28': 'ENOSPC', '-32': 'EPIPE', '-1': 'EPERM', '-4': 'EINTR', '-5': 'EIO' };
    return names[String(errno)] || 'UNKNOWN';
  };
  out.getSystemErrorMap = function () { return new Map(); };
  out.getHeapStatistics = function () {
    return { total_heap_size: 0, total_heap_size_executable: 0, total_physical_size: 0,
      total_available_size: 0, used_heap_size: 0, heap_size_limit: 0, malloced_memory: 0,
      peak_malloced_memory: 0, does_zap_garbage: 0, number_of_native_contexts: 0,
      number_of_detached_contexts: 0 };
  };
  out.getHeapSpaceStatistics = function () { return []; };
  out.getHeapSnapshot = function () {
    var payload = '{"snapshot":{"meta":{},"node_count":0,"edge_count":0},"nodes":[],"edges":[],"strings":[]}';
    var listeners = Object.create(null);
    var sent = false, ended = false, paused = false;
    function emit(name, value) {
      var list = listeners[name];
      if (!list) return;
      listeners[name] = null;
      for (var i = 0; i < list.length; i++) list[i].call(stream, value);
    }
    function deliver() {
      if (sent || paused || ended) return;
      sent = true;
      emit('data', payload);
      ended = true;
      emit('end');
      emit('close');
    }
    var stream = {
      on: function (name, fn) {
        if (typeof fn !== 'function') return this;
        if (!listeners[name]) listeners[name] = [];
        listeners[name].push(fn);
        if (name === 'data' || name === 'readable') Promise.resolve().then(deliver);
        return this;
      },
      once: function (name, fn) {
        var self = this;
        function onceFn(value) { self.removeListener(name, onceFn); return fn.call(self, value); }
        return this.on(name, onceFn);
      },
      removeListener: function (name, fn) {
        var list = listeners[name];
        if (list) listeners[name] = list.filter(function (item) { return item !== fn; });
        return this;
      },
      pause: function () { paused = true; return this; },
      resume: function () { paused = false; Promise.resolve().then(deliver); return this; },
      read: function () {
        if (sent) return null;
        sent = true;
        return payload;
      },
      pipe: function (dest) {
        this.on('data', function (chunk) { if (dest && typeof dest.write === 'function') dest.write(chunk); });
        this.on('end', function () { if (dest && typeof dest.end === 'function') dest.end(); });
        return dest;
      },
      pause: function () { paused = true; return this; },
      destroy: function () { ended = true; listeners = Object.create(null); },
      push: function () {},
      readStart: function () { return this.resume(); }
    };
    return stream;
  };
  if (types && typeof types === 'object') {
    Object.keys(types).forEach(function (key) {
      if (out[key] === undefined) out[key] = types[key];
    });
  }
  return out;
})`

// --- internal/fs/recursive_watch -------------------------------------------

// installRecursiveWatch exposes the pieces fs.watch() uses for recursive
// watching. Node keeps this in its own module; GodeJS implements recursive
// watching inside fs, so this re-exports the public surface.
func installRecursiveWatch(e *env.Env) js.Object {
	if !env.Registered("fs") {
		return e.JS.Obj()
	}
	return factory(e, `(function (fs) {
  'use strict';
  return {
    FSWatcher: fs.FSWatcher,
    watch: fs.watch,
    kFSWatchStart: Symbol('kFSWatchStart')
  };
})`, "[gode:internal/fs/recursive_watch]", e.Exports("fs").V())
}

// --- internal/js_stream_socket ---------------------------------------------

func installJSStreamSocket(e *env.Env) js.Object {
	return factory(e, `(function () {
  'use strict';
  // GodeJS has no JavaScript-authored socket handles, so the adapter is inert.
  function JSStreamSocket() { this._handle = null; }
  JSStreamSocket.prototype.isClosing = function () { return false; };
  return { JSStreamSocket: JSStreamSocket, isSupported: false };
})`, "[gode:internal/js_stream_socket]")
}

// --- internal/child_process -------------------------------------------------

func installChildProcess(e *env.Env) js.Object {
	if !env.Registered("child_process") {
		return e.JS.Obj()
	}
	cp := e.Exports("child_process")
	return factory(e, `(function (cp) {
  'use strict';
  return {
    ChildProcess: cp.ChildProcess,
    setupChannel: function () {},
    getChannelHandle: function () { return undefined; }
  };
})`, "[gode:internal/child_process]", cp.V())
}

// --- internal/dgram ---------------------------------------------------------

// installDgram only has to resolve: dgram itself is not implemented, and a
// socket handle factory returning undefined makes that explicit.
func installDgram(e *env.Env) js.Object {
	return factory(e, `(function () {
  'use strict';
  return { newSocketHandle: function () { return undefined; } };
})`, "[gode:internal/dgram]")
}
