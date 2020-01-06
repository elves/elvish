package vals

import (
	"reflect"
	"testing"

	. "github.com/elves/elvish/pkg/tt"
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
		Args("12", 0).Rets(12),
		Args("0x12", 0).Rets(0x12),
		Args(12.0, 0).Rets(12),
		Args("23", 0.0).Rets(23.0),
		Args("0x23", 0.0).Rets(float64(0x23)),
		Args("x", ' ').Rets('x'),
		Args("foo", "").Rets("foo"),
		Args(someType{"foo"}, someType{}).Rets(someType{"foo"}),
		Args(nil, nil).Rets(nil),

		Args(0.5, 0).Rets(0, errMustBeInteger),
		Args("x", someType{}).Rets(Any, wrongType{"!!vals.someType", "string"}),
		Args(someType{}, 0).Rets(Any, errMustBeInteger),
		Args("x", 0).Rets(Any, cannotParseAs{"integer", "x"}),
		Args(someType{}, 0.0).Rets(Any, errMustBeNumber),
		Args("x", 0.0).Rets(Any, cannotParseAs{"number", "x"}),
		Args(someType{}, ' ').Rets(Any, errMustBeString),
		Args("\xc3\x28", ' ').Rets(Any, errMustBeValidUTF8), // Invalid UTF8
		Args("ab", ' ').Rets(Any, errMustHaveSingleRune),
	})
}

func TestFromGo(t *testing.T) {
	Test(t, Fn("FromGo", FromGo), Table{
		Args(12).Rets("12"),
		Args(1.5).Rets(1.5),
		Args('x').Rets("x"),
		Args(nil).Rets(nil),
		Args(someType{"foo"}).Rets(someType{"foo"}),
	})
}
