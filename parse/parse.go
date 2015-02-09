// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package parse parses elvish source using a recursive descent parser.
//
// Some parser utilities are derived from stdlib package text/template/parse.
package parse

// The Elvish grammar is likely LL(2) (a more careful examinization may prove
// otherwise). The non-terminals are documented in corresponding functions
// using a liberal EBNF variant. In particular, [ option ] and { kleene star }
// are used.
//
// Whitespaces are often significant in Elvish and are never omitted in the
// grammar.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/errutil"
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
	errutil.Throw(errutil.NewContextualError(p.Name, "parsing error", p.text, pos, format, args...))
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
	p.errorf(int(token.Pos), "unexpected %s (%s) in %s", token, token.Typ, context)
}

// stopParse terminates parsing.
func (p *Parser) stopParse() {
	p.lex = nil
}

// A dummy struct used in foundCtx and recoverCtx.
type ctxFound struct {
}

// When the parser is running in completing mode and the next token is EOF,
// foundCtx returns true and sets p.Ctx.Typ to typ if it's not already set.
// Otherwise foundCtx returns false.
func (p *Parser) foundCtx(typ ContextType) bool {
	if p.completing && p.peek().Typ == ItemEOF {
		if p.Ctx.Typ == UnknownContext {
			p.Ctx.Typ = typ
		}
		return true
	}
	return false
}

