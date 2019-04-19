package vals

import (
	"testing"

	"github.com/elves/elvish/tt"
)

var (
	li0 = EmptyList
	li4 = MakeList("foo", "bar", "lorem", "ipsum")
	m   = MakeMapFromKV("foo", "bar", "lorem", "ipsum")
)

var indexTests = tt.Table{
	// String indicies
	Args("abc", "0").Rets("a", nil),
	Args("你好", "0").Rets("你", nil),
	Args("你好", "3").Rets("好", nil),
	Args("你好", "2").Rets(any, errIndexNotAtRuneBoundary),
	Args("abc", "1:2").Rets("b", nil),
	Args("abc", "1:").Rets("bc", nil),
	Args("abc", ":").Rets("abc", nil),
	Args("abc", ":0").Rets("", nil), // i == j == 0 is allowed
	Args("abc", "3:").Rets("", nil), // i == j == n is allowed

	// List indices
	// ============

	// Simple indicies: 0 <= i < n.
	Args(li4, "0").Rets("foo", nil),
	Args(li4, "3").Rets("ipsum", nil),
	Args(li0, "0").Rets(any, anyError),
	Args(li4, "4").Rets(any, anyError),
	Args(li4, "5").Rets(any, anyError),
	// Negative indices: -n <= i < 0.
	Args(li4, "-1").Rets("ipsum", nil),
	Args(li4, "-4").Rets("foo", nil),
	Args(li4, "-5").Rets(any, anyError), // Too negative.
	// Decimal indicies are not allowed even if the value is an integer.
	Args(li4, "0.0").Rets(any, anyError),

	// Slice indicies: 0 <= i <= j <= n.
	Args(li4, "1:3").Rets(eq(MakeList("bar", "lorem")), nil),
	Args(li4, "3:4").Rets(eq(MakeList("ipsum")), nil),
	Args(li4, "0:0").Rets(eq(EmptyList), nil), // i == j == 0 is allowed
	Args(li4, "4:4").Rets(eq(EmptyList), nil), // i == j == n is allowed
	// i defaults to 0
	Args(li4, ":2").Rets(eq(MakeList("foo", "bar")), nil),
	Args(li4, ":-1").Rets(eq(MakeList("foo", "bar", "lorem")), nil),
	// j defaults to n
	Args(li4, "3:").Rets(eq(MakeList("ipsum")), nil),
	Args(li4, "-2:").Rets(eq(MakeList("lorem", "ipsum")), nil),
	// Both indices can be omitted.
	Args(li0, ":").Rets(eq(li0), nil),
	Args(li4, ":").Rets(eq(li4), nil),

	// Malformed list indices.
	Args(li4, "a").Rets(any, anyError),
	Args(li4, "1:3:2").Rets(any, anyError),

	// Map indicies
	// ============

	Args(m, "foo").Rets("bar", nil),
	Args(m, "bad").Rets(any, anyError),

	// StructMap indicies
	Args(testStructMap{"foo", 1.0}, "name").Rets("foo", nil),
	Args(testStructMap{"foo", 1.0}, "bad").Rets(nil, NoSuchKey("bad")),
}

func TestIndex(t *testing.T) {
	tt.Test(t, tt.Fn("Index", Index), indexTests)
}
