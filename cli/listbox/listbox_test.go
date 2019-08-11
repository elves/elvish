package listbox

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var bb = ui.NewBufferBuilder

type itemer struct{ prefix string }

func (it itemer) Item(i int) styled.Text {
	prefix := it.prefix
	if prefix == "" {
		prefix = "item "
	}
	return styled.Plain(fmt.Sprintf("%s%d", prefix, i))
}

var renderTests = []struct {
	name    string
	widget  *Widget
	width   int
	height  int
	wantBuf *ui.BufferBuilder
}{
	{
		"placeholder when Itemer is nil",
		&Widget{Placeholder: styled.Plain("nothing")},
		10, 3,
		bb(10).WritePlain("nothing"),
	},
	{
		"placeholder when NItems is 0",
		&Widget{
			Placeholder: styled.Plain("nothing"),
			State:       State{Itemer: itemer{}},
		},
		10, 3,
		bb(10).WritePlain("nothing"),
	},
	{
		"all items when there is enough height",
		&Widget{State: State{Itemer: itemer{}, NItems: 2, Selected: 0}},
		10, 3,
		bb(10).
			WriteStyled(styled.MakeText("item 0    ", "inverse")).
			Newline().WritePlain("item 1"),
	},
	{
		"long lines cropped",
		&Widget{State: State{Itemer: itemer{}, NItems: 2, Selected: 0}},
		4, 3,
		bb(4).
			WriteStyled(styled.MakeText("item", "inverse")).
			Newline().WritePlain("item"),
	},
	{
		"scrollbar when not showing all items",
		&Widget{State: State{Itemer: itemer{}, NItems: 4, Selected: 0}},
		10, 2,
		bb(10).
			WriteStyled(styled.MakeText("item 0   ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")).
			Newline().WritePlain("item 1   ").
			WriteStyled(styled.MakeText("│", "magenta")),
	},
	{
		"scrollbar when not showing last item in full",
		&Widget{State: State{Itemer: itemer{"item\n"}, NItems: 2, Selected: 0}},
		10, 3,
		bb(10).
			WriteStyled(styled.MakeText("item     ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")).
			Newline().WriteStyled(styled.MakeText("0        ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")).
			Newline().WritePlain("item     ").
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")),
	},
	{
		"scrollbar when not showing only item in full",
		&Widget{State: State{Itemer: itemer{"item\n"}, NItems: 1, Selected: 0}},
		10, 1,
		bb(10).
			WriteStyled(styled.MakeText("item     ", "inverse")).
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")),
	},
}

func TestRender(t *testing.T) {
	for _, test := range renderTests {
		t.Run(test.name, func(t *testing.T) {
			buf := test.widget.Render(test.width, test.height)
			wantBuf := test.wantBuf.Buffer()
			if !reflect.DeepEqual(buf, wantBuf) {
				t.Errorf("got buf %v, want %v", buf, wantBuf)
			}
		})
	}
}

var handleTests = []struct {
	name      string
	widget    *Widget
	events    []term.Event
	wantState State
}{
	{
		"up moving selection up",
		&Widget{State: State{Itemer: itemer{}, NItems: 10, Selected: 1}},
		[]term.Event{term.K(ui.Up)},
		State{Itemer: itemer{}, NItems: 10, Selected: 0},
	},
	{
		"up stopping at 0",
		&Widget{State: State{Itemer: itemer{}, NItems: 10, Selected: 0}},
		[]term.Event{term.K(ui.Up)},
		State{Itemer: itemer{}, NItems: 10, Selected: 0},
	},
	{
		"up moving to last item when selecting after boundary",
		&Widget{State: State{Itemer: itemer{}, NItems: 10, Selected: 11}},
		[]term.Event{term.K(ui.Up)},
		State{Itemer: itemer{}, NItems: 10, Selected: 9},
	},
	{
		"down moving selection down",
		&Widget{State: State{Itemer: itemer{}, NItems: 10, Selected: 1}},
		[]term.Event{term.K(ui.Down)},
		State{Itemer: itemer{}, NItems: 10, Selected: 2},
	},
	{
		"down stopping at n-1",
		&Widget{State: State{Itemer: itemer{}, NItems: 10, Selected: 9}},
		[]term.Event{term.K(ui.Down)},
		State{Itemer: itemer{}, NItems: 10, Selected: 9},
	},
	{
		"down moving to first item when selecting before boundary",
		&Widget{State: State{Itemer: itemer{}, NItems: 10, Selected: -2}},
		[]term.Event{term.K(ui.Down)},
		State{Itemer: itemer{}, NItems: 10, Selected: 0},
	},
	{
		"default no-op accept",
		&Widget{State: State{Itemer: itemer{}, NItems: 10, Selected: 5}},
		[]term.Event{term.K(ui.Enter)},
		State{Itemer: itemer{}, NItems: 10, Selected: 5},
	},
}

func TestHandle(t *testing.T) {
	for _, test := range handleTests {
		t.Run(test.name, func(t *testing.T) {
			w := test.widget
			for _, event := range test.events {
				w.Handle(event)
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
		State:    State{Itemer: itemer{}, NItems: 10, Selected: 5},
		OnAccept: func(i int) { accepted = i },
	}
	w.Handle(term.K(ui.Enter))
	if accepted != 5 {
		t.Errorf("item 5 not accepted")
	}
}
