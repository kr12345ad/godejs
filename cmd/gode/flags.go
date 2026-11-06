package main

// Node.js command-line compatibility. Every flag node --help documents is
// recognized with the same arity and spelling (including --no- negations and
// aliases). Flags that JavaScriptCore can implement are wired in main.go;
// engine- or product-specific flags that have no equivalent fail explicitly
// instead of being silently ignored.

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"godejs/internal/env"
)

type flagKind int

const (
	boolFlag flagKind = iota
	valueFlag
	optionalFlag
)

type flagSpec struct {
	name      string
	kind      flagKind
	takesNext bool // optional flags that consume a following non-flag (only -p)
}

// V8 sampling-profiler flags are intentionally absent. JavaScriptCore has no
// CPU, heap, or tick sampling profiler, so these are not advertised or
// recognized as if they could work.

// rawFlags uses a trailing '=' for a required value and '?' for an optional
// value. Everything else is boolean, including every --no- negation.
var rawFlags = strings.Fields(`
abort-on-uncaught-exception
allow-addons
allow-child-process
allow-ffi
allow-fs-read=
allow-fs-vfs
allow-fs-write=
allow-inspector
allow-net
allow-openssl-store
allow-wasi
allow-worker
bench
bench-isolation=
bench-name-pattern=
bench-reporter=
bench-reporter-destination=
bench-samples=
bench-warmup=
build-sea=
build-snapshot
build-snapshot-config=
check
completion-bash
config-file=
conditions=
debug
debug-port=
diagnostic-dir=
disable-proto=
disable-sigusr1
disable-warning=
disable-wasm-trap-handler
disallow-code-generation-from-strings
dns-result-order=
enable-etw-stack-walking
enable-fips
enable-fips-indicator-events
enable-network-family-autoselection
enable-source-maps
entry-url
env-file=
env-file-if-exists=
eval=
experimental-addon-modules
experimental-bench
experimental-config-file=
experimental-default-config-file=
experimental-eventsource
experimental-import-meta-resolve
experimental-import-text
experimental-inspector-network-resource
experimental-loader=
experimental-network-inspection
experimental-package-map=
experimental-print-required-tla
experimental-sea-config=
experimental-storage-inspection
experimental-stream-iter
experimental-strip-types
experimental-test-coverage
experimental-test-isolation=
experimental-test-module-mocks
experimental-test-tag-filter=
experimental-vfs
experimental-vm-modules
experimental-web-worker
experimental-webstorage
experimental-worker-inspection
expose-gc
force-context-aware
force-fips?
force-node-api-uncaught-exceptions-policy
frozen-intrinsics
heapsnapshot-near-heap-limit=
heapsnapshot-signal=
help
icu-data-dir=
import=
input-type=
insecure-http-parser
inspect?
inspect-brk?
inspect-port=
inspect-publish-uid=
inspect-wait?
interactive
interpreted-frames-native-stack
jitless
loader=
localstorage-file=
max-http-header-size=
max-old-space-size=
max-old-space-size-percentage=
max-semi-space-size=
network-family-autoselection-attempt-timeout=
no-addons
no-async-context-frame
no-deprecation
no-experimental-detect-module
no-experimental-ffi
no-experimental-global-navigator
no-experimental-repl-await
no-experimental-require-module
no-experimental-sqlite
no-experimental-websocket
no-experimental-webstorage
no-extra-info-on-fatal-exception
no-force-async-hooks-checks
no-global-search-paths
no-network-family-autoselection
no-require-module
no-strip-types
no-warnings
no-worker-snapshot
node-memory-debug
openssl-config=
openssl-legacy-provider
openssl-shared-config
pending-deprecation
permission
permission-audit
preserve-symlinks
preserve-symlinks-main
print?
process-timeout=
redirect-warnings=
report-compact
report-dir=
report-exclude-env
report-exclude-network
report-filename=
report-on-fatalerror
report-on-process-timeout
report-on-signal
report-signal=
report-uncaught-exception
require=
run=
secure-heap=
secure-heap-min=
snapshot-blob=
stack-size=
stack-trace-limit=
test
test-concurrency=
test-coverage-branches=
test-coverage-exclude=
test-coverage-functions=
test-coverage-include=
test-coverage-include-all
test-coverage-lines=
test-force-exit
test-global-setup=
test-isolation=
test-name-pattern=
test-only
test-random-seed=
test-randomize
test-reporter=
test-reporter-destination=
test-rerun-failures=
test-shard=
test-skip-pattern=
test-timeout=
test-update-snapshots
throw-deprecation
title=
tls-cipher-list=
tls-keylog=
tls-max-v1.2
tls-max-v1.3
tls-min-v1.0
tls-min-v1.1
tls-min-v1.2
tls-min-v1.3
trace-deprecation
trace-env
trace-env-js-stack
trace-env-native-stack
trace-event-categories=
trace-event-file-pattern=
trace-exit
trace-promises
trace-require-module=
trace-sigint
trace-sync-io
trace-tls
trace-uncaught
trace-warnings
unhandled-rejections=
use-bundled-ca
use-env-proxy
use-largepages=
use-openssl-ca
use-strict
use-system-ca
v8-options
v8-pool-size=
version
vfs-load=
watch
watch-kill-signal=
watch-path=
watch-preserve-output
zero-fill-buffers
`)

