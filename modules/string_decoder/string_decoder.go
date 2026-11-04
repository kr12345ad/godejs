package string_decoder

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("string_decoder", Install) }

// The constructor, prototype wiring and Node's ERR_UNKNOWN_ENCODING shape are
// built in JavaScript; every byte-level decoder lives in Go below.
const wrapperSource = `(function (supported, create) {
  'use strict';
  function StringDecoder(encoding) {
    // Node's StringDecoder is explicitly usable without new: the tests do
    // StringDecoder.call(obj) and inspect the initialized properties. A
    // plain call still needs a receiver, so make one.
    if (this === undefined || this === null) return new StringDecoder(encoding);
    if (encoding === undefined) encoding = 'utf8';
    var enc = String(encoding).toLowerCase();
    if (!supported(enc)) {
      var err = new TypeError('Unknown encoding: ' + encoding);
      err.code = 'ERR_UNKNOWN_ENCODING';
      throw err;
    }
    create.call(this, enc);
  }
  function invoke(self, name, args) {
    if (!Object.prototype.hasOwnProperty.call(self, name)) {
      var err = new TypeError('Value of "this" must be of type StringDecoder');
      err.code = 'ERR_INVALID_THIS';
      throw err;
    }
    return self[name].apply(self, args);
  }
  StringDecoder.prototype.write = function write(buf) { return invoke(this, '_write', [buf]); };
  StringDecoder.prototype.end = function end(buf) { return invoke(this, '_end', [buf]); };
  StringDecoder.prototype.text = function text(buf, offset) { return invoke(this, '_text', [buf, offset]); };
  return StringDecoder;
})`

// Install builds the string_decoder module's exports object.
func Install(e *env.Env) js.Object {
	h := e.JS.Obj()
	h.SetFunc("supported", func(_ js.Object, args []js.Value) (js.Value, error) {
		enc := ""
		if len(args) > 0 {
			enc = args[0].Str()
		}
		_, ok := normalize(enc)
		return e.JS.Bool(ok), nil
	})
	h.SetFunc("create", func(this js.Object, args []js.Value) (js.Value, error) {
		enc := "utf8"
		if len(args) > 0 {
			enc = args[0].Str()
		}
		canonical, ok := normalize(enc)
		if !ok {
			canonical = "utf8"
		}
		c := this.Ctx()
		this.SetHidden("_encoding", c.Str(canonical))
		this.Set("encoding", c.Str(canonical))
		setState(this, canonical, nil, 0, 0)
		this.SetFuncHidden("_write", func(this js.Object, args []js.Value) (js.Value, error) {
			return hostWrite(this, args)
		})
		this.SetFuncHidden("_end", func(this js.Object, args []js.Value) (js.Value, error) {
			return hostEnd(this, args)
		})
		this.SetFuncHidden("_text", func(this js.Object, args []js.Value) (js.Value, error) {
			return hostWrite(this, args)
		})
		return js.Value{}, nil
	})

	fn, err := e.JS.Eval(wrapperSource, "string_decoder.js")
	if err != nil {
		panic(err)
	}
	ctor, err := fn.Object().Call(js.Object{}, h.Get("supported"), h.Get("create"))
	if err != nil {
		panic(err)
	}
	exports := e.JS.Obj()
	exports.Set("StringDecoder", ctor)
	return exports
}

func hostWrite(this js.Object, args []js.Value) (js.Value, error) {
	c := this.Ctx()
	if c == nil || !this.Has("_encoding") {
		return js.Value{}, &js.Error{
			Name:    "TypeError",
			Code:    "ERR_INVALID_THIS",
			Message: `Value of "this" must be of type StringDecoder`,
		}
	}
	enc := encodingOf(this)
	in, err := bufferArg(args, false)
	if err != nil {
		return js.Value{}, err
	}
	// Node's internal text(buf, offset) starts decoding at a byte offset.
	if len(args) > 1 && args[1].IsNumber() {
		off := int(args[1].Int())
		if off >= len(in) {
			in = nil
		} else if off > 0 {
			in = in[off:]
		}
	}
	pending := pendingOf(this)
	// An ASCII-only chunk cannot shrink, so reject a result longer than V8's
	// maximum string length before allocating it.
	if (enc == "utf8" || enc == "latin1") && len(in)+len(pending) > kStringMaxLength {
		if allASCIIBytes(in) && allASCIIBytes(pending) {
			return js.Value{}, errStringTooLong()
		}
	}
	if enc == "utf16le" {
		units, next, need, total := utf16Units(in, pending)
		if len(units) > kStringMaxLength {
			return js.Value{}, errStringTooLong()
		}
		setState(this, enc, next, need, total)
		return unitsValue(c, units), nil
	}
	out, next, need, total := writeText(enc, in, pending)
	if len(out) > kStringMaxLength {
		return js.Value{}, errStringTooLong()
	}
	setState(this, enc, next, need, total)
	return c.Str(out), nil
}

