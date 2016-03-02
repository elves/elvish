package edit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
)

type location struct {
	listing
	candidates []store.Dir
}

func (*location) Mode() ModeType {
	return modeLocation
}

func (l *location) ModeLine(width int) *buffer {
	return l.modeLine(" LOCATION ", width)
}

func (l *location) update(ed *Editor) bool {
	dirs, err := ed.store.FindDirsSubseq(l.filter)
	if err != nil {
		l.candidates = nil
		l.selected = -1
		ed.notify("find directories: %v", err)
		return false
	}
	l.candidates = dirs

	if len(l.candidates) > 0 {
		l.selected = 0
	} else {
		l.selected = -1
	}
	return true
}

func startLocation(ed *Editor) {
	ed.location = location{}
	if ed.location.update(ed) {
		ed.mode = &ed.location
	}
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
	ed.location.backspace(ed)
}

func acceptLocation(ed *Editor) {
	// XXX Maybe we want to use eval.cdInner and increase the score?
	loc := &ed.location
	if len(loc.candidates) > 0 {
		dir := loc.candidates[loc.selected].Path
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
	}
	ed.mode = &ed.insert
}

func locationDefault(ed *Editor) {
	k := ed.lastKey
	if ed.location.handleFilterKey(k) {
		ed.location.update(ed)
	} else {
		startInsert(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}

func (loc *location) prev(cycle bool) {
	loc.listing.prev(cycle, len(loc.candidates))
}

func (loc *location) next(cycle bool) {
	loc.listing.next(cycle, len(loc.candidates))
}

func (loc *location) backspace(ed *Editor) {
	if loc.listing.backspace() {
		loc.update(ed)
	}
}

func (loc *location) List(width, maxHeight int) *buffer {
	get := func(i int) string {
		cand := loc.candidates[i]
		return fmt.Sprintf("%4.0f %s", cand.Score, parse.Quote(cand.Path))
	}
	return loc.listing.list(get, len(loc.candidates), width, maxHeight)
}
