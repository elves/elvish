package util

import (
	"reflect"
	"runtime"
	"testing"
)

func TestFullNames(t *testing.T) {
	var dirs []string
	if runtime.GOOS == "windows" {
		dirs = []string{`C:\`, `C:\Users\`}
	} else {
		dirs = []string{"/", "/usr"}
	}
	for _, dir := range dirs {
		wantNames := ls(dir)
		names := FullNames(dir)
		if !reflect.DeepEqual(names, wantNames) {
			t.Errorf(`FullNames(%q) -> %s, want %s`, dir, names, wantNames)
		}
	}
}
