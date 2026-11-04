package util

import (
	"errors"
	"math"

	"godejs/internal/env"
	"godejs/internal/js"
)

func newPromise(c *js.Context, executor js.Value) (js.Value, error) {
	maker, err := c.Eval(`(function (executor) { return new Promise(executor); })`, "[gode:util]")
	if err != nil {
		return js.Value{}, err
	}
	return maker.Object().Call(js.Object{}, executor)
}

func promisify(original js.Value) (js.Value, error) {
	if !original.IsFunction() {
		return js.Value{}, errors.New(`The "original" argument must be of type function`)
	}
	c := original.Ctx()
	// Functions may carry their own promise-returning form under
	// Symbol.for('nodejs.util.promisify.custom'); returning it keeps
	// promisify(stream.pipeline) reference-equal to stream/promises.pipeline.
	if lookup, err := c.Eval(`(function (fn) {
  return fn[Symbol.for("nodejs.util.promisify.custom")];
})`, "[gode:util]"); err == nil && lookup.IsFunction() {
		custom, callErr := lookup.Object().Call(js.Object{}, original)
		if callErr == nil && custom.Valid() && !custom.IsUndefined() && !custom.IsNull() && !custom.IsFunction() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "util.promisify.custom" argument must be of type function. Received `+describe(custom))
		}
		if callErr == nil && custom.Valid() && custom.IsFunction() {
			// Node marks the custom function as its own promisified form, so
			// promisify(promisify(fn)) stays reference-equal.
			if tag, err := c.Eval(`(function (fn) {
  Object.defineProperty(fn, Symbol.for("nodejs.util.promisify.custom"), {
    __proto__: null, value: fn, enumerable: false, writable: false, configurable: true
  });
  return fn;
})`, "[gode:util]"); err == nil && tag.IsFunction() {
				if r, tagErr := tag.Object().Call(js.Object{}, custom); tagErr == nil {
					return r, nil
				}
			}
			return custom, nil
		}
	}
	// Symbol.for('nodejs.util.promisify.customArgs') names the callback's
	// success values, which promisify then resolves as a single object.
	var argNames js.Value
	if lookup, err := c.Eval(`(function (fn) {
  var names = fn[Symbol.for("nodejs.util.promisify.customArgs")];
  return Array.isArray(names) ? names : undefined;
})`, "[gode:util]"); err == nil && lookup.IsFunction() {
		if r, callErr := lookup.Object().Call(js.Object{}, original); callErr == nil {
			argNames = r.Protect()
		}
	}
	original.Protect()

	wrapper := makeFunc(c, "promisified", func(this js.Object, args []js.Value) (js.Value, error) {
		var resolve, reject js.Value
		executor := makeFunc(c, "executor", func(_ js.Object, a []js.Value) (js.Value, error) {
			if len(a) > 0 {
				resolve = a[0].Protect()
			}
			if len(a) > 1 {
				reject = a[1].Protect()
			}
			return c.Undefined(), nil
		})
		promise, err := newPromise(c, executor)
		if err != nil {
			return js.Value{}, err
		}

		cb := makeFunc(c, "callback", func(_ js.Object, cbArgs []js.Value) (js.Value, error) {
			var errVal js.Value
			if len(cbArgs) > 0 {
				errVal = cbArgs[0]
			}
			if errVal.Valid() && !errVal.IsUndefined() && !errVal.IsNull() {
				if reject.Valid() {
					_, callErr := reject.Object().Call(js.Object{}, errVal)
					return c.Undefined(), callErr
				}
				return c.Undefined(), nil
			}
			val := c.Undefined()
			names := argNames
			if names.Valid() && names.IsObject() && len(cbArgs) > 2 && names.Object().Length() > 0 {
				obj := c.Obj()
				for i := 0; i < names.Object().Length(); i++ {
					if i+1 < len(cbArgs) {
						obj.Set(names.Object().At(i).Str(), cbArgs[i+1])
					}
				}
				val = obj.V()
			} else if len(cbArgs) > 1 {
				val = cbArgs[1]
			}
			if resolve.Valid() {
				_, callErr := resolve.Object().Call(js.Object{}, val)
				return c.Undefined(), callErr
			}
			return c.Undefined(), nil
		})

		callArgs := append(append([]js.Value{}, args...), cb)
		result, callErr := original.Object().Call(this, callArgs...)
		if callErr != nil {
			if reject.Valid() {
				errObj := c.Error(callErr.Error())
				_, _ = reject.Object().Call(js.Object{}, errObj.V())
			}
		} else if result.Valid() && result.IsObject() {
			// Returning a Promise from a callback-style function is almost
			// always a mistake; Node warns with DEP0174.
			if then := result.Object().Get("then"); then.IsFunction() {
				if proc := c.Global().Get("process"); proc.Valid() && proc.IsObject() {
					if ew := proc.Object().Get("emitWarning"); ew.IsFunction() {
						_, _ = ew.Object().Call(proc.Object(),
							c.Str("Calling promisify on a function that returns a Promise is likely a mistake."),
							c.Str("DeprecationWarning"), c.Str("DEP0174"))
					}
				}
			}
		}
		return promise, nil
	})
	// Node keeps the original's prototype chain on the wrapper.
	if setProto, err := c.Eval(`(function (fn, proto) { Object.setPrototypeOf(fn, proto); })`, "[gode:util]"); err == nil && setProto.IsFunction() {
		if getProto, err2 := c.Eval(`(function (fn) { return Object.getPrototypeOf(fn); })`, "[gode:util]"); err2 == nil && getProto.IsFunction() {
			if proto, err3 := getProto.Object().Call(js.Object{}, original); err3 == nil {
				_, _ = setProto.Object().Call(js.Object{}, wrapper, proto)
			}
		}
	}
	return wrapper, nil
}

func callbackify(e *env.Env, original js.Value) (js.Value, error) {
	if !original.IsFunction() {
		return js.Value{}, errors.New(`The "original" argument must be of type function`)
	}
	c := original.Ctx()
	original.Protect()

	wrapper := makeFunc(c, "callbackified", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[len(args)-1].IsFunction() {
			got := "undefined"
			if len(args) > 0 {
				got = describe(args[len(args)-1])
			}
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The last argument must be of type function. Received `+got)
		}
		cb := args[len(args)-1].Protect()
		callArgs := args[:len(args)-1]
		// Both the callback and the receiver outlive this call: they are used
		// from a promise reaction, where JavaScriptCore no longer sees an
		// edge from the Go closure. Unprotect releases both once the callback
		// has been scheduled or invoked.
		receiver := this
		receiverVal := this.Value.Protect()
		release := func() {
			cb.Unprotect()
			receiverVal.Unprotect()
		}

		promise, err := original.Object().Call(receiver, callArgs...)
		if err != nil {
			e.NextTick(cb, c.Error(err.Error()).V())
			release()
			return c.Undefined(), nil
		}
		then := promise.Object().Get("then")
		if !then.IsFunction() {
			e.NextTick(cb, c.Error(`The "original" argument must return a Promise`).V())
			release()
			return c.Undefined(), nil
		}

		// Node binds the callback to the receiver the callbackified function was
		// called with, then runs it on nextTick.
		onFulfilled := makeFunc(c, "onFulfilled", func(_ js.Object, a []js.Value) (js.Value, error) {
			val := c.Undefined()
			if len(a) > 0 {
				val = a[0]
			}
			_, callErr := cb.Object().Call(receiver, c.Null(), val)
			release()
			return c.Undefined(), callErr
		})
		onRejected := makeFunc(c, "onRejected", func(_ js.Object, a []js.Value) (js.Value, error) {
			val := c.Undefined()
			if len(a) > 0 {
				val = a[0]
			}
			// A falsy rejection is wrapped, because a null error is how a
			// callback reports success. Truthy reasons are passed through
			// unchanged, exactly as Node's callbackifyOnRejected does.
			if falsyValue(val) {
				wrapped := c.Error("Promise was rejected with falsy value")
				wrapped.SetStr("code", "ERR_FALSY_VALUE_REJECTION")
				wrapped.Set("reason", val)
				val = wrapped.V()
			}
			_, callErr := cb.Object().Call(receiver, val)
			release()
			return c.Undefined(), callErr
		})
		_, _ = then.Object().Call(promise.Object(), onFulfilled, onRejected)
		return c.Undefined(), nil
	})
	copyFunctionMeta(c, wrapper, original)
	return wrapper, nil
}

