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

// AddOverlayHandler returns a Widget the same as the given Widget, except that
// it always tries to handle events with the given Handler first.
func AddOverlayHandler(w Widget, h Handler) Widget {
	return widgetWithOverlayHandler{w, h}
}

type widgetWithOverlayHandler struct {
	base    Widget
	overlay Handler
}

func (w widgetWithOverlayHandler) Render(width, height int) *ui.Buffer {
	return w.base.Render(width, height)
}

func (w widgetWithOverlayHandler) Handle(event term.Event) bool {
	return w.overlay.Handle(event) || w.base.Handle(event)
}

// RendererTest is a test case to be used in TestRenderer.
type RendererTest struct {
	Name     string
	Renderer Renderer
	Width    int
	Height   int
	Want     interface{ Buffer() *ui.Buffer }
}

// TestRenderer runs the given Renderer tests.
func TestRenderer(t *testing.T, tests []RendererTest) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			buf := test.Renderer.Render(test.Width, test.Height)
			wantBuf := test.Want.Buffer()
			if !reflect.DeepEqual(buf, wantBuf) {
				t.Errorf("got buf %v, want %v", buf, wantBuf)
			}
		})
	}
}
