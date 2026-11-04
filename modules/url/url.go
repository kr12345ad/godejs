// Package url implements Node.js's url module and the WHATWG URL and
// URLSearchParams globals on top of the GodeJS runtime.
package url

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"godejs/internal/env"
	"godejs/internal/js"
)

func init() { env.Register("url", Install) }

// urlPrelude builds the JS-visible classes. All parsing, serialization and
// normalization stays in Go; the classes only wire accessors and iteration.
const urlPrelude = `(function (host) {
  'use strict';
  function makeIterator(owner, kind) {
    var entries = host.paramsEntries.call(owner);
    var i = 0;
    var it = {
      next: function () {
        if (i >= entries.length) return { done: true, value: undefined };
        var pair = entries[i++];
        var value = kind === 2 ? [pair[0], pair[1]] : (kind === 1 ? pair[1] : pair[0]);
        return { done: false, value: value };
      }
    };
    it[Symbol.iterator] = function () { return this; };
    return it;
  }
  class URLSearchParams {
    constructor(init) {
      if (!new.target) {
        throw new TypeError("Class constructor URLSearchParams cannot be invoked without 'new'");
      }
      if (init !== undefined && init !== null && typeof init !== 'string' &&
          !Array.isArray(init) && typeof init[Symbol.iterator] === 'function') {
        init = Array.from(init);
      }
      host.paramsInit.call(this, init);
    }
    get size() { return host.paramsSize.call(this); }
    append(name, value) { host.paramsAppend.call(this, name, value); }
    delete(name, value) { host.paramsDelete.call(this, name, value); }
    get(name) { return host.paramsGet.call(this, name); }
    getAll(name) { return host.paramsGetAll.call(this, name); }
    has(name, value) { return host.paramsHas.call(this, name, value); }
    set(name, value) { host.paramsSet.call(this, name, value); }
    sort() { host.paramsSort.call(this); }
    forEach(callback, thisArg) {
      if (typeof callback !== 'function') {
        throw new TypeError('The "callback" argument must be of type function');
      }
      var entries = host.paramsEntries.call(this);
      for (var i = 0; i < entries.length; i++) {
        callback.call(thisArg, entries[i][1], entries[i][0], this);
      }
    }
    entries() { return makeIterator(this, 2); }
    keys() { return makeIterator(this, 0); }
    values() { return makeIterator(this, 1); }
    [Symbol.iterator]() { return this.entries(); }
    toString() { return host.paramsToString.call(this); }
  }
  Object.defineProperty(URLSearchParams.prototype, Symbol.toStringTag, { value: 'URLSearchParams', configurable: true });

  class URL {
    constructor(input, base) {
      if (!new.target) {
        throw new TypeError("Class constructor URL cannot be invoked without 'new'");
      }
      host.urlInit.call(this, input, base);
    }
    get href() { return host.urlGet.call(this, 'href'); }
    set href(v) { host.urlSet.call(this, 'href', v); }
    get origin() { return host.urlGet.call(this, 'origin'); }
    get protocol() { return host.urlGet.call(this, 'protocol'); }
    set protocol(v) { host.urlSet.call(this, 'protocol', v); }
    get username() { return host.urlGet.call(this, 'username'); }
    set username(v) { host.urlSet.call(this, 'username', v); }
    get password() { return host.urlGet.call(this, 'password'); }
    set password(v) { host.urlSet.call(this, 'password', v); }
    get host() { return host.urlGet.call(this, 'host'); }
    set host(v) { host.urlSet.call(this, 'host', v); }
    get hostname() { return host.urlGet.call(this, 'hostname'); }
    set hostname(v) { host.urlSet.call(this, 'hostname', v); }
    get port() { return host.urlGet.call(this, 'port'); }
    set port(v) { host.urlSet.call(this, 'port', v); }
    get pathname() { return host.urlGet.call(this, 'pathname'); }
    set pathname(v) { host.urlSet.call(this, 'pathname', v); }
    get search() { return host.urlGet.call(this, 'search'); }
    set search(v) { host.urlSet.call(this, 'search', v); }
    get searchParams() { return host.urlParams.call(this); }
    get hash() { return host.urlGet.call(this, 'hash'); }
    set hash(v) { host.urlSet.call(this, 'hash', v); }
    // The legacy accessors Node keeps for url.parse compatibility.
    get auth() {
      var user = this.username, pass = this.password;
      return pass ? user + ':' + pass : user;
    }
    set auth(v) {
      v = String(v);
      var i = v.indexOf(':');
      if (i >= 0) { this.username = v.slice(0, i); this.password = v.slice(i + 1); }
      else { this.username = v; this.password = ''; }
    }
    get path() { return this.pathname + this.search; }
    set path(v) {
      v = String(v);
      var q = v.indexOf('?');
      if (q >= 0) { this.pathname = v.slice(0, q); this.search = v.slice(q); }
      else { this.pathname = v; this.search = ''; }
    }
    get query() { return this.search; }
    set query(v) { this.search = v; }
    get slashes() {
      var href = this.href;
      var colon = href.indexOf(':');
      return colon >= 0 && href.slice(colon + 1, colon + 3) === '//';
    }
    set slashes(v) {
      if (!v) return;
      var href = this.href;
      var colon = href.indexOf(':');
      if (colon >= 0 && href.slice(colon + 1, colon + 3) !== '//') {
        this.href = href.slice(0, colon + 1) + '//' + href.slice(colon + 1);
      }
    }
    toJSON() { return host.urlGet.call(this, 'href'); }
    toString() { return host.urlGet.call(this, 'href'); }
    static canParse(input, base) { return host.urlCanParse(input, base); }
    static parse(input, base) {
      try { return new URL(input, base); } catch (e) { return null; }
    }
    static createObjectURL(blob) {
      if (blob === null || typeof blob !== 'object' || typeof blob.size !== 'number' ||
          typeof blob.type !== 'string' || typeof blob.slice !== 'function') {
        throw new TypeError('The "blob" argument must be an instance of Blob. Received ' +
          (blob === null ? 'null' : typeof blob));
      }
      objectURLCounter++;
      var id = 'blob:nodedata:' + Date.now().toString(16) + '-' + objectURLCounter;
      objectURLs.set(id, blob);
      return id;
    }
    static revokeObjectURL(id) {
      if (arguments.length === 0) {
        var missing = new TypeError('The "id" argument must be specified');
        missing.code = 'ERR_MISSING_ARGS';
        throw missing;
      }
      if (typeof id !== 'string') {
        throw new TypeError('The "id" argument must be of type string. Received ' + describeValue(id));
      }
      objectURLs.delete(id);
    }
  }
  Object.defineProperty(URL.prototype, Symbol.toStringTag, { value: 'URL', configurable: true });

  // Object URLs live in a process-wide Map so buffer.resolveObjectURL can find
  // the Blob that URL.createObjectURL registered without importing url.
  var objectURLs = globalThis.__godeObjectURLs;
  if (!(objectURLs instanceof Map)) {
    objectURLs = new Map();
    globalThis.__godeObjectURLs = objectURLs;
  }
  var objectURLCounter = 0;

  function describeValue(value) {
    if (value === null) return 'null';
    var t = typeof value;
    if (t === 'string') return "type string ('" + value + "')";
    if (t === 'number') return 'type number (' + value + ')';
    if (t === 'boolean') return 'type boolean (' + value + ')';
    if (t === 'undefined') return 'undefined';
    if (Array.isArray(value)) return 'an instance of Array';
    return 'an instance of ' + ((value.constructor && value.constructor.name) || 'Object');
  }

  var BACKSLASH = String.fromCharCode(92);
  var SPECIAL = '^$.*+?()[]{}|';
  function escapeRegExp(text) {
    var out = '';
    for (var i = 0; i < text.length; i++) {
      var c = text[i];
      out += SPECIAL.indexOf(c) >= 0 ? BACKSLASH + c : c;
    }
    return out;
  }

  // compileComponent turns one URLPattern component into a matcher. Star is a
  // wildcard and :name captures one path segment into a named group.
  function compileComponent(value) {
    var source = value === undefined || value === null ? '*' : String(value);
    var names = [];
    var out = '';
    for (var i = 0; i < source.length; i++) {
      var c = source[i];
      if (c === '*') { out += '(.*)'; names.push(null); continue; }
      if (c === ':') {
        var j = i + 1;
        while (j < source.length && /[A-Za-z0-9_]/.test(source[j])) j++;
        if (j > i + 1) {
          out += '([^/]+)';
          names.push(source.slice(i + 1, j));
          i = j - 1;
          continue;
        }
      }
      out += escapeRegExp(c);
    }
    return { source: source, regexp: new RegExp('^' + out + '$', 'i'), names: names };
  }

  class URLPattern {
    constructor(input, baseURL, options) {
      if (!new.target) {
        throw new TypeError("Class constructor URLPattern cannot be invoked without 'new'");
      }
      if (baseURL !== null && typeof baseURL === 'object' && options === undefined) {
        options = baseURL;
        baseURL = undefined;
      }
      options = options || {};
      var parts;
      if (typeof input === 'string') {
        var absolute = null;
        try { absolute = baseURL === undefined ? new URL(input) : new URL(input, baseURL); }
        catch (e) { absolute = null; }
        if (absolute) {
          parts = {
            protocol: absolute.protocol, username: absolute.username,
            password: absolute.password, hostname: absolute.hostname,
            port: absolute.port, pathname: absolute.pathname,
            search: absolute.search, hash: absolute.hash
          };
        } else if (baseURL !== undefined) {
          var base = new URL(baseURL);
          var relative = new URL(input, base);
          parts = {
            protocol: base.protocol, username: '*', password: '*',
            hostname: base.hostname, port: '*', pathname: relative.pathname,
            search: relative.search, hash: relative.hash
          };
        } else {
          throw new TypeError('Invalid URL pattern: ' + input);
        }
      } else if (input !== null && typeof input === 'object') {
        parts = input;
      } else {
        throw new TypeError('The "input" argument must be of type string or an object');
      }
      this._protocol = compileComponent(parts.protocol);
      this._username = compileComponent(parts.username);
      this._password = compileComponent(parts.password);
      this._hostname = compileComponent(parts.hostname);
      this._port = compileComponent(parts.port);
      this._pathname = compileComponent(parts.pathname);
      this._search = compileComponent(parts.search);
      this._hash = compileComponent(parts.hash);
    }
    get protocol() { return this._protocol.source; }
    get username() { return this._username.source; }
    get password() { return this._password.source; }
    get hostname() { return this._hostname.source; }
    get port() { return this._port.source; }
    get pathname() { return this._pathname.source; }
    get search() { return this._search.source; }
    get hash() { return this._hash.source; }
    get hasRegExpGroups() { return false; }
    test(input, baseURL) { return this.exec(input, baseURL) !== null; }
    exec(input, baseURL) {
      var url;
      try { url = baseURL === undefined ? new URL(input) : new URL(input, baseURL); }
      catch (e) { return null; }
      var matched = this;
      function match(component, value) {
        var m = component.regexp.exec(value);
        if (!m) return null;
        var groups = {};
        for (var i = 0; i < component.names.length; i++) {
          if (component.names[i]) {
            groups[component.names[i]] = m[i + 1] === undefined ? '' : m[i + 1];
          }
        }
        return { input: value, groups: groups };
      }
      var result = {
        inputs: [input, baseURL],
        protocol: match(matched._protocol, url.protocol),
        username: match(matched._username, url.username),
        password: match(matched._password, url.password),
        hostname: match(matched._hostname, url.hostname),
        port: match(matched._port, url.port),
        pathname: match(matched._pathname, url.pathname),
        search: match(matched._search, url.search),
        hash: match(matched._hash, url.hash)
      };
      for (var key in result) {
        if (key !== 'inputs' && result[key] === null) return null;
      }
      return result;
    }
  }
  Object.defineProperty(URLPattern.prototype, Symbol.toStringTag, { value: 'URLPattern', configurable: true });

  return {
    URL: URL,
    URLSearchParams: URLSearchParams,
    URLPattern: URLPattern,
    makeURL: function (href) { return new URL(href); },
    makeParams: function () { return new URLSearchParams(); },
  };
})(__godeUrlHost)`

