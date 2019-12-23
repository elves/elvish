// +build !windows,!plan9

package eval

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/elves/elvish/pkg/sys"
)

// Process control functions in Unix.

func putSelfInFg() error {
	if !sys.IsATTY(os.Stdin) {
		return nil
	}
	// If Elvish is in the background, the tcsetpgrp call below will either fail
	// (if the process is in an orphaned process group) or stop the process.
	// Ignoring TTOU fixes that.
	signal.Ignore(syscall.SIGTTOU)
	defer signal.Reset(syscall.SIGTTOU)
	return sys.Tcsetpgrp(0, syscall.Getpgrp())
}

func makeSysProcAttr(bg bool) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: bg}
}
