package env

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"godejs/internal/js"
)

// installStack installs the V8 CallSite layer before any user code runs.
//
// JavaScriptCore only exposes a formatted `fn@file:line:col` stack string, so
// the layer keeps that string as the default `error.stack` (assert, the
// debugger and util.getCallSites all consume it) and only switches to
// structured CallSite objects when Error.prepareStackTrace is a function.
// Every V8 CallSite method is implemented; receiver/function metadata is
// recovered from the caller's source when the engine did not retain it.
func (e *Env) installStack() {
	if _, err := e.JS.Eval(stackPrelude, "[gode:stack]"); err != nil {
		fmt.Fprintf(e.Err(), "gode stack prelude: %v\n", err)
		return
	}
	// CallSite metadata lives on the Go side. Keeping it there avoids a
	// Go-to-JavaScript call while a require() host callback is already on the
	// stack, which JavaScriptCore does not support reliably. The prelude asks
	// for this record lazily, the first time a frame from the file is exposed.
	e.Global().SetFuncHidden("__godeScriptInfo", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsString() {
			return e.JS.Null(), nil
		}
		e.mu.Lock()
		meta := e.stackScripts[args[0].Str()]
		e.mu.Unlock()
		if meta == nil {
			return e.JS.Null(), nil
		}
		obj := e.JS.Obj()
		obj.SetStr("hash", meta.hash)
		obj.SetStr("source", meta.source)
		obj.SetNum("lineOffset", float64(meta.lineOffset))
		obj.SetNum("colShift", float64(meta.colShift))
		if meta.exports.Valid() {
			obj.Set("exports", meta.exports)
		} else {
			obj.Set("exports", e.JS.Undefined())
		}
		return obj.V(), nil
	})
}

// stackScript is the Go-side record the CallSite layer reads on demand.
type stackScript struct {
	hash       string
	source     string
	exports    js.Value
	lineOffset int
	colShift   int
}

// registerScript records the exact source that was compiled, its SHA-256, the
// CommonJS exports object (for the top-level `this`) and the wrapper offsets
// that map engine line numbers back to file coordinates. This is deliberately
// pure Go: the CallSite prelude pulls the record through a host function when
// it first sees a frame from the file.
func (e *Env) registerScript(name, src string, exports js.Value, lineOffset, colShift int) {
	if name == "" || strings.HasPrefix(name, "[") {
		return
	}
	sum := sha256.Sum256([]byte(src))
	e.mu.Lock()
	if e.stackScripts == nil {
		e.stackScripts = map[string]*stackScript{}
	}
	if old := e.stackScripts[name]; old != nil && old.exports.Valid() {
		exports = old.exports
	} else if exports.Valid() {
		// The registry outlives the module, so root the exports object for as
		// long as it can be returned as a frame's receiver.
		exports.Protect()
	}
	e.stackScripts[name] = &stackScript{
		hash:       hex.EncodeToString(sum[:]),
		source:     src,
		exports:    exports,
		lineOffset: lineOffset,
		colShift:   colShift,
	}
	e.mu.Unlock()
}

// stackOffsets measures the wrapper text that precedes a module body.
func stackOffsets(before string) (lineOffset, colShift int) {
	lineOffset = strings.Count(before, "\n")
	colShift = len(before) - strings.LastIndex(before, "\n") - 1
	return
}

