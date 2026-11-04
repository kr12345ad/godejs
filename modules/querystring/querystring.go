// Package querystring implements Node's querystring module on top of
// JavaScriptCore. Parsing and serialization live in Go; only the returned
// values cross into JavaScript.
package querystring

import (
	"math"
	"strings"
	"unicode/utf16"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("querystring", Install) }

// Install builds the querystring exports object.
func Install(e *env.Env) js.Object {
	exports := e.JS.Obj()

	parse := func(_ js.Object, args []js.Value) (js.Value, error) {
		return parseQuery(e, exports, args)
	}
	stringify := func(_ js.Object, args []js.Value) (js.Value, error) {
		return stringifyQuery(e, args)
	}
	unescape := func(_ js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Str(percentDecode(argStr(args, 0), false)), nil
	}

	exports.SetFunc("parse", parse)
	exports.SetFunc("stringify", stringify)
	exports.SetFunc("unescape", unescape)
	exports.SetFunc("unescapeBuffer", func(_ js.Object, args []js.Value) (js.Value, error) {
		return unescapeBufferValue(e, args)
	})
	exports.Set("decode", exports.Get("parse"))
	exports.Set("encode", exports.Get("stringify"))

	// querystring.escape walks UTF-16 code units exactly like Node's encodeStr:
	// a high surrogate consumes the following code unit (even when it is not a
	// low surrogate), and a high surrogate at the end raises ERR_INVALID_URI.
	// Go only ever sees U+FFFD for lone surrogates, so this lives in JavaScript.
	if init, err := e.JS.Eval(`(function (exports) {
  'use strict';
  var hexTable = new Array(256);
  for (var i = 0; i < 256; i++) {
    hexTable[i] = '%' + (i < 16 ? '0' : '') + i.toString(16).toUpperCase();
  }
  function noEscape(c) {
    if (c >= 0x30 && c <= 0x39) return true;
    if (c >= 0x41 && c <= 0x5A) return true;
    if (c >= 0x61 && c <= 0x7A) return true;
    switch (c) {
      case 0x21: case 0x27: case 0x28: case 0x29: case 0x2A: case 0x2D:
      case 0x2E: case 0x5F: case 0x7E:
        return true;
    }
    return false;
  }
  exports.escape = function escape(str) {
    if (typeof str !== 'string') {
      if (typeof str === 'object') str = String(str);
      else str += '';
    }
    var len = str.length;
    if (len === 0) return '';
    var out = '';
    var lastPos = 0;
    var i = 0;
    outer:
    for (; i < len; i++) {
      var c = str.charCodeAt(i);
      while (c < 0x80) {
        if (!noEscape(c)) {
          if (lastPos < i) out += str.slice(lastPos, i);
          lastPos = i + 1;
          out += hexTable[c];
        }
        if (++i === len) break outer;
        c = str.charCodeAt(i);
      }
      if (lastPos < i) out += str.slice(lastPos, i);
      if (c < 0x800) {
        lastPos = i + 1;
        out += hexTable[0xC0 | (c >> 6)] + hexTable[0x80 | (c & 0x3F)];
        continue;
      }
      if (c < 0xD800 || c >= 0xE000) {
        lastPos = i + 1;
        out += hexTable[0xE0 | (c >> 12)] + hexTable[0x80 | ((c >> 6) & 0x3F)] +
               hexTable[0x80 | (c & 0x3F)];
        continue;
      }
      ++i;
      if (i >= len) {
        var err = new URIError('URI malformed');
        err.code = 'ERR_INVALID_URI';
        throw err;
      }
      var c2 = str.charCodeAt(i) & 0x3FF;
      lastPos = i + 1;
      c = 0x10000 + (((c & 0x3FF) << 10) | c2);
      out += hexTable[0xF0 | (c >> 18)] + hexTable[0x80 | ((c >> 12) & 0x3F)] +
             hexTable[0x80 | ((c >> 6) & 0x3F)] + hexTable[0x80 | (c & 0x3F)];
    }
    if (lastPos === 0) return str;
    if (lastPos < len) out += str.slice(lastPos);
    return out;
  };

  // Node's escape runs over every key and string value before formatting, so a
  // lone surrogate anywhere in the object raises ERR_INVALID_URI. The Go
  // stringify cannot see lone surrogates (they arrive as U+FFFD), so validate
  // them in JavaScript first.
  var godeStringify = exports.stringify;
  function checkSurrogates(v) {
    if (typeof v === 'string') { exports.escape(v); return; }
    if (Array.isArray(v)) {
      for (var i = 0; i < v.length; i++) checkSurrogates(v[i]);
    }
  }
  exports.stringify = function stringify(obj, sep, eq, options) {
    var custom = options && typeof options.encodeURIComponent === 'function';
    if (!custom && obj !== null && typeof obj === 'object') {
      var keys = Object.keys(obj);
      for (var i = 0; i < keys.length; i++) {
        checkSurrogates(keys[i]);
        checkSurrogates(obj[keys[i]]);
      }
    }
    return godeStringify.apply(this, arguments);
  };
})`, "[gode:querystring]"); err == nil && init.IsObject() {
		init.Object().Call(js.Object{}, exports.V())
	} else if err != nil {
		panic(err)
	}
	// Node exposes these as the same function objects, not wrappers.
	exports.Set("decode", exports.Get("parse"))
	exports.Set("encode", exports.Get("stringify"))
	return exports
}

