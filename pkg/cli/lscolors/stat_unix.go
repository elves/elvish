//go:build unix

package lscolors

import (
	"os"
	"syscall"
)

func isMultiHardlink(info os.FileInfo) bool {
	// The nlink field from stat considers all the "." and ".." references to
	// directories to be hard links, making all directories technically
	// multi-hardlink (one link from parent, one "." from itself, and one ".."
	// for every subdirectories). However, for the purpose of filename
	// highlighting, only regular files should ever be considered
	// multi-hardlink.
	return !info.IsDir() && info.Sys().(*syscall.Stat_t).Nlink > 1
}
