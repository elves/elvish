//go:build windows

package ewindows

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// https://docs.microsoft.com/en-us/windows/console/readconsoleinput
//
// BOOL WINAPI ReadConsoleInput(
//
//		_In_  HANDLE        hConsoleInput,
//		_Out_ PINPUT_RECORD lpBuffer,
//		_In_  DWORD         nLength,
//		_Out_ LPDWORD       lpNumberOfEventsRead
//	  );
var readConsoleInput = kernel32.NewProc("ReadConsoleInputW")

// ReadConsoleInput input wraps the homonymous Windows API call.
func ReadConsoleInput(h windows.Handle, buf []InputRecord) (int, error) {
	var nr uintptr
	r, _, err := readConsoleInput.Call(uintptr(h),
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), uintptr(unsafe.Pointer(&nr)))
	if r != 0 {
		err = nil
	}
	return int(nr), err
}

// InputEvent is either a KeyEvent, MouseEvent, WindowBufferSizeEvent,
// MenuEvent or FocusEvent.
type InputEvent interface {
	isInputEvent()
}

func (*KeyEvent) isInputEvent()              {}
func (*MouseEvent) isInputEvent()            {}
func (*WindowBufferSizeEvent) isInputEvent() {}
func (*MenuEvent) isInputEvent()             {}
func (*FocusEvent) isInputEvent()            {}

// GetEvent converts InputRecord to InputEvent.
func (input *InputRecord) GetEvent() InputEvent {
	switch input.EventType {
	case KEY_EVENT:
		return (*KeyEvent)(unsafe.Pointer(&input.Event))
	case MOUSE_EVENT:
		return (*MouseEvent)(unsafe.Pointer(&input.Event))
	case WINDOW_BUFFER_SIZE_EVENT:
		return (*WindowBufferSizeEvent)(unsafe.Pointer(&input.Event))
	case MENU_EVENT:
		return (*MenuEvent)(unsafe.Pointer(&input.Event))
	case FOCUS_EVENT:
		return (*FocusEvent)(unsafe.Pointer(&input.Event))
	default:
		return nil
	}
}
