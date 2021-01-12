package parse

import (
	"testing"

	. "github.com/elves/elvish/pkg/tt"
)

func TestQuote(t *testing.T) {
	Test(t, Fn("Quote", Quote).ArgsFmt("(%q)").RetsFmt("%q"), Table{
		// Empty string is single-quoted.
		Args("").Rets(`''`),

		// Bareword when possible.
		Args("x-y:z@h/d").Rets("x-y:z@h/d"),

		// Single quote when there are special characters but no unprintable
		// characters.
		Args("x$y[]ef'").Rets("'x$y[]ef'''"),

		// Tilde needs quoting only leading the expression.
		Args("~x").Rets("'~x'"),
		Args("x~").Rets("x~"),

		// Double quote when there is unprintable char.
		Args("a\nb").Rets(`"a\nb"`),
		Args("\x1b\"\\").Rets(`"\e\"\\"`),

		// Commas and equal signs are always quoted, so that the quoted string is
		// safe for use everywhere.
		Args("a,b").Rets(`'a,b'`),
		Args("a=b").Rets(`'a=b'`),
	})
}

func TestQuoteVariableName(t *testing.T) {
	Test(t, Fn("QuoteVariableName", QuoteVariableName).ArgsFmt("(%q)").RetsFmt("%q"), Table{
		Args("").Rets("''"),
		Args("foo").Rets("foo"),
		Args("a/b").Rets("'a/b'"),
		Args("\x1b").Rets(`"\e"`),
	})
}
