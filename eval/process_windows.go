package eval

import "syscall"

// Nop on Windows.
func putSelfInFg() error { return nil }

// The bitmask for CreationFlags in SysProcAttr to start a process in background.
const detachedProcess = 0x00000008

func makeSysProcAttr(bg bool) *syscall.SysProcAttr {
	flags := uint32(0)
	if bg {
		flags |= detachedProcess
	}
	return &syscall.SysProcAttr{CreationFlags: flags}
}
