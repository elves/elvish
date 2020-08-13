// +build windows

package eval

import (
	"syscall"
)

func exitWaitStatus(exit uint32) syscall.WaitStatus {
	return syscall.WaitStatus{exit}
}
