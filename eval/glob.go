package eval

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/elves/elvish/glob"
	"github.com/elves/elvish/parse"
)

// GlobPattern is en ephemeral Value generated when evaluating tilde and
// wildcards.
type GlobPattern struct {
	glob.Pattern
	Flags GlobFlag
	Buts  []string
}

type GlobFlag uint

const (
	NoMatchOK GlobFlag = 1 << iota
)

func (f GlobFlag) Has(g GlobFlag) bool {
	return (f & g) == g
}

var (
	_ Value   = GlobPattern{}
	_ Indexer = GlobPattern{}
)

var (
	ErrMustFollowWildcard   = errors.New("must follow wildcard")
	ErrModifierMustBeString = errors.New("modifier must be string")
	ErrWildcardNoMatch      = errors.New("wildcard has no match")
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

func (GlobPattern) Kind() string {
	return "glob-pattern"
}

func (gp GlobPattern) Equal(a interface{}) bool {
	return reflect.DeepEqual(gp, a)
}

func (gp GlobPattern) Repr(int) string {
	return fmt.Sprintf("<GlobPattern%v>", gp)
}

func (gp GlobPattern) Index(modifiers []Value) []Value {
	for _, value := range modifiers {
		modifierv, ok := value.(String)
		if !ok {
			throw(ErrModifierMustBeString)
		}
		modifier := string(modifierv)
		switch {
		case modifier == "nomatch-ok":
			gp.Flags |= NoMatchOK
		case strings.HasPrefix(modifier, "but:"):
			gp.Buts = append(gp.Buts, modifier[len("but:"):])
		case modifier == "match-hidden":
			lastSeg := gp.mustGetLastWildSeg()
			gp.Segments[len(gp.Segments)-1] = glob.Wild{
				lastSeg.Type, true, lastSeg.Matchers,
			}
		default:
			if matcher, ok := runeMatchers[modifier]; ok {
				gp.addMatcher(matcher)
			} else if strings.HasPrefix(modifier, "set:") {
				set := modifier[len("set:"):]
				gp.addMatcher(func(r rune) bool {
					return strings.ContainsRune(set, r)
				})
			} else if strings.HasPrefix(modifier, "range:") {
				rangeExpr := modifier[len("range:"):]
				badRangeExpr := fmt.Errorf("bad range modifier: %s", parse.Quote(rangeExpr))
				runes := []rune(rangeExpr)
				if len(runes) != 3 {
					throw(badRangeExpr)
				}
				from, sep, to := runes[0], runes[1], runes[2]
				switch sep {
				case '-':
					gp.addMatcher(func(r rune) bool {
						return from <= r && r <= to
					})
				case '~':
					gp.addMatcher(func(r rune) bool {
						return from <= r && r < to
					})
				default:
					throw(badRangeExpr)
				}
			} else {
				throw(fmt.Errorf("unknown modifier %s", modifierv.Repr(NoPretty)))
			}
		}
	}
	return []Value{gp}
}

func (gp *GlobPattern) mustGetLastWildSeg() glob.Wild {
	if len(gp.Segments) == 0 {
		throw(ErrBadGlobPattern)
	}
	if !glob.IsWild(gp.Segments[len(gp.Segments)-1]) {
		throw(ErrMustFollowWildcard)
	}
	return gp.Segments[len(gp.Segments)-1].(glob.Wild)
}

func (gp *GlobPattern) addMatcher(matcher func(rune) bool) {
	lastSeg := gp.mustGetLastWildSeg()
	gp.Segments[len(gp.Segments)-1] = glob.Wild{
		lastSeg.Type, lastSeg.MatchHidden,
		append(lastSeg.Matchers, matcher),
	}
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

func doGlob(gp GlobPattern, abort <-chan struct{}) []Value {
	but := make(map[string]struct{})
	for _, s := range gp.Buts {
		but[s] = struct{}{}
	}

	vs := make([]Value, 0)
	if !gp.Glob(func(name string) bool {
		select {
		case <-abort:
			logger.Println("glob aborted")
			return false
		default:
		}
		if _, b := but[name]; !b {
			vs = append(vs, String(name))
		}
		return true
	}) {
		throw(ErrInterrupted)
	}
	if len(vs) == 0 && !gp.Flags.Has(NoMatchOK) {
		throw(ErrWildcardNoMatch)
	}
	return vs
}
