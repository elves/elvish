// +build !windows,!plan9

package term

import (
	"testing"

	"github.com/creack/pty"
)

func TestSetupTerminal(t *testing.T) {
	pty, tty, err := pty.Open()
	if err != nil {
		t.Skip("cannot open pty for testing setupTerminal")
	}
	defer pty.Close()
	defer tty.Close()

	_, err = setup(tty, tty)
	if err != nil {
		t.Errorf("setupTerminal returns an error")
	}
	// TODO(xiaq): Test whether the interesting flags in the termios were indeed
	// set.
	// termios, err := sys.TermiosForFd(int(tty.Fd()))
}
