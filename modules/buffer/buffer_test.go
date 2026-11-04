package buffer

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func newEnv(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "buffer", Install)
	return e
}

func TestEncodings(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("hello").toString("hex")`), "68656c6c6f")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("hello").toString("base64")`), "aGVsbG8=")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("hello?").toString("base64url")`), "aGVsbG8_")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("aGVsbG8=", "base64").toString()`), "hello")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("aGVsbG8", "base64").toString()`), "hello")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("aGVsbG8_", "base64url").toString()`), "hello?")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("61006200", "hex").toString("utf16le")`), "ab")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("ab", "utf16le").toString("hex")`), "61006200")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("é", "latin1").toString("hex")`), "e9")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("€", "latin1").toString("hex")`), "ac")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from([0xe9]).toString("ascii")`), "i")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("zz", "hex").length`), "0")

	jstest.Equal(t, jstest.Str(t, e, `Buffer.byteLength("€")`), "3")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.byteLength("ab", "utf16le")`), "4")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.byteLength("aGVsbG8=", "base64")`), "5")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.byteLength("abc", "hex")`), "1")
}

func TestStatics(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `Buffer.alloc(4).toString("hex")`), "00000000")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.alloc(3, "a").toString()`), "aaa")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.allocUnsafe(2).length`), "2")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.allocUnsafeSlow(2).length`), "2")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from([1, 2, 3]).toString("hex")`), "010203")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from(new Uint8Array([4, 5])).toString("hex")`), "0405")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from(Buffer.from("ab")).toString()`), "ab")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.isBuffer(Buffer.from("x"))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.isBuffer("x")`), "false")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.isEncoding("UTF-8")`), "true")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.isEncoding("nope")`), "false")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.poolSize`), "65536")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.concat([Buffer.from("ab"), Buffer.from("cd")]).toString()`), "abcd")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.concat([Buffer.from("ab")], 4).toString("hex")`), "61620000")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.compare(Buffer.from("a"), Buffer.from("b"))`), "-1")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.copyBytesFrom(new Uint8Array([1, 2, 3]), 1, 1).toString("hex")`), "02")

	jstest.Equal(t, jstest.Str(t, e, `buffer.constants.MAX_LENGTH`), "9007199254740991")
	jstest.Equal(t, jstest.Str(t, e, `buffer.constants.MAX_STRING_LENGTH`), "536870888")
	jstest.Equal(t, jstest.Str(t, e, `buffer.kMaxLength`), "9007199254740991")
	jstest.Equal(t, jstest.Str(t, e, `buffer.INSPECT_MAX_BYTES`), "50")
	jstest.Equal(t, jstest.Str(t, e, `buffer.btoa("hi")`), "aGk=")
	jstest.Equal(t, jstest.Str(t, e, `buffer.atob("aGk=")`), "hi")
	jstest.Equal(t, jstest.Str(t, e, `buffer.isAscii(Buffer.from("abc"))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `buffer.isAscii(Buffer.from("é"))`), "false")
	jstest.Equal(t, jstest.Str(t, e, `buffer.isUtf8(Buffer.from("€"))`), "true")
}

func TestArrayBufferViews(t *testing.T) {
	e := newEnv(t)

	// Buffer.from(arrayBuffer) shares memory with the source in both directions.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const ab = new ArrayBuffer(4);
		const ui = new Uint8Array(ab);
		const b = Buffer.from(ab);
		b[0] = 9;
		ui[1] = 7;
		return [b.length, b[0], b[1], ui[0], b.parent === ab, b.buffer === ab].join(":");
	})()`), "4:9:7:9:true:true")

	// byteOffset and length select a window of the ArrayBuffer.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const ab = new Uint8Array([1, 2, 3, 4, 5]).buffer;
		const b = Buffer.from(ab, 1, 3);
		b[0] = 9;
		return b.length + ":" + new Uint8Array(ab)[1];
	})()`), "3:9")

	// slice and subarray are views, so they write through to the original.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);
		b.slice(0, 4).swap32();
		return b.toString("hex");
	})()`), "0403020105060708")

	// Out-of-bounds windows carry Node's code and message.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const ab = new ArrayBuffer(5);
		try { Buffer.from(ab, 6); } catch (err) { return err.code + "|" + err.name + "|" + err.message; }
	})()`), `ERR_BUFFER_OUT_OF_BOUNDS|RangeError|"offset" is outside of buffer bounds`)
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const ab = new ArrayBuffer(5);
		try { Buffer.from(ab, 3, 6); } catch (err) { return err.code + "|" + err.message; }
	})()`), `ERR_BUFFER_OUT_OF_BOUNDS|"length" is outside of buffer bounds`)

	// Non-numeric offsets and lengths count as zero, as in Node.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const ab = new ArrayBuffer(4);
		return Buffer.from(ab, "x").length + ":" + Buffer.from(ab, 0, {}).length;
	})()`), "4:0")
}

