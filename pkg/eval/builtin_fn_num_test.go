package eval_test

import (
	"math"
	"math/big"
	"strings"
	"testing"

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

func TestFloat64(t *testing.T) {
	Test(t,
		That("float64 1").Puts(1.0),
		That("float64 (float64 1)").Puts(1.0),
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
