package url

import (
	"encoding/json"
	"os"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
	"godejs/modules/buffer"
)

func setup(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "url", Install)
	return e
}

func TestURLNormalization(t *testing.T) {
	e := setup(t)
	href := jstest.Str(t, e, `new URL('HTTPS://User:Pass@Example.COM:443/a b/../c?x=1 2#f').href`)
	jstest.Equal(t, href, "https://User:Pass@example.com/c?x=1%202#f")

	cases := map[string]string{
		`u.protocol`: "https:",
		`u.hostname`: "example.com",
		`u.host`:     "example.com",
		`u.port`:     "",
		`u.origin`:   "https://example.com",
		`u.pathname`: "/c",
		`u.search`:   "?x=1%202",
		`u.hash`:     "#f",
		`u.username`: "User",
		`u.password`: "Pass",
	}
	jstest.Eval(t, e, `var u = new URL('HTTPS://User:Pass@Example.COM:443/a b/../c?x=1 2#f')`)
	for expr, want := range cases {
		jstest.Equal(t, jstest.Str(t, e, expr), want)
	}
}

func TestURLEmptyPathAndPort(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `new URL('http://example.com').href`), "http://example.com/")
	jstest.Equal(t, jstest.Str(t, e, `new URL('https://example.com:443/x').href`), "https://example.com/x")
	jstest.Equal(t, jstest.Str(t, e, `new URL('http://example.com:8080').href`), "http://example.com:8080/")
	jstest.Equal(t, jstest.Str(t, e, `new URL('mailto:foo@example.com').href`), "mailto:foo@example.com")
	jstest.Equal(t, jstest.Str(t, e, `new URL('file:///tmp/a b').href`), "file:///tmp/a%20b")
}

func TestRelativeResolution(t *testing.T) {
	e := setup(t)
	cases := [][2]string{
		{`new URL('../x?q=1', 'https://a.com/d/e/f').href`, "https://a.com/d/x?q=1"},
		{`new URL('/abs', 'https://a.com/d/e/f').href`, "https://a.com/abs"},
		{`new URL('//cdn.com/a', 'https://a.com/d').href`, "https://cdn.com/a"},
		{`new URL('?z=9', 'https://a.com/p?q=1#h').href`, "https://a.com/p?z=9"},
		{`new URL('#frag', 'https://a.com/p?q=1').href`, "https://a.com/p?q=1#frag"},
		{`new URL('', 'https://a.com/p?q=1#h').href`, "https://a.com/p?q=1"},
	}
	for _, c := range cases {
		jstest.Equal(t, jstest.Str(t, e, c[0]), c[1])
	}
}

func TestURLSetters(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `var u = new URL('http://example.com/a?x=1#h')`)
	steps := [][2]string{
		{`u.pathname = '/a b'; u.href`, "http://example.com/a%20b?x=1#h"},
		{`u.search = '?q=2'; u.href`, "http://example.com/a%20b?q=2#h"},
		{`u.hash = 'top'; u.href`, "http://example.com/a%20b?q=2#top"},
		{`u.port = '8080'; u.href`, "http://example.com:8080/a%20b?q=2#top"},
		{`u.hostname = 'EXAMPLE.org'; u.href`, "http://example.org:8080/a%20b?q=2#top"},
		{`u.username = 'u'; u.password = 'p'; u.href`, "http://u:p@example.org:8080/a%20b?q=2#top"},
	}
	for _, s := range steps {
		jstest.Equal(t, jstest.Str(t, e, s[0]), s[1])
	}
}

func TestSearchParamsLinkedToURL(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `var u = new URL('https://x/?a=1&a=2&b=3')`)
	for expr, want := range map[string]string{
		`u.searchParams.getAll('a').join(',')`:      "1,2",
		`u.searchParams.get('b')`:                   "3",
		`u.searchParams.has('a', '2')`:              "true",
		`u.searchParams.size`:                       "3",
		`u.searchParams instanceof URLSearchParams`: "true",
	} {
		jstest.Equal(t, jstest.Str(t, e, expr), want)
	}
	jstest.Equal(t, jstest.Str(t, e, `u.searchParams.append('c', '4'); u.href`),
		"https://x/?a=1&a=2&b=3&c=4")
	jstest.Equal(t, jstest.Str(t, e, `u.searchParams.set('a', '9'); u.href`),
		"https://x/?a=9&b=3&c=4")
	jstest.Equal(t, jstest.Str(t, e, `u.searchParams.delete('b'); u.href`),
		"https://x/?a=9&c=4")
	jstest.Equal(t, jstest.Str(t, e, `u.searchParams.sort(); u.href`),
		"https://x/?a=9&c=4")
	jstest.Equal(t, jstest.Str(t, e, `u.searchParams.toString()`), "a=9&c=4")
	jstest.Equal(t, jstest.Str(t, e, `[...u.searchParams.keys()].join(',')`), "a,c")
	jstest.Equal(t, jstest.Str(t, e, `[...u.searchParams.values()].join(',')`), "9,4")
	jstest.Equal(t, jstest.Str(t, e, `[...u.searchParams].join('|')`), "a,9|c,4")
}

