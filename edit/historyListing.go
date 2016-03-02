package edit

import (
	"fmt"

	"github.com/elves/elvish/store"
)

// Command history listing subosytem.

// Interface.

type historyListing struct {
	all      []string
	selected int
}

func (*historyListing) Mode() ModeType {
	return modeHistoryListing
}

func (hl *historyListing) ModeLine(width int) *buffer {
	return makeModeLine(fmt.Sprintf(" HISTORY #%d ", hl.selected), width)
}

func startHistoryListing(ed *Editor) {
	if ed.store == nil {
		ed.notify("store not connected")
		return
	}
	err := initHistoryListing(&ed.historyListing, ed.store)
	if err != nil {
		ed.notify("%s", err)
		return
	}
	ed.mode = &ed.historyListing
}

func histlistPrev(ed *Editor) {
	ed.historyListing.prev()
}

func histlistNext(ed *Editor) {
	ed.historyListing.next()
}

func histlistAppend(ed *Editor) {
	if ed.historyListing.selected != -1 {
		line := ed.historyListing.all[ed.historyListing.selected]
		if len(ed.line) > 0 {
			line = "\n" + line
		}
		ed.insertAtDot(line)
	}
}

func defaultHistoryListing(ed *Editor) {
	ed.mode = &ed.insert
	ed.nextAction = action{typ: reprocessKey}
}

// Implementation.

func initHistoryListing(hl *historyListing, s *store.Store) error {
	seq, err := s.NextCmdSeq()
	if err != nil {
		return err
	}
	cmds, err := s.Cmds(0, seq)
	if err != nil {
		hl.all = nil
		hl.selected = -1
		return err
	}
	hl.all = cmds
	hl.selected = len(hl.all) - 1
	return nil
}

func (hist *historyListing) prev() {
	if len(hist.all) > 0 && hist.selected > 0 {
		hist.selected--
	}
}

func (hist *historyListing) next() {
	if len(hist.all) > 0 && hist.selected < len(hist.all)-1 {
		hist.selected++
	}
}

func (hist *historyListing) List(width, maxHeight int) *buffer {
	b := newBuffer(width)
	if len(hist.all) == 0 {
		b.writes("(no history)", "")
		return b
	}

	low, high := findWindow(len(hist.all), hist.selected, maxHeight)
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		style := ""
		if i == hist.selected {
			style = styleForSelected
		}
		b.writes(ForceWcWidth(hist.all[i], width), style)
	}
	return b
}
