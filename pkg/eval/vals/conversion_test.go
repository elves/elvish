package vals

import (
	"math/big"
	"reflect"
	"testing"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
)

type someType struct {
	Foo string
}

func TestScanToGo_ConcreteTypeDst(t *testing.T) {
	// A wrapper around ScanToGo, to make it easier to test. Instead of
	// supplying a pointer to the destination, an initial value to the
	// destination is supplied and the result is returned.
	scanToGo := func(src any, dstInit any) (any, error) {
		ptr := reflect.New(TypeOf(dstInit))
		err := ScanToGo(src, ptr.Interface())
		return ptr.Elem().Interface(), err
	}

	tt.Test(t, tt.Fn(scanToGo).Named("scanToGo"),
		// int
		Args("12", 0).Rets(12),
		Args("0x12", 0).Rets(0x12),
		Args(12.0, 0).Rets(0, errMustBeInteger),
		Args(0.5, 0).Rets(0, errMustBeInteger),
		Args(someType{}, 0).Rets(tt.Any, errMustBeInteger),
		Args("x", 0).Rets(tt.Any, cannotParseAs{"integer", "x"}),

		// float64
		Args(23, 0.0).Rets(23.0),
		Args(big.NewRat(1, 2), 0.0).Rets(0.5),
		Args(1.2, 0.0).Rets(1.2),
		Args("23", 0.0).Rets(23.0),
		Args("0x23", 0.0).Rets(float64(0x23)),
		Args(someType{}, 0.0).Rets(tt.Any, errMustBeNumber),
		Args("x", 0.0).Rets(tt.Any, cannotParseAs{"number", "x"}),

		// rune
		Args("x", ' ').Rets('x'),
		Args(someType{}, ' ').Rets(tt.Any, errMustBeString),
		Args("\xc3\x28", ' ').Rets(tt.Any, errMustBeValidUTF8), // Invalid UTF8
		Args("ab", ' ').Rets(tt.Any, errMustHaveSingleRune),

		// Other types don't undergo any conversion, as long as the types match
		Args("foo", "").Rets("foo"),
		Args(someType{"foo"}, someType{}).Rets(someType{"foo"}),
		Args(nil, nil).Rets(nil),
		Args("x", someType{}).Rets(tt.Any, WrongType{"!!vals.someType", "string"}),
	)
}

func TestScanToGo_NumDst(t *testing.T) {
	scanToGo := func(src any) (Num, error) {
		var n Num
		err := ScanToGo(src, &n)
		return n, err
	}

	tt.Test(t, tt.Fn(scanToGo).Named("scanToGo"),
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

		Args("bad").Rets(tt.Any, cannotParseAs{"number", "bad"}),
		Args(EmptyList).Rets(tt.Any, errMustBeNumber),
	)
}

func TestScanToGo_InterfaceDst(t *testing.T) {
	scanToGo := func(src any) (any, error) {
		var l List
		err := ScanToGo(src, &l)
		return l, err
	}

	tt.Test(t, tt.Fn(scanToGo).Named("scanToGo"),
		Args(EmptyList).Rets(EmptyList),

		Args("foo").Rets(tt.Any, WrongType{"!!vector.Vector", "string"}),
	)
}

func TestScanToGo_CallableDstAdmitsNil(t *testing.T) {
	type mockCallable interface {
		Call()
	}
	testutil.Set(t, &CallableType, reflect.TypeOf((*mockCallable)(nil)).Elem())
	scanToGo := func(src any) (any, error) {
		var c mockCallable
		err := ScanToGo(src, &c)
		return c, err
	}

	tt.Test(t, tt.Fn(scanToGo).Named("scanToGo"),
		Args(nil).Rets(mockCallable(nil)),
	)
}

func TestScanToGo_ErrorsWithNonPointerDst(t *testing.T) {
	err := ScanToGo("", 1)
	if err == nil {
		t.Errorf("did not return error")
	}
}

func TestScanListToGo(t *testing.T) {
	// A wrapper around ScanListToGo, to make it easier to test.
	scanListToGo := func(src List, dstInit any) (any, error) {
		ptr := reflect.New(TypeOf(dstInit))
		ptr.Elem().Set(reflect.ValueOf(dstInit))
		err := ScanListToGo(src, ptr.Interface())
		return ptr.Elem().Interface(), err
	}

	tt.Test(t, tt.Fn(scanListToGo).Named("scanListToGo"),
		Args(MakeList("1", "2"), []int{}).Rets([]int{1, 2}),
		Args(MakeList("1", "2"), []string{}).Rets([]string{"1", "2"}),

		Args(MakeList("1", "a"), []int{}).Rets([]int{}, cannotParseAs{"integer", "a"}),
	)
}

