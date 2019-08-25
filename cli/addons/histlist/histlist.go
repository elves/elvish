package histlist

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/codearea"
	"github.com/elves/elvish/cli/combobox"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/layout"
	"github.com/elves/elvish/cli/listbox"
	"github.com/elves/elvish/styled"
)

type Config struct {
	Binding clitypes.Handler
	Store   histutil.Store
}

func Start(app *clicore.App, cfg Config) {
	if cfg.Store == nil {
		app.Notify("no history store")
		return
	}
	cmds, err := cfg.Store.AllCmds()
	if err != nil {
		app.Notify("db error: " + err.Error())
	}

	w := combobox.Widget{}
	w.CodeArea.Prompt = layout.ModePrompt("HISTLIST", true)
	w.ListBox.OverlayHandler = cfg.Binding
	w.OnFilter = func(p string) {
		w.ListBox.MutateListboxState(func(s *listbox.State) {
			*s = listbox.MakeState(filter(cmds, p), true)
		})
	}
	w.ListBox.OnAccept = func(i int) {
		items := w.ListBox.CopyListboxState().Items.(items)
		text := items[i].Text
		app.CodeArea.MutateCodeAreaState(func(s *codearea.State) {
			buf := &s.CodeBuffer
			if buf.Content == "" {
				buf.InsertAtDot(text)
			} else {
				buf.InsertAtDot("\n" + text)
			}
		})
	}
	app.MutateAppState(func(s *clicore.State) { s.Listing = &w })
}

type items []histutil.Entry

func filter(allEntries []histutil.Entry, p string) items {
	var entries []histutil.Entry
	for _, entry := range allEntries {
		if strings.Contains(entry.Text, p) {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (it items) Show(i int) styled.Text {
	// TODO: The alignment of the index works up to 10000 entries.
	return styled.Plain(fmt.Sprintf("%4d %s", it[i].Seq, it[i].Text))
}

func (it items) Len() int { return len(it) }
