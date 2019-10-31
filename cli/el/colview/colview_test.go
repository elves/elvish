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
		Given: New(Spec{}),
		Width: 10, Height: 24,
		Want: &ui.Buffer{Width: 10},
	},
	{
		Name: "colview width < number of columns",
		Given: New(Spec{State: State{
			Columns: []el.Widget{
				makeListbox("x", 2, 0), makeListbox("y", 1, 0),
				makeListbox("z", 3, 0), makeListbox("w", 1, 0),
			},
		}}),
		Width: 3, Height: 24,
		Want: &ui.Buffer{Width: 3},
	},
	{
		Name: "colview normal",
		Given: New(Spec{State: State{
			Columns: []el.Widget{
				makeListbox("x", 2, 1),
				makeListbox("y", 1, 0),
				makeListbox("z", 3, -1),
			},
		}}),
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
	return listbox.New(listbox.Spec{
		State: listbox.State{
			Items:    listbox.TestItems{Prefix: prefix, NItems: n},
			Selected: selected,
		}})
}

func TestRender(t *testing.T) {
	el.TestRender(t, renderTests)
}

func TestHandle(t *testing.T) {
	// Channel for recording the place an event was handled. -1 for the widget
	// itself, column index for column.
	handledBy := make(chan int, 10)
	w := New(Spec{
		OverlayHandler: el.MapHandler{
			term.K('a'): func() { handledBy <- -1 },
		},
		State: State{
			Columns: []el.Widget{
				listbox.New(listbox.Spec{
					OverlayHandler: el.MapHandler{
						term.K('a'): func() { handledBy <- 0 },
						term.K('b'): func() { handledBy <- 0 },
					}}),
				listbox.New(listbox.Spec{
					OverlayHandler: el.MapHandler{
						term.K('a'): func() { handledBy <- 1 },
						term.K('b'): func() { handledBy <- 1 },
					}}),
			},
			FocusColumn: 1,
		},
		OnLeft:  func(Widget) { handledBy <- 100 },
		OnRight: func(Widget) { handledBy <- 101 },
	})

	expectHandled := func(event term.Event, wantBy int) {
		t.Helper()
		handled := w.Handle(event)
		if !handled {
			t.Errorf("Handle -> false, want true")
		}
		if by := <-handledBy; by != wantBy {
			t.Errorf("Handled by %d, want %d", by, wantBy)
		}
	}

	expectUnhandled := func(event term.Event) {
		t.Helper()
		handled := w.Handle(event)
		if handled {
			t.Errorf("Handle -> true, want false")
		}
	}

	// Event handled by widget's overlay handler.
	expectHandled(term.K('a'), -1)
	// Event handled by the focused column.
	expectHandled(term.K('b'), 1)
	// Fallback handler for Left
	expectHandled(term.K(ui.Left), 100)
	// Fallback handler for Left
	expectHandled(term.K(ui.Right), 101)
	// No one to handle the event.
	expectUnhandled(term.K('c'))
	// No focused column: event unhandled
	w.MutateColViewState(func(s *State) { s.FocusColumn = -1 })
	expectUnhandled(term.K('b'))
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
