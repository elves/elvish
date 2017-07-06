package eval

import "errors"

// Interrupts returns a channel that is closed when an interrupt signal comes.
func (ec *EvalCtx) Interrupts() <-chan struct{} {
	return ec.intCh
}

var ErrInterrupted = errors.New("interrupted")

// CheckInterrupts checks whether there has been an interrupt, and throws
// ErrInterrupted if that is the case
func (ec *EvalCtx) CheckInterrupts() {
	select {
	case <-ec.Interrupts():
		throw(ErrInterrupted)
	default:
	}
}
