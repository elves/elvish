// Package nodeutil provides utilities for inspecting the AST.
package nodeutil

import (
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[nodeutil] ")

func SimpleCompound(cn *parse.Compound, upto *parse.Indexing, ev *eval.Evaler) (bool, string, error) {
	tilde := false
	head := ""
	for _, in := range cn.Indexings {
		if len(in.Indicies) > 0 {
			return false, "", nil
		}
		switch in.Head.Type {
		case parse.Tilde:
			tilde = true
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			head += in.Head.Value
		case parse.Variable:
			if ev == nil {
				return false, "", nil
			}
			v := PurelyEvalPrimary(in.Head, ev)
			if s, ok := v.(eval.String); ok {
				head += string(s)
			} else {
				return false, "", nil
			}
		default:
			return false, "", nil
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
			return false, "", err
		}
		head = home + head[i:]
	}
	return true, head, nil
}

// PurelyEvalPrimary evaluates a primary node without causing any side effects.
// If this cannot be done, it returns nil.
//
// Currently, only string literals and variables with no @ can be evaluated.
func PurelyEvalPrimary(pn *parse.Primary, ev *eval.Evaler) eval.Value {
	switch pn.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return eval.String(pn.Value)
	case parse.Variable:
		explode, ns, name := eval.ParseVariable(pn.Value)
		if explode {
			return nil
		}
		ec := eval.NewTopEvalCtx(ev, "[pure eval]", "", nil)
		variable := ec.ResolveVar(ns, name)
		if variable != nil {
			return variable.Get()
		}
	}
	return nil
}
