package parse

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"src.elv.sh/pkg/diag"
)

// parser maintains some mutable states of parsing.
//
// NOTE: The src member is assumed to be valid UF-8.
type parser struct {
	srcName string
	src     string
	pos     int
	overEOF int
	errors  []*Error
	warn    io.Writer
}

// Error is a parse error.
type Error = diag.Error[ErrorTag]

// ErrorTag parameterizes [diag.Error] to define [Error].
type ErrorTag struct{}

func (ErrorTag) ErrorTag() string { return "parse error" }

func parse[N Node](ps *parser, n N) parsed[N] {
	begin := ps.pos
	n.n().From = begin
	n.parse(ps)
	n.n().To = ps.pos
	n.n().sourceText = ps.src[begin:ps.pos]
	return parsed[N]{n}
}

type parsed[N Node] struct {
	n N
}

func (p parsed[N]) addAs(ptr *N, parent Node) {
	*ptr = p.n
	addChild(parent, p.n)
}

func (p parsed[N]) addTo(ptr *[]N, parent Node) {
	*ptr = append(*ptr, p.n)
	addChild(parent, p.n)
}

func addChild(p Node, ch Node) {
	p.n().addChild(ch)
	ch.n().parent = p
}

// Tells the parser that parsing is done.
func (ps *parser) done() {
	if ps.pos != len(ps.src) {
		r, _ := utf8.DecodeRuneInString(ps.src[ps.pos:])
		ps.error(fmt.Errorf("unexpected rune %q", r))
	}
}

const eof rune = -1

func (ps *parser) peek() rune {
	if ps.pos == len(ps.src) {
		return eof
	}
	r, _ := utf8.DecodeRuneInString(ps.src[ps.pos:])
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

func (ps *parser) errorp(r diag.Ranger, e error) {
	err := &Error{
		Message: e.Error(),
		Context: *diag.NewContext(ps.srcName, ps.src, r),
		Partial: r.Range().From == len(ps.src),
	}
	ps.errors = append(ps.errors, err)
}

func (ps *parser) error(e error) {
	end := ps.pos
	if end < len(ps.src) {
		end++
	}
	ps.errorp(diag.Ranging{From: ps.pos, To: end}, e)
}

// UnpackErrors returns the constituent parse errors if the given error contains
// one or more parse errors. Otherwise it returns nil.
func UnpackErrors(e error) []*Error {
	if errs := diag.UnpackErrors[ErrorTag](e); len(errs) > 0 {
		return errs
	}
	return nil
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
