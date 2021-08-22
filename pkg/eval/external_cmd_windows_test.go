//go:build windows
// +build windows

package eval_test

import (
	"syscall"
)

func exitWaitStatus(exit uint32) syscall.WaitStatus {
	return syscall.WaitStatus{exit}
}
