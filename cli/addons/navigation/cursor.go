package navigation

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// Cursor represents a cursor for navigating in a potentially virtual filesystem.
type Cursor interface {
	// Current returns a File that represents the current directory.
	Current() (File, error)
	// Parent returns a File that represents the parent directory. It may return
	// nil if the current directory is the root of the filesystem.
	Parent() (File, error)
	// Ascend navigates to the parent directory.
	Ascend() error
	// Descend navigates to the named child directory.
	Descend(name string) error
}

// File represents a potentially virtual file.
type File interface {
	// Name returns the name of the file.
	Name() string
	// Mode returns the file's mode and permissions.
	Mode() os.FileMode
	// DeepMode returns the file's mode and permissions, resolving symlinks.
	DeepMode() (os.FileMode, error)
	// Read returns either a list of File's if the File represents a directory,
	// a (possibly incomplete) slice of bytes if the File represents a normal
	// file, or an error if the File cannot be read.
	Read() ([]File, []byte, error)
}

// NewOSCursor returns a Cursor backed by the OS.
func NewOSCursor() Cursor { return osCursor{} }

type osCursor struct{}

func (c osCursor) Current() (File, error) {
	abs, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	return file{filepath.Base(abs), abs, os.ModeDir}, nil
}

func (c osCursor) Parent() (File, error) {
	if abs, _ := filepath.Abs("."); abs == "/" {
		return emptyDir{}, nil
	}
	abs, err := filepath.Abs("..")
	if err != nil {
		return nil, err
	}
	return file{filepath.Base(abs), abs, os.ModeDir}, nil
}

func (c osCursor) Ascend() error { return os.Chdir("..") }

func (c osCursor) Descend(name string) error { return os.Chdir(name) }

type emptyDir struct{}

func (emptyDir) Name() string                   { return "" }
func (emptyDir) Mode() os.FileMode              { return os.ModeDir }
func (emptyDir) DeepMode() (os.FileMode, error) { return os.ModeDir, nil }
func (emptyDir) Read() ([]File, []byte, error)  { return []File{}, nil, nil }

type file struct {
	name string
	path string
	mode os.FileMode
}

func (f file) Name() string      { return f.name }
func (f file) Mode() os.FileMode { return f.mode }

func (f file) DeepMode() (os.FileMode, error) {
	info, err := os.Stat(f.path)
	if err != nil {
		return 0, err
	}
	return info.Mode(), nil
}

const previewBytes = 64 * 1024

var (
	errDevice     = errors.New("no preview for device file")
	errNamedPipe  = errors.New("no preview for named pipe")
	errSocket     = errors.New("no preview for socket file")
	errCharDevice = errors.New("no preview for char device")
	errNonUTF8    = errors.New("no preview for non-utf8 file")
)

var specialFileModes = []struct {
	mode os.FileMode
	err  error
}{
	{os.ModeDevice, errDevice},
	{os.ModeNamedPipe, errNamedPipe},
	{os.ModeSocket, errSocket},
	{os.ModeCharDevice, errCharDevice},
}

func (f file) Read() ([]File, []byte, error) {
	ff, err := os.Open(f.path)
	if err != nil {
		return nil, nil, err
	}
	defer ff.Close()

	info, err := ff.Stat()
	if err != nil {
		return nil, nil, err
	}

	if info.IsDir() {
		infos, err := ff.Readdir(0)
		if err != nil {
			return nil, nil, err
		}
		files := make([]File, len(infos))
		for i, info := range infos {
			files[i] = file{
				info.Name(),
				filepath.Join(f.path, info.Name()),
				info.Mode(),
			}
		}
		return files, nil, err
	}

	for _, special := range specialFileModes {
		if info.Mode()&special.mode != 0 {
			return nil, nil, special.err
		}
	}

	var buf [previewBytes]byte
	nr, err := ff.Read(buf[:])
	if err != nil && err != io.EOF {
		return nil, nil, err
	}

	content := buf[:nr]
	if !utf8.Valid(content) {
		return nil, nil, errNonUTF8
	}

	return nil, content, nil
}
