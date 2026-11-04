// Package module implements Node's `module` builtin: the CommonJS Module
// constructor, builtin discovery, createRequire, package.json lookup and the
// compile-cache/source-map support APIs.
package module

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("module", Install) }

// Install builds the module module's exports. Like Node, the export is the
// Module constructor itself, with the public helpers attached as statics.
func Install(e *env.Env) js.Object {
	// Module.prototype: the per-instance loading helpers. It is created first
	// because the constructor below installs records with it as prototype.
	proto := e.JS.Obj()
	holder := e.JS.Obj()
	// The Go implementation lives under a private name: Node's Module is a
	// class, so the public constructor has to be a real constructible
	// JavaScript function that delegates here.
	holder.SetFunc("impl", func(this js.Object, args []js.Value) (js.Value, error) {
		filename := ""
		if a := arg(args, 0); a.Valid() && a.IsString() {
			filename = a.Str()
		}
		rec := newModuleRecord(e, filename)
		if p := arg(args, 1); p.Valid() && p.IsObject() && !p.IsNull() {
			rec.Set("parent", p)
			if filename == "" {
				rec.Set("id", p.Object().Get("id"))
			}
		}
		rec.SetProto(proto)
		return rec.V(), nil
	})

	proto.SetFunc("require", func(this js.Object, args []js.Value) (js.Value, error) {
		dir := e.Cwd()
		if self := this.Get("filename"); self.IsString() && self.Str() != "" {
			dir = filepath.Dir(self.Str())
		}
		return e.RequireFrom(dir, arg(args, 0).Str())
	})
	proto.SetFunc("resolve", func(this js.Object, args []js.Value) (js.Value, error) {
		dir := e.Cwd()
		if self := this.Get("filename"); self.IsString() && self.Str() != "" {
			dir = filepath.Dir(self.Str())
		}
		resolved, err := e.RequireResolve(dir, arg(args, 0).Str())
		if err != nil {
			return js.Value{}, notFound(arg(args, 0).Str(), this)
		}
		return e.JS.Str(resolved), nil
	})
	proto.SetFunc("load", func(this js.Object, _ []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})
	m := constructible(e, holder.Get("impl"), proto, "Module")
	// Node's `require('module')` export is the Module constructor itself, and
	// it also exposes itself as `Module.Module`; promzard destructures that.
	m.Set("Module", m.V())
	setGlobalPaths := func() {
		fn, err := e.JS.Eval("(function (m) { var p = process.env.NODE_PATH; m.globalPaths = p ? p.split(':').filter(function (x) { return x !== ''; }) : []; })", "[gode:module]")
		if err == nil {
			_, _ = fn.Object().Call(js.Object{}, m.V())
		}
	}
	m.Set("globalPaths", e.JS.Arr(nil).V())
	setGlobalPaths()
	m.SetFunc("_initPaths", func(js.Object, []js.Value) (js.Value, error) {
		setGlobalPaths()
		return e.JS.Undefined(), nil
	})

	// builtinModules: every require-able builtin without the node: prefix.
	names := append([]string(nil), env.Builtins()...)
	sort.Strings(names)
	values := make([]js.Value, 0, len(names))
	for _, name := range names {
		// Node keeps the internal/* builtins require-able but does not list
		// them in builtinModules.
		if strings.HasPrefix(name, "internal/") {
			continue
		}
		values = append(values, e.JS.Str(name))
	}
	builtins := e.JS.Arr(values)
	m.Set("builtinModules", builtins.V())

	isBuiltin := func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsString() {
			return e.JS.Bool(false), nil
		}
		raw := args[0].Str()
		name := strings.TrimPrefix(raw, "node:")
		// internal/* builtins are require-able but are not public modules.
		if strings.HasPrefix(name, "internal/") {
			return e.JS.Bool(false), nil
		}
		// node:test is public; the bare `test` specifier resolves to the
		// runtime's own harness module and is not a builtin, as in Node.
		if name == "test" && !strings.HasPrefix(raw, "node:") {
			return e.JS.Bool(false), nil
		}
		return e.JS.Bool(env.Registered(name)), nil
	}
	m.SetFunc("isBuiltin", isBuiltin)

	// Module.wrapper is the source prefix/suffix Node compiles CommonJS files
	// with; tests read it to build their own loaders.
	m.Set("wrapper", e.JS.Arr([]js.Value{
		e.JS.Str("(function (exports, require, module, __filename, __dirname) { "),
		e.JS.Str("\n});"),
	}).V())

	// Module._stat reports 0 for files, 1 for directories and a negative value
	// when the path cannot be inspected, matching Node's internal helper.
	m.SetFunc("_stat", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsString() {
			return e.JS.Num(-1), nil
		}
		fi, err := os.Stat(args[0].Str())
		if err != nil {
			return e.JS.Num(-1), nil
		}
		if fi.IsDir() {
			return e.JS.Num(1), nil
		}
		return e.JS.Num(0), nil
	})

	cache := requireCache(e)
	m.Set("_cache", cache)
	m.Set("_extensions", extensions(e))
	m.SetFunc("runMain", func(_ js.Object, args []js.Value) (js.Value, error) {
		file := arg(args, 0).Str()
		if _, err := e.RunFile(file); err != nil {
			return js.Value{}, err
		}
		return e.JS.Undefined(), nil
	})

	m.SetFunc("_nodeModulePaths", func(_ js.Object, args []js.Value) (js.Value, error) {
		return nodeModulePaths(e, arg(args, 0).Str()), nil
	})

	// _resolveLookupPaths returns the directories Node searches for a request.
	// Relative requests resolve against the parent directory (Node represents
	// it as "."); bare specifiers start with the parent's node_modules chain.
	m.SetFunc("_resolveLookupPaths", func(_ js.Object, args []js.Value) (js.Value, error) {
		request := arg(args, 0).Str()
		if request == "." || request == ".." || strings.HasPrefix(request, "./") || strings.HasPrefix(request, "../") {
			return e.JS.Arr([]js.Value{e.JS.Str(".")}).V(), nil
		}
		if p := arg(args, 1); p.Valid() && p.IsObject() && !p.IsNull() {
			dir := filepath.Dir(p.Object().Get("filename").Str())
			values := []js.Value{e.JS.Str(dir)}
			if paths := nodeModulePaths(e, dir); paths.IsArray() {
				for i := 0; i < paths.Object().Length(); i++ {
					values = append(values, paths.Object().At(i))
				}
			}
			return e.JS.Arr(values).V(), nil
		}
		return e.JS.Arr([]js.Value{e.JS.Str("")}).V(), nil
	})

	m.SetFunc("_resolveFilename", func(_ js.Object, args []js.Value) (js.Value, error) {
		parent := js.Object{}
		if p := arg(args, 1); p.Valid() && p.IsObject() && !p.IsNull() {
			parent = p.Object()
		}
		resolved, err := resolveFilename(e, arg(args, 0).Str(), parent, arg(args, 3))
		if err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(resolved), nil
	})

	m.SetFunc("_findPath", func(_ js.Object, args []js.Value) (js.Value, error) {
		paths := arg(args, 1)
		if !paths.IsArray() {
			return e.JS.Bool(false), nil
		}
		request := arg(args, 0).Str()
		for i := 0; i < paths.Object().Length(); i++ {
			if resolved, ok := resolveFromLookupPath(e, paths.Object().At(i).Str(), request); ok {
				return e.JS.Str(resolved), nil
			}
		}
		return e.JS.Bool(false), nil
	})

	m.SetFunc("wrap", func(_ js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Str(wrapScript(arg(args, 0).Str())), nil
	})
	m.SetFunc("wrapSafe", func(_ js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Str(wrapScript(arg(args, 0).Str())), nil
	})

	createRequire := func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return js.Value{}, invalidArg("filename", js.Value{})
		}
		file, err := filenameArg(e, args[0])
		if err != nil {
			return js.Value{}, err
		}
		req := requireFrom(e, filepath.Dir(file))
		if resolve := req.Get("resolve"); resolve.IsFunction() {
			resolve.Object().SetFunc("paths", func(_ js.Object, args []js.Value) (js.Value, error) {
				spec := arg(args, 0).Str()
				if env.Registered(strings.TrimPrefix(spec, "node:")) {
					return e.JS.Null(), nil
				}
				return nodeModulePaths(e, filepath.Dir(file)), nil
			})
		}
		return req.V(), nil
	}
	m.SetFunc("createRequire", createRequire)
	m.SetFunc("createRequireFromPath", createRequire)

	m.SetFunc("findPackageJSON", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return e.JS.Undefined(), nil
		}
		target, err := packageStart(e, args)
		if err != nil || target == "" {
			return e.JS.Undefined(), nil
		}
		if found := nearestPackageJSON(target); found != "" {
			return e.JS.Str(found), nil
		}
		return e.JS.Undefined(), nil
	})

	// Our ESM transform resolves named imports through the CommonJS exports
	// object at load time, so builtin exports are already synchronized.
	m.SetFunc("syncBuiltinESMExports", func(_ js.Object, _ []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})

	installCompileCache(e, m)
	installSourceMaps(e, m)
	installExtended(e, m)
	return m
}

