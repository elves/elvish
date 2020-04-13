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
	runDirs := getRunDirPaths()
	for _, runDir := range runDirs {
		if runDirExistsWithExclusiveAccess(runDir) {
			return runDir, nil
		}
	}

	runDir := runDirs[0]
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

// getRunDirPaths returns an array of paths that's used by Elvish to determine
// where to store the runtime files. The paths are sorted in descending order
// of preference and will always contain at least one path that points to
// TMPDIR/elvish-$uid as the last item in the array.
// When XDG_RUNTIME_DIR is set, the XDG_RUNTIME_DIR/elvish path will be the
// first entry in the arrray.
func getRunDirPaths() []string {
	tmpDirPath := filepath.Join(os.TempDir(), fmt.Sprintf("elvish-%d", os.Getuid()))
	if os.Getenv("XDG_RUNTIME_DIR") != "" {
		xdgDirPath := filepath.Join(os.Getenv("XDG_RUNTIME_DIR"), "elvish")
		return []string{xdgDirPath, tmpDirPath}
	}

	return []string{tmpDirPath}
}

func runDirExistsWithExclusiveAccess(runDir string) bool {
	info, err := os.Stat(runDir)
	if err != nil {
		return false
	}

	err = checkExclusiveAccess(info, os.Getuid())
	return err == nil
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
