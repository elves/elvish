package newedit

import "github.com/elves/elvish/newedit/types"

// The editor's dependencies are not concrete implementations, but interfaces.
// This file defines the interfaces for those dependencies as well as fake
// implementations that are useful in tests.

// Interface for the clicore.Editor dependency.
type editor interface {
	notifier
	State() *types.State
}

// An interface that wraps Notify. It is part of, and smaller than the full
// editor interface. Internal functions that do not need to access the editor
// state can use this interface to make it easier to test.
type notifier interface {
	Notify(string)
}

// An editor implementation that records all Notify calls it has seen, and whose
// state is just a field. Useful in tests.
type fakeEditor struct {
	fakeNotifier
	state types.State
}

func (ed *fakeEditor) State() *types.State { return &ed.state }

// A notifier implementation that records all Notify calls it has seen. Useful
// in tests.
type fakeNotifier struct{ notes []string }

func (n *fakeNotifier) Notify(note string) { n.notes = append(n.notes, note) }
