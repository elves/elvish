package fsutil

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"src.elv.sh/pkg/env"
)

// isExecutable determines whether pathStr refers to an executable file in PATH. It honors E:PATHEXT
// but defaults to extensions ".com", ".exe", ".bat", ".cmd" if that env var isn't set.
func isExecutable(pathStr string) bool {
	var validExts []string
	if pathext := os.Getenv(env.PATHEXT); pathext != "" {
		for _, e := range strings.Split(strings.ToLower(pathext), string(filepath.ListSeparator)) {
			if e == "" {
				continue
			}
			if e[0] != '.' {
				e = "." + e
			}
			validExts = append(validExts, e)
		}
	} else {
		validExts = []string{".com", ".exe", ".bat", ".cmd"}
	}

	ext := strings.ToLower(path.Ext(pathStr))
	for _, valid := range validExts {
		if ext == valid {
			return true
		}
	}
	return false
}

// IsExecutable returns true if the stat object refers to a valid executable on Windows.
// Note that on Windows file permissions are not checked but we do validate the file is not a
// directory and has a recognized extension.
func IsExecutable(stat os.FileInfo) bool {
	return !stat.IsDir() && isExecutable(stat.Name())
}
