package clicore

import (
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

const (
	// Maximum number of buffer updates FakeTTY expect to see.
	maxBufferUpdates = 1024
	// Maximum number of events FakeTTY produces.
	maxEvents = 1024
	// Maximum number of signals FakeTTY produces.
	maxSignals = 1024
)

// An implementation of the TTY interface that is useful in tests.
type fakeTTY struct {
	setup func() (func(), error)
	// Channel that StartRead returns. Can be used to inject additional events.
	eventCh chan term.Event
	// Channel for publishing updates of the main buffer and notes buffer.
	bufCh, notesBufCh chan *ui.Buffer
	// Records history of the main buffer and notes buffer.
	bufs, notesBufs []*ui.Buffer
	// Channel that NotifySignals returns. Can be used to inject signals.
	sigCh chan os.Signal

	sizeMutex sync.RWMutex
	// Predefined sizes.
	height, width int
}

// NewFakeTTY creates a new FakeTTY and a handle for controlling it.
func NewFakeTTY() (TTY, TTYCtrl) {
	tty := &fakeTTY{
		eventCh:    make(chan term.Event, maxEvents),
		sigCh:      make(chan os.Signal, maxSignals),
		bufCh:      make(chan *ui.Buffer, maxBufferUpdates),
		notesBufCh: make(chan *ui.Buffer, maxBufferUpdates),
		height:     24, width: 80,
	}
	return tty, TTYCtrl{tty}
}

// Delegates to the setup function specified using the SetSetup method of
// TTYCtrl, or return a nop function and a nil error.
func (t *fakeTTY) Setup() (func(), error) {
	if t.setup == nil {
		return func() {}, nil
	}
	return t.setup()
}

// Returns the size specified by using the SetSize method of TTYCtrl.
func (t *fakeTTY) Size() (h, w int) {
	t.sizeMutex.RLock()
	defer t.sizeMutex.RUnlock()
	return t.height, t.width
}

// Returns t.eventCh. Events may be injected onto the channel by using the
// TTYCtrl.
func (t *fakeTTY) StartInput() <-chan term.Event {
	return t.eventCh
}

// Nop.
func (t *fakeTTY) SetRawInput(b bool) {}

// Nop.
func (t *fakeTTY) StopInput() { close(t.eventCh) }

// Nop.
func (t *fakeTTY) Newline() {}

// Returns the last recorded buffer.
func (t *fakeTTY) Buffer() *ui.Buffer { return t.bufs[len(t.bufs)-1] }

// Records a nil buffer.
func (t *fakeTTY) ResetBuffer() { t.recordBuf(nil) }

// UpdateBuffer records a new pair of buffers, i.e. sending them to their
// respective channels and appending them to their respective slices.
func (t *fakeTTY) UpdateBuffer(bufNotes, buf *ui.Buffer, _ bool) error {
	t.recordNotesBuf(bufNotes)
	t.recordBuf(buf)
	return nil
}

func (t *fakeTTY) NotifySignals() <-chan os.Signal { return t.sigCh }

func (t *fakeTTY) StopSignals() { close(t.sigCh) }

func (t *fakeTTY) recordBuf(buf *ui.Buffer) {
	t.bufs = append(t.bufs, buf)
	t.bufCh <- buf
}

func (t *fakeTTY) recordNotesBuf(buf *ui.Buffer) {
	t.notesBufs = append(t.notesBufs, buf)
	t.notesBufCh <- buf
}

// TTYCtrl is an interface for controlling a fake terminal.
type TTYCtrl struct{ *fakeTTY }

// SetSetup changes the return values of the Setup method of the fake terminal.
func (t TTYCtrl) SetSetup(restore func(), err error) {
	t.setup = func() (func(), error) {
		return restore, err
	}
}

// SetSize sets the size of the fake terminal.
func (t TTYCtrl) SetSize(h, w int) {
	t.sizeMutex.Lock()
	defer t.sizeMutex.Unlock()
	t.height, t.width = h, w
}

// Inject injects an event to the fake terminal.
func (t TTYCtrl) Inject(e term.Event) { t.eventCh <- e }

// VerifyBuffer verifies that a buffer will appear within the timeout of 1
// second.
func (t TTYCtrl) VerifyBuffer(b *ui.Buffer) bool {
	return verifyBuffer(b, t.bufCh)
}

// VerifyNotesBuffer verifies the a notes buffer will appear within the timeout
// of 1 second.
func (t TTYCtrl) VerifyNotesBuffer(b *ui.Buffer) bool {
	return verifyBuffer(b, t.notesBufCh)
}

// BufferHistory returns a slice of all buffers that have appeared.
func (t TTYCtrl) BufferHistory() []*ui.Buffer { return t.bufs }

// NotesBufferHistory returns a slice of all notes buffers that have appeared.
func (t TTYCtrl) NotesBufferHistory() []*ui.Buffer { return t.notesBufs }

// InjectSignal injects a signal.
func (t TTYCtrl) InjectSignal(sig os.Signal) { t.sigCh <- sig }

var verifyBufferTimeout = time.Second

// Check that an expected buffer will eventually appear. Also useful for waiting
// until the editor reaches a certain state.
func verifyBuffer(want *ui.Buffer, ch <-chan *ui.Buffer) bool {
	for {
		select {
		case buf := <-ch:
			if reflect.DeepEqual(buf, want) {
				return true
			}
		case <-time.After(verifyBufferTimeout):
			return false
		}
	}
}
