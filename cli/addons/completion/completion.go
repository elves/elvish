// Package completion implements the logic to show completion candidates.
package completion

import (
	"strings"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/combobox"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/styled"
)

// Candidate represents a completion candidate.
type Candidate struct {
	// Used in the UI and for filtering.
	ToShow string
	// Used when inserting a candidate.
	ToInsert string
}

type Config struct {
	Binding    el.Handler
	Type       string
	Candidates []Candidate
}

func Start(app *cli.App, cfg Config) {
	w := combobox.Widget{}
	w.CodeArea.Prompt = layout.ModePrompt("COMPLETING "+cfg.Type, true)
	w.ListBox.Horizontal = true
	w.ListBox.OverlayHandler = cfg.Binding
	w.OnFilter = func(p string) {
		w.ListBox.MutateListboxState(func(s *listbox.State) {
			*s = listbox.MakeState(filter(cfg.Candidates, p), false)
		})
	}
	w.ListBox.OnAccept = func(it listbox.Items, i int) {
		text := it.(items)[i].ToInsert
		app.CodeArea.MutateCodeAreaState(func(s *codearea.State) {
			s.CodeBuffer.InsertAtDot(text)
		})
		app.MutateAppState(func(s *cli.State) { s.Listing = nil })
	}
	app.MutateAppState(func(s *cli.State) { s.Listing = &w })
}

type items []Candidate

func filter(all []Candidate, p string) items {
	var filtered []Candidate
	for _, candidate := range all {
		if strings.Contains(candidate.ToShow, p) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func (it items) Show(i int) styled.Text { return styled.Plain(it[i].ToShow) }

func (it items) Len() int { return len(it) }
