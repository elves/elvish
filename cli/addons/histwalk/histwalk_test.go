package histwalk

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
)

var styles = map[rune]ui.Styling{
	'-': ui.Underlined,
}

func TestHistWalk(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	app.CodeArea().MutateState(func(s *codearea.State) {
		s.Buffer = codearea.Buffer{Content: "ls", Dot: 2}
	})

	app.Redraw()
	buf0 := term.NewBufferBuilder(40).Write("ls").SetDotHere().Buffer()
	ttyCtrl.TestBuffer(t, buf0)

	getCfg := func() Config {
		db := &histutil.TestDB{
			//                 0       1        2         3        4         5
			AllCmds: []string{"echo", "ls -l", "echo a", "ls -a", "echo a", "ls -a"},
		}
		return Config{
			Walker: histutil.NewWalker(db, -1, nil, "ls"),
			Binding: el.MapHandler{
				term.K(ui.Up):        func() { Prev(app) },
				term.K(ui.Down):      func() { Next(app) },
				term.K('[', ui.Ctrl): func() { Close(app) },
				term.K(ui.Enter):     func() { Accept(app) },
			},
		}
	}

	Start(app, getCfg())
	buf5 := makeBuf(
		ui.MarkLines(
			"ls -a", styles,
			"  ---",
		),
		" HISTORY #5 ")
	ttyCtrl.TestBuffer(t, buf5)

	ttyCtrl.Inject(term.K(ui.Up))
	// Skips item #3 as it is a duplicate.
	buf1 := makeBuf(
		ui.MarkLines(
			"ls -l", styles,
			"  ---",
		),
		" HISTORY #1 ")
	ttyCtrl.TestBuffer(t, buf1)

	ttyCtrl.Inject(term.K(ui.Down))
	ttyCtrl.TestBuffer(t, buf5)

	ttyCtrl.Inject(term.K('[', ui.Ctrl))
	ttyCtrl.TestBuffer(t, buf0)

	// Start over and accept.
	Start(app, getCfg())
	ttyCtrl.TestBuffer(t, buf5)
	ttyCtrl.Inject(term.K(ui.Enter))
	bufAccepted := term.NewBufferBuilder(40).Write("ls -a").SetDotHere().Buffer()
	ttyCtrl.TestBuffer(t, bufAccepted)
}

func TestHistWalk_NoWalker(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{})
	ttyCtrl.TestNotesBuffer(t, bb().Write("no history walker").Buffer())
}

func TestHistWalk_FallbackHandler(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	db := &histutil.TestDB{
		AllCmds: []string{"ls"},
	}
	Start(app, Config{
		Walker: histutil.NewWalker(db, -1, nil, ""),
	})
	wantBuf := makeBuf(
		ui.MarkLines(
			"ls", styles,
			"--",
		),
		" HISTORY #0 ")
	ttyCtrl.TestBuffer(t, wantBuf)

	ttyCtrl.Inject(term.K(ui.Backspace))
	ttyCtrl.TestBuffer(t, bb().Write("l").SetDotHere().Buffer())
}

func makeBuf(codeArea ui.Text, modeline string) *term.Buffer {
	return bb().
		WriteStyled(codeArea).SetDotHere().
		Newline().
		WriteStyled(layout.ModeLine(modeline, false)).
		Buffer()
}

func bb() *term.BufferBuilder {
	return term.NewBufferBuilder(40)
}

func setup() (cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(24, 40)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := cli.ReadCodeAsync(app)
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}
