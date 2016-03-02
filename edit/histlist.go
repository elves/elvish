package edit

import (
	"fmt"

	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

// Command history listing mode.

// Interface.

type histlist struct {
	listing
	all      []string
	filtered []string
}

func (*histlist) Mode() ModeType {
	return modeHistoryListing
}

func (hl *histlist) ModeLine(width int) *buffer {
	return hl.modeLine(fmt.Sprintf(" HISTORY #%d ", hl.selected), width)
}

func (hl *histlist) update() {
	hl.filtered = nil
	for _, item := range hl.all {
		if util.MatchSubseq(item, hl.filter) {
			hl.filtered = append(hl.filtered, item)
		}
	}
	hl.selected = len(hl.filtered) - 1
}

func (hl *histlist) backspace() {
	if hl.listing.backspace() {
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
	ed.historyListing.prev(false, len(ed.historyListing.all))
}

func histlistNext(ed *Editor) {
	ed.historyListing.next(false, len(ed.historyListing.all))
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
	if ed.historyListing.handleFilterKey(k) {
		ed.historyListing.update()
	} else {
		startInsert(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}

// Implementation.

func initHistoryListing(hl *histlist, s *store.Store) error {
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

func (hist *histlist) List(width, maxHeight int) *buffer {
	get := func(i int) string {
		return ForceWcWidth(hist.filtered[i], width)
	}
	return hist.listing.list(get, len(hist.filtered), width, maxHeight)
}
