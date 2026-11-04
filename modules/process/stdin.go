package process

import (
	"io"
	"os"
	"syscall"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/modules/tty"
)

// stdinState backs process.stdin. The stream itself lives in JavaScript so it
// can extend stream.Readable; Go only supplies chunks. A read that would block
// returns undefined and the stream asks again after a short timer, which keeps
// the event loop free without a reader goroutine touching the JavaScript
// context.
type stdinState struct {
	e       *env.Env
	proc    js.Object
	host    js.Object
	builder js.Value // (Base) -> Stdin instance
	stream  js.Value
	buf     []byte
}

func installStdin(e *env.Env, proc js.Object) {
	st := &stdinState{e: e, proc: proc, buf: make([]byte, 64*1024)}

	host := e.JS.Obj()
	host.SetFunc("read", func(this js.Object, args []js.Value) (js.Value, error) {
		return st.readChunk()
	})
	host.SetFunc("isTTY", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Bool(st.isTTY()), nil
	})
	host.SetFunc("setRawMode", func(this js.Object, args []js.Value) (js.Value, error) {
		if !st.isTTY() {
			return e.JS.Bool(false), nil
		}
		raw := false
		if len(args) > 0 {
			raw = args[0].Bool()
		}
		return e.JS.Bool(tty.SetRawMode(0, raw)), nil
	})
	host.SetFunc("make", func(this js.Object, args []js.Value) (js.Value, error) {
		if st.stream.Valid() {
			return st.stream, nil
		}
		base := e.JS.Undefined()
		if env.Registered("stream") {
			if readable := e.Exports("stream").Get("Readable"); readable.Valid() && readable.IsFunction() {
				base = readable
			}
		}
		out, err := st.builder.Object().Call(js.Object{}, base)
		if err != nil {
			return js.Value{}, err
		}
		st.stream = out
		st.stream.Protect()
		return out, nil
	})
	st.host = host

	// The builder is lazy: it is only invoked once something touches
	// process.stdin, which is after every builtin has been installed.
	fn, err := e.JS.Eval(stdinBuilder, "[gode:process.stdin]")
	if err != nil {
		panic("godejs process.stdin: " + err.Error())
	}
	built, err := fn.Object().Call(js.Object{}, host.V())
	if err != nil {
		panic("godejs process.stdin: " + err.Error())
	}
	// The builder is only reachable from Go until the lazy getter runs, so
	// JavaScriptCore must be told to keep it alive; an unprotected handle is
	// collected and its cell reused during the stream module install below.
	st.builder = built.Protect()

	install, err := e.JS.Eval(stdinGetter, "[gode:process.stdin]")
	if err != nil {
		panic("godejs process.stdin: " + err.Error())
	}
	if _, err := install.Object().Call(js.Object{}, proc.V(), host.V()); err != nil {
		panic("godejs process.stdin: " + err.Error())
	}
}

func (st *stdinState) isTTY() bool {
	f, ok := st.e.In().(*os.File)
	return ok && f.Fd() == 0 && tty.IsTerminal(0)
}

// readChunk returns a Buffer with data, null at end of input, or undefined
// when the read would block.
func (st *stdinState) readChunk() (js.Value, error) {
	r := st.e.In()
	if f, ok := r.(*os.File); !ok || f.Fd() != 0 {
		n, err := r.Read(st.buf)
		if n > 0 {
			return st.chunk(st.buf[:n]), nil
		}
		if err == nil {
			return st.e.JS.Undefined(), nil
		}
		return st.e.JS.Null(), nil
	}

	// select() with a zero timeout reports whether a read would block. Doing
	// that instead of O_NONBLOCK leaves fd 0's flags alone for fs.readFileSync
	// and for child processes that inherit it.
	ready, serr := fdReadable(0)
	if serr != nil {
		// A closed or unusable descriptor is end of input, not a read that
		// will succeed later: retrying would spin the loop forever.
		return st.e.JS.Null(), nil
	}
	if !ready {
		return st.e.JS.Undefined(), nil
	}
	n, err := syscall.Read(0, st.buf)
	if n > 0 {
		return st.chunk(st.buf[:n]), nil
	}
	switch err {
	case nil:
		return st.e.JS.Null(), nil
	case syscall.EAGAIN, syscall.EINTR:
		return st.e.JS.Undefined(), nil
	}
	if err == io.EOF {
		return st.e.JS.Null(), nil
	}
	// EBADF and friends behave like end of input: there is no stream left to
	// read, and retrying forever would spin the loop.
	return st.e.JS.Null(), nil
}

