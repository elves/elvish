package sys

import (
	"os"
	"syscall"
)

func waitForRead(files ...*os.File) (ready []bool, err error) {
	handles := make([]syscall.Handle, len(files))
	for i, file := range files {
		handles[i] = syscall.Handle(file.Fd())
	}
	readyHandle, err := WaitForMultipleObjects(&handles, false, INFINITE)
	if err != nil {
		return nil, err
	}
	ready = make([]bool, len(files))
	for i, handle := range handles {
		if readyHandle == handle {
			ready[i] = true
		}
	}
	return
}
