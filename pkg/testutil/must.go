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

// Calls os.MkdirAll and panics if an error is returned.
func MustMkdirAll(names ...string) {
	for _, name := range names {
		err := os.MkdirAll(name, 0700)
		if err != nil {
			panic(err)
		}
	}
}

// Creates an empty file, and panics if an error occurs.
func MustCreateEmpty(names ...string) {
	for _, name := range names {
		file, err := os.Create(name)
		if err != nil {
			panic(err)
		}
		file.Close()
	}
}

// Calls ioutil.WriteFile and panics if an error occurs.
func MustWriteFile(filename string, data []byte, perm os.FileMode) {
	err := ioutil.WriteFile(filename, data, perm)
	if err != nil {
		panic(err)
	}
}
