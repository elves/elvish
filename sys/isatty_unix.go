// +build !windows,!plan9

package sys

import (
	"os"
	"unsafe"
)

// IsATTY returns true if the given file is a terminal.
func IsATTY(file *os.File) bool {
	var term Termios
	err := Ioctl(int(file.Fd()), getAttrIOCTL, uintptr(unsafe.Pointer(&term)))
	return err == nil
}
