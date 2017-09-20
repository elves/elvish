package util

import (
	"io/ioutil"
	"os"
)

// WithTempDir creates a new temporary directory and runs a function, passing
// the path of the temporary directory. After the function returns, it removes
// the temporary directory. It panics if it cannot make a temporary directory or
// cd into it.
//
// It is useful in tests.
func WithTempDir(f func(string)) {
	tmpdir, err := ioutil.TempDir("", "elvishtest.")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	f(tmpdir)
}

// InTempDir is like WithTempDir, but also cd into the directory before running
// the function, and cd backs after running the function if possible.
//
// It is useful in tests.
func InTempDir(f func(string)) {
	WithTempDir(func(tmpdir string) {
		pwd, err := os.Getwd()
		if err != nil {
			defer os.Chdir(pwd)
		}

		err = os.Chdir(tmpdir)
		if err != nil {
			panic(err)
		}
		f(tmpdir)
	})
}
