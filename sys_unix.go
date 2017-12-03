// +build !windows
// +build !plan9

package main

import (
	"os"
	"syscall"
)

func checkExclusiveAccess(info os.FileInfo, uid int) error {
	stat := info.Sys().(*syscall.Stat_t)
	if int(stat.Uid) != uid {
		return ErrBadOwner
	}
	if stat.Mode&077 != 0 {
		return ErrBadPermission
	}
	return nil
}
