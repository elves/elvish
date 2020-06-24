package eval

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/glob"
	"github.com/elves/elvish/pkg/parse"
)

// GlobPattern is en ephemeral Value generated when evaluating tilde and
// wildcards.
type GlobPattern struct {
	glob.Pattern
	Flags  GlobFlag
	Buts   []string
	TypeCb func(os.FileMode) bool
}

type GlobFlag uint

var typeCbMap = map[string]func(os.FileMode) bool{
	"dir":     os.FileMode.IsDir,
	"regular": os.FileMode.IsRegular,
}

const (
	NoMatchOK GlobFlag = 1 << iota
)

func (f GlobFlag) Has(g GlobFlag) bool {
	return (f & g) == g
}

var (
	_ interface{}     = GlobPattern{}
	_ vals.ErrIndexer = GlobPattern{}
)

var (
	ErrMustFollowWildcard    = errors.New("must follow wildcard")
	ErrModifierMustBeString  = errors.New("modifier must be string")
	ErrWildcardNoMatch       = errors.New("wildcard has no match")
	ErrMultipleTypeModifiers = errors.New("only one type modifier allowed")
	ErrUnknownTypeModifier   = errors.New("unknown type modifier")
)

var runeMatchers = map[string]func(rune) bool{
	"control": unicode.IsControl,
	"digit":   unicode.IsDigit,
	"graphic": unicode.IsGraphic,
	"letter":  unicode.IsLetter,
	"lower":   unicode.IsDigit,
	"mark":    unicode.IsMark,
	"number":  unicode.IsNumber,
	"print":   unicode.IsPrint,
	"punct":   unicode.IsPunct,
	"space":   unicode.IsSpace,
	"symbol":  unicode.IsSymbol,
	"title":   unicode.IsTitle,
	"upper":   unicode.IsUpper,
}

func (gp GlobPattern) Index(k interface{}) (interface{}, error) {
	modifierv, ok := k.(string)
	if !ok {
		return nil, ErrModifierMustBeString
	}
	modifier := modifierv
	switch {
	case modifier == "nomatch-ok":
		gp.Flags |= NoMatchOK
	case strings.HasPrefix(modifier, "but:"):
		gp.Buts = append(gp.Buts, modifier[len("but:"):])
	case modifier == "match-hidden":
		lastSeg, err := gp.lastWildSeg()
		if err != nil {
			return nil, err
		}
		gp.Segments[len(gp.Segments)-1] = glob.Wild{
			lastSeg.Type, true, lastSeg.Matchers,
		}
	case strings.HasPrefix(modifier, "type:"):
		if gp.TypeCb != nil {
			return nil, ErrMultipleTypeModifiers
		}
		switch {
		default:
			typeName := modifier[len("type:"):]
			cb, ok := typeCbMap[typeName]
			if !ok {
				return nil, ErrUnknownTypeModifier
			}
			gp.TypeCb = cb
		}
	default:
		var matcher func(rune) bool
		if m, ok := runeMatchers[modifier]; ok {
			matcher = m
		} else if strings.HasPrefix(modifier, "set:") {
			set := modifier[len("set:"):]
			matcher = func(r rune) bool {
				return strings.ContainsRune(set, r)
			}
		} else if strings.HasPrefix(modifier, "range:") {
			rangeExpr := modifier[len("range:"):]
			badRangeExpr := fmt.Errorf("bad range modifier: %s", parse.Quote(rangeExpr))
			runes := []rune(rangeExpr)
			if len(runes) != 3 {
				return nil, badRangeExpr
			}
			from, sep, to := runes[0], runes[1], runes[2]
			switch sep {
			case '-':
				matcher = func(r rune) bool {
					return from <= r && r <= to
				}
			case '~':
				matcher = func(r rune) bool {
					return from <= r && r < to
				}
			default:
				return nil, badRangeExpr
			}
		} else {
			return nil, fmt.Errorf("unknown modifier %s", vals.Repr(modifierv, vals.NoPretty))
		}
		err := gp.addMatcher(matcher)
		return gp, err
	}
	return gp, nil
}

