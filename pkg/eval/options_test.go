package eval

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

type opts struct {
	Foo string
}

// Equal is required by cmp.Diff, since opts contains unexported fields.
func (o opts) Equal(p opts) bool { return o == p }

func TestScanOptions(t *testing.T) {
	// A wrapper of ScanOptions, to make it easier to test
	wrapper := func(src RawOptions, dstInit any) (any, error) {
		ptr := reflect.New(reflect.TypeOf(dstInit))
		ptr.Elem().Set(reflect.ValueOf(dstInit))
		err := scanOptions(src, ptr.Interface())
		return ptr.Elem().Interface(), err
	}

	tt.Test(t, tt.Fn(wrapper).Named("scanOptions"),
		Args(RawOptions{"foo": "lorem ipsum"}, opts{}).
			Rets(opts{Foo: "lorem ipsum"}, nil),
		Args(RawOptions{"bar": 20}, opts{}).
			Rets(opts{}, UnknownOption{"bar"}),
	)
}
