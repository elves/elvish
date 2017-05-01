package sys

import (
	"syscall"
	"unsafe"
)

func Tcgetpgrp(fd int) (int, error) {
	var pid int
	_, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(syscall.TIOCGPGRP), uintptr(unsafe.Pointer(&pid)))
	if errno == 0 {
		return pid, nil
	}
	return -1, errno
}

func Tcsetpgrp(fd int, pid int) error {
	_, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(syscall.TIOCSPGRP), uintptr(unsafe.Pointer(&pid)))
	if errno == 0 {
		return nil
	}
	return errno
}
