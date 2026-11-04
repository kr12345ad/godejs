package buffer

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf16"

	"godejs/internal/env"
	"godejs/internal/js"
)

const (
	// Node sets Buffer.poolSize to 64 KiB; the pool itself lives in the
	// shared ArrayBuffer and is only an allocation hint.
	poolSize = 65536
	// Node exports kMaxLength as v8::Uint8Array::kMaxLength, which is what
	// every ERR_OUT_OF_RANGE("size") message prints.
	maxLength       = 9007199254740991
	maxStringLength = 536870888
	inspectMaxBytes = 50
)

// module holds the per-environment Buffer constructor and prototype.
type module struct {
	e           *env.Env
	proto       js.Object
	ctor        js.Object
	exports     js.Object
	toPrimitive js.Value
}

// buf is one Buffer value: the live Uint8Array plus a byte slice that aliases
// its backing store, so JavaScript index writes and Go-side mutations observe
// the same memory.
type buf struct {
	m   *module
	obj js.Object
	b   []byte
}

func init() { env.Register("buffer", Install) }

// Install builds the buffer module and installs the Buffer global.
func Install(e *env.Env) js.Object {
	m := &module{e: e, exports: e.JS.Obj()}
	m.proto = e.JS.Obj().Keep()
	// Node's Buffer extends Uint8Array, so keep the typed-array prototype in
	// the chain (length, byteLength, buffer and symbol methods live there).
	uint8Proto := e.Global().Get("Uint8Array").Object().Get("prototype").Object()
	m.proto.SetProto(uint8Proto)

	holder := e.JS.Obj()
	holder.SetFunc("make", func(this js.Object, args []js.Value) (js.Value, error) {
		// Buffer(x) is the deprecated constructor form of Buffer.from, except
		// that a numeric first argument allocates zero-filled memory.
		if len(args) == 0 || !args[0].Valid() || args[0].IsUndefined() {
			return m.newBuffer(nil).obj.V(), nil
		}
		if args[0].IsNumber() {
			// new Buffer(42, 'utf8') is invalid: the legacy numeric form never
			// took an encoding, and Node names the numeric argument as the
			// offending "string" one.
			if len(args) > 1 && args[1].Valid() && args[1].IsString() {
				return js.Value{}, typeErr("The \"string\" argument must be of type string.%s", describeReceived(args[0]))
			}
			return m.alloc(args)
		}
		return m.from(args)
	})
	e.Global().SetHidden("__godeBufferMake", holder.Get("make"))
	// A host function object is not a constructor in JavaScriptCore, so the
	// public Buffer must be a JavaScript function that delegates allocation
	// back to the host. Node's Buffer is a Uint8Array subclass, hence the
	// prototype link; the allocated value already carries Buffer.prototype.
	maker, err := e.JS.Eval(`(function (proto, make) {
  var DEP0005 = 'Buffer() is deprecated due to security and usability issues. ' +
    'Please use the Buffer.alloc(), Buffer.allocUnsafe(), or Buffer.from() ' +
    'methods instead.';
  function Buffer(arg, encodingOrOffset, length) {
    var pending = false;
    try {
      pending = (process.execArgv || []).indexOf('--pending-deprecation') !== -1 ||
        String(process.env.NODE_PENDING_DEPRECATION) === '1';
    } catch (e) {}
    if (new.target === undefined || pending) {
      try {
        process.emitWarning(DEP0005, 'DeprecationWarning', 'DEP0005');
      } catch (e) {}
    }
    return make(arg, encodingOrOffset, length);
  }
  Buffer.prototype = proto;
  try {
    Object.defineProperty(Buffer, 'prototype', {
      writable: false, enumerable: false, configurable: false
    });
  } catch (e) {}
  return Buffer;
})`, "[gode:buffer/ctor]")
	if err != nil {
		panic("buffer: constructor prelude: " + err.Error())
	}
	built, callErr := maker.Object().Call(js.Object{}, m.proto.V(), holder.Get("make"))
	if callErr != nil {
		panic("buffer: constructor: " + callErr.Error())
	}
	m.ctor = built.Object().Keep()
	e.Global().Delete("__godeBufferMake")
	m.proto.Set("constructor", m.ctor.V())

	m.installStatics()
	m.installPrototype()
	m.installInspect(m.exports)

	exp := m.exports
	exp.Set("Buffer", m.ctor.V())
	exp.Set("kMaxLength", e.JS.Num(maxLength))
	exp.Set("kStringMaxLength", e.JS.Num(maxStringLength))
	// Node's documentation also lists the two bounds as module-level
	// constants; expose them under those names as well.
	exp.Set("MAX_LENGTH", e.JS.Num(maxLength))
	exp.Set("MAX_STRING_LENGTH", e.JS.Num(maxStringLength))
	m.installSlowBuffer()
	m.installBlob()
	// Symbol.toPrimitive has no Go-side accessor, so Buffer.from consults it
	// through this helper, after the array-like check exactly like Node.
	toPrim, err := e.JS.Eval(`(function (v) {
  var f = v[Symbol.toPrimitive];
  if (typeof f === 'function') return f.call(v, 'string');
  return undefined;
})`, "[gode:buffer/toPrimitive]")
	if err != nil {
		panic("buffer: toPrimitive prelude: " + err.Error())
	}
	// The helper must be rooted: an unkept function value can be collected,
	// and the dangling reference then calls whatever reuses the address.
	m.toPrimitive = toPrim.Object().Keep().V()
	exp.SetFunc("transcode", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.transcode(args)
	})

	constants := e.JS.Obj()
	constants.Set("MAX_LENGTH", e.JS.Num(maxLength))
	constants.Set("MAX_STRING_LENGTH", e.JS.Num(maxStringLength))
	exp.Set("constants", constants.V())
	exp.SetFunc("atob", func(this js.Object, args []js.Value) (js.Value, error) {
		s := arg(args, 0).Str()
		out, err := decodeBase64(s, false)
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(latin1String(out)), nil
	})
	exp.SetFunc("btoa", func(this js.Object, args []js.Value) (js.Value, error) {
		s := arg(args, 0).Str()
		return e.JS.Str(base64.StdEncoding.EncodeToString(latin1Bytes(s))), nil
	})
	// resolveObjectURL looks up an id created by URL.createObjectURL. The
	// registry is a Map shared through the global object so the url module can
	// populate it without a circular import.
	exp.SetFunc("resolveObjectURL", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		if !v.Valid() || !v.IsString() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"The \"id\" argument must be of type string.")
		}
		registry := e.Global().Get("__godeObjectURLs")
		if !registry.Valid() || !registry.IsObject() {
			return e.JS.Undefined(), nil
		}
		get := registry.Object().Get("get")
		if !get.Valid() || !get.IsFunction() {
			return e.JS.Undefined(), nil
		}
		out, err := get.Object().Call(registry.Object(), v)
		if err != nil {
			return js.Value{}, err
		}
		return out, nil
	})
	// Node's isAscii/isUtf8 only accept typed arrays and ArrayBuffers; strings
	// and every other value are ERR_INVALID_ARG_TYPE. The byte scan itself runs
	// in JavaScript: reading a backing store through the C API pins the buffer,
	// so a later structuredClone(transfer) could not detach it.
	helper, err := e.JS.Eval(asciiHelpersSource, "buffer.isascii.js")
	if err != nil {
		panic(err)
	}
	helpers := helper.Object().Keep()
	checkInput := func(v js.Value) error {
		if !isTypedArrayValue(m, v) && !m.isAnyArrayBuffer(v) {
			return typeErr("The \"input\" argument must be an instance of ArrayBuffer, Buffer, or TypedArray.%s", describeReceived(v))
		}
		return nil
	}
	exp.SetFunc("isAscii", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		if err := checkInput(v); err != nil {
			return js.Value{}, err
		}
		fn := helpers.Get("isAscii")
		r, err := fn.Object().Call(js.Object{}, v)
		if err != nil {
			return js.Value{}, err
		}
		return r, nil
	})
	exp.SetFunc("isUtf8", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		if err := checkInput(v); err != nil {
			return js.Value{}, err
		}
		fn := helpers.Get("isUtf8")
		r, err := fn.Object().Call(js.Object{}, v)
		if err != nil {
			return js.Value{}, err
		}
		return r, nil
	})

	e.Global().Set("Buffer", m.ctor.V())
	m.installLegacyGetters()
	return exp
}

// asciiHelpersSource scans bytes entirely in JavaScript. A detached view or
// buffer has byteLength 0 and is treated as empty, matching Node.
const asciiHelpersSource = `(function () {
  function viewOf(v) {
    if (ArrayBuffer.isView(v)) return new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
    return new Uint8Array(v);
  }
  function isAscii(v) {
    if (v.byteLength === 0) return true;
    var u8 = viewOf(v);
    for (var i = 0; i < u8.length; i++) {
      if (u8[i] > 0x7f) return false;
    }
    return true;
  }
  function isUtf8(v) {
    if (v.byteLength === 0) return true;
    var u8 = viewOf(v), i = 0, len = u8.length;
    while (i < len) {
      var c = u8[i++];
      if (c < 0x80) continue;
      var need, min;
      if (c >= 0xc2 && c <= 0xdf) { need = 1; min = 0x80; }
      else if (c >= 0xe0 && c <= 0xef) { need = 2; min = 0x800; }
      else if (c >= 0xf0 && c <= 0xf4) { need = 3; min = 0x10000; }
      else return false;
      if (i + need > len) return false;
      var cp = c & (0x3f >> need);
      for (var k = 0; k < need; k++) {
        var b = u8[i++];
        if ((b & 0xc0) !== 0x80) return false;
        cp = (cp << 6) | (b & 0x3f);
      }
      if (cp < min || cp > 0x10ffff || (cp >= 0xd800 && cp <= 0xdfff)) return false;
    }
    return true;
  }
  return { isAscii: isAscii, isUtf8: isUtf8 };
})()`

func (m *module) installStatics() {
	c := m.ctor
	c.SetNum("poolSize", poolSize)
	c.SetFunc("isBuffer", func(this js.Object, args []js.Value) (js.Value, error) {
		_, ok := m.asBuffer(arg(args, 0))
		return m.e.JS.Bool(ok), nil
	})
	c.SetFunc("isEncoding", func(this js.Object, args []js.Value) (js.Value, error) {
		_, ok := normalizeEncoding(arg(args, 0).Str())
		return m.e.JS.Bool(ok), nil
	})
	c.SetFunc("from", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.from(args)
	})
	// Buffer.of(...items) is Buffer.from(items) with the arguments as the list.
	c.SetFunc("of", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.from([]js.Value{m.e.JS.Arr(args).V()})
	})
	c.SetFunc("alloc", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.alloc(args)
	})
	c.SetFunc("allocUnsafe", func(this js.Object, args []js.Value) (js.Value, error) {
		n, err := m.sizeArg(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		return m.newBuffer(make([]byte, n)).obj.V(), nil
	})
	c.SetFunc("allocUnsafeSlow", func(this js.Object, args []js.Value) (js.Value, error) {
		n, err := m.sizeArg(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		return m.newBuffer(make([]byte, n)).obj.V(), nil
	})
	c.SetFunc("byteLength", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		if v.IsString() {
			// Node treats an unrecognized encoding as utf8 here.
			enc := "utf8"
			if a := arg(args, 1); a.Valid() && !a.IsUndefined() && a.IsString() && a.Str() != "" {
				if name, ok := normalizeEncoding(a.Str()); ok {
					enc = name
				}
			}
			b, err := encodeString(v.Str(), enc)
			if err != nil {
				return js.Value{}, err
			}
			return m.e.JS.Num(float64(len(b))), nil
		}
		// Buffer, TypedArray, DataView, ArrayBuffer and SharedArrayBuffer all
		// report their length through byteLength. Verify the internal slot
		// first: a plain object with a byteLength property is not accepted, and
		// an ArrayBuffer from another realm has to be recognised by slot too.
		if !v.IsObject() || v.IsFunction() {
			return js.Value{}, typeErr("The \"string\" argument must be of type string or an instance of Buffer or ArrayBuffer.%s", describeReceived(v))
		}
		if !m.isArrayBufferView(v) && !m.isAnyArrayBuffer(v) && v.Tag() != "DataView" {
			return js.Value{}, typeErr("The \"string\" argument must be of type string or an instance of Buffer or ArrayBuffer.%s", describeReceived(v))
		}
		if n := v.Object().Get("byteLength"); n.IsNumber() {
			return m.e.JS.Num(n.Float()), nil
		}
		if data, ok := v.Bytes(); ok {
			return m.e.JS.Num(float64(len(data))), nil
		}
		return m.e.JS.Num(0), nil
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			"The \"string\" argument must be of type string or an instance of Buffer or ArrayBuffer."+describeReceived(v))
	})
	c.SetFunc("concat", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.concat(args)
	})
	c.SetFunc("compare", func(this js.Object, args []js.Value) (js.Value, error) {
		a, err := m.bytesArg(arg(args, 0), "buf1")
		if err != nil {
			return js.Value{}, err
		}
		b, err := m.bytesArg(arg(args, 1), "buf2")
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(bytes.Compare(a, b))), nil
	})
	c.SetFunc("copyBytesFrom", func(this js.Object, args []js.Value) (js.Value, error) {
		src := arg(args, 0)
		// Node counts elements, not bytes, and rejects anything that is not
		// a real typed array (a DataView is not enough).
		if !isTypedArrayValue(m, src) {
			return js.Value{}, typeErr("The \"view\" argument must be an instance of TypedArray.%s", describeReceived(src))
		}
		data, _ := src.Bytes()
		viewLen := src.Object().Get("length").Int()
		if viewLen == 0 {
			return m.newBuffer(nil).obj.V(), nil
		}
		elemSize := len(data) / viewLen
		start := 0
		if a := arg(args, 1); a.Valid() && !a.IsUndefined() {
			n, err := validatedIndex("offset", a)
			if err != nil {
				return js.Value{}, err
			}
			if n >= viewLen {
				return m.newBuffer(nil).obj.V(), nil
			}
			start = n
		}
		end := viewLen
		if a := arg(args, 2); a.Valid() && !a.IsUndefined() {
			n, err := validatedIndex("length", a)
			if err != nil {
				return js.Value{}, err
			}
			// An overlong length clamps instead of throwing, as in Node.
			if end = start + n; end > viewLen {
				end = viewLen
			}
		}
		if end <= start {
			return m.newBuffer(nil).obj.V(), nil
		}
		out := make([]byte, (end-start)*elemSize)
		copy(out, data[start*elemSize:end*elemSize])
		return m.newBuffer(out).obj.V(), nil
	})
}

