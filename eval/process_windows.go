package eval

import "syscall"

// Process control functions in Windows. These are all NOPs.
func ignoreTTOU()        {}
func unignoreTTOU()      {}
func putSelfInFg() error { return nil }

const DETACHED_PROCESS = 0x00000008

func makeSysProcAttr(bg bool) *syscall.SysProcAttr {
	flags := uint32(0)
	if bg {
		flags |= DETACHED_PROCESS
	}
	return &syscall.SysProcAttr{CreationFlags: flags}
}
