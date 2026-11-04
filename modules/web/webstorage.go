package web

// Web Storage: the Storage interface plus the localStorage and sessionStorage
// globals. localStorage persists to a JSON file (per working directory by
// default, or --localstorage-file / GODE_LOCALSTORAGE_FILE); sessionStorage
// lives only for the process.
//
// Node implements localStorage on SQLite behind --experimental-webstorage and
// leaves it undefined without --localstorage-file. GodeJS makes both areas
// available unconditionally and persists localStorage by default, which is the
// behaviour applications written for browsers and Bun expect. The on-disk
// format is private to this runtime; only the JavaScript-visible semantics
// follow the Web Storage standard.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"godejs/internal/env"
	"godejs/internal/js"
)

// storage is one Web Storage area. Keys keep insertion order, matching the
// order the standard requires getItem()/key()/enumeration to observe.
type storage struct {
	mu   sync.Mutex
	keys []string
	vals map[string]string
	path string // "" for an in-memory area
}

type storageFile struct {
	Entries [][2]string `json:"entries"`
}

func newStorage(path string) *storage {
	s := &storage{vals: map[string]string{}, path: path}
	if path == "" {
		return s
	}
	data, err := os.ReadFile(path)
	// A corrupt, unreadable or absurdly large file starts empty rather than
	// aborting startup: storage is auxiliary state, not program input.
	if err != nil || len(data) > 64<<20 {
		return s
	}
	var f storageFile
	if json.Unmarshal(data, &f) != nil {
		return s
	}
	for _, entry := range f.Entries {
		if len(entry) != 2 {
			continue
		}
		if _, ok := s.vals[entry[0]]; !ok {
			s.keys = append(s.keys, entry[0])
		}
		s.vals[entry[0]] = entry[1]
	}
	return s
}

func (s *storage) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.keys)
}

func (s *storage) get(k string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.vals[k]
	return v, ok
}

func (s *storage) has(k string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.vals[k]
	return ok
}

func (s *storage) set(k, v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.vals[k]; !ok {
		s.keys = append(s.keys, k)
	}
	s.vals[k] = v
	s.flushLocked()
}

func (s *storage) remove(k string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.vals[k]; !ok {
		return
	}
	delete(s.vals, k)
	for i, key := range s.keys {
		if key == k {
			s.keys = append(s.keys[:i], s.keys[i+1:]...)
			break
		}
	}
	s.flushLocked()
}

func (s *storage) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.keys) == 0 {
		return
	}
	s.keys = nil
	s.vals = map[string]string{}
	s.flushLocked()
}

func (s *storage) keyAt(i int) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i < 0 || i >= len(s.keys) {
		return "", false
	}
	return s.keys[i], true
}

func (s *storage) names() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.keys...)
}

// flushLocked writes the area to disk. Mutations are rare compared with reads,
// so a synchronous atomic rewrite keeps the on-disk copy exact without needing
// a background writer or an exit hook; a failed write leaves the in-memory
// area authoritative and is retried on the next mutation.
func (s *storage) flushLocked() {
	if s.path == "" {
		return
	}
	f := storageFile{Entries: make([][2]string, 0, len(s.keys))}
	for _, k := range s.keys {
		f.Entries = append(f.Entries, [2]string{k, s.vals[k]})
	}
	data, err := json.Marshal(f)
	if err != nil {
		return
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	tmp, err := os.CreateTemp(dir, ".gode-storage-*")
	if err != nil {
		return
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(name)
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return
	}
	_ = os.Chmod(name, 0o600)
	if err := os.Rename(name, s.path); err != nil {
		os.Remove(name)
	}
}

// localStoragePath resolves the file backing localStorage. Precedence: the
// Node-compatible --localstorage-file flag, GODE_LOCALSTORAGE_FILE, then a
// per-working-directory file under the XDG state directory.
func localStoragePath(e *env.Env) string {
	argv := e.Opts.ExecArgv
	for i := 0; i < len(argv); i++ {
		if argv[i] == "--localstorage-file" && i+1 < len(argv) {
			return argv[i+1]
		}
		if v, ok := strings.CutPrefix(argv[i], "--localstorage-file="); ok {
			return v
		}
	}
	if p := os.Getenv("GODE_LOCALSTORAGE_FILE"); p != "" {
		return p
	}
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			base = filepath.Join(home, ".local", "state")
		} else {
			base = filepath.Join(os.TempDir(), "gode-state")
		}
	}
	sum := sha256.Sum256([]byte(e.Cwd()))
	return filepath.Join(base, "gode", "localstorage", hex.EncodeToString(sum[:8])+".json")
}

