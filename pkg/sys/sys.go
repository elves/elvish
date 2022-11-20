// Package sys provide system utilities with the same API across OSes.
//
// The subpackages eunix and ewindows provide OS-specific utilities.
package sys

import (
	"os"

	"github.com/mattn/go-isatty"
)

const sigsChanBufferSize = 256

// NotifySignals returns a channel on which all signals gets delivered.
func NotifySignals() chan os.Signal { return notifySignals() }

// SIGWINCH is the window size change signal.
const SIGWINCH = sigWINCH

// Winsize queries the size of the terminal referenced by the given file.
func WinSize(file *os.File) (row, col int) { return winSize(file) }

// IsATTY determines whether the given file is a terminal.
func IsATTY(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
