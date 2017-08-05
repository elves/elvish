// Package parse implements the elvish parser.
package parse

//go:generate ./boilerplate.py
//go:generate stringer -type=PrimaryType,RedirMode -output=string.go

import (
	"bytes"
	"errors"
	"fmt"
	"unicode"
)

// Parse parses Elvish source. If the error is not nil, it always has type
// ParseError.
func Parse(srcname, src string) (*Chunk, error) {
	ps := NewParser(srcname, src)
	n := ParseChunk(ps)
	ps.Done()
	return n, ps.Errors()
}

// Errors.
var (
	errUnexpectedRune       = errors.New("unexpected rune")
	errShouldBeForm         = newError("", "form")
	errBadLHS               = errors.New("bad assignment LHS")
	errDuplicateExitusRedir = newError("duplicate exitus redir")
	errBadRedirSign         = newError("bad redir sign", "'<'", "'>'", "'>>'", "'<>'")
	errShouldBeFD           = newError("", "a composite term representing fd")
	errShouldBeFilename     = newError("", "a composite term representing filename")
	errShouldBeArray        = newError("", "spaced")
	errStringUnterminated   = newError("string not terminated")
	errChainedAssignment    = newError("chained assignment not yet supported")
	errInvalidEscape        = newError("invalid escape sequence")
	errInvalidEscapeOct     = newError("invalid escape sequence", "octal digit")
	errInvalidEscapeHex     = newError("invalid escape sequence", "hex digit")
	errInvalidEscapeControl = newError("invalid control sequence", "a rune between @ (0x40) and _(0x5F)")
	errShouldBePrimary      = newError("",
		"single-quoted string", "double-quoted string", "bareword")
	errShouldBeVariableName       = newError("", "variable name")
	errShouldBeRBracket           = newError("", "']'")
	errShouldBeRBrace             = newError("", "'}'")
	errShouldBeBraceSepOrRBracket = newError("", "','", "'}'")
	errShouldBeRParen             = newError("", "')'")
	errShouldBeBackquoteOrLParen  = newError("", "'`'", "'('")
	errShouldBeBackquote          = newError("", "'`'")
	errShouldBeCompound           = newError("", "compound")
	errShouldBeEqual              = newError("", "'='")
	errBothElementsAndPairs       = newError("cannot contain both list elements and map pairs")
)

// Chunk = { PipelineSep | Space } { Pipeline { PipelineSep | Space } }
type Chunk struct {
	node
	Pipelines []*Pipeline
}

func (bn *Chunk) parse(ps *Parser) {
	bn.parseSeps(ps)
	for startsPipeline(ps.peek()) {
		bn.addToPipelines(ParsePipeline(ps))
		if bn.parseSeps(ps) == 0 {
			break
		}
	}
}

func isPipelineSep(r rune) bool {
	return r == '\n' || r == ';'
}

