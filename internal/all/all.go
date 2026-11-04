// Package all wires every builtin module into one registry.
package all

import (
	"godejs/internal/env"
	_ "godejs/internal/napi" // registers the Node-API .node loader
	"godejs/modules/assert"
	"godejs/modules/async_hooks"
	"godejs/modules/buffer"
	"godejs/modules/child_process"
	"godejs/modules/cluster"
	_ "godejs/modules/compat" // registers sys and the JS-only compatibility builtins
	"godejs/modules/console"
	"godejs/modules/crypto"
	"godejs/modules/dgram"
	"godejs/modules/diagnostics_channel"
	"godejs/modules/dns"
	"godejs/modules/domain"
	"godejs/modules/events"
	"godejs/modules/fs"
	godehttp "godejs/modules/http"
	"godejs/modules/http2"
	"godejs/modules/https"
	_ "godejs/modules/internals" // registers the internal/* specifiers
	"godejs/modules/module"
	"godejs/modules/net"
	"godejs/modules/nodetest"
	"godejs/modules/os"
	"godejs/modules/path"
	"godejs/modules/process"
	"godejs/modules/querystring"
	_ "godejs/modules/readline" // registers readline and readline/promises
	"godejs/modules/sqlite"
	"godejs/modules/stream"
	"godejs/modules/string_decoder"
	"godejs/modules/timers"
	"godejs/modules/tls"
	"godejs/modules/tty"
	"godejs/modules/url"
	"godejs/modules/util"
	"godejs/modules/v8"
	"godejs/modules/vm"
	"godejs/modules/wasi"
	"godejs/modules/web"
	"godejs/modules/worker_threads"
	"godejs/modules/zlib"
)

// Info describes one builtin module.
type Info struct {
	Name        string
	Install     env.Installer
	Implemented func() []string
}

// Modules returns every builtin in a stable order.
func Modules() []Info {
	return []Info{
		{Name: "assert", Install: assert.Install, Implemented: assert.Implemented},
		{Name: "async_hooks", Install: async_hooks.Install, Implemented: async_hooks.Implemented},
		{Name: "buffer", Install: buffer.Install, Implemented: buffer.Implemented},
		{Name: "child_process", Install: child_process.Install, Implemented: child_process.Implemented},
		{Name: "cluster", Install: cluster.Install, Implemented: cluster.Implemented},
		{Name: "console", Install: console.Install, Implemented: console.Implemented},
		{Name: "crypto", Install: crypto.Install, Implemented: crypto.Implemented},
		{Name: "dgram", Install: dgram.Install, Implemented: dgram.Implemented},
		{Name: "diagnostics_channel", Install: diagnostics_channel.Install, Implemented: diagnostics_channel.Implemented},
		{Name: "dns", Install: dns.Install, Implemented: dns.Implemented},
		{Name: "domain", Install: domain.Install, Implemented: domain.Implemented},
		{Name: "events", Install: events.Install, Implemented: events.Implemented},
		{Name: "fs", Install: fs.Install, Implemented: fs.Implemented},
		{Name: "http", Install: godehttp.Install, Implemented: godehttp.Implemented},
		{Name: "http3", Install: godehttp.InstallHTTP3, Implemented: godehttp.HTTP3Implemented},
		{Name: "http2", Install: http2.Install, Implemented: http2.Implemented},
		{Name: "https", Install: https.Install, Implemented: https.Implemented},
		{Name: "module", Install: module.Install, Implemented: module.Implemented},
		{Name: "net", Install: net.Install, Implemented: net.Implemented},
		{Name: "os", Install: os.Install, Implemented: os.Implemented},
		{Name: "path", Install: path.Install, Implemented: path.Implemented},
		{Name: "process", Install: process.Install, Implemented: process.Implemented},
		{Name: "querystring", Install: querystring.Install, Implemented: querystring.Implemented},
		{Name: "sqlite", Install: sqlite.Install, Implemented: sqlite.Implemented},
		{Name: "stream", Install: stream.Install, Implemented: stream.Implemented},
		{Name: "string_decoder", Install: string_decoder.Install, Implemented: string_decoder.Implemented},
		{Name: "test", Install: nodetest.Install, Implemented: nodetest.Implemented},
		{Name: "timers", Install: timers.Install, Implemented: timers.Implemented},
		{Name: "tls", Install: tls.Install, Implemented: tls.Implemented},
		{Name: "tty", Install: tty.Install, Implemented: tty.Implemented},
		{Name: "url", Install: url.Install, Implemented: url.Implemented},
		{Name: "util", Install: util.Install, Implemented: util.Implemented},
		{Name: "v8", Install: v8.Install, Implemented: v8.Implemented},
		{Name: "vm", Install: vm.Install, Implemented: vm.Implemented},
		{Name: "wasi", Install: wasi.Install, Implemented: wasi.Implemented},
		{Name: "worker_threads", Install: worker_threads.Install, Implemented: worker_threads.Implemented},
		{Name: "zlib", Install: zlib.Install, Implemented: zlib.Implemented},
	}
}

// InstallGlobals installs web-platform globals Node exposes without a require
// call (TextEncoder, TextDecoder, atob, btoa, structuredClone, performance,
// AbortController, queueMicrotask). It is separate from the module registry
// because these are not importable builtins.
func InstallGlobals(e *env.Env) { web.Install(e) }
