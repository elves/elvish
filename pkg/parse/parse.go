// Package parse implements the elvish parser.
//
// The parser builds a hybrid of AST (abstract syntax tree) and parse tree
// (a.k.a. concrete syntax tree). The AST part only includes parts that are
// semantically significant -- i.e. skipping whitespaces and symbols that do not
// alter the semantics, and is embodied in the fields of each *Node type. The
// parse tree part corresponds to all the text in the original source text, and
// is embodied in the children of each *Node type.
package parse

//go:generate stringer -type=PrimaryType,RedirMode,ExprCtx -output=string.go

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strings"
	"unicode"

	"src.elv.sh/pkg/diag"
)

// Tree represents a parsed tree.
type Tree struct {
	Root   *Chunk
	Source Source
}

// Config keeps configuration options when parsing.
type Config struct {
	// Destination of warnings. If nil, warnings are suppressed.
	WarningWriter io.Writer
}

// Parse parses the given source. The returned error always has type *Error
// if it is not nil.
func Parse(src Source, cfg Config) (Tree, error) {
	tree := Tree{&Chunk{}, src}
	err := ParseAs(src, tree.Root, cfg)
	return tree, err
}

// ParseAs parses the given source as a node, depending on the dynamic type of
// n. If the error is not nil, it always has type *Error.
func ParseAs(src Source, n Node, cfg Config) error {
	ps := &parser{srcName: src.Name, src: src.Code, warn: cfg.WarningWriter}
	ps.parse(n)
	ps.done()
	return ps.assembleError()
}

// Errors.
var (
	errShouldBeForm               = newError("", "form")
	errBadRedirSign               = newError("bad redir sign", "'<'", "'>'", "'>>'", "'<>'")
	errShouldBeFD                 = newError("", "a composite term representing fd")
	errShouldBeFilename           = newError("", "a composite term representing filename")
	errShouldBeArray              = newError("", "spaced")
	errStringUnterminated         = newError("string not terminated")
	errInvalidEscape              = newError("invalid escape sequence")
	errInvalidEscapeOct           = newError("invalid escape sequence", "octal digit")
	errInvalidEscapeOctOverflow   = newError("invalid octal escape sequence", "below 256")
	errInvalidEscapeHex           = newError("invalid escape sequence", "hex digit")
	errInvalidEscapeControl       = newError("invalid control sequence", "a codepoint between 0x3F and 0x5F")
	errShouldBePrimary            = newError("", "single-quoted string", "double-quoted string", "bareword")
	errShouldBeVariableName       = newError("", "variable name")
	errShouldBeRBracket           = newError("", "']'")
	errShouldBeRBrace             = newError("", "'}'")
	errShouldBeBraceSepOrRBracket = newError("", "','", "'}'")
	errShouldBeRParen             = newError("", "')'")
	errShouldBeCompound           = newError("", "compound")
	errShouldBeEqual              = newError("", "'='")
	errShouldBePipe               = newError("", "'|'")
	errBothElementsAndPairs       = newError("cannot contain both list elements and map pairs")
	errShouldBeNewline            = newError("", "newline")
)

// Chunk = { PipelineSep | Space } { Pipeline { PipelineSep | Space } }
type Chunk struct {
	node
	Pipelines []*Pipeline
}

func (bn *Chunk) parse(ps *parser) {
	bn.parseSeps(ps)
	for startsPipeline(ps.peek()) {
		ps.parse(&Pipeline{}).addTo(&bn.Pipelines, bn)
		if bn.parseSeps(ps) == 0 {
			break
		}
	}
}

func isPipelineSep(r rune) bool {
	return r == '\r' || r == '\n' || r == ';'
}

// parseSeps parses pipeline separators along with whitespaces. It returns the
// number of pipeline separators parsed.
func (bn *Chunk) parseSeps(ps *parser) int {
	nseps := 0
	for {
		r := ps.peek()
		if isPipelineSep(r) {
			// parse as a Sep
			parseSep(bn, ps, r)
			nseps++
		} else if IsInlineWhitespace(r) || r == '#' {
			// parse a run of spaces as a Sep
			parseSpaces(bn, ps)
		} else {
			break
		}
	}
	return nseps
}

