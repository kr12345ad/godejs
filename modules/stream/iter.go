package stream

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

// The iterable streams API ("stream/iter") is a separate surface from the
// classic stream module: sources are sync/async iterables of Uint8Array
// batches, and everything is pull-driven. The protocol logic is written in
// JavaScript like the rest of the stream module; Go only installs it and
// hands it the classic stream constructor it needs for interop.
//
// Node behind a flag (--experimental-stream-iter) still exposes the same
// export names, so require('stream/iter') must succeed unconditionally here;
// the experimental warning is intentionally not printed because it would
// pollute stderr that the test suite inspects.
func init() {
	env.Register("stream/iter", InstallIter)
	env.Register("stream/consumers", InstallConsumers)
	env.Register("stream/web", InstallWeb)
}

// InstallIter builds the stream/iter exports object.
func InstallIter(e *env.Env) js.Object {
	return installScript(e, "stream/iter", iterPrelude)
}

// InstallConsumers builds the stream/consumers exports object.
func InstallConsumers(e *env.Env) js.Object {
	return installScript(e, "stream/consumers", consumersPrelude)
}

// InstallWeb builds the stream/web exports object.
func InstallWeb(e *env.Env) js.Object {
	return installScript(e, "stream/web", webPrelude)
}

// installScript evaluates a module bootstrap and returns its exports. Each
// prelude receives (exports, deps) so it can borrow the classic stream
// constructors, Buffer and the pending event emitter without re-implementing
// them or creating an import cycle.
func installScript(e *env.Env, name, src string) js.Object {
	exports := e.JS.Obj()
	deps := e.JS.Obj()
	if env.Registered("stream") {
		deps.Set("stream", e.Exports("stream").V())
	}
	if env.Registered("buffer") {
		buf := e.Exports("buffer")
		deps.Set("buffer", buf.V())
		if ctor := buf.Get("Buffer"); ctor.Valid() && ctor.IsFunction() {
			deps.Set("Buffer", ctor)
		}
	}
	if env.Registered("events") {
		if ctor := e.Exports("events").Get("EventEmitter"); ctor.Valid() {
			deps.Set("EventEmitter", ctor)
		}
	}
	fn, err := e.JS.Eval(src, "[gode:"+name+"]")
	if err != nil {
		panic("godejs " + name + ": " + err.Error())
	}
	if _, err := fn.Object().Call(js.Object{}, exports.V(), deps.V()); err != nil {
		panic("godejs " + name + ": " + err.Error())
	}
	return exports
}

// IterImplemented lists the tracked Node 24 API entries this file provides.
// The package-wide Implemented() lives in stream.go, so this file exposes its
// own list under a distinct name to avoid a duplicate declaration.
func IterImplemented() []string {
	return []string{
		"stream/iter.Stream",
		"stream/iter.from",
		"stream/iter.fromSync",
		"stream/iter.push",
		"stream/iter.duplex",
		"stream/iter.pull",
		"stream/iter.pullSync",
		"stream/iter.pipeTo",
		"stream/iter.pipeToSync",
		"stream/iter.bytes",
		"stream/iter.text",
		"stream/iter.arrayBuffer",
		"stream/iter.array",
		"stream/iter.bytesSync",
		"stream/iter.textSync",
		"stream/iter.arrayBufferSync",
		"stream/iter.arraySync",
		"stream/iter.merge",
		"stream/iter.broadcast",
		"stream/iter.Broadcast",
		"stream/iter.share",
		"stream/iter.shareSync",
		"stream/iter.Share",
		"stream/iter.SyncShare",
		"stream/iter.tap",
		"stream/iter.tapSync",
		"stream/iter.ondrain",
		"stream/iter.fromReadable",
		"stream/iter.fromWritable",
		"stream/iter.toReadable",
		"stream/iter.toReadableSync",
		"stream/iter.toWritable",
		"stream/consumers.arrayBuffer",
		"stream/consumers.blob",
		"stream/consumers.buffer",
		"stream/consumers.bytes",
		"stream/consumers.json",
		"stream/consumers.text",
		"stream/web.ReadableStream",
		"stream/web.WritableStream",
		"stream/web.TransformStream",
		"stream/web.ByteLengthQueuingStrategy",
		"stream/web.CountQueuingStrategy",
	}
}

