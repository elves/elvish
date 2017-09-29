package eval

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/util"
)

// These are used as arguments to scanArg, since Go does not allow taking
// address of literals.
var (
	int_0     int
	ints_0    []int
	float64_0 float64
	String_0  String
)

var scanArgsTestCases = []struct {
	variadic bool
	src      []Value
	dstPtrs  []interface{}
	want     []interface{}
}{
	// Non-variadic (ScanArgs): scanning an int and a String
	{
		src:     []Value{String("20"), String("20")},
		dstPtrs: []interface{}{&int_0, &String_0},
		want:    []interface{}{20, String("20")},
	},
	// Variadic (ScanArgsVariadic): scanning a String and any number of ints
	// (here 2)
	{
		variadic: true,
		src:      []Value{String("a"), String("1"), String("2")},
		dstPtrs:  []interface{}{&String_0, &ints_0},
		want:     []interface{}{String("a"), []int{1, 2}},
	},
	// Variadic (ScanArgsVariadic): scanning a String and any number of ints
	// (here 0)
	{
		variadic: true,
		src:      []Value{String("a")},
		dstPtrs:  []interface{}{&String_0, &ints_0},
		want:     []interface{}{String("a"), []int{}},
	},
}

func TestScanArgs(t *testing.T) {
	for _, tc := range scanArgsTestCases {
		if tc.variadic {
			ScanArgsVariadic(tc.src, tc.dstPtrs...)
		} else {
			ScanArgs(tc.src, tc.dstPtrs...)
		}
		dsts := make([]interface{}, len(tc.dstPtrs))
		for i, ptr := range tc.dstPtrs {
			dsts[i] = indirect(ptr)
		}
		err := compareSlice(tc.want, dsts)
		if err != nil {
			t.Errorf("ScanArgs(%s) got %q, want %v", tc.src, dsts, tc.want)
		}
	}
}

var scanArgsBadTestCases = []struct {
	variadic bool
	src      []Value
	dstPtrs  []interface{}
}{
	// Non-variadic (ScanArgs):
	// Arity mismatch: too few arguments
	{
		src:     []Value{},
		dstPtrs: []interface{}{&String_0},
	},
	// Arity mismatch: too few arguments
	{
		src:     []Value{String("")},
		dstPtrs: []interface{}{&String_0, &int_0},
	},
	// Arity mismatch: too many arguments
	{
		src:     []Value{String("1"), String("2")},
		dstPtrs: []interface{}{&String_0},
	},
	// Type mismatch
	{
		src:     []Value{String("x")},
		dstPtrs: []interface{}{&int_0},
	},

	// Variadic (ScanArgs):
	// Arity mismatch: too few arguments
	{
		src:      []Value{},
		dstPtrs:  []interface{}{&String_0, &ints_0},
		variadic: true,
	},
	// Type mismatch within rest arg
	{
		src:      []Value{String("a"), String("1"), String("lorem")},
		dstPtrs:  []interface{}{&String_0, &ints_0},
		variadic: true,
	},
}

func TestScanArgsBad(t *testing.T) {
	for _, tc := range scanArgsBadTestCases {
		ok := util.ThrowsAny(func() {
			if tc.variadic {
				ScanArgsVariadic(tc.src, tc.dstPtrs)
			} else {
				ScanArgs(tc.src, tc.dstPtrs)
			}
		})
		if !ok {
			t.Errorf("ScanArgs(%v, %v) should throw an error", tc.src, tc.dstPtrs)
		}
	}
}

var scanArgTestCases = []struct {
	source  Value
	destPtr interface{}
	want    interface{}
}{
	{String("20"), &int_0, 20},
	{String("0x20"), &int_0, 0x20},
	{String("20"), &float64_0, 20.0},
	{String("23.33"), &float64_0, 23.33},
	{String(""), &String_0, String("")},
	{String("1"), &String_0, String("1")},
	{String("2.33"), &String_0, String("2.33")},
}

func TestScanArg(t *testing.T) {
	for _, tc := range scanArgTestCases {
		scanArg(tc.source, tc.destPtr)
		if !equals(indirect(tc.destPtr), tc.want) {
			t.Errorf("scanArg(%s) got %q, want %v", tc.source,
				indirect(tc.destPtr), tc.want)
		}
	}
}

func indirect(i interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}