func argVal(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

func argStr(args []js.Value, i int) string {
	v := argVal(args, i)
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return ""
	}
	return v.Str()
}

// stringArg returns the first argument only when it is a JavaScript string,
// matching Node: querystring.parse() of a non-string returns an empty object.
func stringArg(args []js.Value, i int) (string, bool) {
	v := argVal(args, i)
	if !v.Valid() || !v.IsString() {
		return "", false
	}
	return v.Str(), true
}

// optionString mirrors Node's separator handling: a falsy value keeps the
// default, while any other value is coerced with String(). An empty array is
// truthy and coerces to "", which is why querystring.parse('a', null, [])
// yields {”: 'a'} rather than { a: ” }.
func optionString(args []js.Value, i int, def string) string {
	v := argVal(args, i)
	if !v.Valid() || !v.Bool() {
		return def
	}
	return v.Str()
}

// unescapeBufferValue is Node's legacy percent-decoder: it writes raw bytes
// into a Buffer, leaves malformed escapes literal, and only turns '+' into a
// space when decodeSpaces is true.
func unescapeBufferValue(e *env.Env, args []js.Value) (js.Value, error) {
	s := argStr(args, 0)
	decodeSpaces := len(args) > 1 && args[1].Bool()
	units := utf16.Encode([]rune(s))
	out := make([]byte, 0, len(units))
	for i := 0; i < len(units); {
		c := units[i]
		if c == '+' && decodeSpaces {
			out = append(out, ' ')
			i++
			continue
		}
		if c == '%' && i+2 < len(units) {
			hi, lo := unhexUnit(units[i+1]), unhexUnit(units[i+2])
			if hi >= 0 && lo >= 0 {
				out = append(out, byte(hi<<4|lo))
				i += 3
				continue
			}
		}
		out = append(out, byte(c))
		i++
	}
	ctor := e.Global().Get("Buffer")
	if ctor.Valid() && ctor.IsFunction() {
		if from := ctor.Object().Get("from"); from.Valid() && from.IsFunction() {
			if v, err := from.Object().Call(ctor.Object(), e.JS.Bytes(out).V()); err == nil {
				return v, nil
			}
		}
	}
	return e.JS.Bytes(out).V(), nil
}

