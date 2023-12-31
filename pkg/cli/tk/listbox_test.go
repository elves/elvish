package tk

import (
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

var listBoxRenderVerticalTests = []renderTest{
	{
		Name:  "placeholder when Items is nil",
		Given: NewListBox(ListBoxSpec{Placeholder: ui.T("nothing")}),
		Width: 10, Height: 3,
		Want: bb(10).Write("nothing"),
	},
	{
		Name: "placeholder when NItems is 0",
		Given: NewListBox(ListBoxSpec{
			Placeholder: ui.T("nothing"),
			State:       ListBoxState{Items: TestItems{}}}),
		Width: 10, Height: 3,
		Want: bb(10).Write("nothing"),
	},
	{
		Name:  "all items when there is enough height",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 2}, Selected: 0}}),
		Width: 10, Height: 3,
		Want: bb(10).
			Write("item 0    ", ui.Inverse).
			Newline().Write("item 1"),
	},
	{
		Name:  "long lines cropped",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 2}, Selected: 0}}),
		Width: 4, Height: 3,
		Want: bb(4).
			Write("item", ui.Inverse).
			Newline().Write("item"),
	},
	{
		Name:  "scrollbar when not showing all items",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 4}, Selected: 0}}),
		Width: 10, Height: 2,
		Want: bb(10).
			Write("item 0   ", ui.Inverse).
			Write(" ", ui.Inverse, ui.FgMagenta).
			Newline().Write("item 1   ").
			Write("│", ui.FgMagenta),
	},
	{
		Name: "scrollbar when not showing last item in full",
		Given: NewListBox(ListBoxSpec{
			State: ListBoxState{
				Items: TestItems{Prefix: "item\n", NItems: 2}, Selected: 0}}),
		Width: 10, Height: 3,
		Want: bb(10).
			Write("item     ", ui.Inverse).
			Write(" ", ui.Inverse, ui.FgMagenta).
			Newline().Write("0        ", ui.Inverse).
			Write(" ", ui.Inverse, ui.FgMagenta).
			Newline().Write("item     ").
			Write(" ", ui.Inverse, ui.FgMagenta),
	},
	{
		Name: "scrollbar when not showing only item in full",
		Given: NewListBox(ListBoxSpec{
			State: ListBoxState{
				Items: TestItems{Prefix: "item\n", NItems: 1}, Selected: 0}}),
		Width: 10, Height: 1,
		Want: bb(10).
			Write("item     ", ui.Inverse).
			Write(" ", ui.Inverse, ui.FgMagenta),
	},
	{
		Name: "padding",
		Given: NewListBox(
			ListBoxSpec{
				Padding: 1,
				State: ListBoxState{
					Items: TestItems{Prefix: "item\n", NItems: 2}, Selected: 0}}),
		Width: 4, Height: 4,

		Want: bb(4).
			Write(" it ", ui.Inverse).Newline().
			Write(" 0  ", ui.Inverse).Newline().
			Write(" it").Newline().
			Write(" 1").Buffer(),
	},
	{
		Name: "not extending style",
		Given: NewListBox(ListBoxSpec{
			Padding: 1,
			State: ListBoxState{
				Items: TestItems{
					Prefix: "x", NItems: 2,
					Style: ui.Stylings(ui.FgBlue, ui.BgGreen)}}}),
		Width: 6, Height: 2,

		Want: bb(6).
			Write(" ", ui.Inverse).
			Write("x0", ui.FgBlue, ui.BgGreen, ui.Inverse).
			Write("   ", ui.Inverse).
			Newline().
			Write(" ").
			Write("x1", ui.FgBlue, ui.BgGreen).
			Buffer(),
	},
	{
		Name: "extending style",
		Given: NewListBox(ListBoxSpec{
			Padding: 1, ExtendStyle: true,
			State: ListBoxState{Items: TestItems{
				Prefix: "x", NItems: 2,
				Style: ui.Stylings(ui.FgBlue, ui.BgGreen)}}}),
		Width: 6, Height: 2,

		Want: bb(6).
			Write(" x0   ", ui.FgBlue, ui.BgGreen, ui.Inverse).
			Newline().
			Write(" x1   ", ui.FgBlue, ui.BgGreen).
			Buffer(),
	},
}

func TestListBox_Render_Vertical(t *testing.T) {
	testRender(t, listBoxRenderVerticalTests)
}

func TestListBox_Render_Vertical_MutatesState(t *testing.T) {
	// Calling Render alters the First field to reflect the first item rendered.
	w := NewListBox(ListBoxSpec{
		State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 4, First: 0}})
	// Items shown will be 3, 4, 5
	w.Render(10, 3)
	state := w.CopyState()
	if first := state.First; first != 3 {
		t.Errorf("State.First = %d, want 3", first)
	}
	if height := state.ContentHeight; height != 3 {
		t.Errorf("State.Height = %d, want 3", height)
	}
}