// typedArrayTags are the views Buffer.copyBytesFrom accepts. A DataView is
// deliberately absent: Node's isTypedArray rejects it.
var typedArrayTags = map[string]bool{
	"Int8Array": true, "Uint8Array": true, "Uint8ClampedArray": true,
	"Int16Array": true, "Uint16Array": true, "Int32Array": true,
	"Uint32Array": true, "Float32Array": true, "Float64Array": true,
	"BigInt64Array": true, "BigUint64Array": true,
}

func isTypedArrayValue(m *module, v js.Value) bool {
	if !v.IsObject() || v.IsFunction() {
		return false
	}
	// Check the Buffer marker directly: asBuffer aliases the live bytes through
	// the C API, and that pins the backing ArrayBuffer so it can never be
	// detached by structuredClone(transfer).
	if v.Object().Get("_godeBuffer").Bool() {
		return true
	}
	if typedArrayTags[v.Tag()] {
		return true
	}
	// Views created in another realm do not expose this realm's tag.
	return m.isArrayBufferView(v) && v.Tag() != "DataView"
}

// isArrayBufferView runs the real ArrayBuffer.isView, which works across
// realms; a local tag check alone misses vm-created arrays.
func (m *module) isArrayBufferView(v js.Value) bool {
	ab := m.e.Global().Get("ArrayBuffer")
	if !ab.IsFunction() {
		return false
	}
	isView := ab.Object().Get("isView")
	if !isView.IsFunction() {
		return false
	}
	r, err := isView.Object().Call(ab.Object(), v)
	return err == nil && r.Bool()
}

// isAnyArrayBuffer recognises ArrayBuffer and SharedArrayBuffer values,
// including ones created in another context (whose prototype tag differs).
func (m *module) isAnyArrayBuffer(v js.Value) bool {
	if !v.IsObject() || v.IsFunction() {
		return false
	}
	if tag := v.Tag(); tag == "ArrayBuffer" || tag == "SharedArrayBuffer" {
		return true
	}
	obj := m.e.Global().Get("Object")
	if !obj.IsObject() {
		return false
	}
	toString := obj.Object().Get("prototype").Object().Get("toString")
	if !toString.IsFunction() {
		return false
	}
	r, err := toString.Object().Call(v.Object())
	if err != nil {
		return false
	}
	s := r.Str()
	if s == "[object ArrayBuffer]" || s == "[object SharedArrayBuffer]" {
		return true
	}
	if ctor := v.Object().Get("constructor"); ctor.IsFunction() {
		name := ctor.Object().Get("name")
		if name.IsString() && (name.Str() == "ArrayBuffer" || name.Str() == "SharedArrayBuffer") {
			return true
		}
	}
	return false
}

// validatedIndex mirrors Node's validateInteger(value, name, 0): a non-number
// is a type error, while a non-integer or out-of-range value names the range.
func validatedIndex(name string, a js.Value) (int, error) {
	if !a.IsNumber() {
		return 0, typeErr("The \""+name+"\" argument must be of type number.%s", describeReceived(a))
	}
	f := a.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, outOfRange(name, "an integer", a)
	}
	if f < 0 || f > 9007199254740991 {
		return 0, outOfRange(name, sizeRange, a)
	}
	return int(f), nil
}

func (m *module) installPrototype() {
	p := m.proto
	p.SetFuncHidden("toString", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		enc := "utf8"
		if a := arg(args, 0); a.Valid() && !a.IsUndefined() {
			name, ok := normalizeEncoding(a.Str())
			if !ok {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", a.Str()))
			}
			enc = name
		}
		// Node's toString does not use adjustOffset: start <= 0 clamps to 0, a
		// start at or past the end yields '', and end is truncated with NaN
		// mapping to 0, so an unparseable end produces an empty string.
		n := len(b.b)
		startF := 0.0
		if a := arg(args, 1); a.Valid() && !a.IsUndefined() {
			f, _ := m.jsNumber(a)
			switch {
			case math.IsNaN(f), f <= 0:
				startF = 0
			case f >= float64(n):
				return m.e.JS.Str(""), nil
			default:
				startF = math.Trunc(f)
			}
		}
		endF := float64(n)
		if a := arg(args, 2); a.Valid() && !a.IsUndefined() {
			f, _ := m.jsNumber(a)
			switch {
			case math.IsNaN(f):
				endF = 0
			case f > float64(n):
				endF = float64(n)
			default:
				endF = math.Trunc(f)
			}
		}
		if endF <= startF {
			return m.e.JS.Str(""), nil
		}
		start, end := int(startF), int(endF)
		// V8 refuses to materialize strings past kStringMaxLength; JavaScriptCore
		// has its own limit, so report Node's error up front.
		if end-start > maxStringLength {
			return js.Value{}, js.ThrowCode("Error", "ERR_STRING_TOO_LONG",
				fmt.Sprintf("Cannot create a string longer than 0x%x characters", maxStringLength))
		}
		return m.e.JS.Str(decodeString(b.b[start:end], enc)), nil
	})
	p.SetFuncHidden("toJSON", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		data := m.e.JS.Arr(nil)
		for i, v := range b.b {
			data.SetAt(i, m.e.JS.Num(float64(v)))
		}
		obj := m.e.JS.Obj()
		obj.Set("type", m.e.JS.Str("Buffer"))
		obj.Set("data", data.V())
		return obj.V(), nil
	})
	p.SetFuncHidden("equals", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		other, err := m.bytesArg(arg(args, 0), "otherBuffer")
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(bytes.Equal(b.b, other)), nil
	})
	p.SetFuncHidden("compare", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		other, err := m.bytesArg(arg(args, 0), "target")
		if err != nil {
			return js.Value{}, err
		}
		// targetStart/sourceStart have no upper bound in Node (a start past the
		// end is a zero-length range, not an error); only the *End values are
		// range-checked against their buffer.
		tStart, err := m.intArg(arg(args, 1), 0, 0, maxLength, "targetStart")
		if err != nil {
			return js.Value{}, err
		}
		tEnd, err := m.intArg(arg(args, 2), len(other), 0, len(other), "targetEnd")
		if err != nil {
			return js.Value{}, err
		}
		sStart, err := m.intArg(arg(args, 3), 0, 0, maxLength, "sourceStart")
		if err != nil {
			return js.Value{}, err
		}
		sEnd, err := m.intArg(arg(args, 4), len(b.b), 0, len(b.b), "sourceEnd")
		if err != nil {
			return js.Value{}, err
		}
		ts, te := clampRange(len(other), tStart, tEnd)
		ss, se := clampRange(len(b.b), sStart, sEnd)
		return m.e.JS.Num(float64(bytes.Compare(b.b[ss:se], other[ts:te]))), nil
	})
	p.SetFuncHidden("copy", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, typeErr("The \"source\" argument must be an instance of Buffer or Uint8Array.%s", describeReceived(this.V()))
		}
		target := arg(args, 0)
		if !m.isArrayBufferView(target) {
			return js.Value{}, typeErr("The \"target\" argument must be an instance of Buffer or Uint8Array.%s", describeReceived(target))
		}
		data, has := target.Bytes()
		if !has {
			data = nil
		}
		tb := &buf{m: m, obj: target.Object(), b: data}
		// Node coerces the three offsets with ToNumber/toInteger: numeric
		// strings, null, arrays and floats all work, NaN means the default,
		// and a Symbol.toPrimitive hook's throw propagates.
		tStartF, err := m.toInteger(arg(args, 1), 0)
		if err != nil {
			return js.Value{}, err
		}
		sStartF, err := m.toInteger(arg(args, 2), 0)
		if err != nil {
			return js.Value{}, err
		}
		sEndF, err := m.toInteger(arg(args, 3), float64(len(b.b)))
		if err != nil {
			return js.Value{}, err
		}
		if tStartF < 0 {
			return js.Value{}, rangeNum("targetStart", ">= 0", tStartF)
		}
		if sStartF < 0 || sStartF > float64(len(b.b)) {
			return js.Value{}, rangeNum("sourceStart", fmt.Sprintf(">= 0 && <= %d", len(b.b)), sStartF)
		}
		if sEndF < 0 {
			return js.Value{}, rangeNum("sourceEnd", ">= 0", sEndF)
		}
		tStart, sStart, sEnd := int(tStartF), int(sStartF), int(sEndF)
		if tStart >= len(tb.b) || sStart >= sEnd {
			return m.e.JS.Num(0), nil
		}
		n := sEnd - sStart
		if n > len(tb.b)-tStart {
			n = len(tb.b) - tStart
		}
		if sStart+n > len(b.b) {
			n = len(b.b) - sStart
		}
		if n <= 0 {
			return m.e.JS.Num(0), nil
		}
		if err := m.setBytes(tb.obj, b.b[sStart:sStart+n], tStart); err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(n)), nil
	})
	p.SetFuncHidden("fill", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		// A Buffer whose .length was redefined must not let fill reach past the
		// real backing store; Node reports ERR_BUFFER_OUT_OF_BOUNDS instead.
		if n := this.Get("length"); n.IsNumber() && int(n.Int()) != len(b.b) {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_BUFFER_OUT_OF_BOUNDS",
				"Attempt to access memory outside buffer bounds")
		}
		valueArg := arg(args, 0)
		offsetArg := arg(args, 1)
		endArg := arg(args, 2)
		encArg := arg(args, 3)
		// Node fills with zero when the value is omitted or null.
		if !valueArg.Valid() || valueArg.IsUndefined() || valueArg.IsNull() {
			valueArg = m.e.JS.Num(0)
		}
		if valueArg.IsString() {
			// _fill shifts a string in the offset or end position into
			// encoding, exactly like Node does.
			if !offsetArg.Valid() || offsetArg.IsUndefined() || offsetArg.IsString() {
				encArg = offsetArg
				offsetArg = js.Value{}
				endArg = js.Value{}
			} else if endArg.Valid() && endArg.IsString() {
				encArg = endArg
				endArg = js.Value{}
			}
		} else {
			encArg = js.Value{}
		}
		pattern, err := m.fillPattern(valueArg, encArg)
		if err != nil {
			return js.Value{}, err
		}
		if len(b.b) == 0 {
			return b.obj.V(), nil
		}
		// Node's fill rejects negative or past-the-end offsets instead of
		// clamping them the way slice does.
		bound := func(v js.Value, def int, name string) (int, error) {
			if !v.Valid() || v.IsUndefined() {
				return def, nil
			}
			if !v.IsNumber() {
				return 0, typeErr(`The "`+name+`" argument must be of type number.%s`, describeReceived(v))
			}
			f := v.Float()
			if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) || f < 0 || f > float64(len(b.b)) {
				return 0, rangeErr(`The value of "`+name+`" is out of range. It must be >= 0 && <= %d. Received %s`,
					len(b.b), strconv.FormatFloat(f, 'g', -1, 64))
			}
			return int(f), nil
		}
		start, err := bound(offsetArg, 0, "offset")
		if err != nil {
			return js.Value{}, err
		}
		end, err := bound(endArg, len(b.b), "end")
		if err != nil {
			return js.Value{}, err
		}
		if end <= start {
			return b.obj.V(), nil
		}
		filled := make([]byte, end-start)
		for i := range filled {
			filled[i] = pattern[i%len(pattern)]
		}
		if err := m.setBytes(b.obj, filled, start); err != nil {
			return js.Value{}, err
		}
		return b.obj.V(), nil
	})
	p.SetFuncHidden("write", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		first := arg(args, 0)
		if !first.IsString() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"The \"string\" argument must be of type string."+describeReceived(first))
		}
		bufferLength := len(b.b)
		offsetArg, lengthArg, encArg := arg(args, 1), arg(args, 2), arg(args, 3)
		off := 0
		length := bufferLength
		if offsetArg.Valid() && !offsetArg.IsUndefined() {
			if (!lengthArg.Valid() || lengthArg.IsUndefined()) && offsetArg.IsString() {
				// Buffer#write(string, encoding)
				encArg = offsetArg
			} else {
				// Buffer#write(string, offset[, length][, encoding])
				offset, err := m.intArg(offsetArg, 0, 0, bufferLength, "offset")
				if err != nil {
					return js.Value{}, err
				}
				off = offset
				remaining := bufferLength - off
				if !lengthArg.Valid() || lengthArg.IsUndefined() {
					length = remaining
				} else if lengthArg.IsString() {
					encArg = lengthArg
					length = remaining
				} else {
					l, err := m.intArg(lengthArg, 0, 0, bufferLength, "length")
					if err != nil {
						return js.Value{}, err
					}
					if l > remaining {
						l = remaining
					}
					length = l
				}
			}
		}
		enc := "utf8"
		if encArg.Valid() && !encArg.IsUndefined() {
			if !encArg.IsString() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					"The \"encoding\" argument must be of type string."+describeReceived(encArg))
			}
			name, ok := normalizeEncoding(encArg.Str())
			if !ok {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", encArg.Str()))
			}
			enc = name
		}
		data, err := encodeString(first.Str(), enc)
		if err != nil {
			return js.Value{}, err
		}
		if length > len(data) {
			length = len(data)
		}
		if length < 0 {
			length = 0
		}
		// ucs2Write never splits a code unit: an odd byte budget loses one.
		if enc == "utf16le" {
			length &= ^1
		}
		data = data[:length]
		if err := m.setBytes(b.obj, data, off); err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(len(data))), nil
	})
	// Node keeps the per-encoding slice/write helpers on Buffer.prototype.
	// They are documented as internal but are observable and used by tests.
	sliceBytes := func(this js.Object, args []js.Value) ([]byte, error) {
		b, ok := m.self(this)
		if !ok {
			return nil, receiverErr("this is not a Buffer")
		}
		start, end := clampRange(len(b.b), optInt(arg(args, 0), 0), optInt(arg(args, 1), len(b.b)))
		return b.b[start:end], nil
	}
	for name, enc := range map[string]string{
		"asciiSlice":     "latin1",
		"latin1Slice":    "latin1",
		"utf8Slice":      "utf8",
		"hexSlice":       "hex",
		"base64Slice":    "base64",
		"base64urlSlice": "base64url",
		"ucs2Slice":      "utf16le",
	} {
		name, enc := name, enc
		p.SetFuncHidden(name, func(this js.Object, args []js.Value) (js.Value, error) {
			d, err := sliceBytes(this, args)
			if err != nil {
				return js.Value{}, err
			}
			return m.e.JS.Str(decodeString(d, enc)), nil
		})
	}
	writeBytes := func(this js.Object, args []js.Value, enc string) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		s := arg(args, 0)
		if !s.IsString() {
			return js.Value{}, typeErr(`The "string" argument must be of type string.%s`, describeReceived(s))
		}
		off := 0
		if v := arg(args, 1); v.Valid() && !v.IsUndefined() {
			off = int(v.Int())
		}
		n := len(b.b) - off
		if v := arg(args, 2); v.Valid() && !v.IsUndefined() {
			n = int(v.Int())
		}
		if off < 0 || n < 0 || off+n > len(b.b) {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_BUFFER_OUT_OF_BOUNDS", `"offset" is outside of buffer bounds`)
		}
		data, err := encodeString(s.Str(), enc)
		if err != nil {
			return js.Value{}, err
		}
		if n > len(data) {
			n = len(data)
		}
		if enc == "utf16le" {
			n &= ^1
		}
		copy(b.b[off:off+n], data[:n])
		return m.e.JS.Num(float64(n)), nil
	}
	for name, enc := range map[string]string{
		"asciiWrite":     "ascii",
		"latin1Write":    "latin1",
		"utf8Write":      "utf8",
		"hexWrite":       "hex",
		"base64Write":    "base64",
		"base64urlWrite": "base64url",
		"ucs2Write":      "utf16le",
	} {
		name, enc := name, enc
		p.SetFuncHidden(name, func(this js.Object, args []js.Value) (js.Value, error) {
			return writeBytes(this, args, enc)
		})
	}
	p.SetFuncHidden("slice", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		// Buffer#slice delegates to this.subarray in Node, so calling it on a
		// plain Uint8Array keeps that Uint8Array's own subarray and returns a
		// Uint8Array rather than converting it into a Buffer.
		if _, isBuf := m.asBuffer(this.V()); !isBuf {
			if sub := this.Get("subarray"); sub.IsFunction() {
				return sub.Object().Call(this, args...)
			}
		}
		start, end := clampRange(len(b.b), optInt(arg(args, 0), 0), optInt(arg(args, 1), len(b.b)))
		return m.viewRange(this, start, end, b.b)
	})
	p.SetFuncHidden("subarray", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		start, end := clampRange(len(b.b), optInt(arg(args, 0), 0), optInt(arg(args, 1), len(b.b)))
		return m.viewRange(this, start, end, b.b)
	})
	p.SetFuncHidden("includes", func(this js.Object, args []js.Value) (js.Value, error) {
		idx, err := m.search(this, args, true)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(idx >= 0), nil
	})
	p.SetFuncHidden("indexOf", func(this js.Object, args []js.Value) (js.Value, error) {
		idx, err := m.search(this, args, true)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(idx)), nil
	})
	p.SetFuncHidden("lastIndexOf", func(this js.Object, args []js.Value) (js.Value, error) {
		idx, err := m.search(this, args, false)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Num(float64(idx)), nil
	})
	p.SetFuncHidden("swap16", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.swap(this, 2)
	})
	p.SetFuncHidden("swap32", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.swap(this, 4)
	})
	p.SetFuncHidden("swap64", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.swap(this, 8)
	})
	p.SetFuncHidden("toLocaleString", func(this js.Object, args []js.Value) (js.Value, error) {
		b, ok := m.self(this)
		if !ok {
			return js.Value{}, receiverErr("this is not a Buffer")
		}
		return m.e.JS.Str(decodeString(b.b, "utf8")), nil
	})
	m.installReads(p)
	m.installWrites(p)
	// Node exposes both UInt and Uint spellings, and toLocaleString is the
	// very same function object as toString.
	p.Set("toLocaleString", p.Get("toString"))
	for _, name := range []string{
		"readUInt8", "readUInt16LE", "readUInt16BE", "readUInt32LE", "readUInt32BE",
		"readUIntLE", "readUIntBE", "readBigUInt64LE", "readBigUInt64BE",
		"writeUInt8", "writeUInt16LE", "writeUInt16BE", "writeUInt32LE", "writeUInt32BE",
		"writeUIntLE", "writeUIntBE", "writeBigUInt64LE", "writeBigUInt64BE",
	} {
		p.Set(strings.Replace(name, "UInt", "Uint", 1), p.Get(name))
	}
}

