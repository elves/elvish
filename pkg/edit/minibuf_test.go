package edit

import (
	"testing"

	"src.elv.sh/pkg/cli/term"
)

func TestMinibuf(t *testing.T) {
	f := setup(t)

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