func (gp GlobPattern) Concat(v interface{}) (interface{}, error) {
	switch rhs := v.(type) {
	case string:
		gp.append(stringToSegments(rhs)...)
		return gp, nil
	case GlobPattern:
		// We know rhs contains exactly one segment.
		gp.append(rhs.Segments[0])
		gp.Flags |= rhs.Flags
		gp.Buts = append(gp.Buts, rhs.Buts...)
		// This handles illegal cases such as `**[type:regular]x*[type:directory]`.
		if gp.TypeCb != nil && rhs.TypeCb != nil {
			return nil, ErrMultipleTypeModifiers
		}
		if rhs.TypeCb != nil {
			gp.TypeCb = rhs.TypeCb
		}
		return gp, nil
	}

	return nil, vals.ErrConcatNotImplemented
}

func (gp GlobPattern) RConcat(v interface{}) (interface{}, error) {
	switch lhs := v.(type) {
	case string:
		segs := stringToSegments(lhs)
		// We know gp contains exactly one segment.
		segs = append(segs, gp.Segments[0])
		return GlobPattern{Pattern: glob.Pattern{Segments: segs}, Flags: gp.Flags,
			Buts: gp.Buts, TypeCb: gp.TypeCb}, nil
	}

	return nil, vals.ErrConcatNotImplemented
}

func (gp *GlobPattern) lastWildSeg() (glob.Wild, error) {
	if len(gp.Segments) == 0 {
		return glob.Wild{}, ErrBadGlobPattern
	}
	if !glob.IsWild(gp.Segments[len(gp.Segments)-1]) {
		return glob.Wild{}, ErrMustFollowWildcard
	}
	return gp.Segments[len(gp.Segments)-1].(glob.Wild), nil
}

func (gp *GlobPattern) addMatcher(matcher func(rune) bool) error {
	lastSeg, err := gp.lastWildSeg()
	if err != nil {
		return err
	}
	gp.Segments[len(gp.Segments)-1] = glob.Wild{
		lastSeg.Type, lastSeg.MatchHidden,
		append(lastSeg.Matchers, matcher),
	}
	return nil
}

func (gp *GlobPattern) append(segs ...glob.Segment) {
	gp.Segments = append(gp.Segments, segs...)
}

func wildcardToSegment(s string) (glob.Segment, error) {
	switch s {
	case "*":
		return glob.Wild{glob.Star, false, nil}, nil
	case "**":
		return glob.Wild{glob.StarStar, false, nil}, nil
	case "?":
		return glob.Wild{glob.Question, false, nil}, nil
	default:
		return nil, fmt.Errorf("bad wildcard: %q", s)
	}
}

func stringToSegments(s string) []glob.Segment {
	segs := []glob.Segment{}
	for i := 0; i < len(s); {
		j := i
		for ; j < len(s) && s[j] != '/'; j++ {
		}
		if j > i {
			segs = append(segs, glob.Literal{s[i:j]})
		}
		if j < len(s) {
			for ; j < len(s) && s[j] == '/'; j++ {
			}
			segs = append(segs, glob.Slash{})
			i = j
		} else {
			break
		}
	}
	return segs
}

func doGlob(gp GlobPattern, abort <-chan struct{}) ([]interface{}, error) {
	but := make(map[string]struct{})
	for _, s := range gp.Buts {
		but[s] = struct{}{}
	}

	vs := make([]interface{}, 0)
	if !gp.Glob(func(pathInfo glob.PathInfo) bool {
		select {
		case <-abort:
			logger.Println("glob aborted")
			return false
		default:
		}

		if _, ignore := but[pathInfo.Path]; ignore {
			return true
		}

		if gp.TypeCb == nil || (gp.TypeCb)(pathInfo.Info.Mode()) {
			vs = append(vs, pathInfo.Path)
		}
		return true
	}) {
		return nil, ErrInterrupted
	}
	if len(vs) == 0 && !gp.Flags.Has(NoMatchOK) {
		return nil, ErrWildcardNoMatch
	}
	return vs, nil
}
