// +build windows

package sys

import (
	"github.com/mattn/go-isatty"
)

// IsATTY returns true if the given file descriptor is a terminal.
func IsATTY(fd int) bool {
	return isatty.IsTerminal(uintptr(fd)) ||
		isatty.IsCygwinTerminal(uintptr(fd))
}
