package core

import (
	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
)

// An implementation of tty.Reader. Replays a predefined stream of events, and
// records all SetRaw calls.
type fakeReader struct {
	events []tty.Event

	eventCh chan tty.Event
	stopCh  chan struct{}

	setRaws []bool
}

func newFakeReader(events ...tty.Event) *fakeReader {
	return &fakeReader{events, make(chan tty.Event), make(chan struct{}), nil}
}

func (r *fakeReader) SetRaw(b bool) { r.setRaws = append(r.setRaws, b) }

func (r *fakeReader) EventChan() <-chan tty.Event { return r.eventCh }

func (r *fakeReader) Start() {
	go r.run()
}

func (r *fakeReader) run() {
	for _, event := range r.events {
		select {
		case r.eventCh <- event:
		case <-r.stopCh:
			return
		}
	}
	<-r.stopCh
}

func (r *fakeReader) Stop() { r.stopCh <- struct{}{} }

func (r *fakeReader) Close() {}

// An implementation of tty.Writer. Keeps the buffer history.
type fakeWriter struct {
	bufs []*ui.Buffer
}

func newFakeWriter() *fakeWriter {
	return &fakeWriter{[]*ui.Buffer{nil}}
}

func (w *fakeWriter) CurrentBuffer() *ui.Buffer { return w.bufs[len(w.bufs)-1] }

func (w *fakeWriter) ResetCurrentBuffer() { w.bufs = append(w.bufs, nil) }

func (w *fakeWriter) CommitBuffer(_, buf *ui.Buffer, _ bool) error {
	w.bufs = append(w.bufs, buf)
	return nil
}

// An implementation of TTY.
type fakeTTY struct {
	h, w int
}

func newFakeTTY(h, w int) *fakeTTY {
	return &fakeTTY{h, w}
}

func (t *fakeTTY) Setup() (func(), error) { return func() {}, nil }

func (t *fakeTTY) Size() (h, w int) { return t.h, t.w }

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
