package util

import (
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func ls(dir string) []string {
	// BUG: will fail if there are filenames containing newlines.
	output, err := exec.Command("ls", dir).Output()
	mustOK(err)
	names := strings.Split(strings.Trim(string(output), "\n"), "\n")
	for i := range names {
		names[i] = dir + names[i]
	}
	sort.Strings(names)
	return names
}

func TestFullNames(t *testing.T) {
	for _, dir := range []string{"/", "/usr/"} {
		wantNames := ls(dir)
		names := FullNames(dir)
		if !reflect.DeepEqual(names, wantNames) {
			t.Errorf(`FullNames(%q) -> %s, want %s`, dir, names, wantNames)
		}
	}
}
