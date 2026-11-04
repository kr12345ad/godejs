//go:build darwin

package fs

import (
	"syscall"
	"time"
)

// statTimes returns the access, modification, change, and creation times.
// Unlike Linux, Darwin keeps a real birth time in struct stat.
func statTimes(st *syscall.Stat_t) (atime, mtime, ctime, birthtime time.Time) {
	atime = time.Unix(st.Atimespec.Sec, st.Atimespec.Nsec)
	mtime = time.Unix(st.Mtimespec.Sec, st.Mtimespec.Nsec)
	ctime = time.Unix(st.Ctimespec.Sec, st.Ctimespec.Nsec)
	birthtime = time.Unix(st.Birthtimespec.Sec, st.Birthtimespec.Nsec)
	return atime, mtime, ctime, birthtime
}
