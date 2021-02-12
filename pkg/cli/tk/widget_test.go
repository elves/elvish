package tk

import (
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

type testWidget struct {
	// Text to render.
	text ui.Text
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

func TestDummyBindings(t *testing.T) {
	w := Empty{}
	b := DummyBindings{}
	for _, event := range []term.Event{term.K('a'), term.PasteSetting(true)} {
		if b.Handle(w, event) {
			t.Errorf("should not handle")
		}
	}
}

func TestMapBindings(t *testing.T) {
	widgetCh := make(chan Widget, 1)
	w := Empty{}
	b := MapBindings{term.K('a'): func(w Widget) { widgetCh <- w }}
	handled := b.Handle(w, term.K('a'))
	if !handled {
		t.Errorf("should handle")
	}
	if gotWidget := <-widgetCh; gotWidget != w {
		t.Errorf("function called with widget %v, want %v", gotWidget, w)
	}
	handled = b.Handle(w, term.K('b'))
	if handled {
		t.Errorf("should not handle")
	}
}

func TestFuncBindings(t *testing.T) {
	widgetCh := make(chan Widget, 1)
	eventCh := make(chan term.Event, 1)

	h := FuncBindings(func(w Widget, event term.Event) bool {
		widgetCh <- w
		eventCh <- event
		return event == term.K('a')
	})

	w := Empty{}
	event := term.K('a')
	handled := h.Handle(w, event)
	if !handled {
		t.Errorf("should handle")
	}
	if gotWidget := <-widgetCh; gotWidget != w {
		t.Errorf("function called with widget %v, want %v", gotWidget, w)
	}
	if gotEvent := <-eventCh; gotEvent != event {
		t.Errorf("function called with event %v, want %v", gotEvent, event)
	}

	event = term.K('b')
	handled = h.Handle(w, event)
	if handled {
		t.Errorf("should not handle")
	}
	if gotEvent := <-eventCh; gotEvent != event {
		t.Errorf("function called with event %v, want %v", gotEvent, event)
	}
}