// Parse parses the script to construct a representation of the script for
// execution.
func (p *Parser) Parse(text string, completing bool) (err error) {
	defer errutil.Catch(&err)
	defer p.stopParse()

	p.completing = completing
	p.text = text
	p.lex = Lex(p.Name, text)
	p.peekCount = 0

	p.Ctx = &Context{}
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

// chunk parses a chunk. The name is borrowed from Lua.
//
// Chunk = [ [ space ] Pipeline { (";" | "\n") Pipeline } ]
func (p *Parser) chunk() *ChunkNode {
	chunk := newChunk(p.peek().Pos)

	for {
		// Skip leading whitespaces
		token := p.peekNonSpace()
		if p.foundCtx(CommandContext) {
			break
		}
		if token.Typ == ItemSemicolon || token.Typ == ItemEndOfLine {
			// Skip empty pipeline
			p.next()
			continue
		} else if startsPrimary(token.Typ) {
			chunk.append(p.pipeline())
		} else {
			break
		}
	}
	return chunk
}

// pipeline parses a pipeline. The pipeline may not be empty.
//
// Pipeline = Form { '|' Form }
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

// form parses a form. The name is borrowed from Lisp.
//
// Form = Spaced { [ space ] Redir } [ space ]
func (p *Parser) form() *FormNode {
	fm := newForm(p.peekNonSpace().Pos)
	p.Ctx.Form = fm
	fm.Command = p.compound(CommandContext)
	fm.Args = p.spaced()
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

// spaced parses a spaced expression. Many languages separate expressions with
// commas or other punctuations, but shells have a traditional of separating
// them with simple whitespaces.
//
// Spaced = { [ space ] Compound } [ space ]
func (p *Parser) spaced() *SpacedNode {
	// Skip leading spaces
	p.peekNonSpace()
	list := newSpaced(p.peek().Pos)
loop:
	for {
		// Skip space tokens
		p.peekNonSpace()
		if p.foundCtx(NewArgContext) {
			break loop
		}

		if startsPrimary(p.peek().Typ) {
			list.append(p.compound(ArgContext))
		} else {
			break loop
		}
	}
	return list
}

// compound parses a compound expression. The name is borrowed from
// linguistics, where a compound word is roughly some words that run together.
//
// Compound = [ sigil ] Subscript { Subscript } [ space ]
func (p *Parser) compound(ct ContextType) *CompoundNode {
	compound := newCompound(p.peek().Pos, NoSigil)
	if p.peek().Typ == ItemSigil {
		token := p.next()
		compound.Sigil, _ = utf8.DecodeRuneInString(token.Val)
	}
	compound.append(p.subscript())
	for startsPrimary(p.peek().Typ) {
		compound.append(p.subscript())
	}
	p.foundCtx(ct)
	return compound
}

// subscript parses a subscript expression. The subscript part is actually
// optional.
//
// Subscript = Primary [ '[' Compound ']' ]
func (p *Parser) subscript() *SubscriptNode {
	sub := &SubscriptNode{Pos: p.peek().Pos}
	sub.Left = p.primary()
	if p.peek().Typ == ItemLBracket {
		p.next()
		sub.Right = p.compound(SubscriptContext)
		// TODO completion aware
		p.expect(ItemRBracket, "subscript")
	}
	return sub
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

// startsPrimary determines whether a token of type p can start a Primary.
// Frequently used for lookahead, since a Subscript, Compound or Spaced also
// always starts with a Primary.
//
// XXX(xiaq): The case with ItemSigil can be problematic. Compound or Spaced
// may start with it, but it's illegal in Primary or Subscript.
func startsPrimary(p ItemType) bool {
	switch p {
	case ItemBare, ItemSingleQuoted, ItemDoubleQuoted,
		ItemLParen, ItemQuestionLParen, ItemLBracket, ItemLBrace,
		ItemDollar, ItemAmpersand, ItemSigil:
		return true
	default:
		return false
	}
}

// primary parses a primary expression.
//
// Primary = '$' bare
//         = ( bare | single-quoted | double-quoted | Table )
//         = '{' Spaced '}'
//         = Closure
//         = '(' Pipeline ')'
//
// Closure and flat list are distinguished by the first token after the
// opening brace. If startsPrimary(token), it is considered a flat list.
// This implies that whitespaces after opening brace always introduce a
// closure: {echo} is a flat list, { echo } and {|| echo} are closures.
func (p *Parser) primary() (fn *PrimaryNode) {
	fn = newPrimary(p.peek().Pos)
	switch token := p.next(); token.Typ {
	case ItemDollar:
		token := p.next()
		if token.Typ != ItemBare {
			p.unexpected(token, "primary expression of variable")
		}
		fn.Typ = VariablePrimary
		fn.Node = newString(token.Pos, token.Val, token.Val)
		return
	case ItemBare, ItemSingleQuoted, ItemDoubleQuoted:
		text, err := unquote(token)
		if err != nil {
			// BUG(xiaq): When completing, unterminated quoted string results
			// in errors
			p.errorf(int(token.Pos), "%s", err)
		}
		fn.Typ = StringPrimary
		fn.Node = newString(token.Pos, token.Val, text)
		return
	case ItemLBracket:
		p.backup()
		fn.Typ = TablePrimary
		fn.Node = p.table()
		return
	case ItemLBrace:
		if startsPrimary(p.peek().Typ) {
			fn.Typ = ListPrimary
			fn.Node = p.spaced()
			if token := p.next(); token.Typ != ItemRBrace {
				p.unexpected(token, "primary expression of item list")
			}
		} else {
			fn.Typ = ClosurePrimary
			fn.Node = p.closure()
		}
		return
	case ItemLParen, ItemQuestionLParen:
		if token.Typ == ItemLParen {
			fn.Typ = ChanCapturePrimary
		} else {
			fn.Typ = StatusCapturePrimary
		}
		fn.Node = p.pipeline()
		if token := p.next(); token.Typ != ItemRParen {
			p.unexpected(token, "primary expression of pipeline capture")
		}
		return
	default:
		p.unexpected(token, "primary expression")
		return nil
	}
}

// closure parses a closure literal. The opening brace has already been seen.
//
// Closure  = '{' [ space ] [ '|' Spaced '|' [ space ] ] Chunk '}'
func (p *Parser) closure() (tn *ClosureNode) {
	tn = newClosure(p.peek().Pos)
	if p.peekNonSpace().Typ == ItemPipe {
		p.next()
		tn.ArgNames = p.spaced()
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

// table parses a table literal.
//
// Table = '[' { [ space ] TableElement  [ space ] } ']'
//
// TableElement = Compound
//              = '&' Compound [ space ] Compound
func (p *Parser) table() (tn *TableNode) {
	tn = newTable(p.next().Pos)

	for {
		token := p.nextNonSpace()
		// & is used both as key marker and closure leader.
		if token.Typ == ItemAmpersand && p.peek().Typ != ItemLBrace {
			// Key-value pair, add to dict.
			keyCompound := p.compound(TableKeyContext)
			p.peekNonSpace()
			valCompound := p.compound(TableValueContext)
			tn.appendToDict(keyCompound, valCompound)
		} else if startsPrimary(token.Typ) {
			// Single element, add to list.
			p.backup()
			tn.appendToList(p.compound(TableElemContext))
		} else if token.Typ == ItemRBracket {
			return
		} else {
			p.unexpected(token, "table literal")
		}
	}
}

// statusRedir parses a status redirection.
//
// StatusRedir = status-redir-leader [ space ] Compound
//
// The compound expression must consist of a single primary expression which
// in turn must be of type VariablePrimary.
func (p *Parser) statusRedir() string {
	// Skip status-redir-leader
	p.next()

	if token := p.peekNonSpace(); token.Typ != ItemDollar {
		p.errorf(int(token.Pos), "expect variable")
	}
	compound := p.compound(StatusRedirContext)
	if len(compound.Nodes) == 1 && compound.Nodes[0].Right == nil {
		primary := compound.Nodes[0].Left
		if primary.Typ == VariablePrimary {
			return primary.Node.(*StringNode).Text
		}
	}
	p.errorf(int(compound.Pos), "expect variable")
	return ""
}

// redir parses an IO redirection.
//
// Redir = redir-leader [ [ space ] Compound ]
//
// NOTE: The actual grammar is more complex than above, since
// 1) the inner structure of redir-leader is also parsed here
// 2) the Compound is not truly optional, but sometimes required depending on
// the redir-leader.
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
	return newFilenameRedir(leader.Pos, fd, flag, p.compound(RedirFilenameContext))
}
