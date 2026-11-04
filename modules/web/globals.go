package web

// Node exposes these constructors without a require() call. They are builtin
// modules here, so gode only defined them after the first require('url') or
// require('worker_threads'). Instantiating those modules during global setup
// makes `new URL(...)` and `new MessageChannel()` work at startup, as they do
// in Node.

import "godejs/internal/env"

func installModuleGlobals(e *env.Env) {
	g := e.Global()

	// url owns URL, URLSearchParams and URLPattern. It sets them on the global
	// object itself, but only when the module is installed.
	if env.Registered("url") {
		m := e.Exports("url")
		for _, name := range []string{"URL", "URLSearchParams", "URLPattern"} {
			if v := m.Get(name); v.Valid() && !v.IsUndefined() {
				g.Set(name, v)
			}
		}
	}

	// worker_threads owns the message-passing constructors.
	if env.Registered("worker_threads") {
		m := e.Exports("worker_threads")
		for _, name := range []string{"MessageChannel", "MessagePort", "BroadcastChannel"} {
			if v := m.Get(name); v.Valid() && !v.IsUndefined() {
				g.Set(name, v)
			}
		}
	}

	// stream/web ships the WHATWG constructors and installs them as globals
	// itself when its prelude runs. Node has them from startup.
	if env.Registered("stream/web") {
		e.Exports("stream/web")
	}
}
