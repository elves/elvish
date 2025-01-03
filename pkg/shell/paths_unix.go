//go:build unix

package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/fsutil"
)

func defaultConfigHome() (string, error) { return homePath(".config") }

func defaultDataHome() (string, error) { return homePath(".local/share") }

var defaultDataDirs = []string{
	"/usr/local/share/elvish/lib",
	"/usr/share/elvish/lib",
}

func defaultStateHome() (string, error) { return homePath(".local/state") }

func homePath(suffix string) (string, error) {
	home, err := fsutil.GetHome("")
	if err != nil {
		return "", fmt.Errorf("resolve ~/%s: %w", suffix, err)
	}
	return filepath.Join(home, suffix), nil
}

// Returns a "run directory" for storing ephemeral files, which is guaranteed
// to be only accessible to the current user. The path is one of the following:
//
//   - $XDG_RUNTIME_DIR/elvish if $XDG_RUNTIME_DIR is non-empty.
//   - $tmpdir/elvish-$uid otherwise.
func secureRunDir() (string, error) {
	runDir := runDirPath()
	info, err := os.Stat(runDir)
	if err == nil {
		// Already exists; just check if it's secure.
		if !secureAsRunDir(info) {
			return "", fmt.Errorf("existing run directory %v is not secure", runDir)
		}
		return runDir, nil
	}
	// Create new run directory.
	err = os.MkdirAll(runDir, 0700)
	if err != nil {
		return "", fmt.Errorf("create new run directory: %v", err)
	}
	// The OS may have set a different owner or permission bits, so still check
	// if it's secure.
	info, err = os.Stat(runDir)
	if err != nil {
		return "", fmt.Errorf("stat newly created run directory: %v", err)
	}
	if !secureAsRunDir(info) {
		return "", fmt.Errorf("newly created run directory %v is not secure", runDir)
	}
	return runDir, nil
}

func runDirPath() string {
	if os.Getenv(env.XDG_RUNTIME_DIR) != "" {
		return filepath.Join(os.Getenv(env.XDG_RUNTIME_DIR), "elvish")
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("elvish-%d", os.Getuid()))
}

func secureAsRunDir(info os.FileInfo) bool {
	stat := info.Sys().(*syscall.Stat_t)
	return info.IsDir() && int(stat.Uid) == os.Getuid() && stat.Mode&077 == 0
}
