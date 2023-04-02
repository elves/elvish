//go:build openbsd

package os

import (
	"syscall"
	"time"
)

func pathOsMetadata(fi *fileInfo, extra *syscall.Stat_t) {
	fi.ATime = time.Unix(extra.Atim.Unix())
	fi.CTime = time.Unix(extra.Ctim.Unix())
}
