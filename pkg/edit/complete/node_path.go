package complete

import (
	"reflect"

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
//
// TODO: Avoid reflection with generics when Elvish requires Go 1.18.
type typedMatcher struct{ typ reflect.Type }

func typed(n parse.Node) nodesMatcher { return typedMatcher{reflect.TypeOf(n)} }

func (m typedMatcher) matchNodes(ns []parse.Node) ([]parse.Node, bool) {
	if len(ns) > 0 && reflect.TypeOf(ns[0]) == m.typ {
		return ns[1:], true
	}
	return nil, false
}

var (
	aChunk    = typed(&parse.Chunk{})
	aPipeline = typed(&parse.Pipeline{})
	aArray    = typed(&parse.Array{})
	aRedir    = typed(&parse.Redir{})
	aSep      = typed(&parse.Sep{})
)

// Matches one node of a certain type, and stores it into a pointer.
//
// TODO: Avoid reflection with generics when Elvish requires Go 1.18.
type storeMatcher struct {
	p   reflect.Value
	typ reflect.Type
}

func store(p interface{}) nodesMatcher {
	dst := reflect.ValueOf(p).Elem()
	return storeMatcher{dst, dst.Type()}
}

func (m storeMatcher) matchNodes(ns []parse.Node) ([]parse.Node, bool) {
	if len(ns) > 0 && reflect.TypeOf(ns[0]) == m.typ {
		m.p.Set(reflect.ValueOf(ns[0]))
		return ns[1:], true
	}
	return nil, false
}

// Matches an expression that can be evaluated statically. Consumes 3 nodes
// (Primary, Indexing and Compound).
type simpleExprMatcher struct {
	ev       PureEvaler
	s        string
	compound *parse.Compound
	quote    parse.PrimaryType
}

func simpleExpr(ev PureEvaler) *simpleExprMatcher {
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