const stackPrelude = `(function () {
  'use strict';

  var NativeError = Error;
  var sharedPrepare;
  var scripts = Object.create(null);
  var MAX_DECL_SCAN = 400;

  function detectStrict(lines) {
    for (var i = 0; i < lines.length && i < 20; i++) {
      var t = lines[i].trim();
      if (!t) continue;
      if (t.charAt(0) === '/' || t.charAt(0) === '*') continue;
      if (/^(['"])use strict\1\s*;?$/.test(t)) return true;
      if (/^(import|export|const|let|var|function|class|return|if|for|while|switch|try|throw)\b/.test(t)) return false;
    }
    return false;
  }

  function metaFor(file) {
    if (typeof file !== 'string' || file === '') return null;
    var meta = scripts[file];
    if (meta !== undefined) return meta;
    var info = null;
    try {
      if (typeof globalThis.__godeScriptInfo === 'function') {
        info = globalThis.__godeScriptInfo(file);
      }
    } catch (e) { info = null; }
    if (!info) { scripts[file] = null; return null; }
    var text = String(info.source || '');
    var lines = text.split('\n');
    var starts = new Array(lines.length);
    var off = 0;
    for (var i = 0; i < lines.length; i++) {
      starts[i] = off;
      off += lines[i].length + 1;
    }
    meta = {
      hash: info.hash,
      lines: lines,
      starts: starts,
      exports: info.exports,
      strict: detectStrict(lines),
      lineOffset: info.lineOffset | 0,
      colShift: info.colShift | 0
    };
    scripts[file] = meta;
    return meta;
  }

  function escapeRe(s) {
    return String(s).replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }

  function parseFrame(line) {
    if (!line) return null;
    var at = -1;
    for (var i = 0; i < line.length; i++) {
      if (line.charCodeAt(i) !== 64) continue;
      var rest = line.slice(i + 1);
      if (rest === '[native code]' || rest === '' || /^.*:\d+:\d+$/.test(rest)) {
        at = i;
        break;
      }
    }
    if (at === -1) return null;
    var name = line.slice(0, at);
    var loc = line.slice(at + 1);
    var f = { raw: line, name: name, file: null, line: null, col: null, native: false, eval: false };
    if (loc === '[native code]') {
      f.native = true;
      return f;
    }
    if (loc === '') {
      if (name === 'eval code') {
        f.eval = true;
        f.name = 'eval';
        f.file = '';
        f.line = 1;
        f.col = 1;
      }
      return f;
    }
    var m = /^(.*):(\d+):(\d+)$/.exec(loc);
    if (!m) return null;
    f.file = m[1];
    f.line = Number(m[2]);
    f.col = Number(m[3]);
    if (name === 'eval code') {
      f.eval = true;
      f.name = 'eval';
    }
    return f;
  }

  function normalizeFrame(f) {
    if (f.file === null || f.line === null) return f;
    var meta = metaFor(f.file);
    if (!meta) return f;
    var rawLine = f.line;
    if (rawLine === meta.lineOffset + 1) {
      f.col = Math.max(1, f.col - meta.colShift);
    }
    f.line = rawLine - meta.lineOffset;
    if (f.line < 1) {
      f.line = 1;
      f.col = 1;
    }
    return f;
  }

  function parseStack(raw) {
    var out = [];
    var lines = String(raw || '').split('\n');
    for (var i = 0; i < lines.length; i++) {
      var f = parseFrame(lines[i]);
      if (!f) continue;
      if (f.file === '[gode:stack]') continue;
      if (f.file && f.file.indexOf('[gode:') === 0) continue;
      if (f.native && f.name === 'runMain') continue;
      if (f.native && f.name === 'eval' && out.length && out[out.length - 1].eval) continue;
      out.push(normalizeFrame(f));
    }
    return out;
  }

  function lastMatch(re, text) {
    var found = null, m;
    re.lastIndex = 0;
    while ((m = re.exec(text))) {
      found = m;
      if (m.index === re.lastIndex) re.lastIndex++;
    }
    return found;
  }

  function resolveChain(text) {
    if (!text || text === 'this') return undefined;
    var parts = text.split('.');
    var value;
    if (parts[0] === 'globalThis') {
      value = globalThis;
    } else if (parts[0] in globalThis) {
      value = globalThis[parts[0]];
    } else {
      return undefined;
    }
    for (var i = 1; i < parts.length; i++) {
      if (value === null || value === undefined) return undefined;
      value = value[parts[i]];
    }
    return value;
  }

  function receiverBefore(text, idx) {
    var pre = text.slice(0, idx);
    var m = /(new\s+[A-Za-z_$][\w$]*\s*\(\s*\))\s*$/.exec(pre);
    if (m) {
      return { kind: 'new', text: m[1].replace(/\s+/g, ''), type: m[1].replace(/^new\s+/, '').replace(/\s*\(.*$/, '') };
    }
    m = /([A-Za-z_$][\w$]*(?:\s*\.\s*[A-Za-z_$][\w$]*)*)\s*$/.exec(pre);
    if (m) return { kind: 'chain', text: m[1].replace(/\s+/g, '') };
    return null;
  }

  function findDeclaration(frame) {
    if (!frame.name) return { line: 1, col: 1 };
    var meta = metaFor(frame.file);
    if (!meta || frame.line === null) return null;
    var esc = escapeRe(frame.name);
    var patterns = [
      { kind: 'function', re: new RegExp('\\bfunction\\s+' + esc + '\\b') },
      { kind: 'class', re: new RegExp('\\bclass\\s+' + esc + '\\b') },
      { kind: 'get', re: new RegExp('\\bget\\s+' + esc + '\\s*\\(') },
      { kind: 'set', re: new RegExp('\\bset\\s+' + esc + '\\s*\\(') },
      { kind: 'method', re: new RegExp('(^|[^\\w$.])' + esc + '\\s*\\([^)]*\\)\\s*\\{') },
      { kind: 'assign', re: new RegExp('\\b' + esc + '\\s*[:=]\\s*(?:async\\s*)?(?:function\\b|\\()') }
    ];
    var start = Math.min(frame.line, meta.lines.length) - 1;
    var stop = Math.max(0, start - MAX_DECL_SCAN);
    for (var i = start; i >= stop; i--) {
      var line = meta.lines[i];
      for (var p = 0; p < patterns.length; p++) {
        var m = patterns[p].re.exec(line);
        if (!m) continue;
        var col = m.index + (m[1] ? m[1].length : 0) + 1;
        return { kind: patterns[p].kind, line: i + 1, col: col };
      }
    }
    return null;
  }

  function analyse(frame, frames, index, decl) {
    var info = { method: null, type: null, ctor: false, func: undefined, getThis: undefined, getter: null };
    if (frame.eval) {
      info.getThis = globalThis;
      return info;
    }
    if (frame.native) return info;
    if (!frame.name) {
      var meta = metaFor(frame.file);
      info.getThis = meta ? meta.exports : undefined;
      return info;
    }
    if (decl && (decl.kind === 'get' || decl.kind === 'set')) {
      info.method = frame.name;
      info.getter = decl.kind;
    }

    var caller = frames[index + 1];
    var seg = null, base = 0;
    if (caller && !caller.native && !caller.eval && caller.file && caller.line !== null) {
      var callerMeta = metaFor(caller.file);
      var line = callerMeta ? callerMeta.lines[caller.line - 1] : null;
      if (line != null) {
        base = Math.max(0, (caller.col || 1) - 220);
        seg = line.slice(base, Math.min(line.length, (caller.col || 1) + 30));
      }
    }
    if (seg !== null) {
      var esc = escapeRe(frame.name);
      var ctorM = lastMatch(new RegExp('\\bnew\\s+' + esc + '\\s*\\(', 'g'), seg);
      var dotM = lastMatch(new RegExp('\\.\\s*' + esc + '\\s*\\(', 'g'), seg);
      var qdotM = lastMatch(new RegExp('\\?\\.\\s*' + esc + '\\s*\\(', 'g'), seg);
      var compM = lastMatch(new RegExp('\\[\\s*[\'"]' + esc + '[\'"]\\s*\\]\\s*\\(', 'g'), seg);
      var callM = dotM || qdotM || compM;
      if (callM) {
        info.method = frame.name;
        var recv = receiverBefore(seg, callM.index);
        if (recv) {
          if (recv.kind === 'new') {
            info.type = recv.type;
          } else {
            var value = resolveChain(recv.text);
            if (value !== undefined) info.getThis = value;
            if (value && value.constructor && value.constructor.name) info.type = value.constructor.name;
            else if (/^[A-Z]/.test(recv.text.split('.')[0])) info.type = recv.text.split('.')[0];
            else info.type = 'Object';
            if (value !== null && value !== undefined && typeof value[frame.name] === 'function') info.func = value[frame.name];
          }
        }
      } else if (info.getter) {
        var propM = lastMatch(new RegExp('\\.\\s*' + esc + '\\b', 'g'), seg);
        if (propM) {
          var grecv = receiverBefore(seg, propM.index);
          if (grecv && grecv.kind === 'chain') {
            var gval = resolveChain(grecv.text);
            if (gval !== undefined) info.getThis = gval;
            if (gval && gval.constructor && gval.constructor.name) info.type = gval.constructor.name;
            else if (/^[A-Z]/.test(grecv.text.split('.')[0])) info.type = grecv.text.split('.')[0];
            else info.type = 'Object';
            if (gval !== null && gval !== undefined && typeof gval[frame.name] === 'function') info.func = gval[frame.name];
          }
        }
      }
      if (ctorM && (!callM || ctorM.index >= callM.index)) {
        info.ctor = true;
        info.method = null;
        info.type = null;
        info.getThis = undefined;
        var C = globalThis[frame.name];
        if (typeof C === 'function') info.func = C;
      }
    }

    if (!info.method && !info.ctor) {
      var own = globalThis[frame.name];
      if (typeof own === 'function') info.func = own;
      var ownMeta = metaFor(frame.file);
      if (ownMeta && !ownMeta.strict) info.getThis = globalThis;
    }
    return info;
  }

  function CallSite(frames, index, error) {
    this._frames = frames;
    this._index = index;
    this._frame = frames[index];
    this._error = error;
    this._decl = undefined;
    this._info = null;
  }

  CallSite.prototype._declaration = function () {
    if (this._decl === undefined) this._decl = findDeclaration(this._frame);
    return this._decl;
  };

  CallSite.prototype._analysis = function () {
    if (this._info === null) {
      this._info = analyse(this._frame, this._frames, this._index, this._declaration());
    }
    return this._info;
  };

  CallSite.prototype.getThis = function () {
    return this._analysis().getThis;
  };

  CallSite.prototype.getTypeName = function () {
    if (this._frame.native || this._frame.eval) return null;
    if (!this._frame.name) return 'Object';
    return this._analysis().type;
  };

  CallSite.prototype.getFunction = function () {
    return this._analysis().func;
  };

  CallSite.prototype.getFunctionName = function () {
    if (this._frame.eval) return 'eval';
    if (this._frame.native) return this._frame.name || null;
    var info = this._analysis();
    if (info.getter) return info.getter + ' ' + this._frame.name;
    return this._frame.name || null;
  };

  CallSite.prototype.getMethodName = function () {
    return this._analysis().method;
  };

  CallSite.prototype.getFileName = function () {
    if (this._frame.native) return null;
    if (this._frame.eval) return '';
    return this._frame.file;
  };

  CallSite.prototype.getLineNumber = function () {
    if (this._frame.native) return null;
    return this._frame.line;
  };

  CallSite.prototype.getColumnNumber = function () {
    if (this._frame.native) return null;
    return this._frame.col;
  };

  CallSite.prototype.getEvalOrigin = function () {
    if (!this._frame.eval) return undefined;
    for (var i = this._index + 1; i < this._frames.length; i++) {
      var c = this._frames[i];
      if (c.eval || c.native) continue;
      var name = c.name || '<anonymous>';
      var loc = c.file ? c.file + ':' + c.line + ':' + c.col : '<anonymous>';
      return 'eval at ' + name + ' (' + loc + ')';
    }
    return 'eval at <anonymous>';
  };

  CallSite.prototype.isToplevel = function () {
    if (this._frame.eval) return true;
    if (this._frame.native || !this._frame.name) return false;
    var info = this._analysis();
    if (info.method || info.ctor) return false;
    var next = this._frames[this._index + 1];
    return !next || (!next.name && !next.eval && !next.native);
  };

  CallSite.prototype.isEval = function () {
    return this._frame.eval;
  };

  CallSite.prototype.isNative = function () {
    return this._frame.native;
  };

  CallSite.prototype.isConstructor = function () {
    return this._analysis().ctor;
  };

  CallSite.prototype.isAsync = function () {
    return false;
  };

  CallSite.prototype.isPromiseAll = function () {
    return false;
  };

  CallSite.prototype.getPromiseIndex = function () {
    return null;
  };

  CallSite.prototype.getScriptNameOrSourceURL = function () {
    return this._frame.file || '';
  };

  CallSite.prototype.getScriptHash = function () {
    var meta = metaFor(this._frame.file);
    return meta ? meta.hash : '';
  };

  CallSite.prototype.getEnclosingLineNumber = function () {
    if (this._frame.native) return null;
    if (this._frame.eval) return 1;
    var decl = this._declaration();
    return decl ? decl.line : null;
  };

  CallSite.prototype.getEnclosingColumnNumber = function () {
    if (this._frame.native) return null;
    if (this._frame.eval) return 1;
    var decl = this._declaration();
    return decl ? decl.col : null;
  };

  CallSite.prototype.getPosition = function () {
    if (this._frame.native || this._frame.eval) return 0;
    var meta = metaFor(this._frame.file);
    if (!meta || this._frame.line === null) return 0;
    var start = meta.starts[this._frame.line - 1];
    if (start === undefined) return 0;
    return start + Math.max(0, (this._frame.col || 1) - 1);
  };

  CallSite.prototype.toString = function () {
    if (this._frame.native) return 'native';
    var name = this.getFunctionName();
    if (this._frame.eval) {
      return name + ' (' + this.getEvalOrigin() + ', <anonymous>:1:1)';
    }
    var file = this._frame.file || '<anonymous>';
    var loc = file + ':' + this.getLineNumber() + ':' + this.getColumnNumber();
    if (this._analysis().ctor) return 'new ' + this._frame.name + ' (' + loc + ')';
    var info = this._analysis();
    if (info.getter) return name + ' (' + loc + ')';
    if (info.method && info.type) return info.type + '.' + info.method + ' (' + loc + ')';
    if (name) return name + ' (' + loc + ')';
    if (!this._frame.name) return 'Object.<anonymous> (' + loc + ')';
    return '<anonymous> (' + loc + ')';
  };

  function sites(raw, error) {
    var frames = parseStack(raw);
    var out = new Array(frames.length);
    for (var i = 0; i < frames.length; i++) out[i] = new CallSite(frames, i, error);
    return out;
  }

  function installRaw(target, raw) {
    var state = { raw: String(raw || ''), done: false, value: undefined };
    var desc = {
      configurable: true,
      enumerable: false,
      get: function () {
        if (state.done) return state.value;
        var prep = sharedPrepare;
        if (typeof prep === 'function') {
          state.value = prep.call(undefined, this, sites(state.raw, this));
        } else {
          state.value = state.raw;
        }
        state.done = true;
        return state.value;
      },
      set: function (value) {
        state.value = value;
        state.done = true;
      }
    };
    try {
      delete target.stack;
      Object.defineProperty(target, 'stack', desc);
    } catch (e) {
      try { target.stack = state.raw; } catch (e2) {}
    }
  }

  function captureStackTrace(target, constructorOpt) {
    if (target === null || (typeof target !== 'object' && typeof target !== 'function')) {
      throw new TypeError('Cannot capture stack trace on non-object');
    }
    var tmp = {};
    try {
      NativeError.captureStackTrace(tmp);
    } catch (e) {
      tmp.stack = new NativeError().stack;
    }
    var frames = parseStack(String(tmp.stack || ''));
    if (typeof constructorOpt === 'function' && constructorOpt.name) {
      for (var i = 0; i < frames.length; i++) {
        if (!frames[i].native && frames[i].name === constructorOpt.name) {
          frames = frames.slice(i + 1);
          break;
        }
      }
    }
    var raw = '';
    for (var j = 0; j < frames.length; j++) {
      if (j) raw += '\n';
      raw += frames[j].raw;
    }
    installRaw(target, raw);
  }

  function wrap(Native) {
    function Wrapped(message, options) {
      var target = new.target || Wrapped;
      var err = Reflect.construct(Native, arguments, target);
      var raw = '';
      try { raw = String(err.stack || ''); } catch (e) {}
      installRaw(err, raw);
      return err;
    }
    Object.setPrototypeOf(Wrapped, Native);
    try {
      Object.defineProperty(Wrapped, 'name', { configurable: true, value: Native.name });
      Object.defineProperty(Wrapped, 'length', { configurable: true, value: Native.length });
    } catch (e) {}
    Wrapped.prototype = Native.prototype;
    try {
      Object.defineProperty(Native.prototype, 'constructor', {
        value: Wrapped, writable: true, configurable: true
      });
    } catch (e) {}
    Object.defineProperty(Wrapped, 'prepareStackTrace', {
      configurable: true,
      get: function () { return sharedPrepare; },
      set: function (value) { sharedPrepare = value; }
    });
    Object.defineProperty(Wrapped, 'stackTraceLimit', {
      configurable: true,
      enumerable: true,
      get: function () { return NativeError.stackTraceLimit; },
      set: function (value) { NativeError.stackTraceLimit = value; }
    });
    Object.defineProperty(Wrapped, 'captureStackTrace', {
      configurable: true, writable: true, value: captureStackTrace
    });
    return Wrapped;
  }

  try { NativeError.stackTraceLimit = 10; } catch (e) {}

  var names = ['Error', 'EvalError', 'RangeError', 'ReferenceError', 'SyntaxError', 'TypeError', 'URIError', 'AggregateError'];
  for (var i = 0; i < names.length; i++) {
    var Native = globalThis[names[i]];
    if (typeof Native === 'function') globalThis[names[i]] = wrap(Native);
  }
})()`
