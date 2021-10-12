// +build !windows,!plan9

package fsutil

import "os"

// IsExecutable determines whether path refers to an executable file.
func IsExecutable(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return IsExecutableByInfo(fi)
}

func IsExecutableByInfo(info os.FileInfo) bool {
	fm := info.Mode()
	return !fm.IsDir() && (fm&0111 != 0)
}
