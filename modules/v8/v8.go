package v8

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/internal/version"
)

func init() { env.Register("v8", Install) }

// Install builds the v8 module's exports object.
//
// This is deliberately a compatibility shim: GodeJS runs on JavaScriptCore, so
// there is no V8 heap or V8 serializer to reach for. The methods return
// correctly-shaped data so code that probes V8 keeps running; the serializer
// pair speaks a compact wire format of its own that still understands the
// value tags the official serializer tests feed it.
func Install(e *env.Env) js.Object {
	m := e.JS.Obj()

	// cachedDataVersionTag changes whenever flags are set, the way V8's own tag
	// folds the command-line flags into its value.
	var flagVersion uint64

	m.SetFunc("getHeapStatistics", func(this js.Object, args []js.Value) (js.Value, error) {
		return heapStats(e).object(e).V(), nil
	})
	m.SetFunc("getHeapSpaceStatistics", func(this js.Object, args []js.Value) (js.Value, error) {
		return heapSpaces(e).V(), nil
	})
	m.SetFunc("getHeapCodeStatistics", func(this js.Object, args []js.Value) (js.Value, error) {
		return heapCodeStats(e).V(), nil
	})
	m.SetFunc("getCppHeapStatistics", func(this js.Object, args []js.Value) (js.Value, error) {
		// JavaScriptCore exposes no C++ heap accounting.
		o := e.JS.Obj()
		o.SetNum("total_size", 0)
		o.SetNum("used_size", 0)
		return o.V(), nil
	})
	m.SetFunc("setFlagsFromString", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "flags" argument must be of type string. Received undefined`)
		}
		if !args[0].IsString() {
			// Node's invalidArgTypeHelper text carries a leading " Received ".
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "flags" argument must be of type string. Received `+describe(e, args[0]))
		}
		flagVersion++
		return e.JS.Undefined(), nil
	})
	m.SetFunc("cachedDataVersionTag", func(this js.Object, args []js.Value) (js.Value, error) {
		tag := cacheTag() ^ uint32(flagVersion*2654435761)
		tag &= 0x7fffffff
		if tag == 0 {
			tag = 1
		}
		return e.JS.Num(float64(tag)), nil
	})
	m.SetFunc("takeCoverage", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})
	m.SetFunc("stopCoverage", func(this js.Object, args []js.Value) (js.Value, error) {
		return e.JS.Undefined(), nil
	})

	// queryObjects() cannot walk a JavaScriptCore heap, so it reports an empty
	// result. It still validates its arguments and emits the experimental
	// warning Node's tests expect, which is the part callers can observe.
	warnedQuery := false
	m.SetFunc("queryObjects", func(this js.Object, args []js.Value) (js.Value, error) {
		if !warnedQuery {
			warnedQuery = true
			emitWarning(e, "v8.queryObjects() is an experimental feature and might change at any time",
				"ExperimentalWarning")
		}
		if len(args) == 0 || !args[0].IsFunction() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "ctor" argument must be of type function.`+describeArg(e, args, 0))
		}
		format := "count"
		if len(args) > 1 && !args[1].IsUndefined() {
			if !args[1].IsObject() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "options" argument must be of type object.`+describe(e, args[1]))
			}
			if f := args[1].Object().Get("format"); f.Valid() && !f.IsUndefined() {
				format = f.Str()
				if format != "count" && format != "summary" {
					return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_VALUE",
						fmt.Sprintf("The argument 'format' is invalid. Received '%s'", format))
				}
			}
		}
		if format == "summary" {
			return e.JS.Arr(nil).V(), nil
		}
		return e.JS.Num(0), nil
	})
	m.SetFunc("isStringOneByteRepresentation", func(this js.Object, args []js.Value) (js.Value, error) {
		if len(args) == 0 || !args[0].IsString() {
			return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
				`The "content" argument must be of type string.`+describeArg(e, args, 0))
		}
		return e.JS.Bool(oneByteString(args[0].Str())), nil
	})
	m.SetFunc("writeHeapSnapshot", func(this js.Object, args []js.Value) (js.Value, error) {
		name := fmt.Sprintf("Heap.%d.heapsnapshot", os.Getpid())
		if len(args) > 0 && !args[0].IsUndefined() {
			if !args[0].IsString() {
				return js.Value{}, js.ThrowCode("TypeError", "ERR_INVALID_ARG_TYPE",
					`The "filename" argument must be of type string.`+describe(e, args[0]))
			}
			name = args[0].Str()
		}
		path := name
		if !filepath.IsAbs(path) {
			path = filepath.Join(e.Cwd(), path)
		}
		payload := map[string]any{
			"snapshot": map[string]any{
				"engine": "JavaScriptCore",
				"node":   version.NodeVersion,
			},
			"heapStatistics": heapStats(e).plain(),
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return js.Value{}, err
		}
		data = append(data, '\n')
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return js.Value{}, err
		}
		return e.JS.Str(path), nil
	})

	// getHeapSnapshot returns a real readable stream when the stream module is
	// available, so code that pipes the snapshot keeps working.
	var snapFactory js.Value
	m.SetFunc("getHeapSnapshot", func(this js.Object, args []js.Value) (js.Value, error) {
		if !snapFactory.Valid() {
			snapFactory = heapSnapshotFactory(e)
			if snapFactory.Valid() {
				snapFactory.Protect()
			}
		}
		if !snapFactory.Valid() || !snapFactory.IsFunction() {
			return js.Value{}, fmt.Errorf("v8.getHeapSnapshot is unavailable")
		}
		return snapFactory.Object().Call(js.Object{})
	})

	// The serializer and profile APIs are written in JavaScript: cloning
	// arbitrary JS values needs JS itself, not the Go bridge.
	if api, err := e.JS.Eval(serializerPrelude, "[v8]"); err == nil && api.IsObject() {
		obj := api.Object()
		for _, key := range []string{
			"Serializer", "DefaultSerializer", "Deserializer", "DefaultDeserializer",
			"serialize", "deserialize",
		} {
			m.Set(key, obj.Get(key))
		}
	}
	if api, err := e.JS.Eval(extraPrelude, "[v8]"); err == nil && api.IsObject() {
		obj := api.Object()
		for _, key := range []string{
			"startupSnapshot", "isBuildingSnapshot",
			"addSerializeCallback", "addDeserializeCallback", "setDeserializeMainFunction",
			"createHook", "before", "after", "onBefore", "onAfter", "onInit",
			"onSettled", "settled", "init", "setHeapSnapshotNearHeapLimit",
		} {
			m.Set(key, obj.Get(key))
		}
	}

	// GCProfiler is a V8 coverage API. A tiny JS constructor is used because
	// JavaScriptCore does not reliably construct host functions, and this shim
	// only has to expose the documented shape.
	ctor, err := e.JS.Eval(`(function () {
	// JavaScriptCore has no GC observer, but global.gc is our own builtin, so a
	// profiler can record every explicit collection the program asks for. The
	// wrapper is installed for the profiling window only and restored on stop.
	function gcHeapStatistics() {
		return {
			externalMemory: 0, heapSizeLimit: 0, mallocedMemory: 0,
			peakMallocedMemory: 0, totalAvailableSize: 0, totalGlobalHandlesSize: 0,
			totalHeapSize: 0, totalHeapSizeExecutable: 0, totalPhysicalSize: 0,
			usedGlobalHandlesSize: 0, usedHeapSize: 0
		};
	}
	function gcHeapSpaces() {
		return [{
			physicalSpaceSize: 0, spaceAvailableSize: 0, spaceName: 'new_space',
			spaceSize: 0, spaceUsedSize: 0
		}];
	}
	function gcSnapshot() {
		return { heapStatistics: gcHeapStatistics(), heapSpaceStatistics: gcHeapSpaces() };
	}
	function GCProfiler() {
		this.__startTime = 0;
		this.__stopped = false;
		this.__statistics = [];
		this.__originalGC = null;
	}
	GCProfiler.prototype.start = function () {
		this.__startTime = Date.now();
		this.__stopped = false;
		this.__statistics = [];
		this.__originalGC = null;
		var self = this;
		var original = globalThis.gc;
		if (typeof original === 'function' && !original.__godeGCProfiler) {
			var wrapper = function (options) {
				var type = options && options.type;
				self.__statistics.push({
					gcType: type === 'minor' ? 'MinorMarkSweep' : 'MajorGC',
					cost: 0,
					beforeGC: gcSnapshot(),
					afterGC: gcSnapshot()
				});
				return original.apply(this, arguments);
			};
			wrapper.__godeGCProfiler = true;
			globalThis.gc = wrapper;
			this.__originalGC = original;
		}
	};
	GCProfiler.prototype.stop = function () {
		if (this.__stopped) return undefined;
		this.__stopped = true;
		if (this.__originalGC !== null && globalThis.gc && globalThis.gc.__godeGCProfiler) {
			globalThis.gc = this.__originalGC;
		}
		this.__originalGC = null;
		return {
			version: 1,
			startTime: this.__startTime,
			endTime: Date.now(),
			statistics: this.__statistics
		};
	};
	if (typeof Symbol === 'function' && typeof Symbol.dispose === 'symbol') {
		GCProfiler.prototype[Symbol.dispose] = function () { this.stop(); };
	}
	return GCProfiler;
})()`, "[v8]")
	if err == nil {
		m.Set("GCProfiler", ctor)
	}

	return m
}

