package cliedit

// The editor's dependencies are not concrete implementations, but interfaces.
// This file defines the interfaces for those dependencies as well as fake
// implementations that are useful in tests.

// An interface that wraps Notify. It is part of, and smaller than the full app
// interface. Internal functions that do not need to access the app state can
// use this interface to make it easier to test.
type notifier interface {
	Notify(string)
}

// A notifier implementation that records all Notify calls it has seen. Useful
// in tests.
type fakeNotifier struct{ notes []string }

func (n *fakeNotifier) Notify(note string) { n.notes = append(n.notes, note) }
