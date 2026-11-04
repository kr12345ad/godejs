package util

import (
	"encoding/json"
	"fmt"
	"sort"

	"godejs/internal/env"
	"godejs/internal/js"
)

// installExtra adds the parts of Node 24's util that live above the host:
// the MIME classes, diff, debounce/throttle and a handful of small helpers.
// Keeping them in one prelude leaves the Go side responsible for values and
// the errno table only.
func installExtra(e *env.Env, exports js.Object) {
	installErrMap(e, exports)
	installSignalExitCode(e, exports)

	// util._extend is the deprecated Object.assign-alike.
	exports.SetFunc("_extend", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Undefined(), nil
		}
		target := args[0]
		if len(args) > 1 && target.IsObject() && args[1].IsObject() && !args[1].IsNull() {
			src := args[1].Object()
			dst := target.Object()
			for _, key := range src.OwnEnumerableKeys() {
				dst.Set(key, src.Get(key))
			}
		}
		return target, nil
	})

	// util.debug is the same function object as util.debuglog in Node.
	if v := exports.Get("debuglog"); v.Valid() {
		exports.Set("debug", v)
	}

	// markPromiseAsHandled only suppresses the runtime's unhandled-rejection
	// reporting; our runtime does not raise one, so validating the argument is
	// the whole contract.
	exports.SetFunc("markPromiseAsHandled", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || args[0].Tag() != "Promise" {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "promise" argument must be an instance of Promise. Received `+describeArg(args))
		}
		return e.JS.Undefined(), nil
	})

	// setTraceSigInt toggles SIGINT tracing in Node; there is no tracing
	// backend here, but the argument contract still applies.
	exports.SetFunc("setTraceSigInt", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsBool() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "enable" argument must be of type boolean. Received `+describeArg(args))
		}
		return e.JS.Undefined(), nil
	})

	installParseArgs(e, exports)

	impl, err := e.JS.Eval(extraSurfacePrelude, "[gode:util:extra]")
	if err != nil {
		fmt.Fprintf(e.Err(), "util extra prelude: %v\n", err)
		return
	}
	if impl.IsObject() {
		if _, callErr := impl.Object().Call(js.Object{}, exports.V()); callErr != nil {
			fmt.Fprintf(e.Err(), "util extra prelude: %v\n", callErr)
		}
	}
}

// installErrMap exposes util.getSystemErrorMap(): a Map from positive errno to
// [name, message]. Node builds it from libuv's table; our errnoNames is the
// same set of names, and the messages come from libc through errno.Error.
func installErrMap(e *env.Env, exports js.Object) {
	factory, err := e.JS.Eval(`(function (json) {
  var rows = JSON.parse(json);
  var map = new Map();
  for (var i = 0; i < rows.length; i++) map.set(rows[i][0], [rows[i][1], rows[i][2]]);
  return map;
})`, "[gode:util:errmap]")
	if err != nil {
		return
	}
	// The factory is reachable only from Go between calls, so pin it.
	factory = factory.Protect()

	keys := make([]int, 0, len(errnoNames))
	for n := range errnoNames {
		keys = append(keys, n)
	}
	sort.Ints(keys)
	rows := make([][3]any, 0, len(keys))
	for _, n := range keys {
		rows = append(rows, [3]any{n, errnoNames[n], systemErrorMessage(n)})
	}
	data, err := json.Marshal(rows)
	if err != nil {
		return
	}
	payload := string(data)

	var cached js.Value
	exports.SetFunc("getSystemErrorMap", func(this js.Object, args []js.Value) (js.Value, error) {
		if cached.Valid() {
			return cached, nil
		}
		v, callErr := factory.Object().Call(js.Object{}, e.JS.Str(payload))
		if callErr != nil {
			return js.Value{}, callErr
		}
		cached = v.Protect()
		return cached, nil
	})
}