var nodeFlagSpecs = buildFlagSpecs()

func buildFlagSpecs() map[string]flagSpec {
	specs := make(map[string]flagSpec, len(rawFlags))
	for _, raw := range rawFlags {
		spec := flagSpec{name: raw}
		switch {
		case strings.HasSuffix(raw, "="):
			spec.name = strings.TrimSuffix(raw, "=")
			spec.kind = valueFlag
		case strings.HasSuffix(raw, "?"):
			spec.name = strings.TrimSuffix(raw, "?")
			spec.kind = optionalFlag
			spec.takesNext = spec.name == "print"
		default:
			spec.kind = boolFlag
		}
		specs[spec.name] = spec
	}
	return specs
}

var flagAliases = map[string]string{
	"experimental-loader":              "loader",
	"experimental-default-config-file": "experimental-config-file",
	"config-file":                      "experimental-config-file",
	"debug-port":                       "inspect-port",
	"report-directory":                 "report-dir",
	"experimental-test-isolation":      "test-isolation",
	"webstorage":                       "experimental-webstorage",
	"strip-types":                      "experimental-strip-types",
}

var shortFlags = map[string]string{
	"c": "check",
	"d": "debug",
	"e": "eval",
	"h": "help",
	"i": "interactive",
	"p": "print",
	"r": "require",
	"v": "version",
	"C": "conditions",
}

