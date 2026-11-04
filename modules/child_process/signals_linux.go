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

func platformValidSignal(name string) bool {
	switch name {
	case "SIGPWR", "SIGSTKFLT":
		return true
	}
	return false
}

func platformSignalName(sig syscall.Signal) (string, bool) {
	switch sig {
	case syscall.SIGPWR:
		return "SIGPWR", true
	case syscall.SIGSTKFLT:
		return "SIGSTKFLT", true
	}
	return "", false
}
