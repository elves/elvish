package edit

import "github.com/elves/elvish/cli"

// The editor's dependencies are not concrete implementations, but interfaces.
// This file defines the interfaces for those dependencies as well as fake
// implementations that are useful in tests.

// An interface that wraps Notify. It is part of, and smaller than the full app
// interface. Internal functions that do not need to access the app state can
// use this interface to make it easier to test.
type notifier interface {
	Notify(string)
}

// A notifier implementation the wraps an *App. This has to be a pointer to work
// around bootstrapping issues.
type appNotifier struct{ p *cli.App }

func (n appNotifier) Notify(note string) { (*n.p).Notify(note) }

// A notifier implementation that records all Notify calls it has seen. Useful
// in tests.
type fakeNotifier struct{ notes []string }

func (n *fakeNotifier) Notify(note string) { n.notes = append(n.notes, note) }
