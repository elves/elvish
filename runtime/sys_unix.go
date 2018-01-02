// +build !windows,!plan9

package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// getSecureRunDir stats elvish-$uid under the default temp dir, creating it if
// it doesn't yet exist, and return the directory name if it has the correct
// owner and permission.
func getSecureRunDir() (string, error) {
	uid := os.Getuid()
	runDir := filepath.Join(os.TempDir(), fmt.Sprintf("elvish-%d", uid))
	err := os.MkdirAll(runDir, 0700)
	if err != nil {
		return "", fmt.Errorf("mkdir: %v", err)
	}

	info, err := os.Stat(runDir)
	if err != nil {
		return "", err
	}

	return runDir, checkExclusiveAccess(info, uid)
}

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
