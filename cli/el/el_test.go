package el

import (
	"testing"

	"github.com/elves/elvish/cli/term"
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

func (w *testWidget) Render(width, height int) *term.Buffer {
	buf := term.NewBufferBuilder(width).WriteStyled(w.text).Buffer()
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

func TestFuncHandler(t *testing.T) {
	eventCh := make(chan term.Event, 1)
	h := FuncHandler(func(event term.Event) bool {
		eventCh <- event
		return event == term.K('a')
	})

	handled := h.Handle(term.K('a'))
	if !handled {
		t.Errorf("should handle")
	}
	if <-eventCh != term.K('a') {
		t.Errorf("should call func")
	}

	handled = h.Handle(term.K('b'))
	if handled {
		t.Errorf("should not handle")
	}
	if <-eventCh != term.K('b') {
		t.Errorf("should call func")
	}
}
