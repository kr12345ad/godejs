// Package stream implements Node's stream module for GodeJS.
//
// The stream state machine is written in JavaScript and evaluated once per
// environment; the host only provides byte conversion, stdout writes, the
// microtask queue and the default high-water mark. That keeps backpressure
// bookkeeping and event ordering in one readable place while still using the
// runtime's real event loop (env.NextTick), not a bespoke Go scheduler.
package stream

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"unicode/utf16"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.Register("stream", Install)
	// Tests running with --expose-internals import Node's private stream
	// helpers directly. They are thin adapters over the public functions,
	// and their module.exports is callable (a function, not an object).
	env.Register("internal/streams/end-of-stream", func(e *env.Env) js.Object {
		finished := e.Exports("stream").Get("finished")
		v, err := e.JS.Eval(`(function (finished) {
			function eos(stream, options, callback) {
				if (typeof options === 'function') { callback = options; options = {}; }
				if (typeof options === 'boolean') options = { readable: options };
				return finished(stream, options || {}, callback);
			}
			eos.finished = finished;
			return eos;
		})`, "[gode:internal/streams/end-of-stream]")
		if err == nil && v.IsFunction() {
			if r, cerr := v.Object().Call(js.Object{}, finished); cerr == nil && r.Valid() {
				return r.Object()
			}
		}
		return e.JS.Obj()
	})
	env.Register("internal/streams/add-abort-signal", func(e *env.Env) js.Object {
		add := e.Exports("stream").Get("addAbortSignal")
		v, err := e.JS.Eval(`(function (addAbortSignal) {
			function add(signal, stream) { return addAbortSignal(signal, stream); }
			add.addAbortSignalNoValidate = add;
			return add;
		})`, "[gode:internal/streams/add-abort-signal]")
		if err == nil && v.IsFunction() {
			if r, cerr := v.Object().Call(js.Object{}, add); cerr == nil && r.Valid() {
				return r.Object()
			}
		}
		return e.JS.Obj()
	})
	env.Register("internal/streams/pipeline", func(e *env.Env) js.Object {
		pipeline := e.Exports("stream").Get("pipeline")
		v, err := e.JS.Eval(`(function (pipeline) {
			function run(streams, callback) { return pipeline.apply(null, streams.concat([callback])); }
			run.pipeline = pipeline;
			return run;
		})`, "[gode:internal/streams/pipeline]")
		if err == nil && v.IsFunction() {
			if r, cerr := v.Object().Call(js.Object{}, pipeline); cerr == nil && r.Valid() {
				return r.Object()
			}
		}
		return e.JS.Obj()
	})
	env.Register("stream/promises", func(e *env.Env) js.Object {
		if promises := e.Exports("stream").Get("promises"); promises.Valid() && promises.IsObject() {
			return promises.Object()
		}
		return e.JS.Obj()
	})
	// Node's lib/stream.js keeps the pre-streams3 constructors importable under
	// the legacy _stream_* specifiers; test-stream-aliases-legacy asserts every
	// alias is the same object as the public export.
	for _, alias := range []struct{ name, key string }{
		{"_stream_readable", "Readable"},
		{"_stream_writable", "Writable"},
		{"_stream_duplex", "Duplex"},
		{"_stream_transform", "Transform"},
		{"_stream_passthrough", "PassThrough"},
	} {
		alias := alias
		env.Register(alias.name, func(e *env.Env) js.Object {
			if v := e.Exports("stream").Get(alias.key); v.Valid() && v.IsObject() {
				return v.Object()
			}
			return e.JS.Obj()
		})
	}
}

type module struct {
	e          *env.Env
	defaultHWM int
	objectHWM  int
}

// Install builds the stream module's exports object.
// installWebBridges adds stream.fromWeb/toWeb on top of the Node Readable
// class. The runtime has no WhatWG ReadableStream global, so toWeb returns a
// stream-shaped object with the reader protocol rather than a branded
// ReadableStream; fromWeb accepts anything exposing getReader().
func installWebBridges(e *env.Env, exports js.Object) {
	v, err := e.JS.Eval(streamWebPrelude, "[gode:stream web]")
	if err != nil || !v.Valid() || !v.IsFunction() {
		return
	}
	_, _ = v.Object().Call(js.Object{}, exports.V())
}

const streamWebPrelude = `(function (exports) {
  'use strict';
  var Readable = exports.Readable;

  function fromWeb(webStream) {
    if (webStream === null || typeof webStream !== 'object' ||
        typeof webStream.getReader !== 'function') {
      throw new TypeError('The "readableStream" argument must be an instance of ReadableStream');
    }
    var reader = webStream.getReader();
    var out = new Readable();
    out._read = function () {
      var self = this;
      reader.read().then(function (res) {
        if (res.done) { self.push(null); return; }
        var value = res.value;
        self.push(value instanceof Uint8Array ? Buffer.from(value) : value);
      }, function (err) { self.destroy(err); });
    };
    out._destroy = function (err, cb) {
      try { reader.cancel(err); } catch (ignored) {}
      cb(err);
    };
    return out;
  }

  function toBytes(chunk) {
    return chunk instanceof Uint8Array ? chunk : Buffer.from(String(chunk));
  }

  function toWeb(nodeStream) {
    if (nodeStream === null || typeof nodeStream !== 'object' ||
        typeof nodeStream.on !== 'function') {
      throw new TypeError('The "stream" argument must be an instance of Stream');
    }
    var queue = [];
    var waiters = [];
    var ended = false;
    var failure = null;
    nodeStream.on('data', function (chunk) {
      if (waiters.length) {
        waiters.shift().resolve({ value: toBytes(chunk), done: false });
      } else {
        queue.push(chunk);
      }
    });
    nodeStream.on('end', function () {
      ended = true;
      while (waiters.length) {
        waiters.shift().resolve({ value: undefined, done: true });
      }
    });
    nodeStream.on('error', function (err) {
      failure = err;
      while (waiters.length) waiters.shift().reject(err);
    });
    var reader = {
      closed: Promise.resolve(),
      read: function read() {
        if (queue.length) return Promise.resolve({ value: toBytes(queue.shift()), done: false });
        if (failure) return Promise.reject(failure);
        if (ended) return Promise.resolve({ value: undefined, done: true });
        return new Promise(function (resolve, reject) { waiters.push({ resolve: resolve, reject: reject }); });
      },
      cancel: function cancel(reason) {
        try { nodeStream.destroy(reason); } catch (ignored) {}
        return Promise.resolve();
      },
      releaseLock: function releaseLock() {}
    };
    return {
      locked: false,
      getReader: function getReader() { return reader; },
      cancel: reader.cancel,
      pipeThrough: function pipeThrough(transform) { return transform.readable; },
      pipeTo: function pipeTo(dest) {
        var writer = dest.getWriter();
        function pump() {
          return reader.read().then(function (res) {
            if (res.done) { writer.close(); return undefined; }
            return writer.write(res.value).then(pump);
          });
        }
        return pump();
      },
      tee: function tee() { throw new Error('ReadableStream#tee is not supported'); }
    };
  }

  Object.defineProperty(exports, 'fromWeb', { value: fromWeb, writable: true, configurable: true });
  Object.defineProperty(exports, 'toWeb', { value: toWeb, writable: true, configurable: true });
  return exports;
})`

func Install(e *env.Env) js.Object {
	// Node uses 64 KiB for byte streams everywhere except Windows
	// (lib/internal/streams/state.js); object mode stays at 16.
	m := &module{e: e, defaultHWM: 64 * 1024, objectHWM: 16}

	helpers := e.JS.Obj()
	helpers.SetFunc("tick", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsFunction() {
			return e.JS.Undefined(), nil
		}
		e.NextTick(args[0], args[1:]...)
		return e.JS.Undefined(), nil
	})
	helpers.SetFunc("toBuffer", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		enc := "utf8"
		if a := arg(args, 1); a.Valid() && a.IsString() {
			enc = a.Str()
		}
		if v.Valid() && v.IsString() {
			data, err := encodeString(v.Str(), enc)
			if err != nil {
				return js.Value{}, err
			}
			return m.bufferValue(data), nil
		}
		return v, nil
	})
	helpers.SetFunc("decode", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		enc := "utf8"
		if a := arg(args, 1); a.Valid() && a.IsString() {
			enc = a.Str()
		}
		if data, ok := v.Bytes(); ok {
			return e.JS.Str(decodeString(data, enc)), nil
		}
		return e.JS.Str(v.Str()), nil
	})
	helpers.SetFunc("writeOut", func(this js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		if !v.Valid() || v.IsUndefined() || v.IsNull() {
			return e.JS.Bool(true), nil
		}
		if data, ok := v.Bytes(); ok {
			_, err := e.Out().Write(data)
			return e.JS.Bool(err == nil), nil
		}
		if v.IsString() {
			_, err := e.Out().Write([]byte(v.Str()))
			return e.JS.Bool(err == nil), nil
		}
		_, err := e.Out().Write([]byte(v.Str()))
		return e.JS.Bool(err == nil), nil
	})
	helpers.SetFunc("hwm", func(this js.Object, args []js.Value) (js.Value, error) {
		if arg(args, 0).Bool() {
			return e.JS.Num(float64(m.objectHWM)), nil
		}
		return e.JS.Num(float64(m.defaultHWM)), nil
	})
	helpers.SetFunc("setHWM", func(this js.Object, args []js.Value) (js.Value, error) {
		objectMode := arg(args, 0).Bool()
		n := arg(args, 1).Int()
		if n < 0 {
			return js.Value{}, js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
				`The value of "value" is out of range. It must be >= 0. Received `+arg(args, 1).Str())
		}
		if objectMode {
			m.objectHWM = n
		} else {
			m.defaultHWM = n
		}
		return e.JS.Undefined(), nil
	})

	var emitter js.Value
	if env.Registered("events") {
		emitter = e.Exports("events").Get("EventEmitter")
	}

	fn, err := e.JS.Eval(prelude, "stream.js")
	if err != nil {
		panic("godejs stream: " + err.Error())
	}
	args := []js.Value{helpers.V()}
	if emitter.Valid() {
		args = append(args, emitter)
	} else {
		args = append(args, e.JS.Undefined())
	}
	exports, callErr := fn.Object().Call(js.Object{}, args...)
	if callErr != nil {
		panic("godejs stream: " + callErr.Error())
	}
	installWebBridges(e, exports.Object())
	return callableStreamModule(e, exports.Object())
}

// callableStreamModule mirrors Node's lib/stream.js, where module.exports is
// the Stream constructor and every named export is a property on it. Tests do
// `const Stream = require('stream'); new Stream()`, which requires the module
// object itself to be callable.
func callableStreamModule(e *env.Env, exports js.Object) js.Object {
	ctor := exports.Get("Stream")
	if !ctor.Valid() || !ctor.IsFunction() {
		return exports
	}
	copier, err := e.JS.Eval(`(function (src, dst) {
		for (var k in src) {
			if (Object.prototype.hasOwnProperty.call(src, k)) {
				try { dst[k] = src[k]; } catch (ignored) {}
			}
		}
		return dst;
	})`, "[gode:stream exports]")
	if err != nil {
		return exports
	}
	r, cerr := copier.Object().Call(js.Object{}, exports.V(), ctor)
	if cerr != nil || !r.Valid() {
		return exports
	}
	return r.Object()
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"stream.addAbortSignal",
		"stream.compose",
		"stream.finished",
		"stream.from",
		"stream.fromWeb",
		"stream.toWeb",
		"stream.getDefaultHighWaterMark",
		"stream.isDestroyed",
		"stream.isDisturbed",
		"stream.isErrored",
		"stream.isReadable",
		"stream.isWritable",
		"stream.pipeline",
		"stream.push",
		"stream.read",
		"stream.setDefaultHighWaterMark",
	}
}

func arg(args []js.Value, i int) js.Value {
	if i >= len(args) {
		return js.Value{}
	}
	return args[i]
}

func (m *module) bufferValue(b []byte) js.Value {
	g := m.e.Global().Get("Buffer")
	if g.Valid() && g.IsFunction() {
		from := g.Object().Get("from")
		if from.IsFunction() {
			if v, err := from.Object().Call(g.Object(), m.e.JS.Bytes(b).V()); err == nil && v.Valid() {
				return v
			}
		}
	}
	return m.e.JS.Bytes(b).V()
}

func normalizeEncoding(enc string) string {
	switch strings.ToLower(enc) {
	case "", "utf8", "utf-8":
		return "utf8"
	case "utf16le", "utf-16le", "ucs2", "ucs-2":
		return "utf16le"
	case "latin1", "binary":
		return "latin1"
	case "ascii":
		return "ascii"
	case "base64":
		return "base64"
	case "base64url":
		return "base64url"
	case "hex":
		return "hex"
	}
	return "utf8"
}

func encodeString(s, enc string) ([]byte, error) {
	switch normalizeEncoding(enc) {
	case "utf16le":
		units := utf16.Encode([]rune(s))
		out := make([]byte, 0, len(units)*2)
		for _, u := range units {
			out = append(out, byte(u), byte(u>>8))
		}
		return out, nil
	case "latin1":
		out := make([]byte, 0, len(s))
		for _, r := range s {
			out = append(out, byte(r))
		}
		return out, nil
	case "ascii":
		out := make([]byte, 0, len(s))
		for _, r := range s {
			out = append(out, byte(r)&0x7f)
		}
		return out, nil
	case "base64":
		return base64.StdEncoding.DecodeString(s)
	case "base64url":
		return base64.RawURLEncoding.DecodeString(s)
	case "hex":
		if len(s)%2 == 1 {
			s = s[:len(s)-1]
		}
		return hex.DecodeString(s)
	}
	return []byte(s), nil
}

func decodeString(b []byte, enc string) string {
	switch normalizeEncoding(enc) {
	case "utf16le":
		units := make([]uint16, 0, len(b)/2)
		for i := 0; i+1 < len(b); i += 2 {
			units = append(units, uint16(b[i])|uint16(b[i+1])<<8)
		}
		return string(utf16.Decode(units))
	case "latin1":
		var sb strings.Builder
		for _, c := range b {
			sb.WriteRune(rune(c))
		}
		return sb.String()
	case "ascii":
		var sb strings.Builder
		for _, c := range b {
			sb.WriteRune(rune(c & 0x7f))
		}
		return sb.String()
	case "base64":
		return base64.StdEncoding.EncodeToString(b)
	case "base64url":
		return base64.RawURLEncoding.EncodeToString(b)
	case "hex":
		return hex.EncodeToString(b)
	}
	return string(b)
}

