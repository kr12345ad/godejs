package js

/*
#include "godejsc.h"
*/
import "C"

import (
	"errors"
	"math"
	"strings"
	"unicode/utf16"
	"unsafe"
)

// Value is a JavaScript value owned by a Context.
type Value struct {
	c   *Context
	ref uintptr
}

// Object is a JavaScript object (also used for arrays and functions).
type Object struct {
	Value
}

// Throw builds an error that the host-function bridge throws as a named JS
// error (TypeError, RangeError, ...).
func Throw(name, message string) error { return &Error{Name: name, Message: message} }

// ThrowCode is Throw with a Node-style error code.
func ThrowCode(name, code, message string) error {
	return &Error{Name: name, Code: code, Message: message}
}

// ThrownValue wraps an existing JavaScript exception so a host callback
// rethrows that exact object. Rebuilding an error from its message would drop
// custom properties such as `code`, which Node-API addons rely on.
func ThrownValue(v Value) error {
	if !v.Valid() {
		return nil
	}
	msg := "JavaScript exception"
	if v.IsString() {
		msg = v.Str()
	} else if v.IsObject() {
		if m := v.Object().Get("message"); m.IsString() {
			msg = m.Str()
		}
	}
	v.Protect()
	return &Error{Name: "Error", Message: msg, ref: v.ref, ctx: v.c}
}

// CodedError lets an error from outside this package choose the JS error name
// and code it becomes when it crosses into JavaScript.
type CodedError interface {
	error
	JSErrorCode() (name, code string)
}

// Error is a JavaScript exception observed on the Go side.
type Error struct {
	Name    string
	Code    string
	Message string
	// Errno, when set, is attached to the thrown error object as its numeric
	// errno (Node's zlib and fs errors expose one alongside `code`).
	Errno *int
	// Input is Node's error.input for URL validation failures.
	Input  string
	Stack  string
	File   string
	Line   int
	Column int

	// ref and ctx retain the original thrown value when the error came from
	// JavaScript, so host callbacks can rethrow it instead of allocating a new
	// error while an exception is pending.
	ref uintptr
	ctx *Context
}

func (e *Error) Error() string {
	if e.Name != "" && !strings.HasPrefix(e.Message, e.Name+":") {
		return e.Name + ": " + e.Message
	}
	return e.Message
}

// Keep roots the value for the lifetime of the process and returns it. Module
// constructors, prototypes and helper closures are often reachable only from
// Go, so JavaScriptCore would otherwise be free to collect them; a collected
// ref later resolves to whatever object now occupies that address, which shows
// up as a native brand-check error ("Receiver of DataView method must be a
// DataView") or a wrong-receiver TypeError instead of a crash.
func (v Value) Keep() Value {
	v.Protect()
	return v
}

// Keep roots the object and returns it unchanged.
func (o Object) Keep() Object {
	o.Protect()
	return o
}

// Ctx returns the owning context.
func (v Value) Ctx() *Context { return v.c }

// Valid reports whether v holds a JavaScript reference.
func (v Value) Valid() bool { return v.c != nil && v.ref != 0 }

// Object returns the value as an object (no coercion is performed).
func (v Value) Object() Object { return Object{v} }

// Same reports whether two values are strictly equal (JavaScript ===).
func (v Value) Same(other Value) bool {
	if v.c == nil || other.c == nil {
		return v.ref == other.ref
	}
	return C.JSValueIsStrictEqual(u(v.c.ref), u(v.ref), u(other.ref)) != 0
}

// IsUndefined reports undefined.
func (v Value) IsUndefined() bool {
	return v.c != nil && C.JSValueIsUndefined(u(v.c.ref), u(v.ref)) != 0
}

// IsNull reports null.
func (v Value) IsNull() bool {
	return v.c != nil && C.JSValueIsNull(u(v.c.ref), u(v.ref)) != 0
}

// IsBool reports a boolean.
func (v Value) IsBool() bool {
	return v.c != nil && C.JSValueIsBoolean(u(v.c.ref), u(v.ref)) != 0
}

