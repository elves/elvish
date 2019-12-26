package runtime

import (
	"os"

	"github.com/elves/elvish/pkg/util"
)

// Ensures Elvish's data directory exists, creating it if necessary. It returns
// the path to the data directory (never with a trailing slash) and possible
// error.
func ensureDataDir() (string, error) {
	home, err := util.GetHome("")
	if err != nil {
		return "", err
	}
	ddir := home + "/.elvish"
	return ddir, os.MkdirAll(ddir, 0700)
}
