// +build !windows,!plan9

package util

import (
	"os/exec"
	"sort"
	"strings"
)

func ls(dir string) []string {
	output, err := exec.Command("ls", dir).Output()
	mustOK(err)
	names := strings.Split(strings.Trim(string(output), "\n"), "\n")
	for i := range names {
		names[i] = dir + names[i]
	}
	sort.Strings(names)
	return names
}
