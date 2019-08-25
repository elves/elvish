package combobox

import (
	"testing"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/codearea"
	"github.com/elves/elvish/cli/listbox"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var renderTests = []clitypes.RenderTest{
	{
		Name: "rendering codearea and listbox",
		Given: &Widget{
			CodeArea: codearea.Widget{State: codearea.State{
				CodeBuffer: codearea.CodeBuffer{Content: "filter", Dot: 6},
			}},
			ListBox: listbox.Widget{State: listbox.State{
				Items: listbox.TestItems{NItems: 2},
			}},
		},
		Width: 10, Height: 24,
		Want: ui.NewBufferBuilder(10).
			WritePlain("filter").SetDotToCursor().
			Newline().WriteStyled(styled.MakeText("item 0    ", "inverse")).
			Newline().WritePlain("item 1"),
	},
	{
		Name: "calling filter when rendering",
		Given: installOnFilter(&Widget{
			CodeArea: codearea.Widget{State: codearea.State{
				CodeBuffer: codearea.CodeBuffer{Content: "filter", Dot: 6},
			}},
		}),
		Width: 10, Height: 24,
		Want: ui.NewBufferBuilder(10).
			WritePlain("filter").SetDotToCursor().
			Newline().WriteStyled(styled.MakeText("item 0    ", "inverse")).
			Newline().WritePlain("item 1"),
	},
}

func installOnFilter(w *Widget) *Widget {
	w.OnFilter = func(string) {
		w.ListBox.MutateListboxState(func(s *listbox.State) {
			*s = listbox.State{Items: listbox.TestItems{NItems: 2}}
		})
	}
	return w
}

func TestRender(t *testing.T) {
	clitypes.TestRender(t, renderTests)
}

func TestHandle(t *testing.T) {
	var onFilterCalled bool
	var lastFilter string
	w := &Widget{
		ListBox: listbox.Widget{State: listbox.State{
			Items: listbox.TestItems{NItems: 2},
		}},
		OnFilter: func(filter string) {
			onFilterCalled = true
			lastFilter = filter
		},
	}

	handled := w.Handle(term.K(ui.Down))
	if !handled {
		t.Errorf("listbox did not handle")
	}
	if w.ListBox.State.Selected != 1 {
		t.Errorf("listbox state not changed")
	}

	handled = w.Handle(term.K('a'))
	if !handled {
		t.Errorf("codearea did not handle letter key")
	}
	if w.CodeArea.State.CodeBuffer.Content != "a" {
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
