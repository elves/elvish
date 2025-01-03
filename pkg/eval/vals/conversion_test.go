package vals

import (
	"math/big"
	"reflect"
	"testing"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
)

type unknownType struct {
	foo string
}

// An easier to test wrapper around ScanToGo. Takes an initial value rather than
// a pointer, and returns the scanned result back.
func scanToGoWithInit(src any, dstInit any) (any, error) {
	ptr := reflect.New(TypeOf(dstInit))
	err := ScanToGo(src, ptr.Interface())
	return ptr.Elem().Interface(), err
}

// Another easier to test wrapper around ScanToGo. Takes a type parameter rather
// than a pointer, and returns the scanned result back.
func scanToGoOfType[T any](src any) (T, error) {
	var dst T
	err := ScanToGo(src, &dst)
	return dst, err
}

func TestScanToGo_ConcreteTypeDst(t *testing.T) {
	tt.Test(t, tt.Fn(scanToGoWithInit),
		// int
		Args("12", 0).Rets(12),
		Args("0x12", 0).Rets(0x12),
		Args(12.0, 0).Rets(0, errMustBeInteger),
		Args(0.5, 0).Rets(0, errMustBeInteger),
		Args(unknownType{}, 0).Rets(tt.Any, errMustBeInteger),
		Args("x", 0).Rets(tt.Any, cannotParseAs{"integer", "x"}),

		// float64
		Args(23, 0.0).Rets(23.0),
		Args(big.NewRat(1, 2), 0.0).Rets(0.5),
		Args(1.2, 0.0).Rets(1.2),
		Args("23", 0.0).Rets(23.0),
		Args("0x23", 0.0).Rets(float64(0x23)),
		Args(unknownType{}, 0.0).Rets(tt.Any, errMustBeNumber),
		Args("x", 0.0).Rets(tt.Any, cannotParseAs{"number", "x"}),

		// rune
		Args("x", ' ').Rets('x'),
		Args(unknownType{}, ' ').Rets(tt.Any, errMustBeString),
		Args("\xc3\x28", ' ').Rets(tt.Any, errMustBeValidUTF8), // Invalid UTF8
		Args("ab", ' ').Rets(tt.Any, errMustHaveSingleRune),

		// Other types don't undergo any conversion, as long as the types match
		Args("foo", "").Rets("foo"),
		Args(unknownType{"foo"}, unknownType{}).Rets(unknownType{"foo"}),
		Args(nil, nil).Rets(nil),
		Args("x", unknownType{}).Rets(tt.Any, WrongType{"!!vals.unknownType", "string"}),
	)
}

