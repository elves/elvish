package util

import (
	"io/ioutil"
	"os"
)

// InTempDir creates a new temporary directory, cd into it, and runs a function,
// passing the path of the temporary directory. After the function returns, it
// goes back to the original directory if possible, and remove the temporary
// directory. It panics if it cannot make a temporary directory or cd into it.
//
// It is useful in tests.
func InTempDir(f func(string)) {
	tmpdir, err := ioutil.TempDir("", "elvishtest.")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	pwd, err := os.Getwd()
	if err != nil {
		defer os.Chdir(pwd)
	}

	err = os.Chdir(tmpdir)
	if err != nil {
		panic(err)
	}
	f(tmpdir)
}