// Pipeline = Form { '|' Form }
type Pipeline struct {
	node
	Forms      []*Form
	Background bool
}

func (pn *Pipeline) parse(ps *parser) {
	ps.parse(&Form{}).addTo(&pn.Forms, pn)
	for parseSep(pn, ps, '|') {
		parseSpacesAndNewlines(pn, ps)
		if !startsForm(ps.peek()) {
			ps.error(errShouldBeForm)
			return
		}
		ps.parse(&Form{}).addTo(&pn.Forms, pn)
	}
	parseSpaces(pn, ps)
	if ps.peek() == '&' {
		ps.next()
		addSep(pn, ps)
		pn.Background = true
		parseSpaces(pn, ps)
	}
}

func startsPipeline(r rune) bool {
	return startsForm(r)
}

// Form = { Space } { { Assignment } { Space } }
//
//	{ Compound } { Space } { ( Compound | MapPair | Redir ) { Space } }
type Form struct {
	node
	Assignments []*Assignment
	Head        *Compound
	Args        []*Compound
	Opts        []*MapPair
	Redirs      []*Redir
}

func (fn *Form) parse(ps *parser) {
	parseSpaces(fn, ps)
	for startsCompound(ps.peek(), CmdExpr) {
		initial := ps.save()
		cmdNode := &Compound{ExprCtx: CmdExpr}
		parsedCmd := ps.parse(cmdNode)

		if !parsableAsAssignment(cmdNode) {
			parsedCmd.addAs(&fn.Head, fn)
			parseSpaces(fn, ps)
			break
		}
		ps.restore(initial)
		ps.parse(&Assignment{}).addTo(&fn.Assignments, fn)
		parseSpaces(fn, ps)
	}

	if fn.Head == nil {
		if len(fn.Assignments) > 0 {
			// Assignment-only form.
			return
		}
		// Bad form.
		ps.error(fmt.Errorf("bad rune at form head: %q", ps.peek()))
	}

	for {
		r := ps.peek()
		switch {
		case r == '&':
			ps.next()
			hasMapPair := startsCompound(ps.peek(), LHSExpr)
			ps.backup()
			if !hasMapPair {
				// background indicator
				return
			}
			ps.parse(&MapPair{}).addTo(&fn.Opts, fn)
		case startsCompound(r, NormalExpr):
			cn := &Compound{}
			ps.parse(cn)
			if isRedirSign(ps.peek()) {
				// Redir
				ps.parse(&Redir{Left: cn}).addTo(&fn.Redirs, fn)
			} else {
				parsed{cn}.addTo(&fn.Args, fn)
			}
		case isRedirSign(r):
			ps.parse(&Redir{}).addTo(&fn.Redirs, fn)
		default:
			return
		}
		parseSpaces(fn, ps)
	}
}

func parsableAsAssignment(cn *Compound) bool {
	if len(cn.Indexings) == 0 {
		return false
	}
	switch cn.Indexings[0].Head.Type {
	case Braced, SingleQuoted, DoubleQuoted:
		return len(cn.Indexings) >= 2 &&
			strings.HasPrefix(SourceText(cn.Indexings[1]), "=")
	case Bareword:
		name := cn.Indexings[0].Head.Value
		eq := strings.IndexByte(name, '=')
		if eq >= 0 {
			return validBarewordVariableName(name[:eq], true)
		} else {
			return validBarewordVariableName(name, true) &&
				len(cn.Indexings) >= 2 &&
				strings.HasPrefix(SourceText(cn.Indexings[1]), "=")
		}
	default:
		return false
	}
}

func startsForm(r rune) bool {
	return IsInlineWhitespace(r) || startsCompound(r, CmdExpr)
}

