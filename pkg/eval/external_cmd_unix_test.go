//go:build unix

package eval_test

import (
	"syscall"
)

func exitWaitStatus(exit uint32) syscall.WaitStatus {
	// There doesn't seem to be a portable way to construct a WaitStatus, but
	// all Unix platforms Elvish supports have 0 in the lower 8 bits for normal
	// exits, and put exit codes in the higher bits.
	//
	// https://cs.opensource.google/go/go/+/refs/tags/go1.19.3:src/syscall/syscall_linux.go;l=404-411;drc=946b4baaf6521d521928500b2b57429c149854e7
	// https://cs.opensource.google/go/go/+/master:src/syscall/syscall_bsd.go;l=89-93;drc=51297dd6df713b988b5c587e448b27d18ca1bd8a
	return syscall.WaitStatus(exit << 8)
}