// fdReadable reports whether fd has data (or EOF) ready without blocking.
func fdReadable(fd int) (bool, error) {
	var set syscall.FdSet
	set.Bits[fd/64] |= int64(1) << (uint(fd) % 64)
	tv := syscall.Timeval{}
	n, err := syscall.Select(fd+1, &set, nil, nil, &tv)
	if err == syscall.EINTR {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (st *stdinState) chunk(b []byte) js.Value {
	g := st.e.Global().Get("Buffer")
	if g.Valid() && g.IsFunction() {
		if from := g.Object().Get("from"); from.IsFunction() {
			if v, err := from.Object().Call(g.Object(), st.e.JS.Bytes(b).V()); err == nil && v.Valid() {
				return v
			}
		}
	}
	return st.e.JS.Bytes(b).V()
}

// stdinBuilder builds the Stdin class over stream.Readable (or a small
// fallback emitter when stream is unavailable, as in isolated unit tests).
const stdinBuilder = `(function (host) {
  'use strict';

  function FallbackEmitter() {
    this._events = Object.create(null);
    this.readable = true;
    this.destroyed = false;
  }
  function listeners(self, name, create) {
    var list = self._events[name];
    if (!list && create) list = self._events[name] = [];
    return list;
  }
  FallbackEmitter.prototype.on = function (name, fn) {
    listeners(this, name, true).push(fn);
    return this;
  };
  FallbackEmitter.prototype.addListener = FallbackEmitter.prototype.on;
  FallbackEmitter.prototype.prependListener = FallbackEmitter.prototype.on;
  FallbackEmitter.prototype.once = function (name, fn) {
    var self = this;
    function once() {
      self.removeListener(name, once);
      return fn.apply(self, arguments);
    }
    once.listener = fn;
    return this.on(name, once);
  };
  FallbackEmitter.prototype.removeListener = function (name, fn) {
    var list = listeners(this, name, false);
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
    var list = listeners(this, name, false);
    return list ? list.length : 0;
  };
  FallbackEmitter.prototype.emit = function (name) {
    var list = listeners(this, name, false);
    if (!list || !list.length) {
      if (name === 'error') throw arguments[1] instanceof Error ? arguments[1] : new Error('Unhandled error');
      return false;
    }
    var args = Array.prototype.slice.call(arguments, 1);
    list = list.slice();
    for (var i = 0; i < list.length; i++) list[i].apply(this, args);
    return true;
  };
  FallbackEmitter.prototype.push = function (chunk) {
    if (chunk === null) { this.readable = false; this.emit('end'); this.emit('close'); return false; }
    this.emit('data', chunk);
    return true;
  };
  FallbackEmitter.prototype.read = function () { return null; };
  FallbackEmitter.prototype.pause = function () { this._paused = true; return this; };
  FallbackEmitter.prototype.resume = function () { this._paused = false; return this; };
  FallbackEmitter.prototype.setEncoding = function () { return this; };
  FallbackEmitter.prototype.destroy = function (err) {
    if (this.destroyed) return this;
    this.destroyed = true;
    this.readable = false;
    if (err) this.emit('error', err);
    this.emit('close');
    return this;
  };
  FallbackEmitter.prototype.pipe = function (dest) {
    this.on('data', function (chunk) { dest.write(chunk); });
    this.on('end', function () { if (typeof dest.end === 'function') dest.end(); });
    return dest;
  };

  var defer = (typeof setTimeout === 'function')
    ? setTimeout
    : function (fn) { Promise.resolve().then(fn); };

  return function build(Base) {
    var Parent = (typeof Base === 'function') ? Base : FallbackEmitter;

    class Stdin extends Parent {
      constructor() {
        super();
        this.fd = 0;
        this.isTTY = !!host.isTTY();
        this._pending = false;
        this._ended = false;
        // _poll invalidates an in-flight retry timer when the stream is
        // paused or destroyed, so a closed readline cannot keep polling.
        this._poll = 0;
      }
      _read() {
        var self = this;
        if (self._pending || self._ended) return;
        self._pending = true;
        var token = ++self._poll;
        function attempt() {
          if (self._poll !== token || self._ended || self.destroyed) return;
          var chunk;
          try {
            chunk = host.read();
          } catch (err) {
            self._pending = false;
            self.destroy(err);
            return;
          }
          if (chunk === undefined) {
            defer(attempt, 4);
            return;
          }
          self._pending = false;
          if (chunk === null) {
            self._ended = true;
            self.push(null);
            return;
          }
          self.push(chunk);
        }
        attempt();
      }
      pause() {
        this._poll++;
        this._pending = false;
        if (typeof super.pause === 'function') return super.pause();
        return this;
      }
      resume() {
        if (typeof super.resume === 'function') return super.resume();
        return this;
      }
      destroy(err) {
        this._poll++;
        this._pending = false;
        if (typeof super.destroy === 'function') return super.destroy(err);
        return this;
      }
      setRawMode(mode) {
        var on = !(mode === false || mode === undefined || mode === null || mode === 0 || mode === '');
        host.setRawMode(on);
        this.isRaw = on;
        this.rawMode = on ? (mode === 'io' ? 'io' : 'raw') : false;
        return this;
      }
    }

    return new Stdin();
  };
})`

// stdinGetter defines process.stdin as a lazy getter so touching it installs
// stream.Readable on demand (and assigning to it keeps Node's writable slot).
const stdinGetter = `(function (proc, host) {
  'use strict';
  Object.defineProperty(proc, 'stdin', {
    configurable: true,
    enumerable: true,
    get: function () { return host.make(); },
    set: function (value) {
      Object.defineProperty(proc, 'stdin', {
        configurable: true, enumerable: true, writable: true, value: value
      });
    }
  });
})`
