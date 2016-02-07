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
	cutsets []map[rune]int
	errors  *errutil.Errors
}

const eof rune = -1

func (ps *parser) eof() bool {
	return ps.peek() == eof
}

func (ps *parser) peek() rune {
	if ps.pos == len(ps.src) {
		return eof
	}
	r, _ := utf8.DecodeRuneInString(ps.src[ps.pos:])
	if ps.currentCutset()[r] > 0 {
		return eof
	}
	return r
}

func (ps *parser) hasPrefix(prefix string) bool {
	return strings.HasPrefix(ps.src[ps.pos:], prefix)
}

func (ps *parser) next() rune {
	if ps.pos == len(ps.src) {
		ps.overEOF++
		return eof
	}
	r, s := utf8.DecodeRuneInString(ps.src[ps.pos:])
	if ps.currentCutset()[r] > 0 {
		return eof
	}
	ps.pos += s
	return r
}

func (ps *parser) backup() {
	if ps.overEOF > 0 {
		ps.overEOF--
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

func (ps *parser) pushCutset(rs ...rune) {
	ps.cutsets = append(ps.cutsets, map[rune]int{})
	ps.cut(rs...)
}

func (ps *parser) popCutset() {
	n := len(ps.cutsets)
	ps.cutsets[n-1] = nil
	ps.cutsets = ps.cutsets[:n-1]
}

func (ps *parser) currentCutset() map[rune]int {
	return ps.cutsets[len(ps.cutsets)-1]
}

func (ps *parser) cut(rs ...rune) {
	cutset := ps.currentCutset()
	for _, r := range rs {
		cutset[r]++
	}
}

func (ps *parser) uncut(rs ...rune) {
	cutset := ps.currentCutset()
	for _, r := range rs {
		cutset[r]--
	}
}

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
