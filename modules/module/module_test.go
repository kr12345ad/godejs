package module

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/internal/jstest"
)

// setup registers a few stub builtins so builtinModules and createRequire
// have something to resolve in this isolated test binary.
func setup(t *testing.T) *env.Env {
	t.Helper()
	stub := func(e *env.Env) js.Object { return e.JS.Obj() }
	for _, name := range []string{"fs", "path", "crypto"} {
		env.Register(name, stub)
	}
	e := jstest.New(t)
	jstest.Install(t, e, "module", Install)
	return e
}

func TestBuiltinModules(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
		["crypto","fs","module","path"].every(function (n) { return module.builtinModules.indexOf(n) >= 0; }) +
		":" + module.builtinModules.some(function (n) { return n.indexOf("node:") === 0; })`), "true:false")

	// Node keeps the list sorted.
	jstest.Equal(t, jstest.Str(t, e,
		`JSON.stringify(module.builtinModules) === JSON.stringify(module.builtinModules.slice().sort())`), "true")
}

func TestIsBuiltin(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
		module.isBuiltin("fs") + ":" +
		module.isBuiltin("node:fs") + ":" +
		module.isBuiltin("nope") + ":" +
		module.isBuiltin(1) + ":" +
		module.isBuiltin(undefined)`), "true:true:false:false:false")

	// The node: prefix is optional but the prefix itself is not a module.
	jstest.Equal(t, jstest.Str(t, e, `module.isBuiltin("node:")`), "false")
}

func TestCreateRequire(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `
		var req = module.createRequire("/tmp/x.js");
		typeof req + ":" + typeof req.resolve + ":" + typeof req.cache + ":" + (req.main === undefined)`),
		"function:function:object:true")

	// Loading a builtin goes through the bound require.
	jstest.Equal(t, jstest.Str(t, e,
		`Array.isArray(module.createRequire("/tmp/x.js")("node:module").builtinModules)`), "true")
	jstest.Equal(t, jstest.Str(t, e,
		`module.createRequire("/tmp/x.js").resolve("node:module")`), "module")

	// A file: URL is accepted too.
	jstest.Equal(t, jstest.Str(t, e,
		`typeof module.createRequire("file:///tmp/x.js")("node:path")`), "object")
}

func TestCreateRequireRejectsRelative(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e,
		`try { module.createRequire("x.js") } catch (err) { err.name + "|" + err.code }`),
		"TypeError|ERR_INVALID_ARG_VALUE")

	jstest.Equal(t, jstest.Str(t, e,
		`try { module.createRequire(42) } catch (err) { err.name }`), "TypeError")
}

func TestNodeModulePaths(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e,
		`JSON.stringify(module._nodeModulePaths("/usr/lib/node_modules/npm/foo"))`),
		`["/usr/lib/node_modules/npm/foo/node_modules","/usr/lib/node_modules/npm/node_modules","/usr/lib/node_modules","/usr/node_modules","/node_modules"]`)

	// An existing node_modules component is not doubled up.
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(module._nodeModulePaths("/node_modules"))`),
		`["/node_modules"]`)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(module._nodeModulePaths("/"))`),
		`["/node_modules"]`)

	// require.resolve.paths uses the same chain, and null for core modules.
	jstest.Equal(t, jstest.Str(t, e,
		`JSON.stringify(module.createRequire("/usr/lib/app.js").resolve.paths("x"))`),
		`["/usr/lib/node_modules","/usr/node_modules","/node_modules"]`)
	jstest.Equal(t, jstest.Str(t, e,
		`String(module.createRequire("/usr/lib/app.js").resolve.paths("fs"))`), "null")
}

func TestResolveFilename(t *testing.T) {
	e := setup(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dep.js"), []byte("module.exports = 1;"), 0o644); err != nil {
		t.Fatal(err)
	}
	// options.paths is a list of node_modules-style directories to search.
	pkgDir := filepath.Join(dir, "custom", "pkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "index.js"), []byte("module.exports = 2;"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.Global().SetStr("resolveDir", dir)

	jstest.Equal(t, jstest.Str(t, e,
		`module._resolveFilename("./dep.js", { filename: resolveDir + "/main.js" })`),
		filepath.Join(dir, "dep.js"))
	jstest.Equal(t, jstest.Str(t, e, `module._resolveFilename("node:module", {})`), "node:module")
	jstest.Equal(t, jstest.Str(t, e, `module._resolveFilename("fs", {})`), "fs")
	jstest.Equal(t, jstest.Str(t, e,
		`module._resolveFilename("pkg", {}, false, { paths: [resolveDir + "/custom"] })`),
		filepath.Join(pkgDir, "index.js"))

	jstest.Equal(t, jstest.Str(t, e,
		`JSON.stringify(module._findPath("pkg", [resolveDir + "/custom"]))`),
		`"`+filepath.Join(pkgDir, "index.js")+`"`)
	jstest.Equal(t, jstest.Str(t, e, `module._findPath("pkg", [resolveDir + "/nowhere"])`), "false")
}

func TestResolveFilenameErrors(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { module._resolveFilename("./missing.js", { filename: "/tmp/a.js" }); }
		catch (err) { return err.code + "|" + err.message.split("\n")[0]; }
	})()`), "MODULE_NOT_FOUND|Cannot find module './missing.js'")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try {
			module._resolveFilename("./missing.js",
				{ filename: "/tmp/a.js", parent: { filename: "/tmp/b.js", parent: null } });
		} catch (err) { return err.message.split("\n").join("|"); }
	})()`), "Cannot find module './missing.js'|Require stack:|- /tmp/a.js|- /tmp/b.js")
}

