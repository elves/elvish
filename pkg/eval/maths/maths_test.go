package maths

import (
	"fmt"
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
		That(`math:round (float64 Inf)`).Puts(math.Inf(1)),
		That(`math:round (float64 NaN)`).Puts(math.NaN()),

		That(`math:round-to-even 2.1`).Puts(2.0),
		That(`math:round-to-even -2.1`).Puts(-2.0),
		That(`math:round-to-even 2.5`).Puts(2.0),
		That(`math:round-to-even -2.5`).Puts(-2.0),
		That(`math:round-to-even (float64 Inf)`).Puts(math.Inf(1)),
		That(`math:round-to-even (float64 NaN)`).Puts(math.NaN()),

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

		That(`math:cos 0`).Puts(1.0),
		That(`math:cos 1`).Puts(0.5403023058681397174),
		That(fmt.Sprintf(`math:cos %g`, math.Pi)).Puts(-1.0),

		That(`math:sin 0`).Puts(0.0),
		That(`math:sin 1`).Puts(0.84147098480789650665),
		That(fmt.Sprintf(`math:sin %g`, math.Pi)).Puts(0.0),

		That(`math:tan 0`).Puts(0.0),
		That(`math:tan 1`).Puts(1.5574077246549023),
		That(fmt.Sprintf(`math:tan %g`, math.Pi)).Puts(0.0),
	)
}
