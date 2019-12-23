package term

import "os"

// Reader reads terminal events and makes them available on a channel.
type Reader interface {
	// Start starts the Reader.
	Start()
	// EventChan returns the channel onto which the Reader writes events that it
	// has read.
	EventChan() <-chan Event
	// SetRaw requests the next n underlying events to be read uninterpreted. It
	// is applicable to environments where events are represented as a special
	// sequences, such as VT100. It is a no-op if events are delivered as whole
	// units by the terminal, such as Windows consoles.
	SetRaw(n int)
	// Stop stops the Reader.
	Stop()
	// Close releases resources associated with the Reader. It does not affect
	// resources that were used to create it.
	Close()
}

// NewReader creates a new Reader on the given terminal file.
//
// TODO: NewReader should return an error as well. Right now failure to
// initialize Reader panics.
func NewReader(f *os.File) Reader {
	return newReader(f)
}