// installExtended adds the loader-hook and source-map surface Node exposes on
// the module builtin (module.register, registerHooks, SourceMap) plus the
// TypeScript stripping entry point.
func installExtended(e *env.Env, m js.Object) {
	v, err := e.JS.Eval(moduleExtendedPrelude, "[gode:module extended]")
	if err != nil || !v.Valid() || !v.IsFunction() {
		return
	}
	_, _ = v.Object().Call(js.Object{}, m.V())
}

// moduleExtendedPrelude defines module.SourceMap and the loader-hook helpers.
const moduleExtendedPrelude = `(function (m) {
  'use strict';

  var B64 = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';

  function decodeMappings(mappings) {
    var lines = String(mappings || '').split(';');
    var out = [];
    var sourceIndex = 0, originalLine = 0, originalColumn = 0, nameIndex = 0;
    for (var l = 0; l < lines.length; l++) {
      var generatedColumn = 0;
      var segments = lines[l].split(',');
      for (var s = 0; s < segments.length; s++) {
        var seg = segments[s];
        if (!seg) continue;
        var values = [];
        var value = 0, shift = 0;
        for (var i = 0; i < seg.length; i++) {
          var digit = B64.indexOf(seg[i]);
          if (digit < 0) continue;
          value += (digit & 31) << shift;
          if (digit & 32) { shift += 5; continue; }
          var negative = value & 1;
          value >>= 1;
          values.push(negative ? -value : value);
          value = 0; shift = 0;
        }
        if (values.length === 0) continue;
        generatedColumn += values[0];
        if (values.length >= 4) {
          sourceIndex += values[1];
          originalLine += values[2];
          originalColumn += values[3];
          if (values.length >= 5) nameIndex += values[4];
          out.push({
            generatedLine: l + 1, generatedColumn: generatedColumn,
            sourceIndex: sourceIndex, originalLine: originalLine + 1,
            originalColumn: originalColumn, nameIndex: nameIndex
          });
        } else {
          out.push({ generatedLine: l + 1, generatedColumn: generatedColumn });
        }
      }
    }
    return out;
  }

  function SourceMap(payload) {
    if (!(this instanceof SourceMap)) return new SourceMap(payload);
    var map = typeof payload === 'string' ? JSON.parse(payload) : (payload || {});
    this.payload = map;
    this.lineLengths = [];
    this._decoded = decodeMappings(map.mappings);
  }
  SourceMap.prototype.findEntry = function findEntry(lineNumber, columnNumber) {
    var best = null;
    for (var i = 0; i < this._decoded.length; i++) {
      var seg = this._decoded[i];
      if (seg.generatedLine !== lineNumber) continue;
      if (seg.generatedColumn > columnNumber) continue;
      if (seg.sourceIndex === undefined) continue;
      if (!best || seg.generatedColumn >= best.generatedColumn) best = seg;
    }
    if (!best) return {};
    var map = this.payload;
    return {
      generatedLine: best.generatedLine,
      generatedColumn: best.generatedColumn,
      originalSource: (map.sources || [])[best.sourceIndex],
      originalLine: best.originalLine,
      originalColumn: best.originalColumn,
      name: (map.names || [])[best.nameIndex]
    };
  };
  SourceMap.prototype.findOrigin = SourceMap.prototype.findEntry;
  SourceMap.prototype['return'] = function () { return this; };

  function register(specifier, parentURL, options) {
    if (typeof specifier !== 'string') {
      throw new TypeError('The "specifier" argument must be of type string');
    }
    var hooks = options && options.parentURL !== undefined ? options : undefined;
    globalThis.__godeModuleHooks = globalThis.__godeModuleHooks || [];
    globalThis.__godeModuleHooks.push({ specifier: specifier, parentURL: parentURL || null, options: hooks || null });
    return undefined;
  }

  function registerHooks(hooks) {
    if (hooks === null || typeof hooks !== 'object') {
      throw new TypeError('The "hooks" argument must be of type object');
    }
    globalThis.__godeModuleHooks = globalThis.__godeModuleHooks || [];
    globalThis.__godeModuleHooks.push(hooks);
    return {
      deregister: function deregister() {
        var idx = globalThis.__godeModuleHooks.indexOf(hooks);
        if (idx >= 0) globalThis.__godeModuleHooks.splice(idx, 1);
      }
    };
  }

  // stripTypeScriptTypes removes the TypeScript syntax Node's amaro-based
  // implementation handles for the common program forms.
  function stripTypeScriptTypes(code, options) {
    if (typeof code !== 'string') {
      throw new TypeError('The "code" argument must be of type string');
    }
    var src = code;
    src = src.replace(/^\s*(?:export\s+)?(?:declare\s+)?interface\s+[A-Za-z_$][\w$]*(?:\s+extends\s+[^{]+)?\s*\{[\s\S]*?\n\}/gm, '');
    src = src.replace(/^\s*(?:export\s+)?type\s+[A-Za-z_$][\w$]*(?:\s*<[^>]*>)?\s*=[^;]*;?/gm, '');
    src = src.replace(/^\s*declare\s+[^;\n]*;?/gm, '');
    src = src.replace(/^\s*(?:export\s+)?(?:const\s+)?enum\s+([A-Za-z_$][\w$]*)\s*\{[^}]*\}/gm,
      'var $1;');
    src = src.replace(/\s+as\s+(?:const|[A-Za-z_$][\w$.]*(?:<[^>]*>)?)/g, '');
    src = src.replace(/\s+satisfies\s+(?:[A-Za-z_$][\w$.]*(?:<[^>]*>)?)/g, '');
    src = src.replace(/\s*!\s*([.;,)])/g, '$1');
    src = src.replace(/([A-Za-z_$][\w$]*)\s*:\s*(?:string|number|boolean|any|unknown|void|never|object|symbol|bigint|([A-Za-z_$][\w$.]*)(?:<[^>]*>)?(?:\[\])?)\s*([,)=;])/g, '$1$3');
    src = src.replace(/^\s*(?:public|private|protected|readonly)\s+/gm, '');
    if (options && options.mode === 'transform' && !/\bexport\b|\bimport\b/.test(src)) {
      // plain transform mode leaves the module syntax alone
    }
    return { code: src, map: '' };
  }

  Object.defineProperty(m, 'SourceMap', { value: SourceMap, writable: true, configurable: true });
  Object.defineProperty(m, 'register', { value: register, writable: true, configurable: true });
  Object.defineProperty(m, 'registerHooks', { value: registerHooks, writable: true, configurable: true });
  Object.defineProperty(m, 'stripTypeScriptTypes', { value: stripTypeScriptTypes, writable: true, configurable: true });
  return m;
})`

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"module.builtinModules",
		"module.compileCacheStatus",
		"module.createRequire",
		"module.enableCompileCache",
		"module.findPackageJSON",
		"module.findSourceMap",
		"module.flushCompileCache",
		"module.getCompileCacheDir",
		"module.getSourceMapsSupport",
		"module.isBuiltin",
		"module.setSourceMapsSupport",
		"module.syncBuiltinESMExports",
		"module.SourceMap",
		"module.SourceMap#findEntry",
		"module.SourceMap#findOrigin",
		"module.SourceMap.return",
		"module.register",
		"module.registerHooks",
		"module.stripTypeScriptTypes",
	}
}

