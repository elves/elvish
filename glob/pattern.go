package glob

// Pattern is a glob pattern.
type Pattern struct {
	Segments    []Segment
	DirOverride string
}

// Segment is the building block of Pattern.
type Segment interface {
	isSegment()
}

// Slash represents a slash "/".
type Slash struct{}

// Literal is a series of non-slash, non-wildcard characters, that is to be
// matched literally.
type Literal struct {
	Data string
}

// Wild is a wildcard.
type Wild struct {
	Type        WildType
	MatchHidden bool
	Matchers    []func(rune) bool
}

// WildType is the type of a Wild.
type WildType int

// Values for WildType.
const (
	Question = iota
	Star
	StarStar
)

// Match returns whether a rune is within the match set.
func (w Wild) Match(r rune) bool {
	if len(w.Matchers) == 0 {
		return true
	}
	for _, m := range w.Matchers {
		if m(r) {
			return true
		}
	}
	return false
}

func (Literal) isSegment() {}
func (Slash) isSegment()   {}
func (Wild) isSegment()    {}

// IsSlash returns whether a Segment is a Slash.
func IsSlash(seg Segment) bool {
	_, ok := seg.(Slash)
	return ok
}

// IsLiteral returns whether a Segment is a Literal.
func IsLiteral(seg Segment) bool {
	_, ok := seg.(Literal)
	return ok
}

// IsWild returns whether a Segment is a Wild.
func IsWild(seg Segment) bool {
	_, ok := seg.(Wild)
	return ok
}

// IsWild1 returns whether a Segment is a Wild and has the specified type.
func IsWild1(seg Segment, t WildType) bool {
	return IsWild(seg) && seg.(Wild).Type == t
}

// IsWild2 returns whether a Segment is a Wild and has one of the two specified
// types.
func IsWild2(seg Segment, t1, t2 WildType) bool {
	return IsWild(seg) && (seg.(Wild).Type == t1 || seg.(Wild).Type == t2)
}
