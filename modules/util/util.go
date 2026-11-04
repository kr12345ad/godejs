package util

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/modules/web"
)

func init() {
	env.Register("util", Install)
	env.Register("util/types", func(e *env.Env) js.Object {
		if types := e.Exports("util").Get("types"); types.Valid() && types.IsObject() {
			return types.Object()
		}
		return typesObject(e.JS)
	})
}

// Install builds the util module's exports object.
func Install(e *env.Env) js.Object {
	exports := e.JS.Obj()

	// Node's util re-exports the web-platform codecs. The globals are
	// installed by the web prelude, so make sure they exist even when util
	// is installed on its own (tests, embedders).
	web.Install(e)
	exports.Set("TextEncoder", e.Global().Get("TextEncoder"))
	exports.Set("TextDecoder", e.Global().Get("TextDecoder"))

	exports.SetFunc("inspect", func(this js.Object, args []js.Value) (js.Value, error) {
		target := e.JS.Undefined()
		if len(args) > 0 {
			target = args[0]
		}
		opts := defaultInspectOptions
		if len(args) > 1 {
			opts = parseInspectOptions(args[1])
		}
		return e.JS.Str(inspectWith(target, opts)), nil
	})

	// util.inspect.defaultOptions is live: util.format follows mutations of
	// its numericSeparator field, so read it per call instead of caching the
	// value at install time.
	defaultNumericSeparator := func() bool {
		insp := exports.Get("inspect")
		if !insp.IsFunction() {
			return false
		}
		def := insp.Object().Get("defaultOptions")
		if !def.IsObject() {
			return false
		}
		v := def.Object().Get("numericSeparator")
		return v.IsBool() && v.Bool()
	}

	exports.SetFunc("format", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Str(FormatWith(args, defaultNumericSeparator())), nil
	})

	exports.SetFunc("formatWithOptions", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 {
			return e.JS.Str(""), nil
		}
		sep := defaultNumericSeparator()
		if len(args) > 0 && args[0].IsObject() {
			o := args[0].Object()
			if o.Has("numericSeparator") {
				v := o.Get("numericSeparator")
				if v.IsBool() {
					sep = v.Bool()
				}
			}
		}
		return e.JS.Str(FormatWith(args[1:], sep)), nil
	})

	exports.SetFunc("isDeepStrictEqual", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 {
			return e.JS.Bool(false), nil
		}
		// Node 24's third argument opts out of prototype comparison entirely
		// (including for nested values and typed arrays).
		skipPrototype := len(args) > 2 && args[2].Bool()
		return e.JS.Bool(deepStrictEqual(args[0], args[1], skipPrototype)), nil
	})

	exports.SetFunc("inherits", func(this js.Object, args []js.Value) (js.Value, error) {
		var ctorArg, superArg js.Value
		if len(args) > 0 {
			ctorArg = args[0]
		}
		if len(args) > 1 {
			superArg = args[1]
		}
		invalid := func(name, expected, actual string) error {
			return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "`+name+`" `+expected+` must be of type function. Received `+actual)
		}
		if !ctorArg.IsFunction() {
			return js.Value{}, invalid("ctor", "argument", describe(ctorArg))
		}
		// Node only rejects a null superCtor up front; any other non-function
		// value fails on its missing prototype with the property wording.
		if len(args) < 2 || superArg.IsNull() {
			return js.Value{}, invalid("superCtor", "argument", describe(superArg))
		}
		super := superArg.Object()
		superProto := super.Get("prototype")
		if !superProto.IsObject() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "superCtor.prototype" property must be of type object. Received `+describe(superProto))
		}
		ctor := ctorArg.Object()
		if proto := ctor.Get("prototype"); proto.IsObject() {
			proto.Object().SetProto(superProto.Object())
		}
		// super_ is a non-enumerable data property, as Node defines it.
		desc := e.JS.Obj()
		desc.Set("value", super.Value)
		desc.SetBool("enumerable", false)
		desc.SetBool("configurable", true)
		desc.SetBool("writable", true)
		objectCtor := e.Global().Get("Object")
		if define := objectCtor.Object().Get("defineProperty"); define.IsFunction() {
			if _, err := define.Object().Call(objectCtor.Object(), ctorArg, e.JS.Str("super_"), desc.V()); err != nil {
				return js.Value{}, err
			}
		} else {
			ctor.Set("super_", super.Value)
		}
		return e.JS.Undefined(), nil
	})

	exports.SetFunc("promisify", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsFunction() {
			got := "undefined"
			if len(args) > 0 {
				got = describe(args[0])
			}
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "original" argument must be of type function. Received `+got)
		}
		return promisify(args[0])
	})
	// promisify.custom is the shared symbol under which a function can expose
	// its own promise-returning form (Node: SymbolFor).
	if sym, err := e.JS.Eval(`Symbol.for("nodejs.util.promisify.custom")`, "[gode:util]"); err == nil {
		if pv := exports.Get("promisify"); pv.IsFunction() {
			pv.Object().Set("custom", sym)
		}
	}

	exports.SetFunc("callbackify", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsFunction() {
			got := "undefined"
			if len(args) > 0 {
				got = describe(args[0])
			}
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "original" argument must be of type function. Received `+got)
		}
		return callbackify(e, args[0])
	})

	exports.SetFunc("debuglog", func(this js.Object, args []js.Value) (js.Value, error) {
		section := ""
		if len(args) > 0 {
			section = args[0].Str()
		}
		enabled := debugEnabled(e, section)
		fn := makeFunc(e.JS, "debuglog", func(this js.Object, callArgs []js.Value) (js.Value, error) {
			if !enabled {
				return e.JS.Undefined(), nil
			}
			parts := make([]string, 0, len(callArgs))
			for _, a := range callArgs {
				parts = append(parts, a.Str())
			}
			fmt.Fprintf(e.Err(), "%s %d: %s\n", strings.ToUpper(section), os.Getpid(), strings.Join(parts, " "))
			return e.JS.Undefined(), nil
		})
		fn.Object().SetBool("enabled", enabled)
		return fn, nil
	})

	exports.SetFunc("stripVTControlCharacters", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Str(""), nil
		}
		return e.JS.Str(stripVT(args[0].Str())), nil
	})

	exports.SetFunc("toUSVString", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Str(""), nil
		}
		return e.JS.Str(toUSVString(args[0].Str())), nil
	})

	exports.SetFunc("getSystemErrorName", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsNumber() {
			return js.Value{}, errors.New(`The "err" argument must be of type number`)
		}
		f := args[0].Float()
		if err := checkSystemErrno(f, args); err != nil {
			return js.Value{}, err
		}
		name, ok := errnoNames[int(f)]
		if !ok {
			if f < 0 {
				name, ok = errnoNames[int(-f)]
			}
		}
		if !ok {
			return js.Value{}, fmt.Errorf("Unknown system error %d", int(f))
		}
		return e.JS.Str(name), nil
	})

	exports.Set("types", typesObject(e.JS).V())

	// util.parseEnv parses dotenv-style content (Node 21.7+).
	exports.SetFunc("parseEnv", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsString() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "content" argument must be of type string. Received `+describeArg(args))
		}
		pairs, order, err := ParseEnv(args[0].Str())
		if err != nil {
			return js.Value{}, err
		}
		obj := e.JS.Obj()
		for _, key := range order {
			obj.SetStr(key, pairs[key])
		}
		return obj.V(), nil
	})

	// util.getSystemErrorMessage maps an errno to the libc strerror text.
	exports.SetFunc("getSystemErrorMessage", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsNumber() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "err" argument must be of type number. Received `+describeArg(args))
		}
		f := args[0].Float()
		if err := checkSystemErrno(f, args); err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(systemErrorMessage(int(f))), nil
	})

	// _exceptionWithHostPort builds the errors net/first-class modules throw for
	// socket failures. It is not tracked in the docs baseline, but several
	// builtins and tests call it directly.
	exports.SetFunc("_exceptionWithHostPort", func(this js.Object, args []js.Value) (js.Value, error) {
		var errno int
		if len(args) > 0 && args[0].IsNumber() {
			errno = int(args[0].Float())
		}
		syscallName := ""
		if len(args) > 1 {
			syscallName = args[1].Str()
		}
		var address js.Value
		if len(args) > 2 {
			address = args[2]
		}
		port := 0
		if len(args) > 3 && args[3].IsNumber() {
			port = int(args[3].Float())
		}
		additional := ""
		if len(args) > 4 && args[4].IsString() {
			additional = args[4].Str()
		}
		name, ok := errnoNames[absInt(errno)]
		if !ok {
			name = "UNKNOWN"
		}
		msg := syscallName + " " + name
		hasAddress := address.Valid() && !address.IsNull() && !address.IsUndefined()
		if hasAddress {
			msg += " " + address.Str()
			if port > 0 {
				msg += ":" + strconv.Itoa(port)
			}
		}
		if additional != "" {
			msg += " - " + additional
		}
		obj := e.JS.Error(msg)
		obj.Set("errno", e.JS.Num(float64(errno)))
		obj.SetStr("code", name)
		obj.SetStr("syscall", syscallName)
		if address.Valid() {
			obj.Set("address", address)
		}
		if port > 0 {
			obj.SetNum("port", float64(port))
		}
		if stack := obj.Get("stack"); stack.IsString() {
			obj.SetStr("stack", stripHostPortWrapper(stack.Str()))
		}
		return obj.V(), nil
	})

	// The legacy util.is* predicates are deprecated but still part of the
	// Node 24 surface that older modules call.
	legacy := func(name string, pred func(js.Value) bool) {
		exports.SetFunc(name, func(this js.Object, args []js.Value) (js.Value, error) {
			if len(args) == 0 {
				return e.JS.Bool(false), nil
			}
			return e.JS.Bool(pred(args[0])), nil
		})
	}
	legacy("isArray", func(v js.Value) bool { return v.IsArray() })
	legacy("isBoolean", func(v js.Value) bool { return v.IsBool() })
	legacy("isBuffer", func(v js.Value) bool {
		if !v.IsObject() {
			return false
		}
		// Read the global Buffer rather than forcing the builtin to install:
		// util may be installed by a module that is itself mid-install.
		ctor := e.Global().Get("Buffer")
		if !ctor.IsFunction() {
			return false
		}
		isBuf := ctor.Object().Get("isBuffer")
		if !isBuf.IsFunction() {
			return false
		}
		r, err := isBuf.Object().Call(ctor.Object(), v)
		return err == nil && r.Bool()
	})
	legacy("isDate", func(v js.Value) bool { return v.IsDate() })
	legacy("isError", func(v js.Value) bool {
		if !v.IsObject() {
			return false
		}
		tag := rawTag(v)
		return strings.HasSuffix(tag, "Error") || v.IsError()
	})
	legacy("isFunction", func(v js.Value) bool { return v.IsFunction() })
	legacy("isNull", func(v js.Value) bool { return v.IsNull() })
	legacy("isNullOrUndefined", func(v js.Value) bool { return v.IsNull() || v.IsUndefined() })
	legacy("isNumber", func(v js.Value) bool { return v.IsNumber() })
	legacy("isObject", func(v js.Value) bool { return v.IsObject() })
	legacy("isPrimitive", func(v js.Value) bool { return !v.IsObject() })
	legacy("isRegExp", func(v js.Value) bool { return v.IsObject() && rawTag(v) == "RegExp" })
	legacy("isString", func(v js.Value) bool { return v.IsString() })
	legacy("isSymbol", func(v js.Value) bool { return !v.IsObject() && rawTag(v) == "Symbol" })
	legacy("isUndefined", func(v js.Value) bool { return v.IsUndefined() })

	// The JavaScript half of util: the inspect.custom protocol, the remaining
	// util.types predicates, styleText, deprecate and the aborts/scheduling
	// helpers. Keeping these in one prelude makes the host surface small.
	if impl, err := e.JS.Eval(extraPrelude, "[gode:util:extra]"); err == nil && impl.IsObject() {
		impl.Object().Call(js.Object{}, exports.V())
	} else if err != nil {
		fmt.Fprintf(e.Err(), "util prelude: %v\n", err)
	}
	customInspectFn = exports.Get("inspect").Protect()
	customRender = exports.Get("__renderCustom").Protect()
	jsSymbolEntries = exports.Get("__symbolEntries").Protect()
	exports.Delete("__renderCustom")
	exports.Delete("__symbolEntries")

	// getCallSites needs structured access to the current stack. JSC has no
	// public CallSite API, so parse the (stable) `fn@url:line:col` frames of a
	// freshly created Error, dropping the helper's own frame.
	if impl, err := e.JS.Eval(`(function (exports) {
  'use strict';
  function invalidType(name, expected, actual) {
    var e = new TypeError('The "' + name + '" argument must be ' + expected +
      '. Received ' + (actual === null ? 'null' : typeof actual));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function parseFrame(line) {
    var at = line.indexOf('@');
    if (at === -1) return null;
    var rest = line.slice(at + 1);
    var m = /^(.*):(\d+):(\d+)$/.exec(rest);
    if (!m) return null;
    return {
      functionName: line.slice(0, at),
      scriptName: m[1],
      lineNumber: Number(m[2]),
      columnNumber: Number(m[3]),
      scriptId: '0',
    };
  }
  function getCallSites(frameCount, options) {
    if (frameCount === undefined) frameCount = 10;
    if (options === undefined && frameCount !== null && typeof frameCount === 'object') {
      options = frameCount;
      frameCount = 10;
    }
    if (options === undefined) options = {};
    if (options === null || typeof options !== 'object' || Array.isArray(options)) {
      throw invalidType('options', 'of type object', options);
    }
    if (options.sourceMap !== undefined && typeof options.sourceMap !== 'boolean') {
      throw invalidType('options.sourceMap', 'of type boolean', options.sourceMap);
    }
    if (typeof frameCount !== 'number' || !isFinite(frameCount) ||
        Math.floor(frameCount) !== frameCount || frameCount < 1 || frameCount > 200) {
      var e = new RangeError('The value of "frameCount" is out of range. It must be an integer >= 1 && <= 200. Received ' + frameCount);
      e.code = 'ERR_OUT_OF_RANGE';
      throw e;
    }
    var lines = String(new Error().stack || '').split('\n');
    var out = [];
    var seenHelper = false;
    for (var i = 0; i < lines.length && out.length < frameCount; i++) {
      var frame = parseFrame(lines[i]);
      if (!frame) continue;
      if (!seenHelper) {
        seenHelper = true;
        if (frame.functionName === 'getCallSites' || /^\[gode:/.test(frame.scriptName)) {
          continue;
        }
      }
      if (/^\[gode:/.test(frame.scriptName)) continue;
      out.push(frame);
    }
    return out;
  }
  exports.getCallSites = getCallSites;
  exports.__rawStack = function () { return String(new Error().stack || 'EMPTY'); };	})`, "[gode:util]"); err == nil && impl.IsObject() {
		impl.Object().Call(js.Object{}, exports.V())
	} else if err != nil {
		fmt.Fprintf(e.Err(), "util prelude: %v\n", err)
	}
	// The remaining Node 24 surface: MIME types, diff, debounce/throttle and
	// the small host helpers that back them.
	installExtra(e, exports)
	return exports
}

