// +build !windows,!plan9

package sys

import (
	"unsafe"
)

// IsATTY returns true if the given file descriptor is a terminal.
func IsATTY(fd int) bool {
	var term Termios
	err := Ioctl(fd, getAttrIOCTL, uintptr(unsafe.Pointer(&term)))
	return err == nil
}
