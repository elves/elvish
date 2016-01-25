package eval

import (
	"fmt"
	"os"
	"strings"

	"github.com/elves/elvish/parse"
)

// Search tries to resolve an external command and return the full (possibly
// relative) path.
func (ev *Evaler) Search(exe string) (string, error) {
	for _, p := range []string{"/", "./", "../"} {
		if strings.HasPrefix(exe, p) {
			if IsExecutable(exe) {
				return exe, nil
			}
			return "", fmt.Errorf("external command %s not executable", parse.Quote(exe))
		}
	}
	for _, p := range ev.searchPaths {
		full := p + "/" + exe
		if IsExecutable(full) {
			return full, nil
		}
	}
	return "", fmt.Errorf("external command %s not found", parse.Quote(exe))
}

// IsExecutable determines whether path refers to an executable file.
func IsExecutable(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false
	}
	fm := fi.Mode()
	return !fm.IsDir() && (fm&0111 != 0)
}