// preludeV1 is the first-generation implementation, kept only as reference
// for the behaviours the rewrite replaced. It is deliberately unused.
const preludeV1 = `(function (helpers, EventEmitter) {
  'use strict';

  var fallback = typeof EventEmitter !== 'function';
  if (fallback) {
    EventEmitter = function EventEmitter() {};
    EventEmitter.prototype.on = function (name, fn) {
      if (!this._events) this._events = Object.create(null);
      (this._events[name] || (this._events[name] = [])).push(fn);
      if (name === 'newListener') { /* no-op */ }
      return this;
    };
    EventEmitter.prototype.addListener = EventEmitter.prototype.on;
    EventEmitter.prototype.once = function (name, fn) {
      var self = this;
      function wrapped() {
        self.removeListener(name, wrapped);
        return fn.apply(self, arguments);
      }
      wrapped.listener = fn;
      return this.on(name, wrapped);
    };
    EventEmitter.prototype.off = function (name, fn) {
      return this.removeListener(name, fn);
    };
    EventEmitter.prototype.removeListener = function (name, fn) {
      var list = this._events && this._events[name];
      if (!list) return this;
      for (var i = list.length - 1; i >= 0; i--) {
        if (list[i] === fn || list[i].listener === fn) { list.splice(i, 1); break; }
      }
      if (!list.length) delete this._events[name];
      return this;
    };
    EventEmitter.prototype.removeAllListeners = function (name) {
      if (!this._events) return this;
      if (name === undefined) this._events = Object.create(null);
      else delete this._events[name];
      return this;
    };
    EventEmitter.prototype.emit = function (name) {
      var list = this._events && this._events[name];
      if (!list || !list.length) {
        if (name === 'error') {
          var err = arguments[1];
          if (err instanceof Error) throw err;
          throw new Error('Unhandled error. (' + err + ')');
        }
        return false;
      }
      var args = Array.prototype.slice.call(arguments, 1);
      var copy = list.slice();
      for (var i = 0; i < copy.length; i++) copy[i].apply(this, args);
      return true;
    };
    EventEmitter.prototype.listeners = function (name) {
      return (this._events && this._events[name] ? this._events[name].slice() : []);
    };
    EventEmitter.prototype.listenerCount = function (name) {
      var list = this._events && this._events[name];
      return list ? list.length : 0;
    };
    EventEmitter.prototype.eventNames = function () {
      return this._events ? Object.keys(this._events) : [];
    };
    EventEmitter.prototype.setMaxListeners = function (n) { this._maxListeners = n; return this; };
    EventEmitter.prototype.getMaxListeners = function () { return this._maxListeners || 10; };
  }

  function initEmitter(self) {
    if (fallback) self._events = Object.create(null);
    else EventEmitter.call(self);
  }

  function hwm(options) {
    options = options || {};
    if (typeof options.highWaterMark === 'number') return options.highWaterMark;
    return helpers.hwm(!!options.objectMode);
  }

  // ------------------------------------------------------------- Readable --
  function initReadable(self, options) {
    self._readable = true;
    self._objectMode = !!(options && options.objectMode);
    self._hwm = hwm(options);
    self._buffer = [];
    self._flowing = false;
    self._paused = false;
    self._ended = false;
    self._endEmitted = false;
    self._destroyed = false;
    self._encoding = null;
    self._disturbed = false;
    self._flushing = false;
    self._pipes = [];
  }

  function Readable(options) {
    if (!(this instanceof Readable)) return new Readable(options);
    initEmitter(this);
    initReadable(this, options);
  }
  Readable.prototype = Object.create(EventEmitter.prototype);
  Readable.prototype.constructor = Readable;
  Readable.prototype.push = function (chunk, encoding) {
    if (this._destroyed) return false;
    if (chunk === null) {
      this._ended = true;
      this.readableEnded = true;
      this._pump();
      return false;
    }
    var value = chunk;
    if (!this._objectMode) {
      value = helpers.toBuffer(chunk, typeof encoding === 'string' ? encoding : 'utf8');
    }
    this._buffer.push(value);
    var size = this._objectMode ? this._buffer.length : this._bufferSize();
    this._pump();
    return size < this._hwm;
  };
  Readable.prototype._bufferSize = function () {
    var n = 0;
    for (var i = 0; i < this._buffer.length; i++) {
      var b = this._buffer[i];
      n += this._objectMode ? 1 : (b && b.length ? b.length : 0);
    }
    return n;
  };
  Readable.prototype._pump = function () {
    if (this._flushing) return;
    this._flushing = true;
    try {
      while (this._buffer.length && this._flowing && !this._paused && !this._destroyed) {
        var chunk = this._buffer.shift();
        this._disturbed = true;
        if (this._encoding && !this._objectMode) {
          chunk = helpers.decode(chunk, this._encoding);
        }
        this.emit('data', chunk);
      }
      if (this._ended && this._buffer.length === 0 && !this._endEmitted && !this._paused && this._flowing) {
        this._endEmitted = true;
        this.emit('end');
        this.emit('close');
      }
    } finally {
      this._flushing = false;
    }
  };
  Readable.prototype.on = function (name, fn) {
    var out = EventEmitter.prototype.on.call(this, name, fn);
    if (name === 'data') {
      this._flowing = true;
      if (this._paused) this._paused = false;
      var self = this;
      helpers.tick(function () { self._pump(); });
    }
    return out;
  };
  Readable.prototype.addListener = Readable.prototype.on;
  Readable.prototype.once = function (name, fn) {
    var out = EventEmitter.prototype.once.call(this, name, fn);
    if (name === 'data') {
      this._flowing = true;
      var self = this;
      helpers.tick(function () { self._pump(); });
    }
    return out;
  };
  Readable.prototype.pipe = function (dest, options) {
    var self = this;
    if (dest.emit) dest.emit('pipe', this);
    this._pipes.push(dest);
    this.on('data', function (chunk) {
      if (!dest.write) return;
      if (dest.write(chunk) === false && self.pause) self.pause();
    });
    if (!options || options.end !== false) {
      this.on('end', function () { if (dest.end) dest.end(); });
    }
    this.on('error', function (err) { if (dest.emit) dest.emit('error', err); });
    return dest;
  };
  Readable.prototype.unpipe = function (dest) {
    this._pipes = this._pipes.filter(function (d) { return d !== dest; });
    return this;
  };
  Readable.prototype.pause = function () { this._paused = true; this._flowing = false; return this; };
  Readable.prototype.resume = function () {
    this._paused = false;
    this._flowing = true;
    var self = this;
    helpers.tick(function () { self._pump(); });
    return this;
  };
  Readable.prototype.isPaused = function () { return this._paused; };
  Readable.prototype.read = function (size) {
    this._disturbed = true;
    if (!this._buffer.length) return null;
    if (this._objectMode) return this._buffer.shift();
    if (size === undefined || size === null) {
      var all = this._buffer.splice(0, this._buffer.length);
      var total = 0, i;
      for (i = 0; i < all.length; i++) total += all[i].length;
      var out = helpers.toBuffer('', 'utf8');
      if (typeof Buffer !== 'undefined' && Buffer.concat) {
        out = Buffer.concat(all, total);
      } else {
        out = all[0];
      }
      return out;
    }
    var head = this._buffer[0];
    if (head.length <= size) { this._buffer.shift(); return head; }
    var part = head.slice(0, size);
    this._buffer[0] = head.slice(size);
    return part;
  };
  Readable.prototype.destroy = function (err) {
    if (this._destroyed) return this;
    this._destroyed = true;
    this.destroyed = true;
    if (err) { this._errored = true; this.emit('error', err); }
    var self = this;
    helpers.tick(function () { self.emit('close'); });
    return this;
  };
  Readable.prototype.setEncoding = function (enc) { this._encoding = enc; return this; };
  Readable.prototype[Symbol.asyncIterator] = function () {
    var self = this, queue = [], waiters = [], done = false;
    this.on('data', function (chunk) {
      if (waiters.length) waiters.shift()({ value: chunk, done: false });
      else queue.push(chunk);
    });
    this.on('end', function () {
      done = true;
      while (waiters.length) waiters.shift()({ value: undefined, done: true });
    });
    return {
      next: function () {
        if (queue.length) return Promise.resolve({ value: queue.shift(), done: false });
        if (done) return Promise.resolve({ value: undefined, done: true });
        return new Promise(function (resolve) { waiters.push(resolve); });
      }
    };
  };

  // ------------------------------------------------------------- Writable --
  function initWritable(self, options) {
    self._writable = true;
    self._woptions = options || {};
    self._hwmW = hwm(options);
    self._queue = [];
    self._writing = false;
    self._corked = 0;
    self._ending = false;
    self._writableEnded = false;
    self.writableFinished = false;
    self.destroyed = false;
    self._errored = false;
    self._writeFn = self._woptions.write;
  }

  function Writable(options) {
    if (!(this instanceof Writable)) return new Writable(options);
    initEmitter(this);
    initWritable(this, options);
  }
  Writable.prototype = Object.create(EventEmitter.prototype);
  Writable.prototype.constructor = Writable;
  Writable.prototype.write = function (chunk, encoding, cb) {
    if (typeof encoding === 'function') { cb = encoding; encoding = 'utf8'; }
    if (this._ending || this._writableEnded) {
      var err = new Error('write after end');
      err.code = 'ERR_STREAM_WRITE_AFTER_END';
      var self = this;
      helpers.tick(function () { self.emit('error', err); });
      return false;
    }
    this._queue.push([chunk, encoding || 'utf8', cb]);
    this._drainQueue();
    return this._queueLength() < this._hwmW;
  };
  Writable.prototype._queueLength = function () {
    var n = 0;
    for (var i = 0; i < this._queue.length; i++) n += this._queue[i][0] && this._queue[i][0].length ? this._queue[i][0].length : 1;
    return n;
  };
  Writable.prototype._drainQueue = function () {
    if (this._writing || this._corked > 0 || this.destroyed) return;
    var item = this._queue.shift();
    if (!item) {
      if (this._ending && !this.writableFinished) {
        this.writableFinished = true;
        this.writableEnded = true;
        var self = this, cb = this._endCb;
        helpers.tick(function () {
          self.emit('finish');
          if (cb) cb();
        });
      }
      return;
    }
    this._writing = true;
    var self = this;
    var done = function (err) {
      self._writing = false;
      if (err) { self._errored = true; self.emit('error', err); }
      else if (typeof item[2] === 'function') helpers.tick(item[2]);
      self._drainQueue();
    };
    if (typeof this._write === 'function') {
      this._write(item[0], item[1], done);
    } else {
      this._writeDefault(item[0], item[1], done);
    }
  };
  Writable.prototype._writeDefault = function (chunk, encoding, cb) {
    if (typeof this._writeFn === 'function') return this._writeFn(chunk, encoding, cb);
    helpers.writeOut(chunk);
    helpers.tick(cb);
  };
  Writable.prototype.end = function (chunk, encoding, cb) {
    if (typeof chunk === 'function') { cb = chunk; chunk = null; }
    if (typeof encoding === 'function') { cb = encoding; encoding = 'utf8'; }
    if (chunk !== null && chunk !== undefined) this.write(chunk, encoding);
    this._ending = true;
    this._writableEnded = true;
    this.writableEnded = true;
    this._endCb = cb;
    this._drainQueue();
    return this;
  };
  Writable.prototype.cork = function () { this._corked++; return this; };
  Writable.prototype.uncork = function () {
    if (this._corked > 0) this._corked--;
    this._drainQueue();
    return this;
  };
  Writable.prototype.destroy = function (err) {
    if (this.destroyed) return this;
    this.destroyed = true;
    this._destroyed = true;
    if (err) { this._errored = true; this.emit('error', err); }
    return this;
  };
  Writable.prototype.setDefaultEncoding = function () { return this; };

  function copyWritableProto(target) {
    Object.keys(Writable.prototype).forEach(function (key) {
      if (key !== 'constructor') target[key] = Writable.prototype[key];
    });
  }

  // --------------------------------------------------------------- Duplex --
  function Duplex(options) {
    if (!(this instanceof Duplex)) return new Duplex(options);
    initEmitter(this);
    initReadable(this, options);
    initWritable(this, options);
  }
  Duplex.prototype = Object.create(Readable.prototype);
  Duplex.prototype.constructor = Duplex;
  copyWritableProto(Duplex.prototype);

  // ------------------------------------------------------------ Transform --
  function Transform(options) {
    if (!(this instanceof Transform)) return new Transform(options);
    Duplex.call(this, options);
    if (options && typeof options.transform === 'function') this._transform = options.transform;
    if (options && typeof options.flush === 'function') this._flush = options.flush;
  }
  Transform.prototype = Object.create(Duplex.prototype);
  Transform.prototype.constructor = Transform;
  Transform.prototype._transform = function (chunk, encoding, cb) {
    this.push(chunk);
    cb();
  };
  Transform.prototype._write = function (chunk, encoding, cb) {
    var self = this;
    this._transform(chunk, encoding, function (err, out) {
      if (out !== undefined && out !== null) self.push(out);
      cb(err);
    });
  };
  Transform.prototype.end = function (chunk, encoding, cb) {
    if (typeof chunk === 'function') { cb = chunk; chunk = null; }
    if (typeof encoding === 'function') { cb = encoding; encoding = 'utf8'; }
    var self = this;
    Writable.prototype.end.call(this, chunk, encoding, function () {
      var finish = function () {
        self.push(null);
        if (cb) cb();
      };
      if (typeof self._flush === 'function') self._flush(finish);
      else finish();
    });
    return this;
  };

  // ---------------------------------------------------------- PassThrough --
  function PassThrough(options) {
    if (!(this instanceof PassThrough)) return new PassThrough(options);
    Transform.call(this, options);
  }
  PassThrough.prototype = Object.create(Transform.prototype);
  PassThrough.prototype.constructor = PassThrough;
  PassThrough.prototype._transform = function (chunk, encoding, cb) {
    this.push(chunk);
    cb();
  };

  // ------------------------------------------------------------ functions --
  function pipeline() {
    var args = Array.prototype.slice.call(arguments);
    var cb = typeof args[args.length - 1] === 'function' ? args.pop() : null;
    var done = false;
    function finish(err) {
      if (done) return;
      done = true;
      if (cb) helpers.tick(function () { cb(err || null); });
    }
    for (var i = 0; i < args.length - 1; i++) {
      if (args[i] && typeof args[i].pipe === 'function') args[i].pipe(args[i + 1]);
    }
    for (var j = 0; j < args.length; j++) {
      if (args[j] && typeof args[j].on === 'function') {
        args[j].on('error', function (err) { finish(err); });
      }
    }
    var last = args[args.length - 1];
    if (last && typeof last.on === 'function') {
      last.on('finish', function () { finish(null); });
      last.on('end', function () { finish(null); });
    } else {
      finish(null);
    }
    return last;
  }

  function pipelinePromise() {
    var args = Array.prototype.slice.call(arguments);
    return new Promise(function (resolve, reject) {
      args.push(function (err) { if (err) reject(err); else resolve(); });
      pipeline.apply(null, args);
    });
  }

  function finishedPromise(stream, options) {
    return new Promise(function (resolve, reject) {
      finished(stream, options, function (err) { if (err) reject(err); else resolve(); });
    });
  }

  var promises = { pipeline: pipelinePromise, finished: finishedPromise };

  function finished(stream, options, cb) {
    if (typeof options === 'function') { cb = options; options = {}; }
    options = options || {};
    cb = typeof cb === 'function' ? cb : function () {};
    var done = false;
    function fire(err) {
      if (done) return;
      done = true;
      helpers.tick(function () { cb(err || null); });
    }
    if (!stream || typeof stream.on !== 'function') { fire(null); return function () {}; }
    stream.on('end', function () { fire(null); });
    stream.on('finish', function () { fire(null); });
    stream.on('error', function (err) { fire(err); });
    if (options.cleanup !== false) stream.on('close', function () { fire(null); });
    return function () { done = true; };
  }

  function isReadable(s) {
    if (!s || typeof s.read !== 'function') return false;
    if (s.destroyed) return false;
    if (s.readableEnded) return false;
    return s.readable !== false;
  }
  function isWritable(s) {
    if (!s || typeof s.write !== 'function') return false;
    if (s.destroyed) return false;
    if (s.writableFinished || s.writableEnded) return false;
    return s.writable !== false;
  }
  function isDestroyed(s) { return !!(s && s.destroyed); }
  function isDisturbed(s) { return !!(s && (s._disturbed || s.readableDidRead)); }
  function isErrored(s) { return !!(s && (s._errored || s.errored)); }
  function getDefaultHighWaterMark(objectMode) { return helpers.hwm(!!objectMode); }
  function setDefaultHighWaterMark(objectMode, value) { helpers.setHWM(!!objectMode, value); }
  function addAbortSignal(signal, stream) {
    if (!signal) return stream;
    function abort() {
      var err = signal.reason;
      if (!(err instanceof Error)) {
        err = new Error('The operation was aborted');
        err.name = 'AbortError';
        err.code = 'ABORT_ERR';
      }
      stream.destroy(err);
    }
    if (signal.aborted) abort();
    else if (typeof signal.addEventListener === 'function') signal.addEventListener('abort', abort);
    return stream;
  }
  function from(iterable) {
    var r = new Readable({ objectMode: true });
    if (iterable && typeof iterable[Symbol.asyncIterator] === 'function') {
      var it = iterable[Symbol.asyncIterator]();
      function step() {
        it.next().then(function (res) {
          if (res.done) { r.push(null); return; }
          r.push(res.value);
          step();
        }, function (err) { r.destroy(err); });
      }
      step();
    } else {
      var list = [];
      if (iterable && typeof iterable[Symbol.iterator] === 'function') {
        for (var v of iterable) list.push(v);
      }
      for (var i = 0; i < list.length; i++) r.push(list[i]);
      r.push(null);
    }
    return r;
  }
  function compose() {
    var args = Array.prototype.slice.call(arguments);
    for (var i = 0; i < args.length - 1; i++) {
      if (args[i] && typeof args[i].pipe === 'function') args[i].pipe(args[i + 1]);
    }
    return args[0];
  }
  function push(chunk, encoding) {
    if (this && typeof this.push === 'function' && this.push !== push) return this.push(chunk, encoding);
    return false;
  }
  function read(size) {
    if (this && typeof this.read === 'function' && this.read !== read) return this.read(size);
    return null;
  }

  Readable.from = from;
  Duplex.from = from;

  return {
    Readable: Readable,
    Writable: Writable,
    Duplex: Duplex,
    Transform: Transform,
    PassThrough: PassThrough,
    Stream: Readable,
    pipeline: pipeline,
    finished: finished,
    isReadable: isReadable,
    isWritable: isWritable,
    isDestroyed: isDestroyed,
    isDisturbed: isDisturbed,
    isErrored: isErrored,
    getDefaultHighWaterMark: getDefaultHighWaterMark,
    setDefaultHighWaterMark: setDefaultHighWaterMark,
    addAbortSignal: addAbortSignal,
    from: from,
    compose: compose,
    promises: promises,
    push: push,
    read: read
  };
})`

