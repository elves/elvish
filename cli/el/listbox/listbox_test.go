package listbox

import (
	"testing"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var bb = ui.NewBufferBuilder

var renderVerticalTests = []el.RenderTest{
	{
		Name:  "placeholder when Items is nil",
		Given: New(Spec{Placeholder: styled.Plain("nothing")}),
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("nothing"),
	},
	{
		Name: "placeholder when NItems is 0",
		Given: New(Spec{
			Placeholder: styled.Plain("nothing"),
			State:       State{Items: TestItems{}}}),
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("nothing"),
	},
	{
		Name:  "all items when there is enough height",
		Given: New(Spec{State: State{Items: TestItems{NItems: 2}, Selected: 0}}),
		Width: 10, Height: 3,
		Want: bb(10).
			WriteStyled(styled.MakeText("item 0    ", "inverse")).
			Newline().WritePlain("item 1"),
	},
	{
		Name:  "long lines cropped",
		Given: New(Spec{State: State{Items: TestItems{NItems: 2}, Selected: 0}}),
		Width: 4, Height: 3,
		Want: bb(4).
			WriteStyled(styled.MakeText("item", "inverse")).
			Newline().WritePlain("item"),
	},
	{
		Name:  "scrollbar when not showing all items",
		Given: New(Spec{State: State{Items: TestItems{NItems: 4}, Selected: 0}}),
		Width: 10, Height: 2,
		Want: bb(10).
			WriteStyled(styled.MakeText("item 0   ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")).
			Newline().WritePlain("item 1   ").
			WriteStyled(styled.MakeText("│", "magenta")),
	},
	{
		Name: "scrollbar when not showing last item in full",
		Given: New(Spec{
			State: State{
				Items: TestItems{Prefix: "item\n", NItems: 2}, Selected: 0}}),
		Width: 10, Height: 3,
		Want: bb(10).
			WriteStyled(styled.MakeText("item     ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")).
			Newline().WriteStyled(styled.MakeText("0        ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")).
			Newline().WritePlain("item     ").
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")),
	},
	{
		Name: "scrollbar when not showing only item in full",
		Given: New(Spec{
			State: State{
				Items: TestItems{Prefix: "item\n", NItems: 1}, Selected: 0}}),
		Width: 10, Height: 1,
		Want: bb(10).
			WriteStyled(styled.MakeText("item     ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")),
	},
	{
		Name: "padding",
		Given: New(
			Spec{
				Padding: 1,
				State: State{
					Items: TestItems{Prefix: "item\n", NItems: 2}, Selected: 0}}),
		Width: 4, Height: 4,

		Want: bb(4).
			WriteStyled(styled.MakeText(" it ", "inverse")).Newline().
			WriteStyled(styled.MakeText(" 0  ", "inverse")).Newline().
			WritePlain(" it").Newline().
			WritePlain(" 1").Buffer(),
	},
	{
		Name: "not extending style",
		Given: New(Spec{
			Padding: 1,
			State: State{
				Items: TestItems{Prefix: "x", NItems: 2, Styles: "blue bg-green"}}}),
		Width: 6, Height: 2,

		Want: bb(6).
			WriteStyled(styled.MakeText(" ", "inverse")).
			WriteStyled(styled.MakeText("x0", "blue", "bg-green", "inverse")).
			WriteStyled(styled.MakeText("   ", "inverse")).
			Newline().
			WritePlain(" ").
			WriteStyled(styled.MakeText("x1", "blue", "bg-green")).
			Buffer(),
	},
	{
		Name: "extending style",
		Given: New(Spec{
			Padding: 1, ExtendStyle: true,
			State: State{Items: TestItems{
				Prefix: "x", NItems: 2, Styles: "blue bg-green"}}}),
		Width: 6, Height: 2,

		Want: bb(6).
			WriteStyled(styled.MakeText(" x0   ", "blue", "bg-green", "inverse")).
			Newline().
			WriteStyled(styled.MakeText(" x1   ", "blue", "bg-green")).
			Buffer(),
	},
}

func TestRender_Vertical(t *testing.T) {
	el.TestRender(t, renderVerticalTests)
}

func TestRender_Vertical_MutatesState(t *testing.T) {
	// Calling Render alters the First field to reflect the first item rendered.
	w := New(Spec{
		State: State{Items: TestItems{NItems: 10}, Selected: 4, First: 0}})
	// Items shown will be 3, 4, 5
	w.Render(10, 3)
	state := w.CopyState()
	if first := state.First; first != 3 {
		t.Errorf("State.First = %d, want 3", first)
	}
	if height := state.Height; height != 3 {
		t.Errorf("State.Height = %d, want 3", height)
	}
}

var renderHorizontalTests = []el.RenderTest{
	{
		Name:  "placeholder when Items is nil",
		Given: New(Spec{Horizontal: true, Placeholder: styled.Plain("nothing")}),
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("nothing"),
	},
	{
		Name: "placeholder when NItems is 0",
		Given: New(Spec{
			Horizontal: true, Placeholder: styled.Plain("nothing"),
			State: State{Items: TestItems{}}}),
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("nothing"),
	},
	{
		Name: "all items when there is enough space, using minimal height",
		Given: New(Spec{
			Horizontal: true,
			State:      State{Items: TestItems{NItems: 4}, Selected: 0}}),
		Width: 14, Height: 3,
		// Available height is 3, but only need 2 lines.
		Want: bb(14).
			WriteStyled(styled.MakeText("item 0", "inverse")).
			WritePlain("  ").
			WritePlain("item 2").
			Newline().WritePlain("item 1  item 3"),
	},
	{
		Name: "padding",
		Given: New(Spec{
			Horizontal: true, Padding: 1,
			State: State{Items: TestItems{NItems: 4, Prefix: "x"}, Selected: 0}}),
		Width: 14, Height: 3,
		Want: bb(14).
			WriteStyled(styled.MakeText(" x0 ", "inverse")).
			WritePlain("  ").
			WritePlain(" x2").
			Newline().WritePlain(" x1    x3"),
	},
	{
		Name: "extending style",
		Given: New(Spec{
			Horizontal: true, Padding: 1, ExtendStyle: true,
			State: State{Items: TestItems{
				NItems: 2, Prefix: "x", Styles: "blue bg-green"}}}),
		Width: 14, Height: 3,
		Want: bb(14).
			WriteStyled(styled.MakeText(" x0 ", "blue", "bg-green", "inverse")).
			WritePlain("  ").
			WriteStyled(styled.MakeText(" x1 ", "blue", "bg-green")),
	},
	{
		Name: "long lines cropped, with full scrollbar",
		Given: New(Spec{
			Horizontal: true,
			State:      State{Items: TestItems{NItems: 2}, Selected: 0}}),
		Width: 4, Height: 3,
		Want: bb(4).
			WriteStyled(styled.MakeText("item", "inverse")).
			Newline().WritePlain("item").
			Newline().WriteStyled(styled.MakeText("    ", "magenta", "inverse")),
	},
	{
		Name: "scrollbar when not showing all items",
		Given: New(Spec{
			Horizontal: true,
			State:      State{Items: TestItems{NItems: 4}, Selected: 0}}),
		Width: 6, Height: 3,
		Want: bb(6).
			WriteStyled(styled.MakeText("item 0", "inverse")).
			Newline().WritePlain("item 1").
			Newline().
			WriteStyled(styled.MakeText("   ", "inverse", "magenta")).
			WriteStyled(styled.MakeText("━━━", "magenta")),
	},
	{
		Name: "scrollbar when not showing all items",
		Given: New(Spec{
			Horizontal: true,
			State:      State{Items: TestItems{NItems: 4}, Selected: 0}}),
		Width: 10, Height: 3,
		Want: bb(10).
			WriteStyled(styled.MakeText("item 0", "inverse")).WritePlain("  it").
			Newline().WritePlain("item 1  it").
			Newline().
			WriteStyled(styled.MakeText("          ", "inverse", "magenta")),
	},
}

func TestRender_Horizontal(t *testing.T) {
	el.TestRender(t, renderHorizontalTests)
}

func TestRender_Horizontal_MutatesState(t *testing.T) {
	// Calling Render alters the First field to reflect the first item rendered.
	w := New(Spec{
		Horizontal: true,
		State: State{
			Items: TestItems{Prefix: "x", NItems: 10}, Selected: 4, First: 0}})
	// Only a single column of 3 items shown: x3-x5
	w.Render(2, 4)
	state := w.CopyState()
	if first := state.First; first != 3 {
		t.Errorf("State.First = %d, want 3", first)
	}
	if height := state.Height; height != 3 {
		t.Errorf("State.Height = %d, want 3", height)
	}
}

var handleTests = []el.HandleTest{
	{
		Name:  "up moving selection up",
		Given: New(Spec{State: State{Items: TestItems{NItems: 10}, Selected: 1}}),
		Event: term.K(ui.Up),

		WantNewState: State{Items: TestItems{NItems: 10}, Selected: 0},
	},
	{
		Name:  "up stopping at 0",
		Given: New(Spec{State: State{Items: TestItems{NItems: 10}, Selected: 0}}),
		Event: term.K(ui.Up),

		WantNewState: State{Items: TestItems{NItems: 10}, Selected: 0},
	},
	{
		Name:  "up moving to last item when selecting after boundary",
		Given: New(Spec{State: State{Items: TestItems{NItems: 10}, Selected: 11}}),
		Event: term.K(ui.Up),

		WantNewState: State{Items: TestItems{NItems: 10}, Selected: 9},
	},
	{
		Name:  "down moving selection down",
		Given: New(Spec{State: State{Items: TestItems{NItems: 10}, Selected: 1}}),
		Event: term.K(ui.Down),

		WantNewState: State{Items: TestItems{NItems: 10}, Selected: 2},
	},
	{
		Name:  "down stopping at n-1",
		Given: New(Spec{State: State{Items: TestItems{NItems: 10}, Selected: 9}}),
		Event: term.K(ui.Down),

		WantNewState: State{Items: TestItems{NItems: 10}, Selected: 9},
	},
	{
		Name:  "down moving to first item when selecting before boundary",
		Given: New(Spec{State: State{Items: TestItems{NItems: 10}, Selected: -2}}),
		Event: term.K(ui.Down),

		WantNewState: State{Items: TestItems{NItems: 10}, Selected: 0},
	},
	{
		Name:  "enter triggering default no-op accept",
		Given: New(Spec{State: State{Items: TestItems{NItems: 10}, Selected: 5}}),
		Event: term.K(ui.Enter),

		WantNewState: State{Items: TestItems{NItems: 10}, Selected: 5},
	},
	{
		Name:  "other keys not handled",
		Given: New(Spec{State: State{Items: TestItems{NItems: 10}, Selected: 5}}),
		Event: term.K('a'),

		WantUnhandled: true,
	},
	{
		Name: "overlay handler",
		Given: addOverlay(
			New(Spec{State: State{Items: TestItems{NItems: 10}, Selected: 5}}),
			func(w *widget) el.Handler {
				return el.MapHandler{
					term.K('a'): func() { w.State.Selected = 0 },
				}
			}),
		Event: term.K('a'),

		WantNewState: State{Items: TestItems{NItems: 10}, Selected: 0},
	},
}

func addOverlay(w Widget, overlay func(*widget) el.Handler) *widget {
	ww := w.(*widget)
	ww.OverlayHandler = overlay(ww)
	return ww
}

func TestHandle(t *testing.T) {
	el.TestHandle(t, handleTests)
}

func TestHandle_EnterEmitsAccept(t *testing.T) {
	var acceptedItems Items
	var acceptedIndex int
	w := New(Spec{
		OnAccept: func(it Items, i int) {
			acceptedItems = it
			acceptedIndex = i
		},
		State: State{Items: TestItems{NItems: 10}, Selected: 5}})
	w.Handle(term.K(ui.Enter))

	if acceptedItems != (TestItems{NItems: 10}) {
		t.Errorf("OnAccept not passed current Items")
	}
	if acceptedIndex != 5 {
		t.Errorf("OnAccept not passed current selected index")
	}
}

func TestSelect_ChangeState(t *testing.T) {
	// number of items = 10, height = 3
	var tests = []struct {
		name   string
		before int
		f      func(State) int
		after  int
	}{
		{"Next from -1", -1, Next, 0},
		{"Next from 0", 0, Next, 1},
		{"Next from 9", 9, Next, 9},
		{"Next from 10", 10, Next, 9},

		{"NextWrap from -1", -1, NextWrap, 0},
		{"NextWrap from 0", 0, NextWrap, 1},
		{"NextWrap from 9", 9, NextWrap, 0},
		{"NextWrap from 10", 10, NextWrap, 0},

		{"NextPage from -1", -1, NextPage, 2},
		{"NextPage from 0", 0, NextPage, 3},
		{"NextPage from 9", 9, NextPage, 9},
		{"NextPage from 10", 10, NextPage, 9},

		{"Prev from -1", -1, Prev, 0},
		{"Prev from 0", 0, Prev, 0},
		{"Prev from 9", 9, Prev, 8},
		{"Prev from 10", 10, Prev, 9},

		{"PrevWrap from -1", -1, PrevWrap, 9},
		{"PrevWrap from 0", 0, PrevWrap, 9},
		{"PrevWrap from 9", 9, PrevWrap, 8},
		{"PrevWrap from 10", 10, PrevWrap, 9},

		{"PrevPage from -1", -1, PrevPage, 0},
		{"PrevPage from 0", 0, PrevPage, 0},
		{"PrevPage from 9", 9, PrevPage, 6},
		{"PrevPage from 10", 10, PrevPage, 7},

		{"Left from -1", -1, Left, 0},
		{"Left from 0", 0, Left, 0},
		{"Left from 9", 9, Left, 6},
		{"Left from 10", 10, Left, 6},

		{"Right from -1", -1, Right, 3},
		{"Right from 0", 0, Right, 3},
		{"Right from 9", 9, Right, 9},
		{"Right from 10", 10, Right, 9},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := New(Spec{
				State: State{
					Items: TestItems{NItems: 10}, Height: 3,
					Selected: test.before}})
			w.Select(test.f)
			if selected := w.CopyState().Selected; selected != test.after {
				t.Errorf("selected = %d, want %d", selected, test.after)
			}
		})
	}
}

func TestSelect_CallOnSelect(t *testing.T) {
	it := TestItems{NItems: 10}
	gotItemsCh := make(chan Items, 10)
	gotSelectedCh := make(chan int, 10)
	w := New(Spec{
		OnSelect: func(it Items, i int) {
			gotItemsCh <- it
			gotSelectedCh <- i
		},
		State: State{Items: it, Selected: 5}})

	verifyOnSelect := func(wantSelected int) {
		if gotItems := <-gotItemsCh; gotItems != it {
			t.Errorf("Got it = %v, want %v", gotItems, it)
		}
		if gotSelected := <-gotSelectedCh; gotSelected != wantSelected {
			t.Errorf("Got selected = %v, want %v", gotSelected, wantSelected)
		}
	}

	// Test that OnSelect is called during initialization.
	verifyOnSelect(5)
	// Test that OnSelect is called when changing selection.
	w.Select(Next)
	verifyOnSelect(6)
	// Test that OnSelect is not called when index is invalid. Instead of
	// waiting a fixed time to make sure that nothing is sent in the channel, we
	// immediately does another Select with a valid index, and verify that only
	// the valid index is sent.
	w.Select(func(State) int { return -1 })
	w.Select(func(State) int { return 0 })
	verifyOnSelect(0)
}
