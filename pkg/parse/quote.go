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

// QuoteVariableName is like Quote, but quotes s if it contains any character
// that may not appear unquoted in variable names.
func QuoteVariableName(s string) string {
	if s == "" {
		return "''"
	}

	// Keep track of whether it is a valid (unquoted) variable name.
	bare := true
	for s2 := s; len(s2) > 0; {
		r, w := utf8.DecodeRuneInString(s2)
		if r == utf8.RuneError || !unicode.IsPrint(r) {
			// Contains unprintable character; force double quote.
			return quoteDouble(s)
		}
		if !allowedInVariableName(r) {
			bare = false
			break
		}
		s2 = s2[w:]
	}

	if bare {
		return s
	}
	return quoteSingle(s)
}

// QuoteAs returns a representation of s in elvish syntax, preferring the syntax
// specified by q, which must be one of Bareword, SingleQuoted, or DoubleQuoted.
// It returns the quoted string and the actual quoting.
func QuoteAs(s string, q PrimaryType) (string, PrimaryType) {
	if q == DoubleQuoted {
		// Everything can be quoted using double quotes, return directly.
		return quoteDouble(s), DoubleQuoted
	}
	if s == "" {
		return "''", SingleQuoted
	}

	// Keep track of whether it is a valid bareword.
	bare := s[0] != '~'
	for s2 := s; len(s2) > 0; {
		r, w := utf8.DecodeRuneInString(s2)
		if r == utf8.RuneError || !unicode.IsPrint(r) {
			// Contains unprintable character; force double quote.
			return quoteDouble(s), DoubleQuoted
		}
		if !allowedInBareword(r, strictExpr) {
			bare = false
		}
		s2 = s2[w:]
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
	for len(s) > 0 {
		r, w := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError {
			// An invalid UTF-8 sequence was seen -- encode first byte as a hex literal.
			buf.WriteByte('\\')
			buf.WriteByte('x')
			buf.Write(rtohex(rune(s[0]), 2))
			s = s[w:]
			continue
		}

		// The common case -- a valid UTF-8 sequence was decoded so we have a rune.
		if e, ok := doubleUnescape[r]; ok {
			// Handle " and \ as well as things like \n.
			buf.WriteByte('\\')
			buf.WriteRune(e)
		} else if unicode.IsPrint(r) {
			buf.WriteRune(r)
		} else if r <= ' ' {
			// Control char not handled by doubleUnescape.
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
