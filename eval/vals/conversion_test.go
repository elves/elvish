package vals

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/tt"
)

type someType struct {
	foo string
}

var scanToGoTests = []struct {
	src      interface{}
	preScan  interface{}
	postScan interface{}
}{
	{"12", 0, 12},
	{"23", 0.0, 23.0},
	{"x", ' ', 'x'},
	{"foo", "", "foo"},
	{someType{"foo"}, someType{}, someType{"foo"}},
}

func TestScanToGo(t *testing.T) {
	for _, test := range scanToGoTests {
		ptr := reflect.New(reflect.TypeOf(test.preScan))
		ScanToGo(test.src, ptr.Interface())
		dst := ptr.Elem().Interface()
		if dst != test.postScan {
			t.Errorf("Scan %v %v -> %v, want %v", test.src, test.preScan, dst, test.postScan)
		}
	}
}

var fromGoTests = tt.Table{
	tt.Args(12).Rets("12"),
	tt.Args(1.5).Rets("1.5"),
	tt.Args('x').Rets("x"),
	tt.Args(someType{"foo"}).Rets(someType{"foo"}),
}

func TestFromGo(t *testing.T) {
	tt.Test(t, tt.Fn("FromGo", FromGo), fromGoTests)
}
