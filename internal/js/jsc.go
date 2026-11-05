// Package js is a small, hand-written binding to the JavaScriptCore C API.
//
// GodeJS deliberately does not use V8 (or a Go JS interpreter): JavaScriptCore
// is embedded directly, and every host object exposed to JavaScript is built
// out of these primitives.
//
// JavaScriptCore NaN-boxes JSValueRef, so `undefined` can be the machine word
// 0xa. Every handle here is a uintptr and the C prototypes declare pointers as
// uintptr_t; storing one of those words in an unsafe.Pointer slot makes Go's GC
// abort with "invalid pointer found on stack".
package js

/*
#include <stdlib.h>
#include "godejsc.h"

// Exception out-parameters live in C memory. Passing the address of a Go
// stack variable through uintptr_t hides the escape from the compiler, and a
// callback that throws could then observe a stale slot once frames were
// re-laid-out (the write silently vanished, so errors were dropped).
static uintptr_t godeExcAlloc(void) { return (uintptr_t)calloc(1, sizeof(uintptr_t)); }
static void godeExcFree(uintptr_t p) { free((void*)p); }
static uintptr_t godeExcGet(uintptr_t p) { return p ? *(uintptr_t*)p : 0; }

static uintptr_t godeEmptyString(void) {
	return JSStringCreateWithUTF8CString((uintptr_t)"");
}

static uintptr_t godeMakeFunction(uintptr_t ctx, uintptr_t name, uintptr_t exception) {
	return JSObjectMakeFunctionWithCallback(ctx, name, GODE_JS_CALLBACK, exception);
}

static void godeSetProperty(uintptr_t ctx, uintptr_t object, uintptr_t name, uintptr_t value, int attrs) {
	JSObjectSetProperty(ctx, object, name, value, attrs, 0);
}
*/
import "C"

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

// Property attributes from JavaScriptCore.
const (
	attrNone     = 0
	attrReadOnly = 1 << 1 // kJSCPropertyAttributeReadOnly
	attrDontEnum = 1 << 2 // kJSCPropertyAttributeDontEnum
	attrDontDel  = 1 << 3 // kJSCPropertyAttributeDontDelete
	typedUint8   = 3      // kJSTypedArrayTypeUint8Array
)

// Context is one JavaScriptCore global context.
type Context struct {
	ref uintptr

	mu sync.Mutex

	// propNames caches property-name strings. Creating one per lookup
	// allocated a JavaScript string for every property read across the
	// bridge, and request handling performs several of them per request.
	// The cache is bounded so dynamic keys cannot grow it without limit.
	propNames map[string]C.uintptr_t

	// errToString is the shared Error#toString installed on runtime-created
	// errors that carry a Node error code. One function per context keeps the
	// host-callback table from growing with every thrown error.
	errToString Value
}

type hostFunc struct {
	c  *Context
	fn Func
}

var (
	contexts sync.Map // uintptr -> *Context
	hostMu   sync.Mutex
	hosts    = map[int]*hostFunc{}
	// hostSeq numbers host functions across every context. The id is read
	// back from the JavaScript function object at call time and looked up in
	// the global hosts map, so it must be unique process-wide: numbering per
	// context made the second context's first function overwrite the first
	// context's, and calls then ran the wrong Go function with the right
	// receiver.
	hostSeq int
)

// Func is a Go function exposed to JavaScript. Returning an error throws a
// JavaScript Error with that message.
type Func func(this Object, args []Value) (Value, error)

func u(p uintptr) C.uintptr_t { return C.uintptr_t(p) }

// excBox is a C-allocated exception out-parameter. JavaScriptCore writes the
// thrown value into it during a call; value reads it back afterwards.
type excBox struct{ p C.uintptr_t }

func newExcBox() excBox { return excBox{p: C.godeExcAlloc()} }
func (b excBox) addr() C.uintptr_t {
	return b.p
}
func (b excBox) value() uintptr { return uintptr(C.godeExcGet(b.p)) }
func (b excBox) release()       { C.godeExcFree(b.p) }

