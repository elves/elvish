package edit

import (
	"os"
	"unicode/utf8"

	"github.com/elves/elvish/store"
)

type location struct {
	filter     string
	candidates []store.Dir
	current    int
}

func (*location) Mode() ModeType {
	return modeLocation
}

func (l *location) ModeLine(width int) *buffer {
	// TODO keep it one line.
	b := newBuffer(width)
	b.writes(TrimWcWidth(" LOCATION ", width), styleForMode)
	b.writes(" ", "")
	b.writes(l.filter, styleForLocation)
	return b
}

func (l *location) updateCandidates(ed *Editor) bool {
	dirs, err := ed.store.FindDirsSubseq(l.filter)
	if err != nil {
		l.candidates = nil
		l.current = -1
		ed.notify("find directories: %v", err)
		return false
	}
	l.candidates = dirs

	if len(l.candidates) > 0 {
		l.current = 0
	} else {
		l.current = -1
	}
	return true
}

func startLocation(ed *Editor) {
	ed.location = location{}
	if ed.location.updateCandidates(ed) {
		ed.mode = &ed.location
	}
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

func locationBackspace(ed *Editor) {
	loc := &ed.location
	_, size := utf8.DecodeLastRuneInString(loc.filter)
	if size > 0 {
		loc.filter = loc.filter[:len(loc.filter)-size]
		loc.updateCandidates(ed)
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
	ed.mode = &ed.insert
}

func cancelLocation(ed *Editor) {
	ed.mode = &ed.insert
}

func locationDefault(ed *Editor) {
	k := ed.lastKey
	if likeChar(k) {
		ed.location.filter += string(k.Rune)
		ed.location.updateCandidates(ed)
	} else {
		cancelLocation(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}