// installCompileCache exposes Node's compile-cache controls. This runtime has
// no on-disk bytecode cache, so enabling it always reports DISABLED.
func installCompileCache(e *env.Env, m js.Object) {
	status := e.JS.Obj()
	status.SetNum("ENABLED", 0)
	status.SetNum("ALREADY_ENABLED", 1)
	status.SetNum("FAILED", 2)
	status.SetNum("DISABLED", 3)
	m.Set("compileCacheStatus", status.V())

	constants := e.JS.Obj()
	constants.Set("compileCacheStatus", status.V())
	m.Set("constants", constants.V())

	m.SetFunc("enableCompileCache", func(_ js.Object, args []js.Value) (js.Value, error) {
		if d := arg(args, 0); d.Valid() && !d.IsUndefined() && !d.IsString() {
			return js.Value{}, invalidArg("cacheDir", d)
		}
		result := e.JS.Obj()
		result.SetNum("status", 3) // DISABLED
		result.SetStr("message", "Compile cache is not supported by this runtime")
		return result.V(), nil
	})
	m.SetFunc("flushCompileCache", func(_ js.Object, _ []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})
	m.SetFunc("getCompileCacheDir", func(_ js.Object, _ []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})
}

// installSourceMaps exposes Node's source-map support flag API. No source maps
// are produced, but the flags round-trip as they do in Node.
func installSourceMaps(e *env.Env, m js.Object) {
	var enabled, nodeModules, generatedCode bool

	state := func() js.Value {
		obj := e.JS.Obj()
		obj.SetBool("enabled", enabled)
		obj.SetBool("nodeModules", nodeModules)
		obj.SetBool("generatedCode", generatedCode)
		return obj.V()
	}
	m.SetFunc("getSourceMapsSupport", func(_ js.Object, _ []js.Value) (js.Value, error) {
		return state(), nil
	})
	m.SetFunc("setSourceMapsSupport", func(_ js.Object, args []js.Value) (js.Value, error) {
		flag := arg(args, 0)
		if !flag.IsBool() {
			return js.Value{}, invalidArg("enabled", flag)
		}
		enabled = flag.Bool()
		if opts := arg(args, 1); opts.Valid() && !opts.IsUndefined() {
			if !opts.IsObject() || opts.IsNull() {
				return js.Value{}, invalidArg("options", opts)
			}
			o := opts.Object()
			for _, key := range []string{"nodeModules", "generatedCode"} {
				if v := o.Get(key); v.Valid() && !v.IsUndefined() {
					if !v.IsBool() {
						return js.Value{}, invalidArg(key, v)
					}
					if key == "nodeModules" {
						nodeModules = v.Bool()
					} else {
						generatedCode = v.Bool()
					}
				}
			}
		}
		return e.JS.Undefined(), nil
	})
	m.SetFunc("findSourceMap", func(_ js.Object, _ []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})
}

