// Package vm implements Node's vm module.
//
// Every vm context is its own JavaScriptCore global context, so code run with
// runInContext/runInNewContext cannot see the caller's globals, locals or
// builtins. The sandbox object stays in the caller's realm and is mirrored in
// and out of the context on each run: primitives and plain data cross exactly,
// live objects cannot (JavaScriptCore cannot share objects between contexts),
// which is the one place this implementation is narrower than V8's
// contextified globals.
package vm

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("vm", Install) }

// vmContext is one isolated JavaScriptCore context. Sandboxes are mirrored
// rather than shared because values cannot cross JSC contexts directly.
type vmContext struct {
	ctx *js.Context
	// intrinsics are the globals the fresh context already had; they are never
	// mirrored back into a sandbox.
	intrinsics map[string]bool
}

var (
	mu       sync.Mutex
	seq      int
	registry = map[int]*vmContext{}
)

func newContext() (int, *vmContext) {
	mu.Lock()
	seq++
	id := seq
	mu.Unlock()
	jsctx := js.New()
	vc := &vmContext{ctx: jsctx, intrinsics: map[string]bool{}}
	for _, k := range jsctx.Global().OwnKeys() {
		vc.intrinsics[k] = true
	}
	mu.Lock()
	registry[id] = vc
	mu.Unlock()
	return id, vc
}

func lookup(id int) (*vmContext, bool) {
	mu.Lock()
	vc, ok := registry[id]
	mu.Unlock()
	return vc, ok
}

// contextID reports the vm context a contextified sandbox belongs to. The
// mapping lives in a WeakMap rather than in hidden sandbox properties so the
// marker never shows up in Object.keys/getOwnPropertyNames/ownKeys.
func (m *module) contextID(v js.Value) (int, bool) {
	if !v.Valid() || !v.IsObject() || !m.ctxMap.Valid() {
		return 0, false
	}
	get := m.ctxMap.Object().Get("get")
	if !get.IsFunction() {
		return 0, false
	}
	idv, err := get.Object().Call(m.ctxMap.Object(), v)
	if err != nil || !idv.Valid() || !idv.IsNumber() {
		return 0, false
	}
	id := idv.Int()
	if _, ok := lookup(id); !ok {
		return 0, false
	}
	return id, true
}

type module struct {
	e *env.Env
	// ctxMap is a WeakMap<object, number> in the caller's realm mapping a
	// contextified sandbox to the id of its vm context.
	ctxMap js.Value
	// dontContextify is the vm.constants.DONT_CONTEXTIFY symbol; passing it to
	// createContext asks for a fresh global instead of contextifying an object.
	dontContextify js.Value
	// script is the vm.Script constructor built by Install, kept so
	// createScript can construct without another module lookup.
	script js.Object
	// Hidden properties are write-once, so the code cache a Script produces
	// later lives here, keyed by the id hidden on the instance.
	cacheMu  sync.Mutex
	cacheSeq int
	caches   map[int][]byte
}

// Install builds the vm module's exports object.
func Install(e *env.Env) js.Object {
	m := &module{e: e, caches: map[int][]byte{}}
	wm, wmErr := e.JS.Eval(`new WeakMap()`, "[gode:vm]")
	if wmErr != nil {
		panic("godejs vm: " + wmErr.Error())
	}
	m.ctxMap = wm.Keep()
	exports := e.JS.Obj()

	// Node ships two well-known symbols on a frozen, null-prototype object
	// (lib/vm.js vmConstants).
	constsVal, err := e.JS.Eval(`(function () {
		const c = Object.create(null);
		c.USE_MAIN_CONTEXT_DEFAULT_LOADER = Symbol('USE_MAIN_CONTEXT_DEFAULT_LOADER');
		c.DONT_CONTEXTIFY = Symbol('DONT_CONTEXTIFY');
		return Object.freeze(c);
	})()`, "[gode:vm]")
	if err != nil {
		panic("godejs vm: " + err.Error())
	}
	consts := constsVal.Object()
	m.dontContextify = consts.Get("DONT_CONTEXTIFY").Keep()
	exports.Set("constants", consts.V())

	exports.SetFunc("createContext", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.createContext(args)
	})
	exports.SetFunc("isContext", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.isContext(arg(args, 0))
	})
	exports.SetFunc("runInContext", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.runInContext(args)
	})
	exports.SetFunc("runInNewContext", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.runInNewContext(args)
	})
	// Node's runInNewContext is a constructible function; `new` returns the
	// evaluated value because the callback result is an object.
	if run := exports.Get("runInNewContext"); run.IsFunction() {
		if fn, err := e.JS.Eval(`(function (native) {
  return function runInNewContext(code, sandbox, options) {
    return native.call(this, code, sandbox, options);
  };
})`, "[gode:vm]"); err == nil {
			if wrapped, callErr := fn.Object().Call(js.Object{}, run); callErr == nil {
				exports.Set("runInNewContext", wrapped)
			}
		}
	}
	exports.SetFunc("runInThisContext", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.runInThisContext(args)
	})
	exports.SetFunc("createScript", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.createScript(args)
	})
	exports.SetFunc("compileFunction", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.compileFunction(args)
	})
	exports.SetFunc("measureMemory", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.measureMemory(args)
	})

	m.script = m.scriptConstructor().Keep()
	exports.Set("Script", m.script.V())
	// The baseline also tracks vm.DONT_CONTEXTIFY at the top level; Node keeps the
	// canonical symbol on vm.constants but the alias is equivalent.
	exports.Set("DONT_CONTEXTIFY", m.dontContextify)

	// vm.Module and its SourceTextModule/SyntheticModule subclasses. JavaScriptCore
	// has no ESM graph API, so these implement Node's documented observable shape
	// and evaluate module source by stripping the import/export syntax.
	maker, modErr := e.JS.Eval(vmModulesPrelude, "[gode:vm modules]")
	if modErr != nil {
		panic("godejs vm modules: " + modErr.Error())
	}
	modClasses, callErr := maker.Object().Call(js.Object{}, m.ctxMap)
	if callErr != nil {
		panic("godejs vm modules: " + callErr.Error())
	}
	for _, name := range []string{"Module", "SourceTextModule", "SyntheticModule"} {
		exports.Set(name, modClasses.Object().Get(name))
	}
	return exports
}