func (m *module) installReads(p js.Object) {
	read := func(fn func([]byte, int) js.Value) js.Func {
		return func(this js.Object, args []js.Value) (js.Value, error) {
			b, ok := m.self(this)
			if !ok {
				return js.Value{}, receiverErr("this is not a Buffer")
			}
			off, err := m.checkBounds(len(b.b), arg(args, 0), 1)
			if err != nil {
				return js.Value{}, err
			}
			return fn(b.b, off), nil
		}
	}
	readN := func(n int, fn func([]byte) js.Value) js.Func {
		return func(this js.Object, args []js.Value) (js.Value, error) {
			b, ok := m.self(this)
			if !ok {
				return js.Value{}, receiverErr("this is not a Buffer")
			}
			off, err := m.checkBounds(len(b.b), arg(args, 0), n)
			if err != nil {
				return js.Value{}, err
			}
			return fn(b.b[off : off+n]), nil
		}
	}
	readVar := func(signed, be bool) js.Func {
		return func(this js.Object, args []js.Value) (js.Value, error) {
			b, ok := m.self(this)
			if !ok {
				return js.Value{}, receiverErr("this is not a Buffer")
			}
			offArg := arg(args, 0)
			if !offArg.Valid() || offArg.IsUndefined() {
				return js.Value{}, typeErr("The \"offset\" argument must be of type number.%s", describeReceived(offArg))
			}
			n, err := m.checkByteLength(arg(args, 1))
			if err != nil {
				return js.Value{}, err
			}
			off, err := m.checkBounds(len(b.b), offArg, n)
			if err != nil {
				return js.Value{}, err
			}
			if be {
				return m.e.JS.Num(readIntegerBE(b.b[off:off+n], signed)), nil
			}
			return m.e.JS.Num(readInteger(b.b[off:off+n], signed)), nil
		}
	}

	p.SetFuncHidden("readUInt8", read(func(b []byte, off int) js.Value { return m.e.JS.Num(float64(b[off])) }))
	p.SetFuncHidden("readInt8", read(func(b []byte, off int) js.Value { return m.e.JS.Num(float64(int8(b[off]))) }))
	p.SetFuncHidden("readUInt16LE", readN(2, func(b []byte) js.Value { return m.e.JS.Num(float64(binary.LittleEndian.Uint16(b))) }))
	p.SetFuncHidden("readUInt16BE", readN(2, func(b []byte) js.Value { return m.e.JS.Num(float64(binary.BigEndian.Uint16(b))) }))
	p.SetFuncHidden("readUInt32LE", readN(4, func(b []byte) js.Value { return m.e.JS.Num(float64(binary.LittleEndian.Uint32(b))) }))
	p.SetFuncHidden("readUInt32BE", readN(4, func(b []byte) js.Value { return m.e.JS.Num(float64(binary.BigEndian.Uint32(b))) }))
	p.SetFuncHidden("readInt16LE", readN(2, func(b []byte) js.Value { return m.e.JS.Num(float64(int16(binary.LittleEndian.Uint16(b)))) }))
	p.SetFuncHidden("readInt16BE", readN(2, func(b []byte) js.Value { return m.e.JS.Num(float64(int16(binary.BigEndian.Uint16(b)))) }))
	p.SetFuncHidden("readInt32LE", readN(4, func(b []byte) js.Value { return m.e.JS.Num(float64(int32(binary.LittleEndian.Uint32(b)))) }))
	p.SetFuncHidden("readInt32BE", readN(4, func(b []byte) js.Value { return m.e.JS.Num(float64(int32(binary.BigEndian.Uint32(b)))) }))
	p.SetFuncHidden("readFloatLE", readN(4, func(b []byte) js.Value {
		return m.e.JS.Num(float64(math.Float32frombits(binary.LittleEndian.Uint32(b))))
	}))
	p.SetFuncHidden("readFloatBE", readN(4, func(b []byte) js.Value {
		return m.e.JS.Num(float64(math.Float32frombits(binary.BigEndian.Uint32(b))))
	}))
	p.SetFuncHidden("readDoubleLE", readN(8, func(b []byte) js.Value {
		return m.e.JS.Num(math.Float64frombits(binary.LittleEndian.Uint64(b)))
	}))
	p.SetFuncHidden("readDoubleBE", readN(8, func(b []byte) js.Value {
		return m.e.JS.Num(math.Float64frombits(binary.BigEndian.Uint64(b)))
	}))
	p.SetFuncHidden("readUIntLE", readVar(false, false))
	p.SetFuncHidden("readUIntBE", readVar(false, true))
	p.SetFuncHidden("readIntLE", readVar(true, false))
	p.SetFuncHidden("readIntBE", readVar(true, true))
	p.SetFuncHidden("readBigInt64BE", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.readBig(this, args, true, true)
	})
	p.SetFuncHidden("readBigInt64LE", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.readBig(this, args, false, true)
	})
	p.SetFuncHidden("readBigUInt64BE", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.readBig(this, args, true, false)
	})
	p.SetFuncHidden("readBigUInt64LE", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.readBig(this, args, false, false)
	})
}

