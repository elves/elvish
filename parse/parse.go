// Package parse implements the elvish parser.
package parse

//go:generate ./boilerplate.py
//go:generate stringer -type=PrimaryType,RedirMode -output=string.go

import (
	"bytes"
	"errors"
	"unicode"
)

func Parse(src string) (*Chunk, error) {
	ps := &parser{src, 0, 0, nil}
	bn := parseChunk(ps, nil)
	if ps.pos != len(src) {
		ps.error(unexpectedRune)
	}
	var err error
	if ps.errors != nil {
		err = ps.errors
	}
	return bn, err
}

// Errors.
var (
	unexpectedRune       = errors.New("unexpected rune")
	shouldBeForm         = newError("", "form")
	duplicateExitusRedir = newError("duplicate exitus redir")
	shouldBeEqual        = newError("", "=")
	badRedirSign         = newError("bad redir sign", "'<'", "'>'", "'>>'", "'<>'")
	shouldBeFd           = newError("", "a composite term representing fd")
	shouldBeFilename     = newError("", "a composite term representing filename")
	shouldBeArray        = newError("", "spaced")
	StringUnterminated   = newError("string not terminated")
	InvalidEscape        = newError("invalid escape sequence")
	InvalidEscapeOct     = newError("invalid escape sequence", "octal digit")
	InvalidEscapeHex     = newError("invalid escape sequence", "hex digit")
	InvalidEscapeControl = newError("invalid control sequence", "a rune between @ (0x40) and _(0x5F)")
	ShouldBePrimary      = newError("",
		"single-quoted string", "double-quoted string", "bareword")
	shouldBeVariableName       = newError("", "variable name")
	shouldBeAmpersandOrArray   = newError("", "'&'", "spaced")
	shouldBeRBracket           = newError("", "']'")
	shouldBeRBrace             = newError("", "'}'")
	shouldBeAmpersand          = newError("", "'&'")
	shouldBeBraceSepOrRBracket = newError("", "'-'", "','", "'}'")
	shouldBeChunk              = newError("", "chunk")
	shouldBeRParen             = newError("", "')'")
	shouldBeBackquoteOrLParen  = newError("", "'`'", "'('")
	shouldBeBackquote          = newError("", "'`'")
	shouldBeCompound           = newError("", "compound")
)

// Chunk = { PipelineSep | Space } { Pipeline { PipelineSep | Space } }
type Chunk struct {
	node
	Pipelines []*Pipeline
}

func (bn *Chunk) parse(ps *parser, cut runePred) {
	bn.parseSeps(ps, cut)
	for startsPipeline(ps.peek(), cut) {
		bn.addToPipelines(parsePipeline(ps, cut))
		if bn.parseSeps(ps, cut) == 0 {
			break
		}
	}
}

func isPipelineSep(r rune) bool {
	return r == '\n' || r == ';'
}

// parseSeps parses pipeline separators along with whitespaces. It returns the
// number of pipeline separators parsed.
func (bn *Chunk) parseSeps(ps *parser, cut runePred) int {
	nseps := 0
	for {
		r := ps.peek()
		if cut.matches(r) {
			break
		} else if isPipelineSep(r) {
			// parse as a Sep
			parseSep(bn, ps, r)
			nseps += 1
		} else if isSpace(r) {
			// parse a run of spaces as a Sep
			parseSpaces(bn, ps)
		} else if r == '#' {
			// parse a comment as a Sep
			for {
				r := ps.peek()
				if r == EOF || r == '\n' {
					break
				}
				ps.next()
			}
			nseps += 1
		} else {
			break
		}
	}
	return nseps
}

func startsChunk(r rune, cut runePred) bool {
	return isPipelineSep(r) || startsPipeline(r, cut)
}

// Pipeline = Form { '|' Form }
type Pipeline struct {
	node
	Forms []*Form
}

func (pn *Pipeline) parse(ps *parser, cut runePred) {
	pn.addToForms(parseForm(ps, cut))
	for !cut.matches('|') && parseSep(pn, ps, '|') {
		if !startsForm(ps.peek(), cut) {
			ps.error(shouldBeForm)
			return
		}
		pn.addToForms(parseForm(ps, cut))
	}
}

func startsPipeline(r rune, cut runePred) bool {
	return startsForm(r, cut)
}

