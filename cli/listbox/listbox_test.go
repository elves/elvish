package listbox

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var bb = ui.NewBufferBuilder

var renderTests = []clitypes.RenderTest{
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
}

func TestRender(t *testing.T) {
	clitypes.TestRender(t, renderTests)
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
		(&Widget{State: State{Items: TestItems{NItems: 10}, Selected: 5}}).
			AddOverlay(func(w *Widget) clitypes.Handler {
				return clitypes.MapHandler{
					term.K('a'): func() { w.State.Selected = 0 },
				}
			}),
		term.K('a'),
		true,
		State{Items: TestItems{NItems: 10}, Selected: 0},
	},
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
	var accepted int
	w := &Widget{
		State:    State{Items: TestItems{NItems: 10}, Selected: 5},
		OnAccept: func(i int) { accepted = i },
	}
	w.Handle(term.K(ui.Enter))
	if accepted != 5 {
		t.Errorf("item 5 not accepted")
	}
}
