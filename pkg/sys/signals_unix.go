// +build !windows,!plan9,!js

package sys

import (
	"os"
	"os/signal"
	"syscall"
)

func NotifySignals() chan os.Signal {
	// This catches every signal regardless of whether it is ignored.
	sigCh := make(chan os.Signal, sigsChanBufferSize)
	signal.Notify(sigCh)
	// TODO: Remove this if, and when, job control is implemented. This
	// handles the case of running an external command from an interactive
	// prompt.
	//
	// See https://github.com/elves/elvish/issues/988.
	signal.Ignore(syscall.SIGTTIN, syscall.SIGTTOU, syscall.SIGTSTP)
	return sigCh
}
