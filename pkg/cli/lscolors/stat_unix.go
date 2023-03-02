//go:build unix

package lscolors

import (
	"os"
	"syscall"
)

func isMultiHardlink(info os.FileInfo) bool {
	return info.Sys().(*syscall.Stat_t).Nlink > 1
}
