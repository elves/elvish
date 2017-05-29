package edit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/edit/ui"
)

// Command history listing mode.

var _ = registerListingBuiltins("histlist", map[string]func(*Editor){
	"start":                   histlistStart,
	"toggle-dedup":            histlistToggleDedup,
	"toggle-case-sensitivity": histlistToggleCaseSensitivity,
}, func(ed *Editor) *listing { return &ed.histlist.listing })

func init() {
	registerListingBindings(modeHistoryListing, "histlist",
		map[ui.Key]string{
			{'G', ui.Ctrl}: "toggle-case-sensitivity",
			{'D', ui.Ctrl}: "toggle-dedup",
		})
}

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

func newHistlist(cmds []string) *histlist {
	last := make(map[string]int)
	for i, entry := range cmds {
		last[entry] = i
	}
	hl := &histlist{
		all: cmds, last: last, indexWidth: len(strconv.Itoa(len(cmds) - 1))}
	hl.listing = newListing(modeHistoryListing, hl)
	return hl
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

func (*histlist) CursorOnModeLine() bool {
	return true
}

func (hl *histlist) Len() int {
	return len(hl.shown)
}

func (hl *histlist) Show(i int) (string, ui.Styled) {
	return fmt.Sprintf("%d", hl.index[i]), ui.Unstyled(hl.shown[i])
}

func (hl *histlist) Filter(filter string) int {
	hl.updateShown()
	return len(hl.shown) - 1
}

func (hl *histlist) toggleDedup() {
	hl.dedup = !hl.dedup
	hl.updateShown()
}

func (hl *histlist) toggleCaseSensitivity() {
	hl.caseInsensitive = !hl.caseInsensitive
	hl.updateShown()
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

// Editor interface.

func (hl *histlist) Accept(i int, ed *Editor) {
	line := hl.shown[i]
	if len(ed.line) > 0 {
		line = "\n" + line
	}
	ed.insertAtDot(line)
}

func histlistStart(ed *Editor) {
	cmds, err := getCmds(ed)
	if err != nil {
		ed.Notify("%v", err)
		return
	}

	ed.histlist = newHistlist(cmds)
	ed.mode = ed.histlist
}

func getCmds(ed *Editor) ([]string, error) {
	if ed.store == nil {
		return nil, ErrStoreOffline
	}
	seq, err := ed.store.NextCmdSeq()
	if err != nil {
		return nil, err
	}
	return ed.store.GetCmds(0, seq)
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
