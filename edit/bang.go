package edit

import (
	"fmt"
	"strconv"
	"strings"
)

// Bang mode.

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

func (b *bang) Show(i int) (string, styled) {
	entry := b.filtered[i]
	var head string
	if entry.i == -1 {
		head = "M-,"
	} else if b.minus {
		head = fmt.Sprintf("%d", entry.i-len(b.words))
	} else {
		head = fmt.Sprintf("%d", entry.i)
	}
	return head, unstyled(entry.s)
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
	startInsert(ed)
}

func startBang(ed *Editor) {
	_, line, err := ed.store.LastCmd(-1, "")
	if err != nil {
		ed.Notify("db error: %s", err.Error())
		return
	}
	ed.bang = newBang(line)
	ed.mode = ed.bang
}

func bangAltDefault(ed *Editor) {
	l := ed.bang
	if l.handleFilterKey(ed.lastKey) {
		if l.Len() == 1 {
			l.Accept(l.selected, ed)
		}
	} else if ed.lastKey == (Key{',', Alt}) {
		l.Accept(0, ed)
	} else {
		startInsert(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}