var listBoxRenderHorizontalTests = []renderTest{
	{
		Name:  "placeholder when Items is nil",
		Given: NewListBox(ListBoxSpec{Horizontal: true, Placeholder: ui.T("nothing")}),
		Width: 10, Height: 3,
		Want: bb(10).Write("nothing"),
	},
	{
		Name: "placeholder when NItems is 0",
		Given: NewListBox(ListBoxSpec{
			Horizontal: true, Placeholder: ui.T("nothing"),
			State: ListBoxState{Items: TestItems{}}}),
		Width: 10, Height: 3,
		Want: bb(10).Write("nothing"),
	},
	{
		Name: "all items when there is enough space, using minimal height",
		Given: NewListBox(ListBoxSpec{
			Horizontal: true,
			State:      ListBoxState{Items: TestItems{NItems: 4}, Selected: 0}}),
		Width: 14, Height: 3,
		// Available height is 3, but only need 2 lines.
		Want: bb(14).
			Write("item 0", ui.Inverse).
			Write("  ").
			Write("item 2").
			Newline().Write("item 1  item 3"),
	},
	{
		Name: "padding",
		Given: NewListBox(ListBoxSpec{
			Horizontal: true, Padding: 1,
			State: ListBoxState{Items: TestItems{NItems: 4, Prefix: "x"}, Selected: 0}}),
		Width: 14, Height: 3,
		Want: bb(14).
			Write(" x0 ", ui.Inverse).
			Write("  ").
			Write(" x2").
			Newline().Write(" x1    x3"),
	},
	{
		Name: "extending style",
		Given: NewListBox(ListBoxSpec{
			Horizontal: true, Padding: 1, ExtendStyle: true,
			State: ListBoxState{Items: TestItems{
				NItems: 2, Prefix: "x",
				Style: ui.Stylings(ui.FgBlue, ui.BgGreen)}}}),
		Width: 14, Height: 3,
		Want: bb(14).
			Write(" x0 ", ui.FgBlue, ui.BgGreen, ui.Inverse).
			Write("  ").
			Write(" x1 ", ui.FgBlue, ui.BgGreen),
	},
	{
		Name: "long lines cropped, with full scrollbar",
		Given: NewListBox(ListBoxSpec{
			Horizontal: true,
			State:      ListBoxState{Items: TestItems{NItems: 2}, Selected: 0}}),
		Width: 4, Height: 3,
		Want: bb(4).
			Write("item", ui.Inverse).
			Newline().Write("item").
			Newline().Write("    ", ui.FgMagenta, ui.Inverse),
	},
	{
		Name: "scrollbar when not showing all items",
		Given: NewListBox(ListBoxSpec{
			Horizontal: true,
			State:      ListBoxState{Items: TestItems{NItems: 4}, Selected: 0}}),
		Width: 6, Height: 3,
		Want: bb(6).
			Write("item 0", ui.Inverse).
			Newline().Write("item 1").
			Newline().
			Write("   ", ui.Inverse, ui.FgMagenta).
			Write("━━━", ui.FgMagenta),
	},
	{
		Name: "scrollbar when not showing all items",
		Given: NewListBox(ListBoxSpec{
			Horizontal: true,
			State:      ListBoxState{Items: TestItems{NItems: 4}, Selected: 0}}),
		Width: 10, Height: 3,
		Want: bb(10).
			Write("item 0", ui.Inverse).Write("  it").
			Newline().Write("item 1  it").
			Newline().
			Write("          ", ui.Inverse, ui.FgMagenta),
	},
	{
		Name: "not showing scrollbar with height = 1",
		Given: NewListBox(ListBoxSpec{
			Horizontal: true,
			State:      ListBoxState{Items: TestItems{NItems: 4}, Selected: 0}}),
		Width: 10, Height: 1,
		Want: bb(10).
			Write("item 0", ui.Inverse).Write("  it"),
	},
}

func TestListBox_Render_Horizontal(t *testing.T) {
	testRender(t, listBoxRenderHorizontalTests)
}

func TestListBox_Render_Horizontal_MutatesState(t *testing.T) {
	// Calling Render alters the First field to reflect the first item rendered.
	w := NewListBox(ListBoxSpec{
		Horizontal: true,
		State: ListBoxState{
			Items: TestItems{Prefix: "x", NItems: 10}, Selected: 4, First: 0}})
	// Only a single column of 3 items shown: x3-x5
	w.Render(2, 4)
	state := w.CopyState()
	if first := state.First; first != 3 {
		t.Errorf("State.First = %d, want 3", first)
	}
	if height := state.ContentHeight; height != 3 {
		t.Errorf("State.Height = %d, want 3", height)
	}
}

