package parse

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestQuote(t *testing.T) {
	tt.Test(t, tt.Fn("Quote", Quote).ArgsFmt("(%q)"), tt.Table{
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
		Args("\000\x1b\"\\").Rets(`"\x00\e\"\\"`),
		Args("\u0600").Rets(`"\u0600"`),         // Arabic number sign
		Args("\U000110BD").Rets(`"\U000110bd"`), // Kathi number sign

		// Commas and equal signs are always quoted, so that the quoted string is
		// safe for use everywhere.
		Args("a,b").Rets(`'a,b'`),
		Args("a=b").Rets(`'a=b'`),
	})
}

func TestQuoteAs(t *testing.T) {
	tt.Test(t, tt.Fn("QuoteAs", QuoteAs).ArgsFmt("(%q, %s)"), tt.Table{
		// DoubleQuote is always respected.
		Args("", DoubleQuoted).Rets(`""`, DoubleQuoted),
		Args("a", DoubleQuoted).Rets(`"a"`, DoubleQuoted),

		// SingleQuoted is respected when there is no unprintable character.
		Args("", SingleQuoted).Rets(`''`, SingleQuoted),
		Args("a", SingleQuoted).Rets(`'a'`, SingleQuoted),
		Args("\n", SingleQuoted).Rets(`"\n"`, DoubleQuoted),

		// Verify bareword invalid UTF-8 case.
		Args("bad\xffUTF-8", Bareword).Rets(`"bad\xffUTF-8"`, DoubleQuoted),

		// Bareword tested above in TestQuote.
	})
}

func TestQuoteVariableName(t *testing.T) {
	tt.Test(t, tt.Fn("QuoteVariableName", QuoteVariableName).ArgsFmt("(%q)"), tt.Table{
		Args("").Rets("''"),
		Args("foo").Rets("foo"),
		Args("a/b").Rets("'a/b'"),
		Args("\x1b").Rets(`"\e"`),
		Args("bad\xffUTF-8").Rets(`"bad\xffUTF-8"`),
	})
}
