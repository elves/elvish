package tty

import "os"

// Setup sets up the terminal so that it is suitable for the Reader and
// Writer to use. It returns a function that can be used to restore the
// original terminal config.
func Setup(in, out *os.File) (func() error, error) {
	return setup(in, out)
}