func TestURLSearchParams(t *testing.T) {
	e := setup(t)
	for expr, want := range map[string]string{
		`new URLSearchParams('?a=1&b=hello+world&c=%7E').toString()`:          "a=1&b=hello+world&c=%7E",
		`new URLSearchParams({a: 'b c', n: 2}).toString()`:                    "a=b+c&n=2",
		`new URLSearchParams([['x', '1'], ['x', '2']]).getAll('x').join(',')`: "1,2",
		`new URLSearchParams(new Map([['a', '1']])).toString()`:               "a=1",
		`new URLSearchParams(new URLSearchParams('a=1')).toString()`:          "a=1",
	} {
		jstest.Equal(t, jstest.Str(t, e, expr), want)
	}
	jstest.Equal(t, jstest.Str(t, e, `var p = new URLSearchParams('b=2&a=1&b=3'); p.set('b','9'); p.toString()`), "b=9&a=1")
	jstest.Equal(t, jstest.Str(t, e, `p.sort(); p.toString()`), "a=1&b=9")
	jstest.Equal(t, jstest.Str(t, e, `p.has('b')`), "true")
	jstest.Equal(t, jstest.Str(t, e, `p.has('b','3')`), "false")
	jstest.Equal(t, jstest.Str(t, e, `p.delete('b'); p.toString()`), "a=1")
	jstest.Equal(t, jstest.Str(t, e, `var seen=[]; new URLSearchParams('a=1&b=2').forEach(function(v,k){seen.push(k+'='+v)}); seen.join('&')`), "a=1&b=2")
	jstest.Equal(t, jstest.Str(t, e, `p.size`), "1")
}

func TestFileURLConversion(t *testing.T) {
	e := setup(t)
	for expr, want := range map[string]string{
		`url.fileURLToPath('file:///tmp/a%20b')`:              "/tmp/a b",
		`url.fileURLToPath(new URL('file:///tmp/a%20b'))`:     "/tmp/a b",
		`url.fileURLToPath(new URL('file:///tmp/%E2%9C%93'))`: "/tmp/✓",
		`url.pathToFileURL('/tmp/a b').href`:                  "file:///tmp/a%20b",
		`url.pathToFileURL('/tmp/✓').href`:                    "file:///tmp/%E2%9C%93",
		`url.fileURLToPath(url.pathToFileURL('/tmp/x y'))`:    "/tmp/x y",
	} {
		jstest.Equal(t, jstest.Str(t, e, expr), want)
	}
	jstest.Equal(t, jstest.Str(t, e, `try { url.fileURLToPath('http://x/') } catch (e) { e.code }`), "ERR_INVALID_URL_SCHEME")
	jstest.Equal(t, jstest.Str(t, e, `try { url.fileURLToPath('file://example.com/x') } catch (e) { e.code }`), "ERR_INVALID_FILE_URL_HOST")
}

func TestLegacyAPI(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `var p = url.parse('https://u:p@h.com:8443/a?b=1#c')`)
	fields := map[string]string{
		`p.protocol`:    "https:",
		`p.host`:        "h.com:8443",
		`p.hostname`:    "h.com",
		`p.port`:        "8443",
		`p.pathname`:    "/a",
		`p.search`:      "?b=1",
		`p.hash`:        "#c",
		`p.query`:       "b=1",
		`p.path`:        "/a?b=1",
		`url.format(p)`: "https://u:p@h.com:8443/a?b=1#c",
	}
	for expr, want := range fields {
		jstest.Equal(t, jstest.Str(t, e, expr), want)
	}
	jstest.Equal(t, jstest.Str(t, e, `url.parse('http://x/?a=1&a=2', true).query.a.join(',')`), "1,2")
	jstest.Equal(t, jstest.Str(t, e, `url.resolve('https://a.com/d/e/f', '../x')`), "https://a.com/d/x")
	jstest.Equal(t, jstest.Str(t, e, `url.parse('a/b?c=1').pathname`), "a/b")
}

