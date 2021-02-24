package query

import (
	"regexp"
	"strings"
)

// Query represents a compiled query, which can be used to match text.
type Query interface {
	Match(s string) bool
}

type andQuery struct {
	queries []Query
}

func (aq andQuery) Match(s string) bool {
	for _, q := range aq.queries {
		if !q.Match(s) {
			return false
		}
	}
	return true
}

type orQuery struct {
	queries []Query
}

func (oq orQuery) Match(s string) bool {
	for _, q := range oq.queries {
		if q.Match(s) {
			return true
		}
	}
	return false
}

type substringQuery struct {
	pattern    string
	ignoreCase bool
}

func (sq substringQuery) Match(s string) bool {
	if sq.ignoreCase {
		s = strings.ToLower(s)
	}
	return strings.Contains(s, sq.pattern)
}

type regexpQuery struct {
	pattern *regexp.Regexp
}

func (rq regexpQuery) Match(s string) bool {
	return rq.pattern.MatchString(s)
}
