//go:build !linux

// Package napi registers the Node-API addon loader. Native addons need dlopen
// and the exported Node-API entry point, which GodeJS only implements on
// Linux, so other platforms report a clear error instead of linking the C ABI.
package napi

import (
	"godejs/internal/env"
	"godejs/internal/js"
)

func init() {
	env.SetNativeLoader(func(*env.Env, string) (js.Object, error) {
		return js.Object{}, &js.Error{
			Name:    "Error",
			Code:    "ERR_DLOPEN_FAILED",
			Message: "ERR_DLOPEN_FAILED: Node-API addons are only supported on Linux",
		}
	})
}