// prelude is the stream implementation. It follows the shape of Node's
// lib/internal/streams/{readable,writable,duplex,transform,destroy}.js closely
// enough that the public state machines (readableLength, flowing, _read,
// _write, _final, destroy/_destroy, autoDestroy) behave the same way, while
// staying small enough to audit in one place.
const prelude = `(function (helpers, EventEmitter) {
  'use strict';

  var hasEmitter = typeof EventEmitter === 'function';
  var tick = function (fn, arg) {
    if (typeof fn === 'function') helpers.tick(fn, arg);
  };

  if (!hasEmitter) {
    EventEmitter = function EventEmitter() { this._events = Object.create(null); };
    EventEmitter.prototype.on = function (n, fn) {
      (this._events[n] || (this._events[n] = [])).push(fn);
      return this;
    };
    EventEmitter.prototype.addListener = EventEmitter.prototype.on;
    EventEmitter.prototype.prependListener = function (n, fn) {
      (this._events[n] || (this._events[n] = [])).unshift(fn);
      return this;
    };
    EventEmitter.prototype.once = function (n, fn) {
      var self = this;
      function wrapped() { self.removeListener(n, wrapped); return fn.apply(self, arguments); }
      wrapped.listener = fn;
      return this.on(n, wrapped);
    };
    EventEmitter.prototype.prependOnceListener = function (n, fn) {
      var self = this;
      function wrapped() { self.removeListener(n, wrapped); return fn.apply(self, arguments); }
      wrapped.listener = fn;
      return this.prependListener(n, wrapped);
    };
    EventEmitter.prototype.removeListener = function (n, fn) {
      var list = this._events[n];
      if (!list) return this;
      for (var i = list.length - 1; i >= 0; i--) {
        if (list[i] === fn || list[i].listener === fn) { list.splice(i, 1); break; }
      }
      if (!list.length) delete this._events[n];
      return this;
    };
    EventEmitter.prototype.off = EventEmitter.prototype.removeListener;
    EventEmitter.prototype.removeAllListeners = function (n) {
      if (n === undefined) this._events = Object.create(null);
      else delete this._events[n];
      return this;
    };
    EventEmitter.prototype.emit = function (n) {
      var list = this._events[n];
      if (!list || !list.length) {
        if (n === 'error') {
          var err = arguments[1];
          throw err instanceof Error ? err : new Error('Unhandled error. (' + err + ')');
        }
        return false;
      }
      var args = Array.prototype.slice.call(arguments, 1);
      var copy = list.slice();
      for (var i = 0; i < copy.length; i++) copy[i].apply(this, args);
      return true;
    };
    EventEmitter.prototype.listeners = function (n) {
      return this._events[n] ? this._events[n].slice() : [];
    };
    EventEmitter.prototype.listenerCount = function (n) {
      return this._events[n] ? this._events[n].length : 0;
    };
    EventEmitter.prototype.eventNames = function () { return Object.keys(this._events); };
    EventEmitter.prototype.setMaxListeners = function (n) { this._maxListeners = n; return this; };
    EventEmitter.prototype.getMaxListeners = function () { return this._maxListeners || 10; };
  }

  // ------------------------------------------------------------ utilities --
  function coded(ctor, code, message) {
    var e = new (ctor || Error)(message);
    e.code = code;
    return e;
  }
  function errMethodNotImplemented(name) {
    return coded(Error, 'ERR_METHOD_NOT_IMPLEMENTED', 'The ' + name + ' method is not implemented');
  }
  function errInvalidArgType(name, expected, actual) {
    return coded(TypeError, 'ERR_INVALID_ARG_TYPE',
      'The "' + name + '" argument must be ' + expected + '. Received ' + describe(actual));
  }
  function describe(v) {
    if (v === null) return 'null';
    if (v === undefined) return 'undefined';
    var t = typeof v;
    if (t === 'string') return "type string ('" + v + "')";
    if (t === 'number' || t === 'boolean' || t === 'bigint' || t === 'symbol') return 'type ' + t + ' (' + String(v) + ')';
    if (t === 'function') return 'function ' + (v.name || '(anonymous)');
    return 'an instance of ' + ((v.constructor && v.constructor.name) || 'Object');
  }
  function abortError(reason) {
    if (reason instanceof Error) return reason;
    var e = coded(Error, 'ABORT_ERR', 'The operation was aborted');
    e.name = 'AbortError';
    if (reason !== undefined) e.cause = reason;
    return e;
  }
  function isUint8Array(v) {
    return typeof Uint8Array === 'function' && v instanceof Uint8Array;
  }
  function isArrayBufferView(v) {
    return typeof ArrayBuffer === 'function' && typeof ArrayBuffer.isView === 'function' && ArrayBuffer.isView(v);
  }
  function toBytes(v) {
    if (v instanceof Buffer) return v;
    var u = v instanceof Uint8Array ? v : new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
    if (typeof Buffer !== 'undefined' && typeof Buffer.from === 'function') return Buffer.from(u);
    return u;
  }
  // makeDecoder mirrors string_decoder's incremental UTF-8/UTF-16LE decoding:
  // a multibyte character split across two pushes must survive as one glyph
  // instead of becoming replacement characters.
  function makeDecoder(enc) {
    var raw = enc == null ? 'utf8' : String(enc);
    var name = raw.toLowerCase();
    if (typeof Buffer === 'undefined' || typeof Buffer.isEncoding !== 'function' || !Buffer.isEncoding(name)) {
      throw coded(TypeError, 'ERR_UNKNOWN_ENCODING', 'Unknown encoding: ' + raw);
    }
    var canonical = name;
    if (canonical === 'utf-8') canonical = 'utf8';
    else if (canonical === 'utf-16le' || canonical === 'ucs2' || canonical === 'ucs-2') canonical = 'utf16le';
    else if (canonical === 'binary') canonical = 'latin1';
    var pending = null;
    function join(a, b) {
      var out = new Uint8Array(a.length + b.length);
      out.set(a, 0);
      out.set(b, a.length);
      return out;
    }
    function completeUTF8(buf) {
      var i = 0, n = buf.length;
      while (i < n) {
        var b = buf[i], len;
        if (b < 0x80) len = 1;
        else if (b >= 0xc2 && b <= 0xdf) len = 2;
        else if (b >= 0xe0 && b <= 0xef) len = 3;
        else if (b >= 0xf0 && b <= 0xf4) len = 4;
        else { i += 1; continue; }
        if (i + len > n) break;
        i += len;
      }
      return i;
    }
    var decoder = { encoding: canonical };
    decoder.write = function (chunk) {
      if (typeof chunk === 'string') return chunk;
      var buf = toBytes(chunk);
      if (canonical === 'utf16le') {
        if (pending && pending.length) buf = join(pending, buf);
        var keep = buf.length % 2;
        pending = keep ? buf.slice(buf.length - 1) : null;
        var even = keep ? buf.slice(0, buf.length - 1) : buf;
        return helpers.decode(even, 'utf16le');
      }
      if (canonical !== 'utf8') return helpers.decode(buf, canonical);
      if (pending && pending.length) buf = join(pending, buf);
      var take = completeUTF8(buf);
      pending = take < buf.length ? buf.slice(take) : null;
      return take > 0 ? helpers.decode(buf.slice(0, take), 'utf8') : '';
    };
    decoder.end = function () {
      if (!pending || pending.length === 0) return '';
      var out = helpers.decode(pending, canonical);
      pending = null;
      return out;
    };
    return decoder;
  }
  function concat(list, total) {
    if (typeof Buffer !== 'undefined' && typeof Buffer.concat === 'function') return Buffer.concat(list, total);
    var out = new Uint8Array(total);
    var off = 0;
    for (var i = 0; i < list.length; i++) { out.set(list[i], off); off += list[i].length; }
    return out;
  }
  function emptyBuffer() { return helpers.toBuffer('', 'utf8'); }

  function Stream(options) {
    if (!(this instanceof Stream)) return new Stream(options);
    if (hasEmitter) EventEmitter.call(this);
    else this._events = Object.create(null);
  }
  Stream.prototype = Object.create(hasEmitter ? EventEmitter.prototype : Object.prototype);
  Stream.prototype.constructor = Stream;
  // Node aliases off and removeListener to the same function object.
  if (hasEmitter && EventEmitter.prototype.off && EventEmitter.prototype.removeListener &&
      EventEmitter.prototype.off !== EventEmitter.prototype.removeListener) {
    Stream.prototype.off = EventEmitter.prototype.removeListener;
  }

  // legacy.js: Stream.prototype.pipe is the pre-streams3 implementation. It
  // works on any emitter with pause/resume, which is what hand-written
  // Readable/Writable constructors from the 0.8 era rely on, and its listener
  // layout is observable (test-stream-pipe-cleanup counts it exactly).
  Stream.prototype.pipe = function (dest, options) {
    var source = this;

    function ondata(chunk) {
      if (dest.writable && dest.write(chunk) === false && source.pause) {
        source.pause();
      }
    }
    source.on('data', ondata);

    function ondrain() {
      if (source.readable && source.resume) {
        source.resume();
      }
    }
    dest.on('drain', ondrain);

    // If the 'end' option is not supplied, dest.end() will be called when the
    // source gets 'end' or 'close'. Only dest.end() once.
    if (!dest._isStdio && (!options || options.end !== false)) {
      source.on('end', onend);
      source.on('close', onclose);
    }

    var didOnEnd = false;
    function onend() {
      if (didOnEnd) return;
      didOnEnd = true;
      dest.end();
    }

    function onclose() {
      if (didOnEnd) return;
      didOnEnd = true;
      if (typeof dest.destroy === 'function') dest.destroy();
    }

    // Don't leave dangling pipes when there are errors.
    function onerror(er) {
      cleanup();
      var count = typeof this.listenerCount === 'function'
        ? this.listenerCount('error')
        : ((this._events && this._events.error) ? this._events.error.length || 1 : 0);
      if (count === 0) {
        throw er; // Unhandled stream error in pipe.
      }
    }

    source.on('error', onerror);
    dest.on('error', onerror);

    // Remove all the event listeners that were added.
    function cleanup() {
      source.removeListener('data', ondata);
      dest.removeListener('drain', ondrain);
      source.removeListener('end', onend);
      source.removeListener('close', onclose);
      source.removeListener('error', onerror);
      dest.removeListener('error', onerror);
      source.removeListener('end', cleanup);
      source.removeListener('close', cleanup);
      dest.removeListener('close', cleanup);
      dest.emit('unpipe', source);
    }

    // If the user uses functions like unpipe() or removeListener(), the
    // listeners will be removed.
    source.on('end', cleanup);
    source.on('close', cleanup);
    dest.on('close', cleanup);

    dest.emit('pipe', source);

    // Allow for unix-like usage: A.pipe(B).pipe(C)
    return dest;
  };

  function defineAccessor(name, getter, setter) {
    Object.defineProperty(Stream.prototype, name, {
      configurable: true, enumerable: false, get: getter, set: setter || function () {}
    });
  }

  function stateOf(stream) { return stream._readableState || stream._writableState || null; }
  function anyState(stream, name) {
    var r = stream._readableState, w = stream._writableState;
    if (r && r[name] !== undefined) return r[name];
    if (w && w[name] !== undefined) return w[name];
    return undefined;
  }

  defineAccessor('destroyed',
    function () {
      var s = stateOf(this);
      if (s) return !!s.destroyed;
      return this._destroyed === true;
    },
    function (v) {
      if (this._readableState) this._readableState.destroyed = v;
      if (this._writableState) this._writableState.destroyed = v;
      this._destroyed = v;
    });
  defineAccessor('closed', function () {
    var s = stateOf(this);
    return s ? !!s.closed : false;
  });
  defineAccessor('errored', function () {
    var s = stateOf(this);
    return s ? s.errored || null : null;
  });
  defineAccessor('readable', function () {
    var r = this._readableState;
    if (!r) return this._legacyReadable === true;
    if (r.readable === false) return false;
    return !r.destroyed && !r.ended;
  }, function (v) { this._legacyReadable = v; });
  defineAccessor('readableEnded', function () {
    var r = this._readableState;
    return !!r && r.endEmitted;
  });
  defineAccessor('readableFlowing', function () {
    var r = this._readableState;
    return r ? r.flowing : null;
  });
  defineAccessor('readableLength', function () {
    var r = this._readableState;
    return r ? r.length : 0;
  });
  defineAccessor('readableHighWaterMark', function () {
    var r = this._readableState;
    return r ? r.highWaterMark : 0;
  });
  defineAccessor('readableObjectMode', function () {
    var r = this._readableState;
    return !!r && r.objectMode;
  });
  defineAccessor('readableDidRead', function () {
    var r = this._readableState;
    return !!r && r.didRead;
  });
  defineAccessor('readableAborted', function () {
    var r = this._readableState;
    return !!r && r.destroyed && !r.endEmitted;
  });
  defineAccessor('writable', function () {
    var w = this._writableState;
    if (!w) return this._legacyWritable === true;
    if (w.writable === false) return false;
    return !w.destroyed && !w.ending && !w.finished;
  }, function (v) { this._legacyWritable = v; });
  defineAccessor('writableEnded', function () {
    var w = this._writableState;
    return !!w && w.ending;
  });
  defineAccessor('writableFinished', function () {
    var w = this._writableState;
    return !!w && w.finished;
  });
  defineAccessor('writableLength', function () {
    var w = this._writableState;
    return w ? w.length : 0;
  });
  defineAccessor('writableHighWaterMark', function () {
    var w = this._writableState;
    return w ? w.highWaterMark : 0;
  });
  defineAccessor('writableNeedDrain', function () {
    var w = this._writableState;
    return !!w && w.needDrain;
  });
  defineAccessor('writableCorked', function () {
    var w = this._writableState;
    return w ? w.corked : 0;
  });
  defineAccessor('writableObjectMode', function () {
    var w = this._writableState;
    return !!w && w.objectMode;
  });
  defineAccessor('readableEncoding', function () {
    var r = this._readableState;
    return r ? r.encoding : null;
  });
  defineAccessor('readableBuffer', function () {
    var r = this._readableState;
    return r ? r.buffer : null;
  });
  defineAccessor('writableBuffer', function () {
    var w = this._writableState;
    return w ? w.buffer : null;
  });

  // -------------------------------------------------------------- destroy --
  // destroyImpl mirrors lib/internal/streams/destroy.js: mark the state
  // synchronously, run _destroy (or the option hook), then emit error/close on
  // a later tick so listeners registered after destroy() still see them.
  function destroyImpl(stream, err, cb) {
    var fresh = !stream.destroyed;
    if (!fresh) {
      if (typeof cb === 'function') tick(cb, undefined);
      return stream;
    }
    var r = stream._readableState;
    var w = stream._writableState;
    var emitClose = true;
    if (r) { r.destroyed = true; r.errored = err || null; r.errorEmitted = false; emitClose = r.emitClose; }
    if (w) { w.destroyed = true; w.errored = err || null; w.errorEmitted = false; emitClose = w.emitClose; }
    stream._destroyed = true;

    var done = false;
    function finish(err2) {
      if (done) return;
      done = true;
      if (err2) {
        if (r) { r.errored = err2; r.errorEmitted = true; }
        if (w) { w.errored = err2; w.errorEmitted = true; }
        stream.emit('error', err2);
      }
      if (r) r.closed = true;
      if (w) w.closed = true;
      if (emitClose) stream.emit('close');
      // Pending end() callbacks belong to the writer; report the failure the
      // same way Node 24 does rather than dropping the callback.
      if (w && typeof w.onDestroyed === 'function') {
        var pending = w.onDestroyed;
        w.onDestroyed = null;
        pending(err2 || (w.finished ? null : coded(Error, 'ERR_STREAM_DESTROYED', 'Cannot call write after a stream was destroyed')));
      }
      if (typeof cb === 'function') cb(err2);
    }

    tick(function () {
      var hook = stream._destroy;
      if (typeof hook !== 'function') { finish(err || null); return; }
      var called = false;
      var once = function (err2) {
        if (called) { finish(coded(Error, 'ERR_MULTIPLE_CALLBACK', 'Callback called multiple times')); return; }
        called = true;
        finish(err2 || null);
      };
      try {
        var ret = hook.call(stream, err || null, once);
        if (ret && typeof ret.then === 'function') ret.then(function () {}, once);
      } catch (e) { once(e); }
    });
    return stream;
  }

  function defaultDestroy(err, cb) { cb(err || null); }

  function errorOrDestroy(stream, err) {
    var s = stateOf(stream);
    if (s && s.autoDestroy) {
      if (s.destroyed) return;
      destroyImpl(stream, err, null);
      return;
    }
    if (s) s.errored = err;
    tick(function () {
      if (s) s.errorEmitted = true;
      stream.emit('error', err);
    });
  }

  // getHighWaterMark mirrors lib/internal/streams/state.js: an explicit value
  // must be a non-negative integer, and duplexes prefer
  // options.readableHighWaterMark/options.writableHighWaterMark.
  function getHighWaterMark(state, options, duplexKey, isDuplex) {
    var hwm = options.highWaterMark != null ? options.highWaterMark
      : (isDuplex ? options[duplexKey] : null);
    if (hwm != null) {
      if (typeof hwm !== 'number' || hwm !== Math.floor(hwm) || hwm < 0) {
        throw coded(TypeError, 'ERR_INVALID_ARG_VALUE',
          "The property '" + (isDuplex ? 'options.' + duplexKey : 'options.highWaterMark') +
          "' is invalid. Received " + String(hwm));
      }
      return Math.floor(hwm);
    }
    return helpers.hwm(state.objectMode);
  }

  // --------------------------------------------------------------- Readable --
  function ReadableState(options, stream, isDuplex) {
    options = options || {};
    this.objectMode = !!(options.objectMode || (isDuplex && options.readableObjectMode));
    this.highWaterMark = getHighWaterMark(this, options, 'readableHighWaterMark', isDuplex);
    this.buffer = [];
    this.length = 0;
    this.pipes = [];
    this.flowing = null;
    this.ended = false;
    this.endEmitted = false;
    this.reading = false;
    // sync is true only while a synchronous _read() call is in flight, so a
    // push from an async producer starts flowing immediately.
    this.sync = false;
    this.needReadable = false;
    this.emittedReadable = false;
    this.readableListening = false;
    this.didRead = false;
    this.destroyed = false;
    this.errored = null;
    this.errorEmitted = false;
    this.closed = false;
    this.emitClose = options.emitClose !== false;
    this.autoDestroy = options.autoDestroy !== false;
    this.decoder = null;
    this.encoding = null;
    if (options.encoding) {
      this.decoder = makeDecoder(options.encoding);
      this.encoding = this.decoder.encoding;
    }
    this.constructed = false;
    this.closeEmitted = false;
  }

  function initReadable(self, options, isDuplex) {
    options = options || {};
    var state = new ReadableState(options, self, isDuplex);
    self._readableState = state;
    // "readable" is an accessor on Stream.prototype; legacy subclasses that set
    // it directly store the override there instead of shadowing the getter.
    state.readable = options.readable !== false;
    if (options.readable === false) self._legacyReadable = false;
    if (typeof options.read === 'function') self._read = options.read;
    if (typeof options.destroy === 'function') self._destroy = options.destroy;
    state.constructed = typeof options.construct !== 'function';
    if (options.signal) addAbortSignal(options.signal, self);
    return state;
  }

  function construct(stream, fn, writableState) {
    var called = false;
    var finished = false;
    function cb(err) {
      if (called) { errorOrDestroy(stream, coded(Error, 'ERR_MULTIPLE_CALLBACK', 'Callback called multiple times')); return; }
      called = true;
      if (stream._readableState) stream._readableState.constructed = true;
      if (stream._writableState) stream._writableState.constructed = true;
      if (err) { errorOrDestroy(stream, err); return; }
      finished = true;
      if (stream._readableState && stream._readableState.needReadable) maybeReadMore(stream);
    }
    try { fn.call(stream, cb); } catch (e) { cb(e); }
    return finished;
  }

  function Readable(options) {
    if (!(this instanceof Readable)) return new Readable(options);
    Stream.call(this, options);
    initReadable(this, options, false);
    if (options && typeof options.construct === 'function') construct(this, options.construct, null);
  }
  Readable.prototype = Object.create(Stream.prototype);
  Readable.prototype.constructor = Readable;

  Readable.prototype._read = function () {
    errorOrDestroy(this, errMethodNotImplemented('_read()'));
  };
  Readable.prototype._destroy = defaultDestroy;

  function readableAddChunk(stream, chunk, encoding, addToFront) {
    var state = stream._readableState;
    if (chunk === null) {
      state.reading = false;
      onEofChunk(stream, state);
      return false;
    }
    // Node validates the chunk first: an invalid type reports
    // ERR_INVALID_ARG_TYPE asynchronously (errorOrDestroy), while undefined is
    // accepted as an empty push in byte mode.
    var value = chunk;
    if (!state.objectMode) {
      if (typeof chunk === 'string') {
        // A string already in the stream's encoding is decoded text; anything
        // else is bytes in that encoding and goes through the decoder.
        var enc = typeof encoding === 'string' && encoding !== 'buffer' ? encoding : 'utf8';
        if (state.encoding === enc) {
          value = chunk;
        } else {
          value = helpers.toBuffer(chunk, enc);
          if (!addToFront && state.decoder) value = state.decoder.write(value);
        }
      } else if (chunk === undefined) {
        value = undefined;
      } else if (!isArrayBufferView(chunk)) {
        errorOrDestroy(stream, errInvalidArgType('chunk', 'of type string or an instance of Buffer, TypedArray, or DataView', chunk));
        return false;
      } else {
        value = toBytes(chunk);
        if (!addToFront && state.decoder) value = state.decoder.write(value);
      }
    } else if (state.decoder && typeof chunk !== 'string') {
      value = state.decoder.write(toBytes(chunk));
    }
    var len = state.objectMode ? 1 : (value === undefined ? 0 : value.length);
    if (len === 0) {
      state.reading = false;
      return !state.ended && (state.length < state.highWaterMark || state.length === 0);
    }
    // Node only rejects unshift() once the 'end' event has been emitted;
    // before that the chunk is queued ahead of the pending EOF. push() is
    // rejected as soon as the stream has ended.
    if (addToFront) {
      if (state.endEmitted) {
        errorOrDestroy(stream, coded(Error, 'ERR_STREAM_UNSHIFT_AFTER_END_EVENT',
          'stream.unshift() after end event'));
        return false;
      }
      if (state.destroyed || state.errored) return false;
    } else {
      if (state.ended) {
        errorOrDestroy(stream, coded(Error, 'ERR_STREAM_PUSH_AFTER_EOF',
          'stream.push() after EOF'));
        return false;
      }
      if (state.destroyed || state.errored) return false;
    }
    state.reading = false;
    if (addToFront) state.buffer.unshift(value);
    else state.buffer.push(value);
    state.length += len;

    if (state.needReadable && !state.emittedReadable) emitReadable(stream);
    if (state.flowing === true && !state.sync) flow(stream);
    return !state.ended && (state.length < state.highWaterMark || state.length === 0);
  }

  function onEofChunk(stream, state) {
    if (state.ended) return;
    state.ended = true;
    if (state.length === 0) endReadable(stream);
    else emitReadable(stream);
  }

  function emitReadable(stream) {
    var state = stream._readableState;
    if (!state || state.emittedReadable || state.destroyed) return;
    state.emittedReadable = true;
    state.needReadable = false;
    tick(function () {
      state.emittedReadable = false;
      if (state.destroyed) return;
      if (state.length > 0 || state.ended) stream.emit('readable');
    });
  }

  function endReadable(stream) {
    var state = stream._readableState;
    if (!state || state.endEmitted || state.destroyed) return;
    if (state.length > 0) return;
    tick(function () {
      if (state.endEmitted || state.destroyed) return;
      state.endEmitted = true;
      stream.emit('end');
      var w = stream._writableState;
      if (state.autoDestroy && (!w || w.finished)) destroyImpl(stream, null, null);
    });
  }

  function flow(stream) {
    var state = stream._readableState;
    if (!state || state.destroyed) return;
    while (state.flowing === true) {
      var chunk = stream.read();
      if (chunk === null) break;
      stream.emit('data', chunk);
      if (state.destroyed) break;
    }
  }

  function maybeReadMore(stream) {
    var state = stream._readableState;
    if (!state || state.reading || state.ended || state.destroyed || !state.constructed) return;
    if (state.length >= state.highWaterMark) return;
    if (!state.didRead && state.flowing !== true && !state.needReadable && !state.readableListening) return;
    state.reading = true;
    try {
      stream._read(state.highWaterMark);
    } catch (e) {
      state.reading = false;
      errorOrDestroy(stream, e);
    }
  }

  Readable.prototype.push = function (chunk, encoding) {
    return readableAddChunk(this, chunk, encoding, false);
  };
  Readable.prototype.unshift = function (chunk, encoding) {
    return readableAddChunk(this, chunk, encoding, true);
  };

  // Reading more than the high-water mark grows it to the next power of two,
  // so a single big read() does not stay permanently above the limit.
  function computeNewHighWaterMark(n) {
    if (n >= 0x40000000) {
      n = 0x40000000;
    } else {
      n -= 1;
      n |= n >>> 1;
      n |= n >>> 2;
      n |= n >>> 4;
      n |= n >>> 8;
      n |= n >>> 16;
      n += 1;
    }
    return n;
  }

  Readable.prototype.read = function (n) {
    var state = this._readableState;
    var num = n === undefined ? NaN : Number(n);
    if (num !== 0) state.didRead = true;
    if (state.destroyed) return null;
    if (!isNaN(num) && num > state.highWaterMark) {
      state.highWaterMark = computeNewHighWaterMark(num);
    }

    var doRead = false;
    if (isNaN(num) || num === 0) {
      // Never call _read() again once the stream has ended: a source that
      // pushes on every call would otherwise loop forever.
      if (state.length === 0 && !state.ended) doRead = true;
    } else if (num > state.length) {
      if (!state.ended) doRead = true;
      else num = state.length;
    }
    if (doRead && !state.reading && state.constructed && !state.ended) {
      state.reading = true;
      state.sync = true;
      try {
        this._read(state.highWaterMark);
      } catch (e) {
        state.reading = false;
        state.sync = false;
        errorOrDestroy(this, e);
        return null;
      }
      state.sync = false;
    }

    if (state.length === 0) {
      if (state.ended) endReadable(this);
      else if (state.needReadable || state.readableListening) emitReadable(this);
      return null;
    }
    // read(0) is a request to fill the buffer, never a way to consume it;
    // negative sizes are likewise not a read.
    if (num <= 0) return null;

    var ret;
    var size;
    if (state.objectMode) {
      ret = state.buffer.shift();
      state.length -= 1;
      size = 1;
    } else if (state.decoder) {
      // Decoded entries are strings, so reads slice by characters.
      size = isNaN(num) ? state.length : Math.min(num, state.length);
      if (size >= state.length) {
        ret = state.buffer.join('');
        state.buffer.length = 0;
        state.length = 0;
      } else {
        var strParts = [];
        var takenChars = 0;
        while (takenChars < size && state.buffer.length) {
          var strHead = state.buffer[0];
          var wantChars = size - takenChars;
          if (strHead.length <= wantChars) {
            strParts.push(state.buffer.shift());
            takenChars += strHead.length;
          } else {
            strParts.push(strHead.slice(0, wantChars));
            state.buffer[0] = strHead.slice(wantChars);
            takenChars += wantChars;
          }
        }
        state.length -= takenChars;
        ret = strParts.join('');
      }
    } else {
      size = isNaN(num) ? state.length : Math.min(num, state.length);
      if (size >= state.length) {
        ret = concat(state.buffer, state.length);
        state.buffer.length = 0;
        state.length = 0;
      } else {
        var parts = [];
        var taken = 0;
        while (taken < size && state.buffer.length) {
          var head = state.buffer[0];
          var want = size - taken;
          if (head.length <= want) {
            parts.push(state.buffer.shift());
            taken += head.length;
          } else {
            parts.push(head.slice(0, want));
            state.buffer[0] = head.slice(want);
            taken += want;
          }
        }
        state.length -= taken;
        ret = concat(parts, taken);
      }
    }

    if (state.length === 0 && state.ended) endReadable(this);
    else maybeReadMore(this);

    if (state.encoding && !state.objectMode && typeof ret !== 'string') {
      ret = helpers.decode(ret, state.encoding);
    }
    return ret;
  };

  Readable.prototype.setEncoding = function (enc) {
    var state = this._readableState;
    var decoder = makeDecoder(enc);
    state.decoder = decoder;
    state.encoding = decoder.encoding;
    // Convert everything already buffered, so a later read sees text rather
    // than the bytes that arrived before the encoding was set.
    var content = '';
    for (var i = 0; i < state.buffer.length; i++) content += decoder.write(state.buffer[i]);
    if (state.ended) content += decoder.end();
    state.buffer.length = 0;
    if (content !== '') state.buffer.push(content);
    state.length = content.length;
    return this;
  };

  Readable.prototype.on = function (name, fn) {
    var out = Stream.prototype.on.call(this, name, fn);
    var state = this._readableState;
    if (!state) return out;
    if (name === 'data') {
      // Node resumes (and emits 'resume') as soon as a data listener exists,
      // unless the consumer explicitly paused the stream first.
      if (state.flowing !== false) this.resume();
    } else if (name === 'readable') {
      state.readableListening = true;
      state.needReadable = true;
      var s = this;
      tick(function () { maybeReadMore(s); });
    }
    return out;
  };
  Readable.prototype.addListener = Readable.prototype.on;
  Readable.prototype.once = function (name, fn) {
    var out = Stream.prototype.once.call(this, name, fn);
    var state = this._readableState;
    if (!state) return out;
    if (name === 'data') {
      this.resume();
    } else if (name === 'readable') {
      state.readableListening = true;
      state.needReadable = true;
      var self = this;
      tick(function () { maybeReadMore(self); });
    }
    return out;
  };
  Readable.prototype.resume = function () {
    var state = this._readableState;
    if (state.flowing !== true) {
      state.flowing = true;
      state.needReadable = false;
      this.emit('resume');
      var self = this;
      tick(function () { flow(self); });
    }
    return this;
  };
  Readable.prototype.pause = function () {
    var state = this._readableState;
    if (state.flowing !== false) {
      state.flowing = false;
      this.emit('pause');
    }
    return this;
  };
  Readable.prototype.isPaused = function () { return this._readableState.flowing === false; };

  Readable.prototype.pipe = function (dest, options) {
    var self = this;
    var state = this._readableState;
    options = options || {};
    if (state.destroyed) return dest;
    state.pipes.push(dest);
    dest.emit('pipe', this);

    function ondata(chunk) {
      if (dest.writable === false && typeof dest.write !== 'function') return;
      var ok = dest.write(chunk);
      if (ok === false) {
        state.flowing = false;
        self.pause();
        var drained = false;
        var onDrain = function () {
          if (drained) return;
          drained = true;
          dest.removeListener('drain', onDrain);
          if (!state.destroyed && state.pipes.indexOf(dest) !== -1) self.resume();
        };
        dest.once('drain', onDrain);
      }
    }
    function onend() {
      if (options.end === false) return;
      if (typeof dest.end === 'function') dest.end();
    }
    // Node installs one error handler on the destination; if nobody else is
    // listening it re-emits, so an unhandled destination error still becomes an
    // uncaughtException (test-stream-pipe-error-unhandled).
    function onerror(er) {
      self.unpipe(dest);
      if (typeof dest.removeListener === 'function') dest.removeListener('error', onerror);
      if (typeof dest.listenerCount !== 'function' || dest.listenerCount('error') === 0) {
        var st = dest._writableState || dest._readableState;
        if (st && !st.errorEmitted) errorOrDestroy(dest, er);
        else dest.emit('error', er);
      }
    }
    // Both close and finish unpipe, but only once.
    function onclose() {
      dest.removeListener('finish', onfinish);
      self.unpipe(dest);
    }
    function onfinish() {
      dest.removeListener('close', onclose);
      self.unpipe(dest);
    }

    this.on('data', ondata);
    this.once('end', onend);
    // A destination only has to be writable, not an EventEmitter, so every
    // listener hook is optional (test-stream-pipe-without-listenerCount).
    if (typeof dest.prependListener === 'function') dest.prependListener('error', onerror);
    else if (typeof dest.on === 'function') dest.on('error', onerror);
    if (typeof dest.once === 'function') {
      dest.once('close', onclose);
      dest.once('finish', onfinish);
    }
    dest._godeSource = this;
    dest._godePipeCleanup = function () {
      self.removeListener('data', ondata);
      self.removeListener('end', onend);
      dest.removeListener('error', onerror);
      dest.removeListener('close', onclose);
      dest.removeListener('finish', onfinish);
    };
    if (state.flowing !== false) this.resume();
    return dest;
  };

  Readable.prototype.unpipe = function (dest) {
    var state = this._readableState;
    if (!state) return this;
    var targets;
    if (dest) {
      // Unpiping something that was never piped is a no-op in Node: it must
      // not emit 'unpipe' (test-stream-pipe-unpipe-streams relies on that).
      var index = state.pipes.indexOf(dest);
      if (index === -1) return this;
      targets = [dest];
      state.pipes.splice(index, 1);
    } else {
      targets = state.pipes.slice();
      state.pipes = [];
    }
    for (var i = 0; i < targets.length; i++) {
      var d = targets[i];
      if (d && typeof d._godePipeCleanup === 'function') {
        d._godePipeCleanup();
        d._godePipeCleanup = null;
      }
      if (d && typeof d.emit === 'function') d.emit('unpipe', this);
    }
    if (state.pipes.length === 0 && state.flowing !== false) {
      // Node stops flowing once nothing is piped, and announces the pause.
      state.flowing = false;
      state.needReadable = false;
      this.emit('pause');
    }
    return this;
  };

  Readable.prototype.destroy = function (err, cb) { return destroyImpl(this, err, cb); };
  Readable.prototype._undestroy = function () {
    var r = this._readableState, w = this._writableState;
    if (r) { r.destroyed = false; r.errored = null; r.errorEmitted = false; r.closed = false; }
    if (w) {
      w.destroyed = false; w.errored = null; w.errorEmitted = false; w.closed = false;
      w.finalCalled = false; w.finished = false; w.ending = false; w.ended = false;
    }
    this._destroyed = false;
  };

  Readable.prototype.iterator = function (options) {
    options = options || {};
    var stream = this;
    var state = stream._readableState;
    var done = false;

    function cleanup(listeners) {
      stream.removeListener('readable', listeners.readable);
      stream.removeListener('end', listeners.end);
      stream.removeListener('error', listeners.error);
      stream.removeListener('close', listeners.close);
      if (state) state.readableListening = stream.listenerCount('readable') > 0;
    }

    var iterator = {
      next: function () {
        if (done) return Promise.resolve({ value: undefined, done: true });
        if (state && state.destroyed) {
          done = true;
          if (state.errored) return Promise.reject(state.errored);
          return Promise.resolve({ value: undefined, done: true });
        }
        return new Promise(function (resolve, reject) {
          var settled = false;
          var listeners = {};
          function settle(fn, value) {
            if (settled) return;
            settled = true;
            cleanup(listeners);
            fn(value);
          }
          listeners.readable = function () {
            if (settled) return;
            var chunk = stream.read();
            if (chunk === null) return;
            settle(resolve, { value: chunk, done: false });
          };
          listeners.end = function () {
            done = true;
            settle(resolve, { value: undefined, done: true });
          };
          listeners.error = function (err) {
            done = true;
            settle(reject, err);
          };
          listeners.close = function () {
            if (state && state.endEmitted) { listeners.end(); return; }
            done = true;
            settle(reject, coded(Error, 'ERR_STREAM_PREMATURE_CLOSE', 'Premature close'));
          };
          stream.on('readable', listeners.readable);
          stream.on('end', listeners.end);
          stream.on('error', listeners.error);
          stream.on('close', listeners.close);
          listeners.readable();
        });
      },
      return: function (value) {
        done = true;
        if (options.destroyOnReturn !== false && state && !state.destroyed) destroyImpl(stream, null, null);
        return Promise.resolve({ value: value, done: true });
      },
      throw: function (err) {
        done = true;
        if (state && !state.destroyed) destroyImpl(stream, err, null);
        return Promise.reject(err);
      }
    };
    iterator[Symbol.asyncIterator] = function () { return iterator; };
    return iterator;
  };

  Readable.prototype[Symbol.asyncIterator] = function () {
    return this.iterator();
  };

  // Legacy Readable.prototype.wrap adapts an old-style emitter (data/end/error)
  // to a Readable, pausing the source while the readable buffer is full.
  Readable.prototype.wrap = function (stream) {
    var self = this;
    var paused = false;
    stream.on('data', function (chunk) {
      if (self.push(chunk) === false && typeof stream.pause === 'function') {
        paused = true;
        stream.pause();
      }
    });
    stream.on('end', function () { self.push(null); });
    stream.on('error', function (err) { destroyImpl(self, err, null); });
    self._read = function () {
      if (paused && typeof stream.resume === 'function') {
        paused = false;
        stream.resume();
      }
    };
    return this;
  };

  // -------------------------------------------------------------- operators --
  // Node implements these as async generators (lib/internal/streams/operators.js).
  // The same surface is built here from the async iterator plus promise
  // chains, which keeps it independent of async-generator parser support.
  var SKIP = {};
  var STOP = {};

  function stopIterator(it) {
    if (it && typeof it.return === 'function') {
      try { it.return(); } catch (e) { /* the source is already gone */ }
    }
  }

  function eachAsync(source, fn) {
    return new Promise(function (resolve, reject) {
      var it = source.iterator ? source.iterator() : source[Symbol.asyncIterator]();
      var index = 0;
      function step() {
        it.next().then(function (res) {
          if (res.done) { resolve(); return; }
          var out;
          try { out = fn(res.value, index++); } catch (e) { reject(e); return; }
          if (out === STOP) { stopIterator(it); resolve(); return; }
          if (out && typeof out.then === 'function') {
            out.then(function (next) {
              if (next === STOP) { stopIterator(it); resolve(); return; }
              step();
            }, function (err) { reject(err); });
            return;
          }
          step();
        }, function (err) { reject(err); });
      }
      step();
    });
  }

  function operate(source, step, flatten) {
    var out = new Readable({ objectMode: true, read: function () { pump(); } });
    var it = source.iterator ? source.iterator() : source[Symbol.asyncIterator]();
    var queue = [];
    var pumping = false;
    var done = false;

    function flushQueue() {
      while (queue.length) {
        var value = queue.shift();
        if (out.push(value) === false) return false;
      }
      return true;
    }

    function pump() {
      if (done) return;
      if (!flushQueue()) return;
      if (pumping) return;
      pumping = true;
      step(it, function (err, value, eof) {
        pumping = false;
        if (done) return;
        if (err) { done = true; stopIterator(it); destroyImpl(out, err, null); return; }
        if (eof) { done = true; flushQueue(); out.push(null); return; }
        if (value === SKIP) { pump(); return; }
        if (flatten && Array.isArray(value)) {
          for (var i = 0; i < value.length; i++) queue.push(value[i]);
        } else {
          queue.push(value);
        }
        if (flushQueue()) pump();
      });
    }

    out.once('close', function () { done = true; stopIterator(it); });
    pump();
    return out;
  }

  function nextValue(it, cb) {
    it.next().then(function (res) {
      if (res.done) { cb(null, null, true); return; }
      cb(null, res.value, false);
    }, function (err) { cb(err); });
  }

  Readable.prototype.map = function (fn) {
    if (typeof fn !== 'function') throw errInvalidArgType('fn', 'of type function', fn);
    return operate(this, function (it, cb) {
      nextValue(it, function (err, value, eof) {
        if (err || eof) return cb(err, null, eof);
        var out;
        try { out = fn(value); } catch (e) { return cb(e); }
        if (out && typeof out.then === 'function') {
          out.then(function (v) { cb(null, v, false); }, function (e) { cb(e); });
          return;
        }
        cb(null, out, false);
      });
    });
  };

  Readable.prototype.filter = function (fn) {
    if (typeof fn !== 'function') throw errInvalidArgType('fn', 'of type function', fn);
    return operate(this, function (it, cb) {
      (function loop() {
        nextValue(it, function (err, value, eof) {
          if (err || eof) return cb(err, null, eof);
          var keep;
          try { keep = fn(value); } catch (e) { return cb(e); }
          var handle = function (k) {
            if (k) cb(null, value, false);
            else loop();
          };
          if (keep && typeof keep.then === 'function') { keep.then(handle, function (e) { cb(e); }); return; }
          handle(keep);
        });
      })();
    });
  };

  Readable.prototype.flatMap = function (fn) {
    if (typeof fn !== 'function') throw errInvalidArgType('fn', 'of type function', fn);
    return operate(this, function (it, cb) {
      nextValue(it, function (err, value, eof) {
        if (err || eof) return cb(err, null, eof);
        var out;
        try { out = fn(value); } catch (e) { return cb(e); }
        var handle = function (list) {
          if (list && typeof list !== 'string' && typeof list[Symbol.iterator] === 'function') {
            var arr = [];
            for (var v of list) arr.push(v);
            cb(null, arr.length ? arr : SKIP, false);
            return;
          }
          cb(null, SKIP, false);
        };
        if (out && typeof out.then === 'function') { out.then(handle, function (e) { cb(e); }); return; }
        handle(out);
      });
    }, true);
  };

  function toNumberOrZero(value) {
    var n = Number(value);
    if (n !== n) return 0;
    if (n < 0) throw coded(RangeError, 'ERR_OUT_OF_RANGE', 'The value of "number" is out of range. It must be >= 0. Received ' + n);
    return n;
  }

  Readable.prototype.take = function (number) {
    var n = toNumberOrZero(number);
    var taken = 0;
    return operate(this, function (it, cb) {
      if (taken >= n) { stopIterator(it); return cb(null, null, true); }
      nextValue(it, function (err, value, eof) {
        if (err || eof) return cb(err, null, eof);
        taken++;
        if (taken >= n) stopIterator(it);
        cb(null, value, false);
      });
    });
  };

  Readable.prototype.drop = function (number) {
    var n = toNumberOrZero(number);
    var dropped = 0;
    return operate(this, function (it, cb) {
      nextValue(it, function (err, value, eof) {
        if (err || eof) return cb(err, null, eof);
        if (dropped < n) { dropped++; cb(null, SKIP, false); return; }
        cb(null, value, false);
      });
    });
  };

  Readable.prototype.asIndexedPairs = function () {
    var index = 0;
    return operate(this, function (it, cb) {
      nextValue(it, function (err, value, eof) {
        if (err || eof) return cb(err, null, eof);
        cb(null, [index++, value], false);
      });
    });
  };

  Readable.prototype.toArray = function () {
    var out = [];
    return eachAsync(this, function (value) { out.push(value); }).then(function () { return out; });
  };

  Readable.prototype.forEach = function (fn) {
    if (typeof fn !== 'function') throw errInvalidArgType('fn', 'of type function', fn);
    return eachAsync(this, function (value) { return fn(value, { signal: undefined }); });
  };

  Readable.prototype.some = function (fn) {
    if (typeof fn !== 'function') throw errInvalidArgType('fn', 'of type function', fn);
    var found = false;
    return eachAsync(this, function (value) {
      var keep;
      try { keep = fn(value); } catch (e) { throw e; }
      if (keep && typeof keep.then === 'function') {
        return keep.then(function (k) { if (k) { found = true; return STOP; } });
      }
      if (keep) { found = true; return STOP; }
    }).then(function () { return found; });
  };

  Readable.prototype.every = function (fn) {
    if (typeof fn !== 'function') throw errInvalidArgType('fn', 'of type function', fn);
    var failed = false;
    return eachAsync(this, function (value) {
      var keep;
      try { keep = fn(value); } catch (e) { throw e; }
      if (keep && typeof keep.then === 'function') {
        return keep.then(function (k) { if (!k) { failed = true; return STOP; } });
      }
      if (!keep) { failed = true; return STOP; }
    }).then(function () { return !failed; });
  };

  Readable.prototype.find = function (fn) {
    if (typeof fn !== 'function') throw errInvalidArgType('fn', 'of type function', fn);
    var found;
    var has = false;
    return eachAsync(this, function (value) {
      var keep;
      try { keep = fn(value); } catch (e) { throw e; }
      if (keep && typeof keep.then === 'function') {
        return keep.then(function (k) { if (k) { found = value; has = true; return STOP; } });
      }
      if (keep) { found = value; has = true; return STOP; }
    }).then(function () { return has ? found : undefined; });
  };

  Readable.prototype.reduce = function (reducer, initial) {
    if (typeof reducer !== 'function') throw errInvalidArgType('reducer', 'of type function', reducer);
    var hasInitial = arguments.length > 1;
    var acc = initial;
    return eachAsync(this, function (value) {
      if (!hasInitial) { acc = value; hasInitial = true; return; }
      var out = reducer(acc, value);
      if (out && typeof out.then === 'function') return out.then(function (x) { acc = x; });
      acc = out;
    }).then(function () {
      if (!hasInitial) {
        throw coded(Error, 'ERR_MISSING_ARGS', 'Reduce of an empty stream requires an initial value');
      }
      return acc;
    });
  };

  Readable.prototype.compose = function () {
    var stages = [this];
    for (var i = 0; i < arguments.length; i++) stages.push(arguments[i]);
    return compose.apply(null, stages);
  };

  function asyncDisposeImpl() {
    var stream = this;
    return new Promise(function (resolve) {
      if (stream.destroyed) { resolve(); return; }
      var settled = false;
      function done() { if (settled) return; settled = true; resolve(); }
      stream.once('close', done);
      destroyImpl(stream, abortError(), null);
    });
  }
  if (typeof Symbol === 'function' && Symbol.asyncDispose) {
    Readable.prototype[Symbol.asyncDispose] = asyncDisposeImpl;
  }

  // --------------------------------------------------------------- Writable --
  function WritableState(options, isDuplex) {
    options = options || {};
    this.objectMode = !!(options.objectMode || (isDuplex && options.writableObjectMode));
    this.writable = options.writable !== false;
    this.highWaterMark = getHighWaterMark(this, options, 'writableHighWaterMark', isDuplex);
    this.buffer = [];
    this.length = 0;
    this.writing = false;
    this.corked = 0;
    this.sync = true;
    this.ending = false;
    this.ended = false;
    this.finished = false;
    this.finalCalled = false;
    this.needDrain = false;
    this.destroyed = false;
    this.errored = null;
    this.errorEmitted = false;
    this.closed = false;
    this.emitClose = options.emitClose !== false;
    this.autoDestroy = options.autoDestroy !== false;
    this.decodeStrings = options.decodeStrings !== false;
    this.defaultEncoding = options.defaultEncoding || 'utf8';
    this.constructed = false;
    this.onDestroyed = null;
    // Node exposes the queued-request count as state, not a length alias.
    Object.defineProperty(this, 'bufferedRequestCount', {
      configurable: true, enumerable: true,
      get: function () { return this.buffer.length; }
    });
  }

  function initWritable(self, options, isDuplex) {
    options = options || {};
    var state = new WritableState(options, isDuplex);
    self._writableState = state;
    if (typeof options.write === 'function') self._write = options.write;
    if (typeof options.writev === 'function') self._writev = options.writev;
    if (typeof options.final === 'function') self._final = options.final;
    if (typeof options.destroy === 'function') self._destroy = options.destroy;
    state.constructed = typeof options.construct !== 'function';
    if (options.signal) addAbortSignal(options.signal, self);
    return state;
  }

  function Writable(options) {
    if (!(this instanceof Writable)) return new Writable(options);
    Stream.call(this, options);
    initWritable(this, options, false);
    if (options && typeof options.construct === 'function') construct(this, options.construct, null);
  }
  Writable.prototype = Object.create(Stream.prototype);
  Writable.prototype.constructor = Writable;
  Writable.prototype._write = function (chunk, encoding, cb) {
    cb(errMethodNotImplemented('_write()'));
  };
  Writable.prototype._destroy = defaultDestroy;

  function writeChunk(stream, chunk, encoding, cb, isEnd) {
    var state = stream._writableState;
    var er;
    if (state.ending) {
      er = coded(Error, 'ERR_STREAM_WRITE_AFTER_END', 'write after end');
      tick(cb, er);
      errorOrDestroy(stream, er);
      return false;
    }	if (state.destroyed || stream.destroyed) {
      // Node reports write-after-destroy through the callback only: no
      // 'error' event is emitted for a stream that is already destroyed.
      er = coded(Error, 'ERR_STREAM_DESTROYED', 'Cannot call write after a stream was destroyed');
      tick(cb, er);
      return false;
    }
    var type = typeof chunk;
    // Node validates the chunk synchronously and throws; it does not route
    // these programmer errors through the write callback or 'error' event.
    if (chunk === null) {
      throw coded(TypeError, 'ERR_STREAM_NULL_VALUES', 'May not write null values to stream');
    }
    if (!state.objectMode) {
      if (type === 'string') {
        if (state.decodeStrings !== false) chunk = helpers.toBuffer(chunk, encoding || state.defaultEncoding);
        encoding = 'buffer';
      } else if (!isArrayBufferView(chunk)) {
        // Node reports a bad chunk through errorOrDestroy, not a throw; only a
        // null chunk (above) is a synchronous programmer error.
        errorOrDestroy(stream, errInvalidArgType('chunk', 'of type string or an instance of Buffer, TypedArray, or DataView', chunk));
        return false;
      } else {
        chunk = toBytes(chunk);
      }
      if (chunk === undefined) {
        throw coded(TypeError, 'ERR_STREAM_NULL_VALUES', 'May not write null values to stream');
      }
    }

    var len = state.objectMode ? 1 : chunk.length;
    state.length += len;
    var ret = state.length < state.highWaterMark;
    if (!ret) state.needDrain = true;

    if (state.writing || state.corked > 0 || !state.constructed) {
      state.buffer.push({ chunk: chunk, encoding: encoding || state.defaultEncoding, cb: cb, len: len });
    } else {
      state.writing = true;
      doWrite(stream, state, chunk, encoding || state.defaultEncoding, cb, len);
    }
    return ret === false ? false : true;
  }

  function doWrite(stream, state, chunk, encoding, cb, len) {
    var done = false;
    function onwrite(err) {
      if (done) { errorOrDestroy(stream, coded(Error, 'ERR_MULTIPLE_CALLBACK', 'Callback called multiple times')); return; }
      done = true;
      state.writing = false;
      state.sync = false;
      state.length -= len;
      if (state.length < 0) state.length = 0;
      if (err) {
        if (typeof cb === 'function') tick(cb, err);
        errorOrDestroy(stream, err);
        return;
      }
      if (typeof cb === 'function') tick(cb, null);
      clearBuffer(stream, state);
    }
    try {
      stream._write(chunk, encoding, onwrite);
    } catch (e) {
      onwrite(e);
    }
  }

  function clearBuffer(stream, state) {
    if (state.corked > 0 || state.writing || state.buffer.length === 0) {
      needDrain(stream, state);
      if (state.buffer.length === 0 && !state.writing) finishMaybe(stream, state);
      return;
    }
    // A corked batch (or several writes queued behind one in-flight write) is
    // handed to _writev in one call, the way Node batches buffered writes.
    if (typeof stream._writev === 'function' && state.buffer.length > 1) {
      var entries = state.buffer.slice();
      state.buffer.length = 0;
      var total = 0;
      for (var i = 0; i < entries.length; i++) total += entries[i].len;
      state.writing = true;
      doWritev(stream, state, entries, total);
      return;
    }
    var entry = state.buffer.shift();
    state.writing = true;
    doWrite(stream, state, entry.chunk, entry.encoding, entry.cb, entry.len);
  }

  function doWritev(stream, state, entries, total) {
    var done = false;
    function onwrite(err) {
      if (done) { errorOrDestroy(stream, coded(Error, 'ERR_MULTIPLE_CALLBACK', 'Callback called multiple times')); return; }
      done = true;
      state.writing = false;
      state.sync = false;
      state.length -= total;
      if (state.length < 0) state.length = 0;
      for (var i = 0; i < entries.length; i++) {
        if (typeof entries[i].cb === 'function') tick(entries[i].cb, err || null);
      }
      if (err) { errorOrDestroy(stream, err); return; }
      clearBuffer(stream, state);
    }
    var chunks = [];
    for (var j = 0; j < entries.length; j++) {
      chunks.push({ chunk: entries[j].chunk, encoding: entries[j].encoding });
    }
    try {
      stream._writev(chunks, onwrite);
    } catch (e) {
      onwrite(e);
    }
  }

  function needDrain(stream, state) {
    if (state.needDrain && !state.writing && state.length === 0 && state.buffer.length === 0) {
      state.needDrain = false;
      stream.emit('drain');
    }
  }

  function finishMaybe(stream, state) {
    if (state.finished || !state.ending || state.writing || state.buffer.length > 0) {
      needDrain(stream, state);
      return;
    }
    if (state.finished) return;
    if (typeof stream._final === 'function' && !state.finalCalled) {
      state.finalCalled = true;
      var called = false;
      var onfinal = function (err) {
        if (called) { errorOrDestroy(stream, coded(Error, 'ERR_MULTIPLE_CALLBACK', 'Callback called multiple times')); return; }
        called = true;
        if (err) { errorOrDestroy(stream, err); return; }
        afterFinal(stream, state);
      };
      try { stream._final(onfinal); } catch (e) { onfinal(e); }
      return;
    }
    afterFinal(stream, state);
  }

  function afterFinal(stream, state) {
    if (state.finished || state.destroyed) return;
    state.prefinished = true;
    stream.emit('prefinish');
    state.finished = true;
    state.ended = true;
    tick(function () {
      if (state.destroyed) return;
      stream.emit('finish');
      state.onDestroyed = null;
      // autoDestroy closes a Writable once it finishes, and a Duplex once both
      // sides are done (the readable side destroys from endReadable if its
      // 'end' is still pending).
      if (state.autoDestroy) {
        var rState = stream._readableState;
        if (!rState || rState.endEmitted) destroyImpl(stream, null, null);
      }
    });
  }

  Writable.prototype.write = function (chunk, encoding, cb) {
    if (typeof encoding === 'function') { cb = encoding; encoding = null; }
    return writeChunk(this, chunk, encoding, typeof cb === 'function' ? cb : function () {}, false);
  };
  Writable.prototype.end = function (chunk, encoding, cb) {
    if (typeof chunk === 'function') { cb = chunk; chunk = null; encoding = null; }
    else if (typeof encoding === 'function') { cb = encoding; encoding = null; }
    var state = this._writableState;
    if (typeof cb === 'function') {
      if (state.finished) { tick(cb, coded(Error, 'ERR_STREAM_ALREADY_FINISHED', 'Cannot call end after a stream was finished')); return this; }
      if (state.destroyed) { tick(cb, coded(Error, 'ERR_STREAM_DESTROYED', 'Cannot call write after a stream was destroyed')); return this; }
      // Node registers the callback as a one-shot finish listener, so it runs in
      // listener order and also receives a destroy error.
      var called = false;
      var once = function (err) { if (called) return; called = true; cb(err); };
      this.once('finish', function () { once(); });
      state.onDestroyed = function (err) {
        once(err || coded(Error, 'ERR_STREAM_DESTROYED', 'Cannot call write after a stream was destroyed'));
      };
    }
    if (state.ending || state.finished) return this;
    if (chunk !== null && chunk !== undefined) {
      writeChunk(this, chunk, encoding, function () {}, true);
    }
    state.ending = true;
    state.ended = true;
    finishMaybe(this, state);
    return this;
  };
  Writable.prototype.cork = function () { this._writableState.corked++; return this; };
  Writable.prototype.uncork = function () {
    var state = this._writableState;
    if (state.corked > 0) state.corked--;
    if (state.corked === 0) clearBuffer(this, state);
    return this;
  };
  Writable.prototype.setDefaultEncoding = function (enc) {
    if (typeof enc === 'string') enc = enc.toLowerCase();
    if (typeof Buffer === 'undefined' || typeof Buffer.isEncoding !== 'function' || !Buffer.isEncoding(enc)) {
      throw coded(TypeError, 'ERR_UNKNOWN_ENCODING', 'Unknown encoding: ' + enc);
    }
    this._writableState.defaultEncoding = enc;
    return this;
  };
  Writable.prototype.destroy = function (err, cb) { return destroyImpl(this, err, cb); };
  Writable.prototype.destroySoon = function () {
    if (this._writableState.finished) destroyImpl(this, null, null);
    else this.end();
  };
  Writable.prototype._undestroy = Readable.prototype._undestroy;
  if (typeof Symbol === 'function' && Symbol.asyncDispose) {
    Writable.prototype[Symbol.asyncDispose] = asyncDisposeImpl;
  }

  // ----------------------------------------------------------------- Duplex --
  function Duplex(options) {
    if (!(this instanceof Duplex)) return new Duplex(options);
    options = options || {};
    Stream.call(this, options);
    initReadable(this, options, true);
    initWritable(this, options, true);
    if (typeof options.construct === 'function') construct(this, options.construct, null);
    this.allowHalfOpen = options.allowHalfOpen !== false;
    if (options.readable === false) {
      this._readableState.readable = false;
      this._readableState.ended = true;
      this._readableState.endEmitted = true;
    }
    if (options.writable === false) {
      this._writableState.writable = false;
      this._writableState.ending = true;
      this._writableState.ended = true;
      this._writableState.finished = true;
    }
  }
  Duplex.prototype = Object.create(Readable.prototype);
  Duplex.prototype.constructor = Duplex;
  Object.keys(Writable.prototype).forEach(function (key) {
    if (key === 'constructor') return;
    Duplex.prototype[key] = Writable.prototype[key];
  });
  // Node makes every Duplex (and Transform) satisfy 'instanceof Writable'
  // through Symbol.hasInstance, even though its prototype chain is readable-first.
  if (typeof Symbol === 'function' && Symbol.hasInstance) {
    Object.defineProperty(Writable, Symbol.hasInstance, {
      configurable: true,
      value: function (instance) {
        if (Function.prototype[Symbol.hasInstance].call(this, instance)) return true;
        if (this !== Writable) return false;
        return !!instance && instance._writableState instanceof WritableState;
      }
    });
  }

  // -------------------------------------------------------------- Transform --
  function Transform(options) {
    if (!(this instanceof Transform)) return new Transform(options);
    options = options || {};
    Duplex.call(this, options);
    // _read is implemented below, so the first pull must not wait for a
    // constructed state the way a plain Readable would.
    this._readableState.sync = false;
    this._transformState = { pending: null, flushed: false };
    if (typeof options.transform === 'function') this._transform = options.transform;
    if (typeof options.flush === 'function') this._flush = options.flush;
  }
  Transform.prototype = Object.create(Duplex.prototype);
  Transform.prototype.constructor = Transform;
  Transform.prototype._transform = function () {
    throw errMethodNotImplemented('_transform()');
  };
  // Node only reports the write callback complete once the readable side has
  // room; otherwise it parks the callback and _read releases it. That is what
  // makes transform backpressure visible to writers (lib/internal/streams/
  // transform.js).
  Transform.prototype._write = function (chunk, encoding, cb) {
    var self = this;
    var rState = this._readableState;
    var wState = this._writableState;
    var length = rState.length;
    this._transform(chunk, encoding, function (err, data) {
      if (err) { cb(err); return; }
      if (data !== undefined && data !== null) self.push(data);
      if (rState.ended) tick(cb);
      else if (wState.ended || length === rState.length || rState.length < rState.highWaterMark) cb();
      else self._transformState.pending = cb;
    });
  };
  Transform.prototype._read = function () {
    var pending = this._transformState.pending;
    if (pending) {
      this._transformState.pending = null;
      pending();
    }
  };
  Transform.prototype._writev = function (chunks, cb) {
    var self = this;
    var index = 0;
    function next(err) {
      if (err) { cb(err); return; }
      if (index >= chunks.length) { cb(); return; }
      var entry = chunks[index++];
      self._transform(entry.chunk, entry.encoding, function (err2, data) {
        if (err2) { cb(err2); return; }
        if (data !== undefined && data !== null) self.push(data);
        next();
      });
    }
    next();
  };
  Transform.prototype._final = function (cb) {
    var self = this;
    if (typeof this._flush === 'function' && !this.destroyed) {
      this._flush(function (err, data) {
        if (err) { cb(err); return; }
        if (data !== undefined && data !== null) self.push(data);
        self.push(null);
        cb();
      });
      return;
    }
    this.push(null);
    cb();
  };

  function PassThrough(options) {
    if (!(this instanceof PassThrough)) return new PassThrough(options);
    Transform.call(this, options);
  }
  PassThrough.prototype = Object.create(Transform.prototype);
  PassThrough.prototype.constructor = PassThrough;
  PassThrough.prototype._transform = function (chunk, encoding, cb) {
    this.push(chunk);
    cb();
  };

  // -------------------------------------------------------------- helpers --
  function isNodeStream(value) {
    return !!(value && typeof value === 'object' &&
      (value._readableState || value._writableState ||
       typeof value.pipe === 'function' || typeof value.write === 'function'));
  }

  // finished() waits for every side the stream has: 'end' for readable,
  // 'finish' for writable, and reports ERR_STREAM_PREMATURE_CLOSE when the
  // stream closes without reaching them (lib/internal/streams/end-of-stream.js).
  function finished(stream, options, cb) {
    if (typeof options === 'function') { cb = options; options = {}; }
    options = options || {};
    if (options.cleanup !== undefined && typeof options.cleanup !== 'boolean') {
      throw errInvalidArgType('options.cleanup', 'of type boolean', options.cleanup);
    }
    var callback = typeof cb === 'function' ? cb : function () {};
    if (!stream || typeof stream.on !== 'function') { tick(callback); return function () {}; }
    var r = stream._readableState;
    var w = stream._writableState;
    var readable = options.readable;
    var writable = options.writable;
    if (readable === undefined && writable === undefined) {
      readable = !!(r || typeof stream.read === 'function');
      writable = !!(w || (typeof stream.write === 'function' && typeof stream.end === 'function'));
    }
    var ended = !!(r && r.endEmitted);
    var wrote = !!(w && w.finished);
    var done = false;
    var listeners = {};

    function cleanup() {
      // Node only removes eos listeners when cleanup was explicitly requested;
      // by default 'end'/'finish' listeners survive (test-stream-promises).
      if (options.cleanup !== true) return;
      stream.removeListener('end', listeners.end);
      stream.removeListener('finish', listeners.finish);
      stream.removeListener('error', listeners.error);
      stream.removeListener('close', listeners.close);
    }
    function fire(err) {
      if (done) return;
      done = true;
      cleanup();
      tick(function () { callback(err || null); });
    }
    function maybe() {
      if (done) return;
      if (readable && !ended) return;
      if (writable && !wrote) return;
      fire(null);
    }
    function premature() {
      return coded(Error, 'ERR_STREAM_PREMATURE_CLOSE', 'Premature close');
    }

    listeners.end = function () { ended = true; maybe(); };
    listeners.finish = function () { wrote = true; maybe(); };
    listeners.error = function (err) {
      if (options.error === false) return;
      fire(err);
    };
    listeners.close = function () {
      if (readable && !ended) { fire(premature()); return; }
      if (writable && !wrote) { fire(premature()); return; }
      maybe();
    };
    stream.on('end', listeners.end);
    stream.on('finish', listeners.finish);
    stream.on('error', listeners.error);
    stream.on('close', listeners.close);

    if (options.signal) {
      var onAbort = function () { destroyImpl(stream, abortError(options.signal.reason), null); };
      if (options.signal.aborted) onAbort();
      else if (typeof options.signal.addEventListener === 'function') options.signal.addEventListener('abort', onAbort);
    }
    maybe();
    return function () { done = true; cleanup(); };
  }

  // pipeline treats function stages as async generators: Node wraps each one
  // in a Duplex whose writable side feeds the generator and whose readable
  // side is its output (Duplex.from(async function* ...)).
  function stageFromFunction(fn) {
    var input = new PassThrough({ objectMode: true });
    var output = from(fn(input));
    var duplex = new Duplex({
      readableObjectMode: true,
      writableObjectMode: true,
      write: function (chunk, encoding, cb) { input.write(chunk, encoding, cb); },
      final: function (cb) { input.end(); cb(); },
      read: function () { if (typeof output.resume === 'function') output.resume(); }
    });
    output.on('data', function (chunk) {
      if (!duplex.destroyed && duplex.push(chunk) === false && typeof output.pause === 'function') output.pause();
    });
    output.on('end', function () { if (!duplex.destroyed) duplex.push(null); });
    output.on('error', function (err) { if (!duplex.destroyed) destroyImpl(duplex, err, null); });
    duplex._destroy = function (err, cb) {
      if (!input.destroyed) input.destroy(err);
      if (!output.destroyed) output.destroy(err);
      cb(err);
    };
    return duplex;
  }

  function pipeline() {
    var args = Array.prototype.slice.call(arguments);
    var cb = typeof args[args.length - 1] === 'function' ? args.pop() : null;
    var options = {};
    if (args.length && !isNodeStream(args[0]) && !Array.isArray(args[0]) &&
        typeof args[0] !== 'function' && args[0] && typeof args[0] === 'object') {
      options = args.shift();
    } else if (args.length > 1 && !isNodeStream(args[args.length - 1]) &&
               typeof args[args.length - 1] !== 'function' &&
               args[args.length - 1] && typeof args[args.length - 1] === 'object') {
      options = args.pop();
    }
    if (args.length === 1 && Array.isArray(args[0])) args = args[0];
    var streams = [];
    for (var a = 0; a < args.length; a++) {
      streams.push(typeof args[a] === 'function' ? stageFromFunction(args[a]) : args[a]);
    }
    var done = false;
    var errorListeners = [];
    var last = streams.length ? streams[streams.length - 1] : undefined;

    function hasReadableSide(s) { return !!(s && (s._readableState || typeof s.read === 'function')); }
    function hasWritableSide(s) { return !!(s && (s._writableState || typeof s.write === 'function')); }

    // Every stream gets an error listener; the one on a final readable stage is
    // removed when the pipeline completes, which is what Node's eos cleanup
    // leaves behind (test-stream-pipeline-listeners asserts the counts).
    function detachErrorListeners(keepLast) {
      for (var i = 0; i < errorListeners.length; i++) {
        var entry = errorListeners[i];
        if (keepLast && entry.stream === last) continue;
        entry.stream.removeListener('error', entry.fn);
      }
      errorListeners = [];
    }

    function finish(err) {
      if (done) return;
      done = true;
      if (options.signal && typeof options.signal.removeEventListener === 'function') {
        options.signal.removeEventListener('abort', onAbort);
      }
      detachErrorListeners(!hasReadableSide(last));
      if (err) {
        for (var i = 0; i < streams.length; i++) {
          var s = streams[i];
          if (s && typeof s.destroy === 'function' && !s.destroyed) {
            try { s.destroy(err); } catch (e) { /* already destroyed */ }
          }
        }
      }
      if (cb) tick(function () { cb(err || null); });
    }
    function onAbort() { finish(abortError(options.signal.reason)); }

    if (streams.length === 0) { finish(null); return undefined; }
    for (var k = 0; k < streams.length; k++) {
      if (!isNodeStream(streams[k])) {
        finish(errInvalidArgType('streams[' + k + ']', 'an instance of Stream', streams[k]));
        return last;
      }
    }
    if (options.signal) {
      if (options.signal.aborted) { onAbort(); return last; }
      if (typeof options.signal.addEventListener === 'function') options.signal.addEventListener('abort', onAbort);
    }
    for (var i = 0; i < streams.length; i++) {
      (function (stream) {
        var onerr = function (err) { finish(err); };
        errorListeners.push({ stream: stream, fn: onerr });
        stream.on('error', onerr);
      })(streams[i]);
    }
    for (var j = 0; j < streams.length - 1; j++) {
      if (typeof streams[j].pipe === 'function') streams[j].pipe(streams[j + 1]);
    }
    // A final stage that can be written to is awaited as a sink; only a pure
    // readable waits for 'end'.
    finished(last, {
      readable: hasReadableSide(last) && !hasWritableSide(last),
      writable: hasWritableSide(last),
      cleanup: true
    }, function (err) { finish(err || null); });
    return last;
  }

  function pipelinePromise() {
    var args = Array.prototype.slice.call(arguments);
    return new Promise(function (resolve, reject) {
      args.push(function (err) { if (err) reject(err); else resolve(); });
      pipeline.apply(null, args);
    });
  }

  function finishedPromise(stream, options) {
    // Argument errors are thrown synchronously, before the promise exists
    // (test-stream-promises wraps the call in assert.throws).
    options = options || {};
    if (options.cleanup !== undefined && typeof options.cleanup !== 'boolean') {
      throw errInvalidArgType('options.cleanup', 'of type boolean', options.cleanup);
    }
    return new Promise(function (resolve, reject) {
      finished(stream, options, function (err) { if (err) reject(err); else resolve(); });
    });
  }

  function addAbortSignalNoValidate(signal, stream) {
    if (!signal || typeof signal !== 'object' || !('aborted' in signal)) return stream;
    if (signal.aborted) { destroyImpl(stream, abortError(signal.reason), null); return stream; }
    var onAbort = function () { destroyImpl(stream, abortError(signal.reason), null); };
    if (typeof signal.addEventListener === 'function') signal.addEventListener('abort', onAbort, { once: true });
    if (typeof stream.once === 'function') {
      stream.once('close', function () {
        if (typeof signal.removeEventListener === 'function') signal.removeEventListener('abort', onAbort);
      });
    }
    return stream;
  }

  function addAbortSignal(signal, stream) {
    if (!signal || typeof signal !== 'object' || !('aborted' in signal)) {
      throw errInvalidArgType('signal', 'an instance of AbortSignal', signal);
    }
    return addAbortSignalNoValidate(signal, stream);
  }

  function from(iterable, options) {
    options = options || {};
    var objectMode = options.objectMode !== false;
    var opts = {};
    for (var k in options) if (Object.prototype.hasOwnProperty.call(options, k)) opts[k] = options[k];
    opts.objectMode = objectMode;
    var r = new Readable(opts);
    var it = null;
    var stopped = false;
    if (iterable && typeof iterable[Symbol.asyncIterator] === 'function') it = iterable[Symbol.asyncIterator]();
    else if (iterable && typeof iterable[Symbol.iterator] === 'function') it = iterable[Symbol.iterator]();
    else if (typeof iterable === 'string') {
      r.push(iterable);
      r.push(null);
      return r;
    } else {
      r.push(null);
      return r;
    }
    function pump() {
      if (stopped || r.destroyed) return;
      var res;
      try { res = it.next(); } catch (e) { destroyImpl(r, e, null); return; }
      if (res && typeof res.then === 'function') {
        res.then(function (v) {
          if (stopped || r.destroyed) return;
          if (v.done) { r.push(null); return; }
          r.push(v.value);
          pump();
        }, function (err) { destroyImpl(r, err, null); });
        return;
      }
      if (res.done) { r.push(null); return; }
      r.push(res.value);
      pump();
    }
    r._read = function () { pump(); };
    r.once('close', function () {
      stopped = true;
      if (it && typeof it.return === 'function') {
        try { it.return(); } catch (e) { /* the iterable is already gone */ }
      }
    });
    pump();
    return r;
  }

  // compose() returns a Duplex: writes go into the first stage and reads come
  // from the last one, which is what makes compose(a, b).pipe(dest) and
  // source.pipe(compose(a, b)) both work (lib/internal/streams/compose.js).
  function compose() {
    var stages = Array.prototype.slice.call(arguments);
    if (stages.length === 0) {
      throw coded(TypeError, 'ERR_MISSING_ARGS', 'The "streams" argument must be specified');
    }
    if (stages.length === 1) return stages[0];
    for (var k = 0; k < stages.length; k++) {
      if (!isNodeStream(stages[k])) {
        throw errInvalidArgType('streams[' + k + ']', 'an instance of Stream', stages[k]);
      }
    }
    var first = stages[0];
    var last = stages[stages.length - 1];
    var streams = stages;
    var ret = new Duplex({
      readableObjectMode: true,
      writableObjectMode: true,
      write: function (chunk, encoding, cb) {
        if (first && typeof first.write === 'function') first.write(chunk, encoding, cb);
        else cb();
      },
      final: function (cb) {
        if (first && typeof first.end === 'function') first.end();
        cb();
      },
      read: function () {
        if (last && typeof last.resume === 'function') last.resume();
      }
    });
    ret._destroy = function (err, cb) {
      for (var i = 0; i < streams.length; i++) {
        var s = streams[i];
        if (s && typeof s.destroy === 'function' && !s.destroyed) {
          try { s.destroy(err); } catch (e) { /* already destroyed */ }
        }
      }
      cb(err);
    };
    for (var i = 0; i < streams.length - 1; i++) {
      if (typeof streams[i].pipe === 'function') streams[i].pipe(streams[i + 1]);
    }
    last.on('data', function (chunk) {
      if (ret.push(chunk) === false && typeof last.pause === 'function') last.pause();
    });
    last.on('end', function () { ret.push(null); });
    for (var j = 0; j < streams.length; j++) {
      streams[j].on('error', function (err) { destroyImpl(ret, err, null); });
    }
    return ret;
  }

  function duplexPair(options) {
    var a, b;
    a = new Duplex(options);
    b = new Duplex(options);
    a._write = function (chunk, enc, cb) { b.push(chunk); cb(); };
    b._write = function (chunk, enc, cb) { a.push(chunk); cb(); };
    a._read = function () {};
    b._read = function () {};
    return [a, b];
  }

  function isReadable(s) {
    if (!s || typeof s.read !== 'function') return false;
    var st = s._readableState;
    if (!st) return false;
    if (st.readable === false || st.destroyed || st.ended) return false;
    return true;
  }
  function isWritable(s) {
    if (!s || typeof s.write !== 'function') return false;
    var st = s._writableState;
    if (!st) return false;
    if (st.destroyed || st.ending) return false;
    return true;
  }
  function isDestroyed(s) { return !!(s && s.destroyed); }
  function isDisturbed(s) {
    if (!s) return false;
    var st = s._readableState;
    // Node counts a read the consumer never finished (aborted) as disturbed.
    return !!(st && (st.didRead || (st.destroyed && !st.endEmitted)));
  }
  function isErrored(s) { return !!(s && s.errored); }
  function getDefaultHighWaterMark(objectMode) { return helpers.hwm(!!objectMode); }
  function setDefaultHighWaterMark(objectMode, value) { helpers.setHWM(!!objectMode, value); }

  function push(chunk, encoding) {
    if (this && typeof this.push === 'function' && this.push !== push) return this.push(chunk, encoding);
    return false;
  }
  function read(size) {
    if (this && typeof this.read === 'function' && this.read !== read) return this.read(size);
    return null;
  }

  // ---------------------------------------------------------- web bridge --
  // Readable.toWeb/fromWeb and Writable.toWeb/fromWeb adapt node streams to
  // the WHATWG streams in stream/web, matching lib/internal/streams/.../adapters.
  function toWebReadable(stream) {
    if (!stream || typeof stream.on !== 'function') {
      throw errInvalidArgType('stream', 'an instance of ReadableStream', stream);
    }
    return new ReadableStream({
      start: function (controller) {
        stream.on('data', function (chunk) {
          try { controller.enqueue(new Uint8Array(chunk)); } catch (e) { return; }
          if (controller.desiredSize !== null && controller.desiredSize <= 0 &&
              typeof stream.pause === 'function') {
            stream.pause();
          }
        });
        stream.on('end', function () { try { controller.close(); } catch (e) { /* already closed */ } });
        stream.on('error', function (err) { try { controller.error(err); } catch (e) { /* already errored */ } });
      },
      pull: function () { if (typeof stream.resume === 'function') stream.resume(); },
      cancel: function (reason) {
        if (typeof stream.destroy === 'function') stream.destroy(reason || undefined);
      }
    });
  }

  function fromWebReadable(webStream, options) {
    if (!webStream || typeof webStream.getReader !== 'function') {
      throw errInvalidArgType('readableStream', 'an instance of ReadableStream', webStream);
    }
    options = options || {};
    var reader = webStream.getReader();
    var reading = false;
    var r = new Readable({
      objectMode: !!options.objectMode,
      read: function () {
        if (reading) return;
        var self = this;
        reading = true;
        reader.read().then(function (res) {
          reading = false;
          if (self.destroyed) return;
          if (res.done) { self.push(null); return; }
          self.push(res.value);
        }, function (err) {
          reading = false;
          if (!self.destroyed) destroyImpl(self, err, null);
        });
      }
    });
    r._destroy = function (err, cb) {
      var done = function () { cb(err || null); };
      try {
        reader.cancel(err || undefined).then(done, done);
      } catch (e) { done(); }
    };
    return r;
  }

  function toWebWritable(stream) {
    if (!stream || typeof stream.write !== 'function') {
      throw errInvalidArgType('stream', 'an instance of WritableStream', stream);
    }
    function call(fn) {
      return new Promise(function (resolve, reject) {
        fn(function (err) { if (err) reject(err); else resolve(); });
      });
    }
    return new WritableStream({
      write: function (chunk) { return call(function (cb) { stream.write(chunk, cb); }); },
      close: function () {
        return call(function (cb) { if (typeof stream.end === 'function') stream.end(cb); else cb(); });
      },
      abort: function (reason) {
        if (typeof stream.destroy === 'function') stream.destroy(reason || undefined);
      }
    });
  }

  function fromWebWritable(webStream, options) {
    if (!webStream || typeof webStream.getWriter !== 'function') {
      throw errInvalidArgType('writableStream', 'an instance of WritableStream', webStream);
    }
    var writer = webStream.getWriter();
    return new Writable({
      objectMode: !!(options && options.objectMode),
      write: function (chunk, encoding, cb) {
        writer.write(chunk).then(function () { cb(); }, function (err) { cb(err); });
      },
      final: function (cb) {
        writer.close().then(function () { cb(); }, function (err) { cb(err); });
      },
      destroy: function (err, cb) {
        writer.abort(err || undefined).then(function () { cb(err || null); }, function () { cb(err || null); });
      }
    });
  }

  Readable.toWeb = toWebReadable;
  Readable.fromWeb = fromWebReadable;
  Writable.toWeb = toWebWritable;
  Writable.fromWeb = fromWebWritable;
  Duplex.toWeb = function (stream) {
    return { readable: toWebReadable(stream), writable: toWebWritable(stream) };
  };
  Duplex.fromWeb = function (pair, options) {
    pair = pair || {};
    var readable = fromWebReadable(pair.readable, options);
    var writable = fromWebWritable(pair.writable, options);
    var d = new Duplex({
      read: function () { if (typeof readable.resume === 'function') readable.resume(); }
    });
    readable.on('data', function (chunk) {
      if (!d.destroyed && d.push(chunk) === false && typeof readable.pause === 'function') readable.pause();
    });
    readable.on('end', function () { if (!d.destroyed) d.push(null); });
    readable.on('error', function (err) { if (!d.destroyed) destroyImpl(d, err, null); });
    d._write = function (chunk, encoding, cb) { writable.write(chunk, encoding, cb); };
    d._final = function (cb) { writable.end(cb); };
    d._destroy = function (err, cb) { writable.destroy(err); cb(err); };
    return d;
  };

  // stream.destroy(stream[, err]) is Node's public destroyer helper. Without
  // an error it aborts an unfinished stream with an AbortError, matching
  // lib/internal/streams/destroy.js.
  function isFinishedStream(s) {
    var r = s._readableState, w = s._writableState;
    if (r && !r.destroyed && !r.endEmitted) return false;
    if (w && !w.destroyed && !w.finished) return false;
    return true;
  }
  Stream.destroy = function (stream, err) {
    if (!stream || isDestroyed(stream)) return;
    if (!err && !isFinishedStream(stream)) err = abortError(undefined);
    if (typeof stream.destroy === 'function') stream.destroy(err);
    else if (typeof stream.close === 'function') stream.close();
    else if (err) tick(function () { stream.emit('error', err); stream.emit('close'); });
    else tick(function () { stream.emit('close'); });
    if (!stream.destroyed) stream._destroyed = true;
  };

  // Node defines the state accessors on each public prototype, so
  // Object.hasOwn(Readable.prototype, 'readableEnded') is true.
  function ownAccessors(target, names) {
    for (var i = 0; i < names.length; i++) {
      var d = Object.getOwnPropertyDescriptor(Stream.prototype, names[i]);
      if (d) Object.defineProperty(target, names[i], { configurable: true, enumerable: false, get: d.get, set: d.set });
    }
  }
  var readableNames = ['destroyed', 'closed', 'errored', 'readable', 'readableEnded', 'readableFlowing',
    'readableLength', 'readableHighWaterMark', 'readableObjectMode', 'readableDidRead', 'readableAborted',
    'readableEncoding', 'readableBuffer'];
  var writableNames = ['destroyed', 'closed', 'errored', 'writable', 'writableEnded', 'writableFinished',
    'writableLength', 'writableHighWaterMark', 'writableNeedDrain', 'writableCorked', 'writableObjectMode',
    'writableBuffer'];
  ownAccessors(Readable.prototype, readableNames);
  ownAccessors(Writable.prototype, writableNames);
  ownAccessors(Duplex.prototype, readableNames.concat(writableNames));

  Readable.from = from;
  Duplex.from = from;
  Transform.prototype.constructor = Transform;

  // util.promisify returns [Symbol.for('nodejs.util.promisify.custom')] when it
  // exists, which is how stream/promises and util.promisify(stream.pipeline)
  // stay reference-equal (test-stream-promises asserts strict equality).
  var kPromisifyCustom = Symbol.for('nodejs.util.promisify.custom');
  pipeline[kPromisifyCustom] = pipelinePromise;
  finished[kPromisifyCustom] = finishedPromise;

  return {
    Readable: Readable,
    Writable: Writable,
    Duplex: Duplex,
    Transform: Transform,
    PassThrough: PassThrough,
    Stream: Stream,
    pipeline: pipeline,
    finished: finished,
    isReadable: isReadable,
    isWritable: isWritable,
    isDestroyed: isDestroyed,
    isDisturbed: isDisturbed,
    isErrored: isErrored,
    isUint8Array: isUint8Array,
    _isUint8Array: isUint8Array,
    getDefaultHighWaterMark: getDefaultHighWaterMark,
    setDefaultHighWaterMark: setDefaultHighWaterMark,
    addAbortSignal: addAbortSignal,
    addAbortSignalNoValidate: addAbortSignalNoValidate,
    from: from,
    destroy: Stream.destroy,
    compose: compose,
    duplexPair: duplexPair,
    promises: { pipeline: pipelinePromise, finished: finishedPromise },
    push: push,
    read: read
  };
})`
