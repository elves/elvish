package edit

import (
	"testing"

	"github.com/elves/elvish/cli/term"
)

func TestMinibuf(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler, `edit:minibuf:start`)
	f.TestTTY(t,
		"~> \n",
		" MINIBUF  ", Styles,
		"********* ", term.DotHere,
	)
	feedInput(f.TTYCtrl, "edit:insert-at-dot put\n")
	f.TestTTY(t,
		"~> put", Styles,
		"   vvv", term.DotHere,
	)
}
