//go:build linux

package js

import (
	"os"
	"sync"
	"syscall"
)

// JavaScriptCore installs a SIGUSR1 handler for GC and logs a notice when it
// replaces an existing one (Go installs handlers at startup). The notice goes
// straight to fd 2, so it is silenced only for the duration of context
// creation.
var stderrMu sync.Mutex

func silently(fn func()) {
	stderrMu.Lock()
	defer stderrMu.Unlock()

	fd := int(os.Stderr.Fd())
	saved, err := syscall.Dup(fd)
	if err != nil {
		fn()
		return
	}
	null, err := syscall.Open(os.DevNull, syscall.O_WRONLY, 0)
	if err != nil {
		_ = syscall.Close(saved)
		fn()
		return
	}
	if err := syscall.Dup2(null, fd); err != nil {
		_ = syscall.Close(saved)
		_ = syscall.Close(null)
		fn()
		return
	}
	fn()
	_ = syscall.Dup2(saved, fd)
	_ = syscall.Close(saved)
	_ = syscall.Close(null)
}
