package edit

import (
	"testing"

	"github.com/elves/elvish/parse"
)

func TestFindCommandCompleter(t *testing.T) {
	testCompleterFinder(t, "findCommandCompleter", findCommandCompleter, []completerFinderTest{
		// Very beginning, empty seed
		{"", &commandCompleter{"", quotingForEmptySeed, 0, 0}},
		// Very beginning, nonempty seed
		{"a", &commandCompleter{"a", parse.Bareword, 0, 1}},
		// Very beginning, nonempty seed with quoting
		{`"a"`, &commandCompleter{"a", parse.DoubleQuoted, 0, 3}},
		// Very beginning, nonempty seed with partial quoting
		{`"a`, &commandCompleter{"a", parse.DoubleQuoted, 0, 2}},

		// Beginning of output capture, empty seed
		{"a (", &commandCompleter{"", quotingForEmptySeed, 3, 3}},
		// Beginning of output capture, nonempty seed
		{"a (b", &commandCompleter{"b", parse.Bareword, 3, 4}},

		// Beginning of exception capture, empty seed
		{"a ?(", &commandCompleter{"", quotingForEmptySeed, 4, 4}},
		// Beginning of exception capture, nonempty seed
		{"a ?(b", &commandCompleter{"b", parse.Bareword, 4, 5}},

		// Beginning of lambda, empty seed
		{"a { ", &commandCompleter{"", quotingForEmptySeed, 4, 4}},
		// Beginning of lambda, nonempty seed
		{"a { b", &commandCompleter{"b", parse.Bareword, 4, 5}},

		// After another command and pipe, empty seed
		{"a|", &commandCompleter{"", quotingForEmptySeed, 2, 2}},
		// After another command and pipe, nonempty seed
		{"a|b", &commandCompleter{"b", parse.Bareword, 2, 3}},

		// After another pipeline, empty seed
		{"a\n", &commandCompleter{"", quotingForEmptySeed, 2, 2}},
		// After another pipeline, nonempty seed
		{"a\nb", &commandCompleter{"b", parse.Bareword, 2, 3}},
	})
}
