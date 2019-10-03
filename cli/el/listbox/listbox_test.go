package listbox

import (
	"reflect"
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
		Given: &Widget{Placeholder: styled.Plain("nothing")},
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("nothing"),
	},
	{
		Name: "placeholder when NItems is 0",
		Given: &Widget{
			Placeholder: styled.Plain("nothing"),
			State:       State{Items: TestItems{}},
		},
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("nothing"),
	},
	{
		Name:  "all items when there is enough height",
		Given: &Widget{State: State{Items: TestItems{NItems: 2}, Selected: 0}},
		Width: 10, Height: 3,
		Want: bb(10).
			WriteStyled(styled.MakeText("item 0    ", "inverse")).
			Newline().WritePlain("item 1"),
	},
	{
		Name:  "long lines cropped",
		Given: &Widget{State: State{Items: TestItems{NItems: 2}, Selected: 0}},
		Width: 4, Height: 3,
		Want: bb(4).
			WriteStyled(styled.MakeText("item", "inverse")).
			Newline().WritePlain("item"),
	},
	{
		Name:  "scrollbar when not showing all items",
		Given: &Widget{State: State{Items: TestItems{NItems: 4}, Selected: 0}},
		Width: 10, Height: 2,
		Want: bb(10).
			WriteStyled(styled.MakeText("item 0   ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")).
			Newline().WritePlain("item 1   ").
			WriteStyled(styled.MakeText("│", "magenta")),
	},
	{
		Name:  "scrollbar when not showing last item in full",
		Given: &Widget{State: State{Items: TestItems{Prefix: "item\n", NItems: 2}, Selected: 0}},
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
		Name:  "scrollbar when not showing only item in full",
		Given: &Widget{State: State{Items: TestItems{Prefix: "item\n", NItems: 1}, Selected: 0}},
		Width: 10, Height: 1,
		Want: bb(10).
			WriteStyled(styled.MakeText("item     ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")),
	},
	{
		Name: "padding",
		Given: &Widget{
			State: State{
				Items: TestItems{Prefix: "item\n", NItems: 2}, Selected: 0},
			Padding: 1,
		},
		Width: 4, Height: 4,

		Want: bb(4).
			WriteStyled(styled.MakeText(" it ", "inverse")).Newline().
			WriteStyled(styled.MakeText(" 0  ", "inverse")).Newline().
			WritePlain(" it").Newline().
			WritePlain(" 1").Buffer(),
	},
}

func TestRenderVertical(t *testing.T) {
	el.TestRender(t, renderVerticalTests)
}

var renderHorizontalTests = []el.RenderTest{
	{
		Name:  "placeholder when Items is nil",
		Given: &Widget{Horizontal: true, Placeholder: styled.Plain("nothing")},
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("nothing"),
	},
	{
		Name: "placeholder when NItems is 0",
		Given: &Widget{
			Horizontal:  true,
			Placeholder: styled.Plain("nothing"),
			State:       State{Items: TestItems{}},
		},
		Width: 10, Height: 3,
		Want: bb(10).WritePlain("nothing"),
	},
	{
		Name: "all items when there is enough space, using minimal height",
		Given: &Widget{
			Horizontal: true,
			State:      State{Items: TestItems{NItems: 4}, Selected: 0},
		},
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
		Given: &Widget{
			Horizontal: true,
			Padding:    1,
			State:      State{Items: TestItems{NItems: 4, Prefix: "x"}, Selected: 0},
		},
		Width: 14, Height: 3,
		Want: bb(14).
			WriteStyled(styled.MakeText(" x0 ", "inverse")).
			WritePlain("  ").
			WritePlain(" x2").
			Newline().WritePlain(" x1    x3"),
	},
	// TODO(xiaq): Add test for padding.
	{
		Name: "long lines cropped, with full scrollbar",
		Given: &Widget{
			Horizontal: true,
			State:      State{Items: TestItems{NItems: 2}, Selected: 0},
		},
		Width: 4, Height: 3,
		Want: bb(4).
			WriteStyled(styled.MakeText("item", "inverse")).
			Newline().WritePlain("item").
			Newline().WriteStyled(styled.MakeText("    ", "magenta", "inverse")),
	},
	{
		Name: "scrollbar when not showing all items",
		Given: &Widget{
			Horizontal: true,
			State:      State{Items: TestItems{NItems: 4}, Selected: 0},
		},
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
		Given: &Widget{
			Horizontal: true,
			State:      State{Items: TestItems{NItems: 4}, Selected: 0},
		},
		Width: 10, Height: 3,
		Want: bb(10).
			WriteStyled(styled.MakeText("item 0", "inverse")).WritePlain("  it").
			Newline().WritePlain("item 1  it").
			Newline().
			WriteStyled(styled.MakeText("          ", "inverse", "magenta")),
	},
}

func TestRenderHorizontal(t *testing.T) {
	el.TestRender(t, renderHorizontalTests)
}

