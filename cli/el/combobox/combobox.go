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
type Widget struct {
	CodeArea codearea.Widget
	ListBox  listbox.Widget
	OnFilter func(string)

	// Whether filtering has ever been done.
	hasFiltered bool
	// Last filter value.
	lastFilter string
}

var _ = el.Widget(&Widget{})

func (w *Widget) init() {
	if w.OnFilter == nil {
		w.OnFilter = func(string) {}
	}
	if w.ListBox == nil {
		w.ListBox = listbox.New(listbox.Config{})
	}
	if !w.hasFiltered {
		w.OnFilter("")
		w.hasFiltered = true
	}
}

// Render renders the codearea and the listbox below it.
func (w *Widget) Render(width, height int) *ui.Buffer {
	w.init()
	buf := w.CodeArea.Render(width, height)
	bufListBox := w.ListBox.Render(width, height-len(buf.Lines))
	buf.Extend(bufListBox, false)
	return buf
}

// Handle first lets the listbox handle the event, and if it is unhandled, lets
// the codearea handle it. If the codearea has handled the event and the code
// content has changed, it calls OnFilter with the new content.
func (w *Widget) Handle(event term.Event) bool {
	w.init()
	if w.ListBox.Handle(event) {
		return true
	}
	if w.CodeArea.Handle(event) {
		filter := w.CodeArea.CopyState().CodeBuffer.Content
		if filter != w.lastFilter {
			w.OnFilter(filter)
			w.lastFilter = filter
		}
		return true
	}
	return false
}
