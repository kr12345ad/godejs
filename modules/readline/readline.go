// Package readline implements Node's readline builtin. The Interface is built
// in JavaScript because it is mostly event plumbing over any input/output
// pair; Go only supplies the terminal facts and a stdout fallback.
package readline

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.Register("readline", Install)
	env.Register("readline/promises", installPromises)
}

// Install builds readline's exports object.
func Install(e *env.Env) js.Object {
	return build(e, false)
}

func installPromises(e *env.Env) js.Object {
	return build(e, true)
}

func build(e *env.Env, promises bool) js.Object {
	host := e.JS.Obj()
	host.SetFunc("write", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Undefined(), nil
		}
		v := args[0]
		if data, ok := v.Bytes(); ok {
			_, err := e.Out().Write(data)
			return e.JS.Bool(err == nil), nil
		}
		_, err := e.Out().Write([]byte(v.Str()))
		return e.JS.Bool(err == nil), nil
	})

	emitter := e.JS.Undefined()
	if env.Registered("events") {
		if ee := e.Exports("events").Get("EventEmitter"); ee.Valid() && ee.IsFunction() {
			emitter = ee
		}
	}

	fn, err := e.JS.Eval(prelude, "[gode:readline]")
	if err != nil {
		panic("godejs readline: " + err.Error())
	}
	exports, err := fn.Object().Call(js.Object{}, emitter, host.V(), e.JS.Bool(promises))
	if err != nil {
		panic("godejs readline: " + err.Error())
	}
	return exports.Object()
}

// Implemented lists the Node API entries this package provides. readline is
// not part of the tracked parity baseline, so nothing is advertised there.
func Implemented() []string { return nil }