// unsupportedFlags are recognized but cannot be implemented faithfully on this
// runtime. They fail with a clear message rather than changing program
// behaviour silently.
var unsupportedFlags = []struct {
	name   string
	reason string
}{
	{"permission", "the Node permission model is not implemented"},
	{"permission-audit", "the Node permission model is not implemented"},
	{"allow-addons", "the Node permission model is not implemented"},
	{"allow-child-process", "the Node permission model is not implemented"},
	{"allow-ffi", "the Node permission model is not implemented"},
	{"allow-fs-read", "the Node permission model is not implemented"},
	{"allow-fs-vfs", "the Node permission model is not implemented"},
	{"allow-fs-write", "the Node permission model is not implemented"},
	{"allow-inspector", "the Node permission model is not implemented"},
	{"allow-net", "the Node permission model is not implemented"},
	{"allow-openssl-store", "the Node permission model is not implemented"},
	{"allow-wasi", "the Node permission model is not implemented"},
	{"allow-worker", "the Node permission model is not implemented"},
	{"inspect", "the V8 inspector protocol is unavailable on JavaScriptCore"},
	{"inspect-brk", "the V8 inspector protocol is unavailable on JavaScriptCore"},
	{"inspect-wait", "the V8 inspector protocol is unavailable on JavaScriptCore"},
	{"inspect-port", "the V8 inspector protocol is unavailable on JavaScriptCore"},
	{"inspect-publish-uid", "the V8 inspector protocol is unavailable on JavaScriptCore"},
	{"disable-sigusr1", "the V8 inspector protocol is unavailable on JavaScriptCore"},
	{"v8-options", "there is no V8 engine to configure"},
	{"v8-pool-size", "there is no V8 engine to configure"},
	{"jitless", "JavaScriptCore has no jitless mode"},
	{"disallow-code-generation-from-strings", "JavaScriptCore cannot disable eval"},
	{"interpreted-frames-native-stack", "JavaScriptCore has no V8 frame metadata"},
	{"heapsnapshot-signal", "JavaScriptCore heap snapshots are not V8 snapshots"},
	{"heapsnapshot-near-heap-limit", "JavaScriptCore heap snapshots are not V8 snapshots"},
	{"build-snapshot", "V8 startup snapshots are unavailable on JavaScriptCore"},
	{"build-snapshot-config", "V8 startup snapshots are unavailable on JavaScriptCore"},
	{"snapshot-blob", "V8 startup snapshots are unavailable on JavaScriptCore"},
	{"build-sea", "single executable applications are not implemented"},
	{"experimental-sea-config", "single executable applications are not implemented"},
	{"experimental-vfs", "the experimental virtual filesystem is not implemented"},
	{"vfs-load", "the experimental virtual filesystem is not implemented"},
	{"experimental-ffi", "the experimental FFI module is not implemented"},
	{"experimental-package-map", "package map resolution is not implemented"},
	{"experimental-addon-modules", "importable addon modules are not implemented"},
	{"experimental-inspector-network-resource", "the V8 inspector is unavailable"},
	{"experimental-network-inspection", "the V8 inspector is unavailable"},
	{"experimental-storage-inspection", "the V8 inspector is unavailable"},
	{"conditions", "custom export conditions are not implemented"},
	{"loader", "custom ES module loaders are not implemented"},
	{"experimental-config-file", "configuration files are not implemented"},
	{"experimental-eventsource", "the EventSource API is not implemented"},
	{"experimental-strip-types", "TypeScript type stripping is not implemented"},
	{"experimental-import-text", "text import attributes are not implemented"},
	{"experimental-web-worker", "the Web Worker API is not implemented"},
	{"experimental-print-required-tla", "printing pending top-level await is not implemented"},
	{"no-experimental-websocket", "WebSocket is always available; disabling it is not implemented"},
	{"no-experimental-webstorage", "Web Storage is always available; disabling it is not implemented"},
	{"no-experimental-global-navigator", "navigator is always available; disabling it is not implemented"},
	{"no-experimental-sqlite", "node:sqlite is always available; disabling it is not implemented"},
	{"enable-source-maps", "source-map support in stack traces is not implemented"},
	{"insecure-http-parser", "the HTTP parser is Go's; the insecure mode is not implemented"},
	{"max-http-header-size", "configurable HTTP header limits are not implemented"},
	{"dns-result-order", "configurable DNS result ordering is not implemented"},
	{"tls-cipher-list", "Go crypto/tls does not expose Node's cipher list"},
	{"tls-keylog", "TLS key logging is not implemented"},
	{"tls-max-v1.2", "TLS version selection is not implemented"},
	{"tls-max-v1.3", "TLS version selection is not implemented"},
	{"tls-min-v1.0", "TLS version selection is not implemented"},
	{"tls-min-v1.1", "TLS version selection is not implemented"},
	{"tls-min-v1.2", "TLS version selection is not implemented"},
	{"tls-min-v1.3", "TLS version selection is not implemented"},
	{"use-bundled-ca", "Go crypto/tls uses the system trust store"},
	{"use-openssl-ca", "Go crypto/tls does not use OpenSSL"},
	{"use-env-proxy", "proxy environment handling is not implemented"},
	{"enable-network-family-autoselection", "address-family autoselection is fixed by Go's dialer"},
	{"no-network-family-autoselection", "address-family autoselection is fixed by Go's dialer"},
	{"network-family-autoselection-attempt-timeout", "address-family autoselection is fixed by Go's dialer"},
	{"abort-on-uncaught-exception", "core-dumping on uncaught exceptions is not implemented"},
	{"no-addons", "disabling native addons is not implemented"},
	{"force-context-aware", "context-aware addon enforcement is not implemented"},
	{"force-node-api-uncaught-exceptions-policy", "the Node-API exception policy is not implemented"},
	{"no-async-context-frame", "the AsyncContextFrame toggle is not implemented"},
	{"no-force-async-hooks-checks", "async_hooks check toggling is not implemented"},
	{"node-memory-debug", "Node.js internal memory checking is not applicable"},
	{"disable-wasm-trap-handler", "WebAssembly trap handling is provided by JavaScriptCore"},
	{"enable-etw-stack-walking", "ETW stack walking is Windows/V8 specific"},
	{"enable-fips-indicator-events", "FIPS is not available with Go's TLS stack"},
	{"secure-heap", "OpenSSL secure heap is not available"},
	{"secure-heap-min", "OpenSSL secure heap is not available"},
	{"icu-data-dir", "ICU data is linked into JavaScriptCore"},
	{"process-timeout", "process timeout reporting is not implemented"},
	{"report-on-process-timeout", "process timeout reporting is not implemented"},
	{"report-exclude-env", "diagnostic report filtering is not implemented"},
	{"report-exclude-network", "diagnostic report filtering is not implemented"},
	{"trace-sync-io", "synchronous I/O tracing is not implemented"},
	{"trace-promises", "promise tracing is not implemented"},
	{"trace-env", "environment variable tracing is not implemented"},
	{"trace-env-js-stack", "environment variable tracing is not implemented"},
	{"trace-env-native-stack", "environment variable tracing is not implemented"},
	{"trace-sigint", "SIGINT JavaScript stack tracing is not implemented"},
	{"trace-tls", "TLS packet tracing is not implemented"},
	{"no-extra-info-on-fatal-exception", "fatal exception diagnostics are not configurable"},
	{"test-concurrency", "the Node test runner is not implemented"},
	{"test-coverage-branches", "the Node test runner is not implemented"},
	{"test-coverage-exclude", "the Node test runner is not implemented"},
	{"test-coverage-functions", "the Node test runner is not implemented"},
	{"test-coverage-include", "the Node test runner is not implemented"},
	{"test-coverage-include-all", "the Node test runner is not implemented"},
	{"test-coverage-lines", "the Node test runner is not implemented"},
	{"test-force-exit", "the Node test runner is not implemented"},
	{"test-global-setup", "the Node test runner is not implemented"},
	{"test-isolation", "the Node test runner is not implemented"},
	{"test-name-pattern", "the Node test runner is not implemented"},
	{"test-only", "the Node test runner is not implemented"},
	{"test-random-seed", "the Node test runner is not implemented"},
	{"test-randomize", "the Node test runner is not implemented"},
	{"test-reporter", "the Node test runner is not implemented"},
	{"test-reporter-destination", "the Node test runner is not implemented"},
	{"test-rerun-failures", "the Node test runner is not implemented"},
	{"test-shard", "the Node test runner is not implemented"},
	{"test-skip-pattern", "the Node test runner is not implemented"},
	{"test-timeout", "the Node test runner is not implemented"},
	{"test-update-snapshots", "the Node test runner is not implemented"},
	{"experimental-test-coverage", "the Node test runner is not implemented"},
	{"experimental-test-module-mocks", "the Node test runner is not implemented"},
	{"experimental-test-tag-filter", "the Node test runner is not implemented"},
	{"bench", "the benchmark runner is not implemented"},
	{"experimental-bench", "the benchmark runner is not implemented"},
	{"bench-isolation", "the benchmark runner is not implemented"},
	{"bench-name-pattern", "the benchmark runner is not implemented"},
	{"bench-reporter", "the benchmark runner is not implemented"},
	{"bench-reporter-destination", "the benchmark runner is not implemented"},
	{"bench-samples", "the benchmark runner is not implemented"},
	{"bench-warmup", "the benchmark runner is not implemented"},
	{"test", "the Node test runner is not implemented; run test files directly"},
	{"enable-fips", "FIPS mode is not available with Go's TLS stack"},
	{"force-fips", "FIPS mode is not available with Go's TLS stack"},
	{"openssl-config", "this runtime uses Go crypto/tls, not OpenSSL"},
	{"openssl-legacy-provider", "this runtime uses Go crypto/tls, not OpenSSL"},
	{"openssl-shared-config", "this runtime uses Go crypto/tls, not OpenSSL"},
}

