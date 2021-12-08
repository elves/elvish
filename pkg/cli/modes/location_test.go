package modes

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

type locationStore struct {
	storedDirs []storedefs.Dir
	dirsError  error
	chdir      func(dir string) error
	wd         string
}

func (ts locationStore) Dirs(blacklist map[string]struct{}) ([]storedefs.Dir, error) {
	dirs := []storedefs.Dir{}
	for _, dir := range ts.storedDirs {
		if _, ok := blacklist[dir.Path]; ok {
			continue
		}
		dirs = append(dirs, dir)
	}
	return dirs, ts.dirsError
}

func (ts locationStore) Chdir(dir string) error {
	if ts.chdir == nil {
		return nil
	}
	return ts.chdir(dir)
}

func (ts locationStore) Getwd() (string, error) {
	return ts.wd, nil
}

func TestNewLocation_NoStore(t *testing.T) {
	f := Setup()
	defer f.Stop()

	_, err := NewLocation(f.App, LocationSpec{})
	if err != errNoDirectoryHistoryStore {
		t.Error("want errNoDirectoryHistoryStore")
	}
}

func TestNewLocation_StoreError(t *testing.T) {
	f := Setup()
	defer f.Stop()

	_, err := NewLocation(f.App,
		LocationSpec{Store: locationStore{dirsError: errors.New("ERROR")}})
	if err.Error() != "db error: ERROR" {
		t.Error("want db error")
	}
}

func TestLocation_FullWorkflow(t *testing.T) {
	home := testutil.InTempHome(t)
	f := Setup()
	defer f.Stop()

	errChdir := errors.New("mock chdir error")
	chdirCh := make(chan string, 100)
	dirs := []storedefs.Dir{
		{Path: filepath.Join(home, "go"), Score: 200},
		{Path: home, Score: 100},
		{Path: fixPath("/tmp/foo/bar/lorem/ipsum"), Score: 50},
	}
	startLocation(f.App, LocationSpec{Store: locationStore{
		storedDirs: dirs,
		chdir:      func(dir string) error { chdirCh <- dir; return errChdir },
	}})

	// Test UI.
	wantBuf := locationBuf(
		"",
		"200 "+filepath.Join("~", "go"),
		"100 ~",
		" 50 "+fixPath("/tmp/foo/bar/lorem/ipsum"))
	f.TTY.TestBuffer(t, wantBuf)

	// Test filtering.
	f.TTY.Inject(term.K('f'), term.K('o'))

	wantBuf = locationBuf(
		"fo",
		" 50 "+fixPath("/tmp/foo/bar/lorem/ipsum"))
	f.TTY.TestBuffer(t, wantBuf)

	// Test accepting.
	f.TTY.Inject(term.K(ui.Enter))
	// There should be no change to codearea after accepting.
	f.TestTTY(t /* nothing */)
	// Error from Chdir should be sent to notes.
	f.TestTTYNotes(t,
		"error: mock chdir error", Styles,
		"!!!!!!")
	// Chdir should be called.
	wantChdir := fixPath("/tmp/foo/bar/lorem/ipsum")
	select {
	case got := <-chdirCh:
		if got != wantChdir {
			t.Errorf("Chdir called with %s, want %s", got, wantChdir)
		}
	case <-time.After(testutil.Scaled(time.Second)):
		t.Errorf("Chdir not called")
	}
}

func TestLocation_Hidden(t *testing.T) {
	f := Setup()
	defer f.Stop()

	dirs := []storedefs.Dir{
		{Path: fixPath("/usr/bin"), Score: 200},
		{Path: fixPath("/usr"), Score: 100},
		{Path: fixPath("/tmp"), Score: 50},
	}
	startLocation(f.App, LocationSpec{
		Store:         locationStore{storedDirs: dirs},
		IterateHidden: func(f func(string)) { f(fixPath("/usr")) },
	})
	// Test UI.
	wantBuf := locationBuf(
		"",
		"200 "+fixPath("/usr/bin"),
		" 50 "+fixPath("/tmp"))
	f.TTY.TestBuffer(t, wantBuf)
}

func TestLocation_Pinned(t *testing.T) {
	f := Setup()
	defer f.Stop()

	dirs := []storedefs.Dir{
		{Path: fixPath("/usr/bin"), Score: 200},
		{Path: fixPath("/usr"), Score: 100},
		{Path: fixPath("/tmp"), Score: 50},
	}
	startLocation(f.App, LocationSpec{
		Store:         locationStore{storedDirs: dirs},
		IteratePinned: func(f func(string)) { f(fixPath("/home")); f(fixPath("/usr")) },
	})
	// Test UI.
	wantBuf := locationBuf(
		"",
		"  * "+fixPath("/home"),
		"  * "+fixPath("/usr"),
		"200 "+fixPath("/usr/bin"),
		" 50 "+fixPath("/tmp"))
	f.TTY.TestBuffer(t, wantBuf)
}

func TestLocation_HideWd(t *testing.T) {
	f := Setup()
	defer f.Stop()

	dirs := []storedefs.Dir{
		{Path: fixPath("/home"), Score: 200},
		{Path: fixPath("/tmp"), Score: 50},
	}
	startLocation(f.App, LocationSpec{Store: locationStore{storedDirs: dirs, wd: fixPath("/home")}})
	// Test UI.
	wantBuf := locationBuf(
		"",
		" 50 "+fixPath("/tmp"))
	f.TTY.TestBuffer(t, wantBuf)
}

func TestLocation_Workspace(t *testing.T) {
	f := Setup()
	defer f.Stop()

	chdir := ""
	dirs := []storedefs.Dir{
		{Path: fixPath("home/src"), Score: 200},
		{Path: fixPath("ws1/src"), Score: 150},
		{Path: fixPath("ws2/bin"), Score: 100},
		{Path: fixPath("/tmp"), Score: 50},
	}
	startLocation(f.App, LocationSpec{
		Store: locationStore{
			storedDirs: dirs,
			wd:         fixPath("/home/elf/bin"),
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

	wantBuf := locationBuf(
		"",
		"200 "+fixPath("home/src"),
		" 50 "+fixPath("/tmp"))
	f.TTY.TestBuffer(t, wantBuf)

	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t /* nothing */)
	wantChdir := fixPath("/home/elf/src")
	if chdir != wantChdir {
		t.Errorf("got chdir %q, want %q", chdir, wantChdir)
	}
}

func locationBuf(filter string, lines ...string) *term.Buffer {
	b := term.NewBufferBuilder(50).
		Newline(). // empty code area
		WriteStyled(modeLine(" LOCATION ", true)).
		Write(filter).SetDotHere()
	for i, line := range lines {
		b.Newline()
		if i == 0 {
			b.WriteStyled(ui.T(fmt.Sprintf("%-50s", line), ui.Inverse))
		} else {
			b.Write(line)
		}
	}
	return b.Buffer()
}

func fixPath(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}
	if path[0] == '/' {
		path = "C:" + path
	}
	return strings.ReplaceAll(path, "/", "\\")
}

func startLocation(app cli.App, spec LocationSpec) {
	w, err := NewLocation(app, spec)
	startMode(app, w, err)
}
