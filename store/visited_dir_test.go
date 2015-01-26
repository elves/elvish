package store

import (
	"reflect"
	"testing"
)

var (
	dirsToAdd  = []string{"/usr", "/usr/bin", "/usr"}
	wantedDirs = []VisitedDir{VisitedDir{"/usr", 20}, VisitedDir{"/usr/bin", 10}}
)

func TestDir(t *testing.T) {
	for _, path := range dirsToAdd {
		err := tStore.AddVisistedDir(path)
		if err != nil {
			t.Errorf("tStore.AddVisistedDir(%q) => %v, want <nil>", path, err)
		}
	}

	dirs, err := tStore.FindVisitedDirs("usr")
	if err != nil || !reflect.DeepEqual(dirs, wantedDirs) {
		t.Errorf(`tStore.FindVisistedDirs("usr") => (%v, %v), want (%v, <nil>)`,
			dirs, err, wantedDirs)
	}
}
