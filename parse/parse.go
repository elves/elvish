// Derived from stdlib package text/template/parse.

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// das source lexer and parser.
package parse

import (
	"os"
	"fmt"
	"runtime"
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
func Parse(name, text string, tab bool) (t *Tree, err error) {
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

// ErrorContext returns a textual representation of the location of the node in the input text.
func (t *Tree) ErrorContext(n Node) (location, context string) {
	pos := int(n.Position())
	text := t.text[:pos]
	byteNum := strings.LastIndex(text, "\n")
	if byteNum == -1 {
		byteNum = pos // On first line.
	} else {
		byteNum++ // After the newline.
		byteNum = pos - byteNum
	}
	lineNum := 1 + strings.Count(text, "\n")
	context = n.String()
	if len(context) > 20 {
		context = fmt.Sprintf("%.20s...", context)
	}
	return fmt.Sprintf("%s:%d:%d", t.Name, lineNum, byteNum), context
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	format = fmt.Sprintf("das: %s:%d: %s", t.Name, t.lex.lineNumber(), format)
	panic(fmt.Errorf(format, args...))
}

// error terminates processing.
func (t *Tree) error(err error) {
	t.errorf("%s", err)
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
	t.errorf("unexpected %s in %s", token, context)
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if t != nil {
			t.stopParse()
		}
		*errp = e.(error)
	}
	return
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
func (t *Tree) Parse(text string, tab bool) (tree *Tree, err error) {
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

func (t *Tree) pipeline() *ListNode {
	pipe := newList(t.peek().Pos)
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
func (t *Tree) command() *CommandNode {
	cmd := newCommand(t.peek().Pos)
loop:
	for {
		switch t.peekNonSpace().Typ {
		case ItemBare, ItemSingleQuoted, ItemDoubleQuoted:
			cmd.append(t.term())
		case ItemRedirLeader:
			cmd.Redirs = append(cmd.Redirs, t.redir())
		default:
			break loop
		}
	}
	return cmd
}

func unquote(token Item) (string, error) {
	switch token.Typ {
	case ItemBare:
		return token.Val, nil
	case ItemSingleQuoted:
		return strings.Replace(token.Val[1:len(token.Val)-1], "''", "'", -1),
		       nil
	case ItemDoubleQuoted:
		return strconv.Unquote(token.Val)
	default:
		return "", fmt.Errorf("Can't unquote token: %v", token)
	}
}

func (t *Tree) term() Node {
	switch token := t.next(); token.Typ {
	case ItemBare, ItemSingleQuoted, ItemDoubleQuoted:
		text, err := unquote(token)
		if err != nil {
			t.error(err)
		}
		if token.End & MayContinue != 0 {
			t.Ctx = NewArgContext(token.Val)
		} else {
			t.Ctx = nil
		}
		return newString(token.Pos, token.Val, text)
	default:
		t.unexpected(token, "term")
		return nil
	}
}

// redir parses an IO redirection.
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
	var fd, flag int

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
		t.errorf("Unexpected redirection direction %q", dir)
	}

	if len(qual) > 0 {
		// Qualified redirection
		if i := strings.IndexRune(qual, '='); i != -1 {
			// FdRedir or CloseRedir
			lhs := qual[:i]
			rhs := qual[i+1:]
			if len(lhs) > 0 {
				var err error
				fd, err = strconv.Atoi(lhs)
				if err != nil {
					t.errorf("Invalid new fd in qualified redirection %q", lhs)
				}
			}
			if len(rhs) > 0 {
				oldfd, err := strconv.Atoi(rhs)
				if err != nil {
					t.errorf("Invalid old fd in qualified redirection %q", rhs)
				}
				return newFdRedir(fd, oldfd)
			} else {
				return newCloseRedir(fd)
			}
		} else {
			// FilenameRedir with fd altered
			var err error
			fd, err = strconv.Atoi(qual)
			if err != nil {
				t.errorf("Invalid new fd in qualified redirection %q", qual)
			}
		}
	}
	// FilenameRedir
	t.peekNonSpace()
	return newFilenameRedir(fd, flag, t.term())
}
