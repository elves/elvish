package term

import (
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func TestSetupForEval(t *testing.T) {
	// open CONOUT$ manually because os.Stdout is redirected during testing
	out := openFile(t, "CONOUT$", os.O_RDWR, 0)
	defer out.Close()

	// Start with ENABLE_VIRTUAL_TERMINAL_PROCESSING
	initialOutMode := getConsoleMode(t, out) | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	setConsoleMode(t, out, initialOutMode)

	// Clear ENABLE_VIRTUAL_TERMINAL_PROCESSING
	modifiedOutMode := initialOutMode &^ windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	setConsoleMode(t, out, modifiedOutMode)

	// Check that SetupForEval sets ENABLE_VIRTUAL_TERMINAL_PROCESSING without
	// changing other bits
	restore := setupForEval(nil, out)
	if got := getConsoleMode(t, out); got != initialOutMode {
		t.Errorf("got console mode %v, want %v", got, initialOutMode)
	}

	// Check that restore is a no-op
	setConsoleMode(t, out, modifiedOutMode)

	restore()
	if got := getConsoleMode(t, out); got != modifiedOutMode {
		t.Errorf("got console mode %v, want %v", got, modifiedOutMode)
	}
}

func openFile(t *testing.T, name string, flag int, perm os.FileMode) *os.File {
	t.Helper()
	out, err := os.OpenFile(name, flag, perm)
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	return out
}

func setConsoleMode(t *testing.T, file *os.File, mode uint32) {
	t.Helper()
	err := windows.SetConsoleMode(windows.Handle(file.Fd()), mode)
	if err != nil {
		t.Fatal("SetConsoleMode:", err)
	}
}

func getConsoleMode(t *testing.T, file *os.File) uint32 {
	t.Helper()
	var mode uint32
	err := windows.GetConsoleMode(windows.Handle(file.Fd()), &mode)
	if err != nil {
		t.Fatal("GetConsoleMode:", err)
	}
	return mode
}
