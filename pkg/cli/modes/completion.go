package modes

import (
	"errors"
	"sort"
	"strings"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
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
	tags     []string
	tagIndex int
}

// CompletionItem represents a completion item, also known as a candidate.
type CompletionItem struct {
	// Used in the UI and for filtering.
	ToShow ui.Text
	// Used when inserting a candidate.
	ToInsert string
	// TODO
	Tag string
}

type completion struct {
	tk.ComboBox
	attached tk.CodeArea
}

var errNoCandidates = errors.New("no candidates")

// NewCompletion starts the completion UI.
func NewCompletion(app cli.App, cfg CompletionSpec) (Completion, error) {
	itemsByTag := make(map[string][]CompletionItem)
	for _, item := range cfg.Items {
		itemsByTag[item.Tag] = append(itemsByTag[item.Tag], item)
	}
	// TODO presort items?
	for tag := range itemsByTag {
		if strings.TrimSpace(tag) != "" {
			cfg.tags = append(cfg.tags, tag)
		}
	}
	sort.Strings(cfg.tags)
	switch len(cfg.tags) {
	case 1:
		if strings.TrimSpace(cfg.tags[0]) == "" {
			cfg.tags = []string{cfg.Name}
		}
	default:
		cfg.tags = append([]string{cfg.Name}, cfg.tags...)
	}

	codeArea, err := FocusedCodeArea(app)
	if err != nil {
		return nil, err
	}
	if len(cfg.Items) == 0 {
		return nil, errNoCandidates
	}
	var w tk.ComboBox
	w = tk.NewComboBox(tk.ComboBoxSpec{
		CodeArea: tk.CodeAreaSpec{
			Bindings: tk.MapBindings{
				term.K('n', ui.Alt): func(tk.Widget) {
					if len(cfg.tags) == 1 {
						return
					}

					cfg.tagIndex += 1
					if cfg.tagIndex >= len(cfg.tags) {
						cfg.tagIndex = 0
					}
					w.Refilter()
				},
				term.K('p', ui.Alt): func(t tk.Widget) {
					if len(cfg.tags) == 1 {
						return
					}

					cfg.tagIndex -= 1
					if cfg.tagIndex < 0 {
						cfg.tagIndex = len(cfg.tags) - 1
					}
					w.Refilter()
				},
			},
			Prompt: func() ui.Text {
				prompt := modePrompt(" COMPLETING "+cfg.tags[cfg.tagIndex]+" ", false)()
				if cfg.tagIndex > 0 {
					prompt = ui.StyleText(prompt, ui.Inverse)
				}
				prompt = ui.Concat(prompt, ui.T(" "))
				return prompt
			},
			Highlighter: cfg.Filter.Highlighter,
		},
		ListBox: tk.ListBoxSpec{
			Horizontal: true,
			Bindings:   cfg.Bindings,
			OnSelect: func(it tk.Items, i int) {
				text := it.(completionItems)[i].ToInsert
				codeArea.MutateState(func(s *tk.CodeAreaState) {
					s.Pending = tk.PendingCode{
						From: cfg.Replace.From, To: cfg.Replace.To, Content: text}
				})
			},
			OnAccept: func(it tk.Items, i int) {
				codeArea.MutateState((*tk.CodeAreaState).ApplyPending)
				app.PopAddon()
			},
			ExtendStyle: true,
		},
		OnFilter: func(w tk.ComboBox, p string) {
			filtered := cfg.Items
			if cfg.tagIndex > 0 {
				filtered = itemsByTag[cfg.tags[cfg.tagIndex]]
			}
			w.ListBox().Reset(filterCompletionItems(filtered, cfg.Filter.makePredicate(p)), 0)
		},
	})
	return completion{w, codeArea}, nil
}

func (w completion) Dismiss() {
	w.attached.MutateState(func(s *tk.CodeAreaState) { s.Pending = tk.PendingCode{} })
}

type completionItems []CompletionItem

func filterCompletionItems(all []CompletionItem, p func(string) bool) completionItems {
	var filtered []CompletionItem
	for _, candidate := range all {
		if p(unstyle(candidate.ToShow)) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func (it completionItems) Show(i int) ui.Text { return it[i].ToShow }
func (it completionItems) Len() int           { return len(it) }

func unstyle(t ui.Text) string {
	var sb strings.Builder
	for _, seg := range t {
		sb.WriteString(seg.Text)
	}
	return sb.String()
}
