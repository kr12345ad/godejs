// Package compat contains the small compatibility builtins that do not need a
// dedicated Go package; the JS-only surface lives in jsbuiltins.go.
package compat

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.Register("perf_hooks", installPerfHooks)
	env.Register("trace_events", installTraceEvents)
	env.Register("punycode", installPunycode)
	env.Register("sys", installSys)
	env.Register("inspector", installInspector)
	env.Register("repl", installRepl)
	env.Register("sea", installSea)
}

func installSys(e *env.Env) js.Object {
	if exports := e.Exports("util"); exports.Valid() {
		return exports
	}
	return e.JS.Obj()
}

// Implemented lists the tracked Node 24 API entries supplied here.
func Implemented() []string {
	return []string{
		"perf_hooks.performance", "trace_events.createTracing", "punycode",
		"sys", "inspector.open", "repl.start", "sea.isSea",
	}
}
