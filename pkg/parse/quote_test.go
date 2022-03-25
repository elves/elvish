package parse

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestQuote(t *testing.T) {
	tt.Test(t, tt.Fn("Quote", Quote).ArgsFmt("(%q)"), tt.Table{
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
		tt.Args("\x00").Rets(`"\x00"`),
		tt.Args("\u0600").Rets(`"\u0600"`),         // Arabic number sign
		tt.Args("\U000110BD").Rets(`"\U000110bd"`), // Kathi number sign

		// Commas and equal signs are always quoted, so that the quoted string is
		// safe for use everywhere.
		tt.Args("a,b").Rets(`'a,b'`),
		tt.Args("a=b").Rets(`'a=b'`),
	})
}

func TestQuoteAs(t *testing.T) {
	tt.Test(t, tt.Fn("QuoteAs", QuoteAs).ArgsFmt("(%q, %s)"), tt.Table{
		// DoubleQuote is always respected.
		tt.Args("", DoubleQuoted).Rets(`""`, DoubleQuoted),
		tt.Args("a", DoubleQuoted).Rets(`"a"`, DoubleQuoted),

		// SingleQuoted is respected when there is no unprintable character.
		tt.Args("", SingleQuoted).Rets(`''`, SingleQuoted),
		tt.Args("a", SingleQuoted).Rets(`'a'`, SingleQuoted),
		tt.Args("\n", SingleQuoted).Rets(`"\n"`, DoubleQuoted),

		// Bareword tested above in TestQuote.
	})
}

func TestQuoteVariableName(t *testing.T) {
	tt.Test(t, tt.Fn("QuoteVariableName", QuoteVariableName).ArgsFmt("(%q)"), tt.Table{
		tt.Args("").Rets("''"),
		tt.Args("foo").Rets("foo"),
		tt.Args("a/b").Rets("'a/b'"),
		tt.Args("\x1b").Rets(`"\e"`),
	})
}
