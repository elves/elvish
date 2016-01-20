package parse

//go:generate ./boilerplate.py
//go:generate stringer -type=PrimaryType,RedirMode -output=string.go

import "bytes"

type Node interface {
	n() *node
}

type node struct {
	Parent     Node
	Begin, End int
	SourceText string
	Children   []Node
}

func (n *node) n() *node {
	return n
}

func addChild(p Node, ch Node) {
	p.n().Children = append(p.n().Children, ch)
	ch.n().Parent = p
}

// Sep is the catch-all node type for leaf nodes that lack internal structures
// and semantics, and serve solely for syntactic purposes. The parsing of
// separators depend on the Parent node; as such it lacks a genuine parse
// method.
type Sep struct {
	node
}

func addSep(n Node, begin, end int) {
	addChild(n, &Sep{node{Begin: begin, End: end}})
}

func eatRun(rd *reader, r rune) {
	for rd.peek() == r {
		rd.next()
	}
}

func parseSep(n Node, rd *reader, sep rune) bool {
	begin := rd.pos
	if rd.peek() == sep {
		rd.next()
		addSep(n, begin, rd.pos)
		return true
	}
	return false
}

func parseRunAsSep(n Node, rd *reader, isSep func(rune) bool) {
	if !isSep(rd.peek()) {
		return
	}
	begin := rd.pos
	rd.next()
	for isSep(rd.peek()) {
		rd.next()
	}
	addSep(n, begin, rd.pos)
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

func (mpn *MapPair) parse(rd *reader) {
	parseSep(mpn, rd, '&')
	parseSpaces(mpn, rd)
	if !startsCompound(rd.peek()) {
		rd.error = shouldBeCompound
		return
	}
	mpn.setKey(parseCompound(rd))

	parseSpaces(mpn, rd)
	if !startsCompound(rd.peek()) {
		rd.error = shouldBeCompound
		return
	}
	mpn.setValue(parseCompound(rd))
}

type Primary struct {
	node
	Type PrimaryType
	// The unquoted string value. Valid for Bareword, SingleQuoted,
	// DoubleQuoted, Variable and Wildcard.
	Value    string
	List     *Spaced    // Valid for List and Lambda
	Chunk    *Chunk     // Valid for OutputCapture, ExitusCapture and Lambda
	MapPairs []*MapPair // Valid for Map
}

type PrimaryType int

const (
	BadPrimary PrimaryType = iota
	Bareword
	SingleQuoted
	DoubleQuoted
	Variable
	Wildcard
	ExitusCapture
	OutputCapture
	List
	Lambda
	Map
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
// * Anything beyond ASCII
// * Letters and numbers
// * The symbols "-_:"
func allowedInVariableName(r rune) bool {
	return r >= 0x80 ||
		('0' <= r && r <= '9') ||
		('a' <= r && r <= 'z') ||
		('A' <= r && r <= 'Z') ||
		r == '-' || r == '_' || r == ':'
}

var (
	shouldBeVariableName = newError("", "variable name")
)

func (pn *Primary) variable(rd *reader) {
	pn.Type = Variable
	defer func() { pn.Value = rd.src[pn.Begin:rd.pos] }()
	rd.next()
	if !allowedInVariableName(rd.next()) {
		rd.backup()
		rd.error = shouldBeVariableName
		return
	}
	for allowedInVariableName(rd.peek()) {
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

func (pn *Primary) bareword(rd *reader) {
	pn.Type = Bareword
	defer func() { pn.Value = rd.src[pn.Begin:rd.pos] }()
	for allowedInBareword(rd.peek()) {
		rd.next()
	}
}

func isWildcard(r rune) bool {
	return r == '*' || r == '?'
}

func (pn *Primary) wildcard(rd *reader) {
	pn.Type = Wildcard
	for isWildcard(rd.peek()) {
		rd.next()
	}
	pn.Value = rd.src[pn.Begin:rd.pos]
}

var (
	shouldBeChunk  = newError("", "chunk")
	shouldBeRParen = newError("", "')'")
)

func (pn *Primary) exitusCapture(rd *reader) {
	begin := rd.pos
	rd.next()
	rd.next()
	addSep(pn, begin, rd.pos)

	pn.Type = ExitusCapture
	if !startsChunk(rd.peek()) && rd.peek() != ')' {
		rd.error = shouldBeChunk
		return
	}
	pn.setChunk(parseChunk(rd))
	if !parseSep(pn, rd, ')') {
		rd.error = shouldBeRParen
	}
}

func (pn *Primary) outputCapture(rd *reader) {
	pn.Type = OutputCapture
	parseSep(pn, rd, '(')
	if !startsChunk(rd.peek()) && rd.peek() != ')' {
		rd.error = shouldBeChunk
		return
	}
	pn.setChunk(parseChunk(rd))
	if !parseSep(pn, rd, ')') {
		rd.error = shouldBeRParen
	}
}

// List   = '[' { Space } Spaced ']'
// Lambda = List '{' Chunk '}'
// Map    = '[' { Space } '&' { Space } ']'
//        = '[' { Space } { MapPair { Space } } ']'

var (
	shouldBeAmpersandOrSpaced = newError("", "'&'", "spaced")
	shouldBeRBracket          = newError("", "']'")
	shouldBeRBrace            = newError("", "'}'")
	shouldBeAmpersand         = newError("", "'&'")
)

func (pn *Primary) lbracket(rd *reader) {
	parseSep(pn, rd, '[')
	parseSpaces(pn, rd)
	r := rd.peek()
	switch {
	case r == ']' || startsSpaced(r):
		pn.setList(parseSpaced(rd))
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
		if !startsChunk(rd.peek()) && rd.peek() != '}' {
			rd.error = shouldBeChunk
			return
		}
		pn.setChunk(parseChunk(rd))
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
			addSep(pn, amp, rd.pos)
			parseSpaces(pn, rd)
		default:
			// { MapPair { Space } } ']': Wind back
			rd.pos = amp
			for rd.peek() == '&' {
				pn.addToMapPairs(parseMapPair(rd))
				parseSpaces(pn, rd)
			}
		}
		if !parseSep(pn, rd, ']') {
			rd.error = shouldBeRBracket
		}
	default:
		rd.error = shouldBeAmpersandOrSpaced
	}
}

func startsPrimary(r rune) bool {
	return r == '\'' || r == '"' || r == '$' || allowedInBareword(r) ||
		r == '?' || r == '*' || r == '(' || r == '['
}

func (pn *Primary) parse(rd *reader) {
	r := rd.peek()
	if !startsPrimary(r) {
		rd.error = ShouldBePrimary
		return
	}
	switch r {
	case '\'':
		pn.singleQuoted(rd)
	case '"':
		pn.doubleQuoted(rd)
	case '$':
		pn.variable(rd)
	case '*':
		pn.wildcard(rd)
	case '?':
		if rd.hasPrefix("?(") {
			pn.exitusCapture(rd)
		} else {
			pn.wildcard(rd)
		}
	case '(':
		pn.outputCapture(rd)
	case '[':
		pn.lbracket(rd)
	default:
		pn.bareword(rd)
	}
}

// Indexed = Primary { '[' Spaced ']' }
type Indexed struct {
	node
	Head     *Primary
	Indicies []*Spaced
}

func startsIndexed(r rune) bool {
	return startsPrimary(r)
}

var (
	shouldBeSpaced = newError("", "spaced")
)

func (in *Indexed) parse(rd *reader) {
	in.setHead(parsePrimary(rd))
	for parseSep(in, rd, '[') {
		if !startsSpaced(rd.peek()) {
			rd.error = shouldBeSpaced
		}
		in.addToIndicies(parseSpaced(rd))
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

func startsCompound(r rune) bool {
	return startsIndexed(r)
}

func (cn *Compound) parse(rd *reader) {
	for startsIndexed(rd.peek()) {
		cn.addToIndexeds(parseIndexed(rd))
	}
}

// Spaced = { Space } { Indexed { Space } }
type Spaced struct {
	node
	Compounds []*Compound
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func startsSpaced(r rune) bool {
	return isSpace(r) || startsIndexed(r)
}

func (sn *Spaced) parse(rd *reader) {
	parseSpaces(sn, rd)
	for startsCompound(rd.peek()) {
		sn.addToCompounds(parseCompound(rd))
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
func (rn *Redir) parse(rd *reader) {
	begin := rd.pos
	for isRedirSign(rd.peek()) {
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
	addSep(rn, begin, rd.pos)
	parseSpaces(rn, rd)
	if parseSep(rn, rd, '&') {
		rn.SourceIsFd = true
	}
	if !startsCompound(rd.peek()) {
		if rn.SourceIsFd {
			rd.error = shouldBeFd
		} else {
			rd.error = shouldBeFilename
		}
		return
	}
	rn.setSource(parseCompound(rd))
}

// ExitusRedir = '?' '>' { Space } Compound
type ExitusRedir struct {
	node
	Dest *Compound
}

func (ern *ExitusRedir) parse(rd *reader) {
	begin := rd.pos
	rd.next()
	rd.next()
	addSep(ern, begin, rd.pos)
	parseSpaces(ern, rd)
	ern.setDest(parseCompound(rd))
}

// Form = { Space } Compound { Space } { ( Compound | MapPair | Redir | ExitusRedir ) { Space } }
type Form struct {
	node
	Head        *Compound
	Compounds   []*Compound
	MapPairs    []*MapPair
	Redirs      []*Redir
	ExitusRedir *ExitusRedir
}

func startsForm(r rune) bool {
	return isSpace(r) || startsCompound(r)
}

var (
	duplicateExitusRedir = newError("duplicate exitus redir")
)

func (fn *Form) parse(rd *reader) {
	parseSpaces(fn, rd)
	if !startsCompound(rd.peek()) {
		rd.error = shouldBeCompound
		return
	}
	fn.setHead(parseCompound(rd))
	parseSpaces(fn, rd)
	for {
		r := rd.peek()
		switch {
		case r == '&':
			fn.addToMapPairs(parseMapPair(rd))
		case startsCompound(r):
			if rd.hasPrefix("?>") {
				if fn.ExitusRedir != nil {
					rd.error = duplicateExitusRedir
					return
				}
				fn.setExitusRedir(parseExitusRedir(rd))
				continue
			}
			cn := parseCompound(rd)
			if isRedirSign(rd.peek()) {
				// Redir
				rn := parseRedir(rd)
				// XXX(xiaq): Redir.parse doesn't deal with Dest, so we patch
				// it here.
				rn.Begin = cn.Begin
				rn.Dest = cn

				children := rn.Children
				rn.Children = make([]Node, len(children)+1)
				copy(rn.Children[1:], children)
				rn.Children[0] = cn

				fn.addToRedirs(rn)
			} else {
				fn.addToCompounds(cn)
			}
		case isRedirSign(r):
			fn.addToRedirs(parseRedir(rd))
		default:
			return
		}
		parseSpaces(fn, rd)
	}
}

// Pipeline = Form { '|' Form }
type Pipeline struct {
	node
	Forms []*Form
}

func startsPipeline(r rune) bool {
	return startsForm(r)
}

var (
	shouldBeForm = newError("", "form")
)

func (pn *Pipeline) parse(rd *reader) {
	pn.addToForms(parseForm(rd))
	for parseSep(pn, rd, '|') {
		if !startsForm(rd.peek()) {
			rd.error = shouldBeForm
			return
		}
		pn.addToForms(parseForm(rd))
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

func isPipelineSepOrSpace(r rune) bool {
	return isPipelineSep(r) || isSpace(r)
}

func startsChunk(r rune) bool {
	return isPipelineSep(r) || startsPipeline(r)
}

func (bn *Chunk) parse(rd *reader) {
	parseRunAsSep(bn, rd, isPipelineSepOrSpace)
	for startsPipeline(rd.peek()) {
		bn.addToPipelines(parsePipeline(rd))
		parseRunAsSep(bn, rd, isPipelineSepOrSpace)
	}
}

func Parse(src string) (*Chunk, error) {
	rd := &reader{src, 0, nil}
	bn := parseChunk(rd)
	if rd.error != nil {
		return nil, rd.error
	}
	return bn, nil
}
