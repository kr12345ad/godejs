package util

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/internal/jstest"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "util", Install)
	return e
}

func TestInspectPrimitives(t *testing.T) {
	e := setup(t)
	cases := []struct{ expr, want string }{
		{`util.inspect("hello")`, "'hello'"},
		{`util.inspect("it's")`, `'it\'s'`},
		{`util.inspect(42)`, "42"},
		{`util.inspect(-0)`, "-0"},
		{`util.inspect(NaN)`, "NaN"},
		{`util.inspect(Infinity)`, "Infinity"},
		{`util.inspect(null)`, "null"},
		{`util.inspect(undefined)`, "undefined"},
		{`util.inspect(true)`, "true"},
		{`util.inspect(false)`, "false"},
	}
	for _, c := range cases {
		if got := jstest.Str(t, e, c.expr); got != c.want {
			t.Errorf("%s = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestInspectContainers(t *testing.T) {
	e := setup(t)
	cases := []struct{ expr, want string }{
		{`util.inspect([])`, "[]"},
		{`util.inspect([1, 2, 3])`, "[ 1, 2, 3 ]"},
		{`util.inspect([[1, 2]])`, "[ [ 1, 2 ] ]"},
		{`util.inspect({a: 1, b: "x"})`, "{ a: 1, b: 'x' }"},
		{`util.inspect({"a-b": 1})`, "{ 'a-b': 1 }"},
		{`util.inspect({a: {b: 1}})`, "{ a: { b: 1 } }"},
		{`util.inspect({a: {b: {c: 1}}})`, "{ a: { b: { c: 1 } } }"},
		{`util.inspect({a: {b: {c: 1}}}, {depth: null})`, "{ a: { b: { c: 1 } } }"},
		{`util.inspect([1, , 3])`, "[ 1, <1 empty item>, 3 ]"},
		{`util.inspect([1, , , 4])`, "[ 1, <2 empty items>, 4 ]"},
		{`util.inspect(Object.create(null))`, "[Object: null prototype] {}"},
	}
	for _, c := range cases {
		if got := jstest.Str(t, e, c.expr); got != c.want {
			t.Errorf("%s = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestInspectBuiltins(t *testing.T) {
	e := setup(t)
	cases := []struct{ expr, want string }{
		{`util.inspect(new Date(0))`, "1970-01-01T00:00:00.000Z"},
		{`util.inspect(/abc/gi)`, "/abc/gi"},
		{`util.inspect(function foo() {})`, "[Function: foo]"},
		{`util.inspect(function () {})`, "[Function (anonymous)]"},
		{`util.inspect(new Uint8Array([1, 2]))`, "Uint8Array(2) [ 1, 2 ]"},
		{`util.inspect(new Map([["a", 1]]))`, "Map(1) { 'a' => 1 }"},
		{`util.inspect(new Set([1, 2]))`, "Set(2) { 1, 2 }"},
		{`util.inspect(new Map())`, "Map(0) {}"},
		{`util.inspect(new Set())`, "Set(0) {}"},
		{`util.inspect(new Error("boom"))`, "Error: boom"},
	}
	for _, c := range cases {
		if got := jstest.Str(t, e, c.expr); got != c.want {
			t.Errorf("%s = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestInspectBuffer(t *testing.T) {
	e := setup(t)
	e.Global().Set("__bytes", e.JS.Bytes([]byte{1, 2, 255}).V())
	expr := `(() => {
		const b = __bytes;
		Object.defineProperty(b, "_godeBuffer", { value: true });
		return util.inspect(b);
	})()`
	if got := jstest.Str(t, e, expr); got != "<Buffer 01 02 ff>" {
		t.Fatalf("buffer inspect = %q", got)
	}
}

func TestInspectCircular(t *testing.T) {
	e := setup(t)
	expr := `(() => {
		const a = {};
		a.self = a;
		return util.inspect(a);
	})()`
	if got := jstest.Str(t, e, expr); got != "<ref *1> { self: [Circular *1] }" {
		t.Fatalf("circular inspect = %q", got)
	}
}

func TestFormat(t *testing.T) {
	e := setup(t)
	cases := []struct{ expr, want string }{
		{`util.format("%s:%s", "a", "b")`, "a:b"},
		{`util.format("%d", "3.9")`, "3.9"},
		{`util.format("%i", "3.9")`, "3"},
		{`util.format("%f", "3.9")`, "3.9"},
		{`util.format("%j", {a: 1})`, `{"a":1}`},
		{`util.format("%%")`, "%"},
		{`util.format("a", 1)`, "a 1"},
		{`util.format({a: 1})`, "{ a: 1 }"},
		{`util.format("%s", {a: 1})`, "{ a: 1 }"},
		{`util.format("x %s")`, "x %s"},
		{`util.format("%c", "color:red")`, ""},
		{`util.format("%s", "a\nb")`, "a\nb"},
		{`util.formatWithOptions({}, "%s", 1)`, "1"},
	}
	for _, c := range cases {
		if got := jstest.Str(t, e, c.expr); got != c.want {
			t.Errorf("%s = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestDeepStrictEqual(t *testing.T) {
	e := setup(t)
	cases := []struct {
		expr string
		want bool
	}{
		{`util.isDeepStrictEqual({a: 1}, {a: 1})`, true},
		{`util.isDeepStrictEqual({a: 1}, {a: 2})`, false},
		{`util.isDeepStrictEqual([1, [2]], [1, [2]])`, true},
		{`util.isDeepStrictEqual(new Date(0), new Date(0))`, true},
		{`util.isDeepStrictEqual(NaN, NaN)`, true},
		{`util.isDeepStrictEqual(0, -0)`, false},
		{`util.isDeepStrictEqual(new Map([["a", {x: 1}]]), new Map([["a", {x: 1}]]))`, true},
		{`util.isDeepStrictEqual(new Set([1, 2]), new Set([2, 1]))`, true},
		{`util.isDeepStrictEqual(new Uint8Array([1, 2]), new Uint8Array([1, 2]))`, true},
		{`util.isDeepStrictEqual({}, Object.create(null))`, false},
		{`util.isDeepStrictEqual({a: 1, b: 2}, {b: 2, a: 1})`, true},
	}
	for _, c := range cases {
		want := "false"
		if c.want {
			want = "true"
		}
		if got := jstest.Str(t, e, c.expr); got != want {
			t.Errorf("%s = %q, want %q", c.expr, got, want)
		}
	}
}

func TestInherits(t *testing.T) {
	e := setup(t)
	expr := `(() => {
		function A() {}
		A.prototype.hi = function () { return "hi"; };
		function B() {}
		util.inherits(B, A);
		return [new B().hi(), B.super_ === A].join(",");
	})()`
	if got := jstest.Str(t, e, expr); got != "hi,true" {
		t.Fatalf("inherits = %q", got)
	}
}

func TestTypes(t *testing.T) {
	e := setup(t)
	expr := `JSON.stringify([
		util.types.isDate(new Date()),
		util.types.isRegExp(/x/),
		util.types.isMap(new Map()),
		util.types.isSet(new Set()),
		util.types.isPromise(Promise.resolve()),
		util.types.isTypedArray(new Uint8Array(1)),
		util.types.isUint8Array(new Uint8Array(1)),
		util.types.isNativeError(new Error("x")),
		util.types.isAsyncFunction(async function () {}),
		util.types.isGeneratorFunction(function* () {}),
		util.types.isArrayBuffer(new ArrayBuffer(1)),
		util.types.isMap({})
	])`
	want := "[true,true,true,true,true,true,true,true,true,true,true,false]"
	if got := jstest.Str(t, e, expr); got != want {
		t.Fatalf("types = %q", got)
	}
}

func TestStringHelpers(t *testing.T) {
	e := setup(t)
	if got := jstest.Str(t, e, `util.stripVTControlCharacters("\u001b[31mred\u001b[0m")`); got != "red" {
		t.Fatalf("stripVTControlCharacters = %q", got)
	}
	if got := jstest.Str(t, e, `String(util.toUSVString("a\ud800b") === "a\ufffdb")`); got != "true" {
		t.Fatalf("toUSVString = %q", got)
	}
}

func TestGetSystemErrorName(t *testing.T) {
	e := setup(t)
	if got := jstest.Str(t, e, `util.getSystemErrorName(-2)`); got != "ENOENT" {
		t.Fatalf("errno -2 = %q", got)
	}
	if got := jstest.Str(t, e, `util.getSystemErrorName(-13)`); got != "EACCES" {
		t.Fatalf("errno -13 = %q", got)
	}
}

func TestPromisifyAndCallbackify(t *testing.T) {
	e := setup(t)
	if got := jstest.Str(t, e, `typeof util.promisify(function (a, cb) { cb(null, a + 1); })(1).then`); got != "function" {
		t.Fatalf("promisify thenable = %q", got)
	}
	if got := jstest.Str(t, e, `typeof util.callbackify(function () { return Promise.resolve(1); })`); got != "function" {
		t.Fatalf("callbackify = %q", got)
	}
}

func TestDeprecateReturnsValue(t *testing.T) {
	e := setup(t)
	expr := `util.deprecate(function () { return 7; }, "gone")()`
	if got := jstest.Str(t, e, expr); got != "7" {
		t.Fatalf("deprecate = %q", got)
	}
}

func TestGoHelpers(t *testing.T) {
	e := setup(t)
	v := jstest.Eval(t, e, `({ a: 1, b: "x" })`)
	jstest.Equal(t, Inspect(v), "{ a: 1, b: 'x' }")

	got := Format([]js.Value{e.JS.Str("x"), e.JS.Num(1)})
	jstest.Equal(t, got, "x 1")
}

func TestParseEnv(t *testing.T) {
	e := setup(t)
	expr := `JSON.stringify(util.parseEnv("# comment\nA=1\nB='two words'\nC=\"line\\nbreak\"\nexport D=4\nA=5\n"))`
	jstest.Equal(t, jstest.Str(t, e, expr), `{"A":"5","B":"two words","C":"line\nbreak","D":"4"}`)

	// A single-quoted value is literal, double quotes expand escapes.
	jstest.Equal(t, jstest.Str(t, e, `util.parseEnv("X='a\\nb'").X`), `a\nb`)
	jstest.Equal(t, jstest.Str(t, e, `typeof util.parseEnv`), "function")
}

func TestParseEnvHandlesInvalidInput(t *testing.T) {
	e := setup(t)
	// Node's env-file parser skips lines it cannot understand instead of
	// rejecting the whole file (util.parseEnv('NOEQUALS') returns {}).
	jstest.Equal(t, jstest.Str(t, e,
		`JSON.stringify(util.parseEnv("NOEQUALS\nA=1\nbad key=2\n"))`),
		`{"A":"1"}`)
	jstest.Equal(t, jstest.Str(t, e,
		`(function () { try { util.parseEnv(1); } catch (e) { return e.code; } return 'none'; })()`),
		"ERR_INVALID_ARG_TYPE")
}

func TestLegacyTypePredicates(t *testing.T) {
	e := setup(t)
	expr := `[
		util.isArray([]), util.isBoolean(false), util.isDate(new Date(0)),
		util.isError(new Error('x')), util.isFunction(function () {}),
		util.isNull(null), util.isNullOrUndefined(undefined), util.isNumber(1),
		util.isObject({}), util.isPrimitive('s'), util.isRegExp(/x/),
		util.isString('s'), util.isSymbol(Symbol('s')), util.isUndefined(undefined),
		util.isArray({}), util.isNumber('1')
	].map(String).join(',')`
	jstest.Equal(t, jstest.Str(t, e, expr),
		"true,true,true,true,true,true,true,true,true,true,true,true,true,true,false,false")
}

func TestSystemErrorMessage(t *testing.T) {
	e := setup(t)
	// Node's message is the lowercased libc strerror text, and only negative
	// safe integers are accepted.
	jstest.Equal(t, jstest.Str(t, e, `util.getSystemErrorMessage(-2)`), "no such file or directory")
	jstest.Equal(t, jstest.Str(t, e, `util.getSystemErrorMessage(-13)`), "permission denied")

	err := jstest.Throws(t, e, `util.getSystemErrorMessage(13)`)
	jstest.Equal(t, err.Code, "ERR_OUT_OF_RANGE")
	jstest.Equal(t, err.Message,
		`The value of "err" is out of range. It must be a negative integer. Received 13`)

	err = jstest.Throws(t, e, `util.getSystemErrorMessage("x")`)
	jstest.Equal(t, err.Code, "ERR_INVALID_ARG_TYPE")

	jstest.Equal(t, jstest.Str(t, e, `util.getSystemErrorName(-2)`), "ENOENT")
}

func TestInspectFunctionFlavours(t *testing.T) {
	e := setup(t)
	cases := []struct{ expr, want string }{
		{`util.inspect(async function foo() {})`, "[AsyncFunction: foo]"},
		{`util.inspect(async function () {})`, "[AsyncFunction (anonymous)]"},
		{`util.inspect(function* foo() {})`, "[GeneratorFunction: foo]"},
		{`util.inspect(async function* foo() {})`, "[AsyncGeneratorFunction: foo]"},
	}
	for _, c := range cases {
		jstest.Equal(t, jstest.Str(t, e, c.expr), c.want)
	}
}

func TestTextCodecExports(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e,
		`util.TextEncoder === globalThis.TextEncoder && util.TextDecoder === globalThis.TextDecoder ? 'same' : 'different'`),
		"same")
	jstest.Equal(t, jstest.Str(t, e,
		`new util.TextDecoder().decode(new util.TextEncoder().encode('hé'))`), "hé")
	jstest.Equal(t, jstest.Str(t, e,
		`new util.TextDecoder('utf-16le').decode(new Uint8Array([0x68, 0x00, 0xe9, 0x00]))`), "hé")
}

func TestParseArgs(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var r = util.parseArgs({
			args: ['--foo', 'bar', '-v', 'x', 'y'],
			options: { foo: { type: 'string' }, v: { type: 'boolean', short: 'v' } },
			allowPositionals: true
		});
		return [r.values.foo, r.values.v, r.positionals.join(',')].join('|');
	})()`), "bar|true|x,y")
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var r = util.parseArgs({
			args: ['-n5', '--flag=false'],
			options: { n: { type: 'string', short: 'n' }, flag: { type: 'boolean' } }
		});
		return [r.values.n, r.values.flag].join('|');
	})()`), "5|false")
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		try { util.parseArgs({ args: ['--nope'], options: {} }); return 'no'; }
		catch (err) { return err.code; }
	})()`), "ERR_PARSE_ARGS_UNKNOWN_OPTION")
}
