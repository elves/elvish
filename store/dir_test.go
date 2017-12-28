package store

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/store/storedefs"
)

var (
	dirsToAdd  = []string{"/usr/local", "/usr", "/usr/bin", "/usr"}
	black      = map[string]struct{}{"/usr/local": {}}
	wantedDirs = []storedefs.Dir{
		{"/usr", scoreIncrement*scoreDecay*scoreDecay + scoreIncrement},
		{"/usr/bin", scoreIncrement * scoreDecay}}
)

func TestDir(t *testing.T) {
	for _, path := range dirsToAdd {
		err := tStore.AddDir(path, 1)
		if err != nil {
			t.Errorf("tStore.AddDir(%q) => %v, want <nil>", path, err)
		}
	}

	dirs, err := tStore.Dirs(black)
	if err != nil || !reflect.DeepEqual(dirs, wantedDirs) {
		t.Errorf(`tStore.ListDirs() => (%v, %v), want (%v, <nil>)`,
			dirs, err, wantedDirs)
	}
}
