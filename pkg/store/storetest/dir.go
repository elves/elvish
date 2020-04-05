package storetest

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/store"
)

var (
	dirsToAdd  = []string{"/usr/local", "/usr", "/usr/bin", "/usr"}
	black      = map[string]struct{}{"/usr/local": {}}
	wantedDirs = []store.Dir{
		{
			Path:  "/usr",
			Score: store.DirScoreIncrement*store.DirScoreDecay*store.DirScoreDecay + store.DirScoreIncrement,
		},
		{
			Path:  "/usr/bin",
			Score: store.DirScoreIncrement * store.DirScoreDecay,
		},
	}
	dirToDel           = "/usr"
	wantedDirsAfterDel = []store.Dir{
		{
			Path:  "/usr/bin",
			Score: store.DirScoreIncrement * store.DirScoreDecay,
		},
	}
)

// TestDir tests the directory history functionality of a Store.
func TestDir(t *testing.T, tStore store.Store) {
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

	tStore.DelDir("/usr")
	dirs, err = tStore.Dirs(black)
	if err != nil || !reflect.DeepEqual(dirs, wantedDirsAfterDel) {
		t.Errorf(`After DelDir("/usr"), tStore.ListDirs() => (%v, %v), want (%v, <nil>)`,
			dirs, err, wantedDirsAfterDel)
	}
}
