package eval

// Implementation of var/set special forms.

import "github.com/xiaq/elvish/parse"

type varSetForm struct {
	names  []string
	values []*parse.TermNode
}

// checkVarSet checks a var or set special form.
//
// The arguments in the var/set special form must consist of zero or more
// variable factors followed by `=` and then zero or more terms. The number of
// values the terms evaluate to must be equal to the number of names, but
// checkVarSet does not attempt to check this.
func checkVarSet(ch *Checker, args *parse.TermListNode) *varSetForm {
	f := &varSetForm{}
	for i, n := range args.Nodes {
		termReq := "must be a variable name or `=`"
		if len(n.Nodes) != 1 {
			ch.errorf(n, "%s", termReq)
		}
		nf := n.Nodes[0]

		var text string
		if m, ok := nf.Node.(*parse.StringNode); ok {
			text = m.Text
		} else {
			ch.errorf(n, "%s", termReq)
		}

		if nf.Typ == parse.StringFactor && text == "=" {
			f.values = args.Nodes[i+1:]
			break
		} else if nf.Typ == parse.VariableFactor {
			f.names = append(f.names, text)
		} else {
			ch.errorf(n, "%s", termReq)
		}
	}
	ch.checkTerms(f.values)
	return f
}

func checkVar(ch *Checker, fn *parse.FormNode) interface{} {
	f := checkVarSet(ch, fn.Args)
	for _, name := range f.names {
		ch.pushVar(name)
	}
	return f
}

func checkSet(ch *Checker, fn *parse.FormNode) interface{} {
	return checkVarSet(ch, fn.Args)
}

func doSet(ev *Evaluator, names []string, values []Value) string {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(names) != len(values) {
		return "arity mismatch"
	}

	for i, name := range names {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		ev.locals[name] = values[i]
	}

	return ""
}

func var_(ev *Evaluator, a *formAnnotation, ports [2]*port) string {
	f := a.specialAnnotation.(*varSetForm)
	for _, name := range f.names {
		ev.locals[name] = NewString("")
	}
	if f.values != nil {
		return doSet(ev, f.names, ev.evalTermList(
			&parse.TermListNode{0, f.values}))
	}
	return ""
}

func set(ev *Evaluator, a *formAnnotation, ports [2]*port) string {
	f := a.specialAnnotation.(*varSetForm)
	if f.values == nil {
		return "not implemented"
	}
	return doSet(ev, f.names, ev.evalTermList(
		&parse.TermListNode{0, f.values}))
}
