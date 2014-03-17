// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package parse parses elvish source. Derived from stdlib package
// text/template/parse.
package parse

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xiaq/elvish/util"
)

// Parser maintains the states during parsing.
type Parser struct {
	Name       string     // name of the script represented by the tree.
	completing bool       // Whether the parser is running in completing mode
	Root       *ChunkNode // top-level root of the tree.
	Ctx        *Context
	text       string // text parsed to create the script (or its parent)
	// Parsing only; cleared after parse.
	lex       *Lexer
	token     [3]Item // three-token lookahead for parser.
	peekCount int
}

// next returns the next token.
func (p *Parser) next() Item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.NextItem()
	}
	return p.token[p.peekCount]
}

// backup backs the input stream up one token.
func (p *Parser) backup() {
	p.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (p *Parser) backup2(t1 Item) {
	p.token[1] = t1
	p.peekCount = 2
}

// backup3 backs the input stream up three tokens
// The zeroth token is already there.
func (p *Parser) backup3(t2, t1 Item) { // Reverse order: we're pushing back.
	p.token[1] = t1
	p.token[2] = t2
	p.peekCount = 3
}

// peek returns but does not consume the next token.
func (p *Parser) peek() Item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.NextItem()
	return p.token[0]
}

// nextNonSpace returns the next non-space token.
func (p *Parser) nextNonSpace() (token Item) {
	for {
		token = p.next()
		if token.Typ != ItemSpace {
			break
		}
	}
	return token
}

// peekNonSpace returns but does not consume the next non-space token.
func (p *Parser) peekNonSpace() (token Item) {
	for {
		token = p.next()
		if token.Typ != ItemSpace {
			break
		}
	}
	p.backup()
	return token
}

// Parsing.

// NewParser allocates a new parse tree with the given name.
func NewParser(name string) *Parser {
	return &Parser{
		Name: name,
	}
}

// errorf formats the error and terminates processing.
func (p *Parser) errorf(pos int, format string, args ...interface{}) {
	p.Root = nil
	util.Panic(util.NewContextualError(p.Name, p.text, pos, format, args...))
}

// expect consumes the next token and guarantees it has the required type.
func (p *Parser) expect(expected ItemType, context string) Item {
	token := p.nextNonSpace()
	if token.Typ != expected {
		p.unexpected(token, context)
	}
	return token
}

// expectOneOf consumes the next token and guarantees it has one of the required types.
func (p *Parser) expectOneOf(expected1, expected2 ItemType, context string) Item {
	token := p.nextNonSpace()
	if token.Typ != expected1 && token.Typ != expected2 {
		p.unexpected(token, context)
	}
	return token
}

// unexpected complains about the token and terminates processing.
func (p *Parser) unexpected(token Item, context string) {
	p.errorf(int(token.Pos), "unexpected %s in %s", token, context)
}

// stopParse terminates parsing.
func (p *Parser) stopParse() {
	p.lex = nil
}

// A dummy struct used in foundCtx and recoverCtx.
type ctxFound struct {
}

// When the parser is running in completing mode, foundCtx causes Parse to
// terminate immediately by panicking. The panic will be stopped by recoverCtx.
// When the parser is not running in completing mode, it does nothing.
func (p *Parser) foundCtx() {
	if p.completing {
		panic(ctxFound{})
	}
}

// recoverCtx stops the panic and sets p.Ctx only when the panic is caused by
// raiseCtx.
func (p *Parser) recoverCtx() {
	r := recover()
	if r == nil {
		return
	}
	if _, ok := r.(ctxFound); !ok {
		panic(r)
	}
}

// Parse parses the script to construct a representation of the script for
// execution.
func (p *Parser) Parse(text string, completing bool) (err error) {
	defer util.Recover(&err)
	defer p.recoverCtx()
	defer p.stopParse()

	p.completing = completing
	p.text = text
	p.lex = Lex(p.Name, text)
	p.peekCount = 0

	p.Ctx = &Context{CommandContext, nil, newTermList(0), newTerm(0), &FactorNode{Node: newString(0, "", "")}}
	p.Root = p.parse()

	return nil
}

