package location

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store/storedefs"
)

type testStore struct {
	storedDirs []storedefs.Dir
	dirsError  error
	chdir      func(dir string) error
	wd         string
}

func (ts testStore) Dirs(blacklist map[string]struct{}) ([]storedefs.Dir, error) {
	dirs := []storedefs.Dir{}
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
	app, ttyCtrl, teardown := setup()
	defer teardown()

	Start(app, Config{})

	wantNotesBuf := bb().Write("no dir history store").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestStart_StoreError(t *testing.T) {
	app, ttyCtrl, teardown := setup()
	defer teardown()

	Start(app, Config{Store: testStore{dirsError: errors.New("ERROR")}})

	wantNotesBuf := bb().Write("db error: ERROR").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestStart_Hidden(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	dirs := []storedefs.Dir{
		{Path: fix("/usr/bin"), Score: 200},
		{Path: fix("/usr"), Score: 100},
		{Path: fix("/tmp"), Score: 50},
	}
	Start(app, Config{
		Store:         testStore{storedDirs: dirs},
		IterateHidden: func(f func(string)) { f(fix("/usr")) },
	})
	// Test UI.
	wantBuf := listingBuf(
		"",
		"200 "+fix("/usr/bin"), "<- selected",
		" 50 "+fix("/tmp"))
	ttyCtrl.TestBuffer(t, wantBuf)
}

func TestStart_Pinned(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	dirs := []storedefs.Dir{
		{Path: fix("/usr/bin"), Score: 200},
		{Path: fix("/usr"), Score: 100},
		{Path: fix("/tmp"), Score: 50},
	}
	Start(app, Config{
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
	ttyCtrl.TestBuffer(t, wantBuf)
}

func TestStart_HideWd(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	dirs := []storedefs.Dir{
		{Path: fix("/home"), Score: 200},
		{Path: fix("/tmp"), Score: 50},
	}
	Start(app, Config{Store: testStore{storedDirs: dirs, wd: fix("/home")}})
	// Test UI.
	wantBuf := listingBuf(
		"",
		" 50 "+fix("/tmp"), "<- selected")
	ttyCtrl.TestBuffer(t, wantBuf)
}

func TestStart_Workspace(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	chdir := ""
	dirs := []storedefs.Dir{
		{Path: fix("home/src"), Score: 200},
		{Path: fix("ws1/src"), Score: 150},
		{Path: fix("ws2/bin"), Score: 100},
		{Path: fix("/tmp"), Score: 50},
	}
	Start(app, Config{
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
	ttyCtrl.TestBuffer(t, wantBuf)

	ttyCtrl.Inject(term.K(ui.Enter))
	wantBuf = bb().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
	wantChdir := fix("/home/elf/src")
	if chdir != wantChdir {
		t.Errorf("got chdir %q, want %q", chdir, wantChdir)
	}
}

func TestStart_OK(t *testing.T) {
	home, cleanupHome := eval.InTempHome()
	defer cleanupHome()
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	errChdir := errors.New("mock chdir error")
	chdirCh := make(chan string, 100)
	dirs := []storedefs.Dir{
		{Path: filepath.Join(home, "go"), Score: 200},
		{Path: home, Score: 100},
		{Path: fix("/tmp/foo/bar/lorem/ipsum"), Score: 50},
	}
	Start(app, Config{Store: testStore{
		storedDirs: dirs,
		chdir:      func(dir string) error { chdirCh <- dir; return errChdir },
	}})

	// Test UI.
	wantBuf := listingBuf(
		"",
		"200 "+filepath.Join("~", "go"), "<- selected",
		"100 ~",
		" 50 "+fix("/tmp/foo/bar/lorem/ipsum"))
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test filtering.
	ttyCtrl.Inject(term.K('f'), term.K(os.PathSeparator), term.K('l'))

	wantBuf = listingBuf(
		"f"+string(os.PathSeparator)+"l",
		" 50 "+fix("/tmp/foo/bar/lorem/ipsum"), "<- selected")
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test accepting.
	ttyCtrl.Inject(term.K(ui.Enter))
	// There should be no change to codearea after accepting.
	wantBuf = bb().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
	// Error from Chdir should be sent to notes.
	wantNotesBuf := bb().Write("mock chdir error").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
	// Chdir should be called.
	wantChdir := fix("/tmp/foo/bar/lorem/ipsum")
	if got := <-chdirCh; got != wantChdir {
		t.Errorf("Chdir called with %s, want %s", got, wantChdir)
	}
}

func setup() (cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	// Use a smaller TTY size to make diffs easier to see.
	ttyCtrl.SetSize(20, 50)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := cli.ReadCodeAsync(app)
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}

func bb() *ui.BufferBuilder {
	return ui.NewBufferBuilder(50)
}

func listingBuf(filter string, lines ...string) *ui.Buffer {
	b := bb()
	b.Newline() // empty code area
	layout.WriteListing(b, "LOCATION", filter, lines...)
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
