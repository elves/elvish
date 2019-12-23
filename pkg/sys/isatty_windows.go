// +build windows

package sys

import (
	"os"

	"github.com/mattn/go-isatty"
)

// IsATTY returns true if the given file descriptor is a terminal.
func IsATTY(file *os.File) bool {
	fd := uintptr(file.Fd())
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
