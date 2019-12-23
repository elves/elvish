package histwalk

import (
	"testing"

	"github.com/elves/elvish/pkg/cli"
	. "github.com/elves/elvish/pkg/cli/apptest"
	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/el/codearea"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
)

func TestHistWalk(t *testing.T) {
	f := Setup()
	defer f.Stop()

	cli.SetCodeBuffer(f.App, codearea.Buffer{Content: "ls", Dot: 2})
	f.App.Redraw()
	buf0 := f.MakeBuffer("ls", term.DotHere)
	f.TTY.TestBuffer(t, buf0)

	getCfg := func() Config {
		db := &histutil.TestDB{
			//                 0       1        2         3        4         5
			AllCmds: []string{"echo", "ls -l", "echo a", "ls -a", "echo a", "ls -a"},
		}
		return Config{
			Walker: histutil.NewWalker(db, -1, nil, "ls"),
			Binding: el.MapHandler{
				term.K(ui.Up):        func() { Prev(f.App) },
				term.K(ui.Down):      func() { Next(f.App) },
				term.K('[', ui.Ctrl): func() { Close(f.App) },
				term.K(ui.Enter):     func() { Accept(f.App) },
			},
		}
	}

	Start(f.App, getCfg())
	buf5 := f.MakeBuffer(
		"ls -a", Styles,
		"  ___", term.DotHere, "\n",
		" HISTORY #5 ", Styles,
		"************",
	)
	f.TTY.TestBuffer(t, buf5)

	f.TTY.Inject(term.K(ui.Up))
	// Skips item #3 as it is a duplicate.
	buf1 := f.MakeBuffer(
		"ls -l", Styles,
		"  ___", term.DotHere, "\n",
		" HISTORY #1 ", Styles,
		"************",
	)
	f.TTY.TestBuffer(t, buf1)

	f.TTY.Inject(term.K(ui.Down))
	f.TTY.TestBuffer(t, buf5)

	f.TTY.Inject(term.K('[', ui.Ctrl))
	f.TTY.TestBuffer(t, buf0)

	// Start over and accept.
	Start(f.App, getCfg())
	f.TTY.TestBuffer(t, buf5)
	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "ls -a", term.DotHere)
}

func TestHistWalk_NoWalker(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{})
	f.TestTTYNotes(t, "no history walker")
}

func TestHistWalk_NoMatch(t *testing.T) {
	f := Setup()
	defer f.Stop()

	cli.SetCodeBuffer(f.App, codearea.Buffer{Content: "ls", Dot: 2})
	f.App.Redraw()
	buf0 := f.MakeBuffer("ls", term.DotHere)
	f.TTY.TestBuffer(t, buf0)

	db := &histutil.TestDB{AllCmds: []string{"echo 1", "echo 2"}}
	cfg := Config{Walker: histutil.NewWalker(db, -1, nil, "ls")}
	Start(f.App, cfg)
	// Test that an error message has been written to the notes buffer.
	f.TestTTYNotes(t, "end of history")
	// Test that buffer has not changed - histwalk addon is not active.
	f.TTY.TestBuffer(t, buf0)
}

func TestHistWalk_FallbackHandler(t *testing.T) {
	f := Setup()
	defer f.Stop()

	db := &histutil.TestDB{
		AllCmds: []string{"ls"},
	}
	Start(f.App, Config{
		Walker: histutil.NewWalker(db, -1, nil, ""),
	})
	f.TestTTY(t,
		"ls", Styles,
		"__", term.DotHere, "\n",
		" HISTORY #0 ", Styles,
		"************",
	)

	f.TTY.Inject(term.K(ui.Backspace))
	f.TestTTY(t, "l", term.DotHere)
}