const kStringMaxLength = 0x1fffffe8

func errStringTooLong() error {
	return &js.Error{
		Name:    "Error",
		Code:    "ERR_STRING_TOO_LONG",
		Message: "Cannot create a string longer than 0x1fffffe8 characters",
	}
}

func allASCIIBytes(b []byte) bool {
	for _, c := range b {
		if c >= 0x80 {
			return false
		}
	}
	return true
}

func hostEnd(this js.Object, args []js.Value) (js.Value, error) {
	c := this.Ctx()
	if c == nil {
		return js.Value{}, &js.Error{Name: "TypeError", Message: "Illegal invocation"}
	}
	enc := encodingOf(this)
	pending := pendingOf(this)
	if enc == "utf16le" {
		var units []uint16
		if len(args) > 0 && !args[0].IsUndefined() {
			in, err := bufferArg(args, false)
			if err != nil {
				return js.Value{}, err
			}
			var rest []byte
			units, rest, _, _ = utf16Units(in, pending)
			pending = rest
		}
		// At end of input, decode whatever complete units are still held; an
		// odd trailing byte is dropped. Lone surrogates are preserved.
		for i := 0; i+1 < len(pending); i += 2 {
			units = append(units, uint16(pending[i])|uint16(pending[i+1])<<8)
		}
		setState(this, enc, nil, 0, 0)
		this.Set("lastChar", c.Undefined())
		return unitsValue(c, units), nil
	}
	out := ""
	if len(args) > 0 && !args[0].IsUndefined() {
		in, err := bufferArg(args, false)
		if err != nil {
			return js.Value{}, err
		}
		var rest []byte
		out, rest, _, _ = writeText(enc, in, pending)
		pending = rest
	}
	out += finalText(enc, pending)
	setState(this, enc, nil, 0, 0)
	this.Set("lastChar", c.Undefined())
	return c.Str(out), nil
}

func bufferArg(args []js.Value, optional bool) ([]byte, error) {
	if len(args) == 0 || args[0].IsUndefined() || args[0].IsNull() {
		if optional {
			return nil, nil
		}
		received := "undefined"
		if len(args) > 0 && args[0].IsNull() {
			received = "null"
		}
		return nil, &js.Error{
			Name:    "TypeError",
			Code:    "ERR_INVALID_ARG_TYPE",
			Message: `The "buf" argument must be an instance of Buffer, TypedArray, or DataView. Received ` + received,
		}
	}
	if len(args) > 0 {
		if b, ok := args[0].Bytes(); ok {
			return b, nil
		}
	}
	received := "undefined"
	if len(args) > 0 {
		received = describeArg(args[0])
	}
	return nil, &js.Error{
		Name:    "TypeError",
		Code:    "ERR_INVALID_ARG_TYPE",
		Message: `The "buf" argument must be an instance of Buffer, TypedArray, or DataView. Received ` + received,
	}
}

func describeArg(v js.Value) string {
	switch {
	case !v.Valid(), v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		return "type boolean (" + map[bool]string{true: "true", false: "false"}[v.Bool()] + ")"
	case v.IsNumber():
		return "type number (" + strconv.FormatFloat(v.Float(), 'f', -1, 64) + ")"
	case v.IsString():
		return "type string (" + strconv.Quote(v.Str()) + ")"
	}
	return "an instance of " + v.Tag()
}

func encodingOf(o js.Object) string {
	enc := o.Get("_encoding").Str()
	if enc == "" {
		return "utf8"
	}
	return enc
}

func pendingOf(o js.Object) []byte {
	b, ok := o.Get("_pending").Bytes()
	if !ok {
		return nil
	}
	return b
}

func setState(o js.Object, enc string, pending []byte, need, total int) {
	c := o.Ctx()
	if c == nil {
		return
	}
	o.SetHidden("_pending", c.Bytes(pending).V())
	o.Set("lastNeed", c.Num(float64(need)))
	o.Set("lastTotal", c.Num(float64(total)))
	last := make([]byte, 4)
	copy(last, pending)
	o.Set("lastChar", byteBuffer(c, last))
}