// installStorage defines Storage, localStorage, sessionStorage and
// QuotaExceededError on e.
func installStorage(e *env.Env) {
	local := newStorage(localStoragePath(e))
	session := newStorage("")

	pick := func(kind string) *storage {
		if kind == "local" {
			return local
		}
		return session
	}

	host := e.JS.Obj()
	host.SetFuncHidden("length", func(_ js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Num(float64(pick(argString(args, 0)).count())), nil
	})
	host.SetFuncHidden("get", func(_ js.Object, args []js.Value) (js.Value, error) {
		if v, ok := pick(argString(args, 0)).get(argString(args, 1)); ok {
			return e.JS.Str(v), nil
		}
		return e.JS.Undefined(), nil
	})
	host.SetFuncHidden("has", func(_ js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Bool(pick(argString(args, 0)).has(argString(args, 1))), nil
	})
	host.SetFuncHidden("set", func(_ js.Object, args []js.Value) (js.Value, error) {
		pick(argString(args, 0)).set(argString(args, 1), argString(args, 2))
		return e.JS.Undefined(), nil
	})
	host.SetFuncHidden("remove", func(_ js.Object, args []js.Value) (js.Value, error) {
		pick(argString(args, 0)).remove(argString(args, 1))
		return e.JS.Undefined(), nil
	})
	host.SetFuncHidden("clear", func(_ js.Object, args []js.Value) (js.Value, error) {
		pick(argString(args, 0)).clear()
		return e.JS.Undefined(), nil
	})
	host.SetFuncHidden("key", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) < 2 {
			return e.JS.Undefined(), nil
		}
		index := args[1].Float()
		if math.IsNaN(index) || index < 0 {
			return e.JS.Undefined(), nil
		}
		if k, ok := pick(argString(args, 0)).keyAt(int(index)); ok {
			return e.JS.Str(k), nil
		}
		return e.JS.Undefined(), nil
	})
	host.SetFuncHidden("keys", func(_ js.Object, args []js.Value) (js.Value, error) {
		names := pick(argString(args, 0)).names()
		values := make([]js.Value, len(names))
		for i, name := range names {
			values[i] = e.JS.Str(name)
		}
		return e.JS.Arr(values).V(), nil
	})

	e.Global().Set("__godeStorage", host.V())
	if _, err := e.JS.Eval(storagePrelude, "gode:storage"); err != nil {
		panic("webstorage: " + err.Error())
	}
}

func argString(args []js.Value, i int) string {
	if i >= len(args) {
		return ""
	}
	return args[i].Str()
}