// bootstrapJS fills in the explicit-resource-management symbols when the
// JavaScriptCore in use predates them. Node 24 exposes both, and builtins and
// user code feature-detect on them, so they must exist before modules install.
const bootstrapJS = `(function () {
  if (typeof Symbol !== 'function') return;
  if (typeof Symbol.dispose !== 'symbol') {
    Object.defineProperty(Symbol, 'dispose', {
      value: Symbol.for('nodejs.dispose'), configurable: true, writable: true
    });
  }
  if (typeof Symbol.asyncDispose !== 'symbol') {
    Object.defineProperty(Symbol, 'asyncDispose', {
      value: Symbol.for('nodejs.asyncDispose'), configurable: true, writable: true
    });
  }
  // Node's internal error classes render as "Name [ERR_CODE]: message" while
  // system, zlib and user errors with non-ERR codes keep "Name: message".
  if (typeof Error === 'function') {
    function nodeErrorToString() {
      var name = (this && this.name) ? String(this.name) : 'Error';
      var msg = (this && this.message !== undefined && this.message !== null)
        ? String(this.message) : '';
      var code = (this && typeof this.code === 'string') ? this.code : '';
      if (code.indexOf('ERR_') === 0) return name + ' [' + code + ']: ' + msg;
      if (msg === '') return name;
      return name + ': ' + msg;
    }
    var protos = [];
    try { protos.push(Error.prototype); } catch (e) {}
    var names = ['TypeError', 'RangeError', 'SyntaxError', 'ReferenceError', 'EvalError', 'URIError'];
    for (var i = 0; i < names.length; i++) {
      try { if (typeof globalThis[names[i]] === 'function') protos.push(globalThis[names[i]].prototype); } catch (e) {}
    }
    for (var p = 0; p < protos.length; p++) {
      try {
        Object.defineProperty(protos[p], 'toString', {
          value: nodeErrorToString, writable: true, configurable: true, enumerable: false
        });
      } catch (e) {}
    }
  }
})();`

// New creates a fresh JavaScriptCore global context.
func New() *Context {
	// JavaScriptCore's generational collector corrupts the heap while large
	// WebAssembly guests (notably Go's wasip1 runtime) grow memory. The
	// collector option is read when the first context is created. Keep a
	// caller-supplied setting, but default to the safe non-generational
	// collector so WASI guests do not crash the process.
	if os.Getenv("JSC_useGenerationalGC") == "" {
		_ = os.Setenv("JSC_useGenerationalGC", "0")
	}
	var ref C.uintptr_t
	silently(func() { ref = C.JSGlobalContextCreate(0) })
	c := &Context{ref: uintptr(ref)}
	if c.ref != 0 {
		contexts.Store(c.ref, c)
		c.Eval(bootstrapJS, "bootstrap.js")
	}
	runtime.SetFinalizer(c, (*Context).Close)
	return c
}

// Close releases the context. Values from it must not be used afterwards.
func (c *Context) Close() {
	if c == nil || c.ref == 0 {
		return
	}
	contexts.Delete(c.ref)
	hostMu.Lock()
	for id, h := range hosts {
		if h.c == c {
			delete(hosts, id)
		}
	}
	hostMu.Unlock()
	C.JSGlobalContextRelease(u(c.ref))
	c.ref = 0
	runtime.SetFinalizer(c, nil)
}

// Global returns the global object.
func (c *Context) Global() Object {
	return Object{Value{c: c, ref: uintptr(C.JSContextGetGlobalObject(u(c.ref)))}}
}

// Eval runs source in this context. filename is used for stack traces and
// diagnostics.
// CheckSyntax compiles src without executing it, matching node --check.
func (c *Context) CheckSyntax(src, filename string) error {
	script := c.newString(src)
	defer C.JSStringRelease(script)
	var name C.uintptr_t
	if filename != "" {
		name = c.newString(filename)
		defer C.JSStringRelease(name)
	}
	ex := newExcBox()
	defer ex.release()
	if C.JSCheckScriptSyntax(u(c.ref), script, name, 1, ex.addr()) != 0 {
		return nil
	}
	if exception := ex.value(); exception != 0 {
		return c.errorFrom(exception)
	}
	return &Error{Name: "SyntaxError", Message: "Invalid or unexpected token"}
}

