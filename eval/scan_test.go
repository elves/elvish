package eval

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/util"
)

// These are used as arguments to scanArg, since Go does not allow taking
// address of literals.

func intPtr() *int         { var x int; return &x }
func intsPtr() *[]int      { var x []int; return &x }
func float64Ptr() *float64 { var x float64; return &x }
func stringPtr() *string   { var x string; return &x }

var scanArgsTestCases = []struct {
	variadic bool
	src      []interface{}
	dstPtrs  []interface{}
	want     []interface{}
	bad      bool
}{
	// Scanning an int and a String
	{
		src:     []interface{}{"20", "20"},
		dstPtrs: []interface{}{intPtr(), stringPtr()},
		want:    []interface{}{20, "20"},
	},
	// Scanning a String and any number of ints (here 2)
	{
		variadic: true,
		src:      []interface{}{"a", "1", "2"},
		dstPtrs:  []interface{}{stringPtr(), intsPtr()},
		want:     []interface{}{"a", []int{1, 2}},
	},
	// Scanning a String and any number of ints (here 0)
	{
		variadic: true,
		src:      []interface{}{"a"},
		dstPtrs:  []interface{}{stringPtr(), intsPtr()},
		want:     []interface{}{"a", []int{}},
	},

	// Arity mismatch: too few arguments (non-variadic)
	{
		bad:     true,
		src:     []interface{}{},
		dstPtrs: []interface{}{stringPtr()},
	},
	// Arity mismatch: too few arguments (non-variadic)
	{
		bad:     true,
		src:     []interface{}{""},
		dstPtrs: []interface{}{stringPtr(), intPtr()},
	},
	// Arity mismatch: too many arguments (non-variadic)
	{
		bad:     true,
		src:     []interface{}{"1", "2"},
		dstPtrs: []interface{}{stringPtr()},
	},
	// Type mismatch (nonvariadic)
	{
		bad:     true,
		src:     []interface{}{"x"},
		dstPtrs: []interface{}{intPtr()},
	},

	// Arity mismatch: too few arguments (variadic)
	{
		bad:      true,
		src:      []interface{}{},
		dstPtrs:  []interface{}{stringPtr(), intsPtr()},
		variadic: true,
	},
	// Type mismatch within rest arg
	{
		bad:      true,
		src:      []interface{}{"a", "1", "lorem"},
		dstPtrs:  []interface{}{stringPtr(), intsPtr()},
		variadic: true,
	},
}

func TestScanArgs(t *testing.T) {
	for _, tc := range scanArgsTestCases {
		funcToTest := ScanArgs
		if tc.variadic {
			funcToTest = ScanArgsVariadic
		}

		if tc.bad {
			if !util.ThrowsAny(func() { funcToTest(tc.src, tc.dstPtrs...) }) {
				t.Errorf("ScanArgs(%v, %v) should throw an error", tc.src, tc.dstPtrs)
			}
			continue
		}

		funcToTest(tc.src, tc.dstPtrs...)
		dsts := make([]interface{}, len(tc.dstPtrs))
		for i, ptr := range tc.dstPtrs {
			dsts[i] = indirect(ptr)
		}
		err := compareSlice(tc.want, dsts)
		if err != nil {
			t.Errorf("ScanArgs(%v) got %v, want %v", tc.src, dsts, tc.want)
		}
	}
}

var scanArgsBadTestCases = []struct {
	variadic bool
	src      []interface{}
	dstPtrs  []interface{}
}{}

var scanOptsTestCases = []struct {
	bad  bool
	src  map[string]interface{}
	opts []OptToScan
	want []interface{}
}{
	{
		src: map[string]interface{}{
			"foo": "bar",
		},
		opts: []OptToScan{
			{"foo", stringPtr(), "haha"},
		},
		want: []interface{}{
			"bar",
		},
	},
	// Default values.
	{
		src: map[string]interface{}{
			"foo": "bar",
		},
		opts: []OptToScan{
			{"foo", stringPtr(), "haha"},
			{"lorem", stringPtr(), "ipsum"},
		},
		want: []interface{}{
			"bar",
			"ipsum",
		},
	},

	// Unknown option
	{
		bad: true,
		src: map[string]interface{}{
			"foo": "bar",
		},
		opts: []OptToScan{
			{"lorem", stringPtr(), "ipsum"},
		},
	},
}

func TestScanOpts(t *testing.T) {
	for _, tc := range scanOptsTestCases {
		if tc.bad {
			if !util.ThrowsAny(func() { ScanOpts(tc.src, tc.opts...) }) {
				t.Errorf("ScanOpts(%v, %v...) should throw an error",
					tc.src, tc.opts)
			}
			continue
		}

		ScanOpts(tc.src, tc.opts...)
		dsts := make([]interface{}, len(tc.opts))
		for i, opt := range tc.opts {
			dsts[i] = indirect(opt.Ptr)
		}
		err := compareSlice(tc.want, dsts)
		if err != nil {
			t.Errorf("ScanOpts(%v, %v) got %v, want %v", tc.src, tc.opts,
				dsts, tc.want)
		}
	}
}

var scanArgTestCases = []struct {
	source  interface{}
	destPtr interface{}
	want    interface{}
}{
	{"20", intPtr(), 20},
	{"0x20", intPtr(), 0x20},
	{"20", float64Ptr(), 20.0},
	{"23.33", float64Ptr(), 23.33},
	{"", stringPtr(), ""},
	{"1", stringPtr(), "1"},
	{"2.33", stringPtr(), "2.33"},
}

func TestScanArg(t *testing.T) {
	for _, tc := range scanArgTestCases {
		mustScanToGo(tc.source, tc.destPtr)
		if !vals.Equal(indirect(tc.destPtr), tc.want) {
			t.Errorf("scanArg(%s) got %q, want %v", tc.source,
				indirect(tc.destPtr), tc.want)
		}
	}
}

func indirect(i interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}