func TestByteLengthAcrossTypes(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `[
		Buffer.byteLength(Buffer.from("abcd")),
		Buffer.byteLength(new ArrayBuffer(8)),
		Buffer.byteLength(new Int16Array(8)),
		Buffer.byteLength(new DataView(new ArrayBuffer(2))),
		Buffer.byteLength("foo", "nope"),
		Buffer.byteLength("", "hex"),
	].join(":")`), "4:8:16:2:3:0")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { Buffer.byteLength(32); } catch (err) { return err.name + "|" + err.code + "|" + err.message; }
	})()`), `TypeError|ERR_INVALID_ARG_TYPE|The "string" argument must be of type string or an instance of Buffer or ArrayBuffer. Received type number (32)`)
}

func TestNodeValidationMessages(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { Buffer.from([1, 2, 3]).swap16(); } catch (err) { return err.code + "|" + err.name; }
	})()`), "ERR_INVALID_BUFFER_SIZE|RangeError")

	// Array-likes are copied and wrapped modulo 256 the way Node assigns bytes.
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from({length: 3, 0: 257, 1: -1, 2: 1.9}).toString("hex")`), "01ff01")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from({length: NaN}).length`), "0")

	// write() shifts a string offset into the encoding slot.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(4);
		const n = b.write("ab", "utf8");
		return n + ":" + b[0] + b[1] + b[2];
	})()`), "2:97980")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { Buffer.concat([Buffer.from("a"), 5]); } catch (err) { return err.code + "|" + err.message; }
	})()`), `ERR_INVALID_ARG_TYPE|The "list[1]" argument must be an instance of Buffer or Uint8Array. Received type number (5)`)
}

func TestPrototypeMethods(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `Buffer.from([1, 2, 3]).length`), "3")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("hello").toString("utf8", 1, 3)`), "el")
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(Buffer.from([1, 2]))`), `{"type":"Buffer","data":[1,2]}`)
	jstest.Equal(t, jstest.Str(t, e, `Object.keys(Buffer.from("x")).includes("_godeBuffer")`), "false")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("abc").equals(Buffer.from("abc"))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("abc").compare(Buffer.from("abd"))`), "-1")

	jstest.Equal(t, jstest.Str(t, e, `(() => { const a = Buffer.alloc(3); Buffer.from("xy").copy(a, 1); return a.toString("hex"); })()`), "007879")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.alloc(5).fill("ab").toString()`), "ababa")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.alloc(5).fill(255, 1, 3).toString("hex")`), "00ffff0000")
	jstest.Equal(t, jstest.Str(t, e, `(() => { const b = Buffer.alloc(5); b.write("ab", 1); return b.toString("hex"); })()`), "0061620000")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("abc").slice(1).toString()`), "bc")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("abc").subarray(0, 2).toString()`), "ab")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("hello").includes("ell")`), "true")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("hello").indexOf("l")`), "2")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("hello").lastIndexOf("l")`), "3")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("aaa").indexOf("a", 1)`), "1")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from([1, 2, 3, 4]).swap16().toString("hex")`), "02010403")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from([1, 2, 3, 4]).swap32().toString("hex")`), "04030201")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]).swap64().toString("hex")`), "0807060504030201")
}

func TestNumericAccessors(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(8);
		b.writeUInt32LE(0xdeadbeef, 0);
		return b.toString("hex");
	})()`), "efbeadde00000000")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(8);
		b.writeUInt32LE(0xdeadbeef, 0);
		return b.readUInt32LE(0);
	})()`), "3735928559")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(8);
		b.writeUInt16BE(0x1234, 4);
		return b.readUInt16BE(4);
	})()`), "4660")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(4);
		b.writeInt8(-1, 0);
		return b.readInt8(0);
	})()`), "-1")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(8);
		b.writeDoubleLE(1.5, 0);
		return b.readDoubleLE(0);
	})()`), "1.5")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(8);
		b.writeFloatBE(2.5, 0);
		return b.readFloatBE(0);
	})()`), "2.5")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(6);
		b.writeUIntLE(0x123456, 1, 3);
		return b.toString("hex") + ":" + b.readUIntLE(1, 3);
	})()`), "005634120000:1193046")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(2);
		b.writeIntBE(-2, 0, 2);
		return b.toString("hex") + ":" + b.readIntBE(0, 2);
	})()`), "fffe:-2")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.from([0xfe, 0xff]);
		return b.readInt16LE(0);
	})()`), "-2")
}

func TestBigIntAccessors(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(8);
		b.writeBigInt64BE(-2n, 0);
		return b.toString("hex") + ":" + String(b.readBigInt64BE(0));
	})()`), "fffffffffffffffe:-2")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(8);
		b.writeBigUInt64LE(0x0102030405060708n, 0);
		return b.toString("hex") + ":" + String(b.readBigUInt64LE(0));
	})()`), "0807060504030201:72623859790382856")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = Buffer.alloc(8);
		b.writeBigUInt64BE(0xffffffffffffffffn, 0);
		return b.toString("hex") + ":" + String(b.readBigUInt64BE(0));
	})()`), "ffffffffffffffff:18446744073709551615")

	if err := jstest.Throws(t, e, `Buffer.alloc(4).readBigInt64BE(0)`); err.Message == "" {
		t.Fatal("expected a range error")
	}
}

