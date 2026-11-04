//go:build linux

package napi

/*
#include "gode_napi.h"
*/
import "C"

import (
	"sync"

	"godejs/internal/js"
)

type asyncWork struct {
	rt       *runtime
	resource uintptr
	execute  uintptr
	complete uintptr
	data     uintptr

	mu       sync.Mutex
	queued   bool
	running  bool
	canceled bool
	deleted  bool
}

// napi_create_async_work keeps the native callbacks in Go and returns an
// opaque handle to the addon. The execute callback runs off the JS thread;
// completion is posted through Env.NextTick so the complete callback always
// observes the JavaScript isolate on its owning thread.
//
//export napi_create_async_work
func napi_create_async_work(env C.uintptr_t, resource C.uintptr_t, _ *C.char, execute C.uintptr_t, complete C.uintptr_t, data C.uintptr_t, result *C.uintptr_t) C.int {
	rt := runtimeOf(uintptr(env))
	if rt == nil || execute == 0 || result == nil {
		return napiInvalidArg
	}
	mu.Lock()
	id := nextAsync
	nextAsync++
	asyncWorks[id] = &asyncWork{
		rt:       rt,
		resource: uintptr(resource),
		execute:  uintptr(execute),
		complete: uintptr(complete),
		data:     uintptr(data),
	}
	mu.Unlock()
	*result = C.uintptr_t(id)
	return napiOK
}

//export napi_delete_async_work
func napi_delete_async_work(_ C.uintptr_t, work C.uintptr_t) C.int {
	mu.Lock()
	w := asyncWorks[uintptr(work)]
	if w != nil {
		w.mu.Lock()
		w.deleted = true
		running := w.running
		w.mu.Unlock()
		if !running {
			delete(asyncWorks, uintptr(work))
		}
	}
	mu.Unlock()
	if w == nil {
		return napiInvalidArg
	}
	return napiOK
}

//export napi_queue_async_work
func napi_queue_async_work(_ C.uintptr_t, work C.uintptr_t) C.int {
	mu.Lock()
	w := asyncWorks[uintptr(work)]
	mu.Unlock()
	if w == nil {
		return napiInvalidArg
	}
	w.mu.Lock()
	if w.queued || w.running || w.deleted {
		w.mu.Unlock()
		return napiGenericFailure
	}
	w.queued = true
	w.running = true
	rt := w.rt
	execute := w.execute
	complete := w.complete
	data := w.data
	w.mu.Unlock()

	// The completion function is built here, on the JavaScript thread, and
	// only scheduled from the worker goroutine. Touching the JavaScriptCore
	// context from that goroutine would be a data race.
	holder := rt.e.JS.Obj()
	holder.SetFuncHidden("complete", func(_ js.Object, _ []js.Value) (js.Value, error) {
		status := napiOK
		w.mu.Lock()
		canceled := w.canceled
		w.mu.Unlock()
		if canceled {
			status = napiCancelled
		}
		if complete != 0 {
			C.gode_call_complete(C.uintptr_t(complete), C.uintptr_t(rt.id), C.int(status), C.uintptr_t(data))
		}
		w.mu.Lock()
		w.running = false
		w.queued = false
		deleted := w.deleted
		w.mu.Unlock()
		if deleted {
			mu.Lock()
			delete(asyncWorks, uintptr(work))
			mu.Unlock()
		}
		rt.e.DonePendingWork()
		return rt.e.JS.Undefined(), nil
	})
	completeFn := holder.Get("complete")
	// NextTick protects its callback, which is a JavaScriptCore call. Pin the
	// callback here, on the context thread, so the worker below can hand it
	// back without touching JSC itself.
	completeFn.Protect()

	rt.e.AddPendingWork()
	go func() {
		w.mu.Lock()
		canceled := w.canceled
		w.mu.Unlock()
		if !canceled {
			C.gode_call_execute(C.uintptr_t(execute), C.uintptr_t(rt.id), C.uintptr_t(data))
		}

		// Completion must run on the JavaScript thread. EnqueueNative hops
		// back to that thread first; only there is it safe to protect the
		// callback and queue the tick. The pending-work counter keeps RunMain
		// alive until it fires.
		rt.e.EnqueueNative(func() {
			rt.e.NextTick(completeFn)
			completeFn.Unprotect()
		})
	}()
	return napiOK
}

//export napi_cancel_async_work
func napi_cancel_async_work(_ C.uintptr_t, work C.uintptr_t) C.int {
	mu.Lock()
	w := asyncWorks[uintptr(work)]
	mu.Unlock()
	if w == nil {
		return napiInvalidArg
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		return napiGenericFailure
	}
	w.canceled = true
	return napiOK
}
