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

// FormNode holds a form.
type FormNode struct {
	Pos
	Name   *ListNode // A Term
	Args   *ListNode // A ListNode of Term's
	Redirs []Redir
}

func newForm(pos Pos) *FormNode {
	return &FormNode{Pos: pos}
}

// A Term is represented by a ListNode of *FactorNode's.

// A Factor is any of:
// StringFactor: a `a` "a"
// VariableFactor: $a
// TableFactor: [a b c &k v]
// ClosureFactor: {|a| cmd}
// ListFactor: {a b c}
// CaptureFactor: (cmd)
type FactorNode struct {
	Pos
	Typ FactorType
	Node   Node
}

type FactorType int
const (
	StringFactor FactorType = iota
	VariableFactor
	TableFactor
	ClosureFactor
	ListFactor
	CaptureFactor
)

func newFactor(pos Pos) *FactorNode {
	return &FactorNode{Pos: pos}
}

type TablePair struct {
	Key   *ListNode
	Value *ListNode
}

type TableNode struct {
	Pos
	List []*ListNode
	Dict []*TablePair
}

func newTable(pos Pos) *TableNode {
	return &TableNode{Pos: pos}
}

func (tn *TableNode) appendToList(term *ListNode) {
	tn.List = append(tn.List, term)
}

func (tn *TableNode) appendToDict(key *ListNode, value *ListNode) {
	tn.Dict = append(tn.Dict, &TablePair{key, value})
}

type ClosureNode struct {
	Pos
	ArgNames *ListNode
	Chunk    *ListNode
}

func newClosure(pos Pos) *ClosureNode {
	return &ClosureNode{Pos: pos}
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
