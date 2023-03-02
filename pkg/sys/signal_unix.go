//go:build unix

package sys

import (
	"os"
	"os/signal"
	"syscall"
)

func notifySignals() chan os.Signal {
	// This catches every signal regardless of whether it is ignored.
	sigCh := make(chan os.Signal, sigsChanBufferSize)
	signal.Notify(sigCh)
	// Calling signal.Notify will reset the signal ignore status, so we need to
	// call signal.Ignore every time we call signal.Notify.
	//
	// TODO: Remove this if, and when, job control is implemented. This
	// handles the case of running an external command from an interactive
	// prompt.
	//
	// See https://b.elv.sh/988.
	signal.Ignore(syscall.SIGTTIN, syscall.SIGTTOU, syscall.SIGTSTP)
	return sigCh
}
