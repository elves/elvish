package histwalk

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var styles = map[rune]string{
	'-': "underlined",
}

func TestHistWalk(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	db := &histutil.TestDB{
		//                 0       1        2         3        4         5
		AllCmds: []string{"echo", "ls -l", "echo a", "ls -a", "echo a", "ls -a"},
	}
	walker := histutil.NewWalker(db, -1, nil, "ls")

	app.CodeArea().MutateCodeAreaState(func(s *codearea.State) {
		s.CodeBuffer = codearea.CodeBuffer{Content: "ls", Dot: 2}
	})

	app.Redraw()
	buf0 := ui.NewBufferBuilder(40).WritePlain("ls").SetDotToCursor().Buffer()
	ttyCtrl.TestBuffer(t, buf0)

	Start(app, Config{
		Walker: walker,
		Binding: el.MapHandler{
			term.K(ui.Up):        func() { Prev(app) },
			term.K(ui.Down):      func() { Next(app) },
			term.K('[', ui.Ctrl): func() { Close(app) },
		},
	})
	buf5 := makeBuf(
		styled.MarkLines(
			"ls -a", styles,
			"  ---",
		),
		" HISTORY #5 ")
	ttyCtrl.TestBuffer(t, buf5)

	ttyCtrl.Inject(term.K(ui.Up))
	// Skips item #3 as it is a duplicate.
	buf1 := makeBuf(
		styled.MarkLines(
			"ls -l", styles,
			"  ---",
		),
		" HISTORY #1 ")
	ttyCtrl.TestBuffer(t, buf1)

	ttyCtrl.Inject(term.K(ui.Down))
	ttyCtrl.TestBuffer(t, buf5)

	ttyCtrl.Inject(term.K('[', ui.Ctrl))
	ttyCtrl.TestBuffer(t, buf0)
}

func makeBuf(codeArea styled.Text, modeline string) *ui.Buffer {
	return ui.NewBufferBuilder(40).
		WriteStyled(codeArea).
		Newline().SetDotToCursor().
		WriteStyled(layout.ModeLine(modeline, false)).
		Buffer()
}

func setup() (cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(24, 40)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := app.ReadCodeAsync()
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}