// Assignment = Indexing '=' Compound
type Assignment struct {
	node
	Left  *Indexing
	Right *Compound
}

func (an *Assignment) parse(ps *parser) {
	ps.parse(&Indexing{ExprCtx: LHSExpr}).addAs(&an.Left, an)
	head := an.Left.Head
	if !ValidLHSVariable(head, true) {
		ps.errorp(head, errShouldBeVariableName)
	}

	if !parseSep(an, ps, '=') {
		ps.error(errShouldBeEqual)
	}
	ps.parse(&Compound{}).addAs(&an.Right, an)
}

func ValidLHSVariable(p *Primary, allowSigil bool) bool {
	switch p.Type {
	case Braced:
		// TODO(xiaq): check further inside braced expression
		return true
	case SingleQuoted, DoubleQuoted:
		// Quoted variable names may contain anything
		return true
	case Bareword:
		// Bareword variable names may only contain runes that are valid in raw
		// variable names
		return validBarewordVariableName(p.Value, allowSigil)
	default:
		return false
	}
}

func validBarewordVariableName(name string, allowSigil bool) bool {
	if name == "" {
		return false
	}
	if allowSigil && name[0] == '@' {
		name = name[1:]
	}
	for _, r := range name {
		if !allowedInVariableName(r) {
			return false
		}
	}
	return true
}

// Redir = { Compound } { '<'|'>'|'<>'|'>>' } { Space } ( '&'? Compound )
type Redir struct {
	node
	Left      *Compound
	Mode      RedirMode
	RightIsFd bool
	Right     *Compound
}

func (rn *Redir) parse(ps *parser) {
	// The parsing of the Left part is done in Form.parse.
	if rn.Left != nil {
		addChild(rn, rn.Left)
		rn.From = rn.Left.From
	}

	begin := ps.pos
	for isRedirSign(ps.peek()) {
		ps.next()
	}
	sign := ps.src[begin:ps.pos]
	switch sign {
	case "<":
		rn.Mode = Read
	case ">":
		rn.Mode = Write
	case ">>":
		rn.Mode = Append
	case "<>":
		rn.Mode = ReadWrite
	default:
		ps.error(errBadRedirSign)
	}
	addSep(rn, ps)
	parseSpaces(rn, ps)
	if parseSep(rn, ps, '&') {
		rn.RightIsFd = true
	}
	ps.parse(&Compound{}).addAs(&rn.Right, rn)
	if len(rn.Right.Indexings) == 0 {
		if rn.RightIsFd {
			ps.error(errShouldBeFD)
		} else {
			ps.error(errShouldBeFilename)
		}
		return
	}
}

func isRedirSign(r rune) bool {
	return r == '<' || r == '>'
}

// RedirMode records the mode of an IO redirection.
type RedirMode int

// Possible values for RedirMode.
const (
	BadRedirMode RedirMode = iota
	Read
	Write
	ReadWrite
	Append
)

// Filter is the Elvish filter DSL. It uses the same syntax as arguments and
// options to a command.
type Filter struct {
	node
	Args []*Compound
	Opts []*MapPair
}

func (qn *Filter) parse(ps *parser) {
	parseSpaces(qn, ps)
	for {
		r := ps.peek()
		switch {
		case r == '&':
			ps.parse(&MapPair{}).addTo(&qn.Opts, qn)
		case startsCompound(r, NormalExpr):
			ps.parse(&Compound{}).addTo(&qn.Args, qn)
		default:
			return
		}
		parseSpaces(qn, ps)
	}
}

// Compound = { Indexing }
type Compound struct {
	node
	ExprCtx   ExprCtx
	Indexings []*Indexing
}

// ExprCtx represents special contexts of expression parsing.
type ExprCtx int

