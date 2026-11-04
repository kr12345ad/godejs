// Package https exposes the secure entry points of the HTTP implementation.
// The server and client protocol logic lives in modules/http; this package
// only selects the TLS listener and TLS transport paths.
package https

import (
	"godejs/internal/env"
	"godejs/internal/js"

	_ "godejs/modules/http"
)

func init() { env.Register("https", Install) }

func Install(e *env.Env) js.Object {
	http := e.Exports("http")
	m := e.JS.Obj().Keep()
	for _, key := range http.OwnKeys() {
		m.Set(key, http.Get(key))
	}
	m.Set("createServer", http.Get("_createSecureServer"))
	m.Set("request", http.Get("_requestSecure"))
	m.Set("get", http.Get("_getSecure"))
	m.Set("Server", http.Get("Server"))
	m.Set("Agent", http.Get("Agent"))
	m.Set("globalAgent", http.Get("globalAgent"))
	return m
}

func Implemented() []string {
	return []string{
		"https.Agent",
		"https.Server",
		"https.createServer",
		"https.get",
		"https.globalAgent",
		"https.request",
	}
}