// vmModulesPrelude builds vm.Module, vm.SourceTextModule and vm.SyntheticModule.
// It follows lib/vm.js's lifecycle (unlinked -> linked -> evaluating ->
// evaluated/errored) and keeps namespaces and export bags in sync.
const vmModulesPrelude = `(function (ctxMap) {
  'use strict';
  var seq = 0;
  var creatingModule = false;
  var mainContextKey = {};
  var contextCounters = new WeakMap();

  function describeValue(v) {
    if (v === null) return 'null';
    var t = typeof v;
    if (t === 'undefined') return 'undefined';
    if (t === 'string') return "type string ('" + v + "')";
    if (t === 'number' || t === 'boolean' || t === 'bigint') return 'type ' + t + ' (' + String(v) + ')';
    if (t === 'symbol') return 'type symbol (' + String(v) + ')';
    if (t === 'function') return 'function ' + (v.name || '<anonymous>');
    if (Array.isArray(v)) return 'an instance of Array';
    if (v && v.constructor && v.constructor.name) return 'an instance of ' + v.constructor.name;
    return 'an instance of Object';
  }
  function moduleTypeError(name, expected, v) {
    var err = new TypeError('The "' + name + '" argument must be of type ' + expected +
      '. Received ' + describeValue(v));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }
  function moduleInstanceError(name, expected, v) {
    var err = new TypeError('The "' + name + '" argument must be an instance of ' + expected +
      '. Received ' + describeValue(v));
    err.code = 'ERR_INVALID_ARG_TYPE';
    return err;
  }
  function invalidThisError() {
    var err = new TypeError('Illegal invocation');
    err.code = 'ERR_INVALID_THIS';
    return err;
  }
  function moduleStatusError(message) {
    var err = new Error(message);
    err.code = 'ERR_VM_MODULE_STATUS';
    return err;
  }
  function linkFailureError(cause) {
    var err = new Error('Module has not been linked to all dependencies');
    err.code = 'ERR_VM_MODULE_LINK_FAILURE';
    if (cause !== undefined) err.cause = cause;
    return err;
  }
  function checkModuleThis(self) {
    if (!self || !(self instanceof Module) || self._vmModule !== true) throw invalidThisError();
  }
  function nextIdentifier(context) {
    var key = (context === undefined || context === null) ? mainContextKey : context;
    var n = contextCounters.get(key) || 0;
    contextCounters.set(key, n + 1);
    return n;
  }
  function isBufferView(v) {
    if (typeof Buffer === 'function' && typeof Buffer.isBuffer === 'function' && Buffer.isBuffer(v)) return true;
    return typeof ArrayBuffer === 'function' && typeof ArrayBuffer.isView === 'function' && ArrayBuffer.isView(v);
  }
  function splitTop(text, sep) {
    var out = [];
    var depth = 0;
    var cur = '';
    for (var i = 0; i < text.length; i++) {
      var ch = text.charAt(i);
      if (ch === '(' || ch === '[' || ch === '{') depth++;
      else if (ch === ')' || ch === ']' || ch === '}') depth--;
      if (ch === sep && depth === 0) { out.push(cur); cur = ''; continue; }
      cur += ch;
    }
    if (cur.trim() !== '') out.push(cur);
    return out;
  }
  function moduleAttributes(text) {
    var attrs = Object.create(null);
    if (!text) return attrs;
    var re = /([A-Za-z_$][\w$]*)\s*:\s*(['"])([^'"]*)\2/g;
    var m;
    while ((m = re.exec(text)) !== null) attrs[m[1]] = m[3];
    return attrs;
  }
  function attributeKey(attrs) {
    var keys = Object.keys(attrs).sort();
    var parts = [];
    for (var i = 0; i < keys.length; i++) parts.push(keys[i] + '=' + attrs[keys[i]]);
    return parts.join(',');
  }
  function freezeRequest(req) {
    var attrs = Object.create(null);
    var keys = Object.keys(req.attributes || {});
    for (var i = 0; i < keys.length; i++) attrs[keys[i]] = req.attributes[keys[i]];
    Object.freeze(attrs);
    var out = Object.create(null);
    out.specifier = req.specifier;
    out.attributes = attrs;
    out.phase = req.phase;
    return Object.freeze(out);
  }
  function freezeRequests(list) {
    return Object.freeze(list.map(freezeRequest));
  }
  function hasTopLevelAwait(source) {
    var src = String(source);
    var clean = '';
    var i = 0;
    while (i < src.length) {
      var c = src.charAt(i);
      var n = src.charAt(i + 1);
      if (c === '/' && n === '/') {
        while (i < src.length && src.charAt(i) !== '\n') i++;
        continue;
      }
      if (c === '/' && n === '*') {
        i += 2;
        while (i < src.length && !(src.charAt(i) === '*' && src.charAt(i + 1) === '/')) i++;
        i += 2;
        continue;
      }
      if (c === '"' || c === "'" || c === String.fromCharCode(96)) {
        var quote = c;
        i++;
        while (i < src.length && src.charAt(i) !== quote) {
          if (src.charAt(i) === '\\') i++;
          i++;
        }
        i++;
        clean += ' ';
        continue;
      }
      clean += c;
      i++;
    }
    var tokens = clean.match(/[A-Za-z_$][\w$]*|=>|[{}(),;]/g) || [];
    var fnDepth = 0;
    var stack = [];
    var pendingFn = false;
    for (var t = 0; t < tokens.length; t++) {
      var tok = tokens[t];
      if (tok === 'function' || tok === '=>') { pendingFn = true; continue; }
      if (tok === '{') {
        stack.push(pendingFn ? 'f' : 'b');
        if (pendingFn) fnDepth++;
        pendingFn = false;
        continue;
      }
      if (tok === '}') {
        var kind = stack.pop();
        if (kind === 'f') fnDepth--;
        continue;
      }
      if (tok === ';') pendingFn = false;
      if (tok === 'await' && fnDepth === 0) return true;
    }
    return false;
  }
  function parseModuleSource(source) {
    var code = String(source);
    var own = [];
    var reexports = Object.create(null);
    var stars = [];
    var requests = [];
    var seen = Object.create(null);
    var assignments = [];
    var localExports = [];

    var importRe = /import\s+(source\s+)?(?:[^'";]*?\sfrom\s+)?(['"])([^'"]+)\2(\s*(?:with|assert)\s*\{([^}]*)\})?/g;
    var im;
    while ((im = importRe.exec(code)) !== null) {
      var phase = im[1] ? 'source' : 'evaluation';
      var spec = im[3];
      var attrs = moduleAttributes(im[5] || '');
      var key = phase + '|' + spec + '|' + attributeKey(attrs);
      if (seen[key]) continue;
      seen[key] = true;
      requests.push({ specifier: spec, attributes: attrs, phase: phase });
    }
    var reReq = /export\s*(?:\{[^}]*\}|\*)\s*from\s*(['"])([^'"]+)\1(\s*(?:with|assert)\s*\{([^}]*)\})?/g;
    var rm;
    while ((rm = reReq.exec(code)) !== null) {
      var spec2 = rm[2];
      var attrs2 = moduleAttributes(rm[4] || '');
      var key2 = 'evaluation|' + spec2 + '|' + attributeKey(attrs2);
      if (seen[key2]) continue;
      seen[key2] = true;
      requests.push({ specifier: spec2, attributes: attrs2, phase: 'evaluation' });
    }

    code = code.replace(/^[ \t]*import\s+[^;\n]*;?/gm, '');
    code = code.replace(/export\s+default\s+/g, 'exports.default = ');
    if (code.indexOf('exports.default = ') !== -1) own.push('default');

    code = code.replace(/export\s*\{([^}]*)\}\s*from\s*(['"])([^'"]+)\2\s*;?/g,
      function (m, list, q, spec) {
        splitTop(list, ',').forEach(function (part) {
          var bits = part.trim().split(/\s+as\s+/);
          var out = (bits[1] || bits[0] || '').trim();
          if (!out) return;
          own.push(out);
          reexports[out] = spec;
        });
        return '';
      });
    code = code.replace(/export\s*\{([^}]*)\}\s*;?/g, function (m, list) {
      splitTop(list, ',').forEach(function (part) {
        var bits = part.trim().split(/\s+as\s+/);
        var local = (bits[0] || '').trim();
        var out = (bits[1] || bits[0] || '').trim();
        if (!out) return;
        own.push(out);
        localExports.push('exports.' + out + ' = ' + local + ';');
      });
      return '';
    });
    code = code.replace(/export\s*\*\s*from\s*(['"])([^'"]+)\1\s*;?/g,
      function (m, q, spec) { stars.push(spec); return ''; });

    var declRe = /export\s+(const|let|var)\s+([^;\n]+)/g;
    var dm;
    while ((dm = declRe.exec(code)) !== null) {
      splitTop(dm[2], ',').forEach(function (part) {
        var name = part.trim().split('=')[0].trim();
        if (/^[A-Za-z_$][\w$]*$/.test(name)) {
          own.push(name);
          assignments.push(name);
        }
      });
    }
    code = code.replace(/export\s+(const|let|var)\s+/g, '$1 ');

    var fnRe = /export\s+((?:async\s+)?function\s+([A-Za-z_$][\w$]*)|class\s+([A-Za-z_$][\w$]*))/g;
    var fm;
    while ((fm = fnRe.exec(code)) !== null) {
      var fname = fm[2] || fm[3];
      if (fname) { own.push(fname); assignments.push(fname); }
    }
    code = code.replace(/export\s+((?:async\s+)?function|class)\s+/g, '$1 ');

    for (var a = 0; a < assignments.length; a++) {
      code += '\nexports.' + assignments[a] + ' = ' + assignments[a] + ';';
    }
    for (var l = 0; l < localExports.length; l++) code += '\n' + localExports[l];

    return {
      code: code,
      ownNames: own,
      reexports: reexports,
      stars: stars,
      requests: requests,
      hasTLA: hasTopLevelAwait(source)
    };
  }
  function getExportValue(mod, name) {
    if (mod._reexportFrom && mod._reexportFrom[name] !== undefined) {
      var dep = mod._depBySpecifier && mod._depBySpecifier[mod._reexportFrom[name]];
      if (dep) return dep.namespace[name];
      return undefined;
    }
    if (mod._stars) {
      for (var i = 0; i < mod._stars.length; i++) {
        var d = mod._depBySpecifier && mod._depBySpecifier[mod._stars[i]];
        if (d && d._exportNames && d._exportNames.indexOf(name) !== -1) return d.namespace[name];
      }
    }
    if (!mod._initialized) throw new ReferenceError("Cannot access '" + name + "' before initialization");
    return mod._exports[name];
  }
  function makeNamespace(mod) {
    var ns = Object.create(null);
    var names = (mod._exportNames || []).slice().sort();
    for (var i = 0; i < names.length; i++) {
      (function (name) {
        Object.defineProperty(ns, name, {
          enumerable: true,
          configurable: false,
          get: function () { return getExportValue(mod, name); }
        });
      })(names[i]);
    }
    Object.defineProperty(ns, Symbol.toStringTag, { value: 'Module' });
    return Object.freeze(ns);
  }
  function computeExportNames(mod) {
    var names = (mod._ownExportNames || []).slice();
    for (var name in mod._reexportFrom) {
      if (names.indexOf(name) === -1) names.push(name);
    }
    for (var s = 0; s < (mod._stars || []).length; s++) {
      var dep = mod._depBySpecifier && mod._depBySpecifier[mod._stars[s]];
      if (!dep || !dep._exportNames) continue;
      for (var d = 0; d < dep._exportNames.length; d++) {
        var n = dep._exportNames[d];
        if (n !== 'default' && names.indexOf(n) === -1) names.push(n);
      }
    }
    return names;
  }
  function instantiateModule(mod, seen) {
    if (seen.indexOf(mod) !== -1) return;
    seen.push(mod);
    if (mod._status === 'errored') throw linkFailureError(mod._error);
    mod._depBySpecifier = Object.create(null);
    if (mod._linkedModules && mod._requests) {
      for (var i = 0; i < mod._requests.length && i < mod._linkedModules.length; i++) {
        mod._depBySpecifier[mod._requests[i].specifier] = mod._linkedModules[i];
        instantiateModule(mod._linkedModules[i], seen);
      }
    }
    if (mod._exportNames === undefined) mod._exportNames = computeExportNames(mod);
  }

  function stripModuleSyntax(source) {
    var exported = [];
    var specs = [];
    var out = String(source);
    out = out.replace(/^\s*import\s+[^;\n]*?from\s+['"]([^'"]+)['"]\s*;?/gm,
      function (m, spec) { specs.push(spec); return ''; });
    out = out.replace(/^\s*import\s*['"]([^'"]+)['"]\s*;?/gm,
      function (m, spec) { specs.push(spec); return ''; });
    out = out.replace(/^\s*export\s+default\s+/gm, 'globalThis.__vmDefaultExport = ');
    out = out.replace(/^\s*export\s+(const|let|var|function|class)\s+([A-Za-z_$][\w$]*)/gm,
      function (m, kind, name) { exported.push(name); return m.replace(/^\s*export\s+/, ''); });
    out = out.replace(/^\s*export\s*\{([^}]*)\}\s*;?/gm, function (m, list) {
      list.split(',').forEach(function (part) {
        var n = part.trim().split(/\s+as\s+/).pop();
        if (n) exported.push(n);
      });
      return '';
    });
    if (/globalThis\.__vmDefaultExport/.test(out)) exported.push('default');
    return { code: out, exports: exported, specs: specs };
  }

  function Module(options) {
    if (creatingModule !== true) throw new TypeError('Module is not a constructor');
    options = options || {};
    if (options.identifier !== undefined && typeof options.identifier !== 'string') {
      throw moduleTypeError('options.identifier', 'string', options.identifier);
    }
    if (options.context !== undefined &&
        (options.context === null || typeof options.context !== 'object' || !ctxMap.has(options.context))) {
      throw moduleInstanceError('options.context', 'vm.Context', options.context);
    }
    this._vmModule = true;
    this._identifier = options.identifier !== undefined ? options.identifier : 'vm:module(' + nextIdentifier(options.context) + ')';
    this._context = options.context;
    this._status = 'unlinked';
    this._error = undefined;
    this._exports = Object.create(null);
    this._namespace = undefined;
    this._dependencySpecifiers = [];
    this._requests = [];
    this._requestsFrozen = Object.freeze([]);
    this._requestsLinked = false;
    this._linkedModules = [];
    this._exportNames = undefined;
    this._initialized = false;
  }
  Object.defineProperty(Module.prototype, 'identifier', {
    configurable: true,
    get: function () { checkModuleThis(this); return this._identifier; }
  });
  Object.defineProperty(Module.prototype, 'context', {
    configurable: true,
    get: function () { checkModuleThis(this); return this._context; }
  });
  Object.defineProperty(Module.prototype, 'status', {
    configurable: true,
    get: function () { checkModuleThis(this); return this._status; }
  });
  Object.defineProperty(Module.prototype, 'error', {
    configurable: true,
    get: function () {
      checkModuleThis(this);
      if (this._status !== 'errored') throw moduleStatusError('Module status must be errored');
      return this._error;
    }
  });
  Object.defineProperty(Module.prototype, 'namespace', {
    configurable: true,
    get: function () {
      checkModuleThis(this);
      if (this._status === 'unlinked' || this._status === 'linking') {
        throw moduleStatusError('Module status must not be unlinked or linking');
      }
      if (this._namespace === undefined) this._namespace = makeNamespace(this);
      return this._namespace;
    }
  });
  Object.defineProperty(Module.prototype, 'dependencySpecifiers', {
    configurable: true,
    get: function () { checkModuleThis(this); return this._dependencySpecifiers; }
  });
  Object.defineProperty(Module.prototype, 'moduleRequests', {
    configurable: true,
    get: function () { checkModuleThis(this); return this._requestsFrozen; }
  });
  Module.prototype.link = function link(linker) {
    checkModuleThis(this);
    if (typeof linker !== 'function') return Promise.reject(moduleTypeError('linker', 'function', linker));
    if (this._status === 'linked' || this._status === 'evaluated') {
      var already = new Error('Module has already been linked');
      already.code = 'ERR_VM_MODULE_ALREADY_LINKED';
      return Promise.reject(already);
    }
    if (this._status !== 'unlinked') return Promise.reject(moduleStatusError('Module status must be unlinked'));
    this._status = 'linking';
    try {
      if (this._requests.length > 0) {
        var result = linker(this, this._requestsFrozen);
        if (result === undefined) throw linkFailureError();
        var deps = Array.isArray(result) ? result : [result];
        for (var i = 0; i < deps.length; i++) {
          var dep = deps[i];
          if (!dep || !(dep instanceof Module) || dep._vmModule !== true) {
            var notModule = new TypeError('The argument must be an instance of Module');
            notModule.code = 'ERR_VM_MODULE_NOT_MODULE';
            throw notModule;
          }
          if ((dep._context || null) !== (this._context || null)) {
            var different = new TypeError('The modules must be linked to the same context');
            different.code = 'ERR_VM_MODULE_DIFFERENT_CONTEXT';
            throw different;
          }
          if (dep._status === 'errored') throw linkFailureError(dep._error);
        }
        this._linkedModules = deps;
      }
      this._requestsLinked = true;
      this._status = 'linked';
      return Promise.resolve();
    } catch (err) {
      this._status = 'errored';
      this._error = err;
      return Promise.reject(err);
    }
  };
  Module.prototype.instantiate = function instantiate() {
    checkModuleThis(this);
    if (this._status !== 'unlinked' && this._status !== 'linking') {
      throw moduleStatusError('Module status must be one of unlinked or linking');
    }
    if (!this._requestsLinked) throw linkFailureError();
    instantiateModule(this, []);
    this._status = 'linked';
    return undefined;
  };
  Module.prototype.evaluate = function evaluate(options) {
    checkModuleThis(this);
    if (options !== undefined && (options === null || typeof options !== 'object')) {
      return Promise.reject(moduleTypeError('options', 'object', options));
    }
    if (options && options.breakOnSigint !== undefined && typeof options.breakOnSigint !== 'boolean') {
      return Promise.reject(moduleTypeError('options.breakOnSigint', 'boolean', options.breakOnSigint));
    }
    if (options && options.timeout !== undefined) {
      var t = options.timeout;
      if (typeof t !== 'number') return Promise.reject(moduleTypeError('options.timeout', 'number', t));
      if (!(t >= 1) || !isFinite(t) || Math.floor(t) !== t) {
        var range = new RangeError('The value of "options.timeout" is out of range. It must be >= 1 && <= 4294967295. Received ' + String(t));
        range.code = 'ERR_OUT_OF_RANGE';
        return Promise.reject(range);
      }
    }
    if (this._status === 'errored') return Promise.reject(this._error);
    if (this._status === 'evaluated') return Promise.resolve(undefined);
    if (this._status !== 'linked') {
      return Promise.reject(moduleStatusError('Module status must be one of linked, evaluated, or errored'));
    }
    return this._evaluateNow();
  };
  Module.prototype._evaluateNow = function () { return Promise.resolve(undefined); };

  function SourceTextModule(source, options) {
    if (!(this instanceof SourceTextModule)) return new SourceTextModule(source, options);
    if (typeof source !== 'string') throw moduleTypeError('source', 'string', source);
    if (options !== undefined && (options === null || typeof options !== 'object')) {
      throw moduleTypeError('options', 'object', options);
    }
    var opts = options || {};
    if (opts.importModuleDynamically !== undefined && typeof opts.importModuleDynamically !== 'function') {
      throw moduleTypeError('options.importModuleDynamically', 'function', opts.importModuleDynamically);
    }
    if (opts.cachedData !== undefined && !isBufferView(opts.cachedData)) {
      throw moduleInstanceError('options.cachedData', 'Buffer, TypedArray, or DataView', opts.cachedData);
    }
    creatingModule = true;
    try { Module.call(this, opts); } finally { creatingModule = false; }
    var parsed = parseModuleSource(source);
    this._source = source;
    this._code = parsed.code;
    this._ownExportNames = parsed.ownNames;
    this._reexportFrom = parsed.reexports;
    this._stars = parsed.stars;
    this._hasTLA = parsed.hasTLA;
    this._requests = parsed.requests;
    this._dependencySpecifiers = parsed.requests.map(function (req) { return req.specifier; });
    this._requestsFrozen = freezeRequests(parsed.requests);
    this._exportNames = parsed.ownNames.slice();
  }
  Object.setPrototypeOf(SourceTextModule.prototype, Module.prototype);
  Object.defineProperty(SourceTextModule.prototype, 'constructor', {
    value: SourceTextModule, writable: true, configurable: true
  });
  SourceTextModule.prototype.createCachedData = function createCachedData() {
    return Buffer.alloc(0);
  };
  SourceTextModule.prototype.hasTopLevelAwait = function hasTopLevelAwait() {
    return this._hasTLA === true;
  };
  SourceTextModule.prototype.hasAsyncGraph = function hasAsyncGraph() {
    if (this._hasTLA) return true;
    var deps = this._linkedModules || [];
    for (var i = 0; i < deps.length; i++) {
      if (deps[i] && typeof deps[i].hasAsyncGraph === 'function' && deps[i].hasAsyncGraph()) return true;
    }
    return false;
  };
  SourceTextModule.prototype.linkRequests = function linkRequests(requests) {
    checkModuleThis(this);
    if (!Array.isArray(requests)) throw moduleInstanceError('requests', 'Array', requests);
    for (var i = 0; i < requests.length; i++) {
      var dep = requests[i];
      if (!dep || !(dep instanceof Module) || dep._vmModule !== true) {
        throw moduleTypeError('requests[' + i + ']', 'Module', dep);
      }
    }
    if (requests.length !== this._requests.length) throw linkFailureError();
    this._linkedModules = requests.slice();
    this._requestsLinked = true;
    return undefined;
  };
  SourceTextModule.prototype._evaluateNow = function () {
    var mod = this;
    var body = this._code;
    if (this._hasTLA) {
      this._status = 'evaluating';
      var runner;
      try {
        runner = new Function('exports', 'return (async function () {\n' + body + '\n}).call(globalThis);');
      } catch (err) {
        this._status = 'errored';
        this._error = err;
        return Promise.reject(err);
      }
      var pending;
      try { pending = runner(this._exports); }
      catch (err) {
        this._status = 'errored';
        this._error = err;
        return Promise.reject(err);
      }
      return pending.then(function () {
        mod._status = 'evaluated';
        mod._initialized = true;
        mod._namespace = undefined;
        return undefined;
      }, function (err) {
        mod._status = 'errored';
        mod._error = err;
        throw err;
      });
    }
    this._status = 'evaluating';
    try {
      var fn = new Function('exports', body);
      fn(this._exports);
      this._status = 'evaluated';
      this._initialized = true;
      this._namespace = undefined;
      return Promise.resolve(undefined);
    } catch (err) {
      this._status = 'errored';
      this._error = err;
      return Promise.reject(err);
    }
  };

  function SyntheticModule(exportNames, evaluateCallback, options) {
    if (!(this instanceof SyntheticModule)) {
      return new SyntheticModule(exportNames, evaluateCallback, options);
    }
    if (!Array.isArray(exportNames)) {
      var exportErr = new TypeError('The "exportNames" argument must be an Array of unique strings. Received ' +
        describeValue(exportNames));
      exportErr.code = 'ERR_INVALID_ARG_TYPE';
      throw exportErr;
    }
    var seenNames = Object.create(null);
    for (var i = 0; i < exportNames.length; i++) {
      var name = exportNames[i];
      if (typeof name !== 'string') throw moduleTypeError('exportNames[' + i + ']', 'string', name);
      if (seenNames[name]) {
        throw new TypeError("The property 'exportNames." + name + "' is duplicated. Received '" + name + "'");
      }
      seenNames[name] = true;
    }
    if (typeof evaluateCallback !== 'function') {
      throw moduleTypeError('evaluateCallback', 'function', evaluateCallback);
    }
    if (options !== undefined && (options === null || typeof options !== 'object')) {
      throw moduleTypeError('options', 'object', options);
    }
    creatingModule = true;
    try { Module.call(this, options || {}); } finally { creatingModule = false; }
    this._exportNames = exportNames.slice();
    this._ownExportNames = exportNames.slice();
    this._callback = evaluateCallback;
    this._status = 'linked';
    this._requestsLinked = true;
    this._initialized = true;
  }
  Object.setPrototypeOf(SyntheticModule.prototype, Module.prototype);
  Object.defineProperty(SyntheticModule.prototype, 'constructor', {
    value: SyntheticModule, writable: true, configurable: true
  });
  SyntheticModule.prototype.setExport = function setExport(name, value) {
    checkModuleThis(this);
    if (this._exportNames.indexOf(name) === -1) {
      throw new TypeError('The "name" argument must be a valid export name');
    }
    this._exports[name] = value;
    this._namespace = undefined;
    return undefined;
  };
  SyntheticModule.prototype._evaluateNow = function () {
    var mod = this;
    this._status = 'evaluating';
    var result;
    try {
      result = this._callback.call(this);
    } catch (err) {
      this._status = 'errored';
      this._error = err;
      return Promise.reject(err);
    }
    if (result && typeof result.then === 'function') {
      result.then(function () {}, function (err) { mod._error = err; });
    }
    this._status = 'evaluated';
    this._initialized = true;
    this._namespace = undefined;
    return Promise.resolve(undefined);
  };

  // Node's Module inspect shows only the three public fields; the internal
  // bookkeeping properties stay out of the rendered form.
  function showValue(v) {
    if (typeof v === 'string') return "'" + v + "'";
    if (typeof v === 'function') return '[Function: ' + (v.name || '(anonymous)') + ']';
    if (v === null) return 'null';
    if (typeof v === 'object') {
      var keys = Object.keys(v);
      if (keys.length === 0) return '{}';
      return '{ ' + keys.map(function (k) { return k + ': ' + showValue(v[k]); }).join(', ') + ' }';
    }
    return String(v);
  }
  Object.defineProperty(Module.prototype, Symbol.for('nodejs.util.inspect.custom'), {
    configurable: true,
    value: function () {
      checkModuleThis(this);
      return this.constructor.name + ' {\n  status: ' + showValue(this._status) +
        ',\n  identifier: ' + showValue(this._identifier) +
        ',\n  context: ' + showValue(this._context) + '\n}';
    }
  });

  return { Module: Module, SourceTextModule: SourceTextModule, SyntheticModule: SyntheticModule };
})`