// newModuleRecord builds the object shape `new Module()` produces.
func newModuleRecord(e *env.Env, filename string) js.Object {
	o := e.JS.Obj()
	dir := e.Cwd()
	id := "."
	if filename != "" {
		id = filename
		dir = filepath.Dir(filename)
	}
	o.SetStr("id", id)
	o.SetStr("path", dir)
	o.Set("exports", e.JS.Obj().V())
	o.SetBool("loaded", false)
	o.Set("children", e.JS.Arr(nil).V())
	o.Set("paths", nodeModulePaths(e, dir))
	if filename != "" {
		o.SetStr("filename", filename)
	} else {
		o.SetNull("filename")
	}
	o.SetNull("parent")
	return o
}

// requireCache returns the live require.cache object when there is one.
func requireCache(e *env.Env) js.Value {
	if req := e.Global().Get("require"); req.IsObject() {
		if cache := req.Object().Get("cache"); cache.IsObject() && !cache.IsNull() {
			return cache
		}
	}
	return e.JS.Obj().V()
}

// extensions mirrors Module._extensions. GodeJS does not consult this object
// while loading, but the entries load a file the way Node's do.
func extensions(e *env.Env) js.Value {
	obj := e.JS.Obj()
	for _, name := range []string{".js", ".json", ".node"} {
		ext := name
		obj.SetFunc(ext, func(this js.Object, args []js.Value) (js.Value, error) {
			filename := arg(args, 1).Str()
			if filename == "" {
				return e.JS.Undefined(), nil
			}
			var v js.Value
			if ext == ".node" {
				exp, err := e.LoadNative(filename)
				if err != nil {
					return js.Value{}, err
				}
				v = exp.V()
			} else {
				loaded, err := e.RequireFrom(filepath.Dir(filename), filename)
				if err != nil {
					return js.Value{}, err
				}
				v = loaded
			}
			if target := arg(args, 0); target.IsObject() && !target.IsNull() {
				target.Object().Set("exports", v)
			}
			return e.JS.Undefined(), nil
		})
	}
	return obj.V()
}

