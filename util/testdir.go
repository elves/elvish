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
			fmt.Fprintf(os.Stderr, "failed to remove temp dir %s: %v\n", dir, err)
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

// Dir describes the layout of a directory. The keys of the map represent
// filenames. Each value is either a string (for the content of a regular file
// with permission 0644), a File, or a Dir.
type Dir map[string]interface{}

// File describes a file to create.
type File struct {
	Perm    os.FileMode
	Content string
}

// SetupTestDir sets up a temporary directory using the given layout. If wd is
// not empty, it also changes into the given subdirectory. It returns a cleanup
// function to remove the temporary directory and restore the working directory.
//
// It panics if there are any errors.
func SetupTestDir(dir Dir, wd string) func() {
	_, cleanup := InTestDir()
	applyDir(dir, "")
	if wd != "" {
		mustChdir(wd)
	}
	return cleanup
}

func applyDir(dir Dir, prefix string) {
	for name, file := range dir {
		path := filepath.Join(prefix, name)
		switch file := file.(type) {
		case string:
			mustOK(ioutil.WriteFile(path, []byte(file), 0644))
		case File:
			mustOK(ioutil.WriteFile(path, []byte(file.Content), file.Perm))
		case Dir:
			mustOK(os.Mkdir(path, 0755))
			applyDir(file, path)
		default:
			panic(fmt.Sprintf("file is neither string nor Dir: %v", file))
		}
	}
}

func mustChdir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
