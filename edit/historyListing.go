package edit

import (
	"fmt"
	"unicode/utf8"

	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

// Command history listing subosytem.

// Interface.

type historyListing struct {
	filter   string
	all      []string
	filtered []string
	selected int
}

func (*historyListing) Mode() ModeType {
	return modeHistoryListing
}

func (hl *historyListing) ModeLine(width int) *buffer {
	// TODO keep it one line.
	b := newBuffer(width)
	b.writes(TrimWcWidth(fmt.Sprintf(" HISTORY #%d ", hl.selected), width), styleForMode)
	b.writes(" ", "")
	b.writes(hl.filter, styleForFilter)
	b.dot = b.cursor()
	return b
}

func (hl *historyListing) update() {
	hl.filtered = nil
	for _, item := range hl.all {
		if util.MatchSubseq(item, hl.filter) {
			hl.filtered = append(hl.filtered, item)
		}
	}
	hl.selected = len(hl.filtered) - 1
}

func (hl *historyListing) backspace() {
	_, size := utf8.DecodeLastRuneInString(hl.filter)
	if size > 0 {
		hl.filter = hl.filter[:len(hl.filter)-size]
		hl.update()
	}
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

func histlistBackspace(ed *Editor) {
	ed.historyListing.backspace()
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
	k := ed.lastKey
	if likeChar(k) {
		ed.historyListing.filter += string(k.Rune)
		ed.historyListing.update()
	} else {
		startInsert(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
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
	hl.filter = ""
	hl.filtered = cmds
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
	if len(hist.filtered) == 0 {
		b.writes("(no history)", "")
		return b
	}

	low, high := findWindow(len(hist.filtered), hist.selected, maxHeight)
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		style := ""
		if i == hist.selected {
			style = styleForSelected
		}
		b.writes(ForceWcWidth(hist.filtered[i], width), style)
	}
	return b
}
