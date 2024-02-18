package parse

import (
	"bytes"
	"unicode"
	"unicode/utf8"
)

// Quote returns a valid Elvish expression that evaluates to the given string.
// If s is a valid bareword, it is returned as is; otherwise it is quoted,
// preferring the use of single quotes.
func Quote(s string) string {
	s, _ = QuoteAs(s, Bareword)
	return s
}

// QuoteVariableName is like [Quote], but quotes s if it contains any character
// that may not appear unquoted in variable names.
func QuoteVariableName(s string) string {
	if s == "" {
		return "''"
	}

	// Keep track of whether it is a valid (unquoted) variable name.
	bare := true
	for _, r := range s {
		if r == unicode.ReplacementChar || !unicode.IsPrint(r) {
			// Contains invalid UTF-8 sequence or unprintable character; force
			// double quote.
			return quoteDouble(s)
		}
		if !allowedInVariableName(r) {
			bare = false
		}
	}

	if bare {
		return s
	}
	return quoteSingle(s)
}

// QuoteCommandName is like [Quote], but uses the slightly laxer rule for what
// can appear in a command name unquoted, like <.
func QuoteCommandName(s string) string {
	q, _ := quoteAs(s, Bareword, CmdExpr)
	return q
}

// QuoteAs returns a representation of s in Elvish syntax, preferring the syntax
// specified by q, which must be one of Bareword, SingleQuoted, or DoubleQuoted.
// It returns the quoted string and the actual quoting.
func QuoteAs(s string, q PrimaryType) (string, PrimaryType) {
	return quoteAs(s, q, strictExpr)
}

func quoteAs(s string, q PrimaryType, ctx ExprCtx) (string, PrimaryType) {
	if q == DoubleQuoted {
		// Everything can be quoted using double quotes, return directly.
		return quoteDouble(s), DoubleQuoted
	}
	if s == "" {
		return "''", SingleQuoted
	}

	// Keep track of whether it is a valid bareword.
	bare := s[0] != '~'
	for _, r := range s {
		if r == unicode.ReplacementChar || !unicode.IsPrint(r) {
			// Contains invalid UTF-8 sequence or unprintable character; force
			// double quote.
			return quoteDouble(s), DoubleQuoted
		}
		if !allowedInBareword(r, ctx) {
			bare = false
		}
	}

	if q == Bareword && bare {
		return s, Bareword
	}
	return quoteSingle(s), SingleQuoted
}

func quoteSingle(s string) string {
	var buf bytes.Buffer
	buf.WriteByte('\'')
	for _, r := range s {
		buf.WriteRune(r)
		if r == '\'' {
			buf.WriteByte('\'')
		}
	}
	buf.WriteByte('\'')
	return buf.String()
}

// rtohex is optimized for the common cases encountered when encoding Elvish strings and should be
// more efficient than using fmt.Sprintf("%x").
func rtohex(r rune, w int) []byte {
	bytes := make([]byte, w)
	for i := w - 1; i >= 0; i-- {
		d := byte(r % 16)
		r /= 16
		if d <= 9 {
			bytes[i] = '0' + d
		} else {
			bytes[i] = 'a' + d - 10
		}
	}
	return bytes
}

func quoteDouble(s string) string {
	var buf bytes.Buffer
	buf.WriteByte('"')
	for s != "" {
		r, w := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError && w == 1 {
			// An invalid UTF-8 sequence was seen -- encode first byte as a hex literal.
			buf.WriteByte('\\')
			buf.WriteByte('x')
			buf.Write(rtohex(rune(s[0]), 2))
		} else if e, ok := doubleUnescape[r]; ok {
			// This handles the escaping of " and \ too.
			buf.WriteByte('\\')
			buf.WriteRune(e)
		} else if unicode.IsPrint(r) && r != utf8.RuneError {
			// RuneError is technically printable, but don't print it directly
			// to avoid confusion.
			buf.WriteRune(r)
		} else if r <= 0x7f {
			// Unprintable characters in the ASCII range can be escaped with \x
			// since they are one byte in UTF-8.
			buf.WriteByte('\\')
			buf.WriteByte('x')
			buf.Write(rtohex(r, 2))
		} else if r <= 0xffff {
			buf.WriteByte('\\')
			buf.WriteByte('u')
			buf.Write(rtohex(r, 4))
		} else {
			buf.WriteByte('\\')
			buf.WriteByte('U')
			buf.Write(rtohex(r, 8))
		}
		s = s[w:]
	}
	buf.WriteByte('"')
	return buf.String()
}
