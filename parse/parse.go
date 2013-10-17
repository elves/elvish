// Derived from stdlib package text/template/parse.

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// das source lexer and parser.
package parse

import (
	"os"
	"fmt"
	"strings"
	"strconv"
)

// Tree is the representation of a single parsed script.
type Tree struct {
	Name      string    // name of the script represented by the tree.
	Root      Node // top-level root of the tree.
	Ctx       Context
	text      string    // text parsed to create the script (or its parent)
	tab       bool
	// Parsing only; cleared after parse.
	lex       *Lexer
	token     [3]Item // three-token lookahead for parser.
	peekCount int
}

// Parse is shorthand for a New + *Tree.Parse combo.
func Parse(name, text string, tab bool) (t *Tree, err *Error) {
	return New(name).Parse(text, tab)
}

// next returns the next token.
func (t *Tree) next() Item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.NextItem()
	}
	return t.token[t.peekCount]
}

// backup backs the input stream up one token.
func (t *Tree) backup() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *Tree) backup2(t1 Item) {
	t.token[1] = t1
	t.peekCount = 2
}

// backup3 backs the input stream up three tokens
// The zeroth token is already there.
func (t *Tree) backup3(t2, t1 Item) { // Reverse order: we're pushing back.
	t.token[1] = t1
	t.token[2] = t2
	t.peekCount = 3
}

// peek returns but does not consume the next token.
func (t *Tree) peek() Item {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.NextItem()
	return t.token[0]
}

// nextNonSpace returns the next non-space token.
func (t *Tree) nextNonSpace() (token Item) {
	for {
		token = t.next()
		if token.Typ != ItemSpace {
			break
		}
	}
	return token
}

// peekNonSpace returns but does not consume the next non-space token.
func (t *Tree) peekNonSpace() (token Item) {
	for {
		token = t.next()
		if token.Typ != ItemSpace {
			break
		}
	}
	t.backup()
	return token
}

// Parsing.

