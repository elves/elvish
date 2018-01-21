package eval

import "errors"

// Interrupts returns a channel that is closed when an interrupt signal comes.
func (ec *Frame) Interrupts() <-chan struct{} {
	return ec.intCh
}

var ErrInterrupted = errors.New("interrupted")

// IsInterrupted reports whether there has been an interrupt.
func (ec *Frame) IsInterrupted() bool {
	select {
	case <-ec.Interrupts():
		return true
	default:
		return false
	}
}