func (m *module) installWrites(p js.Object) {
	write := func(n int, min, max float64, rangeSpec string, encode func(v float64) []byte) js.Func {
		return func(this js.Object, args []js.Value) (js.Value, error) {
			b, ok := m.self(this)
			if !ok {
				return js.Value{}, receiverErr("this is not a Buffer")
			}
			value := arg(args, 0)
			if err := m.checkInt(value, min, max, rangeSpec); err != nil {
				return js.Value{}, err
			}
			data := encode(value.Float())
			off, err := m.checkBounds(len(b.b), arg(args, 1), n)
			if err != nil {
				return js.Value{}, err
			}
			if err := m.setBytes(b.obj, data, off); err != nil {
				return js.Value{}, err
			}
			return m.e.JS.Num(float64(off + n)), nil
		}
	}
	writeVar := func(be, signed bool) js.Func {
		return func(this js.Object, args []js.Value) (js.Value, error) {
			b, ok := m.self(this)
			if !ok {
				return js.Value{}, receiverErr("this is not a Buffer")
			}
			n, err := m.checkByteLength(arg(args, 2))
			if err != nil {
				return js.Value{}, err
			}
			bits := n * 8
			var min, max float64
			var spec string
			if signed {
				if n > 4 {
					min = -math.Exp2(float64(bits - 1))
					// Node checks an inclusive maximum, so the exclusive
					// bound in the message is one less than 2 ** (bits-1).
					max = math.Exp2(float64(bits-1)) - 1
					spec = fmt.Sprintf(">= -(2 ** %d) and < 2 ** %d", bits-1, bits-1)
				} else {
					min = -math.Exp2(float64(bits - 1))
					max = math.Exp2(float64(bits-1)) - 1
					spec = fmt.Sprintf(">= %d and <= %d", int64(min), int64(max))
				}
			} else {
				if n > 4 {
					min = 0
					max = math.Exp2(float64(bits)) - 1
					spec = fmt.Sprintf(">= 0 and < 2 ** %d", bits)
				} else {
					min = 0
					max = math.Exp2(float64(bits)) - 1
					spec = fmt.Sprintf(">= 0 and <= %d", int64(max))
				}
			}
			value := arg(args, 0)
			if err := m.checkInt(value, min, max, spec); err != nil {
				return js.Value{}, err
			}
			offArg := arg(args, 1)
			if !offArg.Valid() || offArg.IsUndefined() {
				// The explicit-byteLength form has no default offset; Node's
				// checkBounds rejects a missing offset with a type error.
				return js.Value{}, typeErr("The \"offset\" argument must be of type number.%s", describeReceived(offArg))
			}
			off, err := m.checkBounds(len(b.b), offArg, n)
			if err != nil {
				return js.Value{}, err
			}
			data := writeInteger(value.Float(), n, be, signed)
			if err := m.setBytes(b.obj, data, off); err != nil {
				return js.Value{}, err
			}
			return m.e.JS.Num(float64(off + n)), nil
		}
	}

	p.SetFuncHidden("writeUInt8", write(1, 0, 0xff, ">= 0 and <= 255", func(v float64) []byte { return []byte{byte(uint64(v) & 0xff)} }))
	p.SetFuncHidden("writeInt8", write(1, -0x80, 0x7f, ">= -128 and <= 127", func(v float64) []byte { return []byte{byte(int64(v))} }))
	p.SetFuncHidden("writeUInt16LE", write(2, 0, 0xffff, ">= 0 and <= 65535", func(v float64) []byte {
		out := make([]byte, 2)
		binary.LittleEndian.PutUint16(out, uint16(uint64(v)))
		return out
	}))
	p.SetFuncHidden("writeUInt16BE", write(2, 0, 0xffff, ">= 0 and <= 65535", func(v float64) []byte {
		out := make([]byte, 2)
		binary.BigEndian.PutUint16(out, uint16(uint64(v)))
		return out
	}))
	p.SetFuncHidden("writeUInt32LE", write(4, 0, 0xffffffff, ">= 0 and <= 4294967295", func(v float64) []byte {
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, uint32(uint64(v)))
		return out
	}))
	p.SetFuncHidden("writeUInt32BE", write(4, 0, 0xffffffff, ">= 0 and <= 4294967295", func(v float64) []byte {
		out := make([]byte, 4)
		binary.BigEndian.PutUint32(out, uint32(uint64(v)))
		return out
	}))
	p.SetFuncHidden("writeInt16LE", write(2, -0x8000, 0x7fff, ">= -32768 and <= 32767", func(v float64) []byte {
		out := make([]byte, 2)
		binary.LittleEndian.PutUint16(out, uint16(int64(v)))
		return out
	}))
	p.SetFuncHidden("writeInt16BE", write(2, -0x8000, 0x7fff, ">= -32768 and <= 32767", func(v float64) []byte {
		out := make([]byte, 2)
		binary.BigEndian.PutUint16(out, uint16(int64(v)))
		return out
	}))
	p.SetFuncHidden("writeInt32LE", write(4, -0x80000000, 0x7fffffff, ">= -2147483648 and <= 2147483647", func(v float64) []byte {
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, uint32(int64(v)))
		return out
	}))
	p.SetFuncHidden("writeInt32BE", write(4, -0x80000000, 0x7fffffff, ">= -2147483648 and <= 2147483647", func(v float64) []byte {
		out := make([]byte, 4)
		binary.BigEndian.PutUint32(out, uint32(int64(v)))
		return out
	}))
	p.SetFuncHidden("writeFloatLE", write(4, math.Inf(-1), math.Inf(1), "", func(v float64) []byte {
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, math.Float32bits(float32(v)))
		return out
	}))
	p.SetFuncHidden("writeFloatBE", write(4, math.Inf(-1), math.Inf(1), "", func(v float64) []byte {
		out := make([]byte, 4)
		binary.BigEndian.PutUint32(out, math.Float32bits(float32(v)))
		return out
	}))
	p.SetFuncHidden("writeDoubleLE", write(8, math.Inf(-1), math.Inf(1), "", func(v float64) []byte {
		out := make([]byte, 8)
		binary.LittleEndian.PutUint64(out, math.Float64bits(v))
		return out
	}))
	p.SetFuncHidden("writeDoubleBE", write(8, math.Inf(-1), math.Inf(1), "", func(v float64) []byte {
		out := make([]byte, 8)
		binary.BigEndian.PutUint64(out, math.Float64bits(v))
		return out
	}))
	p.SetFuncHidden("writeUIntLE", writeVar(false, false))
	p.SetFuncHidden("writeUIntBE", writeVar(true, false))
	p.SetFuncHidden("writeIntLE", writeVar(false, true))
	p.SetFuncHidden("writeIntBE", writeVar(true, true))
	p.SetFuncHidden("writeBigInt64BE", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.writeBig(this, args, true, true)
	})
	p.SetFuncHidden("writeBigInt64LE", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.writeBig(this, args, false, true)
	})
	p.SetFuncHidden("writeBigUInt64BE", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.writeBig(this, args, true, false)
	})
	p.SetFuncHidden("writeBigUInt64LE", func(this js.Object, args []js.Value) (js.Value, error) {
		return m.writeBig(this, args, false, false)
	})
}

func (m *module) readBig(this js.Object, args []js.Value, be, signed bool) (js.Value, error) {
	b, ok := m.self(this)
	if !ok {
		return js.Value{}, receiverErr("this is not a Buffer")
	}
	off, err := m.checkBounds(len(b.b), arg(args, 0), 8)
	if err != nil {
		return js.Value{}, err
	}
	var u uint64
	if be {
		u = binary.BigEndian.Uint64(b.b[off:])
	} else {
		u = binary.LittleEndian.Uint64(b.b[off:])
	}
	return m.bigintValue(u, signed), nil
}

func (m *module) writeBig(this js.Object, args []js.Value, be, signed bool) (js.Value, error) {
	b, ok := m.self(this)
	if !ok {
		return js.Value{}, receiverErr("this is not a Buffer")
	}
	value, err := m.bigintArg(arg(args, 0), signed)
	if err != nil {
		return js.Value{}, err
	}
	off, err := m.checkBounds(len(b.b), arg(args, 1), 8)
	if err != nil {
		return js.Value{}, err
	}
	dst := b.b[off : off+8]
	if be {
		binary.BigEndian.PutUint64(dst, value)
	} else {
		binary.LittleEndian.PutUint64(dst, value)
	}
	if err := m.setBytes(b.obj, dst, off); err != nil {
		return js.Value{}, err
	}
	return m.e.JS.Num(float64(off + 8)), nil
}

func (m *module) bigintValue(u uint64, signed bool) js.Value {
	s := strconv.FormatUint(u, 10)
	if signed {
		s = strconv.FormatInt(int64(u), 10)
	}
	fn := m.e.Global().Get("BigInt")
	if fn.IsFunction() {
		v, err := fn.Object().Call(js.Object{}, m.e.JS.Str(s))
		if err == nil {
			return v
		}
	}
	return m.e.JS.Num(float64(u))
}

func (m *module) bigintArg(v js.Value, signed bool) (uint64, error) {
	if !v.Valid() || v.IsUndefined() || !v.IsBigInt() {
		return 0, typeErr("The \"value\" argument must be of type bigint.%s", describeReceived(v))
	}
	s := v.Str()
	if s == "" {
		return 0, typeErr("The \"value\" argument must be of type bigint.%s", describeReceived(v))
	}
	if signed {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, outOfRange("value", ">= -(2n ** 63n) and < 2n ** 63n", v)
		}
		return uint64(i), nil
	}
	if strings.HasPrefix(s, "-") {
		return 0, outOfRange("value", ">= 0n and < 2n ** 64n", v)
	}
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, outOfRange("value", ">= 0n and < 2n ** 64n", v)
	}
	return u, nil
}

