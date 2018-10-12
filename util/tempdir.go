package util

import (
	"os"
)

// InTempDir is like WithTempDir, but also cd into the directory before running
// the function, and cd backs after running the function if possible.
//
// It panics if it could not get the working directory or change directory.
//
// It is useful in tests.
func InTempDir(f func(string)) {
	tmpdir, cleanup := TestDir()
	defer cleanup()

	oldpwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	mustChdir(tmpdir)
	defer mustChdir(oldpwd)

	f(tmpdir)
}