// wrapScript produces the wrapper Node compiles CommonJS source with.
func wrapScript(script string) string {
	return "(function (exports, require, module, __filename, __dirname) { " + script + "\n})"
}

// nodeModulePaths returns the node_modules lookup chain for a starting path.
func nodeModulePaths(e *env.Env, from string) js.Value {
	if from == "" {
		from = e.Cwd()
	}
	if !filepath.IsAbs(from) {
		if abs, err := filepath.Abs(from); err == nil {
			from = abs
		}
	}
	from = filepath.Clean(from)
	var values []js.Value
	for dir := from; ; {
		if filepath.Base(dir) != "node_modules" {
			values = append(values, e.JS.Str(filepath.Join(dir, "node_modules")))
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return e.JS.Arr(values).V()
}

// resolveFilename implements Module._resolveFilename for CommonJS requests.
func resolveFilename(e *env.Env, request string, parent js.Object, options js.Value) (string, error) {
	if strings.HasPrefix(request, "node:") {
		if env.Registered(strings.TrimPrefix(request, "node:")) {
			return request, nil
		}
		return "", notFound(request, parent)
	}
	if env.Registered(request) {
		return request, nil
	}
	if request == "" {
		return "", notFound(request, parent)
	}
	if options.Valid() && options.IsObject() && !options.IsNull() {
		if paths := options.Object().Get("paths"); paths.IsArray() {
			for i := 0; i < paths.Object().Length(); i++ {
				if resolved, ok := resolveFromLookupPath(e, paths.Object().At(i).Str(), request); ok {
					return resolved, nil
				}
			}
			return "", notFound(request, parent)
		}
	}
	base := e.Cwd()
	if parent.Valid() {
		if fn := parent.Get("filename"); fn.IsString() && fn.Str() != "" {
			base = filepath.Dir(fn.Str())
		}
	}
	resolved, err := e.RequireResolve(base, request)
	if err != nil {
		return "", notFound(request, parent)
	}
	return resolved, nil
}

// resolveFromLookupPath looks a request up in a node_modules-style directory.
func resolveFromLookupPath(e *env.Env, p, request string) (string, bool) {
	if p == "" {
		return "", false
	}
	if resolved, err := e.RequireResolve(p, request); err == nil {
		return resolved, true
	}
	parent := filepath.Dir(p)
	base := filepath.Base(p)
	if resolved, err := e.RequireResolve(parent, "./"+base+"/"+filepath.ToSlash(request)); err == nil {
		return resolved, true
	}
	return "", false
}

// notFound builds Node's MODULE_NOT_FOUND error, including the require stack.
func notFound(request string, parent js.Object) error {
	msg := "Cannot find module '" + request + "'"
	if stack := requireStack(parent); stack != "" {
		msg += "\nRequire stack:\n" + stack
	}
	return js.ThrowCode("Error", "MODULE_NOT_FOUND", msg)
}

func requireStack(parent js.Object) string {
	var lines []string
	seen := map[string]bool{}
	cur := parent
	for cur.Valid() && cur.Value.IsObject() && !cur.Value.IsNull() {
		fn := cur.Get("filename")
		if fn.IsString() && fn.Str() != "" {
			if seen[fn.Str()] {
				break
			}
			seen[fn.Str()] = true
			lines = append(lines, "- "+fn.Str())
		}
		next := cur.Get("parent")
		if !next.IsObject() || next.IsNull() {
			break
		}
		cur = next.Object()
	}
	return strings.Join(lines, "\n")
}

func invalidArg(name string, v js.Value) error {
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
		"The \""+name+"\" argument must be of type "+argTypeFor(name)+"."+describeReceived(v))
}

// argTypeFor names the expected type in validation messages.
func argTypeFor(name string) string {
	switch name {
	case "enabled":
		return "boolean"
	case "nodeModules", "generatedCode":
		return "boolean"
	case "filename":
		return "string or an instance of URL"
	}
	return "string"
}

// describeReceived renders Node's "Received ..." clause for a bad argument.
func describeReceived(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return " Received undefined"
	}
	if v.IsNull() {
		return " Received null"
	}
	if v.IsObject() && !v.IsFunction() {
		if name := v.Object().Get("constructor").Object().Get("name").Str(); name != "" {
			return " Received an instance of " + name
		}
		return " Received an instance of Object"
	}
	switch {
	case v.IsNumber():
		return fmt.Sprintf(" Received type number (%v)", v.Float())
	case v.IsBool():
		return fmt.Sprintf(" Received type boolean (%v)", v.Bool())
	case v.IsString():
		return " Received type string ('" + v.Str() + "')"
	case v.IsFunction():
		return " Received function " + v.Object().Get("name").Str()
	}
	return " Received type " + v.Tag()
}

