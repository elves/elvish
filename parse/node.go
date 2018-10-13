package parse

import "github.com/elves/elvish/diag"

// Node represents a parse tree as well as an AST.
type Node interface {
	n() *node
	diag.Ranger
	Begin() int
	End() int
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

func (n *node) Parent() Node {
	return n.parent
}

func (n *node) Begin() int {
	return n.begin
}

func (n *node) End() int {
	return n.end
}

// Range returns the range within the original (full) source text that parses
// into the node.
func (n *node) Range() diag.Ranging {
	return diag.Ranging{n.begin, n.end}
}

func (n *node) SourceText() string {
	return n.sourceText
}

func (n *node) Children() []Node {
	return n.children
}
