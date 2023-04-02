//go:build freebsd || netbsd

package os

import (
	"syscall"
	"time"
)

func pathOsMetadata(fi *fileInfo, extra *syscall.Stat_t) {
	fi.ATime = time.Unix(extra.Atimespec.Sec, extra.Atimespec.Nsec)
	fi.BTime = time.Unix(extra.Birthtimespec.Sec, extra.Birthtimespec.Nsec)
	fi.CTime = time.Unix(extra.Ctimespec.Sec, extra.Ctimespec.Nsec)
}
