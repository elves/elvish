package util

import (
	"os"
	"sort"
)

// RootNames returns the result of /*.
func RootStar() []string {
	f, err := os.Open("/")
	if err != nil {
		panic(err)
	}

	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		panic(err)
	}

	var newnames []string
	for _, name := range names {
		if name[0] != '.' {
			newnames = append(newnames, "/"+name)
		}
	}

	sort.Strings(newnames)
	return newnames
}