// New allocates a new parse tree with the given name.
func New(name string) *Tree {
	return &Tree{
		Name:  name,
	}
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(pos int, format string, args ...interface{}) {
	t.Root = nil
	lineno, colno, line := findContext(t.text, pos)
	panic(&Error{t.Name, lineno, colno, line, fmt.Sprintf(format, args...)})
}

// expect consumes the next token and guarantees it has the required type.
func (t *Tree) expect(expected ItemType, context string) Item {
	token := t.nextNonSpace()
	if token.Typ != expected {
		t.unexpected(token, context)
	}
	return token
}

// expectOneOf consumes the next token and guarantees it has one of the required types.
func (t *Tree) expectOneOf(expected1, expected2 ItemType, context string) Item {
	token := t.nextNonSpace()
	if token.Typ != expected1 && token.Typ != expected2 {
		t.unexpected(token, context)
	}
	return token
}

// unexpected complains about the token and terminates processing.
func (t *Tree) unexpected(token Item, context string) {
	t.errorf(int(token.Pos), "unexpected %s in %s", token, context)
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (t *Tree) recover(errp **Error) {
	e := recover()
	if e == nil {
		return
	}
	if _, ok := e.(*Error); !ok {
		panic(e)
	}
	if t != nil {
		t.stopParse()
	}
	*errp = e.(*Error)
}

// stopParse terminates parsing.
func (t *Tree) stopParse() {
	t.lex = nil
	if t.tab {
		t.Root = nil
	} else {
		t.Ctx = nil
	}
}

// Parse parses the script to construct a representation of the script for
// execution.
func (t *Tree) Parse(text string, tab bool) (tree *Tree, err *Error) {
	defer t.recover(&err)

	t.text = text
	t.tab = tab
	t.lex = Lex(t.Name, text)
	t.peekCount = 0

	// TODO This now only parses a pipeline.
	t.Root = t.pipeline()

	t.stopParse()
	return t, nil
}

// Pipeline = [ Command { "|" Command } ]
func (t *Tree) pipeline() *ListNode {
	pipe := newList(t.peek().Pos)
	if t.peekNonSpace().Typ == ItemEOF {
		return pipe
	}
loop:
	for {
		n := t.command()
		pipe.append(n)

		switch token := t.next(); token.Typ {
		case ItemPipe:
			continue loop
		case ItemEndOfLine, ItemEOF:
			break loop
		default:
			t.unexpected(token, "end of pipeline")
		}
	}
	return pipe
}

// command parses a command.
// Command = TermList { [ space ] Redir }
func (t *Tree) command() *CommandNode {
	cmd := newCommand(t.peek().Pos)
	cmd.ListNode = *t.termList()
loop:
	for {
		switch t.peekNonSpace().Typ {
		case ItemRedirLeader:
			cmd.Redirs = append(cmd.Redirs, t.redir())
		default:
			break loop
		}
	}
	return cmd
}

// TermList = [ space ] Term { [ space ] Term } [ space ]
func (t *Tree) termList() *ListNode {
	list := newList(t.peek().Pos)
	list.append(t.term())
loop:
	for {
		if startsFactor(t.peekNonSpace().Typ) {
			list.append(t.term())
		} else {
			break loop
		}
	}
	return list
}

// Term = Factor { Factor }
func (t *Tree) term() *ListNode {
	term := newList(t.peek().Pos)
	term.append(t.factor())
loop:
	for {
		if startsFactor(t.peek().Typ) {
			term.append(t.factor())
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
		return "", fmt.Errorf("Can't unquote token: %v", token)
	}
}

// startsFactor determines whether a token of type t can start a Factor.
// Frequently used for lookahead, since a Term or TermList always starts with
// a Factor.
func startsFactor(t ItemType) bool {
	switch t {
	case ItemBare, ItemSingleQuoted, ItemDoubleQuoted,
			ItemLParen, ItemLBracket,
			ItemDollar:
		return true
	default:
		return false
	}
}

// Factor = '$' Factor
//        = ( bare | single-quoted | double-quoted | Table )
//        = ( '(' TermList ')' )
func (t *Tree) factor() (fn *FactorNode) {
	fn = newFactor(t.peek().Pos)
	for t.peek().Typ == ItemDollar {
		t.next()
		fn.Dollar++
	}
	switch token := t.next(); token.Typ {
	case ItemBare, ItemSingleQuoted, ItemDoubleQuoted:
		text, err := unquote(token)
		if err != nil {
			t.errorf(int(token.Pos), "%s", err)
		}
		if token.End & MayContinue != 0 {
			t.Ctx = NewArgContext(token.Val)
		} else {
			t.Ctx = nil
		}
		fn.Node = newString(token.Pos, token.Val, text)
		return
	case ItemLParen:
		fn.Node = t.termList()
		if token := t.next(); token.Typ != ItemRParen {
			t.unexpected(token, "factor of item list")
		}
		return
	case ItemLBracket:
		fn.Node = t.table()
		return
	default:
		t.unexpected(token, "factor")
		return nil
	}
}

// table parses a table literal. The opening bracket has been seen.
// Table = '[' { [ space ] ( Term [ space ] '=' [ space ] Term | Term ) [ space ] } ']'
// NOTE The '=' is actually special-cased Term.
func (t *Tree) table() (tn *TableNode) {
	tn = newTable(t.peek().Pos)

	for {
		token := t.nextNonSpace()
		if startsFactor(token.Typ) {
			t.backup()
			term := t.term()

			next := t.peekNonSpace()
			if next.Typ == ItemBare && next.Val == "=" {
				t.next()
				// New element of dict part. Skip spaces and find value term.
				t.peekNonSpace()
				valueTerm := t.term()
				tn.appendToDict(term, valueTerm)
			} else {
				// New element of list part.
				tn.appendToList(term)
			}
		} else if token.Typ == ItemRBracket {
			return
		} else {
			t.unexpected(token, "table literal")
		}
	}
}

// redir parses an IO redirection.
// Redir = redir-leader [ [ space ] Term ]
// NOTE The actual grammar is more complex than above, since 1) the inner
// structure of redir-leader is also parsed here, and 2) the Term is not truly
// optional, but sometimes required depending on the redir-leader.
func (t *Tree) redir() Redir {
	leader := t.next()

	// Partition the redirection leader into direction and qualifier parts.
	// For example, if leader.Val == ">>[1=2]", dir == ">>" and qual == "1=2".
	var dir, qual string

	if i := strings.IndexRune(leader.Val, '['); i != -1 {
		dir = leader.Val[:i]
		qual = leader.Val[i+1:len(leader.Val)-1]
	} else {
		dir = leader.Val
	}

	// Determine the flag and default (new) fd from the direction.
	var (
		fd uintptr
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
		t.errorf(int(leader.Pos), "Unexpected redirection direction %q", dir)
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
					t.errorf(int(leader.Pos), "Invalid new fd in qualified redirection %q", lhs)
				}
			}
			if len(rhs) > 0 {
				oldfd, err := Atou(rhs)
				if err != nil {
					// TODO identify precious position
					t.errorf(int(leader.Pos), "Invalid old fd in qualified redirection %q", rhs)
				}
				return NewFdRedir(fd, oldfd)
			} else {
				return newCloseRedir(fd)
			}
		} else {
			// FilenameRedir with fd altered
			var err error
			fd, err = Atou(qual)
			if err != nil {
				// TODO identify precious position
				t.errorf(int(leader.Pos), "Invalid new fd in qualified redirection %q", qual)
			}
		}
	}
	// FilenameRedir
	t.peekNonSpace()
	return newFilenameRedir(fd, flag, t.term())
}