type cliConfig struct {
	present     map[string]bool
	values      map[string][]string
	positionals []string
	execArgv    []string
}

func newCLIConfig() *cliConfig {
	return &cliConfig{
		present: make(map[string]bool),
		values:  make(map[string][]string),
	}
}

func (c *cliConfig) has(name string) bool { return c.present[name] }

func (c *cliConfig) last(name string) (string, bool) {
	values := c.values[name]
	if len(values) == 0 {
		return "", false
	}
	return values[len(values)-1], true
}

func (c *cliConfig) value(name string) string {
	v, _ := c.last(name)
	return v
}

func (c *cliConfig) all(name string) []string {
	return append([]string(nil), c.values[name]...)
}

func lookupLong(name string) (string, flagSpec, bool) {
	if canonical, ok := flagAliases[name]; ok {
		name = canonical
	}
	spec, ok := nodeFlagSpecs[name]
	return name, spec, ok
}

func lookupShort(name string) (string, flagSpec, bool) {
	canonical, ok := shortFlags[name]
	if !ok {
		return "", flagSpec{}, false
	}
	return canonical, nodeFlagSpecs[canonical], true
}

func parseCLI(args []string) (*cliConfig, error) {
	c := newCLIConfig()
	i := 0
	for ; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			i++
			break
		}
		if arg == "-" || !strings.HasPrefix(arg, "-") {
			break
		}
		if strings.HasPrefix(arg, "--") {
			raw := arg[2:]
			name, value := raw, ""
			hasValue := false
			if eq := strings.IndexByte(raw, '='); eq >= 0 {
				name, value, hasValue = raw[:eq], raw[eq+1:], true
			}
			canonical, spec, ok := lookupLong(name)
			if !ok {
				return nil, fmt.Errorf("bad option: --%s", name)
			}
			c.execArgv = append(c.execArgv, arg)
			switch spec.kind {
			case valueFlag:
				if !hasValue {
					if i+1 >= len(args) {
						return nil, fmt.Errorf("bad option: --%s requires an argument", name)
					}
					i++
					value = args[i]
					c.execArgv = append(c.execArgv, value)
				}
				c.values[canonical] = append(c.values[canonical], value)
				c.present[canonical] = true
			case optionalFlag:
				if !hasValue && spec.takesNext && i+1 < len(args) &&
					(args[i+1] == "-" || !strings.HasPrefix(args[i+1], "-")) {
					i++
					value = args[i]
					hasValue = true
					c.execArgv = append(c.execArgv, value)
				}
				c.present[canonical] = true
				if hasValue {
					c.values[canonical] = append(c.values[canonical], value)
				}
			default:
				if hasValue {
					switch strings.ToLower(value) {
					case "true", "1", "":
						c.present[canonical] = true
					case "false", "0":
						c.present[canonical] = false
					default:
						return nil, fmt.Errorf("bad option: --%s=%s", name, value)
					}
				} else {
					c.present[canonical] = true
				}
			}
			continue
		}

		short := arg[1:]
		if short == "" {
			break
		}
		canonical, spec, ok := lookupShort(short)
		if !ok {
			// Support -eCODE and -rMODULE attachments.
			if len(short) > 1 {
				if canon, sp, found := lookupShort(short[:1]); found && sp.kind == valueFlag {
					canonical, spec, ok = canon, sp, true
					c.execArgv = append(c.execArgv, arg[:2], short[1:])
					c.values[canonical] = append(c.values[canonical], short[1:])
					c.present[canonical] = true
					continue
				}
			}
			return nil, fmt.Errorf("bad option: -%s", short)
		}
		c.execArgv = append(c.execArgv, arg)
		switch spec.kind {
		case valueFlag:
			if i+1 >= len(args) {
				return nil, fmt.Errorf("bad option: -%s requires an argument", short)
			}
			i++
			c.execArgv = append(c.execArgv, args[i])
			c.values[canonical] = append(c.values[canonical], args[i])
			c.present[canonical] = true
		case optionalFlag:
			c.present[canonical] = true
			if spec.takesNext && i+1 < len(args) &&
				(args[i+1] == "-" || !strings.HasPrefix(args[i+1], "-")) {
				i++
				c.execArgv = append(c.execArgv, args[i])
				c.values[canonical] = append(c.values[canonical], args[i])
			}
		default:
			c.present[canonical] = true
		}
	}
	c.positionals = append(c.positionals, args[i:]...)
	return c, nil
}

