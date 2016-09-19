package store

import (
	"reflect"
	"testing"
)

var (
	dirsToAdd  = []string{"/usr", "/usr/bin", "/usr"}
	wantedDirs = []Dir{
		Dir{"/usr", scoreIncrement*scoreDecay*scoreDecay + scoreIncrement},
		Dir{"/usr/bin", scoreIncrement * scoreDecay}}
)

func TestDir(t *testing.T) {
	for _, path := range dirsToAdd {
		err := tStore.AddDir(path, 1)
		if err != nil {
			t.Errorf("tStore.AddDir(%q) => %v, want <nil>", path, err)
		}
	}

	dirs, err := tStore.ListDirs()
	if err != nil || !reflect.DeepEqual(dirs, wantedDirs) {
		t.Errorf(`tStore.ListDirs() => (%v, %v), want (%v, <nil>)`,
			dirs, err, wantedDirs)
	}
}
