package core

import (
	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
)

const (
	// Maximum number of buffer updates fakeTTY expect to see.
	maxBufferUpdates = 1024
	// Maximum number of events fakeTTY produces.
	maxEvents = 1024
)

// An implementation of the TTY interface.
type fakeTTY struct {
	// Predefined sizes.
	h, w int

	// Channel returned from StartRead. Can be used to inject additional events.
	eventCh chan tty.Event

	// Channel for publishing buffer updates.
	bufCh chan *ui.Buffer
	// Records buffer history.
	bufs []*ui.Buffer
}

func newFakeTTY() *fakeTTY {
	return newFakeTTYWithSize(24, 80)
}

func newFakeTTYWithSize(h, w int) *fakeTTY {
	return &fakeTTY{
		h, w,
		make(chan tty.Event, maxEvents),
		make(chan *ui.Buffer, maxBufferUpdates), nil,
	}
}

func (t *fakeTTY) Setup() (func(), error) { return func() {}, nil }

func (t *fakeTTY) Size() (h, w int) { return t.h, t.w }

func (t *fakeTTY) StartRead() <-chan tty.Event {
	return t.eventCh
}

func (t *fakeTTY) SetRaw(b bool) {}

func (t *fakeTTY) StopRead() { close(t.eventCh) }

func (t *fakeTTY) Newline() {}

func (t *fakeTTY) Buffer() *ui.Buffer { return t.bufs[len(t.bufs)-1] }

func (t *fakeTTY) ResetBuffer() { t.recordBuf(nil) }

func (t *fakeTTY) UpdateBuffer(_, buf *ui.Buffer, _ bool) error {
	t.recordBuf(buf)
	return nil
}

func (t *fakeTTY) recordBuf(buf *ui.Buffer) {
	t.bufs = append(t.bufs, buf)
	t.bufCh <- buf
}
