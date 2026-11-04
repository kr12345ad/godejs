package napi

/*
#cgo LDFLAGS: -ldl -Wl,--export-dynamic
#include "gode_napi.h"
*/
import "C"

import (
	"fmt"
	"sync"
	"unicode/utf16"
	"unsafe"

	"godejs/internal/env"
	"godejs/internal/js"
)

// The Node-API ABI deliberately treats napi_env and napi_value as opaque
// pointers. Gode passes small integer IDs instead. Addons never dereference
// them, and every operation resolves the ID back to a live JavaScriptCore
// value in Go. This keeps Go pointers out of C memory and keeps the ABI stable
// across Go versions.
type runtime struct {
	id uintptr
	e  *env.Env

	lastException js.Value

	instanceData uintptr

	instanceofFn js.Value
	fnFactory    js.Value
	dateCtor     js.Value
	defineValue  js.Value
	defineAcc    js.Value
	hasOwn       js.Value

	// scopes is the Node-API handle-scope stack. Values created while a
	// scope is open are released when it closes; that is what makes a long
	// running addon stop accumulating protected JavaScriptCore handles.
	scopes  [][]uintptr
	retired []js.Value

	cleanupHooks []*cleanupHook
	finalizers   []finalizer

	externalMemory int64
}

type callbackInfo struct {
	rt        *runtime
	this      js.Value
	newTarget js.Value
	args      []js.Value
	data      uintptr
}

type napiRef struct {
	rt    *runtime
	value js.Value
	count uint32
}

type wrapInfo struct {
	rt       *runtime
	native   uintptr
	finalize uintptr
	hint     uintptr
}

type deferred struct {
	rt      *runtime
	resolve js.Value
	reject  js.Value
}

func init() {
	env.SetNativeLoader(Load)
}

var (
	mu sync.Mutex

	nextEnv     uintptr = 1
	nextValue   uintptr = 1
	nextRef     uintptr = 1
	nextInfo    uintptr = 1
	nextWrap    uintptr = 1
	nextDefer   uintptr = 1
	nextAsync   uintptr = 1
	nextScope   uintptr = 1
	pendingMod  *C.gode_napi_module
	runtimes            = map[uintptr]*runtime{}
	ctxRuntimes         = map[*js.Context]*runtime{}
	values              = map[uintptr]js.Value{}
	refs                = map[uintptr]*napiRef{}
	infos               = map[uintptr]*callbackInfo{}
	wraps               = map[uintptr]*wrapInfo{}
	deferreds           = map[uintptr]*deferred{}
	asyncWorks          = map[uintptr]*asyncWork{}
	handleScopes        = map[uintptr]*runtime{}
	escaped             = map[uintptr]uintptr{}
)

const (
	napiOK                    = 0
	napiInvalidArg            = 1
	napiObjectExpected        = 2
	napiStringExpected        = 3
	napiNameExpected          = 4
	napiFunctionExpected      = 5
	napiNumberExpected        = 6
	napiBooleanExpected       = 7
	napiArrayExpected         = 8
	napiGenericFailure        = 9
	napiPendingException      = 10
	napiCancelled             = 11
	napiBigintExpected        = 17
	napiDateExpected          = 18
	napiArraybufferExpected   = 19
	napiDetachableArraybuffer = 20
)

func newRuntime(e *env.Env) *runtime {
	mu.Lock()
	id := nextEnv
	nextEnv++
	rt := &runtime{id: id, e: e}
	runtimes[id] = rt
	ctxRuntimes[e.JS] = rt
	mu.Unlock()
	e.OnExit(func(int) { rt.shutdown() })
	return rt
}

func runtimeForCtx(c *js.Context) *runtime {
	mu.Lock()
	rt := ctxRuntimes[c]
	mu.Unlock()
	return rt
}

// pushScope opens a handle scope and releases the values retired by the
// previous one. Unprotecting outside the lock keeps JavaScriptCore's
// collector from re-entering our bookkeeping.
func (rt *runtime) pushScope() {
	mu.Lock()
	list := rt.retired
	rt.retired = nil
	rt.scopes = append(rt.scopes, nil)
	mu.Unlock()
	for _, v := range list {
		v.Unprotect()
	}
}

// popScope closes a handle scope. escape is a value id that must survive into
// the parent scope; zero means every value created here is released.
func (rt *runtime) popScope(escape uintptr) {
	mu.Lock()
	if len(rt.scopes) == 0 {
		mu.Unlock()
		return
	}
	top := len(rt.scopes) - 1
	ids := rt.scopes[top]
	rt.scopes = rt.scopes[:top]
	for _, id := range ids {
		if id == 0 || id == escape {
			continue
		}
		if v, ok := values[id]; ok {
			rt.retired = append(rt.retired, v)
			delete(values, id)
		}
	}
	if escape != 0 && len(rt.scopes) > 0 {
		top = len(rt.scopes) - 1
		rt.scopes[top] = append(rt.scopes[top], escape)
	}
	mu.Unlock()
}

func runtimeOf(id uintptr) *runtime {
	mu.Lock()
	rt := runtimes[id]
	mu.Unlock()
	return rt
}

func valueID(v js.Value) uintptr {
	if !v.Valid() {
		return 0
	}
	v.Protect()
	mu.Lock()
	id := nextValue
	nextValue++
	values[id] = v
	if rt := ctxRuntimes[v.Ctx()]; rt != nil && len(rt.scopes) > 0 {
		top := len(rt.scopes) - 1
		rt.scopes[top] = append(rt.scopes[top], id)
	}
	mu.Unlock()
	return id
}

func valueOf(id uintptr) (js.Value, bool) {
	if id == 0 {
		return js.Value{}, false
	}
	mu.Lock()
	v, ok := values[id]
	mu.Unlock()
	return v, ok
}