func (m *module) isDontContextify(v js.Value) bool {
	return v.Valid() && m.dontContextify.Valid() && v.Same(m.dontContextify)
}

func (m *module) createContext(args []js.Value) (js.Value, error) {
	sandboxArg := arg(args, 0)
	fresh := !sandboxArg.Valid() || sandboxArg.IsUndefined() || m.isDontContextify(sandboxArg)
	if !fresh {
		if !sandboxArg.IsObject() {
			return js.Value{}, invalidTypeArg("object", "object", sandboxArg)
		}
		if _, ok := m.contextID(sandboxArg); ok {
			// createContext is idempotent, and Node returns the existing context
			// before it looks at the options (lib/vm.js createContext).
			return sandboxArg, nil
		}
	}
	if len(args) > 1 {
		if _, err := m.contextOptions(args[1]); err != nil {
			return js.Value{}, err
		}
	}

	id, vc := newContext()
	sandbox := m.e.JS.Obj()
	if !fresh {
		sandbox = sandboxArg.Object()
		m.copyInto(vc.ctx.Global(), sandbox, vc.ctx)
	}
	if set := m.ctxMap.Object().Get("set"); set.IsFunction() {
		if _, err := set.Object().Call(m.ctxMap.Object(), sandbox.V(), m.e.JS.Num(float64(id))); err != nil {
			return js.Value{}, err
		}
	}
	return sandbox.V(), nil
}

