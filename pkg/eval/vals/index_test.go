package vals

import (
	"os"
	"testing"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
)

var (
	li0 = EmptyList
	li4 = MakeList("foo", "bar", "lorem", "ipsum")
	m   = MakeMap("foo", "bar", "lorem", "ipsum")
)

func TestIndex(t *testing.T) {
	tt.Test(t, Index,
		// String indices
		Args("abc", "0").Rets("a", nil),
		Args("abc", 0).Rets("a", nil),
		Args("你好", "0").Rets("你", nil),
		Args("你好", "3").Rets("好", nil),
		Args("你好", "2").Rets(tt.Any, errIndexNotAtRuneBoundary),
		// String slices with half-open range.
		Args("abc", "1..2").Rets("b", nil),
		Args("abc", "1..").Rets("bc", nil),
		Args("abc", "..").Rets("abc", nil),
		Args("abc", "..0").Rets("", nil), // i == j == 0 is allowed
		Args("abc", "3..").Rets("", nil), // i == j == n is allowed
		// String slices with closed range.
		Args("abc", "0..=1").Rets("ab", nil),
		Args("abc", "1..=").Rets("bc", nil),
		Args("abc", "..=1").Rets("ab", nil),
		Args("abc", "..=").Rets("abc", nil),
		Args("abc", "..=-1").Rets("abc", nil),
		// String slices not at rune boundary.
		Args("你好", "2..").Rets(tt.Any, errIndexNotAtRuneBoundary),
		Args("你好", "..2").Rets(tt.Any, errIndexNotAtRuneBoundary),

		// List indices
		// ============

		// Simple indices: 0 <= i < n.
		Args(li4, "0").Rets("foo", nil),
		Args(li4, "3").Rets("ipsum", nil),
		Args(li0, "0").Rets(tt.Any, errs.OutOfRange{
			What: "index", ValidLow: "0", ValidHigh: "-1", Actual: "0"}),
		Args(li4, "4").Rets(tt.Any, errs.OutOfRange{
			What: "index", ValidLow: "0", ValidHigh: "3", Actual: "4"}),
		Args(li4, "5").Rets(tt.Any, errs.OutOfRange{
			What: "index", ValidLow: "0", ValidHigh: "3", Actual: "5"}),
		Args(li4, z).Rets(tt.Any,
			errs.OutOfRange{What: "index", ValidLow: "0", ValidHigh: "3", Actual: z}),
		// Negative indices: -n <= i < 0.
		Args(li4, "-1").Rets("ipsum", nil),
		Args(li4, "-4").Rets("foo", nil),
		Args(li4, "-5").Rets(tt.Any, errs.OutOfRange{
			What: "negative index", ValidLow: "-4", ValidHigh: "-1", Actual: "-5"}),
		Args(li4, "-"+z).Rets(tt.Any,
			errs.OutOfRange{What: "negative index", ValidLow: "-4", ValidHigh: "-1", Actual: "-" + z}),
		// Float indices are not allowed even if the value is an integer.
		Args(li4, 0.0).Rets(tt.Any, errIndexMustBeInteger),

		// Integer indices are allowed.
		Args(li4, 0).Rets("foo", nil),
		Args(li4, 3).Rets("ipsum", nil),
		Args(li4, 5).Rets(nil, errs.OutOfRange{
			What: "index", ValidLow: "0", ValidHigh: "3", Actual: "5"}),
		Args(li4, -1).Rets("ipsum", nil),
		Args(li4, -5).Rets(nil, errs.OutOfRange{
			What: "negative index", ValidLow: "-4", ValidHigh: "-1", Actual: "-5"}),

		// Half-open slices.
		Args(li4, "1..3").Rets(eq(MakeList("bar", "lorem")), nil),
		Args(li4, "3..4").Rets(eq(MakeList("ipsum")), nil),
		Args(li4, "0..0").Rets(eq(EmptyList), nil), // i == j == 0 is allowed
		Args(li4, "4..4").Rets(eq(EmptyList), nil), // i == j == n is allowed
		// i defaults to 0
		Args(li4, "..2").Rets(eq(MakeList("foo", "bar")), nil),
		Args(li4, "..-1").Rets(eq(MakeList("foo", "bar", "lorem")), nil),
		// j defaults to n
		Args(li4, "3..").Rets(eq(MakeList("ipsum")), nil),
		Args(li4, "-2..").Rets(eq(MakeList("lorem", "ipsum")), nil),
		// Both indices can be omitted.
		Args(li0, "..").Rets(eq(li0), nil),
		Args(li4, "..").Rets(eq(li4), nil),

		// Closed slices.
		Args(li4, "1..=2").Rets(eq(MakeList("bar", "lorem")), nil),
		Args(li4, "..=1").Rets(eq(MakeList("foo", "bar")), nil),
		Args(li4, "..=-2").Rets(eq(MakeList("foo", "bar", "lorem")), nil),
		Args(li4, "3..=").Rets(eq(MakeList("ipsum")), nil),
		Args(li4, "..=").Rets(eq(li4), nil),
		Args(li4, "..=-1").Rets(eq(li4), nil),

		// Slice index out of range.
		Args(li4, "-5..1").Rets(nil, errs.OutOfRange{
			What: "negative index", ValidLow: "-4", ValidHigh: "-1", Actual: "-5"}),
		Args(li4, "0..5").Rets(nil, errs.OutOfRange{
			What: "index", ValidLow: "0", ValidHigh: "4", Actual: "5"}),
		Args(li4, z+"..").Rets(nil,
			errs.OutOfRange{What: "index", ValidLow: "0", ValidHigh: "4", Actual: z}),
		// Slice index upper < lower
		Args(li4, "3..2").Rets(nil, errs.OutOfRange{
			What: "slice upper index", ValidLow: "3", ValidHigh: "4", Actual: "2"}),
		Args(li4, "-1..-2").Rets(nil,
			errs.OutOfRange{What: "negative slice upper index",
				ValidLow: "-1", ValidHigh: "-1", Actual: "-2"}),

		// Malformed list indices.
		Args(li4, "a").Rets(tt.Any, errIndexMustBeInteger),
		// TODO(xiaq): Make the error more accurate.
		Args(li4, "1:3:2").Rets(tt.Any, errIndexMustBeInteger),

		// Map indices
		// ============

		Args(m, "foo").Rets("bar", nil),
		Args(m, "bad").Rets(tt.Any, NoSuchKey("bad")),

		// Not indexable
		Args(1, "foo").Rets(nil, errNotIndexable),
	)
}

func TestIndex_File(t *testing.T) {
	testutil.InTempDir(t)
	f, err := os.Create("f")
	if err != nil {
		t.Skip("create file:", err)
	}

	tt.Test(t, Index,
		Args(f, "fd").Rets(int(f.Fd()), nil),
		Args(f, "name").Rets(f.Name(), nil),
		Args(f, "x").Rets(nil, NoSuchKey("x")),
	)
}
