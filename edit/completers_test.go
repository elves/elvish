package edit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/parse"
)

type completerFinderTest struct {
	src  string
	want completerIface
}

func testCompleterFinder(t *testing.T, name string, finder completerFinder, tests []completerFinderTest) {
	for _, test := range tests {
		n, err := parse.Parse("[test]", test.src)
		// Ignore error as long is n is non-nil
		if n == nil {
			panic(err)
		}
		leaf := findLeafNode(n, len(test.src))
		got := finder(leaf)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("For %q, %s(leaf) => %v, want %v", test.src, name, got, test.want)
		}
	}
}
