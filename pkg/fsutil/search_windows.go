package fsutil

import (
	"os"
	"path/filepath"
	"strings"

	"src.elv.sh/pkg/env"
)

func isExecutable(stat os.FileInfo) bool {
	return !stat.IsDir() && isExecutableExt(filepath.Ext(stat.Name()))
}

// Determines determines a file name extension is considered executable.
// It honors PATHEXT but defaults to extensions ".com", ".exe", ".bat", ".cmd"
// if that env var isn't set.
func isExecutableExt(ext string) bool {
	validExts := make(map[string]bool)
	if pathext := os.Getenv(env.PATHEXT); pathext != "" {
		for _, e := range filepath.SplitList(strings.ToLower(pathext)) {
			if e == "" {
				continue
			}
			if e[0] != '.' {
				e = "." + e
			}
			validExts[e] = true
		}
	} else {
		validExts = map[string]bool{
			".com": true, ".exe": true, ".bat": true, ".cmd": true}
	}

	return validExts[strings.ToLower(ext)]
}
