package js

/*
#include "godejsc.h"
*/
import "C"

import "unsafe"

//export godeCallback
func godeCallback(ctx, function, thisObject C.uintptr_t, argc C.size_t,
	argv, exception C.uintptr_t) C.uintptr_t {

	undef := C.JSValueMakeUndefined(ctx)

	contextValue, ok := contexts.Load(uintptr(ctx))
	if !ok {
		return undef
	}
	c := contextValue.(*Context)

	id := propertyNumberRef(c, function, "$gode$id")
	hostMu.Lock()
	h, found := hosts[int(id)]
	hostMu.Unlock()
	if !found {
		return undef
	}

	args := make([]Value, 0, int(argc))
	if argc > 0 {
		raw := unsafe.Slice((*uintptr)(unsafe.Pointer(uintptr(argv))), int(argc))
		for _, r := range raw {
			args = append(args, Value{c: c, ref: r})
		}
	}

	var result Value
	var callErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					callErr = err
				} else if s, ok := r.(string); ok {
					callErr = &Error{Name: "Error", Message: s}
				} else {
					callErr = &Error{Name: "Error", Message: "internal error in host function"}
				}
			}
		}()
		result, callErr = h.fn(Object{Value{c: c, ref: uintptr(thisObject)}}, args)
	}()

	if callErr != nil {
		msg := callErr.Error()
		name := "Error"
		code := ""
		if jsErr, ok := callErr.(*Error); ok {
			if jsErr.Message != "" {
				msg = jsErr.Message
			}
			if jsErr.Name != "" {
				name = jsErr.Name
			}
			code = jsErr.Code
		}
		var ref uintptr
		if jsErr, ok := callErr.(*Error); ok && jsErr.ref != 0 && jsErr.ctx == c {
			// Rethrow the value JavaScript already threw. Allocating a new
			// error here can fail while that exception is still pending.
			ref = jsErr.ref
		} else {
			ref = c.makeErrorValue(msg)
			jsErr, _ := callErr.(*Error)
			var errno *int
			if jsErr != nil {
				errno = jsErr.Errno
			}
			if name != "Error" || code != "" || errno != nil {
				errObj := Object{Value{c: c, ref: ref}}
				c.setErrorMeta(errObj, name, code, errno)
				if jsErr != nil && jsErr.Input != "" {
					errObj.SetStr("input", jsErr.Input)
				}
			}
		}
		if exception != 0 {
			*(*uintptr)(unsafe.Pointer(uintptr(exception))) = ref
		}
		// A callback that throws returns NULL; returning undefined alongside
		// the exception slot lets some JavaScriptCore call paths complete as
		// if the call succeeded.
		return 0
	}
	if !result.Valid() {
		return undef
	}
	return C.uintptr_t(result.ref)
}

func propertyNumberRef(c *Context, obj C.uintptr_t, key string) float64 {
	js := c.newString(key)
	defer C.JSStringRelease(js)
	v := C.JSObjectGetProperty(u(c.ref), obj, js, 0)
	if v == 0 {
		return 0
	}
	return float64(C.JSValueToNumber(u(c.ref), v, 0))
}
