//go:build openbsd

package path

import (
	"syscall"
	"time"
)

func pathPlatformMetadata(fi *fileInfo, extra *syscall.Stat_t) {
	fi.ATime = time.Unix(extra.Atim.Unix())
	fi.CTime = time.Unix(extra.Ctim.Unix())
}
