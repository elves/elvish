// Package histlist implements the history listing addon.
package histlist

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/store"
	"github.com/elves/elvish/pkg/ui"
)

// Config contains configurations to start history listing.
type Config struct {
	// Binding provides key binding.
	Binding cli.Handler
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

	w := cli.NewComboBox(cli.ComboBoxSpec{
		CodeArea: cli.CodeAreaSpec{Prompt: func() ui.Text {
			content := " HISTORY "
			if cfg.Dedup() {
				content += "(dedup on) "
			}
			if !cfg.CaseSensitive() {
				content += "(case-insensitive) "
			}
			return cli.ModeLine(content, true)
		}},
		ListBox: cli.ListBoxSpec{
			OverlayHandler: cfg.Binding,
			OnAccept: func(it cli.Items, i int) {
				text := it.(items).entries[i].Text
				app.CodeArea().MutateState(func(s *cli.CodeAreaState) {
					buf := &s.Buffer
					if buf.Content == "" {
						buf.InsertAtDot(text)
					} else {
						buf.InsertAtDot("\n" + text)
					}
				})
				app.MutateState(func(s *cli.State) { s.Addon = nil })
			},
		},
		OnFilter: func(w cli.ComboBox, p string) {
			it := cmdItems.filter(p, cfg.Dedup(), cfg.CaseSensitive())
			w.ListBox().Reset(it, it.Len()-1)
		},
	})

	app.MutateState(func(s *cli.State) { s.Addon = w })
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