func TestModuleRecord(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const m = new module("/tmp/x.js");
		return [m.id, m.filename, m.loaded, typeof m.exports, Array.isArray(m.children), String(m.parent)].join(":");
	})()`), "/tmp/x.js:/tmp/x.js:false:object:true:null")

	jstest.Equal(t, jstest.Str(t, e,
		`typeof module.prototype.require + ":" + typeof module.prototype.resolve`), "function:function")
	jstest.Equal(t, jstest.Str(t, e, `Array.isArray(module.globalPaths)`), "true")
}

func TestWrap(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `module.wrap("1 + 1")`),
		"(function (exports, require, module, __filename, __dirname) { 1 + 1\n})")

	// The wrapper is real source: evaluating it yields a callable function.
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const fn = eval("(" + module.wrap("return 42;") + ")");
		return typeof fn + ":" + fn(null, null, null, null, null);
	})()`), "function:42")
	jstest.Equal(t, jstest.Str(t, e, `module.wrapSafe("2") === module.wrap("2")`), "true")
}

func TestCompileCacheAndSourceMaps(t *testing.T) {
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e, `[
		module.compileCacheStatus.ENABLED,
		module.compileCacheStatus.DISABLED,
		module.enableCompileCache().status,
		String(module.getCompileCacheDir()),
		String(module.flushCompileCache()),
		String(module.findSourceMap("/tmp/x.js")),
	].join(":")`), "0:3:3:undefined:undefined:undefined")

	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(module.getSourceMapsSupport())`),
		`{"enabled":false,"nodeModules":false,"generatedCode":false}`)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		module.setSourceMapsSupport(true, { nodeModules: true });
		return JSON.stringify(module.getSourceMapsSupport());
	})()`), `{"enabled":true,"nodeModules":true,"generatedCode":false}`)

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { module.setSourceMapsSupport("yes"); } catch (err) { return err.name + "|" + err.code; }
	})()`), "TypeError|ERR_INVALID_ARG_TYPE")

	jstest.Equal(t, jstest.Str(t, e, `(() => {
		try { module.enableCompileCache(5); } catch (err) { return err.code; }
	})()`), "ERR_INVALID_ARG_TYPE")
}

func TestFindPackageJSON(t *testing.T) {
	e := setup(t)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	pkg := filepath.Join(root, "package.json")
	if err := os.WriteFile(pkg, []byte(`{"name":"demo"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	e.Global().SetStr("pkgRoot", root)
	e.Global().SetStr("pkgPath", pkg)

	jstest.Equal(t, jstest.Str(t, e,
		`module.findPackageJSON(pkgRoot + "/sub/file.js")`), pkg)
	jstest.Equal(t, jstest.Str(t, e,
		`module.findPackageJSON("file://" + pkgRoot + "/sub/file.js")`), pkg)

	// This runtime has no package.json for builtins; do not invent one.
	jstest.Equal(t, jstest.Str(t, e, `module.findPackageJSON("node:fs") === undefined`), "true")
}

func TestSyncBuiltinESMExports(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e,
		`typeof module.syncBuiltinESMExports + ":" + module.syncBuiltinESMExports()`),
		"function:undefined")
}

func TestImplementedMatchesBaseline(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "parity", "baseline", "module.json"))
	if err != nil {
		t.Fatal(err)
	}
	var baseline []string
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	tracked := map[string]bool{}
	for _, key := range baseline {
		tracked[key] = true
	}

	impl := Implemented()
	seen := map[string]bool{}
	for _, key := range impl {
		if !tracked[key] {
			t.Fatalf("Implemented key %q is not in the baseline", key)
		}
		if seen[key] {
			t.Fatalf("Implemented lists %q twice", key)
		}
		seen[key] = true
	}
	if len(impl) != len(baseline) {
		t.Fatalf("Implemented covers %d of %d baseline keys", len(impl), len(baseline))
	}
}
