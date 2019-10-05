package completion

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

func TestStart(t *testing.T) {
	tty, ttyCtrl := cli.NewFakeTTY()
	app := cli.NewApp(tty)
	codeCh, _ := app.ReadCodeAsync()
	defer func() {
		app.CommitEOF()
		<-codeCh
	}()

	Start(app, Config{
		Type: "WORD",
		Candidates: []Candidate{
			{ToShow: "foo", ToInsert: "foo"},
			{ToShow: "foo bar", ToInsert: "'foo bar'"},
		},
	})

	// Test that the completion combobox is shown correctly.
	wantBuf := ui.NewBufferBuilder(80).
		Newline(). // empty code area
		WriteStyled(styled.MakeText("COMPLETING WORD",
			"bold", "lightgray", "bg-magenta")). // codearea of the combobox
		WritePlain(" ").
		SetDotToCursor().
		Newline().WriteStyled(styled.MakeText("foo", "inverse")). // Selected entry
		WritePlain("  foo bar").
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test the OnFilter handler.
	ttyCtrl.Inject(term.K('b'), term.K('a'))
	wantBuf = ui.NewBufferBuilder(80).
		Newline(). // empty code area
		WriteStyled(styled.MakeText("COMPLETING WORD",
			"bold", "lightgray", "bg-magenta")). // codearea of the combobox
		WritePlain(" ba").
		SetDotToCursor().
		Newline().WriteStyled(styled.MakeText("foo bar", "inverse")). // Selected entry
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)

	// Test the OnAccept handler.
	ttyCtrl.Inject(term.K(ui.Enter))
	wantBuf = ui.NewBufferBuilder(80).
		WritePlain("'foo bar'").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}
