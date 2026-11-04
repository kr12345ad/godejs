package zlib

import (
	"encoding/json"
	"os"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/buffer"
	"godejs/modules/stream"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "buffer", buffer.Install)
	jstest.Install(t, e, "stream", stream.Install)
	jstest.Install(t, e, "zlib", Install)
	return e
}

func TestSyncRoundtrip(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var s = 'hello world';
  return zlib.gunzipSync(zlib.gzipSync(s)).toString('utf8');
})()`), "hello world")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var s = 'hello world';
  return zlib.inflateSync(zlib.deflateSync(s)).toString('utf8');
})()`), "hello world")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var s = 'raw hello';
  return zlib.inflateRawSync(zlib.deflateRawSync(s)).toString('utf8');
})()`), "raw hello")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var b = Buffer.from([0, 255, 1, 2, 3, 10, 13, 128]);
  return zlib.unzipSync(zlib.gzipSync(b)).equals(b) + ':' +
    zlib.unzipSync(zlib.deflateSync(b)).equals(b);
})()`), "true:true")
}

func TestHeadersAndEmpty(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var g = zlib.gzipSync('x');
  var d = zlib.deflateSync('x');
  return g[0] + ',' + g[1] + ',' + d[0];
})()`), "31,139,120")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var empty = Buffer.alloc(0);
  return zlib.gunzipSync(zlib.gzipSync(empty)).length + ':' +
    zlib.inflateSync(zlib.deflateSync(empty)).length;
})()`), "0:0")
}

func TestLargeRoundtrip(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var s = 'abc'.repeat(40000);
  var out = zlib.gunzipSync(zlib.gzipSync(s)).toString('utf8');
  return out === s ? 'ok' : 'bad:' + out.length;
})()`), "ok")
}

func TestCorruptInput(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.gunzipSync(Buffer.from('not gzip')); return 'no'; }
  catch (e) { return e.code + ':' + e.name; }
})()`), "Z_DATA_ERROR:Error")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.inflateSync(Buffer.from([1, 2, 3, 4])); return 'no'; }
  catch (e) { return e.code; }
})()`), "Z_DATA_ERROR")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.gzipSync(1); return 'no'; }
  catch (e) { return e.name; }
})()`), "TypeError")
}

func TestCRC32AndConstants(t *testing.T) {
	e := setup(t)
	// IEEE CRC-32 of '123456789' is 0xCBF43926.
	jstest.Equal(t, jstest.Str(t, e, `zlib.crc32('123456789')`), "3421780262")
	jstest.Equal(t, jstest.Str(t, e, `zlib.crc32('')`), "0")
	jstest.Equal(t, jstest.Str(t, e,
		`[zlib.constants.Z_NO_FLUSH, zlib.constants.Z_SYNC_FLUSH, zlib.constants.Z_FINISH,
		  zlib.constants.Z_BEST_SPEED, zlib.constants.Z_DATA_ERROR].join(',')`), "0,2,4,1,-3")
}

func TestCallbackOrdering(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
globalThis.order = [];
zlib.gzip('hi', function (err, res) {
  order.push('cb:' + zlib.gunzipSync(res).toString('utf8'));
});
order.push('sync');`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `order.join(',')`), "sync,cb:hi")
}

func TestCallbackError(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
globalThis.cerr = '';
zlib.gunzip(Buffer.from('bad'), function (err) {
  cerr = err ? err.code : 'none';
});`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `cerr`), "Z_DATA_ERROR")
}

