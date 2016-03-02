package edit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
)

type location struct {
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
	dirs, err := loc.store.FindDirsSubseq(filter)
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
			store.AddDir(dir, 0.5)
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

// Editor builtins.

func startLocation(ed *Editor) {
	if ed.store == nil {
		ed.notify("%v", ErrStoreOffline)
		return
	}
	loc := &location{ed.store, nil}

	ed.location = listing{modeLocation, loc, 0, ""}
	ed.location.changeFilter("")
	ed.mode = &ed.location
}

func locationPrev(ed *Editor) {
	ed.location.prev(false)
}

func locationCyclePrev(ed *Editor) {
	ed.location.prev(true)
}

func locationNext(ed *Editor) {
	ed.location.next(false)
}

func locationCycleNext(ed *Editor) {
	ed.location.next(true)
}

func locationBackspace(ed *Editor) {
	ed.location.backspace()
}

func acceptLocation(ed *Editor) {
	ed.location.accept(ed)
}

func locationDefault(ed *Editor) {
	ed.location.defaultBinding(ed)
}
