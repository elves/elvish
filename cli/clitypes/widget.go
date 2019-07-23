package clitypes

import (
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

// Widget is the basic component of UI; it knows how to handle events and how to
// render itself.
type Widget interface {
	ui.Renderer
	Handler
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

func (w widgetWithOverlayHandler) Render(bb *ui.BufferBuilder) {
	w.base.Render(bb)
}

func (w widgetWithOverlayHandler) Handle(event term.Event) bool {
	return w.overlay.Handle(event) || w.base.Handle(event)
}
