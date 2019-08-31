package clitypes

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

// Widget is the basic component of UI; it knows how to handle events and how to
// render itself.
type Widget interface {
	Renderer
	Handler
}

// Renderer wraps the Render method.
type Renderer interface {
	// Render onto a region of bound width and height.
	Render(width, height int) *ui.Buffer
}

// Handler wraps the Handle method.
type Handler interface {
	// Try to handle a terminal event and returns whether the event has been
	// handled.
	Handle(event term.Event) bool
}

// DummyHandler is a trivial implementation of Handler.
type DummyHandler struct{}

// Handle always returns false.
func (DummyHandler) Handle(term.Event) bool { return false }

// MapHandler is a map-backed implementation of Handler.
type MapHandler map[term.Event]func()

// Handle handles the event by calling the function corresponding to the event
// in the map. If there is no corresponding function, it returns false.
func (m MapHandler) Handle(event term.Event) bool {
	fn, ok := m[event]
	if ok {
		fn()
	}
	return ok
}

// RenderTest is a test case to be used in TestRenderer.
type RenderTest struct {
	Name   string
	Given  Renderer
	Width  int
	Height int
	Want   interface{ Buffer() *ui.Buffer }
}

// TestRender runs the given Renderer tests.
func TestRender(t *testing.T, tests []RenderTest) {
	t.Helper()
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Helper()
			buf := test.Given.Render(test.Width, test.Height)
			wantBuf := test.Want.Buffer()
			if !reflect.DeepEqual(buf, wantBuf) {
				t.Errorf("Buffer mismatch")
				t.Logf("Got: %s", buf.TTYString())
				t.Logf("Want: %s", wantBuf.TTYString())
			}
		})
	}
}