// fileURLWrap decorates fileURLToPath/fileURLToPathBuffer with the `input`
// property Node sets on ERR_INVALID_FILE_URL_PATH.
const fileURLWrap = `(function (exports) {
  'use strict';
  function wrap(name) {
    var orig = exports[name];
    if (typeof orig !== 'function') return;
    exports[name] = function wrapped(value, options) {
      try {
        return orig(value, options);
      } catch (e) {
        if (e && e.code === 'ERR_INVALID_FILE_URL_PATH' &&
            value !== undefined && value !== null) {
          try { e.input = value instanceof URL ? value : new URL(String(value)); } catch (_) {}
        }
        throw e;
      }
    };
  }
  wrap('fileURLToPath');
  wrap('fileURLToPathBuffer');
})`

type module struct {
	e          *env.Env
	urlCtor    js.Object
	paramsCtor js.Object
	makeURL    js.Object
	makeParams js.Object
	// legacyProto is the prototype shared by url.parse() results and
	// `new url.Url()` instances. Node keeps toJSON there rather than on each
	// object, which keeps the own-key sets identical.
	legacyProto js.Object
}

// Install builds the url module's exports object and installs the URL globals.
func Install(e *env.Env) js.Object {
	m := &module{e: e}
	host := e.JS.Obj()
	m.wire(host)

	g := e.Global()
	g.Set("__godeUrlHost", host.V())
	res, err := e.JS.Eval(urlPrelude, "url.js")
	g.Delete("__godeUrlHost")
	if err != nil {
		panic(err)
	}
	obj := res.Object()
	m.urlCtor = obj.Get("URL").Object().Keep()
	m.paramsCtor = obj.Get("URLSearchParams").Object().Keep()
	m.makeURL = obj.Get("makeURL").Object().Keep()
	m.makeParams = obj.Get("makeParams").Object().Keep()

	exports := e.JS.Obj()
	exports.SetObj("URL", m.urlCtor)
	exports.SetObj("URLSearchParams", m.paramsCtor)

	// The legacy Url constructor initializes every field to null, matching
	// lib/url.js; assignments made by callers then replace them.
	m.legacyProto = e.JS.Obj().Keep()
	m.legacyProto.SetFunc("toJSON", func(this js.Object, _ []js.Value) (js.Value, error) {
		return this.V(), nil
	})
	if urlCtorVal, uerr := e.JS.Eval(`(function () {
		function Url() {
			if (!(this instanceof Url)) return new Url();
			this.protocol = null;
			this.slashes = null;
			this.auth = null;
			this.host = null;
			this.port = null;
			this.hostname = null;
			this.hash = null;
			this.search = null;
			this.query = null;
			this.pathname = null;
			this.path = null;
			this.href = null;
		}
		return Url;
	})()`, "[gode:url legacy]"); uerr == nil && urlCtorVal.IsFunction() {
		ctor := urlCtorVal.Object()
		if proto := ctor.Get("prototype"); proto.IsObject() {
			proto.Object().SetProto(m.legacyProto)
			m.legacyProto.Set("constructor", urlCtorVal)
		}
		exports.Set("Url", urlCtorVal)
	}
	m.wireExports(exports)
	// Node attaches the offending URL as `err.input` on
	// ERR_INVALID_FILE_URL_PATH; do it from JS so the property survives the
	// host-function error bridge.
	if wrapFn, werr := e.JS.Eval(fileURLWrap, "url.js"); werr == nil {
		if _, cerr := wrapFn.Object().Call(js.Object{}, exports.V()); cerr != nil {
			panic(cerr)
		}
	}
	if pattern := obj.Get("URLPattern"); pattern.Valid() {
		exports.Set("URLPattern", pattern)
		g.Set("URLPattern", pattern)
	}

	g.SetObj("URL", m.urlCtor)
	g.SetObj("URLSearchParams", m.paramsCtor)
	return exports
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"url.URL",
		"url.URL#canParse",
		"url.URL#createObjectURL",
		"url.URL#parse",
		"url.URL#revokeObjectURL",
		"url.URL#toJSON",
		"url.URL#toString",
		"url.URL.hash",
		"url.URL.host",
		"url.URL.hostname",
		"url.URL.href",
		"url.URL.origin",
		"url.URL.password",
		"url.URL.pathname",
		"url.URL.port",
		"url.URL.protocol",
		"url.URL.search",
		"url.URL.searchParams",
		"url.URL.username",
		"url.URLSearchParams",
		"url.URLSearchParams#[Symbol.iterator]",
		"url.URLSearchParams#append",
		"url.URLSearchParams#delete",
		"url.URLSearchParams#entries",
		"url.URLSearchParams#forEach",
		"url.URLSearchParams#get",
		"url.URLSearchParams#getAll",
		"url.URLSearchParams#has",
		"url.URLSearchParams#keys",
		"url.URLSearchParams#set",
		"url.URLSearchParams#sort",
		"url.URLSearchParams#toString",
		"url.URLSearchParams#values",
		"url.URLSearchParams.size",
		"url.URLPattern",
		"url.URLPattern#exec",
		"url.URLPattern#test",
		// The deprecated accessors Node documents at the module level.
		"url.auth",
		"url.hash",
		"url.host",
		"url.hostname",
		"url.href",
		"url.path",
		"url.pathname",
		"url.port",
		"url.protocol",
		"url.query",
		"url.search",
		"url.slashes",
		"url.domainToASCII",
		"url.domainToUnicode",
		"url.fileURLToPath",
		"url.fileURLToPathBuffer",
		"url.format",
		"url.parse",
		"url.pathToFileURL",
		"url.resolve",
		"url.resolveObject",
		"url.urlToHttpOptions",
	}
}

