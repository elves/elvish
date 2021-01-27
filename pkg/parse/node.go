package parse

import "src.elv.sh/pkg/diag"

// Node represents a parse tree as well as an AST.
type Node interface {
	diag.Ranger
	parse(*parser)
	n() *node
}

type node struct {
	diag.Ranging
	sourceText string
	parent     Node
	children   []Node
}

func (n *node) n() *node { return n }

func (n *node) addChild(ch Node) { n.children = append(n.children, ch) }

// Range returns the range within the full source text that parses to the node.
func (n *node) Range() diag.Ranging { return n.Ranging }

// Parent returns the parent of a node. It returns nil if the node is the root
// of the parse tree.
func Parent(n Node) Node { return n.n().parent }

// SourceText returns the part of the source text that parses to the node.
func SourceText(n Node) string { return n.n().sourceText }

// Children returns all children of the node in the parse tree.
func Children(n Node) []Node { return n.n().children }