// Parse is a shorthand for constructing a Paser, call Parse and take out its
// Root.
func Parse(name, text string) (*ChunkNode, error) {
	p := NewParser(name)
	err := p.Parse(text, false)
	if err != nil {
		return nil, err
	}
	return p.Root, nil
}

// Complete is a shorthand for constructing a Paser, call Parse and take out
// its Ctx.
func Complete(name, text string) (*Context, error) {
	p := NewParser(name)
	err := p.Parse(text, true)
	if err != nil {
		return nil, err
	}
	return p.Ctx, nil
}

// parse parses a chunk and ensures there are no trailing tokens
func (p *Parser) parse() *ChunkNode {
	chunk := p.chunk()
	if token := p.peekNonSpace(); token.Typ != ItemEOF {
		p.unexpected(token, "end of script")
	}
	return chunk
}

// Chunk = [ [ space ] Pipeline { (";" | "\n") Pipeline } ]
func (p *Parser) chunk() *ChunkNode {
	chunk := newChunk(p.peek().Pos)
	if !startsFactor(p.peekNonSpace().Typ) {
		return chunk
	}

	for {
		// Skip leading whitespaces
		p.peekNonSpace()
		chunk.append(p.pipeline())

		if typ := p.peek().Typ; typ != ItemSemicolon && typ != ItemEndOfLine {
			break
		}

		p.next()
	}
	return chunk
}

// Pipeline = Form { "|" Form }
func (p *Parser) pipeline() *PipelineNode {
	pipe := newPipeline(p.peek().Pos)
	for {
		pipe.append(p.form())
		if p.peek().Typ != ItemPipe {
			break
		}
		p.next()
	}
	return pipe
}

// Form = TermList { [ space ] Redir } [ space ]
func (p *Parser) form() *FormNode {
	fm := newForm(p.peekNonSpace().Pos)
	p.Ctx.Typ = CommandContext
	fm.Command = p.term()
	p.Ctx.CommandTerm = fm.Command
	p.Ctx.Typ = ArgContext
	fm.Args = p.termList()
loop:
	for {
		switch p.peekNonSpace().Typ {
		case ItemRedirLeader:
			fm.Redirs = append(fm.Redirs, p.redir())
		case ItemStatusRedirLeader:
			fm.StatusRedir = p.statusRedir()
		default:
			break loop
		}
	}
	return fm
}

// TermList = { [ space ] Term } [ space ]
func (p *Parser) termList() *TermListNode {
	list := newTermList(p.peek().Pos)
	p.Ctx.PrevTerms = list
loop:
	for {
		switch t := p.peekNonSpace().Typ; {
		case startsFactor(t):
			list.append(p.term())
		case t == ItemEOF:
			pos := p.peek().Pos
			p.Ctx.PrevFactors = newTerm(pos)
			p.Ctx.ThisFactor = &FactorNode{pos, StringFactor, newString(pos, "", "")}
			p.foundCtx()
			fallthrough
		default:
			break loop
		}
	}
	return list
}

// Term = Factor { Factor | [ space ] '^' Factor [ space ] } [ space ]
func (p *Parser) term() *TermNode {
	term := newTerm(p.peek().Pos)
	p.Ctx.PrevFactors = term
	term.append(p.factor())
loop:
	for {
		if startsFactor(p.peek().Typ) {
			term.append(p.factor())
		} else if p.peekNonSpace().Typ == ItemCaret {
			p.next()
			p.peekNonSpace()
			term.append(p.factor())
		} else {
			break loop
		}
	}
	return term
}

