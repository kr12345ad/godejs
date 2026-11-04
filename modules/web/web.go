// Package web installs the web-platform globals Node.js ships but
// JavaScriptCore does not: TextEncoder, TextDecoder, atob/btoa,
// queueMicrotask, structuredClone, performance, AbortController and
// AbortSignal.
//
// It is not a require() builtin. internal/all calls Install directly; the
// classes live in a small JS prelude because they are realm-native objects
// (constructors, prototypes, instanceof) that the host bridge cannot create
// faithfully. Byte codecs and timing live in Go.
package web

import (
	"encoding/base64"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"godejs/internal/env"
	"godejs/internal/js"
)

// Install defines the globals on e. It is idempotent per environment.
func Install(e *env.Env) {
	start := e.StartTime
	if start.IsZero() {
		start = time.Now()
	}

	host := e.JS.Obj()
	host.SetFuncHidden("utf8", func(_ js.Object, args []js.Value) (js.Value, error) {
		s := ""
		if len(args) > 0 {
			s = args[0].Str()
		}
		return e.JS.Bytes([]byte(s)).V(), nil
	})
	host.SetFuncHidden("decode", func(_ js.Object, args []js.Value) (js.Value, error) {
		return hostDecode(e, args)
	})
	host.SetFuncHidden("atob", func(_ js.Object, args []js.Value) (js.Value, error) {
		s := ""
		if len(args) > 0 {
			s = args[0].Str()
		}
		out, err := decodeBase64(s)
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(out), nil
	})
	host.SetFuncHidden("btoa", func(_ js.Object, args []js.Value) (js.Value, error) {
		s := ""
		if len(args) > 0 {
			s = args[0].Str()
		}
		out, err := encodeBase64(s)
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(out), nil
	})
	host.SetFuncHidden("perfNow", func(_ js.Object, _ []js.Value) (js.Value, error) {
		return e.JS.Num(float64(time.Since(start).Nanoseconds()) / 1e6), nil
	})
	host.SetFuncHidden("setTimeout", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 || !args[0].IsFunction() {
			return js.Value{}, js.Throw("TypeError", "The \"callback\" argument must be of type function")
		}
		delay := args[1].Float()
		e.SetTimeout(args[0], delay, false, nil)
		return e.JS.Undefined(), nil
	})

	// The prelude reads the helpers and immediately deletes the global, so no
	// user-visible name survives. The helpers themselves are non-enumerable.
	e.Global().Set("__godeWeb", host.V())
	if _, err := e.JS.Eval(prelude, "node:web"); err != nil {
		panic("web: " + err.Error())
	}

	// Storage and navigator build on the prelude's DOMException; the module
	// globals are independent constructors Node also installs up front.
	installStorage(e)
	installNavigator(e)
	installModuleGlobals(e)
	installModern(e)
	installFetch(e)
	installWebSocket(e)
}

func hostDecode(e *env.Env, args []js.Value) (js.Value, error) {
	encoding := "utf-8"
	if len(args) > 0 {
		encoding = args[0].Str()
	}
	var data, pending []byte
	if len(args) > 1 {
		if b, ok := args[1].Bytes(); ok {
			data = b
		}
	}
	if len(args) > 2 {
		if b, ok := args[2].Bytes(); ok {
			pending = b
		}
	}
	stream := len(args) > 3 && args[3].Bool()
	ignoreBOM := len(args) > 4 && args[4].Bool()
	bomChecked := len(args) > 5 && args[5].Bool()
	fatal := len(args) > 6 && args[6].Bool()

	text, rest, checked, err := decodeText(encoding, data, pending, stream, ignoreBOM, bomChecked, fatal)
	if err != nil {
		return js.Value{}, err
	}
	out := e.JS.Obj()
	out.SetStr("text", text)
	out.Set("pending", e.JS.Bytes(rest).V())
	out.SetBool("bomChecked", checked)
	return out.V(), nil
}

// ---------------------------------------------------------------------------
// atob / btoa

