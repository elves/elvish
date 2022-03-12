//go:build !windows && !plan9
// +build !windows,!plan9

package fsutil

import "os"

func isExecutable(stat os.FileInfo) bool {
	return !stat.IsDir() && stat.Mode()&0o111 != 0
}
