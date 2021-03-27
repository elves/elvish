package mode

import (
	"errors"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/ui"
)

// Completion is a mode specialized for viewing and inserting completion
// candidates. It is based on the ComboBox widget.
type Completion interface {
	tk.ComboBox
}

// CompletionSpec specifies the configuration for the completion mode.
type CompletionSpec struct {
	Bindings tk.Bindings
	Name     string
	Replace  diag.Ranging
	Items    []CompletionItem
	Filter   FilterSpec
}

// CompletionItem represents a completion item, also known as a candidate.
type CompletionItem struct {
	// Used in the UI and for filtering.
	ToShow string
	// Style to use in the UI.
	ShowStyle ui.Style
	// Used when inserting a candidate.
	ToInsert string
}

type completion struct {
	tk.ComboBox
	attached tk.CodeArea
}

var errNoCandidates = errors.New("no candidates")

// NewCompletion starts the completion UI.
func NewCompletion(app cli.App, cfg CompletionSpec) (Completion, error) {
	if len(cfg.Items) == 0 {
		return nil, errNoCandidates
	}
	if cfg.Filter.Maker == nil {
		cfg.Filter.Maker = makeSubstringFilter
	}
	w := tk.NewComboBox(tk.ComboBoxSpec{
		CodeArea: tk.CodeAreaSpec{
			Prompt:      ModePrompt(" COMPLETING "+cfg.Name+" ", true),
			Highlighter: cfg.Filter.Highlighter,
		},
		ListBox: tk.ListBoxSpec{
			Horizontal: true,
			Bindings:   cfg.Bindings,
			OnSelect: func(it tk.Items, i int) {
				text := it.(completionItems)[i].ToInsert
				app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
					s.Pending = tk.PendingCode{
						From: cfg.Replace.From, To: cfg.Replace.To, Content: text}
				})
			},
			OnAccept: func(it tk.Items, i int) {
				app.SetAddon(nil, true)
			},
			ExtendStyle: true,
		},
		OnFilter: func(w tk.ComboBox, p string) {
			w.ListBox().Reset(filterCompletionItems(cfg.Items, cfg.Filter.Maker(p)), 0)
		},
	})
	return completion{w, app.CodeArea()}, nil
}

func (w completion) Close(accept bool) {
	w.attached.MutateState(func(s *tk.CodeAreaState) {
		if accept {
			s.ApplyPending()
		} else {
			s.Pending = tk.PendingCode{}
		}
	})
}

type completionItems []CompletionItem

func filterCompletionItems(all []CompletionItem, p func(string) bool) completionItems {
	var filtered []CompletionItem
	for _, candidate := range all {
		if p(candidate.ToShow) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func (it completionItems) Show(i int) ui.Text {
	return ui.Text{&ui.Segment{Style: it[i].ShowStyle, Text: it[i].ToShow}}
}

func (it completionItems) Len() int { return len(it) }
