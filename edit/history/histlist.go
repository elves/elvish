package history

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
)

// Command history listing mode.

// errStoreOffline is thrown when an operation requires the storage backend, but
// it is offline.
var errStoreOffline = errors.New("store offline")

type histlist struct {
	all             []histutil.Entry
	dedup           bool
	caseInsensitive bool
	last            map[string]int
	shown           []histutil.Entry
	indexWidth      int
}

func newHistlist(cmds []histutil.Entry) *histlist {
	last := make(map[string]int)
	for i, entry := range cmds {
		last[entry.Text] = i
	}
	return &histlist{
		// This has to be here for the initialization to work :(
		all:        cmds,
		dedup:      true,
		last:       last,
		indexWidth: len(strconv.Itoa(len(cmds) - 1)),
	}
}

func (hl *histlist) Teardown() {
	*hl = histlist{}
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
	return fmt.Sprintf("%d", hl.shown[i].Seq), ui.Unstyled(hl.shown[i].Text)
}

func (hl *histlist) Filter(filter string) int {
	hl.shown = nil
	dedup := hl.dedup
	if hl.caseInsensitive {
		filter = strings.ToLower(filter)
	}
	for i, entry := range hl.all {
		fentry := entry.Text
		if hl.caseInsensitive {
			fentry = strings.ToLower(fentry)
		}
		if (!dedup || hl.last[entry.Text] == i) && strings.Contains(fentry, filter) {
			hl.shown = append(hl.shown, entry)
		}
	}
	// TODO: Maintain old selection
	return len(hl.shown) - 1
}

// Editor interface.

func (hl *histlist) Accept(i int, ed eddefs.Editor) {
	line := hl.shown[i].Text
	buffer, _ := ed.Buffer()
	if len(buffer) > 0 {
		line = "\n" + line
	}
	ed.InsertAtDot(line)
}

func (hl *histlist) start(ed eddefs.Editor, fuser *histutil.Fuser, binding eddefs.BindingMap) {
	cmds, err := getCmds(fuser)
	if err != nil {
		ed.Notify("%v", err)
		return
	}

	*hl = *newHistlist(cmds)

	ed.SetModeListing(binding, hl)
}

func getCmds(fuser *histutil.Fuser) ([]histutil.Entry, error) {
	if fuser == nil {
		return nil, errStoreOffline
	}
	return fuser.AllCmds()
}

func (hl *histlist) toggleDedup(ed eddefs.Editor) {
	hl.dedup = !hl.dedup
	ed.RefreshListing()
}

func (hl *histlist) toggleCaseSensitivity(ed eddefs.Editor) {
	hl.caseInsensitive = !hl.caseInsensitive
	ed.RefreshListing()
}
