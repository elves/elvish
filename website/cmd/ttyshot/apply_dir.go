package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// TODO: Add unit test and move this file to pkg/fsutil.

// Dir describes the layout of a directory. The keys of the map represent
// filenames. Each value is either a string (for the content of a regular file
// with permission 0644), a File, or a Dir.
type Dir map[string]any

// File describes a file to create.
type File struct {
	Perm    os.FileMode
	Content string
}

// ApplyDir creates a given filesystem layout.
func ApplyDir(dir Dir, root string) error {
	for name, file := range dir {
		path := filepath.Join(root, name)
		var err error
		switch file := file.(type) {
		case string:
			err = os.WriteFile(path, []byte(file), 0644)
		case File:
			err = os.WriteFile(path, []byte(file.Content), file.Perm)
		case Dir:
			err = os.MkdirAll(path, 0755)
			if err == nil {
				err = ApplyDir(file, path)
			}
		default:
			panic(fmt.Sprintf("file must be string, Dir, or File, got %T", file))
		}
		if err != nil {
			return err
		}
	}
	return nil
}
