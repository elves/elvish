package testutil

import (
	"fmt"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/env"
)

// TempDir creates a temporary directory for testing that will be removed
// after the test finishes. It is different from testing.TB.TempDir in that it
// resolves symlinks in the path of the directory.
//
// It panics if the test directory cannot be created or symlinks cannot be
// resolved. It is only suitable for use in tests.
func TempDir(c Cleanuper) string {
	dir, err := os.MkdirTemp("", "elvishtest.")
	if err != nil {
		panic(err)
	}
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		panic(err)
	}
	c.Cleanup(func() {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to remove temp dir %s: %v\n", dir, err)
		}
	})
	return dir
}

// TempHome is equivalent to Setenv(c, env.HOME, TempDir(c))
func TempHome(c Cleanuper) string {
	return Setenv(c, env.HOME, TempDir(c))
}

// Chdir changes into a directory, and restores the original working directory
// when a test finishes. It returns the directory for easier chaining.
func Chdir(c Cleanuper, dir string) string {
	oldWd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	Must(os.Chdir(dir))
	c.Cleanup(func() {
		Must(os.Chdir(oldWd))
	})
	return dir
}

// InTempDir is equivalent to Chdir(c, TempDir(c)).
func InTempDir(c Cleanuper) string {
	return Chdir(c, TempDir(c))
}

// InTempHome is equivalent to Setenv(c, env.HOME, InTempDir(c))
func InTempHome(c Cleanuper) string {
	return Setenv(c, env.HOME, InTempDir(c))
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

// ApplyDir creates the given filesystem layout in the current directory.
func ApplyDir(dir Dir) {
	applyDir(dir, "")
}

func applyDir(dir Dir, prefix string) {
	for name, file := range dir {
		path := filepath.Join(prefix, name)
		switch file := file.(type) {
		case string:
			Must(os.WriteFile(path, []byte(file), 0644))
		case File:
			Must(os.WriteFile(path, []byte(file.Content), file.Perm))
		case Dir:
			Must(os.MkdirAll(path, 0755))
			applyDir(file, path)
		default:
			panic(fmt.Sprintf("file is neither string, Dir, or Symlink: %v", file))
		}
	}
}
