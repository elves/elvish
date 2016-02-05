package parse

import (
	"bytes"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/errutil"
)

// parser maintains some mutable states of parsing.
//
// NOTE: The str member is assumed to be valid UF-8.
type parser struct {
	src     string
	pos     int
	overEOF int
	errors  *errutil.Errors
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

func (ps *parser) peek() rune {
	if ps.pos == len(ps.src) {
		return EOF
	}
	r, _ := utf8.DecodeRuneInString(ps.src[ps.pos:])
	return r
}

func (ps *parser) hasPrefix(prefix string) bool {
	return strings.HasPrefix(ps.src[ps.pos:], prefix)
}

func (ps *parser) next() rune {
	if ps.pos == len(ps.src) {
		ps.overEOF += 1
		return EOF
	}
	r, s := utf8.DecodeRuneInString(ps.src[ps.pos:])
	ps.pos += s
	return r
}

func (ps *parser) backup() {
	if ps.overEOF > 0 {
		ps.overEOF -= 1
		return
	}
	_, s := utf8.DecodeLastRuneInString(ps.src[:ps.pos])
	ps.pos -= s
}

func (ps *parser) error(e error) {
	if ps.errors == nil {
		ps.errors = &errutil.Errors{}
	}
	ps.errors.Append(&errutil.PosError{ps.pos, ps.pos, e})
}
