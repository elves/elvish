//go:build !windows && !plan9
// +build !windows,!plan9

package fsutil

import "os"

// IsExecutable returns true if the stat object refers to an executable file on UNIX platforms.
func IsExecutable(stat os.FileInfo) bool {
	return !stat.IsDir() && (stat.Mode()&0o111 != 0)
}
