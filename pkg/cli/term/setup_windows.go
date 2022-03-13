package term

import (
	"os"

	"golang.org/x/sys/windows"
	"src.elv.sh/pkg/diag"
)

const (
	inMode = windows.ENABLE_WINDOW_INPUT |
		windows.ENABLE_MOUSE_INPUT | windows.ENABLE_PROCESSED_INPUT
	outMode = windows.ENABLE_PROCESSED_OUTPUT |
		windows.ENABLE_WRAP_AT_EOL_OUTPUT |
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

	errSetIn := windows.SetConsoleMode(hIn, inMode)
	errSetOut := windows.SetConsoleMode(hOut, outMode)
	errVT := setupVT(out)

	return func() error {
		return diag.Errors(
			restoreVT(out),
			windows.SetConsoleMode(hOut, oldOutMode),
			windows.SetConsoleMode(hIn, oldInMode))
	}, diag.Errors(errSetIn, errSetOut, errVT)
}

const outFlagForEval = windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING

func setupForEval(_, out *os.File) func() {
	h := windows.Handle(out.Fd())
	var oldOutMode uint32
	err := windows.GetConsoleMode(h, &oldOutMode)
	if err == nil {
		windows.SetConsoleMode(h, oldOutMode|outFlagForEval)
	}
	return func() {}
}
