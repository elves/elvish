package parse

import "github.com/elves/elvish/diag"

// Node represents a parse tree as well as an AST.
type Node interface {
	diag.Ranger
	Parent() Node
	SourceText() string
	Children() []Node

	n() *node
	parse(*Parser)
	setBegin(int)
	setEnd(int)
	setSourceText(string)
}

type node struct {
	diag.Ranging
	sourceText string
	parent     Node
	children   []Node
}

func (n *node) n() *node {
	return n
}

func (n *node) setBegin(begin int) {
	n.From = begin
}

func (n *node) setEnd(end int) {
	n.To = end
}

func (n *node) setSourceText(source string) {
	n.sourceText = source
}

// Parent returns the parent node. If the node is the root of the syntax tree,
// the parent is nil.
func (n *node) Parent() Node {
	return n.parent
}

// Range returns the range within the original (full) source text that parses
// to the node.
func (n *node) Range() diag.Ranging {
	return diag.Ranging{n.From, n.To}
}

// SourceText returns the part of the source text that parses to the node.
func (n *node) SourceText() string {
	return n.sourceText
}

// Children returns all children of the node in the parse tree.
func (n *node) Children() []Node {
	return n.children
}
