// Package filter implements the Elvish filter DSL.
//
// The filter DSL is a subset of Elvish's expression syntax, and is useful for
// filtering a list of items. It is currently used in the listing modes of the
// interactive editor.
package filter

import (
	"errors"
	"regexp"
	"strings"

	"src.elv.sh/pkg/errutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
)

// Compile parses and compiles a filter.
func Compile(q string) (Filter, error) {
	qn, errParse := parseFilter(q)
	filter, errCompile := compileFilter(qn)
	return filter, errutil.Multi(errParse, errCompile)
}

func parseFilter(q string) (*parse.Filter, error) {
	qn := &parse.Filter{}
	err := parse.ParseAs(parse.Source{Name: "[filter]", Code: q}, qn, parse.Config{})
	return qn, err
}

func compileFilter(qn *parse.Filter) (Filter, error) {
	if len(qn.Opts) > 0 {
		return nil, notSupportedError{"option"}
	}
	qs, err := compileCompounds(qn.Args)
	if err != nil {
		return nil, err
	}
	return andFilter{qs}, nil
}

func compileCompounds(ns []*parse.Compound) ([]Filter, error) {
	qs := make([]Filter, len(ns))
	for i, n := range ns {
		q, err := compileCompound(n)
		if err != nil {
			return nil, err
		}
		qs[i] = q
	}
	return qs, nil
}

func compileCompound(n *parse.Compound) (Filter, error) {
	if pn, ok := cmpd.Primary(n); ok {
		switch pn.Type {
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			s := pn.Value
			ignoreCase := s == strings.ToLower(s)
			return substringFilter{s, ignoreCase}, nil
		case parse.List:
			return compileList(pn.Elements)
		}
	}
	return nil, notSupportedError{cmpd.Shape(n)}
}

var errEmptySubfilter = errors.New("empty subfilter")

func compileList(elems []*parse.Compound) (Filter, error) {
	if len(elems) == 0 {
		return nil, errEmptySubfilter
	}
	head, ok := cmpd.StringLiteral(elems[0])
	if !ok {
		return nil, notSupportedError{"non-literal subfilter head"}
	}
	switch head {
	case "re":
		if len(elems) == 1 {
			return nil, notSupportedError{"re subfilter with no argument"}
		}
		if len(elems) > 2 {
			return nil, notSupportedError{"re subfilter with two or more arguments"}
		}
		arg := elems[1]
		s, ok := cmpd.StringLiteral(arg)
		if !ok {
			return nil, notSupportedError{"re subfilter with " + cmpd.Shape(arg)}
		}
		p, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		return regexpFilter{p}, nil
	case "and":
		qs, err := compileCompounds(elems[1:])
		if err != nil {
			return nil, err
		}
		return andFilter{qs}, nil
	case "or":
		qs, err := compileCompounds(elems[1:])
		if err != nil {
			return nil, err
		}
		return orFilter{qs}, nil
	default:
		return nil, notSupportedError{"head " + parse.SourceText(elems[0])}
	}
}

type notSupportedError struct{ what string }

func (err notSupportedError) Error() string {
	return err.what + " not supported"
}
