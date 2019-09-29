package el

import (
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

type testWidget struct {
	// Text to render.
	text styled.Text
	// Which events to accept.
	accepted []term.Event
	// A record of events that have been handled.
	handled []term.Event
}

func (w *testWidget) Render(width, height int) *ui.Buffer {
	buf := ui.NewBufferBuilder(width).WriteStyled(w.text).Buffer()
	buf.TrimToLines(0, height)
	return buf
}

func (w *testWidget) Handle(e term.Event) bool {
	for _, accept := range w.accepted {
		if e == accept {
			w.handled = append(w.handled, e)
			return true
		}
	}
	return false
}

func TestDummyHandler(t *testing.T) {
	h := DummyHandler{}
	for _, event := range []term.Event{term.K('a'), term.PasteSetting(true)} {
		if h.Handle(event) {
			t.Errorf("should not handle")
		}
	}
}

func TestMapHandler(t *testing.T) {
	var aCalled bool
	h := MapHandler{term.K('a'): func() { aCalled = true }}
	handled := h.Handle(term.K('a'))
	if !handled {
		t.Errorf("should handle")
	}
	if !aCalled {
		t.Errorf("should call callback")
	}
	handled = h.Handle(term.K('b'))
	if handled {
		t.Errorf("should not handle")
	}
}

func TestTestRender(t *testing.T) {
	TestRender(t, []RenderTest{
		{
			Name:  "test",
			Given: &testWidget{text: styled.Plain("test")},
			Width: 10, Height: 10,

			Want: ui.NewBufferBuilder(10).WritePlain("test"),
		},
	})
	// Unable to test the failure branch as that will make the test fail, and
	// *testing.T cannot be constructed externally
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
			Name:  "WantUnhaneld",
			Given: &testHandlerWithState{},
			Event: term.K('x'),

			WantUnhandled: true,
		},
	})
}