// IsNumber reports a number.
func (v Value) IsNumber() bool {
	return v.c != nil && C.JSValueIsNumber(u(v.c.ref), u(v.ref)) != 0
}

// IsString reports a string.
func (v Value) IsString() bool {
	return v.c != nil && C.JSValueIsString(u(v.c.ref), u(v.ref)) != 0
}

// IsObject reports an object.
func (v Value) IsObject() bool {
	return v.c != nil && C.JSValueIsObject(u(v.c.ref), u(v.ref)) != 0
}

// IsArray reports a JS Array.
func (v Value) IsArray() bool {
	return v.c != nil && C.JSValueIsArray(u(v.c.ref), u(v.ref)) != 0
}

// IsDate reports a JS Date.
func (v Value) IsDate() bool {
	return v.c != nil && C.JSValueIsDate(u(v.c.ref), u(v.ref)) != 0
}

// IsFunction reports a callable object.
// IsBigInt reports whether the value is a JavaScript bigint. JavaScriptCore
// classifies values with JSValueGetType, where kJSTypeBigInt is 7 (undefined 0,
// null 1, boolean 2, number 3, string 4, object 5, symbol 6, bigint 7).
func (v Value) IsBigInt() bool {
	if !v.Valid() {
		return false
	}
	return C.JSValueGetType(u(v.c.ref), u(v.ref)) == 7
}

func (v Value) IsFunction() bool {
	if v.c == nil || C.JSValueIsObject(u(v.c.ref), u(v.ref)) == 0 {
		return false
	}
	obj := C.JSValueToObject(u(v.c.ref), u(v.ref), 0)
	return obj != 0 && C.JSObjectIsFunction(u(v.c.ref), obj) != 0
}

// IsError reports a JS Error instance.
func (v Value) IsError() bool {
	if !v.IsObject() {
		return false
	}
	if name, ok := v.Object().Get("name").StrOK(); ok {
		return name == "Error" || strings.HasSuffix(name, "Error")
	}
	return false
}

// Bool coerces to a boolean following JS semantics.
func (v Value) Bool() bool {
	if v.c == nil {
		return false
	}
	return C.JSValueToBoolean(u(v.c.ref), u(v.ref)) != 0
}

// Float coerces to a number following JS semantics.
func (v Value) Float() float64 {
	if v.c == nil {
		return 0
	}
	return float64(C.JSValueToNumber(u(v.c.ref), u(v.ref), 0))
}

// Int coerces to an integer.
func (v Value) Int() int {
	f := v.Float()
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	return int(f)
}

// Str coerces to a string following JS semantics; symbols yield "".
func (v Value) Str() string {
	s, _ := v.StrOK()
	return s
}

// SymbolString renders a symbol the way String(symbol) does: `Symbol(desc)`,
// or `Symbol()` when it has no description. Implicit string conversion throws
// for symbols, so this reads the description property instead.
func (v Value) SymbolString() string {
	if v.c == nil || v.ref == 0 {
		return ""
	}
	obj := C.JSValueToObject(u(v.c.ref), u(v.ref), 0)
	if obj == 0 {
		return ""
	}
	d := v.c.get(obj, "description")
	if d == 0 || C.JSValueIsUndefined(u(v.c.ref), d) != 0 {
		return "Symbol()"
	}
	s := C.JSValueToStringCopy(u(v.c.ref), d, 0)
	if s == 0 {
		return "Symbol()"
	}
	defer C.JSStringRelease(s)
	return "Symbol(" + v.c.stringFrom(s) + ")"
}

// StrOK coerces to a string, reporting failure (for example on a Symbol).
func (v Value) StrOK() (string, bool) {
	if v.c == nil || v.ref == 0 {
		return "", false
	}
	s := C.JSValueToStringCopy(u(v.c.ref), u(v.ref), 0)
	if s == 0 {
		return "", false
	}
	defer C.JSStringRelease(s)
	return v.c.stringFrom(s), true
}

