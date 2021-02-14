package bundled_test

import (
	"os"
	"testing"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/edit"
	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/mods/re"
	"src.elv.sh/pkg/eval/mods/str"
)

func TestEPM(t *testing.T) {
	// A smoke test to ensure that the epm module has no errors.

	TestWithSetup(t, func(ev *eval.Evaler) {
		// TODO: It shouldn't be necessary to do this setup manually. Instead,
		// there should be a function that initializes an Evaler with all the
		// standard modules.
		ev.AddModule("re", re.Ns)
		ev.AddModule("str", str.Ns)
	},
		That("use epm").DoesNothing(),
	)
}

func TestReadlineBinding(t *testing.T) {
	// A smoke test to ensure that the readline-binding module has no errors.

	TestWithSetup(t, func(ev *eval.Evaler) {
		ed := edit.NewEditor(cli.NewTTY(os.Stdin, os.Stderr), ev, nil)
		ev.AddBuiltin(eval.NsBuilder{}.AddNs("edit", ed.Ns()).Ns())
	},
		That("use readline-binding").DoesNothing(),
	)
}
