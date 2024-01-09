package readline_binding_test

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

	TestWithEvalerSetup(t, func(ev *eval.Evaler) {
		mods.AddTo(ev)
		ed := edit.NewEditor(cli.NewTTY(os.Stdin, os.Stderr), ev, nil)
		ev.ExtendBuiltin(eval.BuildNs().AddNs("edit", ed))
	},
		That("use readline-binding").DoesNothing(),
	)
}
