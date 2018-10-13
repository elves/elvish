package parse

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/diag"
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

func (ps *Parser) parse(n Node) parsed {
	begin := ps.pos
	n.setBegin(begin)
	n.parse(ps)
	n.setEnd(ps.pos)
	n.setSourceText(ps.src[begin:ps.pos])
	return parsed{n}
}

var nodeType = reflect.TypeOf((*Node)(nil)).Elem()

type parsed struct {
	n Node
}

func (p parsed) addAs(ptr interface{}, parent Node) {
	dst := reflect.ValueOf(ptr).Elem()
	dst.Set(reflect.ValueOf(p.n)) // *ptr = p.n
	addChild(parent, p.n)
}

func (p parsed) addTo(ptr interface{}, parent Node) {
	dst := reflect.ValueOf(ptr).Elem()
	dst.Set(reflect.Append(dst, reflect.ValueOf(p.n))) // *ptr = append(*ptr, n)
	addChild(parent, p.n)
}

// Tells the parser that parsing is done.
func (ps *Parser) done() {
	if ps.pos != len(ps.src) {
		r, _ := utf8.DecodeRuneInString(ps.src[ps.pos:])
		ps.error(fmt.Errorf("unexpected rune %q", r))
	}
}

// Assembles all parsing errors as one, or returns nil if there were no errors.
func (ps *Parser) assembleError() error {
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
	return r
}

func (ps *Parser) hasPrefix(prefix string) bool {
	return strings.HasPrefix(ps.src[ps.pos:], prefix)
}

func (ps *Parser) next() rune {
	if ps.pos == len(ps.src) {
		ps.overEOF++
		return eof
	}
	r, s := utf8.DecodeRuneInString(ps.src[ps.pos:])
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

func (ps *Parser) errorp(begin, end int, e error) {
	ps.errors.Add(e.Error(), diag.NewSourceRange(ps.srcName, ps.src, begin, end))
}

func (ps *Parser) error(e error) {
	end := ps.pos
	if end < len(ps.src) {
		end++
	}
	ps.errorp(ps.pos, end, e)
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
