# JavaScriptCore in GodeJS

GodeJS does not use V8, goja, or a JavaScript shim. It embeds **JavaScriptCore**
directly through the C API in `internal/js`. The result is a Node-shaped Go
runtime with a different engine underneath.

## What is in the repository

| Path | Responsibility |
| --- | --- |
| `internal/js/jsc.go` | Context creation, evaluation, host functions, errors, promises |
| `internal/js/value.go` | Values, objects, arrays, typed arrays, property access |
| `internal/js/callback.go` | The C callback bridge between JavaScriptCore and Go |
| `internal/js/godejsc.h` | The minimal C API surface GodeJS uses |
| `internal/env/stack.go` | V8-style `CallSite` layer on top of JSC stack strings |
| `modules/zlib`, `modules/crypto`, ... | Go-backed Node builtins, no JavaScriptCore dependency |

The binding is deliberately small and auditable. JavaScriptCore `JSValueRef`s
are NaN-boxed, so handles are stored as `uintptr` rather than `unsafe.Pointer`;
storing a tagged value in a pointer slot makes Go's GC abort. The context is
thread-affine and `cmd/gode` pins its goroutine to one OS thread.

Large WebAssembly guests (notably Go's `wasip1` output) corrupt the
generational collector while growing memory, so `js.New` defaults
`JSC_useGenerationalGC=0` unless the caller explicitly set it.

## Runtime dependency

Release binaries link `libjavascriptcoregtk-4.0.so.18`. Install the runtime
package for your distribution:

```sh
# Debian / Ubuntu
sudo apt install libjavascriptcoregtk-4.0-18

# Alpine (BusyBox ash)
sudo apk add webkit2gtk-4.0

# Fedora
sudo dnf install webkit2gtk4.0
```

Package names move between distributions and releases. If your package manager
cannot find the name above, search for the package that provides
`libjavascriptcoregtk-4.0.so.18`.

## Build from source

You need Go 1.26+, a C compiler, and the JavaScriptCore GTK 4.0 **development**
headers:

```sh
# Debian / Ubuntu
sudo apt install libjavascriptcoregtk-4.0-dev golang gcc

# Alpine
sudo apk add webkit2gtk-4.0-dev go gcc musl-dev

git clone https://github.com/kr12345ad/godejs.git
cd godejs
CGO_ENABLED=1 go build -o gode ./cmd/gode
./gode --version
```

The included installer is POSIX `sh` and works under Alpine's `ash`:

```sh
curl -fsSL https://kr12345ad.github.io/godejs/install.sh | sh -s -- --source
```

## ABI and release builds

The package uses cgo and links the system JavaScriptCore:

```go
#cgo LDFLAGS: -l:libjavascriptcoregtk-4.0.so.18
```

This is why releases are Linux-specific and why the binary is dynamically
linked against JSC. A fully static build is possible only if JavaScriptCore is
built statically for the target platform.

## Stack traces

JavaScriptCore exposes a formatted `name@file:line:column` stack string rather
than V8 `CallSite` objects. `internal/env/stack.go` preserves that default
`error.stack` format and activates structured CallSites only when
`Error.prepareStackTrace` is a function. File names, line/column numbers,
script hashes, eval/native flags, and constructor detection are exact;
receiver/function descriptions are reconstructed from the caller's source.
