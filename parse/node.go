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

// Chunk is a list of Form's.
type Chunk struct {
	Pos
	Nodes []*Pipeline
}

func newChunk(pos Pos, nodes ...*Pipeline) *Chunk {
	return &Chunk{pos, nodes}
}

func (cn *Chunk) isNode() {}

func (cn *Chunk) append(n *Pipeline) {
	cn.Nodes = append(cn.Nodes, n)
}

// Pipeline is a list of Form's.
type Pipeline struct {
	Pos
	Nodes []*Form
}

func newPipeline(pos Pos, nodes ...*Form) *Pipeline {
	return &Pipeline{Pos: pos, Nodes: nodes}
}

func (pn *Pipeline) isNode() {}

func (pn *Pipeline) append(n *Form) {
	pn.Nodes = append(pn.Nodes, n)
}

// Form holds a form.
type Form struct {
	Pos
	Command     *Compound
	Args        *Spaced
	Redirs      []Redir
	StatusRedir string
}

func newForm(pos Pos) *Form {
	return &Form{Pos: pos}
}

func (fn *Form) isNode() {}

// Special value for CompoundNode.Sigil to indicate no sigil is present.
const NoSigil rune = -1

// Compound is a list of Subscript's.
type Compound struct {
	Pos
	Sigil rune
	Nodes []*Subscript
}

func newCompound(pos Pos, sigil rune, nodes ...*Subscript) *Compound {
	return &Compound{pos, sigil, nodes}
}

func (cn *Compound) isNode() {}

func (cn *Compound) append(n *Subscript) {
	cn.Nodes = append(cn.Nodes, n)
}

// Spaced is a list of Compound's.
type Spaced struct {
	Pos
	Nodes []*Compound
}

func newSpaced(pos Pos, nodes ...*Compound) *Spaced {
	return &Spaced{pos, nodes}
}

func (sn *Spaced) isNode() {}

func (sn *Spaced) append(n *Compound) {
	sn.Nodes = append(sn.Nodes, n)
}

// Subscript represents a subscript expression.
type Subscript struct {
	Pos
	Left  *Primary
	Right *Compound
}

func (sn *Subscript) isNode() {}

// Primary represents a primary expression.
type Primary struct {
	Pos
	Typ  PrimaryType
	Node Node
}

// PrimaryType determines the type of a PrimaryNode.
type PrimaryType int

// PrimaryType constants.
const (
	BadPrimary PrimaryType = iota

	StringPrimary        // string literal: a `a` a
	VariablePrimary      // variable: $a
	TablePrimary         // table: [a b c &k v]
	ClosurePrimary       // closure: {|a| cmd}
	ListPrimary          // list: {a b c}
	ChanCapturePrimary   // channel output capture: (cmd1|cmd2)
	StatusCapturePrimary // status capture: ?(cmd1|cmd2)
)

func newPrimary(pos Pos) *Primary {
	return &Primary{Pos: pos}
}

func (pn *Primary) isNode() {}

// TablePair represents a key/value pair in table literal.
type TablePair struct {
	Key   *Compound
	Value *Compound
}

func newTablePair(k, v *Compound) *TablePair {
	return &TablePair{k, v}
}

// Table holds a table literal.
type Table struct {
	Pos
	List []*Compound
	Dict []*TablePair
}

func newTable(pos Pos) *Table {
	return &Table{Pos: pos}
}

func (tn *Table) isNode() {}

func (tn *Table) appendToList(compound *Compound) {
	tn.List = append(tn.List, compound)
}

func (tn *Table) appendToDict(key *Compound, value *Compound) {
	tn.Dict = append(tn.Dict, &TablePair{key, value})
}

// Closure holds a closure literal.
type Closure struct {
	Pos
	ArgNames *Spaced
	Chunk    *Chunk
}

func newClosure(pos Pos) *Closure {
	return &Closure{Pos: pos}
}

func (cn *Closure) isNode() {}

// String holds a string literal. The value has been "unquoted".
type String struct {
	Pos
	Quoted string // The original text of the string, with quotes.
	Text   string // The string, after quote processing.
}

func newString(pos Pos, orig, text string) *String {
	return &String{Pos: pos, Quoted: orig, Text: text}
}

func (sn *String) isNode() {}