func (c *Context) Eval(src, filename string) (Value, error) {
	script := c.newString(src)
	defer C.JSStringRelease(script)
	var name C.uintptr_t
	if filename != "" {
		name = c.newString(filename)
		defer C.JSStringRelease(name)
	}
	ex := newExcBox()
	defer ex.release()
	ref := C.JSEvaluateScript(u(c.ref), script, 0, name, 1, ex.addr())
	if exception := ex.value(); exception != 0 {
		return Value{}, c.errorFrom(exception)
	}
	return Value{c: c, ref: uintptr(ref)}, nil
}

// Promise returns an already-settled Promise: resolved with v when err is nil,
// rejected with err otherwise. Async builtins use it for operations that are
// performed synchronously by the host but must expose Node's promise API.
func (c *Context) Promise(v Value, err error) Value {
	ctor := c.Global().Get("Promise")
	if !ctor.IsObject() {
		return c.Undefined()
	}
	name := "resolve"
	arg := v
	if err != nil {
		name = "reject"
		if custom, ok := err.(JSError); ok {
			if obj := custom.JSErrorObject(c); obj.Valid() {
				arg = obj.V()
			}
		} else {
			// err.Error() already carries "Name: ", so the message has to be
			// read off the *Error or the name ends up in it twice.
			msg := err.Error()
			name, code := "Error", ""
			var errno *int
			if jsErr, ok := err.(*Error); ok {
				if jsErr.Message != "" {
					msg = jsErr.Message
				}
				if jsErr.Name != "" {
					name = jsErr.Name
				}
				code = jsErr.Code
				errno = jsErr.Errno
			}
			ref := c.makeErrorValue(msg)
			obj := Object{Value{c: c, ref: ref}}
			c.setErrorMeta(obj, name, code, errno)
			arg = obj.V()
		}
	}
	fn := ctor.Object().Get(name)
	if !fn.IsFunction() {
		return c.Undefined()
	}
	out, callErr := fn.Object().Call(ctor.Object(), arg)
	if callErr != nil {
		return c.Undefined()
	}
	return out
}

// Undefined returns the undefined value.
func (c *Context) Undefined() Value {
	return Value{c: c, ref: uintptr(C.JSValueMakeUndefined(u(c.ref)))}
}

// Null returns null.
func (c *Context) Null() Value {
	return Value{c: c, ref: uintptr(C.JSValueMakeNull(u(c.ref)))}
}

// Bool returns a boolean value.
func (c *Context) Bool(b bool) Value {
	n := 0
	if b {
		n = 1
	}
	return Value{c: c, ref: uintptr(C.JSValueMakeBoolean(u(c.ref), C.uint8_t(n)))}
}

// Num returns a number value.
func (c *Context) Num(f float64) Value {
	return Value{c: c, ref: uintptr(C.JSValueMakeNumber(u(c.ref), C.double(f)))}
}

// Str returns a string value.
func (c *Context) Str(s string) Value {
	js := c.newString(s)
	defer C.JSStringRelease(js)
	return Value{c: c, ref: uintptr(C.JSValueMakeString(u(c.ref), js))}
}

// ParseJSON parses a JSON document into a value.
func (c *Context) ParseJSON(s string) (Value, error) {
	js := c.newString(s)
	defer C.JSStringRelease(js)
	ref := C.JSValueMakeFromJSONString(u(c.ref), js)
	if ref == 0 {
		return Value{}, &Error{Name: "SyntaxError", Message: "Invalid JSON"}
	}
	return Value{c: c, ref: uintptr(ref)}, nil
}

// Obj creates a plain object.
func (c *Context) Obj() Object {
	return Object{Value{c: c, ref: uintptr(C.JSObjectMake(u(c.ref), 0, 0))}}
}

