package eval_test

import (
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"
	"unsafe"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"

	. "src.elv.sh/pkg/eval/evaltest"
)

const (
	zeros = "0000000000000000000"
	// Values that exceed the range of int64, used for testing bigint.
	z   = "1" + zeros + "0"
	z1  = "1" + zeros + "1" // z+1
	z2  = "1" + zeros + "2" // z+2
	z3  = "1" + zeros + "3" // z+3
	zz  = "2" + zeros + "0" // 2z
	zz1 = "2" + zeros + "1" // 2z+1
	zz2 = "2" + zeros + "2" // 2z+2
	zz3 = "2" + zeros + "3" // 2z+3
)

func TestNum(t *testing.T) {
	Test(t,
		That("num 1").Puts(1),
		That("num "+z).Puts(bigInt(z)),
		That("num 1/2").Puts(big.NewRat(1, 2)),
		That("num 0.1").Puts(0.1),
		That("num (num 1)").Puts(1),
	)
}

func TestExactNum(t *testing.T) {
	Test(t,
		That("exact-num 1").Puts(1),
		That("exact-num 0.125").Puts(big.NewRat(1, 8)),
		That("exact-num inf").Throws(errs.BadValue{
			What: "argument here", Valid: "finite float", Actual: "+Inf"}),
	)
}

func TestInexactNum(t *testing.T) {
	Test(t,
		That("inexact-num 1").Puts(1.0),
		That("inexact-num 1.0").Puts(1.0),
		That("inexact-num (num 1)").Puts(1.0),
		That("inexact-num (num 1.0)").Puts(1.0),
	)
}

func TestNumCmp(t *testing.T) {
	Test(t,
		// int
		That("< 1 2 3").Puts(true),
		That("< 1 3 2").Puts(false),
		// bigint
		That("< "+args(z1, z2, z3)).Puts(true),
		That("< "+args(z1, z3, z2)).Puts(false),
		// bigint and int
		That("< "+args("1", z1)).Puts(true),
		// bigrat
		That("< 1/4 1/3 1/2").Puts(true),
		That("< 1/4 1/2 1/3").Puts(false),
		// bigrat, bigint and int
		That("< "+args("1/2", "1", z1)).Puts(true),
		That("< "+args("1/2", z1, "1")).Puts(false),
		// float64
		That("< 1.0 2.0 3.0").Puts(true),
		That("< 1.0 3.0 2.0").Puts(false),
		// float64, bigrat and int
		That("< 1.0 3/2 2").Puts(true),
		That("< 1.0 2 3/2").Puts(false),

		// Mixing of types not tested for commands below; they share the same
		// code path as <.

		// int
		That("<= 1 1 2").Puts(true),
		That("<= 1 2 1").Puts(false),
		// bigint
		That("<= "+args(z1, z1, z2)).Puts(true),
		That("<= "+args(z1, z2, z1)).Puts(false),
		// bigrat
		That("<= 1/3 1/3 1/2").Puts(true),
		That("<= 1/3 1/2 1/1").Puts(true),
		// float64
		That("<= 1.0 1.0 2.0").Puts(true),
		That("<= 1.0 2.0 1.0").Puts(false),

		// int
		That("== 1 1 1").Puts(true),
		That("== 1 2 1").Puts(false),
		// bigint
		That("== "+args(z1, z1, z1)).Puts(true),
		That("== "+args(z1, z2, z1)).Puts(false),
		// bigrat
		That("== 1/2 1/2 1/2").Puts(true),
		That("== 1/2 1/3 1/2").Puts(false),
		// float64
		That("== 1.0 1.0 1.0").Puts(true),
		That("== 1.0 2.0 1.0").Puts(false),

		// int
		That("!= 1 2 1").Puts(true),
		That("!= 1 1 2").Puts(false),
		// bigint
		That("!= "+args(z1, z2, z1)).Puts(true),
		That("!= "+args(z1, z1, z2)).Puts(false),
		// bigrat
		That("!= 1/2 1/3 1/2").Puts(true),
		That("!= 1/2 1/2 1/3").Puts(false),
		// float64
		That("!= 1.0 2.0 1.0").Puts(true),
		That("!= 1.0 1.0 2.0").Puts(false),

		// int
		That("> 3 2 1").Puts(true),
		That("> 3 1 2").Puts(false),
		// bigint
		That("> "+args(z3, z2, z1)).Puts(true),
		That("> "+args(z3, z1, z2)).Puts(false),
		// bigrat
		That("> 1/2 1/3 1/4").Puts(true),
		That("> 1/2 1/4 1/3").Puts(false),
		// float64
		That("> 3.0 2.0 1.0").Puts(true),
		That("> 3.0 1.0 2.0").Puts(false),

		// int
		That(">= 3 3 2").Puts(true),
		That(">= 3 2 3").Puts(false),
		// bigint
		That(">= "+args(z3, z3, z2)).Puts(true),
		That(">= "+args(z3, z2, z3)).Puts(false),
		// bigrat
		That(">= 1/2 1/2 1/3").Puts(true),
		That(">= 1/2 1/3 1/2").Puts(false),
		// float64
		That(">= 3.0 3.0 2.0").Puts(true),
		That(">= 3.0 2.0 3.0").Puts(false),
	)
}

