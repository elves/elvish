package textview

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var bb = ui.NewBufferBuilder

var renderTests = []el.RenderTest{
	{
		Name: "text fits entirely",
		Given: &Widget{State: State{
			Lines: []string{"line 1", "line 2", "line 3"}}},
		Width: 10, Height: 4,
		Want: bb(10).
			WritePlain("line 1").Newline().
			WritePlain("line 2").Newline().
			WritePlain("line 3").Buffer(),
	},
	{
		Name: "text cropped horizontally",
		Given: &Widget{State: State{
			Lines: []string{"a very long line"}}},
		Width: 10, Height: 4,
		Want: bb(10).
			WritePlain("a very lon").Buffer(),
	},
	{
		Name: "text cropped vertically",
		Given: &Widget{State: State{
			Lines: []string{"line 1", "line 2", "line 3"}}},
		Width: 10, Height: 2,
		Want: bb(10).
			WritePlain("line 1").Newline().
			WritePlain("line 2").Buffer(),
	},
	{
		Name: "text cropped vertically, with scrollbar",
		Given: &Widget{
			Scrollable: true,
			State: State{
				Lines: []string{"line 1", "line 2", "line 3", "line 4"}}},
		Width: 10, Height: 2,
		Want: bb(10).
			WritePlain("line 1   ").
			WriteStyled(styled.MakeText(" ", "inverse", "magenta")).Newline().
			WritePlain("line 2   ").
			WriteStyled(styled.MakeText("│", "magenta")).Buffer(),
	},
	{
		Name: "State.First adjusted to fit text",
		Given: &Widget{State: State{
			First: 2,
			Lines: []string{"line 1", "line 2", "line 3"}}},
		Width: 10, Height: 3,
		Want: bb(10).
			WritePlain("line 1").Newline().
			WritePlain("line 2").Newline().
			WritePlain("line 3").Buffer(),
	},
}

func TestRender(t *testing.T) {
	el.TestRender(t, renderTests)
}

var handleTests = []el.HandleTest{
	{
		Name: "up doing nothing when not scrollable",
		Given: &Widget{
			State: State{Lines: []string{"1", "2", "3", "4"}, First: 1}},
		Event: term.K(ui.Up),

		WantUnhandled: true,
	},
	{
		Name: "up moving window up when scrollable",
		Given: &Widget{
			Scrollable: true,
			State:      State{Lines: []string{"1", "2", "3", "4"}, First: 1}},
		Event: term.K(ui.Up),

		WantNewState: State{Lines: []string{"1", "2", "3", "4"}, First: 0},
	},
	{
		Name: "up doing nothing when already at top",
		Given: &Widget{
			Scrollable: true,
			State:      State{Lines: []string{"1", "2", "3", "4"}, First: 0}},
		Event: term.K(ui.Up),

		WantNewState: State{Lines: []string{"1", "2", "3", "4"}, First: 0},
	},
	{
		Name: "down moving window down when scrollable",
		Given: &Widget{
			Scrollable: true,
			State:      State{Lines: []string{"1", "2", "3", "4"}, First: 1}},
		Event: term.K(ui.Down),

		WantNewState: State{Lines: []string{"1", "2", "3", "4"}, First: 2},
	},
	{
		Name: "down doing nothing when already at bottom",
		Given: &Widget{
			Scrollable: true,
			State:      State{Lines: []string{"1", "2", "3", "4"}, First: 3}},
		Event: term.K(ui.Down),

		WantNewState: State{Lines: []string{"1", "2", "3", "4"}, First: 3},
	},
	{
		Name: "overlay",
		Given: &Widget{
			OverlayHandler: el.MapHandler{term.K('a'): func() {}},
		},
		Event: term.K('a'),

		WantNewState: State{},
	},
}

func TestHandle(t *testing.T) {
	el.TestHandle(t, handleTests)
}

func TestCopyState(t *testing.T) {
	state := State{Lines: []string{"a", "b", "c"}, First: 1}
	w := &Widget{State: state}
	copied := w.CopyState()
	if !reflect.DeepEqual(copied, state) {
		t.Errorf("Got copied state %v, want %v", copied, state)
	}
}
