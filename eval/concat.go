package eval

import "github.com/elves/elvish/glob"

type (
	Concater interface {
		ConcatWith(other interface{}) (v interface{}, ok bool)
	}

	ErrConcater interface {
		ConcatWith(other interface{}) (interface{}, error)
	}
)

type StringConcater string

func (s StringConcater) ConcatWith(other interface{}) (interface{}, bool) {
	switch o := other.(type) {
	case string:
		return string(s) + o, true
	case StringConcater:
		return string(s) + string(o), true
	case GlobPattern:
		segs := stringToSegments(string(s))
		// We know o contains exactly one segment.
		segs = append(segs, o.Segments[0])
		return GlobPattern{glob.Pattern{Segments: segs}, o.Flags, o.Buts}, true
	}

	return nil, false
}
