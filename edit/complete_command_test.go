package edit

import (
	"testing"

	"github.com/elves/elvish/parse"
)

func TestFindCommandComplContext(t *testing.T) {
	testComplContextFinder(t, "findCommandComplContext", findCommandComplContext, []complContextFinderTest{
		// Very beginning, empty seed
		{"", &commandComplContext{
			complContextCommon{"", quotingForEmptySeed, 0, 0}}},
		// Very beginning, nonempty seed
		{"a", &commandComplContext{
			complContextCommon{"a", parse.Bareword, 0, 1}}},
		// Very beginning, nonempty seed with quoting
		{`"a"`, &commandComplContext{
			complContextCommon{"a", parse.DoubleQuoted, 0, 3}}},
		// Very beginning, nonempty seed with partial quoting
		{`"a`, &commandComplContext{
			complContextCommon{"a", parse.DoubleQuoted, 0, 2}}},

		// Beginning of output capture, empty seed
		{"a (", &commandComplContext{
			complContextCommon{"", quotingForEmptySeed, 3, 3}}},
		// Beginning of output capture, nonempty seed
		{"a (b", &commandComplContext{
			complContextCommon{"b", parse.Bareword, 3, 4}}},

		// Beginning of exception capture, empty seed
		{"a ?(", &commandComplContext{
			complContextCommon{"", quotingForEmptySeed, 4, 4}}},
		// Beginning of exception capture, nonempty seed
		{"a ?(b", &commandComplContext{
			complContextCommon{"b", parse.Bareword, 4, 5}}},

		// Beginning of lambda, empty seed
		{"a { ", &commandComplContext{
			complContextCommon{"", quotingForEmptySeed, 4, 4}}},
		// Beginning of lambda, nonempty seed
		{"a { b", &commandComplContext{
			complContextCommon{"b", parse.Bareword, 4, 5}}},

		// After another command and pipe, empty seed
		{"a|", &commandComplContext{
			complContextCommon{"", quotingForEmptySeed, 2, 2}}},
		// After another command and pipe, nonempty seed
		{"a|b", &commandComplContext{
			complContextCommon{"b", parse.Bareword, 2, 3}}},

		// After another pipeline, empty seed
		{"a\n", &commandComplContext{
			complContextCommon{"", quotingForEmptySeed, 2, 2}}},
		// After another pipeline, nonempty seed
		{"a\nb", &commandComplContext{
			complContextCommon{"b", parse.Bareword, 2, 3}}},
	})
}
