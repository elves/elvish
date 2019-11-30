package lastcmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
)

func setup() (cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := cli.ReadCodeAsync(app)
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}

type faultyStore struct{}

var mockError = errors.New("mock error")

func (s faultyStore) LastCmd() (histutil.Entry, error) {
	return histutil.Entry{}, mockError
}

func TestStart_NoStore(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{})
	wantNotesBuf := term.NewBufferBuilder(80).Write("no history store").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestStart_StoreError(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{Store: faultyStore{}})
	wantNotesBuf := term.NewBufferBuilder(80).
		Write("db error: mock error").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestStart_OK(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	store := histutil.NewMemoryStore()
	store.AddCmd(histutil.Entry{Text: "foo,bar,baz", Seq: 0})
	Start(app, Config{
		Store: store,
		Wordifier: func(cmd string) []string {
			return strings.Split(cmd, ",")
		},
	})

	// Test UI.
	wantBuf := term.NewBufferBuilder(80).
		// empty codearea
		Newline().
		// combobox codearea
		WriteStyled(ui.MakeText("LASTCMD",
			"bold", "lightgray", "bg-magenta")).
		Write(" ").
		SetDotHere().
		// first entry is selected
		Newline().WriteStyled(
		ui.MakeText("    foo,bar,baz"+strings.Repeat(" ", 65), "inverse")).
		// unselected entries
		Newline().Write("  0 foo").
		Newline().Write("  1 bar").
		Newline().Write("  2 baz").
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test negative filtering.
	ttyCtrl.Inject(term.K('-'))
	wantBuf = term.NewBufferBuilder(80).
		// empty codearea
		Newline().
		// combobox codearea
		WriteStyled(ui.MakeText("LASTCMD",
			"bold", "lightgray", "bg-magenta")).
		Write(" -").
		SetDotHere().
		// first entry is selected
		Newline().WriteStyled(
		ui.MakeText(" -3 foo"+strings.Repeat(" ", 73), "inverse")).
		// unselected entries
		Newline().Write(" -2 bar").
		Newline().Write(" -1 baz").
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test automatic submission.
	ttyCtrl.Inject(term.K('2')) // -2 bar
	wantBuf = term.NewBufferBuilder(80).
		Write("bar").SetDotHere().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test submission by Enter.
	app.CodeArea().MutateState(func(s *codearea.State) {
		*s = codearea.State{}
	})
	Start(app, Config{
		Store: store,
		Wordifier: func(cmd string) []string {
			return strings.Split(cmd, ",")
		},
	})
	ttyCtrl.Inject(term.K(ui.Enter))
	wantBuf = term.NewBufferBuilder(80).
		Write("foo,bar,baz").SetDotHere().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)

	// Default wordifier.
	app.CodeArea().MutateState(func(s *codearea.State) {
		*s = codearea.State{}
	})
	store.AddCmd(histutil.Entry{Text: "foo bar baz", Seq: 1})
	Start(app, Config{Store: store})
	ttyCtrl.Inject(term.K('0'))
	wantBuf = term.NewBufferBuilder(80).
		Write("foo").SetDotHere().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}