const iterPrelude = `(function (exports, deps) {
  'use strict';
  var streamModule = deps.stream || {};
  var Readable = streamModule.Readable;
  var Writable = streamModule.Writable;
  var BufferCtor = deps.Buffer || (typeof Buffer !== 'undefined' ? Buffer : null);

  var FROM_BATCH_SIZE = 128;

  var toStreamable = Symbol.for('Stream.toStreamable');
  var toAsyncStreamable = Symbol.for('Stream.toAsyncStreamable');
  var broadcastProtocol = Symbol.for('Stream.broadcastProtocol');
  var shareProtocol = Symbol.for('Stream.shareProtocol');
  var shareSyncProtocol = Symbol.for('Stream.shareSyncProtocol');
  var drainableProtocol = Symbol.for('Stream.drainableProtocol');
  var kValidatedSource = Symbol('kValidatedSource');
  var kValidatedTransform = Symbol('kValidatedTransform');

  // ------------------------------------------------------------- validation --
  function typeName(v) {
    if (v === null) return 'null';
    var t = typeof v;
    if (t === 'undefined') return 'undefined';
    if (t === 'number') return 'type number (' + v + ')';
    if (t === 'string') return 'type string (' + v + ')';
    if (t === 'boolean') return 'type boolean (' + v + ')';
    if (t === 'bigint') return 'type bigint (' + v + ')';
    if (t === 'symbol') return 'type symbol';
    if (t === 'function') return 'function ' + (v.name || '(anonymous)');
    if (Array.isArray(v)) return 'an instance of Array';
    if (v.constructor && v.constructor.name) return 'an instance of ' + v.constructor.name;
    return 'an instance of Object';
  }

  function invalidArgType(name, expected, actual) {
    var err = new TypeError('The "' + name + '" argument must be ' + expected +
      '. Received ' + typeName(actual));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }

  function invalidArgValue(name, value, reason) {
    var err = new TypeError("The argument '" + name + "' " + reason +
      '. Received ' + typeName(value));
    err.code = 'ERR_INVALID_ARG_VALUE';
    return err;
  }

  function outOfRange(name, bound, actual) {
    var err = new RangeError('The value of "' + name + '" is out of range. It must be ' +
      bound + '. Received ' + actual);
    err.code = 'ERR_OUT_OF_RANGE';
    return err;
  }

  function validateAbortSignal(signal, name) {
    if (signal === undefined) return;
    if (signal === null || typeof signal !== 'object' || typeof signal.aborted !== 'boolean') {
      throw invalidArgType(name || 'options.signal', 'an instance of AbortSignal', signal);
    }
  }

  function validateFunction(fn, name) {
    if (typeof fn !== 'function') {
      throw invalidArgType(name, 'of type function', fn);
    }
  }

  function throwIfAborted(signal) {
    if (signal && signal.aborted) throw signal.reason;
  }

  // ------------------------------------------------------------ type guards --
  function isUint8Array(v) { return v instanceof Uint8Array; }
  function isAnyArrayBuffer(v) {
    return v instanceof ArrayBuffer ||
      (typeof SharedArrayBuffer !== 'undefined' && v instanceof SharedArrayBuffer);
  }
  function isView(v) { return typeof ArrayBuffer !== 'undefined' && ArrayBuffer.isView(v); }
  function isPromise(v) {
    return v !== null && v !== undefined &&
      (typeof v === 'object' || typeof v === 'function') && typeof v.then === 'function';
  }
  function isPrimitiveChunk(v) {
    return typeof v === 'string' || isAnyArrayBuffer(v) || isView(v);
  }
  function isSyncIterable(v) {
    return v !== null && v !== undefined && typeof v !== 'string' &&
      typeof v[Symbol.iterator] === 'function';
  }
  function isAsyncIterable(v) {
    return v !== null && v !== undefined &&
      typeof v[Symbol.asyncIterator] === 'function';
  }
  function hasProtocol(v, sym) {
    return v !== null && v !== undefined && typeof v[sym] === 'function';
  }
  function isUint8ArrayBatch(v) {
    if (!Array.isArray(v)) return false;
    var len = v.length;
    if (len === 0) return true;
    if (!isUint8Array(v[0])) return false;
    if (len === 1) return true;
    if (!isUint8Array(v[len - 1])) return false;
    for (var i = 1; i < len - 1; i++) {
      if (!isUint8Array(v[i])) return false;
    }
    return true;
  }
  function isTransform(v) {
    return typeof v === 'function' ||
      (v !== null && typeof v === 'object' && typeof v.transform === 'function');
  }
  function isPullOptions(v) {
    return v !== null && typeof v === 'object' && !Array.isArray(v) &&
      typeof v[Symbol.asyncIterator] !== 'function' &&
      typeof v[Symbol.iterator] !== 'function' &&
      typeof v.transform !== 'function' &&
      !isPrimitiveChunk(v);
  }

  function iterableAsync(factory) {
    var obj = {};
    obj[Symbol.asyncIterator] = function () { return factory(); };
    return obj;
  }
  function iterableSync(factory) {
    var obj = {};
    obj[Symbol.iterator] = function () { return factory(); };
    return obj;
  }

  // ----------------------------------------------------------- conversions --
  // Node's iter API encodes strings with TextEncoder, so string chunks come
  // out as plain Uint8Array (not Buffer): tests compare the exact prototype.
  function encodeString(s) {
    if (typeof TextEncoder === 'function') return new TextEncoder().encode(s);
    if (BufferCtor) return BufferCtor.from(s, 'utf8');
    var out = new Uint8Array(s.length);
    for (var i = 0; i < s.length; i++) out[i] = s.charCodeAt(i) & 0xff;
    return out;
  }

  function toUint8Array(chunk) {
    if (typeof chunk === 'string') return encodeString(chunk);
    if (isAnyArrayBuffer(chunk)) return new Uint8Array(chunk);
    if (isUint8Array(chunk)) return chunk;
    if (isView(chunk)) return new Uint8Array(chunk.buffer, chunk.byteOffset, chunk.byteLength);
    return chunk;
  }

  function concatBytes(chunks) {
    var total = 0;
    var i;
    for (i = 0; i < chunks.length; i++) total += chunks[i].byteLength;
    var out = new Uint8Array(total);
    var offset = 0;
    for (i = 0; i < chunks.length; i++) {
      out.set(chunks[i], offset);
      offset += chunks[i].byteLength;
    }
    return out;
  }

  function toArrayBuffer(data) {
    if (data.byteOffset === 0 && data.byteLength === data.buffer.byteLength) {
      return data.buffer;
    }
    return data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength);
  }

  // Some TextDecoder implementations reject the fatal option; probe once so
  // text consumers can still decode instead of throwing at install time.
  var supportsFatalDecoder = (function () {
    try { new TextDecoder('utf-8', { fatal: true }); return true; } catch (_) { return false; }
  })();

  function decodeBytes(chunks, encoding, fatal) {
    var bytes = concatBytes(chunks);
    var enc = encoding || 'utf-8';
    if (typeof TextDecoder === 'function') {
      return new TextDecoder(enc, { fatal: !!fatal }).decode(bytes);
    }
    // Without the web-platform globals, Buffer implements the same encodings.
    if (BufferCtor) return BufferCtor.from(bytes).toString(enc);
    var out = '';
    for (var i = 0; i < bytes.length; i++) out += String.fromCharCode(bytes[i]);
    return out;
  }

  // ------------------------------------------------------- normalization -----
  function* normalizeSyncValue(value) {
    if (isPrimitiveChunk(value)) { yield toUint8Array(value); return; }
    if (hasProtocol(value, toStreamable)) {
      yield* normalizeSyncValue(value[toStreamable]());
      return;
    }
    if (Array.isArray(value)) {
      for (var i = 0; i < value.length; i++) yield* normalizeSyncValue(value[i]);
      return;
    }
    if (isSyncIterable(value)) {
      for (var item of value) yield* normalizeSyncValue(item);
      return;
    }
    throw invalidArgType('value',
      'of type string or an instance of ArrayBuffer, ArrayBufferView, Iterable',
      value);
  }

  function* normalizeSyncSource(source) {
    var batch = [];
    for (var value of source) {
      if (isUint8ArrayBatch(value)) {
        if (batch.length > 0) { yield batch; batch = []; }
        for (var i = 0; i < value.length; i += FROM_BATCH_SIZE) {
          yield value.slice(i, i + FROM_BATCH_SIZE);
        }
        continue;
      }
      if (isUint8Array(value)) {
        batch.push(value);
        if (batch.length === FROM_BATCH_SIZE) { yield batch; batch = []; }
        continue;
      }
      if (batch.length > 0) { yield batch; batch = []; }
      var sub = [];
      for (var chunk of normalizeSyncValue(value)) {
        sub.push(chunk);
        if (sub.length === FROM_BATCH_SIZE) { yield sub; sub = []; }
      }
      if (sub.length > 0) yield sub;
    }
    if (batch.length > 0) yield batch;
  }

  async function* normalizeAsyncValue(value) {
    if (isPromise(value)) {
      yield* normalizeAsyncValue(await value);
      return;
    }
    if (isPrimitiveChunk(value)) { yield toUint8Array(value); return; }
    if (hasProtocol(value, toAsyncStreamable)) {
      var result = value[toAsyncStreamable]();
      yield* normalizeAsyncValue(isPromise(result) ? await result : result);
      return;
    }
    if (hasProtocol(value, toStreamable)) {
      yield* normalizeAsyncValue(value[toStreamable]());
      return;
    }
    if (Array.isArray(value)) {
      for (var i = 0; i < value.length; i++) yield* normalizeAsyncValue(value[i]);
      return;
    }
    if (isAsyncIterable(value)) {
      for await (var item of value) yield* normalizeAsyncValue(item);
      return;
    }
    if (isSyncIterable(value)) {
      for (var sync of value) yield* normalizeAsyncValue(sync);
      return;
    }
    throw invalidArgType('value',
      'of type string or an instance of ArrayBuffer, ArrayBufferView, Iterable, AsyncIterable',
      value);
  }

  async function* normalizeAsyncSource(source) {
    if (isAsyncIterable(source)) {
      for await (var value of source) {
        if (isUint8ArrayBatch(value)) {
          if (value.length > 0) yield value;
          continue;
        }
        if (isUint8Array(value)) { yield [value]; continue; }
        var batch = [];
        for await (var chunk of normalizeAsyncValue(value)) {
          batch.push(chunk);
          if (batch.length === FROM_BATCH_SIZE) { yield batch; batch = []; }
        }
        if (batch.length > 0) yield batch;
      }
      return;
    }
    var syncBatch = [];
    for (var value2 of source) {
      if (isUint8ArrayBatch(value2)) {
        if (syncBatch.length > 0) { yield syncBatch; syncBatch = []; }
        for (var i = 0; i < value2.length; i += FROM_BATCH_SIZE) {
          yield value2.slice(i, i + FROM_BATCH_SIZE);
        }
        continue;
      }
      if (isUint8Array(value2)) {
        syncBatch.push(value2);
        if (syncBatch.length === FROM_BATCH_SIZE) { yield syncBatch; syncBatch = []; }
        continue;
      }
      if (syncBatch.length > 0) { yield syncBatch; syncBatch = []; }
      var sub = [];
      for await (var chunk2 of normalizeAsyncValue(value2)) {
        sub.push(chunk2);
        if (sub.length === FROM_BATCH_SIZE) { yield sub; sub = []; }
      }
      if (sub.length > 0) yield sub;
    }
    if (syncBatch.length > 0) yield syncBatch;
  }

  // -------------------------------------------------------- from / fromSync --
  function fromSync(input) {
    if (input === null || input === undefined) {
      throw invalidArgType('input', 'a non-null value', input);
    }
    if (isPrimitiveChunk(input)) {
      var chunk = toUint8Array(input);
      return iterableSync(function* () { yield [chunk]; });
    }
    if (Array.isArray(input) && input.length > 0 && input.every(isUint8Array)) {
      var whole = input.slice();
      return iterableSync(function* () {
        for (var i = 0; i < whole.length; i += FROM_BATCH_SIZE) {
          yield whole.slice(i, i + FROM_BATCH_SIZE);
        }
      });
    }
    if (hasProtocol(input, toStreamable)) return fromSync(input[toStreamable]());
    if (!isSyncIterable(input)) {
      throw invalidArgType('input',
        'of type string or an instance of ArrayBuffer, ArrayBufferView, Iterable',
        input);
    }
    return iterableSync(function () { return normalizeSyncSource(input); });
  }

  function from(input) {
    if (input === null || input === undefined) {
      throw invalidArgType('input', 'a non-null value', input);
    }
    if (input[kValidatedSource]) return input;
    if (isPrimitiveChunk(input)) {
      var chunk = toUint8Array(input);
      return iterableAsync(async function* () { yield [chunk]; });
    }
    if (Array.isArray(input) && input.length > 0 && input.every(isUint8Array)) {
      var whole = input.slice();
      return iterableAsync(async function* () {
        for (var i = 0; i < whole.length; i += FROM_BATCH_SIZE) {
          yield whole.slice(i, i + FROM_BATCH_SIZE);
        }
      });
    }
    if (hasProtocol(input, toAsyncStreamable)) {
      var result = input[toAsyncStreamable]();
      if (isPromise(result)) {
        return iterableAsync(async function* () {
          yield* from(await result);
        });
      }
      return from(result);
    }
    if (hasProtocol(input, toStreamable)) return from(input[toStreamable]());
    if (!isSyncIterable(input) && !isAsyncIterable(input)) {
      throw invalidArgType('input',
        'of type string or an instance of ArrayBuffer, ArrayBufferView, Iterable, AsyncIterable',
        input);
    }
    return iterableAsync(function () { return normalizeAsyncSource(input); });
  }

  // -------------------------------------------------------------- consumers --
  function collectSync(source, limit) {
    var normalized = fromSync(source);
    var chunks = [];
    var total = 0;
    for (var batch of normalized) {
      for (var i = 0; i < batch.length; i++) {
        var chunk = batch[i];
        if (limit !== undefined) {
          total += chunk.byteLength;
          if (total > limit) throw outOfRange('totalBytes', '<= ' + limit, total);
        }
        chunks.push(chunk);
      }
    }
    return chunks;
  }

  async function collectAsync(source, signal, limit) {
    throwIfAborted(signal);
    var normalized = from(source);
    var chunks = [];
    var total = 0;
    for await (var batch of normalized) {
      throwIfAborted(signal);
      for (var i = 0; i < batch.length; i++) {
        var chunk = batch[i];
        if (limit !== undefined) {
          total += chunk.byteLength;
          if (total > limit) throw outOfRange('totalBytes', '<= ' + limit, total);
        }
        chunks.push(chunk);
      }
    }
    return chunks;
  }

  function consumerOptions(options) {
    if (options === undefined || options === null) return {};
    if (typeof options !== 'object') throw invalidArgType('options', 'of type object', options);
    if (options.limit !== undefined &&
        (typeof options.limit !== 'number' || options.limit < 0 || Math.floor(options.limit) !== options.limit)) {
      throw outOfRange('options.limit', '>= 0', options.limit);
    }
    if (options.encoding !== undefined) {
      if (typeof options.encoding !== 'string') {
        throw invalidArgType('options.encoding', 'of type string', options.encoding);
      }
      // Node validates the label by constructing a TextDecoder and maps any
      // rejection to ERR_INVALID_ARG_VALUE.RangeError.
      try {
        new TextDecoder(options.encoding);
      } catch (e) {
        var bad = new RangeError('The argument \'options.encoding\' must be a valid encoding. Received ' + options.encoding);
        bad.code = 'ERR_INVALID_ARG_VALUE';
        throw bad;
      }
    }
    if (options.signal !== undefined) validateAbortSignal(options.signal);
    return options;
  }

  // Node validates highWaterMark with validateInteger(..., 0), so negative or
  // fractional values are ERR_OUT_OF_RANGE from the iter layer, not the
  // classic Readable constructor.
  function highWaterMarkOption(options) {
    if (options.highWaterMark === undefined) return 64 * 1024;
    var hwm = options.highWaterMark;
    if (typeof hwm !== 'number' || !isFinite(hwm) || Math.floor(hwm) !== hwm || hwm < 0) {
      throw outOfRange('options.highWaterMark', '>= 0', hwm);
    }
    return hwm;
  }

  function validateBackpressure(policy) {
    if (policy === undefined) return 'strict';
    if (policy !== 'strict' && policy !== 'unbounded' && policy !== 'drop-newest' &&
        policy !== 'drop-oldest') {
      var err = new TypeError('The argument \'options.backpressure\' must be one of: ' +
        "'strict', 'unbounded', 'drop-newest', 'drop-oldest'. Received '" + policy + "'");
      err.code = 'ERR_INVALID_ARG_VALUE';
      throw err;
    }
    return policy;
  }

  function bytesSync(source, options) {
    options = consumerOptions(options);
    return concatBytes(collectSync(source, options.limit));
  }

  function textSync(source, options) {
    options = consumerOptions(options);
    return decodeBytes(collectSync(source, options.limit), options.encoding, true);
  }

  function arrayBufferSync(source, options) {
    options = consumerOptions(options);
    return toArrayBuffer(concatBytes(collectSync(source, options.limit)));
  }

  function arraySync(source, options) {
    options = consumerOptions(options);
    return collectSync(source, options.limit);
  }

  async function bytes(source, options) {
    options = consumerOptions(options);
    return concatBytes(await collectAsync(source, options.signal, options.limit));
  }

  async function text(source, options) {
    options = consumerOptions(options);
    return decodeBytes(await collectAsync(source, options.signal, options.limit),
      options.encoding, true);
  }

  async function arrayBuffer(source, options) {
    options = consumerOptions(options);
    return toArrayBuffer(concatBytes(await collectAsync(source, options.signal, options.limit)));
  }

  async function array(source, options) {
    options = consumerOptions(options);
    return collectAsync(source, options.signal, options.limit);
  }

  function tap(callback) {
    validateFunction(callback, 'callback');
    return async function (chunks, options) {
      await callback(chunks, options);
      return chunks;
    };
  }

  function tapSync(callback) {
    validateFunction(callback, 'callback');
    return function (chunks) {
      callback(chunks);
      return chunks;
    };
  }

  function ondrain(drainable) {
    if (drainable === null || drainable === undefined || typeof drainable !== 'object') {
      return null;
    }
    if (typeof drainable[drainableProtocol] !== 'function') return null;
    return drainable[drainableProtocol]();
  }

  function merge() {
    var args = Array.prototype.slice.call(arguments);
    var options = null;
    if (args.length > 0) {
      var last = args[args.length - 1];
      if (isPullOptions(last) && !isSyncIterable(last) && !isAsyncIterable(last) &&
          typeof last[toStreamable] !== 'function' && typeof last[toAsyncStreamable] !== 'function') {
        options = last;
        args.pop();
      }
    }
    if (options && options.signal !== undefined) validateAbortSignal(options.signal);
    var sources = args.map(function (source) { return from(source); });
    var signal = options ? options.signal : undefined;
    return iterableAsync(async function* () {
      throwIfAborted(signal);
      var iterators = sources.map(function (s) { return s[Symbol.asyncIterator](); });
      var pending = new Map();
      function pullOne(index) {
        return Promise.resolve(iterators[index].next()).then(function (result) {
          return { index: index, result: result };
        }, function (error) {
          return { index: index, error: error };
        });
      }
      iterators.forEach(function (_it, i) { pending.set(i, pullOne(i)); });
      while (pending.size > 0) {
        var settled = await Promise.race(Array.from(pending.values()));
        pending.delete(settled.index);
        if (settled.error !== undefined) throw settled.error;
        throwIfAborted(signal);
        if (!settled.result.done) {
          pending.set(settled.index, pullOne(settled.index));
          yield settled.result.value;
        }
      }
    });
  }

  // ------------------------------------------------------------------- push --
  function writeAfterEndError() {
    var err = new Error('write after end');
    err.code = 'ERR_STREAM_WRITE_AFTER_END';
    return err;
  }

  function push(options) {
    if (options !== undefined && options !== null && typeof options !== 'object') {
      throw invalidArgType('options', 'of type object', options);
    }
    var queue = [];
    var waiters = [];
    var ended = false;
    var failure = null;

    function wake() {
      while (waiters.length > 0) {
        if (failure) { waiters.shift().reject(failure); continue; }
        if (queue.length > 0) {
          waiters.shift().resolve({ value: [queue.shift()], done: false });
          continue;
        }
        if (ended) {
          waiters.shift().resolve({ value: undefined, done: true });
          continue;
        }
        break;
      }
    }

    var writer = {
      write: function (chunk) {
        if (failure) return Promise.reject(failure);
        if (ended) return Promise.reject(writeAfterEndError());
        queue.push(toUint8Array(chunk));
        wake();
        return Promise.resolve(true);
      },
      writev: function (chunks) {
        if (failure) return Promise.reject(failure);
        if (ended) return Promise.reject(writeAfterEndError());
        for (var i = 0; i < chunks.length; i++) queue.push(toUint8Array(chunks[i]));
        wake();
        return Promise.resolve(true);
      },
      end: function (chunk) {
        var pending = [];
        if (chunk !== undefined && chunk !== null) pending.push(writer.write(chunk));
        ended = true;
        wake();
        return Promise.all(pending).then(function () {});
      },
      fail: function (error) {
        if (failure || ended) return Promise.resolve();
        failure = error || new Error('writer failed');
        wake();
        return Promise.resolve();
      },
      writeSync: function (chunk) {
        if (failure) throw failure;
        if (ended) throw writeAfterEndError();
        queue.push(toUint8Array(chunk));
        wake();
        return true;
      },
      endSync: function () {
        ended = true;
        wake();
        return 0;
      }
    };
    Object.defineProperty(writer, 'desiredSize', {
      get: function () { return queue.length > 0 ? 0 : 16384; }
    });

    var readable = iterableAsync(function () {
      return {
        next: function () {
          if (failure) return Promise.reject(failure);
          if (queue.length > 0) return Promise.resolve({ value: [queue.shift()], done: false });
          if (ended) return Promise.resolve({ value: undefined, done: true });
          return new Promise(function (resolve, reject) {
            waiters.push({ resolve: resolve, reject: reject });
          });
        },
        return: function () {
          ended = true;
          wake();
          return Promise.resolve({ value: undefined, done: true });
        }
      };
    });

    return { writer: writer, readable: readable };
  }

  // ----------------------------------------------------------------- duplex --
  function createChannel(writer, readable) {
    var closeIterator = readable[Symbol.asyncIterator]();
    var closePromise;
    var channel = {
      get writer() { return writer; },
      get readable() { return readable; },
      close: function () {
        if (closePromise === undefined) {
          closePromise = Promise.resolve().then(function () {
            var endResult = writer.endSync();
            var endPromise = endResult < 0 ? writer.end() : undefined;
            var returnPromise = closeIterator.return();
            return endPromise === undefined
              ? returnPromise
              : Promise.all([endPromise, returnPromise]);
          });
        }
        return closePromise;
      }
    };
    if (typeof Symbol.asyncDispose === 'symbol') {
      channel[Symbol.asyncDispose] = function () { return channel.close(); };
    }
    return channel;
  }

  function duplex(options) {
    if (options !== undefined && options !== null && typeof options !== 'object') {
      throw invalidArgType('options', 'of type object', options);
    }
    options = options || {};
    if (options.signal !== undefined) validateAbortSignal(options.signal);
    var ab = push({ budget: options.budget, backpressure: options.backpressure });
    var ba = push({ budget: options.budget, backpressure: options.backpressure });
    var channelA = createChannel(ab.writer, ba.readable);
    var channelB = createChannel(ba.writer, ab.readable);
    if (options.signal) {
      var abortBoth = function () {
        ab.writer.fail(options.signal.reason);
        ba.writer.fail(options.signal.reason);
      };
      if (options.signal.aborted) abortBoth();
      else if (options.signal.addEventListener) {
        options.signal.addEventListener('abort', abortBoth, { once: true });
      }
    }
    return [channelA, channelB];
  }

  // ------------------------------------------------------ transform results --
  function* flattenTransformYieldSync(value) {
    if (isUint8Array(value)) { yield value; return; }
    if (typeof value === 'string') { yield encodeString(value); return; }
    if (isAnyArrayBuffer(value)) { yield new Uint8Array(value); return; }
    if (isView(value)) { yield toUint8Array(value); return; }
    if (isSyncIterable(value)) {
      for (var item of value) yield* flattenTransformYieldSync(item);
      return;
    }
    throw invalidArgType('value',
      'of type string or an instance of Uint8Array, ArrayBuffer, ArrayBufferView, Iterable',
      value);
  }

  function flattenTransformYieldAsync(value) {
    return iterableAsync(async function* () {
      if (isUint8Array(value)) { yield value; return; }
      if (typeof value === 'string') { yield encodeString(value); return; }
      if (isAnyArrayBuffer(value)) { yield new Uint8Array(value); return; }
      if (isView(value)) { yield toUint8Array(value); return; }
      if (isAsyncIterable(value)) {
        for await (var item of value) yield* flattenTransformYieldAsync(item);
        return;
      }
      if (isSyncIterable(value)) {
        for (var syncItem of value) yield* flattenTransformYieldSync(syncItem);
        return;
      }
      throw invalidArgType('value',
        'of type string or an instance of Uint8Array, ArrayBuffer, ArrayBufferView, Iterable, AsyncIterable',
        value);
    });
  }

  function appendTransformResultSync(target, result) {
    if (result === null || result === undefined) return;
    if (isUint8ArrayBatch(result)) {
      if (result.length > 0) target.push(result);
      return;
    }
    if (isUint8Array(result)) { target.push([result]); return; }
    if (typeof result === 'string') { target.push([encodeString(result)]); return; }
    if (isAnyArrayBuffer(result)) { target.push([new Uint8Array(result)]); return; }
    if (isView(result)) { target.push([toUint8Array(result)]); return; }
    var batch = [];
    if (isSyncIterable(result)) {
      for (var item of result) {
        for (var chunk of flattenTransformYieldSync(item)) batch.push(chunk);
      }
    } else {
      throw invalidArgType('result',
        'an instance of Uint8Array, string, ArrayBuffer, ArrayBufferView, Array, Iterable',
        result);
    }
    if (batch.length > 0) target.push(batch);
  }

  async function appendTransformResultAsync(target, result) {
    if (isPromise(result)) {
      await appendTransformResultAsync(target, await result);
      return;
    }
    if (result === null || result === undefined) return;
    if (isUint8ArrayBatch(result)) {
      if (result.length > 0) target.push(result);
      return;
    }
    if (isUint8Array(result)) { target.push([result]); return; }
    if (typeof result === 'string') { target.push([encodeString(result)]); return; }
    if (isAnyArrayBuffer(result)) { target.push([new Uint8Array(result)]); return; }
    if (isView(result)) { target.push([toUint8Array(result)]); return; }
    var batch = [];
    if (isAsyncIterable(result)) {
      for await (var item of result) {
        for await (var chunk of flattenTransformYieldAsync(item)) batch.push(chunk);
      }
    } else if (isSyncIterable(result)) {
      for (var syncItem of result) {
        for (var syncChunk of flattenTransformYieldSync(syncItem)) batch.push(syncChunk);
      }
    } else {
      throw invalidArgType('result',
        'an instance of Uint8Array, string, ArrayBuffer, ArrayBufferView, Array, Iterable, AsyncIterable',
        result);
    }
    if (batch.length > 0) target.push(batch);
  }

  function* processTransformResultSync(result) {
    var batches = [];
    appendTransformResultSync(batches, result);
    for (var i = 0; i < batches.length; i++) yield batches[i];
  }

  function processTransformResultAsync(result) {
    return iterableAsync(async function* () {
      var batches = [];
      await appendTransformResultAsync(batches, result);
      for (var i = 0; i < batches.length; i++) yield batches[i];
    });
  }

  function normalizeTransformResultSync(result) {
    if (result === null || result === undefined) return null;
    if (isUint8ArrayBatch(result)) return result.length > 0 ? result : null;
    if (isUint8Array(result)) return [result];
    if (typeof result === 'string') return [encodeString(result)];
    if (isAnyArrayBuffer(result)) return [new Uint8Array(result)];
    if (isView(result)) return [toUint8Array(result)];
    var batches = [];
    appendTransformResultSync(batches, result);
    return batches.length > 0 ? batches[0] : null;
  }

  async function normalizeTransformResultAsync(result) {
    if (isPromise(result)) return normalizeTransformResultAsync(await result);
    if (result === null || result === undefined) return null;
    if (isUint8ArrayBatch(result)) return result.length > 0 ? result : null;
    if (isUint8Array(result)) return [result];
    if (typeof result === 'string') return [encodeString(result)];
    if (isAnyArrayBuffer(result)) return [new Uint8Array(result)];
    if (isView(result)) return [toUint8Array(result)];
    var batches = [];
    await appendTransformResultAsync(batches, result);
    return batches.length > 0 ? batches[0] : null;
  }

  // ------------------------------------------------------------- pipelines --
  function* withFlushSync(source) {
    yield* source;
    yield null;
  }

  function withFlushAsync(source) {
    return iterableAsync(async function* () {
      yield* source;
      yield null;
    });
  }

  function* applyStatelessSync(source, transforms) {
    for (var batch of source) {
      var current = batch;
      var dropped = false;
      for (var i = 0; i < transforms.length; i++) {
        var result = transforms[i](current, { signal: undefined });
        if (isPromise(result)) {
          throw invalidArgType('result', 'a synchronous transform result', result);
        }
        if (result === null || result === undefined) { dropped = true; break; }
        if (i === transforms.length - 1) { current = result; continue; }
        current = normalizeTransformResultSync(result);
        if (current === null) { dropped = true; break; }
      }
      if (dropped) continue;
      yield* processTransformResultSync(current);
    }
    var pending = [];
    for (var t = 0; t < transforms.length; t++) {
      var next = [];
      for (var j = 0; j < pending.length; j++) {
        appendTransformResultSync(next, transforms[t](pending[j], { signal: undefined }));
      }
      appendTransformResultSync(next, transforms[t](null, { signal: undefined }));
      pending = next;
    }
    for (var k = 0; k < pending.length; k++) yield pending[k];
  }

  function applyStatelessAsync(source, transforms, signal) {
    return iterableAsync(async function* () {
      for await (var batch of source) {
        throwIfAborted(signal);
        var current = batch;
        var dropped = false;
        for (var i = 0; i < transforms.length; i++) {
          var result = transforms[i](current, { signal: signal });
          if (isPromise(result)) result = await result;
          if (result === null || result === undefined) { dropped = true; break; }
          if (i === transforms.length - 1) { current = result; continue; }
          current = await normalizeTransformResultAsync(result);
          if (current === null) { dropped = true; break; }
        }
        if (dropped) continue;
        yield* processTransformResultAsync(current);
      }
      var pending = [];
      for (var t = 0; t < transforms.length; t++) {
        var next = [];
        for (var j = 0; j < pending.length; j++) {
          await appendTransformResultAsync(next, transforms[t](pending[j], { signal: signal }));
        }
        await appendTransformResultAsync(next, transforms[t](null, { signal: signal }));
        pending = next;
      }
      for (var k = 0; k < pending.length; k++) yield pending[k];
    });
  }

  function* applyStatefulSync(source, transformObject) {
    var output = transformObject.transform.call(transformObject, withFlushSync(source));
    if (isAsyncIterable(output)) {
      throw invalidArgType('result', 'a synchronous transform result', output);
    }
    if (!isSyncIterable(output)) {
      throw invalidArgType('result', 'an instance of Iterable', output);
    }
    for (var item of output) {
      if (item === null || item === undefined) continue;
      if (isUint8ArrayBatch(item)) {
        if (item.length > 0) yield item;
        continue;
      }
      yield* processTransformResultSync(item);
    }
  }

  function applyStatefulAsync(source, transformObject, signal) {
    return iterableAsync(async function* () {
      var output = transformObject.transform.call(transformObject,
        withFlushAsync(source), { signal: signal });
      if (isPromise(output)) output = await output;
      if (isAsyncIterable(output)) {
        for await (var item of output) {
          if (item === null || item === undefined) continue;
          if (isUint8ArrayBatch(item)) {
            if (item.length > 0) yield item;
            continue;
          }
          if (isUint8Array(item)) { yield [item]; continue; }
          var batches = [];
          await appendTransformResultAsync(batches, item);
          for (var i = 0; i < batches.length; i++) yield batches[i];
        }
        return;
      }
      if (isSyncIterable(output)) {
        for (var syncItem of output) {
          if (syncItem === null || syncItem === undefined) continue;
          if (isUint8ArrayBatch(syncItem)) {
            if (syncItem.length > 0) yield syncItem;
            continue;
          }
          var syncBatches = [];
          appendTransformResultSync(syncBatches, syncItem);
          for (var j = 0; j < syncBatches.length; j++) yield syncBatches[j];
        }
        return;
      }
      if (output !== null && output !== undefined) {
        var finalBatches = [];
        await appendTransformResultAsync(finalBatches, output);
        for (var k = 0; k < finalBatches.length; k++) yield finalBatches[k];
      }
    });
  }

  function* createSyncPipeline(source, transforms) {
    var current = source;
    var stateless = [];
    for (var i = 0; i < transforms.length; i++) {
      var transform = transforms[i];
      if (typeof transform === 'function') {
        stateless.push(transform);
        continue;
      }
      if (stateless.length > 0) {
        current = applyStatelessSync(current, stateless);
        stateless = [];
      }
      current = applyStatefulSync(current, transform);
    }
    if (stateless.length > 0) current = applyStatelessSync(current, stateless);
    yield* current;
  }

  function createAsyncPipeline(source, transforms, signal) {
    return iterableAsync(async function* () {
      var current = source;
      var stateless = [];
      for (var i = 0; i < transforms.length; i++) {
        var transform = transforms[i];
        if (typeof transform === 'function') {
          stateless.push(transform);
          continue;
        }
        if (stateless.length > 0) {
          current = applyStatelessAsync(current, stateless, signal);
          stateless = [];
        }
        current = transform[kValidatedTransform]
          ? transform.transform.call(transform, current, { signal: signal })
          : applyStatefulAsync(current, transform, signal);
      }
      if (stateless.length > 0) current = applyStatelessAsync(current, stateless, signal);
      yield* current;
    });
  }

  function parsePullArgs(args) {
    var transforms = args.slice();
    var options = null;
    if (transforms.length > 0 && isPullOptions(transforms[transforms.length - 1])) {
      options = transforms.pop();
    }
    return { transforms: transforms, options: options };
  }

  function checkTransforms(transforms) {
    for (var i = 0; i < transforms.length; i++) {
      if (!isTransform(transforms[i])) {
        throw invalidArgType('transforms[' + i + ']',
          'an instance of Function or an Object with a transform method', transforms[i]);
      }
    }
  }

  function pullSync(source) {
    var parsed = parsePullArgs(Array.prototype.slice.call(arguments, 1));
    checkTransforms(parsed.transforms);
    var normalized = fromSync(source);
    return iterableSync(function () {
      return createSyncPipeline(normalized, parsed.transforms);
    });
  }

  function pull(source) {
    var parsed = parsePullArgs(Array.prototype.slice.call(arguments, 1));
    checkTransforms(parsed.transforms);
    var signal = parsed.options ? parsed.options.signal : undefined;
    if (signal !== undefined) validateAbortSignal(signal);
    var normalized = from(source);
    throwIfAborted(signal);
    return iterableAsync(function () {
      var iterator = createAsyncPipeline(normalized, parsed.transforms, signal)[Symbol.asyncIterator]();
      return {
        next: function (value) { return iterator.next(value); },
        return: function (value) {
          return iterator.return ? iterator.return(value)
            : Promise.resolve({ value: value, done: true });
        },
        throw: function (error) {
          return iterator.throw ? iterator.throw(error) : Promise.reject(error);
        }
      };
    });
  }

  // ----------------------------------------------------------------- pipeTo --
  function parsePipeToArgs(args, method) {
    if (args.length === 0) {
      throw invalidArgValue('args', args, 'pipeTo requires a writer argument');
    }
    var options = null;
    var writerIndex = args.length - 1;
    var last = args[args.length - 1];
    if (isPullOptions(last) && typeof last[method] !== 'function') {
      options = last;
      writerIndex = args.length - 2;
    }
    if (writerIndex < 0) {
      throw invalidArgValue('args', args, 'pipeTo requires a writer argument');
    }
    var writer = args[writerIndex];
    if (writer === null || writer === undefined || typeof writer[method] !== 'function') {
      throw invalidArgType('writer', 'an object with a ' + method + ' method', writer);
    }
    var transforms = args.slice(0, writerIndex);
    checkTransforms(transforms);
    return { transforms: transforms, writer: writer, options: options };
  }

  function pipeToSync(source) {
    var parsed = parsePipeToArgs(Array.prototype.slice.call(arguments, 1), 'writeSync');
    var writer = parsed.writer;
    var options = parsed.options || {};
    var normalized = fromSync(source);
    var pipeline = parsed.transforms.length > 0
      ? createSyncPipeline(normalized, parsed.transforms)
      : normalized;
    var total = 0;
    var hasWritevSync = typeof writer.writevSync === 'function';
    try {
      for (var batch of pipeline) {
        if (hasWritevSync && batch.length > 1) {
          writer.writevSync(batch);
          for (var i = 0; i < batch.length; i++) total += batch[i].byteLength;
        } else {
          for (var j = 0; j < batch.length; j++) {
            writer.writeSync(batch[j]);
            total += batch[j].byteLength;
          }
        }
      }
      if (!options.preventClose) {
        if (typeof writer.endSync === 'function') writer.endSync();
        else if (typeof writer.end === 'function') writer.end();
      }
    } catch (error) {
      if (!options.preventFail && typeof writer.fail === 'function') writer.fail(error);
      throw error;
    }
    return total;
  }

  async function pipeTo(source) {
    var parsed = parsePipeToArgs(Array.prototype.slice.call(arguments, 1), 'write');
    var writer = parsed.writer;
    var options = parsed.options || {};
    var signal = options.signal;
    if (signal !== undefined) validateAbortSignal(signal);
    var pipeline = parsed.transforms.length > 0
      ? pull(source, ...parsed.transforms, signal ? { signal: signal } : {})
      : from(source);
    var total = 0;
    var hasWritev = typeof writer.writev === 'function';
    var hasWritevSync = typeof writer.writevSync === 'function';
    try {
      throwIfAborted(signal);
      for await (var batch of pipeline) {
        throwIfAborted(signal);
        if (hasWritev && batch.length > 1) {
          if (hasWritevSync) writer.writevSync(batch);
          else await writer.writev(batch);
          for (var i = 0; i < batch.length; i++) total += batch[i].byteLength;
          continue;
        }
        for (var j = 0; j < batch.length; j++) {
          if (typeof writer.writeSync === 'function') writer.writeSync(batch[j]);
          else await writer.write(batch[j]);
          total += batch[j].byteLength;
        }
      }
      if (!options.preventClose) {
        if (typeof writer.endSync === 'function') writer.endSync();
        else if (typeof writer.end === 'function') await writer.end();
      }
    } catch (error) {
      if (!options.preventFail && typeof writer.fail === 'function') writer.fail(error);
      throw error;
    }
    return total;
  }

  // ------------------------------------------------------- fan-out helpers --
  function asyncFanout(source) {
    var buffer = [];
    var positions = [];
    var waiters = [];
    var started = false;
    var done = false;
    var failure = null;

    function notify() {
      var pending = waiters;
      waiters = [];
      for (var i = 0; i < pending.length; i++) pending[i]();
    }

    function trim() {
      var min = buffer.length;
      for (var i = 0; i < positions.length; i++) {
        if (positions[i] < min) min = positions[i];
      }
      if (min > 0) {
        buffer = buffer.slice(min);
        for (var j = 0; j < positions.length; j++) positions[j] -= min;
      }
    }

    function start() {
      if (started) return;
      started = true;
      var iterator = source[Symbol.asyncIterator]();
      (async function pump() {
        try {
          while (true) {
            var result = await iterator.next();
            if (result.done) { done = true; break; }
            buffer.push(result.value);
            notify();
          }
        } catch (error) {
          failure = error;
        }
        notify();
      })();
    }

    return {
      add: function () {
        var id = positions.length;
        positions.push(buffer.length);
        start();
        return {
          next: async function () {
            while (true) {
              if (positions[id] < buffer.length) {
                var value = buffer[positions[id]++];
                trim();
                return { value: value, done: false };
              }
              if (failure) throw failure;
              if (done) return { value: undefined, done: true };
              await new Promise(function (resolve) { waiters.push(resolve); });
            }
          },
          return: async function () {
            positions[id] = buffer.length;
            trim();
            return { value: undefined, done: true };
          }
        };
      }
    };
  }

  function syncFanout(source) {
    var buffer = [];
    var positions = [];
    var iterator = source[Symbol.iterator]();
    var done = false;

    function fill(index) {
      while (!done && buffer.length <= index) {
        var result = iterator.next();
        if (result.done) { done = true; break; }
        buffer.push(result.value);
      }
    }
    function trim() {
      var min = buffer.length;
      for (var i = 0; i < positions.length; i++) {
        if (positions[i] < min) min = positions[i];
      }
      if (min > 0) {
        buffer = buffer.slice(min);
        for (var j = 0; j < positions.length; j++) positions[j] -= min;
      }
    }

    return {
      add: function () {
        var id = positions.length;
        positions.push(buffer.length);
        return {
          next: function () {
            fill(positions[id]);
            if (positions[id] < buffer.length) {
              var value = buffer[positions[id]++];
              trim();
              return { value: value, done: false };
            }
            return { value: undefined, done: true };
          },
          return: function () {
            positions[id] = buffer.length;
            trim();
            return { value: undefined, done: true };
          }
        };
      }
    };
  }

  function asAsyncIterable(iterator) {
    var obj = {};
    obj[Symbol.asyncIterator] = function () { return iterator; };
    return obj;
  }
  function asSyncIterable(iterator) {
    var obj = {};
    obj[Symbol.iterator] = function () { return iterator; };
    return obj;
  }

  function Broadcast(source, options) {
    if (options !== undefined && options !== null && typeof options !== 'object') {
      throw invalidArgType('options', 'of type object', options);
    }
    this.source = source;
    this.options = options || {};
    this._fanout = asyncFanout(from(source));
  }
  Broadcast.prototype[Symbol.asyncIterator] = function () { return this._fanout.add(); };
  Broadcast.prototype.push = function () { return asAsyncIterable(this._fanout.add()); };
  Broadcast.prototype.pull = Broadcast.prototype.push;
  Broadcast.prototype[broadcastProtocol] = function () { return this; };
  // node: Broadcast.from(input) accepts a broadcastable protocol object or a
  // plain source; protocol implementations must return an object.
  Broadcast.from = function (input, options) {
    if (input !== null && input !== undefined && hasProtocol(input, broadcastProtocol)) {
      var bc = input[broadcastProtocol](options);
      if (bc === null || typeof bc !== 'object') {
        var err = new TypeError('The "[Symbol.for(\'Stream.broadcastProtocol\')]" method must return an object. Received ' + typeName(bc));
        err.code = 'ERR_INVALID_RETURN_VALUE';
        throw err;
      }
      return { writer: {}, broadcast: bc };
    }
    // Plain source: create a push channel and pump the source into it, so the
    // returned writer feeds every broadcast consumer.
    var queue = [];
    var waiters = [];
    var ended = false;
    var failure = null;
    function wake() {
      var pending = waiters;
      waiters = [];
      for (var i = 0; i < pending.length; i++) pending[i]();
    }
    var channelSource = {};
    channelSource[Symbol.asyncIterator] = function () {
      return {
        next: function () {
          if (queue.length > 0) return Promise.resolve({ value: [queue.shift()], done: false });
          if (failure) return Promise.reject(failure);
          if (ended) return Promise.resolve({ value: undefined, done: true });
          return new Promise(function (resolve) { waiters.push(resolve); });
        },
        return: function () {
          ended = true;
          wake();
          return Promise.resolve({ value: undefined, done: true });
        }
      };
    };
    var bc = new Broadcast(channelSource, options);
    var writer = {
      writeSync: function (chunk) {
        if (failure) throw failure;
        if (ended) throw writeAfterEndError();
        queue.push(toUint8Array(chunk));
        wake();
        return true;
      },
      writevSync: function (chunks) {
        for (var i = 0; i < chunks.length; i++) writer.writeSync(chunks[i]);
        return true;
      },
      endSync: function () { ended = true; wake(); return 0; },
      write: function (chunk) {
        return new Promise(function (resolve, reject) {
          try { writer.writeSync(chunk); resolve(true); } catch (e) { reject(e); }
        });
      },
      writev: function (chunks) {
        return new Promise(function (resolve, reject) {
          try { writer.writevSync(chunks); resolve(true); } catch (e) { reject(e); }
        });
      },
      end: function () { ended = true; wake(); return Promise.resolve(); },
      fail: function (error) { failure = error || new Error('writer failed'); wake(); }
    };
    // Pump the plain source into the channel writer.
    (async function pump() {
      try {
        for await (var batch of from(input)) {
          if (batch.length > 1 && typeof writer.writevSync === 'function') writer.writevSync(batch);
          else for (var i = 0; i < batch.length; i++) writer.writeSync(batch[i]);
        }
        writer.endSync();
      } catch (error) {
        writer.fail(error);
      }
    })();
    return { writer: writer, broadcast: bc };
  };

  function broadcast(source, options) { return new Broadcast(source, options); }

  function Share(source, options) {
    if (options !== undefined && options !== null && typeof options !== 'object') {
      throw invalidArgType('options', 'of type object', options);
    }
    this.source = source;
    this.options = options || {};
    this._fanout = asyncFanout(from(source));
  }
  Share.prototype[Symbol.asyncIterator] = function () { return this._fanout.add(); };
  Share.prototype.pull = function () {
    var consumer = asAsyncIterable(this._fanout.add());
    if (arguments.length > 0) {
      return pull.apply(null, [consumer].concat(Array.prototype.slice.call(arguments)));
    }
    return consumer;
  };
  Share.prototype[shareProtocol] = function () { return this; };
  Share.from = function (input, options) {
    if (input !== null && input !== undefined && hasProtocol(input, shareProtocol)) {
      var shared = input[shareProtocol](options);
      if (shared === null || typeof shared !== 'object') {
        var err = new TypeError('The "[Symbol.for(\'Stream.shareProtocol\')]" method must return an object. Received ' + typeName(shared));
        err.code = 'ERR_INVALID_RETURN_VALUE';
        throw err;
      }
      return shared;
    }
    if (isAsyncIterable(input) || isSyncIterable(input)) return share(input, options);
    throw invalidArgType('input', 'an instance of Shareable, AsyncIterable or Iterable', input);
  };

  function share(source, options) { return new Share(source, options); }

  function SyncShare(source, options) {
    if (options !== undefined && options !== null && typeof options !== 'object') {
      throw invalidArgType('options', 'of type object', options);
    }
    this.source = source;
    this.options = options || {};
    this._fanout = syncFanout(fromSync(source));
  }
  SyncShare.prototype[Symbol.iterator] = function () { return this._fanout.add(); };
  SyncShare.prototype.pull = function () {
    var consumer = asSyncIterable(this._fanout.add());
    if (arguments.length > 0) {
      return pullSync.apply(null, [consumer].concat(Array.prototype.slice.call(arguments)));
    }
    return consumer;
  };
  SyncShare.prototype[shareSyncProtocol] = function () { return this; };
  SyncShare.fromSync = function (input, options) {
    if (input !== null && input !== undefined && hasProtocol(input, shareSyncProtocol)) {
      var shared = input[shareSyncProtocol](options);
      if (shared === null || typeof shared !== 'object') {
        var err = new TypeError('The "[Symbol.for(\'Stream.shareSyncProtocol\')]" method must return an object. Received ' + typeName(shared));
        err.code = 'ERR_INVALID_RETURN_VALUE';
        throw err;
      }
      return shared;
    }
    if (isSyncIterable(input)) return shareSync(input, options);
    throw invalidArgType('input', 'an instance of SyncShareable or Iterable', input);
  };

  function shareSync(source, options) { return new SyncShare(source, options); }

  // --------------------------------------------------------- classic interop --
  function createClassicSource(readable) {
    return iterableAsync(async function* () {
      var queue = [];
      var waiters = [];
      var ended = false;
      var failure = null;

      function wake() {
        var pending = waiters;
        waiters = [];
        for (var i = 0; i < pending.length; i++) pending[i]();
      }
      function drain() {
        var chunk;
        while (true) {
          chunk = readable.read();
          if (chunk === null) break;
          queue.push(chunk);
        }
        wake();
      }
      function onEnd() { ended = true; wake(); }
      function onClose() { ended = true; wake(); }
      function onError(error) { failure = error || new Error('stream error'); wake(); }

      readable.on('readable', drain);
      readable.on('end', onEnd);
      readable.on('close', onClose);
      readable.on('error', onError);
      if (typeof readable.read === 'function') drain();
      try {
        while (true) {
          if (queue.length > 0) {
            var raw = queue.splice(0, queue.length);
            var batch = [];
            for (var i = 0; i < raw.length; i++) {
              if (isUint8Array(raw[i])) {
                batch.push(raw[i]);
              } else {
                for await (var chunk of normalizeAsyncValue(raw[i])) batch.push(chunk);
              }
            }
            if (batch.length > 0) yield batch;
            continue;
          }
          if (failure) throw failure;
          if (ended) return;
          await new Promise(function (resolve) { waiters.push(resolve); });
        }
      } finally {
        if (typeof readable.removeListener === 'function') {
          readable.removeListener('readable', drain);
          readable.removeListener('end', onEnd);
          readable.removeListener('close', onClose);
          readable.removeListener('error', onError);
        }
      }
    });
  }

  function fromReadable(readable) {
    if (readable === null || readable === undefined || typeof readable !== 'object') {
      throw invalidArgType('readable', 'an instance of Readable', readable);
    }
    if (typeof readable.read !== 'function' || typeof readable.on !== 'function') {
      throw invalidArgType('readable', 'an instance of Readable', readable);
    }
    // Node exposes classic Readables through the toAsyncStreamable protocol;
    // honor an implementation supplied by the object itself.
    if (hasProtocol(readable, toAsyncStreamable)) {
      return readable[toAsyncStreamable]();
    }
    return createClassicSource(readable);
  }

  function fromWritable(writable, options) {
    if (writable === null || writable === undefined ||
        typeof writable.write !== 'function' || typeof writable.on !== 'function') {
      throw invalidArgType('writable', 'an instance of Writable', writable);
    }
    if (options !== undefined && options !== null && typeof options !== 'object') {
      throw invalidArgType('options', 'of type object', options);
    }
    options = options || {};
    var policy = validateBackpressure(options.backpressure);
    if (writable.writableObjectMode) {
      var stateErr = new Error('Cannot create a stream/iter Writer from an object-mode Writable');
      stateErr.code = 'ERR_INVALID_STATE';
      throw stateErr;
    }
    if (policy === 'drop-oldest') {
      var policyErr = new TypeError("The argument 'options.backpressure' must not be " +
        "'drop-oldest'. Received 'drop-oldest'");
      policyErr.code = 'ERR_INVALID_ARG_VALUE';
      throw policyErr;
    }
    function writeChunk(chunk) {
      return new Promise(function (resolve, reject) {
        var settled = false;
        var done = function (error) {
          if (settled) return;
          settled = true;
          if (error) reject(error);
          else resolve(true);
        };
        var accepted = writable.write(chunk, done);
        if (accepted === false && typeof writable.once === 'function') {
          writable.once('drain', function () { done(); });
        }
      });
    }
    function isWritable() {
      return !writable.destroyed && !writable.writableFinished && !writable.writableEnded;
    }
    var writer = {
      // The Writer interface includes sync fast paths even when the adapter
      // cannot use them; classic Writables are always asynchronous.
      writeSync: function () { return false; },
      writevSync: function () { return false; },
      endSync: function () { return -1; },
      write: function (chunk) { return writeChunk(toUint8Array(chunk)); },
      writev: async function (chunks) {
        for (var i = 0; i < chunks.length; i++) await writeChunk(toUint8Array(chunks[i]));
        return true;
      },
      end: function (chunk) {
        return new Promise(function (resolve, reject) {
          var onFinish = function () { resolve(); };
          var onError = function (error) { reject(error); };
          if (typeof writable.once === 'function') {
            writable.once('finish', onFinish);
            writable.once('error', onError);
          } else {
            resolve();
          }
          writable.end(chunk === undefined ? undefined : toUint8Array(chunk));
          if (typeof writable.once !== 'function') resolve();
        });
      },
      fail: function (error) {
        if (typeof writable.destroy === 'function') writable.destroy(error);
        return Promise.resolve();
      }
    };
    Object.defineProperty(writer, 'desiredSize', {
      get: function () {
        return typeof writable.writableLength === 'number' ? writable.writableLength : 0;
      }
    });
    Object.defineProperty(writer, 'canWrite', {
      get: function () {
        if (!isWritable()) return null;
        var hwm = typeof writable.writableHighWaterMark === 'number' ?
          writable.writableHighWaterMark : 16384;
        return (typeof writable.writableLength === 'number' ? writable.writableLength : 0) < hwm;
      }
    });
    return writer;
  }

  function toReadable(source, options) {
    if (!isAsyncIterable(source)) {
      throw invalidArgType('source', 'an instance of AsyncIterable', source);
    }
    if (typeof Readable !== 'function') {
      throw invalidArgType('source', 'usable only with the classic stream module installed', source);
    }
    if (options !== undefined && options !== null && typeof options !== 'object') {
      throw invalidArgType('options', 'of type object', options);
    }
    options = options || {};
    var hwm = highWaterMarkOption(options);
    var iterator = source[Symbol.asyncIterator]();
    var ended = false;
    var pumping = false;

    var readable = new Readable({
      highWaterMark: hwm,
      read: function () {
        var self = this;
        if (pumping || ended) return;
        pumping = true;
        (async function pump() {
          try {
            while (true) {
              var result = await iterator.next();
              if (result.done) {
                ended = true;
                pumping = false;
                self.push(null);
                return;
              }
              var batch = result.value;
              if (isUint8Array(batch)) batch = [batch];
              for (var i = 0; i < batch.length; i++) {
                if (!self.push(batch[i])) {
                  pumping = false;
                  return;
                }
              }
            }
          } catch (error) {
            ended = true;
            pumping = false;
            self.destroy(error);
          }
        })();
      }
    });
    if (options.signal && typeof options.signal.addEventListener === 'function') {
      var onAbort = function () { readable.destroy(options.signal.reason); };
      if (options.signal.aborted) onAbort();
      else options.signal.addEventListener('abort', onAbort, { once: true });
    }
    return readable;
  }

  function toReadableSync(source, options) {
    if (!isSyncIterable(source)) {
      throw invalidArgType('source', 'an instance of Iterable', source);
    }
    if (typeof Readable !== 'function') {
      throw invalidArgType('source', 'usable only with the classic stream module installed', source);
    }
    if (options !== undefined && options !== null && typeof options !== 'object') {
      throw invalidArgType('options', 'of type object', options);
    }
    options = options || {};
    var hwm = highWaterMarkOption(options);
    var iterator = source[Symbol.iterator]();
    var batch = null;
    var index = 0;
    var ended = false;
    return new Readable({
      highWaterMark: hwm,
      read: function () {
        while (true) {
          if (batch !== null && index < batch.length) {
            if (!this.push(batch[index++])) return;
            continue;
          }
          batch = null;
          index = 0;
          if (ended) return;
          var result = iterator.next();
          if (result.done) {
            ended = true;
            this.push(null);
            return;
          }
          var value = result.value;
          batch = isUint8Array(value) ? [value] : value;
          if (!Array.isArray(batch)) batch = [toUint8Array(batch)];
        }
      }
    });
  }

  function toWritable(writer) {
    if (writer === null || writer === undefined || typeof writer.write !== 'function') {
      throw invalidArgType('writer', 'an instance of Writer', writer);
    }
    if (typeof Writable !== 'function') {
      throw invalidArgType('writer', 'usable only with the classic stream module installed', writer);
    }
    return new Writable({
      write: function (chunk, encoding, callback) {
        var bytes = typeof chunk === 'string' ? encodeString(chunk) : chunk;
        var finish = function (error) { if (typeof callback === 'function') callback(error); };
        if (typeof writer.writeSync === 'function') {
          try {
            if (writer.writeSync(bytes)) { finish(); return; }
          } catch (error) {
            finish(error);
            return;
          }
        }
        Promise.resolve(writer.write(bytes)).then(function () { finish(); }, function (error) {
          finish(error);
        });
      }
    });
  }

  // ---------------------------------------------------------------- exports --
  var Stream = {
    push: push,
    duplex: duplex,
    from: from,
    fromSync: fromSync,
    pull: pull,
    pullSync: pullSync,
    pipeTo: pipeTo,
    pipeToSync: pipeToSync,
    bytes: bytes,
    text: text,
    arrayBuffer: arrayBuffer,
    array: array,
    bytesSync: bytesSync,
    textSync: textSync,
    arrayBufferSync: arrayBufferSync,
    arraySync: arraySync,
    merge: merge,
    broadcast: broadcast,
    share: share,
    shareSync: shareSync,
    tap: tap,
    tapSync: tapSync,
    ondrain: ondrain,
    toStreamable: toStreamable,
    toAsyncStreamable: toAsyncStreamable,
    broadcastProtocol: broadcastProtocol,
    shareProtocol: shareProtocol,
    shareSyncProtocol: shareSyncProtocol,
    drainableProtocol: drainableProtocol
  };
  Object.freeze(Stream);

  exports.Stream = Stream;
  exports.toStreamable = toStreamable;
  exports.toAsyncStreamable = toAsyncStreamable;
  exports.broadcastProtocol = broadcastProtocol;
  exports.shareProtocol = shareProtocol;
  exports.shareSyncProtocol = shareSyncProtocol;
  exports.drainableProtocol = drainableProtocol;

  exports.push = push;
  exports.duplex = duplex;
  exports.from = from;
  exports.fromSync = fromSync;

  exports.pull = pull;
  exports.pullSync = pullSync;
  exports.pipeTo = pipeTo;
  exports.pipeToSync = pipeToSync;

  exports.bytes = bytes;
  exports.text = text;
  exports.arrayBuffer = arrayBuffer;
  exports.array = array;
  exports.bytesSync = bytesSync;
  exports.textSync = textSync;
  exports.arrayBufferSync = arrayBufferSync;
  exports.arraySync = arraySync;

  exports.merge = merge;
  exports.broadcast = broadcast;
  exports.Broadcast = Broadcast;
  exports.share = share;
  exports.shareSync = shareSync;
  exports.Share = Share;
  exports.SyncShare = SyncShare;

  exports.tap = tap;
  exports.tapSync = tapSync;
  exports.ondrain = ondrain;

  exports.fromReadable = fromReadable;
  exports.fromWritable = fromWritable;

  // classic.js installs this protocol so 'from(readable)' and iter consumers
  // can treat a classic Readable as a byte source. Build it directly instead
  // of calling fromReadable(), which would recurse through the protocol.
  if (typeof Readable === 'function' && Readable.prototype &&
      !Readable.prototype[toAsyncStreamable]) {
    Readable.prototype[toAsyncStreamable] = function () {
      return createClassicSource(this);
    };
  }
  exports.toReadable = toReadable;
  exports.toReadableSync = toReadableSync;
  exports.toWritable = toWritable;
})`

