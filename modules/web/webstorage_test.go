package web

import (
	"path/filepath"
	"testing"

	"godejs/internal/jstest"
)

func TestStorageGlobals(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	e := setup(t)

	jstest.Equal(t, jstest.Str(t, e,
		`[typeof Storage, typeof localStorage, typeof sessionStorage, typeof QuotaExceededError].join(',')`),
		"function,object,object,function")
	jstest.Equal(t, jstest.Str(t, e,
		`localStorage instanceof Storage && sessionStorage instanceof Storage ? 'yes' : 'no'`), "yes")
	// Web IDL requires TypeError for an illegal constructor (Node itself
	// throws a plain Error here, but the platform tests expect the spec type).
	jstest.Equal(t, jstest.Str(t, e,
		`(function () { try { new Storage(); return 'built'; } catch (err) { return err.name + ':' + err.message; } })()`),
		"TypeError:Illegal constructor")
	// Node's Storage carries no Symbol.toStringTag: it reports as a plain object.
	jstest.Equal(t, jstest.Str(t, e, `Object.prototype.toString.call(localStorage)`), "[object Object]")
}

func TestStorageOperations(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	e := setup(t)

	got := jstest.Str(t, e, `
	  var s = sessionStorage;
	  s.setItem('a', 1);
	  s.setItem('b', '');
	  var out = [s.length, s.getItem('a'), s.getItem('b') === '',
	             s.getItem('zz'), s.key(0), s.key(9), s.getItem('a') === '1'];
	  s.removeItem('a');
	  out.push(s.length, s.getItem('a'));
	  JSON.stringify(out)
	`)
	jstest.Equal(t, got, `[2,"1",true,null,"a",null,true,1,null]`)

	jstest.Equal(t, jstest.Str(t, e, `
	  sessionStorage.clear();
	  sessionStorage.setItem('only', 'session');
	  localStorage.clear();
	  localStorage.setItem('only', 'local');
	  JSON.stringify([sessionStorage.getItem('only'), localStorage.getItem('only')])
	`), `["session","local"]`)
}

// Web Storage exposes every entry as a named property, and assignment through
// that view must create items rather than shadow the API.
func TestStorageNamedProperties(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	e := setup(t)

	got := jstest.Str(t, e, `
	  var s = sessionStorage;
	  s.clear();
	  s.setItem('x', '1');
	  s.y = 2;
	  s['z'] = 3;
	  JSON.stringify([s.x, s.y, s.z, s.getItem('y'),
	    Object.prototype.hasOwnProperty.call(s, 'x'), 'x' in s,
	    typeof s.missing, Object.keys(s), s.length])
	`)
	jstest.Equal(t, got, `["1","2","3","2",true,true,"undefined",["x","y","z"],3]`)

	got = jstest.Str(t, e, `
	  var s = sessionStorage;
	  delete s.y;
	  JSON.stringify([s.getItem('y'), Object.keys(s)])
	`)
	jstest.Equal(t, got, `[null,["x","z"]]`)

	// Reserved names still resolve to the API, not to entries.
	jstest.Equal(t, jstest.Str(t, e, `
	  var s = sessionStorage;
	  s.setItem('length', 'not-a-length');
	  s.setItem('getItem', 'not-a-method');
	  JSON.stringify([typeof s.getItem, s.length, s.getItem('length'), s.getItem('getItem')])
	`), `["function",4,"not-a-length","not-a-method"]`)
}

func TestLocalStoragePersistsAcrossEnvs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage.json")
	t.Setenv("GODE_LOCALSTORAGE_FILE", path)

	first := setup(t)
	jstest.Equal(t, jstest.Str(t, first, `
	  localStorage.clear();
	  localStorage.setItem('k', 'v"q');
	  localStorage.setItem('n', '1');
	  localStorage.getItem('k')
	`), `v"q`)

	second := setup(t)
	got := jstest.Str(t, second, `
	  JSON.stringify([localStorage.getItem('k'), localStorage.getItem('n'),
	                  localStorage.length, Object.keys(localStorage)])
	`)
	jstest.Equal(t, got, `["v\"q","1",2,["k","n"]]`)

	// sessionStorage never outlives the process.
	jstest.Equal(t, jstest.Str(t, second, `JSON.stringify([sessionStorage.length, sessionStorage.getItem('k')])`), `[0,null]`)
}

// Web IDL conversions and brand checks: symbols must never stringify into
// storage keys/values, key() takes an unsigned long, and borrowed methods must
// reject receivers that are not Storage areas.
func TestStorageWebIDLSemantics(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	e := setup(t)

	got := jstest.Str(t, e, `
	  var out = [];
	  function t(fn) { try { fn(); out.push('ok'); } catch (err) { out.push(err.name); } }
	  t(function () { new Storage(); });
	  t(function () { sessionStorage.setItem(Symbol('k'), 'v'); });
	  t(function () { sessionStorage.setItem('k', Symbol('v')); });
	  t(function () { sessionStorage.getItem(Symbol('k')); });
	  t(function () { sessionStorage.removeItem(Symbol('k')); });
	  t(function () { sessionStorage.key(Symbol('k')); });
	  sessionStorage.clear();
	  sessionStorage.setItem('a', '1');
	  sessionStorage.setItem('b', '2');
	  out.push(sessionStorage.key(NaN), sessionStorage.key(-1), sessionStorage.key(1.9),
	           sessionStorage.key(Infinity), sessionStorage.key(4294967297));
	  t(function () { Storage.prototype.getItem.call({}, 'a'); });
	  t(function () { Storage.prototype.setItem.call({}, 'a', 'b'); });
	  t(function () { Storage.prototype.clear.call({}); });
	  t(function () { Storage.prototype.key.call({}, 0); });
	  t(function () { Object.getOwnPropertyDescriptor(Storage.prototype, 'length').get.call({}); });
	  out.push(Storage.prototype.getItem.call(sessionStorage, 'a'));
	  JSON.stringify(out)
	`)
	jstest.Equal(t, got, `["TypeError","TypeError","TypeError","TypeError","TypeError","TypeError","a",null,"b","a","b","TypeError","TypeError","TypeError","TypeError","TypeError","1"]`)
}

func TestNavigator(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	e := setup(t)

	got := jstest.Str(t, e, `[
	  navigator.userAgent.indexOf('Node.js/') === 0,
	  typeof navigator.platform,
	  navigator.platform.length > 0,
	  typeof navigator.hardwareConcurrency,
	  navigator.hardwareConcurrency > 0,
	  Array.isArray(navigator.languages),
	  navigator.language === navigator.languages[0],
	  Object.getPrototypeOf(navigator).constructor.name
	].join(',')`)
	jstest.Equal(t, got, "true,string,true,number,true,true,true,Navigator")
}
