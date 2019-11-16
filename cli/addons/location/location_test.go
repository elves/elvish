package location

import (
	"errors"
	"path/filepath"
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

func TestStart_NoStore(t *testing.T) {
	app, ttyCtrl, teardown := setup()
	defer teardown()

	Start(app, Config{})

	wantNotesBuf := bb().WritePlain("no dir history store").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestStart_StoreError(t *testing.T) {
	app, ttyCtrl, teardown := setup()
	defer teardown()

	Start(app, Config{Store: testStore{dirsError: errors.New("ERROR")}})

	wantNotesBuf := bb().WritePlain("db error: ERROR").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestStart_Hidden(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	dirs := []storedefs.Dir{
		{Path: "/usr/bin", Score: 200},
		{Path: "/usr", Score: 100},
		{Path: "/tmp", Score: 50},
	}
	Start(app, Config{
		Store:         testStore{storedDirs: dirs},
		IterateHidden: func(f func(string)) { f("/usr") },
	})
	// Test UI.
	wantBuf := listingBuf(
		"",
		"200 /usr/bin", "<- selected",
		" 50 /tmp")
	ttyCtrl.TestBuffer(t, wantBuf)
}

func TestStart_Pinned(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	dirs := []storedefs.Dir{
		{Path: "/usr/bin", Score: 200},
		{Path: "/usr", Score: 100},
		{Path: "/tmp", Score: 50},
	}
	Start(app, Config{
		Store:         testStore{storedDirs: dirs},
		IteratePinned: func(f func(string)) { f("/home"); f("/usr") },
	})
	// Test UI.
	wantBuf := listingBuf(
		"",
		"  * /home", "<- selected",
		"  * /usr",
		"200 /usr/bin",
		" 50 /tmp")
	ttyCtrl.TestBuffer(t, wantBuf)
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
		{Path: "/tmp", Score: 50},
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
		" 50 /tmp")
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test filtering.
	ttyCtrl.Inject(term.K('t'), term.K('m'), term.K('p'))

	wantBuf = listingBuf(
		"tmp",
		" 50 /tmp", "<- selected")
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test accepting.
	ttyCtrl.Inject(term.K(ui.Enter))
	// There should be no change to codearea after accepting.
	wantBuf = bb().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
	// Error from Chdir should be sent to notes.
	wantNotesBuf := bb().WritePlain("mock chdir error").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
	// Chdir should be called.
	if got := <-chdirCh; got != "/tmp" {
		t.Errorf("Chdir called with %s, want /tmp", got)
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
