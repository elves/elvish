package edit

import (
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
)

func TestInstantAddon_ValueOutput(t *testing.T) {
	f := setup(t)

	evals(f.Evaler, "edit:-instant:start")
	f.TestTTY(t,
		"~> ", term.DotHere, "\n",
		" INSTANT \n", Styles,
		"*********",
	)

	feedInput(f.TTYCtrl, "+")
	f.TestTTY(t,
		"~> +", Styles,
		"   v", term.DotHere, "\n",
		" INSTANT \n", Styles,
		"*********",
		"▶ 0",
	)

	feedInput(f.TTYCtrl, " 1 2")
	f.TestTTY(t,
		"~> + 1 2", Styles,
		"   v    ", term.DotHere, "\n",
		" INSTANT \n", Styles,
		"*********",
		"▶ 3",
	)
}

func TestInstantAddon_ByteOutput(t *testing.T) {
	f := setup(t)

	// We don't want to trigger the evaluation of "e", "ec", and "ech", so we
	// start with a non-empty code buffer.
	f.SetCodeBuffer(tk.CodeBuffer{Content: "echo ", Dot: 5})
	evals(f.Evaler, "edit:-instant:start")
	f.TestTTY(t,
		"~> echo ", Styles,
		"   vvvv ", term.DotHere, "\n",
		" INSTANT \n", Styles,
		"*********",
	)

	feedInput(f.TTYCtrl, "hello")
	f.TestTTY(t,
		"~> echo hello", Styles,
		"   vvvv      ", term.DotHere, "\n",
		" INSTANT \n", Styles,
		"*********",
		"hello",
	)
}
