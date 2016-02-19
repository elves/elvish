package eval

import (
	"fmt"

	"github.com/elves/elvish/glob"
)

// GlobPattern is en ephemeral Value generated when evaluating tilde and
// wildcards.
type GlobPattern glob.Pattern

func (GlobPattern) Kind() string {
	return "glob-pattern"
}

func (gp GlobPattern) Repr(int) string {
	return fmt.Sprintf("<GlobPattern%v>", gp)
}

func (gp *GlobPattern) append(segs ...glob.Segment) {
	gp.Segments = append(gp.Segments, segs...)
}

func wildcardToSegment(s string) glob.Segment {
	switch s {
	case "*":
		return glob.Segment{glob.Star, ""}
	case "**":
		return glob.Segment{glob.StarStar, ""}
	case "?":
		return glob.Segment{glob.Question, ""}
	default:
		throw(fmt.Errorf("bad wildcard: %q", s))
		panic("unreachable")
	}
}

func stringToSegments(s string) []glob.Segment {
	segs := []glob.Segment{}
	for i := 0; i < len(s); {
		j := i
		for ; j < len(s) && s[j] != '/'; j++ {
		}
		if j > i {
			segs = append(segs, glob.Segment{glob.Literal, s[i:j]})
		}
		if j < len(s) {
			for ; j < len(s) && s[j] == '/'; j++ {
			}
			segs = append(segs,
				glob.Segment{glob.Slash, ""})
			i = j
		} else {
			break
		}
	}
	return segs
}

func doGlob(gp GlobPattern) []Value {
	names := glob.Pattern(gp).Glob()
	vs := make([]Value, len(names))
	for i, name := range names {
		vs[i] = String(name)
	}
	return vs
}
