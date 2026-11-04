// Package zlib implements Node's zlib builtin on top of Go's compress
// packages plus pure-Go Brotli and Zstd codecs, so the whole Node 24 surface
// (Deflate, Gzip, Brotli and Zstd) is available.
package zlib

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.Register("zlib", Install)
	// Node exposes the iterable-stream ("stream/iter") compression transforms
	// as a separate public submodule.
	env.Register("zlib/iter", InstallIter)
}

type module struct {
	e       *env.Env
	exports js.Object
	// kMax is buffer.kMaxLength as captured when zlib loads; Node bounds both
	// output buffers and maxOutputLength with it.
	kMax int64
}

// Node's zlib tuning bounds (src/node_zlib.cc).
const (
	zMinChunk          = 64
	zDefaultChunk      = 16384
	zMinWindowBits     = 8
	zMaxWindowBits     = 15
	zDefaultWindowBits = 15
	zMinLevel          = -1
	zMaxLevel          = 9
	zMinMemLevel       = 1
	zMaxMemLevel       = 9
	zDefaultMemLevel   = 8
	zDefaultLevel      = -1
	zDefaultStrategy   = 0
	zFixed             = 4
	zNoFlush           = 0
	zBlock             = 5
	zFinish            = 4
	zSyncFlush         = 2
	zFullFlush         = 3
	zKMaxLength        = 4294967296
)

// Node's compression stream modes.
const (
	modeDeflate = iota + 1
	modeInflate
	modeGzip
	modeGunzip
	modeDeflateRaw
	modeInflateRaw
	modeUnzip
)

// floatInf is Z_MAX_CHUNK, which Node exposes as Infinity.
var floatInf = math.Inf(1)

// zlibOpts is a validated option set.
type zlibOpts struct {
	kind          string
	level         int
	windowBits    int
	memLevel      int
	strategy      int
	chunkSize     int
	flush         int
	finishFlush   int
	maxOutput     int64
	rejectGarbage bool
	info          bool
	dictionary    []byte

	// Brotli parameters (options.params, brotli/encode.h).
	quality     int
	brotliMode  int
	brotliWin   int
	brotliBlock int

	// Zstd parameters (options.params, zstd.h).
	zstdLevel     int
	zstdWindowLog int
	zstdChecksum  bool

	// pledgedSrcSize is a Zstd-only hint; Node fails when it disagrees with
	// the amount of input actually written.
	pledged    int64
	hasPledged bool
}

func defaultOpts(kind string) *zlibOpts {
	o := &zlibOpts{
		kind:        kind,
		level:       zDefaultLevel,
		windowBits:  zDefaultWindowBits,
		memLevel:    zDefaultMemLevel,
		strategy:    zDefaultStrategy,
		chunkSize:   zDefaultChunk,
		flush:       zNoFlush,
		finishFlush: zFinish,
		maxOutput:   zKMaxLength,
		quality:     11, // BROTLI_DEFAULT_QUALITY
		brotliWin:   22, // BROTLI_DEFAULT_WINDOW
		zstdLevel:   3,  // ZSTD_CLEVEL_DEFAULT
	}
	// Brotli and Zstd have their own operation enums; Z_FINISH is not a valid
	// finishFlush for either of them.
	switch kind {
	case "brotli", "brotliDecompress", "zstd", "zstdDecompress":
		o.finishFlush = 2 // BROTLI_OPERATION_FINISH / ZSTD_e_end
	}
	// The decompression side accepts windowBits 0, meaning "read it from the
	// stream header".
	switch kind {
	case "gunzip", "inflate", "unzip":
		o.windowBits = 0
	}
	return o
}

// iterSource implements the stream/iter transform contract for the zlib
// codecs. Each export returns a descriptor whose transform(source, options)
// consumes batches of Uint8Array (with a trailing null flush marker) and yields
// compressed batches. The codecs themselves are the sync entry points above, so
// the semantics of level/windowBits/dictionary stay identical between the two
// surfaces. Brotli and Zstd have no Go standard-library codec, so those names
// exist and throw ERR_ZLIB_UNSUPPORTED when used rather than silently
// mis-encoding data.
const iterSource = `(function (exports, deps) {
  'use strict';
  var zlib = deps.zlib;
  var BufferCtor = typeof Buffer !== 'undefined' ? Buffer : null;

  function describe(v) {
    if (v === null) return 'null';
    var t = typeof v;
    if (t === 'undefined') return 'undefined';
    if (t === 'number') return 'type number (' + v + ')';
    if (t === 'string') return 'type string (\'' + v + '\')';
    if (t === 'boolean') return 'type boolean (' + v + ')';
    if (t === 'bigint') return 'type bigint (' + v + ')';
    if (t === 'symbol') return 'type symbol';
    if (t === 'function') return 'function ' + (v.name || '(anonymous)');
    if (Array.isArray(v)) return 'an instance of Array';
    if (v.constructor && v.constructor.name) return 'an instance of ' + v.constructor.name;
    return 'an instance of Object';
  }

  function invalidType(name, expected, actual) {
    var err = new TypeError('The "' + name + '" argument must be ' + expected +
      '. Received ' + describe(actual));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function checkOptions(options) {
    if (options === undefined || options === null) return undefined;
    if (typeof options !== 'object' || Array.isArray(options)) {
      throw invalidType('options', 'of type object', options);
    }
    var signal = options.signal;
    if (signal !== undefined && signal !== null) {
      if (typeof signal !== 'object' || typeof signal.aborted !== 'boolean') {
        var err = new TypeError(
          'The "options.signal" property must be an instance of AbortSignal. Received ' +
          describe(signal));
        err.code = 'ERR_INVALID_ARG_TYPE';
        throw err;
      }
    }
    return options;
  }

  function concatenate(parts) {
    if (parts.length === 0) return BufferCtor.alloc(0);
    if (parts.length === 1) return parts[0];
    var total = 0;
    var i;
    for (i = 0; i < parts.length; i++) total += parts[i].byteLength;
    return BufferCtor.concat(parts, total);
  }

  function collectSync(source) {
    var parts = [];
    for (var batch of source) {
      if (batch === null || batch === undefined) continue;
      if (!Array.isArray(batch)) throw invalidType('chunks', 'an instance of Array', batch);
      for (var i = 0; i < batch.length; i++) parts.push(batch[i]);
    }
    return concatenate(parts);
  }

  async function collectAsync(source) {
    var parts = [];
    for await (var batch of source) {
      if (batch === null || batch === undefined) continue;
      if (!Array.isArray(batch)) throw invalidType('chunks', 'an instance of Array', batch);
      for (var i = 0; i < batch.length; i++) parts.push(batch[i]);
    }
    return concatenate(parts);
  }

  function unsupported(name) {
    var err = new Error('zlib/iter: ' + name +
      ' requires a Brotli/Zstd codec, which this build does not include');
    err.code = 'ERR_ZLIB_UNSUPPORTED';
    return err;
  }

  function runner(kind) {
    switch (kind) {
      case 'gzip': return function (b, o) { return zlib.gzipSync(b, o); };
      case 'deflate': return function (b, o) { return zlib.deflateSync(b, o); };
      case 'gunzip': return function (b, o) { return zlib.gunzipSync(b, o); };
      case 'inflate': return function (b, o) { return zlib.inflateSync(b, o); };
      case 'brotli': return function (b, o) { return zlib.brotliCompressSync(b, o); };
      case 'brotliDecompress': return function (b, o) { return zlib.brotliDecompressSync(b, o); };
      case 'zstd': return function (b, o) { return zlib.zstdCompressSync(b, o); };
      case 'zstdDecompress': return function (b, o) { return zlib.zstdDecompressSync(b, o); };
    }
    return null;
  }

  function makeTransform(kind, label, isSync) {
    var run = runner(kind);
    return function (options) {
      var opts = checkOptions(options);
      var descriptor = {};
      descriptor.transform = function (source) {
        if (run === null) throw unsupported(label);
        if (isSync) {
          return (function* () {
            var out = run(collectSync(source), opts);
            if (out && out.byteLength > 0) yield [out];
          })();
        }
        return (async function* () {
          var out = run(await collectAsync(source), opts);
          if (out && out.byteLength > 0) yield [out];
        })();
      };
      return descriptor;
    };
  }

  exports.compressGzip = makeTransform('gzip', 'compressGzip', false);
  exports.compressGzipSync = makeTransform('gzip', 'compressGzipSync', true);
  exports.compressDeflate = makeTransform('deflate', 'compressDeflate', false);
  exports.compressDeflateSync = makeTransform('deflate', 'compressDeflateSync', true);
  exports.compressBrotli = makeTransform('brotli', 'compressBrotli', false);
  exports.compressBrotliSync = makeTransform('brotli', 'compressBrotliSync', true);
  exports.compressZstd = makeTransform('zstd', 'compressZstd', false);
  exports.compressZstdSync = makeTransform('zstd', 'compressZstdSync', true);
  exports.decompressGzip = makeTransform('gunzip', 'decompressGzip', false);
  exports.decompressGzipSync = makeTransform('gunzip', 'decompressGzipSync', true);
  exports.decompressDeflate = makeTransform('inflate', 'decompressDeflate', false);
  exports.decompressDeflateSync = makeTransform('inflate', 'decompressDeflateSync', true);
  exports.decompressBrotli = makeTransform('brotliDecompress', 'decompressBrotli', false);
  exports.decompressBrotliSync = makeTransform('brotliDecompress', 'decompressBrotliSync', true);
  exports.decompressZstd = makeTransform('zstdDecompress', 'decompressZstd', false);
  exports.decompressZstdSync = makeTransform('zstdDecompress', 'decompressZstdSync', true);
})`

