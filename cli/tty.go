package cli

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/sys"
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

	// StartInput starts the delivery of terminal events and returns a channel
	// on which events are made available.
	StartInput() <-chan term.Event
	// SetRawInput requests the next n underlying events to be read uninterpreted. It
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
}

const sigsChanBufferSize = 256

// NewTTY returns a new TTY from input and output terminal files.
func NewTTY(in, out *os.File) TTY {
	return &aTTY{in, out, nil, term.NewWriter(out), nil}
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

func (t *aTTY) StartInput() <-chan term.Event {
	t.r = term.NewReader(t.in)
	t.r.Start()
	return t.r.EventChan()
}

func (t *aTTY) SetRawInput(n int) {
	t.r.SetRaw(n)
}

func (t *aTTY) StopInput() {
	t.r.Stop()
	t.r.Close()
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
