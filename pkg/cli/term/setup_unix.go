//go:build unix

package term

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
	"src.elv.sh/pkg/errutil"
	"src.elv.sh/pkg/sys/eunix"
)

func setupForTUIOnce(in, _ *os.File) func() {
	fd := int(in.Fd())
	term, err := eunix.TermiosForFd(fd)
	if err != nil {
		return func() {}
	}
	savedTermios := term.Copy()

	// Turning off IXON frees up Ctrl-Q and Ctrl-S for keybindings, but it's not
	// actually necessary for Elvish to function.
	//
	// We do this in SetupForTUIOnce rather than SetupForTUI so that the user
	// can still use "stty ixon" to turn it on if they wish.
	//
	// Other "nice to have" terminal setups should go here as well.
	term.SetIXON(false)
	term.ApplyToFd(fd)

	return func() { savedTermios.ApplyToFd(fd) }
}

func setupForTUI(in, out *os.File) (func() error, error) {
	// On Unix, use input file for changing termios. All fds pointing to the
	// same terminal are equivalent.

	fd := int(in.Fd())
	term, err := eunix.TermiosForFd(fd)
	if err != nil {
		return nil, fmt.Errorf("can't get terminal attribute: %s", err)
	}

	savedTermios := term.Copy()

	term.SetICanon(false)
	term.SetIExten(false)
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

	var errSetupVT error
	err = setupVT(out)
	if err != nil {
		errSetupVT = fmt.Errorf("can't setup VT: %s", err)
	}

	restore := func() error {
		return errutil.Multi(savedTermios.ApplyToFd(fd), restoreVT(out))
	}

	return restore, errSetupVT
}

func setupForEval(in, out *os.File) func() {
	// There is nothing to set up on Unix, but we try to sanitize the terminal
	// when evaluation finishes.
	return func() { sanitize(in, out) }
}

func sanitize(in, out *os.File) {
	// Some programs use non-blocking IO but do not correctly clear the
	// non-blocking flags after exiting, so we always clear the flag. See #822
	// for an example.
	unix.SetNonblock(int(in.Fd()), false)
	unix.SetNonblock(int(out.Fd()), false)
}
