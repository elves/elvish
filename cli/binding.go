package cli

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/edit/ui"
)

// Binding represents key binding.
type Binding interface {
	// KeyHandler returns a KeyHandler for the given key.
	KeyHandler(ui.Key) KeyHandler
}

// KeyHandler is a function that can handle a key event.
type KeyHandler func(KeyEvent)

// KeyEvent is passed to a KeyHandler, containing information about the event
// and can be used for specifying actions.
type KeyEvent interface {
	// Key returns the key that triggered the KeyEvent.
	Key() ui.Key
	// State returns the State of the app.
	State() *clitypes.State
	// App returns the App.
	App() *App

	// CommitEOF specifies that the app should return from ReadCode with io.EOF
	// after the key handler returns.
	CommitEOF()
	// CommitCode specifies that the app should return from ReadCode after the
	// key handler returns.
	CommitCode()
}

// Internal implementation of KeyHandler interface.
type keyEvent struct {
	key        ui.Key
	app        *App
	commitEOF  bool
	commitLine bool
}

func (ev *keyEvent) Key() ui.Key            { return ev.key }
func (ev *keyEvent) App() *App              { return ev.app }
func (ev *keyEvent) State() *clitypes.State { return ev.app.core.State() }
func (ev *keyEvent) CommitEOF()             { ev.commitEOF = true }
func (ev *keyEvent) CommitCode()            { ev.commitLine = true }

// MapBinding builds a Binding from a map. The map may contain the special
// key ui.Default for a default KeyHandler.
func MapBinding(m map[ui.Key]KeyHandler) Binding {
	return mapBinding(m)
}

type mapBinding map[ui.Key]KeyHandler

func (b mapBinding) KeyHandler(k ui.Key) KeyHandler {
	handler, ok := b[k]
	if ok {
		return handler
	}
	return b[ui.Default]
}

func adaptBinding(b Binding, app *App) func(ui.Key) clitypes.HandlerAction {
	if b == nil {
		return nil
	}
	return func(k ui.Key) clitypes.HandlerAction {
		ev := &keyEvent{k, app, false, false}
		handler := b.KeyHandler(k)
		if handler == nil {
			return clitypes.NoAction
		}
		handler(ev)
		switch {
		case ev.commitEOF:
			return clitypes.CommitEOF
		case ev.commitLine:
			return clitypes.CommitCode
		default:
			return clitypes.NoAction
		}
	}
}
