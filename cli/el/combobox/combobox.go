// Package combobox implements the combobox widget, a combination of a listbox
// and a codearea.
package combobox

import (
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/cli/term"
)

// Widget represents a combobox widget.
type Widget interface {
	el.Widget
	// Returns the embedded codearea widget.
	CodeArea() codearea.Widget
	// Returns the embedded listbox widget.
	ListBox() listbox.Widget
	// Forces the filtering to rerun.
	Refilter()
}

// Spec specifies the configuration and initial state for Widget.
type Spec struct {
	CodeArea codearea.Spec
	ListBox  listbox.Spec
	OnFilter func(Widget, string)
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

// New creates a Widget with the given specification.
func New(spec Spec) Widget {
	if spec.OnFilter == nil {
		spec.OnFilter = func(Widget, string) {}
	}
	w := &widget{
		codeArea: codearea.New(spec.CodeArea),
		listBox:  listbox.New(spec.ListBox),
		OnFilter: spec.OnFilter,
	}
	w.OnFilter(w, "")
	return w
}

// Render renders the codearea and the listbox below it.
func (w *widget) Render(width, height int) *term.Buffer {
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
		filter := w.codeArea.CopyState().Buffer.Content
		if filter != w.lastFilter {
			w.OnFilter(w, filter)
			w.lastFilter = filter
		}
		return true
	}
	return false
}

func (w *widget) Refilter() {
	w.OnFilter(w, w.codeArea.CopyState().Buffer.Content)
}

func (w *widget) CodeArea() codearea.Widget { return w.codeArea }
func (w *widget) ListBox() listbox.Widget   { return w.listBox }
