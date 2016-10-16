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
	listing
	all   []string
	dedup bool
	last  map[string]int
	shown []string
	index []int
}

func (hl *histlist) updateShown() {
	hl.shown = nil
	hl.index = nil
	dedup := hl.dedup
	filter := hl.filter
	for i, entry := range hl.all {
		if (!dedup || hl.last[entry] == i) && strings.Contains(entry, filter) {
			hl.index = append(hl.index, i)
			hl.shown = append(hl.shown, entry)
		}
	}
	hl.selected = len(hl.shown) - 1
}

func (hl *histlist) toggleDedup() {
	hl.dedup = !hl.dedup
	hl.updateShown()
}

func (hl *histlist) Len() int {
	return len(hl.shown)
}

func (hl *histlist) Show(i, width int) styled {
	entry := hl.shown[i]
	return unstyled(util.TrimEachLineWcwidth(entry, width))
}

func (hl *histlist) Filter(filter string) int {
	hl.updateShown()
	return len(hl.shown) - 1
}

func (hl *histlist) Accept(i int, ed *Editor) {
	line := hl.shown[i]
	if len(ed.line) > 0 {
		line = "\n" + line
	}
	ed.insertAtDot(line)
}

func (hl *histlist) ModeTitle(i int) string {
	dedup := ""
	if hl.dedup {
		dedup = " (dedup)"
	}
	return fmt.Sprintf(" HISTORY #%d%s ", hl.index[i], dedup)
}

func startHistlist(ed *Editor) {
	hl, err := newHistlist(ed.store)
	if err != nil {
		ed.Notify("%v", err)
		return
	}

	ed.histlist = hl
	// ed.histlist = newListing(modeHistoryListing, hl)
	ed.mode = ed.histlist
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
	last := make(map[string]int)
	for i, entry := range all {
		last[entry] = i
	}
	hl := &histlist{all: all, last: last}
	hl.listing = newListing(modeHistoryListing, hl)
	return hl, nil
}

func histlistToggleDedup(ed *Editor) {
	if ed.histlist != nil {
		ed.histlist.toggleDedup()
	}
}
