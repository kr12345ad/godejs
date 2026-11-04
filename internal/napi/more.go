package napi

/*
#include "gode_napi.h"
*/
import "C"

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"sync"
	"unsafe"

	"godejs/internal/js"
)

type cleanupHook struct {
	fn  uintptr
	arg uintptr
}

type finalizer struct {
	fn   uintptr
	data uintptr
	hint uintptr
}

var (
	nextAsyncCtx uintptr = 1
	asyncContexts        = map[uintptr]*runtime{}

	nextTSFN uintptr = 1
	tsfns           = map[uintptr]*tsfn{}

	nextCleanupHandle uintptr = 1
	asyncCleanup               = map[uintptr]*cleanupHook{}
)

// shutdown runs the addon's teardown callbacks before the runtime drops its
// state. Node-API finalizers are tied to garbage collection; GodeJS has no
// JavaScriptCore finalizer hook, so they run here instead of never. That still
// gives addons a real place to free native memory.
func (rt *runtime) shutdown() {
	mu.Lock()
	hooks := rt.cleanupHooks
	rt.cleanupHooks = nil
	fins := rt.finalizers
	rt.finalizers = nil
	var wrapFins []wrapInfo
	for id, w := range wraps {
		if w.rt == rt {
			wrapFins = append(wrapFins, *w)
			delete(wraps, id)
		}
	}
	for id, v := range values {
		if v.Ctx() == rt.e.JS {
			delete(values, id)
		}
	}
	for id, r := range refs {
		if r.rt == rt {
			r.value.Unprotect()
			delete(refs, id)
		}
	}
	for id, d := range deferreds {
		if d.rt == rt {
			delete(deferreds, id)
		}
	}
	for id, w := range asyncWorks {
		if w.rt == rt {
			delete(asyncWorks, id)
		}
	}
	delete(runtimes, rt.id)
	delete(ctxRuntimes, rt.e.JS)
	mu.Unlock()

	for _, h := range hooks {
		if h != nil && h.fn != 0 {
			C.gode_call_cleanup(C.uintptr_t(h.fn), C.uintptr_t(h.arg))
		}
	}
	for _, f := range fins {
		if f.fn != 0 {
			C.gode_call_finalize(C.uintptr_t(f.fn), C.uintptr_t(rt.id), C.uintptr_t(f.data), C.uintptr_t(f.hint))
		}
	}
	for _, w := range wrapFins {
		if w.finalize != 0 {
			C.gode_call_finalize(C.uintptr_t(w.finalize), C.uintptr_t(rt.id), C.uintptr_t(w.native), C.uintptr_t(w.hint))
		}
	}
}

func (rt *runtime) construct(name string, args []js.Value) (js.Value, error) {
	ctor := rt.e.JS.Global().Get(name)
	reflect := rt.e.JS.Global().Get("Reflect").Object()
	return reflect.Get("construct").Object().Call(reflect, ctor, rt.e.JS.Arr(args).V())
}

func (rt *runtime) toString(v js.Value) string {
	if !v.Valid() {
		return ""
	}
	if v.IsString() {
		return v.Str()
	}
	out, err := rt.e.JS.Global().Get("String").Object().Call(js.Object{}, v)
	if err != nil {
		return ""
	}
	return out.Str()
}

// liveBytes returns a slice that aliases a buffer's backing store, so a native
// pointer handed to the addon reads and writes what JavaScript sees.
func (rt *runtime) liveBytes(buf js.Value) ([]byte, bool) {
	view, err := rt.construct("Uint8Array", []js.Value{buf})
	if err != nil || !view.Valid() {
		return nil, false
	}
	return view.ByteView()
}

func (rt *runtime) newBuffer(n int) js.Object {
	if b := rt.e.JS.Global().Get("Buffer"); b.IsObject() {
		if alloc := b.Object().Get("alloc"); alloc.IsFunction() {
			if out, err := alloc.Object().Call(b.Object(), rt.e.JS.Num(float64(n))); err == nil && out.IsObject() {
				return out.Object()
			}
		}
	}
	return rt.e.JS.Bytes(make([]byte, n))
}

func isArrayBuffer(v js.Value) bool {
	return v.Valid() && v.IsObject() && v.Tag() == "ArrayBuffer"
}

func isTypedArray(v js.Value) bool {
	if !v.Valid() || !v.IsObject() {
		return false
	}
	t := v.Tag()
	if t == "DataView" || t == "ArrayBuffer" {
		return false
	}
	switch t {
	case "Int8Array", "Uint8Array", "Uint8ClampedArray", "Int16Array", "Uint16Array",
		"Int32Array", "Uint32Array", "Float32Array", "Float64Array", "BigInt64Array", "BigUint64Array":
		return true
	}
	return false
}

