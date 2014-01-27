// Derived from stdlib package text/template/parse.

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// das source lexer and parser.
package parse

import (
	"fmt"
	"github.com/xiaq/das/util"
	"os"
	"strconv"
	"strings"
)

type Parser struct {
	Name string // name of the script represented by the tree.
	Root Node   // top-level root of the tree.
	Ctx  Context
	text string // text parsed to create the script (or its parent)
	tab  bool
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

// Parse parses the script to construct a representation of the script for
// execution.
func (p *Parser) Parse(text string, tab bool) (tree *Parser, err error) {
	defer util.Recover(&err)
	defer p.stopParse()

	p.text = text
	p.tab = tab
	p.lex = Lex(p.Name, text)
	p.peekCount = 0

	p.Root = p.parse()

	p.stopParse()
	return p, nil
}

// parse parses a chunk and ensures there are no trailing tokens
func (p *Parser) parse() *ListNode {
	chunk := p.chunk()
	if token := p.peekNonSpace(); token.Typ != ItemEOF {
		p.unexpected(token, "end of script")
	}
	return chunk
}

// Chunk = [ [ space ] Pipeline { (";" | "\n") Pipeline } ]
func (p *Parser) chunk() *ListNode {
	chunk := newList(p.peek().Pos)
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
func (p *Parser) pipeline() *ListNode {
	pipe := newList(p.peek().Pos)
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
	fm.Name = p.term()
	fm.Args = p.termList()
loop:
	for {
		switch p.peekNonSpace().Typ {
		case ItemRedirLeader:
			fm.Redirs = append(fm.Redirs, p.redir())
		default:
			break loop
		}
	}
	return fm
}

// TermList = { [ space ] Term } [ space ]
func (p *Parser) termList() *ListNode {
	list := newList(p.peek().Pos)
loop:
	for {
		if startsFactor(p.peekNonSpace().Typ) {
			list.append(p.term())
		} else {
			break loop
		}
	}
	return list
}

// Term = Factor { Factor | [ space ] '^' Factor [ space ] } [ space ]
func (p *Parser) term() *ListNode {
	term := newList(p.peek().Pos)
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
		return "", fmt.Errorf("Bad token type (%s)", token.Typ)
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
	switch token := p.next(); token.Typ {
	case ItemDollar:
		token := p.next()
		if token.Typ != ItemBare {
			p.unexpected(token, "factor of variable")
		}
		fn.Typ = VariableFactor
		fn.Node = newString(token.Pos, token.Val, token.Val)
		return
	case ItemBare, ItemSingleQuoted, ItemDoubleQuoted:
		text, err := unquote(token)
		if err != nil {
			p.errorf(int(token.Pos), "%s", err)
		}
		if token.End&MayContinue != 0 {
			p.Ctx = NewArgContext(token.Val)
		} else {
			p.Ctx = nil
		}
		fn.Typ = StringFactor
		fn.Node = newString(token.Pos, token.Val, text)
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
			} else {
				return newCloseRedir(leader.Pos, fd)
			}
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
	return newFilenameRedir(leader.Pos, fd, flag, p.term())
}
