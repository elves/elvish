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
	Command     *CompoundNode
	Args        *SpacedNode
	Redirs      []Redir
	StatusRedir string
}

func newForm(pos Pos) *FormNode {
	return &FormNode{Pos: pos}
}

func (fn *FormNode) isNode() {}

// CompoundNode is a list of SubscriptNode's.
type CompoundNode struct {
	Pos
	Nodes []*SubscriptNode
}

func newCompound(pos Pos, nodes ...*SubscriptNode) *CompoundNode {
	return &CompoundNode{pos, nodes}
}

func (l *CompoundNode) isNode() {}

func (tn *CompoundNode) append(n *SubscriptNode) {
	tn.Nodes = append(tn.Nodes, n)
}

// SpacedNode is a list of CompoundNode's.
type SpacedNode struct {
	Pos
	Nodes []*CompoundNode
}

func newSpaced(pos Pos, nodes ...*CompoundNode) *SpacedNode {
	return &SpacedNode{pos, nodes}
}

func (l *SpacedNode) isNode() {}

func (tn *SpacedNode) append(n *CompoundNode) {
	tn.Nodes = append(tn.Nodes, n)
}

// SubscriptNode represents a subscript expression.
type SubscriptNode struct {
	Pos
	Left  *PrimaryNode
	Right *CompoundNode
}

func (s *SubscriptNode) isNode() {}

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
	Key   *CompoundNode
	Value *CompoundNode
}

// TableNode holds a table literal.
type TableNode struct {
	Pos
	List []*CompoundNode
	Dict []*TablePair
}

func newTable(pos Pos) *TableNode {
	return &TableNode{Pos: pos}
}

func (tn *TableNode) isNode() {}

func (tn *TableNode) appendToList(compound *CompoundNode) {
	tn.List = append(tn.List, compound)
}

func (tn *TableNode) appendToDict(key *CompoundNode, value *CompoundNode) {
	tn.Dict = append(tn.Dict, &TablePair{key, value})
}

// ClosureNode holds a closure literal.
type ClosureNode struct {
	Pos
	ArgNames   *SpacedNode
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