func (m *module) installSlowBuffer() {
	holder := m.e.JS.Obj()
	holder.SetFunc("f", func(this js.Object, args []js.Value) (js.Value, error) {
		n, err := m.sizeArg(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		return m.newBuffer(make([]byte, n)).obj.V(), nil
	})
	m.e.Global().Set("__godeSlowBuffer", holder.Get("f"))
	// A host function cannot be used with `new`, so define the constructor in
	// JavaScript and delegate the allocation back to the host.
	v, err := m.e.JS.Eval(`(function () {
  function SlowBuffer(size) { return __godeSlowBuffer(size); }
  Object.setPrototypeOf(SlowBuffer.prototype, Uint8Array.prototype);
  return SlowBuffer;
})()`, "[gode:buffer/slow]")
	if err != nil {
		panic("buffer: SlowBuffer prelude: " + err.Error())
	}
	m.exports.Set("SlowBuffer", v)
}

// installBlob defines Node's Blob and File classes. They are plain JS classes
// over copied bytes; the stream methods are intentionally absent until GodeJS
// has a web ReadableStream implementation.
func (m *module) installBlob() {
	v, err := m.e.JS.Eval(blobPrelude, "[gode:buffer/blob]")
	if err != nil {
		panic("buffer: Blob prelude: " + err.Error())
	}
	out := v.Object()
	m.exports.Set("Blob", out.Get("Blob"))
	m.exports.Set("File", out.Get("File"))
	m.e.Global().Set("Blob", out.Get("Blob"))
	m.e.Global().Set("File", out.Get("File"))
}

// installInspect defines INSPECT_MAX_BYTES as a validated accessor and adds
// Buffer.prototype.inspect plus the util.inspect.custom hook. The formatting
// matches lib/buffer.js, including the "... N more bytes" suffix and extra
// own properties.
func (m *module) installInspect(exports js.Object) {
	v, err := m.e.JS.Eval(inspectPrelude, "[gode:buffer/inspect]")
	if err != nil {
		panic("buffer: inspect prelude: " + err.Error())
	}
	if _, err := v.Object().Call(js.Object{}, exports.V(), m.proto.V()); err != nil {
		panic("buffer: inspect: " + err.Error())
	}
}

const inspectPrelude = `(function (exports, proto) {
  'use strict';
  var max = 50;
  Object.defineProperty(exports, 'INSPECT_MAX_BYTES', {
    enumerable: true, configurable: true,
    get: function () { return max; },
    set: function (value) {
      if (typeof value !== 'number') {
        var typeErr = new TypeError('The "INSPECT_MAX_BYTES" argument must be of type number. Received ' +
          (value === null ? 'null' : 'type ' + typeof value));
        typeErr.code = 'ERR_INVALID_ARG_TYPE';
        throw typeErr;
      }
      if (value < 0 || value !== value) {
        var rangeErr = new RangeError('The value of "INSPECT_MAX_BYTES" is out of range. ' +
          'It must be >= 0. Received ' + String(value));
        rangeErr.code = 'ERR_OUT_OF_RANGE';
        throw rangeErr;
      }
      max = value;
      try { globalThis.__godeInspectMaxBytes = max; } catch (ignored) {}
    }
  });
  try { globalThis.__godeInspectMaxBytes = max; } catch (ignored) {}
  function inspectBuffer(recurseTimes, ctx, inspectFn) {
    var len = this.length;
    var use = Math.min(max, len);
    var parts = [];
    for (var i = 0; i < use; i++) {
      var h = this[i].toString(16);
      parts.push(h.length === 1 ? '0' + h : h);
    }
    var str = parts.join(' ');
    var remaining = len - max;
    if (remaining > 0) {
      if (str) str += ' ';
      str += '... ' + remaining + ' more byte' + (remaining > 1 ? 's' : '');
    }
    if (ctx) {
      var extras = [];
      var keys = Object.keys(this);
      for (var k = 0; k < keys.length; k++) {
        var key = keys[k];
        if (/^(0|[1-9][0-9]*)$/.test(key)) continue;
        var value = this[key];
        var rendered;
        if (typeof inspectFn === 'function') {
          try {
            rendered = inspectFn(value, { depth: 0, colors: false, compact: true, breakLength: Infinity });
          } catch (e) {
            rendered = String(value);
          }
        } else {
          rendered = String(value);
        }
        extras.push(key + ': ' + rendered);
      }
      if (extras.length) {
        if (str) str += ', ';
        str += extras.join(', ');
      }
    }
    var name = 'Buffer';
    try {
      var ctor = this.constructor;
      if (typeof ctor === 'function' && typeof ctor.name === 'string' && ctor.name) name = ctor.name;
    } catch (ignored) { /* use the default name */ }
    return '<' + name + ' ' + str + '>';
  }
  var custom = Symbol.for('nodejs.util.inspect.custom');
  Object.defineProperty(proto, 'inspect', {
    value: inspectBuffer, writable: true, configurable: true, enumerable: false
  });
  Object.defineProperty(proto, custom, {
    value: inspectBuffer, writable: true, configurable: true
  });
})`

func (m *module) installLegacyGetters() {
	_, _ = m.e.JS.Eval(`(function () {
  // Buffer.prototype is a plain object here (not a Uint8Array like Node's),
  // so reading these on the prototype must answer undefined rather than
  // letting the engine throw "Receiver should be a typed array view".
  function isView(value) {
    return typeof ArrayBuffer === 'function' &&
      typeof ArrayBuffer.isView === 'function' && ArrayBuffer.isView(value);
  }
  Object.defineProperty(Buffer.prototype, 'parent', {
    enumerable: true, configurable: true,
    get: function () { return isView(this) ? this.buffer : undefined; },
  });
  Object.defineProperty(Buffer.prototype, 'offset', {
    enumerable: true, configurable: true,
    get: function () { return isView(this) ? this.byteOffset : undefined; },
  });
})()`, "[gode:buffer/legacy]")
}

func (m *module) transcode(args []js.Value) (js.Value, error) {
	src, ok := arg(args, 0).Bytes()
	if !ok {
		return js.Value{}, typeErr("The \"source\" argument must be an instance of Buffer or Uint8Array.%s", describeReceived(arg(args, 0)))
	}
	if len(src) == 0 {
		return m.newBuffer(nil).obj.V(), nil
	}
	from, ok := normalizeEncoding(arg(args, 1).Str())
	if !ok {
		return js.Value{}, transcodeErr()
	}
	to, ok := normalizeEncoding(arg(args, 2).Str())
	if !ok {
		return js.Value{}, transcodeErr()
	}
	text := decodeString(src, from)
	out, err := encodeString(text, to)
	if err != nil {
		return js.Value{}, transcodeErr()
	}
	return m.newBuffer(out).obj.V(), nil
}

func transcodeErr() error {
	return js.ThrowCode("Error", "U_ILLEGAL_ARGUMENT_ERROR",
		"Unable to transcode Buffer [U_ILLEGAL_ARGUMENT_ERROR]")
}

const blobPrelude = `(function () {
  'use strict';
  const disallowed = /[^\u0020-\u007E]/u;

  function sourceBytes(source, endings) {
    if (source instanceof Blob) return source._bytes.slice();
    if (source instanceof ArrayBuffer) return new Uint8Array(source.slice(0));
    if (ArrayBuffer.isView(source)) {
      const copy = source.buffer.slice(source.byteOffset, source.byteOffset + source.byteLength);
      return new Uint8Array(copy);
    }
    let s = String(source);
    if (endings === 'native') s = s.replace(/\r\n|\r|\n/g, '\n');
    return new Uint8Array(Buffer.from(s, 'utf8'));
  }

  function normalizeType(type) {
    type = String(type);
    return disallowed.test(type) ? '' : type.toLowerCase();
  }

  // singleChunkStream adapts a blob body to the ReadableStream surface tests
  // use: getReader().read() and async iteration both yield one value.
  function singleChunkStream(value) {
    let sent = false;
    const stream = {
      getReader: function () {
        return {
          read: function () {
            if (sent) return Promise.resolve({ done: true, value: undefined });
            sent = true;
            return Promise.resolve({ done: false, value: value });
          },
          cancel: function () { sent = true; return Promise.resolve(); },
          releaseLock: function () {}
        };
      },
      cancel: function () { sent = true; return Promise.resolve(); }
    };
    if (typeof Symbol === 'function' && Symbol.asyncIterator) {
      stream[Symbol.asyncIterator] = function () {
        let done = false;
        return {
          next: function () {
            if (done) return Promise.resolve({ done: true, value: undefined });
            done = true;
            return Promise.resolve({ done: false, value: value });
          }
        };
      };
    }
    return stream;
  }

  class Blob {
    constructor(sources = [], options = {}) {
      if (!Array.isArray(sources)) {
        throw new TypeError('The \"sources\" argument must be an array');
      }
      if (options !== null && options !== undefined && typeof options !== 'object') {
        throw new TypeError('The \"options\" argument must be of type object');
      }
      const opts = options || {};
      const endings = opts.endings === undefined ? 'transparent' : String(opts.endings);
      if (endings !== 'transparent' && endings !== 'native') {
        throw new TypeError('The \"options.endings\" argument must be one of: transparent, native');
      }
      let total = 0;
      const parts = [];
      for (const source of sources) {
        const bytes = sourceBytes(source, endings);
        parts.push(bytes);
        total += bytes.length;
      }
      const bytes = new Uint8Array(total);
      let offset = 0;
      for (const part of parts) {
        bytes.set(part, offset);
        offset += part.length;
      }
      this._bytes = bytes;
      this._type = normalizeType(opts.type === undefined ? '' : opts.type);
    }

    get size() { return this._bytes.length; }
    get type() { return this._type; }

    slice(start = 0, end = this.size, contentType = '') {
      start = Number(start);
      end = Number(end);
      if (Number.isNaN(start)) start = 0;
      if (Number.isNaN(end)) end = this.size;
      start = Math.trunc(start);
      end = Math.trunc(end);
      if (start < 0) start = Math.max(this.size + start, 0);
      else start = Math.min(start, this.size);
      if (end < 0) end = Math.max(this.size + end, 0);
      else end = Math.min(end, this.size);
      const span = Math.max(end - start, 0);
      return new Blob([this._bytes.subarray(start, start + span)], { type: contentType });
    }

    arrayBuffer() {
      return Promise.resolve(this._bytes.slice().buffer);
    }

    bytes() {
      return Promise.resolve(this._bytes.slice());
    }

    text() {
      return this.arrayBuffer().then((buffer) => Buffer.from(buffer).toString('utf8'));
    }

    stream() {
      return singleChunkStream(this._bytes.slice());
    }

    textStream() {
      return singleChunkStream(Buffer.from(this._bytes).toString('utf8'));
    }
  }

  Object.defineProperty(Blob.prototype, Symbol.toStringTag, { value: 'Blob' });

  class File extends Blob {
    constructor(fileBits, fileName, options = {}) {
      if (arguments.length < 2) {
        throw new TypeError('The \"fileBits\" and \"fileName\" arguments are required');
      }
      super(fileBits, options);
      const opts = options || {};
      let lastModified = opts.lastModified;
      if (lastModified === undefined) {
        lastModified = Date.now();
      } else {
        lastModified = Number(lastModified);
        if (Number.isNaN(lastModified)) lastModified = 0;
      }
      this._name = String(fileName);
      this._lastModified = lastModified;
    }
    get name() { return this._name; }
    get lastModified() { return this._lastModified; }
  }

  Object.defineProperty(File.prototype, Symbol.toStringTag, { value: 'File' });
  return { Blob, File };
})()`

func (m *module) newBuffer(data []byte) *buf {
	obj := m.e.JS.Bytes(data)
	obj.SetHidden("_godeBuffer", m.e.JS.Bool(true))
	obj.SetProto(m.proto)
	live, _ := obj.V().ByteView()
	if live == nil {
		// A zero-length or exotic view still needs a usable slice.
		live = []byte{}
	}
	return &buf{m: m, obj: obj, b: live}
}

// self resolves the receiver of a Buffer prototype method. Node's methods are
// generic: the generic-methods test calls them with a plain Uint8Array, so a
// value that is not a Buffer but still exposes typed-array bytes is accepted.
func (m *module) self(this js.Object) (*buf, bool) {
	if b, ok := m.asBuffer(this.V()); ok {
		return b, true
	}
	data, ok := this.V().Bytes()
	if !ok {
		return nil, false
	}
	if data == nil {
		data = []byte{}
	}
	return &buf{m: m, obj: this, b: data}, true
}

func (m *module) asBuffer(v js.Value) (*buf, bool) {
	if !v.IsObject() {
		return nil, false
	}
	o := v.Object()
	if !o.Get("_godeBuffer").Bool() {
		return nil, false
	}
	// Alias the live bytes: a write through buf[i] in JavaScript must be
	// visible to the Go-side operations and the other way round.
	data, ok := v.ByteView()
	if !ok {
		return nil, false
	}
	if data == nil {
		data = []byte{}
	}
	return &buf{m: m, obj: o, b: data}, true
}

func (m *module) bytesArg(v js.Value, name string) ([]byte, error) {
	if b, ok := m.asBuffer(v); ok {
		return b.b, nil
	}
	if data, ok := v.Bytes(); ok {
		return data, nil
	}
	return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		"The \""+name+"\" argument must be an instance of Buffer or Uint8Array."+describeReceived(v))
}

// sizeRange is the range Node prints for every ERR_OUT_OF_RANGE("size").
const sizeRange = ">= 0 && <= 9007199254740991"

func (m *module) sizeArg(args []js.Value, i int) (int, error) {
	v := arg(args, i)
	if !v.Valid() || v.IsUndefined() || !v.IsNumber() {
		return 0, typeErr("The \"size\" argument must be of type number.%s", describeReceived(v))
	}
	f := v.Float()
	// Node truncates a fractional size instead of rejecting it, but refuses
	// anything negative, NaN, Infinite or past MAX_SAFE_INTEGER.
	if f < 0 || math.IsNaN(f) || math.IsInf(f, 0) || f > 9007199254740991 {
		return 0, outOfRange("size", sizeRange, v)
	}
	if f > maxLength {
		return 0, outOfRange("size", sizeRange, v)
	}
	return int(f), nil
}

// intArg validates an optional integer argument the way Node's
// validateInteger does: undefined selects the default, a non-number is an
// ERR_INVALID_ARG_TYPE TypeError, and a non-integer or out-of-range value is an
// ERR_OUT_OF_RANGE RangeError. Buffer methods must reject these instead of
// silently coercing, or callers cannot tell a typo from a valid offset.
func (m *module) intArg(v js.Value, def, min, max int, name string) (int, error) {
	if !v.Valid() || v.IsUndefined() {
		return def, nil
	}
	if !v.IsNumber() {
		return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			"The \""+name+"\" argument must be of type number."+describeReceived(v))
	}
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf("The value of \"%s\" is out of range. It must be an integer. Received %s",
				name, strconv.FormatFloat(f, 'g', -1, 64)))
	}
	n := int(f)
	if n < min || n > max {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			fmt.Sprintf("The value of \"%s\" is out of range. It must be >= %d && <= %d. Received %d",
				name, min, max, n))
	}
	return n, nil
}

// fromString encodes s with the requested encoding argument, validating it
// the way Node does for Buffer.from(string, encoding).
func (m *module) fromString(s string, encArg js.Value) (js.Value, error) {
	// Node only validates the encoding when it is a string: any other type
	// silently means utf8, and only an unknown string encoding throws.
	enc := "utf8"
	if a := encArg; a.Valid() && !a.IsUndefined() && a.IsString() {
		name, ok := normalizeEncoding(a.Str())
		if !ok {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", a.Str()))
		}
		enc = name
	}
	data, err := encodeString(s, enc)
	if err != nil {
		return js.Value{}, err
	}
	return m.newBuffer(data).obj.V(), nil
}

func (m *module) from(args []js.Value) (js.Value, error) {
	v := arg(args, 0)
	switch {
	case v.IsString():
		return m.fromString(v.Str(), arg(args, 1))
	case v.IsArray():
		n := v.Object().Length()
		data := make([]byte, n)
		for i := 0; i < n; i++ {
			f, _ := numberOf(m.e, v.Object().At(i))
			data[i] = byteValue(f)
		}
		return m.newBuffer(data).obj.V(), nil
	}
	if _, ok := m.asBuffer(v); ok {
		data, _ := v.Bytes()
		return m.newBuffer(data).obj.V(), nil
	}
	if m.isAnyArrayBuffer(v) {
		// Buffer.from(arrayBuffer) is a view: writes are visible through both.
		// A faked prototype still reaches this branch, so require a real
		// byteLength and let newView reject anything that is not an ArrayBuffer.
		size := v.Object().Get("byteLength")
		if !size.IsNumber() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"The first argument must be of type string or an instance of Buffer, ArrayBuffer, or Array or an Array-like Object."+describeReceived(v))
		}
		total := size.Float()
		off := 0.0
		if a := arg(args, 1); a.Valid() && !a.IsUndefined() {
			if n, ok := numberOf(m.e, a); ok {
				off = math.Trunc(n)
			}
		}
		length := total - off
		autoLength := true
		if a := arg(args, 2); a.Valid() && !a.IsUndefined() {
			autoLength = false
			length = 0
			if n, ok := numberOf(m.e, a); ok {
				length = math.Trunc(n)
			}
		}
		if off < 0 || off > total {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_BUFFER_OUT_OF_BOUNDS", "\"offset\" is outside of buffer bounds")
		}
		if length < 0 || off+length > total {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_BUFFER_OUT_OF_BOUNDS", "\"length\" is outside of buffer bounds")
		}
		viewLen := int(length)
		if autoLength {
			// Omitting the length gives a length-tracking view, so a resizable
			// ArrayBuffer's Buffer follows resize() the way Node's does.
			viewLen = -1
		}
		view, err := m.newView(v, int(off), viewLen)
		if err != nil {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"The first argument must be of type string or an instance of Buffer, ArrayBuffer, or Array or an Array-like Object."+describeReceived(v))
		}
		return view.V(), nil
	}
	if v.IsObject() && !v.IsFunction() {
		o := v.Object()
		// A boxed primitive (new String(...)) or an object whose valueOf
		// returns a string or a different object converts first, as in Node.
		if valueOf := o.Get("valueOf"); valueOf.IsFunction() {
			r, err := valueOf.Object().Call(o)
			if err != nil {
				return js.Value{}, err
			}
			if r.IsString() {
				return m.fromString(r.Str(), arg(args, 1))
			}
			if r.IsObject() && !r.Same(v) {
				return m.from([]js.Value{r, arg(args, 1), arg(args, 2)})
			}
		}
		if o.Get("type").Str() == "Buffer" {
			data := o.Get("data")
			if data.IsArray() {
				n := data.Object().Length()
				out := make([]byte, n)
				for i := 0; i < n; i++ {
					out[i] = byteValue(data.Object().At(i).Float())
				}
				return m.newBuffer(out).obj.V(), nil
			}
		}
		// Node's fromObject: an object with a length is an array-like (a
		// non-number length means empty), and an object carrying an
		// ArrayBuffer in .buffer is an empty view when it has no length.
		ln := o.Get("length")
		if ln.Valid() && !ln.IsUndefined() {
			if !ln.IsNumber() {
				return m.newBuffer(nil).obj.V(), nil
			}
			f := math.Trunc(ln.Float())
			if math.IsNaN(f) || f <= 0 {
				return m.newBuffer(nil).obj.V(), nil
			}
			n := int(f)
			out := make([]byte, n)
			for i := 0; i < n; i++ {
				v, _ := numberOf(m.e, o.At(i))
				out[i] = byteValue(v)
			}
			return m.newBuffer(out).obj.V(), nil
		}
		if bufv := o.Get("buffer"); bufv.Valid() && m.isAnyArrayBuffer(bufv) {
			return m.newBuffer(nil).obj.V(), nil
		}
	}
	// A Symbol.toPrimitive returning a string converts last, as in Node; a
	// throw inside the user hook propagates instead of becoming an arg error.
	if v.IsObject() && !v.IsFunction() && m.toPrimitive.Valid() {
		r, err := m.toPrimitive.Object().CallWith([]js.Value{v})
		if err != nil {
			return js.Value{}, err
		}
		if r.IsString() {
			return m.fromString(r.Str(), arg(args, 1))
		}
	}
	return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		"The first argument must be of type string or an instance of Buffer, ArrayBuffer, or Array or an Array-like Object."+describeReceived(v))
}

