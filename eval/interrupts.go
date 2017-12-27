package eval

import "errors"

// Interrupts returns a channel that is closed when an interrupt signal comes.
func (ec *Frame) Interrupts() <-chan struct{} {
	return ec.intCh
}

var ErrInterrupted = errors.New("interrupted")

// CheckInterrupts checks whether there has been an interrupt, and throws
// ErrInterrupted if that is the case
func (ec *Frame) CheckInterrupts() {
	select {
	case <-ec.Interrupts():
		throw(ErrInterrupted)
	default:
	}
}