func TestStreamWrapper(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var g = zlib.createGzip();
  var got = null;
  g.on('data', function (d) { got = d; });
  g.end(Buffer.from('stream hello'));
  return zlib.gunzipSync(got).toString('utf8') + ':' + (g instanceof zlib.Gzip);
})()`), "stream hello:true")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var g = zlib.createGzip();
  var parts = [];
  var dest = { write: function (c) { parts.push(c); }, end: function () {} };
  g.pipe(dest);
  g.end('piped');
  return zlib.gunzipSync(Buffer.concat(parts)).toString('utf8');
})()`), "piped")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var g = zlib.createGzip();
  g.write('x');
  var got = null;
  g.on('data', function (d) { got = d; });
  g.flush();
  var n = g.bytesWritten;
  g.reset();
  return zlib.gunzipSync(got).toString('utf8') + ':' + n + ':' + (g instanceof zlib.ZlibBase);
})()`), "x:1:true")
}

func TestConstantsAreFrozen(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  'use strict';
  var out = [];
  try { zlib.constants.Z_OK = 1; out.push('writable'); }
  catch (err) { out.push(err.name); }
  try { zlib.codes = {}; out.push('writable'); }
  catch (err) { out.push('codes:' + err.name); }
  out.push(Object.isFrozen(zlib.codes), Object.isFrozen(zlib.constants));
  out.push(zlib.codes.Z_OK, zlib.codes[0], zlib.codes.Z_DATA_ERROR, zlib.codes[-3]);
  out.push(zlib.constants.Z_MAX_CHUNK);
  return out.join(',');
})()`), "TypeError,codes:TypeError,true,true,0,Z_OK,-3,Z_DATA_ERROR,Infinity")
}

func TestArgumentValidation(t *testing.T) {
	e := setup(t)
	// Option properties are validated when the transform is constructed.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { new zlib.Deflate({ chunkSize: 'test' }); } catch (e) { return e.message; }
})()`), `The "options.chunkSize" property must be of type number. Received type string ('test')`)

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { new zlib.Deflate({ chunkSize: 0 }); } catch (e) { return e.message; }
})()`), `The value of "options.chunkSize" is out of range. It must be >= 64. Received 0`)

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { new zlib.Deflate({ level: -Infinity }); } catch (e) { return e.message; }
})()`), `The value of "options.level" is out of range. It must be a finite number. Received -Infinity`)

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { new zlib.Deflate({ level: -2 }); } catch (e) { return e.code + '|' + e.message; }
})()`), `ERR_OUT_OF_RANGE|The value of "options.level" is out of range. It must be >= -1 and <= 9. Received -2`)

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { new zlib.Deflate({ windowBits: 0 }); } catch (e) { return e.message; }
})()`), `The value of "options.windowBits" is out of range. It must be >= 8 and <= 15. Received 0`)

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { new zlib.Deflate({ dictionary: 'not a buffer' }); } catch (e) { return e.message; }
})()`), `The "options.dictionary" property must be an instance of Buffer, TypedArray, DataView, or ArrayBuffer. Received type string ('not a buffer')`)

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { new zlib.Deflate().params('test'); } catch (e) { return e.message; }
})()`), `The "level" argument must be of type number. Received type string ('test')`)

	// Sync entry points validate the buffer exactly like Node.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.deflateSync(null); } catch (e) { return e.code + '|' + e.message; }
})()`), `ERR_INVALID_ARG_TYPE|The "buffer" argument must be of type string or an instance of Buffer, TypedArray, DataView, or ArrayBuffer. Received null`)

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.deflateSync([1, 2, 3]); } catch (e) { return e.message; }
})()`), `The "buffer" argument must be of type string or an instance of Buffer, TypedArray, DataView, or ArrayBuffer. Received an instance of Array`)

	// The async convenience methods require a callback.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.gzip('abc'); } catch (e) { return e.message; }
})()`), `The "callback" argument must be of type function. Received undefined`)

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.createGzip().flush(99); } catch (e) { return e.code + '|' + e.message; }
})()`), `ERR_OUT_OF_RANGE|The value of "kind" is out of range. It must be >= 0 and <= 5. Received 99`)
}

func TestCRC32Validation(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.crc32('test', null); } catch (e) { return e.code; }
})()`), "ERR_INVALID_ARG_TYPE")
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.crc32(); } catch (e) { return e.code; }
})()`), "ERR_INVALID_ARG_TYPE")
	jstest.Equal(t, jstest.Str(t, e, `zlib.crc32('123456789') === 0xCBF43926 ? 'ok' : 'bad'`), "ok")
}

func TestInfoOption(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var r = zlib.gzipSync('x', { info: true });
  return (r.engine instanceof zlib.Gzip) + ':' + zlib.gunzipSync(r.buffer).toString();
})()`), "true:x")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var r = zlib.deflateSync('y', { info: true });
  return (r.engine instanceof zlib.Deflate) + ':' +
    (zlib.inflateSync(r.buffer, { info: true }).engine instanceof zlib.Inflate);
})()`), "true:true")
}

func TestRejectGarbageAfterEnd(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var d = zlib.gzipSync('a');
  var twice = Buffer.concat([d, d]);
  var plain = zlib.gunzipSync(twice).toString();
  try {
    zlib.gunzipSync(twice, { rejectGarbageAfterEnd: true });
    return plain + ':none';
  } catch (e) {
    return plain + ':' + e.name + ':' + e.code;
  }
})()`), "aa:TypeError:ERR_TRAILING_JUNK_AFTER_STREAM_END")
}