// requireFrom builds the require function createRequire returns.
func requireFrom(e *env.Env, dir string) js.Object {
	holder := e.JS.Obj()
	holder.SetFunc("require", func(_ js.Object, args []js.Value) (js.Value, error) {
		return e.RequireFrom(dir, arg(args, 0).Str())
	})
	req := holder.Get("require").Object()

	req.SetFunc("resolve", func(_ js.Object, args []js.Value) (js.Value, error) {
		spec := arg(args, 0).Str()
		resolved, err := e.RequireResolve(dir, spec)
		if err != nil {
			return js.Value{}, js.ThrowCode("Error", "MODULE_NOT_FOUND", "Cannot find module '"+spec+"'")
		}
		return e.JS.Str(resolved), nil
	})
	req.Set("cache", requireCache(e))
	req.Set("main", e.JS.Undefined())
	return req
}

// filenameArg accepts the shapes Node allows for createRequire: an absolute
// path, a file: URL string, or a URL object.
func filenameArg(e *env.Env, v js.Value) (string, error) {
	if v.IsObject() && !v.IsNull() {
		obj := v.Object()
		if strings.EqualFold(obj.Get("protocol").Str(), "file:") {
			pathname, err := url.PathUnescape(obj.Get("pathname").Str())
			if err != nil {
				return "", invalidFilename(v)
			}
			return pathFromFileURL(pathname, obj.Get("host").Str())
		}
		return "", invalidFilename(v)
	}
	if !v.IsString() {
		return "", invalidFilename(v)
	}
	raw := v.Str()
	if strings.HasPrefix(raw, "file:") {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "file" {
			return "", invalidFilename(v)
		}
		path, err := url.PathUnescape(u.Path)
		if err != nil {
			return "", invalidFilename(v)
		}
		return pathFromFileURL(path, u.Host)
	}
	if !filepath.IsAbs(raw) {
		return "", invalidFilename(v)
	}
	return filepath.Clean(raw), nil
}

