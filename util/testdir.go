package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// TestDir creates a temporary directory for testing. It returns the path of the
// temporary directory and a cleanup function to remove the temporary directory.
// The path has symlinks resolved with filepath.EvalSymlinks.
//
// It panics if the test directory cannot be created or symlinks cannot be
// resolved. It is only suitable for use in tests.
func TestDir() (string, func()) {
	dir, err := ioutil.TempDir("", "elvishtest.")
	if err != nil {
		panic(err)
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		panic(err)
	}
	return dir, func() {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: failed to remove temp dir", dir)
		}
	}
}

// InTestDir is like TestDir, but also changes into the test directory, and the
// cleanup function also changes back to the original working directory.
//
// It panics if it could not get the working directory or change directory. It
// is only suitable for use in tests.
func InTestDir() (string, func()) {
	oldWd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir, cleanup := TestDir()
	mustChdir(dir)
	return dir, func() {
		mustChdir(oldWd)
		cleanup()
	}
}

func mustChdir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
