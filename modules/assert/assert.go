// Package assert implements Node's assert module. The module mirrors Node 24's
// lib/assert.js and lib/internal/assert/* closely: messages, error codes and
// equality semantics are generated in the JavaScript prelude so they match the
// upstream implementation line for line, while Go only exposes Object.is.
package assert

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.Register("assert", Install)
	env.Register("assert/strict", installStrict)
}

// installStrict exposes the same object as assert.strict, matching Node.
func installStrict(e *env.Env) js.Object {
	base := e.Exports("assert")
	if strict := base.Get("strict"); strict.Valid() && strict.IsObject() {
		return strict.Object()
	}
	return base
}

// Install builds the assert module's exports object.
func Install(e *env.Env) js.Object {
	impl := e.JS.Obj()
	impl.SetFunc("same", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 {
			return e.JS.Bool(false), nil
		}
		fn := e.JS.Global().Get("Object").Object().Get("is")
		if fn.IsFunction() {
			if r, err := fn.Object().Call(js.Object{}, args[0], args[1]); err == nil {
				return e.JS.Bool(r.Bool()), nil
			}
		}
		return e.JS.Bool(false), nil
	})

	e.Global().Set("__godeAssertImpl", impl.V())
	v, err := e.JS.Eval(prelude, "node:assert")
	e.Global().Delete("__godeAssertImpl")
	if err != nil {
		panic("assert: " + err.Error())
	}
	return v.Object()
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"assert.Assert",
		"assert.AssertionError",
		"assert.CallTracker",
		"assert.CallTracker#calls",
		"assert.CallTracker#getCalls",
		"assert.CallTracker#report",
		"assert.CallTracker#reset",
		"assert.CallTracker#verify",
		"assert.assert",
		"assert.deepEqual",
		"assert.deepStrictEqual",
		"assert.doesNotMatch",
		"assert.doesNotReject",
		"assert.doesNotThrow",
		"assert.equal",
		"assert.fail",
		"assert.ifError",
		"assert.match",
		"assert.notDeepEqual",
		"assert.notDeepStrictEqual",
		"assert.notEqual",
		"assert.notStrictEqual",
		"assert.ok",
		"assert.partialDeepStrictEqual",
		"assert.rejects",
		"assert.strictEqual",
		"assert.throws",
	}
}

