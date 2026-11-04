Writing a GodeJS Builtin
========================

GodeJS embeds **JavaScriptCore** (not V8, not goja). Go builds every host
object through ``internal/js``; there is no JavaScript shim layer and no Node
source is vendored. The compatibility bar is: **stdout must be byte-identical
to Node.js 24 LTS** for the APIs a module claims.

Package contract
----------------

Each module lives in ``modules/<name>/`` and does exactly this:

.. code-block:: go

   package path

   import (
       "godejs/internal/env"
       "godejs/internal/js"
   )

   func init() { env.Register("path", Install) }

   // Install builds the module's exports object.
   func Install(e *env.Env) js.Object { ... }

   // Implemented lists the tracked keys this package actually provides.
   // Keys must be copied verbatim from parity/baseline/<name>.json.
   func Implemented() []string { return []string{"path.join", "path.sep"} }

* ``init`` is what makes ``require("path")`` and ``require("node:path")`` work.
* If a module also creates globals (``console``, ``process``, ``Buffer``,
  ``setTimeout``, ...), do it in ``Install`` via ``e.Global().Set(...)``.
* Only edit files inside your own package. Never edit ``internal/``, ``cmd/``,
  or another module.

Runtime surface
---------------

.. code-block:: go

   e.JS                 // *js.Context (the JavaScriptCore context)
   e.JS.Obj()           // {} fresh object
   e.JS.Arr([]js.Value) // array
   e.JS.Str(s)          // JS string
   e.JS.Num(f)          // JS number
   e.JS.Bool(b)         // JS boolean
   e.JS.Bytes([]byte)   // Uint8Array (copied)
   e.JS.ParseJSON(s)    // (js.Value, error)
   e.JS.Eval(src, file) // (js.Value, error)

   e.Out(), e.Err(), e.In()       // io.Writer / io.Reader
   e.Argv(), e.Environ(), e.Cwd() // process state
   e.Exit(code)                   // sets exit code; return env.ErrProcessExit to unwind
   e.Global()                     // global object
   e.Obj()                        // fresh object
   e.Exports("util")              // lazily installs and returns another builtin
   e.SetTimeout(fn js.Value, ms float64, repeat bool, args []js.Value) int
   e.ClearTimer(id int)
   e.RefreshTimer(id int, ms float64) bool
   e.TimerActive(id int) bool
   e.NextTick(fn js.Value, args ...js.Value)

Values:

.. code-block:: go

   v.IsUndefined() v.IsNull() v.IsBool() v.IsNumber() v.IsString()
   v.IsObject() v.IsArray() v.IsDate() v.IsFunction() v.IsError()
   v.Bool() v.Float() v.Int() v.Str() v.StrOK() (string, bool)
   v.Bytes() ([]byte, bool)   // Uint8Array only
   v.JSON(indent) (string, error)
   v.Export() any             // plain Go data, useful in host helpers
   v.Tag() string             // "Array", "Date", "Map", "Set", "Uint8Array", ...

Objects:

.. code-block:: go

   o.Get(k) o.Has(k) o.Keys() o.OwnKeys() o.OwnEnumerableKeys()
   o.Set(k,v) o.SetHidden(k,v) o.SetConst(k,v)
   o.SetStr o.SetNum o.SetBool o.SetNull o.SetObj
   o.SetFunc(k, func(this js.Object, args []js.Value) (js.Value, error))
   o.SetFuncHidden(...)
   o.Delete(k) o.Length() o.At(i) o.SetAt(i,v)
   o.Proto() o.SetProto(p)
   o.Call(this js.Object, args ...js.Value) (js.Value, error)

Host functions **return a Go ``error`` to throw**. The error text becomes the
JS Error message, so make it Node-identical (no added prefix). Panics are
converted to JS errors by the runtime, but do not rely on that.

Tests
-----

Use ``internal/jstest`` (no import cycle: it does not import builtin modules).

.. code-block:: go

   func TestJoin(t *testing.T) {
       e := jstest.New(t)
       jstest.Install(t, e, "path", Install)

       jstest.Equal(t, jstest.Str(t, e, `path.join("a", "b")`), "a/b")
       if got := jstest.Str(t, e, `JSON.stringify(path.parse("/x/y.js"))`); got != `{"root":"/","dir":"/x","base":"y.js","ext":".js","name":"y"}` {
           t.Fatalf("parse: %s", got)
       }
   }

``jstest.Throws`` asserts a JS error and returns it. Prefer asserting exact
strings over shapes. For byte-exact stdout, capture the buffer you passed to
``env.Options.Stdout`` and compare with ``==``.

Do not add third-party dependencies. Do not use cgo.

Cross-module contracts
----------------------

* ``buffer``: every Buffer is a ``Uint8Array`` with a **non-enumerable**
  ``_godeBuffer === true`` property. ``Buffer.prototype`` carries the methods.
* ``util.Inspect(v js.Value) string`` is exported by ``modules/util`` and used
  by ``console``. It must produce Node's ``util.inspect`` output (depth 2 by
  default, ``{ a: 1 }``, ``[ 1, 2 ]``, ``'text'``, ``<Buffer 01 02>``).
* ``assert`` implements its own strict deep equality; do not depend on util.
* ``process`` owns ``process.stdout/stderr/stdin``; ``console`` writes through
  ``e.Out()``.

Node fidelity notes
-------------------

* ``console.log("a", 1)`` prints ``a 1``; objects go through ``util.inspect``.
* Strings printed by ``console.log`` are unquoted; strings nested in objects
  are quoted with single quotes; escape ``'``, ``\``, and control characters
  the way Node does.
* Arrays print ``[]``, ``[ 1 ]``, ``[ 1, 2 ]``; objects print ``{}``,
  ``{ a: 1 }``.
* Circular structures print ``<ref *1> { ... }``; implement it rather than
  recursing forever.
* ``path`` on Linux is POSIX. ``path.win32`` must exist and behave like Node's
  win32 implementation.
* Never print debug output, banners, or warnings. A module that cannot support
  an API should omit it instead of silently approximating it.