const (
	// NormalExpr represents a normal expression, namely none of the special
	// ones below. It is the default value.
	NormalExpr ExprCtx = iota
	// CmdExpr represents an expression used as the command in a form. In this
	// context, unquoted <>*^ are treated as bareword characters.
	CmdExpr
	// LHSExpr represents an expression used as the left-hand-side in either
	// assignments or map pairs. In this context, an unquoted = serves as an
	// expression terminator and is thus not treated as a bareword character.
	LHSExpr
	// BracedElemExpr represents an expression used as an element in a braced
	// expression. In this context, an unquoted , serves as an expression
	// terminator and is thus not treated as a bareword character.
	BracedElemExpr
	// strictExpr is only meaningful to allowedInBareword.
	strictExpr
)

func (cn *Compound) parse(ps *parser) {
	cn.tilde(ps)
	for startsIndexing(ps.peek(), cn.ExprCtx) {
		ps.parse(&Indexing{ExprCtx: cn.ExprCtx}).addTo(&cn.Indexings, cn)
	}
}

// tilde parses a tilde if there is one. It is implemented here instead of
// within Primary since a tilde can only appear as the first part of a
// Compound. Elsewhere tildes are barewords.
func (cn *Compound) tilde(ps *parser) {
	if ps.peek() == '~' {
		ps.next()
		base := node{Ranging: diag.Ranging{From: ps.pos - 1, To: ps.pos},
			sourceText: "~", parent: nil, children: nil}
		pn := &Primary{node: base, Type: Tilde, Value: "~"}
		in := &Indexing{node: base}
		parsed{pn}.addAs(&in.Head, in)
		parsed{in}.addTo(&cn.Indexings, cn)
	}
}

func startsCompound(r rune, ctx ExprCtx) bool {
	return startsIndexing(r, ctx)
}

// Indexing = Primary { '[' Array ']' }
type Indexing struct {
	node
	ExprCtx ExprCtx
	Head    *Primary
	Indices []*Array
}

func (in *Indexing) parse(ps *parser) {
	ps.parse(&Primary{ExprCtx: in.ExprCtx}).addAs(&in.Head, in)
	for parseSep(in, ps, '[') {
		if !startsArray(ps.peek()) {
			ps.error(errShouldBeArray)
		}

		ps.parse(&Array{}).addTo(&in.Indices, in)

		if !parseSep(in, ps, ']') {
			ps.error(errShouldBeRBracket)
			return
		}
	}
}

func startsIndexing(r rune, ctx ExprCtx) bool {
	return startsPrimary(r, ctx)
}

// Array = { Space | '\n' } { Compound { Space | '\n' } }
type Array struct {
	node
	Compounds []*Compound
	// When non-empty, records the occurrences of semicolons by the indices of
	// the compounds they appear before. For instance, [; ; a b; c d;] results
	// in Semicolons={0 0 2 4}.
	Semicolons []int
}

func (sn *Array) parse(ps *parser) {
	parseSep := func() { parseSpacesAndNewlines(sn, ps) }

	parseSep()
	for startsCompound(ps.peek(), NormalExpr) {
		ps.parse(&Compound{}).addTo(&sn.Compounds, sn)
		parseSep()
	}
}

func startsArray(r rune) bool {
	return IsWhitespace(r) || startsIndexing(r, NormalExpr)
}

// Primary is the smallest expression unit.
type Primary struct {
	node
	ExprCtx ExprCtx
	Type    PrimaryType
	// The unquoted string value. Valid for Bareword, SingleQuoted,
	// DoubleQuoted, Variable, Wildcard and Tilde.
	Value    string
	Elements []*Compound // Valid for List and Lambda
	Chunk    *Chunk      // Valid for OutputCapture, ExitusCapture and Lambda
	MapPairs []*MapPair  // Valid for Map and Lambda
	Braced   []*Compound // Valid for Braced
}

// PrimaryType is the type of a Primary.
type PrimaryType int

// Possible values for PrimaryType.
const (
	BadPrimary PrimaryType = iota
	Bareword
	SingleQuoted
	DoubleQuoted
	Variable
	Wildcard
	Tilde
	ExceptionCapture
	OutputCapture
	List
	Lambda
	Map
	Braced
)

