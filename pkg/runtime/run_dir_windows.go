package runtime

import (
	"fmt"
	"os"
	"path/filepath"
)

// getSecureRunDir stats elvish-$USERNAME under the default temp dir, creating
// it if it doesn't yet exist, and return the directory name.
func getSecureRunDir() (string, error) {
	username := os.Getenv("USERNAME")

	runDir := filepath.Join(os.TempDir(), "elvish-"+username)
	err := os.MkdirAll(runDir, 0700)
	if err != nil {
		return "", fmt.Errorf("mkdir: %v", err)
	}

	return runDir, nil
}