const prelude = `(function (EventEmitter, host, promisesOnly) {
  'use strict';

  function FallbackEmitter() { this._events = Object.create(null); }
  FallbackEmitter.prototype.on = function (name, fn) {
    (this._events[name] || (this._events[name] = [])).push(fn);
    return this;
  };
  FallbackEmitter.prototype.addListener = FallbackEmitter.prototype.on;
  FallbackEmitter.prototype.once = function (name, fn) {
    var self = this;
    function once() { self.removeListener(name, once); return fn.apply(self, arguments); }
    once.listener = fn;
    return this.on(name, once);
  };
  FallbackEmitter.prototype.removeListener = function (name, fn) {
    var list = this._events[name];
    if (!list) return this;
    for (var i = 0; i < list.length; i++) {
      if (list[i] === fn || list[i].listener === fn) { list.splice(i, 1); break; }
    }
    return this;
  };
  FallbackEmitter.prototype.off = FallbackEmitter.prototype.removeListener;
  FallbackEmitter.prototype.removeAllListeners = function (name) {
    if (name === undefined) this._events = Object.create(null);
    else delete this._events[name];
    return this;
  };
  FallbackEmitter.prototype.listenerCount = function (name) {
    var list = this._events[name];
    return list ? list.length : 0;
  };
  FallbackEmitter.prototype.emit = function (name) {
    var list = this._events[name];
    if (!list || !list.length) {
      if (name === 'error') throw arguments[1] instanceof Error ? arguments[1] : new Error('Unhandled error');
      return false;
    }
    var args = Array.prototype.slice.call(arguments, 1);
    list = list.slice();
    for (var i = 0; i < list.length; i++) list[i].apply(this, args);
    return true;
  };

  var Emitter = (typeof EventEmitter === 'function') ? EventEmitter : FallbackEmitter;

  function typeError(code, message) {
    var err = new TypeError(message);
    err.code = code;
    return err;
  }

  function writeTo(stream, data) {
    if (stream && typeof stream.write === 'function') stream.write(data);
    else host.write(data);
    return true;
  }

  function cursorTo(stream, x, y, callback) {
    if (typeof x !== 'number') {
      throw typeError('ERR_INVALID_ARG_TYPE',
        'The "x" argument must be of type number. Received ' + typeof x);
    }
    if (stream && typeof stream.cursorTo === 'function') return stream.cursorTo(x, y, callback);
    if (y === undefined) writeTo(stream, '\u001b[' + (x + 1) + 'G');
    else writeTo(stream, '\u001b[' + (y + 1) + ';' + (x + 1) + 'H');
    if (typeof callback === 'function') process.nextTick(callback);
    return true;
  }

  function moveCursor(stream, dx, dy, callback) {
    if (stream && typeof stream.moveCursor === 'function') return stream.moveCursor(dx, dy, callback);
    var out = '';
    if (dy < 0) out += '\u001b[' + (-dy) + 'A';
    else if (dy > 0) out += '\u001b[' + dy + 'B';
    if (dx < 0) out += '\u001b[' + (-dx) + 'D';
    else if (dx > 0) out += '\u001b[' + dx + 'C';
    if (out) writeTo(stream, out);
    if (typeof callback === 'function') process.nextTick(callback);
    return true;
  }

  function clearLine(stream, dir, callback) {
    if (stream && typeof stream.clearLine === 'function') return stream.clearLine(dir, callback);
    if (dir < 0) writeTo(stream, '\u001b[1K');
    else if (dir > 0) writeTo(stream, '\u001b[0K');
    else writeTo(stream, '\u001b[2K');
    if (typeof callback === 'function') process.nextTick(callback);
    return true;
  }

  function clearScreenDown(stream, callback) {
    if (stream && typeof stream.clearScreenDown === 'function') return stream.clearScreenDown(callback);
    writeTo(stream, '\u001b[0J');
    if (typeof callback === 'function') process.nextTick(callback);
    return true;
  }

  function keyName(sequence) {
    switch (sequence) {
      case '\u0003': return { name: 'c', ctrl: true };
      case '\r': case '\n': return { name: 'return' };
      case '\t': return { name: 'tab' };
      case '\u007f': case '\b': return { name: 'backspace' };
      case '\u001b': return { name: 'escape' };
      case ' ': return { name: 'space' };
      default:
        if (sequence.length === 1) return { name: sequence.toLowerCase(), shift: sequence !== sequence.toLowerCase() };
        return { name: undefined };
    }
  }

  function emitKeypressEvents(stream, iface) {
    if (!stream || typeof stream.on !== 'function') return;
    var buffer = '';
    stream.on('data', function (chunk) {
      buffer += (typeof chunk === 'string') ? chunk : String(chunk);
      while (buffer.length) {
        var seq = buffer.charAt(0);
        // Keep escape sequences together so arrow keys arrive as one key.
        if (seq === '\u001b' && buffer.length > 1) {
          var match = /^\u001b\[[0-9;]*[A-Za-z~]/.exec(buffer);
          if (match) seq = match[0];
        }
        buffer = buffer.slice(seq.length);
        var info = keyName(seq);
        if (iface && typeof iface.emit === 'function') iface.emit('keypress', seq, info);
        else stream.emit('keypress', seq, info);
      }
    });
  }

  function isReader(input) {
    return !!input && typeof input === 'object' && typeof input.on === 'function';
  }

  class Interface extends Emitter {
    constructor(input, output, options) {
      super();
      if (input && typeof input === 'object' && !isReader(input)) {
        options = input;
        input = options.input;
        output = options.output;
      }
      if (!options || typeof options !== 'object') options = {};
      if (!isReader(input)) {
        throw typeError('ERR_INVALID_ARG_TYPE',
          'The "input" argument must be an instance of stream.Readable. Received ' + typeof input);
      }
      this.input = input;
      this.output = output;
      this.terminal = options.terminal !== undefined ? !!options.terminal : !!(output && output.isTTY);
      this.promptText = options.prompt !== undefined ? String(options.prompt) : '> ';
      this.historySize = typeof options.historySize === 'number' ? options.historySize : 30;
      this.crlfDelay = typeof options.crlfDelay === 'number' ? options.crlfDelay : 100;
      this.history = [];
      this.closed = false;
      this.paused = false;
      this._buffer = '';
      this._pendingCR = false;
      this._question = null;
      this._decoder = null;
      try {
        if (typeof require === 'function') {
          var StringDecoder = require('string_decoder').StringDecoder;
          this._decoder = new StringDecoder('utf8');
        }
      } catch (ignored) { /* string_decoder unavailable */ }

      var self = this;
      this._onData = function (chunk) { self._handleData(chunk); };
      this._onEnd = function () { self._handleEnd(); };
      input.on('data', this._onData);
      input.on('end', this._onEnd);
      if (this.terminal && typeof input.setRawMode === 'function') {
        try { input.setRawMode(true); } catch (ignored) { /* non-tty */ }
      }
    }

    _handleData(chunk) {
      var text;
      if (typeof chunk === 'string') text = chunk;
      else if (this._decoder) text = this._decoder.write(chunk);
      else if (chunk && typeof chunk.toString === 'function') text = chunk.toString('utf8');
      else text = String(chunk);
      if (!text) return;
      if (this._pendingCR) {
        if (text.charAt(0) === '\n') text = text.slice(1);
        this._pendingCR = false;
      }
      this._buffer += text;
      var trailingCR = false;
      if (this._buffer.charAt(this._buffer.length - 1) === '\r') {
        this._buffer = this._buffer.slice(0, -1);
        trailingCR = true;
        this._pendingCR = true;
      }
      var parts = this._buffer.split(/\r\n|\n|\r/);
      this._buffer = parts.pop();
      for (var i = 0; i < parts.length; i++) this._line(parts[i]);
      if (trailingCR) {
        var line = this._buffer;
        this._buffer = '';
        this._line(line);
      }
    }

    _handleEnd() {
      var self = this;
      // Node reports end asynchronously; deferring also means a synchronous
      // close() right after createInterface() wins over an empty input.
      function finish() {
        if (self.closed) return;
        if (self._decoder && typeof self._decoder.end === 'function') {
          var rest = self._decoder.end();
          if (rest) self._buffer += rest;
        }
        if (self._buffer.length || self._pendingCR) {
          var line = self._buffer;
          self._buffer = '';
          self._pendingCR = false;
          self._line(line);
        }
        if (self._question) {
          var pending = self._question;
          self._question = null;
          pending(null);
        }
        self.close();
      }
      if (typeof process !== 'undefined' && process && typeof process.nextTick === 'function') {
        process.nextTick(finish);
      } else {
        finish();
      }
    }

    _line(line) {
      if (this._question) {
        var cb = this._question;
        this._question = null;
        if (this.terminal) writeTo(this.output, '\n');
        cb(line);
        return;
      }
      if (this.historySize > 0) {
        this.history.push(line);
        if (this.history.length > this.historySize) this.history.shift();
      }
      this.emit('line', line);
    }

    setPrompt(prompt) { this.promptText = String(prompt); }

    getPrompt() { return this.promptText; }

    prompt(preserveCursor) {
      if (this.closed) return;
      writeTo(this.output, this.promptText);
    }

    write(data, key) {
      if (this.closed) return;
      writeTo(this.output, data);
    }

    _ask(query, options, cb) {
      if (typeof options === 'function') { cb = options; options = undefined; }
      if (typeof cb !== 'function') {
        throw typeError('ERR_INVALID_ARG_TYPE',
          'The "callback" argument must be of type function. Received ' + typeof cb);
      }
      if (this.closed) {
        var err = new Error('readline was closed');
        err.code = 'ERR_USE_AFTER_CLOSE';
        throw err;
      }
      writeTo(this.output, query === undefined ? '' : String(query));
      this._question = cb;
    }

    question(query, options, cb) { this._ask(query, options, cb); }

    pause() {
      this.paused = true;
      if (this.input && typeof this.input.pause === 'function') this.input.pause();
      this.emit('pause');
      return this;
    }

    resume() {
      this.paused = false;
      if (this.input && typeof this.input.resume === 'function') this.input.resume();
      this.emit('resume');
      return this;
    }

    close() {
      if (this.closed) return;
      // Node pauses the input before tearing the interface down, which also
      // lets tests observe close() -> pause().
      this.pause();
      this.closed = true;
      if (this.input && typeof this.input.removeListener === 'function') {
        this.input.removeListener('data', this._onData);
        this.input.removeListener('end', this._onEnd);
      }
      if (this.terminal && this.input && typeof this.input.setRawMode === 'function') {
        try { this.input.setRawMode(false); } catch (ignored) { /* non-tty */ }
      }
      this.emit('close');
    }

    getCursorPos() { return { rows: 0, cols: this.promptText.length }; }

    [Symbol.asyncIterator]() {
      var self = this;
      var queue = [];
      var waiters = [];
      var finished = false;
      function onLine(line) {
        if (waiters.length) waiters.shift()({ value: line, done: false });
        else queue.push(line);
      }
      function onClose() {
        finished = true;
        while (waiters.length) waiters.shift()({ value: undefined, done: true });
      }
      self.on('line', onLine);
      self.on('close', onClose);
      if (self.closed) onClose();
      return {
        next: function () {
          if (queue.length) return Promise.resolve({ value: queue.shift(), done: false });
          if (finished) return Promise.resolve({ value: undefined, done: true });
          return new Promise(function (resolve) { waiters.push(resolve); });
        },
        return: function () {
          self.removeListener('line', onLine);
          self.removeListener('close', onClose);
          self.close();
          return Promise.resolve({ value: undefined, done: true });
        },
        throw: function (err) { return Promise.reject(err); },
        [Symbol.asyncIterator]: function () { return this; }
      };
    }
  }

  class PromisesInterface extends Interface {
    question(query, options) {
      var self = this;
      return new Promise(function (resolve, reject) {
        try { self._ask(query, options, resolve); } catch (err) { reject(err); }
      });
    }
  }

  function createInterface(input, output, options) {
    return promisesOnly
      ? new PromisesInterface(input, output, options)
      : new Interface(input, output, options);
  }

  var api = {
    Interface: promisesOnly ? PromisesInterface : Interface,
    createInterface: createInterface,
    clearLine: clearLine,
    clearScreenDown: clearScreenDown,
    cursorTo: cursorTo,
    moveCursor: moveCursor
  };
  if (!promisesOnly) api.emitKeypressEvents = emitKeypressEvents;
  return api;
})`
