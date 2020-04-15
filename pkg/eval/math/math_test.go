package math

import (
	"math"
	"testing"

	"github.com/elves/elvish/pkg/eval"
)

var That = eval.That

func TestMath(t *testing.T) {
	setup := func(ev *eval.Evaler) { ev.Builtin.AddNs("math", Ns) }
	eval.TestWithSetup(t, setup,
		That(`put $math:pi`).Puts(math.Pi),
		That(`put $math:e`).Puts(math.E),

		That(`math:abs 2.1`).Puts(2.1),
		That(`math:abs -2.1`).Puts(2.1),

		That(`math:ceil 2.1`).Puts(3.0),
		That(`math:ceil -2.1`).Puts(-2.0),

		That(`math:floor 2.1`).Puts(2.0),
		That(`math:floor -2.1`).Puts(-3.0),

		That(`math:is-inf 1.3`).Puts(false),
		That(`math:is-inf &sign=0 inf`).Puts(true),
		That(`math:is-inf &sign=1 inf`).Puts(true),
		That(`math:is-inf &sign=-1 -inf`).Puts(true),
		That(`math:is-inf &sign=1 -inf`).Puts(false),
		That(`math:is-inf -inf`).Puts(true),
		That(`math:is-inf nan`).Puts(false),
		That(`math:is-inf &sign=0 (float64 inf)`).Puts(true),
		That(`math:is-inf &sign=1 (float64 inf)`).Puts(true),
		That(`math:is-inf &sign=-1 (float64 -inf)`).Puts(true),
		That(`math:is-inf &sign=1 (float64 -inf)`).Puts(false),
		That(`math:is-inf (float64 -inf)`).Puts(true),
		That(`math:is-inf (float64 nan)`).Puts(false),
		That(`math:is-inf (float64 1.3)`).Puts(false),

		That(`math:is-nan 1.3`).Puts(false),
		That(`math:is-nan inf`).Puts(false),
		That(`math:is-nan nan`).Puts(true),
		That(`math:is-nan (float64 inf)`).Puts(false),
		That(`math:is-nan (float64 nan)`).Puts(true),

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

		That(`math:trunc 2.1`).Puts(2.0),
		That(`math:trunc -2.1`).Puts(-2.0),
		That(`math:trunc 2.5`).Puts(2.0),
		That(`math:trunc -2.5`).Puts(-2.0),
		That(`math:trunc (float64 Inf)`).Puts(math.Inf(1)),
		That(`math:trunc (float64 NaN)`).Puts(math.NaN()),

		That(`math:log $math:e`).Puts(1.0),
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
		That(`math:cos 1`).Puts(math.Cos(1.0)),
		That(`math:cos $math:pi`).Puts(-1.0),

		That(`math:cosh 0`).Puts(1.0),
		That(`math:cosh inf`).Puts(math.Inf(1)),
		That(`math:cosh nan`).Puts(math.NaN()),

		That(`math:sin 0`).Puts(0.0),
		That(`math:sin 1`).Puts(math.Sin(1.0)),
		That(`math:sin $math:pi`).Puts(math.Sin(math.Pi)),

		That(`math:sinh 0`).Puts(0.0),
		That(`math:sinh inf`).Puts(math.Inf(1)),
		That(`math:sinh nan`).Puts(math.NaN()),

		That(`math:tan 0`).Puts(0.0),
		That(`math:tan 1`).Puts(math.Tan(1.0)),
		That(`math:tan $math:pi`).Puts(math.Tan(math.Pi)),

		That(`math:tanh 0`).Puts(0.0),
		That(`math:tanh inf`).Puts(1.0),
		That(`math:tanh nan`).Puts(math.NaN()),

		// This block of tests isn't strictly speaking necessary. But it helps
		// ensure that we're not just confirming Go statements such as
		//    math.Tan(math.Pi) == math.Tan(math.Pi)
		// are true. The ops that should return a zero value do not actually
		// do so. Which illustrates why an approximate match is needed.
		That(`math:cos 1`).Puts(eval.Approximately{0.5403023058681397174}),
		That(`math:sin 1`).Puts(eval.Approximately{0.8414709848078965066}),
		That(`math:sin $math:pi`).Puts(eval.Approximately{0.0}),
		That(`math:tan 1`).Puts(eval.Approximately{1.5574077246549023}),
		That(`math:tan $math:pi`).Puts(eval.Approximately{0.0}),

		That(`math:sqrt 0`).Puts(0.0),
		That(`math:sqrt 4`).Puts(2.0),
		That(`math:sqrt -4`).Puts(math.NaN()),

		// Test the inverse trigonometric block of functions.
		That(`math:acos 0`).Puts(math.Acos(0)),
		That(`math:acos 1`).Puts(math.Acos(1)),
		That(`math:acos 1.00001`).Puts(math.NaN()),

		That(`math:asin 0`).Puts(math.Asin(0)),
		That(`math:asin 1`).Puts(math.Asin(1)),
		That(`math:asin 1.00001`).Puts(math.NaN()),

		That(`math:atan 0`).Puts(math.Atan(0)),
		That(`math:atan 1`).Puts(math.Atan(1)),
		That(`math:atan inf`).Puts(math.Pi/2),

		// Test the inverse hyperbolic trigonometric block of functions.
		That(`math:acosh 0`).Puts(math.Acosh(0)),
		That(`math:acosh 1`).Puts(math.Acosh(1)),
		That(`math:acosh nan`).Puts(math.NaN()),

		That(`math:asinh 0`).Puts(math.Asinh(0)),
		That(`math:asinh 1`).Puts(math.Asinh(1)),
		That(`math:asinh inf`).Puts(math.Inf(1)),

		That(`math:atanh 0`).Puts(math.Atanh(0)),
		That(`math:atanh 1`).Puts(math.Inf(1)),
	)
}