var handleTests = []struct {
	name        string
	widget      *Widget
	event       term.Event
	wantHandled bool
	wantState   State
}{
	{
		"up moving selection up",
		&Widget{State: State{Items: TestItems{NItems: 10}, Selected: 1}},
		term.K(ui.Up),
		true,
		State{Items: TestItems{NItems: 10}, Selected: 0},
	},
	{
		"up stopping at 0",
		&Widget{State: State{Items: TestItems{NItems: 10}, Selected: 0}},
		term.K(ui.Up),
		true,
		State{Items: TestItems{NItems: 10}, Selected: 0},
	},
	{
		"up moving to last item when selecting after boundary",
		&Widget{State: State{Items: TestItems{NItems: 10}, Selected: 11}},
		term.K(ui.Up),
		true,
		State{Items: TestItems{NItems: 10}, Selected: 9},
	},
	{
		"down moving selection down",
		&Widget{State: State{Items: TestItems{NItems: 10}, Selected: 1}},
		term.K(ui.Down),
		true,
		State{Items: TestItems{NItems: 10}, Selected: 2},
	},
	{
		"down stopping at n-1",
		&Widget{State: State{Items: TestItems{NItems: 10}, Selected: 9}},
		term.K(ui.Down),
		true,
		State{Items: TestItems{NItems: 10}, Selected: 9},
	},
	{
		"down moving to first item when selecting before boundary",
		&Widget{State: State{Items: TestItems{NItems: 10}, Selected: -2}},
		term.K(ui.Down),
		true,
		State{Items: TestItems{NItems: 10}, Selected: 0},
	},
	{
		"enter triggering default no-op accept",
		&Widget{State: State{Items: TestItems{NItems: 10}, Selected: 5}},
		term.K(ui.Enter),
		true,
		State{Items: TestItems{NItems: 10}, Selected: 5},
	},
	{
		"other keys not handled",
		&Widget{State: State{Items: TestItems{NItems: 10}, Selected: 5}},
		term.K('a'),
		false,
		State{Items: TestItems{NItems: 10}, Selected: 5},
	},
	{
		"overlay handler",
		addOverlay(
			&Widget{State: State{Items: TestItems{NItems: 10}, Selected: 5}},
			func(w *Widget) el.Handler {
				return el.MapHandler{
					term.K('a'): func() { w.State.Selected = 0 },
				}
			}),
		term.K('a'),
		true,
		State{Items: TestItems{NItems: 10}, Selected: 0},
	},
}

func addOverlay(w *Widget, overlay func(*Widget) el.Handler) *Widget {
	w.OverlayHandler = overlay(w)
	return w
}

func TestHandle(t *testing.T) {
	for _, test := range handleTests {
		t.Run(test.name, func(t *testing.T) {
			w := test.widget
			handled := w.Handle(test.event)
			if handled != test.wantHandled {
				t.Errorf("got handled %v, want %v", handled, test.wantHandled)
			}
			if !reflect.DeepEqual(w.State, test.wantState) {
				t.Errorf("got state %v, want %v", w.State, test.wantState)
			}
		})
	}
}

func TestHandle_EnterEmitsAccept(t *testing.T) {
	var acceptedItems Items
	var acceptedIndex int
	w := &Widget{
		State: State{Items: TestItems{NItems: 10}, Selected: 5},
		OnAccept: func(it Items, i int) {
			acceptedItems = it
			acceptedIndex = i
		},
	}
	w.Handle(term.K(ui.Enter))

	if acceptedItems != (TestItems{NItems: 10}) {
		t.Errorf("OnAccept not passed current Items")
	}
	if acceptedIndex != 5 {
		t.Errorf("OnAccept not passed current selected index")
	}
}

func TestSelect_ChangeState(t *testing.T) {
	// number of items = 10
	var tests = []struct {
		name   string
		before int
		f      func(selected, n int) int
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

		{"Prev from -1", -1, Prev, 0},
		{"Prev from 0", 0, Prev, 0},
		{"Prev from 9", 9, Prev, 8},
		{"Prev from 10", 10, Prev, 9},

		{"PrevWrap from -1", -1, PrevWrap, 9},
		{"PrevWrap from 0", 0, PrevWrap, 9},
		{"PrevWrap from 9", 9, PrevWrap, 8},
		{"PrevWrap from 10", 10, PrevWrap, 9},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := &Widget{
				State: State{Items: TestItems{NItems: 10}, Selected: test.before},
			}
			w.Select(test.f)
			if w.State.Selected != test.after {
				t.Errorf("selected = %d, want %d", w.State.Selected, test.after)
			}
		})
	}
}

func TestSelect_CallOnSelect(t *testing.T) {
	it := TestItems{NItems: 10}
	gotItemsCh := make(chan Items, 1)
	gotSelectedCh := make(chan int, 1)
	w := &Widget{
		State: State{Items: it, Selected: 5},
		OnSelect: func(it Items, i int) {
			gotItemsCh <- it
			gotSelectedCh <- i
		},
	}
	w.Select(Next)
	if gotItems := <-gotItemsCh; gotItems != it {
		t.Errorf("Got it = %v, want %v", gotItems, it)
	}
	if gotSelected := <-gotSelectedCh; gotSelected != 6 {
		t.Errorf("Got selected = %v, want 6", gotSelected)
	}
}
