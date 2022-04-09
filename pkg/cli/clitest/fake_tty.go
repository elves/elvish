package clitest

import (
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/testutil"
)

const (
	// Maximum number of buffer updates FakeTTY expect to see.
	fakeTTYBufferUpdates = 4096
	// Maximum number of events FakeTTY produces.
	fakeTTYEvents = 4096
	// Maximum number of signals FakeTTY produces.
	fakeTTYSignals = 4096
)

// An implementation of the cli.TTY interface that is useful in tests.
type fakeTTY struct {
	setup func() (func(), error)
	// Channel that StartRead returns. Can be used to inject additional events.
	eventCh chan term.Event
	// Whether eventCh has been closed.
	eventChClosed bool
	// Mutex for synchronizing writing and closing eventCh.
	eventChMutex sync.Mutex
	// Channel for publishing updates of the main buffer and notes buffer.
	bufCh, notesBufCh chan *term.Buffer
	// Records history of the main buffer and notes buffer.
	bufs, notesBufs []*term.Buffer
	// Mutexes for guarding bufs and notesBufs.
	bufMutex sync.RWMutex
	// Channel that NotifySignals returns. Can be used to inject signals.
	sigCh chan os.Signal
	// Argument that SetRawInput got.
	raw int
	// Number of times the TTY screen has been cleared, incremented in
	// ClearScreen.
	cleared int

	sizeMutex sync.RWMutex
	// Predefined sizes.
	height, width int
}

// Initial size of fake TTY.
const (
	FakeTTYHeight = 20
	FakeTTYWidth  = 50
)

