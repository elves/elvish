package store

import (
	"reflect"
	"testing"
)

var (
	dirsToAdd  = []string{"/usr", "/usr/bin", "/usr"}
	wantedDirs = []Dir{Dir{"/usr", 20}, Dir{"/usr/bin", 10}}
)

func TestDir(t *testing.T) {
	for _, path := range dirsToAdd {
		err := tStore.AddDir(path)
		if err != nil {
			t.Errorf("tStore.AddDir(%q) => %v, want <nil>", path, err)
		}
	}

	dirs, err := tStore.FindDirs("usr")
	if err != nil || !reflect.DeepEqual(dirs, wantedDirs) {
		t.Errorf(`tStore.FindDirs("usr") => (%v, %v), want (%v, <nil>)`,
			dirs, err, wantedDirs)
	}
}
