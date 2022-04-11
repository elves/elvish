package complete

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

type nodePath []parse.Node

// Returns the path of Node's from n to a leaf at position p. Leaf first in the
// returned slice.
func findNodePath(root parse.Node, p int) nodePath {
	n := root
descend:
	for len(parse.Children(n)) > 0 {
		for _, ch := range parse.Children(n) {
			if rg := ch.Range(); rg.From <= p && p <= rg.To {
				n = ch
				continue descend
			}
		}
		return nil
	}
	var path []parse.Node
	for {
		path = append(path, n)
		if n == root {
			break
		}
		n = parse.Parent(n)
	}
	return path
}

func (ns nodePath) match(ms ...nodesMatcher) bool {
	for _, m := range ms {
		ns2, ok := m.matchNodes(ns)
		if !ok {
			return false
		}
		ns = ns2
	}
	return true
}

type nodesMatcher interface {
	// Matches from the beginning of nodes. Returns the remaining nodes and
	// whether the match succeeded.
	matchNodes([]parse.Node) ([]parse.Node, bool)
}

// Matches one node of a given type, without storing it.
type typedMatcher[T parse.Node] struct{}

func (m typedMatcher[T]) matchNodes(ns []parse.Node) ([]parse.Node, bool) {
	if len(ns) > 0 {
		if _, ok := ns[0].(T); ok {
			return ns[1:], true
		}
	}
	return nil, false
}

var (
	aChunk    = typedMatcher[*parse.Chunk]{}
	aPipeline = typedMatcher[*parse.Pipeline]{}
	aArray    = typedMatcher[*parse.Array]{}
	aRedir    = typedMatcher[*parse.Redir]{}
	aSep      = typedMatcher[*parse.Sep]{}
)

// Matches one node of a certain type, and stores it into a pointer.
type storeMatcher[T parse.Node] struct{ p *T }

func store[T parse.Node](p *T) nodesMatcher { return storeMatcher[T]{p} }

func (m storeMatcher[T]) matchNodes(ns []parse.Node) ([]parse.Node, bool) {
	if len(ns) > 0 {
		if n, ok := ns[0].(T); ok {
			*m.p = n
			return ns[1:], true
		}
	}
	return nil, false
}

// Matches an expression that can be evaluated statically. Consumes 3 nodes
// (Primary, Indexing and Compound).
type simpleExprMatcher struct {
	ev       *eval.Evaler
	s        string
	compound *parse.Compound
	quote    parse.PrimaryType
}

func simpleExpr(ev *eval.Evaler) *simpleExprMatcher {
	return &simpleExprMatcher{ev: ev}
}

func (m *simpleExprMatcher) matchNodes(ns []parse.Node) ([]parse.Node, bool) {
	if len(ns) < 3 {
		return nil, false
	}
	primary, ok := ns[0].(*parse.Primary)
	if !ok {
		return nil, false
	}
	indexing, ok := ns[1].(*parse.Indexing)
	if !ok {
		return nil, false
	}
	compound, ok := ns[2].(*parse.Compound)
	if !ok {
		return nil, false
	}
	s, ok := m.ev.PurelyEvalPartialCompound(compound, indexing.To)
	if !ok {
		return nil, false
	}
	m.compound, m.quote, m.s = compound, primary.Type, s
	return ns[3:], true
}
