//go:build darwin

package fs

import "syscall"

// Darwin has no fdatasync syscall; fsync is the portable equivalent.
func fdatasync(fd int) error { return syscall.Fsync(fd) }