func unhexUnit(u uint16) int {
	switch {
	case u >= '0' && u <= '9':
		return int(u - '0')
	case u >= 'a' && u <= 'f':
		return int(u-'a') + 10
	case u >= 'A' && u <= 'F':
		return int(u-'A') + 10
	}
	return -1
}

func parseQuery(e *env.Env, exports js.Object, args []js.Value) (js.Value, error) {
	empty, err := e.JS.Eval("Object.create(null)", "querystring.js")
	if err != nil {
		return js.Value{}, err
	}
	obj := empty.Object()

	str, ok := stringArg(args, 0)
	if !ok || str == "" {
		return obj.V(), nil
	}
	sep := optionString(args, 1, "&")
	eq := optionString(args, 2, "=")

	maxKeys := 1000
	var decode js.Value
	if o := argVal(args, 3); o.Valid() && o.IsObject() {
		opts := o.Object()
		if v := opts.Get("maxKeys"); v.Valid() && v.IsNumber() {
			f := v.Float()
			// Non-finite numbers mean "no limit"; only strings and other
			// non-numbers fall back to Node's default of 1000.
			if math.IsNaN(f) || math.IsInf(f, 0) {
				maxKeys = 0
			} else {
				maxKeys = int(f)
			}
		}
		if v := opts.Get("decodeURIComponent"); v.Valid() && v.IsFunction() {
			decode = v
		}
	}
	if !decode.Valid() {
		// Node reads the module's current unescape on every parse, so assigning
		// querystring.unescape = fn changes how subsequent parses decode keys.
		if u := exports.Get("unescape"); u.Valid() && u.IsFunction() {
			decode = u
		}
	}

	keys := 0
	// Node counts every separator-delimited pair against maxKeys, including
	// empty ones, so a leading '&' consumes a slot without producing a key.
	pairsLeft := maxKeys
	for _, pair := range strings.Split(str, sep) {
		if maxKeys > 0 {
			if pairsLeft <= 0 {
				break
			}
			pairsLeft--
		}
		if pair == "" {
			continue
		}
		key, value := pair, ""
		if i := strings.Index(pair, eq); i >= 0 {
			key, value = pair[:i], pair[i+len(eq):]
		}
		decodedKey, err := decodePart(e, key, decode)
		if err != nil {
			return js.Value{}, err
		}
		decodedValue, err := decodePart(e, value, decode)
		if err != nil {
			return js.Value{}, err
		}
		if !obj.Has(decodedKey) {
			obj.Set(decodedKey, e.JS.Str(decodedValue))
			keys++
			continue
		}
		current := obj.Get(decodedKey)
		if current.IsArray() {
			arr := current.Object()
			arr.SetAt(arr.Length(), e.JS.Str(decodedValue))
			continue
		}
		obj.Set(decodedKey, e.JS.Arr([]js.Value{current, e.JS.Str(decodedValue)}).V())
	}
	return obj.V(), nil
}

func decodePart(e *env.Env, s string, decode js.Value) (string, error) {
	if decode.Valid() && decode.IsFunction() {
		// Node hands a caller-supplied decoder %20 where the query had '+', so
		// decodeURIComponent still turns a plus into a space.
		s = strings.ReplaceAll(s, "+", "%20")
		out, err := decode.Object().Call(js.Object{}, e.JS.Str(s))
		if err != nil {
			// Node swallows errors from a caller-supplied decoder and keeps the
			// raw segment, so one bad escape cannot abort the whole parse.
			return s, nil
		}
		return out.Str(), nil
	}
	// The built-in decoder treats '+' as a space; a caller-supplied
	// decodeURIComponent receives the raw segment, like Node.
	return percentDecode(s, true), nil
}

