package sys

import (
	"os"
	"syscall"
)

// SIGWINCH is the Window size change signal. On Windows this signal does not
// exist, so we use -1, an impossible value for signals.
const SIGWINCH = syscall.Signal(-1)

// GetWinsize queries the size of the terminal referenced by the given file.
func GetWinsize(file *os.File) (row, col int) {
	panic("unimplemented")
}
