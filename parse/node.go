// Derived from stdlib package text/template/parse.

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parse nodes.

package parse

// A Node is an element in the parse tree. The interface is trivial.
// The interface contains an unexported method so that only
// types local to this package can satisfy it.
type Node interface {
	Position() Pos // byte position of start of node in full original input string
	// Make sure only functions in this package can create Nodes.
	meisnode()
}

// Pos represents a byte position in the original input text from which
// this source was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

// meisnode keeps Node implementations local to the package.
// All implementations embed Pos, so this takes care of it.
func (Pos) meisnode() {
}

// Nodes.

// ListNode holds a sequence of nodes.
type ListNode struct {
	Pos
	Nodes []Node
}

func newList(pos Pos) *ListNode {
	return &ListNode{Pos: pos}
}

func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

// CommandNode holds a command, with terms and redirections.
type CommandNode struct {
	ListNode
	Redirs []Redir
}

func newCommand(pos Pos) *CommandNode {
	return &CommandNode{ListNode: *newList(pos)}
}

// A Term is represented by a ListNode of *FactorNode's.

// A Factor is any of the following prefixed by any number of dollars:
// *StringNode (scalar literal a `a` "a", variable evaluation $a $`a`)
//     $`a` and $"a" can be used to refer to variables of funny names.
// *ListNode (flat list (a b c), variable evaluation $(a^b))
//     In case of $(...), the flat list must evaluate to exactly one scalar.
//     Only useful for dynamic constructing of variable names, e.g. $($a$b).
// TODO TableNode (table literal [a b c], ??? $[a b c])
//     What should the syntax of the map part of a table be?
//     What should the semantics of $[a b c] be?
// TODO CommandNode (closure {cmd}, command output substituion ${cmd})

type FactorNode struct {
	Pos
	Dollar int
	Node Node
}

func newFactor(pos Pos) *FactorNode {
	return &FactorNode{Pos: pos}
}

// StringNode holds a string constant. The value has been "unquoted".
type StringNode struct {
	Pos
	Quoted string // The original text of the string, with quotes.
	Text   string // The string, after quote processing.
}

func newString(pos Pos, orig, text string) *StringNode {
	return &StringNode{Pos: pos, Quoted: orig, Text: text}
}