func TestDictionaryRoundtrip(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var dict = Buffer.from('hello world hello world');
  var compressed = zlib.deflateRawSync('hello world', { dictionary: dict });
  return zlib.inflateRawSync(compressed, { dictionary: dict }).toString();
})()`), "hello world")
}

func TestConstructorsWithoutNew(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  return [
    zlib.Deflate() instanceof zlib.Deflate,
    new zlib.Deflate() instanceof zlib.Deflate,
    zlib.Gzip() instanceof zlib.Gzip,
    zlib.createGzip() instanceof zlib.ZlibBase,
  ].join(',');
})()`), "true,true,true,true")
}

func TestTruncatedInput(t *testing.T) {
	e := setup(t)

	// Node reports a truncated stream as Z_BUF_ERROR with errno -5, not as a
	// data error.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var g = zlib.gzipSync('hello world '.repeat(200));
  try { zlib.gunzipSync(g.slice(0, g.length >> 1)); return 'no'; }
  catch (e) { return e.code + ':' + e.errno + ':' + e.message; }
})()`), "Z_BUF_ERROR:-5:unexpected end of file")

	// finishFlush Z_SYNC_FLUSH returns whatever was decoded instead of throwing.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var text = 'hello world '.repeat(200);
  var g = zlib.gzipSync(text);
  var out = zlib.gunzipSync(g.slice(0, g.length >> 1), {
    finishFlush: zlib.constants.Z_SYNC_FLUSH,
  }).toString();
  return text.indexOf(out) === 0 && out.length < text.length ? 'prefix' : 'bad:' + out.length;
})()`), "prefix")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { zlib.inflateSync(Buffer.alloc(0)); return 'no'; }
  catch (e) { return e.code + ':' + e.errno; }
})()`), "Z_BUF_ERROR:-5")
}

func TestProcessChunkAcrossWrites(t *testing.T) {
	e := setup(t)
	// npm's minizlib drives _processChunk once per written chunk with
	// Z_NO_FLUSH and only flushes when the source ends. Each chunk but the
	// last is an incomplete gzip member on its own, so the stream has to keep
	// codec state instead of decoding every chunk in isolation.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var text = 'chunked stream data '.repeat(500);
  var gz = zlib.gzipSync(text);
  var g = zlib.createGunzip();
  var parts = [];
  var step = 97;
  var steps = 0, emptyReturns = 0;
  for (var i = 0; i < gz.length; i += step) {
    var piece = gz.slice(i, Math.min(i + step, gz.length));
    var ret = g._processChunk(piece, zlib.constants.Z_NO_FLUSH);
    // minizlib forwards this return through write(); a falsy value reads as
    // backpressure and stalls the pipe.
    if (ret && ret.length === 0) emptyReturns++;
    parts.push(ret);
    steps++;
  }
  parts.push(g._processChunk(Buffer.alloc(0), zlib.constants.Z_FINISH));
  var kept = parts.filter(function (p) { return p && p.length; });
  return steps === emptyReturns && kept.length === 1 &&
    Buffer.concat(kept).toString() === text ? 'ok' : 'bad';
})()`), "ok")
}

