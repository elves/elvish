package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/sys"
)

// TTY is the type the terminal dependency of the editor needs to satisfy.
type TTY interface {
	// Setup sets up the terminal for the CLI app.
	//
	// This method returns a restore function that undoes the setup, and any
	// error during setup. It only returns fatal errors that make the terminal
	// unsuitable for later operations; non-fatal errors may be reported by
	// showing a warning message, but not returned.
	//
	// This method should be called before any other method is called.
	Setup() (restore func(), err error)

	// ReadEvent reads a terminal event.
	ReadEvent() (term.Event, error)
	// SetRawInput requests the next n ReadEvent calls to read raw events. It
	// is applicable to environments where events are represented as a special
	// sequences, such as VT100. It is a no-op if events are delivered as whole
	// units by the terminal, such as Windows consoles.
	SetRawInput(n int)

	// StopInput causes input delivery to be stopped. When this function
	// returns, the channel previously returned by StartInput will no longer
	// deliver input events.
	StopInput()

	// NotifySignals start relaying signals and returns a channel on which
	// signals are delivered.
	NotifySignals() <-chan os.Signal
	// StopSignals stops the relaying of signals. After this function returns,
	// the channel returned by NotifySignals will no longer deliver signals.
	StopSignals()

	// Size returns the height and width of the terminal.
	Size() (h, w int)

	// Buffer returns the current buffer. The initial value of the current
	// buffer is nil.
	Buffer() *term.Buffer
	// ResetBuffer resets the current buffer to nil without actuating any redraw.
	ResetBuffer()
	// UpdateBuffer updates the current buffer and draw it to the terminal.
	UpdateBuffer(bufNotes, bufMain *term.Buffer, full bool) error
}

// StdTTY is the terminal connected to inputs from stdin and output to stderr.
var StdTTY = NewTTY(os.Stdin, os.Stderr)

type aTTY struct {
	in, out *os.File
	r       term.Reader
	w       term.Writer
	sigCh   chan os.Signal

	rawMutex sync.Mutex
	raw      int
}

const sigsChanBufferSize = 256

// NewTTY returns a new TTY from input and output terminal files.
func NewTTY(in, out *os.File) TTY {
	return &aTTY{in: in, out: out, w: term.NewWriter(out)}
}

func (t *aTTY) Setup() (func(), error) {
	restore, err := term.Setup(t.in, t.out)
	return func() {
		err := restore()
		if err != nil {
			fmt.Println(t.out, "failed to restore terminal properties:", err)
		}
	}, err
}

func (t *aTTY) Size() (h, w int) {
	return sys.GetWinsize(t.out)
}

func (t *aTTY) ReadEvent() (term.Event, error) {
	if t.r == nil {
		t.r = term.NewReader(t.in)
	}
	if t.consumeRaw() {
		return t.r.ReadRawEvent()
	}
	return t.r.ReadEvent()
}

func (t *aTTY) consumeRaw() bool {
	t.rawMutex.Lock()
	defer t.rawMutex.Unlock()
	if t.raw <= 0 {
		return false
	}
	t.raw--
	return true
}

func (t *aTTY) SetRawInput(n int) {
	t.rawMutex.Lock()
	defer t.rawMutex.Unlock()
	t.raw = n
}

func (t *aTTY) StopInput() {
	if t.r != nil {
		t.r.Close()
	}
	t.r = nil
}

func (t *aTTY) Buffer() *term.Buffer {
	return t.w.CurrentBuffer()
}

func (t *aTTY) ResetBuffer() {
	t.w.ResetCurrentBuffer()
}

func (t *aTTY) UpdateBuffer(bufNotes, bufMain *term.Buffer, full bool) error {
	return t.w.CommitBuffer(bufNotes, bufMain, full)
}

func (t *aTTY) NotifySignals() <-chan os.Signal {
	t.sigCh = make(chan os.Signal, sigsChanBufferSize)
	signal.Notify(t.sigCh)
	return t.sigCh
}

func (t *aTTY) StopSignals() {
	signal.Stop(t.sigCh)
	close(t.sigCh)
	t.sigCh = nil
}
