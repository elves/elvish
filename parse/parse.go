// Package parse implements the elvish parser.
package parse

//go:generate ./boilerplate.py
//go:generate stringer -type=PrimaryType,RedirMode -output=string.go

import (
	"bytes"
	"unicode"

	"github.com/elves/elvish/errutil"
)

func addChild(p Node, ch Node) {
	p.n().children = append(p.n().children, ch)
	ch.n().parent = p
}

type runePred func(rune) bool

func (rp runePred) matches(r rune) bool {
	return rp != nil && rp(r)
}

// Sep is the catch-all node type for leaf nodes that lack internal structures
// and semantics, and serve solely for syntactic purposes. The parsing of
// separators depend on the Parent node; as such it lacks a genuine parse
// method.
type Sep struct {
	node
}

func addSep(n Node, rd *reader) {
	var begin int
	ch := n.Children()
	if len(ch) > 0 {
		begin = ch[len(ch)-1].End()
	} else {
		begin = n.Begin()
	}
	addChild(n, &Sep{node{nil, begin, rd.pos, rd.src[begin:rd.pos], nil}})
}

func eatRun(rd *reader, r rune) {
	for rd.peek() == r {
		rd.next()
	}
}

func parseSep(n Node, rd *reader, sep rune) bool {
	if rd.peek() == sep {
		rd.next()
		addSep(n, rd)
		return true
	}
	return false
}

func parseRunAsSep(n Node, rd *reader, isSep func(rune) bool) {
	if !isSep(rd.peek()) {
		return
	}
	rd.next()
	for isSep(rd.peek()) {
		rd.next()
	}
	addSep(n, rd)
}

func parseSpaces(n Node, rd *reader) {
	parseRunAsSep(n, rd, isSpace)
}

// MapPair = '&' { Space } Compound { Space } Compound
type MapPair struct {
	node
	Key, Value *Compound
}

var (
	shouldBeCompound = newError("", "compound")
)

