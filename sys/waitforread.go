package sys

import "os"

// WaitForRead blocks until any of the given files is ready to be read. It
// returns a boolean array indicating which files are ready to be read and
// possible errors.
//
// It is implemented with select(2) on Unix and WaitForMultipleObjects on
// Windows.
func WaitForRead(files ...*os.File) (ready []bool, err error) {
	return waitForRead(files...)
}