func unquote(token Item) (string, error) {
	switch token.Typ {
	case ItemBare:
		return token.Val, nil
	case ItemSingleQuoted:
		return strings.Replace(token.Val[1:len(token.Val)-1], "``", "`", -1),
			nil
	case ItemDoubleQuoted:
		return strconv.Unquote(token.Val)
	default:
		return "", fmt.Errorf("bad token type (%s)", token.Typ)
	}
}

// startsFactor determines whether a token of type p can start a Factor.
// Frequently used for lookahead, since a Term or TermList always starts with
// a Factor.
func startsFactor(p ItemType) bool {
	switch p {
	case ItemBare, ItemSingleQuoted, ItemDoubleQuoted,
		ItemLParen, ItemLBracket, ItemLBrace,
		ItemDollar, ItemAmpersand:
		return true
	default:
		return false
	}
}

// Factor = '$' bare
//        = ( bare | single-quoted | double-quoted | Table )
//        = '{' TermList '}'
//        = Closure
//        = '(' Pipeline ')'
// Closure and flat list are distinguished by the first token after the
// opening brace. If startsFactor(token), it is considered a flat list.
// This implies that whitespaces after opening brace always introduce a
// closure: {echo} is a flat list, { echo } and {|| echo} are closures.
func (p *Parser) factor() (fn *FactorNode) {
	fn = newFactor(p.peek().Pos)
	p.Ctx.ThisFactor = fn
	switch token := p.next(); token.Typ {
	case ItemDollar:
		token := p.next()
		if token.Typ != ItemBare {
			p.unexpected(token, "factor of variable")
		}
		fn.Typ = VariableFactor
		fn.Node = newString(token.Pos, token.Val, token.Val)
		if p.peek().Typ == ItemEOF {
			p.foundCtx()
		}
		return
	case ItemBare, ItemSingleQuoted, ItemDoubleQuoted:
		text, err := unquote(token)
		if err != nil {
			// BUG(xiaq): When completing, unterminated quoted string results
			// in errors
			p.errorf(int(token.Pos), "%s", err)
		}
		fn.Typ = StringFactor
		fn.Node = newString(token.Pos, token.Val, text)
		if p.peek().Typ == ItemEOF {
			p.foundCtx()
		}
		return
	case ItemLBracket:
		fn.Typ = TableFactor
		fn.Node = p.table()
		return
	case ItemLBrace:
		if startsFactor(p.peek().Typ) {
			fn.Typ = ListFactor
			fn.Node = p.termList()
			if token := p.next(); token.Typ != ItemRBrace {
				p.unexpected(token, "factor of item list")
			}
		} else {
			fn.Typ = ClosureFactor
			fn.Node = p.closure()
		}
		return
	case ItemLParen:
		fn.Typ = CaptureFactor
		fn.Node = p.pipeline()
		if token := p.next(); token.Typ != ItemRParen {
			p.unexpected(token, "factor of pipeline capture")
		}
		return
	default:
		p.unexpected(token, "factor")
		return nil
	}
}

// closure parses a closure literal. The opening brace has been seen.
// Closure  = '{' [ space ] [ '|' TermList '|' [ space ] ] Chunk '}'
func (p *Parser) closure() (tn *ClosureNode) {
	tn = newClosure(p.peek().Pos)
	if p.peekNonSpace().Typ == ItemPipe {
		p.next()
		tn.ArgNames = p.termList()
		if token := p.nextNonSpace(); token.Typ != ItemPipe {
			p.unexpected(token, "argument list")
		}
	}
	tn.Chunk = p.chunk()
	if token := p.nextNonSpace(); token.Typ != ItemRBrace {
		p.unexpected(token, "end of closure")
	}
	return
}

