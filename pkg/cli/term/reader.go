package term

import (
	"errors"
	"fmt"
	"os"
)

// Reader reads events from the terminal.
type Reader interface {
	// ReadEvent reads a single event from the terminal.
	ReadEvent() (Event, error)
	// ReadRawEvent reads a single raw event from the terminal. The concept of
	// raw events is applicable where terminal events are represented as escape
	// sequences sequences, such as most modern Unix terminal emulators. If
	// the concept is not applicable, such as in the Windows console, it is
	// equivalent to ReadEvent.
	ReadRawEvent() (Event, error)
	// Close releases resources associated with the Reader. Any outstanding
	// ReadEvent or ReadRawEvent call will be aborted, returning ErrStopped.
	Close()
}

// ErrStopped is returned by Reader when Close is called during a ReadEvent or
// ReadRawEvent method.
var ErrStopped = errors.New("stopped")

var errTimeout = errors.New("timed out")

type seqError struct {
	msg string
	seq string
}

func (err seqError) Error() string {
	return fmt.Sprintf("%s: %q", err.msg, err.seq)
}

// NewReader creates a new Reader on the given terminal file.
//
// TODO: NewReader should return an error as well. Right now failure to
// initialize Reader panics.
func NewReader(f *os.File) Reader {
	return newReader(f)
}

// IsReadErrorRecoverable returns whether an error returned by Reader is
// recoverable.
func IsReadErrorRecoverable(err error) bool {
	if _, ok := err.(seqError); ok {
		return true
	}
	return err == ErrStopped || err == errTimeout
}
