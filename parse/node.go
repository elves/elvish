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
	return &PipelineNode{Pos: pos, Nodes: nodes}
}

func (l *PipelineNode) isNode() {}

func (tn *PipelineNode) append(n *FormNode) {
	tn.Nodes = append(tn.Nodes, n)
}

// FormNode holds a form.
type FormNode struct {
	Pos
	Command     *TermNode
	Args        *TermListNode
	Redirs      []Redir
	StatusRedir string
}

func newForm(pos Pos) *FormNode {
	return &FormNode{Pos: pos}
}

func (fn *FormNode) isNode() {}

// TermNode is a list of PrimaryNode's.
type TermNode struct {
	Pos
	Nodes []*PrimaryNode
}

func newTerm(pos Pos, nodes ...*PrimaryNode) *TermNode {
	return &TermNode{pos, nodes}
}

func (l *TermNode) isNode() {}

func (tn *TermNode) append(n *PrimaryNode) {
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

// PrimaryNode represents a primary expression.
type PrimaryNode struct {
	Pos
	Typ  PrimaryType
	Node Node
}

// PrimaryType determines the type of a PrimaryNode.
type PrimaryType int

// PrimaryType constants.
const (
	StringPrimary        PrimaryType = iota // string literal: a `a` a
	VariablePrimary                         // variable: $a
	TablePrimary                            // table: [a b c &k v]
	ClosurePrimary                          // closure: {|a| cmd}
	ListPrimary                             // list: {a b c}
	OutputCapturePrimary                    // output capture: (cmd1|cmd2)
	StatusCapturePrimary                    // status capture: ?(cmd1|cmd2)
)

func newPrimary(pos Pos) *PrimaryNode {
	return &PrimaryNode{Pos: pos}
}

func (fn *PrimaryNode) isNode() {}

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
	ArgNames   *TermListNode
	Chunk      *ChunkNode
	Annotation interface{}
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