func emitWarning(e *env.Env, message, name string) {
	var process js.Object
	func() {
		defer func() { _ = recover() }()
		process = e.Exports("process")
	}()
	if !process.Valid() {
		return
	}
	warn := process.Get("emitWarning")
	if !warn.Valid() || !warn.IsFunction() {
		return
	}
	_, _ = warn.Object().Call(js.Object{}, e.JS.Str(message), e.JS.Str(name))
}

// heapSnapshotFactory builds getHeapSnapshot() on top of the stream module.
// The factory is created once and reused; without a stream module it falls
// back to a minimal readable-shaped object so the call still succeeds.
func heapSnapshotFactory(e *env.Env) js.Value {
	const payload = `{"snapshot":{"meta":{},"node_count":0,"edge_count":0},"nodes":[],"edges":[],"strings":[]}`
	factory, err := e.JS.Eval(`(function (Readable, snapshot) {
	return function getHeapSnapshot() {
		var sent = false;
		return new Readable({
			read: function () {
				if (sent) { this.push(null); return; }
				sent = true;
				this.push(snapshot);
			}
		});
	};
})`, "[v8]")
	if err == nil && factory.IsFunction() {
		var readable js.Value
		func() {
			defer func() { _ = recover() }()
			readable = e.Exports("stream").Get("Readable")
		}()
		if readable.Valid() && readable.IsFunction() {
			if v, callErr := factory.Object().Call(js.Object{}, readable, e.JS.Str(payload)); callErr == nil {
				return v
			}
		}
	}
	fb, err := e.JS.Eval(`(function () {
	return function getHeapSnapshot() {
		return {
			on: function () { return this; },
			once: function () { return this; },
			pipe: function () { return this; },
			resume: function () { return this; },
			pause: function () { return this; },
			destroy: function () {},
			read: function () { return null; },
			push: function () {}
		};
	};
})`, "[v8]")
	if err != nil || !fb.IsFunction() {
		return js.Value{}
	}
	v, err := fb.Object().Call(js.Object{})
	if err != nil {
		return js.Value{}
	}
	return v
}

