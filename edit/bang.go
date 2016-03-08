package edit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Bang mode.

type bang struct {
	line     string
	words    []string
	filtered []string
}

func (b *bang) Len() int {
	return len(b.filtered)
}

func (b *bang) Show(i, width int) string {
	head := fmt.Sprintf("%-3d ", i)
	if i == 0 {
		head = "A-1 "
	}
	return ForceWcWidth(head+b.filtered[i], width)
}

func (b *bang) Filter(filter string) int {
	b.filtered = nil
	if filter != "" {
		if _, err := strconv.Atoi(filter); err != nil {
			return -1
		}
	}
	// Quite inefficient way to filter by prefix of stringified index.
	if strings.HasPrefix("0", filter) {
		b.filtered = append(b.filtered, b.line)
	}
	for i, word := range b.words {
		if strings.HasPrefix(strconv.Itoa(i+1), filter) {
			b.filtered = append(b.filtered, word)
		}
	}
	if len(b.filtered) == 0 {
		return -1
	}
	return 0
}

func (b *bang) Accept(i int, ed *Editor) {
	ed.insertAtDot(b.filtered[i])
	startInsert(ed)
}

func (b *bang) ModeTitle(i int) string {
	return " BANG "
}

var wordSep = regexp.MustCompile("[ \t]+")

func startBang(ed *Editor) {
	_, line, err := ed.store.LastCmd(-1, "", true)
	if err == nil {
		ed.bang = listing{modeBang, newBang(line), 0, ""}
		ed.bang.changeFilter("")
		ed.mode = &ed.bang
	} else {
		ed.addTip("db error: %s", err.Error())
	}
}

func bangAltDefault(ed *Editor) {
	l := ed.bang
	if l.handleFilterKey(ed.lastKey) {
		if l.provider.Len() == 1 {
			l.provider.Accept(l.selected, ed)
		}
	} else if ed.lastKey == (Key{'1', Alt}) {
		l.provider.Accept(0, ed)
	} else {
		startInsert(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}

func newBang(line string) *bang {
	return &bang{line, wordSep.Split(strings.Trim(line, " \t"), -1), nil}
}
