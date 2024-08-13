package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/sys"
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
	// CloseReader releases resources allocated for reading terminal events.
	CloseReader()

	term.Writer

	// NotifySignals start relaying signals and returns a channel on which
	// signals are delivered.
	NotifySignals() <-chan os.Signal
	// StopSignals stops the relaying of signals. After this function returns,
	// the channel returned by NotifySignals will no longer deliver signals.
	StopSignals()

	// Size returns the height and width of the terminal.
	Size() (h, w int)
}

type aTTY struct {
	in, out *os.File
	r       term.Reader
	term.Writer
	sigCh chan os.Signal

	rawMutex sync.Mutex
	raw      int
}

// NewTTY returns a new TTY from input and output terminal files.
func NewTTY(in, out *os.File) TTY {
	return &aTTY{in: in, out: out, Writer: term.NewWriter(out)}
}

func (t *aTTY) Setup() (func(), error) {
	restore, err := term.SetupForTUI(t.in, t.out)
	return func() {
		err := restore()
		if err != nil {
			fmt.Println(t.out, "failed to restore terminal properties:", err)
		}
	}, err
}

func (t *aTTY) Size() (h, w int) {
	return sys.WinSize(t.out)
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

func (t *aTTY) CloseReader() {
	if t.r != nil {
		t.r.Close()
	}
	t.r = nil
}

func (t *aTTY) NotifySignals() <-chan os.Signal {
	t.sigCh = sys.NotifySignals()
	return t.sigCh
}

func (t *aTTY) StopSignals() {
	signal.Stop(t.sigCh)
	close(t.sigCh)
	t.sigCh = nil
}
