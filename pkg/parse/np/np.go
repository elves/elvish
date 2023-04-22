// Package np provides utilities for working with node paths from a leaf of a
// parse tree to the root.
package np

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

// Path is a path from a leaf in a parse tree to the root.
type Path []parse.Node

// Find finds the path of nodes from the leaf at position p to the root.
func Find(root parse.Node, p int) Path { return find(root, p, false) }

// FindLeft finds the path of nodes from the leaf at position p to the root. If
// p points to the start of one node (p == x.From), FindLeft finds the node to
// the left instead (y s.t. p == y.To).
func FindLeft(root parse.Node, p int) Path { return find(root, p, true) }

func find(root parse.Node, p int, preferLeft bool) Path {
	n := root
descend:
	for len(parse.Children(n)) > 0 {
		for _, ch := range parse.Children(n) {
			r := ch.Range()
			if r.From <= p && p < r.To || preferLeft && p == r.To {
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

// Match matches against matchers, and returns whether all matches have
// succeeded.
func (p Path) Match(ms ...Matcher) bool {
	for _, m := range ms {
		p2, ok := m.Match(p)
		if !ok {
			return false
		}
		p = p2
	}
	return true
}

// Matcher wraps the Match method.
type Matcher interface {
	// Match takes a slice of nodes and returns the remaining nodes and whether
	// the match succeeded.
	Match([]parse.Node) ([]parse.Node, bool)
}

// Typed returns a [Matcher] matching one node of a given type.
func Typed[T parse.Node]() Matcher { return typedMatcher[T]{} }

// Commonly used [Typed] matchers.
var (
	Chunk    = Typed[*parse.Chunk]()
	Pipeline = Typed[*parse.Pipeline]()
	Array    = Typed[*parse.Array]()
	Redir    = Typed[*parse.Redir]()
	Sep      = Typed[*parse.Sep]()
)

type typedMatcher[T parse.Node] struct{}

func (m typedMatcher[T]) Match(ns []parse.Node) ([]parse.Node, bool) {
	if len(ns) > 0 {
		if _, ok := ns[0].(T); ok {
			return ns[1:], true
		}
	}
	return nil, false
}

// Store returns a [Matcher] matching one node of a given type, and stores it
// if a match succeeds.
func Store[T parse.Node](p *T) Matcher { return storeMatcher[T]{p} }

type storeMatcher[T parse.Node] struct{ p *T }

func (m storeMatcher[T]) Match(ns []parse.Node) ([]parse.Node, bool) {
	if len(ns) > 0 {
		if n, ok := ns[0].(T); ok {
			*m.p = n
			return ns[1:], true
		}
	}
	return nil, false
}

// SimpleExpr returns a [Matcher] matching a "simple expression", which consists
// of 3 nodes from the leaf upwards (Primary, Indexing and Compound) and where
// the Compound expression can be evaluated statically using ev.
func SimpleExpr(data *SimpleExprData, ev *eval.Evaler) Matcher {
	return simpleExprMatcher{data, ev}
}

// SimpleExprData contains useful data written by the [SimpleExpr] matcher.
type SimpleExprData struct {
	Value      string
	Compound   *parse.Compound
	PrimarType parse.PrimaryType
}

type simpleExprMatcher struct {
	data *SimpleExprData
	ev   *eval.Evaler
}

func (m simpleExprMatcher) Match(ns []parse.Node) ([]parse.Node, bool) {
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
	value, ok := m.ev.PurelyEvalPartialCompound(compound, indexing.To)
	if !ok {
		return nil, false
	}
	*m.data = SimpleExprData{value, compound, primary.Type}
	return ns[3:], true
}