func (mpn *MapPair) parse(rd *reader, cut runePred) {
	parseSep(mpn, rd, '&')
	parseSpaces(mpn, rd)
	if !startsCompound(rd.peek(), cut) {
		rd.error = shouldBeCompound
		return
	}
	mpn.setKey(parseCompound(rd, cut))

	parseSpaces(mpn, rd)
	if !startsCompound(rd.peek(), cut) {
		rd.error = shouldBeCompound
		return
	}
	mpn.setValue(parseCompound(rd, cut))
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

var (
	StringUnterminated   = newError("string not terminated")
	InvalidEscape        = newError("invalid escape sequence")
	InvalidEscapeOct     = newError("invalid escape sequence", "octal digit")
	InvalidEscapeHex     = newError("invalid escape sequence", "hex digit")
	InvalidEscapeControl = newError("invalid control sequence", "a rune between @ (0x40) and _(0x5F)")
	ShouldBePrimary      = newError("",
		"single-quoted string", "double-quoted string", "bareword")
)

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

func (pn *Primary) singleQuoted(rd *reader) {
	pn.Type = SingleQuoted
	rd.next()
	var buf bytes.Buffer
	defer func() { pn.Value = buf.String() }()
	for {
		switch r := rd.next(); r {
		case EOF:
			rd.error = StringUnterminated
			return
		case '\'':
			if rd.peek() == '\'' {
				// Two consecutive single quotes
				rd.next()
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

func (pn *Primary) doubleQuoted(rd *reader) {
	pn.Type = DoubleQuoted
	rd.next()
	var buf bytes.Buffer
	defer func() { pn.Value = buf.String() }()
	for {
		switch r := rd.next(); r {
		case EOF:
			rd.error = StringUnterminated
			return
		case '"':
			return
		case '\\':
			switch r := rd.next(); r {
			case 'c', '^':
				// Control sequence
				r := rd.next()
				if r < 0x40 || r >= 0x60 {
					rd.backup()
					rd.error = InvalidEscapeControl
					return
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
					d, ok := hexToDigit(rd.next())
					if !ok {
						rd.backup()
						rd.error = InvalidEscapeHex
						return
					}
					rr = rr*16 + d
				}
				buf.WriteRune(rr)
			case '0', '1', '2', '3', '4', '5', '6', '7':
				// 2 more octal digits
				rr := r - '0'
				for i := 0; i < 2; i++ {
					r := rd.next()
					if r < '0' || r > '7' {
						rd.backup()
						rd.error = InvalidEscapeOct
						return
					}
					rr = rr*8 + (r - '0')
				}
				buf.WriteRune(rr)
			default:
				if rr, ok := doubleEscape[r]; ok {
					buf.WriteRune(rr)
				} else {
					rd.backup()
					rd.error = InvalidEscape
				}
			}
		default:
			buf.WriteRune(r)
		}
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

var (
	shouldBeVariableName = newError("", "variable name")
)

func (pn *Primary) variable(rd *reader, cut runePred) {
	pn.Type = Variable
	defer func() { pn.Value = rd.src[pn.begin:rd.pos] }()
	rd.next()
	// The character of the variable name can be anything.
	if r := rd.next(); cut.matches(r) {
		rd.backup()
		rd.error = shouldBeVariableName
		return
	}
	for allowedInVariableName(rd.peek()) && !cut.matches(rd.peek()) {
		rd.next()
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

func (pn *Primary) bareword(rd *reader, cut runePred) {
	pn.Type = Bareword
	defer func() { pn.Value = rd.src[pn.begin:rd.pos] }()
	for allowedInBareword(rd.peek()) && !cut.matches(rd.peek()) {
		rd.next()
	}
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

// Quote returns a representation of s in elvish syntax. Bareword is tried
// first after single quoted string and finally double quoted string.
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

func isWildcard(r rune) bool {
	return r == '*' || r == '?'
}

func (pn *Primary) wildcard(rd *reader) {
	pn.Type = Wildcard
	for isWildcard(rd.peek()) {
		rd.next()
	}
	pn.Value = rd.src[pn.begin:rd.pos]
}

var (
	shouldBeChunk  = newError("", "chunk")
	shouldBeRParen = newError("", "')'")
)

func (pn *Primary) exitusCapture(rd *reader) {
	rd.next()
	rd.next()
	addSep(pn, rd)

	pn.Type = ErrorCapture
	if !startsChunk(rd.peek(), nil) && rd.peek() != ')' {
		rd.error = shouldBeChunk
		return
	}
	pn.setChunk(parseChunk(rd, nil))
	if !parseSep(pn, rd, ')') {
		rd.error = shouldBeRParen
	}
}

var (
	shouldBeBackquoteOrLParen = newError("", "'`'", "'('")
	shouldBeBackquote         = newError("", "'`'")
)

func isBackquote(r rune) bool {
	return r == '`'
}

func (pn *Primary) outputCapture(rd *reader) {
	pn.Type = OutputCapture

	var closer rune
	var shouldBeCloser error
	var cut runePred

	switch rd.next() {
	case '(':
		closer = ')'
		shouldBeCloser = shouldBeRParen
	case '`':
		closer = '`'
		shouldBeCloser = shouldBeBackquote
		cut = isBackquote
	default:
		rd.backup()
		rd.error = shouldBeBackquoteOrLParen
		return
	}
	addSep(pn, rd)

	if !startsChunk(rd.peek(), cut) && rd.peek() != closer {
		rd.error = shouldBeChunk
		return
	}

	pn.setChunk(parseChunk(rd, cut))

	if !parseSep(pn, rd, closer) {
		rd.error = shouldBeCloser
	}
}

// List   = '[' { Space } Array ']'
// Lambda = List '{' Chunk '}'
// Map    = '[' { Space } '&' { Space } ']'
//        = '[' { Space } { MapPair { Space } } ']'

var (
	shouldBeAmpersandOrArray = newError("", "'&'", "spaced")
	shouldBeRBracket         = newError("", "']'")
	shouldBeRBrace           = newError("", "'}'")
	shouldBeAmpersand        = newError("", "'&'")
)

func (pn *Primary) lbracket(rd *reader) {
	parseSep(pn, rd, '[')
	parseSpaces(pn, rd)
	r := rd.peek()
	switch {
	case r == ']' || startsArray(r):
		pn.setList(parseArray(rd))
		if !parseSep(pn, rd, ']') {
			rd.error = shouldBeRBracket
			return
		}
		if !parseSep(pn, rd, '{') {
			// List
			pn.Type = List
			return
		}
		pn.Type = Lambda
		if !startsChunk(rd.peek(), nil) && rd.peek() != '}' {
			rd.error = shouldBeChunk
			return
		}
		pn.setChunk(parseChunk(rd, nil))
		if !parseSep(pn, rd, '}') {
			rd.error = shouldBeRBrace
			return
		}
	case r == '&':
		pn.Type = Map
		// parseSep(pn, rd, '&')
		amp := rd.pos
		rd.next()
		r := rd.peek()
		switch {
		case isSpace(r), r == ']':
			// '&' { Space } ']': '&' is a sep
			addSep(pn, rd)
			parseSpaces(pn, rd)
		default:
			// { MapPair { Space } } ']': Wind back
			rd.pos = amp
			for rd.peek() == '&' {
				pn.addToMapPairs(parseMapPair(rd, nil))
				parseSpaces(pn, rd)
			}
		}
		if !parseSep(pn, rd, ']') {
			rd.error = shouldBeRBracket
		}
	default:
		rd.error = shouldBeAmpersandOrArray
	}
}

func isBracedSep(r rune) bool {
	return r == ',' || r == '-' || isSpace(r)
}

var (
	shouldBeBraceSepOrRBracket = newError("", "'-'", "','", "'}'")
)

// Braced = '{' Compound { (','|'-') Compounds } '}'
// Comma = { Space } [ ',' ] { Space }
func (pn *Primary) braced(rd *reader) {
	pn.Type = Braced
	parseSep(pn, rd, '{')
	// XXX: we don't actually know what happens with an empty Compound.
	pn.addToBraced(parseCompound(rd, isBracedSep))
	for isBracedSep(rd.peek()) {
		if rd.peek() == '-' {
			parseSep(pn, rd, '-')
			pn.IsRange = append(pn.IsRange, true)
		} else {
			parseSpaces(pn, rd)
			// optional, so ignore the return value
			parseSep(pn, rd, ',')
			parseSpaces(pn, rd)
			pn.IsRange = append(pn.IsRange, false)
		}
		pn.addToBraced(parseCompound(rd, isBracedSep))
	}
	if !parseSep(pn, rd, '}') {
		rd.error = shouldBeBraceSepOrRBracket
	}
}

func startsPrimary(r rune, cut runePred) bool {
	if cut.matches(r) {
		return false
	}
	return r == '\'' || r == '"' || r == '$' || allowedInBareword(r) ||
		r == '?' || r == '*' || r == '(' || r == '`' || r == '[' || r == '{'
}

func (pn *Primary) parse(rd *reader, cut runePred) {
	r := rd.peek()
	if !startsPrimary(r, cut) {
		rd.error = ShouldBePrimary
		return
	}
	switch r {
	case '\'':
		pn.singleQuoted(rd)
	case '"':
		pn.doubleQuoted(rd)
	case '$':
		pn.variable(rd, cut)
	case '*':
		pn.wildcard(rd)
	case '?':
		if rd.hasPrefix("?(") {
			pn.exitusCapture(rd)
		} else {
			pn.wildcard(rd)
		}
	case '(', '`':
		pn.outputCapture(rd)
	case '[':
		pn.lbracket(rd)
	case '{':
		pn.braced(rd)
	default:
		pn.bareword(rd, cut)
	}
}

// Indexed = Primary { '[' Array ']' }
type Indexed struct {
	node
	Head     *Primary
	Indicies []*Array
}

func startsIndexed(r rune, cut runePred) bool {
	return startsPrimary(r, cut)
}

var (
	shouldBeArray = newError("", "spaced")
)

func (in *Indexed) parse(rd *reader, cut runePred) {
	in.setHead(parsePrimary(rd, cut))
	for parseSep(in, rd, '[') {
		if !startsArray(rd.peek()) {
			rd.error = shouldBeArray
		}
		in.addToIndicies(parseArray(rd))
		if !parseSep(in, rd, ']') {
			rd.error = shouldBeRBracket
			return
		}
	}
}

// Compound = { Indexed }
type Compound struct {
	node
	Indexeds []*Indexed
}

func startsCompound(r rune, cut runePred) bool {
	return startsIndexed(r, cut)
}

func (cn *Compound) parse(rd *reader, cut runePred) {
	for startsIndexed(rd.peek(), cut) {
		cn.addToIndexeds(parseIndexed(rd, cut))
	}
}

// Array = { Space } { Compound { Space } }
type Array struct {
	node
	Compounds []*Compound
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func startsArray(r rune) bool {
	return isSpace(r) || startsIndexed(r, nil)
}

func (sn *Array) parse(rd *reader) {
	parseSpaces(sn, rd)
	for startsCompound(rd.peek(), nil) {
		sn.addToCompounds(parseCompound(rd, nil))
		parseSpaces(sn, rd)
	}
}

type RedirMode int

const (
	BadRedirMode RedirMode = iota
	Read
	Write
	ReadWrite
	Append
)

// Redir = { Compound } { '<'|'>'|'<>'|'>>' } { Space } ( '&'? Compound )
type Redir struct {
	node
	Dest       *Compound
	Mode       RedirMode
	SourceIsFd bool
	Source     *Compound
}

func isRedirSign(r rune) bool {
	return r == '<' || r == '>'
}

var (
	badRedirSign     = newError("bad redir sign", "'<'", "'>'", "'>>'", "'<>'")
	shouldBeFd       = newError("", "a composite term representing fd")
	shouldBeFilename = newError("", "a composite term representing filename")
)

// XXX(xiaq): The parsing of the Dest part is done in Form.parse.
func (rn *Redir) parse(rd *reader, cut runePred, dest *Compound) {
	if dest != nil {
		rn.Dest = dest
		rn.begin = dest.begin
		addChild(rn, dest)
	}

	begin := rd.pos
	for !cut.matches(rd.peek()) && isRedirSign(rd.peek()) {
		rd.next()
	}
	sign := rd.src[begin:rd.pos]
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
		rd.pos = begin
		rd.error = badRedirSign
		return
	}
	addSep(rn, rd)
	parseSpaces(rn, rd)
	if !cut.matches('&') && parseSep(rn, rd, '&') {
		rn.SourceIsFd = true
	}
	if !startsCompound(rd.peek(), cut) {
		if rn.SourceIsFd {
			rd.error = shouldBeFd
		} else {
			rd.error = shouldBeFilename
		}
		return
	}
	rn.setSource(parseCompound(rd, cut))
}

// ExitusRedir = '?' '>' { Space } Compound
type ExitusRedir struct {
	node
	Dest *Compound
}

func (ern *ExitusRedir) parse(rd *reader, cut runePred) {
	rd.next()
	rd.next()
	addSep(ern, rd)
	parseSpaces(ern, rd)
	ern.setDest(parseCompound(rd, cut))
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

func startsForm(r rune, cut runePred) bool {
	return isSpace(r) || startsCompound(r, cut)
}

var (
	duplicateExitusRedir = newError("duplicate exitus redir")
)

// tryAssignment tries to parse an assignment. If suceeded, it adds the parsed
// assignment to fn.Assignments and returns true. Otherwise it rewinds the
// reader and returns false.
func (fn *Form) tryAssignment(rd *reader, cut runePred) bool {
	begin := rd.pos
	an := parseAssignment(rd, cut)
	if rd.error != nil {
		rd.pos = begin
		rd.error = nil
		return false
	}
	fn.addToAssignments(an)
	return true
}

func (fn *Form) parse(rd *reader, cut runePred) {
	parseSpaces(fn, rd)
	for fn.tryAssignment(rd, cut) {
		parseSpaces(fn, rd)
	}
	if !startsCompound(rd.peek(), cut) {
		if len(fn.Assignments) == 0 {
			rd.error = shouldBeCompound
		}
		return
	}
	fn.setHead(parseCompound(rd, cut))
	parseSpaces(fn, rd)
loop:
	for {
		r := rd.peek()
		switch {
		case cut.matches(r):
			break loop
		case r == '&':
			fn.addToNamedArgs(parseMapPair(rd, cut))
		case startsCompound(r, cut):
			if !cut.matches('>') && rd.hasPrefix("?>") {
				if fn.ExitusRedir != nil {
					rd.error = duplicateExitusRedir
					return
				}
				fn.setExitusRedir(parseExitusRedir(rd, cut))
				continue
			}
			cn := parseCompound(rd, cut)
			if !cut.matches(rd.peek()) && isRedirSign(rd.peek()) {
				// Redir
				fn.addToRedirs(parseRedir(rd, cut, cn))
			} else {
				fn.addToArgs(cn)
			}
		case isRedirSign(r):
			fn.addToRedirs(parseRedir(rd, cut, nil))
		default:
			return
		}
		parseSpaces(fn, rd)
	}
}

// Assignment = Primary '=' Compound
type Assignment struct {
	node
	Dst *Indexed
	Src *Compound
}

var shouldBeEqual = newError("", "=")

func (an *Assignment) parse(rd *reader, cut runePred) {
	cutWithEqual := runePred(func(r rune) bool {
		return cut.matches(r) || r == '='
	})
	an.setDst(parseIndexed(rd, cutWithEqual))
	if !parseSep(an, rd, '=') {
		rd.error = shouldBeEqual
		return
	}
	an.setSrc(parseCompound(rd, cut))
}

// Pipeline = Form { '|' Form }
type Pipeline struct {
	node
	Forms []*Form
}

func startsPipeline(r rune, cut runePred) bool {
	return startsForm(r, cut)
}

var (
	shouldBeForm = newError("", "form")
)

func (pn *Pipeline) parse(rd *reader, cut runePred) {
	pn.addToForms(parseForm(rd, cut))
	for !cut.matches('|') && parseSep(pn, rd, '|') {
		if !startsForm(rd.peek(), cut) {
			rd.error = shouldBeForm
			return
		}
		pn.addToForms(parseForm(rd, cut))
	}
}

// Chunk = { PipelineSep | Space } { Pipeline { PipelineSep | Space } }
type Chunk struct {
	node
	Pipelines []*Pipeline
}

func isPipelineSep(r rune) bool {
	return r == '\n' || r == ';'
}

func startsChunk(r rune, cut runePred) bool {
	return isPipelineSep(r) || startsPipeline(r, cut)
}

// parseSeps parses pipeline separators along with whitespaces. It returns the
// number of pipeline separators parsed.
func (bn *Chunk) parseSeps(rd *reader, cut runePred) int {
	nseps := 0
	for {
		r := rd.peek()
		if cut.matches(r) {
			break
		} else if isPipelineSep(r) {
			// parse as a Sep
			parseSep(bn, rd, r)
			nseps += 1
		} else if isSpace(r) {
			// parse a run of spaces as a Sep
			parseSpaces(bn, rd)
		} else if r == '#' {
			// parse a comment as a Sep
			for {
				r := rd.peek()
				if r == EOF || r == '\n' {
					break
				}
				rd.next()
			}
			nseps += 1
		} else {
			break
		}
	}
	return nseps
}

func (bn *Chunk) parse(rd *reader, cut runePred) {
	bn.parseSeps(rd, cut)
	for startsPipeline(rd.peek(), cut) {
		bn.addToPipelines(parsePipeline(rd, cut))
		if bn.parseSeps(rd, cut) == 0 {
			break
		}
	}
}

func Parse(name, src string) (*Chunk, error) {
	rd := &reader{src, 0, 0, nil}
	bn := parseChunk(rd, nil)
	if rd.error != nil {
		return bn, errutil.NewContextualError(
			name, "syntax error", src, rd.pos, rd.error.Error())
	}
	if rd.pos != len(src) {
		return bn, errutil.NewContextualError(
			name, "syntax error", src, rd.pos, "unexpected rune")
	}
	return bn, nil
}
