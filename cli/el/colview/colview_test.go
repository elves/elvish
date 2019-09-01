package colview

import (
	"testing"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/tt"
)

var renderTests = []el.RenderTest{
	{
		Name:  "colview no column",
		Given: &Widget{},
		Width: 10, Height: 24,
		Want: &ui.Buffer{Width: 10},
	},
	{
		Name: "colview width < number of columns",
		Given: &Widget{State: State{
			Columns: []el.Widget{
				makeListbox("x", 2, 0), makeListbox("y", 1, 0),
				makeListbox("z", 3, 0), makeListbox("w", 1, 0),
			},
		}},
		Width: 3, Height: 24,
		Want: &ui.Buffer{Width: 3},
	},
	{
		Name: "colview normal",
		Given: &Widget{State: State{
			Columns: []el.Widget{
				makeListbox("x", 2, 1),
				makeListbox("y", 1, 0),
				makeListbox("z", 3, -1),
			},
		}},
		Width: 11, Height: 24,
		Want: ui.NewBufferBuilder(11).
			// first line
			WritePlain("x0  ").
			WriteStyled(styled.MakeText("y0 ", "inverse")).
			WritePlain(" z0").
			// second line
			Newline().WriteStyled(styled.MakeText("x1 ", "inverse")).
			WritePlain("     z1").
			// third line
			Newline().WritePlain("        z2"),
	},
}

func makeListbox(prefix string, n, selected int) el.Widget {
	return &listbox.Widget{State: listbox.State{
		Items:    listbox.TestItems{Prefix: prefix, NItems: n},
		Selected: selected,
	}}
}

func TestRender(t *testing.T) {
	el.TestRender(t, renderTests)
}

func TestHandle(t *testing.T) {
	// Channel for recording the place an event was handled. -1 for the widget
	// itself, column index for column.
	handledBy := make(chan int, 10)
	w := &Widget{
		OverlayHandler: el.MapHandler{
			term.K('a'): func() { handledBy <- -1 },
		},
		State: State{
			Columns: []el.Widget{
				&listbox.Widget{OverlayHandler: el.MapHandler{
					term.K('a'): func() { handledBy <- 0 },
					term.K('b'): func() { handledBy <- 0 },
				}},
				&listbox.Widget{OverlayHandler: el.MapHandler{
					term.K('a'): func() { handledBy <- 1 },
					term.K('b'): func() { handledBy <- 1 },
				}},
			},
			FocusColumn: 1,
		},
	}

	// Event handled by widget's overlay handler.
	handled := w.Handle(term.K('a'))
	if !handled {
		t.Errorf("Handle -> false, want true")
	}
	if by := <-handledBy; by != -1 {
		t.Errorf("Handled by %d, want -1", by)
	}

	// Event handled by the focused column.
	handled = w.Handle(term.K('b'))
	if !handled {
		t.Errorf("Handle -> false, want true")
	}
	if by := <-handledBy; by != 1 {
		t.Errorf("Handled by %d, want 1", by)
	}

	// Event unhandled.
	handled = w.Handle(term.K('c'))
	if handled {
		t.Errorf("Handle -> true, want false")
	}

	// No focused column: event unhandled
	w.MutateColViewState(func(s *State) { s.FocusColumn = -1 })
	handled = w.Handle(term.K('b'))
	if handled {
		t.Errorf("Handle -> true, want false")
	}
}

func TestDistribute(t *testing.T) {
	tt.Test(t, tt.Fn("distribute", distribute), tt.Table{
		// Nice integer distributions.
		tt.Args(10, []int{1, 1}).Rets([]int{5, 5}),
		tt.Args(10, []int{2, 3}).Rets([]int{4, 6}),
		tt.Args(10, []int{1, 2, 2}).Rets([]int{2, 4, 4}),
		// Approximate integer distributions.
		tt.Args(10, []int{1, 1, 1}).Rets([]int{3, 3, 4}),
		tt.Args(5, []int{1, 1, 1}).Rets([]int{1, 2, 2}),
	})
}