func (m *module) wire(h js.Object) {
	h.SetFunc("urlInit", m.urlInit)
	h.SetFunc("urlGet", m.urlGet)
	h.SetFunc("urlSet", m.urlSet)
	h.SetFunc("urlParams", m.urlParams)
	h.SetFunc("urlCanParse", func(_ js.Object, args []js.Value) (js.Value, error) {
		_, err := m.parseArgs(args)
		return m.e.JS.Bool(err == nil), nil
	})

	h.SetFunc("paramsInit", m.paramsInit)
	h.SetFunc("paramsSize", func(this js.Object, _ []js.Value) (js.Value, error) {
		return m.e.JS.Num(float64(len(readPairs(this)))), nil
	})
	h.SetFunc("paramsEntries", func(this js.Object, _ []js.Value) (js.Value, error) {
		return pairsArray(this), nil
	})
	h.SetFunc("paramsAppend", func(this js.Object, args []js.Value) (js.Value, error) {
		name, value := argStr(args, 0), argStr(args, 1)
		pairs := readPairs(this)
		pairs = append(pairs, [2]string{name, value})
		writePairs(this, pairs)
		paramsMutated(this)
		return js.Value{}, nil
	})
	h.SetFunc("paramsDelete", func(this js.Object, args []js.Value) (js.Value, error) {
		name := argStr(args, 0)
		pairs := readPairs(this)
		out := pairs[:0]
		if len(args) > 1 && !args[1].IsUndefined() {
			value := args[1].Str()
			for _, p := range pairs {
				if p[0] != name || p[1] != value {
					out = append(out, p)
				}
			}
		} else {
			for _, p := range pairs {
				if p[0] != name {
					out = append(out, p)
				}
			}
		}
		writePairs(this, out)
		paramsMutated(this)
		return js.Value{}, nil
	})
	h.SetFunc("paramsGet", func(this js.Object, args []js.Value) (js.Value, error) {
		name := argStr(args, 0)
		for _, p := range readPairs(this) {
			if p[0] == name {
				return m.e.JS.Str(p[1]), nil
			}
		}
		return m.e.JS.Null(), nil
	})
	h.SetFunc("paramsGetAll", func(this js.Object, args []js.Value) (js.Value, error) {
		name := argStr(args, 0)
		var vals []js.Value
		for _, p := range readPairs(this) {
			if p[0] == name {
				vals = append(vals, m.e.JS.Str(p[1]))
			}
		}
		return m.e.JS.Arr(vals).V(), nil
	})
	h.SetFunc("paramsHas", func(this js.Object, args []js.Value) (js.Value, error) {
		name := argStr(args, 0)
		valueGiven := len(args) > 1 && !args[1].IsUndefined()
		value := ""
		if valueGiven {
			value = args[1].Str()
		}
		for _, p := range readPairs(this) {
			if p[0] == name && (!valueGiven || p[1] == value) {
				return m.e.JS.Bool(true), nil
			}
		}
		return m.e.JS.Bool(false), nil
	})
	h.SetFunc("paramsSet", func(this js.Object, args []js.Value) (js.Value, error) {
		name, value := argStr(args, 0), argStr(args, 1)
		pairs := readPairs(this)
		out := make([][2]string, 0, len(pairs)+1)
		found := false
		for _, p := range pairs {
			if p[0] == name {
				if !found {
					out = append(out, [2]string{name, value})
					found = true
				}
				continue
			}
			out = append(out, p)
		}
		if !found {
			out = append(out, [2]string{name, value})
		}
		writePairs(this, out)
		paramsMutated(this)
		return js.Value{}, nil
	})
	h.SetFunc("paramsSort", func(this js.Object, _ []js.Value) (js.Value, error) {
		pairs := readPairs(this)
		stableSortPairs(pairs)
		writePairs(this, pairs)
		paramsMutated(this)
		return js.Value{}, nil
	})
	h.SetFunc("paramsToString", func(this js.Object, _ []js.Value) (js.Value, error) {
		return m.e.JS.Str(serializeForm(readPairs(this))), nil
	})
}

