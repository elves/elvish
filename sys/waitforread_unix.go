// +build !windows,!plan9

package sys

import "os"

func waitForRead(files ...*os.File) (ready []bool, err error) {
	maxfd := 0
	fdset := NewFdSet()
	for _, file := range files {
		fd := int(file.Fd())
		if maxfd < fd {
			maxfd = fd
		}
		fdset.Set(fd)
	}
	err = Select(maxfd+1, fdset, nil, nil, nil)
	ready = make([]bool, len(files))
	for i, file := range files {
		ready[i] = fdset.IsSet(int(file.Fd()))
	}
	return ready, err
}
