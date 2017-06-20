package storedefs

import (
	"errors"
	"os"

	"github.com/elves/elvish/util"
)

// ErrEmptyHOME is the error returned by EnsureDataDir when the environmental
// variable HOME is empty.
var ErrEmptyHOME = errors.New("environment variable HOME is empty")

// EnsureDataDir ensures Elvish's data directory exists, creating it if
// necessary. It returns the path to the data directory (never with a
// trailing slash) and possible error.
func EnsureDataDir() (string, error) {
	home, err := util.GetHome("")
	if err != nil {
		return "", err
	}
	ddir := home + "/.elvish"
	return ddir, os.MkdirAll(ddir, 0700)
}
