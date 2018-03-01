package eval

import "errors"

// Interrupts returns a channel that is closed when an interrupt signal comes.
func (fm *Frame) Interrupts() <-chan struct{} {
	return fm.intCh
}

var ErrInterrupted = errors.New("interrupted")

// IsInterrupted reports whether there has been an interrupt.
func (fm *Frame) IsInterrupted() bool {
	select {
	case <-fm.Interrupts():
		return true
	default:
		return false
	}
}
