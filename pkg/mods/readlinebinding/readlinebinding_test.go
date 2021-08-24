package readlinebinding_test

import (
	"os"
	"testing"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/edit"
	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods"
)

func TestReadlineBinding(t *testing.T) {
	// A smoke test to ensure that the readline-binding module has no errors.

	TestWithSetup(t, func(ev *eval.Evaler) {
		mods.AddTo(ev)
		ed := edit.NewEditor(cli.NewTTY(os.Stdin, os.Stderr), ev, nil)
		ev.AddBuiltin(eval.NsBuilder{}.AddNs("edit", ed.Ns()).Ns())
	},
		That("use readline-binding").DoesNothing(),
	)
}
