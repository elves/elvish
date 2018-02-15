package eval

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/util"
)

// These are used as arguments to scanArg, since Go does not allow taking
// address of literals.

func intPtr() *int         { var x int; return &x }
func intsPtr() *[]int      { var x []int; return &x }
func float64Ptr() *float64 { var x float64; return &x }
func stringPtr() *string   { var x string; return &x }

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

func indirect(i interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(i)).Interface()
}