var listBoxHandleTests = []handleTest{
	{
		Name:  "up moving selection up",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 1}}),
		Event: term.K(ui.Up),

		WantNewState: ListBoxState{Items: TestItems{NItems: 10}, Selected: 0},
	},
	{
		Name:  "up stopping at 0",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 0}}),
		Event: term.K(ui.Up),

		WantNewState: ListBoxState{Items: TestItems{NItems: 10}, Selected: 0},
	},
	{
		Name:  "up moving to last item when selecting after boundary",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 11}}),
		Event: term.K(ui.Up),

		WantNewState: ListBoxState{Items: TestItems{NItems: 10}, Selected: 9},
	},
	{
		Name:  "down moving selection down",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 1}}),
		Event: term.K(ui.Down),

		WantNewState: ListBoxState{Items: TestItems{NItems: 10}, Selected: 2},
	},
	{
		Name:  "down stopping at n-1",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 9}}),
		Event: term.K(ui.Down),

		WantNewState: ListBoxState{Items: TestItems{NItems: 10}, Selected: 9},
	},
	{
		Name:  "down moving to first item when selecting before boundary",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 10}, Selected: -2}}),
		Event: term.K(ui.Down),

		WantNewState: ListBoxState{Items: TestItems{NItems: 10}, Selected: 0},
	},
	{
		Name:  "enter triggering default no-op accept",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 5}}),
		Event: term.K(ui.Enter),

		WantNewState: ListBoxState{Items: TestItems{NItems: 10}, Selected: 5},
	},
	{
		Name:  "other keys not handled",
		Given: NewListBox(ListBoxSpec{State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 5}}),
		Event: term.K('a'),

		WantUnhandled: true,
	},
	{
		Name: "bindings",
		Given: NewListBox(ListBoxSpec{
			State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 5},
			Bindings: MapBindings{
				term.K('a'): func(w Widget) { w.(*listBox).State.Selected = 0 },
			},
		}),
		Event: term.K('a'),

		WantNewState: ListBoxState{Items: TestItems{NItems: 10}, Selected: 0},
	},
}

func TestListBox_Handle(t *testing.T) {
	testHandle(t, listBoxHandleTests)
}

func TestListBox_Handle_EnterEmitsAccept(t *testing.T) {
	var acceptedItems Items
	var acceptedIndex int
	w := NewListBox(ListBoxSpec{
		OnAccept: func(it Items, i int) {
			acceptedItems = it
			acceptedIndex = i
		},
		State: ListBoxState{Items: TestItems{NItems: 10}, Selected: 5}})
	w.Handle(term.K(ui.Enter))

	if acceptedItems != (TestItems{NItems: 10}) {
		t.Errorf("OnAccept not passed current Items")
	}
	if acceptedIndex != 5 {
		t.Errorf("OnAccept not passed current selected index")
	}
}

func TestListBox_Select_ChangeState(t *testing.T) {
	// number of items = 10, height = 3
	var tests = []struct {
		name   string
		before int
		f      func(ListBoxState) int
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
			w := NewListBox(ListBoxSpec{
				State: ListBoxState{
					Items: TestItems{NItems: 10}, ContentHeight: 3,
					Selected: test.before}})
			w.Select(test.f)
			if selected := w.CopyState().Selected; selected != test.after {
				t.Errorf("selected = %d, want %d", selected, test.after)
			}
		})
	}
}

func TestListBox_Select_CallOnSelect(t *testing.T) {
	it := TestItems{NItems: 10}
	gotItemsCh := make(chan Items, 10)
	gotSelectedCh := make(chan int, 10)
	w := NewListBox(ListBoxSpec{
		OnSelect: func(it Items, i int) {
			gotItemsCh <- it
			gotSelectedCh <- i
		},
		State: ListBoxState{Items: it, Selected: 5}})

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
	w.Select(func(ListBoxState) int { return -1 })
	w.Select(func(ListBoxState) int { return 0 })
	verifyOnSelect(0)
}

func TestListBox_Accept_IndexCheck(t *testing.T) {
	tests := []struct {
		name         string
		nItems       int
		selected     int
		shouldAccept bool
	}{
		{"index in range", 1, 0, true},
		{"index exceeds left boundary", 1, -1, false},
		{"index exceeds right boundary", 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewListBox(ListBoxSpec{
				OnAccept: func(it Items, i int) {
					if !tt.shouldAccept {
						t.Error("should not accept this state")
					}
				},
				State: ListBoxState{
					Items:    TestItems{NItems: tt.nItems},
					Selected: tt.selected,
				},
			})
			w.Accept()
		})
	}
}
