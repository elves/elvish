package eval_test

import (
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
)

func TestExplicitBuiltinModule(t *testing.T) {
	TestWithSetup(t, func(ev *eval.Evaler) { ev.Args = vals.MakeList("a", "b") },
		That("all $args").Puts("a", "b"),
		// Regression test for #1414
		That("use builtin; all $builtin:args").Puts("a", "b"),
	)
}
