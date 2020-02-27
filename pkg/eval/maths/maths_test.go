package maths

import (
	"math"
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

		That(`math:log 2.71828182845904523536028747135266249775724709369995`).Puts(1.0),
		That(`math:log 1`).Puts(0.0),
		That(`math:log 0`).Puts(math.Inf(-1)),
		That(`math:log -1`).Puts(math.NaN()),

		That(`math:log10 10.0`).Puts(1.0),
		That(`math:log10 100.0`).Puts(2.0),
		That(`math:log10 1`).Puts(0.0),
		That(`math:log10 0`).Puts(math.Inf(-1)),
		That(`math:log10 -1`).Puts(math.NaN()),

		That(`math:log2 8`).Puts(3.0),
		That(`math:log2 1024.0`).Puts(10.0),
		That(`math:log2 1`).Puts(0.0),
		That(`math:log2 0`).Puts(math.Inf(-1)),
		That(`math:log2 -1`).Puts(math.NaN()),
	)
}
