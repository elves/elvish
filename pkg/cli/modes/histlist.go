package modes

import (
	"fmt"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/ui"
)

// Histlist is a mode for browsing history and selecting entries to insert. It
// is based on the ComboBox widget.
type Histlist interface {
	tk.ComboBox
}

// HistlistSpec specifies the configuration for the histlist mode.
type HistlistSpec struct {
	// Key bindings.
	Bindings tk.Bindings
	// AllCmds is called to retrieve all commands.
	AllCmds func() ([]storedefs.Cmd, error)
	// Dedup is called to determine whether deduplication should be done.
	// Defaults to true if unset.
	Dedup func() bool
	// Configuration for the filter.
	Filter FilterSpec
	// RPrompt of the code area (first row of the widget).
	CodeAreaRPrompt func() ui.Text
}

// NewHistlist creates a new histlist mode.
func NewHistlist(app cli.App, spec HistlistSpec) (Histlist, error) {
	codeArea, err := FocusedCodeArea(app)
	if err != nil {
		return nil, err
	}
	if spec.AllCmds == nil {
		return nil, errNoHistoryStore
	}
	if spec.Dedup == nil {
		spec.Dedup = func() bool { return true }
	}

	cmds, err := spec.AllCmds()
	if err != nil {
		return nil, fmt.Errorf("db error: %v", err.Error())
	}
	last := map[string]int{}
	for i, cmd := range cmds {
		last[cmd.Text] = i
	}
	cmdItems := histlistItems{cmds, last}

	w := tk.NewComboBox(tk.ComboBoxSpec{
		CodeArea: tk.CodeAreaSpec{
			Prompt: func() ui.Text {
				content := " HISTORY "
				if spec.Dedup() {
					content += "(dedup on) "
				}
				return modeLine(content, true)
			},
			RPrompt:     spec.CodeAreaRPrompt,
			Highlighter: spec.Filter.Highlighter,
		},
		ListBox: tk.ListBoxSpec{
			Bindings: spec.Bindings,
			OnAccept: func(it tk.Items, i int) {
				text := it.(histlistItems).entries[i].Text
				codeArea.MutateState(func(s *tk.CodeAreaState) {
					buf := &s.Buffer
					if buf.Content == "" {
						buf.InsertAtDot(text)
					} else {
						buf.InsertAtDot("\n" + text)
					}
				})
				app.PopAddon()
			},
		},
		OnFilter: func(w tk.ComboBox, p string) {
			it := cmdItems.filter(spec.Filter.makePredicate(p), spec.Dedup())
			w.ListBox().Reset(it, it.Len()-1)
		},
	})
	return w, nil
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

func (it histlistItems) Show(i int) ui.Text {
	entry := it.entries[i]
	// TODO: The alignment of the index works up to 10000 entries.
	return ui.T(fmt.Sprintf("%4d %s", entry.Seq, entry.Text))
}

func (it histlistItems) Len() int { return len(it.entries) }
