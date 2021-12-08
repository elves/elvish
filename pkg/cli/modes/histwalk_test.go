package modes

import (
	"testing"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

func TestHistWalk(t *testing.T) {
	f := Setup(WithSpec(func(spec *cli.AppSpec) {
		spec.CodeAreaState.Buffer = tk.CodeBuffer{Content: "ls", Dot: 2}
	}))
	defer f.Stop()

	f.App.Redraw()
	buf0 := f.MakeBuffer("ls", term.DotHere)
	f.TTY.TestBuffer(t, buf0)

	getCfg := func() HistwalkSpec {
		store := histutil.NewMemStore(
			// 0       1        2         3       4         5
			"echo", "ls -l", "echo a", "echo", "echo a", "ls -a")
		return HistwalkSpec{
			Store:  store,
			Prefix: "ls",
			Bindings: tk.MapBindings{
				term.K(ui.Up):        func(w tk.Widget) { w.(Histwalk).Prev() },
				term.K(ui.Down):      func(w tk.Widget) { w.(Histwalk).Next() },
				term.K('[', ui.Ctrl): func(tk.Widget) { f.App.PopAddon() },
			},
		}
	}

	startHistwalk(f.App, getCfg())
	buf5 := f.MakeBuffer(
		"ls -a", Styles,
		"  ___", term.DotHere, "\n",
		" HISTORY #5 ", Styles,
		"************",
	)
	f.TTY.TestBuffer(t, buf5)

	f.TTY.Inject(term.K(ui.Up))
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
	startHistwalk(f.App, getCfg())
	f.TTY.TestBuffer(t, buf5)
	f.TTY.Inject(term.K(' '))
	f.TestTTY(t, "ls -a ", term.DotHere)
}

func TestHistWalk_FocusedWidgetNotCodeArea(t *testing.T) {
	testFocusedWidgetNotCodeArea(t, func(app cli.App) error {
		store := histutil.NewMemStore("foo")
		_, err := NewHistwalk(app, HistwalkSpec{Store: store})
		return err
	})
}

func TestHistWalk_NoWalker(t *testing.T) {
	f := Setup()
	defer f.Stop()

	startHistwalk(f.App, HistwalkSpec{})
	f.TestTTYNotes(t,
		"error: no history store", Styles,
		"!!!!!!")
}

func TestHistWalk_NoMatch(t *testing.T) {
	f := Setup(WithSpec(func(spec *cli.AppSpec) {
		spec.CodeAreaState.Buffer = tk.CodeBuffer{Content: "ls", Dot: 2}
	}))
	defer f.Stop()

	f.App.Redraw()
	buf0 := f.MakeBuffer("ls", term.DotHere)
	f.TTY.TestBuffer(t, buf0)

	store := histutil.NewMemStore("echo 1", "echo 2")
	cfg := HistwalkSpec{Store: store, Prefix: "ls"}
	startHistwalk(f.App, cfg)
	// Test that an error message has been written to the notes buffer.
	f.TestTTYNotes(t,
		"error: end of history", Styles,
		"!!!!!!")
	// Test that buffer has not changed - histwalk addon is not active.
	f.TTY.TestBuffer(t, buf0)
}

func TestHistWalk_FallbackHandler(t *testing.T) {
	f := Setup()
	defer f.Stop()

	store := histutil.NewMemStore("ls")
	startHistwalk(f.App, HistwalkSpec{Store: store, Prefix: ""})
	f.TestTTY(t,
		"ls", Styles,
		"__", term.DotHere, "\n",
		" HISTORY #0 ", Styles,
		"************",
	)

	f.TTY.Inject(term.K(ui.Backspace))
	f.TestTTY(t, "l", term.DotHere)
}

func startHistwalk(app cli.App, cfg HistwalkSpec) {
	w, err := NewHistwalk(app, cfg)
	if err != nil {
		app.Notify(ErrorText(err))
		return
	}
	app.PushAddon(w)
	app.Redraw()
}