func (o *options) unsupportedReason() string {
	for _, entry := range unsupportedFlags {
		if o.cfg.has(entry.name) {
			return fmt.Sprintf("--%s: %s", entry.name, entry.reason)
		}
	}
	return ""
}

func flagUsage(name string) string {
	spec, ok := nodeFlagSpecs[name]
	if !ok {
		return ""
	}
	switch spec.kind {
	case valueFlag:
		return "--" + name + "=..."
	case optionalFlag:
		return "--" + name + "[=...]"
	default:
		return "--" + name
	}
}

func flagHelpText(w io.Writer) {
	names := make([]string, 0, len(nodeFlagSpecs))
	for name := range nodeFlagSpecs {
		names = append(names, name)
	}
	sort.Strings(names)
	shortFor := make(map[string]string)
	for short, canonical := range shortFlags {
		shortFor[canonical] = "-" + short
	}
	for _, name := range names {
		line := "  "
		if short := shortFor[name]; short != "" {
			line += short + ", "
		}
		line += flagUsage(name)
		if desc := flagDescriptions[name]; desc != "" {
			line += "  " + desc
		}
		fmt.Fprintln(w, line)
	}
}

var flagDescriptions = map[string]string{
	"check":                  "syntax check the script without executing it",
	"completion-bash":        "print a source-able bash completion script",
	"conditions":             "additional user conditions for exports and imports",
	"debug":                  "run the script under gode's line debugger",
	"env-file":               "load environment variables from a file",
	"env-file-if-exists":     "load environment variables if the file exists",
	"eval":                   "evaluate the argument as JavaScript",
	"expose-gc":              "expose the gc() function",
	"frozen-intrinsics":      "freeze JavaScript built-in objects",
	"help":                   "print this help",
	"import":                 "preload an ES module",
	"input-type":             "module type for --eval and stdin (module or commonjs)",
	"interactive":            "enter the REPL after running the script",
	"preserve-symlinks":      "preserve symbolic links when resolving modules",
	"preserve-symlinks-main": "preserve the entry script's symbolic link",
	"print":                  "evaluate and print the result",
	"require":                "preload a CommonJS module",
	"run":                    "run a script from package.json",
	"stack-trace-limit":      "set Error.stackTraceLimit",
	"title":                  "set process.title",
	"trace-exit":             "print a stack trace when the process exits",
	"version":                "print the Node.js compatibility version",
	"watch":                  "restart the script when it changes",
	"watch-kill-signal":      "signal used to stop the child in watch mode",
	"watch-path":             "additional path to watch",
	"watch-preserve-output":  "keep output across restarts in watch mode",
	"disable-proto":          "delete or throw on Object.prototype.__proto__",
	"disable-warning":        "silence a warning code",
	"no-deprecation":         "silence deprecation warnings",
	"no-warnings":            "silence all process warnings",
	"trace-deprecation":      "print a stack trace on deprecation warnings",
	"trace-warnings":         "print a stack trace on process warnings",
	"throw-deprecation":      "throw on deprecation warnings",
	"zero-fill-buffers":      "zero-fill newly allocated Buffers",
	"localstorage-file":      "file used to persist localStorage",
}

