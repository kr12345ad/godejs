package util

import (
	"strings"
	"testing"

	"godejs/internal/jstest"
)

func TestInspectMatchesNode(t *testing.T) {
	e := setup(t)
	cases := []struct{ expr, want string }{
		{`util.inspect({a:1,b:'two',c:[1,2,3],d:{e:true},f:null,g:undefined})`,
			"{\n  a: 1,\n  b: 'two',\n  c: [ 1, 2, 3 ],\n  d: { e: true },\n  f: null,\n  g: undefined\n}"},
		{`util.inspect({a:1,b:2})`, `{ a: 1, b: 2 }`},
		{`util.inspect({nested:{deep:{deeper:{deepest:{x:1}}}}})`, `{ nested: { deep: { deeper: [Object] } } }`},
		{`util.inspect([1,'a',null,undefined,[2,[3]]])`, `[ 1, 'a', null, undefined, [ 2, [ 3 ] ] ]`},
		{`util.inspect({x:{y:{z:{w:1}}}}, {depth:1})`, `{ x: { y: [Object] } }`},
		{`util.inspect([1,[2,[3,[4]]]], {depth:1})`, `[ 1, [ 2, [Array] ] ]`},
		{`util.inspect({a:1}, {depth:0})`, `{ a: 1 }`},
		{`util.inspect([1], {depth:0})`, `[ 1 ]`},
		{`util.inspect({a:{b:1}}, {depth:0})`, `{ a: [Object] }`},
		{`util.inspect({a:{}}, {depth:0})`, `{ a: {} }`},
		{`util.inspect({a:1}, {depth:-1})`, `[Object]`},
		{`util.inspect({}, {depth:-1})`, `{}`},
		{`util.inspect({__proto__:null}, {depth:-1})`, `[Object: null prototype] {}`},
		{`util.inspect({a:1}, {depth:null})`, `{ a: 1 }`},
		{`util.inspect({a:1,b:'two',c:[1,2,3],d:{e:true},f:null,g:undefined}, {breakLength:Infinity})`,
			`{ a: 1, b: 'two', c: [ 1, 2, 3 ], d: { e: true }, f: null, g: undefined }`},
		{`util.inspect({big:'x'.repeat(200)})`,
			"{\n  big: '" + strings.Repeat("x", 200) + "'\n}"},
		{`util.inspect(class Bar {})`, `[class Bar]`},
		{`util.inspect(function foo(){})`, `[Function: foo]`},
		{`util.inspect(function(){})`, `[Function (anonymous)]`},
		{`util.inspect(new Map([['a',1]]))`, `Map(1) { 'a' => 1 }`},
		{`util.inspect(new Set([1,2]))`, `Set(2) { 1, 2 }`},
		{`util.inspect('a'.repeat(100))`, `'` + strings.Repeat("a", 100) + `'`},
		{`util.inspect({a:undefined})`, `{ a: undefined }`},
		{`util.inspect([,1])`, `[ <1 empty item>, 1 ]`},
	}
	for _, c := range cases {
		if got := jstest.Str(t, e, c.expr); got != c.want {
			t.Errorf("%s\n got: %q\nwant: %q", c.expr, got, c.want)
		}
	}
}

func TestInspectLongArrayBreaks(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `util.inspect([1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30])`)
	want := "[\n  1,\n  2,\n  3,\n  4,\n  5,\n  6,\n  7,\n  8,\n  9,\n  10,\n  11,\n  12,\n  13,\n  14,\n  15,\n  16,\n  17,\n  18,\n  19,\n  20,\n  21,\n  22,\n  23,\n  24,\n  25,\n  26,\n  27,\n  28,\n  29,\n  30\n]"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestFormatMultiLineObject(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `util.format('label', {a:1,b:'two',c:[1,2,3],d:{e:true},f:null,g:undefined})`)
	want := "label {\n  a: 1,\n  b: 'two',\n  c: [ 1, 2, 3 ],\n  d: { e: true },\n  f: null,\n  g: undefined\n}"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestFormatPlaceholders(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `util.format('fmt %s', 'x')`), "fmt x")
	jstest.Equal(t, jstest.Str(t, e, `util.format('%s:%d', 'a', 1)`), "a:1")
	jstest.Equal(t, jstest.Str(t, e, `util.format('%j', {a:1})`), `{"a":1}`)
}

