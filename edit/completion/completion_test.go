package completion

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/parse/parseutil"
)

type complContextFinderTest struct {
	src  string
	want complContext
}

func testComplContextFinder(t *testing.T, name string, finder complContextFinder, tests []complContextFinderTest) {
	ev := eval.NewEvaler()
	defer ev.Close()
	for _, test := range tests {
		n, err := parse.Parse("[test]", test.src)
		// Ignore error as long is n is non-nil
		if n == nil {
			panic(err)
		}
		leaf := parseutil.FindLeafNode(n, len(test.src))
		got := finder(leaf, ev)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("For %q, %s(leaf) => %v, want %v", test.src, name, got, test.want)
		}
	}
}