// contextOptions validates the options object createContext accepts.
func (m *module) contextOptions(v js.Value) (js.Object, error) {
	if !v.Valid() || v.IsUndefined() {
		return m.e.JS.Obj(), nil
	}
	// validateObject only checks the typeof result and null, so arrays are
	// accepted as option bags in every vm entry point.
	if !v.IsObject() {
		return js.Object{}, invalidTypeArg("options", "object", v)
	}
	o := v.Object()
	if _, err := stringOpt(o, "options.name", ""); err != nil {
		return js.Object{}, err
	}
	if origin := o.Get("origin"); origin.Valid() && !origin.IsUndefined() {
		if !origin.IsString() {
			return js.Object{}, invalidTypeArg("options.origin", "string", origin)
		}
	}
	if cg := o.Get("codeGeneration"); cg.Valid() && !cg.IsUndefined() {
		if !cg.IsObject() {
			return js.Object{}, invalidTypeArg("options.codeGeneration", "object", cg)
		}
		cgo := cg.Object()
		if _, err := boolOpt(cgo, "options.codeGeneration.strings", true); err != nil {
			return js.Object{}, err
		}
		if _, err := boolOpt(cgo, "options.codeGeneration.wasm", true); err != nil {
			return js.Object{}, err
		}
	}
	if mm := o.Get("microtaskMode"); mm.Valid() && !mm.IsUndefined() {
		// validateOneOf's allowed list is ['afterEvaluate', undefined], and the
		// message prints every entry, undefined included.
		if !mm.IsString() || mm.Str() != "afterEvaluate" {
			return js.Object{}, invalidValue("options.microtaskMode",
				[]string{"'afterEvaluate'", "undefined"}, mm)
		}
	}
	return o, nil
}

func (m *module) isContext(v js.Value) (js.Value, error) {
	if !v.Valid() || !v.IsObject() {
		return js.Value{}, invalidTypeArg("object", "object", v)
	}
	_, ok := m.contextID(v)
	return m.e.JS.Bool(ok), nil
}

func (m *module) runInContext(args []js.Value) (js.Value, error) {
	code, err := codeArg(arg(args, 0))
	if err != nil {
		return js.Value{}, err
	}
	ctx := arg(args, 1)
	opts, err := m.optionsArg(arg(args, 2), "options")
	if err != nil {
		return js.Value{}, err
	}
	if _, err := m.runOptions(opts); err != nil {
		return js.Value{}, err
	}
	// validateContext() reports a missing context as "must be an instance of
	// vm.Context", not the generic "must be of type" form.
	if !ctx.Valid() || !ctx.IsObject() {
		return js.Value{}, invalidInstanceArg("contextifiedObject", "vm.Context", ctx)
	}
	id, ok := m.contextID(ctx)
	if !ok {
		return js.Value{}, invalidInstanceArg("contextifiedObject", "vm.Context", ctx)
	}
	vc, _ := lookup(id)
	return m.runScript(vc, ctx.Object(), code, opts)
}

func (m *module) runInNewContext(args []js.Value) (js.Value, error) {
	code, err := codeArg(arg(args, 0))
	if err != nil {
		return js.Value{}, err
	}
	opts, err := m.optionsArg(arg(args, 2), "options")
	if err != nil {
		return js.Value{}, err
	}
	if _, err := m.runOptions(opts); err != nil {
		return js.Value{}, err
	}
	sandbox := arg(args, 1)
	ctxOpts, err := m.runContextOptions(opts)
	if err != nil {
		return js.Value{}, err
	}
	var ctxVal js.Value
	if !sandbox.Valid() || sandbox.IsUndefined() {
		ctxVal, err = m.createContext(nil)
	} else {
		ctxVal, err = m.createContext([]js.Value{sandbox, ctxOpts.V()})
	}
	if err != nil {
		return js.Value{}, err
	}
	id, _ := m.contextID(ctxVal)
	vc, _ := lookup(id)
	return m.runScript(vc, ctxVal.Object(), code, opts)
}

