package combobox

import (
	"testing"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var renderTests = []el.RenderTest{
	{
		Name: "rendering codearea and listbox",
		Given: New(Spec{
			CodeArea: codearea.Spec{
				State: codearea.State{
					Buffer: codearea.Buffer{Content: "filter", Dot: 6}}},
			ListBox: listbox.Spec{
				State: listbox.State{Items: listbox.TestItems{NItems: 2}}}}),
		Width: 10, Height: 24,
		Want: ui.NewBufferBuilder(10).
			WritePlain("filter").SetDotToCursor().
			Newline().WriteStyled(styled.MakeText("item 0    ", "inverse")).
			Newline().WritePlain("item 1"),
	},
	{
		Name: "calling filter before rendering",
		Given: New(Spec{
			CodeArea: codearea.Spec{
				State: codearea.State{
					Buffer: codearea.Buffer{Content: "filter", Dot: 6}}},
			OnFilter: func(w Widget, filter string) {
				w.ListBox().Reset(listbox.TestItems{NItems: 2}, 0)
			}}),
		Width: 10, Height: 24,
		Want: ui.NewBufferBuilder(10).
			WritePlain("filter").SetDotToCursor().
			Newline().WriteStyled(styled.MakeText("item 0    ", "inverse")).
			Newline().WritePlain("item 1"),
	},
}

func TestRender(t *testing.T) {
	el.TestRender(t, renderTests)
}

func TestHandle(t *testing.T) {
	var onFilterCalled bool
	var lastFilter string
	w := New(Spec{
		OnFilter: func(w Widget, filter string) {
			onFilterCalled = true
			lastFilter = filter
		},
		ListBox: listbox.Spec{
			State: listbox.State{Items: listbox.TestItems{NItems: 2}}}})

	handled := w.Handle(term.K(ui.Down))
	if !handled {
		t.Errorf("listbox did not handle")
	}
	if w.ListBox().CopyState().Selected != 1 {
		t.Errorf("listbox state not changed")
	}

	handled = w.Handle(term.K('a'))
	if !handled {
		t.Errorf("codearea did not handle letter key")
	}
	if w.CodeArea().CopyState().Buffer.Content != "a" {
		t.Errorf("codearea state not changed")
	}
	if lastFilter != "a" {
		t.Errorf("OnFilter not called when codearea content changed")
	}

	onFilterCalled = false
	handled = w.Handle(term.PasteSetting(true))
	if !handled {
		t.Errorf("codearea did not handle PasteSetting")
	}
	if onFilterCalled {
		t.Errorf("OnFilter called when codearea content did not change")
	}
	w.Handle(term.PasteSetting(false))

	handled = w.Handle(term.K('D', ui.Ctrl))
	if handled {
		t.Errorf("key unhandled by codearea and listbox got handled")
	}
}
