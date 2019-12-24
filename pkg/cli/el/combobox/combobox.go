// Package combobox implements the combobox widget, a combination of a listbox
// and a codearea.
package combobox

import (
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/el/codearea"
	"github.com/elves/elvish/pkg/cli/el/listbox"
	"github.com/elves/elvish/pkg/cli/term"
)

// ComboBox is a Widget that combines a ListBox and a CodeArea.
type ComboBox interface {
	el.Widget
	// Returns the embedded codearea widget.
	CodeArea() codearea.CodeArea
	// Returns the embedded listbox widget.
	ListBox() listbox.ListBox
	// Forces the filtering to rerun.
	Refilter()
}

// ComboBoxSpec specifies the configuration and initial state for ComboBox.
type ComboBoxSpec struct {
	CodeArea codearea.CodeAreaSpec
	ListBox  listbox.ListBoxSpec
	OnFilter func(ComboBox, string)
}

type comboBox struct {
	codeArea codearea.CodeArea
	listBox  listbox.ListBox
	OnFilter func(ComboBox, string)

	// Whether filtering has ever been done.
	hasFiltered bool
	// Last filter value.
	lastFilter string
}

// NewComboBox creates a new ComboBox from the given spec.
func NewComboBox(spec ComboBoxSpec) ComboBox {
	if spec.OnFilter == nil {
		spec.OnFilter = func(ComboBox, string) {}
	}
	w := &comboBox{
		codeArea: codearea.NewCodeArea(spec.CodeArea),
		listBox:  listbox.NewListBox(spec.ListBox),
		OnFilter: spec.OnFilter,
	}
	w.OnFilter(w, "")
	return w
}

// Render renders the codearea and the listbox below it.
func (w *comboBox) Render(width, height int) *term.Buffer {
	buf := w.codeArea.Render(width, height)
	bufListBox := w.listBox.Render(width, height-len(buf.Lines))
	buf.Extend(bufListBox, false)
	return buf
}

// Handle first lets the listbox handle the event, and if it is unhandled, lets
// the codearea handle it. If the codearea has handled the event and the code
// content has changed, it calls OnFilter with the new content.
func (w *comboBox) Handle(event term.Event) bool {
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

func (w *comboBox) Refilter() {
	w.OnFilter(w, w.codeArea.CopyState().Buffer.Content)
}

func (w *comboBox) CodeArea() codearea.CodeArea { return w.codeArea }
func (w *comboBox) ListBox() listbox.ListBox    { return w.listBox }