const consumersPrelude = `(function (exports, deps) {
  'use strict';
  var BufferCtor = deps.Buffer || (typeof Buffer !== 'undefined' ? Buffer : null);

  function isUint8Array(v) { return v instanceof Uint8Array; }
  function isAnyArrayBuffer(v) {
    return v instanceof ArrayBuffer ||
      (typeof SharedArrayBuffer !== 'undefined' && v instanceof SharedArrayBuffer);
  }
  function toUint8Array(chunk) {
    if (typeof chunk === 'string') {
      return BufferCtor ? BufferCtor.from(chunk, 'utf8') : new TextEncoder().encode(chunk);
    }
    if (isAnyArrayBuffer(chunk)) return new Uint8Array(chunk);
    if (isUint8Array(chunk)) return chunk;
    if (ArrayBuffer.isView(chunk)) {
      return new Uint8Array(chunk.buffer, chunk.byteOffset, chunk.byteLength);
    }
    return chunk;
  }

  function concat(chunks) {
    var byteChunks = [];
    var total = 0;
    var i;
    for (i = 0; i < chunks.length; i++) {
      if (chunks[i] === null || chunks[i] === undefined) continue;
      var bytes = toUint8Array(chunks[i]);
      byteChunks.push(bytes);
      total += bytes.byteLength;
    }
    var out = new Uint8Array(total);
    var offset = 0;
    for (i = 0; i < byteChunks.length; i++) {
      out.set(byteChunks[i], offset);
      offset += byteChunks[i].byteLength;
    }
    return out;
  }

  async function* iterate(stream) {
    if (stream === null || stream === undefined) {
      throw new TypeError('The "stream" argument must be an instance of AsyncIterable or Readable. Received ' + stream);
    }
    if (typeof stream[Symbol.asyncIterator] === 'function') {
      for await (var chunk of stream) yield chunk;
      return;
    }
    if (typeof stream.getReader === 'function') {
      var reader = stream.getReader();
      try {
        while (true) {
          var result = await reader.read();
          if (result.done) return;
          yield result.value;
        }
      } finally {
        if (typeof reader.releaseLock === 'function') reader.releaseLock();
      }
      return;
    }
    if (typeof stream.on === 'function' && typeof stream.read === 'function') {
      var queue = [];
      var waiters = [];
      var ended = false;
      var failure = null;
      var wake = function () {
        var pending = waiters;
        waiters = [];
        for (var i = 0; i < pending.length; i++) pending[i]();
      };
      var drain = function () {
        var chunk;
        while (true) {
          chunk = stream.read();
          if (chunk === null) break;
          queue.push(chunk);
        }
        wake();
      };
      stream.on('readable', drain);
      stream.on('end', function () { ended = true; wake(); });
      stream.on('close', function () { ended = true; wake(); });
      stream.on('error', function (error) { failure = error; wake(); });
      try {
        while (true) {
          if (queue.length > 0) { yield queue.shift(); continue; }
          if (failure) throw failure;
          if (ended) return;
          await new Promise(function (resolve) { waiters.push(resolve); });
        }
      } finally {
        if (typeof stream.removeListener === 'function') {
          stream.removeListener('readable', drain);
        }
      }
      return;
    }
    throw new TypeError('The "stream" argument must be an instance of AsyncIterable or Readable. Received ' + stream);
  }

  async function blob(stream) {
    var chunks = [];
    for await (var chunk of iterate(stream)) chunks.push(chunk);
    if (typeof Blob === 'function') return new Blob(chunks);
    // Fallback when the Blob global is unavailable: expose just the
    // ArrayBuffer view of the collected bytes, which is what the other
    // consumers in this module read anyway.
    var data = concat(chunks);
    var buffer = data.byteOffset === 0 && data.byteLength === data.buffer.byteLength
      ? data.buffer
      : data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength);
    return {
      size: buffer.byteLength,
      type: '',
      arrayBuffer: function () { return Promise.resolve(buffer); },
      bytes: function () { return Promise.resolve(new Uint8Array(buffer)); },
      text: function () { return Promise.resolve(new TextDecoder().decode(buffer)); }
    };
  }

  async function arrayBuffer(stream) {
    var chunks = [];
    for await (var chunk of iterate(stream)) chunks.push(chunk);
    var bytes = concat(chunks);
    if (bytes.byteOffset === 0 && bytes.byteLength === bytes.buffer.byteLength) {
      return bytes.buffer;
    }
    return bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength);
  }

  async function buffer(stream) {
    var data = await arrayBuffer(stream);
    if (BufferCtor) return BufferCtor.from(data);
    return new Uint8Array(data);
  }

  async function bytes(stream) {
    return new Uint8Array(await arrayBuffer(stream));
  }

  function decodeAllChunks(chunks) {
    var out = '';
    var decoder = typeof TextDecoder === 'function' ? new TextDecoder() : null;
    for (var i = 0; i < chunks.length; i++) {
      var chunk = chunks[i];
      if (typeof chunk === 'string') {
        out += chunk;
      } else if (decoder) {
        out += decoder.decode(chunk, { stream: true });
      } else if (BufferCtor) {
        out += BufferCtor.from(chunk).toString('utf8');
      } else {
        for (var j = 0; j < chunk.length; j++) out += String.fromCharCode(chunk[j]);
      }
    }
    if (decoder) out += decoder.decode(undefined, { stream: false });
    return out;
  }

  async function text(stream) {
    var chunks = [];
    for await (var chunk of iterate(stream)) chunks.push(chunk);
    return decodeAllChunks(chunks);
  }

  async function json(stream) {
    return JSON.parse(await text(stream));
  }

  exports.arrayBuffer = arrayBuffer;
  exports.blob = blob;
  exports.buffer = buffer;
  exports.bytes = bytes;
  exports.text = text;
  exports.json = json;
})`

