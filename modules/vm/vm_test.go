package vm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func newVM(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "vm", Install)
	return e
}

func TestRunInThisContext(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `String(vm.runInThisContext("1 + 1"))`), "2")
	jstest.Equal(t, jstest.Str(t, e, `String(vm.runInThisContext("'a' + 'b'"))`), "ab")
	// Missing code is coerced exactly like Node's `${code}`.
	jstest.Equal(t, jstest.Str(t, e, `String(vm.runInThisContext())`), "undefined")
	jstest.Equal(t, jstest.Str(t, e, `String(vm.runInThisContext(42))`), "42")
	jstest.Equal(t, jstest.Str(t, e, `String(vm.runInThisContext({ toString() { return '7' } }))`), "7")
}

func TestRunInThisContextDoesNotSeeLocals(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const hidden = 1;
		return vm.runInThisContext('typeof hidden');
	})()`), "undefined")
}

func TestRunInNewContext(t *testing.T) {
	e := newVM(t)

	src := `
		const sandbox = { x: 41 };
		const result = vm.runInNewContext("x + 1", sandbox);
		const kept = sandbox.x;
		vm.runInNewContext("x = 7", sandbox);
		JSON.stringify([result, kept, sandbox.x])
	`
	jstest.Equal(t, jstest.Str(t, e, src), "[42,41,7]")
	jstest.Equal(t, jstest.Str(t, e, `String(vm.runInNewContext("1 + 2"))`), "3")
	// The vm realm has none of the host's globals.
	jstest.Equal(t, jstest.Str(t, e,
		`vm.runInNewContext("[typeof process, typeof require, typeof globalThis].join()")`),
		"undefined,undefined,object")
}

func TestCreateContextAndRunInContext(t *testing.T) {
	e := newVM(t)

	src := `
		const ctx = vm.createContext({ a: 2 });
		const isCtx = vm.isContext(ctx);
		const notCtx = vm.isContext({});
		const tripled = vm.runInContext("a * 3", ctx);
		vm.runInContext("b = a + 1", ctx);
		JSON.stringify([isCtx, notCtx, tripled, ctx.b, ctx.a])
	`
	jstest.Equal(t, jstest.Str(t, e, src), "[true,false,6,3,2]")
}

func TestCreateContextIsIdempotent(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const sandbox = { a: 1 };
		const first = vm.createContext(sandbox);
		const second = vm.createContext(first);
		return JSON.stringify([first === sandbox, second === first, vm.isContext(second)]);
	})()`), "[true,true,true]")
}

func TestDontContextify(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const sandbox = { x: 1 };
		const ctx = vm.createContext(vm.constants.DONT_CONTEXTIFY);
		return JSON.stringify([
			ctx !== sandbox,
			vm.isContext(ctx),
			vm.runInContext('typeof globalThis', ctx),
			vm.runInContext('1 + 1', ctx)
		]);
	})()`), `[true,true,"object",2]`)
}

func TestConstantsShape(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const c = vm.constants;
		return JSON.stringify([
			typeof c.DONT_CONTEXTIFY,
			typeof c.USE_MAIN_CONTEXT_DEFAULT_LOADER,
			Object.isFrozen(c),
			Object.getPrototypeOf(c) === null
		]);
	})()`), `["symbol","symbol",true,true]`)
}

func TestScriptClass(t *testing.T) {
	e := newVM(t)

	src := `
		const ctx = vm.createContext({});
		const s = new vm.Script("1 + 1");
		const out = [
			s instanceof vm.Script,
			s.runInThisContext(),
			s.runInContext(ctx),
			s.runInNewContext({}),
			s.createCachedData().byteLength > 0
		];
		JSON.stringify(out)
	`
	jstest.Equal(t, jstest.Str(t, e, src), "[true,2,2,2,true]")
	// A script runs in the caller's realm, so it can see the caller's globals.
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		globalThis.__vmScriptProbe = 5;
		return new vm.Script('__vmScriptProbe + 1').runInThisContext();
	})()`), "6")
}

func TestScriptCachedData(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const source = '(function original() { return "original"; })';
		const script = new vm.Script(source, { produceCachedData: true });
		const data = script.cachedData;
		const same = new vm.Script(source, { cachedData: data });
		const other = new vm.Script('(function other() { return "other"; })', { cachedData: data });
		const sliced = new vm.Script(source, { cachedData: data.slice(4) });
		return JSON.stringify([
			ArrayBuffer.isView(data),
			data.byteLength > 0,
			script.cachedDataRejected,
			same.cachedDataRejected,
			same.runInThisContext()(),
			other.cachedDataRejected,
			sliced.cachedDataRejected
		]);
	})()`), `[true,true,false,false,"original",true,true]`)
}

func TestScriptSourceMapURL(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const withMap = new vm.Script('1\n//# sourceMappingURL=foo.js.map\n');
		const block = new vm.Script('1\n/*# sourceMappingURL=bar.js.map */\n');
		const none = new vm.Script('1');
		return JSON.stringify([withMap.sourceMapURL, block.sourceMapURL, none.sourceMapURL]);
	})()`), `["foo.js.map","bar.js.map",null]`)
}

func TestCreateScript(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const s = vm.createScript('1 + 1');
		return JSON.stringify([s instanceof vm.Script, s.runInThisContext()]);
	})()`), "[true,2]")
}