func TestIteratorsAndLegacyProperties(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `[...Buffer.from([1, 2]).entries()].map((x) => x.join(":")).join(",")`), "0:1,1:2")
	jstest.Equal(t, jstest.Str(t, e, `[...Buffer.from([1, 2]).keys()].join(",")`), "0,1")
	jstest.Equal(t, jstest.Str(t, e, `[...Buffer.from([1, 2]).values()].join(",")`), "1,2")
	jstest.Equal(t, jstest.Str(t, e, `(() => { const b = Buffer.from("abc"); return [b.length, b.byteOffset, b.buffer.byteLength, b.parent === b.buffer, b[1]].join(":"); })()`), "3:0:3:true:98")
	jstest.Equal(t, jstest.Str(t, e, `Buffer.from("abc").toLocaleString()`), "abc")
}

func TestConstantsAndSlowBuffer(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `String(buffer.MAX_LENGTH === buffer.constants.MAX_LENGTH && buffer.MAX_STRING_LENGTH === buffer.constants.MAX_STRING_LENGTH)`), "true")
	jstest.Equal(t, jstest.Str(t, e, `buffer.kStringMaxLength`), "536870888")
	jstest.Equal(t, jstest.Str(t, e, `new buffer.SlowBuffer(3).length`), "3")
	jstest.Equal(t, jstest.Str(t, e, `buffer.SlowBuffer(2) instanceof Buffer`), "true")
}

func TestTranscode(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `buffer.transcode(Buffer.from("hi"), "utf8", "utf16le").toString("hex")`), "68006900")
	jstest.Equal(t, jstest.Str(t, e, `buffer.transcode(Buffer.from("h\u0000i\u0000", "latin1"), "utf16le", "utf8").toString()`), "hi")
	jstest.Equal(t, jstest.Str(t, e, `buffer.transcode(new Uint8Array(0), "utf8", "latin1").length`), "0")
	if err := jstest.Throws(t, e, `buffer.transcode(Buffer.from("a"), "b", "utf8")`); err.Code != "U_ILLEGAL_ARGUMENT_ERROR" {
		t.Fatalf("transcode error code = %q", err.Code)
	}
}

func TestBlobAndFile(t *testing.T) {
	e := newEnv(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const b = new Blob(["hello", Buffer.from(" world")], { type: "TEXT/Plain" });
		return [b.size, b.type, b.slice(6).size, b.slice(6).type].join(":");
	})()`), "11:text/plain:5:")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const f = new File(["x"], "a.txt", { lastModified: 10 });
		return [f.name, f.lastModified, f.size, f instanceof Blob].join(":");
	})()`), "a.txt:10:1:true")

	jstest.Eval(t, e, `globalThis.__blobText = null; new Blob(["hello"]).text().then((t) => { globalThis.__blobText = t; });`)
	if err := e.RunMain(); err != nil {
		t.Fatal(err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__blobText`), "hello")

	jstest.Eval(t, e, `globalThis.__blobBytes = null; new Blob([new Uint8Array([1, 2])]).bytes().then((b) => { globalThis.__blobBytes = Array.from(b).join(","); });`)
	if err := e.RunMain(); err != nil {
		t.Fatal(err)
	}
	jstest.Equal(t, jstest.Str(t, e, `globalThis.__blobBytes`), "1,2")

	jstest.Equal(t, jstest.Str(t, e, `(() => { try { new Blob("hello"); return "no"; } catch (e) { return e.name; } })()`), "TypeError")
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const s = new Blob(["hi"]).stream();
		const t = new Blob(["hi"]).textStream();
		return [typeof s.getReader, typeof t.getReader, typeof s[Symbol.asyncIterator]].join(",");
	})()`), "function,function,function")
	jstest.Equal(t, jstest.Str(t, e, `typeof buffer.resolveObjectURL('blob:nodedata:missing')`), "undefined")
}

func TestErrors(t *testing.T) {
	e := newEnv(t)

	for _, src := range []string{
		`Buffer.alloc(-1)`,
		`Buffer.from("x").toString("nope")`,
		`Buffer.alloc(2).readUInt32LE(0)`,
		`Buffer.alloc(2).writeUInt32LE(1, 0)`,
		`Buffer.from({})`,
		`Buffer.concat([1])`,
		`Buffer.from([1, 2, 3]).swap16()`,
	} {
		if err := jstest.Throws(t, e, src); err.Message == "" {
			t.Fatalf("%s: expected a message", src)
		}
	}
}