const webPrelude = `(function (exports) {
  'use strict';

  // Node installs the WHATWG stream classes as globals; GodeJS ships its own
  // minimal implementation so require('stream/web') never fails when the web
  // platform globals are unavailable.
  if (typeof globalThis.ReadableStream === 'function' &&
      typeof globalThis.WritableStream === 'function' &&
      typeof globalThis.TransformStream === 'function') {
    exports.ReadableStream = globalThis.ReadableStream;
    exports.WritableStream = globalThis.WritableStream;
    exports.TransformStream = globalThis.TransformStream;
    if (typeof globalThis.ByteLengthQueuingStrategy === 'function') {
      exports.ByteLengthQueuingStrategy = globalThis.ByteLengthQueuingStrategy;
    }
    if (typeof globalThis.CountQueuingStrategy === 'function') {
      exports.CountQueuingStrategy = globalThis.CountQueuingStrategy;
    }
    return;
  }

  function noop() {}

  function domException(message, name) {
    if (typeof DOMException === 'function') return new DOMException(message, name);
    var err = new Error(message);
    err.name = name;
    return err;
  }

  function validateStrategy(strategy) {
    if (strategy === undefined || strategy === null) return strategy;
    if (typeof strategy !== 'object') {
      throw new TypeError('strategy must be an object');
    }
    return strategy;
  }

  function ReadableStream(source, strategy) {
    if (!(this instanceof ReadableStream)) return new ReadableStream(source, strategy);
    var self = this;
    this._queue = [];
    this._waiters = [];
    this._closed = false;
    this._errored = null;
    this._controller = null;
    this.locked = false;
    this._disturbed = false;
    strategy = validateStrategy(strategy);

    var controller = {
      get desiredSize() {
        var size = strategy && typeof strategy.size === 'function'
          ? self._queue.length : self._queue.length;
        var hwm = strategy && typeof strategy.highWaterMark === 'number'
          ? strategy.highWaterMark : 1;
        return hwm - size;
      },
      enqueue: function (chunk) {
        if (self._closed) throw new TypeError('Cannot enqueue after close');
        self._queue.push(chunk);
        self._wake();
      },
      close: function () {
        if (self._closed) throw new TypeError('Cannot close twice');
        self._closed = true;
        self._wake();
      },
      error: function (reason) {
        self._errored = reason;
        self._closed = true;
        self._wake();
      }
    };
    this._controller = controller;

    if (source === null || typeof source !== 'object') {
      if (source !== undefined) throw new TypeError('source must be an object');
      source = {};
    }
    if (typeof source.start === 'function') {
      try {
        var started = source.start(controller);
        if (started && typeof started.then === 'function') {
          started.then(noop, function (error) { controller.error(error); });
        }
      } catch (error) {
        controller.error(error);
      }
    }
    this._pull = typeof source.pull === 'function' ? source.pull : null;
    this._cancel = typeof source.cancel === 'function' ? source.cancel : null;
  }

  ReadableStream.prototype._wake = function () {
    var pending = this._waiters;
    this._waiters = [];
    for (var i = 0; i < pending.length; i++) pending[i]();
  };

  ReadableStream.prototype._next = function () {
    var self = this;
    return new Promise(function (resolve, reject) {
      function attempt() {
        if (self._errored) { reject(self._errored); return; }
        if (self._queue.length > 0) {
          resolve({ value: self._queue.shift(), done: false });
          return;
        }
        if (self._closed) { resolve({ value: undefined, done: true }); return; }
        self._waiters.push(attempt);
        if (self._pull) {
          var pulled = self._pull(self._controller);
          if (pulled && typeof pulled.then === 'function') {
            pulled.then(noop, function (error) { self._controller.error(error); });
          }
        }
      }
      attempt();
    });
  };

  ReadableStream.prototype.getReader = function (options) {
    if (this.locked) throw new TypeError('ReadableStream is locked');
    var self = this;
    this.locked = true;
    var released = false;
    return {
      closed: Promise.resolve(),
      get desiredSize() {
        return self._closed ? 0 : self._queue.length === 0 ? 1 : 0;
      },
      cancel: function (reason) {
        self._closed = true;
        self._wake();
        if (self._cancel) return Promise.resolve(self._cancel(reason));
        return Promise.resolve();
      },
      read: function () {
        self._disturbed = true;
        return self._next();
      },
      releaseLock: function () {
        released = true;
        self.locked = false;
      },
      get released() { return released; }
    };
  };

  ReadableStream.prototype.cancel = function (reason) {
    this._closed = true;
    this._wake();
    if (this._cancel) return Promise.resolve(this._cancel(reason));
    return Promise.resolve();
  };

  ReadableStream.prototype[Symbol.asyncIterator] = function () {
    var reader = this.getReader();
    return {
      next: function () { return reader.read(); },
      return: function (value) {
        reader.releaseLock();
        return Promise.resolve({ value: value, done: true });
      }
    };
  };

  ReadableStream.prototype.values = function () {
    return this[Symbol.asyncIterator]();
  };

  ReadableStream.prototype.pipeTo = function (destination, options) {
    options = options || {};
    var reader = this.getReader();
    var writer = destination && typeof destination.getWriter === 'function'
      ? destination.getWriter() : destination;
    return (async function () {
      try {
        for (;;) {
          var result = await reader.read();
          if (result.done) break;
          await writer.write(result.value);
        }
        if (options.preventClose !== true) await writer.close();
      } catch (error) {
        if (options.preventAbort !== true && typeof writer.abort === 'function') {
          try { await writer.abort(error); } catch (ignored) {}
        }
        throw error;
      } finally {
        try { reader.releaseLock(); } catch (ignored) {}
      }
    })();
  };

  ReadableStream.prototype.pipeThrough = function (transform, options) {
    if (!transform || !transform.readable || !transform.writable) {
      throw new TypeError('pipeThrough requires a TransformStream');
    }
    this.pipeTo(transform.writable, options).catch(noop);
    return transform.readable;
  };

  ReadableStream.prototype.tee = function () {
    var source = this;
    var sinks = [];
    function branch() {
      return new ReadableStream({
        start: function (controller) {
          sinks.push(function (chunk, done, failure) {
            if (failure) { controller.error(failure); return; }
            if (done) { controller.close(); return; }
            controller.enqueue(chunk);
          });
        }
      });
    }
    var first = branch();
    var second = branch();
    (async function () {
      var reader = source.getReader();
      try {
        for (;;) {
          var result = await reader.read();
          for (var i = 0; i < sinks.length; i++) {
            sinks[i](result.value, result.done, null);
          }
          if (result.done) break;
        }
      } catch (error) {
        for (var i = 0; i < sinks.length; i++) sinks[i](undefined, true, error);
      }
    })();
    return [first, second];
  };

  function WritableStream(sink, strategy) {
    if (!(this instanceof WritableStream)) return new WritableStream(sink, strategy);
    this._sink = sink || {};
    this._closed = false;
    this._errored = null;
    this.locked = false;
    strategy = validateStrategy(strategy);
  }

  WritableStream.prototype.getWriter = function () {
    if (this.locked) throw new TypeError('WritableStream is locked');
    var self = this;
    this.locked = true;
    var controller = {
      get desiredSize() { return self._errored ? 0 : 65536; },
      error: function (reason) {
        self._errored = reason;
        self._closed = true;
      }
    };
    return {
      get desiredSize() { return controller.desiredSize; },
      closed: Promise.resolve(),
      ready: Promise.resolve(),
      write: function (chunk) {
        if (self._errored) return Promise.reject(self._errored);
        if (self._closed) return Promise.reject(new TypeError('WritableStream is closed'));
        try {
          var result = typeof self._sink.write === 'function'
            ? self._sink.write(chunk, controller) : undefined;
          return Promise.resolve(result).then(function () {});
        } catch (error) {
          return Promise.reject(error);
        }
      },
      close: function () {
        if (self._closed) return Promise.reject(new TypeError('WritableStream is closed'));
        self._closed = true;
        try {
          var result = typeof self._sink.close === 'function' ? self._sink.close() : undefined;
          return Promise.resolve(result).then(function () {});
        } catch (error) {
          return Promise.reject(error);
        }
      },
      abort: function (reason) {
        self._errored = reason;
        self._closed = true;
        try {
          var result = typeof self._sink.abort === 'function' ? self._sink.abort(reason) : undefined;
          return Promise.resolve(result).then(function () {});
        } catch (error) {
          return Promise.reject(error);
        }
      },
      releaseLock: function () { self.locked = false; }
    };
  };

  WritableStream.prototype.abort = function (reason) {
    this._closed = true;
    this._errored = reason;
    if (typeof this._sink.abort === 'function') return Promise.resolve(this._sink.abort(reason));
    return Promise.resolve();
  };

  WritableStream.prototype.close = function () {
    this._closed = true;
    if (typeof this._sink.close === 'function') return Promise.resolve(this._sink.close());
    return Promise.resolve();
  };

  function TransformStream(transformer, writableStrategy, readableStrategy) {
    if (!(this instanceof TransformStream)) {
      return new TransformStream(transformer, writableStrategy, readableStrategy);
    }
    transformer = transformer || {};
    var self = this;
    var readableController = null;
    var readable = new ReadableStream({
      start: function (controller) { readableController = controller; }
    });
    var transform = typeof transformer.transform === 'function'
      ? transformer.transform
      : function (chunk, controller) { controller.enqueue(chunk); };
    var flush = typeof transformer.flush === 'function' ? transformer.flush : null;
    var writable = new WritableStream({
      write: function (chunk, controller) {
        return transform(chunk, readableController);
      },
      close: function () {
        var result = flush ? flush(readableController) : undefined;
        if (result && typeof result.then === 'function') {
          return result.then(function () { readableController.close(); });
        }
        readableController.close();
        return undefined;
      },
      abort: function (reason) { readableController.error(reason); }
    }, writableStrategy);
    this.readable = readable;
    this.writable = writable;
    this._self = self;
  }

  function ByteLengthQueuingStrategy(init) {
    if (!(this instanceof ByteLengthQueuingStrategy)) return new ByteLengthQueuingStrategy(init);
    init = init || {};
    this.highWaterMark = init.highWaterMark;
  }
  ByteLengthQueuingStrategy.prototype.size = function (chunk) {
    return chunk && typeof chunk.byteLength === 'number' ? chunk.byteLength : 0;
  };

  function CountQueuingStrategy(init) {
    if (!(this instanceof CountQueuingStrategy)) return new CountQueuingStrategy(init);
    init = init || {};
    this.highWaterMark = init.highWaterMark;
  }
  CountQueuingStrategy.prototype.size = function () { return 1; };

  exports.ReadableStream = ReadableStream;
  exports.WritableStream = WritableStream;
  exports.TransformStream = TransformStream;
  exports.ByteLengthQueuingStrategy = ByteLengthQueuingStrategy;
  exports.CountQueuingStrategy = CountQueuingStrategy;

  if (typeof globalThis.ReadableStream !== 'function') globalThis.ReadableStream = ReadableStream;
  if (typeof globalThis.WritableStream !== 'function') globalThis.WritableStream = WritableStream;
  if (typeof globalThis.TransformStream !== 'function') globalThis.TransformStream = TransformStream;
  if (typeof globalThis.ByteLengthQueuingStrategy !== 'function') {
    globalThis.ByteLengthQueuingStrategy = ByteLengthQueuingStrategy;
  }
  if (typeof globalThis.CountQueuingStrategy !== 'function') {
    globalThis.CountQueuingStrategy = CountQueuingStrategy;
  }
})`
