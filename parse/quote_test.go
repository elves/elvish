package parse

import (
	"testing"

	"github.com/elves/elvish/tt"
)

var quoteTests = tt.Table{
	// Empty string is single-quoted.
	tt.C("").Rets(`''`),

	// Bareword when possible.
	tt.C("x-y:z@h/d").Rets("x-y:z@h/d"),

	// Single quote when there are special characters but no unprintable
	// characters.
	tt.C("x$y[]ef'").Rets("'x$y[]ef'''"),

	// Tilde needs quoting only leading the expression.
	tt.C("~x").Rets("'~x'"),
	tt.C("x~").Rets("x~"),

	// Double quote when there is unprintable char.
	tt.C("a\nb").Rets(`"a\nb"`),
	tt.C("\x1b\"\\").Rets(`"\e\"\\"`),

	// Commas and equal signs are always quoted, so that the quoted string is
	// safe for use everywhere.
	tt.C("a,b").Rets(`'a,b'`),
	tt.C("a=b").Rets(`'a=b'`),
}

func TestQuote(t *testing.T) {
	tt.Test(t, tt.Fn{"Quote", "(%q)", "%q", Quote}, quoteTests)
}
