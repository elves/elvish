package testutil

import (
	"io"
	"io/ioutil"
	"os"
)

func MustPipe() (*os.File, *os.File) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	return r, w
}

func MustReadAllAndClose(r io.ReadCloser) []byte {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	r.Close()
	return bs
}

// MustMkdirAll calls os.MkdirAll and panics if an error is returned.
func MustMkdirAll(names ...string) {
	for _, name := range names {
		err := os.MkdirAll(name, 0700)
		if err != nil {
			panic(err)
		}
	}
}

// MustCreateEmpty creates an empty file, and panics if an error occurs.
func MustCreateEmpty(names ...string) {
	for _, name := range names {
		file, err := os.Create(name)
		if err != nil {
			panic(err)
		}
		file.Close()
	}
}

// MustWriteFile calls ioutil.WriteFile and panics if an error occurs.
func MustWriteFile(filename string, data []byte, perm os.FileMode) {
	err := ioutil.WriteFile(filename, data, perm)
	if err != nil {
		panic(err)
	}
}

// MustChdir calls os.Chdir and panics if it fails.
func MustChdir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

// Must panics if the error value is not nil. It is typically used like this:
//
//   testutil.Must(a_function())
//
// Where `a_function` returns a single error value. This is useful with
// functions like os.Mkdir to succinctly ensure the test fails to proceed if a
// "can't happen" failure does, in fact, happen.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