func (pn *Primary) parse(ps *parser) {
	r := ps.peek()
	if !startsPrimary(r, pn.ExprCtx) {
		ps.error(errShouldBePrimary)
		return
	}

	// Try bareword early, since it has precedence over wildcard on *
	// when ctx = commandExpr.
	if allowedInBareword(r, pn.ExprCtx) {
		pn.bareword(ps)
		return
	}

	switch r {
	case '\'':
		pn.singleQuoted(ps)
	case '"':
		pn.doubleQuoted(ps)
	case '$':
		pn.variable(ps)
	case '*':
		pn.starWildcard(ps)
	case '?':
		if ps.hasPrefix("?(") {
			pn.exitusCapture(ps)
		} else {
			pn.questionWildcard(ps)
		}
	case '(':
		pn.outputCapture(ps)
	case '[':
		pn.lbracket(ps)
	case '{':
		pn.lbrace(ps)
	default:
		// Parse an empty bareword.
		pn.Type = Bareword
	}
}

func (pn *Primary) singleQuoted(ps *parser) {
	pn.Type = SingleQuoted
	ps.next()
	pn.singleQuotedInner(ps)
}

// Parses a single-quoted string after the opening quote. Sets pn.Value but not
// pn.Type.
func (pn *Primary) singleQuotedInner(ps *parser) {
	var buf bytes.Buffer
	defer func() { pn.Value = buf.String() }()
	for {
		switch r := ps.next(); r {
		case eof:
			ps.error(errStringUnterminated)
			return
		case '\'':
			if ps.peek() == '\'' {
				// Two consecutive single quotes
				ps.next()
				buf.WriteByte('\'')
			} else {
				// End of string
				return
			}
		default:
			buf.WriteRune(r)
		}
	}
}

func (pn *Primary) doubleQuoted(ps *parser) {
	pn.Type = DoubleQuoted
	ps.next()
	pn.doubleQuotedInner(ps)
}

// Parses a double-quoted string after the opening quote. Sets pn.Value but not
// pn.Type.
func (pn *Primary) doubleQuotedInner(ps *parser) {
	var buf bytes.Buffer
	defer func() { pn.Value = buf.String() }()
	for {
		switch r := ps.next(); r {
		case eof:
			ps.error(errStringUnterminated)
			return
		case '"':
			return
		case '\\':
			switch r := ps.next(); r {
			case 'c', '^': // control sequence
				r := ps.next()
				if r < 0x3F || r > 0x5F {
					ps.backup()
					ps.error(errInvalidEscapeControl)
					ps.next()
				}
				if byte(r) == '?' { // special-case: \c? => del
					buf.WriteByte(byte(0x7F))
				} else {
					buf.WriteByte(byte(r - 0x40))
				}
			case 'x', 'u', 'U': // two, four, or eight hex digits
				var n int
				switch r {
				case 'x':
					n = 2
				case 'u':
					n = 4
				case 'U':
					n = 8
				}
				var rr rune
				for i := 0; i < n; i++ {
					d, ok := hexToDigit(ps.next())
					if !ok {
						ps.backup()
						ps.error(errInvalidEscapeHex)
						break
					}
					rr = rr*16 + d
				}
				if r == 'x' {
					buf.WriteByte(byte(rr))
				} else {
					buf.WriteRune(rr)
				}
			case '0', '1', '2', '3', '4', '5', '6', '7': // three octal digits
				rr := r - '0'
				for i := 0; i < 2; i++ {
					r := ps.next()
					if r < '0' || r > '7' {
						ps.backup()
						ps.error(errInvalidEscapeOct)
						break
					}
					rr = rr*8 + (r - '0')
				}
				if rr <= math.MaxUint8 {
					buf.WriteByte(byte(rr))
				} else {
					r := diag.Ranging{From: ps.pos - 4, To: ps.pos}
					ps.errorp(r, errInvalidEscapeOctOverflow)
				}
			default:
				if rr, ok := doubleEscape[r]; ok {
					buf.WriteRune(rr)
				} else {
					ps.backup()
					ps.error(errInvalidEscape)
					ps.next()
				}
			}
		default:
			buf.WriteRune(r)
		}
	}
}