// NewFakeTTY creates a new FakeTTY and a handle for controlling it. The initial
// size of the terminal is FakeTTYHeight and FakeTTYWidth.
func NewFakeTTY() (cli.TTY, TTYCtrl) {
	tty := &fakeTTY{
		eventCh:    make(chan term.Event, fakeTTYEvents),
		sigCh:      make(chan os.Signal, fakeTTYSignals),
		bufCh:      make(chan *term.Buffer, fakeTTYBufferUpdates),
		notesBufCh: make(chan *term.Buffer, fakeTTYBufferUpdates),
		height:     FakeTTYHeight, width: FakeTTYWidth,
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

// Returns next event from t.eventCh.
func (t *fakeTTY) ReadEvent() (term.Event, error) {
	return <-t.eventCh, nil
}

// Records the argument.
func (t *fakeTTY) SetRawInput(n int) {
	t.raw = n
}

// Closes eventCh.
func (t *fakeTTY) CloseReader() {
	t.eventChMutex.Lock()
	defer t.eventChMutex.Unlock()
	close(t.eventCh)
	t.eventChClosed = true
}

// Returns the last recorded buffer.
func (t *fakeTTY) Buffer() *term.Buffer {
	t.bufMutex.RLock()
	defer t.bufMutex.RUnlock()
	return t.bufs[len(t.bufs)-1]
}

// Records a nil buffer.
func (t *fakeTTY) ResetBuffer() {
	t.bufMutex.Lock()
	defer t.bufMutex.Unlock()
	t.recordBuf(nil)
}

// UpdateBuffer records a new pair of buffers, i.e. sending them to their
// respective channels and appending them to their respective slices.
func (t *fakeTTY) UpdateBuffer(bufNotes, buf *term.Buffer, _ bool) error {
	t.bufMutex.Lock()
	defer t.bufMutex.Unlock()
	t.recordNotesBuf(bufNotes)
	t.recordBuf(buf)
	return nil
}

func (t *fakeTTY) HideCursor() {
}

func (t *fakeTTY) ShowCursor() {
}

func (t *fakeTTY) ClearScreen() {
	t.cleared++
}

func (t *fakeTTY) NotifySignals() <-chan os.Signal { return t.sigCh }

func (t *fakeTTY) StopSignals() { close(t.sigCh) }

func (t *fakeTTY) recordBuf(buf *term.Buffer) {
	t.bufs = append(t.bufs, buf)
	t.bufCh <- buf
}

func (t *fakeTTY) recordNotesBuf(buf *term.Buffer) {
	t.notesBufs = append(t.notesBufs, buf)
	t.notesBufCh <- buf
}

// TTYCtrl is an interface for controlling a fake terminal.
type TTYCtrl struct{ *fakeTTY }

// GetTTYCtrl takes a TTY and returns a TTYCtrl and true, if the TTY is a fake
// terminal. Otherwise it returns an invalid TTYCtrl and false.
func GetTTYCtrl(t cli.TTY) (TTYCtrl, bool) {
	fake, ok := t.(*fakeTTY)
	return TTYCtrl{fake}, ok
}

// SetSetup sets the return values of the Setup method of the fake terminal.
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

// Inject injects events to the fake terminal.
func (t TTYCtrl) Inject(events ...term.Event) {
	for _, event := range events {
		t.inject(event)
	}
}

func (t TTYCtrl) inject(event term.Event) {
	t.eventChMutex.Lock()
	defer t.eventChMutex.Unlock()
	if !t.eventChClosed {
		t.eventCh <- event
	}
}

// EventCh returns the underlying channel for delivering events.
func (t TTYCtrl) EventCh() chan term.Event {
	return t.eventCh
}

// InjectSignal injects signals.
func (t TTYCtrl) InjectSignal(sigs ...os.Signal) {
	for _, sig := range sigs {
		t.sigCh <- sig
	}
}

// RawInput returns the argument in the last call to the SetRawInput method of
// the TTY.
func (t TTYCtrl) RawInput() int {
	return t.raw
}

// ScreenCleared returns the number of times ClearScreen has been called on the
// TTY.
func (t TTYCtrl) ScreenCleared() int {
	return t.cleared
}

// TestBuffer verifies that a buffer will appear within 100ms, and aborts the
// test if it doesn't.
func (t TTYCtrl) TestBuffer(tt *testing.T, b *term.Buffer) {
	tt.Helper()
	ok := testBuffer(b, t.bufCh)
	if !ok {
		tt.Logf("wanted buffer not shown:\n%s", b.TTYString())

		t.bufMutex.RLock()
		defer t.bufMutex.RUnlock()
		lastBuf := t.LastBuffer()
		tt.Logf("Last buffer: %s", lastBuf.TTYString())
		if lastBuf == nil {
			bufs := t.BufferHistory()
			for i := len(bufs) - 1; i >= 0; i-- {
				if bufs[i] != nil {
					tt.Logf("Last non-nil buffer: %s", bufs[i].TTYString())
					return
				}
			}
		}
		tt.FailNow()
	}
}

// TestNotesBuffer verifies that a notes buffer will appear within 100ms, and
// aborts the test if it doesn't.
func (t TTYCtrl) TestNotesBuffer(tt *testing.T, b *term.Buffer) {
	tt.Helper()
	ok := testBuffer(b, t.notesBufCh)
	if !ok {
		tt.Logf("wanted notes buffer not shown:\n%s", b.TTYString())

		t.bufMutex.RLock()
		defer t.bufMutex.RUnlock()
		bufs := t.NotesBufferHistory()
		tt.Logf("There has been %d notes buffers. None-nil ones are:", len(bufs))
		for i, buf := range bufs {
			if buf != nil {
				tt.Logf("#%d:\n%s", i, buf.TTYString())
			}
		}
		tt.FailNow()
	}
}

// BufferHistory returns a slice of all buffers that have appeared.
func (t TTYCtrl) BufferHistory() []*term.Buffer {
	t.bufMutex.RLock()
	defer t.bufMutex.RUnlock()
	return t.bufs
}

// LastBuffer returns the last buffer that has appeared.
func (t TTYCtrl) LastBuffer() *term.Buffer {
	t.bufMutex.RLock()
	defer t.bufMutex.RUnlock()
	if len(t.bufs) == 0 {
		return nil
	}
	return t.bufs[len(t.bufs)-1]
}

// NotesBufferHistory returns a slice of all notes buffers that have appeared.
func (t TTYCtrl) NotesBufferHistory() []*term.Buffer {
	t.bufMutex.RLock()
	defer t.bufMutex.RUnlock()
	return t.notesBufs
}

func (t TTYCtrl) LastNotesBuffer() *term.Buffer {
	t.bufMutex.RLock()
	defer t.bufMutex.RUnlock()
	if len(t.notesBufs) == 0 {
		return nil
	}
	return t.notesBufs[len(t.notesBufs)-1]
}

// Tests that an buffer appears on the channel within 100ms.
func testBuffer(want *term.Buffer, ch <-chan *term.Buffer) bool {
	timeout := time.After(testutil.Scaled(100 * time.Millisecond))
	for {
		select {
		case buf := <-ch:
			if reflect.DeepEqual(buf, want) {
				return true
			}
		case <-timeout:
			return false
		}
	}
}
