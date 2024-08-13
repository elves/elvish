package term

import (
	"fmt"
	"os"

	"src.elv.sh/pkg/sys"
	"src.elv.sh/pkg/wcwidth"
)

// SetupForTUIOnce sets up the terminal once for a whole interactive session. It
// returns a function that can be used to restore the original terminal config.
func SetupForTUIOnce(in, out *os.File) func() {
	return setupForTUIOnce(in, out)
}

// SetupForTUI sets up the terminal so that it is suitable for the TUI
// application (as implemented by pkg/cli). It returns a function that can be
// used to restore the original terminal config.
func SetupForTUI(in, out *os.File) (func() error, error) {
	return setupForTUI(in, out)
}

// SetupForEval sets up the terminal for evaluating Elvish code, whether or not
// we are in an interactive session. It returns a function to call after the
// evaluation finishes.
func SetupForEval(in, out *os.File) func() {
	return setupForEval(in, out)
}

const (
	lackEOLRune    = '\u23ce'
	lackEOL        = "\033[7m" + string(lackEOLRune) + "\033[m"
	enableSGRMouse = false
)

// setupVT performs setup for VT-like terminals.
func setupVT(out *os.File) error {
	_, width := sys.WinSize(out)

	s := ""
	/*
		Write a lackEOLRune if the cursor is not in the leftmost column. This is
		done as follows:

		1. Turn on autowrap;

		2. Write lackEOL along with enough padding, so that the total width is
		   equal to the width of the screen.

		   If the cursor was in the first column, we are still in the same line,
		   just off the line boundary. Otherwise, we are now in the next line.

		3. Rewind to the first column, write one space and rewind again. If the
		   cursor was in the first column to start with, we have just erased the
		   LackEOL character. Otherwise, we are now in the next line and this is
		   a no-op. The LackEOL character remains.
	*/
	s += fmt.Sprintf("\033[?7h%s%*s\r \r", lackEOL, width-wcwidth.OfRune(lackEOLRune), "")

	/*
		Turn off autowrap.

		The terminals sometimes has different opinions about how wide some
		characters are (notably emojis and some dingbats) with elvish. When that
		happens, elvish becomes wrong about where the cursor is when it writes
		its output, and the effect can be disastrous.

		If we turn off autowrap, the terminal won't insert any newlines behind
		the scene, so elvish is always right about which line the cursor is.
		With a bit more caution, this can restrict the consequence of the
		mismatch within one line.
	*/
	s += "\033[?7l"

	// Turn on SGR-style mouse tracking.
	if enableSGRMouse {
		s += "\033[?1000;1006h"
	}

	// Enable bracketed paste.
	s += "\033[?2004h"

	_, err := out.WriteString(s)
	return err
}

// restoreVT performs restore for VT-like terminals.
func restoreVT(out *os.File) error {
	s := ""
	// Turn on autowrap.
	s += "\033[?7h"
	// Turn off mouse tracking.
	if enableSGRMouse {
		s += "\033[?1000;1006l"
	}
	// Disable bracketed paste.
	s += "\033[?2004l"
	// Move the cursor to the first row, even if we haven't written anything
	// visible. This is because the terminal driver might not be smart enough to
	// recognize some escape sequences as invisible and wrongly assume that we
	// are not in the first column, which can mess up with tabs. See
	// https://src.elv.sh/pkg/issues/629 for an example.
	s += "\r"
	_, err := out.WriteString(s)
	return err
}
