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

type emitTests struct {
	source       string
	wantStylings []styling
}

// In the test cases, commands that start with x are bad, everything else is
// good.
func goodFormHead(head string) bool { return !strings.HasPrefix(head, "x") }

// This just tests the Highlight method itself, its dependencies are tested
// below.
var emitAllTests = []emitTests{
	//01234
	{"x 'y'", []styling{
		{0, 1, styleForBadCommand},
		{0, 1, styleForPrimary[parse.Bareword]},
		{2, 5, styleForPrimary[parse.SingleQuoted]},
	}},
}

func TestEmitAll(t *testing.T) {
	test(t, "form", emitAllTests,
		func(e *Emitter, src string) {
			n := &parse.Chunk{}
			parse.As("<test>", src, n)
			e.EmitAll(n)
		})
}

var formTests = []emitTests{
	// Temporary assignments.
	{"a=1 b=2", []styling{
		{0, 1, styleForGoodVariable},
		{4, 5, styleForGoodVariable}}},
	// Normal assignments,
	{"a b = 1 2", []styling{
		{0, 1, styleForGoodVariable},
		{2, 3, styleForGoodVariable}}},
	// Good commands.
	{"a", []styling{{0, 1, styleForGoodCommand}}},
	// Bad commands.
	{"xabc", []styling{{0, 4, styleForBadCommand}}},
	{"'xa'", []styling{{0, 4, styleForBadCommand}}},

	// "while"
	// Highlighting "else"
	{"while $false { } else { }", []styling{
		{0, 5, styleForGoodCommand},
		{17, 21, styleForSep["else"]}}},

	// "for".
	// Highlighting variable.
	//012345678901
	{"for x [] { }", []styling{
		{0, 3, styleForGoodCommand},
		{4, 5, styleForGoodVariable}}},
	// Highlighting variable, incomplete form.
	//01234
	{"for x", []styling{
		{0, 3, styleForGoodCommand},
		{4, 5, styleForGoodVariable}}},
	// Highlighting variable and "else".
	//012345678901234567890
	{"for x [] { } else { }", []styling{
		{0, 3, styleForGoodCommand},
		{4, 5, styleForGoodVariable},
		{13, 17, styleForSep["else"]}}},

	// "try".
	// Highlighting except-variable.
	//01234567890123456789
	{"try { } except x { }", []styling{
		{0, 3, styleForGoodCommand},
		{8, 14, styleForSep["except"]},
		{15, 16, styleForGoodVariable},
	}},
	// Highlighting except-variable, incomplete form.
	//0123456789012345
	{"try { } except x", []styling{
		{0, 3, styleForGoodCommand},
		{8, 14, styleForSep["except"]},
		{15, 16, styleForGoodVariable},
	}},
	// Highlighting "else" and "finally".
	//0123456789012345678901234567
	{"try { } else { } finally { }", []styling{
		{0, 3, styleForGoodCommand},
		{8, 12, styleForSep["else"]},
		{17, 24, styleForSep["finally"]},
	}},
}

func TestForm(t *testing.T) {
	test(t, "form", formTests,
		func(e *Emitter, src string) {
			n := &parse.Form{}
			parse.As("<test>", src, n)
			e.form(n)
		})
}

var primaryTests = []emitTests{
	{"what", []styling{{0, 4, styleForPrimary[parse.Bareword]}}},
	{"$var", []styling{{0, 4, styleForPrimary[parse.Variable]}}},
	{"'a'", []styling{{0, 3, styleForPrimary[parse.SingleQuoted]}}},
	{`"x"`, []styling{{0, 3, styleForPrimary[parse.DoubleQuoted]}}},
}

func TestPrimary(t *testing.T) {
	test(t, "primary", primaryTests,
		func(e *Emitter, src string) {
			n := &parse.Primary{}
			parse.As("<test>", src, n)
			e.primary(n)
		})
}

var sepTests = []emitTests{
	{">", []styling{{0, 1, styleForSep[">"]}}},
	{"# comment", []styling{{0, 9, styleForComment}}},
}

func TestSep(t *testing.T) {
	test(t, "sep", sepTests,
		func(e *Emitter, src string) {
			e.sep(parse.NewSep(src, 0, len(src)))
		})
}

func test(t *testing.T, what string, tests []emitTests, f func(*Emitter, string)) {

	for _, test := range tests {
		var stylings []styling
		e := &Emitter{goodFormHead, func(b, e int, s string) {
			stylings = append(stylings, styling{b, e, s})
		}}

		f(e, test.source)

		if !reflect.DeepEqual(stylings, test.wantStylings) {
			t.Errorf("%s %q gets stylings %v, want %v", what, test.source,
				stylings, test.wantStylings)
		}
	}
}
