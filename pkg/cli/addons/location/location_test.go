package location

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/elves/elvish/pkg/cli"
	. "github.com/elves/elvish/pkg/cli/apptest"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/store"
	"github.com/elves/elvish/pkg/ui"
)

type testStore struct {
	storedDirs []store.Dir
	dirsError  error
	chdir      func(dir string) error
	wd         string
}

func (ts testStore) Dirs(blacklist map[string]struct{}) ([]store.Dir, error) {
	dirs := []store.Dir{}
	for _, dir := range ts.storedDirs {
		if _, ok := blacklist[dir.Path]; ok {
			continue
		}
		dirs = append(dirs, dir)
	}
	return dirs, ts.dirsError
}

func (ts testStore) Chdir(dir string) error {
	if ts.chdir == nil {
		return nil
	}
	return ts.chdir(dir)
}

func (ts testStore) Getwd() (string, error) {
	return ts.wd, nil
}

func TestStart_NoStore(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{})

	f.TestTTYNotes(t, "no dir history store")
}

func TestStart_StoreError(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{Store: testStore{dirsError: errors.New("ERROR")}})

	f.TestTTYNotes(t, "db error: ERROR")
}

func TestStart_Hidden(t *testing.T) {
	f := Setup()
	defer f.Stop()

	dirs := []store.Dir{
		{Path: fix("/usr/bin"), Score: 200},
		{Path: fix("/usr"), Score: 100},
		{Path: fix("/tmp"), Score: 50},
	}
	Start(f.App, Config{
		Store:         testStore{storedDirs: dirs},
		IterateHidden: func(f func(string)) { f(fix("/usr")) },
	})
	// Test UI.
	wantBuf := listingBuf(
		"",
		"200 "+fix("/usr/bin"), "<- selected",
		" 50 "+fix("/tmp"))
	f.TTY.TestBuffer(t, wantBuf)
}

func TestStart_Pinned(t *testing.T) {
	f := Setup()
	defer f.Stop()

	dirs := []store.Dir{
		{Path: fix("/usr/bin"), Score: 200},
		{Path: fix("/usr"), Score: 100},
		{Path: fix("/tmp"), Score: 50},
	}
	Start(f.App, Config{
		Store:         testStore{storedDirs: dirs},
		IteratePinned: func(f func(string)) { f(fix("/home")); f(fix("/usr")) },
	})
	// Test UI.
	wantBuf := listingBuf(
		"",
		"  * "+fix("/home"), "<- selected",
		"  * "+fix("/usr"),
		"200 "+fix("/usr/bin"),
		" 50 "+fix("/tmp"))
	f.TTY.TestBuffer(t, wantBuf)
}

func TestStart_HideWd(t *testing.T) {
	f := Setup()
	defer f.Stop()

	dirs := []store.Dir{
		{Path: fix("/home"), Score: 200},
		{Path: fix("/tmp"), Score: 50},
	}
	Start(f.App, Config{Store: testStore{storedDirs: dirs, wd: fix("/home")}})
	// Test UI.
	wantBuf := listingBuf(
		"",
		" 50 "+fix("/tmp"), "<- selected")
	f.TTY.TestBuffer(t, wantBuf)
}

func TestStart_Workspace(t *testing.T) {
	f := Setup()
	defer f.Stop()

	chdir := ""
	dirs := []store.Dir{
		{Path: fix("home/src"), Score: 200},
		{Path: fix("ws1/src"), Score: 150},
		{Path: fix("ws2/bin"), Score: 100},
		{Path: fix("/tmp"), Score: 50},
	}
	Start(f.App, Config{
		Store: testStore{
			storedDirs: dirs,
			wd:         fix("/home/elf/bin"),
			chdir: func(dir string) error {
				chdir = dir
				return nil
			},
		},
		IterateWorkspaces: func(f func(kind, pattern string) bool) {
			if runtime.GOOS == "windows" {
				// Invalid patterns are ignored.
				f("ws1", `C:\\usr\\[^\\+`)
				f("home", `C:\\home\\[^\\]+`)
				f("ws2", `C:\\tmp\[^\]+`)
			} else {
				// Invalid patterns are ignored.
				f("ws1", "/usr/[^/+")
				f("home", "/home/[^/]+")
				f("ws2", "/tmp/[^/]+")
			}
		},
	})

	wantBuf := listingBuf(
		"",
		"200 "+fix("home/src"), "<- selected",
		" 50 "+fix("/tmp"))
	f.TTY.TestBuffer(t, wantBuf)

	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t /* nothing */)
	wantChdir := fix("/home/elf/src")
	if chdir != wantChdir {
		t.Errorf("got chdir %q, want %q", chdir, wantChdir)
	}
}

func TestStart_OK(t *testing.T) {
	home, cleanupHome := eval.InTempHome()
	defer cleanupHome()
	f := Setup()
	defer f.Stop()

	errChdir := errors.New("mock chdir error")
	chdirCh := make(chan string, 100)
	dirs := []store.Dir{
		{Path: filepath.Join(home, "go"), Score: 200},
		{Path: home, Score: 100},
		{Path: fix("/tmp/foo/bar/lorem/ipsum"), Score: 50},
	}
	Start(f.App, Config{Store: testStore{
		storedDirs: dirs,
		chdir:      func(dir string) error { chdirCh <- dir; return errChdir },
	}})

	// Test UI.
	wantBuf := listingBuf(
		"",
		"200 "+filepath.Join("~", "go"), "<- selected",
		"100 ~",
		" 50 "+fix("/tmp/foo/bar/lorem/ipsum"))
	f.TTY.TestBuffer(t, wantBuf)

	// Test filtering.
	f.TTY.Inject(term.K('f'), term.K(os.PathSeparator), term.K('l'))

	wantBuf = listingBuf(
		"f"+string(os.PathSeparator)+"l",
		" 50 "+fix("/tmp/foo/bar/lorem/ipsum"), "<- selected")
	f.TTY.TestBuffer(t, wantBuf)

	// Test accepting.
	f.TTY.Inject(term.K(ui.Enter))
	// There should be no change to codearea after accepting.
	f.TestTTY(t /* nothing */)
	// Error from Chdir should be sent to notes.
	f.TestTTYNotes(t, "mock chdir error")
	// Chdir should be called.
	wantChdir := fix("/tmp/foo/bar/lorem/ipsum")
	if got := <-chdirCh; got != wantChdir {
		t.Errorf("Chdir called with %s, want %s", got, wantChdir)
	}
}

func listingBuf(filter string, lines ...string) *term.Buffer {
	b := term.NewBufferBuilder(50)
	b.Newline() // empty code area
	cli.WriteListing(b, " LOCATION ", filter, lines...)
	return b.Buffer()
}

func fix(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}
	if path[0] == '/' {
		path = "C:" + path
	}
	return strings.ReplaceAll(path, "/", "\\")
}
