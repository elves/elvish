//go:build windows

package fsutil

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
	"src.elv.sh/pkg/env"
)

var (
	DefaultConfigHome = roamingAppData
	DefaultDataHome   = localAppData
	DefaultDataDirs   = []string{}
	DefaultStateHome  = localAppData
)

func roamingAppData() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_RoamingAppData, windows.KF_FLAG_CREATE)
}

func localAppData() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_LocalAppData, windows.KF_FLAG_CREATE)
}

// SecureRunDir stats elvish-$USERNAME under the default temp dir, creating it if it doesn't yet
// exist, and return the directory name.
func SecureRunDir() (string, error) {
	username := os.Getenv(env.USERNAME)

	runDir := filepath.Join(os.TempDir(), "elvish-"+username)
	err := os.MkdirAll(runDir, 0700)
	if err != nil {
		return "", fmt.Errorf("mkdir: %v", err)
	}

	return runDir, nil
}
