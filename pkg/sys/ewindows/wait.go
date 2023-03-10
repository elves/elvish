//go:build windows

package ewindows

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	INFINITE = 0xFFFFFFFF
)

const (
	WAIT_OBJECT_0    = 0
	WAIT_ABANDONED_0 = 0x00000080
	WAIT_TIMEOUT     = 0x00000102
	WAIT_FAILED      = 0xFFFFFFFF
)

var (
	waitForMultipleObjects = kernel32.NewProc("WaitForMultipleObjects")
	errTimeout             = errors.New("WaitForMultipleObjects timeout")
)

// WaitForMultipleObjects blocks until any of the objects is triggered or
// timeout.
//
// DWORD WINAPI WaitForMultipleObjects(
//
//	_In_       DWORD  nCount,
//	_In_ const HANDLE *lpHandles,
//	_In_       BOOL   bWaitAll,
//	_In_       DWORD  dwMilliseconds
//
// );
func WaitForMultipleObjects(handles []windows.Handle, waitAll bool,
	timeout uint32) (trigger int, abandoned bool, err error) {

	count := uintptr(len(handles))
	ret, _, err := waitForMultipleObjects.Call(count,
		uintptr(unsafe.Pointer(&handles[0])), boolToUintptr(waitAll), uintptr(timeout))
	switch {
	case WAIT_OBJECT_0 <= ret && ret < WAIT_OBJECT_0+count:
		return int(ret - WAIT_OBJECT_0), false, nil
	case WAIT_ABANDONED_0 <= ret && ret < WAIT_ABANDONED_0+count:
		return int(ret - WAIT_ABANDONED_0), true, nil
	case ret == WAIT_TIMEOUT:
		return -1, false, errTimeout
	default:
		return -1, false, err
	}
}

func boolToUintptr(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
