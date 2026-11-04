package querystring

import (
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "qs", Install)
	return e
}

func TestParseBasics(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a=1&b=2"))`), `{"a":"1","b":"2"}`)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a"))`), `{"a":""}`)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("=v"))`), `{"":"v"}`)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a=b=c"))`), `{"a":"b=c"}`)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a=1&&b=2&"))`), `{"a":"1","b":"2"}`)
	jstest.Equal(t, jstest.Str(t, e, `Object.getPrototypeOf(qs.parse("a=1")) === null`), "true")
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse(""))`), `{}`)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a=1;b=2", ";"))`), `{"a":"1","b":"2"}`)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a:1|b:2", "|", ":"))`), `{"a":"1","b":"2"}`)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("__proto__=x"))`), `{"__proto__":"x"}`)
}

func TestStringifyPrimitivesOnly(t *testing.T) {
	e := setup(t)
	// Node only serializes strings, finite numbers, bigints and booleans;
	// objects (a RegExp here) contribute an empty value.
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({ regexp: /./g })`), "regexp=")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({ a: 1, b: true, c: 'x' })`), "a=1&b=true&c=x")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({ a: NaN, b: Infinity })`), "a=&b=")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({ a: 10n })`), "a=10")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({ a: [1, 2], b: { c: 1 } })`), "a=1&a=2&b=")
	// Node 24 emits every own key: values that are not strings, finite
	// numbers, bigints or booleans format as an empty value.
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({ a: undefined, b: null, c: () => {} })`), "a=&b=&c=")
}

func TestParseDuplicates(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a=1&a=2&a=3"))`), `{"a":["1","2","3"]}`)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a&a"))`), `{"a":["",""]}`)
}

func TestParseMaxKeys(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e,
		`JSON.stringify(qs.parse("a=1&b=2&c=3", null, null, {maxKeys: 2}))`), `{"a":"1","b":"2"}`)
	jstest.Equal(t, jstest.Str(t, e,
		`JSON.stringify(qs.parse("a=1&b=2", null, null, {maxKeys: 0}))`), `{"a":"1","b":"2"}`)
	jstest.Equal(t, jstest.Str(t, e,
		`JSON.stringify(qs.parse("a=1&a=2&b=3", null, null, {maxKeys: 1}))`), `{"a":"1"}`)
}

func TestDecode(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `qs.parse("a=hello+world").a`), "hello world")
	jstest.Equal(t, jstest.Str(t, e, `qs.parse("a=%E2%82%AC").a`), "\u20ac")
	jstest.Equal(t, jstest.Str(t, e, `qs.parse("a=%zz%20").a`), "%zz ")
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a%20b=c%20d"))`), `{"a b":"c d"}`)
	jstest.Equal(t, jstest.Str(t, e, `qs.parse("a+b=c+d")["a b"]`), "c d")
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.parse("a%2Bb=c"))`), `{"a+b":"c"}`)
}

func TestCustomDecoder(t *testing.T) {
	e := setup(t)
	jstest.Str(t, e, `var calls = 0`)
	jstest.Equal(t, jstest.Str(t, e,
		`JSON.stringify(qs.parse("a=1", null, null, {decodeURIComponent: function (s) { calls++; return s.toUpperCase(); }}))`),
		`{"A":"1"}`)
	jstest.Equal(t, jstest.Str(t, e, `calls`), "2")
}

func TestStringify(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({a: 1, b: "two"})`), "a=1&b=two")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({})`), "")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({b: 2, a: 1})`), "b=2&a=1")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({a: [1, 2], b: null, c: undefined, d: ""})`), "a=1&a=2&b=&c=&d=")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({a: [null, undefined, 1]})`), "a=&a=&a=1")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({a: true, b: 0})`), "a=true&b=0")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify(null)`), "")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify("x")`), "")
	jstest.Equal(t, jstest.Str(t, e,
		`qs.stringify({a: "hello world", b: "x+y", c: "\u20ac"})`),
		"a=hello%20world&b=x%2By&c=%E2%82%AC")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({a: 1, b: 2}, ";")`), "a=1;b=2")
	jstest.Equal(t, jstest.Str(t, e, `qs.stringify({a: 1}, null, ":")`), "a:1")
}

func TestCustomEncoder(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e,
		`qs.stringify({a: "x", b: "y"}, null, null, {encodeURIComponent: function (s) { return "[" + s + "]"; }})`),
		"[a]=[x]&[b]=[y]")
}

func TestAliasesAreIdentical(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `qs.decode === qs.parse`), "true")
	jstest.Equal(t, jstest.Str(t, e, `qs.encode === qs.stringify`), "true")
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(qs.decode("a=1"))`), `{"a":"1"}`)
	jstest.Equal(t, jstest.Str(t, e, `qs.encode({a: 1})`), "a=1")
}

func TestEscapeUnescape(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `qs.escape("a b+c/d?e=f&g")`), "a%20b%2Bc%2Fd%3Fe%3Df%26g")
	jstest.Equal(t, jstest.Str(t, e, `qs.escape("!'()*-._~Az9")`), "!'()*-._~Az9")
	jstest.Equal(t, jstest.Str(t, e, `qs.escape("\u20ac")`), "%E2%82%AC")
	jstest.Equal(t, jstest.Str(t, e, `qs.unescape("a%20b%2Bc")`), "a b+c")
	jstest.Equal(t, jstest.Str(t, e, `qs.unescape("%E2%82%AC")`), "\u20ac")
	jstest.Equal(t, jstest.Str(t, e, `qs.unescape("%zz")`), "%zz")
	jstest.Equal(t, jstest.Str(t, e, `qs.unescape("a+b")`), "a+b")
}

func TestImplementedMatchesBaseline(t *testing.T) {
	want := map[string]bool{
		"querystring.decode":    true,
		"querystring.encode":    true,
		"querystring.escape":    true,
		"querystring.parse":     true,
		"querystring.stringify": true,
		"querystring.unescape":  true,
	}
	got := Implemented()
	if len(got) != len(want) {
		t.Fatalf("Implemented(): got %d keys, want %d", len(got), len(want))
	}
	for _, k := range got {
		if !want[k] {
			t.Fatalf("unexpected implemented key %q", k)
		}
	}
}
