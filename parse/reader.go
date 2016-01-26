package parse

import (
	"bytes"
	"errors"
	"strings"
	"unicode/utf8"
)

// reader provides helpers for maintaining a current position within a string.
//
// NOTE: The str member is assumed to be valid UF-8.
type reader struct {
	src     string
	pos     int
	overEOF int
	error   error
}

const (
	EOF rune = -1 - iota
	ParseError
)

func newError(text string, shouldbe ...string) error {
	if len(shouldbe) == 0 {
		return errors.New(text)
	}
	var buf bytes.Buffer
	if len(text) > 0 {
		buf.WriteString(text + ", ")
	}
	buf.WriteString("should be " + shouldbe[0])
	for i, opt := range shouldbe[1:] {
		if i == len(shouldbe)-2 {
			buf.WriteString(" or ")
		} else {
			buf.WriteString(", ")
		}
		buf.WriteString(opt)
	}
	return errors.New(buf.String())
}

func (rd *reader) peek() rune {
	if rd.error != nil {
		return ParseError
	}
	if rd.pos == len(rd.src) {
		return EOF
	}
	r, _ := utf8.DecodeRuneInString(rd.src[rd.pos:])
	return r
}

func (rd *reader) hasPrefix(prefix string) bool {
	return strings.HasPrefix(rd.src[rd.pos:], prefix)
}

func (rd *reader) next() rune {
	if rd.error != nil {
		return ParseError
	}
	if rd.pos == len(rd.src) {
		rd.overEOF += 1
		return EOF
	}
	r, s := utf8.DecodeRuneInString(rd.src[rd.pos:])
	rd.pos += s
	return r
}

func (rd *reader) backup() {
	if rd.overEOF > 0 {
		rd.overEOF -= 1
		return
	}
	_, s := utf8.DecodeLastRuneInString(rd.src[:rd.pos])
	rd.pos -= s
}
