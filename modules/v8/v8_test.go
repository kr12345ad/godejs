package v8

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"godejs/internal/env"
	"godejs/internal/jstest"
)

func newEnv(t *testing.T) *env.Env {
	t.Helper()
	e := jstest.New(t)
	jstest.Install(t, e, "v8", Install)
	return e
}

func TestHeapStatisticsShape(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `Object.keys(v8.getHeapStatistics()).sort().join(",")`)
	want := "does_zap_garbage,external_memory,heap_size_limit,malloced_memory," +
		"number_of_detached_contexts,number_of_native_contexts,peak_malloced_memory," +
		"total_available_size,total_global_handles_size,total_heap_size," +
		"total_heap_size_executable,total_physical_size,used_global_handles_size,used_heap_size"
	jstest.Equal(t, got, want)

	jstest.Equal(t, jstest.Str(t, e, `typeof v8.getHeapStatistics().heap_size_limit === "number"`), "true")
	jstest.Equal(t, jstest.Str(t, e, `v8.getHeapStatistics().heap_size_limit > 0`), "true")
	jstest.Equal(t, jstest.Str(t, e, `v8.getHeapStatistics().used_global_handles_size >= 0`), "true")
}

func TestHeapSpaceStatistics(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `v8.getHeapSpaceStatistics().length >= 4`), "true")
	jstest.Equal(t, jstest.Str(t, e, `v8.getHeapSpaceStatistics().every(function (s) {
		return typeof s.space_name === "string" &&
			typeof s.space_size === "number" &&
			typeof s.space_used_size === "number" &&
			typeof s.space_available_size === "number" &&
			typeof s.physical_space_size === "number";
	})`), "true")
	jstest.Equal(t, jstest.Str(t, e, `v8.getHeapSpaceStatistics()[0].space_name`), "new_space")
}

func TestHeapCodeStatistics(t *testing.T) {
	e := newEnv(t)
	got := jstest.Str(t, e, `Object.keys(v8.getHeapCodeStatistics()).sort().join(",")`)
	jstest.Equal(t, got, "bytecode_and_metadata_size,code_and_metadata_size,cpu_profiler_metadata_size,external_script_source_size")
}

func TestCppHeapStatistics(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `JSON.stringify(v8.getCppHeapStatistics())`),
		`{"total_size":0,"used_size":0}`)
}

func TestNoOps(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `typeof v8.setFlagsFromString("--noopt")`), "undefined")
	jstest.Equal(t, jstest.Str(t, e, `typeof v8.takeCoverage()`), "undefined")
	jstest.Equal(t, jstest.Str(t, e, `typeof v8.stopCoverage()`), "undefined")
	jstest.Equal(t, jstest.Str(t, e, `v8.queryObjects(function () {})`), "0")
	jstest.Equal(t, jstest.Str(t, e, `v8.queryObjects(function () {}, { format: "summary" }).length`), "0")
}

func TestCachedDataVersionTag(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `v8.cachedDataVersionTag() === v8.cachedDataVersionTag()`), "true")
	jstest.Equal(t, jstest.Str(t, e, `v8.cachedDataVersionTag() > 0`), "true")
	jstest.Equal(t, jstest.Str(t, e, `v8.cachedDataVersionTag() % 1 === 0`), "true")
}

func TestIsStringOneByteRepresentation(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `v8.isStringOneByteRepresentation("abc")`), "true")
	jstest.Equal(t, jstest.Str(t, e, `v8.isStringOneByteRepresentation("latin1-\u00e9")`), "true")
	jstest.Equal(t, jstest.Str(t, e, `v8.isStringOneByteRepresentation("\u20ac")`), "false")
	jstest.Equal(t, jstest.Str(t, e, `v8.isStringOneByteRepresentation("")`), "true")

	err := jstest.Throws(t, e, `v8.isStringOneByteRepresentation(42)`)
	if !strings.Contains(err.Message, `must be of type string`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteHeapSnapshot(t *testing.T) {
	e := newEnv(t)
	name := "Heap.test.heapsnapshot"
	got := jstest.Str(t, e, `v8.writeHeapSnapshot("`+name+`")`)
	want := filepath.Join(e.Cwd(), name)
	jstest.Equal(t, got, want)

	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("snapshot not written: %v", err)
	}
	if !strings.Contains(string(data), `"total_heap_size"`) {
		t.Fatalf("snapshot missing heap statistics: %s", data)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("snapshot is not JSON: %v", err)
	}
	hs, ok := payload["heapStatistics"].(map[string]any)
	if !ok {
		t.Fatalf("heapStatistics missing: %#v", payload)
	}
	if _, ok := hs["used_heap_size"]; !ok {
		t.Fatalf("used_heap_size missing: %#v", hs)
	}
}

