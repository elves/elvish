// +build !windows,!plan9

package sys

import (
	"os"
)

// IsATTY returns true if the given file is a terminal.
func IsATTY(file *os.File) bool {
	_, err := TermiosForFd(int(file.Fd()))
	return err == nil
}
