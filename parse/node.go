package parse

import "github.com/elves/elvish/diag"

// Node represents a parse tree as well as an AST.
type Node interface {
	parse(*parser)

	diag.Ranger
	SourceText() string
	Parent() Node
	Children() []Node

	setFrom(int)
	setTo(int)
	setSourceText(string)
	setParent(Node)
	addChild(Node)
}

type node struct {
	diag.Ranging
	sourceText string
	parent     Node
	children   []Node
}

func (n *node) setFrom(begin int) { n.From = begin }

func (n *node) setTo(end int) { n.To = end }

func (n *node) setSourceText(source string) { n.sourceText = source }

func (n *node) setParent(p Node) { n.parent = p }

func (n *node) addChild(ch Node) { n.children = append(n.children, ch) }

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
