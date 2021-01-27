// +build !windows,!plan9

package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"src.elv.sh/pkg/env"
)

// Returns a "run directory" for storing ephemeral files, which is guaranteed
// to be only accessible to the current user.
//
// The path of the run directory is either $XDG_RUNTIME_DIR/elvish or
// $tmpdir/elvish-$uid (where $tmpdir is the system temporary directory). The
// former is used if the XDG_RUNTIME_DIR environment variable exists and the
// latter directory does not exist.
func getSecureRunDir() (string, error) {
	runDirs := getRunDirCandidates()
	for _, runDir := range runDirs {
		if checkExclusiveAccess(runDir) {
			return runDir, nil
		}
	}
	runDir := runDirs[0]
	err := os.MkdirAll(runDir, 0700)
	if err != nil {
		return "", fmt.Errorf("mkdir: %v", err)
	}
	if !checkExclusiveAccess(runDir) {
		return "", fmt.Errorf("cannot create %v as a secure run directory", runDir)
	}
	return runDir, nil
}

// Returns one or more candidates for the run directory, in descending order of
// preference.
func getRunDirCandidates() []string {
	tmpDirPath := filepath.Join(os.TempDir(), fmt.Sprintf("elvish-%d", os.Getuid()))
	if os.Getenv(env.XDG_RUNTIME_DIR) != "" {
		xdgDirPath := filepath.Join(os.Getenv(env.XDG_RUNTIME_DIR), "elvish")
		return []string{xdgDirPath, tmpDirPath}
	}
	return []string{tmpDirPath}
}

func checkExclusiveAccess(runDir string) bool {
	info, err := os.Stat(runDir)
	if err != nil {
		return false
	}
	stat := info.Sys().(*syscall.Stat_t)
	return info.IsDir() && int(stat.Uid) == os.Getuid() && stat.Mode&077 == 0
}
