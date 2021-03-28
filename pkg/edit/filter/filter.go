package filter

import (
	"regexp"
	"strings"
)

// Filter represents a compiled filter, which can be used to match text.
type Filter interface {
	Match(s string) bool
}

type andFilter struct {
	queries []Filter
}

func (aq andFilter) Match(s string) bool {
	for _, q := range aq.queries {
		if !q.Match(s) {
			return false
		}
	}
	return true
}

type orFilter struct {
	queries []Filter
}

func (oq orFilter) Match(s string) bool {
	for _, q := range oq.queries {
		if q.Match(s) {
			return true
		}
	}
	return false
}

type substringFilter struct {
	pattern    string
	ignoreCase bool
}

func (sq substringFilter) Match(s string) bool {
	if sq.ignoreCase {
		s = strings.ToLower(s)
	}
	return strings.Contains(s, sq.pattern)
}

type regexpFilter struct {
	pattern *regexp.Regexp
}

func (rq regexpFilter) Match(s string) bool {
	return rq.pattern.MatchString(s)
}