// runContextOptions maps the run* option names (contextName, contextOrigin,
// contextCodeGeneration) onto the createContext option names
// (lib/vm.js getContextOptions).
func (m *module) runContextOptions(opts js.Object) (js.Object, error) {
	out := m.e.JS.Obj()
	if v := opts.Get("contextName"); v.Valid() && !v.IsUndefined() {
		if !v.IsString() {
			return js.Object{}, invalidTypeArg("options.contextName", "string", v)
		}
		out.Set("name", v)
	}
	if v := opts.Get("contextOrigin"); v.Valid() && !v.IsUndefined() {
		if !v.IsString() {
			return js.Object{}, invalidTypeArg("options.contextOrigin", "string", v)
		}
		out.Set("origin", v)
	}
	if v := opts.Get("contextCodeGeneration"); v.Valid() && !v.IsUndefined() {
		if !v.IsObject() {
			return js.Object{}, invalidTypeArg("options.contextCodeGeneration", "object", v)
		}
		cg := v.Object()
		if _, err := boolOpt(cg, "options.contextCodeGeneration.strings", true); err != nil {
			return js.Object{}, err
		}
		if _, err := boolOpt(cg, "options.contextCodeGeneration.wasm", true); err != nil {
			return js.Object{}, err
		}
		out.Set("codeGeneration", v)
	}
	if v := opts.Get("microtaskMode"); v.Valid() && !v.IsUndefined() {
		if !v.IsString() || v.Str() != "afterEvaluate" {
			return js.Object{}, invalidValue("options.microtaskMode", []string{"'afterEvaluate'"}, v)
		}
		out.Set("microtaskMode", v)
	}
	return out, nil
}

func (m *module) runInThisContext(args []js.Value) (js.Value, error) {
	code, err := codeArg(arg(args, 0))
	if err != nil {
		return js.Value{}, err
	}
	opts, err := m.optionsArg(arg(args, 1), "options")
	if err != nil {
		return js.Value{}, err
	}
	if _, err := m.runOptions(opts); err != nil {
		return js.Value{}, err
	}
	return m.evalHere(code, opts)
}

// evalHere runs source in the caller's context, which is what
// runInThisContext and Script#runInThisContext do.
func (m *module) evalHere(code string, opts js.Object) (js.Value, error) {
	filename, err := stringOpt(opts, "options.filename", "evalmachine.<anonymous>")
	if err != nil {
		return js.Value{}, err
	}
	lineOffset, err := int32Opt(opts, "options.lineOffset")
	if err != nil {
		return js.Value{}, err
	}
	if lineOffset > 0 {
		code = strings.Repeat("\n", lineOffset) + code
	}
	return m.e.JS.Eval(code, filename)
}

// runOptions validates the options shared by every run* entry point
// (lib/vm.js getRunInContextArgs).
func (m *module) runOptions(opts js.Object) (int, error) {
	if _, err := uint32Opt(opts, "options.timeout", true); err != nil {
		return 0, err
	}
	if _, err := boolOpt(opts, "options.displayErrors", true); err != nil {
		return 0, err
	}
	if _, err := boolOpt(opts, "options.breakOnSigint", false); err != nil {
		return 0, err
	}
	return 0, nil
}

func (m *module) runScript(vc *vmContext, sandbox js.Object, code string, opts js.Object) (js.Value, error) {
	filename, err := stringOpt(opts, "options.filename", "evalmachine.<anonymous>")
	if err != nil {
		return js.Value{}, err
	}
	lineOffset, err := int32Opt(opts, "options.lineOffset")
	if err != nil {
		return js.Value{}, err
	}
	if lineOffset > 0 {
		code = strings.Repeat("\n", lineOffset) + code
	}

	g := vc.ctx.Global()
	m.copyInto(g, sandbox, vc.ctx)

	before := snapshot(g, vc.intrinsics)
	v, err := vc.ctx.Eval(code, filename)
	m.syncBack(sandbox, g, before, vc.intrinsics)
	if err != nil {
		return m.e.JS.Undefined(), normalizeRealmError(err)
	}
	return m.transfer(v, m.e.JS), nil
}

// createScript is `new Script(code, options)`; the constructor is a class, so
// it can only be invoked through Reflect.construct.
func (m *module) createScript(args []js.Value) (js.Value, error) {
	reflect := m.e.Global().Get("Reflect").Object().Get("construct")
	if !reflect.IsFunction() {
		return js.Value{}, js.ThrowCode("Error", "", "Reflect.construct is not available")
	}
	return reflect.Object().Call(js.Object{}, m.script.V(), m.e.JS.Arr(args).V())
}

func (m *module) compileFunction(args []js.Value) (js.Value, error) {
	// compileFunction validates the code as a string (Script and run* still
	// template-coerce it), so an omitted argument is a type error rather than
	// the string "undefined".
	codeV := arg(args, 0)
	if !codeV.Valid() || !codeV.IsString() {
		return js.Value{}, invalidTypeArg("code", "string", codeV)
	}
	code := codeV.Str()
	var params []string
	if pv := arg(args, 1); pv.Valid() && !pv.IsUndefined() {
		// validateStringArray starts with validateArray, whose message is the
		// "instance of Array" form.
		if !pv.IsArray() {
			return js.Value{}, invalidInstanceArg("params", "Array", pv)
		}
		arr := pv.Object()
		for i := 0; i < arr.Length(); i++ {
			item := arr.At(i)
			if !item.IsString() {
				return js.Value{}, invalidTypeArg("params["+strconv.Itoa(i)+"]", "string", item)
			}
			params = append(params, item.Str())
		}
	}

	opts, err := m.objectArg(arg(args, 2), "options")
	if err != nil {
		return js.Value{}, err
	}
	filename, err := stringOpt(opts, "options.filename", "")
	if err != nil {
		return js.Value{}, err
	}
	lineOffset, err := int32Opt(opts, "options.lineOffset")
	if err != nil {
		return js.Value{}, err
	}
	columnOffset, err := int32Opt(opts, "options.columnOffset")
	if err != nil {
		return js.Value{}, err
	}
	cachedData, err := optCachedData(opts, "options.cachedData")
	if err != nil {
		return js.Value{}, err
	}
	produceCached, err := boolOpt(opts, "options.produceCachedData", false)
	if err != nil {
		return js.Value{}, err
	}
	if pc := opts.Get("parsingContext"); pc.Valid() && !pc.IsUndefined() {
		_, ok := m.contextID(pc)
		if !pc.IsObject() || !ok {
			return js.Value{}, invalidInstanceArg("options.parsingContext", "Context", pc)
		}
	}

	var extensions []js.Value
	if ext := opts.Get("contextExtensions"); ext.Valid() && !ext.IsUndefined() {
		if !ext.IsArray() {
			return js.Value{}, invalidTypeArg("options.contextExtensions", "Array", ext)
		}
		arr := ext.Object()
		for i := 0; i < arr.Length(); i++ {
			item := arr.At(i)
			// Node allows null entries (kValidateObjectAllowNullable); they
			// simply contribute no bindings, so no `with` block is emitted.
			if item.Valid() && item.IsNull() {
				continue
			}
			if !item.Valid() || !item.IsObject() {
				return js.Value{}, invalidTypeArg(
					"options.contextExtensions["+strconv.Itoa(i)+"]", "object", item)
			}
			extensions = append(extensions, item)
		}
	}

	if lineOffset > 0 {
		code = strings.Repeat("\n", lineOffset) + code
	}
	if columnOffset > 0 {
		// V8 shifts the first line's reported columns by columnOffset; the
		// closest equivalent is indenting that line inside the function body.
		code = strings.Repeat(" ", columnOffset) + code
	}
	// contextExtensions behave like the scope chain V8 installs around the
	// compiled body, so a sloppy-mode `with` around each extension is the
	// closest construct JavaScript offers.
	var b strings.Builder
	b.WriteString("(function (__godeExtensions) {\nreturn (function (")
	b.WriteString(strings.Join(params, ", "))
	b.WriteString(") {\n")
	for i := range extensions {
		fmt.Fprintf(&b, "with (__godeExtensions[%d]) {\n", i)
	}
	b.WriteString(code)
	b.WriteString("\n")
	for range extensions {
		b.WriteString("}\n")
	}
	b.WriteString("});\n})")

	fn, err := m.e.JS.Eval(b.String(), filename)
	if err != nil {
		return js.Value{}, err
	}
	// The wrapper always has to be invoked: it returns the compiled function
	// itself, so skipping the call yielded the wrapper instead of the body.
	fn, err = fn.Object().Call(js.Object{}, m.e.JS.Arr(extensions).V())
	if err != nil {
		return js.Value{}, err
	}
	// V8 hands back a function carrying the code-cache properties; the tests
	// read cachedData/cachedDataProduced/cachedDataRejected off the result.
	if fn.Valid() && fn.IsFunction() {
		fo := fn.Object()
		if produceCached {
			fo.Set("cachedData", m.buffer(cacheBytes(codeV.Str())))
			fo.Set("cachedDataProduced", m.e.JS.Bool(true))
		}
		if cachedData != nil {
			fo.Set("cachedData", m.buffer(cachedData))
			fo.Set("cachedDataRejected", m.e.JS.Bool(!cacheValid(cachedData, codeV.Str())))
		}
	}
	return fn, nil
}