// byteBuffer prefers a real Buffer, which is what Node exposes as lastChar;
// it falls back to a Uint8Array when the buffer module is not installed. It
// goes through Buffer.from: calling the Buffer constructor without `new` is
// the deprecated form and emits a spurious DEP0005 warning.
func byteBuffer(c *js.Context, b []byte) js.Value {
	if ctor := c.Global().Get("Buffer"); ctor.Valid() && ctor.IsFunction() {
		if from := ctor.Object().Get("from"); from.Valid() && from.IsFunction() {
			if v, err := from.Object().Call(ctor.Object(), c.Bytes(b).V()); err == nil && v.Valid() {
				return v
			}
		}
	}
	return c.Bytes(b).V()
}

func normalize(enc string) (string, bool) {
	switch strings.ToLower(enc) {
	case "utf8", "utf-8":
		return "utf8", true
	case "utf16le", "utf-16le", "ucs2", "ucs-2":
		return "utf16le", true
	case "latin1", "binary":
		return "latin1", true
	case "base64":
		return "base64", true
	case "base64url":
		return "base64url", true
	case "ascii":
		return "ascii", true
	case "hex":
		return "hex", true
	}
	return "", false
}

func writeText(enc string, in, pending []byte) (string, []byte, int, int) {
	switch enc {
	case "utf8":
		return utf8Write(in, pending)

	case "base64":
		return base64Write(in, pending, base64.StdEncoding)
	case "base64url":
		return base64Write(in, pending, base64.RawURLEncoding)
	case "hex":
		return hex.EncodeToString(join(pending, in)), nil, 0, 0
	case "latin1":
		return latin1String(join(pending, in)), nil, 0, 0
	case "ascii":
		return asciiString(join(pending, in)), nil, 0, 0
	}
	return "", nil, 0, 0
}

func finalText(enc string, pending []byte) string {
	switch enc {
	case "utf8":
		if len(pending) == 0 {
			return ""
		}
		// A held UTF-8 prefix is one incomplete sequence; Node reports a
		// single replacement character for it.
		return "\uFFFD"
	case "utf16le":
		if len(pending) < 2 {
			return ""
		}
		return unitsString([]uint16{uint16(pending[0]) | uint16(pending[1])<<8})
	case "base64":
		return base64.StdEncoding.EncodeToString(pending)
	case "base64url":
		return base64.RawURLEncoding.EncodeToString(pending)
	}
	return ""
}

// utf8Write decodes bytes the way JavaScriptCore does: an invalid sequence is
// reported as one replacement character for its maximal subpart, but a
// structurally valid encoding of a surrogate is reported one byte at a time.
func utf8Write(in, pending []byte) (string, []byte, int, int) {
	buf := join(pending, in)
	var sb strings.Builder
	i := 0
	for i < len(buf) {
		c := buf[i]
		if c < 0x80 {
			sb.WriteByte(c)
			i++
			continue
		}
		// Buffering uses the structural shape even for out-of-range leads such
		// as F6, which Node waits on before deciding how many replacements to
		// emit for the broken prefix.
		n := utf8StructLen(c)
		if n == 0 {
			sb.WriteRune(utf8.RuneError)
			i++
			continue
		}
		// Count structurally valid continuation bytes. The decoder buffers such
		// a prefix even when the lead byte itself is out of range (Node holds
		// F6 9B), then decides how many replacement characters to emit when the
		// sequence breaks.
		k := 1
		for k < n && i+k < len(buf) && buf[i+k]&0xC0 == 0x80 {
			k++
		}
		if i+n > len(buf) {
			// The sequence may still be completed by a later chunk.
			if k == len(buf)-i {
				rest := append([]byte{}, buf[i:]...)
				return sb.String(), rest, n - len(rest), n
			}
			// Broken prefix: one replacement for the strictly valid part, so
			// every invalid byte is reported separately.
			v := utf8ValidPrefix(c, buf[i:i+k])
			if v == 0 {
				v = 1
			}
			sb.WriteRune(utf8.RuneError)
			i += v
			continue
		}
		if k < n {
			v := utf8ValidPrefix(c, buf[i:i+k])
			if v == 0 {
				v = 1
			}
			sb.WriteRune(utf8.RuneError)
			i += v
			continue
		}
		r, size := utf8.DecodeRune(buf[i : i+n])
		if r == utf8.RuneError && size == 1 {
			sb.WriteRune(utf8.RuneError)
			i++
			continue
		}
		sb.WriteRune(r)
		i += size
	}
	return sb.String(), nil, 0, 0
}

func utf8StructLen(b byte) int {
	switch {
	case b < 0x80:
		return 1
	case b&0xE0 == 0xC0:
		return 2
	case b&0xF0 == 0xE0:
		return 3
	case b&0xF8 == 0xF0:
		return 4
	}
	return 0
}

