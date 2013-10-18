package util

import (
	"os"
	"time"
	"errors"
	"unsafe"
	"syscall"
)

var (
	Timeout = errors.New("timed out")
	FdTooBig = errors.New("fd exceeds FD_SETSIZE")
)

type TimedReader struct {
	File *os.File
	Timeout time.Duration
	nfds int
	set syscall.FdSet
}

func NewTimedReader(f *os.File) (*TimedReader, error) {
	fd := f.Fd()
	if fd >= syscall.FD_SETSIZE {
		return nil, FdTooBig
	}
	tr := &TimedReader{File: f, Timeout: -1, nfds: int(fd) + 1}
	bitLength := unsafe.Sizeof(tr.set.Bits[0]) * 8
	tr.set.Bits[fd / bitLength] |= 1 << (fd % bitLength)
	return tr, nil
}

func (tr *TimedReader) Read(p []byte) (n int, err error) {
	if tr.Timeout < 0 {
		// Timeout is turned off
		return tr.File.Read(p)
	}

	tv := syscall.NsecToTimeval(int64(tr.Timeout))
	set := tr.set // Make a copy since syscall.Select will modify it

	nfd, err := syscall.Select(tr.nfds, &set, nil, nil, &tv)
	if err != nil {
		return 0, err
	}
	if nfd == 0 {
		return 0, Timeout
	}
	return tr.File.Read(p)
}
