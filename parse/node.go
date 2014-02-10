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
	isNode()       // tag function
}

// Pos represents a byte position in the original input text from which
// this source was parsed.
type Pos int

// Position returns p itself.
func (p Pos) Position() Pos {
	return p
}

// Nodes.

// ListNode holds a sequence of nodes.
type ListNode struct {
	Pos
	Nodes []Node
}

func newList(pos Pos, nodes ...Node) *ListNode {
	return &ListNode{Pos: pos, Nodes: nodes}
}

func (l *ListNode) isNode() {}

func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

// FormNode holds a form.
type FormNode struct {
	Pos
	Command *TermNode
	Args    *TermListNode
	Redirs  []Redir
}

func newForm(pos Pos) *FormNode {
	return &FormNode{Pos: pos}
}

func (fn *FormNode) isNode() {}

// TermNode is a ListNode of FactorNode's.
type TermNode struct {
	ListNode
}

func newTerm(pos Pos, nodes ...Node) *TermNode {
	return &TermNode{ListNode{Pos: pos, Nodes: nodes}}
}

// TermListNode is a ListNode of TermNode's.
type TermListNode struct {
	ListNode
}

func newTermList(pos Pos, nodes ...Node) *TermListNode {
	return &TermListNode{ListNode{Pos: pos, Nodes: nodes}}
}

// FactorNode represents a factor.
type FactorNode struct {
	Pos
	Typ  FactorType
	Node Node
}

// FactorType determines the type of a FactorNode.
type FactorType int

// FactorType constants.
const (
	StringFactor   FactorType = iota // string literal: a `a` a
	VariableFactor                   // variable: $a
	TableFactor                      // table: [a b c &k v]
	ClosureFactor                    // closure: {|a| cmd}
	ListFactor                       // list: {a b c}
	CaptureFactor                    // pipeline capture: (cmd1|cmd2)
)

func newFactor(pos Pos) *FactorNode {
	return &FactorNode{Pos: pos}
}

func (fn *FactorNode) isNode() {}

// TablePair represents a key/value pair in table literal.
type TablePair struct {
	Key   *TermNode
	Value *TermNode
}

// TableNode holds a table literal.
type TableNode struct {
	Pos
	List []*TermNode
	Dict []*TablePair
}

func newTable(pos Pos) *TableNode {
	return &TableNode{Pos: pos}
}

func (tn *TableNode) isNode() {}

func (tn *TableNode) appendToList(term *TermNode) {
	tn.List = append(tn.List, term)
}

func (tn *TableNode) appendToDict(key *TermNode, value *TermNode) {
	tn.Dict = append(tn.Dict, &TablePair{key, value})
}

// ClosureNode holds a closure literal.
type ClosureNode struct {
	Pos
	ArgNames *TermListNode
	Chunk    *ListNode
}

func newClosure(pos Pos) *ClosureNode {
	return &ClosureNode{Pos: pos}
}

func (cn *ClosureNode) isNode() {}

// StringNode holds a string literal. The value has been "unquoted".
type StringNode struct {
	Pos
	Quoted string // The original text of the string, with quotes.
	Text   string // The string, after quote processing.
}

func newString(pos Pos, orig, text string) *StringNode {
	return &StringNode{Pos: pos, Quoted: orig, Text: text}
}

func (sn *StringNode) isNode() {}