// newView wraps a window of an ArrayBuffer as a Buffer without copying, so the
// memory stays shared with the original typed array or ArrayBuffer.
// newView wraps a window of an ArrayBuffer without copying. A negative length
// omits the length argument, producing a length-tracking view for resizable
// ArrayBuffers.
func (m *module) newView(ab js.Value, off, length int) (js.Object, error) {
	ctor := m.e.Global().Get("Uint8Array")
	construct := m.e.Global().Get("Reflect").Object().Get("construct")
	bits := []js.Value{ab, m.e.JS.Num(float64(off))}
	if length >= 0 {
		bits = append(bits, m.e.JS.Num(float64(length)))
	}
	args := m.e.JS.Arr(bits)
	res, err := construct.Object().Call(js.Object{}, ctor, args.V())
	if err != nil {
		return js.Object{}, err
	}
	obj := res.Object()
	obj.SetHidden("_godeBuffer", m.e.JS.Bool(true))
	obj.SetProto(m.proto)
	return obj, nil
}

// viewRange returns a Buffer view over this buffer's bytes. Node's slice and
// subarray share memory with the original.
func (m *module) viewRange(this js.Object, start, end int, bytes []byte) (js.Value, error) {
	backing := this.Get("buffer")
	base := optInt(this.Get("byteOffset"), 0)
	if backing.IsObject() && (backing.Tag() == "ArrayBuffer" || backing.Tag() == "SharedArrayBuffer") {
		view, err := m.newView(backing, base+start, end-start)
		if err == nil {
			return view.V(), nil
		}
	}
	// Without a usable ArrayBuffer fall back to a copy, which is still a Buffer.
	out := make([]byte, end-start)
	copy(out, bytes[start:end])
	return m.newBuffer(out).obj.V(), nil
}

// searchOffset clamps a search offset the way V8 does for indexOf/includes/
// lastIndexOf: infinities saturate, NaN falls back to the default, and the
// remaining values are truncated and clamped into [-length, length].
func searchOffset(v js.Value, def, length int) int {
	if !v.Valid() || v.IsUndefined() || !v.IsNumber() {
		return def
	}
	f := v.Float()
	switch {
	case math.IsNaN(f):
		return def
	case math.IsInf(f, 1):
		return length
	case math.IsInf(f, -1):
		return -length
	}
	if f >= float64(length) {
		return length
	}
	if f <= float64(-length) {
		return -length
	}
	return int(math.Trunc(f))
}

// offsetArg validates a read/write offset the way Node's checkBounds does:
// undefined means zero, a non-number is an ERR_INVALID_ARG_TYPE TypeError, and
// a fractional or infinite value is the "must be an integer" RangeError.
func (m *module) offsetArg(v js.Value) (int, error) {
	if !v.Valid() || v.IsUndefined() {
		return 0, nil
	}
	if !v.IsNumber() {
		return 0, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			"The \"offset\" argument must be of type number."+describeReceived(v))
	}
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
			"The value of \"offset\" is out of range. It must be an integer. Received "+
				strconv.FormatFloat(f, 'g', -1, 64))
	}
	return int(f), nil
}

// boundsError mirrors Node's internal/buffer.js boundsError. room is the
// largest offset that still fits the element (length - byteLength); a negative
// room means the buffer is too small for the element at all, which Node
// reports as ERR_BUFFER_OUT_OF_BOUNDS instead of an offset problem.
func (m *module) boundsError(off, room int) error {
	if room < 0 {
		return js.ThrowCode("RangeError", "ERR_BUFFER_OUT_OF_BOUNDS",
			"Attempt to access memory outside buffer bounds")
	}
	return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
		fmt.Sprintf("The value of \"offset\" is out of range. It must be >= 0 and <= %d. Received %d", room, off))
}

// checkBounds is a port of Node's checkBounds/boundsError pair. Node probes
// buf[offset] and buf[offset+byteLength-1]; a non-integer, negative, infinite
// or too-large offset falls through to the shared error text.
func (m *module) checkBounds(bufLen int, v js.Value, byteLength int) (int, error) {
	if !v.Valid() || v.IsUndefined() {
		v = m.e.JS.Num(0)
	}
	if !v.IsNumber() {
		return 0, typeErr("The \"offset\" argument must be of type number.%s", describeReceived(v))
	}
	f := v.Float()
	room := bufLen - byteLength
	if math.IsNaN(f) || (!math.IsInf(f, 0) && f != math.Trunc(f)) {
		return 0, outOfRange("offset", "an integer", v)
	}
	if math.IsInf(f, 0) || f < 0 || f > float64(room) {
		return 0, m.boundsErrorAt(v, room)
	}
	return int(f), nil
}

// checkByteLength validates the explicit byteLength of read*LE/BE and
// write*LE/BE. Node dispatches on 1..6 and reports every other value through
// boundsError(byteLength, 6, 'byteLength').
func (m *module) checkByteLength(v js.Value) (int, error) {
	if !v.Valid() || v.IsUndefined() || !v.IsNumber() {
		return 0, typeErr("The \"byteLength\" argument must be of type number.%s", describeReceived(v))
	}
	f := v.Float()
	if math.IsNaN(f) || (!math.IsInf(f, 0) && f != math.Trunc(f)) {
		return 0, outOfRange("byteLength", "an integer", v)
	}
	if math.IsInf(f, 0) || f < 1 || f > 6 {
		return 0, outOfRange("byteLength", ">= 1 and <= 6", v)
	}
	return int(f), nil
}

// boundsError mirrors Node's internal/buffer.js boundsError(value, length):
// non-integers name the integer requirement, a negative room means the buffer
// is too small, and everything else reports the valid range.
func (m *module) boundsErrorAt(v js.Value, room int) error {
	if v.IsNumber() {
		f := v.Float()
		if math.IsNaN(f) || (!math.IsInf(f, 0) && f != math.Trunc(f)) {
			return outOfRange("offset", "an integer", v)
		}
	}
	if room < 0 {
		return js.ThrowCode("RangeError", "ERR_BUFFER_OUT_OF_BOUNDS",
			"Attempt to access memory outside buffer bounds")
	}
	return outOfRange("offset", fmt.Sprintf(">= 0 and <= %d", room), v)
}

// checkInt mirrors Node's checkInt: it rejects out-of-range values with the
// range text the C++ helpers use before the offset is even examined.
func (m *module) checkInt(value js.Value, min, max float64, rangeSpec string) error {
	f := value.Float()
	if math.IsNaN(f) {
		return nil
	}
	if f > max || f < min {
		return outOfRange("value", rangeSpec, value)
	}
	return nil
}

// jsNumber coerces a value with the global Number() (which invokes valueOf and
// toString hooks, and lets their throws propagate).
func (m *module) jsNumber(v js.Value) (float64, error) {
	if v.IsNumber() {
		return v.Float(), nil
	}
	num := m.e.Global().Get("Number")
	if !num.IsFunction() {
		return 0, nil
	}
	r, err := num.Object().Call(js.Object{}, v)
	if err != nil {
		return 0, err
	}
	return r.Float(), nil
}

// adjustOffset is Node's adjustOffset from lib/buffer.js, used by toString and
// the range accessors: Math.trunc, NaN means 0, negatives count from the end
// and then clamp, and values past the length clamp to it.
func (m *module) adjustOffset(v js.Value, length int) int {
	f, err := m.jsNumber(v)
	if err != nil {
		return 0
	}
	f = math.Trunc(f)
	if f == 0 {
		return 0
	}
	if math.IsNaN(f) {
		return 0
	}
	if f < 0 {
		f += float64(length)
		if f < 0 {
			return 0
		}
		return int(f)
	}
	if f < float64(length) {
		return int(f)
	}
	return length
}

// toInteger converts a value the way Node's internal toInteger does: Number()
// coercion (so numeric strings and arrays work), NaN or an unsafe magnitude
// falls back to def, and fractions floor. A user hook's throw propagates.
func (m *module) toInteger(v js.Value, def float64) (float64, error) {
	if !v.Valid() || v.IsUndefined() {
		return def, nil
	}
	if v.IsNumber() {
		f := v.Float()
		if math.IsNaN(f) || f < -9007199254740991 || f > 9007199254740991 {
			return def, nil
		}
		if f == math.Trunc(f) {
			return f, nil
		}
		return math.Floor(f), nil
	}
	num := m.e.Global().Get("Number")
	if !num.IsFunction() {
		return def, nil
	}
	r, err := num.Object().Call(js.Object{}, v)
	if err != nil {
		return 0, err
	}
	f := r.Float()
	if math.IsNaN(f) || f < -9007199254740991 || f > 9007199254740991 {
		return def, nil
	}
	if f == math.Trunc(f) {
		return f, nil
	}
	return math.Floor(f), nil
}

// rangeNum builds ERR_OUT_OF_RANGE for a value that has already been coerced
// to a number, which is what Node reports for copy offsets.
func rangeNum(name, spec string, f float64) error {
	v := strconv.FormatFloat(f, 'f', -1, 64)
	return rangeErr("The value of %q is out of range. It must be %s. Received %s", name, spec, v)
}

// numberOf converts a value as JS Number() would, reporting false for NaN.
func numberOf(e *env.Env, v js.Value) (float64, bool) {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return 0, false
	}
	if v.IsNumber() {
		f := v.Float()
		return f, !math.IsNaN(f)
	}
	num := e.Global().Get("Number")
	if num.IsFunction() {
		if r, err := num.Object().Call(js.Object{}, v); err == nil && r.IsNumber() {
			f := r.Float()
			return f, !math.IsNaN(f)
		}
	}
	return 0, false
}

// byteValue narrows a JS number to a byte the way Node's Buffer assignment
// does: non-numbers become zero and the rest wraps modulo 256.
func byteValue(f float64) byte {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	n := int64(math.Trunc(f)) % 256
	if n < 0 {
		n += 256
	}
	return byte(n)
}

// describeReceived renders Node's "Received ..." clause for a bad argument.
func describeReceived(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return " Received undefined"
	}
	if v.IsNull() {
		return " Received null"
	}
	if v.IsFunction() {
		return " Received function " + v.Object().Get("name").Str()
	}
	if v.IsObject() {
		o := v.Object()
		if o.IsNullPrototype() {
			// Without a constructor Node falls back to inspecting the value.
			return " Received [Object: null prototype] {}"
		}
		if name := o.Get("constructor").Object().Get("name").Str(); name != "" {
			return " Received an instance of " + name
		}
		return " Received an instance of Object"
	}
	switch {
	case v.IsNumber():
		return " Received type number (" + strconv.FormatFloat(v.Float(), 'g', -1, 64) + ")"
	case v.IsBool():
		return " Received type boolean (" + strconv.FormatBool(v.Bool()) + ")"
	case v.IsString():
		return " Received type string ('" + v.Str() + "')"
	}
	// Only symbols and bigints reach here; everything else returns above.
	// Node prints `type <typeof> (<inspect>)`, lowercase with the value.
	if v.IsBigInt() {
		return " Received type bigint (" + v.Str() + "n)"
	}
	return " Received type symbol (" + v.SymbolString() + ")"
}

func (m *module) alloc(args []js.Value) (js.Value, error) {
	n, err := m.sizeArg(args, 0)
	if err != nil {
		return js.Value{}, err
	}
	if n > 0xFFFFFFFF {
		return js.Value{}, js.Throw("RangeError", fmt.Sprintf("Invalid typed array length: %d", n))
	}
	data := make([]byte, n)
	fill := arg(args, 1)
	if fill.Valid() && !fill.IsUndefined() {
		pattern, err := m.fillPattern(fill, arg(args, 2))
		if err != nil {
			return js.Value{}, err
		}
		if len(pattern) > 0 {
			for i := range data {
				data[i] = pattern[i%len(pattern)]
			}
		}
	}
	return m.newBuffer(data).obj.V(), nil
}

