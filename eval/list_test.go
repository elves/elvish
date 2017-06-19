package eval

import (
	"testing"

	"github.com/elves/elvish/util"
)

var parseAndFixListIndexTests = []struct {
	name string
	// input
	expr string
	len  int
	// output
	shouldPanic, isSlice bool
	begin, end           int
}{
	{
		name: "stringIndex",
		expr: "a", len: 0,
		shouldPanic: true,
	},
	{
		name: "floatIndex",
		expr: "1.0", len: 0,
		shouldPanic: true,
	},
	{
		name: "emptyZeroIndex",
		expr: "0", len: 0,
		shouldPanic: true,
	},
	{
		name: "emptyPosIndex",
		expr: "1", len: 0,
		shouldPanic: true,
	},
	{
		name: "emptyNegIndex",
		expr: "-1", len: 0,
		shouldPanic: true,
	},
	{
		name: "emptySliceAbbrevBoth",
		expr: ":", len: 0,
		shouldPanic: true,
	},
	{
		name: "i<-n",
		expr: "-2", len: 1,
		shouldPanic: true,
	},
	{
		name: "i=-n",
		expr: "-1", len: 1,
		begin: 0, end: 0,
	},
	{
		name: "-n<i<0",
		expr: "-1", len: 2,
		begin: 1, end: 0,
	},
	{
		name: "i=0",
		expr: "0", len: 2,
		begin: 0, end: 0,
	},
	{
		name: "0<i<n",
		expr: "1", len: 2,
		begin: 1, end: 0,
	},
	{
		name: "i=n",
		expr: "1", len: 1,
		shouldPanic: true,
	},
	{
		name: "i>n",
		expr: "2", len: 1,
		shouldPanic: true,
	},
	{
		name: "sliceAbbrevBoth",
		expr: ":", len: 1,
		isSlice: true, begin: 0, end: 1,
	},
	{
		name: "sliceAbbrevBegin",
		expr: ":1", len: 1,
		isSlice: true, begin: 0, end: 1,
	},
	{
		name: "sliceAbbrevEnd",
		expr: "0:", len: 1,
		isSlice: true, begin: 0, end: 1,
	},
	{
		name: "sliceNegEnd",
		expr: "0:-1", len: 1,
		isSlice: true, begin: 0, end: 0,
	},
	{
		name: "sliceBeginEqualEnd",
		expr: "1:1", len: 2,
		isSlice: true, begin: 1, end: 1,
	},
	{
		name: "sliceBeginAboveEnd",
		expr: "1:0", len: 2,
		shouldPanic: true,
	},
}

func TestParseAndFixListIndex(t *testing.T) {
	checkEqual := func(name, value string, want, got interface{}) {
		if want != got {
			t.Errorf("%s value: [%s] want: [%v] got: [%v]",
				name, value, want, got)
		}
	}

	for _, item := range parseAndFixListIndexTests {
		var (
			isSlice    bool
			begin, end int
		)

		if err := util.PCall(func() {
			isSlice, begin, end = ParseAndFixListIndex(item.expr, item.len)
		}); err != nil {
			checkEqual(item.name, "shouldPanic", item.shouldPanic, err != nil)
			continue
		}
		checkEqual(item.name, "isSlice", item.isSlice, isSlice)
		checkEqual(item.name, "begin", item.begin, begin)
		checkEqual(item.name, "end", item.end, end)
	}

}