type stats struct {
	TotalHeapSize            float64
	TotalHeapSizeExecutable  float64
	TotalPhysicalSize        float64
	TotalAvailableSize       float64
	UsedHeapSize             float64
	HeapSizeLimit            float64
	MallocedMemory           float64
	PeakMallocedMemory       float64
	DoesZapGarbage           float64
	NumberOfNativeContexts   float64
	NumberOfDetachedContexts float64
	TotalGlobalHandlesSize   float64
	UsedGlobalHandlesSize    float64
	ExternalMemory           float64
}

func heapStats(e *env.Env) stats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	const limit = 4 * 1024 * 1024 * 1024
	used := float64(m.HeapAlloc)
	available := float64(limit) - used
	if available < 0 {
		available = 0
	}
	malloced := float64(m.Sys) - float64(m.HeapSys)
	if malloced < 0 {
		malloced = 0
	}
	return stats{
		TotalHeapSize:            float64(m.HeapSys),
		TotalHeapSizeExecutable:  0,
		TotalPhysicalSize:        float64(m.HeapSys),
		TotalAvailableSize:       available,
		UsedHeapSize:             used,
		HeapSizeLimit:            limit,
		MallocedMemory:           malloced,
		PeakMallocedMemory:       0,
		DoesZapGarbage:           0,
		NumberOfNativeContexts:   1,
		NumberOfDetachedContexts: 0,
		TotalGlobalHandlesSize:   0,
		UsedGlobalHandlesSize:    0,
		ExternalMemory:           float64(m.OtherSys),
	}
}

func (s stats) object(e *env.Env) js.Object {
	o := e.JS.Obj()
	o.SetNum("total_heap_size", s.TotalHeapSize)
	o.SetNum("total_heap_size_executable", s.TotalHeapSizeExecutable)
	o.SetNum("total_physical_size", s.TotalPhysicalSize)
	o.SetNum("total_available_size", s.TotalAvailableSize)
	o.SetNum("used_heap_size", s.UsedHeapSize)
	o.SetNum("heap_size_limit", s.HeapSizeLimit)
	o.SetNum("malloced_memory", s.MallocedMemory)
	o.SetNum("peak_malloced_memory", s.PeakMallocedMemory)
	o.SetNum("does_zap_garbage", s.DoesZapGarbage)
	o.SetNum("number_of_native_contexts", s.NumberOfNativeContexts)
	o.SetNum("number_of_detached_contexts", s.NumberOfDetachedContexts)
	o.SetNum("total_global_handles_size", s.TotalGlobalHandlesSize)
	o.SetNum("used_global_handles_size", s.UsedGlobalHandlesSize)
	o.SetNum("external_memory", s.ExternalMemory)
	return o
}

func (s stats) plain() map[string]any {
	return map[string]any{
		"total_heap_size":             s.TotalHeapSize,
		"total_heap_size_executable":  s.TotalHeapSizeExecutable,
		"total_physical_size":         s.TotalPhysicalSize,
		"total_available_size":        s.TotalAvailableSize,
		"used_heap_size":              s.UsedHeapSize,
		"heap_size_limit":             s.HeapSizeLimit,
		"malloced_memory":             s.MallocedMemory,
		"peak_malloced_memory":        s.PeakMallocedMemory,
		"does_zap_garbage":            s.DoesZapGarbage,
		"number_of_native_contexts":   s.NumberOfNativeContexts,
		"number_of_detached_contexts": s.NumberOfDetachedContexts,
		"total_global_handles_size":   s.TotalGlobalHandlesSize,
		"used_global_handles_size":    s.UsedGlobalHandlesSize,
		"external_memory":             s.ExternalMemory,
	}
}

