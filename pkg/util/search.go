package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DontSearch determines whether the path to an external command should be
// taken literally and not searched.
func DontSearch(exe string) bool {
	return exe == ".." || strings.ContainsRune(exe, filepath.Separator) ||
		strings.ContainsRune(exe, '/')
}

// IsExecutable determines whether path refers to an executable file.
func IsExecutable(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	fm := fi.Mode()
	return !fm.IsDir() && IsExecutableFile(fi)
}

var osHasNoExecutableBit bool

func init() {
	osHasNoExecutableBit = runtime.GOOS == "windows"
}

// IsExecutableFile returns true if the item denoted by info is executable on the runtime platform
func IsExecutableFile(info os.FileInfo) bool {
	return (info.Mode()&0111 != 0 || osHasNoExecutableBit)
}
