package term

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func TestSetupGlobalTerminal(t *testing.T) {
	in, out, release, err := createStdInOut()
	if err != nil {
		t.Errorf("cannot open stdin/stdout %v", err)
		return
	}
	defer release()

	initialOutMode, _ := getConsoleMode(out)
	initialOutMode = initialOutMode &^ windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	setConsoleMode(out, initialOutMode)

	// check that mode is for control sequences
	restore := setupGlobal(in, out)
	err = assertConsoleMode(
		out,
		initialOutMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	if err != nil {
		t.Errorf("got err %v, want nil", err)
		return
	}

	// check that mode is restored
	restore()
	err = assertConsoleMode(
		out,
		initialOutMode)
	if err != nil {
		t.Errorf("got err %v, want nil", err)
		return
	}
}

func TestSanitizeTerminal(t *testing.T) {
	in, out, release, err := createStdInOut()
	if err != nil {
		t.Errorf("cannot open stdin/stdout %v", err)
		return
	}
	defer release()

	initialOutMode, _ := getConsoleMode(out)
	setConsoleMode(out, initialOutMode)

	setupGlobal(in, out)

	// break console mode
	setConsoleMode(out, initialOutMode&^windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)

	sanitize(in, out)

	// check that sanitized
	err = assertConsoleMode(
		out,
		initialOutMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	if err != nil {
		t.Errorf("got err %v, want nil", err)
		return
	}
}

func assertConsoleMode(file *os.File, want uint32) error {
	got, err := getConsoleMode(file)
	if err != nil {
		return err
	} else if got != want {
		return fmt.Errorf("got %b, want %b", got, want)
	} else {
		return nil
	}
}

// open stdin/stdout manually because os.Stdin/os.Stdout cannot use in testing
func createStdInOut() (*os.File, *os.File, func(), error) {
	in, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, nil, err
	}
	out, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, nil, err
	}
	release := func() {
		in.Close()
		out.Close()
	}
	return in, out, release, nil
}

func setConsoleMode(file *os.File, mode uint32) error {
	err := windows.SetConsoleMode(windows.Handle(file.Fd()), mode)
	return err
}

func getConsoleMode(file *os.File) (uint32, error) {
	var mode uint32
	err := windows.GetConsoleMode(windows.Handle(file.Fd()), &mode)
	return mode, err
}