func stringifyQuery(e *env.Env, args []js.Value) (js.Value, error) {
	objVal := argVal(args, 0)
	if !objVal.Valid() || !objVal.IsObject() || objVal.IsNull() {
		return e.JS.Str(""), nil
	}
	sep := optionString(args, 1, "&")
	eq := optionString(args, 2, "=")

	var encode js.Value
	if o := argVal(args, 3); o.Valid() && o.IsObject() {
		if v := o.Object().Get("encodeURIComponent"); v.Valid() && v.IsFunction() {
			encode = v
		}
	}

	obj := objVal.Object()
	var b strings.Builder
	first := true
	add := func(key, value string) error {
		k, err := encodePart(e, key, encode)
		if err != nil {
			return err
		}
		v, err := encodePart(e, value, encode)
		if err != nil {
			return err
		}
		if !first {
			b.WriteString(sep)
		}
		first = false
		b.WriteString(k)
		b.WriteString(eq)
		b.WriteString(v)
		return nil
	}

	for _, key := range obj.OwnEnumerableKeys() {
		v := obj.Get(key)
		// Node emits every own key: undefined, null, functions and the rest
		// format as an empty value instead of being dropped.
		if !v.Valid() {
			continue
		}
		if v.IsArray() {
			arr := v.Object()
			n := arr.Length()
			for i := 0; i < n; i++ {
				item := arr.At(i)
				if !item.Valid() {
					continue
				}
				if err := add(key, stringifyPrimitive(e, item)); err != nil {
					return js.Value{}, err
				}
			}
			continue
		}
		if err := add(key, stringifyPrimitive(e, v)); err != nil {
			return js.Value{}, err
		}
	}
	return e.JS.Str(b.String()), nil
}

// stringifyPrimitive follows Node's querystring.stringify rule: only strings,
// finite numbers, bigints and booleans contribute text; every other value
// (objects, symbols, NaN/Infinity) serializes as an empty string.
func stringifyPrimitive(e *env.Env, v js.Value) string {
	switch {
	case v.IsString():
		return v.Str()
	case v.IsBool():
		if v.Bool() {
			return "true"
		}
		return "false"
	case v.IsNumber():
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return ""
		}
		return jsNumberText(e, v)
	}
	if v.Ctx() == nil || v.IsObject() {
		return ""
	}
	// BigInt is the one remaining primitive that serializes; String(10n) is
	// `10`, without the trailing n.
	if s := jsNumberText(e, v); isBigIntText(s) {
		return s
	}
	return ""
}

func isBigIntText(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '-' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func jsNumberText(e *env.Env, v js.Value) string {
	s := e.JS.Global().Get("String")
	if s.IsFunction() {
		if r, err := s.Object().Call(js.Object{}, v); err == nil {
			return r.Str()
		}
	}
	return v.Str()
}

func encodePart(e *env.Env, s string, encode js.Value) (string, error) {
	if encode.Valid() && encode.IsFunction() {
		out, err := encode.Object().Call(js.Object{}, e.JS.Str(s))
		if err != nil {
			return "", err
		}
		return out.Str(), nil
	}
	return percentEncode(s), nil
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func hexVal(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}

// percentDecode decodes %XX escapes as UTF-8 bytes. Malformed escapes are kept
// verbatim, matching Node's fallback decoder. When plusIsSpace is set, '+' is
// decoded as a space (the querystring.parse rule).
func percentDecode(s string, plusIsSpace bool) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '+' && plusIsSpace {
			b.WriteByte(' ')
			continue
		}
		if c == '%' && i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
			b.WriteByte(hexVal(s[i+1])<<4 | hexVal(s[i+2]))
			i += 2
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// percentEncode matches Node's querystring.escape / encodeURIComponent: UTF-8
// bytes are escaped except for the unreserved set.
func percentEncode(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUnreserved(c) {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&0x0F])
	}
	return b.String()
}

func isUnreserved(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	case c == '-', c == '_', c == '.', c == '!', c == '~', c == '*', c == '\'', c == '(', c == ')':
		return true
	}
	return false
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"querystring.decode",
		"querystring.encode",
		"querystring.escape",
		"querystring.parse",
		"querystring.stringify",
		"querystring.unescape",
	}
}
