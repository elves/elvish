package parse

import (
	"bytes"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/util"
)

// Parser maintains some mutable states of parsing.
//
// NOTE: The str member is assumed to be valid UF-8.
type Parser struct {
	srcName string
	src     string
	pos     int
	overEOF int
	cutsets []map[rune]int
	errors  Error
}

// NewParser creates a new parser from a piece of source text and its name.
func NewParser(srcname, src string) *Parser {
	return &Parser{srcname, src, 0, 0, []map[rune]int{{}}, Error{}}
}

// Done tells the parser that parsing has completed.
func (ps *Parser) Done() {
	if ps.pos != len(ps.src) {
		ps.error(errUnexpectedRune)
	}
}

// Errors gets the parsing errors after calling one of the parse* functions. If
// the return value is not nil, it is always of type Error.
func (ps *Parser) Errors() error {
	if len(ps.errors.Entries) > 0 {
		return &ps.errors
	}
	return nil
}

// Source returns the source code that is being parsed.
func (ps *Parser) Source() string {
	return ps.src
}

const eof rune = -1

func (ps *Parser) peek() rune {
	if ps.pos == len(ps.src) {
		return eof
	}
	r, _ := utf8.DecodeRuneInString(ps.src[ps.pos:])
	if ps.currentCutset()[r] > 0 {
		return eof
	}
	return r
}

func (ps *Parser) hasPrefix(prefix string) bool {
	return strings.HasPrefix(ps.src[ps.pos:], prefix)
}

// findWord looks ahead for [a-z]* that is also a valid compound. If the
// lookahead fails, it returns an empty string. It is useful for looking for
// command leaders.
func (ps *Parser) findPossibleLeader() string {
	rest := ps.src[ps.pos:]
	i := strings.IndexFunc(rest, func(r rune) bool {
		return r < 'a' || r > 'z'
	})
	if i == -1 {
		// The whole rest is just one possible leader.
		return rest
	}
	r, _ := utf8.DecodeRuneInString(rest[i:])
	if startsPrimary(r, false) {
		return ""
	}
	return rest[:i]
}

func (ps *Parser) next() rune {
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

func (ps *Parser) backup() {
	if ps.overEOF > 0 {
		ps.overEOF--
		return
	}
	_, s := utf8.DecodeLastRuneInString(ps.src[:ps.pos])
	ps.pos -= s
}

func (ps *Parser) advance(c int) {
	ps.pos += c
	if ps.pos > len(ps.src) {
		ps.overEOF = ps.pos - len(ps.src)
		ps.pos = len(ps.src)
	}
}

func (ps *Parser) errorp(begin, end int, e error) {
	ps.errors.Add(e.Error(), util.SourceContext{ps.srcName, ps.src, begin, end, nil})
}

func (ps *Parser) error(e error) {
	end := ps.pos
	if end < len(ps.src) {
		end++
	}
	ps.errorp(ps.pos, end, e)
}

func (ps *Parser) pushCutset(rs ...rune) {
	ps.cutsets = append(ps.cutsets, map[rune]int{})
	ps.cut(rs...)
}

func (ps *Parser) popCutset() {
	n := len(ps.cutsets)
	ps.cutsets[n-1] = nil
	ps.cutsets = ps.cutsets[:n-1]
}

func (ps *Parser) currentCutset() map[rune]int {
	return ps.cutsets[len(ps.cutsets)-1]
}

func (ps *Parser) cut(rs ...rune) {
	cutset := ps.currentCutset()
	for _, r := range rs {
		cutset[r]++
	}
}

func (ps *Parser) uncut(rs ...rune) {
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
