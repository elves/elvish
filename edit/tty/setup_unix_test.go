// +build !windows,plan9

package tty

import (
	"testing"

	"github.com/kr/pty"
)

func TestSetupTerminal(t *testing.T) {
	pty, tty, err := pty.Open()
	if err != nil {
		t.Errorf("cannot open pty for testing setupTerminal")
	}
	defer pty.Close()
	defer tty.Close()

	_, err = setup(tty)
	if err != nil {
		t.Errorf("setupTerminal returns an error")
	}
	// TODO(xiaq): Test whether the interesting flags in the termios were indeed
	// set.
	// termios, err := sys.NewTermiosFromFd(int(tty.Fd()))
}
