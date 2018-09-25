package newedit

import "github.com/elves/elvish/newedit/types"

// The interface this package uses to access core.Editor.
//
// TODO(xiaq): Maybe this should be in the types package.
type editor interface {
	notifier
	State() *types.State
}

// An editor implementation that does nothing. Useful in tests.
type dummyEditor struct{ dummyNotifier }

func (dummyEditor) State() *types.State { return &types.State{} }

// An editor implementation wrapping fakeNotifier. Useful in tests.
type fakeEditor struct {
	fakeNotifier
	state types.State
}

func (ed *fakeEditor) State() *types.State { return &ed.state }

// Wraps the Notify method.
type notifier interface {
	Notify(string)
}

// A notifier implementation that does nothing. Useful in tests.
type dummyNotifier struct{}

func (dummyNotifier) Notify(_ string) {}

// A notifier implementation that records all Notify calls it has seen. Useful
// in tests.
type fakeNotifier struct{ notes []string }

func (n *fakeNotifier) Notify(note string) { n.notes = append(n.notes, note) }
