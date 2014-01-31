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
	Position() Pos        // byte position of start of node in full original input string
	isomorph(n Node) bool // compares two Nodes, ignoring Pos
}

// Pos represents a byte position in the original input text from which
// this source was parsed.
type Pos int

// Position returns p itself.
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

func newList(pos Pos, nodes ...Node) *ListNode {
	return &ListNode{Pos: pos, Nodes: nodes}
}

func (l *ListNode) isomorph(n Node) bool {
	if l2, ok := n.(*ListNode); ok {
		if len(l.Nodes) == len(l2.Nodes) {
			for i := range l.Nodes {
				if !l.Nodes[i].isomorph(l2.Nodes[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

// FormNode holds a form.
type FormNode struct {
	Pos
	Command *ListNode // A Term
	Args    *ListNode // A ListNode of Term's
	Redirs  []Redir
}

func newForm(pos Pos) *FormNode {
	return &FormNode{Pos: pos}
}

func (fn *FormNode) isomorph(n Node) bool {
	if fn2, ok := n.(*FormNode); ok {
		if fn.Command.isomorph(fn2.Command) && fn.Args.isomorph(fn2.Args) && len(fn.Redirs) == len(fn2.Redirs) {
			for i := range fn.Redirs {
				if !fn.Redirs[i].isomorph(fn2.Redirs[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// A Term is represented by a ListNode of *FactorNode's.

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

func (fn *FactorNode) isomorph(n Node) bool {
	if fn2, ok := n.(*FactorNode); ok {
		return fn.Typ == fn2.Typ && fn.Node.isomorph(fn2.Node)
	}
	return false
}

// TablePair represents a key/value pair in table literal.
type TablePair struct {
	Key   *ListNode
	Value *ListNode
}

// TableNode holds a table literal.
type TableNode struct {
	Pos
	List []*ListNode
	Dict []*TablePair
}

func newTable(pos Pos) *TableNode {
	return &TableNode{Pos: pos}
}

func (tn *TableNode) isomorph(n Node) bool {
	if tn2, ok := n.(*TableNode); ok {
		if len(tn.List) == len(tn2.List) && len(tn.Dict) == len(tn2.Dict) {
			for i := range tn.List {
				if !tn.List[i].isomorph(tn2.List[i]) {
					return false
				}
			}
			for i := range tn.Dict {
				if !tn.Dict[i].Key.isomorph(tn2.Dict[i].Key) {
					return false
				}
				if !tn.Dict[i].Value.isomorph(tn2.Dict[i].Value) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (tn *TableNode) appendToList(term *ListNode) {
	tn.List = append(tn.List, term)
}

func (tn *TableNode) appendToDict(key *ListNode, value *ListNode) {
	tn.Dict = append(tn.Dict, &TablePair{key, value})
}

// ClosureNode holds a closure literal.
type ClosureNode struct {
	Pos
	ArgNames *ListNode
	Chunk    *ListNode
}

func newClosure(pos Pos) *ClosureNode {
	return &ClosureNode{Pos: pos}
}

func (cn *ClosureNode) isomorph(n Node) bool {
	if cn2, ok := n.(*ClosureNode); ok {
		return cn.ArgNames.isomorph(cn2.ArgNames) && cn.Chunk.isomorph(cn2.Chunk)
	}
	return false
}

// StringNode holds a string literal. The value has been "unquoted".
type StringNode struct {
	Pos
	Quoted string // The original text of the string, with quotes.
	Text   string // The string, after quote processing.
}

func newString(pos Pos, orig, text string) *StringNode {
	return &StringNode{Pos: pos, Quoted: orig, Text: text}
}

func (sn *StringNode) isomorph(n Node) bool {
	if sn2, ok := n.(*StringNode); ok {
		return sn.Quoted == sn2.Quoted && sn.Text == sn2.Text
	}
	return false
}