// transformSource builds the Transform-shaped wrapper Node programs expect
// from createGzip and friends.
const transformSource = `(function (h, deps) {
  'use strict';
  // Node's zlib streams are Transform streams (lib/zlib.js does
  // ObjectSetPrototypeOf(ZlibBase.prototype, Transform.prototype)), so the
  // whole readable/writable machinery applies to them.
  var Transform = deps && typeof deps.Transform === 'function' ? deps.Transform : null;
  function codedError(name, code, message) {
    var err = new Error(message);
    err.name = name;
    err.code = code;
    return err;
  }
  function unsupportedError(name) {
    return codedError('Error', 'ERR_ZLIB_UNSUPPORTED',
      'zlib: ' + name + ' is not supported by this build (no Brotli or Zstd codec)');
  }
  function describe(v) {
    if (v === null) return 'null';
    var t = typeof v;
    if (t === 'undefined') return 'undefined';
    if (t === 'number') return 'type number (' + v + ')';
    if (t === 'string') return "type string ('" + v + "')";
    if (t === 'boolean') return 'type boolean (' + v + ')';
    if (t === 'bigint') return 'type bigint (' + v + ')';
    if (t === 'function') return 'function ' + (v.name || '(anonymous)');
    if (Array.isArray(v)) return 'an instance of Array';
    return 'an instance of ' + ((v.constructor && v.constructor.name) || 'Object');
  }
  function ZlibBase(kind, options) {
    if (!(this instanceof ZlibBase)) {
      return new ZlibBase(kind, options);
    }
    if (Transform) {
      // Gives the instance _readableState/_writableState and the real stream
      // state machine; the codec itself still runs through the host.
      Transform.call(this, { autoDestroy: true });
    }
    this._kind = kind;
    this._normalized = h.checkOptions(kind, options);
    this._options = options || {};
    this._chunks = [];
    this._pending = 0;
    // Output that no 'data' listener consumed stays available to read(), the
    // way a Transform's readable side buffers it.
    this._readBuf = [];
    this._readLen = 0;
    this._listeners = Object.create(null);
    this._dest = null;
    this._ended = false;
    this._closed = false;
    this._destroyed = false;
    this._error = null;
    this.bytesWritten = 0;
    this.bytesRead = 0;
    this._chunkSize = this._normalized.chunkSize;
    this._maxOutputLength = this._normalized.maxOutputLength;
    this._defaultFlushFlag = this._normalized.flush;
    this._defaultFullFlushFlag = h.fullFlushFlag(this._kind);
    this._finishFlushFlag = this._normalized.finishFlush;
    this._level = this._normalized.level;
    this._strategy = this._normalized.strategy;
    this._rejectGarbageAfterEnd = this._normalized.rejectGarbageAfterEnd === true;
    this._info = this._normalized.info === true;
    this._outOffset = 0;
    this.errored = null;
    this._closeEmitted = false;
    var owner = this;
    // Node keeps a native handle on the stream; tests reach it through
    // _handle (and reset() while a write is pending must fail).
    this._handle = {
      reset: function () {
        if (owner._pending > 0) {
          throw codedError('Error', 'ERR_ZLIB_RESET',
            'Cannot reset zlib stream while a write is in progress');
        }
        owner._chunks = [];
        owner._pending = 0;
        return owner;
      },
      close: function () { owner._handle = null; }
    };
    this.writableEnded = false;
    this.writableFinished = false;
    this.readableEnded = false;
  }
  ZlibBase.prototype.on = function (event, fn) {
    (this._listeners[event] || (this._listeners[event] = [])).push(fn);
    return this;
  };
  ZlibBase.prototype.addListener = ZlibBase.prototype.on;
  ZlibBase.prototype.prependListener = function (event, fn) {
    (this._listeners[event] || (this._listeners[event] = [])).unshift(fn);
    return this;
  };
  ZlibBase.prototype.once = function (event, fn) {
    var self = this;
    function wrapper() {
      self.removeListener(event, wrapper);
      return fn.apply(self, arguments);
    }
    return this.on(event, wrapper);
  };
  ZlibBase.prototype.removeListener = function (event, fn) {
    var list = this._listeners[event];
    if (list) {
      var i = list.indexOf(fn);
      if (i >= 0) list.splice(i, 1);
    }
    return this;
  };
  ZlibBase.prototype.off = ZlibBase.prototype.removeListener;
  ZlibBase.prototype.removeAllListeners = function (event) {
    if (event === undefined) this._listeners = Object.create(null);
    else delete this._listeners[event];
    return this;
  };
  ZlibBase.prototype.listeners = function (event) {
    return (this._listeners[event] || []).slice();
  };
  ZlibBase.prototype.listenerCount = function (event) {
    var list = this._listeners[event];
    return list ? list.length : 0;
  };
  ZlibBase.prototype.eventNames = function () { return Object.keys(this._listeners); };
  ZlibBase.prototype.emit = function (event) {
    var list = this._listeners[event];
    if (!list || list.length === 0) {
      if (event === 'error') this._error = arguments[1];
      return false;
    }
    var args = Array.prototype.slice.call(arguments, 1);
    var copy = list.slice();
    for (var i = 0; i < copy.length; i++) copy[i].apply(this, args);
    return true;
  };
  ZlibBase.prototype._emitClose = function () {
    if (this._closeEmitted) return;
    this._closeEmitted = true;
    this._closed = true;
    this.emit('close');
  };
  ZlibBase.prototype.write = function (chunk, encoding, cb) {
    if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
    // Node's Writable reports a write after the stream was destroyed through
    // the callback (or an 'error' event) instead of throwing at the call site.
    if (this._destroyed || this._closed) {
      var destroyedErr = codedError('Error', 'ERR_STREAM_DESTROYED',
        'Cannot call write after a stream was destroyed');
      if (typeof cb === 'function') { h.later(cb, destroyedErr); return false; }
      var destroyedSelf = this;
      h.later(function () { destroyedSelf.emit('error', destroyedErr); }, null);
      return false;
    }
    if (this._ended) {
      throw codedError('Error', 'ERR_STREAM_WRITE_AFTER_END', 'write after end');
    }
    if (typeof chunk === 'string') {
      chunk = h.toBuffer(chunk, typeof encoding === 'string' ? encoding : 'utf8');
    } else {
      var bytes = h.asBuffer(chunk, 'chunk');
      if (bytes === null) {
        throw h.invalidChunk(chunk);
      }
      chunk = bytes;
    }
    this._chunks.push(chunk);
    this._pending += chunk.length;
    this.bytesWritten += chunk.length;
    if (typeof cb === 'function') h.later(cb, null);
    // A transform reports backpressure the same way Node's ZlibBase does:
    // false once the pending input reaches chunkSize, then 'drain' when the
    // codec has consumed it.
    var ok = this._pending < this._chunkSize;
    if (!ok) this._needDrain = true;
    return ok;
  };
  ZlibBase.prototype._maybeDrain = function () {
    if (!this._needDrain) return;
    this._needDrain = false;
    var self = this;
    h.later(function () {
      if (!self._closed) self.emit('drain');
    }, null);
  };
  ZlibBase.prototype.push = function (chunk) {
    if (chunk !== null && chunk !== undefined) this.emit('data', chunk);
    return true;
  };
  // Node's zlib streams hand the readable side one output buffer per chunkSize
  // (the libuv output buffer), so 'data' listeners never see one giant chunk.
  ZlibBase.prototype._emitData = function (buf) {
    if (!buf || buf.byteLength === 0) return;
    // lib/zlib.js counts bytes as they arrive and fails the stream once the
    // declared maxOutputLength is exceeded.
    if (this._maxOutputLength > 0 && this.bytesRead + buf.byteLength > this._maxOutputLength) {
      var tooLarge = codedError('RangeError', 'ERR_BUFFER_TOO_LARGE',
        'Cannot create a Buffer larger than ' + this._maxOutputLength + ' bytes');
      this._error = tooLarge;
      this.emit('error', tooLarge);
      return;
    }
    this.bytesRead += buf.byteLength;
    var size = this._chunkSize > 0 ? this._chunkSize : buf.byteLength;
    var pieces;
    if (buf.byteLength <= size) {
      pieces = [buf];
    } else {
      pieces = [];
      for (var i = 0; i < buf.byteLength; i += size) {
        pieces.push(buf.slice(i, Math.min(i + size, buf.byteLength)));
      }
    }
    var hasDataListener = this.listenerCount('data') > 0;
    for (var j = 0; j < pieces.length; j++) {
      var piece = pieces[j];
      if (hasDataListener) {
        this.emit('data', piece);
      } else {
        this._readBuf.push(piece);
        this._readLen += piece.byteLength;
      }
      if (this._dest) {
        try { this._dest.write(piece); } catch (e) { /* destination failure surfaces there */ }
      }
    }
  };
  ZlibBase.prototype._finishOutput = function (cb, flushFlag) {
    var out, err;
    try {
      out = this._unsupported ? null :
        h.run(this._kind, this._chunks, this._normalized, flushFlag);
    } catch (e) {
      err = e;
    }
    if (this._unsupported) err = unsupportedError(this._unsupported);
    this._chunks = [];
    this._pending = 0;
    if (err) {
      this._error = err;
      this.errored = err;
      // Node reports codec failures from the async process callback, so a
      // listener attached after end() (finished(), once('error')) sees them.
      var failed = this;
      h.later(function () {
        failed.emit('error', err);
        failed._emitClose();
      }, null);
      if (typeof cb === 'function') h.later(cb, err);
      return false;
    }
    this._emitData(out);
    this._maybeDrain();
    return true;
  };
  ZlibBase.prototype.end = function (chunk, encoding, cb) {
    if (typeof chunk === 'function') { cb = chunk; chunk = undefined; }
    else if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
    if (chunk !== undefined && chunk !== null) this.write(chunk, encoding);
    this.writableEnded = true;
    this._ended = true;
    var ok = this._finishOutput(cb, this._finishFlushFlag);
    if (!ok) { this.writableFinished = false; return this; }
    this.writableFinished = true;
    this.readableEnded = true;
    if (this._dest) {
      try { this._dest.end(); } catch (e) { /* ignore destination teardown races */ }
    }
    var self = this;
    // Node emits end, finish and close on later ticks. Deferring them lets a
    // listener attached right after end() -- finished(), pipe(), once() --
    // still observe the events instead of racing a synchronous emit.
    h.later(function () {
      self.emit('end');
      self.emit('finish');
      self._emitClose();
    }, null);
    if (typeof cb === 'function') h.later(cb, null);
    return this;
  };
  ZlibBase.prototype.flush = function (kind, cb) {
    if (typeof kind === 'function' || (kind === undefined && !cb)) {
      cb = kind;
      kind = h.fullFlushFlag(this._kind);
    }
    kind = h.checkFlush(kind, this._kind);
    if (this.writableFinished) {
      if (typeof cb === 'function') h.later(cb, null);
      return this;
    }
    if (this.writableEnded) {
      if (typeof cb === 'function') this.once('end', cb);
      return this;
    }
    var out, err;
    try {
      out = this._unsupported ? null :
        h.run(this._kind, this._chunks, this._normalized, kind);
    } catch (e) {
      err = e;
    }
    if (this._unsupported) err = unsupportedError(this._unsupported);
    this._chunks = [];
    this._pending = 0;
    if (err) {
      this._error = err;
      this.emit('error', err);
      if (typeof cb === 'function') h.later(cb, err);
      return this;
    }
    if (out && out.length > 0) this._emitData(out);
    this._maybeDrain();
    if (typeof cb === 'function') h.later(cb, null);
    return this;
  };
  ZlibBase.prototype.close = function (cb) {
    // Node implements close() as finished(this, cb) followed by destroy().
    if (this._closeEmitted) {
      if (typeof cb === 'function') h.later(cb, null);
      return this;
    }
    if (typeof cb === 'function') this.once('close', cb);
    this.destroy();
    return this;
  };
  ZlibBase.prototype.destroy = function (err, cb) {
    if (typeof err === 'function') { cb = err; err = undefined; }
    if (this._destroyed) {
      if (typeof cb === 'function') h.later(cb, err || null);
      return this;
    }
    this._destroyed = true;
    this._ended = true;
    this._handle = null;
    this._closed = true;
    this._chunks = [];
    this._pending = 0;
    if (err) {
      this._error = err;
      this.errored = err;
      this.emit('error', err);
    }
    var self = this;
    h.later(function () { self._emitClose(); }, null);
    if (typeof cb === 'function') h.later(cb, err || null);
    return this;
  };
  ZlibBase.prototype.reset = function () {
    if (!this._handle) {
      var e = new Error('zlib binding closed');
      e.name = 'AssertionError';
      e.code = 'ERR_ASSERTION';
      throw e;
    }
    return this._handle.reset();
  };
  ZlibBase.prototype.params = function (level, strategy, cb) {
    if (typeof strategy === 'function') { cb = strategy; strategy = undefined; }
    var checked = h.checkParams(level, strategy);
    if (checked.level !== undefined) this._level = checked.level;
    if (checked.strategy !== undefined) this._strategy = checked.strategy;
    this._normalized.level = this._level;
    if (checked.strategy !== undefined) this._normalized.strategy = checked.strategy;
    if (typeof cb === 'function') h.later(cb, null);
    return this;
  };
  ZlibBase.prototype._processChunk = function (chunk, flushFlag, cb) {
    if (this._outOffset > this._chunkSize) {
      throw codedError('RangeError', 'ERR_OUT_OF_RANGE',
        'The value of "outOffset" is out of range. It must be <= "chunkSize"');
    }
    if (chunk !== undefined && chunk !== null) {
      this._chunks.push(chunk);
      this._pending += chunk.length || 0;
    }
    // _processChunk is the binding-level entry point minizlib drives: it is
    // called once per written chunk with Z_NO_FLUSH and once with the finish
    // flush at end(). The host codecs are whole-buffer, so a non-finishing
    // call only buffers; the finish flush decodes the complete stream. Decoding
    // each lone chunk (and clearing the queue) made every multi-chunk gzip
    // stream fail with Z_BUF_ERROR: unexpected end of file.
    var finishing = flushFlag === this._finishFlushFlag;
    if (!finishing) {
      // Node's binding returns an empty Buffer, not undefined, when the codec
      // has no output yet. minizlib forwards that value through write(), and a
      // caller must see truthy (no backpressure) or its pipe stalls: tar's
      // parser pauses the source on a falsy write and no drain ever follows.
      var empty = Buffer.alloc(0);
      if (typeof cb === 'function') { h.later(cb, null, empty); return; }
      return empty;
    }
    var result, err;
    try {
      result = this._unsupported ? null :
        h.run(this._kind, this._chunks, this._normalized, flushFlag);
    } catch (e) {
      err = e;
    }
    if (this._unsupported) err = unsupportedError(this._unsupported);
    this._chunks = [];
    this._pending = 0;
    if (err) {
      if (typeof cb === 'function') { h.later(cb, err); return; }
      throw err;
    }
    if (typeof cb === 'function') {
      h.later(cb, null, result || Buffer.alloc(0));
      return;
    }
    return result || Buffer.alloc(0);
  };
  ZlibBase.prototype.pipe = function (dest) {
    this._dest = dest;
    return dest;
  };
  ZlibBase.prototype.unpipe = function () {
    this._dest = null;
    return this;
  };
  ZlibBase.prototype.read = function (size) {
    if (this._readLen === 0) return null;
    if (size === undefined || size === null || size >= this._readLen) {
      var parts = this._readBuf;
      this._readBuf = [];
      this._readLen = 0;
      return parts.length === 1 ? parts[0] : Buffer.concat(parts);
    }
    var want = size;
    var taken = [];
    var got = 0;
    while (got < want && this._readBuf.length) {
      var head = this._readBuf[0];
      var need = want - got;
      if (head.length <= need) {
        taken.push(this._readBuf.shift());
        got += head.length;
      } else {
        taken.push(head.slice(0, need));
        this._readBuf[0] = head.slice(need);
        got += need;
      }
    }
    this._readLen -= got;
    return taken.length === 1 ? taken[0] : Buffer.concat(taken);
  };
  ZlibBase.prototype.isPaused = function () { return false; };
  ZlibBase.prototype.pause = function () { return this; };
  ZlibBase.prototype.resume = function () { return this; };
  ZlibBase.prototype.setEncoding = function () { return this; };

  // Transform hooks, mirroring lib/zlib.js. The custom write()/end() paths
  // above stay in charge of ordering, but anything that drives the stream
  // machinery directly (a Transform reader, pipeline internals) still gets
  // correct behaviour from these.
  ZlibBase.prototype._transform = function (chunk, encoding, callback) {
    var buf;
    if (typeof chunk === 'string') buf = h.toBuffer(chunk, encoding || 'utf8');
    else buf = h.asBuffer(chunk);
    if (!buf) { callback(h.invalidChunk(chunk)); return; }
    this._chunks.push(buf);
    this._pending += buf.length;
    this.bytesWritten += buf.length;
    callback();
  };
  ZlibBase.prototype._flush = function (callback) {
    var out = null, err = null;
    try {
      out = this._unsupported ? null :
        h.run(this._kind, this._chunks, this._normalized, this._finishFlushFlag);
    } catch (e) { err = e; }
    this._chunks = [];
    this._pending = 0;
    if (this._unsupported) err = unsupportedError(this._unsupported);
    if (err) { callback(err); return; }
    if (out) this._emitData(out);
    callback();
  };
  ZlibBase.prototype._final = function (callback) { callback(); };
  ZlibBase.prototype._destroy = function (err, callback) {
    this._chunks = [];
    this._pending = 0;
    callback(err);
  };

  if (Transform) {
    Object.setPrototypeOf(ZlibBase.prototype, Transform.prototype);
    Object.setPrototypeOf(ZlibBase, Transform);
  }

  function makeClass(name, kind) {
    function C(options) {
      // Support both new C(...) and the inheritable pattern Node documents:
      // Object.setPrototypeOf(Sub.prototype, C.prototype); C.call(this, ...).
      if (!(this instanceof C)) return new C(options);
      ZlibBase.call(this, kind, options);
    }
    C.prototype = Object.create(ZlibBase.prototype);
    C.prototype.constructor = C;
    Object.defineProperty(C, 'name', { value: name, configurable: true });
    if (Transform) Object.setPrototypeOf(C, ZlibBase);
    return C;
  }

  // Brotli and Zstd parameter bags are validated the way lib/zlib.js validates
  // them even though this build has no codec to configure.
  function checkParamsMap(params, maxKey, invalidCode) {
    if (params === undefined || params === null) return;
    if (typeof params !== 'object') {
      throw codedError('TypeError', 'ERR_INVALID_ARG_TYPE',
        'The "options.params" property must be of type object. Received ' + describe(params));
    }
    Object.keys(params).forEach(function (origKey) {
      var key = +origKey;
      if (isNaN(key) || key < 0 || key > maxKey) {
        throw codedError('TypeError', invalidCode, 'Invalid parameter: ' + origKey);
      }
      var value = params[origKey];
      if (typeof value !== 'number' && typeof value !== 'boolean') {
        throw codedError('TypeError', 'ERR_INVALID_ARG_TYPE',
          'The "options.params[key]" property must be of type number. Received ' + describe(value));
      }
    });
  }
  function makeUnsupportedClass(name, kind, maxParam, invalidCode) {
    function C(options) {
      if (!new.target) return new C(options);
      if (options && options.params !== undefined) {
        checkParamsMap(options.params, maxParam, invalidCode);
      }
      ZlibBase.call(this, kind, options);
      this._unsupported = name;
    }
    C.prototype = Object.create(ZlibBase.prototype);
    C.prototype.constructor = C;
    Object.defineProperty(C, 'name', { value: name, configurable: true });
    if (Transform) Object.setPrototypeOf(C, ZlibBase);
    return C;
  }
  var Gzip = makeClass('Gzip', 'gzip');
  var Gunzip = makeClass('Gunzip', 'gunzip');
  var Deflate = makeClass('Deflate', 'deflate');
  var Inflate = makeClass('Inflate', 'inflate');
  var DeflateRaw = makeClass('DeflateRaw', 'deflateRaw');
  var InflateRaw = makeClass('InflateRaw', 'inflateRaw');
  var Unzip = makeClass('Unzip', 'unzip');
  var BrotliCompress = makeClass('BrotliCompress', 'brotli');
  var BrotliDecompress = makeClass('BrotliDecompress', 'brotliDecompress');
  var ZstdCompress = makeClass('ZstdCompress', 'zstd');
  var ZstdDecompress = makeClass('ZstdDecompress', 'zstdDecompress');

  return {
    ZlibBase: ZlibBase, Gzip: Gzip, Gunzip: Gunzip, Deflate: Deflate,
    Inflate: Inflate, DeflateRaw: DeflateRaw, InflateRaw: InflateRaw,
    Unzip: Unzip,
    BrotliCompress: BrotliCompress, BrotliDecompress: BrotliDecompress,
    ZstdCompress: ZstdCompress, ZstdDecompress: ZstdDecompress,
    createGzip: function (o) { return new Gzip(o); },
    createGunzip: function (o) { return new Gunzip(o); },
    createDeflate: function (o) { return new Deflate(o); },
    createInflate: function (o) { return new Inflate(o); },
    createDeflateRaw: function (o) { return new DeflateRaw(o); },
    createInflateRaw: function (o) { return new InflateRaw(o); },
    createUnzip: function (o) { return new Unzip(o); },
    createBrotliCompress: function (o) { return new BrotliCompress(o); },
    createBrotliDecompress: function (o) { return new BrotliDecompress(o); },
    createZstdCompress: function (o) { return new ZstdCompress(o); },
    createZstdDecompress: function (o) { return new ZstdDecompress(o); }
  };
})`