func (rt *runtime) undefined() uintptr { return valueID(rt.e.JS.Undefined()) }
func (rt *runtime) null() uintptr      { return valueID(rt.e.JS.Null()) }

func (rt *runtime) takeException() js.Value {
	mu.Lock()
	v := rt.lastException
	rt.lastException = js.Value{}
	mu.Unlock()
	return v
}

func (rt *runtime) setException(v js.Value) {
	if !v.Valid() {
		return
	}
	v.Protect()
	mu.Lock()
	rt.lastException = v
	mu.Unlock()
}

func (rt *runtime) setError(err error) {
	if err == nil {
		return
	}
	rt.setException(rt.e.JS.ErrorValue(err))
}

type thrown struct {
	rt    *runtime
	value js.Value
}

func (t *thrown) Error() string {
	if t.value.IsObject() {
		if m := t.value.Object().Get("message"); m.IsString() {
			return m.Str()
		}
	}
	return t.value.Str()
}

func (t *thrown) JSErrorObject(c *js.Context) js.Object {
	if t.value.Valid() && t.value.Ctx() == c && t.value.IsObject() {
		return t.value.Object()
	}
	return js.Object{}
}

func cString(p *C.char) string {
	if p == nil {
		return ""
	}
	return C.GoString(p)
}

func setUint(out *C.uintptr_t, v uintptr) bool {
	if out == nil {
		return false
	}
	*out = C.uintptr_t(v)
	return true
}

// wrapFunction turns a host callback into a JavaScript function that forwards
// the call's `this`, `new.target` and argument list. new.target is what makes
// napi_get_new_target and class constructors behave like Node's.
func (rt *runtime) wrapFunction(host js.Value, name string) js.Value {
	if !rt.fnFactory.Valid() {
		rt.fnFactory, _ = rt.e.JS.Eval(`(function (host, name) {
	var F = function () {
		var a = [];
		for (var i = 0; i < arguments.length; i++) a.push(arguments[i]);
		return host(this, new.target, a);
	};
	try { Object.defineProperty(F, "name", { value: name, configurable: true }); } catch (e) {}
	return F;
})`, "[napi:function]")
	}
	if !rt.fnFactory.Valid() {
		return rt.e.JS.Undefined()
	}
	out, err := rt.fnFactory.Object().Call(js.Object{}, host, rt.e.JS.Str(name))
	if err != nil {
		return rt.e.JS.Undefined()
	}
	return out
}

func (rt *runtime) callback(name string, cb uintptr, data uintptr) js.Value {
	fnName := name
	if fnName == "" {
		fnName = "napi"
	}
	holder := rt.e.JS.Obj()
	holder.SetFuncHidden("host", func(_ js.Object, argv []js.Value) (js.Value, error) {
		var thisVal, newTarget js.Value
		var args []js.Value
		if len(argv) > 0 {
			thisVal = argv[0]
		}
		if len(argv) > 1 {
			newTarget = argv[1]
		}
		if len(argv) > 2 && argv[2].IsArray() {
			arr := argv[2].Object()
			n := arr.Length()
			args = make([]js.Value, 0, n)
			for i := 0; i < n; i++ {
				args = append(args, arr.At(i))
			}
		}
		rt.pushScope()
		mu.Lock()
		id := nextInfo
		nextInfo++
		infos[id] = &callbackInfo{rt: rt, this: thisVal, newTarget: newTarget, args: args, data: data}
		mu.Unlock()
		res := C.gode_call_callback(C.uintptr_t(cb), C.uintptr_t(rt.id), C.uintptr_t(id))
		var out js.Value
		if res != 0 {
			out, _ = valueOf(uintptr(res))
		}
		mu.Lock()
		delete(infos, id)
		mu.Unlock()
		rt.popScope(uintptr(res))
		if exc := rt.takeException(); exc.Valid() {
			return js.Value{}, js.ThrownValue(exc)
		}
		if !out.Valid() {
			out = rt.e.JS.Undefined()
		}
		return out, nil
	})
	return rt.wrapFunction(holder.Get("host"), fnName)
}

// Load opens a .node addon, resolves its registration function, and invokes
// it. The returned JavaScript object is the addon's module.exports value.
func Load(e *env.Env, path string) (js.Object, error) {
	rt := newRuntime(e)
	exports := e.JS.Obj()
	exports.Protect()
	exportsID := valueID(exports.V())

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	h := C.gode_dlopen(cpath)
	if h == nil {
		return js.Object{}, &js.Error{Name: "Error", Code: "ERR_DLOPEN_FAILED",
			Message: "ERR_DLOPEN_FAILED: " + C.GoString(C.gode_dlerror()) + ", dlopen '" + path + "'"}
	}

	name := C.CString("napi_register_module_v1")
	fn := C.gode_dlsym(h, name)
	C.free(unsafe.Pointer(name))
	if fn == 0 {
		mu.Lock()
		mod := pendingMod
		pendingMod = nil
		mu.Unlock()
		if mod != nil {
			fn = C.gode_module_register_func(mod)
		}
	}
	if fn == 0 {
		return js.Object{}, &js.Error{Name: "Error", Code: "ERR_DLOPEN_FAILED",
			Message: "ERR_DLOPEN_FAILED: module did not export napi_register_module_v1, '" + path + "'"}
	}
	rt.pushScope()
	res := C.gode_call_register(unsafe.Pointer(uintptr(fn)), C.uintptr_t(rt.id), C.uintptr_t(exportsID))
	var regOut js.Value
	if res != 0 {
		regOut, _ = valueOf(uintptr(res))
	}
	rt.popScope(uintptr(res))
	if exc := rt.takeException(); exc.Valid() {
		return js.Object{}, js.ThrownValue(exc)
	}
	if !regOut.Valid() {
		return exports, nil
	}
	if v := regOut; v.IsObject() {
		return v.Object(), nil
	}
	v, ok := valueOf(uintptr(res))
	if !ok || !v.IsObject() {
		return exports, nil
	}
	return v.Object(), nil
}