// Arr creates an array.
func (c *Context) Arr(values []Value) Object {
	var argv C.uintptr_t
	refs := make([]uintptr, len(values))
	for i, v := range values {
		refs[i] = v.ref
	}
	if len(refs) > 0 {
		argv = C.uintptr_t(uintptr(unsafe.Pointer(&refs[0])))
	}
	ex := newExcBox()
	defer ex.release()
	ref := C.JSObjectMakeArray(u(c.ref), C.size_t(len(values)), argv, ex.addr())
	if ex.value() != 0 || ref == 0 {
		ref = C.JSObjectMakeArray(u(c.ref), 0, 0, 0)
	}
	return Object{Value{c: c, ref: uintptr(ref)}}
}

// Error creates a JavaScript Error object.
func (c *Context) Error(msg string) Object {
	return Object{Value{c: c, ref: c.makeErrorValue(msg)}}
}

// JSError is implemented by errors that carry Node-specific fields (code,
// errno, syscall, path) and therefore know how to build their own JavaScript
// error object. The bridge consults it so promise rejections and thrown
// exceptions keep the same shape as their synchronous counterparts.
type JSError interface {
	error
	JSErrorObject(c *Context) Object
}

// nodeErrorString renders an error the way Node's own error classes do:
// `Name [CODE]: message` for Node-internal ERR_* codes, and the plain
// ECMAScript `Name: message` for everything else. Node's system errors
// (ENOENT), zlib errors (Z_DATA_ERROR) and other foreign codes keep the plain
// form even though they carry a `code` property.
func nodeErrorString(e Object) string {
	name := "Error"
	if v := e.Get("name"); v.Valid() && v.IsString() && v.Str() != "" {
		name = v.Str()
	}
	message := ""
	if v := e.Get("message"); v.Valid() && !v.IsUndefined() && !v.IsNull() {
		message = v.Str()
	}
	code := ""
	if v := e.Get("code"); v.Valid() && v.IsString() {
		code = v.Str()
	}
	if strings.HasPrefix(code, "ERR_") {
		return name + " [" + code + "]: " + message
	}
	if message == "" {
		return name
	}
	return name + ": " + message
}

// nodeErrorToString returns this context's shared Node-style Error#toString.
func (c *Context) nodeErrorToString() Value {
	c.mu.Lock()
	fn := c.errToString
	c.mu.Unlock()
	if fn.Valid() {
		return fn
	}
	fn = c.fnObject("toString", func(this Object, _ []Value) (Value, error) {
		return c.Str(nodeErrorString(this)), nil
	}).Protect()
	c.mu.Lock()
	c.errToString = fn
	c.mu.Unlock()
	return fn
}

// setErrorMeta attaches the fields Node puts on its own errors. The toString
// override is installed non-enumerable, the way Node keeps it on the class
// prototype, so deepStrictEqual comparisons of error objects stay honest.
// errorClasses are the built-in JavaScript error classes. JavaScriptCore's
// JSObjectMakeError always produces a plain Error, but Node's RangeErrors and
// TypeErrors really do inherit from those classes: without re-pointing the
// prototype, `err instanceof RangeError` and `err.constructor.name` disagree
// with Node, and assert.throws(fn, RangeError) fails.
var errorClasses = map[string]bool{
	"RangeError": true, "TypeError": true, "SyntaxError": true,
	"ReferenceError": true, "EvalError": true, "URIError": true,
}

// applyErrorPrototype moves a freshly built error object onto the prototype of
// the class it names. Unknown names keep Error.prototype, which is also what
// Node does for errors that merely set a custom `name`.
func (c *Context) applyErrorPrototype(obj Object, name string) {
	if !errorClasses[name] {
		return
	}
	global := C.JSContextGetGlobalObject(u(c.ref))
	ctor := c.newString(name)
	defer C.JSStringRelease(ctor)
	ctorVal := C.JSObjectGetProperty(u(c.ref), global, ctor, 0)
	if ctorVal == 0 || C.JSValueIsObject(u(c.ref), ctorVal) == 0 {
		return
	}
	pname := c.newString("prototype")
	defer C.JSStringRelease(pname)
	proto := C.JSObjectGetProperty(u(c.ref), ctorVal, pname, 0)
	if proto == 0 || C.JSValueIsObject(u(c.ref), proto) == 0 {
		return
	}
	C.JSObjectSetPrototype(u(c.ref), u(obj.ref), proto)
}

