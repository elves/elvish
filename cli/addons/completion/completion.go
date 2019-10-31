// Package completion implements the UI for showing, filtering and inserting
// completion candidates.
package completion

import (
	"strings"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/combobox"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/styled"
)

// Item represents a completion item, also known as a candidate.
type Item struct {
	// Used in the UI and for filtering.
	ToShow string
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
func Start(app *cli.App, cfg Config) {
	w := combobox.New(combobox.Spec{
		CodeArea: codearea.Spec{
			Prompt: layout.ModePrompt("COMPLETING "+cfg.Name, true),
		},
		ListBox: listbox.Spec{
			Horizontal:     true,
			OverlayHandler: cfg.Binding,
			OnAccept: func(it listbox.Items, i int) {
				text := it.(items)[i].ToInsert
				app.CodeArea.MutateCodeAreaState(func(s *codearea.State) {
					// TODO(xiaq): This is not correct when cfg.Replace has non-zero
					// width.
					s.CodeBuffer.InsertAtDot(text)
				})
				app.MutateAppState(func(s *cli.State) { s.Listing = nil })
			},
		},
		OnFilter: func(w combobox.Widget, p string) {
			w.ListBox().Reset(filter(cfg.Items, p), 0)
		},
	})
	// TODO(xiaq): Maintain PendingCode of the code area widget.
	app.MutateAppState(func(s *cli.State) { s.Listing = w })
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

func (it items) Show(i int) styled.Text { return styled.Plain(it[i].ToShow) }

func (it items) Len() int { return len(it) }