func TestArithmeticCommands(t *testing.T) {
	Test(t,
		// No argument
		That("+").Puts(0),
		// int
		That("+ 233100 233").Puts(233333),
		// bigint
		That("+ "+args(z, z1)).Puts(bigInt(zz1)),
		// bigint and int
		That("+ 1 2 "+z).Puts(bigInt(z3)),
		// bigrat
		That("+ 1/2 1/3 1/4").Puts(big.NewRat(13, 12)),
		// bigrat, bigint and int
		That("+ 1/2 1/2 1 "+z).Puts(bigInt(z2)),
		// float64
		That("+ 0.5 0.25 1.0").Puts(1.75),
		// float64 and other types
		That("+ 0.5 1/4 1").Puts(1.75),

		// Mixing of types not tested for commands below; they share the same
		// code path as +.

		That("-").Throws(ErrorWithType(errs.ArityMismatch{})),
		// One argument - negation
		That("- 233").Puts(-233),
		That("- "+z).Puts(bigInt("-"+z)),
		That("- 1/2").Puts(big.NewRat(-1, 2)),
		That("- 1.0").Puts(-1.0),
		// int
		That("- 20 10 2").Puts(8),
		// bigint
		That("- "+args(zz3, z1)).Puts(bigInt(z2)),
		// bigrat
		That("- 1/2 1/3").Puts(big.NewRat(1, 6)),
		// float64
		That("- 2.0 1.0 0.5").Puts(0.5),

		// No argument
		That("*").Puts(1),
		// int
		That("* 2 7 4").Puts(56),
		// bigint
		That("* 2 "+z1).Puts(bigInt(zz2)),
		// bigrat
		That("* 1/2 1/3").Puts(big.NewRat(1, 6)),
		// float64
		That("* 2.0 0.5 1.75").Puts(1.75),
		// 0 * non-infinity
		That("* 0 1/2 1.0").Puts(0),
		// 0 * infinity
		That("* 0 +Inf").Puts(math.NaN()),

		// One argument - inversion
		That("/ 2").Puts(big.NewRat(1, 2)),
		That("/ "+z).Puts(bigRat("1/"+z)),
		That("/ 2.0").Puts(0.5),
		// int
		That("/ 233333 353").Puts(661),
		That("/ 3 4 2").Puts(big.NewRat(3, 8)),
		// bigint
		That("/ "+args(zz, z)).Puts(2),
		That("/ "+args(zz, "2")).Puts(bigInt(z)),
		That("/ "+args(z1, z)).Puts(bigRat(z1+"/"+z)),
		// float64
		That("/ 1.0 2.0 4.0").Puts(0.125),
		// 0 / non-zero
		That("/ 0 1/2 0.1").Puts(0),
		// anything / 0
		That("/ 0 0").Throws(ErrDivideByZero, "/ 0 0"),
		That("/ 1 0").Throws(ErrDivideByZero, "/ 1 0"),
		That("/ 1.0 0").Throws(ErrDivideByZero, "/ 1.0 0"),

		That("% 23 7").Puts(2),
		That("% 1 0").Throws(ErrDivideByZero, "% 1 0"),
	)
}

func TestRandint(t *testing.T) {
	Test(t,
		That("randint 1 2").Puts(1),
		That("randint 1").Puts(0),
		That("var i = (randint 10 100); and (<= 10 $i) (< $i 100)").Puts(true),
		That("var i = (randint 10); and (<= 0 $i) (< $i 10)").Puts(true),

		That("randint 2 1").Throws(
			errs.BadValue{What: "high value", Valid: "larger than 2", Actual: "1"},
			"randint 2 1"),
		That("randint").Throws(ErrorWithType(errs.ArityMismatch{}), "randint"),
		That("randint 1 2 3").Throws(ErrorWithType(errs.ArityMismatch{}), "randint 1 2 3"),
	)
}

func TestRandSeed(t *testing.T) {
	//lint:ignore SA1019 Reseed to make other RNG-dependent tests non-deterministic
	defer rand.Seed(time.Now().UTC().UnixNano())

	Test(t,
		// Observe that the effect of -randseed is making randint deterministic
		That("fn f { -randseed 0; randint 10 }; eq (f) (f)").Puts(true),
	)
}

var (
	maxInt = 1<<((unsafe.Sizeof(0)*8)-1) - 1
	minInt = -maxInt - 1

	maxDenseIntInFloat = float64(1 << 53)
)

