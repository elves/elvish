package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// WithTempDirs creates a requested number of temporary directories and runs a
// function, passing the paths of the temporary directories; the passed paths
// all have their symlinks resolved using filepath.EvalSymlinks. After the
// function returns, it removes the temporary directories. It panics if it
// cannot make a temporary directory, and prints an error message to stderr if
// it cannot remove the temporary directories.
//
// It is useful in tests.
func WithTempDirs(n int, f func([]string)) {
	tmpdirs := make([]string, n)
	for i := range tmpdirs {
		tmpdir, err := ioutil.TempDir("", "elvishtest.")
		if err != nil {
			panic(err)
		}
		tmpdirs[i], err = filepath.EvalSymlinks(tmpdir)
		if err != nil {
			panic(err)
		}
	}
	defer func() {
		for _, tmpdir := range tmpdirs {
			err := os.RemoveAll(tmpdir)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Warning: failed to remove temp dir", tmpdir)
			}
		}
	}()

	f(tmpdirs)
}

// WithTempDir is like with WithTempDirs, except that it always creates one
// temporary directory and pass the function a string instead of []string.
func WithTempDir(f func(string)) {
	WithTempDirs(1, func(s []string) {
		f(s[0])
	})
}

// InTempDir is like WithTempDir, but also cd into the directory before running
// the function, and cd backs after running the function if possible.
//
// It panics if it could not get the working directory or change directory.
//
// It is useful in tests.
func InTempDir(f func(string)) {
	WithTempDir(func(tmpdir string) {
		oldpwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		mustChdir(tmpdir)
		defer mustChdir(oldpwd)

		f(tmpdir)
	})
}

func mustChdir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