const prelude = `(function () {
  'use strict';
  var impl = globalThis.__godeAssertImpl;
  delete globalThis.__godeAssertImpl;

  var BT = String.fromCharCode(96);
  var NO_EXCEPTION = {};

  // ---------------------------------------------------------------------------
  // util.inspect
  // ---------------------------------------------------------------------------

  function show(v) {
    if (typeof v === 'string') return "'" + v.replace(/\\/g, '\\\\').replace(/'/g, "\\'") + "'";
    if (typeof v === 'bigint') return String(v) + 'n';
    if (typeof v === 'symbol') return v.toString();
    if (v === null) return 'null';
    if (v === undefined) return 'undefined';
    if (typeof v === 'function') return v.name ? '[Function: ' + v.name + ']' : '[Function (anonymous)]';
    if (Array.isArray(v)) return v.length === 0 ? '[]' : '[ ' + v.map(show).join(', ') + ' ]';
    if (v instanceof Date) return v.toISOString();
    if (v instanceof RegExp) return String(v);
    if (v instanceof Error) return (v.name || 'Error') + ': ' + v.message;
    if (typeof v === 'object') {
      try {
        var keys = Object.keys(v);
        return keys.length === 0 ? '{}' : '{ ' + keys.map(function (k) { return k + ': ' + show(v[k]); }).join(', ') + ' }';
      } catch (_) {
        return String(v);
      }
    }
    return String(v);
  }

  var cachedInspect;
  function inspectValue(v) {
    if (cachedInspect === undefined) {
      cachedInspect = false;
      try {
        var util = require('util');
        if (util && typeof util.inspect === 'function') cachedInspect = util.inspect;
      } catch (_) { cachedInspect = false; }
    }
    if (cachedInspect) {
      try {
        var out = cachedInspect(v, {
          compact: false,
          customInspect: false,
          depth: 1000,
          maxArrayLength: Infinity,
          showHidden: false,
          showProxy: false,
          sorted: true,
          getters: true
        });
        if (typeof out === 'string') return out;
      } catch (_) { /* fall through to show() */ }
    }
    return show(v);
  }

  // ---------------------------------------------------------------------------
  // internal/errors + internal/validators message construction
  // ---------------------------------------------------------------------------

  function formatList(array, type) {
    type = type || 'and';
    switch (array.length) {
      case 0: return '';
      case 1: return '' + array[0];
      case 2: return array[0] + ' ' + type + ' ' + array[1];
      case 3: return array[0] + ', ' + array[1] + ', ' + type + ' ' + array[2];
      default:
        return array.slice(0, -1).join(', ') + ', ' + type + ' ' + array[array.length - 1];
    }
  }

  function determineSpecificType(value) {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    var type = typeof value;
    switch (type) {
      case 'bigint': return 'type bigint (' + value + 'n)';
      case 'number':
        if (value === 0) return 1 / value === -Infinity ? 'type number (-0)' : 'type number (0)';
        if (value !== value) return 'type number (NaN)';
        if (value === Infinity) return 'type number (Infinity)';
        if (value === -Infinity) return 'type number (-Infinity)';
        return 'type number (' + value + ')';
      case 'boolean': return value ? 'type boolean (true)' : 'type boolean (false)';
      case 'symbol': return 'type symbol (' + String(value) + ')';
      case 'function': return 'function ' + value.name;
      case 'object': {
        var name;
        try { name = value.constructor && value.constructor.name; } catch (_) { name = ''; }
        if (typeof name === 'string' && name !== '') return 'an instance of ' + name;
        return inspectValue(value);
      }
      case 'string':
        if (value.length > 28) value = value.slice(0, 25) + '...';
        if (value.indexOf("'") === -1) return "type string ('" + value + "')";
        return 'type string (' + JSON.stringify(value) + ')';
      default:
        value = show(value);
        if (value.length > 28) value = value.slice(0, 25) + '...';
        return 'type ' + type + ' (' + value + ')';
    }
  }

  function makeNamed(Ctor, code, message) {
    var e = new Ctor(message);
    e.code = code;
    return e;
  }

  var kTypes = ['string', 'number', 'boolean', 'bigint', 'symbol', 'function', 'object', 'undefined'];

  function invalidArgType(name, expected, actual) {
    if (!Array.isArray(expected)) expected = [expected];
    var msg = 'The ';
    if (name.slice(-9) === ' argument') {
      msg += name + ' ';
    } else {
      msg += '"' + name + '" ' + (name.indexOf('.') === -1 ? 'argument' : 'property') + ' ';
    }
    msg += 'must be ';
    var types = [], instances = [], other = [];
    for (var i = 0; i < expected.length; i++) {
      var value = expected[i];
      if (kTypes.indexOf(value) !== -1) types.push(value);
      else if (/^[A-Z]/.test(value)) instances.push(value);
      else other.push(value);
    }
    if (instances.length > 0) {
      var pos = types.indexOf('object');
      if (pos !== -1) { types.splice(pos, 1); instances.push('Object'); }
    }
    if (types.length > 0) {
      msg += (types.length > 1 ? 'one of type ' : 'of type ') + formatList(types, 'or');
      if (instances.length > 0 || other.length > 0) msg += ' or ';
    }
    if (instances.length > 0) {
      msg += 'an instance of ' + formatList(instances, 'or');
      if (other.length > 0) msg += ' or ';
    }
    if (other.length > 0) {
      if (other.length > 1) msg += 'one of ' + formatList(other, 'or');
      else {
        if (other[0].charAt(0) !== other[0].charAt(0).toUpperCase()) msg += 'an ';
        msg += other[0];
      }
    }
    msg += '. Received ' + determineSpecificType(actual);
    return makeNamed(TypeError, 'ERR_INVALID_ARG_TYPE', msg);
  }

  function invalidArgValue(name, value, reason) {
    var inspected = inspectValue(value);
    if (inspected.length > 128) inspected = inspected.slice(0, 128) + '...';
    var type = name.indexOf('.') !== -1 ? 'property' : 'argument';
    return makeNamed(TypeError, 'ERR_INVALID_ARG_VALUE',
      'The ' + type + " '" + name + "' " + (reason || 'is invalid') +
      '. Received ' + inspected);
  }

  function missingArgs(a, b) {
    var args = Array.prototype.slice.call(arguments).map(function (x) { return '"' + x + '"'; });
    return makeNamed(TypeError, 'ERR_MISSING_ARGS',
      'The ' + formatList(args) + ' argument' + (args.length > 1 ? 's' : '') + ' must be specified');
  }

  function ambiguousArgument(name, detail) {
    return makeNamed(TypeError, 'ERR_AMBIGUOUS_ARGUMENT',
      'The "' + name + '" argument is ambiguous. ' + detail);
  }

  function outOfRange(name, range, value) {
    var received;
    if (typeof value === 'number' && Number.isInteger(value) && Math.abs(value) > Math.pow(2, 32)) {
      received = String(value).replace(/\B(?=(\d{3})+(?!\d))/g, '_');
    } else {
      received = inspectValue(value);
    }
    return makeNamed(RangeError, 'ERR_OUT_OF_RANGE',
      'The value of "' + name + '" is out of range. It must be ' + range + '. Received ' + received);
  }

  function validateUint32(value, name, positive) {
    if (typeof value !== 'number') throw invalidArgType(name, 'number', value);
    if (!Number.isInteger(value)) throw outOfRange(name, 'an integer', value);
    var min = positive ? 1 : 0;
    var max = 4294967295;
    if (value < min || value > max) throw outOfRange(name, '>= ' + min + ' && <= ' + max, value);
  }

  function validateFunction(value, name) {
    if (typeof value !== 'function') throw invalidArgType(name, 'Function', value);
  }

  function constructRequired(name) {
    return makeNamed(TypeError, 'ERR_CONSTRUCT_CALL_REQUIRED',
      'Class constructor ' + name + ' cannot be invoked without ' + BT + 'new' + BT);
  }

  function unavailableDuringExit() {
    return makeNamed(Error, 'ERR_UNAVAILABLE_DURING_EXIT', 'CallTracker is not available during process exit');
  }

  // ---------------------------------------------------------------------------
  // internal/util/comparisons (deep equality)
  // ---------------------------------------------------------------------------

  function hasOwn(o, key) {
    return Object.prototype.hasOwnProperty.call(o, key);
  }

  function ownKeys(o) {
    var keys = Object.keys(o);
    if (typeof Object.getOwnPropertySymbols === 'function') {
      var syms = Object.getOwnPropertySymbols(o);
      for (var i = 0; i < syms.length; i++) {
        var d = Object.getOwnPropertyDescriptor(o, syms[i]);
        if (d && d.enumerable) keys.push(syms[i]);
      }
    }
    return keys;
  }

  function isSharedArrayBuffer(v) {
    try {
      return typeof SharedArrayBuffer === 'function' && v instanceof SharedArrayBuffer;
    } catch (_) {
      return false;
    }
  }

  function tagOf(v) {
    var tag = Object.prototype.toString.call(v).slice(8, -1);
    // JavaScriptCore does not always expose SharedArrayBuffer's builtin tag;
    // the internal-slot check keeps the two buffer kinds distinct.
    if (tag === 'ArrayBuffer' && isSharedArrayBuffer(v)) return 'SharedArrayBuffer';
    return tag;
  }

  // A boxed primitive is compared through its intrinsic valueOf, never through
  // a userland (or deleted) valueOf property.
  function boxedValueOf(v, tag) {
    try {
      switch (tag) {
        case 'Number': return { ok: true, value: Number.prototype.valueOf.call(v) };
        case 'String': return { ok: true, value: String.prototype.valueOf.call(v) };
        case 'Boolean': return { ok: true, value: Boolean.prototype.valueOf.call(v) };
        case 'BigInt': return { ok: true, value: BigInt.prototype.valueOf.call(v) };
        case 'Symbol': return { ok: true, value: Symbol.prototype.valueOf.call(v) };
      }
    } catch (_) { /* not a real boxed primitive */ }
    return { ok: false };
  }

  // Node's partial byte comparison is a subsequence match: every expected byte
  // must appear in the actual bytes, in order, with room for the rest.
  function partialBytes(a, b) {
    var la = a.length, lb = b.length;
    if (la < lb) return false;
    var oa = 0;
    for (var ob = 0; ob < lb; ob++) {
      while (a[oa] !== b[ob]) {
        oa++;
        if (oa > la - lb + ob) return false;
      }
      oa++;
    }
    return true;
  }

  function byteView(v) {
    try {
      if (tagOf(v) === 'ArrayBuffer' || tagOf(v) === 'SharedArrayBuffer') {
        return new Uint8Array(v);
      }
      if (ArrayBuffer.isView(v)) {
        return new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
      }
    } catch (_) { }
    return null;
  }

  function bytesEqual(a, b) {
    var x = byteView(a), y = byteView(b);
    if (x === null || y === null || x.length !== y.length) return false;
    for (var i = 0; i < x.length; i++) if (x[i] !== y[i]) return false;
    return true;
  }

  function isErrorValue(v) {
    return tagOf(v) === 'Error' || v instanceof Error;
  }

  function copyError(source) {
    var target = Object.assign(Object.create(Object.getPrototypeOf(source)), source);
    Object.defineProperty(target, 'message', { __proto__: null, value: source.message });
    if (hasOwn(source, 'cause')) {
      var cause = source.cause;
      if (isErrorValue(cause)) cause = copyError(cause);
      Object.defineProperty(target, 'cause', { __proto__: null, value: cause });
    }
    return target;
  }

  function isDeepEqual(actual, expected, strict, partial, skipPrototype) {
    return innerDeepEqual(actual, expected, strict, partial, !!skipPrototype, new Map());
  }

  function innerDeepEqual(val1, val2, strict, partial, skipProto, memo) {
    if (Object.is(val1, val2)) return true;
    if (val1 === val2) {
      if (val1 === 0) return strict ? false : true;
      return true;
    }
    var type1 = typeof val1, type2 = typeof val2;
    if (type1 !== 'object' || val1 === null || type2 !== 'object' || val2 === null) {
      if (strict) return false;
      // A boxed primitive never equals a primitive in loose mode either:
      // only two primitives may coerce.
      var obj1 = type1 === 'object' && val1 !== null;
      var obj2 = type2 === 'object' && val2 !== null;
      if (obj1 !== obj2) return false;
      return val1 == val2 || (val1 !== val1 && val2 !== val2); // eslint-disable-line eqeqeq
    }
    // Only pairs already known to be equal (or currently being compared, which
    // assumes equality so cycles terminate) are memoised. A recorded false
    // would make a repeated unequal pair look equal.
    var pairs = memo.get(val1);
    if (pairs !== undefined) {
      for (var p = 0; p < pairs.length; p++) {
        if (pairs[p] === val2) return true;
      }
    } else {
      pairs = [];
      memo.set(val1, pairs);
    }
    pairs.push(val2);
    var result = objectEqual(val1, val2, strict, partial, skipProto, memo);
    if (!result) {
      var at = pairs.indexOf(val2);
      if (at !== -1) pairs.splice(at, 1);
    }
    return result;
  }

  function objectEqual(val1, val2, strict, partial, skipProto, memo) {
    var tag = tagOf(val1);
    if (tag !== tagOf(val2)) return false;
    // The global object and process are hosts with internal identity; a plain
    // object that copies their properties is never deep-equal to them.
    if (typeof globalThis !== 'undefined' && (val1 === globalThis) !== (val2 === globalThis)) return false;
    try {
      if (typeof process !== 'undefined' && ((val1 === process) !== (val2 === process))) return false;
    } catch (_) { }

    // Each built-in branch validates its own payload and then falls through to
    // the shared own-property comparison, so extra enumerable properties (and
    // symbols) still make two otherwise identical values unequal.
    switch (tag) {
      case 'Date': {
        var t1 = val1.getTime(), t2 = val2.getTime();
        if (t1 !== t2 && (t1 === t1 || t2 === t2)) return false;
        break;
      }
      case 'RegExp':
        if (val1.source !== val2.source || val1.flags !== val2.flags ||
            val1.lastIndex !== val2.lastIndex) return false;
        break;
      case 'Number':
      case 'String':
      case 'Boolean':
      case 'BigInt':
      case 'Symbol': {
        // Boxed primitives compare by their intrinsic value. A fake boxed
        // primitive (one that only borrows the prototype) has no internal
        // slot and is never equal to a real one.
        var box1 = boxedValueOf(val1, tag);
        var box2 = boxedValueOf(val2, tag);
        if (!box1.ok || !box2.ok) return false;
        if (!Object.is(box1.value, box2.value)) return false;
        break;
      }
      case 'Promise':
      case 'WeakMap':
      case 'WeakSet':
      case 'WeakRef':
      case 'FinalizationRegistry':
      case 'Function':
        return false;
      case 'ArrayBuffer':
      case 'SharedArrayBuffer':
      case 'DataView':
        if (partial) {
          if (!partialBytes(byteView(val1), byteView(val2))) return false;
        } else if (!bytesEqual(val1, val2)) return false;
        break;
      case 'Map':
        if (!mapEqual(val1, val2, strict, partial, skipProto, memo)) return false;
        break;
      case 'Set':
        if (!setEqual(val1, val2, strict, partial, skipProto, memo)) return false;
        break;
      default:
        break;
    }

    if (ArrayBuffer.isView(val1) || ArrayBuffer.isView(val2)) {
      if (!ArrayBuffer.isView(val1) || !ArrayBuffer.isView(val2)) return false;
      if (tagOf(val1) !== tagOf(val2)) return false;
      if (strict && !skipProto && Object.getPrototypeOf(val1) !== Object.getPrototypeOf(val2)) return false;
      var viewTag = tagOf(val1);
      if (!strict && (viewTag === 'Float32Array' || viewTag === 'Float64Array' || viewTag === 'Float16Array')) {
        // Loose mode compares float elements numerically, so +0 equals -0 and
        // NaN equals NaN (matching areSimilarFloatArrays).
        if (val1.length !== val2.length) return false;
        for (var fi = 0; fi < val1.length; fi++) {
          var f1 = val1[fi], f2 = val2[fi];
          if (f1 !== f2 && (f1 === f1 || f2 === f2)) return false;
        }
        return keysEqual(val1, val2, memo, strict, partial, skipProto, true);
      }
      if (partial) {
        if (!partialBytes(byteView(val1), byteView(val2))) return false;
      } else if (!bytesEqual(val1, val2)) {
        return false;
      }
      return keysEqual(val1, val2, memo, strict, partial, skipProto, true);
    }

    if (strict && !skipProto && Object.getPrototypeOf(val1) !== Object.getPrototypeOf(val2)) {
      return false;
    }

    if (isErrorValue(val1) || isErrorValue(val2)) {
      if (!isErrorValue(val1) || !isErrorValue(val2)) return false;
      if (val1.name !== val2.name) return false;
      if (val1.message !== val2.message) return false;
      var c1 = hasOwn(val1, 'cause'), c2 = hasOwn(val2, 'cause');
      if (c1 !== c2) return false;
      if (c1) {
        if (!innerDeepEqual(val1.cause, val2.cause, strict, partial, skipProto, memo)) return false;
      }
      if (hasOwn(val1, 'errors') !== hasOwn(val2, 'errors')) return false;
      if (hasOwn(val1, 'errors')) {
        if (!innerDeepEqual(val1.errors, val2.errors, strict, partial, skipProto, memo)) return false;
      }
    }

    if (Array.isArray(val1)) {
      if (partial) {
        if (!partialArrayEquiv(val1, val2, strict, skipProto, memo)) return false;
        return keysEqual(val1, val2, memo, strict, true, skipProto, true);
      }
      var len1 = val1.length, len2 = val2.length;
      if (len1 !== len2) return false;
      for (var i = 0; i < len1; i++) {
        var h1 = hasOwn(val1, i), h2 = hasOwn(val2, i);
        if (h1 !== h2) return false;
        if (h1 && !innerDeepEqual(val1[i], val2[i], strict, false, skipProto, memo)) return false;
      }
      return keysEqual(val1, val2, memo, strict, false, skipProto, true);
    }

    return keysEqual(val1, val2, memo, strict, partial, skipProto, false);
  }

  // Node's partial array comparison is an ordered subsequence match over the
  // expected elements, with sparse holes comparing as absent entries.
  function partialArrayEquiv(a, b, strict, skipProto, memo) {
    var aPos = 0;
    for (var i = 0; i < b.length; i++) {
      var isSparse = b[i] === undefined && !hasOwn(b, i);
      if (isSparse) return partialSparseArrayEquiv(a, b, strict, skipProto, memo, aPos, i);
      while (!(isSparse = a[aPos] === undefined && !hasOwn(a, aPos)) &&
             !innerDeepEqual(a[aPos], b[i], strict, false, skipProto, memo)) {
        aPos++;
        if (aPos > a.length - b.length + i) return false;
      }
      if (isSparse) return partialSparseArrayEquiv(a, b, strict, skipProto, memo, aPos, i);
      aPos++;
    }
    return true;
  }

  function partialSparseArrayEquiv(a, b, strict, skipProto, memo, startA, startB) {
    var keysA = Object.keys(a);
    var keysB = Object.keys(b);
    var lenA = keysA.length - startA;
    var lenB = keysB.length - startB;
    if (lenA < lenB) return false;
    var aPos = startA;
    for (var i = 0; i < lenB; i++) {
      var keyB = keysB[startB + i];
      while (!innerDeepEqual(a[keysA[aPos]], b[keyB], strict, false, skipProto, memo)) {
        aPos++;
        if (aPos > keysA.length - lenB + i) return false;
      }
      aPos++;
    }
    return true;
  }

  function keysEqual(a, b, memo, strict, partial, skipProto, skipIndices) {
    var keysB = ownKeys(b);
    for (var i = 0; i < keysB.length; i++) {
      var key = keysB[i];
      if (skipIndices && isArrayIndex(key)) continue;
      // Loose deep equality ignores symbol-keyed properties, as Node does.
      if (typeof key === 'symbol' && !strict) continue;
      if (!hasOwn(a, key)) return false;
      if (!innerDeepEqual(readKey(a, key), readKey(b, key), strict, partial, skipProto, memo)) return false;
    }
    if (!partial) {
      var keysA = ownKeys(a);
      for (var j = 0; j < keysA.length; j++) {
        var k = keysA[j];
        if (skipIndices && isArrayIndex(k)) continue;
        if (typeof k === 'symbol' && !strict) continue;
        if (!hasOwn(b, k)) return false;
      }
    }
    return true;
  }

  function isArrayIndex(key) {
    return typeof key === 'string' && key !== '' && String(Number(key) >>> 0) === key && Number(key) !== 4294967295;
  }

  function readKey(obj, key) {
    try { return obj[key]; } catch (e) { return e; }
  }

  function mapEqual(a, b, strict, partial, skipProto, memo) {
    if (partial ? a.size < b.size : a.size !== b.size) return false;
    var actualEntries = Array.from(a.entries());
    var expectedEntries = Array.from(b.entries());
    var used = new Array(actualEntries.length);
    for (var i = 0; i < expectedEntries.length; i++) {
      var found = false;
      for (var j = 0; j < actualEntries.length; j++) {
        if (used[j]) continue;
        if (!innerDeepEqual(actualEntries[j][0], expectedEntries[i][0], strict, false, skipProto, memo)) continue;
        if (!innerDeepEqual(actualEntries[j][1], expectedEntries[i][1], strict, false, skipProto, memo)) continue;
        used[j] = true;
        found = true;
        break;
      }
      if (!found) return false;
    }
    return true;
  }

  function setEqual(a, b, strict, partial, skipProto, memo) {
    if (partial ? a.size < b.size : a.size !== b.size) return false;
    var actualValues = Array.from(a.values());
    var expectedValues = Array.from(b.values());
    var used = new Array(actualValues.length);
    for (var i = 0; i < expectedValues.length; i++) {
      var found = false;
      for (var j = 0; j < actualValues.length; j++) {
        if (used[j]) continue;
        if (!innerDeepEqual(actualValues[j], expectedValues[i], strict, false, skipProto, memo)) continue;
        used[j] = true;
        found = true;
        break;
      }
      if (!found) return false;
    }
    return true;
  }

  // ---------------------------------------------------------------------------
  // internal/assert/assertion_error
  // ---------------------------------------------------------------------------

  var kReadableOperator = {
    deepStrictEqual: 'Expected values to be strictly deep-equal:',
    partialDeepStrictEqual: 'Expected values to be partially and strictly deep-equal:',
    strictEqual: 'Expected values to be strictly equal:',
    strictEqualObject: 'Expected "actual" to be reference-equal to "expected":',
    deepEqual: 'Expected values to be loosely deep-equal:',
    notDeepStrictEqual: 'Expected "actual" not to be strictly deep-equal to:',
    notStrictEqual: 'Expected "actual" to be strictly unequal to:',
    notStrictEqualObject: 'Expected "actual" not to be reference-equal to "expected":',
    notDeepEqual: 'Expected "actual" not to be loosely deep-equal to:',
    notIdentical: 'Values have same structure but are not reference-equal:',
    notDeepEqualUnequal: 'Expected values not to be loosely deep-equal:'
  };
  var kMethodsWithCustomMessageDiff = {
    deepStrictEqual: true, strictEqual: true, partialDeepStrictEqual: true
  };
  var kMaxShortStringLength = 12;
  var kMaxLongStringLength = 512;
  var kOps = { DELETE: -1, NOP: 0, INSERT: 1 };
  var kNopLinesToCollapse = 5;

  function linesEqual(actual, expected, checkComma) {
    if (actual === expected) return true;
    if (checkComma) return (actual + ',') === expected || actual === (expected + ',');
    return false;
  }

  function myersDiff(actual, expected, checkComma) {
    var actualLength = actual.length;
    var expectedLength = expected.length;
    var max = actualLength + expectedLength;
    var v = new Int32Array(2 * max + 1);
    var trace = [];
    for (var level = 0; level <= max; level++) {
      trace.push(new Int32Array(v));
      for (var diagonal = -level; diagonal <= level; diagonal += 2) {
        var offset = diagonal + max;
        var previousOffset = v[offset - 1];
        var nextOffset = v[offset + 1];
        var x = diagonal === -level ||
          (diagonal !== level && previousOffset < nextOffset) ? nextOffset : previousOffset + 1;
        var y = x - diagonal;
        while (x < actualLength && y < expectedLength &&
               linesEqual(actual[x], expected[y], checkComma)) {
          x++;
          y++;
        }
        v[offset] = x;
        if (x >= actualLength && y >= expectedLength) {
          return myersBacktrack(trace, actual, expected, checkComma);
        }
      }
    }
    return [];
  }

  function myersBacktrack(trace, actual, expected, checkComma) {
    var actualLength = actual.length;
    var expectedLength = expected.length;
    var max = actualLength + expectedLength;
    var x = actualLength;
    var y = expectedLength;
    var result = [];
    for (var level = trace.length - 1; level >= 0; level--) {
      var v = trace[level];
      var diagonal = x - y;
      var offset = diagonal + max;
      var prevDiagonal;
      if (diagonal === -level || (diagonal !== level && v[offset - 1] < v[offset + 1])) {
        prevDiagonal = diagonal + 1;
      } else {
        prevDiagonal = diagonal - 1;
      }
      var prevX = v[prevDiagonal + max];
      var prevY = prevX - prevDiagonal;
      while (x > prevX && y > prevY) {
        var item = actual[x - 1];
        var value = checkComma && item.slice(-1) !== ',' ? expected[y - 1] : item;
        result.push([kOps.NOP, value]);
        x--;
        y--;
      }
      if (level > 0) {
        if (x > prevX) {
          result.push([kOps.INSERT, actual[--x]]);
        } else {
          result.push([kOps.DELETE, expected[--y]]);
        }
      }
    }
    return result;
  }

  function printMyersDiff(diff, operator) {
    var message = '';
    var nopCount = 0;
    for (var i = diff.length - 1; i >= 0; i--) {
      var operation = diff[i][0];
      var value = diff[i][1];
      var previousOperation = i < diff.length - 1 ? diff[i + 1][0] : null;
      if (previousOperation === kOps.NOP && operation !== previousOperation) {
        if (nopCount === kNopLinesToCollapse + 1) {
          message += '  ' + diff[i + 1][1] + '\n';
        } else if (nopCount === kNopLinesToCollapse + 2) {
          message += '  ' + diff[i + 2][1] + '\n';
          message += '  ' + diff[i + 1][1] + '\n';
        } else if (nopCount >= kNopLinesToCollapse + 3) {
          message += '...\n';
          message += '  ' + diff[i + 1][1] + '\n';
        }
        nopCount = 0;
      }
      if (operation === kOps.INSERT) {
        // Node's partial diff uses a leading space only when colors are on;
        // with colors disabled (always, here) it is a plus like the others.
        message += '+ ' + value + '\n';
      } else if (operation === kOps.DELETE) {
        message += '- ' + value + '\n';
      } else if (operation === kOps.NOP) {
        if (nopCount < kNopLinesToCollapse) message += '  ' + value + '\n';
        nopCount++;
      }
    }
    message = message.replace(/\s+$/, '');
    return { message: '\n' + message, skipped: false };
  }

  function getStackedDiff(actual, expected) {
    return { message: '\n+ ' + actual + '\n- ' + expected };
  }

  function getSimpleDiff(originalActual, actual, originalExpected, expected) {
    var stringsLen = String(actual).length + String(expected).length;
    if (typeof originalActual === 'string') stringsLen -= 2;
    if (typeof originalExpected === 'string') stringsLen -= 2;
    if (stringsLen <= kMaxShortStringLength && (originalActual !== 0 || originalExpected !== 0)) {
      return { message: actual + ' !== ' + expected, header: '' };
    }
    return getStackedDiff(actual, expected);
  }

  function getErrorMessage(operator, message) {
    return message || kReadableOperator[operator];
  }

  function checkOperator(actual, expected, operator) {
    if (operator === 'strictEqual' &&
        ((typeof actual === 'object' && actual !== null &&
          typeof expected === 'object' && expected !== null) ||
         (typeof actual === 'function' && typeof expected === 'function'))) {
      return 'strictEqualObject';
    }
    return operator;
  }

  function createErrDiff(actual, expected, operator, customMessage, diffType) {
    diffType = diffType === undefined ? 'simple' : diffType;
    operator = checkOperator(actual, expected, operator);
    var message = '';
    var skipped = false;
    var inspectedActual = String(inspectValue(actual));
    var inspectedExpected = String(inspectValue(expected));
    var splitActual = inspectedActual.split('\n');
    var splitExpected = inspectedExpected.split('\n');
    var showSimpleDiff = splitActual.length <= 1 && splitExpected.length <= 1 &&
      (typeof actual !== 'object' || actual === null ||
       typeof expected !== 'object' || expected === null);
    var header = '+ actual - expected';

    if (showSimpleDiff) {
      var simple = getSimpleDiff(actual, splitActual[0], expected, splitExpected[0]);
      message = simple.message;
      if (simple.header !== undefined) header = simple.header;
    } else if (inspectedActual === inspectedExpected) {
      operator = 'notIdentical';
      if (splitActual.length > 50 && diffType !== 'full') {
        message = splitActual.slice(0, 50).join('\n') + '\n...}';
        skipped = true;
      } else {
        message = splitActual.join('\n');
      }
      header = '';
    } else {
      var checkCommaDisparity = actual != null && typeof actual === 'object';
      var diff = myersDiff(splitActual, splitExpected, checkCommaDisparity);
      var printed = printMyersDiff(diff, operator);
      message = printed.message;
      if (printed.skipped) skipped = true;
    }

    var headerMessage = getErrorMessage(operator, customMessage) + '\n' + header;
    var skippedMessage = skipped ? '\n... Skipped lines' : '';
    return headerMessage + skippedMessage + '\n' + message + '\n';
  }

  class AssertionError extends Error {
    constructor(options) {
      if (options === null || Array.isArray(options) || typeof options !== 'object') {
        throw invalidArgType('options', 'Object', options);
      }
      var message = options.message;
      var operator = options.operator;
      var actual = options.actual;
      var expected = options.expected;
      var details = options.details;
      var diff = options.diff === undefined ? 'simple' : options.diff;
      var text;

      if (message != null) {
        if (kMethodsWithCustomMessageDiff[operator]) {
          text = createErrDiff(actual, expected, operator, message, diff);
        } else {
          text = String(message);
        }
      } else if (kMethodsWithCustomMessageDiff[operator]) {
        if (isErrorValue(actual) && isErrorValue(expected) &&
            typeof actual.stack === 'string' && typeof expected.stack === 'string') {
          actual = copyError(actual);
          expected = copyError(expected);
        }
        text = createErrDiff(actual, expected, operator, message, diff);
      } else if (operator === 'notDeepStrictEqual' || operator === 'notStrictEqual') {
        var base = kReadableOperator[operator];
        var res = String(inspectValue(actual)).split('\n');
        if (operator === 'notStrictEqual' &&
            ((typeof actual === 'object' && actual !== null) || typeof actual === 'function')) {
          base = kReadableOperator.notStrictEqualObject;
        }
        if (res.length > 50 && diff !== 'full') {
          res[46] = '...';
          while (res.length > 47) res.pop();
        }
        if (res.length === 1) {
          text = base + (res[0].length > 5 ? '\n\n' : ' ') + res[0];
        } else {
          text = base + '\n\n' + res.join('\n') + '\n';
        }
      } else {
        var resActual = String(inspectValue(actual));
        var other = String(inspectValue(expected));
        var knownOperator = kReadableOperator[operator];
        if (operator === 'notDeepEqual' && resActual === other) {
          resActual = knownOperator + '\n\n' + resActual;
          if (resActual.length > 1024 && diff !== 'full') {
            resActual = resActual.slice(0, 1021) + '...';
          }
          text = resActual;
        } else {
          if (resActual.length > kMaxLongStringLength && diff !== 'full') {
            resActual = resActual.slice(0, 509) + '...';
          }
          if (other.length > kMaxLongStringLength && diff !== 'full') {
            other = other.slice(0, 509) + '...';
          }
          if (operator === 'deepEqual') {
            resActual = knownOperator + '\n\n' + resActual + '\n\nshould loosely deep-equal\n\n';
          } else {
            var newOp = kReadableOperator[operator + 'Unequal'];
            if (newOp) {
              resActual = newOp + '\n\n' + resActual + '\n\nshould not loosely deep-equal\n\n';
            } else {
              other = ' ' + operator + ' ' + other;
            }
          }
          text = resActual + other;
        }
      }

      super(text);
      this.generatedMessage = !message;
      // The descriptor must have a null prototype: ToPropertyDescriptor reads
      // inherited get/set, so a userland Object.prototype.get would otherwise
      // make defineProperty throw "value present with getter or setter".
      Object.defineProperty(this, 'name', {
        __proto__: null,
        value: 'AssertionError [ERR_ASSERTION]',
        enumerable: false,
        writable: true,
        configurable: true
      });
      this.code = 'ERR_ASSERTION';
      if (details) {
        this.actual = undefined;
        this.expected = undefined;
        this.operator = undefined;
        for (var i = 0; i < details.length; i++) {
          this['message ' + i] = details[i].message;
          this['actual ' + i] = details[i].actual;
          this['expected ' + i] = details[i].expected;
          this['operator ' + i] = details[i].operator;
          this['stack trace ' + i] = details[i].stack;
        }
      } else {
        this.actual = actual;
        this.expected = expected;
        this.operator = operator;
      }
      void this.stack;
      this.name = 'AssertionError';
      this.diff = diff;
      // JavaScriptCore's stack holds only frames. Node's stack begins with the
      // coded header and starts at the caller (stackStartFn), so compose that
      // shape and drop the assertion implementation's own frames.
      var rawStack = String(this.stack || '');
      var stackHeader = 'AssertionError [ERR_ASSERTION]: ' + text;
      var rawLines = rawStack.split('\n');
      var keptFrames = [];
      for (var li = 0; li < rawLines.length; li++) {
        var frameLine = rawLines[li];
        if (frameLine === '' || frameLine === stackHeader) continue;
        if (frameLine.indexOf('@node:assert:') !== -1) continue;
        if (frameLine.indexOf('@[gode:') !== -1) continue;
        keptFrames.push(frameLine);
      }
      try {
        Object.defineProperty(this, 'stack', {
          __proto__: null,
          value: stackHeader + (keptFrames.length ? '\n' + keptFrames.join('\n') : ''),
          writable: true,
          configurable: true
        });
      } catch (_) {
        try {
          this.stack = stackHeader + (keptFrames.length ? '\n' + keptFrames.join('\n') : '');
        } catch (_) { /* leave the engine's stack in place */ }
      }
    }

    toString() {
      return this.name + ' [' + this.code + ']: ' + this.message;
    }
  }

  function innerFail(obj) {
    if (obj.message instanceof Error) throw obj.message;
    throw new AssertionError(obj);
  }

  function optionsOf(self) {
    if (self === null || self === undefined) return undefined;
    try { return self._options; } catch (_) { return undefined; }
  }

  function diffOf(self) {
    var options = optionsOf(self);
    return options === undefined ? undefined : options.diff;
  }

  function skipProtoOf(self) {
    var options = optionsOf(self);
    return options === undefined ? false : !!options.skipPrototype;
  }

  // ---------------------------------------------------------------------------
  // module level assertions (lib/assert.js + internal/assert/utils.js)
  // ---------------------------------------------------------------------------

  // callExpressionAt recovers the call expression that produced the error from
  // its source line and column (for example assert.ok('')) from a dense line. Node
  // tokenizes the line with acorn; this is the same slice for ordinary
  // identifier/member callees without a parser dependency.
  function callExpressionAt(line, column) {
    var i = Number(column) || 0;
    if (i > 0) i -= 1;
    if (i < 0) i = 0;
    if (i > line.length) i = line.length;
    var open = -1;
    for (var p = i; p < line.length && p < i + 80; p++) {
      var ch = line.charAt(p);
      if (ch === '(') { open = p; break; }
      if (ch === ';') break;
    }
    if (open === -1) {
      var back = line.lastIndexOf('(', i);
      if (back !== -1 && i - back <= 3) open = back;
    }
    if (open === -1) return null;
    var start = open;
    while (start > 0) {
      var c = line.charAt(start - 1);
      // Non-ASCII code units are identifier characters (the first-line fixture
      // uses an identifier spelled ässört).
      if (c.charCodeAt(0) > 127 || /[A-Za-z0-9_$]/.test(c)) { start--; continue; }
      if (c === '.' || c === '?') { start--; continue; }
      if (c === ']') {
        var depth = 0, j = start - 1, matched = false;
        for (; j >= 0; j--) {
          if (line.charAt(j) === ']') depth++;
          else if (line.charAt(j) === '[') {
            depth--;
            if (depth === 0) { matched = true; break; }
          }
        }
        if (matched) { start = j; continue; }
      }
      if (c === ' ') {
        var k = start - 1;
        while (k > 0 && line.charAt(k - 1) === ' ') k--;
        if (k > 0 && (line.charAt(k - 1) === '.' || line.charAt(k - 1) === '?')) {
          start = k;
          continue;
        }
      }
      break;
    }
    var parens = 0, quote = null;
    for (var q = open; q < line.length; q++) {
      var qc = line.charAt(q);
      if (quote) {
        if (qc === '\\') { q++; continue; }
        if (qc === quote) quote = null;
        continue;
      }
      if (qc === '"' || qc === "'" || qc === BT) { quote = qc; continue; }
      if (qc === '(') parens++;
      else if (qc === ')') {
        parens--;
        if (parens === 0) return line.slice(start, q + 1);
      }
    }
    return line.slice(start);
  }

  // getErrorSourceExpression reads the source text of the caller's statement,
  // which is what Node puts into a generated assert.ok() message.
  function falsyExpressionMessage() {
    try {
      var util = require('util');
      var fs = require('fs');
      if (util && typeof util.getCallSites === 'function') {
        var sites = util.getCallSites(10);
        for (var i = 0; i < sites.length; i++) {
          var site = sites[i];
          var file = site.scriptName;
          if (!file || file.indexOf('node:') === 0 || file.indexOf('[gode:') === 0) continue;
          var text = fs.readFileSync(file, 'utf8');
          var line = text.split('\n')[site.lineNumber - 1];
          if (line === undefined) continue;
          line = line.replace(/[\x00-\x08\x0b\x0c\x0e-\x1f]/g, '');
          var expression = callExpressionAt(line, site.columnNumber);
          if (expression === null) expression = line.replace(/^\s+/, '').replace(/\s+$/, '');
          if (expression.indexOf(';') !== -1 || expression.length > 160) {
            var matches = line.match(/(?:[A-Za-z0-9_$.\u0080-\uFFFF]+)\.ok\([^)]*\)/g);
            if (matches && matches.length) expression = matches[matches.length - 1];
          }
          if (expression === '') continue;
          return 'The expression evaluated to a falsy value:\n\n  ' + expression + '\n';
        }
      }
    } catch (_) { /* fall back to the generic message */ }
    return 'The expression evaluated to a falsy value:';
  }

  function innerOk(argLen, value, message) {
    if (value) return;
    var generatedMessage = false;
    if (argLen === 0) {
      generatedMessage = true;
      message = 'No value argument passed to ' + BT + 'assert.ok()' + BT;
    } else if (message === null || message === undefined) {
      generatedMessage = true;
      message = falsyExpressionMessage();
    } else if (message instanceof Error) {
      throw message;
    }
    var err = new AssertionError({
      actual: value,
      expected: true,
      message: message,
      operator: '=='
    });
    err.generatedMessage = generatedMessage;
    throw err;
  }

  function ok(value, message) {
    innerOk(arguments.length, value, message);
  }

  function equal(actual, expected, message) {
    if (arguments.length < 2) throw missingArgs('actual', 'expected');
    if (actual != expected && (!Number.isNaN(actual) || !Number.isNaN(expected))) { // eslint-disable-line eqeqeq
      innerFail({
        actual: actual, expected: expected, message: message, operator: '==',
        diff: diffOf(this)
      });
    }
  }

  function notEqual(actual, expected, message) {
    if (arguments.length < 2) throw missingArgs('actual', 'expected');
    if (actual == expected || (Number.isNaN(actual) && Number.isNaN(expected))) { // eslint-disable-line eqeqeq
      innerFail({
        actual: actual, expected: expected, message: message, operator: '!=',
        diff: diffOf(this)
      });
    }
  }

  function deepEqual(actual, expected, message) {
    if (arguments.length < 2) throw missingArgs('actual', 'expected');
    if (!isDeepEqual(actual, expected, false, false, false)) {
      innerFail({
        actual: actual, expected: expected, message: message, operator: 'deepEqual',
        diff: diffOf(this)
      });
    }
  }

  function notDeepEqual(actual, expected, message) {
    if (arguments.length < 2) throw missingArgs('actual', 'expected');
    if (isDeepEqual(actual, expected, false, false, false)) {
      innerFail({
        actual: actual, expected: expected, message: message, operator: 'notDeepEqual',
        diff: diffOf(this)
      });
    }
  }

  function deepStrictEqual(actual, expected, message) {
    if (arguments.length < 2) throw missingArgs('actual', 'expected');
    if (!isDeepEqual(actual, expected, true, false, skipProtoOf(this))) {
      innerFail({
        actual: actual, expected: expected, message: message, operator: 'deepStrictEqual',
        diff: diffOf(this)
      });
    }
  }

  function notDeepStrictEqual(actual, expected, message) {
    if (arguments.length < 2) throw missingArgs('actual', 'expected');
    if (isDeepEqual(actual, expected, true, false, skipProtoOf(this))) {
      innerFail({
        actual: actual, expected: expected, message: message, operator: 'notDeepStrictEqual',
        diff: diffOf(this)
      });
    }
  }

  function strictEqual(actual, expected, message) {
    if (arguments.length < 2) throw missingArgs('actual', 'expected');
    if (!Object.is(actual, expected)) {
      innerFail({
        actual: actual, expected: expected, message: message, operator: 'strictEqual',
        diff: diffOf(this)
      });
    }
  }

  function notStrictEqual(actual, expected, message) {
    if (arguments.length < 2) throw missingArgs('actual', 'expected');
    if (Object.is(actual, expected)) {
      innerFail({
        actual: actual, expected: expected, message: message, operator: 'notStrictEqual',
        diff: diffOf(this)
      });
    }
  }

  function partialDeepStrictEqual(actual, expected, message) {
    if (arguments.length < 2) throw missingArgs('actual', 'expected');
    if (!isDeepEqual(actual, expected, true, true, false)) {
      innerFail({
        actual: actual, expected: expected, message: message, operator: 'partialDeepStrictEqual',
        diff: diffOf(this)
      });
    }
  }

  // ---------------------------------------------------------------------------
  // throws / rejects / doesNotThrow / doesNotReject
  // ---------------------------------------------------------------------------

  class Comparison {
    constructor(obj, keys, actual) {
      for (var i = 0; i < keys.length; i++) {
        var key = keys[i];
        if (!(key in obj)) continue;
        if (actual !== undefined && typeof actual[key] === 'string' &&
            obj[key] instanceof RegExp && obj[key].test(actual[key])) {
          this[key] = actual[key];
        } else {
          this[key] = obj[key];
        }
      }
    }
  }

  function isErrorLike(v) {
    return isErrorValue(v) ||
      (v !== null && typeof v === 'object' &&
       typeof v.name === 'string' && typeof v.message === 'string');
  }

  function compareExceptionKey(actual, expected, key, message, keys, fnName) {
    if (!(key in actual) || !isDeepEqual(actual[key], expected[key], true, false, false)) {
      if (!message) {
        var a = new Comparison(actual, keys);
        var b = new Comparison(expected, keys, actual);
        var err = new AssertionError({
          actual: a,
          expected: b,
          operator: 'deepStrictEqual'
        });
        err.actual = actual;
        err.expected = expected;
        err.operator = fnName;
        throw err;
      }
      innerFail({
        actual: actual, expected: expected, message: message, operator: fnName
      });
    }
  }

  function expectedException(actual, expected, message, fnName) {
    var generatedMessage = false;
    var throwError = false;

    if (typeof expected !== 'function') {
      if (expected instanceof RegExp) {
        var str = String(actual);
        if (expected.test(str)) return;
        if (!message) {
          generatedMessage = true;
          message = 'The input did not match the regular expression ' +
            inspectValue(expected) + '. Input:\n\n' + inspectValue(str) + '\n';
        }
        throwError = true;
      } else if (typeof actual !== 'object' || actual === null) {
        var plain = new AssertionError({
          actual: actual, expected: expected, message: message,
          operator: 'deepStrictEqual'
        });
        plain.operator = fnName;
        throw plain;
      } else {
        var keys = Object.keys(expected);
        if (expected instanceof Error) {
          keys.push('name', 'message');
        } else if (keys.length === 0) {
          throw invalidArgValue('error', expected, 'may not be an empty object');
        }
        for (var i = 0; i < keys.length; i++) {
          if (typeof actual[keys[i]] === 'string' &&
              expected[keys[i]] instanceof RegExp &&
              expected[keys[i]].test(actual[keys[i]])) {
            continue;
          }
          compareExceptionKey(actual, expected, keys[i], message, keys, fnName);
        }
        return;
      }
    } else if (expected.prototype !== undefined && actual instanceof expected) {
      return;
    } else if (Error.isPrototypeOf(expected)) {
      // Error.isPrototypeOf is Object.prototype.isPrototypeOf with 'this'
      // bound to Error: it matches an error class (including subclasses of
      // AssertionError), not an error instance.
      if (!message) {
        generatedMessage = true;
        message = 'The error is expected to be an instance of "' + expected.name + '". Received ';
        if (isErrorLike(actual)) {
          var name = (actual.constructor && actual.constructor.name) || actual.name;
          if (expected.name === name) {
            message += 'an error with identical name but a different prototype.';
          } else {
            message += '"' + name + '"';
          }
          if (actual.message) message += '\n\nError message:\n\n' + actual.message;
        } else {
          message += '"' + inspectValue(actual) + '"';
        }
      }
      throwError = true;
    } else {
      var res = expected.call({}, actual);
      if (res !== true) {
        if (!message) {
          generatedMessage = true;
          var fname = expected.name ? '"' + expected.name + '" ' : '';
          message = 'The ' + fname + 'validation function is expected to return "true". Received ' + inspectValue(res);
          if (isErrorLike(actual)) message += '\n\nCaught error:\n\n' + String(actual);
        }
        throwError = true;
      }
    }

    if (throwError) {
      var err = new AssertionError({
        actual: actual, expected: expected, message: message, operator: fnName
      });
      err.generatedMessage = generatedMessage;
      throw err;
    }
  }

  function expectsError(fnName, actual, args) {
    var error = args[0];
    var message = args[1];
    var hasMessage = args.length >= 2;

    if (typeof error === 'string') {
      if (hasMessage) {
        throw invalidArgType('error', ['Object', 'Error', 'Function', 'RegExp'], error);
      }
      if (actual !== NO_EXCEPTION && actual !== null && typeof actual === 'object') {
        if (actual.message === error) {
          throw ambiguousArgument('error/message',
            'The error message "' + actual.message + '" is identical to the message.');
        }
      } else if (actual === error) {
        throw ambiguousArgument('error/message',
          'The error "' + actual + '" is identical to the message.');
      }
      message = error;
      error = undefined;
    } else if (error !== null && error !== undefined &&
               typeof error !== 'object' && typeof error !== 'function') {
      throw invalidArgType('error', ['Object', 'Error', 'Function', 'RegExp'], error);
    }

    if (actual === NO_EXCEPTION) {
      var details = '';
      if (error && error.name) details += ' (' + error.name + ')';
      details += message ? ': ' + message : '.';
      var fnType = fnName === 'rejects' ? 'rejection' : 'exception';
      innerFail({
        actual: undefined,
        expected: error,
        operator: fnName,
        message: 'Missing expected ' + fnType + details
      });
    }

    if (!error) return;
    expectedException(actual, error, message, fnName);
  }

  function hasMatchingError(actual, expected) {
    if (typeof expected !== 'function') {
      if (expected instanceof RegExp) return expected.test(String(actual));
      throw invalidArgType('expected', ['Function', 'RegExp'], expected);
    }
    if (expected.prototype !== undefined && actual instanceof expected) return true;
    if (Error.isPrototypeOf(expected)) return false;
    return expected.call({}, actual) === true;
  }

  function expectsNoError(fnName, actual, args) {
    if (actual === NO_EXCEPTION) return;
    var error = args[0];
    var message = args[1];
    if (typeof error === 'string') {
      message = error;
      error = undefined;
    }
    if (!error || hasMatchingError(actual, error)) {
      var details = message ? ': ' + message : '.';
      var fnType = fnName === 'doesNotReject' ? 'rejection' : 'exception';
      innerFail({
        actual: actual,
        expected: error,
        operator: fnName,
        message: 'Got unwanted ' + fnType + details + '\n' +
          'Actual message: "' + (actual === null || actual === undefined ? actual : actual.message) + '"'
      });
    }
    throw actual;
  }

  function getActual(fn) {
    validateFunction(fn, 'fn');
    try {
      fn();
    } catch (e) {
      return e;
    }
    return NO_EXCEPTION;
  }

  function checkIsPromise(obj) {
    return (obj !== null && typeof obj === 'object' &&
      typeof obj.then === 'function' && typeof obj.catch === 'function');
  }

  function waitForActual(promiseFn) {
    var result;
    if (typeof promiseFn === 'function') {
      result = promiseFn();
      if (!checkIsPromise(result)) {
        return Promise.reject(makeNamed(TypeError, 'ERR_INVALID_RETURN_VALUE',
          'Expected instance of Promise to be returned from the "promiseFn" function but got ' +
          determineSpecificType(result) + '.'));
      }
    } else if (checkIsPromise(promiseFn)) {
      result = promiseFn;
    } else {
      return Promise.reject(invalidArgType('promiseFn', ['Function', 'Promise'], promiseFn));
    }
    return Promise.resolve(result).then(function (value) {
      return NO_EXCEPTION;
    }, function (e) {
      return e;
    });
  }

  function throws(fn) {
    var args = Array.prototype.slice.call(arguments, 1);
    expectsError('throws', getActual(fn), args);
  }

  function doesNotThrow(fn) {
    var args = Array.prototype.slice.call(arguments, 1);
    expectsNoError('doesNotThrow', getActual(fn), args);
  }

  function rejects(fn) {
    var args = Array.prototype.slice.call(arguments, 1);
    var actual;
    try {
      actual = waitForActual(fn);
    } catch (e) {
      return Promise.reject(e);
    }
    return actual.then(function (resolved) {
      expectsError('rejects', resolved, args);
    });
  }

  function doesNotReject(fn) {
    var args = Array.prototype.slice.call(arguments, 1);
    var actual;
    try {
      actual = waitForActual(fn);
    } catch (e) {
      return Promise.reject(e);
    }
    return actual.then(function (resolved) {
      expectsNoError('doesNotReject', resolved, args);
    });
  }

  function ifError(err) {
    if (err !== null && err !== undefined) {
      var message = 'ifError got unwanted exception: ';
      if (typeof err === 'object' && typeof err.message === 'string') {
        if (err.message.length === 0 && err.constructor) {
          message += err.constructor.name;
        } else {
          message += err.message;
        }
      } else if (err && typeof err === 'object' && !Array.isArray(err)) {
        var objKeys = Object.keys(err);
        if (objKeys.length === 1 &&
            (typeof err[objKeys[0]] !== 'object' || err[objKeys[0]] === null)) {
          message += '{ ' + objKeys[0] + ': ' + inspectValue(err[objKeys[0]]) + ' }';
        } else {
          message += inspectValue(err);
        }
      } else {
        message += inspectValue(err);
      }

      var newErr = new AssertionError({
        actual: err,
        expected: null,
        operator: 'ifError',
        message: message,
        diff: diffOf(this)
      });

      var origStack = err.stack;
      if (typeof origStack === 'string' && typeof newErr.stack === 'string') {
        var origStackStart = origStack.indexOf('\n    at');
        if (origStackStart !== -1) {
          var originalFrames = origStack.slice(origStackStart + 1).split('\n');
          var newFrames = String(newErr.stack).split('\n');
          for (var i = 0; i < originalFrames.length; i++) {
            var pos = newFrames.indexOf(originalFrames[i]);
            if (pos !== -1) {
              newFrames = newFrames.slice(0, pos);
              break;
            }
          }
          newErr.stack = newFrames.join('\n') + '\n' + originalFrames.join('\n');
        }
      }
      throw newErr;
    }
  }

  function internalMatch(string, regexp, message, isMatch, fnName) {
    if (!(regexp instanceof RegExp)) {
      throw invalidArgType('regexp', 'RegExp', regexp);
    }
    if (typeof string !== 'string' || regexp.test(string) !== isMatch) {
      if (message instanceof Error) throw message;
      var generatedMessage = !message;
      if (!message) {
        message = typeof string !== 'string' ?
          'The "string" argument must be of type string. Received type ' +
            typeof string + ' (' + inspectValue(string) + ')' :
          (isMatch ?
            'The input did not match the regular expression ' :
            'The input was expected to not match the regular expression ') +
            inspectValue(regexp) + '. Input:\n\n' + inspectValue(string) + '\n';
      }
      var err = new AssertionError({
        actual: string,
        expected: regexp,
        message: message,
        operator: fnName
      });
      err.generatedMessage = generatedMessage;
      throw err;
    }
  }

  function match(string, regexp, message) {
    internalMatch(string, regexp, message, true, 'match');
  }

  function doesNotMatch(string, regexp, message) {
    internalMatch(string, regexp, message, false, 'doesNotMatch');
  }

  var failWarned = false;

  function fail(actual, expected, message, operator, stackStartFn) {
    var argsLen = arguments.length;
    var internalMessage = false;
    if (actual == null && argsLen <= 1) { // eslint-disable-line eqeqeq
      internalMessage = true;
      message = 'Failed';
    } else if (argsLen === 1) {
      message = actual;
      actual = undefined;
    } else {
      if (!failWarned) {
        failWarned = true;
        if (typeof process !== 'undefined' && typeof process.emitWarning === 'function') {
          process.emitWarning(
            'assert.fail() with more than one argument is deprecated. ' +
              'Please use assert.strictEqual() instead or only pass a message.',
            'DeprecationWarning',
            'DEP0094');
        }
      }
      if (argsLen === 2) operator = '!=';
    }
    if (message instanceof Error) throw message;
    var err = new AssertionError({
      actual: actual,
      expected: expected,
      operator: operator === undefined ? 'fail' : operator,
      message: message,
      diff: diffOf(this)
    });
    if (internalMessage) err.generatedMessage = true;
    throw err;
  }

  // ---------------------------------------------------------------------------
  // CallTracker (lib/internal/assert/calltracker.js)
  // ---------------------------------------------------------------------------

  // JavaScriptCore words frozen-array mutation as "Attempted to assign to
  // readonly property", while Node guarantees "object is not extensible". The
  // non-enumerable own push reproduces Node's wording without changing the
  // array's keys, length or prototype (so deepStrictEqual still matches).
  function freezeArrayLikeNode(arr) {
    try {
      Object.defineProperty(arr, 'push', {
        __proto__: null,
        value: function () {
          throw new TypeError('Cannot add property ' + arr.length + ', object is not extensible');
        },
        writable: false,
        enumerable: false,
        configurable: false
      });
    } catch (_) { /* leave the frozen array's native push */ }
    return Object.freeze(arr);
  }

  // Node freezes each call record; a Proxy keeps V8's "object is not
  // extensible" wording when user code mutates one (JavaScriptCore words the
  // same failure differently).
  function frozenCallRecord(thisArg, args) {
    var target = Object.freeze({ thisArg: thisArg, arguments: args });
    try {
      return new Proxy(target, {
        __proto__: null,
        set: function (t, prop) {
          throw new TypeError('Cannot add property ' + String(prop) + ', object is not extensible');
        },
        defineProperty: function (t, prop) {
          throw new TypeError('Cannot add property ' + String(prop) + ', object is not extensible');
        }
      });
    } catch (_) {
      return target;
    }
  }

  class CallTrackerContext {
    constructor(expected, name, stackTrace) {
      this.calls = [];
      this.expected = expected;
      this.name = name;
      this.stackTrace = stackTrace;
    }
    track(thisArg, args) {
      var copy = freezeArrayLikeNode(Array.prototype.slice.call(args));
      this.calls.push(frozenCallRecord(thisArg, copy));
    }
    get delta() { return this.calls.length - this.expected; }
    reset() { this.calls = []; }
    getCalls() { return freezeArrayLikeNode(this.calls.slice()); }
    report() {
      if (this.delta !== 0) {
        return {
          message: 'Expected the ' + this.name + ' function to be executed ' +
            this.expected + ' time(s) but was executed ' +
            this.calls.length + ' time(s).',
          actual: this.calls.length,
          expected: this.expected,
          operator: this.name,
          stack: this.stackTrace
        };
      }
    }
  }

  class CallTracker {
    constructor() {
      this._contexts = new Set();
      this._tracked = new WeakMap();
    }
    _getTracked(tracked) {
      if (!this._tracked.has(tracked)) {
        throw invalidArgValue('tracked', tracked, 'is not a tracked function');
      }
      return this._tracked.get(tracked);
    }
    reset(tracked) {
      if (tracked === undefined) {
        this._contexts.forEach(function (context) { context.reset(); });
        return;
      }
      this._getTracked(tracked).reset();
    }
    getCalls(tracked) {
      return this._getTracked(tracked).getCalls();
    }
    calls(fn, expected) {
      if (typeof process !== 'undefined' && process._exiting) throw unavailableDuringExit();
      if (expected === undefined) expected = 1;
      if (typeof fn === 'number') {
        expected = fn;
        fn = function () {};
      } else if (fn === undefined) {
        fn = function () {};
      }
      validateUint32(expected, 'expected', true);
      if (typeof fn !== 'function') throw invalidArgType('fn', 'Function', fn);
      var context = new CallTrackerContext(expected, fn.name || 'calls', new Error());
      var tracked = new Proxy(fn, {
        // Null prototype: the Proxy machinery reads its hooks through the
        // handler, and an inherited Object.prototype.get must not be mistaken
        // for a trap (the official test installs one as a sentinel).
        __proto__: null,
        apply: function (target, thisArg, argList) {
          context.track(thisArg, argList);
          return Reflect.apply(target, thisArg, argList);
        }
      });
      this._contexts.add(context);
      this._tracked.set(tracked, context);
      return tracked;
    }
    report() {
      var errors = [];
      this._contexts.forEach(function (context) {
        var entry = context.report();
        if (entry !== undefined) errors.push(entry);
      });
      return errors;
    }
    verify() {
      var errors = this.report();
      if (errors.length === 0) return;
      var message = errors.length === 1 ?
        errors[0].message :
        'Functions were not called the expected number of times';
      throw new AssertionError({ message: message, details: errors });
    }
  }

  // ---------------------------------------------------------------------------
  // Assert class (lib/assert.js)
  // ---------------------------------------------------------------------------

  function Assert(options) {
    if (!new.target) throw constructRequired('Assert');
    options = Object.assign({ __proto__: null, strict: true, skipPrototype: false }, options);
    if (options.diff !== undefined) {
      if (options.diff !== 'simple' && options.diff !== 'full') {
        throw invalidArgValue('options.diff', options.diff, "must be one of: 'simple', 'full'");
      }
    }
    this.AssertionError = AssertionError;
    Object.defineProperty(this, '_options', {
      value: options,
      enumerable: false,
      configurable: false,
      writable: false
    });
    if (options.strict) {
      this.equal = this.strictEqual;
      this.deepEqual = this.deepStrictEqual;
      this.notEqual = this.notStrictEqual;
      this.notDeepEqual = this.notDeepStrictEqual;
    }
  }

  Assert.prototype.ok = function okProto() { innerOk(arguments.length, arguments[0], arguments[1]); };
  Assert.prototype.equal = equal;
  Assert.prototype.notEqual = notEqual;
  Assert.prototype.deepEqual = deepEqual;
  Assert.prototype.notDeepEqual = notDeepEqual;
  Assert.prototype.deepStrictEqual = deepStrictEqual;
  Assert.prototype.notDeepStrictEqual = notDeepStrictEqual;
  Assert.prototype.strictEqual = strictEqual;
  Assert.prototype.notStrictEqual = notStrictEqual;
  Assert.prototype.partialDeepStrictEqual = partialDeepStrictEqual;
  Assert.prototype.fail = fail;
  Assert.prototype.throws = throws;
  Assert.prototype.rejects = rejects;
  Assert.prototype.doesNotThrow = doesNotThrow;
  Assert.prototype.doesNotReject = doesNotReject;
  Assert.prototype.ifError = ifError;
  Assert.prototype.match = match;
  Assert.prototype.doesNotMatch = doesNotMatch;

  // ---------------------------------------------------------------------------
  // module wiring
  // ---------------------------------------------------------------------------

  var assert = ok;

  assert.ok = assert;
  assert.AssertionError = AssertionError;
  assert.CallTracker = CallTracker;
  assert.Assert = Assert;
  assert.fail = fail;
  assert.equal = equal;
  assert.notEqual = notEqual;
  assert.deepEqual = deepEqual;
  assert.notDeepEqual = notDeepEqual;
  assert.deepStrictEqual = deepStrictEqual;
  assert.notDeepStrictEqual = notDeepStrictEqual;
  assert.strictEqual = strictEqual;
  assert.notStrictEqual = notStrictEqual;
  assert.partialDeepStrictEqual = partialDeepStrictEqual;
  assert.match = match;
  assert.doesNotMatch = doesNotMatch;
  assert.throws = throws;
  assert.rejects = rejects;
  assert.doesNotThrow = doesNotThrow;
  assert.doesNotReject = doesNotReject;
  assert.ifError = ifError;

  function strict() {
    innerOk(arguments.length, arguments[0], arguments[1]);
  }

  var strictAssert = Object.assign(strict, assert, {
    equal: assert.strictEqual,
    deepEqual: assert.deepStrictEqual,
    notEqual: assert.notStrictEqual,
    notDeepEqual: assert.notDeepStrictEqual
  });
  strictAssert.Assert = Assert;
  strictAssert.strict = strictAssert;

  assert.strict = strictAssert;

  return assert;
})()`
