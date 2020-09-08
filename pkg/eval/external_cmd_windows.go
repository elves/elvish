// +build windows

package eval

import (
	"syscall"
)

func ExitWaitStatus(exit uint32) syscall.WaitStatus {
	return syscall.WaitStatus{exit}
}
