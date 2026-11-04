// Package tty implements Node's tty builtin. TTY streams are net.Socket
// subclasses, matching Node, but their I/O goes straight to the underlying
// file descriptor: a terminal cannot be multiplexed through the socket pump.
// Terminal detection uses the termios ioctl rather than the file mode, so
// character devices such as /dev/null are not mistaken for terminals.
package tty

import (
	"os"
	"syscall"
	"unsafe"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("tty", Install) }

// module holds the environment while the exports object is built.
type module struct {
	e       *env.Env
	exports js.Object
}

// Install builds the tty exports object.
func Install(e *env.Env) js.Object {
	m := &module{e: e, exports: e.JS.Obj()}

	host := e.JS.Obj()
	host.SetFunc("isatty", func(_ js.Object, args []js.Value) (js.Value, error) {
		fd := argAt(args, 0)
		if !fd.IsNumber() {
			return e.JS.Bool(false), nil
		}
		n := fd.Int()
		if n < 0 || n > 2147483647 {
			return e.JS.Bool(false), nil
		}
		return e.JS.Bool(IsTerminal(int(n))), nil
	})
	host.SetFunc("write", func(_ js.Object, args []js.Value) (js.Value, error) {
		fd := int(argAt(args, 0).Int())
		text := argAt(args, 1).Str()
		var err error
		switch fd {
		case 1:
			_, err = e.Out().Write([]byte(text))
		case 2:
			_, err = e.Err().Write([]byte(text))
		default:
			f := os.NewFile(uintptr(fd), "tty")
			if f == nil {
				return e.JS.Str("EBADF: bad file descriptor"), nil
			}
			_, err = f.Write([]byte(text))
		}
		if err != nil {
			return e.JS.Str(err.Error()), nil
		}
		return e.JS.Str(""), nil
	})
	host.SetFunc("windowSize", func(_ js.Object, args []js.Value) (js.Value, error) {
		fd := int(argAt(args, 0).Int())
		cols, rows, ok := windowSize(fd)
		if !ok {
			return e.JS.Undefined(), nil
		}
		o := e.JS.Obj()
		o.SetNum("columns", float64(cols))
		o.SetNum("rows", float64(rows))
		return o.V(), nil
	})
	host.SetFunc("setRawMode", func(_ js.Object, args []js.Value) (js.Value, error) {
		fd := int(argAt(args, 0).Int())
		raw := argAt(args, 1).Bool()
		return e.JS.Bool(setRawMode(fd, raw)), nil
	})

	socket := e.Exports("net").Get("Socket")
	if !socket.Valid() || !socket.IsFunction() {
		socket = e.JS.Undefined()
	}
	events := e.JS.Undefined()
	if ee := e.Exports("events").Get("EventEmitter"); ee.Valid() && ee.IsFunction() {
		events = ee
	}

	fn, err := e.JS.Eval(prelude, "[gode:tty]")
	if err != nil {
		panic("godejs tty: " + err.Error())
	}
	api, err := fn.Object().Call(js.Object{}, host.V(), socket, events)
	if err != nil {
		panic("godejs tty: " + err.Error())
	}
	built := api.Object()
	for _, name := range built.OwnKeys() {
		m.exports.Set(name, built.Get(name))
	}
	return m.exports
}

func argAt(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

type winsize struct {
	Rows   uint16
	Cols   uint16
	Xpixel uint16
	Ypixel uint16
}

// IsTerminal reports whether fd is an open terminal. Node's isatty asks libuv
// to guess the handle type; a successful termios read is the same signal.
func IsTerminal(fd int) bool {
	var t syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&t)))
	return errno == 0
}

func windowSize(fd int) (cols, rows int, ok bool) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 0, 0, false
	}
	return int(ws.Cols), int(ws.Rows), true
}

// SetRawMode switches a terminal between raw and canonical mode. It is
// exported for process.stdin, whose setRawMode() drives the same ioctl.
func SetRawMode(fd int, raw bool) bool { return setRawMode(fd, raw) }

func setRawMode(fd int, raw bool) bool {
	var t syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&t))); errno != 0 {
		return false
	}
	if raw {
		t.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.IEXTEN | syscall.ISIG
		t.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
		t.Cc[syscall.VMIN] = 1
		t.Cc[syscall.VTIME] = 0
	} else {
		t.Lflag |= syscall.ICANON | syscall.ECHO | syscall.IEXTEN | syscall.ISIG
		t.Iflag |= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&t)))
	return errno == 0
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"tty.ReadStream",
		"tty.ReadStream#setRawMode",
		"tty.ReadStream.isRaw",
		"tty.ReadStream.isTTY",
		"tty.ReadStream.rawMode",
		"tty.WriteStream",
		"tty.WriteStream#clearLine",
		"tty.WriteStream#clearScreenDown",
		"tty.WriteStream#cursorTo",
		"tty.WriteStream#getColorDepth",
		"tty.WriteStream#getWindowSize",
		"tty.WriteStream#hasColors",
		"tty.WriteStream#moveCursor",
		"tty.WriteStream.columns",
		"tty.WriteStream.isTTY",
		"tty.WriteStream.rows",
		"tty.isatty",
	}
}

