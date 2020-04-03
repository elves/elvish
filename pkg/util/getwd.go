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
	if home == "" || home == "/" {
		// If home is "" or "/", do not abbreviate because (1) it is likely a
		// problem with the environment and (2) it will make the path actually
		// longer.
		return path
	}
	if err == nil {
		if path == home {
			return "~"
		} else if strings.HasPrefix(path, home+pathSep) {
			return "~" + path[len(home):]
		}
	}
	return path
}