// InstallIter builds the zlib/iter module: the stream/iter variants of the
// compression transforms. Node exposes each one as a descriptor object with a
// `transform(source, options)` method (lib/internal/streams/iter/transform.js),
// so the protocol glue lives in JavaScript on top of the sync entry points this
// package already installs.
func InstallIter(e *env.Env) js.Object {
	exports := e.JS.Obj()
	deps := e.JS.Obj()
	deps.Set("zlib", e.Exports("zlib").V())
	factory, err := e.JS.Eval(iterSource, "zlib_iter.js")
	if err != nil {
		panic("godejs zlib/iter: " + err.Error())
	}
	if _, err := factory.Object().Call(js.Object{}, exports.V(), deps.V()); err != nil {
		panic("godejs zlib/iter: " + err.Error())
	}
	return exports
}

// Install builds the zlib exports object.
func Install(e *env.Env) js.Object {
	m := &module{e: e}
	// lib/zlib.js captures buffer.kMaxLength when it is first loaded and uses
	// it as the default maxOutputLength and output-buffer bound.
	m.kMax = boundMaxLength(e)
	exports := e.JS.Obj()
	m.exports = exports

	host := e.JS.Obj()
	host.SetFunc("encode", func(this js.Object, args []js.Value) (js.Value, error) {
		s := arg(args, 0).Str()
		return e.JS.Bytes([]byte(s)).V(), nil
	})
	host.SetFunc("toBuffer", func(this js.Object, args []js.Value) (js.Value, error) {
		s := arg(args, 0).Str()
		return m.bytesValue([]byte(s)), nil
	})
	host.SetFunc("asBuffer", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		if data, ok := v.Bytes(); ok {
			return m.bytesValue(data), nil
		}
		return e.JS.Null(), nil
	})
	host.SetFunc("invalidChunk", func(this js.Object, args []js.Value) (js.Value, error) {
		err := invalidType("chunk", "of type string or an instance of Buffer, TypedArray, or DataView", arg(args, 0))
		return e.JS.ErrorValue(err), nil
	})
	host.SetFunc("later", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsFunction() {
			return js.Value{}, js.Throw("TypeError", "The \"callback\" argument must be of type function")
		}
		e.NextTick(args[0], args[1:]...)
		return e.JS.Undefined(), nil
	})
	host.SetFunc("fullFlushFlag", func(this js.Object, args []js.Value) (js.Value, error) {
		// Node's ZlibBase picks the codec's own "flush everything" operation:
		// BROTLI_OPERATION_FLUSH and ZSTD_e_flush are both 1, Z_FULL_FLUSH is 3.
		switch arg(args, 0).Str() {
		case "brotli", "brotliDecompress", "zstd", "zstdDecompress":
			return e.JS.Num(1), nil
		}
		return e.JS.Num(float64(zFullFlush)), nil
	})
	host.SetFunc("checkOptions", func(this js.Object, args []js.Value) (js.Value, error) {
		o, err := parseOptions(arg(args, 0).Str(), arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		return m.normalizedObject(o).V(), nil
	})
	host.SetFunc("checkFlush", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		// Brotli and Zstd use their own operation enums, which end at 3 and 2.
		maxFlush := int64(zBlock)
		switch arg(args, 1).Str() {
		case "brotli", "brotliDecompress":
			maxFlush = 3
		case "zstd", "zstdDecompress":
			maxFlush = 2
		}
		// NaN means "use the default flush", exactly like an omitted kind.
		if v.Valid() && !v.IsUndefined() && !(v.IsNumber() && math.IsNaN(v.Float())) {
			if err := checkRanges(v, "kind", zNoFlush, maxFlush); err != nil {
				return js.Value{}, err
			}
			return e.JS.Num(float64(v.Int())), nil
		}
		return e.JS.Num(float64(zFullFlush)), nil
	})
	host.SetFunc("checkParams", func(this js.Object, args []js.Value) (js.Value, error) {
		out := e.JS.Obj()
		level := arg(args, 0)
		strategy := arg(args, 1)
		if !level.Valid() || level.IsUndefined() || (level.IsNumber() && math.IsNaN(level.Float())) {
			return out.V(), nil
		}
		if err := checkRanges(level, "level", zMinLevel, zMaxLevel); err != nil {
			return js.Value{}, err
		}
		if strategy.Valid() && !strategy.IsUndefined() &&
			!(strategy.IsNumber() && math.IsNaN(strategy.Float())) {
			if err := checkRanges(strategy, "strategy", zDefaultStrategy, zFixed); err != nil {
				return js.Value{}, err
			}
		}
		out.SetNum("level", float64(level.Int()))
		if strategy.Valid() && !strategy.IsUndefined() {
			out.SetNum("strategy", float64(strategy.Int()))
		}
		return out.V(), nil
	})
	host.SetFunc("run", func(this js.Object, args []js.Value) (js.Value, error) {
		kind := arg(args, 0).Str()
		data, err := collectChunks(arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		o, err := parseOptions(kind, arg(args, 2))
		if err != nil {
			return js.Value{}, err
		}
		if fv := arg(args, 3); fv.Valid() && fv.IsNumber() {
			o.finishFlush = fv.Int()
		}
		out, err := m.process(kind, data, o)
		if err != nil {
			return js.Value{}, err
		}
		return m.bytesValue(out), nil
	})

	// zlib streams are Transform streams; hand the prelude the real class so
	// the prototype chain and stream state machines line up with Node.
	deps := e.JS.Obj()
	streamExports := e.Exports("stream")
	if t := streamExports.Get("Transform"); t.Valid() {
		deps.Set("Transform", t)
	}
	if s := streamExports.Get("Stream"); s.Valid() {
		deps.Set("Stream", s)
	}
	source, err := e.JS.Eval(transformSource, "zlib_stream.js")
	if err != nil {
		panic(err)
	}
	classes, err := source.Object().Call(js.Object{}, host.V(), deps.V())
	if err != nil {
		panic(err)
	}
	obj := classes.Object()

	for _, name := range []string{
		"ZlibBase", "Gzip", "Gunzip", "Deflate", "Inflate", "DeflateRaw",
		"InflateRaw", "Unzip", "BrotliCompress", "BrotliDecompress",
		"ZstdCompress", "ZstdDecompress",
	} {
		exports.Set(name, obj.Get(name))
	}
	for _, name := range []string{
		"createGzip", "createGunzip", "createDeflate", "createInflate",
		"createDeflateRaw", "createInflateRaw", "createUnzip",
		"createBrotliCompress", "createBrotliDecompress",
		"createZstdCompress", "createZstdDecompress",
	} {
		exports.Set(name, obj.Get(name))
	}

	// Sync API.
	sync := func(name, kind string) {
		exports.SetFunc(name, func(this js.Object, args []js.Value) (js.Value, error) {
			o, err := parseOptions(kind, arg(args, 1))
			if err != nil {
				return js.Value{}, err
			}
			data, err := m.input(arg(args, 0))
			if err != nil {
				return js.Value{}, err
			}
			out, err := m.process(kind, data, o)
			if err != nil {
				return js.Value{}, err
			}
			buf := m.bytesValue(out)
			if o.info {
				res := e.JS.Obj()
				res.Set("buffer", buf)
				res.Set("engine", m.engine(kind, o))
				return res.V(), nil
			}
			return buf, nil
		})
	}
	sync("gzipSync", "gzip")
	sync("gunzipSync", "gunzip")
	sync("deflateSync", "deflate")
	sync("inflateSync", "inflate")
	sync("deflateRawSync", "deflateRaw")
	sync("inflateRawSync", "inflateRaw")
	sync("unzipSync", "unzip")
	sync("brotliCompressSync", "brotli")
	sync("brotliDecompressSync", "brotliDecompress")
	sync("zstdCompressSync", "zstd")
	sync("zstdDecompressSync", "zstdDecompress")

	// Callback API: the host does the work synchronously and delivers the
	// result on the next tick, the same ordering as Node's threadpool callback.
	async := func(name, kind string) {
		exports.SetFunc(name, func(this js.Object, args []js.Value) (js.Value, error) {
			opts := arg(args, 1)
			cb := arg(args, 2)
			if opts.Valid() && opts.IsFunction() {
				// (buffer, callback)
				cb = opts
				opts = js.Value{}
			}
			if !cb.Valid() || !cb.IsFunction() {
				return js.Value{}, invalidType("callback", "of type function", cb)
			}
			o, err := parseOptions(kind, opts)
			if err != nil {
				return js.Value{}, err
			}
			data, err := m.input(arg(args, 0))
			if err != nil {
				m.deliver(cb, err, nil)
				return e.JS.Undefined(), nil
			}
			out, err := m.process(kind, data, o)
			if err != nil {
				m.deliver(cb, err, nil)
				return e.JS.Undefined(), nil
			}
			if o.info {
				buf := m.bytesValue(out)
				res := e.JS.Obj()
				res.Set("buffer", buf)
				res.Set("engine", m.engine(kind, o))
				m.deliverValue(cb, nil, res.V())
				return e.JS.Undefined(), nil
			}
			m.deliver(cb, nil, out)
			return e.JS.Undefined(), nil
		})
	}
	async("gzip", "gzip")
	async("gunzip", "gunzip")
	async("deflate", "deflate")
	async("inflate", "inflate")
	async("deflateRaw", "deflateRaw")
	async("inflateRaw", "inflateRaw")
	async("unzip", "unzip")
	async("brotliCompress", "brotli")
	async("brotliDecompress", "brotliDecompress")
	async("zstdCompress", "zstd")
	async("zstdDecompress", "zstdDecompress")

	exports.SetFunc("crc32", func(this js.Object, args []js.Value) (js.Value, error) {
		// lib/zlib.js accepts a string or an ArrayBufferView and rejects a bare
		// ArrayBuffer, with the type list in that order.
		first := arg(args, 0)
		var data []byte
		if first.Valid() && first.IsString() {
			data = []byte(first.Str())
		} else if b, ok := first.Bytes(); ok {
			data = b
		} else {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				fmt.Sprintf("The %q argument must be of type Buffer, TypedArray, DataView, or string.", "data")+received(first))
		}
		seed := uint32(0)
		if v := arg(args, 1); v.Valid() && !v.IsUndefined() {
			if !v.IsNumber() {
				return js.Value{}, invalidType("value", "of type number", v)
			}
			f := v.Float()
			if math.IsNaN(f) || f < 0 || f > float64(^uint32(0)) || f != math.Trunc(f) {
				return js.Value{}, outOfRange("value", ">= 0 and <= 4294967295", v)
			}
			seed = uint32(f)
		}
		return e.JS.Num(float64(crc32.Update(seed, crc32.IEEETable, data))), nil
	})

	consts := constants(e)
	exports.SetObj("constants", consts)

	// Node exposes zlib.constants (frozen), zlib.codes (frozen name<->value
	// map) and the deprecated top-level constants. The aliases are installed
	// from JavaScript so the properties are non-writable.
	finalize, ferr := e.JS.Eval(`(function (exports, constants) {
  'use strict';
  var codes = {};
  var names = ['Z_OK', 'Z_STREAM_END', 'Z_NEED_DICT', 'Z_ERRNO', 'Z_STREAM_ERROR',
               'Z_DATA_ERROR', 'Z_MEM_ERROR', 'Z_BUF_ERROR', 'Z_VERSION_ERROR'];
  for (var i = 0; i < names.length; i++) {
    var key = names[i];
    codes[key] = constants[key];
    codes[codes[key]] = key;
  }
  Object.freeze(codes);
  Object.freeze(constants);
  Object.defineProperty(exports, 'constants', {
    value: constants, writable: false, configurable: false, enumerable: true
  });
  Object.defineProperty(exports, 'codes', {
    value: codes, writable: false, configurable: false, enumerable: true
  });
  var keys = Object.keys(constants);
  for (var j = 0; j < keys.length; j++) {
    var k = keys[j];
    // lib/zlib.js keeps BROTLI_* out of the deprecated top-level aliases but
    // does republish the ZSTD_* ones.
    if (k.indexOf('BROTLI') === 0) continue;
    if (Object.prototype.hasOwnProperty.call(exports, k)) continue;
    Object.defineProperty(exports, k, {
      value: constants[k], writable: false, enumerable: false, configurable: false
    });
  }
  return exports;
})`, "zlib_constants.js")
	if ferr != nil {
		panic(ferr)
	}
	if _, ferr := finalize.Object().Call(js.Object{}, exports.V(), consts.V()); ferr != nil {
		panic(ferr)
	}
	return exports
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"zlib.Deflate",
		"zlib.DeflateRaw",
		"zlib.Gunzip",
		"zlib.Gzip",
		"zlib.Inflate",
		"zlib.InflateRaw",
		"zlib.Unzip",
		"zlib.ZlibBase",
		"zlib.ZlibBase#close",
		"zlib.ZlibBase#flush",
		"zlib.ZlibBase#params",
		"zlib.ZlibBase#reset",
		"zlib.ZlibBase.bytesWritten",
		"zlib.constants",
		"zlib.crc32",
		// Brotli and Zstd, including the zlib/iter transforms that the Node 24
		// documentation lists under zlib.
		"zlib.BrotliCompress",
		"zlib.BrotliDecompress",
		"zlib.ZstdCompress",
		"zlib.ZstdDecompress",
		"zlib.brotliCompress",
		"zlib.brotliCompressSync",
		"zlib.brotliDecompress",
		"zlib.brotliDecompressSync",
		"zlib.zstdCompress",
		"zlib.zstdCompressSync",
		"zlib.zstdDecompress",
		"zlib.zstdDecompressSync",
		"zlib.compressBrotli",
		"zlib.compressBrotliSync",
		"zlib.compressDeflate",
		"zlib.compressDeflateSync",
		"zlib.compressGzip",
		"zlib.compressGzipSync",
		"zlib.compressZstd",
		"zlib.compressZstdSync",
		"zlib.decompressBrotli",
		"zlib.decompressBrotliSync",
		"zlib.decompressDeflate",
		"zlib.decompressDeflateSync",
		"zlib.decompressGzip",
		"zlib.decompressGzipSync",
		"zlib.decompressZstd",
		"zlib.decompressZstdSync",
		"zlib.createBrotliCompress",
		"zlib.createBrotliDecompress",
		"zlib.createZstdCompress",
		"zlib.createZstdDecompress",
		"zlib.createDeflate",
		"zlib.createDeflateRaw",
		"zlib.createGunzip",
		"zlib.createGzip",
		"zlib.createInflate",
		"zlib.createInflateRaw",
		"zlib.createUnzip",
		"zlib.deflate",
		"zlib.deflateRaw",
		"zlib.deflateRawSync",
		"zlib.deflateSync",
		"zlib.gunzip",
		"zlib.gunzipSync",
		"zlib.gzip",
		"zlib.gzipSync",
		"zlib.inflate",
		"zlib.inflateRaw",
		"zlib.inflateRawSync",
		"zlib.inflateSync",
		"zlib.unzip",
		"zlib.unzipSync",
	}
}