// heapSpaces reports the space names Node 24 exposes. Only the sizes are
// Go-runtime estimates; the inventory is what callers and tests rely on.
func heapSpaces(e *env.Env) js.Object {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	total := float64(m.HeapSys)
	used := float64(m.HeapAlloc)
	type space struct {
		name      string
		size      float64
		used      float64
		available float64
	}
	spaces := []space{
		{"new_space", total / 4, 0, total / 4},
		{"old_space", total, used, total - used},
		{"code_space", 0, 0, 0},
		{"large_object_space", 0, 0, 0},
		{"read_only_space", 0, 0, 0},
		{"new_large_object_space", 0, 0, 0},
		{"code_large_object_space", 0, 0, 0},
		{"trusted_space", 0, 0, 0},
		{"trusted_large_object_space", 0, 0, 0},
		{"shared_space", 0, 0, 0},
		{"shared_large_object_space", 0, 0, 0},
		{"shared_trusted_space", 0, 0, 0},
		{"shared_trusted_large_object_space", 0, 0, 0},
	}
	items := make([]js.Value, 0, len(spaces))
	for _, s := range spaces {
		o := e.JS.Obj()
		o.SetStr("space_name", s.name)
		o.SetNum("space_size", s.size)
		o.SetNum("space_used_size", s.used)
		o.SetNum("space_available_size", s.available)
		o.SetNum("physical_space_size", s.size)
		items = append(items, o.V())
	}
	return e.JS.Arr(items)
}

func heapCodeStats(e *env.Env) js.Object {
	o := e.JS.Obj()
	o.SetNum("code_and_metadata_size", 0)
	o.SetNum("bytecode_and_metadata_size", 0)
	o.SetNum("external_script_source_size", 0)
	o.SetNum("cpu_profiler_metadata_size", 0)
	return o
}

func cacheTag() uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(version.NodeVersion + "+javascriptcore+" + version.Gode))
	return h.Sum32() & 0x7fffffff
}

func oneByteString(s string) bool {
	for _, r := range s {
		if r > 0xFF {
			return false
		}
	}
	return true
}

// describe renders a value the way Node's invalidArgTypeHelper does, e.g.
// "type number (1)" or "undefined".
func describe(e *env.Env, v js.Value) string {
	switch {
	case !v.Valid() || v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		if v.Bool() {
			return "type boolean (true)"
		}
		return "type boolean (false)"
	case v.IsNumber():
		return "type number (" + formatNumber(v.Float()) + ")"
	case v.IsBigInt():
		return "type bigint"
	case v.IsString():
		return "type string ('" + v.Str() + "')"
	case v.IsFunction():
		return "function"
	case v.IsObject():
		if ctor := v.Object().Get("constructor"); ctor.Valid() && ctor.IsFunction() {
			if name := ctor.Object().Get("name"); name.Valid() && name.IsString() && name.Str() != "" {
				return "an instance of " + name.Str()
			}
		}
		return "type object"
	default:
		return "unknown"
	}
}

// describeArg is describe for an optional argument slot.
func describeArg(e *env.Env, args []js.Value, i int) string {
	if i >= len(args) {
		return "undefined"
	}
	return describe(e, args[i])
}

