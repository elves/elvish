package core

import (
	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
)

const (
	// Maximum number of events fakeTTY produces.
	maxEvents = 1024
)

// An implementation of the TTY interface.
type fakeTTY struct {
	// Predefined sizes.
	h, w int
	// Predefined events.
	events []tty.Event

	// Records buffer history.
	bufs []*ui.Buffer
	// Records SetRaw calls.
	setRaws []bool

	// Channel returned from StartRead. Can be used to inject additional events.
	eventCh chan tty.Event
}

func newFakeTTY(h, w int, events []tty.Event) *fakeTTY {
	return &fakeTTY{
		h, w, events,
		nil, nil,
		make(chan tty.Event, maxEvents)}
}

func (t *fakeTTY) Setup() (func(), error) { return func() {}, nil }

func (t *fakeTTY) Size() (h, w int) { return t.h, t.w }

func (t *fakeTTY) StartRead() <-chan tty.Event {
	for _, event := range t.events {
		t.eventCh <- event
	}
	return t.eventCh
}

func (t *fakeTTY) SetRaw(b bool) { t.setRaws = append(t.setRaws, b) }

func (t *fakeTTY) StopRead() {}

func (t *fakeTTY) Buffer() *ui.Buffer { return t.bufs[len(t.bufs)-1] }

func (t *fakeTTY) ResetBuffer() { t.bufs = append(t.bufs, nil) }

func (t *fakeTTY) UpdateBuffer(_, buf *ui.Buffer, _ bool) error {
	t.bufs = append(t.bufs, buf)
	return nil
}

// An implementation of Mode. The HandleKey method returns CommitCode after a
// certain number of key events and keeps the key event history.
type fakeMode struct {
	nkeys int

	keys []ui.Key
}

func newFakeMode(n int) *fakeMode {
	return &fakeMode{n, nil}
}

func (m *fakeMode) ModeLine() ui.Renderer          { return nil }
func (m *fakeMode) ModeRenderFlag() ModeRenderFlag { return 0 }

func (m *fakeMode) HandleKey(k ui.Key, st *State) HandlerAction {
	m.keys = append(m.keys, k)
	if len(m.keys) == m.nkeys {
		return CommitCode
	}
	return 0
}