//export napi_module_register
func napi_module_register(mod *C.gode_napi_module) {
	mu.Lock()
	pendingMod = mod
	mu.Unlock()
}

//export napi_get_version
func napi_get_version(_ C.uintptr_t, result *C.uint32_t) C.int {
	if result == nil {
		return napiInvalidArg
	}
	*result = 10
	return napiOK
}

//export napi_get_node_version
func napi_get_node_version(_ C.uintptr_t, result **C.gode_napi_node_version) C.int {
	if result == nil {
		return napiInvalidArg
	}
	*result = C.gode_node_version()
	return napiOK
}

//export napi_get_last_error_info
func napi_get_last_error_info(_ C.uintptr_t, result **C.gode_napi_error_info) C.int {
	if result == nil {
		return napiInvalidArg
	}
	*result = C.gode_last_error_info()
	return napiOK
}

//export napi_get_undefined
func napi_get_undefined(env C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || !setUint(result, rt.undefined()) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_get_null
func napi_get_null(env C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || !setUint(result, rt.null()) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_get_global
func napi_get_global(env C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || !setUint(result, valueID(rt.e.JS.Global().V())) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_get_boolean
func napi_get_boolean(env C.uintptr_t, value C.uint8_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || !setUint(result, valueID(rt.e.JS.Bool(value != 0))) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_create_object
func napi_create_object(env C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || !setUint(result, valueID(rt.e.JS.Obj().V())) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_create_array
func napi_create_array(env C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || !setUint(result, valueID(rt.e.JS.Arr(nil).V())) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_create_array_with_length
func napi_create_array_with_length(env C.uintptr_t, length C.size_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil {
		return napiInvalidArg
	}
	arr := rt.e.JS.Arr(nil)
	if length > 0 {
		arr.SetAt(int(length)-1, rt.e.JS.Undefined())
	}
	if !setUint(result, valueID(arr.V())) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_create_double
func napi_create_double(env C.uintptr_t, value C.double, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || !setUint(result, valueID(rt.e.JS.Num(float64(value)))) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_create_int32
func napi_create_int32(env C.uintptr_t, value C.int32_t, result *C.uintptr_t) C.int {
	return napi_create_double(env, C.double(value), result)
}

//export napi_create_uint32
func napi_create_uint32(env C.uintptr_t, value C.uint32_t, result *C.uintptr_t) C.int {
	return napi_create_double(env, C.double(value), result)
}

//export napi_create_int64
func napi_create_int64(env C.uintptr_t, value C.int64_t, result *C.uintptr_t) C.int {
	return napi_create_double(env, C.double(value), result)
}

//export napi_create_uint64
func napi_create_uint64(env C.uintptr_t, value C.uint64_t, result *C.uintptr_t) C.int {
	return napi_create_double(env, C.double(value), result)
}

//export napi_create_string_utf8
func napi_create_string_utf8(env C.uintptr_t, str *C.char, length C.size_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || str == nil || !setUint(result, valueID(rt.e.JS.Str(cStringN(str, length)))) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_create_string_latin1
func napi_create_string_latin1(env C.uintptr_t, str *C.char, length C.size_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || str == nil {
		return napiInvalidArg
	}
	b := []byte(cStringN(str, length))
	runes := make([]rune, len(b))
	for i, c := range b {
		runes[i] = rune(c)
	}
	if !setUint(result, valueID(rt.e.JS.Str(string(runes)))) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_create_string_utf16
func napi_create_string_utf16(env C.uintptr_t, str *C.uint16_t, length C.size_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || str == nil {
		return napiInvalidArg
	}
	units := unsafe.Slice((*uint16)(unsafe.Pointer(str)), int(length))
	if !setUint(result, valueID(rt.e.JS.Str(string(utf16.Decode(units))))) {
		return napiInvalidArg
	}
	return napiOK
}

func cStringN(p *C.char, n C.size_t) string {
	if p == nil {
		return ""
	}
	if uint64(n) > 1<<40 {
		return C.GoString(p)
	}
	if n == 0 {
		return ""
	}
	return C.GoStringN(p, C.int(n))
}

//export napi_get_value_double
func napi_get_value_double(env C.uintptr_t, value C.uintptr_t, result *C.double) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil || !v.IsNumber() {
		return napiNumberExpected
	}
	*result = C.double(v.Float())
	return napiOK
}

//export napi_get_value_int32
func napi_get_value_int32(env C.uintptr_t, value C.uintptr_t, result *C.int32_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil || !v.IsNumber() {
		return napiNumberExpected
	}
	*result = C.int32_t(v.Int())
	return napiOK
}

//export napi_get_value_uint32
func napi_get_value_uint32(env C.uintptr_t, value C.uintptr_t, result *C.uint32_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil || !v.IsNumber() {
		return napiNumberExpected
	}
	*result = C.uint32_t(v.Int())
	return napiOK
}

//export napi_get_value_int64
func napi_get_value_int64(env C.uintptr_t, value C.uintptr_t, result *C.int64_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil || !v.IsNumber() {
		return napiNumberExpected
	}
	*result = C.int64_t(v.Float())
	return napiOK
}

//export napi_get_value_uint64
func napi_get_value_uint64(env C.uintptr_t, value C.uintptr_t, result *C.uint64_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil || !v.IsNumber() {
		return napiNumberExpected
	}
	*result = C.uint64_t(v.Float())
	return napiOK
}

//export napi_get_value_bool
func napi_get_value_bool(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil || !v.IsBool() {
		return napiBooleanExpected
	}
	if v.Bool() {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

func copyBytes(buf *C.char, size C.size_t, b []byte) C.size_t {
	if buf == nil || size == 0 {
		return C.size_t(len(b))
	}
	n := len(b)
	if n >= int(size) {
		n = int(size) - 1
	}
	dst := unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(size))
	copy(dst, b[:n])
	dst[n] = 0
	return C.size_t(n)
}

//export napi_get_value_string_utf8
func napi_get_value_string_utf8(env C.uintptr_t, value C.uintptr_t, buf *C.char, size C.size_t, result *C.size_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || !v.IsString() {
		return napiStringExpected
	}
	n := copyBytes(buf, size, []byte(v.Str()))
	if result != nil {
		*result = n
	}
	return napiOK
}

//export napi_get_value_string_latin1
func napi_get_value_string_latin1(env C.uintptr_t, value C.uintptr_t, buf *C.char, size C.size_t, result *C.size_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || !v.IsString() {
		return napiStringExpected
	}
	s := v.Str()
	b := make([]byte, 0, len(s))
	for _, r := range s {
		b = append(b, byte(r))
	}
	n := copyBytes(buf, size, b)
	if result != nil {
		*result = n
	}
	return napiOK
}

//export napi_get_value_string_utf16
func napi_get_value_string_utf16(env C.uintptr_t, value C.uintptr_t, buf *C.uint16_t, size C.size_t, result *C.size_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || !v.IsString() {
		return napiStringExpected
	}
	units := utf16.Encode([]rune(v.Str()))
	n := len(units)
	if buf != nil && size > 0 {
		if n >= int(size) {
			n = int(size) - 1
		}
		dst := unsafe.Slice((*uint16)(unsafe.Pointer(buf)), int(size))
		copy(dst, units[:n])
		dst[n] = 0
	}
	if result != nil {
		*result = C.size_t(n)
	}
	return napiOK
}

//export napi_typeof
func napi_typeof(env C.uintptr_t, value C.uintptr_t, result *C.int32_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil {
		return napiInvalidArg
	}
	var t int32 = 0
	switch {
	case v.IsUndefined():
		t = 0
	case v.IsNull():
		t = 1
	case v.IsBool():
		t = 2
	case v.IsNumber():
		t = 3
	case v.IsString():
		t = 4
	case v.IsBigInt():
		t = 9
	case v.IsFunction():
		t = 7
	case v.IsObject():
		t = 6
	case v.SymbolString() != "":
		t = 5
	}
	*result = C.int32_t(t)
	return napiOK
}

//export napi_create_function
func napi_create_function(env C.uintptr_t, utf8name *C.char, _ C.size_t, cb C.uintptr_t, data C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || cb == 0 || !setUint(result, valueID(rt.callback(cString(utf8name), uintptr(cb), uintptr(data)))) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_get_cb_info
func napi_get_cb_info(_ C.uintptr_t, info C.uintptr_t, argc *C.size_t, argv *C.uintptr_t, thisArg *C.uintptr_t, data *C.uintptr_t) C.int {
	mu.Lock()
	ci := infos[uintptr(info)]
	mu.Unlock()
	if ci == nil {
		return napiInvalidArg
	}
	actual := len(ci.args)
	if argc != nil {
		max := int(*argc)
		*argc = C.size_t(actual)
		if argv != nil {
			n := actual
			if n > max {
				n = max
			}
			out := unsafe.Slice(argv, n)
			for i := 0; i < n; i++ {
				out[i] = C.uintptr_t(valueID(ci.args[i]))
			}
		}
	}
	if thisArg != nil {
		*thisArg = C.uintptr_t(valueID(ci.this))
	}
	if data != nil {
		*data = C.uintptr_t(ci.data)
	}
	return napiOK
}

//export napi_get_new_target
func napi_get_new_target(_ C.uintptr_t, info C.uintptr_t, result *C.uintptr_t) C.int {
	mu.Lock()
	ci := infos[uintptr(info)]
	mu.Unlock()
	if result == nil || ci == nil {
		return napiInvalidArg
	}
	if ci.newTarget.Valid() {
		*result = C.uintptr_t(valueID(ci.newTarget))
	} else {
		*result = 0
	}
	return napiOK
}

func (rt *runtime) call(recv, fn js.Value, args []js.Value) (js.Value, C.int) {
	if !fn.IsFunction() {
		return js.Value{}, napiFunctionExpected
	}
	var this js.Object
	if recv.Valid() && recv.IsObject() {
		this = recv.Object()
	}
	out, err := fn.Object().Call(this, args...)
	if err != nil {
		rt.setError(err)
		return js.Value{}, napiPendingException
	}
	return out, napiOK
}

//export napi_call_function
func napi_call_function(env C.uintptr_t, recv C.uintptr_t, fn C.uintptr_t, argc C.size_t, argv *C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil {
		return napiInvalidArg
	}
	fv, ok := valueOf(uintptr(fn))
	if !ok {
		return napiFunctionExpected
	}
	rv, _ := valueOf(uintptr(recv))
	args := make([]js.Value, 0, int(argc))
	if argc > 0 {
		if argv == nil {
			return napiInvalidArg
		}
		refs := unsafe.Slice(argv, int(argc))
		for _, ref := range refs {
			if v, ok := valueOf(uintptr(ref)); ok {
				args = append(args, v)
			} else {
				args = append(args, rt.e.JS.Undefined())
			}
		}
	}
	out, status := rt.call(rv, fv, args)
	if status == napiOK && result != nil {
		*result = C.uintptr_t(valueID(out))
	}
	return status
}

//export napi_new_instance
func napi_new_instance(env C.uintptr_t, ctor C.uintptr_t, argc C.size_t, argv *C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	cv, ok := valueOf(uintptr(ctor))
	if rt == nil || !ok || !cv.IsFunction() {
		return napiFunctionExpected
	}
	args := make([]js.Value, 0, int(argc))
	if argc > 0 {
		if argv == nil {
			return napiInvalidArg
		}
		for _, ref := range unsafe.Slice(argv, int(argc)) {
			if v, ok := valueOf(uintptr(ref)); ok {
				args = append(args, v)
			}
		}
	}
	reflect := rt.e.JS.Global().Get("Reflect").Object()
	construct := reflect.Get("construct")
	out, err := construct.Object().Call(reflect, cv, rt.e.JS.Arr(args).V())
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	if !setUint(result, valueID(out)) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_instanceof
func napi_instanceof(env C.uintptr_t, object C.uintptr_t, ctor C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	ov, ok1 := valueOf(uintptr(object))
	cv, ok2 := valueOf(uintptr(ctor))
	if rt == nil || !ok1 || !ok2 || result == nil || !cv.IsFunction() {
		return napiInvalidArg
	}
	if !rt.instanceofFn.Valid() {
		rt.instanceofFn, _ = rt.e.JS.Eval("(function (o, c) { return o instanceof c; })", "[napi:instanceof]")
	}
	out, err := rt.instanceofFn.Object().Call(js.Object{}, ov, cv)
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	if out.Bool() {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_strict_equals
func napi_strict_equals(env C.uintptr_t, a C.uintptr_t, b C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	av, ok1 := valueOf(uintptr(a))
	bv, ok2 := valueOf(uintptr(b))
	if rt == nil || !ok1 || !ok2 || result == nil {
		return napiInvalidArg
	}
	if av.Same(bv) {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

func objectOf(id uintptr) (js.Object, bool) {
	v, ok := valueOf(id)
	if !ok || !v.IsObject() {
		return js.Object{}, false
	}
	return v.Object(), true
}

//export napi_set_property
func napi_set_property(env C.uintptr_t, object C.uintptr_t, key C.uintptr_t, value C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	kv, ok2 := valueOf(uintptr(key))
	vv, ok3 := valueOf(uintptr(value))
	if rt == nil || !ok || !ok2 || !ok3 {
		return napiInvalidArg
	}
	obj.Set(kv.Str(), vv)
	return napiOK
}

//export napi_get_property
func napi_get_property(env C.uintptr_t, object C.uintptr_t, key C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	kv, ok2 := valueOf(uintptr(key))
	if rt == nil || !ok || !ok2 || !setUint(result, valueID(obj.Get(kv.Str()))) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_has_property
func napi_has_property(env C.uintptr_t, object C.uintptr_t, key C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	kv, ok2 := valueOf(uintptr(key))
	if rt == nil || !ok || !ok2 || result == nil {
		return napiInvalidArg
	}
	if obj.Has(kv.Str()) {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_has_own_property
func napi_has_own_property(env C.uintptr_t, object C.uintptr_t, key C.uintptr_t, result *C.uint8_t) C.int {
	return napi_has_property(env, object, key, result)
}

//export napi_delete_property
func napi_delete_property(env C.uintptr_t, object C.uintptr_t, key C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	kv, ok2 := valueOf(uintptr(key))
	if rt == nil || !ok || !ok2 {
		return napiInvalidArg
	}
	obj.Delete(kv.Str())
	if result != nil {
		*result = 1
	}
	return napiOK
}

//export napi_set_named_property
func napi_set_named_property(env C.uintptr_t, object C.uintptr_t, name *C.char, value C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	vv, ok2 := valueOf(uintptr(value))
	if rt == nil || !ok || !ok2 {
		return napiInvalidArg
	}
	obj.Set(cString(name), vv)
	return napiOK
}

//export napi_get_named_property
func napi_get_named_property(env C.uintptr_t, object C.uintptr_t, name *C.char, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	if rt == nil || !ok || !setUint(result, valueID(obj.Get(cString(name)))) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_has_named_property
func napi_has_named_property(env C.uintptr_t, object C.uintptr_t, name *C.char, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	if rt == nil || !ok || result == nil {
		return napiInvalidArg
	}
	if obj.Has(cString(name)) {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_set_element
func napi_set_element(env C.uintptr_t, object C.uintptr_t, index C.uint32_t, value C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	vv, ok2 := valueOf(uintptr(value))
	if rt == nil || !ok || !ok2 {
		return napiInvalidArg
	}
	obj.SetAt(int(index), vv)
	return napiOK
}

//export napi_get_element
func napi_get_element(env C.uintptr_t, object C.uintptr_t, index C.uint32_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	if rt == nil || !ok || !setUint(result, valueID(obj.At(int(index)))) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_has_element
func napi_has_element(env C.uintptr_t, object C.uintptr_t, index C.uint32_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	if rt == nil || !ok || result == nil {
		return napiInvalidArg
	}
	if obj.Has(fmt.Sprintf("%d", index)) {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_delete_element
func napi_delete_element(env C.uintptr_t, object C.uintptr_t, index C.uint32_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	if rt == nil || !ok {
		return napiInvalidArg
	}
	obj.Delete(fmt.Sprintf("%d", index))
	if result != nil {
		*result = 1
	}
	return napiOK
}

//export napi_is_array
func napi_is_array(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil {
		return napiInvalidArg
	}
	if v.IsArray() {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_get_array_length
func napi_get_array_length(env C.uintptr_t, value C.uintptr_t, result *C.uint32_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil || !v.IsArray() {
		return napiArrayExpected
	}
	*result = C.uint32_t(v.Object().Length())
	return napiOK
}

//export napi_get_prototype
func napi_get_prototype(env C.uintptr_t, object C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	if rt == nil || !ok || !setUint(result, valueID(obj.Proto().V())) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_get_property_names
func napi_get_property_names(env C.uintptr_t, object C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	if rt == nil || !ok {
		return napiInvalidArg
	}
	keys := obj.Keys()
	vals := make([]js.Value, len(keys))
	for i, key := range keys {
		vals[i] = rt.e.JS.Str(key)
	}
	if !setUint(result, valueID(rt.e.JS.Arr(vals).V())) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_get_all_property_names
func napi_get_all_property_names(env C.uintptr_t, object C.uintptr_t, keyMode C.int32_t, keyFilter C.int32_t, keyConversion C.int32_t, result *C.uintptr_t) C.int {
	return napi_get_property_names(env, object, result)
}

func (rt *runtime) defineValueProperty(obj js.Object, name string, value js.Value, attrs int32) {
	if !rt.defineValue.Valid() {
		rt.defineValue, _ = rt.e.JS.Eval(`(function (o, n, v, w, e, c) {
	Object.defineProperty(o, n, { value: v, writable: !!w, enumerable: !!e, configurable: !!c });
})`, "[napi:define]")
	}
	if !rt.defineValue.Valid() {
		return
	}
	b := func(bit int32) js.Value { return rt.e.JS.Bool(attrs&bit != 0) }
	_, _ = rt.defineValue.Object().Call(js.Object{}, obj.V(), rt.e.JS.Str(name), value, b(1), b(2), b(4))
}

func (rt *runtime) defineAccessorProperty(obj js.Object, name string, getter, setter js.Value, attrs int32) {
	if !rt.defineAcc.Valid() {
		rt.defineAcc, _ = rt.e.JS.Eval(`(function (o, n, g, s, e, c) {
	Object.defineProperty(o, n, { get: g, set: s, enumerable: !!e, configurable: !!c });
})`, "[napi:accessor]")
	}
	if !rt.defineAcc.Valid() {
		return
	}
	if !getter.Valid() {
		getter = rt.e.JS.Undefined()
	}
	if !setter.Valid() {
		setter = rt.e.JS.Undefined()
	}
	_, _ = rt.defineAcc.Object().Call(js.Object{}, obj.V(), rt.e.JS.Str(name), getter, setter,
		rt.e.JS.Bool(attrs&2 != 0), rt.e.JS.Bool(attrs&4 != 0))
}

// defineDescriptor applies one napi_property_descriptor to target. Methods and
// values become data properties; getter/setter pairs become accessors. The
// descriptor attributes map directly onto writable/enumerable/configurable.
func (rt *runtime) defineDescriptor(target js.Object, desc *C.gode_napi_property_descriptor) {
	name := cString(desc.utf8name)
	if desc.name != 0 {
		if nv, ok := valueOf(uintptr(desc.name)); ok {
			if s, ok := nv.StrOK(); ok {
				name = s
			}
		}
	}
	if name == "" {
		return
	}
	// napi_property_descriptor carries native C function pointers for method,
	// getter and setter (not napi_value handles). Each one becomes a real
	// JavaScript function bound to the descriptor's data pointer.
	if desc.method != 0 {
		fv := rt.callback(name, uintptr(desc.method), uintptr(desc.data))
		rt.defineValueProperty(target, name, fv, int32(desc.attributes))
		return
	}
	if desc.value != 0 {
		if vv, ok := valueOf(uintptr(desc.value)); ok {
			rt.defineValueProperty(target, name, vv, int32(desc.attributes))
		}
		return
	}
	if desc.getter != 0 || desc.setter != 0 {
		var g, s js.Value
		if desc.getter != 0 {
			g = rt.callback("get "+name, uintptr(desc.getter), uintptr(desc.data))
		}
		if desc.setter != 0 {
			s = rt.callback("set "+name, uintptr(desc.setter), uintptr(desc.data))
		}
		rt.defineAccessorProperty(target, name, g, s, int32(desc.attributes))
	}
}

//export napi_define_properties
func napi_define_properties(env C.uintptr_t, object C.uintptr_t, count C.size_t, props *C.gode_napi_property_descriptor) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	if rt == nil || !ok || count > 0 && props == nil {
		return napiInvalidArg
	}
	descs := unsafe.Slice(props, int(count))
	for i := range descs {
		rt.defineDescriptor(obj, &descs[i])
	}
	return napiOK
}

//export napi_create_error
func napi_create_error(env C.uintptr_t, code C.uintptr_t, msg C.uintptr_t, result *C.uintptr_t) C.int {
	return createError(env, code, msg, result, "Error")
}

func createError(env C.uintptr_t, code C.uintptr_t, msg C.uintptr_t, result *C.uintptr_t, name string) C.int {
	rt := runtimeOf(uintptr(env))
	mv, ok := valueOf(uintptr(msg))
	if rt == nil || !ok || !setUint(result, 0) {
		return napiInvalidArg
	}
	obj := rt.e.JS.Error(mv.Str())
	obj.SetStr("name", name)
	if cv, ok := valueOf(uintptr(code)); ok {
		obj.SetStr("code", cv.Str())
	}
	*result = C.uintptr_t(valueID(obj.V()))
	return napiOK
}

//export napi_create_type_error
func napi_create_type_error(env C.uintptr_t, code C.uintptr_t, msg C.uintptr_t, result *C.uintptr_t) C.int {
	return createError(env, code, msg, result, "TypeError")
}

//export napi_create_range_error
func napi_create_range_error(env C.uintptr_t, code C.uintptr_t, msg C.uintptr_t, result *C.uintptr_t) C.int {
	return createError(env, code, msg, result, "RangeError")
}

//export napi_throw
func napi_throw(env C.uintptr_t, error C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(error))
	if rt == nil || !ok {
		return napiInvalidArg
	}
	rt.setException(v)
	return napiOK
}

//export napi_throw_error
func napi_throw_error(env C.uintptr_t, code *C.char, msg *C.char) C.int {
	return throwError(env, code, msg, "Error")
}

//export napi_throw_type_error
func napi_throw_type_error(env C.uintptr_t, code *C.char, msg *C.char) C.int {
	return throwError(env, code, msg, "TypeError")
}

//export napi_throw_range_error
func napi_throw_range_error(env C.uintptr_t, code *C.char, msg *C.char) C.int {
	return throwError(env, code, msg, "RangeError")
}

func throwError(env C.uintptr_t, code *C.char, msg *C.char, name string) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil {
		return napiInvalidArg
	}
	obj := rt.e.JS.Error(cString(msg))
	obj.SetStr("name", name)
	if code != nil {
		obj.SetStr("code", cString(code))
	}
	rt.setException(obj.V())
	return napiOK
}

//export napi_is_error
func napi_is_error(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil {
		return napiInvalidArg
	}
	if v.IsError() {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_is_exception_pending
func napi_is_exception_pending(env C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	if rt.lastException.Valid() {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_get_and_clear_last_exception
func napi_get_and_clear_last_exception(env C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil {
		return napiInvalidArg
	}
	v := rt.takeException()
	if !setUint(result, valueID(v)) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_create_reference
func napi_create_reference(env C.uintptr_t, value C.uintptr_t, initial C.uint32_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil {
		return napiInvalidArg
	}
	v.Protect()
	mu.Lock()
	id := nextRef
	nextRef++
	refs[id] = &napiRef{rt: rt, value: v, count: uint32(initial)}
	mu.Unlock()
	*result = C.uintptr_t(id)
	return napiOK
}

//export napi_delete_reference
func napi_delete_reference(_ C.uintptr_t, ref C.uintptr_t) C.int {
	mu.Lock()
	r := refs[uintptr(ref)]
	delete(refs, uintptr(ref))
	mu.Unlock()
	if r != nil {
		r.value.Unprotect()
	}
	return napiOK
}

//export napi_reference_ref
func napi_reference_ref(_ C.uintptr_t, ref C.uintptr_t, result *C.uint32_t) C.int {
	mu.Lock()
	r := refs[uintptr(ref)]
	if r != nil {
		r.count++
	}
	mu.Unlock()
	if r == nil {
		return napiInvalidArg
	}
	if result != nil {
		*result = C.uint32_t(r.count)
	}
	return napiOK
}

//export napi_reference_unref
func napi_reference_unref(_ C.uintptr_t, ref C.uintptr_t, result *C.uint32_t) C.int {
	mu.Lock()
	r := refs[uintptr(ref)]
	if r != nil && r.count > 0 {
		r.count--
	}
	mu.Unlock()
	if r == nil {
		return napiInvalidArg
	}
	if result != nil {
		*result = C.uint32_t(r.count)
	}
	return napiOK
}

//export napi_get_reference_value
func napi_get_reference_value(_ C.uintptr_t, ref C.uintptr_t, result *C.uintptr_t) C.int {
	mu.Lock()
	r := refs[uintptr(ref)]
	mu.Unlock()
	if r == nil || !setUint(result, valueID(r.value)) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_open_handle_scope
func napi_open_handle_scope(env C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	rt.pushScope()
	mu.Lock()
	id := nextScope
	nextScope++
	handleScopes[id] = rt
	mu.Unlock()
	*result = C.uintptr_t(id)
	return napiOK
}

//export napi_close_handle_scope
func napi_close_handle_scope(_ C.uintptr_t, scope C.uintptr_t) C.int {
	mu.Lock()
	rt := handleScopes[uintptr(scope)]
	delete(handleScopes, uintptr(scope))
	mu.Unlock()
	if rt == nil {
		return napiInvalidArg
	}
	rt.popScope(0)
	return napiOK
}

//export napi_open_escapable_handle_scope
func napi_open_escapable_handle_scope(env C.uintptr_t, result *C.uintptr_t) C.int {
	return napi_open_handle_scope(env, result)
}

//export napi_close_escapable_handle_scope
func napi_close_escapable_handle_scope(env C.uintptr_t, scope C.uintptr_t) C.int {
	mu.Lock()
	rt := handleScopes[uintptr(scope)]
	esc := escaped[uintptr(scope)]
	delete(handleScopes, uintptr(scope))
	delete(escaped, uintptr(scope))
	mu.Unlock()
	if rt == nil {
		return napiInvalidArg
	}
	rt.popScope(esc)
	return napiOK
}

//export napi_escape_handle
func napi_escape_handle(_ C.uintptr_t, scope C.uintptr_t, escapee C.uintptr_t, result *C.uintptr_t) C.int {
	if result == nil {
		return napiInvalidArg
	}
	mu.Lock()
	if _, ok := handleScopes[uintptr(scope)]; !ok {
		mu.Unlock()
		return napiInvalidArg
	}
	if escaped[uintptr(scope)] != 0 {
		mu.Unlock()
		return 12 // napi_escape_called_twice
	}
	escaped[uintptr(scope)] = uintptr(escapee)
	mu.Unlock()
	*result = escapee
	return napiOK
}

//export napi_wrap
func napi_wrap(env C.uintptr_t, object C.uintptr_t, native C.uintptr_t, finalize C.uintptr_t, hint C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	obj, ok := objectOf(uintptr(object))
	if rt == nil || !ok {
		return napiInvalidArg
	}
	mu.Lock()
	id := nextWrap
	nextWrap++
	wraps[id] = &wrapInfo{rt: rt, native: uintptr(native), finalize: uintptr(finalize), hint: uintptr(hint)}
	mu.Unlock()
	obj.SetHidden("$godeNapiWrap", rt.e.JS.Num(float64(id)))
	if result != nil {
		*result = C.uintptr_t(id)
	}
	return napiOK
}

func wrapOf(obj js.Object) *wrapInfo {
	v := obj.Get("$godeNapiWrap")
	if !v.IsNumber() {
		return nil
	}
	mu.Lock()
	w := wraps[uintptr(v.Int())]
	mu.Unlock()
	return w
}

//export napi_unwrap
func napi_unwrap(_ C.uintptr_t, object C.uintptr_t, result *C.uintptr_t) C.int {
	obj, ok := objectOf(uintptr(object))
	if !ok || result == nil {
		return napiInvalidArg
	}
	w := wrapOf(obj)
	if w == nil {
		return napiInvalidArg
	}
	*result = C.uintptr_t(w.native)
	return napiOK
}

//export napi_remove_wrap
func napi_remove_wrap(_ C.uintptr_t, object C.uintptr_t, result *C.uintptr_t) C.int {
	obj, ok := objectOf(uintptr(object))
	if !ok {
		return napiInvalidArg
	}
	w := wrapOf(obj)
	if w == nil {
		return napiInvalidArg
	}
	mu.Lock()
	delete(wraps, uintptr(obj.Get("$godeNapiWrap").Int()))
	mu.Unlock()
	obj.Delete("$godeNapiWrap")
	if result != nil {
		*result = C.uintptr_t(w.native)
	}
	return napiOK
}

//export napi_create_external
func napi_create_external(env C.uintptr_t, data C.uintptr_t, finalize C.uintptr_t, hint C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil {
		return napiInvalidArg
	}
	obj := rt.e.JS.Obj()
	mu.Lock()
	id := nextWrap
	nextWrap++
	wraps[id] = &wrapInfo{rt: rt, native: uintptr(data), finalize: uintptr(finalize), hint: uintptr(hint)}
	mu.Unlock()
	obj.SetHidden("$godeNapiExternal", rt.e.JS.Num(float64(id)))
	if !setUint(result, valueID(obj.V())) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_get_value_external
func napi_get_value_external(_ C.uintptr_t, value C.uintptr_t, result *C.uintptr_t) C.int {
	obj, ok := objectOf(uintptr(value))
	if !ok || result == nil {
		return napiInvalidArg
	}
	v := obj.Get("$godeNapiExternal")
	if !v.IsNumber() {
		return napiInvalidArg
	}
	mu.Lock()
	w := wraps[uintptr(v.Int())]
	mu.Unlock()
	if w == nil || !setUint(result, w.native) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_create_promise
func napi_create_promise(env C.uintptr_t, resultDeferred *C.uintptr_t, resultPromise *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || resultDeferred == nil || resultPromise == nil {
		return napiInvalidArg
	}
	holder := rt.e.JS.Obj()
	var resolve, reject js.Value
	holder.SetFuncHidden("executor", func(_ js.Object, args []js.Value) (js.Value, error) {
		if len(args) > 0 {
			resolve = args[0]
		}
		if len(args) > 1 {
			reject = args[1]
		}
		return rt.e.JS.Undefined(), nil
	})
	executor := holder.Get("executor")
	promise, err := rt.construct("Promise", []js.Value{executor})
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	mu.Lock()
	id := nextDefer
	nextDefer++
	deferreds[id] = &deferred{rt: rt, resolve: resolve, reject: reject}
	mu.Unlock()
	*resultDeferred = C.uintptr_t(id)
	*resultPromise = C.uintptr_t(valueID(promise))
	return napiOK
}

//export napi_resolve_deferred
func napi_resolve_deferred(_ C.uintptr_t, dfd C.uintptr_t, value C.uintptr_t) C.int {
	mu.Lock()
	d := deferreds[uintptr(dfd)]
	delete(deferreds, uintptr(dfd))
	mu.Unlock()
	v, ok := valueOf(uintptr(value))
	if d == nil || !ok || !d.resolve.IsFunction() {
		return napiInvalidArg
	}
	if _, err := d.resolve.Object().Call(js.Object{}, v); err != nil {
		d.rt.setError(err)
		return napiPendingException
	}
	return napiOK
}

//export napi_reject_deferred
func napi_reject_deferred(_ C.uintptr_t, dfd C.uintptr_t, value C.uintptr_t) C.int {
	mu.Lock()
	d := deferreds[uintptr(dfd)]
	delete(deferreds, uintptr(dfd))
	mu.Unlock()
	v, ok := valueOf(uintptr(value))
	if d == nil || !ok || !d.reject.IsFunction() {
		return napiInvalidArg
	}
	if _, err := d.reject.Object().Call(js.Object{}, v); err != nil {
		d.rt.setError(err)
		return napiPendingException
	}
	return napiOK
}

//export napi_is_promise
func napi_is_promise(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil {
		return napiInvalidArg
	}
	if v.Tag() == "Promise" {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_run_script
func napi_run_script(env C.uintptr_t, script C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(script))
	if rt == nil || !ok || !v.IsString() {
		return napiStringExpected
	}
	out, err := rt.e.RunScript(v.Str(), "[napi:run_script]")
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	if !setUint(result, valueID(out)) {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_adjust_external_memory
func napi_adjust_external_memory(env C.uintptr_t, delta C.int64_t, result *C.int64_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	mu.Lock()
	rt.externalMemory += int64(delta)
	if rt.externalMemory < 0 {
		rt.externalMemory = 0
	}
	n := rt.externalMemory
	mu.Unlock()
	*result = C.int64_t(n)
	return napiOK
}

//export napi_create_date
func napi_create_date(env C.uintptr_t, value C.double, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	if !rt.dateCtor.Valid() {
		rt.dateCtor = rt.e.JS.Global().Get("Date")
	}
	reflect := rt.e.JS.Global().Get("Reflect").Object()
	out, err := reflect.Get("construct").Object().Call(reflect,
		rt.dateCtor, rt.e.JS.Arr([]js.Value{rt.e.JS.Num(float64(value))}).V())
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

//export napi_set_instance_data
func napi_set_instance_data(env C.uintptr_t, data C.uintptr_t, finalize C.uintptr_t, hint C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil {
		return napiInvalidArg
	}
	rt.instanceData = uintptr(data)
	return napiOK
}

//export napi_get_instance_data
func napi_get_instance_data(env C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	*result = C.uintptr_t(rt.instanceData)
	return napiOK
}

