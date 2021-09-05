// Package tk is the toolkit for the cli package.
//
// This package defines three basic interfaces - Renderer, Handler and Widget -
// and numerous implementations of these interfaces.
package tk

import (
	"src.elv.sh/pkg/cli/term"
)

// Widget is the basic component of UI; it knows how to handle events and how to
// render itself.
type Widget interface {
	Renderer
	MaxHeighter
	Handler
}

// Renderer wraps the Render method.
type Renderer interface {
	// Render renders onto a region of bound width and height.
	Render(width, height int) *term.Buffer
}

// MaxHeighter wraps the MaxHeight method.
type MaxHeighter interface {
	// MaxHeight returns the maximum height needed when rendering onto a region
	// of bound width and height. The returned value may be larger than the
	// height argument.
	MaxHeight(width, height int) int
}

// Handler wraps the Handle method.
type Handler interface {
	// Try to handle a terminal event and returns whether the event has been
	// handled.
	Handle(event term.Event) bool
}

// Bindings is the interface for key bindings.
type Bindings interface {
	Handle(Widget, term.Event) bool
}

// DummyBindings is a trivial Bindings implementation.
type DummyBindings struct{}

// Handle always returns false.
func (DummyBindings) Handle(w Widget, event term.Event) bool {
	return false
}

// MapBindings is a map-backed Bindings implementation.
type MapBindings map[term.Event]func(Widget)

// Handle handles the event by calling the function corresponding to the event
// in the map. If there is no corresponding function, it returns false.
func (m MapBindings) Handle(w Widget, event term.Event) bool {
	fn, ok := m[event]
	if ok {
		fn(w)
	}
	return ok
}

// FuncBindings is a function-based Bindings implementation.
type FuncBindings func(Widget, term.Event) bool

// Handle handles the event by calling the function.
func (f FuncBindings) Handle(w Widget, event term.Event) bool {
	return f(w, event)
}
