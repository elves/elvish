// Package cmpd contains utilities for working with compound nodes.
package cmpd

import (
	"fmt"

	"src.elv.sh/pkg/parse"
)

// Primary returns a primary node and true if that's the only child of the
// compound node. Otherwise it returns nil and false.
func Primary(n *parse.Compound) (*parse.Primary, bool) {
	if n != nil && len(n.Indexings) == 1 && len(n.Indexings[0].Indices) == 0 {
		return n.Indexings[0].Head, true
	}
	return nil, false
}

// StringLiteral returns the value of a string literal and true if that's the
// only child of the compound node. Otherwise it returns "" and false.
func StringLiteral(n *parse.Compound) (string, bool) {
	if pn, ok := Primary(n); ok {
		switch pn.Type {
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			return pn.Value, true
		}
	}
	return "", false
}

// Lambda returns a lambda primary node and true if that's the only child of the
// compound node. Otherwise it returns nil and false.
func Lambda(n *parse.Compound) (*parse.Primary, bool) {
	if pn, ok := Primary(n); ok {
		if pn.Type == parse.Lambda {
			return pn, true
		}
	}
	return nil, false
}

// StringLiteralOrError is like StringLiteral, but returns an error suitable as
// a compiler error when StringLiteral would return false.
func StringLiteralOrError(n *parse.Compound, what string) (string, error) {
	s, ok := StringLiteral(n)
	if !ok {
		return "", fmt.Errorf("%s must be string literal, found %s", what, Shape(n))
	}
	return s, nil
}

// Shape describes the shape of the compound node.
func Shape(n *parse.Compound) string {
	if len(n.Indexings) == 0 {
		return "empty expression"
	}
	if len(n.Indexings) > 1 {
		return "compound expression"
	}
	in := n.Indexings[0]
	if len(in.Indices) > 0 {
		return "indexing expression"
	}
	pn := in.Head
	return "primary expression of type " + pn.Type.String()
}