const storagePrelude = `(function (host) {
  'use strict';
  delete globalThis.__godeStorage;

  var INTERNAL = {};
  var AREA_KIND = new WeakMap();

  function areaKind(receiver) {
    var kind = AREA_KIND.get(receiver);
    if (kind !== 'local' && kind !== 'session') {
      throw new TypeError('Illegal invocation');
    }
    return kind;
  }

  function toDOMString(value) {
    if (typeof value === 'symbol') {
      throw new TypeError('Cannot convert a Symbol value to a string');
    }
    return String(value);
  }

  function toUnsignedLong(value) {
    if (typeof value === 'symbol') {
      throw new TypeError('Cannot convert a Symbol value to a number');
    }
    if (typeof value === 'bigint') {
      throw new TypeError('Cannot convert a BigInt value to a number');
    }
    return Number(value) >>> 0;
  }

  function required(args, count, method) {
    if (args.length < count) {
      throw new TypeError("Failed to execute '" + method + "' on 'Storage': " +
        count + ' argument' + (count === 1 ? '' : 's') +
        ' required, but only ' + args.length + ' present.');
    }
  }

  // Storage is a legacy platform object: the standard methods live on the
  // prototype and every other property name is a storage entry. A Proxy
  // supplies that named-property view, the way browsers and Node expose it.
  class Storage {
    constructor(token, area) {
      if (token !== INTERNAL) throw new TypeError('Illegal constructor');
      Object.defineProperty(this, '_kind', {
        value: area, writable: false, enumerable: false, configurable: true
      });
    }
    get length() { return host.length(areaKind(this)); }
    key(index) {
      required(arguments, 1, 'key');
      var k = host.key(areaKind(this), toUnsignedLong(index));
      return k === undefined ? null : k;
    }
    getItem(key) {
      required(arguments, 1, 'getItem');
      var v = host.get(areaKind(this), toDOMString(key));
      return v === undefined ? null : v;
    }
    setItem(key, value) {
      required(arguments, 2, 'setItem');
      host.set(areaKind(this), toDOMString(key), toDOMString(value));
    }
    removeItem(key) {
      required(arguments, 1, 'removeItem');
      host.remove(areaKind(this), toDOMString(key));
    }
    clear() { host.clear(areaKind(this)); }
  }

  var RESERVED = { length: true };
  function isReserved(name) {
    return RESERVED[name] === true || name in Storage.prototype ||
      Object.prototype.hasOwnProperty.call(this, name);
  }

  function makeArea(kind) {
    var target = new Storage(INTERNAL, kind);
    var area = new Proxy(target, {
      get: function (t, key, receiver) {
        if (typeof key === 'symbol' || isReserved.call(t, key)) {
          return Reflect.get(t, key, receiver);
        }
        var v = host.get(kind, String(key));
        return v === undefined ? undefined : v;
      },
      set: function (t, key, value, receiver) {
        if (typeof key === 'symbol' || isReserved.call(t, key)) {
          if (key === 'length') return true;
          return Reflect.set(t, key, value, receiver);
        }
        host.set(kind, String(key), toDOMString(value));
        return true;
      },
      has: function (t, key) {
        if (typeof key === 'symbol' || isReserved.call(t, key)) return true;
        return host.has(kind, String(key));
      },
      deleteProperty: function (t, key) {
        if (typeof key === 'symbol' || isReserved.call(t, key)) {
          if (key === 'length') return true;
          return Reflect.deleteProperty(t, key);
        }
        host.remove(kind, String(key));
        return true;
      },
      ownKeys: function () { return host.keys(kind); },
      getOwnPropertyDescriptor: function (t, key) {
        if (typeof key === 'symbol' || isReserved.call(t, key)) {
          return Object.getOwnPropertyDescriptor(t, key);
        }
        if (!host.has(kind, String(key))) return undefined;
        return {
          value: host.get(kind, String(key)),
          writable: true, enumerable: true, configurable: true
        };
      }
    });
    AREA_KIND.set(area, kind);
    return area;
  }

  globalThis.Storage = Storage;
  globalThis.localStorage = makeArea('local');
  globalThis.sessionStorage = makeArea('session');

  if (typeof DOMException === 'function') {
    class QuotaExceededError extends DOMException {
      constructor(message, name) {
        super(message === undefined ? 'The quota has been exceeded.' : String(message),
              name === undefined ? 'QuotaExceededError' : String(name));
      }
    }
    globalThis.QuotaExceededError = QuotaExceededError;
  }
})(globalThis.__godeStorage);`