// table parses a table literal. The opening bracket has been seen.
// Table = '[' { [ space ] ( '& 'Term [ space ] Term | Term ) [ space ] } ']'
func (p *Parser) table() (tn *TableNode) {
	tn = newTable(p.peek().Pos)

	for {
		token := p.nextNonSpace()
		// & is used both as key marker and closure leader.
		if token.Typ == ItemAmpersand && p.peek().Typ != ItemLBrace {
			// Key-value pair, add to dict.
			keyTerm := p.term()
			p.peekNonSpace()
			valueTerm := p.term()
			tn.appendToDict(keyTerm, valueTerm)
		} else if startsFactor(token.Typ) {
			// Single element, add to list.
			p.backup()
			tn.appendToList(p.term())
		} else if token.Typ == ItemRBracket {
			return
		} else {
			p.unexpected(token, "table literal")
		}
	}
}

// statusRedir parses a status redirection.
// StatusRedir = status-redir-leader [ space ] Term
// The term must consist of a single factor which in turn must be of type
// VariableFactor.
func (p *Parser) statusRedir() string {
	// Skip status-redir-leader
	p.next()

	if token := p.peekNonSpace(); token.Typ != ItemDollar {
		p.errorf(int(token.Pos), "expect variable")
	}
	term := p.term()
	if len(term.Nodes) == 1 {
		factor := term.Nodes[0]
		if factor.Typ == VariableFactor {
			return factor.Node.(*StringNode).Text
		}
	}
	p.errorf(int(term.Pos), "expect variable")
	return ""
}

// redir parses an IO redirection.
// Redir = redir-leader [ [ space ] Term ]
// NOTE The actual grammar is more complex than above, since 1) the inner
// structure of redir-leader is also parsed here, and 2) the Term is not truly
// optional, but sometimes required depending on the redir-leader.
func (p *Parser) redir() Redir {
	leader := p.next()

	// Partition the redirection leader into direction and qualifier parts.
	// For example, if leader.Val == ">>[1=2]", dir == ">>" and qual == "1=2".
	var dir, qual string

	if i := strings.IndexRune(leader.Val, '['); i != -1 {
		dir = leader.Val[:i]
		qual = leader.Val[i+1 : len(leader.Val)-1]
	} else {
		dir = leader.Val
	}

	// Determine the flag and default (new) fd from the direction.
	var (
		fd   uintptr
		flag int
	)

	switch dir {
	case "<":
		flag = os.O_RDONLY
		fd = 0
	case "<>":
		flag = os.O_RDWR | os.O_CREATE
		fd = 0
	case ">":
		flag = os.O_WRONLY | os.O_CREATE
		fd = 1
	case ">>":
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		fd = 1
	default:
		p.errorf(int(leader.Pos), "Unexpected redirection direction %q", dir)
	}

	if len(qual) > 0 {
		// Qualified redirection
		if i := strings.IndexRune(qual, '='); i != -1 {
			// FdRedir or CloseRedir
			lhs := qual[:i]
			rhs := qual[i+1:]
			if len(lhs) > 0 {
				var err error
				fd, err = Atou(lhs)
				if err != nil {
					// TODO identify precious position
					p.errorf(int(leader.Pos), "Invalid new fd in qualified redirection %q", lhs)
				}
			}
			if len(rhs) > 0 {
				oldfd, err := Atou(rhs)
				if err != nil {
					// TODO identify precious position
					p.errorf(int(leader.Pos), "Invalid old fd in qualified redirection %q", rhs)
				}
				return NewFdRedir(leader.Pos, fd, oldfd)
			}
			return newCloseRedir(leader.Pos, fd)
		} else {
			// FilenameRedir with fd altered
			var err error
			fd, err = Atou(qual)
			if err != nil {
				// TODO identify precious position
				p.errorf(int(leader.Pos), "Invalid new fd in qualified redirection %q", qual)
			}
		}
	}
	// FilenameRedir
	p.peekNonSpace()
	p.Ctx.Typ = RedirFilenameContext
	p.Ctx.PrevTerms = nil
	return newFilenameRedir(leader.Pos, fd, flag, p.term())
}
