package eval

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/elves/elvish/parse"
)

// Search tries to resolve an external command and return the full (possibly
// relative) path.
func (ev *Evaler) Search(exe string) (string, error) {
	if DontSearch(exe) {
		if IsExecutable(exe) {
			return exe, nil
		}
		return "", fmt.Errorf("external command %s not executable", parse.Quote(exe))
	}
	for _, p := range ev.searchPaths {
		full := p + "/" + exe
		if IsExecutable(full) {
			return full, nil
		}
	}
	return "", fmt.Errorf("external command %s not found", parse.Quote(exe))
}

// AllExecutables writes the names of all executable files in the search path
// to a channel.
func (ev *Evaler) AllExecutables(names chan<- string) {
	for _, dir := range ev.searchPaths {
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
	return exe == ".." ||
		strings.HasPrefix(exe, "/") ||
		strings.HasPrefix(exe, "./") ||
		strings.HasPrefix(exe, "../")
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
