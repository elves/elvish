//go:build linux

package os

import (
	"syscall"
	"time"
)

func pathOsMetadata(fi *fileInfo, extra *syscall.Stat_t) {
	fi.ATime = time.Unix(int64(extra.Atim.Sec), int64(extra.Atim.Nsec))
	fi.CTime = time.Unix(int64(extra.Ctim.Sec), int64(extra.Ctim.Nsec))
}
