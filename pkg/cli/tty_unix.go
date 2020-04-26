// +build !windows,!plan9,!js

package cli

import (
	"os"
	"os/signal"
	"syscall"
)

func (t *aTTY) NotifySignals() <-chan os.Signal {
	// This implicitly catches every signal regardless of whether it is
	// currently ignored.
	t.sigCh = make(chan os.Signal, sigsChanBufferSize)
	signal.Notify(t.sigCh)
	// TODO: Remove this if, and when, job control is implemented. This
	// handles the case of running an external command from an interactive
	// prompt.
	//
	// See https://github.com/elves/elvish/issues/988. See also setupShell()
	// in pkg/shell/shell.go.
	signal.Ignore(syscall.SIGTTIN, syscall.SIGTTOU, syscall.SIGTSTP)
	return t.sigCh
}
