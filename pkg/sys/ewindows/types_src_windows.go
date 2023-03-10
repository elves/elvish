//go:build ignore

package ewindows

/*
#include <windows.h>
*/
import "C"

type (
	Coord       C.COORD
	InputRecord C.INPUT_RECORD

	KeyEvent              C.KEY_EVENT_RECORD
	MouseEvent            C.MOUSE_EVENT_RECORD
	WindowBufferSizeEvent C.WINDOW_BUFFER_SIZE_RECORD
	MenuEvent             C.MENU_EVENT_RECORD
	FocusEvent            C.FOCUS_EVENT_RECORD
)

const (
	KEY_EVENT                = C.KEY_EVENT
	MOUSE_EVENT              = C.MOUSE_EVENT
	WINDOW_BUFFER_SIZE_EVENT = C.WINDOW_BUFFER_SIZE_EVENT
	MENU_EVENT               = C.MENU_EVENT
	FOCUS_EVENT              = C.FOCUS_EVENT
)
