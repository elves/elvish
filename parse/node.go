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

// ChunkNode is a list of FormNode's.
type ChunkNode struct {
	Pos
	Nodes []*PipelineNode
}

func newChunk(pos Pos, nodes ...*PipelineNode) *ChunkNode {
	return &ChunkNode{pos, nodes}
}

func (l *ChunkNode) isNode() {}

func (tn *ChunkNode) append(n *PipelineNode) {
	tn.Nodes = append(tn.Nodes, n)
}

// PipelineNode is a list of FormNode's.
type PipelineNode struct {
	Pos
	Nodes []*FormNode
}

func newPipeline(pos Pos, nodes ...*FormNode) *PipelineNode {
	return &PipelineNode{pos, nodes}
}

func (l *PipelineNode) isNode() {}

func (tn *PipelineNode) append(n *FormNode) {
	tn.Nodes = append(tn.Nodes, n)
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

// TermNode is a list of FactorNode's.
type TermNode struct {
	Pos
	Nodes []*FactorNode
}

func newTerm(pos Pos, nodes ...*FactorNode) *TermNode {
	return &TermNode{pos, nodes}
}

func (l *TermNode) isNode() {}

func (tn *TermNode) append(n *FactorNode) {
	tn.Nodes = append(tn.Nodes, n)
}

// TermListNode is a list of TermNode's.
type TermListNode struct {
	Pos
	Nodes []*TermNode
}

func newTermList(pos Pos, nodes ...*TermNode) *TermListNode {
	return &TermListNode{pos, nodes}
}

func (l *TermListNode) isNode() {}

func (tn *TermListNode) append(n *TermNode) {
	tn.Nodes = append(tn.Nodes, n)
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
	Chunk    *ChunkNode
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
