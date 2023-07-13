package modes

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"

	"src.elv.sh/pkg/cli/lscolors"
	"src.elv.sh/pkg/ui"
)

// NavigationCursor represents a cursor for navigating in a potentially virtual
// filesystem.
type NavigationCursor interface {
	// Current returns a File that represents the current directory.
	Current() (NavigationFile, error)
	// Parent returns a File that represents the parent directory. It may return
	// nil if the current directory is the root of the filesystem.
	Parent() (NavigationFile, error)
	// Ascend navigates to the parent directory.
	Ascend() error
	// Descend navigates to the named child directory.
	Descend(name string) error
}

// NavigationFile represents a potentially virtual file.
type NavigationFile interface {
	// Name returns the name of the file.
	Name() string
	// ShowName returns a styled filename.
	ShowName() ui.Text
	// IsDirDeep returns whether the file is itself a directory or a symlink to
	// a directory.
	IsDirDeep() bool
	// Read returns either a list of File's if the File represents a directory,
	// a (possibly incomplete) slice of bytes if the File represents a normal
	// file, or an error if the File cannot be read.
	Read() ([]NavigationFile, []byte, error)
}

// NewOSNavigationCursor returns a NavigationCursor backed by the OS.
func NewOSNavigationCursor(chdir func(string) error) NavigationCursor {
	return osCursor{chdir, lscolors.GetColorist()}
}

type osCursor struct {
	chdir    func(string) error
	colorist lscolors.Colorist
}

func (c osCursor) Current() (NavigationFile, error) {
	abs, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	return file{filepath.Base(abs), abs, os.ModeDir, c.colorist}, nil
}

func (c osCursor) Parent() (NavigationFile, error) {
	if abs, _ := filepath.Abs("."); abs == "/" {
		return emptyDir{}, nil
	}
	abs, err := filepath.Abs("..")
	if err != nil {
		return nil, err
	}
	return file{filepath.Base(abs), abs, os.ModeDir, c.colorist}, nil
}

func (c osCursor) Ascend() error { return c.chdir("..") }

func (c osCursor) Descend(name string) error { return c.chdir(name) }

type emptyDir struct{}

func (emptyDir) Name() string                            { return "" }
func (emptyDir) ShowName() ui.Text                       { return nil }
func (emptyDir) IsDirDeep() bool                         { return true }
func (emptyDir) Read() ([]NavigationFile, []byte, error) { return []NavigationFile{}, nil, nil }

type file struct {
	name     string
	path     string
	mode     os.FileMode
	colorist lscolors.Colorist
}

func (f file) Name() string { return f.name }

func (f file) ShowName() ui.Text {
	sgrStyle := f.colorist.GetStyle(f.path)
	return ui.Text{&ui.Segment{
		Style: ui.StyleFromSGR(sgrStyle), Text: f.name}}
}

func (f file) IsDirDeep() bool {
	if f.mode.IsDir() {
		// File itself is a directory; return true and save a stat call.
		return true
	}
	info, err := os.Stat(f.path)
	return err == nil && info.IsDir()
}

const previewBytes = 64 * 1024

var (
	errNamedPipe  = errors.New("no preview for named pipe")
	errDevice     = errors.New("no preview for device file")
	errSocket     = errors.New("no preview for socket file")
	errCharDevice = errors.New("no preview for char device")
	errNonUTF8    = errors.New("no preview for non-utf8 file")
)

var specialFileModes = []struct {
	mode os.FileMode
	err  error
}{
	{os.ModeNamedPipe, errNamedPipe},
	{os.ModeDevice, errDevice},
	{os.ModeSocket, errSocket},
	{os.ModeCharDevice, errCharDevice},
}

func (f file) Read() ([]NavigationFile, []byte, error) {
	// On Unix, opening a named pipe for reading is blocking when there are no
	// writers, so we need to do this check at the very beginning of this
	// function.
	//
	// TODO: There is still a chance that the file has changed between when
	// f.mode is populated and the os.Open call below, in which case the os.Open
	// call can still block. This can be fixed by opening the file in async mode
	// and setting a timeout on the reads. Reading the file asynchronously is
	// also desirable behavior more generally for the navigation mode to remain
	// usable on slower filesystems.
	if f.mode&os.ModeNamedPipe != 0 {
		return nil, nil, errNamedPipe
	}

	ff, err := os.Open(f.path)
	if err != nil {
		return nil, nil, err
	}
	defer ff.Close()

	info, err := ff.Stat()
	if err != nil {
		return nil, nil, err
	}

	for _, special := range specialFileModes {
		if info.Mode()&special.mode != 0 {
			return nil, nil, special.err
		}
	}

	if info.IsDir() {
		infos, err := ff.Readdir(0)
		if err != nil {
			return nil, nil, err
		}
		files := make([]NavigationFile, len(infos))
		for i, info := range infos {
			files[i] = file{
				info.Name(),
				filepath.Join(f.path, info.Name()),
				info.Mode(),
				f.colorist,
			}
		}
		return files, nil, err
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
