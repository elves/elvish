package histlist

import (
	"errors"
	"fmt"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

func TestStart_NoStore(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{})
	wantNotesBuf := bb().WritePlain("no history store").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

type faultyStore struct{}

var mockError = errors.New("mock error")

func (s faultyStore) AllCmds() ([]histutil.Entry, error) { return nil, mockError }

func TestStart_StoreError(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{Store: faultyStore{}})
	wantNotesBuf := bb().WritePlain("db error: mock error").Buffer()
	ttyCtrl.TestNotesBuffer(t, wantNotesBuf)
}

func TestStart_OK(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	store := histutil.NewMemoryStore()
	store.AddCmd(histutil.Entry{Text: "foo", Seq: 0})
	store.AddCmd(histutil.Entry{Text: "bar", Seq: 1})
	store.AddCmd(histutil.Entry{Text: "baz", Seq: 2})
	Start(app, Config{Store: store})

	// Test UI.
	ttyCtrl.TestBuffer(t,
		makeListingBuf(
			"",
			"   0 foo",
			"   1 bar",
			"   2 baz"))

	// Test filtering.
	ttyCtrl.Inject(term.K('b'))
	ttyCtrl.TestBuffer(t,
		makeListingBuf(
			"b",
			"   1 bar",
			"   2 baz"))

	// Test accepting.
	ttyCtrl.Inject(term.K(ui.Enter))
	wantBufAccepted1 := bb().
		// codearea now contains selected entry
		WritePlain("baz").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufAccepted1)

	// Test accepting when there is already some text.
	store.AddCmd(histutil.Entry{Text: "baz2"})
	Start(app, Config{Store: store})
	ttyCtrl.Inject(term.K(ui.Enter))
	wantBufAccepted2 := bb().
		WritePlain("baz").
		// codearea now contains newly inserted entry on a separate line
		Newline().WritePlain("baz2").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufAccepted2)
}

func TestStart_Dedup(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	store := histutil.NewMemoryStore()
	store.AddCmd(histutil.Entry{Text: "ls", Seq: 0})
	store.AddCmd(histutil.Entry{Text: "echo", Seq: 1})
	store.AddCmd(histutil.Entry{Text: "ls", Seq: 2})

	// No dedup
	Start(app, Config{Store: store, Dedup: func() bool { return false }})
	ttyCtrl.TestBuffer(t,
		makeListingBuf(
			"",
			"   0 ls",
			"   1 echo",
			"   2 ls"))
	app.MutateState(func(s *cli.State) { s.Addon = nil })

	// With dedup
	Start(app, Config{Store: store, Dedup: func() bool { return true }})
	ttyCtrl.TestBuffer(t,
		makeListingBuf(
			"",
			"   1 echo",
			"   2 ls"))
}

func TestStart_CaseSensitive(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	store := histutil.NewMemoryStore()
	store.AddCmd(histutil.Entry{Text: "ls", Seq: 0})
	store.AddCmd(histutil.Entry{Text: "LS", Seq: 1})

	// Case sensitive
	Start(app, Config{Store: store, CaseSensitive: func() bool { return true }})
	ttyCtrl.Inject(term.K('l'))
	ttyCtrl.TestBuffer(t,
		makeListingBuf(
			"l",
			"   0 ls"))
	app.MutateState(func(s *cli.State) { s.Addon = nil })

	// Case insensitive
	Start(app, Config{Store: store, CaseSensitive: func() bool { return false }})
	ttyCtrl.Inject(term.K('l'))
	ttyCtrl.TestBuffer(t,
		makeListingBuf(
			"l",
			"   0 ls",
			"   1 LS"))
}

func setup() (cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(24, 50)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := cli.ReadCodeAsync(app)
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}

func bb() *ui.BufferBuilder { return ui.NewBufferBuilder(50) }

func makeListingBuf(filter string, lines ...string) *ui.Buffer {
	b := bb().Newline().
		WriteStyled(layout.ModeLine("HISTLIST", true)).
		WritePlain(filter).SetDotToCursor()
	for i, line := range lines {
		b.Newline()
		if i < len(lines)-1 {
			b.WritePlain(line)
		} else {
			b.WriteStyled(
				styled.MakeText(fmt.Sprintf("%-50s", line), "inverse"))
		}
	}
	return b.Buffer()
}
