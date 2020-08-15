package eval

import (
	"os"

	"github.com/elves/elvish/pkg/util"
)

// Chdir changes the current directory. On success it also updates the PWD
// environment variable and records the new directory in the directory history.
// It runs the functions in beforeChdir immediately before changing the
// directory, and the functions in afterChdir immediately after (if chdir was
// successful). It returns nil as long as the directory changing part succeeds.
func (ev *Evaler) Chdir(path string) error {
	for _, hook := range ev.beforeChdir {
		hook(path)
	}

	err := os.Chdir(path)
	if err != nil {
		return err
	}

	for _, hook := range ev.afterChdir {
		hook(path)
	}

	pwd, err := os.Getwd()
	if err != nil {
		logger.Println("getwd after cd:", err)
		return nil
	}
	os.Setenv(util.EnvPWD, pwd)

	return nil
}
