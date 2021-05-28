package term

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
	"src.elv.sh/pkg/diag"
)

const (
	wantedInMode = windows.ENABLE_WINDOW_INPUT |
		windows.ENABLE_MOUSE_INPUT | windows.ENABLE_PROCESSED_INPUT
	wantedOutMode = windows.ENABLE_PROCESSED_OUTPUT |
		windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	wantedGlobalOutMode = windows.ENABLE_PROCESSED_OUTPUT |
		windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
)

var (
	user32                = syscall.NewLazyDLL("user32.dll")
	procGetKeyboardLayout = user32.NewProc("GetKeyboardLayout")
	procVkKeyScanExA      = user32.NewProc("VkKeyScanExA")

	currentKeyboardLayoutHasAltGr = false
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

	layout, _, err := syscall.Syscall(procGetKeyboardLayout.Addr(), 1, 0, 0, 0)
	if err == windows.ERROR_SUCCESS {

		// Shamelessly stolen from
		// https://stackoverflow.com/questions/54588823/detect-if-the-keyboard-layout-has-altgr-on-it-under-windows
		for char := 0x20; char <= 0xff; char += 1 {
			scancode, _, err := syscall.Syscall(procVkKeyScanExA.Addr(), 2, uintptr(char), layout, 0)

			if err == windows.ERROR_SUCCESS && scancode&0x0600 == 0x0600 {
				// At least one ASCII char requires CTRL and ALT to be pressed
				currentKeyboardLayoutHasAltGr = true
				break
			}
		}
	}

	return func() error {
		return diag.Errors(
			restoreVT(out),
			windows.SetConsoleMode(hOut, oldOutMode),
			windows.SetConsoleMode(hIn, oldInMode))
	}, diag.Errors(errSetIn, errSetOut, errVT)
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
