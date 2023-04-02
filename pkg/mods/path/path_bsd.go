//go:build freebsd || netbsd

package path

import (
	"syscall"
	"time"
)

func pathPlatformMetadata(fi *fileInfo, extra *syscall.Stat_t) {
	fi.ATime = time.Unix(extra.Atimespec.Sec, extra.Atimespec.Nsec)
	fi.BTime = time.Unix(extra.Birthtimespec.Sec, extra.Birthtimespec.Nsec)
	fi.CTime = time.Unix(extra.Ctimespec.Sec, extra.Ctimespec.Nsec)
}
