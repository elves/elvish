package edit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type complContextFinderTest struct {
	src  string
	want complContext
}

func testComplContextFinder(t *testing.T, name string, finder complContextFinder, tests []complContextFinderTest) {
	ev := eval.NewEvaler(api.NewClient("/invalid"), nil, "", make(map[string]eval.Namespace))
	for _, test := range tests {
		n, err := parse.Parse("[test]", test.src)
		// Ignore error as long is n is non-nil
		if n == nil {
			panic(err)
		}
		leaf := findLeafNode(n, len(test.src))
		got := finder(leaf, ev)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("For %q, %s(leaf) => %v, want %v", test.src, name, got, test.want)
		}
	}
}
