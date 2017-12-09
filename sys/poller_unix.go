// +build !windows

package sys

import "syscall"

type Poller struct {
	rdset *FdSet
	wdset *FdSet
	maxfd int
	rfds  []int
	wfds  []int
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func (poller *Poller) Init(rfds []uintptr, wfds []uintptr) error {
	poller.rdset = NewFdSet()
	poller.wdset = NewFdSet()
	poller.maxfd = -1
	for _, rfd := range rfds {
		poller.rfds = append(poller.rfds, int(rfd))
		poller.maxfd = max(poller.maxfd, int(rfd))
	}
	for _, wfd := range wfds {
		poller.wfds = append(poller.wfds, int(wfd))
		poller.maxfd = max(poller.maxfd, int(wfd))
	}
	return nil
}

func (poller *Poller) Poll(timeout *syscall.Timeval) (*[]uintptr, *[]uintptr, error) {
	poller.rdset.Set(poller.rfds...)
	poller.wdset.Set(poller.wfds...)
	err := Select(poller.maxfd+1, poller.rdset, poller.wdset, nil, timeout)
	if err != nil {
		return nil, nil, err
	}
	rfds := []uintptr{}
	wfds := []uintptr{}
	for _, rfd := range poller.rfds {
		if poller.rdset.IsSet(rfd) {
			rfds = append(rfds, uintptr(rfd))
		}
	}
	for _, wfd := range poller.wfds {
		if poller.rdset.IsSet(wfd) {
			wfds = append(wfds, uintptr(wfd))
		}
	}
	return &rfds, &wfds, nil
}