// Form = { Space } Compound { Space } { ( Compound | MapPair | Redir | ExitusRedir ) { Space } }
type Form struct {
	node
	Assignments []*Assignment
	Head        *Compound
	Args        []*Compound
	NamedArgs   []*MapPair
	Redirs      []*Redir
	ExitusRedir *ExitusRedir
}

func (fn *Form) parse(ps *parser, cut runePred) {
	parseSpaces(fn, ps)
	for fn.tryAssignment(ps, cut) {
		parseSpaces(fn, ps)
	}
	if !startsCompound(ps.peek(), cut) {
		if len(fn.Assignments) > 0 {
			return
		} else {
			ps.error(shouldBeCompound)
		}
	}
	fn.setHead(parseCompound(ps, cut))
	parseSpaces(fn, ps)
loop:
	for {
		r := ps.peek()
		switch {
		case cut.matches(r):
			break loop
		case r == '&':
			fn.addToNamedArgs(parseMapPair(ps, cut))
		case startsCompound(r, cut):
			if !cut.matches('>') && ps.hasPrefix("?>") {
				if fn.ExitusRedir != nil {
					ps.error(duplicateExitusRedir)
					// Parse the duplicate redir anyway.
					addChild(fn, parseExitusRedir(ps, cut))
				} else {
					fn.setExitusRedir(parseExitusRedir(ps, cut))
				}
				continue
			}
			cn := parseCompound(ps, cut)
			if !cut.matches(ps.peek()) && isRedirSign(ps.peek()) {
				// Redir
				fn.addToRedirs(parseRedir(ps, cut, cn))
			} else {
				fn.addToArgs(cn)
			}
		case isRedirSign(r):
			fn.addToRedirs(parseRedir(ps, cut, nil))
		default:
			return
		}
		parseSpaces(fn, ps)
	}
}

// tryAssignment tries to parse an assignment. If suceeded, it adds the parsed
// assignment to fn.Assignments and returns true. Otherwise it rewinds the
// parser and returns false.
func (fn *Form) tryAssignment(ps *parser, cut runePred) bool {
	if !startsIndexing(ps.peek(), cut) || ps.peek() == '=' {
		return false
	}

	begin := ps.pos
	var ok bool
	an := parseAssignment(ps, cut, &ok)
	if !ok {
		ps.pos = begin
		return false
	}
	fn.addToAssignments(an)
	return true
}

func startsForm(r rune, cut runePred) bool {
	return isSpace(r) || startsCompound(r, cut)
}

// Assignment = Primary '=' Compound
type Assignment struct {
	node
	Dst *Indexing
	Src *Compound
}

func (an *Assignment) parse(ps *parser, cut runePred, pok *bool) {
	cutWithEqual := runePred(func(r rune) bool {
		return cut.matches(r) || r == '='
	})
	an.setDst(parseIndexing(ps, cutWithEqual))
	if !parseSep(an, ps, '=') {
		*pok = false
		return
	}
	an.setSrc(parseCompound(ps, cut))
	*pok = true
}

// ExitusRedir = '?' '>' { Space } Compound
type ExitusRedir struct {
	node
	Dest *Compound
}

func (ern *ExitusRedir) parse(ps *parser, cut runePred) {
	ps.next()
	ps.next()
	addSep(ern, ps)
	parseSpaces(ern, ps)
	ern.setDest(parseCompound(ps, cut))
}

// Redir = { Compound } { '<'|'>'|'<>'|'>>' } { Space } ( '&'? Compound )
type Redir struct {
	node
	Dest       *Compound
	Mode       RedirMode
	SourceIsFd bool
	Source     *Compound
}

func (rn *Redir) parse(ps *parser, cut runePred, dest *Compound) {
	// The parsing of the Dest part is done in Form.parse.
	if dest != nil {
		rn.Dest = dest
		rn.begin = dest.begin
		addChild(rn, dest)
	}

	begin := ps.pos
	for !cut.matches(ps.peek()) && isRedirSign(ps.peek()) {
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
		ps.error(badRedirSign)
	}
	addSep(rn, ps)
	parseSpaces(rn, ps)
	if !cut.matches('&') && parseSep(rn, ps, '&') {
		rn.SourceIsFd = true
	}
	if !startsCompound(ps.peek(), cut) {
		if rn.SourceIsFd {
			ps.error(shouldBeFd)
		} else {
			ps.error(shouldBeFilename)
		}
		return
	}
	rn.setSource(parseCompound(ps, cut))
}