func (m *module) concat(args []js.Value) (js.Value, error) {
	list := arg(args, 0)
	if !list.IsArray() {
		return js.Value{}, typeErr("The \"list\" argument must be an instance of Array.%s", describeReceived(list))
	}
	n := list.Object().Length()
	// Node returns its shared empty buffer before it validates totalLength, so
	// Buffer.concat([], 3.5) is 0 long rather than a range error.
	if n == 0 {
		return m.newBuffer(nil).obj.V(), nil
	}
	parts := make([][]byte, 0, n)
	total := 0
	for i := 0; i < n; i++ {
		item := list.Object().At(i)
		var data []byte
		if b, ok := m.asBuffer(item); ok {
			data = b.b
		} else if raw, ok := item.Bytes(); ok {
			data = raw
		} else {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				fmt.Sprintf("The \"list[%d]\" argument must be an instance of Buffer or Uint8Array.", i)+describeReceived(item))
		}
		parts = append(parts, data)
		total += len(data)
	}
	want := total
	if v := arg(args, 1); v.Valid() && !v.IsUndefined() {
		if !v.IsNumber() {
			return js.Value{}, typeErr("The \"length\" argument must be of type number.%s", describeReceived(v))
		}
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
			return js.Value{}, outOfRange("length", "an integer", v)
		}
		if f < 0 || f > maxLength {
			return js.Value{}, outOfRange("length", sizeRange, v)
		}
		want = int(f)
	}
	out := make([]byte, want)
	pos := 0
	for _, p := range parts {
		if pos >= want {
			break
		}
		pos += copy(out[pos:], p)
	}
	return m.newBuffer(out).obj.V(), nil
}

func invalidFillValue(v js.Value) error {
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
		"The argument 'value' is invalid. Received "+describeReceived(v))
}

func (m *module) fillPattern(v, encArg js.Value) ([]byte, error) {
	switch {
	case v.IsNumber():
		return []byte{byteValue(v.Float())}, nil
	case v.IsString():
		enc := "utf8"
		if encArg.Valid() && !encArg.IsUndefined() {
			if !encArg.IsString() {
				return nil, typeErr("The \"encoding\" argument must be of type string.%s", describeReceived(encArg))
			}
			name, ok := normalizeEncoding(encArg.Str())
			if !ok {
				return nil, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", encArg.Str()))
			}
			enc = name
		}
		if v.Str() == "" {
			// Empty fill values fill with zero, not with nothing.
			return []byte{0}, nil
		}
		if enc == "hex" {
			s := v.Str()
			if len(s)%2 != 0 {
				return nil, invalidFillValue(v)
			}
			for i := 0; i < len(s); i++ {
				if _, ok := hexVal(s[i]); !ok {
					return nil, invalidFillValue(v)
				}
			}
		}
		return encodeString(v.Str(), enc)
	case v.IsArray():
		// Node's binding treats an array fill value like a byte list; an empty
		// array fills with zeros, matching Buffer.alloc(5, []).
		n := v.Object().Length()
		if n == 0 {
			return []byte{0}, nil
		}
		out := make([]byte, n)
		for i := 0; i < n; i++ {
			f, _ := numberOf(m.e, v.Object().At(i))
			out[i] = byteValue(f)
		}
		return out, nil
	case v.IsObject() && !v.IsNull():
		if !m.isArrayBufferView(v) {
			return nil, typeErr("The \"value\" argument must be of type string or an instance of Buffer, TypedArray, or DataView.%s", describeReceived(v))
		}
		data, ok := v.Bytes()
		if !ok {
			return nil, typeErr("The \"value\" argument must be of type string or an instance of Buffer, TypedArray, or DataView.%s", describeReceived(v))
		}
		if len(data) == 0 {
			// An empty Buffer/TypedArray cannot fill anything.
			return nil, invalidFillValue(v)
		}
		return data, nil
	}
	return nil, typeErr("The \"value\" argument must be of type string or an instance of Buffer, TypedArray, or DataView.%s", describeReceived(v))
}

func (m *module) searchNeedle(v, encArg js.Value) ([]byte, error) {
	switch {
	case v.IsNumber():
		return []byte{byte(uint64(v.Float()) & 0xff)}, nil
	case v.IsString():
		enc := "utf8"
		if encArg.Valid() && !encArg.IsUndefined() && encArg.IsString() {
			name, ok := normalizeEncoding(encArg.Str())
			if !ok {
				return nil, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", encArg.Str()))
			}
			enc = name
		}
		return encodeString(v.Str(), enc)
	case v.IsObject():
		data, ok := v.Bytes()
		if !ok {
			return nil, typeErr("The \"value\" argument must be of type string or an instance of Buffer or Uint8Array.%s", describeReceived(v))
		}
		return data, nil
	}
	return nil, typeErr("The \"value\" argument must be of type string or an instance of Buffer or Uint8Array.%s", describeReceived(v))
}

// isUint8ArrayValue mirrors Node's isUint8Array: Buffers and Uint8Array views
// are searchable values, other typed arrays are not.
func isUint8ArrayValue(v js.Value) bool {
	if !v.IsObject() || v.IsFunction() {
		return false
	}
	tag := v.Tag()
	return tag == "Uint8Array" || tag == "Buffer" || v.Object().Get("_godeBuffer").Bool()
}

// indexOfOffset is a direct port of Node's IndexOfOffset: it converts the
// user's byteOffset into a usable start position or reports that no match is
// possible (-1).
func indexOfOffset(length int, offset int64, needleLen int, forward bool) int64 {
	l := int64(length)
	if offset < 0 {
		if offset+l >= 0 {
			return l + offset
		}
		if forward || needleLen == 0 {
			return 0
		}
		return -1
	}
	if offset+int64(needleLen) <= l {
		return offset
	}
	if needleLen == 0 {
		return l
	}
	if forward {
		return -1
	}
	return l - 1
}