// installSignalExitCode implements util.convertProcessSignalToExitCode: POSIX
// reports signal deaths as 128 + the signal number. The accepted names are the
// keys of os.constants.signals, which is what Node validates against too.
func installSignalExitCode(e *env.Env, exports js.Object) {
	exports.SetFunc("convertProcessSignalToExitCode", func(this js.Object, args []js.Value) (js.Value, error) {
		invalid := func() error {
			return js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
				"The argument 'signalCode' is invalid. Received "+describeArg(args))
		}
		if len(args) == 0 || !args[0].IsString() {
			return js.Value{}, invalid()
		}
		signals := e.Exports("os").Get("constants").Object().Get("signals")
		if !signals.Valid() || !signals.IsObject() {
			return js.Value{}, invalid()
		}
		number := signals.Object().Get(args[0].Str())
		if !number.Valid() || !number.IsNumber() {
			return js.Value{}, invalid()
		}
		return e.JS.Num(128 + number.Float()), nil
	})
}

// extraSurfacePrelude installs util.MIMEType, util.MIMEParams, util.diff,
// util.debounce and util.throttle. It follows Node 24's implementations in
// lib/internal/mime.js, lib/internal/util/diff.js (myers_diff.js) and
// lib/internal/util/{debounce,throttle}.js so that behavior, argument
// validation and error codes stay the same.
const extraSurfacePrelude = `(function (exports) {
  'use strict';

  var TIMEOUT_MAX = 2147483647;

  function describeJS(value) {
    if (value === undefined) return 'undefined';
    if (value === null) return 'null';
    var t = typeof value;
    if (t === 'number') return 'type number (' + String(value) + ')';
    if (t === 'string') return 'type string (' + JSON.stringify(value) + ')';
    if (t === 'boolean') return 'type boolean (' + String(value) + ')';
    if (t === 'bigint') return 'type bigint (' + String(value) + 'n)';
    if (t === 'symbol') return 'type symbol (' + String(value) + ')';
    if (t === 'function') return 'function ' + (value.name || '(anonymous)');
    if (Array.isArray(value)) return 'an instance of Array';
    var ctor = value.constructor && value.constructor.name ? value.constructor.name : 'Object';
    return 'an instance of ' + ctor;
  }
  function invalidType(name, expected, actual) {
    var e = new TypeError('The "' + name + '" argument must be ' + expected + '. Received ' + describeJS(actual));
    e.code = 'ERR_INVALID_ARG_TYPE';
    return e;
  }
  function invalidValue(name, value) {
    var e = new TypeError('The argument \'' + name + '\' is invalid. Received ' + describeJS(value));
    e.code = 'ERR_INVALID_ARG_VALUE';
    return e;
  }
  function outOfRange(name, actual, bound) {
    var e = new RangeError('The value of "' + name + '" is out of range. It must be ' + bound + '. Received ' + actual);
    e.code = 'ERR_OUT_OF_RANGE';
    return e;
  }
  function validateFunction(value, name) {
    if (typeof value !== 'function') throw invalidType(name, 'of type function', value);
  }
  function validateObject(value, name) {
    if (value === null || typeof value !== 'object' || Array.isArray(value)) {
      throw invalidType(name, 'of type object', value);
    }
  }
  function validateBoolean(value, name) {
    if (typeof value !== 'boolean') throw invalidType(name, 'of type boolean', value);
  }
  function validateInteger(value, name, min, max) {
    if (typeof value !== 'number') throw invalidType(name, 'of type number', value);
    if (typeof value === 'number' && (!isFinite(value) || Math.floor(value) !== value ||
        value < min || (max !== undefined && value > max))) {
      var bound = 'an integer >= ' + min;
      if (max !== undefined) bound += ' && <= ' + max;
      throw outOfRange(name, value, bound);
    }
  }
  function validateAbortSignal(signal, name) {
    if (signal === undefined) return;
    if (signal === null || typeof signal !== 'object' ||
        typeof signal.aborted !== 'boolean' ||
        typeof signal.addEventListener !== 'function') {
      throw invalidType(name, 'an instance of AbortSignal', signal);
    }
  }
  function validateOneOf(value, name, allowed) {
    if (allowed.indexOf(value) === -1) throw invalidValue(name, value);
  }
  function abortError(message, cause) {
    var e = new Error(message === undefined ? 'The operation was aborted' : message);
    e.name = 'AbortError';
    e.code = 'ABORT_ERR';
    if (cause !== undefined) e.cause = cause;
    return e;
  }
  function addAbortListener(signal, listener) {
    signal.addEventListener('abort', listener, { once: true });
  }
  function withResolvers() {
    var resolve, reject;
    var promise = new Promise(function (res, rej) { resolve = res; reject = rej; });
    return { promise: promise, resolve: resolve, reject: reject };
  }
  function defineMethod(target, name, value) {
    Object.defineProperty(target, name, {
      value: value, writable: true, enumerable: true, configurable: true
    });
  }
  function monoNow() {
    if (typeof performance !== 'undefined' && performance && typeof performance.now === 'function') {
      return performance.now();
    }
    return Date.now();
  }

  // ------------------------------------------------------------- MIME -----
  // Port of lib/internal/mime.js. The only deliberate differences are that
  // instances carry their state under symbols instead of private fields and
  // that placeholders are never substituted for user callbacks.
  var NOT_HTTP_TOKEN = /[^!#$%&'*+\-.^_\x60|~A-Za-z0-9]/;
  var NOT_HTTP_QUOTED = /[^\t\u0020-~\u0080-\u00FF]/;
  var END_BEGINNING_WS = /[^\r\n\t ]|$/;
  var START_ENDING_WS = /[\r\n\t ]*$/;
  var EQUALS_SEMICOLON_OR_END = /[;=]|$/;
  var QUOTED_VALUE = /^(?:([\\]$)|[\\][\s\S]|[^"])*(?:(")|$)/u;

  var NO_THROW = Symbol('kNoThrow');
  var TYPE = Symbol('type');
  var SUBTYPE = Symbol('subtype');
  var PARAMS = Symbol('params');
  var PARAMS_DATA = Symbol('MIMEParams data');
  var PARAMS_STRING = Symbol('MIMEParams string');
  var PARAMS_PARSED = Symbol('MIMEParams parsed');

  function hidden(target, key, value) {
    Object.defineProperty(target, key, {
      value: value, writable: true, enumerable: false, configurable: true
    });
  }
  function mimeSyntax(what, value, index) {
    var where = index !== undefined && index !== -1 ? ' at ' + index : '';
    var e = new TypeError('The MIME syntax for a ' + what + ' in "' + String(value) +
      '" is invalid' + where + ' (ERR_INVALID_MIME_SYNTAX)');
    e.code = 'ERR_INVALID_MIME_SYNTAX';
    return e;
  }
  function toASCIILower(str) {
    if (!/[^\x00-\x7f]/.test(str)) return str.toLowerCase();
    var out = '';
    for (var i = 0; i < str.length; i++) {
      var c = str[i];
      out += c >= 'A' && c <= 'Z' ? c.toLowerCase() : c;
    }
    return out;
  }
  function parseTypeAndSubtype(str, noThrow) {
    var position = str.search(END_BEGINNING_WS);
    var typeEnd = str.indexOf('/', position);
    var trimmedType = typeEnd === -1 ? str.slice(position) : str.slice(position, typeEnd);
    var invalidTypeIndex = trimmedType.search(NOT_HTTP_TOKEN);
    if (trimmedType === '' || invalidTypeIndex !== -1 || typeEnd === -1) {
      if (noThrow === NO_THROW) return null;
      throw mimeSyntax('type', str, invalidTypeIndex);
    }
    position = typeEnd + 1;
    var type = toASCIILower(trimmedType);
    var subtypeEnd = str.indexOf(';', position);
    var rawSubtype = subtypeEnd === -1 ? str.slice(position) : str.slice(position, subtypeEnd);
    position += rawSubtype.length;
    if (subtypeEnd !== -1) position += 1;
    var trimmedSubtype = rawSubtype.slice(0, rawSubtype.search(START_ENDING_WS));
    var invalidSubtypeIndex = trimmedSubtype.search(NOT_HTTP_TOKEN);
    if (trimmedSubtype === '' || invalidSubtypeIndex !== -1) {
      if (noThrow === NO_THROW) return null;
      throw mimeSyntax('subtype', str, invalidSubtypeIndex);
    }
    return [type, toASCIILower(trimmedSubtype), position];
  }
  function removeBackslashes(str) {
    var ret = '';
    var i;
    for (i = 0; i < str.length - 1; i++) {
      var c = str[i];
      if (c === '\\') {
        i++;
        ret += str[i];
      } else {
        ret += c;
      }
    }
    if (i === str.length - 1) ret += str[i];
    return ret;
  }
  function encodeValue(value) {
    if (value.length === 0) return '""';
    if (value.search(NOT_HTTP_TOKEN) === -1) return value;
    return '"' + value.replace(/["\\]/g, function (c) { return '\\' + c; }) + '"';
  }
  function parseParams(params) {
    if (params[PARAMS_PARSED]) return;
    var data = params[PARAMS_DATA];
    var str = params[PARAMS_STRING] || '';
    var position = 0;
    var endOfSource = str.slice(position).search(START_ENDING_WS) + position;
    while (position < endOfSource) {
      position += str.slice(position).search(END_BEGINNING_WS);
      var afterName = str.slice(position).search(EQUALS_SEMICOLON_OR_END) + position;
      var parameterString = toASCIILower(str.slice(position, afterName));
      position = afterName;
      if (position < endOfSource) {
        var ch = str.charAt(position);
        position += 1;
        if (ch === ';') continue;
      }
      if (position >= endOfSource) break;
      var first = str.charAt(position);
      var parameterValue = null;
      if (first === '"') {
        position += 1;
        var match = QUOTED_VALUE.exec(str.slice(position));
        position += match[0].length;
        var inside = (match[1] || match[2]) ? match[0].slice(0, -1) : match[0];
        parameterValue = removeBackslashes(inside);
        if (match[1]) parameterValue += '\\';
      } else {
        var valueEnd = str.indexOf(';', position);
        var rawValue = valueEnd === -1 ? str.slice(position) : str.slice(position, valueEnd);
        position += rawValue.length;
        var trimmedValue = rawValue.slice(0, rawValue.search(START_ENDING_WS));
        if (trimmedValue === '') continue;
        parameterValue = trimmedValue;
      }
      if (parameterString !== '' &&
          parameterString.search(NOT_HTTP_TOKEN) === -1 &&
          parameterValue.search(NOT_HTTP_QUOTED) === -1 &&
          data.has(parameterString) === false) {
        data.set(parameterString, parameterValue);
      }
      position++;
    }
    params[PARAMS_PARSED] = true;
  }
  function instantiateMimeParams(str) {
    var params = new MIMEParams();
    hidden(params, PARAMS_STRING, str);
    hidden(params, PARAMS_PARSED, false);
    return params;
  }

  class MIMEParams {
    constructor() {
      hidden(this, PARAMS_DATA, new Map());
      hidden(this, PARAMS_STRING, null);
      hidden(this, PARAMS_PARSED, true);
    }
    delete(name) {
      parseParams(this);
      this[PARAMS_DATA].delete(toASCIILower(String(name)));
    }
    get(name) {
      parseParams(this);
      var key = toASCIILower(String(name));
      return this[PARAMS_DATA].has(key) ? this[PARAMS_DATA].get(key) : null;
    }
    has(name) {
      parseParams(this);
      return this[PARAMS_DATA].has(toASCIILower(String(name)));
    }
    set(name, value) {
      parseParams(this);
      name = toASCIILower(String(name));
      value = String(value);
      var invalidNameIndex = name.search(NOT_HTTP_TOKEN);
      if (name.length === 0 || invalidNameIndex !== -1) {
        throw mimeSyntax('parameter name', name, invalidNameIndex);
      }
      var invalidValueIndex = value.search(NOT_HTTP_QUOTED);
      if (invalidValueIndex !== -1) {
        throw mimeSyntax('parameter value', value, invalidValueIndex);
      }
      this[PARAMS_DATA].set(name, value);
    }
    entries() { parseParams(this); return this[PARAMS_DATA].entries(); }
    keys() { parseParams(this); return this[PARAMS_DATA].keys(); }
    values() { parseParams(this); return this[PARAMS_DATA].values(); }
    toString() {
      parseParams(this);
      var ret = '';
      var self = this;
      this[PARAMS_DATA].forEach(function (value, key) {
        if (ret.length) ret += ';';
        ret += key + '=' + encodeValue(value);
      });
      return ret;
    }
  }
  MIMEParams.prototype[Symbol.iterator] = MIMEParams.prototype.entries;
  Object.defineProperty(MIMEParams.prototype, 'toJSON', {
    value: MIMEParams.prototype.toString,
    writable: true, enumerable: false, configurable: true
  });

  class MIMEType {
    constructor(string, noThrowSymbol) {
      if (noThrowSymbol != null && noThrowSymbol !== NO_THROW) {
        throw new TypeError('Illegal constructor');
      }
      string = String(string);
      var data = parseTypeAndSubtype(string, noThrowSymbol);
      hidden(this, TYPE, data === null ? undefined : data[0]);
      hidden(this, SUBTYPE, data === null ? undefined : data[1]);
      hidden(this, PARAMS, data === null ? undefined : instantiateMimeParams(string.slice(data[2])));
    }
    static parse(string) {
      var mt = new MIMEType(string, NO_THROW);
      return mt.type ? mt : null;
    }
    get type() { return this[TYPE]; }
    set type(value) {
      value = String(value);
      var invalidIndex = value.search(NOT_HTTP_TOKEN);
      if (value.length === 0 || invalidIndex !== -1) {
        throw mimeSyntax('type', value, invalidIndex);
      }
      this[TYPE] = toASCIILower(value);
    }
    get subtype() { return this[SUBTYPE]; }
    set subtype(value) {
      value = String(value);
      var invalidIndex = value.search(NOT_HTTP_TOKEN);
      if (value.length === 0 || invalidIndex !== -1) {
        throw mimeSyntax('subtype', value, invalidIndex);
      }
      this[SUBTYPE] = toASCIILower(value);
    }
    get essence() { return this[TYPE] + '/' + this[SUBTYPE]; }
    get params() { return this[PARAMS]; }
    toString() {
      var ret = this[TYPE] + '/' + this[SUBTYPE];
      var paramStr = this[PARAMS] === undefined ? '' : this[PARAMS].toString();
      if (paramStr.length) ret += ';' + paramStr;
      return ret;
    }
  }
  Object.defineProperty(MIMEType.prototype, 'toJSON', {
    value: MIMEType.prototype.toString,
    writable: true, enumerable: false, configurable: true
  });

  exports.MIMEParams = MIMEParams;
  exports.MIMEType = MIMEType;

  // ------------------------------------------------------------- diff -----
  var DIFF_DELETE = -1;
  var DIFF_NOP = 0;
  var DIFF_INSERT = 1;

  function myersDiff(actual, expected) {
    var actualLength = actual.length;
    var expectedLength = expected.length;
    var max = actualLength + expectedLength;
    if (max > 2147483647) {
      throw new RangeError('The value of "myersDiff input size" is out of range. It must be < 2^31. Received ' + max);
    }
    var v = new Int32Array(2 * max + 1);
    var trace = [];
    for (var diffLevel = 0; diffLevel <= max; diffLevel++) {
      trace.push(new Int32Array(v));
      for (var diagonalIndex = -diffLevel; diagonalIndex <= diffLevel; diagonalIndex += 2) {
        var offset = diagonalIndex + max;
        var previousOffset = v[offset - 1];
        var nextOffset = v[offset + 1];
        var x = diagonalIndex === -diffLevel ||
          (diagonalIndex !== diffLevel && previousOffset < nextOffset) ?
          nextOffset : previousOffset + 1;
        var y = x - diagonalIndex;
        while (x < actualLength && y < expectedLength && actual[x] === expected[y]) {
          x++;
          y++;
        }
        v[offset] = x;
        if (x >= actualLength && y >= expectedLength) {
          return diffBacktrack(trace, actual, expected);
        }
      }
    }
    return [];
  }
  function diffBacktrack(trace, actual, expected) {
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
        result.push([DIFF_NOP, actual[x - 1]]);
        x--;
        y--;
      }
      if (diffLevel > 0) {
        if (x > prevX) {
          result.push([DIFF_INSERT, actual[--x]]);
        } else {
          result.push([DIFF_DELETE, expected[--y]]);
        }
      }
    }
    return result;
  }
  function validateDiffInput(value, name) {
    if (typeof value === 'string') return;
    if (Array.isArray(value)) {
      for (var i = 0; i < value.length; i++) {
        if (typeof value[i] !== 'string') {
          throw invalidType(name + '[' + i + ']', 'of type string', value[i]);
        }
      }
      return;
    }
    throw invalidType(name, 'of type string or an instance of Array', value);
  }
  exports.diff = function diff(actual, expected) {
    if (actual === expected) return [];
    validateDiffInput(actual, 'actual');
    validateDiffInput(expected, 'expected');
    return myersDiff(actual, expected).reverse();
  };

  // --------------------------------------------------------- debounce -----
  function normalizeOptions(options) {
    if (options === undefined) return {};
    validateObject(options, 'options');
    return options;
  }

  exports.debounce = function debounce(fn, wait, options) {
    validateFunction(fn, 'fn');
    validateInteger(wait, 'wait', 0, TIMEOUT_MAX);
    var opts = normalizeOptions(options);
    var leading = opts.leading === undefined ? false : opts.leading;
    var rejectOnCancel = opts.rejectOnCancel === undefined ? false : opts.rejectOnCancel;
    var signal = opts.signal;
    validateBoolean(leading, 'options.leading');
    validateBoolean(rejectOnCancel, 'options.rejectOnCancel');
    validateAbortSignal(signal, 'options.signal');
    if (signal && signal.aborted) {
      throw abortError(undefined, signal.reason);
    }

    var args;
    var abortErr;
    var pendingCalls = [];
    var refed = true;
    var timeout;

    function rejectPending(error) {
      var calls = pendingCalls;
      pendingCalls = [];
      for (var i = 0; i < calls.length; i++) calls[i].reject(error);
    }
    function cancelWithError(error) {
      if (timeout === undefined) return;
      clearTimeout(timeout);
      timeout = undefined;
      args = undefined;
      rejectPending(error);
    }
    function cancel(reason) {
      if (timeout === undefined) return;
      cancelWithError(reason === undefined ? abortError() : abortError(undefined, reason));
    }
    function abort() {
      if (abortErr === undefined) {
        abortErr = signal.reason === undefined ? abortError() : abortError(undefined, signal.reason);
      }
      cancelWithError(abortErr);
    }
    function ref() {
      refed = true;
      if (timeout && typeof timeout.ref === 'function') timeout.ref();
      return debounced;
    }
    function unref() {
      refed = false;
      if (timeout && typeof timeout.unref === 'function') timeout.unref();
      return debounced;
    }
    function invoke(callArgs, calls) {
      if (signal && signal.aborted) {
        abort();
        for (var i = 0; i < calls.length; i++) calls[i].reject(abortErr);
        return;
      }
      var result;
      try {
        result = fn.apply(debounced, callArgs);
      } catch (error) {
        for (var j = 0; j < calls.length; j++) calls[j].reject(error);
        return;
      }
      for (var k = 0; k < calls.length; k++) calls[k].resolve(result);
    }
    function invokePending() {
      var callArgs = args;
      args = undefined;
      var calls = pendingCalls;
      pendingCalls = [];
      if (calls.length !== 0) invoke(callArgs, calls);
    }
    function onTimeout() {
      timeout = undefined;
      invokePending();
    }
    function flush() {
      if (pendingCalls.length === 0) return;
      clearTimeout(timeout);
      timeout = undefined;
      invokePending();
    }
    function debounced() {
      var callArgs = Array.prototype.slice.call(arguments);
      if (signal && signal.aborted) abort();
      if (abortErr !== undefined) return Promise.reject(abortErr);
      var invokeNow = leading && timeout === undefined;
      if (timeout !== undefined) {
        if (pendingCalls.length !== 0) {
          if (rejectOnCancel) rejectPending(abortError('The debounced call was superseded'));
        }
        if (typeof timeout.refresh === 'function') timeout.refresh();
      } else {
        timeout = setTimeout(onTimeout, wait);
        if (!refed && typeof timeout.unref === 'function') timeout.unref();
      }
      if (signal && signal.aborted) {
        abort();
        return Promise.reject(abortErr);
      }
      var call = withResolvers();
      if (invokeNow) {
        invoke(callArgs, [call]);
      } else {
        pendingCalls.push(call);
        args = callArgs;
      }
      return call.promise;
    }
    defineMethod(debounced, 'cancel', cancel);
    defineMethod(debounced, 'flush', flush);
    defineMethod(debounced, 'ref', ref);
    defineMethod(debounced, 'unref', unref);
    Object.defineProperty(debounced, 'pending', {
      get: function () {
        return pendingCalls.length === 0 ? null : pendingCalls[pendingCalls.length - 1].promise;
      },
      enumerable: false, configurable: true
    });
    Object.defineProperty(debounced, 'pendingCount', {
      get: function () { return pendingCalls.length; },
      enumerable: false, configurable: true
    });
    Object.defineProperty(debounced, 'length', { value: fn.length, configurable: true });
    Object.defineProperty(debounced, 'name', { value: fn.name, configurable: true });
    if (signal !== undefined) addAbortListener(signal, abort);
    return debounced;
  };

  // --------------------------------------------------------- throttle -----
  exports.throttle = function throttle(fn, limit, interval, options) {
    validateFunction(fn, 'fn');
    validateInteger(limit, 'limit', 1);
    validateInteger(interval, 'interval', 0, TIMEOUT_MAX);
    var opts = normalizeOptions(options);
    var concurrency = opts.concurrency === undefined ? Infinity : opts.concurrency;
    var maxPending = opts.maxPending === undefined ? Infinity : opts.maxPending;
    var overflow = opts.overflow === undefined ? 'queue' : opts.overflow;
    var signal = opts.signal;
    var strict = opts.strict === undefined ? false : opts.strict;

    if (concurrency !== Infinity) validateInteger(concurrency, 'options.concurrency', 1);
    if (maxPending !== Infinity) validateInteger(maxPending, 'options.maxPending', 0);
    validateOneOf(overflow, 'options.overflow', ['queue', 'drop']);
    validateAbortSignal(signal, 'options.signal');
    validateBoolean(strict, 'options.strict');
    if (signal && signal.aborted) {
      throw abortError(undefined, signal.reason);
    }

    var abortErr;
    var activeCount = 0;
    var pendingCalls = [];
    var pendingIndex = 0;
    var refed = true;
    var strictIndex = 0;
    var strictTicks = [];
    var timeout;
    var windowCount = 0;
    var windowStart;

    function pendingCount() { return pendingCalls.length - pendingIndex; }
    function dequeue() {
      var call = pendingCalls[pendingIndex];
      pendingCalls[pendingIndex++] = undefined;
      if (pendingIndex === pendingCalls.length) {
        pendingCalls = [];
        pendingIndex = 0;
      } else if (pendingIndex > 1024 && pendingIndex * 2 >= pendingCalls.length) {
        pendingCalls = pendingCalls.slice(pendingIndex);
        pendingIndex = 0;
      }
      return call;
    }
    function rejectPending(error) {
      var calls = pendingCalls;
      var start = pendingIndex;
      pendingCalls = [];
      pendingIndex = 0;
      for (var i = start; i < calls.length; i++) calls[i].reject(error);
    }
    function resetLimiter() {
      strictIndex = 0;
      strictTicks = [];
      windowCount = 0;
      windowStart = undefined;
    }
    function cancelWithError(error) {
      if (timeout !== undefined) {
        clearTimeout(timeout);
        timeout = undefined;
      }
      resetLimiter();
      rejectPending(error);
    }
    function cancel(reason) {
      cancelWithError(reason === undefined ? abortError() : abortError(undefined, reason));
    }
    function abort() {
      if (abortErr === undefined) {
        abortErr = signal.reason === undefined ? abortError() : abortError(undefined, signal.reason);
      }
      cancelWithError(abortErr);
    }
    function ref() {
      refed = true;
      if (timeout && typeof timeout.ref === 'function') timeout.ref();
      return throttled;
    }
    function unref() {
      refed = false;
      if (timeout && typeof timeout.unref === 'function') timeout.unref();
      return throttled;
    }
    function hasImmediateCapacity() {
      if (signal && signal.aborted) return false;
      if (pendingCount() !== 0 || activeCount >= concurrency) return false;
      var now = monoNow();
      if (strict) {
        var index = strictIndex;
        while (index < strictTicks.length && now - strictTicks[index] >= interval) index++;
        return strictTicks.length - index < limit;
      }
      return windowStart === undefined || now - windowStart >= interval || windowCount < limit;
    }
    function pruneStrictTicks(now) {
      while (strictIndex < strictTicks.length && now - strictTicks[strictIndex] >= interval) {
        strictIndex++;
      }
      if (strictIndex === strictTicks.length) {
        strictIndex = 0;
        strictTicks = [];
      } else if (strictIndex > 1024 && strictIndex * 2 >= strictTicks.length) {
        strictTicks = strictTicks.slice(strictIndex);
        strictIndex = 0;
      }
    }
    function hasCapacity(now) {
      if (strict) {
        pruneStrictTicks(now);
        return strictTicks.length - strictIndex < limit;
      }
      if (windowStart === undefined || now - windowStart >= interval) {
        windowCount = 0;
        windowStart = now;
      }
      return windowCount < limit;
    }
    function reserve(now) {
      if (strict) strictTicks.push(now);
      else windowCount++;
    }
    function onSettled() {
      activeCount--;
      drain();
    }
    function invoke(call) {
      activeCount++;
      var resultPromise;
      try {
        resultPromise = Promise.resolve(fn.apply(throttled, call.args));
      } catch (error) {
        resultPromise = Promise.reject(error);
      }
      call.resolve(resultPromise);
      resultPromise.then(onSettled, onSettled);
    }
    function drop() {
      var e = new Error('Throttled');
      e.code = 'ERR_THROTTLED';
      var rejected = Promise.reject(e);
      // Node marks dropped calls handled so a throttle overflow never
      // surfaces as an unhandled rejection; keep the returned promise itself.
      rejected.catch(function () {});
      return rejected;
    }
    function schedule() {
      if (timeout !== undefined || pendingCount() === 0 || activeCount >= concurrency) return;
      var now = monoNow();
      if (hasCapacity(now)) {
        drain();
        return;
      }
      var delay = strict ?
        strictTicks[strictIndex] + interval - now :
        windowStart + interval - now;
      timeout = setTimeout(onTimeout, delay);
      if (!refed && typeof timeout.unref === 'function') timeout.unref();
    }
    function drain() {
      if (signal && signal.aborted) {
        abort();
        return;
      }
      while (pendingCount() !== 0 && activeCount < concurrency) {
        var now = monoNow();
        if (!hasCapacity(now)) break;
        var call = dequeue();
        reserve(now);
        invoke(call);
      }
      schedule();
    }
    function onTimeout() {
      timeout = undefined;
      drain();
    }
    function throttled() {
      var callArgs = Array.prototype.slice.call(arguments);
      if (signal && signal.aborted) abort();
      if (abortErr !== undefined) return Promise.reject(abortErr);
      var now = monoNow();
      var canInvoke = pendingCount() === 0 && activeCount < concurrency && hasCapacity(now);
      if (!canInvoke && (overflow === 'drop' || pendingCount() >= maxPending)) {
        return drop();
      }
      var call = withResolvers();
      call.args = callArgs;
      if (canInvoke) {
        reserve(now);
        invoke(call);
      } else {
        pendingCalls.push(call);
        schedule();
      }
      return call.promise;
    }
    defineMethod(throttled, 'cancel', cancel);
    defineMethod(throttled, 'hasImmediateCapacity', hasImmediateCapacity);
    defineMethod(throttled, 'ref', ref);
    defineMethod(throttled, 'unref', unref);
    Object.defineProperty(throttled, 'pending', {
      get: function () {
        return pendingCount() === 0 ? null : pendingCalls[pendingCalls.length - 1].promise;
      },
      enumerable: false, configurable: true
    });
    Object.defineProperty(throttled, 'pendingCount', {
      get: pendingCount, enumerable: false, configurable: true
    });
    Object.defineProperty(throttled, 'activeCount', {
      get: function () { return activeCount; }, enumerable: false, configurable: true
    });
    Object.defineProperty(throttled, 'length', { value: fn.length, configurable: true });
    Object.defineProperty(throttled, 'name', { value: fn.name, configurable: true });
    if (signal !== undefined) addAbortListener(signal, abort);
    return throttled;
  };
})`
