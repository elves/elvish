package eval

import (
	"strings"

	"src.elv.sh/pkg/parse"
)

func (ev *Evaler) PurelyEvalCompound(cn *parse.Compound) (string, bool) {
	return ev.PurelyEvalPartialCompound(cn, -1)
}

func (ev *Evaler) PurelyEvalPartialCompound(cn *parse.Compound, upto int) (string, bool) {
	tilde := false
	head := ""
	for _, in := range cn.Indexings {
		if len(in.Indices) > 0 {
			return "", false
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
				return "", false
			}
			v := ev.PurelyEvalPrimary(in.Head)
			if s, ok := v.(string); ok {
				head += s
			} else {
				return "", false
			}
		default:
			return "", false
		}
	}
	if tilde {
		i := strings.Index(head, "/")
		if i == -1 {
			i = len(head)
		}
		uname := head[:i]
		home, err := getHome(uname)
		if err != nil {
			return "", false
		}
		head = home + head[i:]
	}
	return head, true
}

// PurelyEvalPrimary evaluates a primary node without causing any side effects.
// If this cannot be done, it returns nil.
//
// Currently, only string literals and variables with no @ can be evaluated.
func (ev *Evaler) PurelyEvalPrimary(pn *parse.Primary) any {
	switch pn.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return pn.Value
	case parse.Variable:
		sigil, qname := SplitSigil(pn.Value)
		if sigil != "" {
			return nil
		}
		fm := &Frame{Evaler: ev, local: ev.Global(), up: new(Ns)}
		ref := resolveVarRef(fm, qname, nil)
		if ref != nil {
			variable := deref(fm, ref)
			if variable == nil {
				return nil
			}
			return variable.Get()
		}
	}
	return nil
}
