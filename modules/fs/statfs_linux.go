//go:build linux

package fs

import "syscall"

type statfsValues struct {
	fsType, bsize, frsize, blocks, bfree, bavail, files, ffree int64
}

func statfsFields(st *syscall.Statfs_t) statfsValues {
	return statfsValues{
		fsType: int64(st.Type),
		bsize:  int64(st.Bsize),
		frsize: int64(st.Frsize),
		blocks: int64(st.Blocks),
		bfree:  int64(st.Bfree),
		bavail: int64(st.Bavail),
		files:  int64(st.Files),
		ffree:  int64(st.Ffree),
	}
}