// Bytes returns the bytes of a Uint8Array (a copy). The second result is
// false for values that are not Uint8Arrays.
// Bytes returns the raw bytes of any ArrayBufferView (every TypedArray kind
// and DataView). Rather than mapping JavaScriptCore's version-dependent
// element-type enum, it reads buffer/byteOffset/byteLength and views the range
// as a Uint8Array, which is exact for every element width.
// ArrayBufferBytes copies the bytes of an ArrayBuffer without creating a
// Uint8Array view. A view returned through the C API stays alive inside the
// bridge and keeps the buffer from being detached, which breaks
// structuredClone(buffer, { transfer: [buffer] }). Detached or empty buffers
// yield an empty slice.
func (v Value) ArrayBufferBytes() ([]byte, bool) {
	if v.c == nil || !v.IsObject() {
		return nil, false
	}
	obj := C.JSValueToObject(u(v.c.ref), u(v.ref), 0)
	if obj == 0 {
		return nil, false
	}
	n := C.JSObjectGetArrayBufferByteLength(u(v.c.ref), obj, 0)
	ptr := C.JSObjectGetArrayBufferBytesPtr(u(v.c.ref), obj, 0)
	if n == 0 || ptr == 0 {
		return []byte{}, true
	}
	out := make([]byte, int(n))
	copy(out, unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), int(n)))
	return out, true
}

func (v Value) Bytes() ([]byte, bool) {
	// ByteView reads properties and builds a Uint8Array, and the copy below
	// runs while JavaScriptCore's collector may be working on another thread.
	// Pin the view so its backing store cannot be reclaimed mid-copy.
	v.Protect()
	defer v.Unprotect()
	live, ok := v.ByteView()
	if !ok {
		return nil, false
	}
	out := make([]byte, len(live))
	copy(out, live)
	return out, true
}

// ByteView returns the bytes of an ArrayBufferView as a slice that ALIASES the
// backing store. Writing through the slice changes what JavaScript sees, and a
// JavaScript write is visible through the slice. The second result is false
// for values that are not ArrayBufferViews.
//
// The slice stays valid while the view itself is alive; it must not be kept
// after the value can be collected, and it must not be resliced across calls
// whose ArrayBuffer may have been detached or grown.
func (v Value) ByteView() ([]byte, bool) {
	if v.c == nil || !v.IsObject() {
		return nil, false
	}
	obj := C.JSValueToObject(u(v.c.ref), u(v.ref), 0)
	if obj == 0 {
		return nil, false
	}
	buffer := v.c.get(obj, "buffer")
	lengthVal := v.c.get(obj, "byteLength")
	offsetVal := v.c.get(obj, "byteOffset")
	// The property values are raw references. Looking up the constructor and
	// building the Uint8Array below can collect an unprotected buffer.
	for _, ref := range []C.uintptr_t{buffer, lengthVal, offsetVal} {
		if ref != 0 {
			C.JSValueProtect(u(v.c.ref), ref)
			defer C.JSValueUnprotect(u(v.c.ref), ref)
		}
	}
	if buffer == 0 || lengthVal == 0 || C.JSValueIsObject(u(v.c.ref), buffer) == 0 ||
		C.JSValueIsNumber(u(v.c.ref), lengthVal) == 0 {
		return nil, false
	}
	length := int(C.JSValueToNumber(u(v.c.ref), lengthVal, 0))
	if length == 0 {
		return []byte{}, true
	}
	// Fast path: read the ArrayBuffer's backing store directly. Building a
	// Uint8Array through the constructor allocated a JavaScript object for
	// every response body copied across the bridge, which showed up directly
	// in HTTP throughput. The stored-length check also rejects detached
	// buffers, which fall through to the constructor path below.
	if bufObj := C.JSValueToObject(u(v.c.ref), buffer, 0); bufObj != 0 {
		stored := C.JSObjectGetArrayBufferByteLength(u(v.c.ref), bufObj, 0)
		base := C.JSObjectGetArrayBufferBytesPtr(u(v.c.ref), bufObj, 0)
		shift := 0
		if offsetVal != 0 && C.JSValueIsNumber(u(v.c.ref), offsetVal) != 0 {
			shift = int(C.JSValueToNumber(u(v.c.ref), offsetVal, 0))
		}
		if base != 0 && shift >= 0 && int(stored) >= shift+length {
			return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(base)+uintptr(shift))), length), true
		}
	}
	glob := C.JSContextGetGlobalObject(u(v.c.ref))
	ctor := v.c.get(glob, "Uint8Array")
	if ctor == 0 || C.JSObjectIsFunction(u(v.c.ref), ctor) == 0 {
		return nil, false
	}
	// Uint8Array(buffer, byteOffset, length) views the same bytes.
	args := []C.uintptr_t{buffer, offsetVal, lengthVal}
	ex := newExcBox()
	defer ex.release()
	ref := C.JSObjectCallAsConstructor(u(v.c.ref), ctor, 3,
		C.uintptr_t(uintptr(unsafe.Pointer(&args[0]))), ex.addr())
	if ref == 0 || ex.value() != 0 {
		return nil, false
	}
	view := C.JSValueToObject(u(v.c.ref), ref, 0)
	ptr := C.JSObjectGetTypedArrayBytesPtr(u(v.c.ref), view, 0)
	if ptr == 0 {
		return nil, false
	}
	// JavaScriptCore hands back the backing store of the typed array, not the
	// window the view selects, so a subarray/slice would otherwise read the
	// parent's first bytes. The view above is a Uint8Array, so its byteOffset
	// is already a byte count.
	shift := 0
	if offsetVal != 0 && C.JSValueIsNumber(u(v.c.ref), offsetVal) != 0 {
		shift = int(C.JSValueToNumber(u(v.c.ref), offsetVal, 0))
	}
	if shift < 0 {
		return nil, false
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr)+uintptr(shift))), length), true
}

