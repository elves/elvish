// +build !windows,!plan9

package sys

import (
	"os"
	"time"
)

// WaitForRead blocks until any of the given files is ready to be read or
// timeout. A negative timeout means no timeout. It returns a boolean array
// indicating which files are ready to be read and any possible error.
func WaitForRead(timeout time.Duration, files ...*os.File) (ready []bool, err error) {
	maxfd := 0
	fdset := NewFdSet()
	for _, file := range files {
		fd := int(file.Fd())
		if maxfd < fd {
			maxfd = fd
		}
		fdset.Set(fd)
	}
	err = Select(maxfd+1, fdset, nil, nil, timeout)
	ready = make([]bool, len(files))
	for i, file := range files {
		ready[i] = fdset.IsSet(int(file.Fd()))
	}
	return ready, err
}
