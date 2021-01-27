package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/env"
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
	Must(os.Chdir(dir))
	return dir, func() {
		Must(os.Chdir(oldWd))
		cleanup()
	}
}

// InTempHome is like InTestDir, but it also sets HOME to the temporary
// directory and restores the original HOME in cleanup.
func InTempHome() (string, func()) {
	oldHome := os.Getenv(env.HOME)
	tmpHome, cleanup := InTestDir()
	os.Setenv(env.HOME, tmpHome)

	return tmpHome, func() {
		os.Setenv(env.HOME, oldHome)
		cleanup()
	}
}

// Dir describes the layout of a directory. The keys of the map represent
// filenames. Each value is either a string (for the content of a regular file
// with permission 0644), a File, or a Dir.
type Dir map[string]interface{}

// Symlink defines the target path of a symlink to be created.
type Symlink struct{ Target string }

// File describes a file to create.
type File struct {
	Perm    os.FileMode
	Content string
}

// ApplyDir creates the given filesystem layout in the current directory.
func ApplyDir(dir Dir) {
	applyDir(dir, "")
}

func applyDir(dir Dir, prefix string) {
	for name, file := range dir {
		path := filepath.Join(prefix, name)
		switch file := file.(type) {
		case string:
			Must(ioutil.WriteFile(path, []byte(file), 0644))
		case File:
			Must(ioutil.WriteFile(path, []byte(file.Content), file.Perm))
		case Dir:
			Must(os.Mkdir(path, 0755))
			applyDir(file, path)
		case Symlink:
			Must(os.Symlink(file.Target, path))
		default:
			panic(fmt.Sprintf("file is neither string, Dir, or Symlink: %v", file))
		}
	}
}