// a table for the simple double-quote escape sequences.
var doubleEscape = map[rune]rune{
	// same as golang
	'a': '\a', 'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r',
	't': '\t', 'v': '\v', '\\': '\\', '"': '"',
	// additional
	'e': '\033',
}

var doubleUnescape = map[rune]rune{}

func init() {
	for k, v := range doubleEscape {
		doubleUnescape[v] = k
	}
}

func hexToDigit(r rune) (rune, bool) {
	switch {
	case '0' <= r && r <= '9':
		return r - '0', true
	case 'a' <= r && r <= 'f':
		return r - 'a' + 10, true
	case 'A' <= r && r <= 'F':
		return r - 'A' + 10, true
	default:
		return -1, false
	}
}

func (pn *Primary) variable(ps *parser) {
	pn.Type = Variable
	ps.next()
	switch r := ps.next(); r {
	case eof:
		ps.backup()
		ps.error(errShouldBeVariableName)
		ps.next()
	case '\'':
		pn.singleQuotedInner(ps)
	case '"':
		pn.doubleQuotedInner(ps)
	default:
		defer func() { pn.Value = ps.src[pn.From+1 : ps.pos] }()
		if !allowedInVariableName(r) && r != '@' {
			ps.backup()
			ps.error(errShouldBeVariableName)
		}
		for allowedInVariableName(ps.peek()) {
			ps.next()
		}
	}
}

// The following are allowed in variable names:
// * Anything beyond ASCII that is printable
// * Letters and numbers
// * The symbols "-_:~"
func allowedInVariableName(r rune) bool {
	return (r >= 0x80 && unicode.IsPrint(r)) ||
		('0' <= r && r <= '9') ||
		('a' <= r && r <= 'z') ||
		('A' <= r && r <= 'Z') ||
		r == '-' || r == '_' || r == ':' || r == '~'
}

func (pn *Primary) starWildcard(ps *parser) {
	pn.Type = Wildcard
	for ps.peek() == '*' {
		ps.next()
	}
	pn.Value = ps.src[pn.From:ps.pos]
}

func (pn *Primary) questionWildcard(ps *parser) {
	pn.Type = Wildcard
	if ps.peek() == '?' {
		ps.next()
	}
	pn.Value = ps.src[pn.From:ps.pos]
}

func (pn *Primary) exitusCapture(ps *parser) {
	ps.next()
	ps.next()
	addSep(pn, ps)

	pn.Type = ExceptionCapture

	ps.parse(&Chunk{}).addAs(&pn.Chunk, pn)

	if !parseSep(pn, ps, ')') {
		ps.error(errShouldBeRParen)
	}
}

func (pn *Primary) outputCapture(ps *parser) {
	pn.Type = OutputCapture
	parseSep(pn, ps, '(')

	ps.parse(&Chunk{}).addAs(&pn.Chunk, pn)

	if !parseSep(pn, ps, ')') {
		ps.error(errShouldBeRParen)
	}
}

// List   = '[' { Space } { Compound } ']'
//        = '[' { Space } { MapPair { Space } } ']'
// Map    = '[' { Space } '&' { Space } ']'
// Lambda = '[' { Space } { (Compound | MapPair) { Space } } ']' '{' Chunk '}'

