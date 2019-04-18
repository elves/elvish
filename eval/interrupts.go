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

// Starts to listen to terminal interrupts. Returns a channel that is closed
// when a SIGINT or SIGQUIT has been received, and a cleanup function that
// should be called to stop listening and clean up the resource.
func listenInterrupts() (chan struct{}, func()) {
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