func toUint16(b []byte) []uint16 {
	out := make([]uint16, len(b)/2)
	for i := range out {
		out[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return out
}

func indexUint16(hay, needle []uint16, start int) int {
	if len(needle) == 0 || start < 0 || start > len(hay) {
		return -1
	}
	for i := start; i+len(needle) <= len(hay); i++ {
		match := true
		for j := range needle {
			if hay[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func lastIndexUint16(hay, needle []uint16, start int) int {
	if len(needle) == 0 {
		return -1
	}
	if start > len(hay)-len(needle) {
		start = len(hay) - len(needle)
	}
	for i := start; i >= 0; i-- {
		match := true
		for j := range needle {
			if hay[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// search implements Buffer#indexOf/includes/lastIndexOf the way Node's
// bidirectionalIndexOf does: (value, byteOffset, end, encoding), where a
// string in the end position is the encoding, offsets follow ToNumber/NaN
// rules, and a UTF-16 search aligns to whole two-byte code units.
func (m *module) search(this js.Object, args []js.Value, forward bool) (int, error) {
	b, ok := m.self(this)
	if !ok {
		return 0, receiverErr("this is not a Buffer")
	}
	hay := b.b
	val := arg(args, 0)
	offsetArg := arg(args, 1)
	endArg := arg(args, 2)
	encArg := arg(args, 3)
	// Buffer.prototype.indexOf shifts a string end argument into encoding and
	// resets the end to the buffer length.
	if endArg.Valid() && endArg.IsString() {
		encArg = endArg
		endArg = js.Value{}
	}
	hayLength := len(hay)
	searchEnd := hayLength
	if endArg.Valid() && !endArg.IsUndefined() {
		if f, ok := numberOf(m.e, endArg); ok {
			searchEnd = int(math.Trunc(f))
			if searchEnd < 0 {
				searchEnd = 0
			}
			if searchEnd > hayLength {
				searchEnd = hayLength
			}
		}
	}
	// A string byteOffset doubles as the encoding when none was supplied yet.
	if offsetArg.Valid() && offsetArg.IsString() {
		encArg = offsetArg
		offsetArg = js.Value{}
	}
	var offset int64
	if !offsetArg.Valid() || offsetArg.IsUndefined() {
		if forward {
			offset = 0
		} else {
			offset = int64(hayLength)
		}
	} else if offsetArg.IsNull() {
		// ToNumber(null) is 0; Node searches from the start, not the whole buffer.
		offset = 0
	} else if f, ok := numberOf(m.e, offsetArg); ok {
		if f > 0x7fffffff {
			f = 0x7fffffff
		}
		if f < -0x80000000 {
			f = -0x80000000
		}
		offset = int64(math.Trunc(f))
	} else if forward {
		offset = 0
	} else {
		offset = int64(hayLength)
	}

	ucs2 := false
	var needle []byte
	switch {
	case val.IsNumber():
		needle = []byte{byteValue(val.Float())}
	case val.IsString():
		enc := "utf8"
		if encArg.Valid() && !encArg.IsUndefined() {
			name, ok := normalizeEncoding(encArg.Str())
			if !ok {
				return 0, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", encArg.Str()))
			}
			enc = name
		}
		ucs2 = enc == "utf16le"
		data, err := encodeString(val.Str(), enc)
		if err != nil {
			return 0, err
		}
		needle = data
	case isUint8ArrayValue(val):
		data, ok := val.Bytes()
		if !ok {
			return 0, invalidSearchValue(val)
		}
		needle = data
		if encArg.Valid() && !encArg.IsUndefined() && encArg.IsString() {
			if name, ok := normalizeEncoding(encArg.Str()); ok && name == "utf16le" {
				ucs2 = true
			}
		}
	default:
		return 0, invalidSearchValue(val)
	}

	needleLen := len(needle)
	if ucs2 {
		hayLength &= ^1
		searchEnd &= ^1
	}
	opt := indexOfOffset(hayLength, offset, needleLen, forward)
	if needleLen == 0 {
		clamped := opt
		if int64(searchEnd) < clamped {
			clamped = int64(searchEnd)
		}
		return int(clamped), nil
	}
	if hayLength == 0 || opt <= -1 {
		return -1, nil
	}
	off := opt
	if !forward && off >= int64(searchEnd) {
		if searchEnd == 0 {
			return -1, nil
		}
		off = int64(searchEnd - 1)
	} else if forward && off >= int64(searchEnd) {
		return -1, nil
	}
	if (forward && off+int64(needleLen) > int64(searchEnd)) || int64(needleLen) > int64(searchEnd) {
		return -1, nil
	}

	if ucs2 {
		if searchEnd < 2 || needleLen < 2 {
			return -1, nil
		}
		hayUnits := toUint16(hay[:searchEnd])
		needleUnits := toUint16(needle[:needleLen&^1])
		start := int(off / 2)
		var idx int
		if forward {
			idx = indexUint16(hayUnits, needleUnits, start)
		} else {
			idx = lastIndexUint16(hayUnits, needleUnits, start)
		}
		if idx < 0 {
			return -1, nil
		}
		return idx * 2, nil
	}
	hayPart := hay[:searchEnd]
	if forward {
		idx := bytes.Index(hayPart[off:], needle)
		if idx < 0 {
			return -1, nil
		}
		return int(off) + idx, nil
	}
	hi := searchEnd
	if hi > int(off)+needleLen {
		hi = int(off) + needleLen
	}
	idx := bytes.LastIndex(hayPart[:hi], needle)
	if idx < 0 {
		return -1, nil
	}
	return idx, nil
}

func invalidSearchValue(v js.Value) error {
	return typeErr("The \"value\" argument must be one of type number or string or an instance of Buffer or Uint8Array.%s", describeReceived(v))
}

func (m *module) swap(this js.Object, size int) (js.Value, error) {
	b, ok := m.self(this)
	if !ok {
		return js.Value{}, receiverErr("this is not a Buffer")
	}
	if len(b.b)%size != 0 {
		return js.Value{}, js.ThrowCode("RangeError", "ERR_INVALID_BUFFER_SIZE",
			fmt.Sprintf("Buffer size must be a multiple of %d-bits", size*8))
	}
	out := make([]byte, len(b.b))
	for i := 0; i < len(b.b); i += size {
		for j := 0; j < size; j++ {
			out[i+j] = b.b[i+size-1-j]
		}
	}
	if err := m.setBytes(b.obj, out, 0); err != nil {
		return js.Value{}, err
	}
	return b.obj.V(), nil
}

// uint8View builds a plain Uint8Array over a window of an ArrayBuffer without
// changing the prototype of the result.
func (m *module) uint8View(ab js.Value, off, length int) (js.Object, error) {
	ctor := m.e.Global().Get("Uint8Array")
	construct := m.e.Global().Get("Reflect").Object().Get("construct")
	args := m.e.JS.Arr([]js.Value{ab, m.e.JS.Num(float64(off)), m.e.JS.Num(float64(length))})
	res, err := construct.Object().Call(js.Object{}, ctor, args.V())
	if err != nil {
		return js.Object{}, err
	}
	return res.Object(), nil
}

// setBytes copies raw bytes into any ArrayBufferView. Writing through a
// Uint8Array view over the target's backing store is required for wide views
// (a Uint16Array target must receive packed bytes, not 16-bit elements).
func (m *module) setBytes(dst js.Object, data []byte, off int) error {
	if len(data) == 0 {
		return nil
	}
	backing := dst.Get("buffer")
	if backing.Valid() && m.isAnyArrayBuffer(backing) {
		byteOffset := dst.Get("byteOffset").Int()
		view, err := m.uint8View(backing, byteOffset+off, len(data))
		if err != nil {
			return err
		}
		set := m.e.Global().Get("Uint8Array").Object().Get("prototype").Object().Get("set")
		if !set.IsFunction() {
			return fmt.Errorf("Uint8Array.prototype.set is unavailable")
		}
		_, err = set.Object().Call(view, m.e.JS.Bytes(data).V())
		return err
	}
	set := m.e.Global().Get("Uint8Array").Object().Get("prototype").Object().Get("set")
	if !set.IsFunction() {
		return fmt.Errorf("Uint8Array.prototype.set is unavailable")
	}
	_, err := set.Object().Call(dst, m.e.JS.Bytes(data).V(), m.e.JS.Num(float64(off)))
	return err
}

func normalizeEncoding(enc string) (string, bool) {
	switch strings.ToLower(enc) {
	case "utf8", "utf-8":
		return "utf8", true
	case "utf16le", "utf-16le", "ucs2", "ucs-2":
		return "utf16le", true
	case "latin1", "binary":
		return "latin1", true
	case "ascii":
		return "ascii", true
	case "hex":
		return "hex", true
	case "base64":
		return "base64", true
	case "base64url":
		return "base64url", true
	}
	return "", false
}

func encodeString(s, enc string) ([]byte, error) {
	switch enc {
	case "utf8":
		return []byte(s), nil
	case "utf16le":
		units := utf16.Encode([]rune(s))
		out := make([]byte, len(units)*2)
		for i, u := range units {
			binary.LittleEndian.PutUint16(out[i*2:], u)
		}
		return out, nil
	case "latin1", "ascii":
		units := utf16.Encode([]rune(s))
		out := make([]byte, len(units))
		mask := uint16(0xff)
		if enc == "ascii" {
			mask = 0x7f
		}
		for i, u := range units {
			out[i] = byte(u & mask)
		}
		return out, nil
	case "hex":
		out := make([]byte, 0, len(s)/2)
		for i := 0; i+1 < len(s); i += 2 {
			hi, ok1 := hexVal(s[i])
			lo, ok2 := hexVal(s[i+1])
			if !ok1 || !ok2 {
				break
			}
			out = append(out, hi<<4|lo)
		}
		return out, nil
	case "base64":
		return decodeBase64(s, false)
	case "base64url":
		return decodeBase64(s, true)
	}
	return nil, js.ThrowCode("TypeError", "ERR_UNKNOWN_ENCODING", fmt.Sprintf("Unknown encoding: %s", enc))
}

func decodeString(b []byte, enc string) string {
	switch enc {
	case "utf8":
		return string(b)
	case "utf16le":
		n := len(b) / 2
		units := make([]uint16, n)
		for i := 0; i < n; i++ {
			units[i] = binary.LittleEndian.Uint16(b[i*2:])
		}
		return string(utf16.Decode(units))
	case "latin1":
		return latin1String(b)
	case "ascii":
		out := make([]rune, len(b))
		for i, c := range b {
			out[i] = rune(c & 0x7f)
		}
		return string(out)
	case "hex":
		return hex.EncodeToString(b)
	case "base64":
		return base64.StdEncoding.EncodeToString(b)
	case "base64url":
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return string(b)
}

func decodeBase64(s string, urlSafe bool) ([]byte, error) {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r == '+', r == '-':
			if urlSafe {
				return '-'
			}
			return '+'
		case r == '/', r == '_':
			if urlSafe {
				return '_'
			}
			return '/'
		case r == '=':
			return '='
		}
		return -1
	}, s)
	if pad := len(clean) % 4; pad != 0 {
		clean += strings.Repeat("=", 4-pad)
	}
	var enc *base64.Encoding
	if urlSafe {
		enc = base64.URLEncoding
	} else {
		enc = base64.StdEncoding
	}
	out, err := enc.DecodeString(clean)
	if err != nil && len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func hexVal(c byte) (byte, bool) {
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

func latin1String(b []byte) string {
	out := make([]rune, len(b))
	for i, c := range b {
		out[i] = rune(c)
	}
	return string(out)
}

func latin1Bytes(s string) []byte {
	units := utf16.Encode([]rune(s))
	out := make([]byte, len(units))
	for i, u := range units {
		out[i] = byte(u)
	}
	return out
}

func isASCIIString(s string) bool {
	for _, r := range s {
		if r > 0x7f {
			return false
		}
	}
	return true
}

func isASCIIBytes(b []byte) bool {
	for _, c := range b {
		if c > 0x7f {
			return false
		}
	}
	return true
}

func readInteger(b []byte, signed bool) float64 {
	var x uint64
	for i := len(b) - 1; i >= 0; i-- {
		x = x<<8 | uint64(b[i])
	}
	if signed {
		shift := uint(64 - 8*len(b))
		return float64(int64(x<<shift) >> shift)
	}
	return float64(x)
}

func readIntegerBE(b []byte, signed bool) float64 {
	var x uint64
	for _, c := range b {
		x = x<<8 | uint64(c)
	}
	if signed {
		shift := uint(64 - 8*len(b))
		return float64(int64(x<<shift) >> shift)
	}
	return float64(x)
}

func writeInteger(v float64, n int, be, signed bool) []byte {
	x := uint64(int64(v))
	if !signed {
		x = uint64(v)
	}
	mask := uint64(1)<<(8*n) - 1
	if n >= 8 {
		mask = ^uint64(0)
	}
	x &= mask
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		if be {
			out[i] = byte(x >> (8 * (n - 1 - i)))
		} else {
			out[i] = byte(x >> (8 * i))
		}
	}
	return out
}

func clampRange(n, start, end int) (int, int) {
	if start < 0 {
		start = maxInt(n+start, 0)
	}
	if end < 0 {
		end = maxInt(n+end, 0)
	}
	if start > n {
		start = n
	}
	if end > n {
		end = n
	}
	if end < start {
		end = start
	}
	return start, end
}

func normalizeOffset(n, off int) int {
	if off < 0 {
		off = n + off
	}
	if off < 0 {
		off = 0
	}
	if off > n {
		off = n
	}
	return off
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

func optInt(v js.Value, def int) int {
	if !v.Valid() || v.IsUndefined() {
		return def
	}
	return v.Int()
}

// typeErr builds Node's ERR_INVALID_ARG_TYPE: a TypeError carrying the code
// so String(err) renders as `TypeError [ERR_INVALID_ARG_TYPE]: ...`.
func typeErr(format string, a ...any) error {
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE", fmt.Sprintf(format, a...))
}

// rangeErr builds Node's ERR_OUT_OF_RANGE.
func rangeErr(format string, a ...any) error {
	return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE", fmt.Sprintf(format, a...))
}

// receiverErr is a plain TypeError for a bad `this`. Node reports these from
// the TypedArray internals with no error code, so they carry none either.
func receiverErr(format string, a ...any) error {
	return js.Throw("TypeError", fmt.Sprintf(format, a...))
}

// outOfRange renders Node's ERR_OUT_OF_RANGE message for a named value:
// `The value of "x" is out of range. It must be <range>. Received <value>`.
func outOfRange(name, rangeSpec string, input js.Value) error {
	return rangeErr("The value of %q is out of range. It must be %s. Received %s",
		name, rangeSpec, rangeValue(input))
}

// rangeValue renders the offending value the way Node's ERR_OUT_OF_RANGE does:
// integers past 2**32 get digit separators, everything else is inspected.
func rangeValue(v js.Value) string {
	if v.IsBigInt() {
		// Node prints a BigInt with grouped digits and its `n` suffix.
		return addSeparators(v.Str()) + "n"
	}
	if v.IsNumber() {
		f := v.Float()
		switch {
		case math.IsNaN(f):
			return "NaN"
		case math.IsInf(f, 1):
			return "Infinity"
		case math.IsInf(f, -1):
			return "-Infinity"
		}
		if f == math.Trunc(f) && math.Abs(f) > 4294967296 {
			return addSeparators(strconv.FormatFloat(f, 'f', -1, 64))
		}
		if f == math.Trunc(f) && math.Abs(f) < 1e21 {
			return strconv.FormatFloat(f, 'f', -1, 64)
		}
		return strconv.FormatFloat(f, 'g', -1, 64)
	}
	if v.IsString() {
		return "'" + v.Str() + "'"
	}
	if v.IsBool() {
		return strconv.FormatBool(v.Bool())
	}
	return strings.TrimPrefix(describeReceived(v), " Received type ")
}

// addSeparators groups digits in threes, which is how Node prints integers
// larger than 2**32 inside an error message.
func addSeparators(val string) string {
	start := 0
	if strings.HasPrefix(val, "-") {
		start = 1
	}
	// Direct port of addNumericalSeparator in lib/internal/errors.js: walk
	// three digits at a time from the right, keeping at least one leading digit.
	i := len(val)
	res := ""
	for i >= start+4 {
		res = "_" + val[i-3:i] + res
		i -= 3
	}
	return val[:i] + res
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"buffer.Buffer",
		"buffer.Buffer.alloc",
		"buffer.Buffer.allocUnsafe",
		"buffer.Buffer.allocUnsafeSlow",
		"buffer.Buffer.byteLength",
		"buffer.Buffer.compare",
		"buffer.Buffer.concat",
		"buffer.Buffer.copyBytesFrom",
		"buffer.Buffer.from",
		"buffer.Buffer.isBuffer",
		"buffer.Buffer.isEncoding",
		"buffer.Buffer.poolSize",
		"buffer.Buffer#compare",
		"buffer.Buffer#copy",
		"buffer.Buffer#equals",
		"buffer.Buffer#fill",
		"buffer.Buffer#includes",
		"buffer.Buffer#indexOf",
		"buffer.Buffer#lastIndexOf",
		"buffer.Blob",
		"buffer.Blob#arrayBuffer",
		"buffer.Blob#bytes",
		"buffer.Blob#slice",
		"buffer.Blob#stream",
		"buffer.Blob#text",
		"buffer.Blob#textStream",
		"buffer.Blob.size",
		"buffer.Blob.type",
		"buffer.File",
		"buffer.File.lastModified",
		"buffer.File.name",
		"buffer.Buffer#entries",
		"buffer.Buffer#keys",
		"buffer.Buffer#values",
		"buffer.Buffer#readBigInt64BE",
		"buffer.Buffer#readBigInt64LE",
		"buffer.Buffer#readBigUInt64BE",
		"buffer.Buffer#readBigUInt64LE",
		"buffer.Buffer#writeBigInt64BE",
		"buffer.Buffer#writeBigInt64LE",
		"buffer.Buffer#writeBigUInt64BE",
		"buffer.Buffer#writeBigUInt64LE",
		"buffer.Buffer.buffer",
		"buffer.Buffer.byteOffset",
		"buffer.Buffer.index",
		"buffer.Buffer.length",
		"buffer.Buffer.parent",
		"buffer.Buffer#readDoubleBE",
		"buffer.Buffer#readDoubleLE",
		"buffer.Buffer#readFloatBE",
		"buffer.Buffer#readFloatLE",
		"buffer.Buffer#readInt16BE",
		"buffer.Buffer#readInt16LE",
		"buffer.Buffer#readInt32BE",
		"buffer.Buffer#readInt32LE",
		"buffer.Buffer#readInt8",
		"buffer.Buffer#readIntBE",
		"buffer.Buffer#readIntLE",
		"buffer.Buffer#readUInt16BE",
		"buffer.Buffer#readUInt16LE",
		"buffer.Buffer#readUInt32BE",
		"buffer.Buffer#readUInt32LE",
		"buffer.Buffer#readUInt8",
		"buffer.Buffer#readUIntBE",
		"buffer.Buffer#readUIntLE",
		"buffer.Buffer#slice",
		"buffer.Buffer#subarray",
		"buffer.Buffer#swap16",
		"buffer.Buffer#swap32",
		"buffer.Buffer#swap64",
		"buffer.Buffer#toJSON",
		"buffer.Buffer#toString",
		"buffer.Buffer#write",
		"buffer.Buffer#writeDoubleBE",
		"buffer.Buffer#writeDoubleLE",
		"buffer.Buffer#writeFloatBE",
		"buffer.Buffer#writeFloatLE",
		"buffer.Buffer#writeInt16BE",
		"buffer.Buffer#writeInt16LE",
		"buffer.Buffer#writeInt32BE",
		"buffer.Buffer#writeInt32LE",
		"buffer.Buffer#writeInt8",
		"buffer.Buffer#writeIntBE",
		"buffer.Buffer#writeIntLE",
		"buffer.Buffer#writeUInt16BE",
		"buffer.Buffer#writeUInt16LE",
		"buffer.Buffer#writeUInt32BE",
		"buffer.Buffer#writeUInt32LE",
		"buffer.Buffer#writeUInt8",
		"buffer.Buffer#writeUIntBE",
		"buffer.Buffer#writeUIntLE",
		"buffer.INSPECT_MAX_BYTES",
		"buffer.MAX_LENGTH",
		"buffer.MAX_STRING_LENGTH",
		"buffer.SlowBuffer",
		"buffer.atob",
		"buffer.btoa",
		"buffer.constants",
		"buffer.isAscii",
		"buffer.isUtf8",
		"buffer.kMaxLength",
		"buffer.kStringMaxLength",
		"buffer.resolveObjectURL",
		"buffer.transcode",
	}
}
