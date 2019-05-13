package clicore

import (
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
)

// FakeTTY is an implementation of the TTY interface that is useful in tests.
type FakeTTY struct {
	// Callback to be returned from Setup.
	RestoreFunc func()
	// Error to be returned from Setup.
	SetupErr error

	// Channel returned from StartRead. Can be used to inject additional events.
	EventCh chan term.Event

	// Channel for publishing updates of the main buffer and notes buffer.
	BufCh, NotesBufCh chan *ui.Buffer
	// Records history of the main buffer and notes buffer.
	Bufs, NotesBufs []*ui.Buffer

	sizeMutex sync.RWMutex
	// Predefined sizes.
	height, width int
}

// NewFakeTTY creates a new FakeTTY.
func NewFakeTTY() *FakeTTY {
	return &FakeTTY{
		RestoreFunc: func() {},
		EventCh:     make(chan term.Event, maxEvents),
		BufCh:       make(chan *ui.Buffer, maxBufferUpdates),
		NotesBufCh:  make(chan *ui.Buffer, maxBufferUpdates),
		height:      24, width: 80,
	}
}

// Setup returns t.RestoreFunc and t.SetupErr.
func (t *FakeTTY) Setup() (func(), error) {
	return t.RestoreFunc, t.SetupErr
}

// Size returns the size previously set by SetSize.
func (t *FakeTTY) Size() (h, w int) {
	t.sizeMutex.RLock()
	defer t.sizeMutex.RUnlock()
	return t.height, t.width
}

// SetSize sets the size that will be returned by Size.
func (t *FakeTTY) SetSize(h, w int) {
	t.sizeMutex.Lock()
	defer t.sizeMutex.Unlock()
	t.height, t.width = h, w
}

// StartInput returns t.EventCh.
func (t *FakeTTY) StartInput() <-chan term.Event {
	return t.EventCh
}

// SetRawInput does nothing.
func (t *FakeTTY) SetRawInput(b bool) {}

// StopInput closes t.EventCh
func (t *FakeTTY) StopInput() { close(t.EventCh) }

// Newline does nothing.
func (t *FakeTTY) Newline() {}

// Buffer returns the last recorded buffer.
func (t *FakeTTY) Buffer() *ui.Buffer { return t.Bufs[len(t.Bufs)-1] }

// ResetBuffer records a nil buffer.
func (t *FakeTTY) ResetBuffer() { t.recordBuf(nil) }

// UpdateBuffer records a new pair of buffers, i.e. sending them to their
// respective channels and appending them to their respective slices.
func (t *FakeTTY) UpdateBuffer(bufNotes, buf *ui.Buffer, _ bool) error {
	t.recordNotesBuf(bufNotes)
	t.recordBuf(buf)
	return nil
}

func (t *FakeTTY) recordBuf(buf *ui.Buffer) {
	t.Bufs = append(t.Bufs, buf)
	t.BufCh <- buf
}

func (t *FakeTTY) recordNotesBuf(buf *ui.Buffer) {
	t.NotesBufs = append(t.NotesBufs, buf)
	t.NotesBufCh <- buf
}

// VerifyBuffer verifies that a buffer will appear within one second.
func (t *FakeTTY) VerifyBuffer(b *ui.Buffer) bool {
	return verifyBuffer(b, t.BufCh)
}

// VerifyNotesBuffer verifies that a notes buffer will appear within one second.
func (t *FakeTTY) VerifyNotesBuffer(b *ui.Buffer) bool {
	return verifyBuffer(b, t.NotesBufCh)
}

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
