package parse

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func TestQuote(t *testing.T) {
	tt.Test(t, tt.Fn(Quote).ArgsFmt("(%q)"),
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
		Args("\x00").Rets(`"\x00"`),
		Args("\x7f").Rets(`"\x7f"`),
		Args("\u0090").Rets(`"\u0090"`),
		Args("\u0600").Rets(`"\u0600"`),         // Arabic number sign
		Args("\ufffd").Rets(`"\ufffd"`),         // Unicode replacement character
		Args("\U000110BD").Rets(`"\U000110bd"`), // Kathi number sign

		// String containing characters that can be single-quoted are
		// double-quoted when it also contains unprintable characters.
		Args("$\n").Rets(`"$\n"`),

		// Commas and equal signs are always quoted, so that the quoted string is
		// safe for use everywhere.
		Args("a,b").Rets(`'a,b'`),
		Args("a=b").Rets(`'a=b'`),

		// Double quote strings containing invalid UTF-8 sequences with \x.
		Args("bad\xffUTF-8").Rets(`"bad\xffUTF-8"`),
	)
}

func TestQuoteAs(t *testing.T) {
	tt.Test(t, tt.Fn(QuoteAs).ArgsFmt("(%q, %s)"),
		// DoubleQuote is always respected.
		Args("", DoubleQuoted).Rets(`""`, DoubleQuoted),
		Args("a", DoubleQuoted).Rets(`"a"`, DoubleQuoted),

		// SingleQuoted is respected when there is no unprintable character.
		Args("", SingleQuoted).Rets(`''`, SingleQuoted),
		Args("a", SingleQuoted).Rets(`'a'`, SingleQuoted),
		Args("\n", SingleQuoted).Rets(`"\n"`, DoubleQuoted),

		// Bareword tested above in TestQuote.
	)
}

func TestQuoteVariableName(t *testing.T) {
	tt.Test(t, tt.Fn(QuoteVariableName).ArgsFmt("(%q)"),
		Args("").Rets("''"),
		Args("foo").Rets("foo"),
		Args("a/b").Rets("'a/b'"),
		Args("\x1b").Rets(`"\e"`),
		Args("bad\xffUTF-8").Rets(`"bad\xffUTF-8"`),
		Args("$\n").Rets(`"$\n"`),
	)
}

func TestQuoteCommandName(t *testing.T) {
	tt.Test(t, tt.Fn(QuoteCommandName).ArgsFmt("(%q)"),
		Args("<").Rets("<"),
		Args("foo").Rets("foo"),
		Args("$").Rets(`'$'`),
	)
}