// JSON serializes the value with JSON.stringify semantics.
func (v Value) JSON(indent int) (string, error) {
	if v.c == nil {
		return "", errors.New("invalid value")
	}
	ex := newExcBox()
	defer ex.release()
	s := C.JSValueCreateJSONString(u(v.c.ref), u(v.ref), C.uint(indent), ex.addr())
	if exception := ex.value(); exception != 0 {
		return "", v.c.errorFrom(exception)
	}
	if s == 0 {
		return "", nil
	}
	defer C.JSStringRelease(s)
	return v.c.stringFrom(s), nil
}

// Protect pins the value so the garbage collector cannot reclaim it.
func (v Value) Protect() Value {
	if v.Valid() {
		C.JSValueProtect(u(v.c.ref), u(v.ref))
	}
	return v
}

// Unprotect releases a pin taken by Protect.
func (v Value) Unprotect() {
	if v.Valid() {
		C.JSValueUnprotect(u(v.c.ref), u(v.ref))
	}
}

// Export converts a value into plain Go data (for tests and host APIs).
func (v Value) Export() any { return v.export(0) }

func (v Value) export(depth int) any {
	if v.c == nil || v.ref == 0 || depth > 64 {
		return nil
	}
	switch {
	case v.IsUndefined() || v.IsNull():
		return nil
	case v.IsBool():
		return v.Bool()
	case v.IsNumber():
		return v.Float()
	case v.IsString():
		return v.Str()
	case v.IsArray():
		o := v.Object()
		n := o.Length()
		out := make([]any, n)
		for i := 0; i < n; i++ {
			out[i] = o.At(i).export(depth + 1)
		}
		return out
	}
	if b, ok := v.Bytes(); ok {
		return b
	}
	if v.IsFunction() {
		return "[Function]"
	}
	o := v.Object()
	m := map[string]any{}
	for _, k := range o.Keys() {
		m[k] = o.Get(k).export(depth + 1)
	}
	return m
}

// V returns o as a plain Value.
func (o Object) V() Value { return o.Value }

// Ctx returns the owning context.
func (o Object) Ctx() *Context { return o.Value.c }

// Get reads a property.
func (o Object) Get(key string) Value {
	if o.c == nil || o.ref == 0 || !o.Value.IsObject() {
		return Value{}
	}
	return Value{c: o.c, ref: uintptr(o.c.get(u(o.ref), key))}
}

