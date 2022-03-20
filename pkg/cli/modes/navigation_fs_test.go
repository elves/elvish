package modes

import (
	"errors"
	"strings"

	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

var (
	errCannotCd   = errors.New("cannot cd")
	errNoSuchFile = errors.New("no such file")
	errNoSuchDir  = errors.New("no such directory")
)

type testCursor struct {
	root testutil.Dir
	pwd  []string

	currentErr, parentErr, ascendErr, descendErr error
}

func (c *testCursor) Current() (NavigationFile, error) {
	if c.currentErr != nil {
		return nil, c.currentErr
	}
	return getDirFile(c.root, c.pwd)
}

func (c *testCursor) Parent() (NavigationFile, error) {
	if c.parentErr != nil {
		return nil, c.parentErr
	}
	parent := c.pwd
	if len(parent) > 0 {
		parent = parent[:len(parent)-1]
	}
	return getDirFile(c.root, parent)
}

func (c *testCursor) Ascend() error {
	if c.ascendErr != nil {
		return c.ascendErr
	}
	if len(c.pwd) > 0 {
		c.pwd = c.pwd[:len(c.pwd)-1]
	}
	return nil
}

func (c *testCursor) Descend(name string) error {
	if c.descendErr != nil {
		return c.descendErr
	}
	pwdCopy := append([]string{}, c.pwd...)
	childPath := append(pwdCopy, name)
	if _, err := getDirFile(c.root, childPath); err == nil {
		c.pwd = childPath
		return nil
	}
	return errCannotCd
}

func getFile(root testutil.Dir, path []string) (NavigationFile, error) {
	var f any = root
	for _, p := range path {
		d, ok := f.(testutil.Dir)
		if !ok {
			return nil, errNoSuchFile
		}
		f = d[p]
	}
	name := ""
	if len(path) > 0 {
		name = path[len(path)-1]
	}
	return testFile{name, f}, nil
}

func getDirFile(root testutil.Dir, path []string) (NavigationFile, error) {
	f, err := getFile(root, path)
	if err != nil {
		return nil, err
	}
	if !f.IsDirDeep() {
		return nil, errNoSuchDir
	}
	return f, nil
}

type testFile struct {
	name string
	data any
}

func (f testFile) Name() string { return f.name }

func (f testFile) ShowName() ui.Text {
	// The style matches that of LS_COLORS in the test code.
	switch {
	case f.IsDirDeep():
		return ui.T(f.name, ui.FgBlue)
	case strings.HasSuffix(f.name, ".png"):
		return ui.T(f.name, ui.FgRed)
	default:
		return ui.T(f.name)
	}
}

func (f testFile) IsDirDeep() bool {
	_, ok := f.data.(testutil.Dir)
	return ok
}

func (f testFile) Read() ([]NavigationFile, []byte, error) {
	if dir, ok := f.data.(testutil.Dir); ok {
		files := make([]NavigationFile, 0, len(dir))
		for name, data := range dir {
			files = append(files, testFile{name, data})
		}
		return files, nil, nil
	}
	return nil, []byte(f.data.(string)), nil
}
