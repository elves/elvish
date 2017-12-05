package util

import (
	"os"
	"path/filepath"
	"strings"
)

var pathSep = string(filepath.Separator)

// Getwd returns path of the working directory in a format suitable as the
// prompt.
func Getwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "?"
	}
	return TildeAbbr(pwd)
}

// TildeAbbr abbreviates the user's home directory to ~.
func TildeAbbr(path string) string {
	home, err := GetHome("")
	if err == nil {
		if path == home {
			return "~"
		} else if strings.HasPrefix(path, home+pathSep) {
			return "~" + path[len(home):]
		}
	}
	return path
}
