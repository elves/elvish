package eval

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// ErrInterrupted is thrown when the execution is interrupted by a signal.
var ErrInterrupted = errors.New("interrupted")

// ListenInterrupts returns a Context that is canceled when SIGINT or SIGQUIT
// has been received by the process. It also returns a function to cancel the
// Context, which should be called when it is no longer needed.
func ListenInterrupts() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(sigCh)
	}()

	return ctx, func() {
		cancel()
	}
}
