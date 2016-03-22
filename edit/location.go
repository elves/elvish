package edit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
)

// Location mode.

type location struct {
	listing
	store      *store.Store
	candidates []store.Dir
}

func (loc *location) Len() int {
	return len(loc.candidates)
}

func (loc *location) Show(i, width int) string {
	cand := loc.candidates[i]
	return fmt.Sprintf("%4.0f %s", cand.Score, parse.Quote(cand.Path))
}

func (loc *location) Filter(filter string) int {
	dirs, err := loc.store.FindDirsLoose(filter)
	if err != nil {
		loc.candidates = nil
		// XXX Should report error
		// ed.notify("find directories: %v", err)
		return -1
	}
	loc.candidates = dirs
	if len(dirs) == 0 {
		return -1
	}
	return 0
}

func (loc *location) Accept(i int, ed *Editor) {
	dir := loc.candidates[i].Path
	err := os.Chdir(dir)
	if err == nil {
		store := ed.store
		go func() {
			store.Waits.Add(1)
			// XXX Error ignored.
			store.AddDir(dir, 1)
			store.Waits.Done()
			Logger.Println("added dir to store:", dir)
		}()
	} else {
		ed.notify("%v", err)
	}
	ed.mode = &ed.insert
}

func (loc *location) ModeTitle(i int) string {
	return " LOCATION "
}

func startLocation(ed *Editor) {
	if ed.store == nil {
		ed.notify("%v", ErrStoreOffline)
		return
	}
	loc := &location{store: ed.store}
	loc.listing = newListing(modeLocation, loc)

	ed.location = loc
	ed.mode = ed.location
}
