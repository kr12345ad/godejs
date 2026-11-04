package string_decoder

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "string_decoder", Install)
	return e
}

func setBytes(t *testing.T, e *env.Env, name string, b []byte) {
	t.Helper()
	e.Global().Set(name, e.JS.Bytes(b).V())
}

func newDecoder(t *testing.T, e *env.Env, name string, enc ...string) {
	t.Helper()
	if len(enc) == 0 {
		e.Global().Set(name, jstest.Eval(t, e, `new string_decoder.StringDecoder()`))
		return
	}
	e.Global().Set("__enc", e.JS.Str(enc[0]))
	e.Global().Set(name, jstest.Eval(t, e, `new string_decoder.StringDecoder(__enc)`))
}

func TestUTF8SplitSequence(t *testing.T) {
	e := setup(t)
	setBytes(t, e, "b1", []byte{0xE2, 0x82})
	setBytes(t, e, "b2", []byte{0xAC})
	newDecoder(t, e, "d", "utf8")

	jstest.Equal(t, jstest.Str(t, e, `d.write(b1)`), "")
	jstest.Equal(t, jstest.Str(t, e, `d.lastNeed`), "1")
	jstest.Equal(t, jstest.Str(t, e, `d.lastTotal`), "3")
	jstest.Equal(t, jstest.Str(t, e, `d.write(b2)`), "\u20ac")
	jstest.Equal(t, jstest.Str(t, e, `d.end()`), "")
}

func TestUTF8EndFlushesReplacement(t *testing.T) {
	e := setup(t)
	setBytes(t, e, "b", []byte{0xE2, 0x82})
	newDecoder(t, e, "d", "utf8")

	jstest.Equal(t, jstest.Str(t, e, `d.write(b)`), "")
	jstest.Equal(t, jstest.Str(t, e, `d.end()`), "\uFFFD")
}

func TestUTF8PassThrough(t *testing.T) {
	e := setup(t)
	setBytes(t, e, "b", []byte("hello \u20ac"))
	newDecoder(t, e, "d", "utf8")

	jstest.Equal(t, jstest.Str(t, e, `d.write(b)`), "hello \u20ac")
	jstest.Equal(t, jstest.Str(t, e, `d.end()`), "")
}

func TestUTF16LEOddByte(t *testing.T) {
	e := setup(t)
	setBytes(t, e, "b1", []byte{0x61})
	setBytes(t, e, "b2", []byte{0x00})
	newDecoder(t, e, "d", "utf16le")

	jstest.Equal(t, jstest.Str(t, e, `d.write(b1)`), "")
	jstest.Equal(t, jstest.Str(t, e, `d.lastNeed`), "1")
	jstest.Equal(t, jstest.Str(t, e, `d.write(b2)`), "a")

	setBytes(t, e, "b3", []byte{0x61, 0x00, 0x62, 0x00})
	newDecoder(t, e, "d2", "utf16le")
	jstest.Equal(t, jstest.Str(t, e, `d2.write(b3)`), "ab")
}

func TestUTF16LEDanglingByteIsDroppedOnEnd(t *testing.T) {
	e := setup(t)
	setBytes(t, e, "b", []byte{0x61})
	newDecoder(t, e, "d", "utf16le")

	jstest.Equal(t, jstest.Str(t, e, `d.write(b)`), "")
	jstest.Equal(t, jstest.Str(t, e, `d.end()`), "")
}

func TestBase64AndBase64URL(t *testing.T) {
	e := setup(t)
	setBytes(t, e, "a", []byte{0x61})
	setBytes(t, e, "b", []byte{0x62})
	setBytes(t, e, "abc", []byte{0x61, 0x62, 0x63})
	setBytes(t, e, "hi", []byte{0xFB, 0xFF})

	newDecoder(t, e, "d", "base64")
	jstest.Equal(t, jstest.Str(t, e, `d.write(a)`), "")
	jstest.Equal(t, jstest.Str(t, e, `d.write(b)`), "")
	jstest.Equal(t, jstest.Str(t, e, `d.end()`), "YWI=")

	newDecoder(t, e, "d2", "base64")
	jstest.Equal(t, jstest.Str(t, e, `d2.write(abc)`), "YWJj")
	jstest.Equal(t, jstest.Str(t, e, `d2.end()`), "")

	newDecoder(t, e, "d3", "base64url")
	jstest.Equal(t, jstest.Str(t, e, `d3.write(hi)`), "")
	jstest.Equal(t, jstest.Str(t, e, `d3.end()`), "-_8")
}

func TestHexLatin1ASCII(t *testing.T) {
	e := setup(t)
	setBytes(t, e, "hexb", []byte{0xDE, 0xAD})
	setBytes(t, e, "lat", []byte{0xE9, 0x41})
	setBytes(t, e, "asc", []byte{0xC1})

	newDecoder(t, e, "dh", "hex")
	jstest.Equal(t, jstest.Str(t, e, `dh.write(hexb)`), "dead")
	jstest.Equal(t, jstest.Str(t, e, `dh.end()`), "")

	newDecoder(t, e, "dl", "latin1")
	jstest.Equal(t, jstest.Str(t, e, `dl.write(lat)`), "\u00e9A")

	newDecoder(t, e, "da", "ascii")
	jstest.Equal(t, jstest.Str(t, e, `da.write(asc)`), "A")
}

func TestAliasesAndDefault(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `new string_decoder.StringDecoder().encoding`), "utf8")
	jstest.Equal(t, jstest.Str(t, e, `new string_decoder.StringDecoder('utf-8').encoding`), "utf8")
	jstest.Equal(t, jstest.Str(t, e, `new string_decoder.StringDecoder('binary').encoding`), "latin1")
	jstest.Equal(t, jstest.Str(t, e, `new string_decoder.StringDecoder('ucs2').encoding`), "utf16le")
	jstest.Equal(t, jstest.Str(t, e, `String(new string_decoder.StringDecoder() instanceof string_decoder.StringDecoder)`), "true")
}

func TestUnknownEncoding(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `(function () {
		try {
			new string_decoder.StringDecoder('wat');
			return 'no throw';
		} catch (err) {
			return err.name + '|' + err.code + '|' + err.message;
		}
	})()`)
	want := "TypeError|ERR_UNKNOWN_ENCODING|Unknown encoding: wat"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	got = jstest.Str(t, e, `(function () {
		try { new string_decoder.StringDecoder(null); }
		catch (err) { return err.code + '|' + err.message; }
		return 'no throw';
	})()`)
	jstest.Equal(t, got, "ERR_UNKNOWN_ENCODING|Unknown encoding: null")
}

func TestImplementedCoversBaseline(t *testing.T) {
	got := Implemented()
	if len(got) != 3 {
		t.Fatalf("Implemented() = %v", got)
	}
}
