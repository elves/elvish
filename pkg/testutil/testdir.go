package testutil

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/must"
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
	must.Chdir(dir)
	c.Cleanup(func() {
		must.Chdir(oldWd)
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
type Dir map[string]any

// File describes a file to create.
type File struct {
	Perm    os.FileMode
	Content string
}

// ApplyDir creates the given filesystem layout in the current directory.
func ApplyDir(dir Dir) {
	ApplyDirIn(dir, "")
}

// ApplyDirIn creates the given filesystem layout in a given directory.
func ApplyDirIn(dir Dir, root string) {
	for name, file := range dir {
		path := filepath.Join(root, name)
		switch file := file.(type) {
		case string:
			must.OK(os.WriteFile(path, []byte(file), 0644))
		case File:
			must.OK(os.WriteFile(path, []byte(file.Content), file.Perm))
		case Dir:
			must.OK(os.MkdirAll(path, 0755))
			ApplyDirIn(file, path)
		default:
			panic(fmt.Sprintf("file is neither string, Dir, or Symlink: %v", file))
		}
	}
}

// fs.FS implementation for Dir.

func (dir Dir) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		return newFsDir(".", dir), nil
	}
	currentDir := dir
	currentName := name
	for {
		first, rest, moreLevels := strings.Cut(currentName, "/")
		file, ok := currentDir[first]
		if !ok {
			return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
		if !moreLevels {
			return newFsFileOrDir(name, file), nil
		}
		if nextDir, ok := file.(Dir); ok {
			currentDir = nextDir
			currentName = rest
		} else {
			return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
	}
}

func newFsFileOrDir(name string, x any) fs.File {
	switch x := x.(type) {
	case Dir:
		return newFsDir(name, x)
	case File:
		return fsFile{newFsFileInfo(path.Base(name), x).(fileInfo), strings.NewReader(x.Content)}
	case string:
		return fsFile{newFsFileInfo(path.Base(name), x).(fileInfo), strings.NewReader(x)}
	default:
		panic(fmt.Sprintf("file is neither string, File or Dir: %v", x))
	}
}

func newFsFileInfo(basename string, x any) fs.FileInfo {
	switch x := x.(type) {
	case Dir:
		return dirInfo{basename}
	case File:
		return fileInfo{basename, x.Perm, len(x.Content)}
	case string:
		return fileInfo{basename, 0o644, len(x)}
	default:
		panic(fmt.Sprintf("file is neither string, File or Dir: %v", x))
	}
}

type fsDir struct {
	info    dirInfo
	readErr error
	entries []fs.DirEntry
}

var errIsDir = errors.New("is a directory")

func newFsDir(name string, dir Dir) *fsDir {
	info := dirInfo{path.Base(name)}
	readErr := &fs.PathError{Op: "read", Path: name, Err: errIsDir}
	entries := make([]fs.DirEntry, 0, len(dir))
	for name, file := range dir {
		entries = append(entries, fs.FileInfoToDirEntry(newFsFileInfo(name, file)))
	}
	return &fsDir{info, readErr, entries}
}

func (fd *fsDir) Stat() (fs.FileInfo, error) { return fd.info, nil }
func (fd *fsDir) Read([]byte) (int, error)   { return 0, fd.readErr }
func (fd *fsDir) Close() error               { return nil }

func (fd *fsDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if n <= 0 || (n >= len(fd.entries) && len(fd.entries) != 0) {
		ret := fd.entries
		fd.entries = nil
		return ret, nil
	}
	if len(fd.entries) == 0 {
		return nil, io.EOF
	}
	ret := fd.entries[:n]
	fd.entries = fd.entries[n:]
	return ret, nil
}

type dirInfo struct{ basename string }

var t0 = time.Unix(0, 0).UTC()

func (di dirInfo) Name() string    { return di.basename }
func (dirInfo) Size() int64        { return 0 }
func (dirInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o755 }
func (dirInfo) ModTime() time.Time { return t0 }
func (dirInfo) IsDir() bool        { return true }
func (dirInfo) Sys() any           { return nil }

type fsFile struct {
	info fileInfo
	*strings.Reader
}

func (ff fsFile) Stat() (fs.FileInfo, error) {
	return ff.info, nil
}

func (ff fsFile) Close() error { return nil }

type fileInfo struct {
	basename string
	perm     fs.FileMode
	size     int
}

func (fi fileInfo) Name() string      { return fi.basename }
func (fi fileInfo) Size() int64       { return int64(fi.size) }
func (fi fileInfo) Mode() fs.FileMode { return fi.perm }
func (fileInfo) ModTime() time.Time   { return t0 }
func (fileInfo) IsDir() bool          { return false }
func (fileInfo) Sys() any             { return nil }