func TestScanListElementsToGo(t *testing.T) {
	// A wrapper around ScanListElementsToGo, to make it easier to test.
	scanListElementsToGo := func(src List, inits ...any) ([]any, error) {
		ptrs := make([]any, len(inits))
		for i, init := range inits {
			if o, ok := init.(optional); ok {
				// Wrapping the init value with Optional translates to wrapping
				// the pointer with Optional.
				ptrs[i] = Optional(reflect.New(TypeOf(o.ptr)).Interface())
			} else {
				ptrs[i] = reflect.New(TypeOf(init)).Interface()
			}
		}
		err := ScanListElementsToGo(src, ptrs...)
		vals := make([]any, len(ptrs))
		for i, ptr := range ptrs {
			if o, ok := ptr.(optional); ok {
				vals[i] = reflect.ValueOf(o.ptr).Elem().Interface()
			} else {
				vals[i] = reflect.ValueOf(ptr).Elem().Interface()
			}
		}
		return vals, err
	}

	tt.Test(t, tt.Fn(scanListElementsToGo).Named("scanListElementsToGo"),
		Args(MakeList("1", "2"), 0, 0).Rets([]any{1, 2}),
		Args(MakeList("1", "2"), "", "").Rets([]any{"1", "2"}),
		Args(MakeList("1", "2"), 0, Optional(0)).Rets([]any{1, 2}),
		Args(MakeList("1"), 0, Optional(0)).Rets([]any{1, 0}),

		Args(MakeList("a"), 0).Rets([]any{0},
			cannotParseAs{"integer", "a"}),
		Args(MakeList("1"), 0, 0).Rets([]any{0, 0},
			errs.ArityMismatch{What: "list elements",
				ValidLow: 2, ValidHigh: 2, Actual: 1}),
		Args(MakeList("1"), 0, 0, Optional(0)).Rets([]any{0, 0, 0},
			errs.ArityMismatch{What: "list elements",
				ValidLow: 2, ValidHigh: 3, Actual: 1}),
	)
}

type aStruct struct {
	Foo int
	bar any
}

// Equal is required by cmp.Diff, since aStruct contains unexported fields.
func (a aStruct) Equal(b aStruct) bool { return a == b }

func TestScanMapToGo(t *testing.T) {
	// A wrapper around ScanMapToGo, to make it easier to test.
	scanMapToGo := func(src Map, dstInit any) (any, error) {
		ptr := reflect.New(TypeOf(dstInit))
		ptr.Elem().Set(reflect.ValueOf(dstInit))
		err := ScanMapToGo(src, ptr.Interface())
		return ptr.Elem().Interface(), err
	}

	tt.Test(t, tt.Fn(scanMapToGo).Named("scanMapToGo"),
		Args(MakeMap("foo", "1"), aStruct{}).Rets(aStruct{Foo: 1}),
		// More fields is OK
		Args(MakeMap("foo", "1", "bar", "x"), aStruct{}).Rets(aStruct{Foo: 1}),
		// Fewer fields is OK
		Args(MakeMap(), aStruct{}).Rets(aStruct{}),
		// Unexported fields are ignored
		Args(MakeMap("bar", 20), aStruct{bar: 10}).Rets(aStruct{bar: 10}),

		// Conversion error
		Args(MakeMap("foo", "a"), aStruct{}).
			Rets(aStruct{}, cannotParseAs{"integer", "a"}),
	)
}

func TestFromGo(t *testing.T) {
	tt.Test(t, FromGo,
		// BigInt -> int, when in range
		Args(bigInt(z)).Rets(bigInt(z)),
		Args(big.NewInt(100)).Rets(100),
		// BigRat -> BigInt or int, when denominator is 1
		Args(bigRat(z1+"/"+z)).Rets(bigRat(z1+"/"+z)),
		Args(bigRat(z+"/1")).Rets(bigInt(z)),
		Args(bigRat("2/1")).Rets(2),
		// rune -> string
		Args('x').Rets("x"),

		// Other types don't undergo any conversion
		Args(nil).Rets(nil),
		Args(someType{"foo"}).Rets(someType{"foo"}),
	)
}
