package maths

import (
	"testing"

	"github.com/elves/elvish/pkg/eval"
)

func TestMath(t *testing.T) {
	setup := func(ev *eval.Evaler) { ev.Builtin.AddNs("math", Ns) }
	That := eval.That
	eval.TestWithSetup(t, setup,
		That(`math:abs 2.1`).Puts(2.1),
		That(`math:abs -2.1`).Puts(2.1),

		That(`math:ceil 2.1`).Puts(3.0),
		That(`math:ceil -2.1`).Puts(-2.0),

		That(`math:floor 2.1`).Puts(2.0),
		That(`math:floor -2.1`).Puts(-3.0),

		That(`math:round 2.1`).Puts(2.0),
		That(`math:round -2.1`).Puts(-2.0),
		That(`math:round 2.5`).Puts(3.0),
		That(`math:round -2.5`).Puts(-3.0),
	)
}