func (c *Context) setErrorMeta(obj Object, name, code string, errno *int) {
	if name != "" && name != "Error" {
		obj.SetStr("name", name)
		c.applyErrorPrototype(obj, name)
	}
	if code != "" {
		obj.SetStr("code", code)
		obj.SetHidden("toString", c.nodeErrorToString())
	}
	if errno != nil {
		obj.SetNum("errno", float64(*errno))
	}
}

// ErrorValue turns a Go error into a JavaScript Error object. A *Error that
// came from this context keeps its original identity and code.
func (c *Context) ErrorValue(err error) Value {
	if jsErr, ok := err.(*Error); ok && jsErr.ref != 0 && jsErr.ctx == c {
		return Value{c: c, ref: jsErr.ref}
	}
	if custom, ok := err.(JSError); ok && custom.JSErrorObject(c).Valid() {
		return custom.JSErrorObject(c).V()
	}
	msg := err.Error()
	name, code := "Error", ""
	if coded, ok := err.(CodedError); ok {
		name, code = coded.JSErrorCode()
	}
	if jsErr, ok := err.(*Error); ok {
		if jsErr.Message != "" {
			msg = jsErr.Message
		}
		if jsErr.Name != "" {
			name = jsErr.Name
		}
		code = jsErr.Code
	}
	ref := c.makeErrorValue(msg)
	if ref == 0 {
		return c.Undefined()
	}
	obj := Object{Value{c: c, ref: ref}}
	var errno *int
	if jsErr, ok := err.(*Error); ok {
		errno = jsErr.Errno
	}
	c.setErrorMeta(obj, name, code, errno)
	if jsErr, ok := err.(*Error); ok && jsErr.Input != "" {
		obj.SetStr("input", jsErr.Input)
	}
	return obj.V()
}

func (c *Context) makeErrorValue(msg string) uintptr {
	js := c.newString(msg)
	defer C.JSStringRelease(js)
	arg := C.JSValueMakeString(u(c.ref), js)
	ex := newExcBox()
	defer ex.release()
	return uintptr(C.JSObjectMakeError(u(c.ref), 1,
		C.uintptr_t(uintptr(unsafe.Pointer(&arg))), ex.addr()))
}

// Bytes creates a Uint8Array copying b.
func (c *Context) Bytes(b []byte) Object {
	ref := C.JSObjectMakeTypedArray(u(c.ref), typedUint8, C.size_t(len(b)), 0)
	if len(b) > 0 {
		ptr := C.JSObjectGetTypedArrayBytesPtr(u(c.ref), ref, 0)
		if ptr != 0 {
			copy(unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), len(b)), b)
		}
	}
	return Object{Value{c: c, ref: uintptr(ref)}}
}

func (c *Context) newString(s string) C.uintptr_t {
	if s == "" {
		return C.godeEmptyString()
	}
	units := utf16Encode(s)
	return C.JSStringCreateWithCharacters(
		C.uintptr_t(uintptr(unsafe.Pointer(&units[0]))), C.size_t(len(units)))
}

func (c *Context) stringFrom(s C.uintptr_t) string {
	if s == 0 {
		return ""
	}
	n := C.JSStringGetLength(s)
	if n == 0 {
		return ""
	}
	ptr := C.JSStringGetCharactersPtr(s)
	if ptr == 0 {
		return ""
	}
	return utf16Decode(unsafe.Slice((*uint16)(unsafe.Pointer(uintptr(ptr))), int(n)))
}

