// Package combobox implements the combobox widget, a combination of a listbox
// and a codearea.
package combobox

import (
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

// Widget represents a combobox widget.
type Widget interface {
	el.Widget
	CodeArea() codearea.Widget
	ListBox() listbox.Widget
}

// Config keeps the configuration for Widget.
type Config struct {
	CodeArea codearea.Config
	ListBox  listbox.Config
	OnFilter func(Widget, string)
}

// InitState keeps the initial state of Widget.
type InitState struct {
	CodeArea codearea.State
	ListBox  listbox.State
}

type widget struct {
	codeArea codearea.Widget
	listBox  listbox.Widget
	OnFilter func(Widget, string)

	// Whether filtering has ever been done.
	hasFiltered bool
	// Last filter value.
	lastFilter string
}

// New creates a Widget with the given config.
func New(cfg Config) Widget {
	return NewWithState(cfg, InitState{})
}

// NewWithState creates a Widget with the given config and initial state.
func NewWithState(cfg Config, state InitState) Widget {
	if cfg.OnFilter == nil {
		cfg.OnFilter = func(Widget, string) {}
	}
	w := &widget{
		codeArea: codearea.NewWithState(cfg.CodeArea, state.CodeArea),
		listBox:  listbox.NewWithState(cfg.ListBox, state.ListBox),
		OnFilter: cfg.OnFilter,
	}
	w.OnFilter(w, "")
	return w
}

// Render renders the codearea and the listbox below it.
func (w *widget) Render(width, height int) *ui.Buffer {
	buf := w.codeArea.Render(width, height)
	bufListBox := w.listBox.Render(width, height-len(buf.Lines))
	buf.Extend(bufListBox, false)
	return buf
}

// Handle first lets the listbox handle the event, and if it is unhandled, lets
// the codearea handle it. If the codearea has handled the event and the code
// content has changed, it calls OnFilter with the new content.
func (w *widget) Handle(event term.Event) bool {
	if w.listBox.Handle(event) {
		return true
	}
	if w.codeArea.Handle(event) {
		filter := w.codeArea.CopyState().CodeBuffer.Content
		if filter != w.lastFilter {
			w.OnFilter(w, filter)
			w.lastFilter = filter
		}
		return true
	}
	return false
}

func (w *widget) CodeArea() codearea.Widget { return w.codeArea }
func (w *widget) ListBox() listbox.Widget   { return w.listBox }
