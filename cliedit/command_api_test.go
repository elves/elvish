package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

func TestCommandMode(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler, `edit:insert:binding[Ctrl-'['] = $edit:command:start~`)
	feedInput(f.TTYCtrl, "echo")
	f.TTYCtrl.Inject(term.K('[', ui.Ctrl))
	f.TTYCtrl.TestBuffer(t,
		bb().WritePlain("~> ").WriteString("echo", "green").SetDotHere().
			Newline().WriteStyled(layout.ModeLine(" COMMAND ", false)).Buffer())

	f.TTYCtrl.Inject(term.K('b'))
	f.TTYCtrl.TestBuffer(t,
		bb().WritePlain("~> ").SetDotHere().WriteString("echo", "green").
			Newline().WriteStyled(layout.ModeLine(" COMMAND ", false)).Buffer())
}
