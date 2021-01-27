package fsutil

import (
	"io/ioutil"
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

// IsExecutable determines whether path refers to an executable file.
func IsExecutable(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	fm := fi.Mode()
	return !fm.IsDir() && (fm&0111 != 0)
}

// EachExternal calls f for each name that can resolve to an external command.
//
// BUG: EachExternal may generate the same command multiple command it it
// appears in multiple directories in PATH.
//
// BUG: EachExternal doesn't work on Windows since it relies on the execution
// permission bit, which doesn't exist on Windows.
func EachExternal(f func(string)) {
	for _, dir := range searchPaths() {
		// TODO(xiaq): Ignore error.
		infos, _ := ioutil.ReadDir(dir)
		for _, info := range infos {
			if !info.IsDir() && (info.Mode()&0111 != 0) {
				f(info.Name())
			}
		}
	}
}

func searchPaths() []string {
	return strings.Split(os.Getenv(env.PATH), ":")
}