func TestGCProfiler(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `typeof v8.GCProfiler`), "function")
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const p = new v8.GCProfiler();
		p.start();
		const s = p.stop();
		return p instanceof v8.GCProfiler &&
			Array.isArray(s.statistics) &&
			s.statistics.length === 0 &&
			typeof s.version === "number" &&
			s.endTime >= s.startTime &&
			p.stop() === undefined;
	})()`), "true")

	// Explicit collections inside the profiling window are recorded, with the
	// V8 GC names Node reports, and global.gc is restored afterwards.
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		const original = function () { return 'native'; };
		globalThis.gc = original;
		const p = new v8.GCProfiler();
		p.start();
		globalThis.gc({ type: 'minor', execution: 'sync' });
		globalThis.gc();
		const s = p.stop();
		const ok = s.statistics.length === 2 &&
			s.statistics[0].gcType === 'MinorMarkSweep' &&
			s.statistics[1].gcType === 'MajorGC' &&
			s.statistics[0].afterGC.heapStatistics.usedHeapSize === 0 &&
			s.statistics[0].afterGC.heapSpaceStatistics[0].spaceName === 'new_space' &&
			globalThis.gc === original;
		delete globalThis.gc;
		return ok;
	})()`), "true")
}

func TestSetFlagsFromStringValidatesAndChangesTag(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		try { v8.setFlagsFromString(1); } catch (err) { return err.code + ':' + err.name; }
		return 'no-throw';
	})()`), "ERR_INVALID_ARG_TYPE:TypeError")

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		try { v8.setFlagsFromString(); } catch (err) { return err.message; }
		return 'no-throw';
	})()`), `The "flags" argument must be of type string. Received undefined`)

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var before = v8.cachedDataVersionTag();
		v8.setFlagsFromString('--allow_natives_syntax');
		var after = v8.cachedDataVersionTag();
		return before !== after && typeof after === 'number';
	})()`), "true")
}

func TestSerializeRoundTripsValues(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `v8.serialize(1)[0]`), "255")

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var back = v8.deserialize(v8.serialize({ a: 1, b: 'two', c: [1, 2, 3] }));
		return back.a === 1 && back.b === 'two' && back.c.length === 3 && back.c[2] === 3;
	})()`), "true")

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var value = {};
		value.self = value;
		value.list = [1, value];
		var back = v8.deserialize(v8.serialize(value));
		return back.self === back && back.list[1] === back;
	})()`), "true")

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var back = v8.deserialize(v8.serialize(new Uint8Array([1, 2, 3])));
		return back instanceof Uint8Array && back.length === 3 && back[2] === 3;
	})()`), "true")

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var back = v8.deserialize(v8.serialize(new Set([1, 2])));
		return back.size === 2 && back.has(2);
	})()`), "true")

	jstest.Equal(t, jstest.Str(t, e, `(function () {
		try { v8.deserialize('not a buffer'); } catch (err) { return err.message; }
		return 'no-throw';
	})()`), "buffer must be a TypedArray or a DataView")
}

func TestStartCpuProfile(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var handle = v8.startCpuProfile();
		var first = handle.stop();
		return typeof first === 'string' && first.length > 0 && handle.stop() === undefined;
	})()`), "true")
}

func TestStartupSnapshotIsUnavailable(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `v8.startupSnapshot.isBuildingSnapshot()`), "false")
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		try { v8.startupSnapshot.addSerializeCallback(function () {}); } catch (err) { return err.code; }
		return 'no-throw';
	})()`), "ERR_NOT_BUILDING_SNAPSHOT")
}

func TestHeapSnapshotIsStreamLike(t *testing.T) {
	e := newEnv(t)
	jstest.Equal(t, jstest.Str(t, e, `(function () {
		var first = v8.getHeapSnapshot();
		var second = v8.getHeapSnapshot();
		return typeof first.on === 'function' && typeof second.pipe === 'function';
	})()`), "true")
}

func TestImplementedKeysAreTracked(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "parity", "baseline", "v8.json"))
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	var baseline []string
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatalf("parse baseline: %v", err)
	}
	tracked := make(map[string]bool, len(baseline))
	for _, k := range baseline {
		tracked[k] = true
	}
	for _, k := range Implemented() {
		if !tracked[k] {
			t.Fatalf("Implemented() key %q is not in the Node 24 baseline", k)
		}
	}
}
