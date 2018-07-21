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
		func(e *Emitter, ps *parse.Parser) {
			e.EmitAll(parse.ParseChunk(ps))
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
		func(e *Emitter, ps *parse.Parser) {
			e.form(parse.ParseForm(ps))
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
		func(e *Emitter, ps *parse.Parser) {
			e.primary(parse.ParsePrimary(ps, parse.NormalExpr))
		})
}

var sepTests = []emitTests{
	{">", []styling{{0, 1, styleForSep[">"]}}},
	{"# comment", []styling{{0, 9, styleForComment}}},
}

func TestSep(t *testing.T) {
	test(t, "sep", sepTests,
		func(e *Emitter, ps *parse.Parser) {
			src := ps.Source()
			e.sep(parse.NewSep(src, 0, len(src)))
		})
}

func test(t *testing.T, what string,
	tests []emitTests, f func(*Emitter, *parse.Parser)) {

	for _, test := range tests {
		var stylings []styling
		e := &Emitter{goodFormHead, func(b, e int, s string) {
			stylings = append(stylings, styling{b, e, s})
		}}
		ps := parse.NewParser("<test>", test.source)

		f(e, ps)

		if !reflect.DeepEqual(stylings, test.wantStylings) {
			t.Errorf("%s %q gets stylings %v, want %v", what, test.source,
				stylings, test.wantStylings)
		}
	}
}
