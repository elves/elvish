package eval

import (
	"os"
)

// AddDirer wraps the AddDir function.
type AddDirer interface {
	// AddDir adds a directory with the given weight to some storage.
	AddDir(dir string, weight float64) error
}

// Chdir changes the current directory. On success it also updates the PWD
// environment variable and records the new directory in the directory history.
// It returns nil as long as the directory changing part succeeds.
func Chdir(path string, store AddDirer) error {
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
	if store != nil {
		go func() {
			err := store.AddDir(pwd, 1)
			if err != nil {
				logger.Println("Failed to save dir to history:", err)
			}
		}()
	}
	return nil
}
