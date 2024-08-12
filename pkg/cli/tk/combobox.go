package tk

import (
	"src.elv.sh/pkg/cli/term"
)

// ComboBox is a Widget that combines a ListBox and a CodeArea.
type ComboBox interface {
	Widget
	// Returns the embedded codearea widget.
	CodeArea() CodeArea
	// Returns the embedded listbox widget.
	ListBox() ListBox
	// Forces the filtering to rerun.
	Refilter()
}

// ComboBoxSpec specifies the configuration and initial state for ComboBox.
type ComboBoxSpec struct {
	CodeArea CodeAreaSpec
	ListBox  ListBoxSpec
	OnFilter func(ComboBox, string)
}

type comboBox struct {
	codeArea CodeArea
	listBox  ListBox
	OnFilter func(ComboBox, string)

	// Last filter value.
	lastFilter string
}

// NewComboBox creates a new ComboBox from the given spec.
func NewComboBox(spec ComboBoxSpec) ComboBox {
	if spec.OnFilter == nil {
		spec.OnFilter = func(ComboBox, string) {}
	}
	w := &comboBox{
		codeArea: NewCodeArea(spec.CodeArea),
		listBox:  NewListBox(spec.ListBox),
		OnFilter: spec.OnFilter,
	}
	w.OnFilter(w, "")
	return w
}

// Render renders the codearea and the listbox below it.
func (w *comboBox) Render(width, height int) *term.Buffer {
	// TODO: Test the behavior of Render when height is very small
	// (https://b.elv.sh/1820)
	if height == 1 {
		return w.listBox.Render(width, height)
	}
	buf := w.codeArea.Render(width, height-1)
	bufListBox := w.listBox.Render(width, height-len(buf.Lines))
	buf.Extend(bufListBox, false)
	return buf
}

func (w *comboBox) MaxHeight(width, height int) int {
	return w.codeArea.MaxHeight(width, height) + w.listBox.MaxHeight(width, height)
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

func (w *comboBox) CodeArea() CodeArea { return w.codeArea }
func (w *comboBox) ListBox() ListBox   { return w.listBox }
