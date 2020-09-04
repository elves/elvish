// +build !windows,!plan9,!js

package eval_test

import (
	"syscall"
)

func exitWaitStatus(exit uint32) syscall.WaitStatus {
	// The exit<<8 is gross but I can't find any exported symbols that would
	// allow us to construct WaitStatus. So assume legacy UNIX encoding
	// for a process that exits normally; i.e., not due to a signal.
	return syscall.WaitStatus(exit << 8)
}
