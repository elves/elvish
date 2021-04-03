package vals

import (
	"testing"

	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/tt"
)

var (
	li0 = EmptyList
	li4 = MakeList("foo", "bar", "lorem", "ipsum")
	m   = MakeMap("foo", "bar", "lorem", "ipsum")
)

func TestIndex(t *testing.T) {
	Test(t, Fn("Index", Index), Table{
		// String indices
		Args("abc", "0").Rets("a", nil),
		Args("abc", int64(0)).Rets("a", nil),
		Args("你好", "0").Rets("你", nil),
		Args("你好", "3").Rets("好", nil),
		Args("你好", "2").Rets(Any, errIndexNotAtRuneBoundary),
		// String slices with half-open range.
		Args("abc", "1..2").Rets("b", nil),
		Args("abc", "1..").Rets("bc", nil),
		Args("abc", "..").Rets("abc", nil),
		Args("abc", "..0").Rets("", nil), // i == j == 0 is allowed
		Args("abc", "3..").Rets("", nil), // i == j == n is allowed
		// String slices with half-open range, using deprecated syntax.
		Args("abc", "1:2").Rets("b", nil),
		Args("abc", "1:").Rets("bc", nil),
		Args("abc", ":").Rets("abc", nil),
		Args("abc", ":0").Rets("", nil), // i == j == 0 is allowed
		Args("abc", "3:").Rets("", nil), // i == j == n is allowed
		// String slices with closed range.
		Args("abc", "0..=1").Rets("ab", nil),
		Args("abc", "1..=").Rets("bc", nil),
		Args("abc", "..=1").Rets("ab", nil),
		Args("abc", "..=").Rets("abc", nil),

		// List indices
		// ============

		// Simple indices: 0 <= i < n.
		Args(li4, "0").Rets("foo", nil),
		Args(li4, "3").Rets("ipsum", nil),
		Args(li0, "0").Rets(Any, errs.OutOfRange{
			What: "index here", ValidLow: "0", ValidHigh: "-1", Actual: "0"}),
		Args(li4, "4").Rets(Any, errs.OutOfRange{
			What: "index here", ValidLow: "0", ValidHigh: "3", Actual: "4"}),
		Args(li4, "5").Rets(Any, errs.OutOfRange{
			What: "index here", ValidLow: "0", ValidHigh: "3", Actual: "5"}),
		// Negative indices: -n <= i < 0.
		Args(li4, "-1").Rets("ipsum", nil),
		Args(li4, "-4").Rets("foo", nil),
		Args(li4, "-5").Rets(Any, errs.OutOfRange{
			What: "negative index here", ValidLow: "-4", ValidHigh: "-1", Actual: "-5"}),
		// Float indices are not allowed even if the value is an integer.
		Args(li4, 0.0).Rets(Any, errIndexMustBeInteger),

		// Integer indices are allowed.
		Args(li4, int64(0)).Rets("foo", nil),
		Args(li4, int64(3)).Rets("ipsum", nil),
		Args(li4, int64(5)).Rets(nil, errs.OutOfRange{
			What: "index here", ValidLow: "0", ValidHigh: "3", Actual: "5"}),
		Args(li4, int64(-1)).Rets("ipsum", nil),
		Args(li4, int64(-5)).Rets(nil, errs.OutOfRange{
			What: "negative index here", ValidLow: "-4", ValidHigh: "-1", Actual: "-5"}),

		// Half-open slices.
		Args(li4, "1..3").Rets(Eq(MakeList("bar", "lorem")), nil),
		Args(li4, "3..4").Rets(Eq(MakeList("ipsum")), nil),
		Args(li4, "0..0").Rets(Eq(EmptyList), nil), // i == j == 0 is allowed
		Args(li4, "4..4").Rets(Eq(EmptyList), nil), // i == j == n is allowed
		// i defaults to 0
		Args(li4, "..2").Rets(Eq(MakeList("foo", "bar")), nil),
		Args(li4, "..-1").Rets(Eq(MakeList("foo", "bar", "lorem")), nil),
		// j defaults to n
		Args(li4, "3..").Rets(Eq(MakeList("ipsum")), nil),
		Args(li4, "-2..").Rets(Eq(MakeList("lorem", "ipsum")), nil),
		// Both indices can be omitted.
		Args(li0, "..").Rets(Eq(li0), nil),
		Args(li4, "..").Rets(Eq(li4), nil),

		// Half-open slices using deprecated syntax.
		Args(li4, "1:3").Rets(Eq(MakeList("bar", "lorem")), nil),
		Args(li4, "3:4").Rets(Eq(MakeList("ipsum")), nil),
		Args(li4, "0:0").Rets(Eq(EmptyList), nil), // i == j == 0 is allowed
		Args(li4, "4:4").Rets(Eq(EmptyList), nil), // i == j == n is allowed
		Args(li4, ":2").Rets(Eq(MakeList("foo", "bar")), nil),
		Args(li4, ":-1").Rets(Eq(MakeList("foo", "bar", "lorem")), nil),
		Args(li4, "3:").Rets(Eq(MakeList("ipsum")), nil),
		Args(li4, "-2:").Rets(Eq(MakeList("lorem", "ipsum")), nil),
		Args(li0, ":").Rets(Eq(li0), nil),
		Args(li4, ":").Rets(Eq(li4), nil),

		// Closed slices.
		Args(li4, "1..=2").Rets(Eq(MakeList("bar", "lorem")), nil),
		Args(li4, "..=1").Rets(Eq(MakeList("foo", "bar")), nil),
		Args(li4, "..=-2").Rets(Eq(MakeList("foo", "bar", "lorem")), nil),
		Args(li4, "3..=").Rets(Eq(MakeList("ipsum")), nil),
		Args(li4, "..=").Rets(Eq(li4), nil),

		// Index out of range.
		Args(li4, "-5:1").Rets(nil, errs.OutOfRange{
			What: "negative index here", ValidLow: "-4", ValidHigh: "-1", Actual: "-5"}),
		Args(li4, "0:5").Rets(nil, errs.OutOfRange{
			What: "index here", ValidLow: "0", ValidHigh: "4", Actual: "5"}),
		Args(li4, "3:2").Rets(nil, errs.OutOfRange{
			What: "slice upper index here", ValidLow: "3", ValidHigh: "4", Actual: "2"}),

		// Malformed list indices.
		Args(li4, "a").Rets(Any, errIndexMustBeInteger),
		// TODO(xiaq): Make the error more accurate.
		Args(li4, "1:3:2").Rets(Any, errIndexMustBeInteger),

		// Map indices
		// ============

		Args(m, "foo").Rets("bar", nil),
		Args(m, "bad").Rets(Any, NoSuchKey("bad")),
	})
}

func TestCheckDeprecatedIndex(t *testing.T) {
	Test(t, Fn("CheckDeprecatedIndex", CheckDeprecatedIndex), Table{
		Args("ab", "1:2").Rets("using : for slice is deprecated; use .. instead"),
		Args("ab", "1").Rets(""),
		Args("ab", "1..2").Rets(""),
		Args("ab", 1.0).Rets(""),
		Args(li4, "1:2").Rets("using : for slice is deprecated; use .. instead"),
		Args(li4, "1").Rets(""),
		Args(li4, "1..2").Rets(""),
		Args(li4, 1.0).Rets(""),
	})
}