func (m *module) wireExports(exports js.Object) {
	ctx := m.e.JS

	exports.SetFunc("parse", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.legacyParse(args)
	})
	exports.SetFunc("format", func(_ js.Object, args []js.Value) (js.Value, error) {
		v := arg(args, 0)
		// Node accepts a string URL, which it parses first, and rejects every
		// other non-object with ERR_INVALID_ARG_TYPE.
		if v.Valid() && v.IsString() {
			obj, perr := m.legacyParse([]js.Value{v})
			if perr != nil {
				return js.Value{}, perr
			}
			if obj.Valid() && obj.IsObject() {
				return ctx.Str(legacyFormat(obj.Object())), nil
			}
			return ctx.Str(""), nil
		}
		if !v.Valid() || v.IsUndefined() || v.IsNull() || !v.IsObject() || v.IsFunction() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "urlObject" argument must be of type object or string. Received `+describe(v))
		}
		obj := v.Object()
		// A WHATWG URL carries the hidden $protocol slot; legacy Url objects do
		// not. Only URL instances honour the format() options object.
		if obj.Get("$protocol").IsString() {
			out, ferr := m.formatURLObject(obj, arg(args, 1))
			if ferr != nil {
				return js.Value{}, ferr
			}
			return ctx.Str(out), nil
		}
		return ctx.Str(legacyFormat(obj)), nil
	})
	exports.SetFunc("resolve", func(_ js.Object, args []js.Value) (js.Value, error) {
		from := argStr(args, 0)
		to := argStr(args, 1)
		if strings.HasSuffix(from, ":") && strings.HasPrefix(to, from+"//") {
			return ctx.Str(to), nil
		}
		if strings.HasSuffix(to, ":") && strings.HasPrefix(from, to+"//") {
			return ctx.Str(from), nil
		}
		if strings.HasPrefix(from, "//") {
			host := strings.TrimPrefix(from, "//")
			if !strings.ContainsAny(host, "/?#") {
				if i := strings.Index(to, "://"); i > 0 {
					return ctx.Str(to[:i+3] + host), nil
				}
			}
		}
		if strings.HasPrefix(to, "//") {
			host := strings.TrimPrefix(to, "//")
			if !strings.ContainsAny(host, "/?#") {
				if i := strings.Index(from, "://"); i > 0 {
					res := from[:i+3] + host
					if isSpecial(from[:i+1]) {
						res += "/"
					}
					return ctx.Str(res), nil
				}
			}
		}
		if i := strings.IndexByte(to, ':'); i > 0 && !strings.HasPrefix(to[i:], "://") && strings.HasPrefix(from, to[:i]+"://") {
			if base, err := parseURL(from, nil); err == nil {
				rest := to[i+1:]
				if rest != "" && !strings.HasPrefix(rest, "/") && rest[0] != '#' && rest[0] != '?' {
					p := base.pathname
					if j := strings.LastIndex(p, "/"); j >= 0 {
						p = p[:j+1]
					} else {
						p = "/"
					}
					return ctx.Str(to[:i+1] + "//" + base.hostport() + p + rest), nil
				}
			}
		}
		if i := strings.IndexByte(from, ':'); i > 0 && !strings.HasPrefix(from[i:], "://") && strings.HasPrefix(to, from[:i]+"://") {
			if base, err := parseURL(to, nil); err == nil {
				rest := from[i+1:]
				if rest != "" && !strings.HasPrefix(rest, "/") {
					p := base.pathname
					if j := strings.LastIndex(p, "/"); j >= 0 {
						p = p[:j+1]
					} else {
						p = "/"
					}
					return ctx.Str(from[:i+1] + "//" + base.hostport() + p + rest), nil
				}
			}
		}
		if strings.HasPrefix(to, "/") {
			if i := strings.IndexByte(from, ':'); i > 0 && (i+1 >= len(from) || from[i+1] != '/') {
				p := normalizePath(to)
				for strings.HasPrefix(p, "../") {
					p = strings.TrimPrefix(p, "../")
				}
				if !strings.HasPrefix(p, "/") {
					p = "/" + p
				}
				return ctx.Str(from[:i+1] + p), nil
			}
		}
		if strings.HasPrefix(from, "/") {
			if i := strings.IndexByte(to, ':'); i > 0 && i+1 < len(to) && to[i+1] != '/' {
				p := normalizePath(from)
				for strings.HasPrefix(p, "../") {
					p = strings.TrimPrefix(p, "../")
				}
				if !strings.HasPrefix(p, "/") {
					p = "/" + p
				}
				return ctx.Str(to[:i+1] + p), nil
			}
		}
		if scheme := strings.TrimSuffix(from, ":."); scheme != from && strings.HasPrefix(to, scheme+":") {
			return ctx.Str(scheme + ":"), nil
		}
		if scheme := strings.TrimSuffix(to, ":."); scheme != to && strings.HasPrefix(from, scheme+":") {
			return ctx.Str(scheme + ":"), nil
		}
		if i := strings.Index(from, ":"); i > 0 && !strings.Contains(from[:i], "/") && !strings.HasPrefix(from[i:], "://") {
			scheme, rest := from[:i], from[i+1:]
			if rest == "." && strings.HasPrefix(to, scheme+":") {
				return ctx.Str(scheme + ":"), nil
			}
		}
		base, err := parseURL(from, nil)
		if err == nil {
			// A different scheme with `:#fragment` or a single `/` denotes an
			// authority: https:#frag -> https:///#frag, https:/p -> https://p.
			if i := strings.Index(to, ":#"); i > 0 && !strings.ContainsAny(to[:i], "/:") && to[:i]+":" != base.protocol {
				return ctx.Str(to[:i] + ":///" + to[i+1:]), nil
			}
			if i := strings.Index(to, ":/"); i > 0 && !strings.HasPrefix(to[i:], "://") && !strings.ContainsAny(to[:i], "/:") && to[:i]+":" != base.protocol {
				return ctx.Str(to[:i] + "://" + to[i+2:]), nil
			}
			if i := strings.Index(to, ":/"); i > 0 && !strings.HasPrefix(to[i:], "://") && !strings.ContainsAny(to[:i], "/:") && to[:i]+":" == base.protocol {
				// Same scheme with one slash: keep the base authority.
				return ctx.Str(base.protocol + "//" + base.hostport() + to[i+1:]), nil
			}
			if i := strings.Index(to, ":#"); i > 0 && !strings.ContainsAny(to[:i], "/:") && to[:i]+":" == base.protocol {
				// Same-scheme #fragment keeps the base href, including its exact
				// path/search spelling that the URL parser would normalize.
				baseHref := from
				if j := strings.IndexByte(baseHref, '#'); j >= 0 {
					baseHref = baseHref[:j]
				}
				return ctx.Str(baseHref + to[i+1:]), nil
			}
			p, err := parseURL(to, base)
			if err == nil {
				return ctx.Str(p.href()), nil
			}
		}
		// Node's legacy resolver also accepts relative bases; join them as a
		// POSIX path when the WHATWG parser cannot help. An absolute target
		// replaces the base path entirely.
		if i := strings.Index(from, ":"); i > 0 && !strings.Contains(from[:i], "/") && !strings.HasPrefix(from[i:], "://") {
			// Opaque base (foo:a/b): resolve against its path and keep the scheme.
			scheme, rest := from[:i], from[i+1:]
			if rest == "." && strings.HasPrefix(to, scheme+":") {
				return ctx.Str(scheme + ":"), nil
			}
			dir := rest
			if j := strings.LastIndex(dir, "/"); j >= 0 {
				dir = dir[:j+1]
			} else {
				dir = ""
			}
			resolved := normalizeRelativePath(dir + to)
			for strings.HasPrefix(resolved, "../") {
				resolved = strings.TrimPrefix(resolved, "../")
			}
			return ctx.Str(scheme + ":" + resolved), nil
		}
		if strings.HasPrefix(from, "/") {
			if strings.HasPrefix(to, "/") {
				return ctx.Str(normalizePath(to)), nil
			}
			dir := from
			if i := strings.LastIndex(dir, "/"); i >= 0 {
				dir = dir[:i+1]
			}
			return ctx.Str(normalizePath(dir + to)), nil
		}
		// Relative base: apply the legacy path rules, which keep leading
		// ".." segments instead of clamping them the way URL paths do.
		dir := from
		if i := strings.LastIndex(dir, "/"); i >= 0 {
			dir = dir[:i+1]
		} else {
			dir = ""
		}
		return ctx.Str(normalizeRelativePath(dir + to)), nil
	})
	// url.resolveObject(source, relative) returns the *relative* argument
	// untouched when the source is falsy, and a parsed Url object otherwise.
	exports.SetFunc("resolveObject", func(_ js.Object, args []js.Value) (js.Value, error) {
		source := argStr(args, 0)
		if source == "" {
			if len(args) > 1 {
				return args[1], nil
			}
			return ctx.Undefined(), nil
		}
		var href string
		base, err := parseURL(source, nil)
		if err == nil {
			if p, err := parseURL(argStr(args, 1), base); err == nil {
				href = p.href()
			}
		}
		if href == "" {
			href = argStr(args, 1)
		}
		obj, parseErr := m.legacyParse([]js.Value{ctx.Str(href), ctx.Bool(false)})
		if parseErr == nil && obj.Valid() && obj.IsObject() {
			return obj, nil
		}
		return ctx.Str(href), nil
	})
	exports.SetFunc("fileURLToPath", func(_ js.Object, args []js.Value) (js.Value, error) {
		p, err := m.filePathFromURL(argURL(args), arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		return ctx.Str(p), nil
	})
	exports.SetFunc("fileURLToPathBuffer", func(_ js.Object, args []js.Value) (js.Value, error) {
		p, err := m.filePathFromURL(argURL(args), arg(args, 1))
		if err != nil {
			return js.Value{}, err
		}
		return m.bufferFrom([]byte(p)), nil
	})
	exports.SetFunc("pathToFileURL", func(_ js.Object, args []js.Value) (js.Value, error) {
		var path string
		if len(args) > 0 && args[0].IsString() {
			path = args[0].Str()
		} else if len(args) > 0 {
			if b, ok := args[0].Bytes(); ok {
				path = string(b)
			} else {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					"The \"path\" argument must be of type string or an instance of Buffer or URL. Received "+describe(args[0]))
			}
		} else {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				"The \"path\" argument must be of type string. Received undefined")
		}
		windows := false
		if len(args) > 1 && args[1].Valid() && args[1].IsObject() {
			if w := args[1].Object().Get("windows"); w.IsBool() {
				windows = w.Bool()
			}
		}
		var href string
		if windows {
			var err error
			href, err = windowsFileHref(path)
			if err != nil {
				return js.Value{}, err
			}
		} else {
			trailingSlash := strings.HasSuffix(path, "/")
			if !filepath.IsAbs(path) {
				path = filepath.Join(m.e.Cwd(), path)
			}
			abs, err := filepath.Abs(path)
			if err == nil {
				path = abs
			}
			slashPath := filepath.ToSlash(path)
			if trailingSlash && !strings.HasSuffix(slashPath, "/") {
				slashPath += "/"
			}
			href = "file://" + encodeFilePath(slashPath)
		}
		v, err := m.makeURL.Call(js.Object{}, ctx.Str(href))
		if err != nil {
			return js.Value{}, err
		}
		return v, nil
	})
	exports.SetFunc("domainToASCII", func(_ js.Object, args []js.Value) (js.Value, error) {
		return ctx.Str(idnaASCII(argStr(args, 0))), nil
	})
	exports.SetFunc("domainToUnicode", func(_ js.Object, args []js.Value) (js.Value, error) {
		return ctx.Str(idnaUnicode(argStr(args, 0))), nil
	})
	exports.SetFunc("urlToHttpOptions", func(_ js.Object, args []js.Value) (js.Value, error) {
		return m.urlToHTTPOptions(args)
	})
}

// --- URL object state ------------------------------------------------------

func setParts(o js.Object, p *urlParts) {
	c := o.Ctx()
	o.SetHidden("$protocol", c.Str(p.protocol))
	o.SetHidden("$username", c.Str(p.username))
	o.SetHidden("$password", c.Str(p.password))
	o.SetHidden("$host", c.Str(p.host))
	o.SetHidden("$port", c.Str(p.port))
	o.SetHidden("$pathname", c.Str(p.pathname))
	o.SetHidden("$search", c.Str(p.search))
	o.SetHidden("$hash", c.Str(p.hash))
	o.SetHidden("$hasSearch", c.Bool(p.hasSearch))
	o.SetHidden("$hasHash", c.Bool(p.hasHash))
	o.SetHidden("$auth", c.Bool(p.hasAuth))
	o.SetHidden("$opaque", c.Bool(p.opaque))
}

func getParts(o js.Object) *urlParts {
	return &urlParts{
		protocol:  o.Get("$protocol").Str(),
		username:  o.Get("$username").Str(),
		password:  o.Get("$password").Str(),
		host:      o.Get("$host").Str(),
		port:      o.Get("$port").Str(),
		pathname:  o.Get("$pathname").Str(),
		search:    o.Get("$search").Str(),
		hash:      o.Get("$hash").Str(),
		hasSearch: o.Get("$hasSearch").Bool(),
		hasHash:   o.Get("$hasHash").Bool(),
		hasAuth:   o.Get("$auth").Bool(),
		opaque:    o.Get("$opaque").Bool(),
	}
}

func (m *module) parseArgs(args []js.Value) (*urlParts, error) {
	if len(args) == 0 {
		return nil, urlError{}
	}
	input, ok := stringish(args[0])
	if !ok {
		return nil, urlError{}
	}
	var base *urlParts
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		if v, ok := stringish(args[1]); ok {
			p, err := parseURL(v, nil)
			if err != nil {
				return nil, err
			}
			base = p
		} else {
			return nil, urlError{}
		}
	}
	return parseURL(input, base)
}

func stringish(v js.Value) (string, bool) {
	if v.IsUndefined() || v.IsNull() {
		return "", false
	}
	if v.IsObject() && v.Object().Get("$protocol").IsString() {
		return getParts(v.Object()).href(), true
	}
	s, ok := v.StrOK()
	return s, ok
}

func (m *module) urlInit(this js.Object, args []js.Value) (js.Value, error) {
	p, err := m.parseArgs(args)
	if err != nil {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_URL", "Invalid URL")
	}
	setParts(this, p)
	return js.Value{}, nil
}

