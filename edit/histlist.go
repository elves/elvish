package edit

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

// Command history listing mode.

var ErrStoreOffline = errors.New("store offline")

type histlist struct {
	listing
	all             []string
	dedup           bool
	caseInsensitive bool
	last            map[string]int
	shown           []string
	index           []int
	indexWidth      int
}

func (hl *histlist) updateShown() {
	hl.shown = nil
	hl.index = nil
	dedup := hl.dedup
	filter := hl.filter
	if hl.caseInsensitive {
		filter = strings.ToLower(filter)
	}
	for i, entry := range hl.all {
		fentry := entry
		if hl.caseInsensitive {
			fentry = strings.ToLower(entry)
		}
		if (!dedup || hl.last[entry] == i) && strings.Contains(fentry, filter) {
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

func (hl *histlist) toggleCaseSensitivity() {
	hl.caseInsensitive = !hl.caseInsensitive
	hl.updateShown()
}

func (hl *histlist) Len() int {
	return len(hl.shown)
}

func (hl *histlist) Show(i, width int) styled {
	entry := hl.shown[i]
	lines := strings.Split(entry, "\n")

	var b bytes.Buffer

	first := fmt.Sprintf("%*d %s", hl.indexWidth, hl.index[i], lines[0])
	b.WriteString(util.ForceWcwidth(first, width))

	indent := strings.Repeat(" ", hl.indexWidth+1)
	for _, line := range lines[1:] {
		b.WriteByte('\n')
		b.WriteString(util.ForceWcwidth(indent+line, width))
	}

	return unstyled(b.String())
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
	s := " HISTORY "
	if hl.dedup {
		s += "(dedup on) "
	}
	if hl.caseInsensitive {
		s += "(case-insensitive) "
	}
	return s
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
	hl := &histlist{all: all, last: last, indexWidth: len(strconv.Itoa(len(all) - 1))}
	hl.listing = newListing(modeHistoryListing, hl)
	return hl, nil
}

func histlistToggleDedup(ed *Editor) {
	if ed.histlist != nil {
		ed.histlist.toggleDedup()
	}
}

func histlistToggleCaseSensitivity(ed *Editor) {
	if ed.histlist != nil {
		ed.histlist.toggleCaseSensitivity()
	}
}
