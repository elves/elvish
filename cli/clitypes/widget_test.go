package clitypes

import (
	"reflect"
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

func TestAddOverlayHandler(t *testing.T) {
	base := testWidget{
		text:     styled.Plain("base"),
		accepted: []term.Event{term.K('a'), term.K('b')},
	}
	overlay := testWidget{
		text:     styled.Plain("overlay"),
		accepted: []term.Event{term.K(ui.Up), term.K(ui.Down)},
	}
	w := AddOverlayHandler(&base, &overlay)

	buf := w.Render(10, 10)
	wantBuf := ui.NewBufferBuilder(10).WritePlain("base").Buffer()
	if !reflect.DeepEqual(buf, wantBuf) {
		t.Errorf("should render like base")
	}

	if !w.Handle(term.K('a')) || base.handled[0] != term.K('a') {
		t.Errorf("base did not handle")
	}
	if !w.Handle(term.K(ui.Up)) || overlay.handled[0] != term.K(ui.Up) {
		t.Errorf("overlay did not handle")
	}
	if w.Handle(term.K(ui.PageUp)) {
		t.Errorf("should not handle")
	}
}

func TestTestRender(t *testing.T) {
	TestRender(t, []RenderTest{{
		"a", &testWidget{text: styled.Plain("test")},
		10, 10,
		ui.NewBufferBuilder(10).WritePlain("test"),
	}})
	// Unable to test the failure branch as that will make the test fail, and
	// *testing.T cannot be constructed externally
}