// jsSymbolEntries lists an object's enumerable symbol keys for the inspector.
var jsSymbolEntries js.Value

// extraPrelude installs the JavaScript-only parts of util. It is written as
// one prelude so the host stays responsible for I/O and values only.
const extraPrelude = `(function (exports) {
  'use strict';
  var customSymbol = Symbol.for('nodejs.util.inspect.custom');
  exports.custom = customSymbol;
  if (typeof exports.inspect === 'function') {
    exports.inspect.custom = customSymbol;
    // Node exposes the default options as a mutable object. The host
    // renderer reads its own defaults, so setting fields here is inert but
    // keeps feature probes and option merging code working.
    exports.inspect.defaultOptions = {
      breakLength: 128,
      colors: false,
      compact: 3,
      depth: 2,
      maxArrayLength: 100,
      maxStringLength: 10000,
      showHidden: false,
      showProxy: false,
      sorted: false,
      getters: false,
      numericSeparator: false
    };
  }

  function describe(v) {
    if (v === undefined) return 'undefined';
    if (v === null) return 'null';
    var t = typeof v;
    if (t === 'number') return 'type number (' + String(v) + ')';
    if (t === 'string') return 'type string (' + JSON.stringify(v) + ')';
    if (t === 'boolean') return 'type boolean (' + String(v) + ')';
    if (t === 'bigint') return 'type bigint (' + String(v) + 'n)';
    if (t === 'symbol') return 'type symbol (' + String(v) + ')';
    if (t === 'function') return 'function ' + (v.name || '(anonymous)');
    if (Array.isArray(v)) return 'an instance of Array';
    var name = v.constructor && v.constructor.name ? v.constructor.name : 'Object';
    return 'an instance of ' + name;
  }
  function invalidType(name, expected, actual) {
    var e = new TypeError('The "' + name + '" argument must be ' + expected +
      '. Received ' + describe(actual));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function invalidValue(name, value, reason) {
    var e = new TypeError('The argument \'' + name + '\' ' + reason +
      '. Received ' + describe(value));
    e.code = 'ERR_INVALID_ARG_VALUE';
    return e;
  }
  function outOfRange(name, actual, bound) {
    var e = new RangeError('The value of "' + name + '" is out of range. It must be ' +
      bound + '. Received ' + actual);
    e.code = 'ERR_OUT_OF_RANGE';
    return e;
  }

  // --- inspect.custom ----------------------------------------------------
  function symbolEntries(value) {
    var syms = Object.getOwnPropertySymbols(value);
    var out = [];
    for (var i = 0; i < syms.length; i++) {
      if (Object.prototype.propertyIsEnumerable.call(value, syms[i])) {
        // The symbol itself is kept so equality can distinguish two symbols
        // with the same description.
        out.push([String(syms[i]), Reflect.get(value, syms[i]), syms[i]]);
      }
    }
    return out;
  }

  function renderCustom(value, depth, options, inspect) {
    if (value === null || typeof value !== 'object') return { hit: false };
    var hook = value[customSymbol];
    if (typeof hook !== 'function' || hook === inspect) return { hit: false };
    var ctor = Object.getOwnPropertyDescriptor(value, 'constructor');
    if (ctor && ctor.value && ctor.value.prototype === value) return { hit: false };
    var text = hook.call(value, depth, options, inspect);
    if (typeof text !== 'string') return { hit: false };
    return { hit: true, text: text };
  }

  exports.__symbolEntries = symbolEntries;
  exports.__renderCustom = renderCustom;

  // --- util.types --------------------------------------------------------
  var types = exports.types;
  if (types) {
    var tag = function (v) { return Object.prototype.toString.call(v); };
    var def = function (name, fn) { types[name] = fn; };
    def('isAnyArrayBuffer', function (v) {
      return v instanceof ArrayBuffer ||
        (typeof SharedArrayBuffer === 'function' && v instanceof SharedArrayBuffer);
    });
    def('isSharedArrayBuffer', function (v) {
      return typeof SharedArrayBuffer === 'function' && v instanceof SharedArrayBuffer;
    });
    def('isArrayBufferView', function (v) { return ArrayBuffer.isView(v); });
    def('isDataView', function (v) { return v instanceof DataView; });
    def('isArgumentsObject', function (v) { return tag(v) === '[object Arguments]'; });
    def('isGeneratorObject', function (v) { return tag(v) === '[object Generator]'; });
    def('isMapIterator', function (v) { return tag(v) === '[object Map Iterator]'; });
    def('isSetIterator', function (v) { return tag(v) === '[object Set Iterator]'; });
    def('isNumberObject', function (v) { return tag(v) === '[object Number]'; });
    def('isStringObject', function (v) { return tag(v) === '[object String]'; });
    def('isBooleanObject', function (v) { return tag(v) === '[object Boolean]'; });
    def('isBigIntObject', function (v) { return tag(v) === '[object BigInt]'; });
    def('isSymbolObject', function (v) { return tag(v) === '[object Symbol]'; });
    def('isWeakMap', function (v) { return v instanceof WeakMap; });
    def('isWeakSet', function (v) { return v instanceof WeakSet; });
    def('isUint8ClampedArray', function (v) { return tag(v) === '[object Uint8ClampedArray]'; });
    def('isFloat16Array', function (v) { return tag(v) === '[object Float16Array]'; });
    def('isInt8Array', function (v) { return tag(v) === '[object Int8Array]'; });
    def('isInt16Array', function (v) { return tag(v) === '[object Int16Array]'; });
    def('isInt32Array', function (v) { return tag(v) === '[object Int32Array]'; });
    def('isUint16Array', function (v) { return tag(v) === '[object Uint16Array]'; });
    def('isUint32Array', function (v) { return tag(v) === '[object Uint32Array]'; });
    def('isFloat32Array', function (v) { return tag(v) === '[object Float32Array]'; });
    def('isFloat64Array', function (v) { return tag(v) === '[object Float64Array]'; });
    def('isModuleNamespaceObject', function (v) { return tag(v) === '[object Module]'; });
    def('isExternal', function () { return false; });
    def('isProxy', function () { return false; });
    def('isKeyObject', function (v) {
      if (v === null || typeof v !== 'object') return false;
      // KeyObjects register their id in a WeakMap behind this prototype method;
      // an object merely inheriting the prototype has no registered id.
      var own = v.$godeOwnId;
      if (typeof own !== 'function') return false;
      try { return typeof own.call(v) === 'number'; } catch (e) { return false; }
    });
    def('isCryptoKey', function (v) { return tag(v) === '[object CryptoKey]'; });
  }

  // --- util.inspect.colors + util.styleText ------------------------------
  var ESC = '\u001b[';
  var colors = {
    reset: [0, 0], bold: [1, 22], dim: [2, 22], italic: [3, 23],
    underline: [4, 24], blink: [5, 25], inverse: [7, 27], hidden: [8, 28],
    strikethrough: [9, 29], doubleunderline: [21, 24],
    black: [30, 39], red: [31, 39], green: [32, 39], yellow: [33, 39],
    blue: [34, 39], magenta: [35, 39], cyan: [36, 39], white: [37, 39],
    bgBlack: [40, 49], bgRed: [41, 49], bgGreen: [42, 49], bgYellow: [43, 49],
    bgBlue: [44, 49], bgMagenta: [45, 49], bgCyan: [46, 49], bgWhite: [47, 49],
    framed: [51, 54], overlined: [53, 55],
    gray: [90, 39], redBright: [91, 39], greenBright: [92, 39],
    yellowBright: [93, 39], blueBright: [94, 39], magentaBright: [95, 39],
    cyanBright: [96, 39], whiteBright: [97, 39],
    bgGray: [100, 49], bgRedBright: [101, 49], bgGreenBright: [102, 49],
    bgYellowBright: [103, 49], bgBlueBright: [104, 49], bgMagentaBright: [105, 49],
    bgCyanBright: [106, 49], bgWhiteBright: [107, 49]
  };
  var aliases = {
    faint: 'dim', swapcolors: 'inverse', swapColors: 'inverse',
    conceal: 'hidden', strikeThrough: 'strikethrough', crossedout: 'strikethrough',
    crossedOut: 'strikethrough', doubleUnderline: 'doubleunderline',
    grey: 'gray', blackBright: 'gray', bgGrey: 'bgGray', bgBlackBright: 'bgGray'
  };
  function styleOf(key) {
    if (typeof key !== 'string') return undefined;
    return colors[key] || colors[aliases[key]];
  }
  if (typeof exports.inspect === 'function') exports.inspect.colors = colors;

  function shouldColorize(stream) {
    var env = (typeof process !== 'undefined' && process.env) || {};
    if (env.FORCE_COLOR !== undefined && env.FORCE_COLOR !== '0') return true;
    if (env.NODE_DISABLE_COLORS !== undefined) return false;
    if (env.NO_COLOR !== undefined) return false;
    if (typeof process !== 'undefined' && process.stdout && stream === process.stdout) {
      return process.stdout.isTTY === true;
    }
    return !!(stream && stream.isTTY === true);
  }
  // Node's replaceCloseCode never rewrites a close sequence that ends the
  // string: an inner style's final reset must stay in place, and the outer
  // style then contributes its own close. Rewriting it would leave the
  // reopened style without a matching reset (visible in nested styleText).
  function replaceClose(text, closeCode, openCode, keepClose) {
    var index = text.indexOf(closeCode);
    if (index === -1) return text;
    var replacement = keepClose ? closeCode + openCode : openCode;
    var result = '';
    var lastIndex = 0;
    do {
      var afterClose = index + closeCode.length;
      if (afterClose >= text.length) break;
      result += text.slice(lastIndex, index) + replacement;
      lastIndex = afterClose;
      index = text.indexOf(closeCode, afterClose);
    } while (index !== -1);
    return result + text.slice(lastIndex);
  }
  function styleText(format, text, options) {
    var validateStream = options && options.validateStream === false ? false : true;
    if (typeof text !== 'string') throw invalidType('text', 'of type string', text);
    var formatArray = Array.isArray(format) ? format : [format];
    var codes = [];
    for (var i = 0; i < formatArray.length; i++) {
      var key = formatArray[i];
      if (key === 'none') continue;
      if (typeof key === 'string' && key[0] === '#') {
        if (!/^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/.test(key)) {
          throw invalidValue('format', key, 'must be a valid hex color (#RGB or #RRGGBB)');
        }
        codes.push(['hex', key]);
        continue;
      }
      var style = styleOf(key);
      if (!style) {
        throw invalidValue('format', key, 'must be one of: ' +
          Object.keys(colors).concat(Object.keys(aliases)).join(', '));
      }
      codes.push(['code', style[0], style[1]]);
    }
    var skipColorize = false;
    if (validateStream) {
      var stream = (options && options.stream) ||
        (typeof process !== 'undefined' ? process.stdout : undefined);
      if (!stream || typeof stream !== 'object' ||
          (stream.isTTY === undefined && typeof stream.write !== 'function')) {
        throw invalidType('stream', ['ReadableStream', 'WritableStream', 'Stream'], stream);
      }
      skipColorize = !shouldColorize(stream);
    }
    var openCodes = '';
    var closeCodes = '';
    var processed = text;
    for (var j = 0; j < codes.length; j++) {
      var entry = codes[j];
      if (entry[0] === 'hex') {
        var hex = entry[1].slice(1);
        if (hex.length === 3) hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
        var r = parseInt(hex.slice(0, 2), 16);
        var g = parseInt(hex.slice(2, 4), 16);
        var b = parseInt(hex.slice(4, 6), 16);
        var hexOpen = ESC + '38;2;' + r + ';' + g + ';' + b + 'm';
        var hexClose = ESC + '39m';
        openCodes += hexOpen;
        closeCodes = hexClose + closeCodes;
        processed = replaceClose(processed, hexClose, hexOpen, false);
        continue;
      }
      var openSeq = ESC + entry[1] + 'm';
      var closeSeq = ESC + entry[2] + 'm';
      openCodes += openSeq;
      closeCodes = closeSeq + closeCodes;
      processed = replaceClose(processed, closeSeq, openSeq, entry[2] === 22);
    }
    if (skipColorize) return text;
    return openCodes + processed + closeCodes;
  }
  exports.styleText = styleText;

  // --- util.deprecate ----------------------------------------------------
  var emittedCodes = Object.create(null);
  exports.deprecate = function deprecate(fn, msg, code, options) {
    if (typeof fn !== 'function') throw invalidType('fn', 'of type function', fn);
    if (msg === undefined) msg = 'Deprecated';
    if (typeof msg !== 'string') throw invalidType('msg', 'of type string', msg);
    if (code !== undefined && typeof code !== 'string') {
      throw invalidType('code', 'of type string', code);
    }
    var modifyPrototype = options === undefined || options.modifyPrototype !== false;
    var emitted = false;
    function wrapped() {
      if (!emitted && !(code !== undefined && emittedCodes[code])) {
        emitted = true;
        if (code !== undefined) emittedCodes[code] = true;
        if (typeof process !== 'undefined' && typeof process.emitWarning === 'function') {
          process.emitWarning(msg, 'DeprecationWarning', code, wrapped);
        }
      }
      return fn.apply(this, arguments);
    }
    Object.defineProperty(wrapped, 'name', { value: fn.name, configurable: true });
    Object.defineProperty(wrapped, 'length', { value: fn.length, configurable: true });
    if (modifyPrototype && fn.prototype !== undefined) {
      wrapped.prototype = fn.prototype;
      Object.setPrototypeOf(wrapped, fn);
    }
    return wrapped;
  };

  // --- util.sleep --------------------------------------------------------
  exports.sleep = function sleep(msec) {
    if (typeof msec !== 'number') throw invalidType('msec', 'of type number', msec);
    if (!isFinite(msec) || msec < 0 || msec > 4294967295 ||
        Math.floor(msec) !== msec) {
      throw outOfRange('msec', msec, '>= 0 && <= 4294967295');
    }
    return new Promise(function (resolve) { setTimeout(resolve, msec); });
  };

  // --- abort helpers -----------------------------------------------------
  function abortError(cause) {
    var e = new Error('The operation was aborted');
    e.name = 'AbortError';
    e.code = 'ABORT_ERR';
    if (cause !== undefined) e.cause = cause;
    return e;
  }
  exports.aborted = function aborted(signal, resource) {
    if (signal === null || signal === undefined ||
        typeof signal.addEventListener !== 'function') {
      throw invalidType('signal', 'an instance of AbortSignal', signal);
    }
    return new Promise(function (resolve, reject) {
      function onAbort() {
        if (resource && typeof resource.close === 'function') {
          try { resource.close(); } catch (e) {}
        }
        reject(abortError(signal.reason));
      }
      if (signal.aborted) return onAbort();
      signal.addEventListener('abort', onAbort, { once: true });
    });
  };
  function markTransferable(proto) {
    try { Object.defineProperty(proto, Symbol.for('nodejs.transferable'), { value: true }); } catch (e) {}
  }
  exports.transferableAbortController = function transferableAbortController() {
    var c = new AbortController();
    markTransferable(Object.getPrototypeOf(c));
    markTransferable(Object.getPrototypeOf(c.signal));
    return c;
  };
  exports.transferableAbortSignal = function transferableAbortSignal(signal) {
    if (signal === null || signal === undefined ||
        typeof signal.addEventListener !== 'function') {
      throw invalidType('signal', 'an instance of AbortSignal', signal);
    }
    markTransferable(Object.getPrototypeOf(signal));
    return signal;
  };
})`

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"util._extend",
		"util.aborted",
		"util.callbackify",
		"util.convertProcessSignalToExitCode",
		"util.debounce",
		"util.debug",
		"util.debuglog",
		"util.deprecate",
		"util.diff",
		"util.format",
		"util.formatWithOptions",
		"util.getSystemErrorMap",
		"util.getCallSites",
		"util.getSystemErrorMessage",
		"util.getSystemErrorName",
		"util.parseArgs",
		"util.parseEnv",
		"util.isArray",
		"util.inherits",
		"util.inspect",
		"util.MIMEParams",
		"util.MIMEParams#[Symbol.iterator]",
		"util.MIMEParams#delete",
		"util.MIMEParams#entries",
		"util.MIMEParams#get",
		"util.MIMEParams#has",
		"util.MIMEParams#keys",
		"util.MIMEParams#set",
		"util.MIMEParams#values",
		"util.MIMEType",
		"util.MIMEType#parse",
		"util.MIMEType#toJSON",
		"util.MIMEType#toString",
		"util.MIMEType.essence",
		"util.MIMEType.params",
		"util.MIMEType.subtype",
		"util.MIMEType.type",
		"util.markPromiseAsHandled",
		"util.isDeepStrictEqual",
		"util.promisify",
		"util.setTraceSigInt",
		"util.stripVTControlCharacters",
		"util.throttle",
		"util.styleText",
		"util.transferableAbortController",
		"util.transferableAbortSignal",
		"util.TextDecoder",
		"util.TextDecoder#decode",
		"util.TextDecoder.encoding",
		"util.TextDecoder.fatal",
		"util.TextDecoder.ignoreBOM",
		"util.TextEncoder",
		"util.TextEncoder#encode",
		"util.TextEncoder#encodeInto",
		"util.TextEncoder.encoding",
		"util.toUSVString",
		"util.types",
	}
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// stripHostPortWrapper drops the _exceptionWithHostPort frame from the stack;
// Node uses Error.captureStackTrace to achieve the same result.
func stripHostPortWrapper(stack string) string {
	lines := strings.Split(stack, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if strings.Contains(line, "_exceptionWithHostPort") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// makeFunc turns a Go host function into a JavaScript function value.
func makeFunc(c *js.Context, name string, f js.Func) js.Value {
	holder := c.Obj()
	holder.SetFunc(name, f)
	return holder.Get(name)
}

func debugEnabled(e *env.Env, section string) bool {
	for _, entry := range e.Environ() {
		if !strings.HasPrefix(entry, "NODE_DEBUG=") {
			continue
		}
		for _, name := range strings.FieldsFunc(strings.TrimPrefix(entry, "NODE_DEBUG="), func(r rune) bool { return r == ',' || r == ' ' }) {
			if strings.EqualFold(name, section) || name == "*" {
				return true
			}
		}
	}
	return false
}

func typesObject(c *js.Context) js.Object {
	t := c.Obj()
	simple := func(name string, pred func(js.Value) bool) {
		t.SetFunc(name, func(this js.Object, args []js.Value) (js.Value, error) {
			if len(args) == 0 {
				return c.Bool(false), nil
			}
			return c.Bool(pred(args[0])), nil
		})
	}
	simple("isDate", func(v js.Value) bool { return v.IsDate() })
	simple("isRegExp", func(v js.Value) bool { return v.Tag() == "RegExp" })
	simple("isMap", func(v js.Value) bool { return v.Tag() == "Map" })
	simple("isSet", func(v js.Value) bool { return v.Tag() == "Set" })
	simple("isPromise", func(v js.Value) bool { return v.Tag() == "Promise" })
	simple("isTypedArray", func(v js.Value) bool { return typedArrayTags[v.Tag()] })
	simple("isUint8Array", func(v js.Value) bool { return v.Tag() == "Uint8Array" })
	simple("isNativeError", func(v js.Value) bool { return v.IsError() })
	// KeyObject has no own symbol brand we can see from another module, so
	// recognise the public shape crypto.KeyObject exposes.
	simple("isKeyObject", func(v js.Value) bool {
		if !v.IsObject() {
			return false
		}
		o := v.Object()
		if cn := o.Get("constructor"); cn.IsFunction() && cn.Object().Get("name").Str() == "KeyObject" {
			return true
		}
		typ := o.Get("type")
		if !typ.IsString() {
			return false
		}
		switch typ.Str() {
		case "secret", "private", "public":
			return o.Get("export").IsFunction()
		}
		return false
	})
	simple("isAsyncFunction", func(v js.Value) bool { return rawTag(v) == "AsyncFunction" })
	simple("isGeneratorFunction", func(v js.Value) bool { return rawTag(v) == "GeneratorFunction" })
	simple("isArrayBuffer", func(v js.Value) bool { return v.Tag() == "ArrayBuffer" })
	simple("isBigInt64Array", func(v js.Value) bool { return v.Tag() == "BigInt64Array" })
	simple("isBigUint64Array", func(v js.Value) bool { return v.Tag() == "BigUint64Array" })
	simple("isBoxedPrimitive", func(v js.Value) bool {
		if !v.IsObject() {
			return false
		}
		switch v.Tag() {
		case "Boolean", "Number", "String", "Symbol", "BigInt":
			return true
		}
		return false
	})
	return t
}

// rawTag always asks Object.prototype.toString; js.Value.Tag reports
// "Function" for callables, which would hide AsyncFunction/GeneratorFunction.
func rawTag(v js.Value) string {
	if !v.Valid() {
		return ""
	}
	fn := v.Ctx().Global().Get("Object").Object().Get("prototype").Object().Get("toString")
	if !fn.IsFunction() {
		return v.Tag()
	}
	r, err := fn.Object().Call(js.Object{Value: v})
	if err != nil {
		return v.Tag()
	}
	tag := r.Str()
	tag = strings.TrimPrefix(tag, "[object ")
	tag = strings.TrimSuffix(tag, "]")
	return tag
}

// vtSequence is Node's ansi-regex (lib/internal/util/inspect.js): an OSC
// sequence with BEL, ESC-backslash or C1 ST termination, or a CSI sequence.
var vtSequence = regexp.MustCompile(`(?:\x1b\](?s:.*?)(?:\x07|\x1b\\|\x9c))|[\x1b\x9b][\[\]()#;?]*(?:[0-9]{1,4}(?:[;:][0-9]{0,4})*)?[0-9A-PR-TZcf-nq-uy=><~]`)

func stripVT(s string) string {
	if !strings.ContainsAny(s, "\x1b\x9b") {
		return s
	}
	return vtSequence.ReplaceAllString(s, "")
}

// toUSVString replaces lone surrogates with U+FFFD. JavaScript strings are
// converted through UTF-16 already, so replacements have happened by here.
func toUSVString(s string) string { return s }

var errnoNames = map[int]string{
	1:   "EPERM",
	2:   "ENOENT",
	3:   "ESRCH",
	4:   "EINTR",
	5:   "EIO",
	6:   "ENXIO",
	7:   "E2BIG",
	8:   "ENOEXEC",
	9:   "EBADF",
	10:  "ECHILD",
	11:  "EAGAIN",
	12:  "ENOMEM",
	13:  "EACCES",
	14:  "EFAULT",
	16:  "EBUSY",
	17:  "EEXIST",
	18:  "EXDEV",
	19:  "ENODEV",
	20:  "ENOTDIR",
	21:  "EISDIR",
	22:  "EINVAL",
	23:  "ENFILE",
	24:  "EMFILE",
	25:  "ENOTTY",
	26:  "ETXTBSY",
	27:  "EFBIG",
	28:  "ENOSPC",
	29:  "ESPIPE",
	30:  "EROFS",
	31:  "EMLINK",
	32:  "EPIPE",
	33:  "EDOM",
	34:  "ERANGE",
	39:  "ENOTEMPTY",
	40:  "ELOOP",
	95:  "EOPNOTSUPP",
	98:  "EADDRINUSE",
	99:  "EADDRNOTAVAIL",
	103: "ECONNABORTED",
	104: "ECONNRESET",
	107: "ENOTCONN",
	110: "ETIMEDOUT",
	111: "ECONNREFUSED",
}

// checkSystemErrno mirrors util's validateNumber plus the safe-integer range
// check both getSystemErrorMessage and getSystemErrorName perform.
func checkSystemErrno(f float64, args []js.Value) error {
	if math.Trunc(f) != f || f >= 0 || f > 9007199254740991 || f < -9007199254740991 {
		return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			`The value of "err" is out of range. It must be a negative integer. Received `+
				groupedNumber(f, renderValue(args)))
	}
	return nil
}

// groupedNumber reproduces the digit separators Node prints for large
// integers in range errors (100_000_000_000_000_000_000).
func groupedNumber(f float64, fallback string) string {
	if math.Trunc(f) != f || math.Abs(f) < 10000 {
		return fallback
	}
	digits := strconv.FormatInt(int64(f), 10)
	if f < 0 {
		digits = digits[1:]
	}
	var out []byte
	for i, c := range []byte(digits) {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, '_')
		}
		out = append(out, c)
	}
	if f < 0 {
		return "-" + string(out)
	}
	return string(out)
}

// renderValue prints a value the way Node's ERR_OUT_OF_RANGE messages do:
// numbers appear verbatim (13, 1.5, 100000000000000000000) rather than with
// the "type number (13)" wording used by type errors.
func renderValue(args []js.Value) string {
	if len(args) == 0 {
		return "undefined"
	}
	v := args[0]
	if v.IsNumber() {
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	return describe(v)
}