func utf8SeqLen(b byte) int {
	switch {
	case b < 0x80:
		return 1
	case b >= 0xC2 && b <= 0xDF:
		return 2
	case b >= 0xE0 && b <= 0xEF:
		return 3
	case b >= 0xF0 && b <= 0xF4:
		return 4
	}
	return 0
}

// utf8ContOK reports whether b may appear at position pos (1-based) of a
// sequence with the given lead byte. The UTF-8 second-byte ranges exclude
// overlong forms, surrogates and code points above U+10FFFF.
// utf8ValidPrefix counts the leading bytes of p that form a strictly valid
// prefix of a sequence led by lead. Zero means the lead itself is invalid.
func utf8ValidPrefix(lead byte, p []byte) int {
	n := utf8SeqLen(lead)
	if n == 0 {
		return 0
	}
	v := 1
	for v < n && v < len(p) && utf8ContOK(lead, v, p[v]) {
		v++
	}
	return v
}

func utf8ContOK(lead byte, pos int, b byte) bool {
	if b&0xC0 != 0x80 {
		return false
	}
	if pos != 1 {
		return true
	}
	switch lead {
	case 0xE0:
		return b >= 0xA0
	case 0xED:
		return b <= 0x9F
	case 0xF0:
		return b >= 0x90
	case 0xF4:
		return b <= 0x8F
	}
	return true
}

func utf8PrefixOK(p []byte) bool {
	for i := 1; i < len(p); i++ {
		if p[i]&0xC0 != 0x80 {
			return false
		}
	}
	return true
}

// utf16Units decodes complete UTF-16 code units the way Node's native decoder
// does. A high surrogate that lands exactly on a chunk boundary is held until
// a full code unit can follow it; an odd trailing byte is held on its own,
// while the complete units before it (even a lone high surrogate) are emitted.
func utf16Units(in, pending []byte) ([]uint16, []byte, int, int) {
	buf := join(pending, in)
	// A high surrogate buffered at a previous boundary still needs two bytes;
	// one byte is not enough to decide whether a low surrogate follows.
	if len(pending) == 2 && len(in) < 2 {
		u := uint16(pending[0]) | uint16(pending[1])<<8
		if u >= 0xD800 && u <= 0xDBFF {
			return nil, buf, len(buf), 4
		}
	}
	end := len(buf) - len(buf)%2
	units := make([]uint16, 0, end/2)
	for i := 0; i+1 < end; i += 2 {
		units = append(units, uint16(buf[i])|uint16(buf[i+1])<<8)
	}
	if len(buf)%2 == 1 {
		hold := append([]byte{}, buf[end:]...)
		return units, hold, 1, 2
	}
	if len(units) > 0 {
		last := units[len(units)-1]
		if last >= 0xD800 && last <= 0xDBFF {
			units = units[:len(units)-1]
			hold := append([]byte{}, buf[end-2:]...)
			return units, hold, len(hold), 4
		}
	}
	return units, nil, 0, 0
}

// unitsValue builds a JavaScript string directly from UTF-16 code units so
// lone surrogates survive (a Go string would replace them with U+FFFD).
func unitsValue(c *js.Context, units []uint16) js.Value {
	if len(units) == 0 {
		return c.Str("")
	}
	android := c.Global().Get("String")
	fn := android.Object().Get("fromCharCode")
	args := make([]js.Value, len(units))
	for i, u := range units {
		args[i] = c.Num(float64(u))
	}
	if v, err := fn.Object().Call(android.Object(), args...); err == nil {
		return v
	}
	return c.Str(string(utf16.Decode(units)))
}

func unitsString(units []uint16) string {
	return string(utf16.Decode(units))
}

func base64Write(in, pending []byte, enc *base64.Encoding) (string, []byte, int, int) {
	buf := join(pending, in)
	n := len(buf) % 3
	complete := len(buf) - n
	out := enc.EncodeToString(buf[:complete])
	if n == 0 {
		return out, nil, 0, 0
	}
	next := append([]byte{}, buf[complete:]...)
	return out, next, 3 - n, 3
}

func latin1String(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		sb.WriteRune(rune(c))
	}
	return sb.String()
}

func asciiString(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		sb.WriteRune(rune(c & 0x7F))
	}
	return sb.String()
}

func join(a, b []byte) []byte {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	out := make([]byte, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"string_decoder.StringDecoder",
		"string_decoder.StringDecoder#end",
		"string_decoder.StringDecoder#write",
	}
}
