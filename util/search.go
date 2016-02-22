package util

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

var (
	ErrNotExecutable = errors.New("not executable")
	ErrNotFound      = errors.New("not found")
)

// Search tries to resolve an external command and return the full (possibly
// relative) path.
func Search(paths []string, exe string) (string, error) {
	if DontSearch(exe) {
		if IsExecutable(exe) {
			return exe, nil
		}
		return "", ErrNotExecutable
	}
	for _, p := range paths {
		full := p + "/" + exe
		if IsExecutable(full) {
			return full, nil
		}
	}
	return "", ErrNotFound
}

// AllExecutables writes the names of all executable files in the search path
// to a channel.
func AllExecutables(paths []string, names chan<- string) {
	for _, dir := range paths {
		// XXX Ignore error
		infos, _ := ioutil.ReadDir(dir)
		for _, info := range infos {
			if !info.IsDir() && (info.Mode()&0111 != 0) {
				names <- info.Name()
			}
		}
	}
}

// DontSearch determines whether the path to an external command should be
// taken literally and not searched.
func DontSearch(exe string) bool {
	return exe == ".." || strings.ContainsRune(exe, '/')
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