func (m *module) urlGet(this js.Object, args []js.Value) (js.Value, error) {
	field := ""
	if len(args) > 0 {
		field = args[0].Str()
	}
	p := getParts(this)
	c := m.e.JS
	switch field {
	case "href":
		return c.Str(p.href()), nil
	case "origin":
		return c.Str(p.origin()), nil
	case "protocol":
		return c.Str(p.protocol), nil
	case "username":
		return c.Str(p.username), nil
	case "password":
		return c.Str(p.password), nil
	case "host":
		return c.Str(p.hostport()), nil
	case "hostname":
		return c.Str(p.host), nil
	case "port":
		return c.Str(p.port), nil
	case "pathname":
		return c.Str(p.pathname), nil
	case "search":
		if p.search == "" {
			return c.Str(""), nil
		}
		return c.Str("?" + p.search), nil
	case "hash":
		if p.hash == "" {
			return c.Str(""), nil
		}
		return c.Str("#" + p.hash), nil
	}
	return c.Undefined(), nil
}

func (m *module) urlSet(this js.Object, args []js.Value) (js.Value, error) {
	field := ""
	if len(args) > 0 {
		field = args[0].Str()
	}
	value := ""
	if len(args) > 1 {
		if v, ok := args[1].StrOK(); ok {
			value = v
		} else {
			return js.Value{}, nil
		}
	}
	p := getParts(this)
	switch field {
	case "href":
		next, err := parseURL(value, nil)
		if err != nil {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_URL", "Invalid URL")
		}
		p = next
	case "protocol":
		scheme := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ":")
		if scheme == "" || !validScheme(scheme) {
			return js.Value{}, nil
		}
		p.protocol = scheme + ":"
	case "username":
		p.username = encodeUserinfo(value)
	case "password":
		p.password = encodeUserinfo(value)
	case "host", "hostname":
		hostValue := value
		port := p.port
		if field == "host" {
			if v, ok := splitHostPort(value); ok {
				hostValue, port = v[0], v[1]
			}
		}
		hostValue = idnaASCII(strings.ToLower(hostValue))
		if strings.ContainsAny(hostValue, " \t\r\n/?#@") {
			return js.Value{}, nil
		}
		if field == "host" {
			if n, err := strconv.Atoi(port); err == nil && n >= 0 && n <= 65535 {
				if strconv.Itoa(n) == defaultPorts[p.protocol] {
					port = ""
				} else {
					port = strconv.Itoa(n)
				}
			} else {
				port = ""
			}
		}
		p.host = hostValue
		p.port = port
	case "port":
		if value == "" {
			p.port = ""
		} else if n, err := strconv.Atoi(value); err == nil && n >= 0 && n <= 65535 {
			port := strconv.Itoa(n)
			if port == defaultPorts[p.protocol] {
				port = ""
			}
			p.port = port
		}
	case "pathname":
		path := encodePath(value, isSpecial(p.protocol))
		path = ensureLeadingSlash(normalizePath(path))
		p.pathname = path
	case "search":
		raw := strings.TrimPrefix(value, "?")
		p.search = encodeQuery(raw, isSpecial(p.protocol))
		p.hasSearch = p.search != ""
	case "hash":
		raw := strings.TrimPrefix(value, "#")
		p.hash = encodeFragment(raw)
		p.hasHash = p.hash != ""
	}
	setParts(this, p)
	if pp := this.Get("$params"); pp.IsObject() {
		if field == "href" || field == "search" {
			writePairs(pp.Object(), parseForm(p.search))
		}
	}
	return js.Value{}, nil
}

func validScheme(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if i == 0 {
			if !isAlpha(c) {
				return false
			}
			continue
		}
		if !isAlpha(c) && !isDigit(c) && c != '+' && c != '-' && c != '.' {
			return false
		}
	}
	return s != ""
}

func splitHostPort(v string) ([2]string, bool) {
	if strings.HasPrefix(v, "[") {
		if i := strings.IndexByte(v, ']'); i >= 0 {
			host := v[:i+1]
			if i+1 < len(v) && v[i+1] == ':' {
				return [2]string{host, v[i+2:]}, true
			}
			return [2]string{host, ""}, true
		}
	}
	if i := strings.LastIndexByte(v, ':'); i >= 0 {
		return [2]string{v[:i], v[i+1:]}, true
	}
	return [2]string{v, ""}, true
}

// --- URL.searchParams ------------------------------------------------------

func (m *module) urlParams(this js.Object, _ []js.Value) (js.Value, error) {
	if p := this.Get("$params"); p.IsObject() {
		return p, nil
	}
	inst, err := m.makeParams.Call(js.Object{})
	if err != nil {
		return js.Value{}, err
	}
	params := inst.Object()
	params.SetHidden("$url", this.V())
	writePairs(params, parseForm(this.Get("$search").Str()))
	this.SetHidden("$params", params.V())
	return inst, nil
}

func paramsMutated(o js.Object) {
	urlv := o.Get("$url")
	if !urlv.IsObject() {
		return
	}
	url := urlv.Object()
	raw := serializeForm(readPairs(o))
	c := url.Ctx()
	url.SetHidden("$search", c.Str(raw))
	url.SetHidden("$hasSearch", c.Bool(raw != ""))
}

func readPairs(o js.Object) [][2]string {
	arr := o.Get("$pairs")
	if !arr.IsArray() {
		return nil
	}
	a := arr.Object()
	n := a.Length()
	out := make([][2]string, 0, n)
	for i := 0; i < n; i++ {
		pair := a.At(i)
		if !pair.IsArray() {
			continue
		}
		p := pair.Object()
		name, _ := p.At(0).StrOK()
		value, _ := p.At(1).StrOK()
		out = append(out, [2]string{name, value})
	}
	return out
}

func writePairs(o js.Object, pairs [][2]string) {
	c := o.Ctx()
	arr := c.Arr(nil)
	for i, p := range pairs {
		pair := c.Arr(nil)
		pair.SetAt(0, c.Str(p[0]))
		pair.SetAt(1, c.Str(p[1]))
		arr.SetAt(i, pair.V())
	}
	o.SetHidden("$pairs", arr.V())
}

func pairsArray(o js.Object) js.Value {
	arr := o.Get("$pairs")
	if arr.IsArray() {
		return arr
	}
	return o.Ctx().Arr(nil).V()
}

func stableSortPairs(pairs [][2]string) {
	// Insertion sort is stable and URLSearchParams lists are small.
	for i := 1; i < len(pairs); i++ {
		for j := i; j > 0 && compareUTF16(pairs[j][0], pairs[j-1][0]) < 0; j-- {
			pairs[j], pairs[j-1] = pairs[j-1], pairs[j]
		}
	}
}

func (m *module) paramsInit(this js.Object, args []js.Value) (js.Value, error) {
	var init js.Value
	if len(args) > 0 {
		init = args[0]
	}
	if !init.Valid() || init.IsUndefined() || init.IsNull() {
		writePairs(this, nil)
		return js.Value{}, nil
	}
	if init.IsString() {
		raw := strings.TrimPrefix(init.Str(), "?")
		writePairs(this, parseForm(raw))
		return js.Value{}, nil
	}
	if init.IsArray() {
		pairs := init.Object()
		out := make([][2]string, 0, pairs.Length())
		for i := 0; i < pairs.Length(); i++ {
			pair := pairs.At(i)
			if !pair.IsArray() || pair.Object().Length() < 2 {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_TUPLE",
					"Each query pair must be an iterable [name, value] tuple")
			}
			p := pair.Object()
			out = append(out, [2]string{p.At(0).Str(), p.At(1).Str()})
		}
		writePairs(this, out)
		return js.Value{}, nil
	}
	if !init.IsObject() {
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_TUPLE",
			"Each query pair must be an iterable [name, value] tuple")
	}
	src := init.Object()
	if src.Get("$pairs").IsArray() {
		writePairs(this, readPairs(src))
		return js.Value{}, nil
	}
	out := make([][2]string, 0)
	for _, key := range src.OwnEnumerableKeys() {
		v := src.Get(key)
		value := ""
		if v.IsArray() {
			parts := make([]string, 0, v.Object().Length())
			for i := 0; i < v.Object().Length(); i++ {
				parts = append(parts, v.Object().At(i).Str())
			}
			value = strings.Join(parts, ",")
		} else {
			value = v.Str()
		}
		out = append(out, [2]string{key, value})
	}
	writePairs(this, out)
	return js.Value{}, nil
}

// --- file URLs -------------------------------------------------------------

func argURL(args []js.Value) js.Value {
	if len(args) == 0 {
		return js.Value{}
	}
	return args[0]
}

