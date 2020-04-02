// +build !windows,!plan9

package shell

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

var (
	errBadOwner      = errors.New("bad owner")
	errBadPermission = errors.New("bad permission")
)

// getSecureRunDir stats elvish under the directory defined by the
// XDG_RUNTIME_DIR environment variable, creating it if it doesn't yet exist,
// and return the directory name if it has the correct owner and permission.
// If XDG_RUNTIME_DIR is not set it falls back to elvish-$uid under the default
// temp dir.
func getSecureRunDir() (string, error) {
	runDir := getRunDirPath()
	err := os.MkdirAll(runDir, 0700)
	if err != nil {
		return "", fmt.Errorf("mkdir: %v", err)
	}

	info, err := os.Stat(runDir)
	if err != nil {
		return "", err
	}

	return runDir, checkExclusiveAccess(info, os.Getuid())
}

// getRunDirPath returns a path that's used by Elvish to store runtime files.
// This path will be either XDG_RUNTIME_DIR/elvish when the XDG_RUNTIME_DIR
// environment variable is set, or TMPDIR/elvish-$uid otherwise.
func getRunDirPath() string {
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		return filepath.Join(os.Getenv("XDG_RUNTIME_DIR"), "elvish")
	}

	uid := os.Getuid()
	return filepath.Join(os.TempDir(), fmt.Sprintf("elvish-%d", uid))
}

func checkExclusiveAccess(info os.FileInfo, uid int) error {
	stat := info.Sys().(*syscall.Stat_t)
	if int(stat.Uid) != uid {
		return errBadOwner
	}
	if stat.Mode&077 != 0 {
		return errBadPermission
	}
	return nil
}
