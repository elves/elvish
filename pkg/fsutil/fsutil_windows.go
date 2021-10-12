package fsutil

import (
	"os"
	"path"
)

// IsExecutable determines whether path refers to an executable file.
func IsExecutable(pathStr string) bool {
	ext := path.Ext(pathStr)
	return ext == ".exe" || ext == ".cmd" || ext == ".bat"
}

func IsExecutableByInfo(info os.FileInfo) bool {
	return IsExecutable(info.Name())
}