func TestCompileFunction(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e,
		`String(vm.compileFunction("return a + b", ["a", "b"])(2, 3))`), "5")
	jstest.Equal(t, jstest.Str(t, e,
		`typeof vm.compileFunction("return 1")`), "function")
	jstest.Equal(t, jstest.Str(t, e,
		`vm.compileFunction("return 1").name`), "")
	jstest.Equal(t, jstest.Str(t, e,
		`vm.compileFunction("return a + b").length`), "0")
	jstest.Equal(t, jstest.Str(t, e,
		`String(vm.compileFunction("return a + b", ["a", "b"]).length)`), "2")
}

func TestCompileFunctionContextExtensions(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `String(vm.compileFunction(
		'return a + b',
		[],
		{ contextExtensions: [{ a: 1 }, { b: 2 }] }
	)())`), "3")
	// Node's own test only exercises object entries; a null entry is accepted
	// by its validator but never reaches V8, so we test the real shape.
	jstest.Equal(t, jstest.Str(t, e, `String(vm.compileFunction(
		'return typeof a + ":" + (b === 5)',
		[],
		{ contextExtensions: [{ a: 1 }, { b: 5 }] }
	)())`), "number:true")
}

func TestMeasureMemoryAndConstants(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `typeof vm.measureMemory().then`), "function")
	jstest.Equal(t, jstest.Str(t, e, `typeof vm.constants`), "object")
}

func TestArgumentValidation(t *testing.T) {
	e := newVM(t)

	that := func(expr, want string) {
		t.Helper()
		jstest.Equal(t, jstest.Str(t, e, `(function () {
			try { `+expr+`; return 'no error'; }
			catch (err) { return err.code + '|' + err.name; }
		})()`), want)
	}

	that(`vm.createContext('string is not supported')`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.createContext(1)`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.isContext(null)`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.isContext(1)`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.runInContext('1', {})`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.runInContext('1')`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.runInNewContext('1', null)`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.runInThisContext('1', { timeout: -1 })`, "ERR_OUT_OF_RANGE|RangeError")
	that(`vm.runInThisContext('1', { timeout: 1.5 })`, "ERR_OUT_OF_RANGE|RangeError")
	that(`vm.runInThisContext('1', { displayErrors: 'yes' })`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.runInThisContext('1', { breakOnSigint: 1 })`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`new vm.Script('1', { filename: 5 })`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`new vm.Script('1', { lineOffset: 'x' })`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`new vm.Script('1', { columnOffset: 3000000000 })`, "ERR_OUT_OF_RANGE|RangeError")
	that(`new vm.Script('1', { cachedData: 'ohai' })`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`new vm.Script('1', { produceCachedData: 'yes' })`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.compileFunction('1', 'not-array')`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.compileFunction('1', [1])`, "ERR_INVALID_ARG_TYPE|TypeError")
	that(`vm.measureMemory({ mode: 'nope' })`, "ERR_INVALID_ARG_VALUE|TypeError")
	that(`vm.measureMemory({ execution: 'nope' })`, "ERR_INVALID_ARG_VALUE|TypeError")
	that(`vm.createContext({}, { microtaskMode: 'later' })`, "ERR_INVALID_ARG_VALUE|TypeError")
	that(`vm.createContext({}, { name: 5 })`, "ERR_INVALID_ARG_TYPE|TypeError")

	// createContext accepts arrays (kValidateObjectAllowArray).
	jstest.Equal(t, jstest.Str(t, e,
		`String(vm.isContext(vm.createContext([0, 1, 2, 3])))`), "true")
}

func TestRunOptionsAsString(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const err = (function () {
			try { vm.runInThisContext('throw new Error("boom")', 'named.js'); }
			catch (e) { return e.message; }
		})();
		return err;
	})()`), "boom")
}

func TestErrorsCrossContexts(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const err = vm.runInNewContext('new Error("boom")', {});
		const date = vm.runInNewContext('new Date(0)', {});
		return JSON.stringify([
			err instanceof Error,
			err.message,
			date instanceof Date,
			date.getTime()
		]);
	})()`), `[true,"boom",true,0]`)
}

func TestThrowingCodePropagates(t *testing.T) {
	e := newVM(t)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		try { vm.runInThisContext('throw new TypeError("nope")'); }
		catch (e) { return e.name + ':' + e.message; }
	})()`), "TypeError:nope")
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		try { vm.runInNewContext('throw new RangeError("radius")', {}); }
		catch (e) { return e.name + ':' + e.message; }
	})()`), "RangeError:radius")
}

func TestImplementedKeysExistInBaseline(t *testing.T) {
	// Every advertised key must be one the baseline tracks, and the module
	// currently claims the whole tracked surface.
	data, err := os.ReadFile(filepath.Join("..", "..", "parity", "baseline", "vm.json"))
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		t.Fatal(err)
	}
	tracked := map[string]bool{}
	for _, k := range keys {
		tracked[k] = true
	}
	for _, k := range Implemented() {
		if !tracked[k] {
			t.Fatalf("Implemented() lists untracked key %q", k)
		}
	}
	if len(Implemented()) != len(keys) {
		t.Fatalf("Implemented covers %d of %d baseline keys", len(Implemented()), len(keys))
	}
}