func isRedirSign(r rune) bool {
	return r == '<' || r == '>'
}

type RedirMode int

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

func (cn *Compound) parse(ps *parser, cut runePred) {
	for startsIndexing(ps.peek(), cut) {
		cn.addToIndexings(parseIndexing(ps, cut))
	}
}

func startsCompound(r rune, cut runePred) bool {
	return startsIndexing(r, cut)
}

// Indexing = Primary { '[' Array ']' }
type Indexing struct {
	node
	Head     *Primary
	Indicies []*Array
}

func (in *Indexing) parse(ps *parser, cut runePred) {
	in.setHead(parsePrimary(ps, cut))
	for parseSep(in, ps, '[') {
		if !startsArray(ps.peek()) {
			ps.error(shouldBeArray)
		}
		in.addToIndicies(parseArray(ps))
		if !parseSep(in, ps, ']') {
			ps.error(shouldBeRBracket)
			return
		}
	}
}

func startsIndexing(r rune, cut runePred) bool {
	return startsPrimary(r, cut)
}

// Array = { Space } { Compound { Space } }
type Array struct {
	node
	Compounds []*Compound
}

func (sn *Array) parse(ps *parser) {
	parseSpaces(sn, ps)
	for startsCompound(ps.peek(), nil) {
		sn.addToCompounds(parseCompound(ps, nil))
		parseSpaces(sn, ps)
	}
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func startsArray(r rune) bool {
	return isSpace(r) || startsIndexing(r, nil)
}

type Primary struct {
	node
	Type PrimaryType
	// The unquoted string value. Valid for Bareword, SingleQuoted,
	// DoubleQuoted, Variable and Wildcard.
	Value    string
	List     *Array      // Valid for List and Lambda
	Chunk    *Chunk      // Valid for OutputCapture, ExitusCapture and Lambda
	MapPairs []*MapPair  // Valid for Map
	Braced   []*Compound // Valid for Braced
	IsRange  []bool      // Valid for Braced
}

type PrimaryType int

const (
	BadPrimary PrimaryType = iota
	Bareword
	SingleQuoted
	DoubleQuoted
	Variable
	Wildcard
	ErrorCapture
	OutputCapture
	List
	Lambda
	Map
	Braced
)

func (pn *Primary) parse(ps *parser, cut runePred) {
	r := ps.peek()
	if !startsPrimary(r, cut) {
		ps.error(ShouldBePrimary)
		return
	}
	switch r {
	case '\'':
		pn.singleQuoted(ps)
	case '"':
		pn.doubleQuoted(ps)
	case '$':
		pn.variable(ps, cut)
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
		pn.bareword(ps, cut)
	}
}

func (pn *Primary) singleQuoted(ps *parser) {
	pn.Type = SingleQuoted
	ps.next()
	var buf bytes.Buffer
	defer func() { pn.Value = buf.String() }()
	for {
		switch r := ps.next(); r {
		case EOF:
			ps.error(StringUnterminated)
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
	var buf bytes.Buffer
	defer func() { pn.Value = buf.String() }()
	for {
		switch r := ps.next(); r {
		case EOF:
			ps.error(StringUnterminated)
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
					ps.error(InvalidEscapeControl)
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
						ps.error(InvalidEscapeHex)
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
						ps.error(InvalidEscapeOct)
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
					ps.error(InvalidEscape)
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

func (pn *Primary) variable(ps *parser, cut runePred) {
	pn.Type = Variable
	defer func() { pn.Value = ps.src[pn.begin:ps.pos] }()
	ps.next()
	// The character of the variable name can be anything.
	if r := ps.next(); cut.matches(r) {
		ps.backup()
		ps.error(shouldBeVariableName)
		ps.next()
	}
	for allowedInVariableName(ps.peek()) && !cut.matches(ps.peek()) {
		ps.next()
	}
}

// The following are allowed in variable names:
// * Anything beyond ASCII that is printable
// * Letters and numbers
// * The symbols "-_:"
func allowedInVariableName(r rune) bool {
	return (r >= 0x80 && unicode.IsPrint(r)) ||
		('0' <= r && r <= '9') ||
		('a' <= r && r <= 'z') ||
		('A' <= r && r <= 'Z') ||
		r == '-' || r == '_' || r == ':'
}

func (pn *Primary) wildcard(ps *parser) {
	pn.Type = Wildcard
	for isWildcard(ps.peek()) {
		ps.next()
	}
	pn.Value = ps.src[pn.begin:ps.pos]
}

func isWildcard(r rune) bool {
	return r == '*' || r == '?'
}

func (pn *Primary) exitusCapture(ps *parser) {
	ps.next()
	ps.next()
	addSep(pn, ps)

	pn.Type = ErrorCapture
	if !startsChunk(ps.peek(), nil) && ps.peek() != ')' {
		ps.error(shouldBeChunk)
		return
	}
	pn.setChunk(parseChunk(ps, nil))
	if !parseSep(pn, ps, ')') {
		ps.error(shouldBeRParen)
	}
}

func (pn *Primary) outputCapture(ps *parser) {
	pn.Type = OutputCapture

	var closer rune
	var shouldBeCloser error
	var cut runePred

	switch ps.next() {
	case '(':
		closer = ')'
		shouldBeCloser = shouldBeRParen
	case '`':
		closer = '`'
		shouldBeCloser = shouldBeBackquote
		cut = isBackquote
	default:
		ps.backup()
		ps.error(shouldBeBackquoteOrLParen)
		ps.next()
		return
	}
	addSep(pn, ps)

	if !startsChunk(ps.peek(), cut) && ps.peek() != closer {
		ps.error(shouldBeChunk)
		return
	}

	pn.setChunk(parseChunk(ps, cut))

	if !parseSep(pn, ps, closer) {
		ps.error(shouldBeCloser)
	}
}

func isBackquote(r rune) bool {
	return r == '`'
}

// List   = '[' { Space } Array ']'
// Lambda = List '{' Chunk '}'
// Map    = '[' { Space } '&' { Space } ']'
//        = '[' { Space } { MapPair { Space } } ']'

func (pn *Primary) lbracket(ps *parser) {
	parseSep(pn, ps, '[')
	parseSpaces(pn, ps)
	r := ps.peek()
	switch {
	case r == ']' || startsArray(r):
		pn.setList(parseArray(ps))
		if !parseSep(pn, ps, ']') {
			ps.error(shouldBeRBracket)
			return
		}
		if parseSep(pn, ps, '{') {
			pn.lambda(ps)
		} else {
			pn.Type = List
		}

	case r == '&':
		pn.Type = Map
		// parseSep(pn, ps, '&')
		amp := ps.pos
		ps.next()
		r := ps.peek()
		switch {
		case isSpace(r), r == ']':
			// '&' { Space } ']': '&' is a sep
			addSep(pn, ps)
			parseSpaces(pn, ps)
		default:
			// { MapPair { Space } } ']': Wind back
			ps.pos = amp
			for ps.peek() == '&' {
				pn.addToMapPairs(parseMapPair(ps, nil))
				parseSpaces(pn, ps)
			}
		}
		if !parseSep(pn, ps, ']') {
			ps.error(shouldBeRBracket)
		}
	default:
		ps.error(shouldBeAmpersandOrArray)
	}
}

// lambda parses a lambda expression. The opening brace has been seen.
func (pn *Primary) lambda(ps *parser) {
	pn.Type = Lambda
	if !startsChunk(ps.peek(), nil) && ps.peek() != '}' {
		ps.error(shouldBeChunk)
	}
	pn.setChunk(parseChunk(ps, nil))
	if !parseSep(pn, ps, '}') {
		ps.error(shouldBeRBrace)
	}
}

// Braced = '{' Compound { (','|'-') Compounds } '}'
// Comma = { Space } [ ',' ] { Space }
func (pn *Primary) lbrace(ps *parser) {
	parseSep(pn, ps, '{')

	if r := ps.peek(); r == ';' || r == '\n' || isSpace(r) {
		pn.lambda(ps)
		return
	}

	pn.Type = Braced

	// XXX: we don't actually know what happens with an empty Compound.
	pn.addToBraced(parseCompound(ps, isBracedSep))
	for isBracedSep(ps.peek()) {
		if ps.peek() == '-' {
			parseSep(pn, ps, '-')
			pn.IsRange = append(pn.IsRange, true)
		} else {
			parseSpaces(pn, ps)
			// optional, so ignore the return value
			parseSep(pn, ps, ',')
			parseSpaces(pn, ps)
			pn.IsRange = append(pn.IsRange, false)
		}
		pn.addToBraced(parseCompound(ps, isBracedSep))
	}
	if !parseSep(pn, ps, '}') {
		ps.error(shouldBeBraceSepOrRBracket)
	}
}

func isBracedSep(r rune) bool {
	return r == ',' || r == '-' || isSpace(r)
}

func (pn *Primary) bareword(ps *parser, cut runePred) {
	pn.Type = Bareword
	defer func() { pn.Value = ps.src[pn.begin:ps.pos] }()
	for allowedInBareword(ps.peek()) && !cut.matches(ps.peek()) {
		ps.next()
	}
}

// The following are allowed in barewords:
// * Anything allowed in variable names
// * The symbols "%+,./=@"
func allowedInBareword(r rune) bool {
	return allowedInVariableName(r) ||
		r == '%' || r == '+' || r == ',' ||
		r == '.' || r == '/' || r == '=' || r == '@'
}

func startsPrimary(r rune, cut runePred) bool {
	if cut.matches(r) {
		return false
	}
	return r == '\'' || r == '"' || r == '$' || allowedInBareword(r) ||
		r == '?' || r == '*' || r == '(' || r == '`' || r == '[' || r == '{'
}

// MapPair = '&' { Space } Compound { Space } Compound
type MapPair struct {
	node
	Key, Value *Compound
}

func (mpn *MapPair) parse(ps *parser, cut runePred) {
	parseSep(mpn, ps, '&')
	parseSpaces(mpn, ps)
	if !startsCompound(ps.peek(), cut) {
		ps.error(shouldBeCompound)
		return
	}
	mpn.setKey(parseCompound(ps, cut))

	parseSpaces(mpn, ps)
	if !startsCompound(ps.peek(), cut) {
		ps.error(shouldBeCompound)
		return
	}
	mpn.setValue(parseCompound(ps, cut))
}

// Sep is the catch-all node type for leaf nodes that lack internal structures
// and semantics, and serve solely for syntactic purposes. The parsing of
// separators depend on the Parent node; as such it lacks a genuine parse
// method.
type Sep struct {
	node
}

func addSep(n Node, ps *parser) {
	var begin int
	ch := n.Children()
	if len(ch) > 0 {
		begin = ch[len(ch)-1].End()
	} else {
		begin = n.Begin()
	}
	addChild(n, &Sep{node{nil, begin, ps.pos, ps.src[begin:ps.pos], nil}})
}

func eatRun(ps *parser, r rune) {
	for ps.peek() == r {
		ps.next()
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

func parseRunAsSep(n Node, ps *parser, isSep func(rune) bool) {
	if !isSep(ps.peek()) {
		return
	}
	ps.next()
	for isSep(ps.peek()) {
		ps.next()
	}
	addSep(n, ps)
}

func parseSpaces(n Node, ps *parser) {
	parseRunAsSep(n, ps, isSpace)
}

// Helpers.

func addChild(p Node, ch Node) {
	p.n().children = append(p.n().children, ch)
	ch.n().parent = p
}

type runePred func(rune) bool

func (rp runePred) matches(r rune) bool {
	return rp != nil && rp(r)
}

// Quote returns a representation of s in elvish syntax. Bareword is tried
// first, then single quoted string and finally double quoted string.
func Quote(s string) string {
	bare := true
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return quoteDouble(s)
		}
		if !allowedInBareword(r) {
			bare = false
		}
	}
	if bare {
		return s
	}
	return quoteSingle(s)
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
	for _, r := range s {
		if r == '\\' || r == '"' {
			buf.WriteByte('\\')
			buf.WriteRune(r)
		} else if !unicode.IsPrint(r) {
			buf.WriteByte('\\')
			if r <= 0xff {
				buf.WriteByte('x')
				buf.Write(rtohex(r, 2))
			} else if r <= 0xffff {
				buf.WriteByte('u')
				buf.Write(rtohex(r, 4))
			} else {
				buf.WriteByte('U')
				buf.Write(rtohex(r, 8))
			}
		} else {
			buf.WriteRune(r)
		}
	}
	buf.WriteByte('"')
	return buf.String()
}
