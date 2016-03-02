package edit

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

// Command history listing mode.

var ErrStoreOffline = errors.New("store offline")

type histlist struct {
	all      []string
	filtered []string
}

func (hl *histlist) Len() int {
	return len(hl.filtered)
}

func (hl *histlist) Show(i, width int) string {
	return ForceWcWidth(hl.filtered[i], width)
}

func (hl *histlist) Filter(filter string) int {
	hl.filtered = nil
	for _, item := range hl.all {
		if util.MatchSubseq(item, filter) {
			hl.filtered = append(hl.filtered, item)
		}
	}
	// Select the last entry.
	return len(hl.filtered) - 1
}

func (hl *histlist) Accept(i int, ed *Editor) {
	line := hl.all[i]
	if len(ed.line) > 0 {
		line = "\n" + line
	}
	ed.insertAtDot(line)
}

func (hl *histlist) ModeTitle(i int) string {
	return fmt.Sprintf(" HISTORY #%d ", i)
}

func newHistlist(s *store.Store) (*histlist, error) {
	if s == nil {
		return nil, ErrStoreOffline
	}
	seq, err := s.NextCmdSeq()
	if err != nil {
		return nil, err
	}
	all, err := s.Cmds(0, seq)
	if err != nil {
		return nil, err
	}
	return &histlist{all, nil}, nil
}

// Editor builtins.

func startHistoryListing(ed *Editor) {
	hl, err := newHistlist(ed.store)
	if err != nil {
		ed.notify("%v", err)
		return
	}

	ed.histlist = listing{modeHistoryListing, hl, 0, ""}
	ed.histlist.changeFilter("")
	ed.mode = &ed.histlist
}

func histlistPrev(ed *Editor) {
	ed.histlist.prev(false)
}

func histlistNext(ed *Editor) {
	ed.histlist.next(false)
}

func histlistBackspace(ed *Editor) {
	ed.histlist.backspace()
}

func histlistAppend(ed *Editor) {
	ed.histlist.accept(ed)
}

func histlistDefault(ed *Editor) {
	ed.histlist.defaultBinding(ed)
}
