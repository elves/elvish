// Package must contains simple functions that panic on errors.
//
// It should only be used in tests and rare places where errors are provably
// impossible.
package must

import (
	"io"
	"os"
	"path/filepath"
)

// OK panics if the error value is not nil. It is intended for use with
// functions that return just an error.
func OK(err error) {
	if err != nil {
		panic(err)
	}
}

// OK1 panics if the error value is not nil. It is intended for use with
// functions that return one value and an error.
func OK1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// OK2 panics if the error value is not nil. It is intended for use with
// functions that return two values and an error.
func OK2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return v1, v2
}

// Pipe wraps os.Pipe.
func Pipe() (*os.File, *os.File) {
	return OK2(os.Pipe())
}

// Chdir wraps os.Chdir.
func Chdir(dir string) {
	OK(os.Chdir(dir))
}

// ReadAll wraps io.ReadAll and io.Closer.Close.
func ReadAllAndClose(r io.ReadCloser) []byte {
	v := OK1(io.ReadAll(r))
	OK(r.Close())
	return v
}

// ReadFile wraps os.ReadFile.
func ReadFile(fname string) []byte {
	return OK1(os.ReadFile(fname))
}

// ReadFileString converts the result of ReadFile to a string.
func ReadFileString(fname string) string {
	return string(ReadFile(fname))
}

// MkdirAll calls os.MkdirAll for each argument.
func MkdirAll(names ...string) {
	for _, name := range names {
		OK(os.MkdirAll(name, 0700))
	}
}

// CreateEmpty creates empty file, after creating all ancestor directories that
// don't exist.
func CreateEmpty(names ...string) {
	for _, name := range names {
		OK(os.MkdirAll(filepath.Dir(name), 0700))
		file := OK1(os.Create(name))
		OK(file.Close())
	}
}

// WriteFile writes data to a file, after creating all ancestor directories that
// don't exist.
func WriteFile(filename, data string) {
	OK(os.MkdirAll(filepath.Dir(filename), 0700))
	OK(os.WriteFile(filename, []byte(data), 0600))
}