func TestRange(t *testing.T) {
	Test(t,
		// Basic argument sanity checks.
		That("range").Throws(ErrorWithType(errs.ArityMismatch{})),
		That("range 0 1 2").Throws(ErrorWithType(errs.ArityMismatch{})),

		// Int count up.
		That("range 3").Puts(0, 1, 2),
		That("range 1 3").Puts(1, 2),
		// Int count down.
		That("range -1 10 &step=3").Puts(-1, 2, 5, 8),
		That("range 3 -3").Puts(3, 2, 1, 0, -1, -2),
		// Near maxInt or minInt.
		That("range "+args(maxInt-2, maxInt)).Puts(maxInt-2, maxInt-1),
		That("range "+args(maxInt, maxInt-2)).Puts(maxInt, maxInt-1),
		That("range "+args(minInt, minInt+2)).Puts(minInt, minInt+1),
		That("range "+args(minInt+2, minInt)).Puts(minInt+2, minInt+1),
		// Invalid step given the "start" and "end" values of the range.
		That("range &step=-1 1").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-1"}),
		That("range &step=1 1 0").
			Throws(errs.BadValue{What: "step", Valid: "negative", Actual: "1"}),
		thatOutputErrorIsBubbled("range 2"),

		// Big int count up.
		That("range "+z+" "+z3).Puts(bigInt(z), bigInt(z1), bigInt(z2)),
		That("range "+z+" "+z3+" &step=2").Puts(bigInt(z), bigInt(z2)),
		// Big int count down.
		That("range "+z3+" "+z).Puts(bigInt(z3), bigInt(z2), bigInt(z1)),
		That("range "+z3+" "+z+" &step=-2").Puts(bigInt(z3), bigInt(z1)),
		// Invalid big int step.
		That("range &step=-"+z+" 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-" + z}),
		That("range &step="+z+" 10 0").
			Throws(errs.BadValue{What: "step", Valid: "negative", Actual: z}),
		thatOutputErrorIsBubbled("range "+z+" "+z1),

		// Rational count up.
		That("range 23/10").Puts(0, 1, 2),
		That("range 1/10 23/10").Puts(
			big.NewRat(1, 10), big.NewRat(11, 10), big.NewRat(21, 10)),
		That("range 23/10 1/10").Puts(
			big.NewRat(23, 10), big.NewRat(13, 10), big.NewRat(3, 10)),
		That("range 1/10 9/10 &step=3/10").Puts(
			big.NewRat(1, 10), big.NewRat(4, 10), big.NewRat(7, 10)),
		// Rational count down.
		That("range 9/10 0/10 &step=-3/10").Puts(
			big.NewRat(9, 10), big.NewRat(6, 10), big.NewRat(3, 10)),
		// Invalid rational step.
		That("range &step=-1/2 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-1/2"}),
		That("range &step=1/2 10 0").
			Throws(errs.BadValue{What: "step", Valid: "negative", Actual: "1/2"}),
		thatOutputErrorIsBubbled("range 1/2 3/2"),

		// Float64 count up.
		That("range 1.2").Puts(0.0, 1.0),
		That("range &step=0.5 1 3").Puts(1.0, 1.5, 2.0, 2.5),
		// Float64 count down.
		That("range 1.2 -1.2").Puts(1.2, Approximately(0.2), Approximately(-0.8)),
		That("range &step=-0.5 3 1").Puts(3.0, 2.5, 2.0, 1.5),
		// Near maxDenseIntInFloat.
		That("range "+args(maxDenseIntInFloat-2, "+inf")).
			Puts(maxDenseIntInFloat-2, maxDenseIntInFloat-1, maxDenseIntInFloat),
		That("range "+args(maxDenseIntInFloat, maxDenseIntInFloat-2)).
			Puts(maxDenseIntInFloat, maxDenseIntInFloat-1),
		// Invalid float64 step.
		That("range &step=-0.5 10").
			Throws(errs.BadValue{What: "step", Valid: "positive", Actual: "-0.5"}),
		That("range &step=0.5 10 0").
			Throws(errs.BadValue{What: "step", Valid: "negative", Actual: "0.5"}),
		thatOutputErrorIsBubbled("range 1.2"),
	)
}

func bigInt(s string) *big.Int {
	z, ok := new(big.Int).SetString(s, 0)
	if !ok {
		panic("cannot parse as big int: " + s)
	}
	return z
}

func bigRat(s string) *big.Rat {
	z, ok := new(big.Rat).SetString(s)
	if !ok {
		panic("cannot parse as big rat: " + s)
	}
	return z
}

func args(vs ...any) string {
	s := make([]string, len(vs))
	for i, v := range vs {
		s[i] = vals.ToString(v)
	}
	return strings.Join(s, " ")
}
