package edit

import (
	"errors"
	"fmt"
	"strings"

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
	entry := hl.filtered[i]
	if i := strings.IndexRune(entry, '\n'); i != -1 {
		entry = entry[:i] + "…"
	}
	return ForceWcWidth(entry, width)
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

func startHistoryListing(ed *Editor) {
	hl, err := newHistlist(ed.store)
	if err != nil {
		ed.notify("%v", err)
		return
	}

	ed.histlist = newListing(modeHistoryListing, hl)
	ed.mode = &ed.histlist
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