// engine builds the object an `info: true` call returns as `engine`. It is a
// fresh transform instance, which is all the one-shot convenience methods need
// beyond its constructor identity (`result.engine instanceof zlib.Gzip`).
func (m *module) engine(kind string, o *zlibOpts) js.Value {
	// During Install the module is not yet in the registry, so read the
	// exports object being built instead of re-entering Exports("zlib").
	ctor := m.exports.Get(engineClass(kind))
	if !ctor.Valid() || !ctor.IsFunction() {
		return m.e.JS.Obj().V()
	}
	factory, err := m.e.JS.Eval("(function (C) { return new C(); })", "[zlib:engine]")
	if err != nil {
		return m.e.JS.Obj().V()
	}
	inst, err := factory.Object().Call(js.Object{}, ctor)
	if err != nil {
		return m.e.JS.Obj().V()
	}
	return inst
}

func engineClass(kind string) string {
	switch kind {
	case "gzip":
		return "Gzip"
	case "gunzip":
		return "Gunzip"
	case "deflate":
		return "Deflate"
	case "inflate":
		return "Inflate"
	case "deflateRaw":
		return "DeflateRaw"
	case "inflateRaw":
		return "InflateRaw"
	case "unzip":
		return "Unzip"
	case "brotli":
		return "BrotliCompress"
	case "brotliDecompress":
		return "BrotliDecompress"
	case "zstd":
		return "ZstdCompress"
	case "zstdDecompress":
		return "ZstdDecompress"
	}
	return "ZlibBase"
}

