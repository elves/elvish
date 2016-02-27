package edit

import "github.com/elves/elvish/store"

// Command history listing subosytem.

// Interface.

type historyListing struct {
	all     []string
	current int
}

func (*historyListing) Mode() ModeType {
	return modeHistoryListing
}

func (hl *historyListing) ModeLine(width int) *buffer {
	return makeModeLine(" LISTING HISTORY ", width)
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
		hl.current = -1
		return err
	}
	hl.all = cmds
	hl.current = len(hl.all) - 1
	return nil
}

func (hist *historyListing) prev() {
	if len(hist.all) > 0 && hist.current > 0 {
		hist.current--
	}
}

func (hist *historyListing) next() {
	if len(hist.all) > 0 && hist.current < len(hist.all)-1 {
		hist.current++
	}
}

func (hist *historyListing) List(width, maxHeight int) *buffer {
	b := newBuffer(width)
	if len(hist.all) == 0 {
		b.writes("(no history)", "")
		return b
	}

	low, high := findWindow(len(hist.all), hist.current, maxHeight)
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		style := ""
		if i == hist.current {
			style = styleForSelected
		}
		b.writes(TrimWcWidth(hist.all[i], width), style)
	}
	return b
}
