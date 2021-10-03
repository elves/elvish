package shell

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
	"src.elv.sh/pkg/env"
)

func newRCPath() (string, error) {
	d, err := roamingAppData()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "elvish", "rc.elv"), nil
}

func newLibPaths() ([]string, error) {
	local, err := localAppData()
	if err != nil {
		return nil, err
	}
	localLib := filepath.Join(local, "elvish", "lib")

	roaming, err := roamingAppData()
	if err != nil {
		return nil, err
	}
	roamingLib := filepath.Join(roaming, "elvish", "lib")

	return []string{roamingLib, localLib}, nil
}

func newDBPath() (string, error) {
	d, err := localAppData()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "elvish", "db.bolt"), nil
}

func localAppData() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_LocalAppData, windows.KF_FLAG_CREATE)
}

func roamingAppData() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_RoamingAppData, windows.KF_FLAG_CREATE)
}

// getSecureRunDir stats elvish-$USERNAME under the default temp dir, creating
// it if it doesn't yet exist, and return the directory name.
func secureRunDir() (string, error) {
	username := os.Getenv(env.USERNAME)

	runDir := filepath.Join(os.TempDir(), "elvish-"+username)
	err := os.MkdirAll(runDir, 0700)
	if err != nil {
		return "", fmt.Errorf("mkdir: %v", err)
	}

	return runDir, nil
}
