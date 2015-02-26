package sys

import "C"
import (
	"os"
	"syscall"
)

// Ioctl wraps the ioctl syscall.
func Ioctl(fd int, req int, arg uintptr) error {
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL, uintptr(fd), uintptr(req), arg)
	if e != 0 {
		return os.NewSyscallError("ioctl", e)
	}
	return nil
}
