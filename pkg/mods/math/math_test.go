package math

import (
	"math"
	"math/big"
	"strconv"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
)

const (
	zeros = "0000000000000000000"
	// Values that exceed the range of int64, used for testing BigInt.
	z  = "1" + zeros + "0"
	z1 = "1" + zeros + "1" // z+1
	z2 = "1" + zeros + "2" // z+2
)

var minIntString = strconv.Itoa(minInt)

func TestMath(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddNs("math", Ns))
	}
	TestWithEvalerSetup(t, setup,
		That("math:abs 2").Puts(2),
		That("math:abs -2").Puts(2),
		That("math:abs "+minIntString).Puts(bigInt(minIntString[1:])),
		That("math:abs "+z).Puts(bigInt(z)),
		That("math:abs -"+z).Puts(bigInt(z)),
		That("math:abs -1/2").Puts(big.NewRat(1, 2)),
		That("math:abs 1/2").Puts(big.NewRat(1, 2)),
		That("math:abs 2.1").Puts(2.1),
		That("math:abs -2.1").Puts(2.1),

		That("math:ceil 2").Puts(2),
		That("math:ceil "+z).Puts(bigInt(z)),
		That("math:ceil 3/2").Puts(2),
		That("math:ceil -3/2").Puts(-1),
		That("math:ceil 2.1").Puts(3.0),
		That("math:ceil -2.1").Puts(-2.0),

		That("math:floor 2").Puts(2),
		That("math:floor "+z).Puts(bigInt(z)),
		That("math:floor 3/2").Puts(1),
		That("math:floor -3/2").Puts(-2),
		That("math:floor 2.1").Puts(2.0),
		That("math:floor -2.1").Puts(-3.0),

		That("math:round 2").Puts(2),
		That("math:round "+z).Puts(bigInt(z)),
		That("math:round 1/3").Puts(0),
		That("math:round 1/2").Puts(1),
		That("math:round 2/3").Puts(1),
		That("math:round -1/3").Puts(0),
		That("math:round -1/2").Puts(-1),
		That("math:round -2/3").Puts(-1),
		That("math:round 2.1").Puts(2.0),
		That("math:round 2.5").Puts(3.0),

		That("math:round-to-even 2").Puts(2),
		That("math:round-to-even "+z).Puts(bigInt(z)),
		That("math:round-to-even 1/3").Puts(0),
		That("math:round-to-even 2/3").Puts(1),
		That("math:round-to-even -1/3").Puts(0),
		That("math:round-to-even -2/3").Puts(-1),
		That("math:round-to-even 2.5").Puts(2.0),
		That("math:round-to-even -2.5").Puts(-2.0),

		That("math:round-to-even 1/2").Puts(0),
		That("math:round-to-even 3/2").Puts(2),
		That("math:round-to-even 5/2").Puts(2),
		That("math:round-to-even 7/2").Puts(4),
		That("math:round-to-even -1/2").Puts(0),
		That("math:round-to-even -3/2").Puts(-2),
		That("math:round-to-even -5/2").Puts(-2),
		That("math:round-to-even -7/2").Puts(-4),

		That("math:trunc 2").Puts(2),
		That("math:trunc "+z).Puts(bigInt(z)),
		That("math:trunc 3/2").Puts(1),
		That("math:trunc -3/2").Puts(-1),
		That("math:trunc 2.1").Puts(2.0),
		That("math:trunc -2.1").Puts(-2.0),

		That("math:is-inf 1.3").Puts(false),
		That("math:is-inf &sign=0 inf").Puts(true),
		That("math:is-inf &sign=1 inf").Puts(true),
		That("math:is-inf &sign=-1 -inf").Puts(true),
		That("math:is-inf &sign=1 -inf").Puts(false),
		That("math:is-inf -inf").Puts(true),
		That("math:is-inf nan").Puts(false),
		That("math:is-inf 1").Puts(false),
		That("math:is-inf "+z).Puts(false),
		That("math:is-inf 1/2").Puts(false),

		That("math:is-nan 1.3").Puts(false),
		That("math:is-nan inf").Puts(false),
		That("math:is-nan nan").Puts(true),
		That("math:is-nan 1").Puts(false),
		That("math:is-nan "+z).Puts(false),
		That("math:is-nan 1/2").Puts(false),

		That("math:max").Throws(
			errs.ArityMismatch{What: "arguments", ValidLow: 1, ValidHigh: -1, Actual: 0},
			"math:max"),
		That("math:max 42").Puts(42),
		That("math:max -3 3 10 -4").Puts(10),
		That("math:max 2 10 "+z).Puts(bigInt(z)),
		That("math:max "+z1+" "+z2+" "+z).Puts(bigInt(z2)),
		That("math:max 1/2 1/3 2/3").Puts(big.NewRat(2, 3)),
		That("math:max 1.0 2.0").Puts(2.0),
		That("math:max 3 NaN 5").Puts(math.NaN()),

		That("math:min").Throws(
			errs.ArityMismatch{What: "arguments", ValidLow: 1, ValidHigh: -1, Actual: 0},
			"math:min"),
		That("math:min 42").Puts(42),
		That("math:min -3 3 10 -4").Puts(-4),
		That("math:min 2 10 "+z).Puts(2),
		That("math:min "+z1+" "+z2+" "+z).Puts(bigInt(z)),
		That("math:min 1/2 1/3 2/3").Puts(big.NewRat(1, 3)),
		That("math:min 1.0 2.0").Puts(1.0),
		That("math:min 3 NaN 5").Puts(math.NaN()),

		// base is int, exp is int
		That("math:pow 2 0").Puts(1),
		That("math:pow 2 1").Puts(2),
		That("math:pow 2 -1").Puts(big.NewRat(1, 2)),
		That("math:pow 2 3").Puts(8),
		That("math:pow 2 -3").Puts(big.NewRat(1, 8)),
		// base is *big.Rat, exp is int
		That("math:pow 2/3 0").Puts(1),
		That("math:pow 2/3 1").Puts(big.NewRat(2, 3)),
		That("math:pow 2/3 -1").Puts(big.NewRat(3, 2)),
		That("math:pow 2/3 3").Puts(big.NewRat(8, 27)),
		That("math:pow 2/3 -3").Puts(big.NewRat(27, 8)),
		// exp is *big.Rat
		That("math:pow 4 1/2").Puts(2.0),
		// exp is float64
		That("math:pow 2 2.0").Puts(4.0),
		That("math:pow 1/2 2.0").Puts(0.25),
		// base is float64
		That("math:pow 2.0 2").Puts(4.0),

		// Tests below this line are tests against simple bindings for Go's math package.

		That("put $math:pi").Puts(math.Pi),
		That("put $math:e").Puts(math.E),

		That("math:trunc 2.1").Puts(2.0),
		That("math:trunc -2.1").Puts(-2.0),
		That("math:trunc 2.5").Puts(2.0),
		That("math:trunc -2.5").Puts(-2.0),
		That("math:trunc (num Inf)").Puts(math.Inf(1)),
		That("math:trunc (num NaN)").Puts(math.NaN()),

		That("math:log $math:e").Puts(1.0),
		That("math:log 1").Puts(0.0),
		That("math:log 0").Puts(math.Inf(-1)),
		That("math:log -1").Puts(math.NaN()),

		That("math:log10 10.0").Puts(1.0),
		That("math:log10 100.0").Puts(2.0),
		That("math:log10 1").Puts(0.0),
		That("math:log10 0").Puts(math.Inf(-1)),
		That("math:log10 -1").Puts(math.NaN()),

		That("math:log2 8").Puts(3.0),
		That("math:log2 1024.0").Puts(10.0),
		That("math:log2 1").Puts(0.0),
		That("math:log2 0").Puts(math.Inf(-1)),
		That("math:log2 -1").Puts(math.NaN()),

		That("math:cos 0").Puts(1.0),
		That("math:cos 1").Puts(math.Cos(1.0)),
		That("math:cos $math:pi").Puts(-1.0),

		That("math:cosh 0").Puts(1.0),
		That("math:cosh inf").Puts(math.Inf(1)),
		That("math:cosh nan").Puts(math.NaN()),

		That("math:sin 0").Puts(0.0),
		That("math:sin 1").Puts(math.Sin(1.0)),
		That("math:sin $math:pi").Puts(math.Sin(math.Pi)),

		That("math:sinh 0").Puts(0.0),
		That("math:sinh inf").Puts(math.Inf(1)),
		That("math:sinh nan").Puts(math.NaN()),

		That("math:tan 0").Puts(0.0),
		That("math:tan 1").Puts(math.Tan(1.0)),
		That("math:tan $math:pi").Puts(math.Tan(math.Pi)),

		That("math:tanh 0").Puts(0.0),
		That("math:tanh inf").Puts(1.0),
		That("math:tanh nan").Puts(math.NaN()),

		// This block of tests isn't strictly speaking necessary. But it helps
		// ensure that we're not just confirming Go statements such as
		//    math.Tan(math.Pi) == math.Tan(math.Pi)
		// are true. The ops that should return a zero value do not actually
		// do so. Which illustrates why an approximate match is needed.
		That("math:cos 1").Puts(Approximately(0.5403023058681397174)),
		That("math:sin 1").Puts(Approximately(0.8414709848078965066)),
		That("math:sin $math:pi").Puts(Approximately(0.0)),
		That("math:tan 1").Puts(Approximately(1.5574077246549023)),
		That("math:tan $math:pi").Puts(Approximately(0.0)),

		That("math:sqrt 0").Puts(0.0),
		That("math:sqrt 4").Puts(2.0),
		That("math:sqrt -4").Puts(math.NaN()),

		// Test the inverse trigonometric block of functions.
		That("math:acos 0").Puts(math.Acos(0)),
		That("math:acos 1").Puts(math.Acos(1)),
		That("math:acos 1.00001").Puts(math.NaN()),

		That("math:asin 0").Puts(math.Asin(0)),
		That("math:asin 1").Puts(math.Asin(1)),
		That("math:asin 1.00001").Puts(math.NaN()),

		That("math:atan 0").Puts(math.Atan(0)),
		That("math:atan 1").Puts(math.Atan(1)),
		That("math:atan inf").Puts(math.Pi/2),

		// Test the inverse hyperbolic trigonometric block of functions.
		That("math:acosh 0").Puts(math.Acosh(0)),
		That("math:acosh 1").Puts(math.Acosh(1)),
		That("math:acosh nan").Puts(math.NaN()),

		That("math:asinh 0").Puts(math.Asinh(0)),
		That("math:asinh 1").Puts(math.Asinh(1)),
		That("math:asinh inf").Puts(math.Inf(1)),

		That("math:atanh 0").Puts(math.Atanh(0)),
		That("math:atanh 1").Puts(math.Inf(1)),
	)
}

func bigInt(s string) *big.Int {
	z, ok := new(big.Int).SetString(s, 0)
	if !ok {
		panic("cannot parse as big int: " + s)
	}
	return z
}
