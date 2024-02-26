package fsutil

import (
	"os"
	"path/filepath"
	"strings"

	"src.elv.sh/pkg/env"
)

// DontSearch determines whether the path to an external command should be
// taken literally and not searched.
func DontSearch(exe string) bool {
	// TODO: Remove ".." after implicit cd is removed.
	return exe == ".." || strings.ContainsRune(exe, filepath.Separator) ||
		strings.ContainsRune(exe, '/')
}

// IsExecutable returns whether the FileInfo refers to an executable file.
//
// This is determined by permission bits on Unix, and by file name on Windows.
func IsExecutable(stat os.FileInfo) bool {
	return isExecutable(stat)
}

// EachExternal calls f for each executable file found while scanning the directories of $E:PATH.
//
// NOTE: EachExternal may generate the same command multiple times; once for each time it appears in
// $E:PATH. That is, no deduplication of the files found by scanning $E:PATH is performed.
func EachExternal(f func(string)) {
	for _, dir := range searchPaths() {
		files, err := os.ReadDir(dir)
		if err != nil {
			// In practice this rarely happens. There isn't much we can reasonably do when it does
			// happen other than silently ignore the invalid directory.
			continue
		}
		for _, file := range files {
			stat, err := file.Info()
			if err == nil && IsExecutable(stat) {
				f(stat.Name())
			}
		}
	}
}

func searchPaths() []string {
	return strings.Split(os.Getenv(env.PATH), string(filepath.ListSeparator))
}