func typedArrayType(tag string) (int32, int) {
	switch tag {
	case "Int8Array":
		return 0, 1
	case "Uint8Array":
		return 1, 1
	case "Uint8ClampedArray":
		return 2, 1
	case "Int16Array":
		return 3, 2
	case "Uint16Array":
		return 4, 2
	case "Int32Array":
		return 5, 4
	case "Uint32Array":
		return 6, 4
	case "Float32Array":
		return 7, 4
	case "Float64Array":
		return 8, 8
	case "BigInt64Array":
		return 9, 8
	case "BigUint64Array":
		return 10, 8
	}
	return -1, 0
}

func typedArrayName(t int32) string {
	switch t {
	case 0:
		return "Int8Array"
	case 1:
		return "Uint8Array"
	case 2:
		return "Uint8ClampedArray"
	case 3:
		return "Int16Array"
	case 4:
		return "Uint16Array"
	case 5:
		return "Int32Array"
	case 6:
		return "Uint32Array"
	case 7:
		return "Float32Array"
	case 8:
		return "Float64Array"
	case 9:
		return "BigInt64Array"
	case 10:
		return "BigUint64Array"
	}
	return ""
}

func ptrOf(b []byte) C.uintptr_t {
	if len(b) == 0 {
		return 0
	}
	return C.uintptr_t(uintptr(unsafe.Pointer(&b[0])))
}

func (rt *runtime) bigintFromString(s string) (js.Value, error) {
	ctor := rt.e.JS.Global().Get("BigInt")
	return ctor.Object().Call(js.Object{}, rt.e.JS.Str(s))
}

func (rt *runtime) bigintString(v js.Value) string {
	return rt.toString(v)
}

func bigIntWords(v js.Value) ([]uint64, int, bool) {
	s := v.Str()
	if s == "" {
		s = "0"
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		// Very large BigInts stringify in exponent form on some engines.
		if f, _, err := big.ParseFloat(s, 10, 256, big.ToNearestEven); err == nil {
			if i, acc := f.Int(nil); acc == big.Exact {
				n = i
				ok = true
			}
		}
	}
	if !ok {
		return nil, 0, false
	}
	sign := 0
	if n.Sign() < 0 {
		sign = 1
		n.Neg(n)
	}
	mask := new(big.Int).SetUint64(^uint64(0))
	var words []uint64
	for n.Sign() > 0 {
		w := new(big.Int).And(n, mask)
		words = append(words, w.Uint64())
		n.Rsh(n, 64)
	}
	if len(words) == 0 {
		words = []uint64{0}
	}
	return words, sign, true
}

// --- coercion -------------------------------------------------------------

func coerceTo(env C.uintptr_t, value C.uintptr_t, result *C.uintptr_t, ctor string) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok || result == nil {
		return napiInvalidArg
	}
	out, err := rt.e.JS.Global().Get(ctor).Object().Call(js.Object{}, v)
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

//export napi_coerce_to_bool
func napi_coerce_to_bool(env C.uintptr_t, value C.uintptr_t, result *C.uintptr_t) C.int {
	return coerceTo(env, value, result, "Boolean")
}

//export napi_coerce_to_number
func napi_coerce_to_number(env C.uintptr_t, value C.uintptr_t, result *C.uintptr_t) C.int {
	return coerceTo(env, value, result, "Number")
}

//export napi_coerce_to_object
func napi_coerce_to_object(env C.uintptr_t, value C.uintptr_t, result *C.uintptr_t) C.int {
	return coerceTo(env, value, result, "Object")
}

//export napi_coerce_to_string
func napi_coerce_to_string(env C.uintptr_t, value C.uintptr_t, result *C.uintptr_t) C.int {
	return coerceTo(env, value, result, "String")
}

// --- buffers and binary views --------------------------------------------

func isBuffer(v js.Value) bool {
	if !v.Valid() || !v.IsObject() {
		return false
	}
	if b := v.Ctx().Global().Get("Buffer"); b.IsObject() {
		if fn := b.Object().Get("isBuffer"); fn.IsFunction() {
			if out, err := fn.Object().Call(b.Object(), v); err == nil {
				return out.Bool()
			}
		}
	}
	return false
}

