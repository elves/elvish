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
	WAIT_OBJECT_0    = 0
	WAIT_ABANDONED_0 = 0x00000080
	WAIT_TIMEOUT     = 0x00000102
	WAIT_FAILED      = 0xFFFFFFFF
)

var (
	kernel32               = syscall.MustLoadDLL("kernel32.dll")
	waitForMultipleObjects = kernel32.MustFindProc("WaitForMultipleObjects")
	errTimeout             = errors.New("WaitForMultipleObjects timeout")
)

// DWORD WINAPI WaitForMultipleObjects(
//   _In_       DWORD  nCount,
//   _In_ const HANDLE *lpHandles,
//   _In_       BOOL   bWaitAll,
//   _In_       DWORD  dwMilliseconds
// );
func WaitForMultipleObjects(handles *[]syscall.Handle, waitAll bool, milliseconds uint32) (syscall.Handle, error) {
	count := len(*handles)
	ret, _, err := waitForMultipleObjects.Call(uintptr(count),
		uintptr(unsafe.Pointer(handles)), boolToUintptr(waitAll), uintptr(milliseconds))
	if err != nil {
		return syscall.InvalidHandle, err
	}
	if WAIT_OBJECT_0 <= ret && ret < WAIT_OBJECT_0+uintptr(count) {
		return (*handles)[ret-WAIT_OBJECT_0], nil
	}
	return syscall.InvalidHandle, errTimeout
}

func boolToUintptr(b bool) uintptr {
	if b {
		return 1
	} else {
		return 0
	}
}
