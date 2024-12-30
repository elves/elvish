package edit

import (
	"fmt"
	"strconv"
	"strings"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

func startLastcmd(ed *Editor, c etk.Context) {
	cursor := ed.histStore.Cursor("")
	cursor.Prev()
	cmd, err := cursor.Get()
	if err != nil {
		// TODO: Report
		return
	}

	items := makeLastCmdItems(cmd.Text)
	pushAddon(c, etk.WithInit(comps.ComboBox,
		"query/prompt", addonPromptText(" LASTCMD "),
		"gen-list", func(f string) (comps.ListItems, int) {
			return items, 0
		},
		"binding", etkBindingFromBindingMap(ed, &ed.lastcmdBinding),
	), 1)
}

type lastcmdItems struct {
	negFilter bool
	entries   []lastcmdItem
}

func makeLastCmdItems(cmd string) lastcmdItems {
	// TODO: Use Elvish's lexer as the basis for a better wordifier
	words := strings.Fields(cmd)
	entries := make([]lastcmdItem, len(words)+1)
	entries[0] = lastcmdItem{content: cmd}
	for i, word := range words {
		entries[i+1] = lastcmdItem{strconv.Itoa(i), strconv.Itoa(i - len(words)), word}
	}
	// TODO: Filter
	return filterLastcmdItems(entries, "")
}

type lastcmdItem struct {
	posIndex string
	negIndex string
	content  string
}

func filterLastcmdItems(allEntries []lastcmdItem, p string) lastcmdItems {
	if p == "" {
		return lastcmdItems{false, allEntries}
	}
	var entries []lastcmdItem
	negFilter := strings.HasPrefix(p, "-")
	for _, entry := range allEntries {
		if (negFilter && strings.HasPrefix(entry.negIndex, p)) ||
			(!negFilter && strings.HasPrefix(entry.posIndex, p)) {
			entries = append(entries, entry)
		}
	}
	return lastcmdItems{negFilter, entries}
}

func (it lastcmdItems) Len() int      { return len(it.entries) }
func (it lastcmdItems) Get(i int) any { return it.entries[i] }

func (it lastcmdItems) Show(i int) ui.Text {
	index := ""
	entry := it.entries[i]
	if it.negFilter {
		index = entry.negIndex
	} else {
		index = entry.posIndex
	}
	// NOTE: We now use a hardcoded width of 3 for the index, which will work as
	// long as the command has less than 1000 words (when filter is positive) or
	// 100 words (when filter is negative).
	return ui.T(fmt.Sprintf("%3s %s", index, entry.content))
}
