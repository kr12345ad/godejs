package web

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

// installModern adds the platform classes Node installs at startup that were
// not part of the original web prelude: MessageEvent, explicit resource
// management stacks, text transform streams and compression streams.
func installModern(e *env.Env) {
	fn, err := e.JS.Eval(modernPrelude, "[gode:web modern]")
	if err != nil || !fn.Valid() || !fn.IsFunction() {
		return
	}
	_, _ = fn.Object().Call(js.Object{})
}

const modernPrelude = `(function () {
  'use strict';

  function load(name) {
    if (typeof require === 'function') return require(name);
    if (typeof globalThis.require === 'function') return globalThis.require(name);
    throw new Error('require is unavailable');
  }

  function toBytes(value) {
    if (value instanceof Uint8Array) return value;
    if (typeof value === 'string') return new TextEncoder().encode(value);
    if (typeof ArrayBuffer !== 'undefined' && value instanceof ArrayBuffer) {
      return new Uint8Array(value);
    }
    if (typeof ArrayBuffer !== 'undefined' && ArrayBuffer.isView(value)) {
      return new Uint8Array(value.buffer, value.byteOffset, value.byteLength);
    }
    return new TextEncoder().encode(String(value));
  }

  function concatBytes(chunks) {
    var length = 0;
    for (var i = 0; i < chunks.length; i++) length += chunks[i].byteLength;
    var out = new Uint8Array(length);
    var offset = 0;
    for (var j = 0; j < chunks.length; j++) {
      out.set(chunks[j], offset);
      offset += chunks[j].byteLength;
    }
    return out;
  }

  // --- MessageEvent ------------------------------------------------------
  if (typeof globalThis.MessageEvent !== 'function' && typeof globalThis.Event === 'function') {
    class MessageEvent extends Event {
      constructor(type, init) {
        init = init || {};
        super(type, init);
        function define(name, value) {
          Object.defineProperty(this, name, {
            value: value, writable: false, enumerable: true, configurable: true
          });
        }
        define.call(this, 'data', init.data === undefined ? null : init.data);
        define.call(this, 'origin', init.origin === undefined ? '' : String(init.origin));
        define.call(this, 'lastEventId',
          init.lastEventId === undefined ? '' : String(init.lastEventId));
        define.call(this, 'source', init.source === undefined ? null : init.source);
        define.call(this, 'ports', init.ports === undefined ? [] : init.ports);
      }
    }
    Object.defineProperty(MessageEvent.prototype, Symbol.toStringTag, {
      value: 'MessageEvent', configurable: true
    });
    globalThis.MessageEvent = MessageEvent;
  }

  // --- explicit resource management --------------------------------------
  var disposeSymbol = typeof Symbol.dispose === 'symbol'
    ? Symbol.dispose : Symbol('Symbol.dispose');
  var asyncDisposeSymbol = typeof Symbol.asyncDispose === 'symbol'
    ? Symbol.asyncDispose : Symbol('Symbol.asyncDispose');
  if (typeof Symbol.dispose !== 'symbol') {
    Object.defineProperty(Symbol, 'dispose', { value: disposeSymbol, configurable: true });
  }
  if (typeof Symbol.asyncDispose !== 'symbol') {
    Object.defineProperty(Symbol, 'asyncDispose', {
      value: asyncDisposeSymbol, configurable: true
    });
  }

  var SuppressedError = globalThis.SuppressedError || class SuppressedError extends Error {
    constructor(error, suppressed, message) {
      super(message);
      this.name = 'SuppressedError';
      this.error = error;
      this.suppressed = suppressed;
    }
  };
  globalThis.SuppressedError = SuppressedError;

  function suppressed(previous, error) {
    return previous === null ? error
      : new SuppressedError(error, previous, 'An error was suppressed during disposal');
  }

  class DisposableStack {
    constructor() { this._head = null; this._disposed = false; }
    get disposed() { return this._disposed; }
    _check() {
      if (this._disposed) {
        throw new ReferenceError('DisposableStack has already been disposed');
      }
    }
    use(value) {
      this._check();
      if (value === null || value === undefined) return value;
      var fn = value[disposeSymbol];
      if (typeof fn !== 'function') throw new TypeError('Object is not disposable');
      this._head = { run: fn.bind(value), next: this._head };
      return value;
    }
    adopt(value, onDispose) {
      this._check();
      if (typeof onDispose !== 'function') throw new TypeError('onDispose must be a function');
      this._head = { run: function () { return onDispose(value); }, next: this._head };
      return value;
    }
    defer(onDispose) {
      this._check();
      if (typeof onDispose !== 'function') throw new TypeError('onDispose must be a function');
      this._head = { run: onDispose, next: this._head };
    }
    move() {
      this._check();
      var next = new DisposableStack();
      next._head = this._head;
      this._head = null;
      this._disposed = true;
      return next;
    }
    dispose() {
      if (this._disposed) return undefined;
      this._disposed = true;
      var error = null;
      while (this._head !== null) {
        var node = this._head;
        this._head = node.next;
        try { node.run(); } catch (thrown) { error = suppressed(error, thrown); }
      }
      if (error !== null) throw error;
      return undefined;
    }
    [disposeSymbol]() { this.dispose(); }
  }
  Object.defineProperty(DisposableStack.prototype, Symbol.toStringTag,
    { value: 'DisposableStack', configurable: true });

  class AsyncDisposableStack {
    constructor() { this._head = null; this._disposed = false; }
    get disposed() { return this._disposed; }
    _check() {
      if (this._disposed) {
        throw new ReferenceError('AsyncDisposableStack has already been disposed');
      }
    }
    use(value) {
      this._check();
      if (value === null || value === undefined) return value;
      var syncFn = value[disposeSymbol];
      var asyncFn = value[asyncDisposeSymbol];
      var fn = typeof asyncFn === 'function' ? asyncFn : syncFn;
      if (typeof fn !== 'function') throw new TypeError('Object is not disposable');
      this._head = { run: fn.bind(value), next: this._head };
      return value;
    }
    adopt(value, onDispose) {
      this._check();
      if (typeof onDispose !== 'function') throw new TypeError('onDispose must be a function');
      this._head = { run: function () { return onDispose(value); }, next: this._head };
      return value;
    }
    defer(onDispose) {
      this._check();
      if (typeof onDispose !== 'function') throw new TypeError('onDispose must be a function');
      this._head = { run: onDispose, next: this._head };
    }
    move() {
      this._check();
      var next = new AsyncDisposableStack();
      next._head = this._head;
      this._head = null;
      this._disposed = true;
      return next;
    }
    async disposeAsync() {
      if (this._disposed) return undefined;
      this._disposed = true;
      var error = null;
      while (this._head !== null) {
        var node = this._head;
        this._head = node.next;
        try { await node.run(); } catch (thrown) { error = suppressed(error, thrown); }
      }
      if (error !== null) throw error;
      return undefined;
    }
    [asyncDisposeSymbol]() { return this.disposeAsync(); }
  }
  Object.defineProperty(AsyncDisposableStack.prototype, Symbol.toStringTag,
    { value: 'AsyncDisposableStack', configurable: true });

  globalThis.DisposableStack = DisposableStack;
  globalThis.AsyncDisposableStack = AsyncDisposableStack;

  // --- text transform streams --------------------------------------------
  function Transform() {
    if (typeof globalThis.TransformStream !== 'function') {
      throw new Error('TransformStream is unavailable');
    }
    return globalThis.TransformStream;
  }

  if (typeof globalThis.TextEncoderStream !== 'function') {
    class TextEncoderStream {
      constructor() {
        var encoder = new TextEncoder();
        var stream = new (Transform())({
          transform: function (chunk, controller) {
            controller.enqueue(encoder.encode(String(chunk)));
          }
        });
        this._readable = stream.readable;
        this._writable = stream.writable;
      }
      get encoding() { return 'utf-8'; }
      get readable() { return this._readable; }
      get writable() { return this._writable; }
    }
    Object.defineProperty(TextEncoderStream.prototype, Symbol.toStringTag,
      { value: 'TextEncoderStream', configurable: true });
    globalThis.TextEncoderStream = TextEncoderStream;
  }

  if (typeof globalThis.TextDecoderStream !== 'function') {
    class TextDecoderStream {
      constructor(label, options) {
        var decoder = new TextDecoder(label, options);
        var chunks = [];
        var stream = new (Transform())({
          transform: function (chunk, controller) { chunks.push(toBytes(chunk)); },
          flush: function (controller) {
            var text = decoder.decode(concatBytes(chunks));
            if (text.length > 0) controller.enqueue(text);
          }
        });
        this._readable = stream.readable;
        this._writable = stream.writable;
        this.encoding = decoder.encoding;
        this.fatal = !!decoder.fatal;
        this.ignoreBOM = !!decoder.ignoreBOM;
      }
      get readable() { return this._readable; }
      get writable() { return this._writable; }
    }
    Object.defineProperty(TextDecoderStream.prototype, Symbol.toStringTag,
      { value: 'TextDecoderStream', configurable: true });
    globalThis.TextDecoderStream = TextDecoderStream;
  }

  // --- compression streams ------------------------------------------------
  var COMPRESSION_FORMATS = { gzip: true, deflate: true, 'deflate-raw': true, br: true };

  function compressionFormat(format) {
    var name = String(format);
    if (!COMPRESSION_FORMATS[name]) {
      throw new TypeError("Unsupported compression format: '" + name + "'");
    }
    return name;
  }

  function compressedClass(name, compress) {
    var klass = function (format) {
      if (!(this instanceof klass)) return new klass(format);
      var normalized = compressionFormat(format);
      var chunks = [];
      var stream = new (Transform())({
        transform: function (chunk, controller) { chunks.push(toBytes(chunk)); },
        flush: function (controller) {
          var output = compress(normalized, concatBytes(chunks));
          controller.enqueue(
            new Uint8Array(output.buffer, output.byteOffset, output.byteLength));
        }
      });
      this._readable = stream.readable;
      this._writable = stream.writable;
    };
    try { Object.defineProperty(klass, 'name', { value: name }); } catch (ignored) {}
    Object.defineProperty(klass.prototype, 'readable', {
      get: function () { return this._readable; }, configurable: true
    });
    Object.defineProperty(klass.prototype, 'writable', {
      get: function () { return this._writable; }, configurable: true
    });
    Object.defineProperty(klass.prototype, Symbol.toStringTag,
      { value: name, configurable: true });
    return klass;
  }

  if (typeof globalThis.CompressionStream !== 'function') {
    globalThis.CompressionStream = compressedClass('CompressionStream',
      function (format, input) {
        var zlib = load('zlib');
        if (format === 'gzip') return zlib.gzipSync(input);
        if (format === 'deflate') return zlib.deflateSync(input);
        if (format === 'deflate-raw') {
          return typeof zlib.deflateRawSync === 'function'
            ? zlib.deflateRawSync(input) : zlib.deflateSync(input, { raw: true });
        }
        return zlib.brotliCompressSync(input);
      });
  }

  if (typeof globalThis.DecompressionStream !== 'function') {
    globalThis.DecompressionStream = compressedClass('DecompressionStream',
      function (format, input) {
        var zlib = load('zlib');
        if (format === 'gzip') return zlib.gunzipSync(input);
        if (format === 'deflate') return zlib.inflateSync(input);
        if (format === 'deflate-raw') {
          return typeof zlib.inflateRawSync === 'function'
            ? zlib.inflateRawSync(input) : zlib.inflateSync(input, { raw: true });
        }
        return zlib.brotliDecompressSync(input);
      });
  }
})`
