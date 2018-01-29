package eval

import (
	"errors"
	"strings"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

var ErrImpure = errors.New("expression is impure")

func PurelyEvalCompound(cn *parse.Compound) (string, error) {
	return (*Evaler)(nil).PurelyEvalCompound(cn)
}

func (ev *Evaler) PurelyEvalCompound(cn *parse.Compound) (string, error) {
	return ev.PurelyEvalPartialCompound(cn, nil)
}

func (ev *Evaler) PurelyEvalPartialCompound(cn *parse.Compound, upto *parse.Indexing) (string, error) {
	tilde := false
	head := ""
	for _, in := range cn.Indexings {
		if len(in.Indicies) > 0 {
			return "", ErrImpure
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

		if in == upto {
			break
		}
	}
	if tilde {
		i := strings.Index(head, "/")
		if i == -1 {
			i = len(head)
		}
		uname := head[:i]
		home, err := util.GetHome(uname)
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
		explode, ns, name := ParseVariableRef(pn.Value)
		if explode {
			return nil
		}
		ec := NewTopFrame(ev, NewInternalSource("[purely eval]"), nil)
		variable := ec.ResolveVar(ns, name)
		if variable != nil {
			return variable.Get()
		}
	}
	return nil
}
