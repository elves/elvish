package edit

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
)

type location struct {
	filter     string
	candidates []store.Dir
	selected   int
}

func (*location) Mode() ModeType {
	return modeLocation
}

func (l *location) ModeLine(width int) *buffer {
	// TODO keep it one line.
	b := newBuffer(width)
	b.writes(TrimWcWidth(" LOCATION ", width), styleForMode)
	b.writes(" ", "")
	b.writes(l.filter, styleForFilter)
	b.dot = b.cursor()
	return b
}

func (l *location) updateCandidates(ed *Editor) bool {
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
	if ed.location.updateCandidates(ed) {
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
	if likeChar(k) {
		ed.location.filter += string(k.Rune)
		ed.location.updateCandidates(ed)
	} else {
		startInsert(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}

func (loc *location) prev(cycle bool) {
	loc.selected--
	if loc.selected == -1 {
		if cycle {
			loc.selected = len(loc.candidates) - 1
		} else {
			loc.selected++
		}
	}
}

func (loc *location) next(cycle bool) {
	loc.selected++
	if loc.selected == len(loc.candidates) {
		if cycle {
			loc.selected = 0
		} else {
			loc.selected--
		}
	}
}

func (loc *location) backspace(ed *Editor) {
	_, size := utf8.DecodeLastRuneInString(loc.filter)
	if size > 0 {
		loc.filter = loc.filter[:len(loc.filter)-size]
		loc.updateCandidates(ed)
	}
}

func (loc *location) List(width, maxHeight int) *buffer {
	b := newBuffer(width)
	if len(loc.candidates) == 0 {
		b.writes("(no match)", "")
		return b
	}
	low, high := findWindow(len(loc.candidates), loc.selected, maxHeight)
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		text := fmt.Sprintf("%4.0f %s", loc.candidates[i].Score, parse.Quote(loc.candidates[i].Path))
		style := ""
		if i == loc.selected {
			style = styleForSelected
		}
		b.writes(TrimWcWidth(text, width), style)
	}
	return b
}