// Has reports whether the property exists (including inherited).
func (o Object) Has(key string) bool {
	if o.c == nil || o.ref == 0 || !o.Value.IsObject() {
		return false
	}
	js := o.c.newString(key)
	defer C.JSStringRelease(js)
	return C.JSObjectHasProperty(u(o.c.ref), u(o.ref), js) != 0
}

// Set writes an enumerable, writable, configurable property.
func (o Object) Set(key string, v Value) { o.setAttr(key, v, attrNone) }

// SetHidden writes a non-enumerable, non-configurable property.
func (o Object) SetHidden(key string, v Value) { o.setAttr(key, v, attrDontEnum|attrDontDel) }

// SetConst writes a non-enumerable read-only property.
func (o Object) SetConst(key string, v Value) { o.setAttr(key, v, attrReadOnly|attrDontEnum) }

func (o Object) setAttr(key string, v Value, attrs int) {
	if o.c == nil || o.ref == 0 || !o.Value.IsObject() || v.ref == 0 {
		return
	}
	o.c.set(u(o.ref), key, u(v.ref), attrs)
}

// SetStr sets a string property.
func (o Object) SetStr(key, s string) { o.Set(key, o.c.Str(s)) }

// SetNum sets a number property.
func (o Object) SetNum(key string, f float64) { o.Set(key, o.c.Num(f)) }

// SetBool sets a boolean property.
func (o Object) SetBool(key string, b bool) { o.Set(key, o.c.Bool(b)) }

// SetNull sets a null property.
func (o Object) SetNull(key string) { o.Set(key, o.c.Null()) }

// SetObj sets an object property.
func (o Object) SetObj(key string, v Object) { o.Set(key, v.Value) }

// SetFunc exposes a Go function as a JS function.
func (o Object) SetFunc(key string, fn Func) { o.Set(key, o.c.fnObject(key, fn)) }

// SetFuncHidden exposes a Go function as a non-enumerable JS function.
func (o Object) SetFuncHidden(key string, fn Func) {
	o.setAttr(key, o.c.fnObject(key, fn), attrDontEnum)
}

// Delete removes a property.
func (o Object) Delete(key string) {
	if o.c == nil {
		return
	}
	js := o.c.newString(key)
	defer C.JSStringRelease(js)
	C.JSObjectDeleteProperty(u(o.c.ref), u(o.ref), js, 0)
}

// Keys returns property names (including inherited enumerable ones).
func (o Object) Keys() []string {
	if o.c == nil || o.ref == 0 {
		return nil
	}
	arr := C.JSObjectCopyPropertyNames(u(o.c.ref), u(o.ref))
	if arr == 0 {
		return nil
	}
	defer C.JSPropertyNameArrayRelease(arr)
	n := int(C.JSPropertyNameArrayGetCount(arr))
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		s := C.JSPropertyNameArrayGetNameAtIndex(arr, C.size_t(i))
		if s != 0 {
			out = append(out, o.c.stringFrom(s))
		}
	}
	return out
}

// OwnKeys returns own property names (non-enumerable included).
func (o Object) OwnKeys() []string {
	v, err := o.ObjectCtor().Get("getOwnPropertyNames").Object().Call(Object{}, o.Value)
	if err != nil || !v.IsArray() {
		return nil
	}
	arr := v.Object()
	n := arr.Length()
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, arr.At(i).Str())
	}
	return out
}

// OwnEnumerableKeys returns the keys Object.keys would return.
func (o Object) OwnEnumerableKeys() []string {
	v, err := o.ObjectCtor().Get("keys").Object().Call(Object{}, o.Value)
	if err != nil || !v.IsArray() {
		return nil
	}
	arr := v.Object()
	n := arr.Length()
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, arr.At(i).Str())
	}
	return out
}

// ObjectCtor returns the realm's Object constructor.
func (o Object) ObjectCtor() Object {
	return o.c.Global().Get("Object").Object()
}