func decodeBase64(input string) (string, error) {
	var b strings.Builder
	for _, r := range input {
		switch r {
		case ' ', '\t', '\n', '\f', '\r':
			continue
		}
		if r > 0x7f || !isBase64Char(byte(r)) {
			return "", invalidCharacter("The string to be decoded is not correctly encoded.")
		}
		b.WriteByte(byte(r))
	}
	s := b.String()
	if len(s)%4 == 1 {
		return "", invalidCharacter("The string to be decoded is not correctly encoded.")
	}
	for len(s)%4 != 0 {
		s += "="
	}
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", invalidCharacter("The string to be decoded is not correctly encoded.")
	}
	var out strings.Builder
	for _, c := range raw {
		out.WriteRune(rune(c))
	}
	return out.String(), nil
}

func isBase64Char(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' ||
		c >= '0' && c <= '9' || c == '+' || c == '/' || c == '='
}

func encodeBase64(input string) (string, error) {
	buf := make([]byte, 0, len(input))
	for _, r := range input {
		if r > 0xFF {
			return "", invalidCharacter("The string to be encoded contains characters outside of the Latin1 range.")
		}
		buf = append(buf, byte(r))
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

func invalidCharacter(msg string) error {
	return js.ThrowCode("InvalidCharacterError", "ERR_INVALID_CHARACTER_ERR", msg)
}

// ---------------------------------------------------------------------------
// TextDecoder

func decodeText(enc string, data, pending []byte, stream, ignoreBOM, bomChecked, fatal bool) (string, []byte, bool, error) {
	buf := make([]byte, 0, len(pending)+len(data))
	buf = append(buf, pending...)
	buf = append(buf, data...)

	switch enc {
	case "utf-8":
		if !ignoreBOM && !bomChecked {
			if len(buf) < 3 {
				if utf8BOMPrefix(buf) && stream {
					return "", buf, false, nil
				}
				bomChecked = true
			} else {
				if buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
					buf = buf[3:]
				}
				bomChecked = true
			}
		}
		text, rest, err := decodeUTF8(buf, stream, fatal)
		return text, rest, bomChecked, err
	case "utf-16le", "utf-16be":
		little := enc == "utf-16le"
		if !ignoreBOM && !bomChecked {
			if len(buf) < 2 {
				partial := len(buf) == 1 && ((little && buf[0] == 0xFF) || (!little && buf[0] == 0xFE))
				if partial && stream {
					return "", buf, false, nil
				}
				bomChecked = true
			} else {
				if little && buf[0] == 0xFF && buf[1] == 0xFE {
					buf = buf[2:]
				}
				if !little && buf[0] == 0xFE && buf[1] == 0xFF {
					buf = buf[2:]
				}
				bomChecked = true
			}
		}
		text, rest, err := decodeUTF16(buf, little, stream, fatal, enc)
		if err != nil {
			return "", nil, bomChecked, err
		}
		if !stream && len(rest) > 0 {
			text += "\uFFFD"
			rest = nil
		}
		return text, rest, bomChecked, nil
	case "windows-1252":
		return decodeWindows1252(buf), nil, bomChecked, nil
	}
	return "", nil, bomChecked, js.Throw("TypeError", "The encoding is not supported")
}

func utf8BOMPrefix(b []byte) bool {
	bom := []byte{0xEF, 0xBB, 0xBF}
	if len(b) > len(bom) {
		return false
	}
	for i := range b {
		if b[i] != bom[i] {
			return false
		}
	}
	return true
}

func decodeUTF8(buf []byte, stream, fatal bool) (string, []byte, error) {
	var sb strings.Builder
	i := 0
	for i < len(buf) {
		c := buf[i]
		if c < 0x80 {
			sb.WriteByte(c)
			i++
			continue
		}
		n := utf8SeqLen(c)
		if n == 0 {
			if fatal {
				return "", nil, invalidUTF8()
			}
			sb.WriteRune(utf8.RuneError)
			i++
			continue
		}
		if i+n > len(buf) {
			ok := true
			for j := i + 1; j < len(buf); j++ {
				if buf[j]&0xC0 != 0x80 {
					ok = false
					break
				}
			}
			if ok {
				if stream {
					return sb.String(), append([]byte{}, buf[i:]...), nil
				}
				if fatal {
					return "", nil, invalidUTF8()
				}
				sb.WriteRune(utf8.RuneError)
				return sb.String(), nil, nil
			}
			if fatal {
				return "", nil, invalidUTF8()
			}
			sb.WriteRune(utf8.RuneError)
			i++
			continue
		}
		r, size := utf8.DecodeRune(buf[i : i+n])
		if r == utf8.RuneError && size <= 1 {
			if fatal {
				return "", nil, invalidUTF8()
			}
			sb.WriteRune(utf8.RuneError)
			i++
			continue
		}
		sb.WriteRune(r)
		i += size
	}
	return sb.String(), nil, nil
}

func utf8SeqLen(c byte) int {
	switch {
	case c >= 0xC2 && c <= 0xDF:
		return 2
	case c >= 0xE0 && c <= 0xEF:
		return 3
	case c >= 0xF0 && c <= 0xF4:
		return 4
	}
	return 0
}

func invalidUTF8() error {
	return js.ThrowCode("TypeError", "ERR_ENCODING_INVALID_ENCODED_DATA",
		"The encoded data was not valid for encoding utf-8")
}

func decodeUTF16(buf []byte, little, stream, fatal bool, label string) (string, []byte, error) {
	end := len(buf) &^ 1
	var rest []byte
	if end < len(buf) {
		rest = append([]byte{}, buf[end:]...)
	}
	units := make([]uint16, 0, end/2)
	for i := 0; i < end; i += 2 {
		if little {
			units = append(units, uint16(buf[i])|uint16(buf[i+1])<<8)
		} else {
			units = append(units, uint16(buf[i])<<8|uint16(buf[i+1]))
		}
	}
	if stream && len(units) > 0 {
		last := units[len(units)-1]
		if last >= 0xD800 && last <= 0xDBFF {
			units = units[:len(units)-1]
			// Keep the high surrogate and any odd trailing byte, so the
			// next chunk stays byte-aligned with the original stream.
			rest = append([]byte{}, buf[end-2:]...)
		}
	}
	if fatal {
		for i := 0; i < len(units); i++ {
			u := units[i]
			if u >= 0xD800 && u <= 0xDBFF {
				if i+1 >= len(units) || units[i+1] < 0xDC00 || units[i+1] > 0xDFFF {
					return "", nil, invalidUTF16(label)
				}
				i++
				continue
			}
			if u >= 0xDC00 && u <= 0xDFFF {
				return "", nil, invalidUTF16(label)
			}
		}
	}
	return string(utf16.Decode(units)), rest, nil
}

func invalidUTF16(label string) error {
	return js.ThrowCode("TypeError", "ERR_ENCODING_INVALID_ENCODED_DATA",
		"The encoded data was not valid for encoding "+label)
}

// windows1252 maps the C1 range to the characters the WHATWG encoding
// standard assigns to them. Every other byte maps to the same code point.
var windows1252 = [32]rune{
	'\u20AC', '\u0081', '\u201A', '\u0192', '\u201E', '\u2026', '\u2020', '\u2021',
	'\u02C6', '\u2030', '\u0160', '\u2039', '\u0152', '\u008D', '\u017D', '\u008F',
	'\u0090', '\u2018', '\u2019', '\u201C', '\u201D', '\u2022', '\u2013', '\u2014',
	'\u02DC', '\u2122', '\u0161', '\u203A', '\u0153', '\u009D', '\u017E', '\u0178',
}

func decodeWindows1252(buf []byte) string {
	var sb strings.Builder
	for _, c := range buf {
		if c >= 0x80 && c <= 0x9F {
			sb.WriteRune(windows1252[c-0x80])
			continue
		}
		sb.WriteRune(rune(c))
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
// JavaScript prelude

const prelude = `(function (host) {
  'use strict';
  delete globalThis.__godeWeb;

  var EMPTY = new Uint8Array(0);

  function typeError(msg) { return new TypeError(msg); }

  // --- TextEncoder -------------------------------------------------------
  class TextEncoder {
    get encoding() { return 'utf-8'; }
    encode(input) {
      return host.utf8(String(input === undefined ? '' : input));
    }
    encodeInto(source, destination) {
      if (!(destination instanceof Uint8Array)) {
        throw typeError('The "destination" argument must be an instance of Uint8Array');
      }
      var s = String(source === undefined ? '' : source);
      var written = 0;
      var read = 0;
      for (var i = 0; i < s.length;) {
        var cp = s.codePointAt(i);
        var ch = String.fromCodePoint(cp);
        var bytes = host.utf8(ch);
        if (written + bytes.length > destination.length) break;
        destination.set(bytes, written);
        written += bytes.length;
        i += cp > 0xFFFF ? 2 : 1;
        read = i;
      }
      return { read: read, written: written };
    }
  }

  // --- TextDecoder -------------------------------------------------------
  var WINDOWS_1252 = 'windows-1252';

  function canonicalLabel(label) {
    var l = String(label === undefined ? 'utf-8' : label).trim().toLowerCase();
    switch (l) {
      case 'unicode-1-1-utf-8':
      case 'unicode11utf8':
      case 'unicode20utf8':
      case 'utf-8':
      case 'utf8':
      case 'x-unicode20utf8':
        return 'utf-8';
      case 'ansi_x3.4-1968':
      case 'ascii':
      case 'cp1252':
      case 'cp819':
      case 'csisolatin1':
      case 'ibm819':
      case 'iso-8859-1':
      case 'iso-ir-100':
      case 'iso8859-1':
      case 'iso88591':
      case 'iso_8859-1':
      case 'iso_8859-1:1987':
      case 'l1':
      case 'latin1':
      case 'us-ascii':
      case 'windows-1252':
      case 'x-cp1252':
        return WINDOWS_1252;
      case 'csunicode':
      case 'iso-10646-ucs-2':
      case 'ucs-2':
      case 'unicode':
      case 'unicodefeff':
      case 'utf-16':
      case 'utf-16le':
        return 'utf-16le';
      case 'unicodefffe':
      case 'utf-16be':
        return 'utf-16be';
    }
    var err = new RangeError("The encoding label provided ('" + l + "') is invalid.");
    err.code = 'ERR_ENCODING_NOT_SUPPORTED';
    throw err;
  }

  function toBytes(input) {
    if (input === undefined) return EMPTY;
    if (input instanceof ArrayBuffer) return new Uint8Array(input);
    if (typeof SharedArrayBuffer !== 'undefined' && input instanceof SharedArrayBuffer) {
      return new Uint8Array(input);
    }
    if (ArrayBuffer.isView(input)) {
      return new Uint8Array(input.buffer, input.byteOffset, input.byteLength);
    }
    throw typeError('The "input" argument must be an instance of ArrayBuffer or ArrayBufferView');
  }

  class TextDecoder {
    constructor(label, options) {
      this.encoding = canonicalLabel(label);
      this.fatal = !!(options && options.fatal);
      this.ignoreBOM = !!(options && options.ignoreBOM);
      this._pending = EMPTY;
      this._bomChecked = false;
    }
    decode(input, options) {
      var stream = !!(options && options.stream);
      var bytes = toBytes(input);
      var out = host.decode(this.encoding, bytes, this._pending, stream, this.ignoreBOM, this._bomChecked, this.fatal);
      this._pending = out.pending;
      this._bomChecked = out.bomChecked;
      return out.text;
    }
  }

  // --- atob / btoa -------------------------------------------------------
  function atob(input) {
    return host.atob(String(input));
  }
  function btoa(input) {
    return host.btoa(String(input));
  }

  // --- queueMicrotask ----------------------------------------------------
  function queueMicrotask(callback) {
    if (typeof callback !== 'function') {
      throw typeError('The "callback" argument must be of type function');
    }
    Promise.resolve().then(callback);
  }

  // --- structuredClone ---------------------------------------------------
  function dataCloneError(value) {
    var desc;
    try { desc = String(value); } catch (_) { desc = 'value'; }
    var err = new Error(desc + ' could not be cloned.');
    err.name = 'DataCloneError';
    err.code = 25;
    return err;
  }

  function cloneValue(value, seen) {
    if (value === null) return null;
    var t = typeof value;
    if (t === 'undefined' || t === 'boolean' || t === 'number' || t === 'string' || t === 'bigint') {
      return value;
    }
    if (t === 'symbol' || t === 'function') throw dataCloneError(value);
    if (seen.has(value)) return seen.get(value);

    if (value instanceof Error) {
      var clonedErr = new value.constructor(value.message);
      clonedErr.name = value.name;
      seen.set(value, clonedErr);
      if ('cause' in value) clonedErr.cause = cloneValue(value.cause, seen);
      return clonedErr;
    }
    if (Array.isArray(value)) {
      var arr = [];
      seen.set(value, arr);
      for (var i = 0; i < value.length; i++) {
        if (Object.prototype.hasOwnProperty.call(value, i)) {
          arr[i] = cloneValue(value[i], seen);
        }
      }
      return arr;
    }
    if (value instanceof Date) return new Date(value.getTime());
    if (value instanceof RegExp) {
      var re = new RegExp(value.source, value.flags);
      re.lastIndex = value.lastIndex;
      return re;
    }
    if (value instanceof Map) {
      var m = new Map();
      seen.set(value, m);
      value.forEach(function (v, k) { m.set(cloneValue(k, seen), cloneValue(v, seen)); });
      return m;
    }
    if (value instanceof Set) {
      var set = new Set();
      seen.set(value, set);
      value.forEach(function (v) { set.add(cloneValue(v, seen)); });
      return set;
    }
    if (value instanceof ArrayBuffer) return value.slice(0);
    if (typeof SharedArrayBuffer !== 'undefined' && value instanceof SharedArrayBuffer) {
      return value.slice(0);
    }
    if (ArrayBuffer.isView(value)) {
      var copy = value.buffer.slice(0);
      if (value instanceof DataView) {
        return new DataView(copy, value.byteOffset, value.byteLength);
      }
      // Buffer is a Uint8Array subclass with a non-standard constructor;
      // the structured clone algorithm yields a plain Uint8Array.
      var Ctor = value instanceof Uint8Array ? Uint8Array : value.constructor;
      return new Ctor(copy, value.byteOffset, value.length);
    }
    if (value instanceof Promise || value instanceof WeakMap || value instanceof WeakSet) {
      throw dataCloneError(value);
    }

    var obj = {};
    seen.set(value, obj);
    Object.keys(value).forEach(function (k) {
      obj[k] = cloneValue(value[k], seen);
    });
    return obj;
  }

  function structuredClone(value, options) {
    var result = cloneValue(value, new Map());
    if (options && options.transfer != null) {
      var list = options.transfer;
      if (typeof list === 'object' && typeof list.length === 'number') {
        for (var i = 0; i < list.length; i++) {
          var item = list[i];
          if (!(item instanceof ArrayBuffer) &&
              !(typeof SharedArrayBuffer !== 'undefined' && item instanceof SharedArrayBuffer)) {
            throw dataCloneError(item);
          }
          // Node detaches every transferred ArrayBuffer after copying its bytes
          // into the clone; JavaScriptCore exposes that as transfer().
          if (typeof item.transfer === 'function') item.transfer();
        }
      }
    }
    return result;
  }

  // --- performance -------------------------------------------------------
  var perfStart = host.perfNow();
  var marks = [];
  var measures = [];

  function findMark(name) {
    for (var i = marks.length - 1; i >= 0; i--) {
      if (marks[i].name === name) return marks[i];
    }
    return null;
  }

  class Performance {}
  Object.defineProperty(Performance.prototype, Symbol.toStringTag, {
    value: 'Performance', configurable: true
  });

  var performance = {
    timeOrigin: Date.now() - perfStart,
    now: function () { return host.perfNow(); },
    mark: function (name) {
      var entry = { name: String(name), entryType: 'mark', startTime: this.now(), duration: 0, detail: null };
      marks.push(entry);
      return entry;
    },
    measure: function (name, startMark, endMark) {
      var start = this.now();
      var end = start;
      if (typeof startMark === 'string') {
        var foundStart = findMark(startMark);
        if (foundStart) start = foundStart.startTime;
      } else if (startMark && typeof startMark === 'object' && startMark.startTime !== undefined) {
        start = startMark.startTime;
      }
      if (typeof endMark === 'string') {
        var foundEnd = findMark(endMark);
        if (foundEnd) end = foundEnd.startTime;
      } else if (endMark && typeof endMark === 'object' && endMark.startTime !== undefined) {
        end = endMark.startTime;
      }
      var entry = { name: String(name), entryType: 'measure', startTime: start, duration: end - start, detail: null };
      measures.push(entry);
      return entry;
    },
    getEntries: function () {
      return marks.concat(measures).slice().sort(function (a, b) { return a.startTime - b.startTime; });
    },
    getEntriesByName: function (name, type) {
      return this.getEntries().filter(function (e) {
        return e.name === String(name) && (type === undefined || e.entryType === String(type));
      });
    },
    getEntriesByType: function (type) {
      return this.getEntries().filter(function (e) { return e.entryType === String(type); });
    },
    clearMarks: function (name) {
      if (name === undefined) { marks.length = 0; return; }
      marks = marks.filter(function (e) { return e.name !== String(name); });
    },
    clearMeasures: function (name) {
      if (name === undefined) { measures.length = 0; return; }
      measures = measures.filter(function (e) { return e.name !== String(name); });
    },
    toJSON: function () {
      return { nodeTiming: {}, timeOrigin: this.timeOrigin };
    }
  };

  // --- AbortController / AbortSignal -------------------------------------
  var INTERNAL = {};

  function abortError() {
    var err = new Error('This operation was aborted');
    err.name = 'AbortError';
    err.code = 'ABORT_ERR';
    return err;
  }

  function timeoutError() {
    var err = new Error('The operation was aborted due to timeout');
    err.name = 'TimeoutError';
    err.code = 'ERR_TIMEOUT';
    return err;
  }

  class AbortSignal {
    constructor(token) {
      if (token !== INTERNAL) throw typeError('Illegal constructor');
      this._listeners = [];
      this.aborted = false;
      this.reason = undefined;
      this.onabort = null;
    }
    addEventListener(type, listener, options) {
      if (type !== 'abort' || typeof listener !== 'function') return;
      this._listeners.push({
        fn: listener,
        once: !!(options && options.once),
        resist: !!(options && options.resistStopPropagation)
      });
    }
    removeEventListener(type, listener) {
      if (type !== 'abort') return;
      this._listeners = this._listeners.filter(function (l) { return l.fn !== listener; });
    }
    dispatchEvent(event) {
      this._fire(event || { type: 'abort', target: this });
      return true;
    }
    throwIfAborted() {
      if (this.aborted) throw this.reason;
    }
    _fire(event) {
      var list = this._listeners.slice();
      for (var i = 0; i < list.length; i++) {
        var l = list[i];
        // resistStopPropagation marks Node's internal abort listeners, which
        // still run when user code stopped immediate propagation.
        if (event && event._stopImmediate && !l.resist) break;
        if (l.once) this.removeEventListener('abort', l.fn);
        l.fn.call(this, event);
      }
      if (typeof this.onabort === 'function') this.onabort.call(this, event);
    }
    _makeEvent() {
      var ev;
      try { ev = new globalThis.Event('abort'); } catch (e) { ev = { type: 'abort' }; }
      if (typeof ev.stopImmediatePropagation !== 'function') {
        ev.stopImmediatePropagation = function () { ev._stopImmediate = true; };
      }
      if (ev._stopImmediate === undefined) ev._stopImmediate = false;
      try { ev._target = this; ev.target = this; ev.currentTarget = this; } catch (e) {}
      return ev;
    }
    _abort(reason) {
      if (this.aborted) return;
      this.aborted = true;
      this.reason = reason;
      this._fire(this._makeEvent());
    }
    static abort(reason) {
      var s = new AbortSignal(INTERNAL);
      s._abort(reason === undefined ? abortError() : reason);
      return s;
    }
    static timeout(ms) {
      var n = Number(ms);
      if (!isFinite(n) || n < 0) {
        var err = new RangeError('The "delay" argument must be a non-negative finite number.');
        err.code = 'ERR_OUT_OF_RANGE';
        throw err;
      }
      var c = new AbortController();
      host.setTimeout(function () { c.abort(timeoutError()); }, n);
      return c.signal;
    }
    static any(signals) {
      if (signals == null || typeof signals[Symbol.iterator] !== 'function') {
        throw typeError('The "signals" argument must be an iterable of AbortSignal objects');
      }
      var c = new AbortController();
      for (var s of signals) {
        if (!(s instanceof AbortSignal)) {
          throw typeError('The "signals" argument must be an iterable of AbortSignal objects');
        }
        if (s.aborted) {
          c.abort(s.reason);
          break;
        }
        s.addEventListener('abort', (function (sig) {
          return function () { c.abort(sig.reason); };
        })(s), { once: true });
      }
      return c.signal;
    }
  }

  class AbortController {
    constructor() {
      this.signal = new AbortSignal(INTERNAL);
    }
    abort(reason) {
      this.signal._abort(reason === undefined ? abortError() : reason);
    }
  }

  // DOMException is a web-platform global Node installs from internal/deps;
  // JavaScriptCore has no DOM, so provide the shape Node tests rely on:
  // an Error subclass with name, message and the legacy numeric code.
  var DOM_EXCEPTION_CODES = {
    IndexSizeError: 1, HierarchyRequestError: 3, WrongDocumentError: 4,
    InvalidCharacterError: 5, NoModificationAllowedError: 7,
    NotFoundError: 8, NotSupportedError: 9, InvalidStateError: 11,
    SyntaxError: 12, InvalidModificationError: 13, NamespaceError: 14,
    InvalidAccessError: 15, TypeMismatchError: 17, SecurityError: 18,
    NetworkError: 19, AbortError: 20, URLMismatchError: 21,
    QuotaExceededError: 22, TimeoutError: 23, InvalidNodeTypeError: 24,
    DataCloneError: 25
  };
  class DOMException extends Error {
    constructor(message, name) {
      super(message === undefined ? '' : String(message));
      var n = name === undefined ? 'Error' : String(name);
      Object.defineProperty(this, 'name', {
        value: n, writable: true, enumerable: false, configurable: true
      });
      this.code = DOM_EXCEPTION_CODES[n] || 0;
    }
  }
  tagClass(DOMException, 'DOMException');

  // JavaScriptCore already provides Event/EventTarget/CustomEvent; give them
  // the Symbol.toStringTag Node's implementations carry so that
  // Object.prototype.toString reports the class name.
  function tagClass(ctor, name) {
    if (typeof ctor !== 'function' || !ctor.prototype) return;
    try {
      Object.defineProperty(ctor.prototype, Symbol.toStringTag, {
        value: name, writable: false, enumerable: false, configurable: true
      });
    } catch (e) {}
  }
  tagClass(globalThis.Event, 'Event');
  tagClass(globalThis.EventTarget, 'EventTarget');
  tagClass(globalThis.CustomEvent, 'CustomEvent');

  globalThis.TextEncoder = TextEncoder;
  globalThis.TextDecoder = TextDecoder;
  globalThis.atob = atob;
  globalThis.btoa = btoa;
  globalThis.queueMicrotask = queueMicrotask;
  globalThis.structuredClone = structuredClone;
  Object.setPrototypeOf(performance, Performance.prototype);
  globalThis.performance = performance;
  // Node's V8 string limit: String.prototype.repeat must reject results that
  // would exceed it with RangeError: Invalid string length. JavaScriptCore
  // has a larger limit, so enforce the Node-visible bound before allocating.
  var MAX_STRING_LENGTH = 536870888;
  var originalRepeat = String.prototype.repeat;
  String.prototype.repeat = function repeat(count) {
    var n = Number(count);
    if (isFinite(n) && Math.floor(n) === n && n > 0 &&
        this.length * n > MAX_STRING_LENGTH) {
      throw new RangeError('Invalid string length');
    }
    return originalRepeat.call(this, count);
  };

  globalThis.AbortSignal = AbortSignal;
  globalThis.AbortController = AbortController;
  globalThis.DOMException = DOMException;

  // JavaScriptCore only exposes SharedArrayBuffer when the embedder enables
  // it, so alias the non-shared type: every construction and view path still
  // works, and worker_threads never actually shares memory in GodeJS.
  if (typeof globalThis.SharedArrayBuffer === 'undefined' &&
      typeof globalThis.ArrayBuffer === 'function') {
    globalThis.SharedArrayBuffer = globalThis.ArrayBuffer;
  }
})(globalThis.__godeWeb);`
