package web

// navigator: Node has exposed a Navigator global since v21. GodeJS matches its
// shape and values (a Navigator instance with userAgent, platform,
// hardwareConcurrency, language and languages) using host facts rather than
// process APIs, so it works before any require() call.

import (
	"os"
	"runtime"
	"strings"

	"godejs/internal/env"
	"godejs/internal/js"
	"godejs/internal/version"
)

// nodeUserAgent matches Node's own value: "Node.js/<major>".
func nodeUserAgent() string {
	major := version.Node
	if i := strings.IndexByte(major, '.'); i >= 0 {
		major = major[:i]
	}
	return "Node.js/" + major
}

// navigatorPlatform matches uname's "sysname machine" form, which is what
// Node reports ("Linux x86_64").
func navigatorPlatform() string {
	osName := map[string]string{
		"linux": "Linux", "darwin": "Darwin", "windows": "Windows",
		"freebsd": "FreeBSD", "openbsd": "OpenBSD", "netbsd": "NetBSD",
		"sunos": "SunOS", "aix": "AIX",
	}[runtime.GOOS]
	if osName == "" {
		osName = runtime.GOOS
	}
	arch := map[string]string{
		"amd64": "x86_64", "arm64": "aarch64", "386": "i686", "arm": "armv7l",
		"ppc64le": "ppc64le", "s390x": "s390x", "riscv64": "riscv64",
	}[runtime.GOARCH]
	if arch == "" {
		arch = runtime.GOARCH
	}
	return osName + " " + arch
}

// navigatorLanguages derives BCP-47 tags from POSIX locale variables. ICU
// resolves these from the same environment, so this reproduces Node's values
// without linking ICU.
func navigatorLanguages() (string, []string) {
	raw := os.Getenv("LC_ALL")
	if raw == "" {
		raw = os.Getenv("LC_MESSAGES")
	}
	if raw == "" {
		raw = os.Getenv("LANG")
	}
	raw = strings.SplitN(raw, ".", 2)[0]
	if i := strings.IndexByte(raw, '@'); i >= 0 {
		raw = raw[:i]
	}
	raw = strings.ReplaceAll(raw, "_", "-")
	if raw == "" || raw == "C" || raw == "POSIX" || raw == "C-UTF-8" {
		return "en-US", []string{"en-US"}
	}
	// Drop an unknown territory suffix instead of inventing a language.
	parts := strings.Split(raw, "-")
	parts[0] = strings.ToLower(parts[0])
	if len(parts) > 1 {
		parts[1] = strings.ToUpper(parts[1])
	}
	tag := strings.Join(parts[:min(len(parts), 2)], "-")
	languages := []string{tag}
	if base := parts[0]; base != tag {
		languages = append(languages, base)
	}
	return tag, languages
}

// installNavigator defines the navigator global.
func installNavigator(e *env.Env) {
	lang, languages := navigatorLanguages()

	nav := e.JS.Obj()
	nav.SetStr("userAgent", nodeUserAgent())
	nav.SetStr("platform", navigatorPlatform())
	nav.Set("hardwareConcurrency", e.JS.Num(float64(runtime.NumCPU())))
	nav.SetStr("language", lang)
	values := make([]js.Value, len(languages))
	for i, tag := range languages {
		values[i] = e.JS.Str(tag)
	}
	nav.Set("languages", e.JS.Arr(values).V())

	e.Global().Set("__godeNavigator", nav.V())
	if _, err := e.JS.Eval(navigatorPrelude, "gode:navigator"); err != nil {
		panic("navigator: " + err.Error())
	}
}

const navigatorPrelude = `(function (v) {
  'use strict';
  delete globalThis.__godeNavigator;

  class Navigator {
    get userAgent() { return v.userAgent; }
    get platform() { return v.platform; }
    get hardwareConcurrency() { return v.hardwareConcurrency; }
    get language() { return v.language; }
    get languages() { return v.languages; }
  }

  var navigator = new Navigator();
  Object.defineProperty(globalThis, 'navigator', {
    value: navigator, writable: false, enumerable: true, configurable: true
  });
})(globalThis.__godeNavigator);`
