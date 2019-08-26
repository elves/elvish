package lastcmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/codearea"
	"github.com/elves/elvish/cli/combobox"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/layout"
	"github.com/elves/elvish/cli/listbox"
	"github.com/elves/elvish/styled"
)

type Config struct {
	Binding   clitypes.Handler
	Store     histutil.Store
	Wordifier func(string) []string
}

func Start(app *cli.App, cfg Config) {
	if cfg.Store == nil {
		app.Notify("no history store")
		return
	}
	cmd, err := cfg.Store.LastCmd()
	if err != nil {
		app.Notify("db error: " + err.Error())
	}
	wordifier := cfg.Wordifier
	if wordifier == nil {
		wordifier = strings.Fields
	}
	cmdText := cmd.Text
	words := wordifier(cmdText)
	entries := make([]entry, len(words)+1)
	entries[0] = entry{content: cmdText}
	for i, word := range words {
		entries[i+1] = entry{strconv.Itoa(i), strconv.Itoa(i - len(words)), word}
	}

	w := combobox.Widget{}
	w.CodeArea.Prompt = layout.ModePrompt("LASTCMD", true)
	w.ListBox.OverlayHandler = cfg.Binding
	w.OnFilter = func(p string) {
		w.ListBox.MutateListboxState(func(s *listbox.State) {
			*s = listbox.MakeState(filter(entries, p), false)
		})
	}
	w.ListBox.OnAccept = func(it listbox.Items, i int) {
		text := it.(items).entries[i].content
		app.CodeArea.MutateCodeAreaState(func(s *codearea.State) {
			s.CodeBuffer.InsertAtDot(text)
		})
	}
	app.MutateAppState(func(s *cli.State) { s.Listing = &w })
}

type items struct {
	negFilter bool
	entries   []entry
}

type entry struct {
	posIndex string
	negIndex string
	content  string
}

func filter(allEntries []entry, p string) items {
	var entries []entry
	negFilter := strings.HasPrefix(p, "-")
	for _, entry := range allEntries {
		if (negFilter && strings.HasPrefix(entry.negIndex, p)) ||
			(!negFilter && strings.HasPrefix(entry.posIndex, p)) {
			entries = append(entries, entry)
		}
	}
	return items{negFilter, entries}
}

func (it items) Show(i int) styled.Text {
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
	return styled.Plain(fmt.Sprintf("%3s %s", index, entry.content))
}

func (it items) Len() int { return len(it.entries) }