// mergeEnvFiles applies --env-file and --env-file-if-exists in order.
func mergeEnvFiles(base, required, optional []string) ([]string, error) {
	order := make([]string, 0, len(base))
	values := make(map[string]string, len(base))
	set := func(key, value string) {
		if _, ok := values[key]; !ok {
			order = append(order, key)
		}
		values[key] = value
	}
	for _, entry := range base {
		if i := strings.IndexByte(entry, '='); i > 0 {
			set(entry[:i], entry[i+1:])
		}
	}
	load := func(path string, missingOK bool) error {
		data, err := os.ReadFile(path)
		if err != nil {
			if missingOK && os.IsNotExist(err) {
				return nil
			}
			return err
		}
		for n, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
			eq := strings.IndexByte(line, '=')
			if eq <= 0 {
				return fmt.Errorf("%s:%d: invalid environment line", path, n+1)
			}
			key := strings.TrimSpace(line[:eq])
			value := strings.TrimSpace(line[eq+1:])
			if len(value) >= 2 {
				quote := value[0]
				if (quote == '"' || quote == '\'') && value[len(value)-1] == quote {
					value = value[1 : len(value)-1]
					if quote == '"' {
						value = strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\"`, `"`, `\\`, `\`).Replace(value)
					}
				}
			}
			set(key, value)
		}
		return nil
	}
	for _, path := range required {
		if err := load(path, false); err != nil {
			return nil, fmt.Errorf("--env-file: %w", err)
		}
	}
	for _, path := range optional {
		if err := load(path, true); err != nil {
			return nil, fmt.Errorf("--env-file-if-exists: %w", err)
		}
	}
	out := make([]string, 0, len(order))
	for _, key := range order {
		out = append(out, key+"="+values[key])
	}
	return out, nil
}

// printCompletion writes a bash completion script listing every flag.
func printCompletion(w io.Writer) {
	names := make([]string, 0, len(nodeFlagSpecs))
	for name := range nodeFlagSpecs {
		names = append(names, name)
	}
	short := make([]string, 0, len(shortFlags))
	for s := range shortFlags {
		short = append(short, s)
	}
	sort.Strings(short)
	sort.Strings(names)
	words := make([]string, 0, len(names)+len(short))
	for _, s := range short {
		words = append(words, "-"+s)
	}
	for _, n := range names {
		words = append(words, "--"+n)
	}
	fmt.Fprintln(w, "_gode_complete() {")
	fmt.Fprintln(w, `  local cur="${COMP_WORDS[COMP_CWORD]}"`)
	fmt.Fprintf(w, "  COMPREPLY=( $(compgen -W %q -- \"$cur\") )\n", strings.Join(words, " "))
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w, "complete -F _gode_complete gode node")
}

// formatForInputWithSource applies Node's module-syntax detection to --eval
// and stdin input when --input-type was not given explicitly. Node accepts
// static import/export in `-e` snippets, just as it does in .mjs files.
func formatForInputWithSource(inputType, file, source string) string {
	if inputType == "" && (file == "[eval]" || file == "[stdin]") && env.HasModuleSyntax(source) {
		return "module"
	}
	return formatForInput(inputType, file)
}

// formatForInput maps --input-type and a filename to a loader format.
func formatForInput(inputType, file string) string {
	switch inputType {
	case "module", "esm":
		return "module"
	case "commonjs", "cjs":
		return "commonjs"
	}
	switch {
	case strings.HasSuffix(file, ".mjs"), strings.HasSuffix(file, ".mts"):
		return "module"
	}
	return "commonjs"
}

// quoteJS returns a JavaScript string literal.
func quoteJS(s string) string {
	return strconv.Quote(s)
}
