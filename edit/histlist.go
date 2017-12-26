package edit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/edit/ui"
)

// Command history listing mode.

var _ = registerBuiltins(modeHistoryListing, map[string]func(*Editor){
	"start":                   histlistStart,
	"toggle-dedup":            histlistToggleDedup,
	"toggle-case-sensitivity": histlistToggleCaseSensitivity,
})

func init() {
	registerBindings(modeHistoryListing, modeHistoryListing,
		map[ui.Key]string{
			{'G', ui.Ctrl}: "toggle-case-sensitivity",
			{'D', ui.Ctrl}: "toggle-dedup",
		})
}

// ErrStoreOffline is thrown when an operation requires the storage backend, but
// it is offline.
var ErrStoreOffline = errors.New("store offline")

type histlist struct {
	all             []string
	dedup           bool
	caseInsensitive bool
	last            map[string]int
	shown           []string
	index           []int
	indexWidth      int
}

func newHistlist(cmds []string) *listing {
	last := make(map[string]int)
	for i, entry := range cmds {
		last[entry] = i
	}
	hl := &histlist{
		// This has to be here for the initialization to work :(
		all:        cmds,
		dedup:      true,
		last:       last,
		indexWidth: len(strconv.Itoa(len(cmds) - 1)),
	}
	l := newListing(modeHistoryListing, hl)
	return &l
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
	hl.shown = nil
	hl.index = nil
	dedup := hl.dedup
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
	// TODO: Maintain old selection
	return len(hl.shown) - 1
}

// Editor interface.

func (hl *histlist) Accept(i int, ed *Editor) {
	line := hl.shown[i]
	if len(ed.buffer) > 0 {
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

	ed.mode = newHistlist(cmds)
}

func getCmds(ed *Editor) ([]string, error) {
	if ed.daemon == nil {
		return nil, ErrStoreOffline
	}
	return ed.historyFuser.AllCmds()
}

func histlistToggleDedup(ed *Editor) {
	if l, hl, ok := getHistlist(ed); ok {
		hl.dedup = !hl.dedup
		l.refresh()
	}
}

func histlistToggleCaseSensitivity(ed *Editor) {
	if l, hl, ok := getHistlist(ed); ok {
		hl.caseInsensitive = !hl.caseInsensitive
		l.refresh()
	}
}

func getHistlist(ed *Editor) (*listing, *histlist, bool) {
	if l, ok := ed.mode.(*listing); ok {
		if hl, ok := l.provider.(*histlist); ok {
			return l, hl, true
		}
	}
	return nil, nil, false
}
