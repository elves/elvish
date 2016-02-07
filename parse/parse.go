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
	ps := &parser{src, 0, 0, []map[rune]int{{}}, nil}
	bn := parseChunk(ps)
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

func (bn *Chunk) parse(ps *parser) {
	bn.parseSeps(ps)
	for startsPipeline(ps.peek()) {
		bn.addToPipelines(parsePipeline(ps))
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
func (bn *Chunk) parseSeps(ps *parser) int {
	nseps := 0
	for {
		r := ps.peek()
		if isPipelineSep(r) {
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

func startsChunk(r rune) bool {
	return isPipelineSep(r) || startsPipeline(r)
}

// Pipeline = Form { '|' Form }
type Pipeline struct {
	node
	Forms []*Form
}

func (pn *Pipeline) parse(ps *parser) {
	pn.addToForms(parseForm(ps))
	for parseSep(pn, ps, '|') {
		if !startsForm(ps.peek()) {
			ps.error(shouldBeForm)
			return
		}
		pn.addToForms(parseForm(ps))
	}
}

func startsPipeline(r rune) bool {
	return startsForm(r)
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

func (fn *Form) parse(ps *parser) {
	parseSpaces(fn, ps)
	for fn.tryAssignment(ps) {
		parseSpaces(fn, ps)
	}
	if !startsCompound(ps.peek()) {
		if len(fn.Assignments) > 0 {
			return
		} else {
			ps.error(shouldBeCompound)
		}
	}
	fn.setHead(parseCompound(ps))
	parseSpaces(fn, ps)

	for {
		r := ps.peek()
		switch {
		case r == '&':
			fn.addToNamedArgs(parseMapPair(ps))
		case startsCompound(r):
			if ps.hasPrefix("?>") {
				if fn.ExitusRedir != nil {
					ps.error(duplicateExitusRedir)
					// Parse the duplicate redir anyway.
					addChild(fn, parseExitusRedir(ps))
				} else {
					fn.setExitusRedir(parseExitusRedir(ps))
				}
				continue
			}
			cn := parseCompound(ps)
			if isRedirSign(ps.peek()) {
				// Redir
				fn.addToRedirs(parseRedir(ps, cn))
			} else {
				fn.addToArgs(cn)
			}
		case isRedirSign(r):
			fn.addToRedirs(parseRedir(ps, nil))
		default:
			return
		}
		parseSpaces(fn, ps)
	}
}

// tryAssignment tries to parse an assignment. If suceeded, it adds the parsed
// assignment to fn.Assignments and returns true. Otherwise it rewinds the
// parser and returns false.
func (fn *Form) tryAssignment(ps *parser) bool {
	if !startsIndexing(ps.peek()) || ps.peek() == '=' {
		return false
	}

	begin := ps.pos
	var ok bool
	an := parseAssignment(ps, &ok)
	if !ok {
		ps.pos = begin
		return false
	}
	fn.addToAssignments(an)
	return true
}

func startsForm(r rune) bool {
	return isSpace(r) || startsCompound(r)
}

// Assignment = Primary '=' Compound
type Assignment struct {
	node
	Dst *Indexing
	Src *Compound
}

func (an *Assignment) parse(ps *parser, pok *bool) {
	ps.cut('=')
	an.setDst(parseIndexing(ps))
	ps.uncut('=')

	if !parseSep(an, ps, '=') {
		*pok = false
		return
	}
	an.setSrc(parseCompound(ps))
	*pok = true
}

// ExitusRedir = '?' '>' { Space } Compound
type ExitusRedir struct {
	node
	Dest *Compound
}

func (ern *ExitusRedir) parse(ps *parser) {
	ps.next()
	ps.next()
	addSep(ern, ps)
	parseSpaces(ern, ps)
	ern.setDest(parseCompound(ps))
}

// Redir = { Compound } { '<'|'>'|'<>'|'>>' } { Space } ( '&'? Compound )
type Redir struct {
	node
	Dest       *Compound
	Mode       RedirMode
	SourceIsFd bool
	Source     *Compound
}

func (rn *Redir) parse(ps *parser, dest *Compound) {
	// The parsing of the Dest part is done in Form.parse.
	if dest != nil {
		rn.Dest = dest
		rn.begin = dest.begin
		addChild(rn, dest)
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
		ps.error(badRedirSign)
	}
	addSep(rn, ps)
	parseSpaces(rn, ps)
	if parseSep(rn, ps, '&') {
		rn.SourceIsFd = true
	}
	if !startsCompound(ps.peek()) {
		if rn.SourceIsFd {
			ps.error(shouldBeFd)
		} else {
			ps.error(shouldBeFilename)
		}
		return
	}
	rn.setSource(parseCompound(ps))
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

func (cn *Compound) parse(ps *parser) {
	cn.tilde(ps)
	for startsIndexing(ps.peek()) {
		cn.addToIndexings(parseIndexing(ps))
	}
}

// tilde parses a tilde if there is one. It is implemented here instead of
// within Primary since a tilde can only appear as the first part of a
// Compound. Elsewhere tildes are barewords.
func (cn *Compound) tilde(ps *parser) {
	if ps.peek() == '~' {
		ps.next()
		base := node{nil, ps.pos - 1, ps.pos, "~", nil}
		pn := &Primary{node: base, Type: Tilde, Value: "~"}
		in := &Indexing{node: base}
		in.setHead(pn)
		cn.addToIndexings(in)
	}
}

func startsCompound(r rune) bool {
	return startsIndexing(r)
}

// Indexing = Primary { '[' Array ']' }
type Indexing struct {
	node
	Head     *Primary
	Indicies []*Array
}

func (in *Indexing) parse(ps *parser) {
	in.setHead(parsePrimary(ps))
	for parseSep(in, ps, '[') {
		if !startsArray(ps.peek()) {
			ps.error(shouldBeArray)
		}

		ps.pushCutset()
		in.addToIndicies(parseArray(ps))
		ps.popCutset()

		if !parseSep(in, ps, ']') {
			ps.error(shouldBeRBracket)
			return
		}
	}
}

func startsIndexing(r rune) bool {
	return startsPrimary(r)
}

// Array = { Space } { Compound { Space } }
type Array struct {
	node
	Compounds []*Compound
}

func (sn *Array) parse(ps *parser) {
	parseSpaces(sn, ps)
	for startsCompound(ps.peek()) {
		sn.addToCompounds(parseCompound(ps))
		parseSpaces(sn, ps)
	}
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func startsArray(r rune) bool {
	return isSpace(r) || startsIndexing(r)
}

type Primary struct {
	node
	Type PrimaryType
	// The unquoted string value. Valid for Bareword, SingleQuoted,
	// DoubleQuoted, Variable, Wildcard and Tilde.
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
	Tilde
	ErrorCapture
	OutputCapture
	List
	Lambda
	Map
	Braced
)

func (pn *Primary) parse(ps *parser) {
	r := ps.peek()
	if !startsPrimary(r) {
		ps.error(ShouldBePrimary)
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
		pn.bareword(ps)
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

func (pn *Primary) variable(ps *parser) {
	pn.Type = Variable
	defer func() { pn.Value = ps.src[pn.begin+1 : ps.pos] }()
	ps.next()
	// The character of the variable name can be anything.
	if ps.next() == EOF {
		ps.backup()
		ps.error(shouldBeVariableName)
		ps.next()
	}
	for allowedInVariableName(ps.peek()) {
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

	ps.pushCutset()
	pn.setChunk(parseChunk(ps))
	ps.popCutset()

	if !parseSep(pn, ps, ')') {
		ps.error(shouldBeRParen)
	}
}

func (pn *Primary) outputCapture(ps *parser) {
	pn.Type = OutputCapture

	var closer rune
	var shouldBeCloser error

	switch ps.next() {
	case '(':
		closer = ')'
		shouldBeCloser = shouldBeRParen
	case '`':
		closer = '`'
		shouldBeCloser = shouldBeBackquote
	default:
		ps.backup()
		ps.error(shouldBeBackquoteOrLParen)
		ps.next()
		return
	}
	addSep(pn, ps)

	if closer == '`' {
		ps.pushCutset(closer)
	} else {
		ps.pushCutset()
	}
	pn.setChunk(parseChunk(ps))
	ps.popCutset()

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
	ps.pushCutset()

	switch {
	case r == '&':
		pn.Type = Map
		// parseSep(pn, ps, '&')
		amp := ps.pos
		ps.next()
		r := ps.peek()
		switch {
		case isSpace(r), r == ']', r == EOF:
			// '&' { Space } ']': '&' is a sep
			addSep(pn, ps)
			parseSpaces(pn, ps)
		default:
			// { MapPair { Space } } ']': Wind back
			ps.pos = amp
			for ps.peek() == '&' {
				pn.addToMapPairs(parseMapPair(ps))
				parseSpaces(pn, ps)
			}
		}
		ps.popCutset()
		if !parseSep(pn, ps, ']') {
			ps.error(shouldBeRBracket)
		}
	default:
		pn.setList(parseArray(ps))
		ps.popCutset()

		if !parseSep(pn, ps, ']') {
			ps.error(shouldBeRBracket)
			return
		}
		if parseSep(pn, ps, '{') {
			pn.lambda(ps)
		} else {
			pn.Type = List
		}
	}
}

// lambda parses a lambda expression. The opening brace has been seen.
func (pn *Primary) lambda(ps *parser) {
	pn.Type = Lambda
	ps.pushCutset()
	pn.setChunk(parseChunk(ps))
	ps.popCutset()
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
	ps.pushCutset(',', '-')
	pn.addToBraced(parseCompound(ps))
	ps.popCutset()

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
		ps.pushCutset(',', '-')
		pn.addToBraced(parseCompound(ps))
		ps.popCutset()
	}
	if !parseSep(pn, ps, '}') {
		ps.error(shouldBeBraceSepOrRBracket)
	}
}

func isBracedSep(r rune) bool {
	return r == ',' || r == '-' || isSpace(r)
}

func (pn *Primary) bareword(ps *parser) {
	pn.Type = Bareword
	defer func() { pn.Value = ps.src[pn.begin:ps.pos] }()
	for allowedInBareword(ps.peek()) {
		ps.next()
	}
}

// The following are allowed in barewords:
// * Anything allowed in variable names
// * The symbols "%+,./=@~"
func allowedInBareword(r rune) bool {
	return allowedInVariableName(r) ||
		r == '%' || r == '+' || r == ',' || r == '.' ||
		r == '/' || r == '=' || r == '@' || r == '~'
}

func startsPrimary(r rune) bool {
	return r == '\'' || r == '"' || r == '$' || allowedInBareword(r) ||
		r == '?' || r == '*' || r == '(' || r == '`' || r == '[' || r == '{'
}

// MapPair = '&' { Space } Compound { Space } Compound
type MapPair struct {
	node
	Key, Value *Compound
}

func (mpn *MapPair) parse(ps *parser) {
	parseSep(mpn, ps, '&')

	parseSpaces(mpn, ps)
	mpn.setKey(parseCompound(ps))
	if len(mpn.Key.Indexings) == 0 {
		ps.error(shouldBeCompound)
	}

	parseSpaces(mpn, ps)
	mpn.setValue(parseCompound(ps))
	if len(mpn.Value.Indexings) == 0 {
		ps.error(shouldBeCompound)
	}
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
		if !unicode.IsPrint(r) && r != '\n' {
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
