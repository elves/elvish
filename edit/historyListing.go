package edit

import "github.com/elves/elvish/store"

// Command history listing subosytem.

// Interface.

type historyListing struct {
	all []string
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
	cmds, err := s.Cmds(seq-100, seq)
	if err != nil {
		return err
	}
	hl.all = cmds
	return nil
}
