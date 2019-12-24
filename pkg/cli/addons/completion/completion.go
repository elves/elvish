// Package completion implements the UI for showing, filtering and inserting
// completion candidates.
package completion

import (
	"strings"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/el/codearea"
	"github.com/elves/elvish/pkg/cli/el/combobox"
	"github.com/elves/elvish/pkg/cli/el/layout"
	"github.com/elves/elvish/pkg/cli/el/listbox"
	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/ui"
)

// Item represents a completion item, also known as a candidate.
type Item struct {
	// Used in the UI and for filtering.
	ToShow string
	// Style to use in the UI.
	ShowStyle ui.Style
	// Used when inserting a candidate.
	ToInsert string
}

// Config keeps the configuration for the completion UI.
type Config struct {
	Binding el.Handler
	Name    string
	Replace diag.Ranging
	Items   []Item
}

// Start starts the completion UI.
func Start(app cli.App, cfg Config) {
	if len(cfg.Items) == 0 {
		app.Notify("no candidates")
		return
	}
	w := combobox.NewComboBox(combobox.ComboBoxSpec{
		CodeArea: codearea.CodeAreaSpec{
			Prompt: layout.ModePrompt(" COMPLETING "+cfg.Name+" ", true),
		},
		ListBox: listbox.ListBoxSpec{
			Horizontal:     true,
			OverlayHandler: cfg.Binding,
			OnSelect: func(it listbox.Items, i int) {
				text := it.(items)[i].ToInsert
				app.CodeArea().MutateState(func(s *codearea.CodeAreaState) {
					s.Pending = codearea.PendingCode{
						From: cfg.Replace.From, To: cfg.Replace.To, Content: text}
				})
			},
			OnAccept: func(it listbox.Items, i int) {
				app.CodeArea().MutateState(func(s *codearea.CodeAreaState) {
					s.ApplyPending()
				})
				app.MutateState(func(s *cli.State) { s.Addon = nil })
			},
			ExtendStyle: true,
		},
		OnFilter: func(w combobox.ComboBox, p string) {
			w.ListBox().Reset(filter(cfg.Items, p), 0)
		},
	})
	app.MutateState(func(s *cli.State) { s.Addon = w })
	app.Redraw()
}

// Close closes the completion UI.
func Close(app cli.App) {
	app.CodeArea().MutateState(
		func(s *codearea.CodeAreaState) { s.Pending = codearea.PendingCode{} })
	app.MutateState(func(s *cli.State) { s.Addon = nil })
	app.Redraw()
}

type items []Item

func filter(all []Item, p string) items {
	var filtered []Item
	for _, candidate := range all {
		if strings.Contains(candidate.ToShow, p) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func (it items) Show(i int) ui.Text {
	return ui.Text{&ui.Segment{Style: it[i].ShowStyle, Text: it[i].ToShow}}
}

func (it items) Len() int { return len(it) }
