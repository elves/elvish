package eval

import (
	"reflect"
	"testing"
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
	src      []Value
	dstPtrs  []interface{}
	variadic bool
	want     []interface{}
}{
	{nil, nil, false, nil},
	{[]Value{String("20")}, []interface{}{&int_0}, false, []interface{}{20}},
	{[]Value{String("a"), String("1"), String("2")},
		[]interface{}{&String_0, &ints_0}, true, []interface{}{
			String("a"), []int{1, 2}}},
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
