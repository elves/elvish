package eval

import (
	"errors"
	"strings"

	"github.com/elves/elvish/pkg/fsutil"
	"github.com/elves/elvish/pkg/parse"
)

var ErrImpure = errors.New("expression is impure")

func (ev *Evaler) PurelyEvalCompound(cn *parse.Compound) (string, error) {
	return ev.PurelyEvalPartialCompound(cn, -1)
}

func (ev *Evaler) PurelyEvalPartialCompound(cn *parse.Compound, upto int) (string, error) {
	tilde := false
	head := ""
	for _, in := range cn.Indexings {
		if len(in.Indicies) > 0 {
			return "", ErrImpure
		}
		if upto >= 0 && in.To > upto {
			break
		}
		switch in.Head.Type {
		case parse.Tilde:
			tilde = true
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			head += in.Head.Value
		case parse.Variable:
			if ev == nil {
				return "", ErrImpure
			}
			v := ev.PurelyEvalPrimary(in.Head)
			if s, ok := v.(string); ok {
				head += s
			} else {
				return "", ErrImpure
			}
		default:
			return "", ErrImpure
		}
	}
	if tilde {
		i := strings.Index(head, "/")
		if i == -1 {
			i = len(head)
		}
		uname := head[:i]
		home, err := fsutil.GetHome(uname)
		if err != nil {
			return "", err
		}
		head = home + head[i:]
	}
	return head, nil
}

// PurelyEvalPrimary evaluates a primary node without causing any side effects.
// If this cannot be done, it returns nil.
//
// Currently, only string literals and variables with no @ can be evaluated.
func (ev *Evaler) PurelyEvalPrimary(pn *parse.Primary) interface{} {
	switch pn.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return pn.Value
	case parse.Variable:
		sigil, qname := SplitSigil(pn.Value)
		if sigil != "" {
			return nil
		}
		fm := NewTopFrame(ev, parse.Source{Name: "[purely-eval]"}, nil)
		ref := resolveVarRef(fm, qname, nil)
		if ref != nil {
			return deref(fm, ref).Get()
		}
	}
	return nil
}
