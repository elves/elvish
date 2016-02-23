package util

import (
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestRootStar(t *testing.T) {
	// NOTE: will fail if there are newlines in /*.
	want, err := exec.Command("ls", "/").Output()
	mustOK(err)
	wantNames := strings.Split(strings.Trim(string(want), "\n"), "\n")
	for i := range wantNames {
		wantNames[i] = "/" + wantNames[i]
	}

	names := RootStar()

	sort.Strings(wantNames)
	sort.Strings(names)

	if !reflect.DeepEqual(names, wantNames) {
		t.Errorf("RootStar() -> %s, want %s", names, wantNames)
	}
}
