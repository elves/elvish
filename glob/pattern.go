package glob

// Pattern is a glob pattern.
type Pattern struct {
	Segments    []Segment
	DirOverride string
}

type Segment interface {
	isSegment()
}

type Slash struct{}

type Literal struct {
	Data string
}

type Wild struct {
	Type        WildType
	MatchHidden bool
}

// WildType is the type of a Wild.
type WildType int

// Values for WildType.
const (
	Question = iota
	Star
	StarStar
)

func (Literal) isSegment() {}
func (Slash) isSegment()   {}
func (Wild) isSegment()    {}

func IsSlash(seg Segment) bool {
	_, ok := seg.(Slash)
	return ok
}

func IsLiteral(seg Segment) bool {
	_, ok := seg.(Literal)
	return ok
}

func IsWild(seg Segment) bool {
	_, ok := seg.(Wild)
	return ok
}

func IsWild1(seg Segment, t WildType) bool {
	return IsWild(seg) && seg.(Wild).Type == t
}

func IsWild2(seg Segment, t1, t2 WildType) bool {
	return IsWild(seg) && (seg.(Wild).Type == t1 || seg.(Wild).Type == t2)
}