// prelude builds the JavaScript surface. It receives (host, net.Socket,
// EventEmitter) and returns the tty exports. The cursor and color helpers are
// ports of Node's internal/tty.js and internal/readline/callbacks.js.
const prelude = `(function (host, Socket, EventEmitter) {
  'use strict';

  var proc = (typeof process === 'object' && process !== null) ? process : null;
  var COLORS_2 = 1;
  var COLORS_16 = 4;
  var COLORS_256 = 8;
  var COLORS_16m = 24;

  var TERM_ENVS = {
    eterm: COLORS_16, cons25: COLORS_16, console: COLORS_16, cygwin: COLORS_16,
    dtterm: COLORS_16, gnome: COLORS_16, hurd: COLORS_16, jfbterm: COLORS_16,
    konsole: COLORS_16, kterm: COLORS_16, mlterm: COLORS_16, mosh: COLORS_16m,
    putty: COLORS_16, st: COLORS_16, 'rxvt-unicode-24bit': COLORS_16m,
    terminator: COLORS_16m, 'xterm-kitty': COLORS_16m
  };
  var CI_ENVS = {
    APPVEYOR: COLORS_256, BUILDKITE: COLORS_256, CIRCLECI: COLORS_16m,
    DRONE: COLORS_256, GITEA_ACTIONS: COLORS_16m, GITHUB_ACTIONS: COLORS_16m,
    GITLAB_CI: COLORS_256, TRAVIS: COLORS_256
  };
  var TERM_PATTERNS = [/ansi/, /color/, /linux/, /direct/, /^con[0-9]*x[0-9]/,
                       /^rxvt/, /^screen/, /^xterm/, /^vt100/, /^vt220/];

  function hasOwn(o, k) {
    return Object.prototype.hasOwnProperty.call(o, k);
  }

  function describe(v) {
    if (v === null) return 'null';
    if (v === undefined) return 'undefined';
    if (typeof v === 'string') return "type string ('" + v + "')";
    if (typeof v === 'object') {
      var name = v.constructor && v.constructor.name ? v.constructor.name : 'Object';
      return 'an instance of ' + name;
    }
    return 'type ' + typeof v + ' (' + String(v) + ')';
  }

  function makeError(name, code, message) {
    var Err = globalThis[name] || Error;
    var err = new Err(message);
    err.code = code;
    return err;
  }

  function getColorDepth(env) {
    if (!env) env = (proc && proc.env) || {};
    if (env.FORCE_COLOR !== undefined) {
      switch (env.FORCE_COLOR) {
        case '': case '1': case 'true': return COLORS_16;
        case '2': return COLORS_256;
        case '3': return COLORS_16m;
        default: return COLORS_2;
      }
    }
    if ((env.NODE_DISABLE_COLORS !== undefined && env.NODE_DISABLE_COLORS !== '') ||
        (env.NO_COLOR !== undefined && env.NO_COLOR !== '') ||
        env.TERM === 'dumb') {
      return COLORS_2;
    }
    if (proc && proc.platform === 'win32') return COLORS_16;
    if (env.TMUX) return COLORS_16m;
    if (hasOwn(env, 'TF_BUILD') && hasOwn(env, 'AGENT_NAME')) return COLORS_16;
    if (hasOwn(env, 'CI')) {
      for (var ci in CI_ENVS) {
        if (hasOwn(env, ci)) return CI_ENVS[ci];
      }
      if (env.CI_NAME === 'codeship') return COLORS_256;
      return COLORS_2;
    }
    if ('TEAMCITY_VERSION' in env) {
      return /^(9\.(0*[1-9]\d*)\.|\d{2,}\.)/.test(env.TEAMCITY_VERSION) ? COLORS_16 : COLORS_2;
    }
    switch (env.TERM_PROGRAM) {
      case 'iTerm.app':
        if (!env.TERM_PROGRAM_VERSION || /^[0-2]\./.test(env.TERM_PROGRAM_VERSION)) return COLORS_256;
        return COLORS_16m;
      case 'HyperTerm':
      case 'MacTerm':
        return COLORS_16m;
      case 'Apple_Terminal':
        return COLORS_256;
    }
    if (env.COLORTERM === 'truecolor' || env.COLORTERM === '24bit') return COLORS_16m;
    if (env.TERM) {
      if (/truecolor/.test(env.TERM)) return COLORS_16m;
      if (/^xterm-256/.test(env.TERM)) return COLORS_256;
      var termEnv = String(env.TERM).toLowerCase();
      if (TERM_ENVS[termEnv]) return TERM_ENVS[termEnv];
      for (var i = 0; i < TERM_PATTERNS.length; i++) {
        if (TERM_PATTERNS[i].test(termEnv)) return COLORS_16;
      }
    }
    if (env.COLORTERM) return COLORS_16;
    return COLORS_2;
  }

  function validateInteger(value, name, min) {
    if (typeof value !== 'number' || !Number.isInteger(value)) {
      throw makeError('TypeError', 'ERR_INVALID_ARG_TYPE',
        'The "' + name + '" argument must be of type number. Received ' + describe(value));
    }
    if (value < min) {
      throw makeError('RangeError', 'ERR_OUT_OF_RANGE',
        'The value of "' + name + '" is out of range. It must be >= ' + min + '. Received ' + value);
    }
  }

  function hasColors(count, env) {
    if (env === undefined && (count === undefined || (typeof count === 'object' && count !== null))) {
      env = count;
      count = 16;
    } else {
      validateInteger(count, 'count', 2);
    }
    return count <= Math.pow(2, getColorDepth(env));
  }

  function checkFd(fd) {
    if ((fd >> 0) !== fd || fd < 0) {
      throw makeError('TypeError', 'ERR_INVALID_FD',
        "The argument 'fd' must be a 32-bit signed integer. Received " + describe(fd));
    }
    return fd;
  }

  function isatty(fd) {
    return Number.isInteger(fd) && fd >= 0 && fd <= 2147483647 && host.isatty(fd);
  }

  function nextTick(cb) {
    if (proc && typeof proc.nextTick === 'function') return proc.nextTick(cb);
    return (globalThis.setTimeout || function (f) { f(); })(cb, 0);
  }

  function writeTo(stream, data) {
    if (stream && typeof stream.write === 'function') stream.write(data);
    return true;
  }

  function cursorTo(stream, x, y, callback) {
    if (typeof x !== 'number') {
      throw makeError('TypeError', 'ERR_INVALID_ARG_TYPE',
        'The "x" argument must be of type number. Received ' + describe(x));
    }
    if (y !== undefined && typeof y !== 'number') {
      throw makeError('TypeError', 'ERR_INVALID_ARG_TYPE',
        'The "y" argument must be of type number. Received ' + describe(y));
    }
    if (y === undefined) writeTo(stream, '\u001b[' + (x + 1) + 'G');
    else writeTo(stream, '\u001b[' + (y + 1) + ';' + (x + 1) + 'H');
    if (typeof callback === 'function') nextTick(callback);
    return true;
  }

  function moveCursor(stream, dx, dy, callback) {
    if (dx === null || dx === undefined) dx = 0;
    if (dy === null || dy === undefined) dy = 0;
    if (typeof dx !== 'number') {
      throw makeError('TypeError', 'ERR_INVALID_ARG_TYPE',
        'The "dx" argument must be of type number. Received ' + describe(dx));
    }
    if (typeof dy !== 'number') {
      throw makeError('TypeError', 'ERR_INVALID_ARG_TYPE',
        'The "dy" argument must be of type number. Received ' + describe(dy));
    }
    var out = '';
    if (dy < 0) out += '\u001b[' + (-dy) + 'A';
    else if (dy > 0) out += '\u001b[' + dy + 'B';
    if (dx < 0) out += '\u001b[' + (-dx) + 'D';
    else if (dx > 0) out += '\u001b[' + dx + 'C';
    if (out) writeTo(stream, out);
    if (typeof callback === 'function') nextTick(callback);
    return true;
  }

  function clearLine(stream, dir, callback) {
    if (dir === undefined || dir === null) dir = 0;
    if (typeof dir !== 'number') {
      throw makeError('TypeError', 'ERR_INVALID_ARG_TYPE',
        'The "dir" argument must be of type number. Received ' + describe(dir));
    }
    if (dir < 0) writeTo(stream, '\u001b[1K');
    else if (dir === 0) writeTo(stream, '\u001b[2K');
    else writeTo(stream, '\u001b[0K');
    if (typeof callback === 'function') nextTick(callback);
    return true;
  }

  function clearScreenDown(stream, callback) {
    writeTo(stream, '\u001b[0J');
    if (typeof callback === 'function') nextTick(callback);
    return true;
  }

  var Base = typeof Socket === 'function' ? Socket : EventEmitter;
  if (typeof Base !== 'function') throw new Error('tty: no stream base class');

  class ReadStream extends Base {
    constructor(fd, options) {
      checkFd(fd);
      super(options);
      this.fd = fd;
      this.isTTY = isatty(fd);
      this.isRaw = false;
      this.rawMode = false;
      this._ttyDestroyed = false;
    }
    setRawMode(mode) {
      var rawMode;
      if (mode === 'io' || mode === 'raw') {
        rawMode = mode;
      } else if (typeof mode === 'string') {
        throw makeError('TypeError', 'ERR_INVALID_ARG_VALUE',
          "The argument 'mode' is invalid. Received '" + mode + "'");
      } else {
        rawMode = mode ? 'raw' : false;
      }
      if (!host.setRawMode(this.fd, rawMode !== false)) {
        if (typeof this.listenerCount === 'function' && this.listenerCount('error') > 0) {
          this.emit('error', makeError('Error', 'EIO', 'setRawMode'));
        }
        return this;
      }
      this.isRaw = rawMode !== false;
      this.rawMode = rawMode;
      return this;
    }
    pause() { this.readable = false; return this; }
    resume() { this.readable = true; return this; }
    destroy(err) {
      if (this._ttyDestroyed) return this;
      this._ttyDestroyed = true;
      this.destroyed = true;
      this.readable = false;
      if (err) this.emit('error', err);
      this.emit('close');
      return this;
    }
  }
  ReadStream.prototype.isTTY = true;

  class WriteStream extends Base {
    constructor(fd) {
      checkFd(fd);
      super();
      this.fd = fd;
      this.isTTY = isatty(fd);
      var size = host.windowSize(fd);
      if (size) {
        this.columns = size.columns;
        this.rows = size.rows;
      }
      this._ttyDestroyed = false;
    }
    write(chunk, encoding, cb) {
      if (typeof encoding === 'function') { cb = encoding; encoding = undefined; }
      var text = chunk;
      if (typeof text !== 'string') {
        if (text && typeof text.toString === 'function') text = text.toString(encoding || 'utf8');
        else text = String(text);
      }
      var err = host.write(this.fd, text);
      if (err) throw makeError('Error', 'EIO', err);
      if (typeof cb === 'function') nextTick(cb);
      return true;
    }
    end(chunk, encoding, cb) {
      if (typeof chunk === 'function') { cb = chunk; chunk = undefined; }
      if (chunk !== undefined && chunk !== null) this.write(chunk, encoding);
      if (typeof cb === 'function') nextTick(cb);
      return this;
    }
    destroy(err) {
      if (this._ttyDestroyed) return this;
      this._ttyDestroyed = true;
      this.destroyed = true;
      this.writable = false;
      if (err) this.emit('error', err);
      this.emit('close');
      return this;
    }
    getWindowSize() { return [this.columns, this.rows]; }
    _refreshSize() {
      var oldCols = this.columns;
      var oldRows = this.rows;
      var size = host.windowSize(this.fd);
      if (!size) {
        this.emit('error', makeError('Error', 'EIO', 'getWindowSize'));
        return;
      }
      if (oldCols !== size.columns || oldRows !== size.rows) {
        this.columns = size.columns;
        this.rows = size.rows;
        this.emit('resize');
      }
    }
    getColorDepth(env) { return getColorDepth(env); }
    hasColors(count, env) { return hasColors(count, env); }
    cursorTo(x, y, cb) { return cursorTo(this, x, y, cb); }
    moveCursor(dx, dy, cb) { return moveCursor(this, dx, dy, cb); }
    clearLine(dir, cb) { return clearLine(this, dir, cb); }
    clearScreenDown(cb) { return clearScreenDown(this, cb); }
  }
  WriteStream.prototype.isTTY = true;

  // Node's tty constructors work with and without new. A class cannot be
  // called without new, so expose wrappers that share the class prototype
  // (keeping instanceof both tty.WriteStream and net.Socket true).
  function callable(cls) {
    function C(fd, options) {
      if (!(this instanceof C)) return new cls(fd, options);
      return new cls(fd, options);
    }
    C.prototype = cls.prototype;
    C.prototype.constructor = C;
    Object.setPrototypeOf(C, cls);
    return C;
  }

  return {
    isatty: isatty,
    ReadStream: callable(ReadStream),
    WriteStream: callable(WriteStream)
  };
})`