// parseSeps parses pipeline separators along with whitespaces. It returns the
// number of pipeline separators parsed.
func (bn *Chunk) parseSeps(ps *Parser) int {
	nseps := 0
	for {
		r := ps.peek()
		if isPipelineSep(r) {
			// parse as a Sep
			parseSep(bn, ps, r)
			nseps++
		} else if IsSpace(r) {
			// parse a run of spaces as a Sep
			parseSpaces(bn, ps)
		} else if r == '#' {
			// parse a comment as a Sep
			for {
				r := ps.peek()
				if r == eof || r == '\n' {
					break
				}
				ps.next()
			}
			addSep(bn, ps)
			nseps++
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

func (pn *Pipeline) parse(ps *Parser) {
	pn.addToForms(ParseForm(ps))
	for parseSep(pn, ps, '|') {
		parseSpacesAndNewlines(pn, ps)
		if !startsForm(ps.peek()) {
			ps.error(errShouldBeForm)
			return
		}
		pn.addToForms(ParseForm(ps))
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
//        { Compound } { Space } { ( Compound | MapPair | Redir | ExitusRedir ) { Space } }
type Form struct {
	node
	Assignments []*Assignment
	Head        *Compound
	// Left-hand-sides for the spacey assignment. Right-hand-sides are in Args.
	Vars        []*Compound
	Args        []*Compound
	Opts        []*MapPair
	Redirs      []*Redir
	ExitusRedir *ExitusRedir
}

func (fn *Form) parse(ps *Parser) {
	parseSpaces(fn, ps)
	for fn.tryAssignment(ps) {
		parseSpaces(fn, ps)
	}

	// Parse head.
	if !startsCompound(ps.peek(), true) {
		if len(fn.Assignments) > 0 {
			// Assignment-only form.
			return
		}
		// Bad form.
		ps.error(fmt.Errorf("bad rune at form head: %q", ps.peek()))
	}
	fn.setHead(ParseCompound(ps, true))
	parseSpaces(fn, ps)

	for {
		r := ps.peek()
		switch {
		case r == '&':
			ps.next()
			hasMapPair := startsCompound(ps.peek(), false)
			ps.backup()
			if !hasMapPair {
				// background indicator
				return
			}
			fn.addToOpts(ParseMapPair(ps))
		case startsCompound(r, false):
			if ps.hasPrefix("?>") {
				if fn.ExitusRedir != nil {
					ps.error(errDuplicateExitusRedir)
					// Parse the duplicate redir anyway.
					addChild(fn, ParseExitusRedir(ps))
				} else {
					fn.setExitusRedir(ParseExitusRedir(ps))
				}
				continue
			}
			cn := ParseCompound(ps, false)
			if isRedirSign(ps.peek()) {
				// Redir
				fn.addToRedirs(ParseRedir(ps, cn))
			} else if cn.sourceText == "=" {
				// Spacey assignment.
				// Turn the equal sign into a Sep.
				addChild(fn, NewSep(ps.src, cn.begin, cn.end))
				// Turn the head and preceding arguments into LHSs.
				addLHS := func(cn *Compound) {
					if len(cn.Indexings) == 1 && checkVariableInAssignment(cn.Indexings[0].Head, ps) {
						fn.Vars = append(fn.Vars, cn)
					} else {
						ps.errorp(cn.begin, cn.end, errBadLHS)
					}
				}
				if fn.Head != nil {
					addLHS(fn.Head)
				} else {
					ps.error(errChainedAssignment)
				}
				fn.Head = nil
				for _, cn := range fn.Args {
					addLHS(cn)
				}
				fn.Args = nil
			} else {
				fn.addToArgs(cn)
			}
		case isRedirSign(r):
			fn.addToRedirs(ParseRedir(ps, nil))
		default:
			return
		}
		parseSpaces(fn, ps)
	}
}

// tryAssignment tries to parse an assignment. If succeeded, it adds the parsed
// assignment to fn.Assignments and returns true. Otherwise it rewinds the
// parser and returns false.
func (fn *Form) tryAssignment(ps *Parser) bool {
	if !startsIndexing(ps.peek(), false) || ps.peek() == '=' {
		return false
	}

	pos := ps.pos
	errorEntries := ps.errors.Entries
	an := ParseAssignment(ps)
	// If errors were added, revert
	if len(ps.errors.Entries) > len(errorEntries) {
		ps.errors.Entries = errorEntries
		ps.pos = pos
		return false
	}
	fn.addToAssignments(an)
	return true
}

func startsForm(r rune) bool {
	return IsSpace(r) || startsCompound(r, true)
}

// Assignment = Indexing '=' Compound
type Assignment struct {
	node
	Left  *Indexing
	Right *Compound
}

func (an *Assignment) parse(ps *Parser) {
	ps.cut('=')
	an.setLeft(ParseIndexing(ps, false))
	head := an.Left.Head
	if !checkVariableInAssignment(head, ps) {
		ps.errorp(head.Begin(), head.End(), errShouldBeVariableName)
	}
	ps.uncut('=')

	if !parseSep(an, ps, '=') {
		ps.error(errShouldBeEqual)
	}
	an.setRight(ParseCompound(ps, false))
}

func checkVariableInAssignment(p *Primary, ps *Parser) bool {
	if p.Type == Braced {
		// XXX don't check further inside braced expression
		return true
	}
	if p.Type != Bareword && p.Type != SingleQuoted && p.Type != DoubleQuoted {
		return false
	}
	if p.Value == "" {
		return false
	}
	for _, r := range p.Value {
		// XXX special case '&' and '@'.
		if !allowedInVariableName(r) && r != '&' && r != '@' {
			return false
		}
	}
	return true
}

// ExitusRedir = '?' '>' { Space } Compound
type ExitusRedir struct {
	node
	Dest *Compound
}

func (ern *ExitusRedir) parse(ps *Parser) {
	ps.next()
	ps.next()
	addSep(ern, ps)
	parseSpaces(ern, ps)
	ern.setDest(ParseCompound(ps, false))
}

// Redir = { Compound } { '<'|'>'|'<>'|'>>' } { Space } ( '&'? Compound )
type Redir struct {
	node
	Left      *Compound
	Mode      RedirMode
	RightIsFd bool
	Right     *Compound
}

func (rn *Redir) parse(ps *Parser, dest *Compound) {
	// The parsing of the Left part is done in Form.parse.
	if dest != nil {
		rn.setLeft(dest)
		rn.begin = dest.begin
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
	rn.setRight(ParseCompound(ps, false))
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

// Compound = { Indexing }
type Compound struct {
	node
	Indexings []*Indexing
}

func (cn *Compound) parse(ps *Parser, head bool) {
	cn.tilde(ps)
	for startsIndexing(ps.peek(), head) {
		cn.addToIndexings(ParseIndexing(ps, head))
	}
}

// tilde parses a tilde if there is one. It is implemented here instead of
// within Primary since a tilde can only appear as the first part of a
// Compound. Elsewhere tildes are barewords.
func (cn *Compound) tilde(ps *Parser) {
	if ps.peek() == '~' {
		ps.next()
		base := node{nil, ps.pos - 1, ps.pos, "~", nil}
		pn := &Primary{node: base, Type: Tilde, Value: "~"}
		in := &Indexing{node: base}
		in.setHead(pn)
		cn.addToIndexings(in)
	}
}

func startsCompound(r rune, head bool) bool {
	return startsIndexing(r, head)
}

// Indexing = Primary { '[' Array ']' }
type Indexing struct {
	node
	Head     *Primary
	Indicies []*Array
}

func (in *Indexing) parse(ps *Parser, head bool) {
	in.setHead(ParsePrimary(ps, head))
	for parseSep(in, ps, '[') {
		if !startsArray(ps.peek()) {
			ps.error(errShouldBeArray)
		}

		ps.pushCutset()
		in.addToIndicies(ParseArray(ps, false))
		ps.popCutset()

		if !parseSep(in, ps, ']') {
			ps.error(errShouldBeRBracket)
			return
		}
	}
}

func startsIndexing(r rune, head bool) bool {
	return startsPrimary(r, head)
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

func (sn *Array) parse(ps *Parser, allowSemicolon bool) {
	parseSep := func() {
		parseSpacesAndNewlines(sn, ps)
		if allowSemicolon {
			for parseSep(sn, ps, ';') {
				sn.Semicolons = append(sn.Semicolons, len(sn.Compounds))
			}
			parseSpacesAndNewlines(sn, ps)
		}
	}

	parseSep()
	for startsCompound(ps.peek(), false) {
		sn.addToCompounds(ParseCompound(ps, false))
		parseSep()
	}
}

func IsSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func startsArray(r rune) bool {
	return IsSpaceOrNewline(r) || startsIndexing(r, false)
}

// Primary is the smallest expression unit.
type Primary struct {
	node
	Type PrimaryType
	// The unquoted string value. Valid for Bareword, SingleQuoted,
	// DoubleQuoted, Variable, Wildcard and Tilde.
	Value    string
	Elements []*Compound // Valid for List and Labda
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

func (pn *Primary) parse(ps *Parser, head bool) {
	r := ps.peek()
	if !startsPrimary(r, head) {
		ps.error(errShouldBePrimary)
		return
	}

	// Try bareword early, since it has precedence over wildcard on *
	// when head is true.
	if allowedInBareword(r, head) {
		pn.bareword(ps, head)
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
		pn.wildcard(ps)
	case '?':
		if ps.hasPrefix("?(") {
			pn.exitusCapture(ps)
		} else {
			pn.wildcard(ps)
		}
	case '(', '`':
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

func (pn *Primary) singleQuoted(ps *Parser) {
	pn.Type = SingleQuoted
	ps.next()
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

func (pn *Primary) doubleQuoted(ps *Parser) {
	pn.Type = DoubleQuoted
	ps.next()
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
			case 'c', '^':
				// Control sequence
				r := ps.next()
				if r < 0x40 || r >= 0x60 {
					ps.backup()
					ps.error(errInvalidEscapeControl)
					ps.next()
				}
				buf.WriteByte(byte(r - 0x40))
			case 'x', 'u', 'U':
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
				buf.WriteRune(rr)
			case '0', '1', '2', '3', '4', '5', '6', '7':
				// 2 more octal digits
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
				buf.WriteRune(rr)
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

func (pn *Primary) variable(ps *Parser) {
	pn.Type = Variable
	defer func() { pn.Value = ps.src[pn.begin+1 : ps.pos] }()
	ps.next()
	// The character of the variable name can be anything.
	if ps.next() == eof {
		ps.backup()
		ps.error(errShouldBeVariableName)
		ps.next()
	}
	for allowedInVariableName(ps.peek()) {
		ps.next()
	}
}

// The following are allowed in variable names:
// * Anything beyond ASCII that is printable
// * Letters and numbers
// * The symbols "-_:&"
func allowedInVariableName(r rune) bool {
	return (r >= 0x80 && unicode.IsPrint(r)) ||
		('0' <= r && r <= '9') ||
		('a' <= r && r <= 'z') ||
		('A' <= r && r <= 'Z') ||
		r == '-' || r == '_' || r == ':' || r == '&'
}

func (pn *Primary) wildcard(ps *Parser) {
	pn.Type = Wildcard
	for isWildcard(ps.peek()) {
		ps.next()
	}
	pn.Value = ps.src[pn.begin:ps.pos]
}

func isWildcard(r rune) bool {
	return r == '*' || r == '?'
}

func (pn *Primary) exitusCapture(ps *Parser) {
	ps.next()
	ps.next()
	addSep(pn, ps)

	pn.Type = ExceptionCapture

	ps.pushCutset()
	pn.setChunk(ParseChunk(ps))
	ps.popCutset()

	if !parseSep(pn, ps, ')') {
		ps.error(errShouldBeRParen)
	}
}

func (pn *Primary) outputCapture(ps *Parser) {
	pn.Type = OutputCapture

	var closer rune
	var shouldBeCloser error

	switch ps.next() {
	case '(':
		closer = ')'
		shouldBeCloser = errShouldBeRParen
	case '`':
		closer = '`'
		shouldBeCloser = errShouldBeBackquote
	default:
		ps.backup()
		ps.error(errShouldBeBackquoteOrLParen)
		ps.next()
		return
	}
	addSep(pn, ps)

	if closer == '`' {
		ps.pushCutset(closer)
	} else {
		ps.pushCutset()
	}
	pn.setChunk(ParseChunk(ps))
	ps.popCutset()

	if !parseSep(pn, ps, closer) {
		ps.error(shouldBeCloser)
	}
}

// List   = '[' { Space } { Compound } ']'
//        = '[' { Space } { MapPair { Space } } ']'
// Map    = '[' { Space } '&' { Space } ']'
// Lambda = '[' { Space } { (Compound | MapPair) { Space } } ']' '{' Chunk '}'

func (pn *Primary) lbracket(ps *Parser) {
	parseSep(pn, ps, '[')
	parseSpacesAndNewlines(pn, ps)

	loneAmpersand := false
	ps.pushCutset()
items:
	for {
		r := ps.peek()
		switch {
		case r == '&':
			ps.next()
			hasMapPair := startsCompound(ps.peek(), false)
			if !hasMapPair {
				loneAmpersand = true
				addSep(pn, ps)
				parseSpacesAndNewlines(pn, ps)
				break items
			}
			ps.backup()
			pn.addToMapPairs(ParseMapPair(ps))
		case startsCompound(r, false):
			pn.addToElements(ParseCompound(ps, false))
		default:
			break items
		}
		parseSpacesAndNewlines(pn, ps)
	}
	ps.popCutset()

	if !parseSep(pn, ps, ']') {
		ps.error(errShouldBeRBracket)
	}
	if parseSep(pn, ps, '{') {
		pn.lambda(ps)
	} else {
		if loneAmpersand || len(pn.MapPairs) > 0 {
			if len(pn.Elements) > 0 {
				ps.error(errBothElementsAndPairs)
			}
			pn.Type = Map
		} else {
			pn.Type = List
		}
	}
}

// lambda parses a lambda expression. The opening brace has been seen.
func (pn *Primary) lambda(ps *Parser) {
	pn.Type = Lambda
	ps.pushCutset()
	pn.setChunk(ParseChunk(ps))
	ps.popCutset()
	if !parseSep(pn, ps, '}') {
		ps.error(errShouldBeRBrace)
	}
}

// Braced = '{' Compound { BracedSep Compounds } '}'
// BracedSep = { Space | '\n' } [ ',' ] { Space | '\n' }
func (pn *Primary) lbrace(ps *Parser) {
	parseSep(pn, ps, '{')

	if r := ps.peek(); r == ';' || r == '\n' || IsSpace(r) {
		pn.lambda(ps)
		return
	}

	pn.Type = Braced

	ps.pushCutset()
	defer ps.popCutset()

	// XXX: The compound can be empty, which allows us to parse {,foo}.
	// Allowing compounds to be empty can be fragile in other cases.
	ps.cut(',')
	pn.addToBraced(ParseCompound(ps, false))
	ps.uncut(',')

	for isBracedSep(ps.peek()) {
		parseSpacesAndNewlines(pn, ps)
		// optional, so ignore the return value
		parseSep(pn, ps, ',')
		parseSpacesAndNewlines(pn, ps)

		ps.cut(',')
		pn.addToBraced(ParseCompound(ps, false))
		ps.uncut(',')
	}
	if !parseSep(pn, ps, '}') {
		ps.error(errShouldBeBraceSepOrRBracket)
	}
}

func isBracedSep(r rune) bool {
	return r == ',' || IsSpaceOrNewline(r)
}

func (pn *Primary) bareword(ps *Parser, head bool) {
	pn.Type = Bareword
	defer func() { pn.Value = ps.src[pn.begin:ps.pos] }()
	for allowedInBareword(ps.peek(), head) {
		ps.next()
	}
}

// The following are allowed in barewords:
// * Anything allowed in variable names, except &
// * The symbols "%+,./=@~!\"
// * The symbols "<>*^", if the bareword is in head
//
// The seemingly weird inclusion of \ is for easier path manipulation in
// Windows.
func allowedInBareword(r rune, head bool) bool {
	return (r != '&' && allowedInVariableName(r)) ||
		r == '%' || r == '+' || r == ',' || r == '.' ||
		r == '/' || r == '=' || r == '@' || r == '~' || r == '!' || r == '\\' ||
		(head && (r == '<' || r == '>' || r == '*' || r == '^'))
}

func startsPrimary(r rune, head bool) bool {
	return r == '\'' || r == '"' || r == '$' || allowedInBareword(r, head) ||
		r == '?' || r == '*' || r == '(' || r == '`' || r == '[' || r == '{'
}

// MapPair = '&' { Space } Compound { Space } Compound
type MapPair struct {
	node
	Key, Value *Compound
}

func (mpn *MapPair) parse(ps *Parser) {
	parseSep(mpn, ps, '&')

	// Parse key part, cutting on '='.
	ps.cut('=')
	mpn.setKey(ParseCompound(ps, false))
	if len(mpn.Key.Indexings) == 0 {
		ps.error(errShouldBeCompound)
	}
	ps.uncut('=')

	if parseSep(mpn, ps, '=') {
		parseSpacesAndNewlines(mpn, ps)
		// Parse value part.
		mpn.setValue(ParseCompound(ps, false))
		// The value part can be empty.
	}
}

// Sep is the catch-all node type for leaf nodes that lack internal structures
// and semantics, and serve solely for syntactic purposes. The parsing of
// separators depend on the Parent node; as such it lacks a genuine parse
// method.
type Sep struct {
	node
}

func NewSep(src string, begin, end int) *Sep {
	return &Sep{node{nil, begin, end, src[begin:end], nil}}
}

func addSep(n Node, ps *Parser) {
	var begin int
	ch := n.Children()
	if len(ch) > 0 {
		begin = ch[len(ch)-1].End()
	} else {
		begin = n.Begin()
	}
	addChild(n, NewSep(ps.src, begin, ps.pos))
}

func parseSep(n Node, ps *Parser, sep rune) bool {
	if ps.peek() == sep {
		ps.next()
		addSep(n, ps)
		return true
	}
	return false
}

func parseSpaces(n Node, ps *Parser) {
	if !IsSpace(ps.peek()) {
		return
	}
	ps.next()
	for IsSpace(ps.peek()) {
		ps.next()
	}
	addSep(n, ps)
}

func parseSpacesAndNewlines(n Node, ps *Parser) {
	// TODO parse comments here.
	if !IsSpaceOrNewline(ps.peek()) {
		return
	}
	ps.next()
	for IsSpaceOrNewline(ps.peek()) {
		ps.next()
	}
	addSep(n, ps)
}

func IsSpaceOrNewline(r rune) bool {
	return IsSpace(r) || r == '\n'
}

func addChild(p Node, ch Node) {
	p.n().children = append(p.n().children, ch)
	ch.n().parent = p
}
