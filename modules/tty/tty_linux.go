//go:build linux

package tty

import (
	"syscall"
	"unsafe"
)

func isTerminal(fd int) bool {
	var t syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&t)))
	return errno == 0
}

func windowSize(fd int) (cols, rows int, ok bool) {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 0, 0, false
	}
	return int(ws.Cols), int(ws.Rows), true
}

func setRawMode(fd int, raw bool) bool {
	var t syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&t))); errno != 0 {
		return false
	}
	if raw {
		t.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.IEXTEN | syscall.ISIG
		t.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
		t.Cc[syscall.VMIN] = 1
		t.Cc[syscall.VTIME] = 0
	} else {
		t.Lflag |= syscall.ICANON | syscall.ECHO | syscall.IEXTEN | syscall.ISIG
		t.Iflag |= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&t)))
	return errno == 0
}
