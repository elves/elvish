package tk

import (
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/tt"
	"src.elv.sh/pkg/ui"
)

var colViewRenderTests = []renderTest{
	{
		Name:  "colview no column",
		Given: NewColView(ColViewSpec{}),
		Width: 10, Height: 24,
		Want: &term.Buffer{Width: 10},
	},
	{
		Name: "colview width < number of columns",
		Given: NewColView(ColViewSpec{State: ColViewState{
			Columns: []Widget{
				makeListbox("x", 2, 0), makeListbox("y", 1, 0),
				makeListbox("z", 3, 0), makeListbox("w", 1, 0),
			},
		}}),
		Width: 3, Height: 24,
		Want: &term.Buffer{Width: 3},
	},
	{
		Name: "colview normal",
		Given: NewColView(ColViewSpec{State: ColViewState{
			Columns: []Widget{
				makeListbox("x", 2, 1),
				makeListbox("y", 1, 0),
				makeListbox("z", 3, -1),
			},
		}}),
		Width: 11, Height: 24,
		Want: term.NewBufferBuilder(11).
			// first line
			Write("x0  ").
			Write("y0 ", ui.Inverse).
			Write(" z0").
			// second line
			Newline().Write("x1 ", ui.Inverse).
			Write("     z1").
			// third line
			Newline().Write("        z2"),
	},
}

func makeListbox(prefix string, n, selected int) Widget {
	return NewListBox(ListBoxSpec{
		State: ListBoxState{
			Items:    TestItems{Prefix: prefix, NItems: n},
			Selected: selected,
		}})
}

func TestColView_Render(t *testing.T) {
	testRender(t, colViewRenderTests)
}

func TestColView_Handle(t *testing.T) {
	// Channel for recording the place an event was handled. -1 for the widget
	// itself, column index for column.
	handledBy := make(chan int, 10)
	w := NewColView(ColViewSpec{
		Bindings: MapBindings{
			term.K('a'): func(Widget) { handledBy <- -1 },
		},
		State: ColViewState{
			Columns: []Widget{
				NewListBox(ListBoxSpec{
					Bindings: MapBindings{
						term.K('a'): func(Widget) { handledBy <- 0 },
						term.K('b'): func(Widget) { handledBy <- 0 },
					}}),
				NewListBox(ListBoxSpec{
					Bindings: MapBindings{
						term.K('a'): func(Widget) { handledBy <- 1 },
						term.K('b'): func(Widget) { handledBy <- 1 },
					}}),
			},
			FocusColumn: 1,
		},
		OnLeft:  func(ColView) { handledBy <- 100 },
		OnRight: func(ColView) { handledBy <- 101 },
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
	w.MutateState(func(s *ColViewState) { s.FocusColumn = -1 })
	expectUnhandled(term.K('b'))
}

func TestDistribute(t *testing.T) {
	tt.Test(t, distribute,
		// Nice integer distributions.
		Args(10, []int{1, 1}).Rets([]int{5, 5}),
		Args(10, []int{2, 3}).Rets([]int{4, 6}),
		Args(10, []int{1, 2, 2}).Rets([]int{2, 4, 4}),
		// Approximate integer distributions.
		Args(10, []int{1, 1, 1}).Rets([]int{3, 3, 4}),
		Args(5, []int{1, 1, 1}).Rets([]int{1, 2, 2}),
	)
}
