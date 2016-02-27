package edit

import (
	"os"
	"unicode"

	"github.com/elves/elvish/store"
)

type location struct {
	filter     string
	candidates []store.Dir
	current    int
}

func (l *location) updateCandidates(store *store.Store) error {
	dirs, err := store.FindDirs(l.filter)
	if err != nil {
		return err
	}
	l.candidates = dirs

	if len(l.candidates) > 0 {
		l.current = 0
	} else {
		l.current = -1
	}
	return nil
}

func startLocation(ed *Editor) {
	ed.location = location{}
	ed.location.updateCandidates(ed.store)
	ed.mode = modeLocation
}

func locationPrev(ed *Editor) {
	if len(ed.location.candidates) > 0 && ed.location.current > 0 {
		ed.location.current--
	}
}

func locationNext(ed *Editor) {
	if len(ed.location.candidates) > 0 && ed.location.current < len(ed.location.candidates)-1 {
		ed.location.current++
	}
}

func acceptLocation(ed *Editor) {
	// XXX Maybe we want to use eval.cdInner and increase the score?
	loc := &ed.location
	if len(loc.candidates) > 0 {
		err := os.Chdir(loc.candidates[loc.current].Path)
		if err != nil {
			ed.notify("%v", err)
		}
	}
	ed.mode = modeInsert
}

func cancelLocation(ed *Editor) {
	ed.mode = modeInsert
}

func locationDefault(ed *Editor) {
	k := ed.lastKey
	if k.Mod == 0 && k.Rune > 0 && unicode.IsGraphic(k.Rune) {
		ed.location.filter += string(k.Rune)
		ed.location.updateCandidates(ed.store)
	} else {
		cancelLocation(ed)
		ed.nextAction = action{actionType: reprocessKey}
	}
}