func pathFromFileURL(pathname, host string) (string, error) {
	if host != "" && host != "localhost" {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_FILE_URL_HOST",
			"File URL host must be \"localhost\" or empty on linux")
	}
	if pathname == "" {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_FILE_URL_PATH",
			"File URL path must be absolute")
	}
	return filepath.FromSlash(pathname), nil
}

func invalidFilename(v js.Value) error {
	// Node reports the value with util.inspect, so an empty object is {} and
	// strings keep their quotes.
	received := "'" + v.Str() + "'"
	if v.IsObject() {
		received = "{}"
		if v.IsArray() {
			received = "[]"
		}
	}
	return js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
		"The argument 'filename' must be a file URL object, file URL string, or absolute path string. Received "+
			received)
}

// packageStart resolves findPackageJSON's specifier to a filesystem start
// point. A builtin has no package.json in this runtime, so it yields "".
func packageStart(e *env.Env, args []js.Value) (string, error) {
	spec := args[0].Str()
	base := e.Cwd()
	if len(args) > 1 && args[1].IsString() && args[1].Str() != "" {
		base = args[1].Str()
		if strings.HasPrefix(base, "file:") {
			if u, err := url.Parse(base); err == nil {
				if p, err := url.PathUnescape(u.Path); err == nil {
					base = p
				}
			}
		}
	}
	if strings.HasPrefix(spec, "file:") {
		u, err := url.Parse(spec)
		if err != nil {
			return "", err
		}
		p, err := url.PathUnescape(u.Path)
		if err != nil {
			return "", err
		}
		return p, nil
	}
	if filepath.IsAbs(spec) {
		return filepath.Clean(spec), nil
	}
	if strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../") || spec == "." || spec == ".." {
		return filepath.Clean(filepath.Join(base, spec)), nil
	}
	if env.Registered(strings.TrimPrefix(spec, "node:")) {
		return "", nil
	}
	resolved, err := e.RequireResolve(base, spec)
	if err != nil {
		return "", nil
	}
	return resolved, nil
}