func TestURLToHTTPOptions(t *testing.T) {
	e := setup(t)
	jstest.Eval(t, e, `var o = url.urlToHttpOptions(new URL('http://u:p@h.example:8080/a?b=1#c'))`)
	for expr, want := range map[string]string{
		`o.protocol`: "http:",
		`o.hostname`: "h.example",
		`o.hash`:     "#c",
		`o.search`:   "?b=1",
		`o.pathname`: "/a",
		`o.path`:     "/a?b=1",
		`o.port`:     "8080",
		`o.auth`:     "u:p",
		`o.href`:     "http://u:p@h.example:8080/a?b=1#c",
	} {
		jstest.Equal(t, jstest.Str(t, e, expr), want)
	}
}

func TestDomainConversion(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `url.domainToASCII('EXAMPLE.COM')`), "example.com")
	jstest.Equal(t, jstest.Str(t, e, `url.domainToUnicode('example.com')`), "example.com")
	jstest.Equal(t, jstest.Str(t, e, `url.domainToASCII('español.com').slice(0, 4)`), "xn--")
	jstest.Equal(t, jstest.Str(t, e, `url.domainToUnicode(url.domainToASCII('español.com'))`), "español.com")
	jstest.Equal(t, jstest.Str(t, e, `url.domainToUnicode(url.domainToASCII('münchen.de'))`), "münchen.de")
}

func TestInvalidURL(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `try { new URL('not a url') } catch (e) { e.name + '|' + e.code }`), "TypeError|ERR_INVALID_URL")
	jstest.Equal(t, jstest.Str(t, e, `URL.canParse('not a url')`), "false")
	jstest.Equal(t, jstest.Str(t, e, `URL.canParse('/x', 'https://a.com')`), "true")
	jstest.Equal(t, jstest.Str(t, e, `URL.parse('not a url') === null`), "true")
	jstest.Equal(t, jstest.Str(t, e, `URL.parse('/x', 'https://a.com').href`), "https://a.com/x")
}

func TestGlobals(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `globalThis.URL === url.URL`), "true")
	jstest.Equal(t, jstest.Str(t, e, `globalThis.URLSearchParams === url.URLSearchParams`), "true")
	jstest.Equal(t, jstest.Str(t, e, `typeof new URL('https://x/').toString`), "function")
}

func TestImplementedKeysCoverBaseline(t *testing.T) {
	data, err := os.ReadFile("../../parity/baseline/url.json")
	if err != nil {
		t.Fatalf("baseline: %v", err)
	}
	var baseline []string
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatalf("baseline: %v", err)
	}
	tracked := map[string]bool{}
	for _, k := range baseline {
		tracked[k] = true
	}
	seen := map[string]bool{}
	for _, k := range Implemented() {
		if !tracked[k] {
			t.Errorf("Implemented() lists untracked key %q", k)
		}
		if seen[k] {
			t.Errorf("duplicate key %q", k)
		}
		seen[k] = true
	}
	if len(seen) != len(tracked) {
		t.Errorf("Implemented() covers %d of %d baseline keys", len(seen), len(tracked))
	}
}

func TestLegacyAccessors(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var u = new URL('https://user:pw@example.com/a/b?q=1#frag');
		return [u.auth, u.path, u.query, u.slashes].join('|');
	})()`), "user:pw|/a/b?q=1|?q=1|true")
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var u = new URL('https://example.com/a');
		u.auth = 'me:secret';
		u.path = '/x?y=2';
		return [u.username, u.password, u.pathname, u.search].join('|');
	})()`), "me|secret|/x|?y=2")
}

func TestURLPattern(t *testing.T) {
	e := setup(t)
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var p = new URLPattern({ pathname: '/users/:id', hostname: 'example.com' });
		var m = p.exec('https://example.com/users/42?x=1');
		return [p.test('https://example.com/users/42'), m.pathname.groups.id,
			p.test('https://example.com/nope')].join('|');
	})()`), "true|42|false")
}

func TestObjectURLs(t *testing.T) {
	e := jstest.New(t)
	jstest.Install(t, e, "buffer", buffer.Install)
	jstest.Install(t, e, "url", Install)
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var blob = new Blob(['hello']);
		var id = URL.createObjectURL(blob);
		var back = buffer.resolveObjectURL(id);
		var same = back === blob;
		URL.revokeObjectURL(id);
		return [same, buffer.resolveObjectURL(id) === undefined].join('|');
	})()`), "true|true")
}
