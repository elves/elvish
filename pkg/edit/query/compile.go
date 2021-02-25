// Package query implements the Elvish query language.
//
// The Elvish query language is a DSL for filtering data, such as in history
// listing mode and location mode of the interactive editor. It is a subset of
// Elvish's expression syntax, currently a very small one.
package query

import (
	"errors"
	"regexp"
	"strings"

	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
)

// Compile parses and compiles a query.
func Compile(q string) (Query, error) {
	qn := &parse.Query{}
	err := parse.ParseAs(parse.Source{Code: q}, qn, parse.Config{})
	if err != nil {
		return nil, err
	}
	return compileQuery(qn)
}

func compileQuery(qn *parse.Query) (Query, error) {
	if len(qn.Opts) > 0 {
		return nil, notSupportedError{"option"}
	}
	qs, err := compileCompounds(qn.Args)
	if err != nil {
		return nil, err
	}
	return andQuery{qs}, nil
}

func compileCompounds(ns []*parse.Compound) ([]Query, error) {
	qs := make([]Query, len(ns))
	for i, n := range ns {
		q, err := compileCompound(n)
		if err != nil {
			return nil, err
		}
		qs[i] = q
	}
	return qs, nil
}

func compileCompound(n *parse.Compound) (Query, error) {
	if pn, ok := cmpd.Primary(n); ok {
		switch pn.Type {
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			s := pn.Value
			ignoreCase := s == strings.ToLower(s)
			return substringQuery{s, ignoreCase}, nil
		case parse.List:
			return compileList(pn.Elements)
		}
	}
	return nil, notSupportedError{cmpd.Shape(n)}
}

var errEmptySubquery = errors.New("empty subquery")

func compileList(elems []*parse.Compound) (Query, error) {
	if len(elems) == 0 {
		return nil, errEmptySubquery
	}
	head, ok := cmpd.StringLiteral(elems[0])
	if !ok {
		return nil, notSupportedError{"non-literal subquery head"}
	}
	switch head {
	case "re":
		if len(elems) == 1 {
			return nil, notSupportedError{"re subquery with no argument"}
		}
		if len(elems) > 2 {
			return nil, notSupportedError{"re subquery with two or more arguments"}
		}
		arg := elems[1]
		s, ok := cmpd.StringLiteral(arg)
		if !ok {
			return nil, notSupportedError{"re subquery with " + cmpd.Shape(arg)}
		}
		p, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		return regexpQuery{p}, nil
	case "and":
		qs, err := compileCompounds(elems[1:])
		if err != nil {
			return nil, err
		}
		return andQuery{qs}, nil
	case "or":
		qs, err := compileCompounds(elems[1:])
		if err != nil {
			return nil, err
		}
		return orQuery{qs}, nil
	default:
		return nil, notSupportedError{"head " + parse.SourceText(elems[0])}
	}
}

type notSupportedError struct{ what string }

func (err notSupportedError) Error() string {
	return err.what + " not supported"
}
