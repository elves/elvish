package edit

import (
	"fmt"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/ui"
)

func startHistlist(ed *Editor, c etk.Context) {
	cmds, err := ed.histStore.AllCmds()
	if err != nil {
		// TODO: Handle
		return
	}

	last := map[string]int{}
	for i, cmd := range cmds {
		last[cmd.Text] = i
	}
	cmdItems := histlistItems{cmds, last}

	pushAddon(c, withAfterReact(
		etk.WithInit(comps.ComboBox,
			"query/prompt", addonPromptText(" HISTORY "),
			"gen-list", func(f string) (comps.ListItems, int) {
				// TODO: Implement filtering
				return cmdItems, len(cmdItems.entries) - 1
			},
			"binding", etkBindingFromBindingMap(ed, &ed.histlistBinding),
		),
		func(c etk.Context, r etk.Reaction) etk.Reaction {
			// TODO: Implement insertion of selected item
			return r
		},
	), true)
}

type histlistItems struct {
	entries []storedefs.Cmd
	last    map[string]int
}

func (it histlistItems) filter(p func(string) bool, dedup bool) histlistItems {
	var filtered []storedefs.Cmd
	for i, entry := range it.entries {
		text := entry.Text
		if dedup && it.last[text] != i {
			continue
		}
		if p(text) {
			filtered = append(filtered, entry)
		}
	}
	return histlistItems{filtered, nil}
}

func (it histlistItems) Len() int      { return len(it.entries) }
func (it histlistItems) Get(i int) any { return it.entries[i] }

func (it histlistItems) Show(i int) ui.Text {
	entry := it.entries[i]
	// TODO: The alignment of the index works up to 10000 entries.
	return ui.T(fmt.Sprintf("%4d %s", entry.Seq, entry.Text))
}
