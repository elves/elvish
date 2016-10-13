package util

import (
	"os"
	"sort"
)

// FullNames returns the full names of non-hidden files under a directory. The
// directory name should end in a slash. If the directory cannot be listed, it
// returns nil.
//
// The output should be the same as globbing dir + "*". It is used for testing
// globbing.
func FullNames(dir string) []string {
	f, err := os.Open(dir)
	if err != nil {
		return nil
	}

	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil
	}

	fullnames := make([]string, 0, len(names))
	for _, name := range names {
		if name[0] != '.' {
			fullnames = append(fullnames, dir+name)
		}
	}

	sort.Strings(fullnames)
	return fullnames
}
