package testutil

import (
	"io"
	"os"
	"path/filepath"
)

// MustPipe calls os.Pipe. It panics if an error occurs.
func MustPipe() (*os.File, *os.File) {
	r, w, err := os.Pipe()
	Must(err)
	return r, w
}

// MustReadAllAndClose reads all bytes and closes the ReadCloser. It panics if
// an error occurs.
func MustReadAllAndClose(r io.ReadCloser) []byte {
	bs, err := io.ReadAll(r)
	Must(err)
	Must(r.Close())
	return bs
}

// MustMkdirAll calls os.MkdirAll for each argument. It panics if an error
// occurs.
func MustMkdirAll(names ...string) {
	for _, name := range names {
		Must(os.MkdirAll(name, 0700))
	}
}

// MustCreateEmpty creates empty file, after creating all ancestor directories
// that don't exist. It panics if an error occurs.
func MustCreateEmpty(names ...string) {
	for _, name := range names {
		Must(os.MkdirAll(filepath.Dir(name), 0700))
		file, err := os.Create(name)
		Must(err)
		Must(file.Close())
	}
}

// MustWriteFile writes data to a file, after creating all ancestor directories
// that don't exist. It panics if an error occurs.
func MustWriteFile(filename, data string) {
	Must(os.MkdirAll(filepath.Dir(filename), 0700))
	Must(os.WriteFile(filename, []byte(data), 0600))
}

// MustChdir calls os.Chdir and panics if it fails.
func MustChdir(dir string) {
	Must(os.Chdir(dir))
}

// Must panics if the error value is not nil. It is typically used like this:
//
//   testutil.Must(someFunction(...))
//
// Where someFunction returns a single error value. This is useful with
// functions like os.Mkdir to succinctly ensure the test fails to proceed if an
// operation required for the test setup results in an error.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
