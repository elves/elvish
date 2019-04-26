package vals

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/tt"
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

var scanToGoTests = tt.Table{
	Args("12", 0).Rets(12),
	Args("0x12", 0).Rets(0x12),
	Args("23", 0.0).Rets(23.0),
	Args("0x23", 0.0).Rets(float64(0x23)),
	Args("x", ' ').Rets('x'),
	Args("foo", "").Rets("foo"),
	Args(someType{"foo"}, someType{}).Rets(someType{"foo"}),
	Args(nil, nil).Rets(nil),

	Args("x", someType{}).Rets(any, anyError),
	Args(someType{}, 0).Rets(any, anyError),
	Args("x", 0).Rets(any, anyError),
	Args(someType{}, 0.0).Rets(any, anyError),
	Args("x", 0.0).Rets(any, anyError),
	Args(someType{}, ' ').Rets(any, anyError),
	Args("\xc3\x28", ' ').Rets(any, anyError), // Invalid UTF8
	Args("ab", ' ').Rets(any, anyError),
}

func TestScanToGo(t *testing.T) {
	tt.Test(t, tt.Fn("ScanToGo", scanToGo2), scanToGoTests)
}

var fromGoTests = tt.Table{
	tt.Args(12).Rets("12"),
	tt.Args(1.5).Rets(1.5),
	tt.Args('x').Rets("x"),
	tt.Args(nil).Rets(nil),
	tt.Args(someType{"foo"}).Rets(someType{"foo"}),
}

func TestFromGo(t *testing.T) {
	tt.Test(t, tt.Fn("FromGo", FromGo), fromGoTests)
}