func (m *module) measureMemory(args []js.Value) (js.Value, error) {
	opts, err := m.objectArg(arg(args, 0), "options")
	if err != nil {
		return js.Value{}, err
	}
	if mode := opts.Get("mode"); mode.Valid() && !mode.IsUndefined() {
		if !mode.IsString() || (mode.Str() != "summary" && mode.Str() != "detailed") {
			return js.Value{}, invalidValue("options.mode",
				[]string{"'summary'", "'detailed'"}, mode)
		}
	}
	if exec := opts.Get("execution"); exec.Valid() && !exec.IsUndefined() {
		if !exec.IsString() || (exec.Str() != "default" && exec.Str() != "eager") {
			return js.Value{}, invalidValue("options.execution",
				[]string{"'default'", "'eager'"}, exec)
		}
	}
	stats := m.e.JS.Obj()
	total := m.e.JS.Obj()
	total.SetNum("jsMemoryEstimate", 0)
	total.Set("jsMemoryRange", m.e.JS.Arr([]js.Value{m.e.JS.Num(0), m.e.JS.Num(0)}).V())
	stats.Set("total", total.V())
	stats.Set("WebAssembly", m.e.JS.Obj().V())
	return m.e.JS.Promise(stats.V(), nil), nil
}

// scriptConstructor builds vm.Script. The class lives in JavaScript so
// instanceof, getters and prototype methods behave; every operation stays in
// Go.
func (m *module) scriptConstructor() js.Object {
	host := m.e.JS.Obj()
	host.SetFunc("init", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.scriptInit(args)
	})
	host.SetFunc("cachedData", func(_ js.Object, args []js.Value) (js.Value, error) {
		inst, err := scriptInstance(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		if data := m.scriptCache(inst); data != nil {
			return m.buffer(data), nil
		}
		return m.e.JS.Null(), nil
	})
	host.SetFunc("cachedDataProduced", func(_ js.Object, args []js.Value) (js.Value, error) {
		inst, err := scriptInstance(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(m.scriptCache(inst) != nil), nil
	})
	host.SetFunc("cachedDataRejected", func(_ js.Object, args []js.Value) (js.Value, error) {
		inst, err := scriptInstance(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		return m.e.JS.Bool(inst.Get("$godeScriptRejected").Bool()), nil
	})
	host.SetFunc("sourceMapURL", func(_ js.Object, args []js.Value) (js.Value, error) {
		inst, err := scriptInstance(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		src := inst.Get("$godeScriptSource").Str()
		if url := sourceMapOf(src); url != "" {
			return m.e.JS.Str(url), nil
		}
		// Node reports undefined (not null) when the script has no source map.
		return m.e.JS.Undefined(), nil
	})
	host.SetFunc("createCachedData", func(this js.Object, args []js.Value) (js.Value, error) {
		inst, err := scriptInstance(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		src := inst.Get("$godeScriptSource").Str()
		buf := cacheBytes(src)
		m.setScriptCache(inst, buf)
		return m.buffer(buf), nil
	})
	host.SetFunc("runThis", func(_ js.Object, args []js.Value) (js.Value, error) {
		inst, err := scriptInstance(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		opts, err := m.optionsArg(arg(args, 1), "options")
		if err != nil {
			return js.Value{}, err
		}
		if _, err := m.runOptions(opts); err != nil {
			return js.Value{}, err
		}
		code := inst.Get("$godeScriptSource").Str()
		filename := inst.Get("$godeScriptFilename").Str()
		lineOffset := inst.Get("$godeScriptLineOffset").Int()
		if lineOffset > 0 {
			code = strings.Repeat("\n", lineOffset) + code
		}
		return m.e.JS.Eval(code, filename)
	})
	host.SetFunc("runCtx", func(_ js.Object, args []js.Value) (js.Value, error) {
		inst, err := scriptInstance(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		ctx := arg(args, 1)
		opts, err := m.optionsArg(arg(args, 2), "options")
		if err != nil {
			return js.Value{}, err
		}
		if _, err := m.runOptions(opts); err != nil {
			return js.Value{}, err
		}
		if !ctx.Valid() || !ctx.IsObject() {
			return js.Value{}, invalidInstanceArg("contextifiedObject", "vm.Context", ctx)
		}
		id, ok := m.contextID(ctx)
		if !ok {
			return js.Value{}, invalidInstanceArg("contextifiedObject", "vm.Context", ctx)
		}
		vc, _ := lookup(id)
		code := inst.Get("$godeScriptSource").Str()
		if off := inst.Get("$godeScriptLineOffset").Int(); off > 0 {
			code = strings.Repeat("\n", off) + code
		}
		return m.runScriptRaw(vc, ctx.Object(), code, inst.Get("$godeScriptFilename").Str())
	})
	host.SetFunc("runNew", func(_ js.Object, args []js.Value) (js.Value, error) {
		inst, err := scriptInstance(args, 0)
		if err != nil {
			return js.Value{}, err
		}
		sandbox := arg(args, 1)
		opts, err := m.optionsArg(arg(args, 2), "options")
		if err != nil {
			return js.Value{}, err
		}
		if _, err := m.runOptions(opts); err != nil {
			return js.Value{}, err
		}
		ctxOpts, err := m.runContextOptions(opts)
		if err != nil {
			return js.Value{}, err
		}
		var ctxVal js.Value
		if !sandbox.Valid() || sandbox.IsUndefined() {
			ctxVal, err = m.createContext(nil)
		} else {
			ctxVal, err = m.createContext([]js.Value{sandbox, ctxOpts.V()})
		}
		if err != nil {
			return js.Value{}, err
		}
		id, _ := m.contextID(ctxVal)
		vc, _ := lookup(id)
		code := inst.Get("$godeScriptSource").Str()
		if off := inst.Get("$godeScriptLineOffset").Int(); off > 0 {
			code = strings.Repeat("\n", off) + code
		}
		return m.runScriptRaw(vc, ctxVal.Object(), code, inst.Get("$godeScriptFilename").Str())
	})

	// The class body is a Go raw string, so it cannot contain backticks:
	// coercion of `code` happens in codeArg instead of a template literal.
	maker, err := m.e.JS.Eval(`(function (host) {
		class Script {
			constructor(code, options) {
				host.init(this, code, options);
			}
			get cachedData() { return host.cachedData(this); }
			get cachedDataProduced() { return host.cachedDataProduced(this); }
			get cachedDataRejected() { return host.cachedDataRejected(this); }
			get sourceMapURL() { return host.sourceMapURL(this); }
			createCachedData() { return host.createCachedData(this); }
			runInThisContext(options) { return host.runThis(this, options); }
			runInContext(contextifiedObject, options) {
				return host.runCtx(this, contextifiedObject, options);
			}
			runInNewContext(contextObject, options) {
				return host.runNew(this, contextObject, options);
			}
		}
		return Script;
	})`, "[gode:vm]")
	if err != nil {
		panic("godejs vm: " + err.Error())
	}
	ctor, err := maker.Object().Call(js.Object{}, host.V())
	if err != nil {
		panic("godejs vm: " + err.Error())
	}
	return ctor.Object()
}

func (m *module) scriptInit(args []js.Value) (js.Value, error) {
	if len(args) < 2 || !args[0].IsObject() {
		return js.Value{}, js.ThrowCode("TypeError", "", "Script receiver missing")
	}
	instance := args[0].Object()
	code, err := codeArg(args[1])
	if err != nil {
		return js.Value{}, err
	}
	var opts js.Object
	if len(args) > 2 {
		opts, err = m.optionsArg(args[2], "options")
	} else {
		opts, err = m.e.JS.Obj(), nil
	}
	if err != nil {
		return js.Value{}, err
	}
	filename, err := stringOpt(opts, "options.filename", "evalmachine.<anonymous>")
	if err != nil {
		return js.Value{}, err
	}
	lineOffset, err := int32Opt(opts, "options.lineOffset")
	if err != nil {
		return js.Value{}, err
	}
	if _, err := int32Opt(opts, "options.columnOffset"); err != nil {
		return js.Value{}, err
	}
	cached, err := optCachedData(opts, "options.cachedData")
	if err != nil {
		return js.Value{}, err
	}
	produce, err := boolOpt(opts, "options.produceCachedData", false)
	if err != nil {
		return js.Value{}, err
	}

	instance.SetHidden("$godeScriptSource", m.e.JS.Str(code))
	instance.SetHidden("$godeScriptFilename", m.e.JS.Str(filename))
	instance.SetHidden("$godeScriptLineOffset", m.e.JS.Num(float64(lineOffset)))
	instance.SetHidden("$godeScriptRejected", m.e.JS.Bool(cached != nil && !cacheValid(cached, code)))

	m.cacheMu.Lock()
	m.cacheSeq++
	id := m.cacheSeq
	m.cacheMu.Unlock()
	instance.SetHidden("$godeScriptId", m.e.JS.Num(float64(id)))
	if produce {
		m.setScriptCache(instance, cacheBytes(code))
	}
	return m.e.JS.Undefined(), nil
}

func (m *module) scriptCache(inst js.Object) []byte {
	id := inst.Get("$godeScriptId")
	if !id.Valid() || !id.IsNumber() {
		return nil
	}
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()
	return m.caches[id.Int()]
}

func (m *module) setScriptCache(inst js.Object, data []byte) {
	id := inst.Get("$godeScriptId")
	if !id.Valid() || !id.IsNumber() {
		return
	}
	m.cacheMu.Lock()
	m.caches[id.Int()] = data
	m.cacheMu.Unlock()
}

// runScriptRaw is runScript without option plumbing, used by Script methods
// whose filename/lineOffset were fixed at construction time.
func (m *module) runScriptRaw(vc *vmContext, sandbox js.Object, code, filename string) (js.Value, error) {
	g := vc.ctx.Global()
	m.copyInto(g, sandbox, vc.ctx)
	before := snapshot(g, vc.intrinsics)
	v, err := vc.ctx.Eval(code, filename)
	m.syncBack(sandbox, g, before, vc.intrinsics)
	if err != nil {
		return m.e.JS.Undefined(), normalizeRealmError(err)
	}
	return m.transfer(v, m.e.JS), nil
}

func scriptInstance(args []js.Value, i int) (js.Object, error) {
	if i >= len(args) || !args[i].IsObject() {
		return js.Object{}, js.ThrowCode("TypeError", "", "Script receiver missing")
	}
	o := args[i].Object()
	if !o.Get("$godeScriptSource").IsString() {
		return js.Object{}, js.ThrowCode("TypeError", "", "Script receiver missing")
	}
	return o, nil
}

// --- value helpers -------------------------------------------------------

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

// codeArg coerces the code argument the way Node's `${code}` does, which
// throws for symbols rather than stringifying them.
func codeArg(v js.Value) (string, error) {
	if !v.Valid() {
		return "undefined", nil
	}
	s, ok := v.StrOK()
	if !ok {
		return "", js.ThrowCode("TypeError", "", "Cannot convert a Symbol value to a string")
	}
	return s, nil
}

func (m *module) optionsArg(v js.Value, name string) (js.Object, error) {
	if !v.Valid() || v.IsUndefined() {
		return m.e.JS.Obj(), nil
	}
	if v.IsString() {
		o := m.e.JS.Obj()
		o.SetStr("filename", v.Str())
		return o, nil
	}
	return m.objectArg(v, name)
}

// objectArg is validateObject: options that do not accept the string
// shorform (compileFunction, measureMemory) go through this.
func (m *module) objectArg(v js.Value, name string) (js.Object, error) {
	if !v.Valid() || v.IsUndefined() {
		return m.e.JS.Obj(), nil
	}
	if !v.IsObject() {
		return js.Object{}, invalidTypeArg(name, "object", v)
	}
	return v.Object(), nil
}

func keyOf(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i+1:]
	}
	return name
}

func stringOpt(o js.Object, name, def string) (string, error) {
	v := o.Get(keyOf(name))
	if !v.Valid() || v.IsUndefined() {
		return def, nil
	}
	if !v.IsString() {
		return "", invalidTypeArg(name, "string", v)
	}
	return v.Str(), nil
}

func boolOpt(o js.Object, name string, def bool) (bool, error) {
	v := o.Get(keyOf(name))
	if !v.Valid() || v.IsUndefined() {
		return def, nil
	}
	if !v.IsBool() {
		return false, invalidTypeArg(name, "boolean", v)
	}
	return v.Bool(), nil
}

// int32Opt mirrors validateInt32 (lib/internal/validators.js): a non-number
// is a type error, a non-integer or out-of-range number an out-of-range error.
func int32Opt(o js.Object, name string) (int, error) {
	v := o.Get(keyOf(name))
	if !v.Valid() || v.IsUndefined() {
		return 0, nil
	}
	if !v.IsNumber() {
		return 0, invalidTypeArg(name, "number", v)
	}
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, outOfRange(name, "an integer", v)
	}
	if f < math.MinInt32 || f > math.MaxInt32 {
		return 0, outOfRange(name, ">= -2147483648 && <= 2147483647", v)
	}
	return int(f), nil
}

// uint32Opt mirrors validateUint32; positive demands a minimum of 1, which is
// what run* options do for `timeout`.
func uint32Opt(o js.Object, name string, positive bool) (int, error) {
	v := o.Get(keyOf(name))
	if !v.Valid() || v.IsUndefined() {
		return 0, nil
	}
	if !v.IsNumber() {
		return 0, invalidTypeArg(name, "number", v)
	}
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, outOfRange(name, "an integer", v)
	}
	min := 0.0
	if positive {
		min = 1
	}
	if f < min || f > 4294967295 {
		want := ">= 0 && <= 4294967295"
		if positive {
			want = ">= 1 && <= 4294967295"
		}
		return 0, outOfRange(name, want, v)
	}
	return int(f), nil
}

func outOfRange(name, want string, v js.Value) error {
	return js.ThrowCode("RangeError", "ERR_OUT_OF_RANGE",
		`The value of "`+name+`" is out of range. It must be `+want+`. Received `+numString(v))
}

func optCachedData(o js.Object, name string) ([]byte, error) {
	v := o.Get(keyOf(name))
	if !v.Valid() || v.IsUndefined() {
		return nil, nil
	}
	b, ok := v.Bytes()
	if !ok {
		return nil, invalidInstanceArg(name, "Buffer, TypedArray, or DataView", v)
	}
	return b, nil
}

func invalidTypeArg(name, expected string, v js.Value) error {
	// Node's validators word properties of an options object as "property",
	// and arrays as "an instance of Array".
	kind := "argument"
	if strings.HasPrefix(name, "options.") {
		kind = "property"
	}
	if expected == "Array" {
		return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "`+name+`" `+kind+` must be an instance of Array. Received `+describe(v))
	}
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		`The "`+name+`" `+kind+` must be of type `+expected+`. Received `+describe(v))
}

func invalidInstanceArg(name, expected string, v js.Value) error {
	kind := "argument"
	if strings.HasPrefix(name, "options.") {
		kind = "property"
	}
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		`The "`+name+`" `+kind+` must be an instance of `+expected+`. Received `+describe(v))
}

func invalidValue(name string, allowed []string, v js.Value) error {
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
		"The argument '"+name+"' must be one of: "+strings.Join(allowed, ", ")+
			". Received "+inspect(v))
}

// inspect renders a value the way ERR_INVALID_ARG_VALUE's message expects.
func inspect(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return "undefined"
	}
	if v.IsNull() {
		return "null"
	}
	if v.IsString() {
		return "'" + v.Str() + "'"
	}
	if v.IsNumber() {
		return numString(v)
	}
	if v.IsBool() {
		if v.Bool() {
			return "true"
		}
		return "false"
	}
	if s, err := v.JSON(0); err == nil && s != "" {
		return s
	}
	return describe(v)
}

func describe(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return "undefined"
	}
	if v.IsNull() {
		return "null"
	}
	if v.IsBool() {
		if v.Bool() {
			return "type boolean (true)"
		}
		return "type boolean (false)"
	}
	if v.IsNumber() {
		return "type number (" + numString(v) + ")"
	}
	if v.IsString() {
		return "type string ('" + v.Str() + "')"
	}
	if v.IsFunction() {
		if name := v.Object().Get("name").Str(); name != "" {
			return "function " + name
		}
		return "function"
	}
	if v.IsObject() {
		name := "Object"
		if c := v.Object().Get("constructor"); c.IsFunction() {
			if n := c.Object().Get("name").Str(); n != "" {
				name = n
			}
		}
		return "an instance of " + name
	}
	return "an instance of Object"
}

func numString(v js.Value) string {
	f := v.Float()
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// --- cached data ---------------------------------------------------------

var cacheMagic = []byte("GODEVM1\n")

// cacheBytes is the serialized "code cache" a Script produces. V8 stores
// bytecode here; GodeJS recompiles, so the cache only proves the source it was
// produced from, which is exactly what cachedDataRejected reports.
func cacheBytes(code string) []byte {
	sum := sha256.Sum256([]byte(code))
	out := make([]byte, 0, len(cacheMagic)+8+16)
	out = append(out, cacheMagic...)
	var n [8]byte
	binary.BigEndian.PutUint64(n[:], uint64(len(code)))
	out = append(out, n[:]...)
	out = append(out, sum[:16]...)
	return out
}

func cacheValid(data []byte, code string) bool {
	want := cacheBytes(code)
	if len(data) != len(want) {
		return false
	}
	for i := range want {
		if data[i] != want[i] {
			return false
		}
	}
	return true
}

func (m *module) buffer(b []byte) js.Value {
	ctor := m.e.Global().Get("Buffer")
	if ctor.IsFunction() {
		from := ctor.Object().Get("from")
		if from.IsFunction() {
			if v, err := from.Object().Call(ctor.Object(), m.e.JS.Bytes(b).V()); err == nil {
				return v
			}
		}
	}
	return m.e.JS.Bytes(b).V()
}

var sourceMapRe = regexp.MustCompile(
	`(?m)(?://[#@][ \t]*sourceMappingURL=([^\s]+)|/\*[#@][ \t]*sourceMappingURL=([^\s]+)[ \t]*\*/)`)

func sourceMapOf(code string) string {
	matches := sourceMapRe.FindAllStringSubmatch(code, -1)
	if len(matches) == 0 {
		return ""
	}
	last := matches[len(matches)-1]
	if len(last) > 1 && last[1] != "" {
		return last[1]
	}
	if len(last) > 2 {
		return last[2]
	}
	return ""
}

// --- realm mirroring -----------------------------------------------------

// copyInto copies a sandbox's own properties into a vm global. Own keys, not
// enumerable keys: Node contextifies non-enumerable sandbox properties too.
func (m *module) copyInto(dst js.Object, src js.Object, dstCtx *js.Context) {
	for _, k := range src.OwnKeys() {
		if strings.HasPrefix(k, "$gode") {
			continue
		}
		dst.Set(k, m.transfer(src.Get(k), dstCtx))
	}
}

// snapshot fingerprints the sandbox-visible globals so syncBack can tell what
// the run actually changed.
func snapshot(g js.Object, intrinsics map[string]bool) map[string]string {
	out := make(map[string]string, len(g.OwnKeys()))
	for _, k := range g.OwnKeys() {
		if strings.HasPrefix(k, "$gode") || intrinsics[k] {
			continue
		}
		out[k] = fingerprint(g.Get(k))
	}
	return out
}

// syncBack copies the global properties a run added or changed back to the
// sandbox. Comparing fingerprints keeps untouched values (and their identity
// in the caller's realm) intact.
func (m *module) syncBack(sandbox js.Object, g js.Object, before map[string]string, intrinsics map[string]bool) {
	for _, k := range g.OwnKeys() {
		if strings.HasPrefix(k, "$gode") || intrinsics[k] {
			continue
		}
		fp := fingerprint(g.Get(k))
		if old, ok := before[k]; ok && old == fp {
			continue
		}
		sandbox.Set(k, m.transfer(g.Get(k), m.e.JS))
	}
}

func fingerprint(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return "u"
	}
	if v.IsNull() {
		return "n"
	}
	if v.IsBool() {
		if v.Bool() {
			return "b1"
		}
		return "b0"
	}
	if v.IsNumber() {
		return "d" + strconv.FormatFloat(v.Float(), 'g', -1, 64)
	}
	if v.IsString() {
		return "s" + v.Str()
	}
	if s, err := v.JSON(0); err == nil {
		return "j" + s
	}
	return "x"
}

// normalizeRealmError rewrites the JavaScriptCore wording of a missing global
// binding to the wording V8 uses, so vm errors read like Node's.
func normalizeRealmError(err error) error {
	je, ok := err.(*js.Error)
	if !ok || je.Name != "ReferenceError" {
		return err
	}
	const prefix = "Can't find variable: "
	i := strings.Index(je.Message, prefix)
	if i < 0 {
		return err
	}
	name := strings.TrimSpace(je.Message[i+len(prefix):])
	if name == "" {
		return err
	}
	je.Message = name + " is not defined"
	return err
}

// transfer moves a value between realms. Primitives cross directly; functions
// become host callbacks that marshal their arguments back across the boundary;
// arrays, typed arrays and plain objects are copied structurally. Exotic
// objects (process, promises, class instances) still degrade to undefined.
func (m *module) transfer(v js.Value, dst *js.Context) js.Value {
	return m.transferValue(v, dst, 0)
}

func (m *module) transferValue(v js.Value, dst *js.Context, depth int) js.Value {
	if !v.Valid() || v.IsUndefined() {
		return dst.Undefined()
	}
	if v.Ctx() == dst {
		return v
	}
	if v.IsNull() {
		return dst.Null()
	}
	if v.IsBool() {
		return dst.Bool(v.Bool())
	}
	if v.IsNumber() {
		return dst.Num(v.Float())
	}
	if v.IsString() {
		return dst.Str(v.Str())
	}
	if v.IsError() {
		return transferError(v, dst)
	}
	if v.IsDate() {
		if ms, err := v.Object().Get("getTime").Object().Call(v.Object()); err == nil && ms.IsNumber() {
			if d, err := dst.Eval("new Date("+numString(ms)+")", "[gode:vm]"); err == nil {
				return d
			}
		}
	}
	if v.IsFunction() {
		return m.bridgeFunction(v, dst)
	}
	if tag := v.Tag(); tag == "ArrayBuffer" || tag == "SharedArrayBuffer" {
		if n := v.Object().Get("byteLength"); n.IsNumber() && n.Float() == 0 {
			if ab, err := dst.Eval("new ArrayBuffer(0)", "[gode:vm]"); err == nil {
				return ab
			}
		}
	}
	if depth >= 4 {
		return dst.Undefined()
	}
	if v.IsArray() {
		o := v.Object()
		n := o.Length()
		if n < 0 || n > 100000 {
			return dst.Undefined()
		}
		vals := make([]js.Value, n)
		for i := 0; i < n; i++ {
			vals[i] = m.transferValue(o.At(i), dst, depth+1)
		}
		return dst.Arr(vals).V()
	}
	if name, n, ok := typedArrayInfo(v); ok {
		if n >= 0 && n <= 65536 {
			parts := make([]string, n)
			o := v.Object()
			for i := 0; i < n; i++ {
				parts[i] = strconv.FormatFloat(o.At(i).Float(), 'g', -1, 64)
			}
			if out, err := dst.Eval("new "+name+"(["+strings.Join(parts, ",")+"])", "[gode:vm]"); err == nil {
				return out
			}
		}
	}
	if isPlainObject(v) {
		o := v.Object()
		out := dst.Obj()
		for _, k := range o.OwnEnumerableKeys() {
			out.Set(k, m.transferValue(o.Get(k), dst, depth+1))
		}
		return out.V()
	}
	s, err := v.JSON(0)
	if err != nil || s == "" {
		return dst.Undefined()
	}
	parsed, err := dst.ParseJSON(s)
	if err != nil {
		return dst.Undefined()
	}
	return parsed
}

// bridgeFunction exposes a function from another realm to dst as a host
// function, marshalling the arguments and the return value across the boundary.
func (m *module) bridgeFunction(fn js.Value, dst *js.Context) js.Value {
	holder := dst.Obj()
	holder.SetFunc("bridged", func(_ js.Object, args []js.Value) (js.Value, error) {
		src := fn.Ctx()
		callArgs := make([]js.Value, len(args))
		for i, a := range args {
			callArgs[i] = m.transferValue(a, src, 1)
		}
		res, err := fn.Object().Call(js.Object{}, callArgs...)
		if err != nil {
			return js.Value{}, err
		}
		return m.transferValue(res, dst, 1), nil
	})
	return holder.Get("bridged")
}

// isPlainObject reports whether v is an object literal (or null-prototype
// object) rather than a class instance with host state.
func isPlainObject(v js.Value) bool {
	if !v.Valid() || !v.IsObject() {
		return false
	}
	p := v.Object().Proto()
	if !p.V().Valid() {
		return true
	}
	objCtor := v.Ctx().Global().Get("Object")
	if !objCtor.Valid() || !objCtor.IsObject() {
		return false
	}
	proto := objCtor.Object().Get("prototype")
	return proto.Valid() && p.Same(proto)
}

// typedArrayInfo reports a typed array's constructor name and length.
func typedArrayInfo(v js.Value) (string, int, bool) {
	if !v.Valid() || !v.IsObject() {
		return "", 0, false
	}
	ab := v.Ctx().Global().Get("ArrayBuffer")
	if !ab.Valid() || !ab.IsObject() {
		return "", 0, false
	}
	isView := ab.Object().Get("isView")
	if !isView.IsFunction() {
		return "", 0, false
	}
	res, err := isView.Object().Call(ab.Object(), v)
	if err != nil || !res.Bool() {
		return "", 0, false
	}
	name := v.Object().Get("constructor").Object().Get("name").Str()
	switch name {
	case "Uint8Array", "Uint8ClampedArray", "Int8Array", "Uint16Array", "Int16Array",
		"Uint32Array", "Int32Array", "Float32Array", "Float64Array":
	default:
		return "", 0, false
	}
	return name, v.Object().Get("length").Int(), true
}

func transferError(v js.Value, dst *js.Context) js.Value {
	o := v.Object()
	name := o.Get("name").Str()
	if name == "" {
		name = "Error"
	}
	err := &js.Error{Name: name, Message: o.Get("message").Str()}
	if code := o.Get("code"); code.Valid() && code.IsString() {
		err.Code = code.Str()
	}
	out := dst.ErrorValue(err)
	if out.Valid() && out.IsObject() {
		if stack := o.Get("stack").Str(); stack != "" {
			out.Object().SetStr("stack", stack)
		}
	}
	return out
}

// Implemented lists the tracked Node 24 API entries this package provides.
// Module/SourceTextModule/SyntheticModule are also absent from Node unless the
// --experimental-vm-modules flag is on, which GodeJS does not accept.
func Implemented() []string {
	return []string{
		"vm.Script",
		"vm.Script#createCachedData",
		"vm.Script#runInContext",
		"vm.Script#runInNewContext",
		"vm.Script#runInThisContext",
		"vm.Script.cachedDataRejected",
		"vm.Script.sourceMapURL",
		"vm.DONT_CONTEXTIFY",
		"vm.compileFunction",
		"vm.constants",
		"vm.createContext",
		"vm.isContext",
		"vm.measureMemory",
		"vm.runInContext",
		"vm.runInNewContext",
		"vm.runInThisContext",
		"vm.Module",
		"vm.Module#evaluate",
		"vm.Module#link",
		"vm.Module.error",
		"vm.Module.identifier",
		"vm.Module.namespace",
		"vm.Module.status",
		"vm.SourceTextModule",
		"vm.SourceTextModule#createCachedData",
		"vm.SourceTextModule#hasAsyncGraph",
		"vm.SourceTextModule#hasTopLevelAwait",
		"vm.SourceTextModule#instantiate",
		"vm.SourceTextModule#linkRequests",
		"vm.SourceTextModule.dependencySpecifiers",
		"vm.SourceTextModule.moduleRequests",
		"vm.SyntheticModule",
		"vm.SyntheticModule#setExport",
	}
}
