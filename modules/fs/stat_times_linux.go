//go:build linux

package fs

import (
	"syscall"
	"time"
)

// statTimes returns the access, modification, change, and creation times.
// Linux has no creation time in stat(2), so birthtime follows ctime.
func statTimes(st *syscall.Stat_t) (atime, mtime, ctime, birthtime time.Time) {
	atime = time.Unix(st.Atim.Sec, st.Atim.Nsec)
	mtime = time.Unix(st.Mtim.Sec, st.Mtim.Nsec)
	ctime = time.Unix(st.Ctim.Sec, st.Ctim.Nsec)
	return atime, mtime, ctime, ctime
}
