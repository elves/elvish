package edit

import "github.com/elves/elvish/store"

// Command history listing subosytem.

// Interface.

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
	ed.mode = modeHistoryListing
}

func defaultHistoryListing(ed *Editor) {
	ed.mode = modeInsert
	ed.nextAction = action{actionType: reprocessKey}
}

// Implementation.

type historyListing struct {
	all []string
}

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
