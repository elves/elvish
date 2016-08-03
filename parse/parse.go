// Package parse implements the elvish parser.
package parse

//go:generate ./boilerplate.py
//go:generate stringer -type=PrimaryType,RedirMode,ControlKind -output=string.go

import (
	"bytes"
	"errors"
	"fmt"
	"unicode"
)

// Parse parses elvish source.
func Parse(src string) (*Chunk, error) {
	ps := &parser{src, 0, 0, []map[rune]int{{}}, 0, nil}
	bn := parseChunk(ps)
	if ps.pos != len(src) {
		ps.error(errUnexpectedRune)
	}
	var err error
	if ps.errors != nil {
		err = ps.errors
	}
	return bn, err
}

// Errors.
var (
	errUnexpectedRune         = errors.New("unexpected rune")
	errShouldBeForm           = newError("", "form")
	errDuplicateExitusRedir   = newError("duplicate exitus redir")
	errShouldBeThen           = newError("", "then")
	errShouldBeElifOrElseOrFi = newError("", "elif", "else", "fi")
	errShouldBeFi             = newError("", "fi")
	errShouldBeTried          = newError("", "tried")
	errShouldBeDo             = newError("", "do")
	errShouldBeDone           = newError("", "done")
	errShouldBeIn             = newError("", "in")
	errShouldBePipelineSep    = newError("", "';'", "newline")
	errShouldBeEnd            = newError("", "end")
	errBadRedirSign           = newError("bad redir sign", "'<'", "'>'", "'>>'", "'<>'")
	errShouldBeFD             = newError("", "a composite term representing fd")
	errShouldBeFilename       = newError("", "a composite term representing filename")
	errShouldBeArray          = newError("", "spaced")
	errStringUnterminated     = newError("string not terminated")
	errInvalidEscape          = newError("invalid escape sequence")
	errInvalidEscapeOct       = newError("invalid escape sequence", "octal digit")
	errInvalidEscapeHex       = newError("invalid escape sequence", "hex digit")
	errInvalidEscapeControl   = newError("invalid control sequence", "a rune between @ (0x40) and _(0x5F)")
	errShouldBePrimary        = newError("",
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
	errArgListAllowNoSemicolon    = newError("argument list doesn't allow semicolons")
)

// Chunk = { PipelineSep | Space } { Pipeline { PipelineSep | Space } }
type Chunk struct {
	node
	Pipelines []*Pipeline
}

func (bn *Chunk) parse(ps *parser) {
	bn.parseSeps(ps)
	for startsPipeline(ps.peek()) {
		leader, starter := findLeader(ps)
		if leader != "" && !starter && ps.controls > 0 {
			// We found a non-starting leader and there is a control block that
			// has not been closed. Stop parsing this chunk. We don't check the
			// validity of the leader; the checking is done where the control
			// block is parsed (e.g. (*Form).parseIf).
			break
		}
		// We have more chance to check for validity of the leader, but
		// eventually it will be checked in (*Form).parse. So we don't check it
		// here, for more uniform error reporting and recovery.

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
			nseps++
		} else if isSpace(r) {
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

func (pn *Pipeline) parse(ps *parser) {
	pn.addToForms(parseForm(ps))
	for parseSep(pn, ps, '|') {
		parseSpacesAndNewlines(pn, ps)
		if !startsForm(ps.peek()) {
			ps.error(errShouldBeForm)
			return
		}
		pn.addToForms(parseForm(ps))
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

// findLeader look aheads a command leader. It returns the leader and whether
// it starts a control block.
func findLeader(ps *parser) (string, bool) {
	switch leader := ps.findPossibleLeader(); leader {
	case "if", "while", "for", "try", "begin":
		// Starting leaders are always legal.
		return leader, true
	case "then", "elif", "else", "fi", "do", "done", "except", "finally", "tried", "end":
		return leader, false
	default:
		// There is no leader.
		return "", false
	}
}

func hasLeader(ps *parser, l string) bool {
	s, _ := findLeader(ps)
	return s == l
}

// Form = { Space } { { Assignment } { Space } }
//        { Compound | Control } { Space } { ( Compound | MapPair | Redir | ExitusRedir ) { Space } }
type Form struct {
	node
	Assignments []*Assignment
	Control     *Control
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
	leader, starter := findLeader(ps)
	if leader != "" && starter {
		// Parse Control.
		fn.setControl(parseControl(ps, leader))
		parseSpaces(fn, ps)
	} else {
		if leader != "" {
			ps.error(fmt.Errorf("bogus command leader %q ignored", leader))
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
		fn.setHead(parseCompound(ps, true))
		parseSpaces(fn, ps)
	}

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
			fn.addToNamedArgs(parseMapPair(ps))
		case startsCompound(r, false):
			if ps.hasPrefix("?>") {
				if fn.ExitusRedir != nil {
					ps.error(errDuplicateExitusRedir)
					// Parse the duplicate redir anyway.
					addChild(fn, parseExitusRedir(ps))
				} else {
					fn.setExitusRedir(parseExitusRedir(ps))
				}
				continue
			}
			cn := parseCompound(ps, false)
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
	if !startsIndexing(ps.peek(), false) || ps.peek() == '=' {
		return false
	}

	pos := ps.pos
	errors := ps.errors
	an := parseAssignment(ps)
	if ps.errors != errors {
		ps.errors = errors
		ps.pos = pos
		return false
	}
	fn.addToAssignments(an)
	return true
}

func startsForm(r rune) bool {
	return isSpace(r) || startsCompound(r, true)
}

// Assignment = Primary '=' Compound
type Assignment struct {
	node
	Dst *Indexing
	Src *Compound
}

func (an *Assignment) parse(ps *parser) {
	ps.cut('=')
	an.setDst(parseIndexing(ps, false))
	head := an.Dst.Head
	if !checkVariableInAssignment(head, ps) {
		ps.errorp(head.Begin(), head.End(), errShouldBeVariableName)
	}
	ps.uncut('=')

	if !parseSep(an, ps, '=') {
		ps.error(errShouldBeEqual)
	}
	an.setSrc(parseCompound(ps, false))
}

func checkVariableInAssignment(p *Primary, ps *parser) bool {
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

// Control = IfControl | WhileControl | ForControl | BeginControl
// IfControl = If Chunk Then Chunk { Elif Chunk Then Chunk } [ Else Chunk ] Fi
// WhileControl = While Chunk Do Chunk [ Else Chunk ] Done
// ForControl = For Primary In Array PipelineSep Do Chunk [ Else Chunk ] Done
// BeginControl = Begin Chunk Done
// If = "if" Space { Space }
// (Similiar for Then, Elif, Else, Fi, While, Do, Done, For, Begin, End)
type Control struct {
	node
	Kind        ControlKind
	Condition   *Chunk    // Valid for WhileControl.
	Iterator    *Indexing // Valid for ForControl.
	Array       *Array    // Valid for ForControl.
	Body        *Chunk    // Valid for all except IfControl.
	Conditions  []*Chunk  // Valid for IfControl.
	Bodies      []*Chunk  // Valid for IfControl.
	ElseBody    *Chunk    // Valid for IfControl, WhileControl and ForControl.
	ExceptBody  *Chunk    // Valid for TryControl.
	ExceptVar   *Indexing // Valid for TryControl.
	FinallyBody *Chunk    // Valid for TryControl.
}

// ControlKind identifies which control structure a Control represents.
type ControlKind int

// Possible values of ControlKind.
const (
	BadControl ControlKind = iota
	IfControl
	WhileControl
	ForControl
	TryControl
	BeginControl
)

func (ctrl *Control) parse(ps *parser, leader string) {
	ps.advance(len(leader))
	addSep(ctrl, ps)

	ps.controls++
	defer func() { ps.controls-- }()

	consumeLeader := func() string {
		leader, _ := findLeader(ps)
		if len(leader) > 0 {
			ps.advance(len(leader))
			addSep(ctrl, ps)
		}
		return leader
	}

	doElseDone := func() {
		parseSpaces(ctrl, ps)
		if consumeLeader() != "do" {
			ps.error(errShouldBeDo)
		}
		ctrl.setBody(parseChunk(ps))
		if hasLeader(ps, "else") {
			consumeLeader()
			ctrl.setElseBody(parseChunk(ps))
		}
		if consumeLeader() != "done" {
			ps.error(errShouldBeDone)
		}
	}

	switch leader {
	case "if":
		ctrl.Kind = IfControl
		ctrl.addToConditions(parseChunk(ps))
		if consumeLeader() != "then" {
			ps.error(errShouldBeThen)
		}
		ctrl.addToBodies(parseChunk(ps))
	Elifs:
		for {
			switch consumeLeader() {
			case "fi":
				break Elifs
			case "elif":
				ctrl.addToConditions(parseChunk(ps))
				if consumeLeader() != "then" {
					ps.error(errShouldBeThen)
				}
				ctrl.addToBodies(parseChunk(ps))
			case "else":
				ctrl.setElseBody(parseChunk(ps))
				if consumeLeader() != "fi" {
					ps.error(errShouldBeFi)
				}
				break Elifs
			default:
				ps.error(errShouldBeElifOrElseOrFi)
				break Elifs
			}
		}
	case "while":
		ctrl.Kind = WhileControl
		ctrl.setCondition(parseChunk(ps))
		doElseDone()
	case "for":
		ctrl.Kind = ForControl
		parseSpacesAndNewlines(ctrl, ps)
		ctrl.setIterator(parseIndexing(ps, false))
		parseSpacesAndNewlines(ctrl, ps)
		if ps.findPossibleLeader() == "in" {
			ps.advance(len("in"))
			addSep(ctrl, ps)
		} else {
			ps.error(errShouldBeIn)
		}
		ctrl.setArray(parseArray(ps, false))
		switch ps.peek() {
		case '\n', ';':
			ps.next()
			addSep(ctrl, ps)
		default:
			ps.error(errShouldBePipelineSep)
		}
		doElseDone()
	case "try":
		ctrl.Kind = TryControl
		ctrl.setBody(parseChunk(ps))
		if hasLeader(ps, "except") {
			consumeLeader()
			parseSpaces(ctrl, ps)
			if startsIndexing(ps.peek(), false) {
				ctrl.setExceptVar(parseIndexing(ps, false))
				parseSpaces(ctrl, ps)
			}
			switch ps.peek() {
			case '\n', ';':
				ps.next()
				addSep(ctrl, ps)
			default:
				ps.error(errShouldBePipelineSep)
			}
			ctrl.setExceptBody(parseChunk(ps))
		}
		if hasLeader(ps, "else") {
			consumeLeader()
			ctrl.setElseBody(parseChunk(ps))
		}
		if hasLeader(ps, "finally") {
			consumeLeader()
			ctrl.setFinallyBody(parseChunk(ps))
		}
		if consumeLeader() != "tried" {
			ps.error(errShouldBeTried)
		}
	case "begin":
		ctrl.Kind = BeginControl
		ctrl.setBody(parseChunk(ps))
		if consumeLeader() != "end" {
			ps.error(errShouldBeEnd)
		}
	default:
		ps.error(fmt.Errorf("unknown leader %q; parser bug", leader))
	}
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
	ern.setDest(parseCompound(ps, false))
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
		rn.setDest(dest)
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
		rn.SourceIsFd = true
	}
	rn.setSource(parseCompound(ps, false))
	if len(rn.Source.Indexings) == 0 {
		if rn.SourceIsFd {
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

func (cn *Compound) parse(ps *parser, head bool) {
	cn.tilde(ps)
	for startsIndexing(ps.peek(), head) {
		cn.addToIndexings(parseIndexing(ps, head))
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

func startsCompound(r rune, head bool) bool {
	return startsIndexing(r, head)
}

// Indexing = Primary { '[' Array ']' }
type Indexing struct {
	node
	Head     *Primary
	Indicies []*Array
}

func (in *Indexing) parse(ps *parser, head bool) {
	in.setHead(parsePrimary(ps, head))
	for parseSep(in, ps, '[') {
		if !startsArray(ps.peek()) {
			ps.error(errShouldBeArray)
		}

		ps.pushCutset()
		in.addToIndicies(parseArray(ps, false))
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
	// When non-empty, records the occurences of semicolons by the indices of
	// the compounds they appear before. For instance, [; ; a b; c d;] results
	// in Semicolons={0 0 2 4}.
	Semicolons []int
}

func (sn *Array) parse(ps *parser, allowSemicolon bool) {
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
		sn.addToCompounds(parseCompound(ps, false))
		parseSep()
	}
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func startsArray(r rune) bool {
	return isSpaceOrNewline(r) || startsIndexing(r, false)
}

// Primary is the smallest expression unit.
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
	PredCapture
	OutputCapture
	List
	Lambda
	Map
	Braced
)

func (pn *Primary) parse(ps *parser, head bool) {
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

func (pn *Primary) singleQuoted(ps *parser) {
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

func (pn *Primary) doubleQuoted(ps *parser) {
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

func (pn *Primary) variable(ps *parser) {
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

	pn.Type = PredCapture

	ps.pushCutset()
	pn.setChunk(parseChunk(ps))
	ps.popCutset()

	if !parseSep(pn, ps, ')') {
		ps.error(errShouldBeRParen)
	}
}

func (pn *Primary) outputCapture(ps *parser) {
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
	pn.setChunk(parseChunk(ps))
	ps.popCutset()

	if !parseSep(pn, ps, closer) {
		ps.error(shouldBeCloser)
	}
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
		case isSpace(r), r == ']', r == eof:
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
			ps.error(errShouldBeRBracket)
		}
	default:
		pn.setList(parseArray(ps, true))
		ps.popCutset()

		if !parseSep(pn, ps, ']') {
			ps.error(errShouldBeRBracket)
		}
		if parseSep(pn, ps, '{') {
			if len(pn.List.Semicolons) > 0 {
				ps.errorp(pn.List.Begin(), pn.List.End(), errArgListAllowNoSemicolon)
			}
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
		ps.error(errShouldBeRBrace)
	}
}

// Braced = '{' Compound { BracedSep Compounds } '}'
// BracedSep = { Space | '\n' } [ ',' ] { Space | '\n' }
func (pn *Primary) lbrace(ps *parser) {
	parseSep(pn, ps, '{')

	if r := ps.peek(); r == ';' || r == '\n' || isSpace(r) {
		pn.lambda(ps)
		return
	}

	pn.Type = Braced

	ps.pushCutset()
	defer ps.popCutset()

	// XXX: The compound can be empty, which allows us to parse {,foo}.
	// Allowing compounds to be empty can be fragile in other cases.
	ps.cut(',')
	pn.addToBraced(parseCompound(ps, false))
	ps.uncut(',')

	for isBracedSep(ps.peek()) {
		parseSpacesAndNewlines(pn, ps)
		// optional, so ignore the return value
		parseSep(pn, ps, ',')
		parseSpacesAndNewlines(pn, ps)

		ps.cut(',')
		pn.addToBraced(parseCompound(ps, false))
		ps.uncut(',')
	}
	if !parseSep(pn, ps, '}') {
		ps.error(errShouldBeBraceSepOrRBracket)
	}
}

func isBracedSep(r rune) bool {
	return r == ',' || isSpaceOrNewline(r)
}

func (pn *Primary) bareword(ps *parser, head bool) {
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

func (mpn *MapPair) parse(ps *parser) {
	parseSep(mpn, ps, '&')

	// Parse key part, cutting on '='.
	ps.cut('=')
	mpn.setKey(parseCompound(ps, false))
	if len(mpn.Key.Indexings) == 0 {
		ps.error(errShouldBeCompound)
	}
	ps.uncut('=')

	if parseSep(mpn, ps, '=') {
		// Parse value part.
		mpn.setValue(parseCompound(ps, false))
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

func parseSep(n Node, ps *parser, sep rune) bool {
	if ps.peek() == sep {
		ps.next()
		addSep(n, ps)
		return true
	}
	return false
}

func parseSpaces(n Node, ps *parser) {
	if !isSpace(ps.peek()) {
		return
	}
	ps.next()
	for isSpace(ps.peek()) {
		ps.next()
	}
	addSep(n, ps)
}

func parseSpacesAndNewlines(n Node, ps *parser) {
	// TODO parse comments here.
	if !isSpaceOrNewline(ps.peek()) {
		return
	}
	ps.next()
	for isSpaceOrNewline(ps.peek()) {
		ps.next()
	}
	addSep(n, ps)
}

func isSpaceOrNewline(r rune) bool {
	return isSpace(r) || r == '\n'
}

func addChild(p Node, ch Node) {
	p.n().children = append(p.n().children, ch)
	ch.n().parent = p
}
