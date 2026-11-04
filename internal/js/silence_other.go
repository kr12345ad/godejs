//go:build !linux

package js

func silently(fn func()) { fn() }
