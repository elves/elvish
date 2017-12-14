package tty

import (
	"os"

	"golang.org/x/sys/windows"
)

const wantedMode = windows.ENABLE_WINDOW_INPUT |
	windows.ENABLE_MOUSE_INPUT | windows.ENABLE_PROCESSED_INPUT

func setup(file *os.File) (func() error, error) {
	handle := windows.Handle(file.Fd())

	var oldMode uint32
	err := windows.GetConsoleMode(handle, &oldMode)
	if err != nil {
		return nil, err
	}
	err = windows.SetConsoleMode(handle, wantedMode)
	if err != nil {
		return nil, err
	}
	return func() error {
		return windows.SetConsoleMode(handle, oldMode)
	}, nil
}
