//go:build unix

package fsutil

import (
	"os"
	"os/exec"
)

func isExecutable(stat os.FileInfo) bool {
	return !stat.IsDir() && stat.Mode()&0o111 != 0
}

func searchExecutable(name string) (string, error) {
	return exec.LookPath(name)
}
