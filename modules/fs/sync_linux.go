//go:build linux

package fs

import "syscall"

// fdatasync flushes file data without forcing unrelated metadata updates.
func fdatasync(fd int) error { return syscall.Fdatasync(fd) }
