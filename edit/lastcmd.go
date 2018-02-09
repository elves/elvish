package edit

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// LastCmd mode.

type lastcmd struct {
	line     string
	words    []string
	filtered []lastcmdEntry
	minus    bool
}

type lastcmdEntry struct {
	i int
	s string
}

func init() { atEditorInit(initLastcmd) }

func initLastcmd(ed *editor, ns eval.Ns) {
	subns := eval.Ns{
		"binding": eval.NewVariableFromPtr(&ed.lastcmdBinding),
	}
	subns.AddBuiltinFns("edit:lastcmd:", map[string]interface{}{
		"start":       func() { lastcmdStart(ed) },
		"alt-default": func() { lastcmdAltDefault(ed) },
	})
	ns.AddNs("lastcmd", subns)
}

func newLastCmd(line string) *lastcmd {
	return &lastcmd{line, wordify(line), nil, false}
}

func (b *lastcmd) ModeTitle(int) string {
	return " LASTCMD "
}

func (b *lastcmd) Len() int {
	return len(b.filtered)
}

func (b *lastcmd) Show(i int) (string, ui.Styled) {
	entry := b.filtered[i]
	var head string
	if entry.i == -1 {
		head = "M-1"
	} else if b.minus {
		head = fmt.Sprintf("%d", entry.i-len(b.words))
	} else {
		head = fmt.Sprintf("%d", entry.i)
	}
	return head, ui.Unstyled(entry.s)
}

func (b *lastcmd) Filter(filter string) int {
	b.filtered = nil
	b.minus = len(filter) > 0 && filter[0] == '-'
	if filter == "" || filter == "-" {
		b.filtered = append(b.filtered, lastcmdEntry{-1, b.line})
	} else if _, err := strconv.Atoi(filter); err != nil {
		return -1
	}
	// Quite inefficient way to filter by prefix of stringified index.
	n := len(b.words)
	for i, word := range b.words {
		if filter == "" ||
			(!b.minus && strings.HasPrefix(strconv.Itoa(i), filter)) ||
			(b.minus && strings.HasPrefix(strconv.Itoa(i-n), filter)) {
			b.filtered = append(b.filtered, lastcmdEntry{i, word})
		}
	}
	if len(b.filtered) == 0 {
		return -1
	}
	return 0
}

// Editor interface.

func (b *lastcmd) Accept(i int, ed eddefs.Editor) {
	ed.InsertAtDot(b.filtered[i].s)
	ed.SetModeInsert()
}

func lastcmdStart(ed *editor) {
	logger.Println("lastcmd-alt-start")
	_, cmd, err := ed.daemon.PrevCmd(-1, "")
	if err != nil {
		ed.Notify("db error: %s", err.Error())
		return
	}
	ed.SetModeListing(&ed.lastcmdBinding, newLastCmd(cmd))
}

func lastcmdAltDefault(ed *editor) {
	l, lc := getLastcmd(ed)
	if l == nil {
		return
	}
	logger.Println("lastcmd-alt-default")
	if ed.lastKey == (ui.Key{'1', ui.Alt}) {
		lc.Accept(0, ed)
		logger.Println("accepting")
	} else if l.handleFilterKey(ed.lastKey) {
		if lc.Len() == 1 {
			lc.Accept(l.selected, ed)
			logger.Println("accepting")
		}
	} else {
		ed.SetModeInsert()
		ed.SetAction(reprocessKey)
	}
}

func getLastcmd(ed *editor) (*listingMode, *lastcmd) {
	if l, ok := ed.mode.(*listingMode); ok {
		if lc, ok := l.provider.(*lastcmd); ok {
			return l, lc
		}
	}
	return nil, nil
}