// process performs one zlib operation.
func (m *module) process(kind string, data []byte, o *zlibOpts) ([]byte, error) {
	// Node allocates its output buffer at chunkSize, so a chunk larger than
	// buffer.kMaxLength fails before any codec work happens. lib/zlib.js
	// captures kMaxLength when zlib is first loaded, which is why the
	// kmaxlength tests set it only for the duration of `require('zlib')`.
	kMax := m.kMax
	if kMax <= 0 {
		kMax = boundMaxLength(m.e)
	}
	if kMax > 0 && int64(o.chunkSize) > kMax {
		return nil, bufferTooLarge(kMax)
	}
	out, err := compress(kind, data, o)
	if err != nil {
		return nil, err
	}
	if o.maxOutput > 0 && int64(len(out)) > o.maxOutput {
		return nil, bufferTooLarge(o.maxOutput)
	}
	if kMax > 0 && int64(len(out)) > kMax {
		return nil, bufferTooLarge(kMax)
	}
	return out, nil
}

// boundMaxLength reads buffer.kMaxLength the way lib/zlib.js captures it at
// load time. A build without the buffer module keeps the zlib defaults.
func boundMaxLength(e *env.Env) int64 {
	exports := e.Exports("buffer")
	if !exports.Valid() {
		return 0
	}
	v := exports.Get("kMaxLength")
	if !v.Valid() || !v.IsNumber() {
		return 0
	}
	f := v.Float()
	if f <= 0 || math.IsInf(f, 0) || math.IsNaN(f) {
		return 0
	}
	return int64(f)
}

func bufferTooLarge(max int64) error {
	return js.ThrowCode("RangeError", "ERR_BUFFER_TOO_LARGE",
		fmt.Sprintf("Cannot create a Buffer larger than %d bytes", max))
}

// compress performs one zlib operation. kind is gzip, gunzip, deflate,
// inflate, deflateRaw, inflateRaw or unzip.
func compress(kind string, data []byte, o *zlibOpts) ([]byte, error) {
	if o == nil {
		o = defaultOpts(kind)
	}
	var buf bytes.Buffer
	switch kind {
	case "gzip":
		// An empty member is a fixed 20-byte container; Go's encoder emits a
		// 23-byte stored block for it, so write the canonical bytes instead.
		if len(data) == 0 {
			return []byte{0x1f, 0x8b, 8, 0, 0, 0, 0, 0, 0, 3, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil
		}
		w, err := gzip.NewWriterLevel(&buf, clampLevel(o.level))
		if err != nil {
			return nil, levelErr(err)
		}
		if _, err := w.Write(data); err != nil {
			return nil, dataErr(err)
		}
		if err := w.Close(); err != nil {
			return nil, dataErr(err)
		}
		return buf.Bytes(), nil
	case "deflate":
		w, err := zlib.NewWriterLevelDict(&buf, clampLevel(o.level), o.dictionary)
		if err != nil {
			return nil, levelErr(err)
		}
		if _, err := w.Write(data); err != nil {
			return nil, dataErr(err)
		}
		if err := w.Close(); err != nil {
			return nil, dataErr(err)
		}
		return buf.Bytes(), nil
	case "deflateRaw":
		w, err := flate.NewWriterDict(&buf, clampLevel(o.level), o.dictionary)
		if err != nil {
			return nil, levelErr(err)
		}
		if _, err := w.Write(data); err != nil {
			return nil, dataErr(err)
		}
		if err := w.Close(); err != nil {
			return nil, dataErr(err)
		}
		return buf.Bytes(), nil
	case "gunzip":
		return inflateGzip(data, o)
	case "inflate":
		return inflateZlib(data, o)
	case "inflateRaw":
		return inflateRawStream(data, o)
	case "unzip":
		if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
			return inflateGzip(data, o)
		}
		return inflateZlib(data, o)
	case "brotli":
		return compressBrotli(data, o)
	case "brotliDecompress":
		return decompressBrotli(data, o)
	case "zstd":
		return compressZstd(data, o)
	case "zstdDecompress":
		return decompressZstd(data, o)
	}
	return nil, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE", "Unknown zlib operation: "+kind)
}

// parseParams validates an options.params bag against the parameter space of
// the selected codec, the way Node's ZlibBase does. Integer-like keys are
// re-serialised so a key such as "00" is rejected, matching Node.
func parseParams(v js.Value, kind string, o *zlibOpts) error {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return nil
	}
	if !v.IsObject() || v.IsFunction() {
		return invalidType("options.params", "of type object", v)
	}
	brotliKind := kind == "brotli" || kind == "brotliDecompress"
	zstdKind := kind == "zstd" || kind == "zstdDecompress"
	if !brotliKind && !zstdKind {
		return nil
	}
	maxKey := 8
	code := "ERR_BROTLI_INVALID_PARAM"
	label := "Brotli"
	if zstdKind {
		label = "zstd"
		code = "ERR_ZSTD_INVALID_PARAM"
		maxKey = 402
		if kind == "zstdDecompress" {
			maxKey = 100
		}
	}
	obj := v.Object()
	for _, origKey := range obj.OwnEnumerableKeys() {
		key, err := strconv.Atoi(origKey)
		if err != nil || key < 0 || key > maxKey || strconv.Itoa(key) != origKey {
			return js.ThrowCode("RangeError", code, origKey+" is not a valid "+label+" parameter")
		}
		val := obj.Get(origKey)
		if !val.IsNumber() && !val.IsBool() {
			return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				fmt.Sprintf("The \"options.params[%s]\" property must be of type number. Received %s", origKey, received(val)))
		}
		n := 0
		if val.IsNumber() {
			n = val.Int()
		} else if val.Bool() {
			n = 1
		}
		if brotliKind {
			switch key {
			case 0: // BROTLI_PARAM_MODE
				if n < 0 || n > 2 {
					return initFailed("Initialization failed")
				}
				o.brotliMode = n
			case 1: // BROTLI_PARAM_QUALITY
				if n < 0 || n > 11 {
					return initFailed("Initialization failed")
				}
				o.quality = n
			case 2: // BROTLI_PARAM_LGWIN
				if n != 0 && (n < 10 || n > 24) {
					return initFailed("Initialization failed")
				}
				o.brotliWin = n
			case 3: // BROTLI_PARAM_LGBLOCK
				if n != 0 && (n < 16 || n > 24) {
					return initFailed("Initialization failed")
				}
				o.brotliBlock = n
			case 4: // BROTLI_PARAM_DISABLE_LITERAL_CONTEXT_MODELING
				if n != 0 && n != 1 {
					return initFailed("Initialization failed")
				}
			}
			continue
		}
		if kind == "zstdDecompress" {
			if key == 100 { // ZSTD_d_windowLogMax
				if n != 0 && n < 10 {
					return initFailed("Setting parameter failed")
				}
				o.zstdWindowLog = n
			}
			continue
		}
		switch key {
		case 100: // ZSTD_c_compressionLevel
			if n < -131072 || n > 22 {
				return initFailed("Setting parameter failed")
			}
			o.zstdLevel = n
		case 101: // ZSTD_c_windowLog
			if n != 0 && (n < 10 || n > 31) {
				return initFailed("Setting parameter failed")
			}
			o.zstdWindowLog = n
		case 107: // ZSTD_c_strategy
			if n < 1 || n > 9 {
				return initFailed("Setting parameter failed")
			}
		case 160, 200, 201, 202: // long-distance mode and boolean flags
			if n != 0 && n != 1 {
				return initFailed("Setting parameter failed")
			}
			if key == 201 {
				o.zstdChecksum = n != 0
			}
		}
	}
	return nil
}

// initFailed is the error libbrotli/libzstd report for a parameter the codec
// cannot accept, which Node surfaces as ERR_ZLIB_INITIALIZATION_FAILED.
func initFailed(message string) error {
	return js.ThrowCode("Error", "ERR_ZLIB_INITIALIZATION_FAILED", message)
}

func compressBrotli(data []byte, o *zlibOpts) ([]byte, error) {
	opts := brotli.WriterOptions{Quality: o.quality}
	if o.brotliWin >= 10 && o.brotliWin <= 24 {
		opts.LGWin = o.brotliWin
	}
	var buf bytes.Buffer
	w := brotli.NewWriterOptions(&buf, opts)
	if _, err := w.Write(data); err != nil {
		return nil, js.ThrowCode("Error", "Z_DATA_ERROR", err.Error())
	}
	if err := w.Close(); err != nil {
		return nil, js.ThrowCode("Error", "Z_DATA_ERROR", err.Error())
	}
	return buf.Bytes(), nil
}

