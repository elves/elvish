package el

import (
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
)

// Unable to test the failure branches, as we cannot construct a valid
// *testing.T instance from outside the testing package :(

func TestTestRender(t *testing.T) {
	TestRender(t, []RenderTest{
		{
			Name:  "test",
			Given: &testWidget{text: ui.NewText("test")},
			Width: 10, Height: 10,

			Want: term.NewBufferBuilder(10).Write("test"),
		},
	})
}

type testHandlerWithState struct {
	State testHandlerState
}

type testHandlerState struct {
	last  term.Event
	total int
}

func (h *testHandlerWithState) Handle(e term.Event) bool {
	if e == term.K('x') {
		return false
	}
	h.State.last = e
	h.State.total++
	return true
}

func TestTestHandle(t *testing.T) {
	TestHandle(t, []HandleTest{
		{
			Name:  "WantNewState",
			Given: &testHandlerWithState{},
			Event: term.K('a'),

			WantNewState: testHandlerState{last: term.K('a'), total: 1},
		},
		{
			Name:   "Multiple events",
			Given:  &testHandlerWithState{},
			Events: []term.Event{term.K('a'), term.K('b')},

			WantNewState: testHandlerState{last: term.K('b'), total: 2},
		},
		{
			Name:  "WantUnhaneld",
			Given: &testHandlerWithState{},
			Event: term.K('x'),

			WantUnhandled: true,
		},
	})
}
