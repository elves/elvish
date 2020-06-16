// +build !windows

package progtest

import (
	"github.com/creack/pty"
	"github.com/elves/elvish/pkg/util"
)

// SetupInteractive sets up a test fixture for use by an interactive elvish
// shell. That is, one that reads commands from a tty/pty.
//
// The caller is responsible for calling the CleanupInteractive method of the
// returned Fixture.
func SetupInteractive() *Fixture {
	_, dirCleanup := util.InTestDir()
	pty, tty, err := pty.Open()
	if err != nil {
		panic(err)
	}
	pty_pipe := pipe{w: tty, r: pty}
	return &Fixture{[3]*pipe{&pty_pipe, makePipe(), makePipe()}, dirCleanup}
}