func decompressBrotli(data []byte, o *zlibOpts) ([]byte, error) {
	out, err := io.ReadAll(brotli.NewReader(bytes.NewReader(data)))
	if err == nil {
		return out, nil
	}
	// libbrotli stops at the end of the first stream and leaves whatever
	// follows it untouched. The Go reader instead reports the leftover bytes
	// as excessive input, so a completed stream still yields its output; Node
	// discards that trailing data unless rejectGarbageAfterEnd is set.
	if len(out) > 0 && strings.Contains(err.Error(), "excessive input") {
		if o != nil && o.rejectGarbage {
			return nil, trailingJunk()
		}
		return out, nil
	}
	return nil, js.ThrowCode("Error", "Z_DATA_ERROR", err.Error())
}

func compressZstd(data []byte, o *zlibOpts) ([]byte, error) {
	if o.hasPledged && o.pledged != int64(len(data)) {
		errno := 72 // ZSTD_error_srcSize_wrong
		return nil, &js.Error{
			Name:    "Error",
			Code:    "ZSTD_error_srcSize_wrong",
			Message: "Pledged source size does not match the actual source size",
			Errno:   &errno,
		}
	}
	opts := []zstd.EOption{
		zstd.WithEncoderConcurrency(1),
		zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(o.zstdLevel)),
	}
	if o.zstdChecksum {
		opts = append(opts, zstd.WithEncoderCRC(true))
	}
	if o.zstdWindowLog >= 10 && o.zstdWindowLog <= 31 {
		opts = append(opts, zstd.WithWindowSize(1<<uint(o.zstdWindowLog)))
	}
	enc, err := zstd.NewWriter(nil, opts...)
	if err != nil {
		return nil, js.ThrowCode("Error", "ZSTD_error_GENERIC", err.Error())
	}
	defer enc.Close()
	return enc.EncodeAll(data, nil), nil
}

func decompressZstd(data []byte, o *zlibOpts) ([]byte, error) {
	out, err := zstdDecodeAll(data, o)
	if err == nil {
		return out, nil
	}
	if strings.Contains(err.Error(), "window") {
		errno := 16 // ZSTD_error_frameParameter_windowTooLarge
		return nil, &js.Error{
			Name:    "Error",
			Code:    "ZSTD_error_frameParameter_windowTooLarge",
			Message: "Frame requires too much memory for decoding",
			Errno:   &errno,
		}
	}
	// The decoder advances to the next frame after one ends, so bytes that are
	// not a frame surface as a magic/EOF error. Node stops at the end of the
	// last complete frame and discards the remainder unless asked to reject it.
	// Corruption inside a frame is never treated as trailing junk.
	msg := err.Error()
	trailing := errors.Is(err, zstd.ErrMagicMismatch) || strings.Contains(msg, "magic number mismatch")
	truncated := errors.Is(err, io.ErrUnexpectedEOF) || strings.Contains(msg, "unexpected EOF")
	if trailing || truncated {
		if prefix, ok := zstdCompletePrefix(data, o); ok {
			if o != nil && o.rejectGarbage {
				return nil, trailingJunk()
			}
			return prefix, nil
		}
	}
	if truncated {
		return nil, bufErr()
	}
	return nil, js.ThrowCode("Error", "Z_DATA_ERROR", err.Error())
}

func zstdDecodeAll(data []byte, o *zlibOpts) ([]byte, error) {
	opts := []zstd.DOption{zstd.WithDecoderConcurrency(1)}
	if o != nil && o.zstdWindowLog >= 10 {
		opts = append(opts, zstd.WithDecoderMaxWindow(uint64(1)<<uint(o.zstdWindowLog)))
	}
	dec, err := zstd.NewReader(nil, opts...)
	if err != nil {
		return nil, js.ThrowCode("Error", "ZSTD_error_GENERIC", err.Error())
	}
	defer dec.Close()
	return dec.DecodeAll(data, nil)
}

// zstdCompletePrefix finds the longest prefix that decodes as complete frames
// while leaving at most a short tail of non-frame bytes. The search is bounded
// because it only has to cover junk suffixes (and partial frame headers), not
// arbitrary input sizes.
func zstdCompletePrefix(data []byte, o *zlibOpts) ([]byte, bool) {
	// Bound the work: retrying a huge buffer dozens of times would be worse
	// than reporting the original error.
	if len(data) > 4<<20 {
		return nil, false
	}
	maxTrim := len(data) - 1
	if maxTrim > 64 {
		maxTrim = 64
	}
	for trim := 1; trim <= maxTrim; trim++ {
		out, err := zstdDecodeAll(data[:len(data)-trim], o)
		if err == nil && len(out) > 0 {
			return out, true
		}
	}
	return nil, false
}

// zlibErrno maps a zlib return code onto the negative errno Node attaches to
// the same errors (zlib.types / node_zlib.cc).
func zlibErrno(code string) (float64, bool) {
	switch code {
	case "Z_OK":
		return 0, true
	case "Z_STREAM_END":
		return 1, true
	case "Z_NEED_DICT":
		return 2, true
	case "Z_ERRNO":
		return -1, true
	case "Z_STREAM_ERROR":
		return -2, true
	case "Z_DATA_ERROR":
		return -3, true
	case "Z_MEM_ERROR":
		return -4, true
	case "Z_BUF_ERROR":
		return -5, true
	case "Z_VERSION_ERROR":
		return -6, true
	}
	return 0, false
}

func clampLevel(level int) int {
	if level < zMinLevel || level > zMaxLevel {
		return zDefaultLevel
	}
	return level
}

func inflateGzip(data []byte, o *zlibOpts) ([]byte, error) {
	var out []byte
	rest := data
	for {
		// Check the magic ourselves: Go's reader reports a short header as an
		// unexpected EOF, while Node reports a wrong magic as a data error even
		// when the stream is too short to contain the full header.
		if len(rest) < 2 {
			return nil, bufErr()
		}
		if rest[0] != 0x1f || rest[1] != 0x8b {
			return nil, headerErr()
		}
		if len(rest) >= 4 && rest[2] != 8 {
			return nil, methodErr()
		}
		br := bytes.NewReader(rest)
		r, err := gzip.NewReader(br)
		if err != nil {
			return nil, dataErr(err)
		}
		// One member at a time, so trailing junk can be told apart from a real
		// concatenated member.
		r.Multistream(false)
		chunk, err := io.ReadAll(r)
		// With finishFlush Z_SYNC_FLUSH Node returns whatever was decoded
		// instead of failing on an incomplete stream.
		if err != nil && o.finishFlush != zSyncFlush {
			return nil, dataErr(err)
		}
		if cerr := r.Close(); cerr != nil && err == nil && o.finishFlush != zSyncFlush {
			return nil, dataErr(cerr)
		}
		out = append(out, chunk...)
		rest = rest[len(rest)-br.Len():]
		// Node ignores zero padding after the last member.
		for len(rest) > 0 && rest[0] == 0 {
			rest = rest[1:]
		}
		if len(rest) == 0 {
			return out, nil
		}
		if o.rejectGarbage {
			return nil, trailingJunk()
		}
		if rest[0] == 0x1f && len(rest) >= 2 && rest[1] == 0x8b {
			continue
		}
		// Anything else is garbage after the stream end; Node discards it.
		return out, nil
	}
}

// methodErr is zlib's Z_DATA_ERROR for a gzip member whose compression method
// is not deflate.
func methodErr() error {
	errno := -3 // Z_DATA_ERROR
	return &js.Error{Name: "Error", Code: "Z_DATA_ERROR", Message: "unknown compression method", Errno: &errno}
}

func inflateZlib(data []byte, o *zlibOpts) ([]byte, error) {
	if len(data) < 2 {
		return nil, bufErr()
	}
	if err := checkZlibHeader(data); err != nil {
		return nil, err
	}
	br := bytes.NewReader(data)
	r, err := zlib.NewReaderDict(br, o.dictionary)
	if err != nil {
		return nil, dataErr(err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		if o.finishFlush == zSyncFlush {
			return out, nil
		}
		return nil, dataErr(err)
	}
	if err := r.Close(); err != nil {
		return nil, dataErr(err)
	}
	if o.rejectGarbage && br.Len() > 0 {
		return nil, trailingJunk()
	}
	return out, nil
}

// checkZlibHeader validates the two-byte zlib header the way zlib itself does
// before any decompression work happens.
func checkZlibHeader(data []byte) error {
	cmf, flg := data[0], data[1]
	if cmf&0x0f != 8 || cmf>>4 > 7 {
		return headerErr()
	}
	if (uint16(cmf)<<8|uint16(flg))%31 != 0 {
		return headerErr()
	}
	if flg&0x20 != 0 && len(data) < 6 {
		return headerErr()
	}
	return nil
}

// bufErr is zlib's Z_BUF_ERROR: the input ended before the stream was complete.
func bufErr() error {
	errno := -5 // Z_BUF_ERROR
	return &js.Error{Name: "Error", Code: "Z_BUF_ERROR", Message: "unexpected end of file", Errno: &errno}
}

// headerErr is zlib's Z_DATA_ERROR for an unreadable stream header.
func headerErr() error {
	errno := -3 // Z_DATA_ERROR
	return &js.Error{Name: "Error", Code: "Z_DATA_ERROR", Message: "incorrect header check", Errno: &errno}
}

func inflateRawStream(data []byte, o *zlibOpts) ([]byte, error) {
	if len(data) == 0 {
		return nil, bufErr()
	}
	br := bytes.NewReader(data)
	r := flate.NewReaderDict(br, o.dictionary)
	out, err := io.ReadAll(r)
	if err != nil {
		if o.finishFlush == zSyncFlush {
			return out, nil
		}
		return nil, dataErr(err)
	}
	if err := r.Close(); err != nil {
		return nil, dataErr(err)
	}
	if o.rejectGarbage && br.Len() > 0 {
		return nil, trailingJunk()
	}
	return out, nil
}

func trailingJunk() error {
	return js.ThrowCode("TypeError", "ERR_TRAILING_JUNK_AFTER_STREAM_END",
		"Trailing junk found after the end of the compressed stream")
}

func levelErr(err error) error {
	return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE", "The value of \"level\" is out of range. It must be >= -1 and <= 9")
}

// dataErr maps Go compression failures onto the single Node error code zlib
// uses for malformed input.
func dataErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) ||
		strings.Contains(err.Error(), "unexpected EOF") ||
		strings.Contains(err.Error(), "unexpected end of file") {
		return bufErr()
	}
	msg := "incorrect header check"
	switch {
	case errors.Is(err, gzip.ErrHeader), errors.Is(err, zlib.ErrHeader):
		msg = "incorrect header check"
	case errors.Is(err, gzip.ErrChecksum), errors.Is(err, zlib.ErrChecksum):
		msg = "incorrect data check"
	case errors.Is(err, zlib.ErrDictionary):
		msg = "need dictionary"
	}
	var corrupt flate.CorruptInputError
	if errors.As(err, &corrupt) {
		msg = "invalid distance too far back"
	}
	return &js.Error{Name: "Error", Code: "Z_DATA_ERROR", Message: msg}
}

func (m *module) bytesValue(data []byte) js.Value {
	raw := m.e.JS.Bytes(data)
	ctor := m.e.Global().Get("Buffer")
	if ctor.Valid() && ctor.IsFunction() {
		from := ctor.Object().Get("from")
		if from.IsFunction() {
			if v, err := from.Object().Call(ctor.Object(), raw.V()); err == nil {
				return v
			}
		}
	}
	return raw.V()
}

