package core

import (
	"reflect"
	"sync"
	"time"

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
	sizeMutex sync.RWMutex
	// Predefined sizes.
	height, width int
	// Callback to be returned from Setup.
	restoreFunc func()
	// Error to be returned from Setup.
	setupErr error

	// Channel returned from StartRead. Can be used to inject additional events.
	eventCh chan tty.Event

	// Channel for publishing updates of the main buffer and notes buffer.
	bufCh, notesBufCh chan *ui.Buffer
	// Records history of the main buffer and notes buffer.
	bufs, notesBufs []*ui.Buffer
}

func newFakeTTY() *fakeTTY {
	return &fakeTTY{
		height: 24, width: 80,
		restoreFunc: func() {},
		eventCh:     make(chan tty.Event, maxEvents),
		bufCh:       make(chan *ui.Buffer, maxBufferUpdates),
		notesBufCh:  make(chan *ui.Buffer, maxBufferUpdates),
	}
}

func (t *fakeTTY) Setup() (func(), error) {
	return t.restoreFunc, t.setupErr
}

func (t *fakeTTY) Size() (h, w int) {
	t.sizeMutex.RLock()
	defer t.sizeMutex.RUnlock()
	return t.height, t.width
}

func (t *fakeTTY) setSize(h, w int) {
	t.sizeMutex.Lock()
	defer t.sizeMutex.Unlock()
	t.height, t.width = h, w
}

func (t *fakeTTY) StartInput() <-chan tty.Event {
	return t.eventCh
}

func (t *fakeTTY) SetRawInput(b bool) {}

func (t *fakeTTY) StopInput() { close(t.eventCh) }

func (t *fakeTTY) Newline() {}

func (t *fakeTTY) Buffer() *ui.Buffer { return t.bufs[len(t.bufs)-1] }

func (t *fakeTTY) ResetBuffer() { t.recordBuf(nil) }

func (t *fakeTTY) UpdateBuffer(bufNotes, buf *ui.Buffer, _ bool) error {
	t.recordNotesBuf(bufNotes)
	t.recordBuf(buf)
	return nil
}

func (t *fakeTTY) recordBuf(buf *ui.Buffer) {
	t.bufs = append(t.bufs, buf)
	t.bufCh <- buf
}

func (t *fakeTTY) recordNotesBuf(buf *ui.Buffer) {
	t.notesBufs = append(t.notesBufs, buf)
	t.notesBufCh <- buf
}

var checkBufferTimeout = time.Second

// Check that an expected buffer will eventually appear. Also useful for waiting
// until the editor reaches a certain state.
func checkBuffer(want *ui.Buffer, ch <-chan *ui.Buffer) bool {
	for {
		select {
		case buf := <-ch:
			if reflect.DeepEqual(buf, want) {
				return true
			}
		case <-time.After(checkBufferTimeout):
			return false
		}
	}
}
