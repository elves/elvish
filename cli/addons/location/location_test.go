package location

import (
	"errors"
	"strings"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
)

func setup() (cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	// Use a smaller TTY size to make diffs easier to see.
	ttyCtrl.SetSize(20, 50)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := app.ReadCodeAsync()
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}

type testStore struct {
	dirs  func() ([]storedefs.Dir, error)
	chdir func(dir string) error
}

func (ts testStore) Dirs() ([]storedefs.Dir, error) {
	if ts.dirs == nil {
		return nil, nil
	}
	return ts.dirs()
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

	wantNotesBuf := ui.NewBufferBuilder(50).
		WritePlain("no dir history store").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestStart_StoreError(t *testing.T) {
	app, ttyCtrl, teardown := setup()
	defer teardown()

	mockError := errors.New("mock error")
	Start(app, Config{Store: testStore{dirs: func() ([]storedefs.Dir, error) {
		return nil, mockError
	}}})

	wantNotesBuf := ui.NewBufferBuilder(50).
		WritePlain("db error: mock error").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestStart_OK(t *testing.T) {
	app, ttyCtrl, teardown := setup()
	defer teardown()

	errChdir := errors.New("mock chdir error")
	chdirCh := make(chan string, 100)
	dirs := []storedefs.Dir{
		{Path: "/home/elf/go", Score: 200},
		{Path: "/home/elf", Score: 100},
		{Path: "/tmp", Score: 50},
	}
	Start(app, Config{Store: testStore{
		dirs:  func() ([]storedefs.Dir, error) { return dirs, nil },
		chdir: func(dir string) error { chdirCh <- dir; return errChdir },
	}})

	// Test UI.
	wantBuf := ui.NewBufferBuilder(50).
		// empty codearea
		Newline().
		// combobox prompt
		WriteStyled(
			styled.MakeText("LOCATION", "bold", "lightgray", "bg-magenta")).
		WritePlain(" ").SetDotToCursor().
		// items sorted by score in descending order; first selected
		Newline().
		WriteStyled(styled.MakeText(
			"200 /home/elf/go"+strings.Repeat(" ", 34), "inverse")).
		Newline().WritePlain("100 /home/elf").
		Newline().WritePlain(" 50 /tmp").
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test filtering.
	ttyCtrl.Inject(term.K('t'), term.K('m'), term.K('p'))
	wantBuf = ui.NewBufferBuilder(50).
		// empty codearea
		Newline().
		// combobox prompt
		WriteStyled(
			styled.MakeText("LOCATION", "bold", "lightgray", "bg-magenta")).
		WritePlain(" tmp").SetDotToCursor().
		// items sorted by score in descending order; first selected
		Newline().
		WriteStyled(styled.MakeText(
			" 50 /tmp"+strings.Repeat(" ", 42), "inverse")).
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test accepting.
	ttyCtrl.Inject(term.K(ui.Enter))
	// There should be no change to codearea after accepting.
	wantBuf = ui.NewBufferBuilder(50).Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
	// Error from Chdir should be sent to notes.
	wantNotesBuf := ui.NewBufferBuilder(50).WritePlain("mock chdir error").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
	// Chdir should be called.
	if got := <-chdirCh; got != "/tmp" {
		t.Errorf("Chdir called with %s, want /tmp", got)
	}
}