func TestWriteAfterEnd(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  var g = zlib.createGzip();
  g.end('x');
  try { g.write('y'); return 'no'; }
  catch (e) { return e.code + '|' + e.message; }
})()`), "ERR_STREAM_WRITE_AFTER_END|write after end")
}

func TestStreamCloseAndChunkSize(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `
globalThis.closed = false;
globalThis.sizes = [];
var stream = zlib.createGzip({ chunkSize: 64 });
stream.on('close', function () { closed = true; });
stream.on('data', function (d) { sizes.push(d.length); });
var payload = Buffer.alloc(6000);
for (var i = 0; i < payload.length; i++) payload[i] = (i ^ (i >> 3) ^ (i >> 7)) & 0xff;
stream.end(payload);`)
	if err := e.RunMain(); err != nil {
		t.Fatalf("run main: %v", err)
	}
	jstest.Equal(t, jstest.Str(t, e, `String(closed)`), "true")
	jstest.Equal(t, jstest.Str(t, e,
		`sizes.length > 1 && sizes.every(function (n) { return n <= 64; })`), "true")
}

func TestIterTransforms(t *testing.T) {
	e := setup(t)
	jstest.Install(t, e, "ziter", InstallIter)

	// The sync descriptor follows the stream/iter transform contract: it takes a
	// sync iterable of Uint8Array batches and returns one.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  function source(buf) {
    return (function* () { yield [buf]; yield null; })();
  }
  function collect(it) {
    var parts = [];
    for (var batch of it) {
      for (var i = 0; i < batch.length; i++) parts.push(batch[i]);
    }
    return Buffer.concat(parts);
  }
  var compressed = collect(ziter.compressGzipSync().transform(source(Buffer.from('iter hello'))));
  var inflated = collect(ziter.decompressGzipSync().transform(source(compressed)));
  return inflated.toString();
})()`), "iter hello")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  return [
    typeof ziter.compressGzip,
    typeof ziter.compressDeflate,
    typeof ziter.decompressDeflate,
    typeof ziter.compressGzipSync().transform,
  ].join(',');
})()`), "function,function,function,function")

	// Brotli and Zstd run on the bundled pure-Go codecs, including the iter
	// transforms.
	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  function source(buf) { return (function* () { yield [buf]; yield null; })(); }
  function collect(it) {
    var parts = [];
    for (var batch of it) { if (batch === null) continue; for (var i = 0; i < batch.length; i++) parts.push(batch[i]); }
    return Buffer.concat(parts);
  }
  var out = collect(ziter.compressBrotliSync().transform(source(Buffer.from('hello brotli'))));
  var back = collect(ziter.decompressBrotliSync().transform(source(out)));
  return back.toString();
})()`), "hello brotli")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  function source(buf) { return (function* () { yield [buf]; yield null; })(); }
  function collect(it) {
    var parts = [];
    for (var batch of it) { if (batch === null) continue; for (var i = 0; i < batch.length; i++) parts.push(batch[i]); }
    return Buffer.concat(parts);
  }
  var out = collect(ziter.compressZstdSync().transform(source(Buffer.from('hello zstd'))));
  var back = collect(ziter.decompressZstdSync().transform(source(out)));
  return back.toString();
})()`), "hello zstd")

	jstest.Equal(t, jstest.Str(t, e, `
(function () {
  try { ziter.compressGzip('nope'); return 'no'; }
  catch (e) { return e.code + '|' + e.message; }
})()`), `ERR_INVALID_ARG_TYPE|The "options" argument must be of type object. Received type string ('nope')`)
}

func TestImplementedIsSubset(t *testing.T) {
	data, err := os.ReadFile("../../parity/baseline/zlib.json")
	if err != nil {
		t.Fatalf("baseline: %v", err)
	}
	var baseline []string
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatalf("baseline: %v", err)
	}
	tracked := map[string]bool{}
	for _, k := range baseline {
		tracked[k] = true
	}
	seen := map[string]bool{}
	for _, k := range Implemented() {
		if !tracked[k] {
			t.Errorf("Implemented lists untracked key %q", k)
		}
		if seen[k] {
			t.Errorf("duplicate key %q", k)
		}
		seen[k] = true
	}
}
