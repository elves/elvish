//go:build unix

package term

import (
	"os"
	"testing"

	"github.com/creack/pty"
)

func TestSetupForTUIOnce(t *testing.T) {
	_, tty := setupPTY(t)

	setupForTUIOnce(tty, tty)
	// TODO: Test whether the interesting flags in the termios were indeed set.
}

func TestSetupForTUI(t *testing.T) {
	_, tty := setupPTY(t)

	_, err := setupForTUI(tty, tty)
	if err != nil {
		t.Errorf("setupForTUI returns an error")
	}
	// TODO: Test whether the interesting flags in the termios were indeed set.
	// termios, err := eunix.TermiosForFd(int(tty.Fd()))
}

func setupPTY(t *testing.T) (ptySide, ttySide *os.File) {
	t.Helper()
	ptySide, ttySide, err := pty.Open()
	if err != nil {
		t.Skip("cannot open pty")
	}
	t.Cleanup(func() {
		ptySide.Close()
		ttySide.Close()
	})
	return ptySide, ttySide
}
