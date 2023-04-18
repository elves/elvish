//go:build unix

package os

import (
	"errors"

	"golang.org/x/sys/unix"
)

// isDirNotEmpty returns a bool that indicates whether the error corresponds to a
// platform specific syscall error that indicates a directory is not empty.
func isDirNotEmpty(err error) bool {
	return errors.Is(err, unix.ENOTEMPTY)
}