// formatURLObject implements Node's url.format(url[, options]) for WHATWG URL
// instances: options.auth/fragment/search/unicode default to true/true/true/
// false and are only read when truthy, exactly like the reference code.
func (m *module) formatURLObject(o js.Object, options js.Value) (string, error) {
	p := getParts(o)
	auth, fragment, search, unicode := true, true, true, false
	if options.Valid() && options.Bool() {
		if !options.IsObject() || options.IsFunction() {
			return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "options" argument must be of type object. Received `+describe(options))
		}
		opts := options.Object()
		read := func(key string, dst *bool) {
			v := opts.Get(key)
			if v.Valid() && !v.IsUndefined() && !v.IsNull() {
				*dst = v.Bool()
			}
		}
		read("auth", &auth)
		read("fragment", &fragment)
		read("search", &search)
		read("unicode", &unicode)
	}
	if !auth {
		p.username, p.password = "", ""
	}
	if !search {
		p.search, p.hasSearch = "", false
	}
	if !fragment {
		p.hash, p.hasHash = "", false
	}
	if unicode && p.host != "" {
		if u := idnaUnicode(p.host); u != "" {
			p.host = u
		}
	}
	return p.href(), nil
}

func (m *module) filePathFromURL(v js.Value, options js.Value) (string, error) {
	windows := runtime.GOOS == "windows"
	if options.Valid() && options.IsObject() {
		if w := options.Object().Get("windows"); w.IsBool() {
			windows = w.Bool()
		}
	}
	if !v.Valid() || v.IsFunction() || (!v.IsObject() && !v.IsString()) {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			"The \"url\" argument must be of type string or an instance of URL. Received "+describe(v))
	}
	var protocol, pathname, host string
	if v.IsString() {
		p, err := parseURL(v.Str(), nil)
		if err != nil {
			return "", js.ThrowCode("TypeError", "ERR_INVALID_URL", "Invalid URL")
		}
		protocol, pathname, host = p.protocol, p.pathname, p.host
	} else {
		obj := v.Object()
		if pv := obj.Get("$protocol"); pv.IsString() {
			protocol = pv.Str()
			pathname = obj.Get("$pathname").Str()
			host = obj.Get("$host").Str()
		} else {
			// A URL from another realm, or any object exposing the getters.
			pv = obj.Get("protocol")
			pathname = obj.Get("pathname").Str()
			host = obj.Get("hostname").Str()
			if !pv.IsString() || pv.Str() == "" {
				return "", js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					"The \"url\" argument must be of type string or an instance of URL. Received "+describe(v))
			}
			protocol = pv.Str()
		}
	}
	if protocol != "file:" {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_URL_SCHEME", "The URL must be of scheme file")
	}
	if host != "" && host != "localhost" {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_FILE_URL_HOST",
			`The URL host must be "localhost" or empty on linux`)
	}
	lower := strings.ToLower(pathname)
	// Node rejects encoded separators: always %2F, and %5C only where the
	// backslash is a path separator (Windows).
	if strings.Contains(lower, "%2f") || (windows && strings.Contains(lower, "%5c")) {
		return "", js.ThrowCode("TypeError", "ERR_INVALID_FILE_URL_PATH",
			"File URL path must not include encoded / characters")
	}
	return decodePercent(pathname), nil
}

func (m *module) bufferFrom(b []byte) js.Value {
	if env.Registered("buffer") {
		buf := m.e.Exports("buffer")
		fn := buf.Get("Buffer")
		if fn.IsFunction() {
			from := fn.Object().Get("from")
			if from.IsFunction() {
				if v, err := from.Object().Call(fn.Object(), m.e.JS.Bytes(b).V()); err == nil {
					return v
				}
			}
		}
	}
	return m.e.JS.Bytes(b).V()
}

// --- urlToHttpOptions ------------------------------------------------------

// urlToHTTPOptions mirrors Node's internal implementation: property values are
// copied through unchanged (so a spread copy of a URL yields undefined), the
// IPv6 bracket pair is stripped from hostname, and auth is only present when a
// username or password was set.
func (m *module) urlToHTTPOptions(args []js.Value) (js.Value, error) {
	if len(args) == 0 || !args[0].IsObject() || args[0].IsNull() {
		received := "undefined"
		if len(args) > 0 {
			received = urlArgType(args[0])
		}
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			"The \"url\" argument must be of type object. Received "+received)
	}
	u := args[0].Object()
	c := m.e.JS
	opts := c.Obj()
	for _, key := range u.OwnEnumerableKeys() {
		opts.Set(key, u.Get(key))
	}
	hostname := u.Get("hostname")
	if hostname.IsString() && strings.HasPrefix(hostname.Str(), "[") && strings.HasSuffix(hostname.Str(), "]") {
		hostname = c.Str(hostname.Str()[1 : len(hostname.Str())-1])
	}
	pathname, search := u.Get("pathname"), u.Get("search")
	path := ""
	if pathname.IsString() {
		path = pathname.Str()
	}
	if search.IsString() {
		path += search.Str()
	}
	opts.Set("protocol", u.Get("protocol"))
	opts.Set("hostname", hostname)
	opts.Set("hash", u.Get("hash"))
	opts.Set("search", search)
	opts.Set("pathname", pathname)
	opts.SetStr("path", path)
	opts.Set("href", u.Get("href"))
	if port := u.Get("port"); !(port.IsString() && port.Str() == "") {
		n := math.NaN()
		if port.IsString() {
			if f, err := strconv.ParseFloat(port.Str(), 64); err == nil {
				n = f
			}
		}
		opts.Set("port", c.Num(n))
	}
	username, password := u.Get("username"), u.Get("password")
	if (username.IsString() && username.Str() != "") || (password.IsString() && password.Str() != "") {
		opts.SetStr("auth", decodeURIComponentPart(c, usernameStr(username))+":"+decodeURIComponentPart(c, usernameStr(password)))
	}
	return opts.V(), nil
}

func usernameStr(v js.Value) string {
	if v.IsString() {
		return v.Str()
	}
	if v.Valid() && !v.IsUndefined() && !v.IsNull() {
		return v.Str()
	}
	return ""
}

func decodeURIComponentPart(c *js.Context, s string) string {
	fn := c.Global().Get("decodeURIComponent")
	if !fn.IsFunction() {
		return s
	}
	if v, err := fn.Object().Call(js.Object{}, c.Str(s)); err == nil {
		return v.Str()
	}
	return s
}

func urlArgType(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return "undefined"
	}
	if v.IsString() {
		return "type string ('" + v.Str() + "')"
	}
	return describe(v)
}

// windowsFileHref emulates pathToFileURL(..., { windows: true }) well enough
// for UNC, extended-length and drive-letter paths.
func windowsFileHref(path string) (string, error) {
	// Extended-length prefixes are dropped before anything else, like Node's
	// pathToFileURL does.
	raw := path
	if strings.HasPrefix(raw, `\\?\UNC\`) {
		raw = `\\` + raw[len(`\\?\UNC\`):]
	} else if strings.HasPrefix(raw, `\\?\`) {
		raw = raw[len(`\\?\`):]
	}
	missing := func() error {
		return js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
			"The argument 'path' must be a string, Buffer, or URL without null bytes. Received "+strconv.Quote(path))
	}
	if strings.HasPrefix(raw, `\\`) {
		// UNC path: the hostname runs to the next backslash. Node passes the
		// hostname to the URL parser, so a '/', '#' or '?' inside it truncates
		// the host and discards the rest of that segment.
		rest := raw[2:]
		end := strings.IndexByte(rest, '\\')
		if end < 0 {
			return "", missing()
		}
		host := rest[:end]
		host = strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == '\t' {
				return -1
			}
			return r
		}, host)
		if i := strings.IndexAny(host, "/?#"); i >= 0 {
			host = host[:i]
		}
		if host == "" {
			return "", missing()
		}
		for _, bad := range []string{" ", "@", ":", "[", "]"} {
			if strings.Contains(host, bad) {
				return "", js.ThrowCode("TypeError", "ERR_INVALID_URL", "Invalid URL")
			}
		}
		tail := strings.ReplaceAll(rest[end:], `\`, "/")
		if tail == "/" || tail == "" {
			return "", missing()
		}
		return "file://" + host + encodeFilePath(tail), nil
	}
	p := strings.ReplaceAll(raw, `\`, "/")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "file://" + encodeFilePath(p), nil
}

// --- legacy API ------------------------------------------------------------

func (m *module) legacyParse(args []js.Value) (js.Value, error) {
	if len(args) == 0 || !args[0].IsString() {
		var v js.Value
		if len(args) > 0 {
			v = args[0]
		}
		return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
			`The "url" argument must be of type string.`+m.invalidArgReceived(v))
	}
	original := argStr(args, 0)
	if strings.IndexByte(original, 0) >= 0 {
		return js.Value{}, &js.Error{Name: "TypeError", Code: "ERR_INVALID_URL", Message: "Invalid URL", Input: original}
	}
	raw := strings.TrimSpace(original)
	parseQuery := len(args) > 1 && args[1].Bool()
	slashesDenoteHost := len(args) > 2 && args[2].Bool()
	if strings.HasPrefix(raw, "//") {
		rest := raw[2:]
		authPart := rest
		if k := strings.IndexAny(rest, "/?#"); k >= 0 {
			authPart = rest[:k]
		}
		if strings.ContainsAny(authPart, "@:.") {
			if p, err := parseURL("http:"+raw, nil); err == nil {
				p.protocol = ""
				obj := m.legacyObject(p, parseQuery)
				obj.SetStr("href", legacyFormat(obj))
				return obj.V(), nil
			}
		}
	}
	if i := strings.Index(raw, ":"); i > 0 && validScheme(raw[:i]) && !strings.HasPrefix(raw[i:], "//") {
		rest := raw[i+1:]
		if at := strings.IndexByte(rest, '@'); at > 0 && !strings.Contains(rest[:at], "/") {
			auth := rest[:at]
			after := rest[at+1:]
			host, tail := after, ""
			if j := strings.IndexAny(after, "/?#"); j >= 0 {
				host, tail = after[:j], after[j:]
			}
			obj := m.legacyRelative(tail, parseQuery)
			obj.SetStr("protocol", strings.ToLower(raw[:i])+":")
			obj.SetStr("auth", auth)
			obj.SetStr("host", host)
			obj.SetStr("hostname", host)
			pathname := tail
			if j := strings.IndexAny(pathname, "?#"); j >= 0 {
				pathname = pathname[:j]
			}
			if pathname == "" {
				obj.SetNull("pathname")
			} else {
				obj.SetStr("pathname", pathname)
			}
			if tail == "" {
				obj.SetNull("path")
			} else {
				obj.SetStr("path", tail)
			}
			obj.SetStr("href", raw)
			return obj.V(), nil
		}
	}
	if i := strings.Index(raw, ":"); i > 0 && i+1 < len(raw) && raw[i+1] != '/' && raw[i+1] != '\\' && validScheme(raw[:i]) {
		rest := raw[i+1:]
		if j := strings.IndexByte(rest, '/'); j > 0 {
			host, tail := rest[:j], rest[j:]
			obj := m.legacyRelative(tail, parseQuery)
			obj.SetStr("protocol", strings.ToLower(raw[:i])+":")
			obj.SetStr("host", host)
			obj.SetStr("hostname", host)
			obj.SetStr("href", raw)
			return obj.V(), nil
		}
		obj := m.legacyRelative(rest, parseQuery)
		obj.SetStr("protocol", strings.ToLower(raw[:i])+":")
		obj.SetStr("href", raw)
		return obj.V(), nil
	}
	if !strings.Contains(raw, ":") && !(slashesDenoteHost && strings.HasPrefix(raw, "//")) {
		return m.legacyRelative(raw, parseQuery).V(), nil
	}
	// Node's legacy parser keeps the fragment verbatim; the WHATWG parser
	// percent-encodes characters such as '#', so strip it before parsing and
	// reattach the original text afterwards.
	rawHash := ""
	if i := strings.Index(raw, "#"); i >= 0 {
		rawHash = raw[i:]
		raw = raw[:i]
	}
	if specialProtocol(raw) {
		// Special URLs treat backslashes as path separators, but only before
		// the query; backslashes inside a query are percent-encoded normally.
		head, tail := raw, ""
		if i := strings.IndexByte(raw, '?'); i >= 0 {
			head, tail = raw[:i], raw[i:]
		}
		raw = strings.ReplaceAll(head, "\\", "/") + tail
	}
	// Legacy parse decodes the userinfo and raises URIError for malformed
	// percent escapes there; stray '%' in paths is tolerated.
	if i := strings.Index(raw, "://"); i >= 0 {
		rest := raw[i+3:]
		if at := strings.IndexByte(rest, '@'); at >= 0 {
			if userinfo := rest[:at]; hasMalformedPercent(userinfo) {
				if _, derr := m.e.JS.Eval("decodeURIComponent("+strconv.Quote(userinfo)+")", "[gode:url]"); derr != nil {
					return js.Value{}, derr
				}
			}
		}
	}
	if i := strings.Index(raw, ":/"); i > 0 && !strings.HasPrefix(raw[i:], "://") && validScheme(raw[:i]) {
		rest := raw[i+1:]
		obj := m.legacyRelative(rest, parseQuery)
		if rawHash != "" {
			obj.SetStr("hash", "#"+encodeLegacyHash(strings.TrimPrefix(rawHash, "#")))
			obj.SetStr("href", raw+rawHash)
		} else {
			obj.SetStr("href", raw)
		}
		obj.SetStr("protocol", strings.ToLower(raw[:i])+":")
		return obj.V(), nil
	}
	if i := strings.Index(raw, "//"); i >= 0 {
		rest := raw[i+2:]
		if j := strings.IndexByte(rest, ';'); j >= 0 {
			if k := strings.IndexAny(rest, "/?#"); k < 0 || j < k {
				raw = raw[:i+2] + rest[:j] + "/" + rest[j:]
			}
		}
	}
	p, err := parseURL(raw, nil)
	if err != nil {
		if fixed, ok := encodeWonkyLegacy(raw); ok {
			if p2, e2 := parseURL(fixed, nil); e2 == nil {
				p, err = p2, nil
			}
		}
	}
	if err != nil {
		// The legacy parser is deliberately lenient: malformed strings still
		// produce a Url object rather than throwing (except for the explicit
		// NUL/percent/IDNA cases handled above).
		obj := m.legacyRelative(raw, parseQuery)
		if i := strings.Index(raw, ":"); i > 0 && validScheme(raw[:i]) {
			obj.SetStr("protocol", strings.ToLower(raw[:i])+":")
		}
		return obj.V(), nil
	}
	if strings.HasPrefix(strings.ToLower(raw), "file://") {
		rest := raw[7:]
		if i := strings.IndexAny(rest, "/?#"); i > 0 {
			p.host = strings.ToLower(rest[:i])
		}
	}
	if m.invalidIDNA(raw) {
		return js.Value{}, &js.Error{Name: "TypeError", Code: "ERR_INVALID_URL", Message: "Invalid URL", Input: raw}
	}
	if hostTooLong(p.host) {
		q := *p
		q.host, q.port, q.username, q.password = "", "", "", ""
		q.hasAuth = true
		p = &q
	}
	obj := m.legacyObject(p, parseQuery)
	if rawHash != "" {
		rawHash = "#" + encodeLegacyHash(strings.TrimPrefix(rawHash, "#"))
		obj.SetStr("hash", rawHash)
		obj.SetStr("href", legacyFormat(obj))
	}
	return obj.V(), nil
}

func (m *module) legacyObject(p *urlParts, parseQuery bool) js.Object {
	c := m.e.JS
	obj := c.Obj()
	obj.SetProto(m.legacyParseProto())
	search := ""
	if p.hasSearch {
		search = "?" + p.search
	}
	hash := ""
	if p.hasHash {
		hash = "#" + p.hash
	}
	if p.protocol == "" {
		obj.SetNull("protocol")
	} else {
		obj.SetStr("protocol", p.protocol)
	}
	obj.SetBool("slashes", p.hasAuth)
	auth := ""
	if p.username != "" || p.password != "" {
		auth = p.username
		if p.password != "" {
			auth += ":" + p.password
		}
	}
	if auth == "" {
		obj.SetNull("auth")
	} else {
		obj.SetStr("auth", decodePercentString(auth))
	}
	if hostport := p.hostport(); hostport == "" {
		if p.protocol == "file:" && p.hasAuth {
			obj.SetStr("host", "")
		} else {
			obj.SetNull("host")
		}
	} else {
		obj.SetStr("host", hostport)
	}
	if p.port == "" {
		obj.SetNull("port")
	} else {
		obj.SetStr("port", p.port)
	}
	if p.host == "" {
		if p.protocol == "file:" && p.hasAuth {
			obj.SetStr("hostname", "")
		} else {
			obj.SetNull("hostname")
		}
	} else {
		obj.SetStr("hostname", p.host)
	}
	if p.hasHash {
		obj.SetStr("hash", hash)
	} else {
		obj.SetNull("hash")
	}
	if p.hasSearch {
		obj.SetStr("search", search)
	} else {
		obj.SetNull("search")
	}
	if parseQuery {
		obj.SetObj("query", m.queryObject(parseForm(p.search)))
	} else if p.hasSearch {
		obj.SetStr("query", p.search)
	} else {
		obj.SetNull("query")
	}
	pathname := p.pathname
	if strings.HasPrefix(pathname, "/;") {
		pathname = pathname[1:]
	}
	obj.SetStr("pathname", pathname)
	path := pathname
	if p.hasSearch {
		path += "?" + p.search
	}
	obj.SetStr("path", path)
	obj.SetStr("href", legacyFormat(obj))
	return obj
}

func (m *module) legacyRelative(raw string, parseQuery bool) js.Object {
	c := m.e.JS
	path, search, hash, hasSearch, hasHash := splitTail(raw)
	path = encodePath(path, false)
	search = encodeQuery(search, false)
	hash = encodeLegacyHash(hash)
	obj := c.Obj()
	obj.SetProto(m.legacyParseProto())
	obj.SetNull("protocol")
	obj.SetNull("slashes")
	obj.SetNull("auth")
	obj.SetNull("host")
	obj.SetNull("port")
	obj.SetNull("hostname")
	if hasHash {
		obj.SetStr("hash", "#"+hash)
	} else {
		obj.SetNull("hash")
	}
	if hasSearch {
		obj.SetStr("search", "?"+search)
	} else {
		obj.SetNull("search")
	}
	if parseQuery {
		obj.SetObj("query", m.queryObject(parseForm(search)))
	} else if hasSearch {
		obj.SetStr("query", search)
	} else {
		obj.SetNull("query")
	}
	obj.SetStr("pathname", path)
	pathPart := path
	if hasSearch {
		pathPart += "?" + search
	}
	obj.SetStr("path", pathPart)
	href := pathPart
	if hasHash {
		href += "#" + hash
	}
	obj.SetStr("href", href)
	return obj
}

// legacyParseProto returns Url.prototype, which is also what new url.Url()
// uses. deepStrictEqual compares prototypes, so parse results must match it.
func decodePercentString(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) && isHexDigit(s[i+1]) && isHexDigit(s[i+2]) {
			hi := strings.IndexByte("0123456789abcdef", byte(strings.ToLower(s[i+1 : i+2])[0]))
			lo := strings.IndexByte("0123456789abcdef", byte(strings.ToLower(s[i+2 : i+3])[0]))
			b.WriteByte(byte(hi<<4 | lo))
			i += 2
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// encodeWonkyLegacy handles inputs such as http://google.com" onload="x/,
// where the authority ends at an invalid character and the remainder is a path.
func encodeWonkyLegacy(raw string) (string, bool) {
	i := strings.Index(raw, "://")
	if i < 0 {
		return "", false
	}
	rest := raw[i+3:]
	for j := 0; j < len(rest); j++ {
		c := rest[j]
		if c == ' ' || c == '"' || c == '<' || c == '>' || c == '`' {
			tail := rest[j:]
			tail = strings.ReplaceAll(tail, " ", "%20")
			tail = strings.ReplaceAll(tail, "\"", "%22")
			return raw[:i+3] + rest[:j] + "/" + tail, true
		}
	}
	return "", false
}

