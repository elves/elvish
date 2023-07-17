package eval

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/glob"
	"src.elv.sh/pkg/parse"
)

// An ephemeral value generated when evaluating tilde and wildcards.
type globPattern struct {
	glob.Pattern
	Flags  globFlag
	Buts   []string
	TypeCb func(os.FileMode) bool
}

type globFlag uint

var typeCbMap = map[string]func(os.FileMode) bool{
	"dir":     os.FileMode.IsDir,
	"regular": os.FileMode.IsRegular,
}

const (
	// noMatchOK indicates that the "nomatch-ok" glob index modifier was
	// present.
	noMatchOK globFlag = 1 << iota
)

func (f globFlag) Has(g globFlag) bool {
	return (f & g) == g
}

var _ vals.ErrIndexer = globPattern{}

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

func (gp globPattern) Kind() string { return "glob-pattern" }

func (gp globPattern) Index(k any) (any, error) {
	modifierv, ok := k.(string)
	if !ok {
		return nil, ErrModifierMustBeString
	}
	modifier := modifierv
	switch {
	case modifier == "nomatch-ok":
		gp.Flags |= noMatchOK
	case strings.HasPrefix(modifier, "but:"):
		gp.Buts = append(gp.Buts, modifier[len("but:"):])
	case modifier == "match-hidden":
		lastSeg, err := gp.lastWildSeg()
		if err != nil {
			return nil, err
		}
		gp.Segments[len(gp.Segments)-1] = glob.Wild{
			Type: lastSeg.Type, MatchHidden: true, Matchers: lastSeg.Matchers,
		}
	case strings.HasPrefix(modifier, "type:"):
		if gp.TypeCb != nil {
			return nil, ErrMultipleTypeModifiers
		}
		typeName := modifier[len("type:"):]
		cb, ok := typeCbMap[typeName]
		if !ok {
			return nil, ErrUnknownTypeModifier
		}
		gp.TypeCb = cb
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
			return nil, fmt.Errorf("unknown modifier %s", vals.ReprPlain(modifierv))
		}
		err := gp.addMatcher(matcher)
		return gp, err
	}
	return gp, nil
}

func (gp globPattern) Concat(v any) (any, error) {
	switch rhs := v.(type) {
	case string:
		var segs []glob.Segment
		segs = append(segs, gp.Segments...)
		segs = append(segs, stringToSegments(rhs)...)
		return globPattern{Pattern: glob.Pattern{Segments: segs}, Flags: gp.Flags,
			Buts: gp.Buts, TypeCb: gp.TypeCb}, nil
	case globPattern:
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

func (gp globPattern) RConcat(v any) (any, error) {
	switch lhs := v.(type) {
	case string:
		segs := stringToSegments(lhs)
		// We know gp contains exactly one segment.
		segs = append(segs, gp.Segments[0])
		return globPattern{Pattern: glob.Pattern{Segments: segs}, Flags: gp.Flags,
			Buts: gp.Buts, TypeCb: gp.TypeCb}, nil
	}

	return nil, vals.ErrConcatNotImplemented
}

func (gp *globPattern) lastWildSeg() (glob.Wild, error) {
	if len(gp.Segments) == 0 {
		return glob.Wild{}, ErrBadglobPattern
	}
	if !glob.IsWild(gp.Segments[len(gp.Segments)-1]) {
		return glob.Wild{}, ErrMustFollowWildcard
	}
	return gp.Segments[len(gp.Segments)-1].(glob.Wild), nil
}

func (gp *globPattern) addMatcher(matcher func(rune) bool) error {
	lastSeg, err := gp.lastWildSeg()
	if err != nil {
		return err
	}
	gp.Segments[len(gp.Segments)-1] = glob.Wild{
		Type: lastSeg.Type, MatchHidden: lastSeg.MatchHidden,
		Matchers: append(lastSeg.Matchers, matcher),
	}
	return nil
}

func (gp *globPattern) append(segs ...glob.Segment) {
	gp.Segments = append(gp.Segments, segs...)
}

func wildcardToSegment(s string) (glob.Segment, error) {
	switch s {
	case "*":
		return glob.Wild{Type: glob.Star, MatchHidden: false, Matchers: nil}, nil
	case "**":
		return glob.Wild{Type: glob.StarStar, MatchHidden: false, Matchers: nil}, nil
	case "?":
		return glob.Wild{Type: glob.Question, MatchHidden: false, Matchers: nil}, nil
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
			segs = append(segs, glob.Literal{Data: s[i:j]})
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

func doGlob(ctx context.Context, gp globPattern) ([]any, error) {
	but := make(map[string]struct{})
	for _, s := range gp.Buts {
		but[s] = struct{}{}
	}

	vs := make([]any, 0)
	if !gp.Glob(func(pathInfo glob.PathInfo) bool {
		select {
		case <-ctx.Done():
			logger.Println("glob aborted")
			return false
		default:
		}

		if _, ignore := but[pathInfo.Path]; ignore {
			return true
		}

		if gp.TypeCb == nil || gp.TypeCb(pathInfo.Info.Mode()) {
			vs = append(vs, pathInfo.Path)
		}
		return true
	}) {
		return nil, ErrInterrupted
	}
	if len(vs) == 0 && !gp.Flags.Has(noMatchOK) {
		return nil, ErrWildcardNoMatch
	}
	return vs, nil
}
