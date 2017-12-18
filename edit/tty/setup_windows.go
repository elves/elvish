package tty

import (
	"os"

	"github.com/elves/elvish/util"
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

	errSetIn := windows.SetConsoleMode(hIn, wantedInMode)
	errSetOut := windows.SetConsoleMode(hOut, wantedOutMode)
	errVT := setupVT(out)

	return func() error {
		return util.Errors(
			windows.SetConsoleMode(hIn, oldInMode),
			windows.SetConsoleMode(hOut, oldOutMode),
			restoreVT(out))
	}, util.Errors(errSetIn, errSetOut, errVT)
}
