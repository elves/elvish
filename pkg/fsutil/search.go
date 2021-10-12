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
	return exe == ".." || strings.ContainsRune(exe, filepath.Separator) ||
		strings.ContainsRune(exe, '/')
}

// EachExternal calls f for each name that can resolve to an external command.
//
// BUG: EachExternal may generate the same command multiple command it it
// appears in multiple directories in PATH.
func EachExternal(f func(string)) {
	for _, dir := range searchPaths() {
		// TODO(xiaq): Ignore error.
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			info, err := file.Info()
			if err == nil && IsExecutableByInfo(info) {
				f(info.Name())
			}
		}
	}
}

func searchPaths() []string {
	return strings.Split(os.Getenv(env.PATH), string(filepath.ListSeparator))
}
