package edit

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/edit/ui"
)

// Bang mode.

var _ = registerListingBuiltins("bang", map[string]func(*Editor){
	"start":       bangStart,
	"alt-default": bangAltDefault,
}, func(ed *Editor) *listing { return &ed.bang.listing })

func init() {
	registerListingBindings(modeBang, "bang", map[ui.Key]string{
		ui.Default: "alt-default",
	})
}

type bangEntry struct {
	i int
	s string
}

type bang struct {
	listing
	line     string
	words    []string
	filtered []bangEntry
	minus    bool
}

func newBang(line string) *bang {
	b := &bang{listing{}, line, wordify(line), nil, false}
	b.listing = newListing(modeBang, b)
	return b
}

func (b *bang) ModeTitle(int) string {
	return " LASTCMD "
}

func (b *bang) Len() int {
	return len(b.filtered)
}

func (b *bang) Show(i int) (string, ui.Styled) {
	entry := b.filtered[i]
	var head string
	if entry.i == -1 {
		head = "M-,"
	} else if b.minus {
		head = fmt.Sprintf("%d", entry.i-len(b.words))
	} else {
		head = fmt.Sprintf("%d", entry.i)
	}
	return head, ui.Unstyled(entry.s)
}

func (b *bang) Filter(filter string) int {
	b.filtered = nil
	b.minus = len(filter) > 0 && filter[0] == '-'
	if filter == "" || filter == "-" {
		b.filtered = append(b.filtered, bangEntry{-1, b.line})
	} else if _, err := strconv.Atoi(filter); err != nil {
		return -1
	}
	// Quite inefficient way to filter by prefix of stringified index.
	n := len(b.words)
	for i, word := range b.words {
		if filter == "" ||
			(!b.minus && strings.HasPrefix(strconv.Itoa(i), filter)) ||
			(b.minus && strings.HasPrefix(strconv.Itoa(i-n), filter)) {
			b.filtered = append(b.filtered, bangEntry{i, word})
		}
	}
	if len(b.filtered) == 0 {
		return -1
	}
	return 0
}

// Editor interface.

func (b *bang) Accept(i int, ed *Editor) {
	ed.insertAtDot(b.filtered[i].s)
	insertStart(ed)
}

func bangStart(ed *Editor) {
	_, cmd, err := ed.daemon.PrevCmd(-1, "")
	if err != nil {
		ed.Notify("db error: %s", err.Error())
		return
	}
	ed.bang = newBang(cmd)
	ed.mode = ed.bang
}

func bangAltDefault(ed *Editor) {
	l := ed.bang
	if l.handleFilterKey(ed.lastKey) {
		if l.Len() == 1 {
			l.Accept(l.selected, ed)
		}
	} else if ed.lastKey == (ui.Key{',', ui.Alt}) {
		l.Accept(0, ed)
	} else {
		insertStart(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}