// input accepts everything Node's zlibBufferSync accepts: strings, every
// ArrayBufferView and ArrayBuffer itself.
func (m *module) input(v js.Value) ([]byte, error) {
	if v.Valid() && v.IsString() {
		return []byte(v.Str()), nil
	}
	if data, ok := v.Bytes(); ok {
		// A typed array can lie about its length. Node validates the declared
		// byteLength against the backing store and throws ERR_OUT_OF_RANGE
		// instead of reading past the real bytes.
		if v.Valid() && v.IsObject() && !v.IsArray() {
			if bl := v.Object().Get("byteLength"); bl.Valid() && bl.IsNumber() &&
				int64(bl.Int()) != int64(len(data)) {
				return nil, outOfRange("buffer", "of type TypedArray or DataView", v)
			}
		}
		return data, nil
	}
	if v.Valid() && v.IsObject() && !v.IsArray() && !v.IsFunction() {
		// ArrayBuffer: Buffer.from(ab) shares the bytes and is a Uint8Array.
		if data, ok := m.arrayBufferBytes(v); ok {
			return data, nil
		}
	}
	return nil, invalidType("buffer", "of type string or an instance of Buffer, TypedArray, DataView, or ArrayBuffer", v)
}

// arrayBufferBytes converts an ArrayBuffer through Buffer.from, which throws,
// rather than wrapping arbitrary objects.
func (m *module) arrayBufferBytes(v js.Value) ([]byte, bool) {
	ctor := m.e.Global().Get("Buffer")
	if !ctor.Valid() || !ctor.IsFunction() {
		return nil, false
	}
	from := ctor.Object().Get("from")
	if !from.Valid() || !from.IsFunction() {
		return nil, false
	}
	buf, err := from.Object().Call(ctor.Object(), v)
	if err != nil {
		return nil, false
	}
	return buf.Bytes()
}

func (m *module) normalizedObject(o *zlibOpts) js.Object {
	out := m.e.JS.Obj()
	out.SetNum("level", float64(o.level))
	out.SetNum("windowBits", float64(o.windowBits))
	out.SetNum("memLevel", float64(o.memLevel))
	out.SetNum("strategy", float64(o.strategy))
	out.SetNum("chunkSize", float64(o.chunkSize))
	out.SetNum("flush", float64(o.flush))
	out.SetNum("finishFlush", float64(o.finishFlush))
	out.SetNum("maxOutputLength", float64(o.maxOutput))
	out.SetBool("rejectGarbageAfterEnd", o.rejectGarbage)
	out.SetBool("info", o.info)
	if o.dictionary != nil {
		out.Set("dictionary", m.bytesValue(o.dictionary))
	}
	// The JS stream keeps this normalised bag and hands it back to the host on
	// every flush, so codec parameters have to survive the round trip.
	if o.kind == "brotli" || o.kind == "brotliDecompress" {
		params := m.e.JS.Obj()
		params.SetNum("0", float64(o.brotliMode))
		params.SetNum("1", float64(o.quality))
		params.SetNum("2", float64(o.brotliWin))
		params.SetNum("3", float64(o.brotliBlock))
		out.Set("params", params.V())
	}
	if o.kind == "zstd" {
		params := m.e.JS.Obj()
		params.SetNum("100", float64(o.zstdLevel))
		if o.zstdWindowLog > 0 {
			params.SetNum("101", float64(o.zstdWindowLog))
		}
		if o.zstdChecksum {
			params.SetNum("201", 1)
		}
		out.Set("params", params.V())
	}
	if o.hasPledged {
		out.SetNum("pledgedSrcSize", float64(o.pledged))
	}
	return out
}

// parseOptions validates an options bag exactly the way Node's Zlib/ZlibBase
// constructors do.
func parseOptions(kind string, opts js.Value) (*zlibOpts, error) {
	o := defaultOpts(kind)
	if !opts.Valid() || opts.IsUndefined() || opts.IsNull() {
		return o, nil
	}
	if !opts.IsObject() || opts.IsFunction() {
		return o, nil
	}
	obj := opts.Object()

	// ZlibBase options. Node treats undefined and NaN as "keep the default".
	if v := obj.Get("chunkSize"); v.Valid() && !v.IsUndefined() {
		if !v.IsNumber() {
			return nil, invalidType("options.chunkSize", "of type number", v)
		}
		f := v.Float()
		if math.IsInf(f, 0) {
			return nil, outOfRange("options.chunkSize", "a finite number", v)
		}
		if !math.IsNaN(f) {
			n := v.Int()
			if n < zMinChunk {
				return nil, outOfRange("options.chunkSize", ">= 64", v)
			}
			if int64(n) > zKMaxLength {
				return nil, outOfRange("options.chunkSize", ">= 64 and <= 4294967296", v)
			}
			o.chunkSize = n
		}
	}
	flushMax := int64(zBlock)
	switch kind {
	case "brotli", "brotliDecompress":
		flushMax = 3 // BROTLI_OPERATION_EMIT_METADATA
	case "zstd", "zstdDecompress":
		flushMax = 2 // ZSTD_e_end
	}
	if ok, err := rangeValue(obj.Get("flush"), "options.flush", zNoFlush, flushMax); err != nil {
		return nil, err
	} else if ok {
		o.flush = obj.Get("flush").Int()
	}
	if ok, err := rangeValue(obj.Get("finishFlush"), "options.finishFlush", zNoFlush, flushMax); err != nil {
		return nil, err
	} else if ok {
		o.finishFlush = obj.Get("finishFlush").Int()
	}
	if ok, err := rangeValue(obj.Get("maxOutputLength"), "options.maxOutputLength", 1, zKMaxLength); err != nil {
		return nil, err
	} else if ok {
		o.maxOutput = int64(obj.Get("maxOutputLength").Int())
	}
	if v := obj.Get("rejectGarbageAfterEnd"); v.Valid() && !v.IsUndefined() {
		if !v.IsBool() {
			return nil, invalidType("options.rejectGarbageAfterEnd", "of type boolean", v)
		}
		o.rejectGarbage = v.Bool()
	}
	if v := obj.Get("info"); v.Valid() && !v.IsUndefined() {
		if !v.IsBool() {
			return nil, invalidType("options.info", "of type boolean", v)
		}
		o.info = v.Bool()
	}

	// Zlib options.
	if ok, err := rangeValue(obj.Get("level"), "options.level", zMinLevel, zMaxLevel); err != nil {
		return nil, err
	} else if ok {
		o.level = obj.Get("level").Int()
	}
	if v := obj.Get("windowBits"); v.Valid() && !v.IsUndefined() {
		decompress := kind == "gunzip" || kind == "inflate" || kind == "unzip"
		if decompress && (v.IsNull() || (v.IsNumber() && v.Float() == 0)) {
			o.windowBits = 0
		} else {
			min := zMinWindowBits
			if kind == "gzip" {
				min = zMinWindowBits + 1
			}
			if ok, err := rangeValue(v, "options.windowBits", int64(min), zMaxWindowBits); err != nil {
				return nil, err
			} else if ok {
				o.windowBits = v.Int()
			}
		}
	}
	if ok, err := rangeValue(obj.Get("memLevel"), "options.memLevel", zMinMemLevel, zMaxMemLevel); err != nil {
		return nil, err
	} else if ok {
		o.memLevel = obj.Get("memLevel").Int()
	}
	if ok, err := rangeValue(obj.Get("strategy"), "options.strategy", zDefaultStrategy, zFixed); err != nil {
		return nil, err
	} else if ok {
		o.strategy = obj.Get("strategy").Int()
	}
	if err := parseParams(obj.Get("params"), kind, o); err != nil {
		return nil, err
	}
	if kind == "zstd" {
		if v := obj.Get("pledgedSrcSize"); v.Valid() && !v.IsUndefined() {
			if !v.IsNumber() {
				return nil, invalidType("options.pledgedSrcSize", "of type number", v)
			}
			f := v.Float()
			// Node validates pledgedSrcSize as a non-negative safe integer.
			if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 || f != math.Trunc(f) || f > 9007199254740991 {
				return nil, outOfRange("options.pledgedSrcSize", ">= 0 and <= Number.MAX_SAFE_INTEGER", v)
			}
			o.pledged = int64(f)
			o.hasPledged = true
		}
	}
	if v := obj.Get("dictionary"); v.Valid() && !v.IsUndefined() {
		if data, ok := dictionaryBytes(v); ok {
			o.dictionary = data
		} else {
			return nil, invalidType("options.dictionary", "an instance of Buffer, TypedArray, DataView, or ArrayBuffer", v)
		}
	}
	return o, nil
}

// dictionaryBytes accepts every shape Node's dictionary option accepts:
// Buffer/TypedArray/DataView directly and ArrayBuffer through Buffer.from.
func dictionaryBytes(v js.Value) ([]byte, bool) {
	if data, ok := v.Bytes(); ok {
		return data, true
	}
	if !v.IsObject() || v.IsArray() || v.IsFunction() {
		return nil, false
	}
	c := v.Ctx()
	if c == nil {
		return nil, false
	}
	ctor := c.Global().Get("Buffer")
	if !ctor.Valid() || !ctor.IsFunction() {
		return nil, false
	}
	from := ctor.Object().Get("from")
	if !from.Valid() || !from.IsFunction() {
		return nil, false
	}
	buf, err := from.Object().Call(ctor.Object(), v)
	if err != nil {
		return nil, false
	}
	return buf.Bytes()
}

// finiteNumber implements Node's validateFiniteNumber for option properties.
func finiteNumber(v js.Value, name string) error {
	if !v.IsNumber() {
		return invalidType(name, "of type number", v)
	}
	if math.IsInf(v.Float(), 0) || math.IsNaN(v.Float()) {
		return outOfRange(name, "a finite number", v)
	}
	return nil
}

// checkRanges implements Node's checkRangesOrGetDefault validation half.
func checkRanges(v js.Value, name string, lower, upper int64) error {
	_, err := rangeValue(v, name, lower, upper)
	return err
}

// rangeValue is Node's checkRangesOrGetDefault decision: undefined and NaN mean
// "keep the caller's default", non-numbers are a type error, infinities are a
// range error, and anything outside [lower, upper] is a range error too.
func rangeValue(v js.Value, name string, lower, upper int64) (bool, error) {
	if !v.Valid() || v.IsUndefined() {
		return false, nil
	}
	if !v.IsNumber() {
		return false, invalidType(name, "of type number", v)
	}
	f := v.Float()
	if math.IsNaN(f) {
		return false, nil
	}
	if math.IsInf(f, 0) {
		return false, outOfRange(name, "a finite number", v)
	}
	if f < float64(lower) || f > float64(upper) {
		return false, outOfRange(name, fmt.Sprintf(">= %d and <= %d", lower, upper), v)
	}
	return true, nil
}

func outOfRange(name, bound string, v js.Value) error {
	return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
		fmt.Sprintf("The value of \"%s\" is out of range. It must be %s. Received %s",
			name, bound, inspectValue(v)))
}

// invalidType renders Node's ERR_INVALID_ARG_TYPE message: option paths are
// called "property" rather than "argument".
func invalidType(name, expected string, v js.Value) error {
	word := "argument"
	if strings.Contains(name, ".") {
		word = "property"
	}
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		fmt.Sprintf("The \"%s\" %s must be %s.%s", name, word, expected, received(v)))
}

