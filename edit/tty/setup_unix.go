// +build !windows,!plan9

package tty

import (
	"fmt"
	"os"

	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

const flushInputDuringSetup = false

func setup(in, out *os.File) (func() error, error) {
	// On Unix, use input file for changing termios. All fds pointing to the
	// same terminal are equivalent.

	fd := int(in.Fd())
	term, err := sys.NewTermiosFromFd(fd)
	if err != nil {
		return nil, fmt.Errorf("can't get terminal attribute: %s", err)
	}

	savedTermios := term.Copy()

	term.SetICanon(false)
	term.SetEcho(false)
	term.SetVMin(1)
	term.SetVTime(0)

	// Enforcing crnl translation on readline. Assuming user won't set
	// inlcr or -onlcr, otherwise we have to hardcode all of them here.
	term.SetICRNL(true)

	err = term.ApplyToFd(fd)
	if err != nil {
		return nil, fmt.Errorf("can't set up terminal attribute: %s", err)
	}

	var errFlushInput error
	if flushInputDuringSetup {
		err = sys.FlushInput(fd)
		if err != nil {
			errFlushInput = fmt.Errorf("can't flush input: %s", err)
		}
	}

	var errSetupVT error
	err = setupVT(out)
	if err != nil {
		errSetupVT = fmt.Errorf("can't setup VT: %s", err)
	}

	restore := func() error {
		return util.Errors(savedTermios.ApplyToFd(fd), restoreVT(out))
	}

	return restore, util.Errors(errFlushInput, errSetupVT)
}