// Tag returns the value's built-in class tag, e.g. "Array", "Date", "Map",
// "Uint8Array", derived from Object.prototype.toString.
func (v Value) Tag() string {
	if v.c == nil {
		return ""
	}
	if v.IsArray() {
		return "Array"
	}
	if v.IsFunction() {
		return "Function"
	}
	fn := v.c.Global().Get("Object").Object().Get("prototype").Object().Get("toString")
	if !fn.IsFunction() {
		return ""
	}
	r, err := fn.Object().Call(Object{Value: v})
	if err != nil {
		return ""
	}
	tag := r.Str()
	tag = strings.TrimPrefix(tag, "[object ")
	tag = strings.TrimSuffix(tag, "]")
	return tag
}

// Length returns the value of the "length" property.
func (o Object) Length() int { return o.Get("length").Int() }

// At reads an indexed element.
func (o Object) At(i int) Value {
	if o.c == nil || o.ref == 0 {
		return Value{}
	}
	return Value{c: o.c, ref: uintptr(C.JSObjectGetPropertyAtIndex(u(o.c.ref), u(o.ref), C.uint(i), 0))}
}

// SetAt writes an indexed element.
func (o Object) SetAt(i int, v Value) {
	if o.c == nil || v.ref == 0 {
		return
	}
	C.JSObjectSetPropertyAtIndex(u(o.c.ref), u(o.ref), C.uint(i), u(v.ref), 0)
}

// IsNullPrototype reports whether Object.getPrototypeOf(o) is null.
func (o Object) IsNullPrototype() bool {
	fn := o.ObjectCtor().Get("getPrototypeOf")
	if !fn.IsFunction() {
		return false
	}
	v, err := fn.Object().Call(Object{}, o.Value)
	return err == nil && v.IsNull()
}

// Proto returns the prototype object.
func (o Object) Proto() Object {
	if o.c == nil {
		return Object{}
	}
	return Object{Value{c: o.c, ref: uintptr(C.JSObjectGetPrototype(u(o.c.ref), u(o.ref)))}}
}

// SetProto replaces the prototype.
func (o Object) SetProto(p Object) bool {
	if o.c == nil || p.c == nil {
		return false
	}
	return C.JSObjectSetPrototype(u(o.c.ref), u(o.ref), u(p.ref)) != 0
}

// Call invokes the object as a function.
func (o Object) Call(this Object, args ...Value) (Value, error) {
	if o.c == nil || !o.IsFunction() {
		return Value{}, &Error{Name: "TypeError", Message: "value is not a function"}
	}
	var argv C.uintptr_t
	if len(args) > 0 {
		// The call into JavaScriptCore is synchronous, so a stack array can
		// hold the argument vector. This avoids an allocation on every host
		// call, which is the hottest path between Go and JavaScript.
		var fixed [8]uintptr
		var refs []uintptr
		if len(args) <= len(fixed) {
			refs = fixed[:len(args)]
		} else {
			refs = make([]uintptr, len(args))
		}
		for i, a := range args {
			refs[i] = a.ref
		}
		argv = C.uintptr_t(uintptr(unsafe.Pointer(&refs[0])))
	}
	var thisRef C.uintptr_t
	if this.c != nil && this.ref != 0 {
		thisRef = u(this.ref)
	}
	ex := newExcBox()
	defer ex.release()
	ref := C.JSObjectCallAsFunction(u(o.c.ref), u(o.ref), thisRef,
		C.size_t(len(args)), argv, ex.addr())
	if exception := ex.value(); exception != 0 {
		return Value{}, o.c.errorFrom(exception)
	}
	return Value{c: o.c, ref: uintptr(ref)}, nil
}

// CallWith invokes the object with a Go slice of arguments.
func (o Object) CallWith(args []Value) (Value, error) { return o.Call(Object{}, args...) }

// IsArrayIndex reports whether s is a canonical array index.
func IsArrayIndex(s string) bool {
	if s == "" || len(s) > 10 {
		return false
	}
	if s == "0" {
		return true
	}
	if s[0] == '0' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func utf16Encode(s string) []uint16 { return utf16.Encode([]rune(s)) }

func utf16Decode(units []uint16) string { return string(utf16.Decode(units)) }