func (pn *Primary) lbracket(ps *parser) {
	parseSep(pn, ps, '[')
	parseSpacesAndNewlines(pn, ps)

	loneAmpersand := false
items:
	for {
		r := ps.peek()
		switch {
		case r == '&':
			ps.next()
			hasMapPair := startsCompound(ps.peek(), LHSExpr)
			if !hasMapPair {
				loneAmpersand = true
				addSep(pn, ps)
				parseSpacesAndNewlines(pn, ps)
				break items
			}
			ps.backup()
			ps.parse(&MapPair{}).addTo(&pn.MapPairs, pn)
		case startsCompound(r, NormalExpr):
			ps.parse(&Compound{}).addTo(&pn.Elements, pn)
		default:
			break items
		}
		parseSpacesAndNewlines(pn, ps)
	}

	if !parseSep(pn, ps, ']') {
		ps.error(errShouldBeRBracket)
	}
	if loneAmpersand || len(pn.MapPairs) > 0 {
		if len(pn.Elements) > 0 {
			// TODO(xiaq): Add correct position information.
			ps.error(errBothElementsAndPairs)
		}
		pn.Type = Map
	} else {
		pn.Type = List
	}
}

// lambda parses a lambda expression. The opening brace has been seen.
func (pn *Primary) lambda(ps *parser) {
	pn.Type = Lambda
	parseSpacesAndNewlines(pn, ps)
	if parseSep(pn, ps, '|') {
		parseSpacesAndNewlines(pn, ps)
	items:
		for {
			r := ps.peek()
			switch {
			case r == '&':
				ps.parse(&MapPair{}).addTo(&pn.MapPairs, pn)
			case startsCompound(r, NormalExpr):
				ps.parse(&Compound{}).addTo(&pn.Elements, pn)
			default:
				break items
			}
			parseSpacesAndNewlines(pn, ps)
		}
		if !parseSep(pn, ps, '|') {
			ps.error(errShouldBePipe)
		}
	}
	ps.parse(&Chunk{}).addAs(&pn.Chunk, pn)
	if !parseSep(pn, ps, '}') {
		ps.error(errShouldBeRBrace)
	}
}

// Braced = '{' Compound { BracedSep Compounds } '}'
// BracedSep = { Space | '\n' } [ ',' ] { Space | '\n' }
func (pn *Primary) lbrace(ps *parser) {
	parseSep(pn, ps, '{')

	if r := ps.peek(); r == ';' || r == '\r' || r == '\n' || r == '|' || IsInlineWhitespace(r) {
		pn.lambda(ps)
		return
	}

	pn.Type = Braced

	// TODO(xiaq): The compound can be empty, which allows us to parse {,foo}.
	// Allowing compounds to be empty can be fragile in other cases.
	ps.parse(&Compound{ExprCtx: BracedElemExpr}).addTo(&pn.Braced, pn)

	for isBracedSep(ps.peek()) {
		parseSpacesAndNewlines(pn, ps)
		// optional, so ignore the return value
		parseSep(pn, ps, ',')
		parseSpacesAndNewlines(pn, ps)

		ps.parse(&Compound{ExprCtx: BracedElemExpr}).addTo(&pn.Braced, pn)
	}
	if !parseSep(pn, ps, '}') {
		ps.error(errShouldBeBraceSepOrRBracket)
	}
}

func isBracedSep(r rune) bool {
	return r == ',' || IsWhitespace(r)
}

func (pn *Primary) bareword(ps *parser) {
	pn.Type = Bareword
	defer func() { pn.Value = ps.src[pn.From:ps.pos] }()
	for allowedInBareword(ps.peek(), pn.ExprCtx) {
		ps.next()
	}
}

// allowedInBareword returns where a rune is allowed in barewords in the given
// expression context. The special strictExpr context queries whether the rune
// is allowed in all contexts.
//
// The following are allowed in barewords:
//
// * Anything allowed in variable names
// * The symbols "./\@%+!"
// * The symbol "=", if ctx != lhsExpr && ctx != strictExpr
// * The symbol ",", if ctx != bracedExpr && ctx != strictExpr
// * The symbols "<>*^", if ctx = commandExpr
//
// The seemingly weird inclusion of \ is for easier path manipulation in
// Windows.
func allowedInBareword(r rune, ctx ExprCtx) bool {
	return allowedInVariableName(r) || r == '.' || r == '/' ||
		r == '\\' || r == '@' || r == '%' || r == '+' || r == '!' ||
		(ctx != LHSExpr && ctx != strictExpr && r == '=') ||
		(ctx != BracedElemExpr && ctx != strictExpr && r == ',') ||
		(ctx == CmdExpr && (r == '<' || r == '>' || r == '*' || r == '^'))
}

