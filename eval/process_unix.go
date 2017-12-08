// +build !windows,!plan9

package eval

import (
	"os/signal"
	"syscall"

	"github.com/elves/elvish/sys"
)

// Process control functions in Unix.

func ignoreTTOU() {
	signal.Ignore(syscall.SIGTTOU)
}

func unignoreTTOU() {
	signal.Reset(syscall.SIGTTOU)
}

func putSelfInFg() error {
	return sys.Tcsetpgrp(0, syscall.Getpgrp())
}

func makeSysProcAttr(bg bool) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: bg}
}
