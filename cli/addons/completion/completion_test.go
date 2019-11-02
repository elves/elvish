package completion

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

func TestStart(t *testing.T) {
	tty, ttyCtrl := cli.NewFakeTTY()
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := app.ReadCodeAsync()
	defer func() {
		app.CommitEOF()
		<-codeCh
	}()

	cfg := Config{
		Name:    "WORD",
		Replace: diag.Ranging{From: 0, To: 0},
		Items: []Item{
			{ToShow: "foo", ToInsert: "foo"},
			{ToShow: "foo bar", ToInsert: "'foo bar'"},
		},
	}
	Start(app, cfg)

	// Test that the completion combobox is shown correctly.
	wantBufStarted := ui.NewBufferBuilder(80).
		WriteStyled(styled.MakeText("foo", "underlined")). // code area
		Newline().
		WriteStyled(layout.ModeLine("COMPLETING WORD", true)).
		SetDotToCursor().
		Newline().WriteStyled(styled.MakeText("foo", "inverse")). // Selected entry
		WritePlain("  foo bar").
		Buffer()
	ttyCtrl.TestBuffer(t, wantBufStarted)

	// Test the OnFilter handler.
	ttyCtrl.Inject(term.K('b'), term.K('a'))
	wantBufFiltering := ui.NewBufferBuilder(80).
		WriteStyled(styled.MakeText("'foo bar'", "underlined")). // code area
		Newline().
		WriteStyled(layout.ModeLine("COMPLETING WORD", true)).
		WritePlain("ba").SetDotToCursor().
		Newline().WriteStyled(styled.MakeText("foo bar", "inverse")). // Selected entry
		Buffer()
	ttyCtrl.TestBuffer(t, wantBufFiltering)

	// Test the OnAccept handler.
	ttyCtrl.Inject(term.K(ui.Enter))
	wantBufAccepted := ui.NewBufferBuilder(80).
		WritePlain("'foo bar'").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBufAccepted)

	// Test Close first we need to start over.
	app.CodeArea().MutateState(
		func(s *codearea.State) { *s = codearea.State{} })
	Start(app, cfg)
	ttyCtrl.TestBuffer(t, wantBufStarted)
	Close(app)
	wantBufClosed := ui.NewBufferBuilder(80).Buffer()
	ttyCtrl.TestBuffer(t, wantBufClosed)
}
