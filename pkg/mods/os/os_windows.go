//go:build windows

package os

import (
	"errors"

	"golang.org/x/sys/windows"
)

// isDirNotEmpty returns a bool that indicates whether the error corresponds to a
// platform specific syscall error that indicates a directory is not empty.
func isDirNotEmpty(err error) bool {
	return errors.Is(err, windows.ERROR_DIR_NOT_EMPTY)
}
