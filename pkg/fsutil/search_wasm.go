package fsutil

import "os"

func isExecutable(stat os.FileInfo) bool {
	return true
}
