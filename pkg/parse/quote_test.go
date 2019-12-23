package parse

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

var quoteTests = tt.Table{
	// Empty string is single-quoted.
	tt.Args("").Rets(`''`),

	// Bareword when possible.
	tt.Args("x-y:z@h/d").Rets("x-y:z@h/d"),

	// Single quote when there are special characters but no unprintable
	// characters.
	tt.Args("x$y[]ef'").Rets("'x$y[]ef'''"),

	// Tilde needs quoting only leading the expression.
	tt.Args("~x").Rets("'~x'"),
	tt.Args("x~").Rets("x~"),

	// Double quote when there is unprintable char.
	tt.Args("a\nb").Rets(`"a\nb"`),
	tt.Args("\x1b\"\\").Rets(`"\e\"\\"`),

	// Commas and equal signs are always quoted, so that the quoted string is
	// safe for use everywhere.
	tt.Args("a,b").Rets(`'a,b'`),
	tt.Args("a=b").Rets(`'a=b'`),
}

func TestQuote(t *testing.T) {
	tt.Test(t, tt.Fn("Quote", Quote).ArgsFmt("(%q)").RetsFmt("%q"), quoteTests)
}
