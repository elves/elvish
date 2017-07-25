// +build windows

package sys

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	INFINITE uint32 = 0xFFFFFFFF
)

const (
	MAXIMUM_WAIT_OBJECTS int = 64
)

const (
	WAIT_OBJECT_0    uint32 = 0
	WAIT_ABANDONED_0 uint32 = 0x00000080
	WAIT_TIMEOUT     uint32 = 0x00000102
	WAIT_FAILED      uint32 = 0xFFFFFFFF
)

var (
	kernel32, _               = syscall.LoadLibrary("kernel32.dll")
	waitForMultipleObjects, _ = syscall.GetProcAddress(kernel32, "WaitForMultipleObjects")
)

func boolToUint32(b bool) uint32 {
	if b {
		return 1
	} else {
		return 0
	}
}

// DWORD WINAPI WaitForMultipleObjects(
//   _In_       DWORD  nCount,
//   _In_ const HANDLE *lpsyscall.Handles,
//   _In_       BOOL   bWaitAll,
//   _In_       DWORD  dwMilliseconds
// );
func WaitForMultipleObjects(handles *[]syscall.Handle, waitAll bool, milliseconds uint32) (handle syscall.Handle, err error) {
	count := uint32(len(*handles))
	r1, _, e1 := syscall.Syscall6(waitForMultipleObjects, 4, uintptr(count), uintptr(unsafe.Pointer(handles)), uintptr(boolToUint32(waitAll)), uintptr(milliseconds), 0, 0)
	if e1 != 0 {
		return syscall.Handle(syscall.InvalidHandle), e1
	}
	ret := uint32(r1)
	if WAIT_OBJECT_0 <= ret && ret <= WAIT_OBJECT_0+count-1 {
		return syscall.Handle((*handles)[ret-WAIT_OBJECT_0]), nil
	}
	return syscall.Handle(syscall.InvalidHandle), errors.New("timeout")
}

type Poller struct {
	rfds []syscall.Handle
	wfds []syscall.Handle
}

// TODO better option is IOCP
func (poller *Poller) Init(rfds []uintptr, wfds []uintptr) error {
	for _, rfd := range rfds {
		poller.rfds = append(poller.rfds, syscall.Handle(rfd))
	}
	for _, wfd := range wfds {
		poller.wfds = append(poller.wfds, syscall.Handle(wfd))
	}
	return nil
}

// WaitForMultipleObjects is O(n) and limited by MAXIMUM_WAIT_OBJECTS
func (poller *Poller) Poll(timeout *syscall.Timeval) (*[]uintptr, *[]uintptr, error) {
	var milliseconds uint32
	if timeout == nil || uint64(timeout.Sec*1000+timeout.Usec/1000) > uint64(INFINITE) {
		milliseconds = INFINITE
	} else {
		milliseconds = uint32(timeout.Sec*1000 + timeout.Usec/1000)
	}
	handles := append(poller.rfds, poller.wfds...)
	handle, err := WaitForMultipleObjects(&handles, false, milliseconds)
	if err != nil {
		return nil, nil, err
	}
	// this method is not truely trust, fd may be in both set
	rfds := []uintptr{}
	wfds := []uintptr{}
	for _, rfd := range poller.rfds {
		if rfd == handle {
			rfds = append(rfds, uintptr(handle))
			break
		}
	}
	for _, wfd := range poller.wfds {
		if wfd == handle {
			wfds = append(wfds, uintptr(handle))
			break
		}
	}
	return &rfds, &wfds, nil
}