func (c *Context) errorFrom(exception uintptr) error {
	e := &Error{Name: "Error", Message: "JavaScript exception", ref: exception, ctx: c}
	if C.JSValueIsObject(u(c.ref), u(exception)) != 0 {
		obj := C.JSValueToObject(u(c.ref), u(exception), 0)
		if obj != 0 {
			if s, ok := c.propertyString(obj, "name"); ok {
				e.Name = s
			}
			if s, ok := c.propertyString(obj, "message"); ok {
				e.Message = s
			}
			if s, ok := c.propertyString(obj, "code"); ok {
				e.Code = s
			}
			if s, ok := c.propertyString(obj, "stack"); ok {
				e.Stack = s
			}
			if s, ok := c.propertyString(obj, "sourceURL"); ok {
				e.File = s
			}
			if n, ok := c.propertyNumber(obj, "line"); ok {
				e.Line = int(n)
			}
			if n, ok := c.propertyNumber(obj, "column"); ok {
				e.Column = int(n)
			}
		}
	}
	if e.Message == "JavaScript exception" && e.Name == "Error" {
		s := C.JSValueToStringCopy(u(c.ref), u(exception), 0)
		if s != 0 {
			e.Message = c.stringFrom(s)
			C.JSStringRelease(s)
		}
	}
	if e.Line == 0 {
		e.Line = 1
	}
	return e
}

func (c *Context) propertyString(obj C.uintptr_t, key string) (string, bool) {
	v := c.get(obj, key)
	if v == 0 || C.JSValueIsString(u(c.ref), v) == 0 {
		return "", false
	}
	s := C.JSValueToStringCopy(u(c.ref), v, 0)
	if s == 0 {
		return "", false
	}
	defer C.JSStringRelease(s)
	return c.stringFrom(s), true
}

func (c *Context) propertyNumber(obj C.uintptr_t, key string) (float64, bool) {
	v := c.get(obj, key)
	if v == 0 || C.JSValueIsNumber(u(c.ref), v) == 0 {
		return 0, false
	}
	return float64(C.JSValueToNumber(u(c.ref), v, 0)), true
}

// maxCachedNames bounds the property-name cache so user-controlled keys
// cannot grow it forever.
const maxCachedNames = 4096

// propName returns a cached JSString for key. The second result is false when
// the cache is full, in which case the caller owns (and must release) a fresh
// string.
func (c *Context) propName(key string) (C.uintptr_t, bool) {
	if s, ok := c.propNames[key]; ok {
		return s, true
	}
	if len(c.propNames) >= maxCachedNames {
		return 0, false
	}
	if c.propNames == nil {
		c.propNames = make(map[string]C.uintptr_t, 128)
	}
	s := c.newString(key)
	c.propNames[key] = s
	return s, true
}

func (c *Context) get(obj C.uintptr_t, key string) C.uintptr_t {
	if s, ok := c.propName(key); ok {
		return C.JSObjectGetProperty(u(c.ref), obj, s, 0)
	}
	js := c.newString(key)
	defer C.JSStringRelease(js)
	return C.JSObjectGetProperty(u(c.ref), obj, js, 0)
}

func (c *Context) fnObject(name string, fn Func) Value {
	n := c.newString(name)
	defer C.JSStringRelease(n)
	ex := newExcBox()
	defer ex.release()
	ref := C.godeMakeFunction(u(c.ref), n, ex.addr())
	if ref == 0 {
		return c.Undefined()
	}
	hostMu.Lock()
	hostSeq++
	id := hostSeq
	hosts[id] = &hostFunc{c: c, fn: fn}
	hostMu.Unlock()
	c.set(ref, "$gode$id", C.JSValueMakeNumber(u(c.ref), C.double(id)),
		attrDontEnum|attrDontDel|attrReadOnly)
	return Value{c: c, ref: uintptr(ref)}
}

func (c *Context) set(obj C.uintptr_t, key string, val C.uintptr_t, attrs int) {
	js := c.newString(key)
	defer C.JSStringRelease(js)
	C.godeSetProperty(u(c.ref), obj, js, val, C.int(attrs))
}
