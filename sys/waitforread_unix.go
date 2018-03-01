// +build !windows,!plan9

package sys

import "os"

// WaitForRead blocks until any of the given files is ready to be read. It
// returns a boolean array indicating which files are ready to be read and
// possible errors.
//
// It is implemented with select(2) on Unix and WaitForMultipleObjects on
// Windows.
func WaitForRead(files ...*os.File) (ready []bool, err error) {
	maxfd := 0
	fdset := NewFdSet()
	for _, file := range files {
		fd := int(file.Fd())
		if maxfd < fd {
			maxfd = fd
		}
		fdset.Set(fd)
	}
	err = Select(maxfd+1, fdset, nil, nil)
	ready = make([]bool, len(files))
	for i, file := range files {
		ready[i] = fdset.IsSet(int(file.Fd()))
	}
	return ready, err
}