func startsPrimary(r rune, ctx ExprCtx) bool {
	return r == '\'' || r == '"' || r == '$' || allowedInBareword(r, ctx) ||
		r == '?' || r == '*' || r == '(' || r == '[' || r == '{'
}

// MapPair = '&' { Space } Compound { Space } Compound
type MapPair struct {
	node
	Key, Value *Compound
}

func (mpn *MapPair) parse(ps *parser) {
	parseSep(mpn, ps, '&')

	ps.parse(&Compound{ExprCtx: LHSExpr}).addAs(&mpn.Key, mpn)
	if len(mpn.Key.Indexings) == 0 {
		ps.error(errShouldBeCompound)
	}

	if parseSep(mpn, ps, '=') {
		parseSpacesAndNewlines(mpn, ps)
		// Parse value part. It can be empty.
		ps.parse(&Compound{}).addAs(&mpn.Value, mpn)
	}
}

// Sep is the catch-all node type for leaf nodes that lack internal structures
// and semantics, and serve solely for syntactic purposes. The parsing of
// separators depend on the Parent node; as such it lacks a genuine parse
// method.
type Sep struct {
	node
}

// NewSep makes a new Sep.
func NewSep(src string, begin, end int) *Sep {
	return &Sep{node: node{diag.Ranging{From: begin, To: end}, src[begin:end], nil, nil}}
}

func (*Sep) parse(*parser) {
	// A no-op, only to satisfy the Node interface.
}

func addSep(n Node, ps *parser) {
	var begin int
	ch := Children(n)
	if len(ch) > 0 {
		begin = ch[len(ch)-1].Range().To
	} else {
		begin = n.Range().From
	}
	if begin < ps.pos {
		addChild(n, NewSep(ps.src, begin, ps.pos))
	}
}

func parseSep(n Node, ps *parser, sep rune) bool {
	if ps.peek() == sep {
		ps.next()
		addSep(n, ps)
		return true
	}
	return false
}

func parseSpaces(n Node, ps *parser) {
	parseSpacesInner(n, ps, false)
}

func parseSpacesAndNewlines(n Node, ps *parser) {
	parseSpacesInner(n, ps, true)
}

func parseSpacesInner(n Node, ps *parser, newlines bool) {
spaces:
	for {
		r := ps.peek()
		switch {
		case IsInlineWhitespace(r):
			ps.next()
		case newlines && IsWhitespace(r):
			ps.next()
		case r == '#':
			// Comment is like inline whitespace as long as we don't include the
			// trailing newline.
			ps.next()
			for {
				r := ps.peek()
				if r == eof || r == '\r' || r == '\n' {
					break
				}
				ps.next()
			}
		case r == '^':
			// Line continuation is like inline whitespace.
			ps.next()
			switch ps.peek() {
			case '\r':
				ps.next()
				if ps.peek() == '\n' {
					ps.next()
				}
			case '\n':
				ps.next()
			case eof:
				ps.error(errShouldBeNewline)
			default:
				ps.backup()
				break spaces
			}
		default:
			break spaces
		}
	}
	addSep(n, ps)
}

// IsInlineWhitespace reports whether r is an inline whitespace character.
// Currently this includes space (Unicode 0x20) and tab (Unicode 0x9).
func IsInlineWhitespace(r rune) bool {
	return r == ' ' || r == '\t'
}

// IsWhitespace reports whether r is a whitespace. Currently this includes
// inline whitespace characters and newline (Unicode 0xa).
func IsWhitespace(r rune) bool {
	return IsInlineWhitespace(r) || r == '\r' || r == '\n'
}

func addChild(p Node, ch Node) {
	p.n().addChild(ch)
	ch.n().parent = p
}