// falsyValue mirrors JavaScript truthiness for the rejection reasons Node
// special-cases.
func falsyValue(v js.Value) bool {
	if !v.Valid() || v.IsUndefined() || v.IsNull() {
		return true
	}
	if v.IsBool() {
		return !v.Bool()
	}
	if v.IsNumber() {
		f := v.Float()
		return f == 0 || math.IsNaN(f)
	}
	if v.IsString() {
		return v.Str() == ""
	}
	if v.Tag() == "BigInt" {
		return jsToString(v) == "0"
	}
	return false
}

// copyFunctionMeta implements Node's descriptor copy for callbackify: a
// numeric length is incremented and a string name gains the Callbackified
// suffix, while manipulated (non-number/non-string) values are left alone.
func copyFunctionMeta(c *js.Context, wrapper, original js.Value) {
	objCtor := c.Global().Get("Object")
	define := objCtor.Object().Get("defineProperty")
	if !define.IsFunction() {
		return
	}
	set := func(prop string, value js.Value) {
		if !value.Valid() {
			return
		}
		desc := c.Obj()
		desc.Set("value", value)
		desc.SetBool("writable", false)
		desc.SetBool("enumerable", false)
		desc.SetBool("configurable", true)
		_, _ = define.Object().Call(objCtor.Object(), wrapper, c.Str(prop), desc.V())
	}
	length := original.Object().Get("length")
	if length.IsNumber() {
		set("length", c.Num(length.Float()+1))
	}
	name := original.Object().Get("name")
	if name.IsString() {
		set("name", c.Str(name.Str()+"Callbackified"))
	}
}
