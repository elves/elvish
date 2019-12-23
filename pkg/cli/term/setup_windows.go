package term

import (
	"os"

	"github.com/elves/elvish/pkg/util"
	"golang.org/x/sys/windows"
)

const (
	wantedInMode = windows.ENABLE_WINDOW_INPUT |
		windows.ENABLE_MOUSE_INPUT | windows.ENABLE_PROCESSED_INPUT
	wantedOutMode = windows.ENABLE_PROCESSED_OUTPUT |
		windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	wantedGlobalOutMode = windows.ENABLE_PROCESSED_OUTPUT |
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
			restoreVT(out),
			windows.SetConsoleMode(hOut, oldOutMode),
			windows.SetConsoleMode(hIn, oldInMode))
	}, util.Errors(errSetIn, errSetOut, errVT)
}

func setupGlobal() func() {
	hOut := windows.Handle(os.Stderr.Fd())
	var oldOutMode uint32
	err := windows.GetConsoleMode(hOut, &oldOutMode)
	if err != nil {
		return func() {}
	}
	err = windows.SetConsoleMode(hOut, wantedGlobalOutMode)
	if err != nil {
		return func() {}
	}
	return func() { windows.SetConsoleMode(hOut, oldOutMode) }
}

func sanitize(in, out *os.File) {}