func formatNumber(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func argType(v js.Value) string {
	switch {
	case v.IsUndefined():
		return "undefined"
	case v.IsNull():
		return "null"
	case v.IsBool():
		return "boolean"
	case v.IsNumber():
		return "number"
	case v.IsString():
		return "string"
	case v.IsFunction():
		return "function"
	case v.IsObject():
		return "object"
	default:
		return "unknown"
	}
}

// Implemented lists the tracked Node 24 API entries this package provides.
func Implemented() []string {
	return []string{
		"v8.Deserializer",
		"v8.Deserializer#_readHostObject",
		"v8.Deserializer#getWireFormatVersion",
		"v8.Deserializer#readDouble",
		"v8.Deserializer#readHeader",
		"v8.Deserializer#readRawBytes",
		"v8.Deserializer#readUint32",
		"v8.Deserializer#readUint64",
		"v8.Deserializer#readValue",
		"v8.Deserializer#transferArrayBuffer",
		"v8.DefaultDeserializer",
		"v8.DefaultSerializer",
		"v8.GCProfiler",
		"v8.GCProfiler#[Symbol.dispose]",
		"v8.GCProfiler#start",
		"v8.GCProfiler#stop",
		"v8.Serializer",
		"v8.Serializer#_getDataCloneError",
		"v8.Serializer#_getSharedArrayBufferId",
		"v8.Serializer#_setTreatArrayBufferViewsAsHostObjects",
		"v8.Serializer#_writeHostObject",
		"v8.Serializer#releaseBuffer",
		"v8.Serializer#transferArrayBuffer",
		"v8.Serializer#writeDouble",
		"v8.Serializer#writeHeader",
		"v8.Serializer#writeRawBytes",
		"v8.Serializer#writeUint32",
		"v8.Serializer#writeUint64",
		"v8.Serializer#writeValue",
		"v8.addDeserializeCallback",
		"v8.addSerializeCallback",
		"v8.after",
		"v8.before",
		"v8.cachedDataVersionTag",
		"v8.createHook",
		"v8.deserialize",
		"v8.getCppHeapStatistics",
		"v8.getHeapCodeStatistics",
		"v8.getHeapSnapshot",
		"v8.getHeapSpaceStatistics",
		"v8.getHeapStatistics",
		"v8.init",
		"v8.isBuildingSnapshot",
		"v8.isStringOneByteRepresentation",
		"v8.onAfter",
		"v8.onBefore",
		"v8.onInit",
		"v8.onSettled",
		"v8.queryObjects",
		"v8.serialize",
		"v8.setDeserializeMainFunction",
		"v8.setFlagsFromString",
		"v8.setHeapSnapshotNearHeapLimit",
		"v8.settled",
		"v8.stopCoverage",
		"v8.takeCoverage",
		"v8.writeHeapSnapshot",
	}
}

// serializerPrelude implements the v8 Serializer/Deserializer surface. The
// wire format starts with the two-byte V8 version header (0xFF + version) and
// then uses the V8 value tags for the shapes the serializer understands, so a
// value written here can still be read back by this runtime and the handful of
// historical byte sequences in Node's tests decode correctly.
const serializerPrelude = `(function () {
  'use strict';

  var HEADER_VERSION = 0x0f;

  var T = {
    UNDEFINED: 0x5f, NULL: 0x30, TRUE: 0x54, FALSE: 0x46,
    INT: 0x53, DOUBLE: 0x44, BIGINT: 0x5a, STRING: 0x22,
    OBJECT: 0x6f, OBJECT_END: 0x7b, ARRAY: 0x41, ARRAY_END: 0x24,
    DATE: 0x64, REGEXP: 0x52, TYPED: 0x5c, MAP: 0x3b, MAP_END: 0x3a,
    SET: 0x27, SET_END: 0x2c, REF: 0x5e, HOST: 0x68
  };

  var TYPED_NAMES = [
    'Int8Array', 'Uint8Array', 'Uint8ClampedArray', 'Int16Array',
    'Uint16Array', 'Int32Array', 'Uint32Array', 'Float32Array',
    'Float64Array', 'BigInt64Array', 'BigUint64Array', 'Float16Array'
  ];

  function utf8Encode(str) {
    if (typeof TextEncoder === 'function') return new TextEncoder().encode(str);
    var out = [];
    for (var i = 0; i < str.length; i++) {
      var c = str.charCodeAt(i);
      if (c < 0x80) {
        out.push(c);
      } else if (c < 0x800) {
        out.push(0xc0 | (c >> 6), 0x80 | (c & 0x3f));
      } else {
        out.push(0xe0 | (c >> 12), 0x80 | ((c >> 6) & 0x3f), 0x80 | (c & 0x3f));
      }
    }
    return new Uint8Array(out);
  }

  function utf8Decode(bytes) {
    if (typeof TextDecoder === 'function') return new TextDecoder().decode(bytes);
    var s = '';
    for (var i = 0; i < bytes.length; i++) s += String.fromCharCode(bytes[i]);
    return s;
  }

  function makeBuffer(bytes) {
    if (typeof Buffer === 'function' && typeof Buffer.from === 'function') {
      return Buffer.from(bytes);
    }
    var out = new Uint8Array(bytes.length);
    for (var i = 0; i < bytes.length; i++) out[i] = bytes[i];
    return out;
  }

  function toBytes(src) {
    if (src && typeof ArrayBuffer !== 'undefined' && ArrayBuffer.isView &&
        ArrayBuffer.isView(src)) {
      return new Uint8Array(src.buffer, src.byteOffset, src.byteLength);
    }
    return null;
  }

  function typedIndex(name) {
    for (var i = 0; i < TYPED_NAMES.length; i++) {
      if (TYPED_NAMES[i] === name) return i;
    }
    return -1;
  }

  function typedFrom(typeId, bytes) {
    if (typeId === 13) return makeBuffer(bytes);
    var copy = new Uint8Array(bytes.length);
    for (var i = 0; i < bytes.length; i++) copy[i] = bytes[i];
    if (typeId === 12) return new DataView(copy.buffer);
    var name = TYPED_NAMES[typeId];
    var ctor = name && typeof globalThis !== 'undefined' ? globalThis[name] : undefined;
    if (typeof ctor !== 'function') {
      throw new Error('Unable to deserialize cloned data: unsupported typed array');
    }
    return new ctor(copy.buffer);
  }

  function hostName(value) {
    if (value && value.constructor && typeof value.constructor.name === 'string' &&
        value.constructor.name !== '') {
      return value.constructor.name;
    }
    return 'Object';
  }

  function isHostObject(value) {
    return !!value && value.constructor !== undefined &&
      hostName(value) === 'JSStream';
  }

  function defaultWriteHostObject(value) {
    throw new Error('Unserializable host object: ' + hostName(value) + ' {}');
  }

  function defaultReadHostObject() {
    throw new Error('Unserializable host object');
  }

  function defaultCloneError(message) {
    var err = new Error(message);
    err.name = 'DataCloneError';
    err.code = 25;
    return err;
  }

  function Writer() { this.bytes = []; }
  Writer.prototype.byte = function (b) { this.bytes.push(b & 0xff); };
  Writer.prototype.u32 = function (v) {
    v = v >>> 0;
    this.byte(v); this.byte(v >>> 8); this.byte(v >>> 16); this.byte(v >>> 24);
  };
  Writer.prototype.varint = function (v) {
    v = v >>> 0;
    while (v >= 0x80) { this.byte((v & 0x7f) | 0x80); v = v >>> 7; }
    this.byte(v);
  };
  Writer.prototype.double = function (v) {
    var dv = new DataView(new ArrayBuffer(8));
    dv.setFloat64(0, v, true);
    for (var i = 0; i < 8; i++) this.byte(dv.getUint8(i));
  };
  Writer.prototype.raw = function (bytes) {
    for (var i = 0; i < bytes.length; i++) this.byte(bytes[i]);
  };
  Writer.prototype.string = function (s) {
    var bytes = utf8Encode(String(s));
    this.byte(T.STRING);
    this.varint(bytes.length);
    this.raw(bytes);
  };

  function Reader(bytes) {
    this.bytes = bytes;
    this.pos = 0;
    this.version = 0;
    this.refs = [];
  }
  Reader.prototype.need = function (n) {
    if (this.pos + n > this.bytes.length) {
      throw new Error('Unable to deserialize cloned data: unexpected end of buffer');
    }
  };
  Reader.prototype.peek = function () {
    this.need(1);
    return this.bytes[this.pos];
  };
  Reader.prototype.byte = function () {
    this.need(1);
    return this.bytes[this.pos++];
  };
  Reader.prototype.varint = function () {
    var result = 0, shift = 0, b;
    do {
      b = this.byte();
      result += (b & 0x7f) * Math.pow(2, shift);
      shift += 7;
      if (shift > 42) throw new Error('Unable to deserialize cloned data: invalid varint');
    } while (b & 0x80);
    return result;
  };
  Reader.prototype.varintSigned = function () {
    var raw = this.varint();
    return (raw >>> 1) ^ -(raw & 1);
  };
  Reader.prototype.double = function () {
    this.need(8);
    var dv = new DataView(this.bytes.buffer, this.bytes.byteOffset + this.pos, 8);
    this.pos += 8;
    return dv.getFloat64(0, true);
  };
  Reader.prototype.string = function () {
    if (this.byte() !== T.STRING) {
      throw new Error('Unable to deserialize cloned data: expected a string');
    }
    var n = this.varint();
    this.need(n);
    var s = utf8Decode(this.bytes.subarray(this.pos, this.pos + n));
    this.pos += n;
    return s;
  };
  Reader.prototype.bigint = function () {
    var n = this.varint();
    this.need(n);
    var digits = '';
    for (var i = 0; i < n; i++) digits += String.fromCharCode(this.bytes[this.pos + i]);
    this.pos += n;
    return BigInt(digits);
  };

  function Encoder(writer, serializer) {
    this.w = writer;
    this.ser = serializer;
    this.ids = new WeakMap();
    this.nextId = 0;
    this.treatViewsAsHost = false;
  }
  Encoder.prototype.cloneError = function (value) {
    var message = '#<' + hostName(value) + '> could not be cloned.';
    if (typeof this.ser._getDataCloneError === 'function') {
      return this.ser._getDataCloneError(message);
    }
    return defaultCloneError(message);
  };
  Encoder.prototype.register = function (value) {
    var id = this.nextId;
    this.nextId += 1;
    this.ids.set(value, id);
    return id;
  };
  Encoder.prototype.write = function (value) {
    var w = this.w;
    var type = typeof value;
    if (type === 'undefined') { w.byte(T.UNDEFINED); return; }
    if (value === null) { w.byte(T.NULL); return; }
    if (type === 'boolean') { w.byte(value ? T.TRUE : T.FALSE); return; }
    if (type === 'number') {
      if (Number.isInteger(value) && value >= -2147483648 && value <= 2147483647 &&
          !Object.is(value, -0)) {
        w.byte(T.INT);
        w.varint(((value << 1) ^ (value >> 31)) >>> 0);
      } else {
        w.byte(T.DOUBLE);
        w.double(value);
      }
      return;
    }
    if (type === 'bigint') {
      var digits = value.toString(10);
      w.byte(T.BIGINT);
      w.varint(digits.length);
      for (var d = 0; d < digits.length; d++) w.byte(digits.charCodeAt(d));
      return;
    }
    if (type === 'string') { w.string(value); return; }
    if (type === 'function' || type === 'symbol') {
      throw this.cloneError(value);
    }

    if (isHostObject(value)) {
      if (typeof this.ser._writeHostObject !== 'function') {
        throw new Error('Unserializable host object: ' + hostName(value) + ' {}');
      }
      w.byte(T.HOST);
      this.ser._writeHostObject(value);
      return;
    }

    if (typeof Buffer === 'function' && typeof Buffer.isBuffer === 'function' &&
        Buffer.isBuffer(value)) {
      var bufBytes = new Uint8Array(value.buffer, value.byteOffset, value.byteLength);
      w.byte(T.TYPED);
      w.varint(13);
      w.varint(bufBytes.length);
      w.raw(bufBytes);
      return;
    }

    var viewBytes = toBytes(value);
    if (viewBytes) {
      var viewName = value.constructor && value.constructor.name ?
        value.constructor.name : '';
      var kind = viewName === 'DataView' ? 12 : typedIndex(viewName);
      if (kind < 0) kind = 1;
      w.byte(T.TYPED);
      w.varint(kind);
      w.varint(viewBytes.length);
      w.raw(viewBytes);
      return;
    }

    var seen = this.ids.get(value);
    if (seen !== undefined) {
      w.byte(T.REF);
      w.varint(seen);
      return;
    }

    if (value instanceof Date) {
      w.byte(T.DATE);
      w.double(value.getTime());
      return;
    }
    if (value instanceof RegExp) {
      w.byte(T.REGEXP);
      w.string(value.source);
      w.string(value.flags);
      return;
    }
    if (typeof Map === 'function' && value instanceof Map) {
      this.register(value);
      var encoder = this;
      w.byte(T.MAP);
      value.forEach(function (v, k) { encoder.write(k); encoder.write(v); });
      w.byte(T.MAP_END);
      w.varint(value.size);
      return;
    }
    if (typeof Set === 'function' && value instanceof Set) {
      this.register(value);
      var enc = this;
      w.byte(T.SET);
      value.forEach(function (v) { enc.write(v); });
      w.byte(T.SET_END);
      w.varint(value.size);
      return;
    }
    this.register(value);
    if (Array.isArray(value)) {
      w.byte(T.ARRAY);
      for (var i = 0; i < value.length; i++) this.write(value[i]);
      w.byte(T.ARRAY_END);
      w.varint(value.length);
      return;
    }
    w.byte(T.OBJECT);
    var keys = Object.keys(value);
    for (var k = 0; k < keys.length; k++) {
      w.string(keys[k]);
      this.write(value[keys[k]]);
    }
    w.byte(T.OBJECT_END);
    w.varint(keys.length);
  };

  function decodeValue(r, des) {
    var tag = r.byte();
    switch (tag) {
      case T.UNDEFINED: return undefined;
      case T.NULL: return null;
      case T.TRUE: return true;
      case T.FALSE: return false;
      case T.INT: return r.varintSigned();
      case T.DOUBLE: return r.double();
      case T.BIGINT: return r.bigint();
      case T.STRING: {
        r.pos -= 1;
        return r.string();
      }
      case T.REF: return r.refs[r.varint()];
      case T.HOST: return des._readHostObject();
      case T.DATE: return new Date(r.double());
      case T.REGEXP: {
        var source = r.string();
        var flags = r.string();
        return new RegExp(source, flags);
      }
      case T.OBJECT: {
        var obj = {};
        r.refs.push(obj);
        for (;;) {
          if (r.peek() === T.OBJECT_END) { r.pos += 1; r.varint(); return obj; }
          var key = r.string();
          Object.defineProperty(obj, key, {
            value: decodeValue(r, des),
            enumerable: true, configurable: true, writable: true
          });
        }
      }
      case T.ARRAY: {
        var arr = [];
        r.refs.push(arr);
        for (;;) {
          if (r.peek() === T.ARRAY_END) { r.pos += 1; r.varint(); return arr; }
          arr.push(decodeValue(r, des));
        }
      }
      case T.MAP: {
        var map = new Map();
        r.refs.push(map);
        for (;;) {
          if (r.peek() === T.MAP_END) { r.pos += 1; r.varint(); return map; }
          var mk = decodeValue(r, des);
          map.set(mk, decodeValue(r, des));
        }
      }
      case T.SET: {
        var set = new Set();
        r.refs.push(set);
        for (;;) {
          if (r.peek() === T.SET_END) { r.pos += 1; r.varint(); return set; }
          set.add(decodeValue(r, des));
        }
      }
      case T.TYPED: {
        var typeId = r.varint();
        var length = r.varint();
        r.need(length);
        var bytes = r.bytes.subarray(r.pos, r.pos + length);
        r.pos += length;
        return typedFrom(typeId, bytes);
      }
      default:
        throw new Error('Unable to deserialize cloned data: unknown tag ' + tag);
    }
  }

  function constructRequired(name) {
    var err = new TypeError("Class constructor " + name + " cannot be invoked without 'new'");
    err.code = 'ERR_CONSTRUCT_CALL_REQUIRED';
    return err;
  }

  function Serializer() {
    if (!(this instanceof Serializer)) throw constructRequired('Serializer');
    this._w = new Writer();
    this._enc = new Encoder(this._w, this);
    this._transfers = {};
  }
  Serializer.prototype._writeHostObject = defaultWriteHostObject;
  Serializer.prototype._getDataCloneError = defaultCloneError;
  Serializer.prototype._getSharedArrayBufferId = function () { return undefined; };
  Serializer.prototype._setTreatArrayBufferViewsAsHostObjects = function (flag) {
    this._enc.treatViewsAsHost = !!flag;
  };
  Serializer.prototype.writeHeader = function () {
    this._w.byte(0xff);
    this._w.byte(HEADER_VERSION);
  };
  Serializer.prototype.writeValue = function (value) { this._enc.write(value); };
  Serializer.prototype.releaseBuffer = function () {
    var out = makeBuffer(this._w.bytes);
    this._w = new Writer();
    this._enc = new Encoder(this._w, this);
    return out;
  };
  Serializer.prototype.transferArrayBuffer = function (id, buffer) {
    this._transfers[id] = buffer;
  };
  Serializer.prototype.writeUint32 = function (value) { this._w.u32(value); };
  Serializer.prototype.writeUint64 = function (hi, lo) {
    this._w.u32(hi);
    this._w.u32(lo);
  };
  Serializer.prototype.writeDouble = function (value) { this._w.double(value); };
  Serializer.prototype.writeRawBytes = function (source) {
    var bytes = toBytes(source);
    if (!bytes) throw new TypeError('source must be a TypedArray or a DataView');
    this._w.raw(bytes);
  };

  function Deserializer(buffer) {
    if (!(this instanceof Deserializer)) throw constructRequired('Deserializer');
    var bytes = toBytes(buffer);
    if (!bytes) throw new TypeError('buffer must be a TypedArray or a DataView');
    this._r = new Reader(bytes);
    this._transfers = {};
  }
  Deserializer.prototype._readHostObject = defaultReadHostObject;
  Deserializer.prototype.readHeader = function () {
    if (this._r.byte() !== 0xff) {
      throw new Error('Unable to deserialize cloned data.');
    }
    this._r.version = this._r.byte();
  };
  Deserializer.prototype.getWireFormatVersion = function () { return this._r.version; };
  Deserializer.prototype.readValue = function () { return decodeValue(this._r, this); };
  Deserializer.prototype.readUint32 = function () {
    this._r.need(4);
    var b = this._r.bytes;
    var p = this._r.pos;
    var v = (b[p] | (b[p + 1] << 8) | (b[p + 2] << 16) | (b[p + 3] << 24)) >>> 0;
    this._r.pos = p + 4;
    return v;
  };
  Deserializer.prototype.readUint64 = function () {
    var hi = this.readUint32();
    var lo = this.readUint32();
    return [hi, lo];
  };
  Deserializer.prototype.readDouble = function () {
    if (this._r.pos + 8 > this._r.bytes.length) {
      throw new Error('ReadDouble() failed');
    }
    return this._r.double();
  };
  Deserializer.prototype.readRawBytes = function (length) {
    this._r.need(length);
    var out = makeBuffer(this._r.bytes.subarray(this._r.pos, this._r.pos + length));
    this._r.pos += length;
    return out;
  };
  Deserializer.prototype.transferArrayBuffer = function (id, buffer) {
    this._transfers[id] = buffer;
  };

  function DefaultSerializer() { Serializer.call(this); }
  DefaultSerializer.prototype = Object.create(Serializer.prototype);
  DefaultSerializer.prototype.constructor = DefaultSerializer;

  function DefaultDeserializer(buffer) { Deserializer.call(this, buffer); }
  DefaultDeserializer.prototype = Object.create(Deserializer.prototype);
  DefaultDeserializer.prototype.constructor = DefaultDeserializer;

  function serialize(value) {
    var ser = new DefaultSerializer();
    ser.writeHeader();
    ser.writeValue(value);
    return ser.releaseBuffer();
  }

  function deserialize(buffer) {
    var des = new DefaultDeserializer(buffer);
    des.readHeader();
    return des.readValue();
  }

  return {
    Serializer: Serializer,
    DefaultSerializer: DefaultSerializer,
    Deserializer: Deserializer,
    DefaultDeserializer: DefaultDeserializer,
    serialize: serialize,
    deserialize: deserialize
  };
})()`

// extraPrelude holds the startup-snapshot and promise-hook shims, which only
// need declared shapes on a runtime that is not V8.
const extraPrelude = `(function () {
  'use strict';

  function notBuildingSnapshot() {
    var err = new Error('This API is only available while building a snapshot');
    err.code = 'ERR_NOT_BUILDING_SNAPSHOT';
    return err;
  }

  function isBuildingSnapshot() { return false; }

  function forbid() { throw notBuildingSnapshot(); }

  function createHook() {
    return { enable: function () {}, disable: function () {} };
  }

  function setHeapSnapshotNearHeapLimit(limit) {
    if (typeof limit !== 'number') {
      var typeErr = new TypeError('The "limit" argument must be of type number. Received ' +
        typeof limit);
      typeErr.code = 'ERR_INVALID_ARG_TYPE';
      throw typeErr;
    }
    if (!(limit >= 1)) {
      var rangeErr = new RangeError('The value of "limit" is out of range. It must be >= 1. Received ' +
        limit);
      rangeErr.code = 'ERR_OUT_OF_RANGE';
      throw rangeErr;
    }
  }

  var startupSnapshot = {
    isBuildingSnapshot: isBuildingSnapshot,
    addSerializeCallback: forbid,
    addDeserializeCallback: forbid,
    setDeserializeMainFunction: forbid
  };

  return {
    startupSnapshot: startupSnapshot,
    isBuildingSnapshot: isBuildingSnapshot,
    addSerializeCallback: forbid,
    addDeserializeCallback: forbid,
    setDeserializeMainFunction: forbid,
    createHook: createHook,
    before: function () {},
    after: function () {},
    onBefore: function () {},
    onAfter: function () {},
    onInit: function () {},
    onSettled: function () {},
    settled: function () { return Promise.resolve(); },
    init: function () {},
    setHeapSnapshotNearHeapLimit: setHeapSnapshotNearHeapLimit
  };
})()`
