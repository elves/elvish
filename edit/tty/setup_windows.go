package tty

import (
	"os"

	"golang.org/x/sys/windows"
)

const (
	wantedInMode = windows.ENABLE_WINDOW_INPUT |
		windows.ENABLE_MOUSE_INPUT | windows.ENABLE_PROCESSED_INPUT
	wantedOutMode = windows.ENABLE_PROCESSED_OUTPUT |
		windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
)

func setup(in, out *os.File) (func() error, error) {
	hIn := windows.Handle(in.Fd())
	hOut := windows.Handle(out.Fd())

	var oldInMode, oldOutMode uint32
	err := windows.GetConsoleMode(hIn, &oldInMode)
	if err != nil {
		return nil, err
	}
	err = windows.GetConsoleMode(hOut, &oldOutMode)
	if err != nil {
		return nil, err
	}

	err = windows.SetConsoleMode(hIn, wantedInMode)
	if err != nil {
		return nil, err
	}
	err = windows.SetConsoleMode(hOut, wantedOutMode)
	if err != nil {
		windows.SetConsoleMode(hIn, oldInMode)
		return nil, err
	}

	return func() error {
		err1 := windows.SetConsoleMode(hIn, oldInMode)
		err2 := windows.SetConsoleMode(hOut, oldOutMode)
		if err1 != nil {
			// TODO when err2 != nil, return both
			return err1
		}
		return err2
	}, nil
}
