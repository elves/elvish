package completion

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/ui"
)

func TestStart(t *testing.T) {
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(24, 40)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := cli.ReadCodeAsync(app)
	defer func() {
		app.CommitEOF()
		<-codeCh
	}()

	cfg := Config{
		Name:    "WORD",
		Replace: diag.Ranging{From: 0, To: 0},
		Items: []Item{
			{ToShow: "foo", ToInsert: "foo"},
			{ToShow: "foo bar", ToInsert: "'foo bar'",
				ShowStyle: styled.Style{Foreground: "blue"}},
		},
	}
	Start(app, cfg)

	// Test that the completion combobox is shown correctly.
	wantBufStarted := term.NewBufferBuilder(40).
		Write("foo", "underlined"). // code area
		Newline().
		WriteStyled(layout.ModeLine("COMPLETING WORD", true)).
		SetDotHere().
		Newline().Write("foo", "inverse"). // Selected entry
		Write("  ").
		Write("foo bar", "blue").
		Buffer()
	ttyCtrl.TestBuffer(t, wantBufStarted)

	// Test the OnFilter handler.
	ttyCtrl.Inject(term.K('b'), term.K('a'))
	wantBufFiltering := term.NewBufferBuilder(40).
		Write("'foo bar'", "underlined"). // code area
		Newline().
		WriteStyled(layout.ModeLine("COMPLETING WORD", true)).
		Write("ba").SetDotHere().
		Newline().Write("foo bar", "blue", "inverse"). // Selected entry
		Buffer()
	ttyCtrl.TestBuffer(t, wantBufFiltering)

	// Test the OnAccept handler.
	ttyCtrl.Inject(term.K(ui.Enter))
	wantBufAccepted := term.NewBufferBuilder(40).
		Write("'foo bar'").SetDotHere().Buffer()
	ttyCtrl.TestBuffer(t, wantBufAccepted)

	// Test Close first we need to start over.
	app.CodeArea().MutateState(
		func(s *codearea.State) { *s = codearea.State{} })
	Start(app, cfg)
	ttyCtrl.TestBuffer(t, wantBufStarted)
	Close(app)
	wantBufClosed := term.NewBufferBuilder(40).Buffer()
	ttyCtrl.TestBuffer(t, wantBufClosed)
}
