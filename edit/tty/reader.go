package tty

import "os"

// Reader reads terminal events and makes them available on a channel.
type Reader interface {
	// SetRaw turns the raw mode on or off. In raw mode, the Reader does not
	// decode special sequences, but simply deliver them as RawRune events. If
	// the Reader is in the middle of reading one event, it takes effect after
	// this event is fully read. On platforms (i.e. Windows) where events are
	// not encoded as special sequences, SetRaw has no effect.
	SetRaw(bool)
	// EventChan returns the channel onto which the Reader writes events that it
	// has read.
	EventChan() <-chan Event
	// Start starts the Reader.
	Start()
	// Stop stops the Reader.
	Stop()
	// Close releases resources associated with the Reader. It does not affect
	// resources used to create it.
	Close()
}

// NewReader creates a new Reader on the given terminal file.
// TODO: NewReader should return an error as well. Right now failure to
// initialize Reader panics.
func NewReader(f *os.File) Reader {
	return newReader(f)
}
