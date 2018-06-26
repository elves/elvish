package core

import (
	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
)

// An implementation of tty.Reader. Replays a predefined stream of events, and
// records all SetRaw calls.

type fakeTTY struct {
	h, w   int
	events []tty.Event

	eventCh chan tty.Event
	stopCh  chan struct{}

	bufs    []*ui.Buffer
	setRaws []bool
}

func newFakeTTY(h, t int, events []tty.Event) *fakeTTY {
	return &fakeTTY{
		h, t, events, make(chan tty.Event), make(chan struct{}), nil, nil}
}

func (t *fakeTTY) Setup() (func(), error) { return func() {}, nil }

func (t *fakeTTY) Size() (h, w int) { return t.h, t.w }

func (t *fakeTTY) StartRead() <-chan tty.Event {
	go t.run()
	return t.eventCh
}

func (t *fakeTTY) run() {
	for _, event := range t.events {
		select {
		case t.eventCh <- event:
		case <-t.stopCh:
			return
		}
	}
	<-t.stopCh
}

func (t *fakeTTY) SetRaw(b bool) { t.setRaws = append(t.setRaws, b) }

func (t *fakeTTY) StopRead() { t.stopCh <- struct{}{} }

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
