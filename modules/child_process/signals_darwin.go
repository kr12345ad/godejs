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