func hasMalformedPercent(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			continue
		}
		if i+2 >= len(s) || !isHexDigit(s[i+1]) || !isHexDigit(s[i+2]) {
			return true
		}
		i += 2
	}
	return false
}

func isHexDigit(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

type invalidURLError struct{ input string }

func (e *invalidURLError) Error() string { return "Invalid URL" }

func (e *invalidURLError) JSErrorObject(c *js.Context) js.Object {
	fn, err := c.Eval("(function (input) { const e = new TypeError('Invalid URL'); e.code = 'ERR_INVALID_URL'; e.input = input; return e; })", "[gode:url]")
	if err != nil {
		return c.Obj()
	}
	v, err := fn.Object().Call(js.Object{}, c.Str(e.input))
	if err != nil {
		return c.Obj()
	}
	return v.Object()
}

func encodeLegacyHash(s string) string {
	return encodeWith(s, func(c byte) bool {
		return isAlnum(c) || strings.IndexByte("-._~!$&'()*+,;=:@/?#", c) >= 0
	})
}

func hostTooLong(host string) bool { return len(host) > 255 }

// invalidIDNA rejects hosts whose NFKD form contains characters IDNA forbids.
func (m *module) invalidIDNA(raw string) bool {
	i := strings.Index(raw, "//")
	if i < 0 {
		return false
	}
	rest := raw[i+2:]
	end := strings.IndexAny(rest, "/?#")
	if end < 0 {
		end = len(rest)
	}
	auth := rest[:end]
	if at := strings.LastIndexByte(auth, '@'); at >= 0 {
		auth = auth[at+1:]
	}
	if strings.HasPrefix(auth, "[") {
		return false
	}
	if strings.ContainsRune(auth, '\u00AD') {
		return true
	}
	if j := strings.LastIndexByte(auth, ':'); j >= 0 {
		auth = auth[:j]
	}
	if isASCII(auth) {
		return false
	}
	fn, err := m.e.JS.Eval("(function (h) { return h.normalize('NFKD'); })", "[gode:url]")
	if err != nil {
		return false
	}
	v, err := fn.Object().Call(js.Object{}, m.e.JS.Str(auth))
	if err != nil {
		return false
	}
	return strings.ContainsAny(v.Str(), "#%/:?@[\\]^|")
}

func specialProtocol(raw string) bool {
	lower := strings.ToLower(raw)
	for _, p := range []string{"http:", "https:", "ws:", "wss:", "ftp:", "file:"} {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func (m *module) legacyParseProto() js.Object {
	if u := m.e.Exports("url").Get("Url"); u.IsFunction() {
		if p := u.Object().Get("prototype"); p.IsObject() && !p.IsNull() {
			return p.Object()
		}
	}
	return m.legacyProto
}

func (m *module) invalidArgReceived(v js.Value) string {
	if !v.Valid() || v.IsUndefined() {
		return " Received undefined"
	}
	if v.IsNull() {
		return " Received null"
	}
	switch {
	case v.IsBool():
		return fmt.Sprintf(" Received type boolean (%t)", v.Bool())
	case v.IsNumber():
		return " Received type number (" + strconv.FormatFloat(v.Float(), 'f', -1, 64) + ")"
	case v.IsString():
		return fmt.Sprintf(" Received type string (%q)", v.Str())
	case v.IsFunction():
		return " Received function " + v.Object().Get("name").Str()
	case v.Tag() == "Symbol":
		if fn, err := m.e.JS.Eval("(function (s) { return String(s); })", "[gode:url]"); err == nil {
			if s, err := fn.Object().Call(js.Object{}, v); err == nil {
				return " Received type symbol (" + s.Str() + ")"
			}
		}
		return " Received type symbol ()"
	default:
		return " Received an instance of " + v.Tag()
	}
}

func (m *module) queryObject(pairs [][2]string) js.Object {
	o := m.e.JS.Obj()
	if proto, err := m.e.JS.Eval("Object.create(null)", "[gode:url-query]"); err == nil && proto.IsObject() {
		o = proto.Object()
	}
	for _, p := range pairs {
		cur := o.Get(p[0])
		if cur.IsUndefined() {
			o.SetStr(p[0], p[1])
			continue
		}
		if cur.IsArray() {
			cur.Object().SetAt(cur.Object().Length(), m.e.JS.Str(p[1]))
			continue
		}
		arr := m.e.JS.Arr([]js.Value{cur, m.e.JS.Str(p[1])})
		o.Set(p[0], arr.V())
	}
	return o
}

func legacyFormat(obj js.Object) string {
	if obj.Get("$protocol").IsString() {
		return getParts(obj).href()
	}
	get := func(key string) string {
		v := obj.Get(key)
		if v.IsString() {
			return v.Str()
		}
		return ""
	}
	protocol := get("protocol")
	if protocol != "" && !strings.HasSuffix(protocol, ":") {
		protocol += ":"
	}
	host := get("host")
	hostname, port := get("hostname"), get("port")
	if host == "" && hostname != "" {
		if strings.Contains(hostname, ":") && !strings.HasPrefix(hostname, "[") {
			host = "[" + hostname + "]"
		} else {
			host = hostname
		}
		if port != "" {
			host += ":" + port
		}
	}
	auth := get("auth")
	if auth == "" {
		user, pass := get("username"), get("password")
		if user != "" || pass != "" {
			auth = user
			if pass != "" {
				auth += ":" + pass
			}
		}
	}
	pathname := get("pathname")
	if pathname != "" && host != "" && !strings.HasPrefix(pathname, "/") {
		pathname = "/" + pathname
	}
	pathname = encodePath(pathname, false)
	search := get("search")
	if search != "" && !strings.HasPrefix(search, "?") && !strings.HasPrefix(search, "#") {
		search = "?" + search
	}
	search = strings.ReplaceAll(search, "#", "%23")
	if search == "" {
		q := obj.Get("query")
		switch {
		case q.IsString():
			if q.Str() != "" {
				search = "?" + q.Str()
			}
		case q.IsObject():
			var pairs [][2]string
			for _, k := range q.Object().OwnEnumerableKeys() {
				v := q.Object().Get(k)
				pairs = append(pairs, [2]string{k, v.Str()})
			}
			if len(pairs) > 0 {
				search = "?" + serializeForm(pairs)
			}
		}
	}
	hash := get("hash")
	if hash != "" && !strings.HasPrefix(hash, "#") {
		hash = "#" + hash
	}
	slashes := obj.Get("slashes").Bool() || (specialProtocol(protocol) && (host != "" || protocol == "file:"))
	var b strings.Builder
	b.WriteString(protocol)
	if slashes {
		b.WriteString("//")
	}
	if auth != "" {
		user, pass, hasPass := strings.Cut(auth, ":")
		auth = encodeAuthComponent(user)
		if hasPass {
			auth += ":" + encodeAuthComponent(pass)
		}
		b.WriteString(auth)
		b.WriteString("@")
	}
	b.WriteString(host)
	if pathname != "" {
		if !strings.HasPrefix(pathname, "/") && host != "" {
			b.WriteString("/")
		}
		b.WriteString(pathname)
	}
	b.WriteString(search)
	b.WriteString(hash)
	return b.String()
}

// --- small helpers ---------------------------------------------------------

func arg(args []js.Value, i int) js.Value {
	if i < len(args) {
		return args[i]
	}
	return js.Value{}
}

func argStr(args []js.Value, i int) string {
	if i >= len(args) {
		return ""
	}
	return args[i].Str()
}

func describe(v js.Value) string {
	if !v.Valid() {
		return "undefined"
	}
	switch {
	case v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsString():
		return fmt.Sprintf("type string (%q)", v.Str())
	case v.IsNumber():
		return "type number (" + strconv.FormatFloat(v.Float(), 'f', -1, 64) + ")"
	case v.IsBool():
		return "type boolean"
	}
	return "an instance of " + v.Tag()
}
