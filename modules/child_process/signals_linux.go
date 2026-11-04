//go:build linux

package child_process

import "syscall"

func platformSignal(name string) (syscall.Signal, bool) {
	switch name {
	case "SIGPWR":
		return syscall.SIGPWR, true
	case "SIGSTKFLT":
		return syscall.SIGSTKFLT, true
	}
	return 0, false
}
