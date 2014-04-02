package eval

// Builtin special forms.

import "github.com/xiaq/elvish/parse"

type builtinSpecialImpl func(*Evaluator, *formAnnotation, [2]*port) string
type builtinSpecialCheck func(*Checker, *parse.FormNode) interface{}

type builtinSpecial struct {
	fn          builtinSpecialImpl
	check       builtinSpecialCheck
	streamTypes [2]StreamType
}

var builtinSpecials map[string]builtinSpecial

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]builtinSpecial{
		"var": builtinSpecial{var_, checkVar, [2]StreamType{}},
		"set": builtinSpecial{set, checkSet, [2]StreamType{}},
		"del": builtinSpecial{del, checkDel, [2]StreamType{}},
	}
}

type varSetForm struct {
	names  []string
	types  []Type
	values []*parse.TermNode
}

type delForm struct {
	names []string
}

// checkVarSet checks a var or set special form. If v is true, a var special
// form is being checked.
//
// The arguments in the var/set special form must consist of zero or more
// variable factors followed by `=` and then zero or more terms. The number of
// values the terms evaluate to must be equal to the number of names, but
// checkVarSet does not attempt to check this.
func checkVarSet(ch *Checker, args *parse.TermListNode, v bool) *varSetForm {
	f := &varSetForm{}
	lastTyped := 0
	for i, n := range args.Nodes {
		termReq := ""
		if v {
			termReq = "must be a variable, literal type name or literal `=`"
		} else {
			termReq = "must be a variable or literal `=`"
		}
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

		if nf.Typ == parse.StringFactor {
			if text == "=" {
				f.values = args.Nodes[i+1:]
				break
			} else if t := typenames[text]; v && t != nil {
				if i == 0 {
					ch.errorf(n, "type name must follow variables")
				}
				for j := lastTyped; j < i; j++ {
					f.types = append(f.types, t)
				}
				lastTyped = i
			} else {
				ch.errorf(n, "%s", termReq)
			}
		} else if nf.Typ == parse.VariableFactor {
			f.names = append(f.names, text)
		} else {
			ch.errorf(n, "%s", termReq)
		}
	}
	if v {
		if len(f.types) != len(f.names) {
			ch.errorf(args, "Some variables lack type")
		}
		// Skip checking of variable terms
		ch.checkTerms(f.values)
	} else {
		// Do conventional checking of all terms, including ensuring that
		// variables can be resolved
		ch.checkTermList(args)
	}
	return f
}

func checkVar(ch *Checker, fn *parse.FormNode) interface{} {
	f := checkVarSet(ch, fn.Args, true)
	for i, name := range f.names {
		ch.pushVar(name, f.types[i])
	}
	return f
}

func checkSet(ch *Checker, fn *parse.FormNode) interface{} {
	return checkVarSet(ch, fn.Args, false)
}

func doSet(ev *Evaluator, names []string, values []Value) string {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(names) != len(values) {
		return "arity mismatch"
	}

	for i, name := range names {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		*ev.resolveVar(name) = values[i]
	}

	return ""
}

func var_(ev *Evaluator, a *formAnnotation, ports [2]*port) string {
	f := a.specialAnnotation.(*varSetForm)
	for i, name := range f.names {
		ev.scope[name] = valuePtr(f.types[i].Default())
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

func checkDel(ch *Checker, fn *parse.FormNode) interface{} {
	// Do conventional checking of all terms, including ensuring that
	// variables can be resolved
	ch.checkTermList(fn.Args)
	f := &delForm{}
	for _, n := range fn.Args.Nodes {
		termReq := "must be a varible"
		if len(n.Nodes) != 1 {
			ch.errorf(n, "%s", termReq)
		}
		nf := n.Nodes[0]
		if nf.Typ != parse.VariableFactor {
			ch.errorf(n, "%s", termReq)
		}
		name := nf.Node.(*parse.StringNode).Text
		if !ch.hasVarOnThisScope(name) {
			ch.errorf(n, "can only delete variable on current scope")
		}
		ch.popVar(name)
		f.names = append(f.names, name)
	}
	return f
}

func del(ev *Evaluator, a *formAnnotation, ports [2]*port) string {
	f := a.specialAnnotation.(*delForm)
	for _, name := range f.names {
		delete(ev.scope, name)
	}
	return ""
}
