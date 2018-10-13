package parse

import "github.com/elves/elvish/diag"

// Node represents a parse tree as well as an AST.
type Node interface {
	n() *node
	diag.Ranger
	Parent() Node
	SourceText() string
	Children() []Node
}

type node struct {
	parent     Node
	begin, end int
	sourceText string
	children   []Node
}

func (n *node) n() *node {
	return n
}

// Parent returns the parent node. If the node is the root of the syntax tree,
// the parent is nil.
func (n *node) Parent() Node {
	return n.parent
}

// Range returns the range within the original (full) source text that parses
// to the node.
func (n *node) Range() diag.Ranging {
	return diag.Ranging{n.begin, n.end}
}

// SourceText returns the part of the source text that parses to the node.
func (n *node) SourceText() string {
	return n.sourceText
}

// Children returns all children of the node in the parse tree.
func (n *node) Children() []Node {
	return n.children
}
