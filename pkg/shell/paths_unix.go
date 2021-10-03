//go:build !windows && !plan9
// +build !windows,!plan9

package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/fsutil"
)

func newRCPath() (string, error) {
	return xdgHomePath(env.XDG_CONFIG_HOME, ".config", "elvish/rc.elv")
}

const elvishLib = "elvish/lib"

func newLibPaths() ([]string, error) {
	var paths []string
	libConfig, errConfig := xdgHomePath(env.XDG_CONFIG_HOME, ".config", elvishLib)
	if errConfig == nil {
		paths = append(paths, libConfig)
	}
	libData, errData := xdgHomePath(env.XDG_DATA_HOME, ".local/share", elvishLib)
	if errData == nil {
		paths = append(paths, libData)
	}

	libSystem := os.Getenv(env.XDG_DATA_DIRS)
	if libSystem == "" {
		libSystem = "/usr/local/share:/usr/share"
	}
	for _, p := range filepath.SplitList(libSystem) {
		paths = append(paths, filepath.Join(p, elvishLib))
	}

	return paths, diag.Errors(errConfig, errData)
}

func newDBPath() (string, error) {
	return xdgHomePath(env.XDG_STATE_HOME, ".local/state", "elvish/db.bolt")
}

func xdgHomePath(envName, fallback, suffix string) (string, error) {
	dir := os.Getenv(envName)
	if dir == "" {
		home, err := fsutil.GetHome("")
		if err != nil {
			return "", fmt.Errorf("resolve ~/%s/%s: %w", fallback, suffix, err)
		}
		dir = filepath.Join(home, fallback)
	}
	return filepath.Join(dir, suffix), nil
}

// Returns a "run directory" for storing ephemeral files, which is guaranteed
// to be only accessible to the current user.
//
// The path of the run directory is either $XDG_RUNTIME_DIR/elvish or
// $tmpdir/elvish-$uid (where $tmpdir is the system temporary directory). The
// former is used if the XDG_RUNTIME_DIR environment variable exists and the
// latter directory does not exist.
func secureRunDir() (string, error) {
	runDirs := runDirCandidates()
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
func runDirCandidates() []string {
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
