//go:build darwin

package child_process

import "syscall"

func platformSignal(name string) (syscall.Signal, bool) {
	switch name {
	case "SIGINFO":
		return syscall.SIGINFO, true
	case "SIGEMT":
		return syscall.SIGEMT, true
	}
	return 0, false
}

func platformValidSignal(name string) bool {
	switch name {
	case "SIGINFO", "SIGEMT":
		return true
	}
	return false
}

func platformSignalName(sig syscall.Signal) (string, bool) {
	switch sig {
	case syscall.SIGINFO:
		return "SIGINFO", true
	case syscall.SIGEMT:
		return "SIGEMT", true
	}
	return "", false
}