//export napi_is_buffer
func napi_is_buffer(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	_, ok := valueOf(uintptr(value))
	if runtimeOf(uintptr(env)) == nil || !ok || result == nil {
		return napiInvalidArg
	}
	if isBuffer(mustValue(value)) {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

func mustValue(id C.uintptr_t) js.Value {
	v, _ := valueOf(uintptr(id))
	return v
}

//export napi_is_typedarray
func napi_is_typedarray(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	if runtimeOf(uintptr(env)) == nil || result == nil {
		return napiInvalidArg
	}
	if isTypedArray(mustValue(value)) {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_is_dataview
func napi_is_dataview(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	if runtimeOf(uintptr(env)) == nil || result == nil {
		return napiInvalidArg
	}
	v := mustValue(value)
	if v.Valid() && v.IsObject() && v.Tag() == "DataView" {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_is_arraybuffer
func napi_is_arraybuffer(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	if runtimeOf(uintptr(env)) == nil || result == nil {
		return napiInvalidArg
	}
	if isArrayBuffer(mustValue(value)) {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_get_buffer_info
func napi_get_buffer_info(env C.uintptr_t, value C.uintptr_t, data *C.uintptr_t, length *C.size_t) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil || !isBuffer(v) {
		return napiInvalidArg
	}
	b, _ := v.ByteView()
	if data != nil {
		*data = ptrOf(b)
	}
	if length != nil {
		*length = C.size_t(len(b))
	}
	return napiOK
}

//export napi_create_buffer
func napi_create_buffer(env C.uintptr_t, length C.size_t, data *C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	obj := rt.newBuffer(int(length))
	if data != nil {
		b, _ := obj.Value.ByteView()
		*data = ptrOf(b)
	}
	*result = C.uintptr_t(valueID(obj.V()))
	return napiOK
}

//export napi_create_buffer_copy
func napi_create_buffer_copy(env C.uintptr_t, length C.size_t, src C.uintptr_t, resultData *C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	obj := rt.newBuffer(int(length))
	b, _ := obj.Value.ByteView()
	if int(length) > 0 && src != 0 {
		copy(b, unsafe.Slice((*byte)(unsafe.Pointer(uintptr(src))), int(length)))
	}
	if resultData != nil {
		*resultData = ptrOf(b)
	}
	*result = C.uintptr_t(valueID(obj.V()))
	return napiOK
}

//export napi_create_external_buffer
func napi_create_external_buffer(env C.uintptr_t, length C.size_t, src C.uintptr_t, finalize C.uintptr_t, hint C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	obj := rt.newBuffer(int(length))
	b, _ := obj.Value.ByteView()
	if int(length) > 0 && src != 0 {
		copy(b, unsafe.Slice((*byte)(unsafe.Pointer(uintptr(src))), int(length)))
	}
	if finalize != 0 {
		mu.Lock()
		rt.finalizers = append(rt.finalizers, finalizer{fn: uintptr(finalize), data: uintptr(src), hint: uintptr(hint)})
		mu.Unlock()
	}
	*result = C.uintptr_t(valueID(obj.V()))
	return napiOK
}

//export napi_get_arraybuffer_info
func napi_get_arraybuffer_info(env C.uintptr_t, value C.uintptr_t, data *C.uintptr_t, length *C.size_t) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil || !isArrayBuffer(v) {
		return napiArraybufferExpected
	}
	b, _ := rt.liveBytes(v)
	if data != nil {
		*data = ptrOf(b)
	}
	if length != nil {
		*length = C.size_t(len(b))
	}
	return napiOK
}

//export napi_create_arraybuffer
func napi_create_arraybuffer(env C.uintptr_t, length C.size_t, data *C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	out, err := rt.construct("ArrayBuffer", []js.Value{rt.e.JS.Num(float64(length))})
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	if data != nil {
		b, _ := rt.liveBytes(out)
		*data = ptrOf(b)
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

//export napi_create_external_arraybuffer
func napi_create_external_arraybuffer(env C.uintptr_t, external C.uintptr_t, length C.size_t, finalize C.uintptr_t, hint C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	out, err := rt.construct("ArrayBuffer", []js.Value{rt.e.JS.Num(float64(length))})
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	b, _ := rt.liveBytes(out)
	if int(length) > 0 && external != 0 {
		copy(b, unsafe.Slice((*byte)(unsafe.Pointer(uintptr(external))), int(length)))
	}
	if finalize != 0 {
		mu.Lock()
		rt.finalizers = append(rt.finalizers, finalizer{fn: uintptr(finalize), data: uintptr(external), hint: uintptr(hint)})
		mu.Unlock()
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

//export napi_detach_arraybuffer
func napi_detach_arraybuffer(env C.uintptr_t, value C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil {
		return napiInvalidArg
	}
	if !isArrayBuffer(v) {
		return napiArraybufferExpected
	}
	fn, err := rt.e.JS.Eval(`(function (b) {
	try { structuredClone(b, { transfer: [b] }); return true; } catch (e) { return false; }
})`, "[napi:detach]")
	if err != nil || !fn.IsFunction() {
		return napiDetachableArraybuffer
	}
	out, err := fn.Object().Call(js.Object{}, v)
	if err != nil || !out.Bool() {
		return napiDetachableArraybuffer
	}
	return napiOK
}

//export napi_is_detached_arraybuffer
func napi_is_detached_arraybuffer(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	if !isArrayBuffer(v) {
		return napiArraybufferExpected
	}
	d := v.Object().Get("detached")
	if d.IsBool() && d.Bool() {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

//export napi_get_typedarray_info
func napi_get_typedarray_info(env C.uintptr_t, value C.uintptr_t, typ *C.int32_t, length *C.size_t, data *C.uintptr_t, arraybuffer *C.uintptr_t, offset *C.size_t) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil || !isTypedArray(v) {
		return napiInvalidArg
	}
	t, _ := typedArrayType(v.Tag())
	obj := v.Object()
	b, _ := v.ByteView()
	if typ != nil {
		*typ = C.int32_t(t)
	}
	if length != nil {
		*length = C.size_t(obj.Get("length").Int())
	}
	if data != nil {
		*data = ptrOf(b)
	}
	if arraybuffer != nil {
		*arraybuffer = C.uintptr_t(valueID(obj.Get("buffer")))
	}
	if offset != nil {
		*offset = C.size_t(obj.Get("byteOffset").Int())
	}
	return napiOK
}

//export napi_create_typedarray
func napi_create_typedarray(env C.uintptr_t, typ C.int32_t, length C.size_t, arraybuffer C.uintptr_t, offset C.size_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	buf, ok := valueOf(uintptr(arraybuffer))
	if rt == nil || !ok || result == nil || !isArrayBuffer(buf) {
		return napiInvalidArg
	}
	name := typedArrayName(int32(typ))
	if name == "" {
		return napiInvalidArg
	}
	out, err := rt.construct(name, []js.Value{buf, rt.e.JS.Num(float64(offset)), rt.e.JS.Num(float64(length))})
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

//export napi_get_dataview_info
func napi_get_dataview_info(env C.uintptr_t, value C.uintptr_t, length *C.size_t, data *C.uintptr_t, arraybuffer *C.uintptr_t, offset *C.size_t) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil || !v.Valid() || !v.IsObject() || v.Tag() != "DataView" {
		return napiInvalidArg
	}
	b, _ := v.ByteView()
	if length != nil {
		*length = C.size_t(v.Object().Get("byteLength").Int())
	}
	if data != nil {
		*data = ptrOf(b)
	}
	if arraybuffer != nil {
		*arraybuffer = C.uintptr_t(valueID(v.Object().Get("buffer")))
	}
	if offset != nil {
		*offset = C.size_t(v.Object().Get("byteOffset").Int())
	}
	return napiOK
}

//export napi_create_dataview
func napi_create_dataview(env C.uintptr_t, length C.size_t, arraybuffer C.uintptr_t, offset C.size_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	buf, ok := valueOf(uintptr(arraybuffer))
	if rt == nil || !ok || result == nil || !isArrayBuffer(buf) {
		return napiInvalidArg
	}
	out, err := rt.construct("DataView", []js.Value{buf, rt.e.JS.Num(float64(offset)), rt.e.JS.Num(float64(length))})
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

// --- dates ----------------------------------------------------------------

//export napi_get_date_value
func napi_get_date_value(env C.uintptr_t, value C.uintptr_t, result *C.double) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil || result == nil || !v.Valid() || !v.IsDate() {
		return napiDateExpected
	}
	getTime := rt.e.JS.Global().Get("Date").Object().Get("prototype").Object().Get("getTime")
	out, err := getTime.Object().Call(v.Object())
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.double(out.Float())
	return napiOK
}

//export napi_is_date
func napi_is_date(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	if runtimeOf(uintptr(env)) == nil || result == nil {
		return napiInvalidArg
	}
	if mustValue(value).IsDate() {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

// --- BigInt ---------------------------------------------------------------

//export napi_create_bigint_int64
func napi_create_bigint_int64(env C.uintptr_t, value C.int64_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	out, err := rt.bigintFromString(strconv.FormatInt(int64(value), 10))
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

//export napi_create_bigint_uint64
func napi_create_bigint_uint64(env C.uintptr_t, value C.uint64_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	out, err := rt.bigintFromString(strconv.FormatUint(uint64(value), 10))
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

//export napi_create_bigint_words
func napi_create_bigint_words(env C.uintptr_t, sign C.int32_t, count C.size_t, words *C.uint64_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil || count > 0 && words == nil {
		return napiInvalidArg
	}
	n := new(big.Int)
	if count > 0 {
		raw := unsafe.Slice((*uint64)(unsafe.Pointer(words)), int(count))
		for i := len(raw) - 1; i >= 0; i-- {
			n.Lsh(n, 64)
			n.Or(n, new(big.Int).SetUint64(raw[i]))
		}
	}
	if sign != 0 {
		n.Neg(n)
	}
	out, err := rt.bigintFromString(n.String())
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

//export napi_get_value_bigint_int64
func napi_get_value_bigint_int64(env C.uintptr_t, value C.uintptr_t, result *C.int64_t, lossless *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil || result == nil || !v.Valid() || !v.IsBigInt() {
		return napiBigintExpected
	}
	n, ok := new(big.Int).SetString(rt.bigintString(v), 10)
	if !ok {
		return napiBigintExpected
	}
	lost := C.uint8_t(1)
	if !n.IsInt64() {
		lost = 0
		// Match Node's wrap-around to the low 64 bits.
		m := new(big.Int).And(n, new(big.Int).SetUint64(^uint64(0)))
		n = m
		if n.Bit(63) != 0 {
			n.Sub(n, new(big.Int).Lsh(big.NewInt(1), 64))
		}
	}
	*result = C.int64_t(n.Int64())
	if lossless != nil {
		*lossless = lost
	}
	return napiOK
}

//export napi_get_value_bigint_uint64
func napi_get_value_bigint_uint64(env C.uintptr_t, value C.uintptr_t, result *C.uint64_t, lossless *C.uint8_t) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil || result == nil || !v.Valid() || !v.IsBigInt() {
		return napiBigintExpected
	}
	n, ok := new(big.Int).SetString(rt.bigintString(v), 10)
	if !ok {
		return napiBigintExpected
	}
	lost := C.uint8_t(1)
	if !n.IsUint64() {
		lost = 0
		n = new(big.Int).And(n, new(big.Int).SetUint64(^uint64(0)))
	}
	*result = C.uint64_t(n.Uint64())
	if lossless != nil {
		*lossless = lost
	}
	return napiOK
}

//export napi_get_value_bigint_words
func napi_get_value_bigint_words(env C.uintptr_t, value C.uintptr_t, sign *C.int32_t, count *C.size_t, words *C.uint64_t) C.int {
	rt := runtimeOf(uintptr(env))
	v := mustValue(value)
	if rt == nil || count == nil || !v.Valid() || !v.IsBigInt() {
		return napiBigintExpected
	}
	list, s, ok := bigIntWords(v)
	if !ok {
		return napiBigintExpected
	}
	needed := len(list)
	have := int(*count)
	if words == nil {
		*count = C.size_t(needed)
		if sign != nil {
			*sign = C.int32_t(s)
		}
		return napiOK
	}
	n := have
	if n > needed {
		n = needed
	}
	dst := unsafe.Slice((*uint64)(unsafe.Pointer(words)), have)
	copy(dst, list[:n])
	for i := n; i < have; i++ {
		dst[i] = 0
	}
	*count = C.size_t(n)
	if sign != nil {
		*sign = C.int32_t(s)
	}
	return napiOK
}

//export napi_is_bigint
func napi_is_bigint(env C.uintptr_t, value C.uintptr_t, result *C.uint8_t) C.int {
	if runtimeOf(uintptr(env)) == nil || result == nil {
		return napiInvalidArg
	}
	if mustValue(value).IsBigInt() {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

// --- symbols and type tags -----------------------------------------------

//export napi_create_symbol
func napi_create_symbol(env C.uintptr_t, description C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	sym := rt.e.JS.Global().Get("Symbol")
	args := []js.Value{}
	if d, ok := valueOf(uintptr(description)); ok {
		args = append(args, rt.e.JS.Str(rt.toString(d)))
	}
	out, err := sym.Object().Call(js.Object{}, args...)
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

//export napi_symbol
func napi_symbol(env C.uintptr_t, description C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	d, _ := valueOf(uintptr(description))
	forFn := rt.e.JS.Global().Get("Symbol").Object().Get("for")
	out, err := forFn.Object().Call(js.Object{}, rt.e.JS.Str(rt.toString(d)))
	if err != nil {
		rt.setError(err)
		return napiPendingException
	}
	*result = C.uintptr_t(valueID(out))
	return napiOK
}

func tagString(tag *C.gode_napi_type_tag) string {
	return strconv.FormatUint(uint64(tag.lower), 16) + ":" + strconv.FormatUint(uint64(tag.upper), 16)
}

//export napi_type_tag_object
func napi_type_tag_object(env C.uintptr_t, object C.uintptr_t, tag *C.gode_napi_type_tag) C.int {
	obj, ok := objectOf(uintptr(object))
	if runtimeOf(uintptr(env)) == nil || !ok || tag == nil {
		return napiInvalidArg
	}
	obj.SetHidden("$godeNapiTag", obj.Ctx().Str(tagString(tag)))
	return napiOK
}

//export napi_check_object_type_tag
func napi_check_object_type_tag(env C.uintptr_t, object C.uintptr_t, tag *C.gode_napi_type_tag, result *C.uint8_t) C.int {
	obj, ok := objectOf(uintptr(object))
	if runtimeOf(uintptr(env)) == nil || !ok || tag == nil || result == nil {
		return napiInvalidArg
	}
	v := obj.Get("$godeNapiTag")
	if v.IsString() && v.Str() == tagString(tag) {
		*result = 1
	} else {
		*result = 0
	}
	return napiOK
}

// --- classes and object integrity ---------------------------------------

//export napi_define_class
func napi_define_class(env C.uintptr_t, utf8name *C.char, _ C.size_t, ctor C.uintptr_t, data C.uintptr_t, count C.size_t, props *C.gode_napi_property_descriptor, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || ctor == 0 || result == nil || count > 0 && props == nil {
		return napiInvalidArg
	}
	name := cString(utf8name)
	if name == "" {
		name = "NativeClass"
	}
	fn := rt.callback(name, uintptr(ctor), uintptr(data))
	if !fn.IsFunction() {
		return napiGenericFailure
	}
	proto := fn.Object().Get("prototype")
	descs := unsafe.Slice(props, int(count))
	for i := range descs {
		target := proto.Object()
		if descs[i].attributes&1024 != 0 {
			target = fn.Object()
		}
		rt.defineDescriptor(target, &descs[i])
	}
	*result = C.uintptr_t(valueID(fn))
	return napiOK
}

//export napi_object_freeze
func napi_object_freeze(env C.uintptr_t, object C.uintptr_t) C.int {
	obj, ok := objectOf(uintptr(object))
	rt := runtimeOf(uintptr(env))
	if rt == nil || !ok {
		return napiInvalidArg
	}
	if _, err := obj.ObjectCtor().Get("freeze").Object().Call(obj.ObjectCtor(), obj.V()); err != nil {
		rt.setError(err)
		return napiPendingException
	}
	return napiOK
}

//export napi_object_seal
func napi_object_seal(env C.uintptr_t, object C.uintptr_t) C.int {
	obj, ok := objectOf(uintptr(object))
	rt := runtimeOf(uintptr(env))
	if rt == nil || !ok {
		return napiInvalidArg
	}
	if _, err := obj.ObjectCtor().Get("seal").Object().Call(obj.ObjectCtor(), obj.V()); err != nil {
		rt.setError(err)
		return napiPendingException
	}
	return napiOK
}

// --- teardown hooks -------------------------------------------------------

//export napi_add_env_cleanup_hook
func napi_add_env_cleanup_hook(env C.uintptr_t, fn C.uintptr_t, arg C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || fn == 0 {
		return napiInvalidArg
	}
	mu.Lock()
	rt.cleanupHooks = append(rt.cleanupHooks, &cleanupHook{fn: uintptr(fn), arg: uintptr(arg)})
	mu.Unlock()
	return napiOK
}

//export napi_remove_env_cleanup_hook
func napi_remove_env_cleanup_hook(env C.uintptr_t, fn C.uintptr_t, arg C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil {
		return napiInvalidArg
	}
	mu.Lock()
	for i, h := range rt.cleanupHooks {
		if h.fn == uintptr(fn) && h.arg == uintptr(arg) {
			rt.cleanupHooks = append(rt.cleanupHooks[:i], rt.cleanupHooks[i+1:]...)
			break
		}
	}
	mu.Unlock()
	return napiOK
}

//export napi_add_async_cleanup_hook
func napi_add_async_cleanup_hook(env C.uintptr_t, fn C.uintptr_t, arg C.uintptr_t, remove *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || fn == 0 {
		return napiInvalidArg
	}
	h := &cleanupHook{fn: uintptr(fn), arg: uintptr(arg)}
	mu.Lock()
	rt.cleanupHooks = append(rt.cleanupHooks, h)
	handle := nextCleanupHandle
	nextCleanupHandle++
	asyncCleanup[handle] = h
	mu.Unlock()
	if remove != nil {
		*remove = C.uintptr_t(handle)
	}
	return napiOK
}

//export napi_remove_async_cleanup_hook
func napi_remove_async_cleanup_hook(handle C.uintptr_t) C.int {
	mu.Lock()
	h := asyncCleanup[uintptr(handle)]
	delete(asyncCleanup, uintptr(handle))
	if h != nil {
		for id, rt := range runtimes {
			_ = id
			for i, cur := range rt.cleanupHooks {
				if cur == h {
					rt.cleanupHooks = append(rt.cleanupHooks[:i], rt.cleanupHooks[i+1:]...)
					mu.Unlock()
					return napiOK
				}
			}
		}
	}
	mu.Unlock()
	return napiOK
}

//export napi_add_finalizer
func napi_add_finalizer(env C.uintptr_t, _ C.uintptr_t, data C.uintptr_t, finalize C.uintptr_t, hint C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || finalize == 0 {
		return napiInvalidArg
	}
	mu.Lock()
	rt.finalizers = append(rt.finalizers, finalizer{fn: uintptr(finalize), data: uintptr(data), hint: uintptr(hint)})
	mu.Unlock()
	if result != nil {
		*result = 0
	}
	return napiOK
}

//export napi_fatal_exception
func napi_fatal_exception(env C.uintptr_t, value C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	v, ok := valueOf(uintptr(value))
	if rt == nil || !ok {
		return napiInvalidArg
	}
	rt.setException(v)
	return napiOK
}

//export napi_fatal_error
func napi_fatal_error(location *C.char, locationLen C.size_t, message *C.char, messageLen C.size_t) C.int {
	where := cString(location)
	if where == "" {
		where = "GodeJS"
	}
	fmt.Fprintf(os.Stderr, "FATAL ERROR: %s %s\n", where, cString(message))
	os.Exit(1)
	return napiOK
}

// --- async contexts -------------------------------------------------------

//export napi_async_init
func napi_async_init(env C.uintptr_t, resource C.uintptr_t, name C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil {
		return napiInvalidArg
	}
	// Keep the resource and name alive for the lifetime of the context, as
	// Node does; addons pass them back through the async hooks API.
	if v, ok := valueOf(uintptr(resource)); ok {
		v.Protect()
	}
	if v, ok := valueOf(uintptr(name)); ok {
		v.Protect()
	}
	mu.Lock()
	id := nextAsyncCtx
	nextAsyncCtx++
	asyncContexts[id] = rt
	mu.Unlock()
	*result = C.uintptr_t(id)
	return napiOK
}

//export napi_async_destroy
func napi_async_destroy(env C.uintptr_t, ctx C.uintptr_t) C.int {
	if runtimeOf(uintptr(env)) == nil {
		return napiInvalidArg
	}
	mu.Lock()
	delete(asyncContexts, uintptr(ctx))
	mu.Unlock()
	return napiOK
}

//export napi_make_callback
func napi_make_callback(env C.uintptr_t, _ C.uintptr_t, recv C.uintptr_t, fn C.uintptr_t, argc C.size_t, argv *C.uintptr_t, result *C.uintptr_t) C.int {
	return napi_call_function(env, recv, fn, argc, argv, result)
}

//export napi_open_callback_scope
func napi_open_callback_scope(env C.uintptr_t, resource C.uintptr_t, asyncCtx C.uintptr_t, result *C.uintptr_t) C.int {
	if runtimeOf(uintptr(env)) == nil || result == nil {
		return napiInvalidArg
	}
	*result = 1
	return napiOK
}

//export napi_close_callback_scope
func napi_close_callback_scope(env C.uintptr_t, _ C.uintptr_t) C.int {
	if runtimeOf(uintptr(env)) == nil {
		return napiInvalidArg
	}
	return napiOK
}

// napi_get_uv_event_loop asks for the libuv loop that Node exposes. GodeJS has
// no libuv, so there is no pointer that could be handed out safely; callers get
// a failure instead of a dangling handle.
//
//export napi_get_uv_event_loop
func napi_get_uv_event_loop(_ C.uintptr_t, result *C.uintptr_t) C.int {
	if result != nil {
		*result = 0
	}
	return napiGenericFailure
}

// --- thread-safe functions ------------------------------------------------

type tsfn struct {
	rt           *runtime
	fn           js.Value
	callJS       uintptr
	context      uintptr
	finalize     uintptr
	finalizeData uintptr
	maxQueue     int
	threads      int
	envRef       bool
	closed       bool
	queue        []uintptr
	pending      []uintptr
	tick         js.Value

	mu   sync.Mutex
	cond *sync.Cond
}

func (t *tsfn) dispatch(data uintptr) {
	t.mu.Lock()
	t.pending = append(t.pending, data)
	t.mu.Unlock()
	// dispatch runs on a caller's thread; NextTick protects a JSC value.
	t.rt.e.EnqueueNative(func() { t.rt.e.NextTick(t.tick) })
}

//export napi_create_threadsafe_function
func napi_create_threadsafe_function(env C.uintptr_t, fn C.uintptr_t, asyncResource C.uintptr_t, asyncName C.uintptr_t, maxQueue C.size_t, initial C.size_t, finalizeData C.uintptr_t, finalize C.uintptr_t, context C.uintptr_t, callJS C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || result == nil || initial == 0 {
		return napiInvalidArg
	}
	fv, _ := valueOf(uintptr(fn))
	t := &tsfn{
		rt:           rt,
		fn:           fv,
		callJS:       uintptr(callJS),
		context:      uintptr(context),
		finalize:     uintptr(finalize),
		finalizeData: uintptr(finalizeData),
		maxQueue:     int(maxQueue),
		threads:      int(initial),
		envRef:       true,
	}
	if t.maxQueue == 0 {
		t.maxQueue = 1 << 20
	}
	if t.fn.Valid() {
		t.fn.Protect()
	}
	t.cond = sync.NewCond(&t.mu)

	holder := rt.e.JS.Obj()
	holder.SetFuncHidden("tick", func(_ js.Object, _ []js.Value) (js.Value, error) {
		rt.pushScope()
		defer rt.popScope(0)
		t.mu.Lock()
		if len(t.pending) == 0 {
			t.mu.Unlock()
			return rt.e.JS.Undefined(), nil
		}
		data := t.pending[0]
		t.pending = t.pending[1:]
		t.mu.Unlock()
		if t.callJS != 0 {
			var cb uintptr
			if t.fn.Valid() {
				cb = valueID(t.fn)
			}
			C.gode_call_tsfn(C.uintptr_t(t.callJS), C.uintptr_t(rt.id), C.uintptr_t(cb),
				C.uintptr_t(t.context), C.uintptr_t(data))
		} else if t.fn.IsFunction() {
			if _, err := t.fn.Object().Call(js.Object{}); err != nil {
				return js.Value{}, err
			}
		}
		return rt.e.JS.Undefined(), nil
	})
	t.tick = holder.Get("tick")
	if t.tick.Valid() {
		t.tick.Protect()
	}

	mu.Lock()
	id := nextTSFN
	nextTSFN++
	tsfns[id] = t
	mu.Unlock()

	rt.e.AddPendingWork()
	*result = C.uintptr_t(id)

	go func() {
		for {
			t.mu.Lock()
			for len(t.queue) == 0 && !t.closed {
				t.cond.Wait()
			}
			if len(t.queue) == 0 && t.closed {
				t.mu.Unlock()
				break
			}
			data := t.queue[0]
			t.queue = t.queue[1:]
			t.cond.Broadcast()
			t.mu.Unlock()
			t.dispatch(data)
		}
		if t.finalize != 0 {
			C.gode_call_finalize(C.uintptr_t(t.finalize), C.uintptr_t(rt.id),
				C.uintptr_t(t.finalizeData), C.uintptr_t(t.context))
		}
		if t.fn.Valid() {
			t.fn.Unprotect()
		}
		if t.tick.Valid() {
			t.tick.Unprotect()
		}
		if t.envRef {
			rt.e.DonePendingWork()
		}
		mu.Lock()
		delete(tsfns, id)
		mu.Unlock()
	}()
	return napiOK
}

//export napi_call_threadsafe_function
func napi_call_threadsafe_function(fn C.uintptr_t, data C.uintptr_t, mode C.int32_t) C.int {
	mu.Lock()
	t := tsfns[uintptr(fn)]
	mu.Unlock()
	if t == nil {
		return napiInvalidArg
	}
	t.mu.Lock()
	for !t.closed && len(t.queue) >= t.maxQueue {
		if mode == 0 {
			t.mu.Unlock()
			return 15 // napi_queue_full
		}
		t.cond.Wait()
	}
	if t.closed {
		t.mu.Unlock()
		return 16 // napi_closing
	}
	t.queue = append(t.queue, uintptr(data))
	t.cond.Signal()
	t.mu.Unlock()
	return napiOK
}

//export napi_get_threadsafe_function_context
func napi_get_threadsafe_function_context(fn C.uintptr_t, result *C.uintptr_t) C.int {
	mu.Lock()
	t := tsfns[uintptr(fn)]
	mu.Unlock()
	if t == nil || result == nil {
		return napiInvalidArg
	}
	*result = C.uintptr_t(t.context)
	return napiOK
}

//export napi_ref_threadsafe_function
func napi_ref_threadsafe_function(_ C.uintptr_t, fn C.uintptr_t) C.int {
	mu.Lock()
	t := tsfns[uintptr(fn)]
	mu.Unlock()
	if t == nil {
		return napiInvalidArg
	}
	t.mu.Lock()
	if !t.envRef {
		t.envRef = true
		t.rt.e.AddPendingWork()
	}
	t.mu.Unlock()
	return napiOK
}

//export napi_unref_threadsafe_function
func napi_unref_threadsafe_function(_ C.uintptr_t, fn C.uintptr_t) C.int {
	mu.Lock()
	t := tsfns[uintptr(fn)]
	mu.Unlock()
	if t == nil {
		return napiInvalidArg
	}
	t.mu.Lock()
	if t.envRef {
		t.envRef = false
		t.rt.e.DonePendingWork()
	}
	t.mu.Unlock()
	return napiOK
}

//export napi_release_threadsafe_function
func napi_release_threadsafe_function(fn C.uintptr_t, mode C.int32_t) C.int {
	mu.Lock()
	t := tsfns[uintptr(fn)]
	mu.Unlock()
	if t == nil {
		return napiInvalidArg
	}
	t.mu.Lock()
	if t.threads > 0 {
		t.threads--
	}
	if t.threads == 0 && !t.closed {
		t.closed = true
		t.cond.Broadcast()
	}
	t.mu.Unlock()
	return napiOK
}