// received implements common.invalidArgTypeHelper's "Received ..." suffix, which
// is what Node appends to ERR_INVALID_ARG_TYPE messages.
func received(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return " Received undefined"
	case v.IsNull():
		return " Received null"
	case v.IsFunction():
		name := ""
		if n := v.Object().Get("name"); n.IsString() {
			name = n.Str()
		}
		return " Received function " + name
	case v.IsObject():
		if name := ctorName(v); name != "" {
			return " Received an instance of " + name
		}
		return " Received [object Object]"
	case v.IsBool():
		return " Received type boolean (" + strconv.FormatBool(v.Bool()) + ")"
	case v.IsNumber():
		return " Received type number (" + inspectNumber(v.Float()) + ")"
	case v.IsString():
		return " Received type string (" + inspectString(v.Str()) + ")"
	}
	return " Received " + inspectValue(v)
}

func ctorName(v js.Value) string {
	if !v.IsObject() {
		return ""
	}
	c := v.Object().Get("constructor")
	if c.Valid() && c.IsFunction() {
		if n := c.Object().Get("name"); n.IsString() {
			return n.Str()
		}
	}
	return ""
}

func inspectNumber(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case f == math.Trunc(f) && math.Abs(f) < 1e18:
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// inspectString quotes a string the way util.inspect does, including Node's
// 28-character truncation in argument error messages.
func inspectString(s string) string {
	quoted := "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(s) + "'"
	if len(quoted) > 28 {
		return quoted[:25] + "..."
	}
	return quoted
}

func inspectValue(v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		return strconv.FormatBool(v.Bool())
	case v.IsNumber():
		return inspectNumber(v.Float())
	case v.IsString():
		return inspectString(v.Str())
	}
	return v.Str()
}

func collectChunks(v js.Value) ([]byte, error) {
	if !v.Valid() || !v.IsArray() {
		return nil, js.Throw("TypeError", "The \"chunks\" argument must be an array")
	}
	n := v.Object().Length()
	var out []byte
	for i := 0; i < n; i++ {
		b, ok := v.Object().At(i).Bytes()
		if !ok {
			return nil, js.Throw("TypeError", "The \"chunk\" argument must be an instance of Buffer, TypedArray, or DataView")
		}
		out = append(out, b...)
	}
	return out, nil
}

func (m *module) deliver(cb js.Value, err error, out []byte) {
	var values []js.Value
	if err != nil {
		values = []js.Value{m.errorValue(err)}
	} else {
		values = []js.Value{m.e.JS.Null(), m.bytesValue(out)}
	}
	m.e.NextTick(cb, values...)
}

func (m *module) deliverValue(cb js.Value, err error, v js.Value) {
	var values []js.Value
	if err != nil {
		values = []js.Value{m.errorValue(err)}
	} else {
		values = []js.Value{m.e.JS.Null(), v}
	}
	m.e.NextTick(cb, values...)
}

func (m *module) errorValue(err error) js.Value {
	msg := err.Error()
	if je, ok := err.(*js.Error); ok {
		if je.Message != "" {
			msg = je.Message
		}
	}
	obj := m.e.JS.Error(msg)
	if je, ok := err.(*js.Error); ok {
		if je.Name != "" {
			obj.SetStr("name", je.Name)
		}
		if je.Code != "" {
			obj.SetStr("code", je.Code)
		}
		if je.Errno != nil {
			// Codec errors (Zstd's ZSTD_error_*) carry their own numeric errno.
			obj.SetNum("errno", float64(*je.Errno))
		} else if n, ok := zlibErrno(je.Code); ok {
			obj.SetNum("errno", n)
		}
	}
	return obj.V()
}

func constants(e *env.Env) js.Object {
	o := e.JS.Obj()
	values := map[string]int{
		// Flush levels.
		"Z_NO_FLUSH": 0, "Z_PARTIAL_FLUSH": 1, "Z_SYNC_FLUSH": 2,
		"Z_FULL_FLUSH": 3, "Z_FINISH": 4, "Z_BLOCK": 5, "Z_TREES": 6,
		// Return codes.
		"Z_OK": 0, "Z_STREAM_END": 1, "Z_NEED_DICT": 2,
		"Z_ERRNO": -1, "Z_STREAM_ERROR": -2, "Z_DATA_ERROR": -3,
		"Z_MEM_ERROR": -4, "Z_BUF_ERROR": -5, "Z_VERSION_ERROR": -6,
		// Compression levels.
		"Z_NO_COMPRESSION": 0, "Z_BEST_SPEED": 1, "Z_BEST_COMPRESSION": 9,
		"Z_DEFAULT_COMPRESSION": -1,
		// Strategies.
		"Z_FILTERED": 1, "Z_HUFFMAN_ONLY": 2, "Z_RLE": 3, "Z_FIXED": 4,
		"Z_DEFAULT_STRATEGY": 0,
		// Data types.
		"Z_BINARY": 0, "Z_TEXT": 1, "Z_ASCII": 1, "Z_UNKNOWN": 2,
		"Z_DEFLATED": 8, "Z_NULL": 0,
		// Tuning bounds.
		"Z_MIN_CHUNK": 64, "Z_DEFAULT_CHUNK": 16384,
		"Z_MIN_WINDOWBITS": 8, "Z_MAX_WINDOWBITS": 15, "Z_DEFAULT_WINDOWBITS": 15,
		"Z_MIN_LEVEL": -1, "Z_MAX_LEVEL": 9, "Z_DEFAULT_LEVEL": -1,
		"Z_MIN_MEMLEVEL": 1, "Z_MAX_MEMLEVEL": 9, "Z_DEFAULT_MEMLEVEL": 8,
		// Node's stream modes.
		"DEFLATE": 1, "INFLATE": 2, "GZIP": 3, "GUNZIP": 4,
		"DEFLATERAW": 5, "INFLATERAW": 6, "UNZIP": 7,
		"BROTLI_DECODE": 8, "BROTLI_ENCODE": 9,
		"ZSTD_COMPRESS": 10, "ZSTD_DECOMPRESS": 11,
		// Brotli parameters, modes and operations.
		"BROTLI_PARAM_MODE": 0, "BROTLI_PARAM_QUALITY": 1, "BROTLI_PARAM_LGWIN": 2,
		"BROTLI_PARAM_LGBLOCK": 3, "BROTLI_PARAM_DISABLE_LITERAL_CONTEXT_MODELING": 4,
		"BROTLI_PARAM_SIZE_HINT": 5, "BROTLI_PARAM_LARGE_WINDOW": 6,
		"BROTLI_PARAM_NPOSTFIX": 7, "BROTLI_PARAM_NDIRECT": 8,
		"BROTLI_MODE_GENERIC": 0, "BROTLI_MODE_TEXT": 1, "BROTLI_MODE_FONT": 2,
		"BROTLI_OPERATION_PROCESS": 0, "BROTLI_OPERATION_FLUSH": 1,
		"BROTLI_OPERATION_FINISH": 2, "BROTLI_OPERATION_EMIT_METADATA": 3,
		"BROTLI_DEFAULT_QUALITY": 11, "BROTLI_DEFAULT_WINDOW": 22,
		"BROTLI_DEFAULT_MODE": 0, "BROTLI_MIN_WINDOW_BITS": 10,
		"BROTLI_MAX_WINDOW_BITS": 24, "BROTLI_MIN_INPUT_BLOCK_BITS": 16,
		"BROTLI_MAX_INPUT_BLOCK_BITS": 24, "BROTLI_MIN_QUALITY": 0,
		"BROTLI_MAX_QUALITY": 11,
		// Brotli decoder parameters (brotli/decode.h).
		"BROTLI_DECODER_PARAM_DISABLE_RING_BUFFER_REALLOCATION": 0,
		"BROTLI_DECODER_PARAM_LARGE_WINDOW":                     1,
		"BROTLI_DECODER_RESULT_ERROR":                           0, "BROTLI_DECODER_RESULT_SUCCESS": 1,
		"BROTLI_DECODER_RESULT_NEEDS_MORE_INPUT":  2,
		"BROTLI_DECODER_RESULT_NEEDS_MORE_OUTPUT": 3,
		// Zstd end directives.
		"ZSTD_e_continue": 0, "ZSTD_e_flush": 1, "ZSTD_e_end": 2,
		// Zstd compression and decompression parameters (zstd.h).
		"ZSTD_c_compressionLevel": 100, "ZSTD_c_windowLog": 101,
		"ZSTD_c_hashLog": 102, "ZSTD_c_chainLog": 103, "ZSTD_c_searchLog": 104,
		"ZSTD_c_minMatch": 105, "ZSTD_c_targetLength": 106, "ZSTD_c_strategy": 107,
		"ZSTD_c_enableLongDistanceMatching": 160, "ZSTD_c_ldmHashLog": 161,
		"ZSTD_c_ldmMinMatch": 162, "ZSTD_c_ldmBucketSizeLog": 163,
		"ZSTD_c_ldmHashRateLog": 164, "ZSTD_c_contentSizeFlag": 200,
		"ZSTD_c_checksumFlag": 201, "ZSTD_c_dictIDFlag": 202,
		"ZSTD_c_nbWorkers": 400, "ZSTD_c_jobSize": 401, "ZSTD_c_overlapLog": 402,
		"ZSTD_d_windowLogMax": 100,
		// Zstd error codes (zstd.h ZSTD_ErrorCode), which Node publishes as
		// constants and attaches to errors as both code and errno.
		"ZSTD_error_no_error": 0, "ZSTD_error_GENERIC": 1,
		"ZSTD_error_prefix_unknown": 10, "ZSTD_error_version_unsupported": 12,
		"ZSTD_error_frameParameter_unsupported":    14,
		"ZSTD_error_frameParameter_windowTooLarge": 16,
		"ZSTD_error_corruption_detected":           20, "ZSTD_error_checksum_wrong": 22,
		"ZSTD_error_literals_headerWrong": 24, "ZSTD_error_dictionary_corrupted": 30,
		"ZSTD_error_dictionary_wrong": 32, "ZSTD_error_dictionaryCreation_failed": 34,
		"ZSTD_error_parameter_unsupported":             40,
		"ZSTD_error_parameter_combination_unsupported": 41,
		"ZSTD_error_parameter_outOfBound":              42, "ZSTD_error_tableLog_tooLarge": 44,
		"ZSTD_error_maxSymbolValue_tooLarge": 46, "ZSTD_error_maxSymbolValue_tooSmall": 48,
		"ZSTD_error_stabilityCondition_notRespected": 50, "ZSTD_error_stage_wrong": 60,
		"ZSTD_error_init_missing": 62, "ZSTD_error_memory_allocation": 64,
		"ZSTD_error_workSpace_tooSmall": 66, "ZSTD_error_dstSize_tooSmall": 70,
		"ZSTD_error_srcSize_wrong": 72, "ZSTD_error_dstBuffer_null": 74,
		"ZSTD_error_noForwardProgress_destFull":   80,
		"ZSTD_error_noForwardProgress_inputEmpty": 82,
	}
	for k, v := range values {
		o.SetNum(k, float64(v))
	}
	// Z_MAX_CHUNK is Infinity in Node.
	o.SetNum("Z_MAX_CHUNK", floatInf)
	return o
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}