func TestInspectCustomHook(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `typeof util.inspect.custom`), "symbol")
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		class Point {
			constructor() { this.x = 1; this.y = 2; }
			[util.inspect.custom](depth, options, inspect) {
				return 'Point<' + this.x + ',' + this.y + '>';
			}
		}
		return util.inspect(new Point());
	})()`), "Point<1,2>")
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		const v = {};
		v[util.inspect.custom] = () => 'custom';
		return util.inspect({ v });
	})()`), "{ v: custom }")
}

func TestInspectClassPrefixAndSymbols(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(() => {
		class Foo { constructor() { this.a = 1; } }
		return util.inspect(new Foo());
	})()`), "Foo { a: 1 }")
	jstest.Equal(t, jstest.Str(t, e, `util.inspect({ [Symbol('k')]: 1 })`), "{ [Symbol(k)]: 1 }")
	jstest.Equal(t, jstest.Str(t, e, `util.inspect(Object.create(null))`), "[Object: null prototype] {}")
	jstest.Equal(t, jstest.Str(t, e, `util.inspect(new (class extends Object {})())`), "{}")
}

func TestInspectMaxArrayLength(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `util.inspect([1,2,3,4,5], { maxArrayLength: 3 })`),
		"[ 1, 2, 3, ... 2 more items ]")
	jstest.Equal(t, jstest.Str(t, e, `util.inspect([1,2], { maxArrayLength: 5 })`), "[ 1, 2 ]")
}

func TestStyleText(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `util.styleText('red', 'x', { validateStream: false })`),
		"\x1b[31mx\x1b[39m")
	jstest.Equal(t, jstest.Str(t, e, `util.styleText('grey', 'x', { validateStream: false })`),
		"\x1b[90mx\x1b[39m")
	jstest.Equal(t, jstest.Str(t, e, `util.styleText(['bold', 'red'], 'x', { validateStream: false })`),
		"\x1b[1m\x1b[31mx\x1b[39m\x1b[22m")
	jstest.Equal(t, jstest.Str(t, e, `util.styleText('none', 'x', { validateStream: false })`), "x")
	// Nested styles are re-opened after the inner close code.
	jstest.Equal(t, jstest.Str(t, e, `util.styleText('red', 'A' + util.styleText('blue', 'B', { validateStream: false }) + 'C', { validateStream: false })`),
		"\x1b[31mA\x1b[34mB\x1b[31mC\x1b[39m")
	jstest.Throws(t, e, `util.styleText('invalid', 'x')`)
	jstest.Throws(t, e, `util.styleText('red', 5)`)
}

func TestTypesExtras(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `String(util.types.isMapIterator(new Map().keys()))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(util.types.isSetIterator(new Set().values()))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(util.types.isArrayBufferView(new Uint8Array(1)))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(util.types.isNumberObject(Object(1)))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(util.types.isAnyArrayBuffer(new ArrayBuffer(1)))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(util.types.isWeakMap(new WeakMap()))`), "true")
	jstest.Equal(t, jstest.Str(t, e, `String(util.types.isAnyArrayBuffer({}))`), "false")
}

func TestDeprecateValidation(t *testing.T) {
	e := setup(t)
	// A non-string code is rejected before any warning is emitted.
	err := jstest.Throws(t, e, `util.deprecate(() => {}, 'message', 5)`)
	if err.Code != "ERR_INVALID_ARG_TYPE" {
		t.Fatalf("code = %q, want ERR_INVALID_ARG_TYPE", err.Code)
	}
}

func TestInspectErrorFirstLine(t *testing.T) {
	e := setup(t)
	got := jstest.Str(t, e, `util.inspect(new Error('boom'))`)
	if !strings.HasPrefix(got, "Error: boom") {
		t.Fatalf("got %q", got)
	}
}
