package edit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// Command history listing mode.

// errStoreOffline is thrown when an operation requires the storage backend, but
// it is offline.
var errStoreOffline = errors.New("store offline")

type histlist struct {
	all             []string
	dedup           bool
	caseInsensitive bool
	last            map[string]int
	shown           []string
	index           []int
	indexWidth      int
}

func init() { atEditorInit(initHistlist) }

func initHistlist(ed *editor, ns eval.Ns) {
	subns := eval.Ns{
		"binding": eval.NewVariableFromPtr(&ed.histlistBinding),
	}
	subns.AddBuiltinFns("edit:histlist:", map[string]interface{}{
		"start":                   func() { histlistStart(ed) },
		"toggle-dedup":            func() { histlistToggleDedup(ed) },
		"toggle-case-sensitivity": func() { histlistToggleCaseSensitivity(ed) },
	})
	ns.AddNs("histlist", subns)
}

func newHistlist(cmds []string) *histlist {
	last := make(map[string]int)
	for i, entry := range cmds {
		last[entry] = i
	}
	return &histlist{
		// This has to be here for the initialization to work :(
		all:        cmds,
		dedup:      true,
		last:       last,
		indexWidth: len(strconv.Itoa(len(cmds) - 1)),
	}
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

func (hl *histlist) Accept(i int, ed eddefs.Editor) {
	line := hl.shown[i]
	buffer, _ := ed.Buffer()
	if len(buffer) > 0 {
		line = "\n" + line
	}
	ed.InsertAtDot(line)
}

func histlistStart(ed *editor) {
	cmds, err := getCmds(ed)
	if err != nil {
		ed.Notify("%v", err)
		return
	}

	ed.SetModeListing(ed.histlistBinding, newHistlist(cmds))
}

func getCmds(ed *editor) ([]string, error) {
	if ed.daemon == nil {
		return nil, errStoreOffline
	}
	return ed.hist.fuser.AllCmds()
}

func histlistToggleDedup(ed *editor) {
	if l, hl, ok := getHistlist(ed); ok {
		hl.dedup = !hl.dedup
		l.refresh()
	}
}

func histlistToggleCaseSensitivity(ed *editor) {
	if l, hl, ok := getHistlist(ed); ok {
		hl.caseInsensitive = !hl.caseInsensitive
		l.refresh()
	}
}

func getHistlist(ed *editor) (*listingMode, *histlist, bool) {
	if l, ok := ed.mode.(*listingMode); ok {
		if hl, ok := l.provider.(*histlist); ok {
			return l, hl, true
		}
	}
	return nil, nil, false
}
