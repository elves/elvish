// Package histlist implements the history listing addon.
package histlist

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/ui"
)

// Config contains configurations to start history listing.
type Config struct {
	// Key bindings.
	Bindings tk.Bindings
	// Store provides the source of all commands.
	Store Store
	// Dedup is called to determine whether deduplication should be done.
	// Defaults to true if unset.
	Dedup func() bool
	// CaseSensitive is called to determine whether the filter should be
	// case-sensitive. Defaults to true if unset.
	CaseSensitive func() bool
}

// Store wraps the AllCmds method. It is a subset of histutil.Store.
type Store interface {
	AllCmds() ([]store.Cmd, error)
}

var _ = Store(histutil.Store(nil))

// Start starts history listing.
func Start(app cli.App, cfg Config) {
	if cfg.Store == nil {
		app.Notify("no history store")
		return
	}
	if cfg.Dedup == nil {
		cfg.Dedup = func() bool { return true }
	}
	if cfg.CaseSensitive == nil {
		cfg.CaseSensitive = func() bool { return true }
	}

	cmds, err := cfg.Store.AllCmds()
	if err != nil {
		app.Notify("db error: " + err.Error())
	}
	last := map[string]int{}
	for i, cmd := range cmds {
		last[cmd.Text] = i
	}
	cmdItems := items{cmds, last}

	w := tk.NewComboBox(tk.ComboBoxSpec{
		CodeArea: tk.CodeAreaSpec{Prompt: func() ui.Text {
			content := " HISTORY "
			if cfg.Dedup() {
				content += "(dedup on) "
			}
			if !cfg.CaseSensitive() {
				content += "(case-insensitive) "
			}
			return mode.Line(content, true)
		}},
		ListBox: tk.ListBoxSpec{
			Bindings: cfg.Bindings,
			OnAccept: func(it tk.Items, i int) {
				text := it.(items).entries[i].Text
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
			it := cmdItems.filter(p, cfg.Dedup(), cfg.CaseSensitive())
			w.ListBox().Reset(it, it.Len()-1)
		},
	})

	app.SetAddon(w, false)
	app.Redraw()
}

type items struct {
	entries []store.Cmd
	last    map[string]int
}

func (it items) filter(p string, dedup, caseSensitive bool) items {
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
	return items{filtered, nil}
}

func (it items) Show(i int) ui.Text {
	entry := it.entries[i]
	// TODO: The alignment of the index works up to 10000 entries.
	return ui.T(fmt.Sprintf("%4d %s", entry.Seq, entry.Text))
}

func (it items) Len() int { return len(it.entries) }
