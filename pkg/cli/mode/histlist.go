package mode

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/store"
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
	AllCmds func() ([]store.Cmd, error)
	// Dedup is called to determine whether deduplication should be done.
	// Defaults to true if unset.
	Dedup func() bool
	// CaseSensitive is called to determine whether the filter should be
	// case-sensitive. Defaults to true if unset.
	CaseSensitive func() bool
}

// NewHistlist creates a new histlist mode.
func NewHistlist(app cli.App, spec HistlistSpec) (Histlist, error) {
	if spec.AllCmds == nil {
		return nil, errNoHistoryStore
	}
	if spec.Dedup == nil {
		spec.Dedup = func() bool { return true }
	}
	if spec.CaseSensitive == nil {
		spec.CaseSensitive = func() bool { return true }
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
		CodeArea: tk.CodeAreaSpec{Prompt: func() ui.Text {
			content := " HISTORY "
			if spec.Dedup() {
				content += "(dedup on) "
			}
			if !spec.CaseSensitive() {
				content += "(case-insensitive) "
			}
			return ModeLine(content, true)
		}},
		ListBox: tk.ListBoxSpec{
			Bindings: spec.Bindings,
			OnAccept: func(it tk.Items, i int) {
				text := it.(histlistItems).entries[i].Text
				app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
					buf := &s.Buffer
					if buf.Content == "" {
						buf.InsertAtDot(text)
					} else {
						buf.InsertAtDot("\n" + text)
					}
				})
				app.SetAddon(nil, false)
			},
		},
		OnFilter: func(w tk.ComboBox, p string) {
			it := cmdItems.filter(p, spec.Dedup(), spec.CaseSensitive())
			w.ListBox().Reset(it, it.Len()-1)
		},
	})
	return w, nil
}

type histlistItems struct {
	entries []store.Cmd
	last    map[string]int
}

func (it histlistItems) filter(p string, dedup, caseSensitive bool) histlistItems {
	if p == "" && !dedup {
		return it
	}
	if !caseSensitive {
		p = strings.ToLower(p)
	}
	var filtered []store.Cmd
	for i, entry := range it.entries {
		text := entry.Text
		if dedup && it.last[text] != i {
			continue
		}
		if !caseSensitive {
			text = strings.ToLower(text)
		}
		if strings.Contains(text, p) {
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
