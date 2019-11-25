// Package el defines interfaces for CLI UI elements and related utilities.
package el

import (
	"github.com/elves/elvish/cli/term"
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
	Render(width, height int) *term.Buffer
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

// FuncHandler is a function-based implementation of Handler.
type FuncHandler func(term.Event) bool

// Handle handles the event by calling the function.
func (f FuncHandler) Handle(event term.Event) bool {
	return f(event)
}
