package vals

import (
	"math/big"
	"reflect"
	"testing"

	. "src.elv.sh/pkg/tt"
)

type someType struct {
	foo string
}

// A wrapper around ScanToGo, to make it easier to test. Instead of supplying a
// pointer to the destination, an initial value to the destination is supplied
// and the result is returned.
func scanToGo2(src interface{}, dstInit interface{}) (interface{}, error) {
	ptr := reflect.New(TypeOf(dstInit))
	err := ScanToGo(src, ptr.Interface())
	return ptr.Elem().Interface(), err
}

func TestScanToGo(t *testing.T) {
	Test(t, Fn("ScanToGo", scanToGo2), Table{
		// int
		Args("12", 0).Rets(12),
		Args("0x12", 0).Rets(0x12),
		Args(12.0, 0).Rets(0, errMustBeInteger),
		Args(0.5, 0).Rets(0, errMustBeInteger),
		Args(someType{}, 0).Rets(Any, errMustBeInteger),
		Args("x", 0).Rets(Any, cannotParseAs{"integer", "x"}),

		// float64
		Args(23, 0.0).Rets(23.0),
		Args(big.NewRat(1, 2), 0.0).Rets(0.5),
		Args(1.2, 0.0).Rets(1.2),
		Args("23", 0.0).Rets(23.0),
		Args("0x23", 0.0).Rets(float64(0x23)),
		Args(someType{}, 0.0).Rets(Any, errMustBeNumber),
		Args("x", 0.0).Rets(Any, cannotParseAs{"number", "x"}),

		// Num is tested below

		// rune
		Args("x", ' ').Rets('x'),
		Args(someType{}, ' ').Rets(Any, errMustBeString),
		Args("\xc3\x28", ' ').Rets(Any, errMustBeValidUTF8), // Invalid UTF8
		Args("ab", ' ').Rets(Any, errMustHaveSingleRune),

		// Other types don't undergo any conversion, as long as the types match
		Args("foo", "").Rets("foo"),
		Args(someType{"foo"}, someType{}).Rets(someType{"foo"}),
		Args(nil, nil).Rets(nil),
		Args("x", someType{}).Rets(Any, WrongType{"!!*vals.someType", "string"}),
	})
}

func scanToGoNum(src interface{}) (Num, error) {
	var n Num
	err := ScanToGo(src, &n)
	return n, err
}

func TestScanToGoNum(t *testing.T) {
	Test(t, Fn("ScanToGo", scanToGoNum), Table{
		// Strings are automatically converted
		Args("12").Rets(12),
		Args(z).Rets(bigInt(z)),
		Args("1/2").Rets(big.NewRat(1, 2)),
		Args("12.0").Rets(12.0),
		// Already numbers
		Args(12).Rets(12),
		Args(bigInt(z)).Rets(bigInt(z)),
		Args(big.NewRat(1, 2)).Rets(big.NewRat(1, 2)),
		Args(12.0).Rets(12.0),
	})
}

func TestFromGo(t *testing.T) {
	Test(t, Fn("FromGo", FromGo), Table{
		// BigInt -> int, when in range
		Args(bigInt(z)).Rets(bigInt(z)),
		Args(big.NewInt(100)).Rets(100),
		// BigRat -> BigInt or int, when denominator is 1
		Args(bigRat(z1 + "/" + z)).Rets(bigRat(z1 + "/" + z)),
		Args(bigRat(z + "/1")).Rets(bigInt(z)),
		Args(bigRat("2/1")).Rets(2),
		// rune -> string
		Args('x').Rets("x"),

		// Other types don't undergo any conversion
		Args(nil).Rets(nil),
		Args(someType{"foo"}).Rets(someType{"foo"}),
	})
}
