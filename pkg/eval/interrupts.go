package eval

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// Interrupts returns a channel that is closed when an interrupt signal comes.
func (fm *Frame) Interrupts() <-chan struct{} {
	return fm.intCh
}

// ErrInterrupted is thrown when the execution is interrupted by a signal.
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

// ListenInterrupts returns a channel that is closed when SIGINT or SIGQUIT
// has been received by the process. It also returns a function that should be
// called when the channel is no longer needed.
func ListenInterrupts() (<-chan struct{}, func()) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGQUIT)
	// Channel to return, closed after receiving the first SIGINT or SIGQUIT.
	intCh := make(chan struct{})

	// Closed in the cleanup function to request the relaying goroutine to stop.
	stop := make(chan struct{})
	// Closed in the relaying goroutine to signal that it has stopped.
	stopped := make(chan struct{})

	go func() {
		closed := false
	loop:
		for {
			select {
			case <-sigCh:
				if !closed {
					close(intCh)
					closed = true
				}
			case <-stop:
				break loop
			}
		}
		signal.Stop(sigCh)
		close(stopped)
	}()

	return intCh, func() {
		close(stop)
		<-stopped
	}
}