func TestScanToGo_NumDst(t *testing.T) {
	tt.Test(t, tt.Fn(scanToGoOfType[Num]),
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

func scanToFieldMapOpts(src any, opt ScanOpt) (fieldMap, error) {
	var dst fieldMap
	err := ScanToGoOpts(src, &dst, opt)
	if err != nil {
		// ScanToGoOpt may have set some fields and return an error. To simplify
		// tests, always return an empty fieldMap in this case.
		return fieldMap{}, err
	}
	return dst, err
}

func TestScanToGo_MapToFieldMap(t *testing.T) {
	tt.Test(t, tt.Fn(scanToFieldMapOpts),
		// ScanOpt(0)
		Args(MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", 23), ScanOpt(0)).
			Rets(fieldMap{Foo: "lorem", Bar: "ipsum", FooBar: 23}, nil),
		// Missing key is not OK
		Args(MakeMap("foo", "lorem", "bar", "ipsum"), ScanOpt(0)).
			Rets(fieldMap{},
				errs.BadValue{What: "value",
					Valid:  "map with keys being exactly [foo bar foo-bar]",
					Actual: "[&bar=ipsum &foo=lorem]"}),
		// Extra key is not OK
		Args(MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", 23, "more", "x"), ScanOpt(0)).
			Rets(fieldMap{},
				errs.BadValue{What: "value",
					Valid:  "map with keys being exactly [foo bar foo-bar]",
					Actual: "[&bar=ipsum &foo=lorem &foo-bar=(num 23) &more=x]"}),
		// Mismatched type is not OK
		Args(MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", "bad"), ScanOpt(0)).
			Rets(fieldMap{}, cannotParseAs{"integer", "bad"}),

		// AllowMissingMapKey
		Args(MakeMap("foo", "lorem"), AllowMissingMapKey).
			Rets(fieldMap{Foo: "lorem"}),
		// Extra key is not OK - len(map) > len(fieldMap)
		Args(MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", 23, "more", "x"), AllowMissingMapKey).
			Rets(fieldMap{},
				errs.BadValue{What: "value",
					Valid:  "map with keys constrained to [foo bar foo-bar]",
					Actual: "[&bar=ipsum &foo=lorem &foo-bar=(num 23) &more=x]"}),
		// Extra key is not OK - len(map) < len(fieldMap)
		Args(MakeMap("foo", "lorem", "more", "x"), AllowMissingMapKey).
			Rets(fieldMap{},
				errs.BadValue{What: "value",
					Valid:  "map with keys constrained to [foo bar foo-bar]",
					Actual: "[&foo=lorem &more=x]"}),
		// Mismatched type is not OK
		Args(MakeMap("foo-bar", "bad"), AllowMissingMapKey).
			Rets(fieldMap{}, cannotParseAs{"integer", "bad"}),

		// AllowExtraMapKey
		Args(MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", 23, "extra", ""), AllowExtraMapKey).
			Rets(fieldMap{Foo: "lorem", Bar: "ipsum", FooBar: 23}),
		// Missing key is not OK - len(map) < len(fieldMap)
		Args(MakeMap("foo", "lorem", "bar", "ipsum"), AllowExtraMapKey).
			Rets(fieldMap{},
				errs.BadValue{What: "value",
					Valid:  "map with keys containing at least [foo bar foo-bar]",
					Actual: "[&bar=ipsum &foo=lorem]"}),
		// Missing key is not OK - len(map) > len(fieldMap)
		Args(MakeMap("foo", "lorem", "bar", "ipsum", "more1", "1", "more2", "2"), AllowExtraMapKey).
			Rets(fieldMap{},
				errs.BadValue{What: "value",
					Valid:  "map with keys containing at least [foo bar foo-bar]",
					Actual: "[&bar=ipsum &foo=lorem &more1=1 &more2=2]"}),
		// Mismatched type is not OK
		Args(MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", "bad"), AllowExtraMapKey).
			Rets(fieldMap{}, cannotParseAs{"integer", "bad"}),

		// AllowMissingMapKey | AllowExtraMapKey
		// Mismatched type is not OK
		Args(MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", "bad"), AllowMissingMapKey|AllowExtraMapKey).
			Rets(fieldMap{}, cannotParseAs{"integer", "bad"}),
	)
}

func TestScanToGo_FieldMapToFieldMap(t *testing.T) {
	tt.Test(t, tt.Fn(scanToFieldMapOpts),
		Args(fieldMap{Foo: "lorem", Bar: "ipsum", FooBar: 23}, ScanOpt(0)).
			Rets(fieldMap{Foo: "lorem", Bar: "ipsum", FooBar: 23}, nil),
		Args(fieldMap2{Foo: "lorem", Bar: "ipsum", FooBar: 23}, ScanOpt(0)).
			Rets(fieldMap{Foo: "lorem", Bar: "ipsum", FooBar: 23}, nil),
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
	scanToGo := func(src any) (any, error) {
		var c mockCallable
		err := ScanToGo(src, &c)
		return c, err
	}

	tt.Test(t, tt.Fn(scanToGo).Named("scanToGo"),
		Args(nil).Rets(mockCallable(nil)),
	)
}

func TestScanToGo_PanicsWithNonPointerDst(t *testing.T) {
	x := testutil.Recover(func() {
		ScanToGo("", 1)
	})
	if x == nil {
		t.Errorf("did not panic")
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
		Args(unknownType{"foo"}).Rets(unknownType{"foo"}),
	)
}

func BenchmarkScanToGo_ScanIntToInt(b *testing.B) {
	benchmarkScanToGo[int](b, 1)
}

func BenchmarkScanToGo_ScanStringToInt(b *testing.B) {
	benchmarkScanToGo[int](b, "1")
}

func BenchmarkScanToGo_ScanFieldMapToFieldMap(b *testing.B) {
	benchmarkScanToGo[fieldMap](b, fieldMap{})
}

func BenchmarkScanToGo_ScanMapToFieldMap(b *testing.B) {
	benchmarkScanToGo[fieldMap](b, MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", 23))
}

func benchmarkScanToGo[D any](b *testing.B, src any) {
	for range b.N {
		var dst D
		ScanToGo(src, &dst)
	}
}
