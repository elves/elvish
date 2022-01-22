package term

import (
	"os"

	"golang.org/x/sys/windows"
	"src.elv.sh/pkg/diag"
)

const (
	wantedInMode = windows.ENABLE_WINDOW_INPUT |
		windows.ENABLE_MOUSE_INPUT | windows.ENABLE_PROCESSED_INPUT
	wantedOutMode = windows.ENABLE_PROCESSED_OUTPUT |
		windows.ENABLE_WRAP_AT_EOL_OUTPUT |
		windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING

	additionalGlobalOutMode = windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
)

var globalOldInMode, globalOldOutMode uint32

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
		return diag.Errors(
			restoreVT(out),
			windows.SetConsoleMode(hOut, oldOutMode),
			windows.SetConsoleMode(hIn, oldInMode))
	}, diag.Errors(errSetIn, errSetOut, errVT)
}

func setupGlobal(in, out *os.File) func() {
	hIn := windows.Handle(in.Fd())
	hOut := windows.Handle(out.Fd())

	err := windows.GetConsoleMode(hIn, &globalOldInMode)
	if err != nil {
		return func() {}
	}
	err = windows.GetConsoleMode(hOut, &globalOldOutMode)
	if err != nil {
		return func() {}
	}

	windows.SetConsoleMode(hIn, globalOldInMode)
	windows.SetConsoleMode(hOut, globalOldOutMode|additionalGlobalOutMode)

	return func() {
		windows.SetConsoleMode(hIn, globalOldInMode)
		windows.SetConsoleMode(hOut, globalOldOutMode)
	}
}

func sanitize(in, out *os.File) {
	hIn := windows.Handle(in.Fd())
	hOut := windows.Handle(out.Fd())

	windows.SetConsoleMode(hIn, globalOldInMode)
	windows.SetConsoleMode(hOut, globalOldOutMode|additionalGlobalOutMode)
}