// nearestPackageJSON walks up from start to the filesystem root and returns
// the first existing package.json, or "".
func nearestPackageJSON(start string) string {
	dir := start
	if st, err := os.Stat(start); err == nil && !st.IsDir() {
		dir = filepath.Dir(start)
	} else if err != nil {
		dir = filepath.Dir(start)
	}
	for {
		candidate := filepath.Join(dir, "package.json")
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

// constructible turns a host implementation function into a real JavaScript
// function whose prototype is proto, so `new Module(id)` behaves like Node's
// class while every static still hangs off the same object. A Go-backed
// function object is not constructible in JavaScriptCore, which is why the
// public constructor is built from JavaScript and only delegates to Go.
func constructible(e *env.Env, impl js.Value, proto js.Object, name string) js.Object {
	makeCls, err := e.JS.Eval(`(function (impl, proto, name) {
  var Cls = function (...args) { return impl.apply(this, args); };
  Object.defineProperty(Cls, 'name', { value: name, configurable: true });
  Cls.prototype = proto;
  Object.defineProperty(proto, 'constructor', {
    value: Cls, writable: true, configurable: true, enumerable: false
  });
  return Cls;
})`, "[gode:module/ctor]")
	if err != nil {
		panic("godejs module: " + err.Error())
	}
	built, callErr := makeCls.Object().Call(js.Object{}, impl, proto.V(), e.JS.Str(name))
	if callErr != nil {
		panic("godejs module: " + callErr.Error())
	}
	return built.Object().Keep()
}
