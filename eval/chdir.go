package eval

import (
	"os"

	"github.com/elves/elvish/daemon/api"
)

// Chdir changes the current directory. On success it also updates the PWD
// environment variable and records the new directory in the directory history.
// It returns nil as long as the directory changing part succeeds.
func Chdir(path string, daemon *api.Client) error {
	err := os.Chdir(path)
	if err != nil {
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		logger.Println("getwd after cd:", err)
		return nil
	}
	os.Setenv("PWD", pwd)
	if daemon != nil {
		daemon.Waits().Add(1)
		go func() {
			daemon.AddDir(pwd, 1)
			daemon.Waits().Done()
		}()
	}
	return nil
}
