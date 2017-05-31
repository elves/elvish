package highlight

import (
	"reflect"
	"strings"
	"testing"

	"github.com/elves/elvish/parse"
)

type styling struct {
	begin int
	end   int
	style string
}

type highlightTest struct {
	source       string
	wantStylings []styling
}

// In the test cases against Highlighter.form, commands that start with x are
// bad, everything else is good.
func goodFormHead(head string) bool { return !strings.HasPrefix(head, "x") }

var highlightFormTests = []highlightTest{
	// Temporary assignments.
	{"a=1 b=2", []styling{
		{0, 1, styleForGoodVariable.String()},
		{4, 5, styleForGoodVariable.String()}}},
	// Normal assignments,
	{"a b = 1 2", []styling{
		{0, 1, styleForGoodVariable.String()},
		{2, 3, styleForGoodVariable.String()}}},
	// Good commands.
	{"a", []styling{{0, 1, styleForGoodCommand.String()}}},
	// Bad commands.
	{"xabc", []styling{{0, 4, styleForBadCommand.String()}}},
	{"'xa'", []styling{{0, 4, styleForBadCommand.String()}}},

	// "for".
	// Highlighting variable.
	//012345678901
	{"for x [] { }", []styling{
		{0, 3, styleForGoodCommand.String()},
		{4, 5, styleForGoodVariable.String()}}},
	// Highlighting variable, incomplete form.
	//01234
	{"for x", []styling{
		{0, 3, styleForGoodCommand.String()},
		{4, 5, styleForGoodVariable.String()}}},
	// Highlighting variable and "else".
	//012345678901234567890
	{"for x [] { } else { }", []styling{
		{0, 3, styleForGoodCommand.String()},
		{4, 5, styleForGoodVariable.String()},
		{13, 17, styleForSep["else"]}}},

	// "try".
	// Highlighting except-variable.
	//01234567890123456789
	{"try { } except x { }", []styling{
		{0, 3, styleForGoodCommand.String()},
		{8, 14, styleForSep["except"]},
		{15, 16, styleForGoodVariable.String()},
	}},
	// Highlighting except-variable, incomplete form.
	//0123456789012345
	{"try { } except x", []styling{
		{0, 3, styleForGoodCommand.String()},
		{8, 14, styleForSep["except"]},
		{15, 16, styleForGoodVariable.String()},
	}},
	// Highlighting "else" and "finally".
	//0123456789012345678901234567
	{"try { } else { } finally { }", []styling{
		{0, 3, styleForGoodCommand.String()},
		{8, 12, styleForSep["else"]},
		{17, 24, styleForSep["finally"]},
	}},
}

func TestHighlightForm(t *testing.T) {
	testWithGoodFormHead(t, "form", highlightFormTests,
		func(hl *Highlighter, ps *parse.Parser) {
			hl.form(parse.ParseForm(ps))
		}, goodFormHead)
}

var highlightPrimaryTests = []highlightTest{
	{"what", []styling{{0, 4, styleForPrimary[parse.Bareword].String()}}},
	{"$var", []styling{{0, 4, styleForPrimary[parse.Variable].String()}}},
	{"'a'", []styling{{0, 3, styleForPrimary[parse.SingleQuoted].String()}}},
	{`"x"`, []styling{{0, 3, styleForPrimary[parse.DoubleQuoted].String()}}},
}

func TestHighlightPrimary(t *testing.T) {
	test(t, "primary", highlightPrimaryTests,
		func(hl *Highlighter, ps *parse.Parser) {
			hl.primary(parse.ParsePrimary(ps, false))
		})
}

var highlightSepTests = []highlightTest{
	{">", []styling{{0, 1, styleForSep[">"]}}},
	{"# comment", []styling{{0, 9, styleForComment.String()}}},
}

func TestHighlightSep(t *testing.T) {
	test(t, "sep", highlightSepTests,
		func(hl *Highlighter, ps *parse.Parser) {
			src := ps.Source()
			hl.sep(parse.NewSep(src, 0, len(src)))
		})
}

func test(t *testing.T, what string,
	tests []highlightTest, f func(*Highlighter, *parse.Parser)) {

	testWithGoodFormHead(t, what, tests, f, nil)
}

func testWithGoodFormHead(t *testing.T, what string, tests []highlightTest,
	f func(*Highlighter, *parse.Parser), goodFormHead func(string) bool) {

	for _, test := range tests {
		var stylings []styling
		hl := &Highlighter{goodFormHead, func(b, e int, s string) {
			stylings = append(stylings, styling{b, e, s})
		}}
		ps := parse.NewParser("<test>", test.source)

		f(hl, ps)

		if !reflect.DeepEqual(stylings, test.wantStylings) {
			t.Errorf("%s %q gets stylings %v, want %v", what, test.source,
				stylings, test.wantStylings)
		}
	}
}
