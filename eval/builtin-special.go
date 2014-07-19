package eval

// Builtin special forms.

import "github.com/xiaq/elvish/parse"

type strOp func(*Evaluator) string
type builtinSpecialCompile func(*Compiler, *parse.FormNode) strOp

type builtinSpecial struct {
	compile     builtinSpecialCompile
	streamTypes [2]StreamType
}

var builtinSpecials map[string]builtinSpecial

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]builtinSpecial{
		"var": builtinSpecial{compileVar, [2]StreamType{}},
		"set": builtinSpecial{compileSet, [2]StreamType{}},
		"del": builtinSpecial{compileDel, [2]StreamType{}},
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

func checkSetType(cp *Compiler, args *parse.TermListNode, f *varSetForm, vop valuesOp) {
	if len(f.names) != len(vop.ts) {
		cp.errorf(args, "number of variables doesn't match that of values")
	}
	for i, name := range f.names {
		if _, ok := vop.ts[i].(AnyType); ok {
			// TODO Check type soundness at runtime
			continue
		}
		if cp.tryResolveVar(name) != vop.ts[i] {
			cp.errorf(f.values[i], "type mismatch")
		}
	}
}

// compileVarSet compiles a var or set special form. If v is true, a var special
// form is being compiled.
//
// The arguments in the var/set special form must consist of zero or more
// variable factors followed by `=` and then zero or more terms. The number of
// values the terms evaluate to must be equal to the number of names, but
// compileVarSet does not attempt to compile this.
func compileVarSet(cp *Compiler, args *parse.TermListNode, v bool) strOp {
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
			cp.errorf(n, "%s", termReq)
		}
		nf := n.Nodes[0]

		var text string
		if m, ok := nf.Node.(*parse.StringNode); ok {
			text = m.Text
		} else {
			cp.errorf(n, "%s", termReq)
		}

		if nf.Typ == parse.StringFactor {
			if text == "=" {
				f.values = args.Nodes[i+1:]
				break
			} else if t := typenames[text]; v && t != nil {
				if i == 0 {
					cp.errorf(n, "type name must follow variables")
				}
				for j := lastTyped; j < i; j++ {
					f.types = append(f.types, t)
				}
				lastTyped = i
			} else {
				cp.errorf(n, "%s", termReq)
			}
		} else if nf.Typ == parse.VariableFactor {
			if !v {
				// For set, ensure that the variable can be resolved
				cp.resolveVar(text, nf)
			}
			f.names = append(f.names, text)
		} else {
			cp.errorf(n, "%s", termReq)
		}
	}
	if v {
		if len(f.types) != len(f.names) {
			cp.errorf(args, "Some variables lack type")
		}
		for i, name := range f.names {
			cp.pushVar(name, f.types[i])
		}
		var vop valuesOp
		if f.values != nil {
			vop = cp.compileTerms(f.values)
			checkSetType(cp, args, f, vop)
		}
		return func(ev *Evaluator) string {
			for i, name := range f.names {
				ev.scope[name] = valuePtr(f.types[i].Default())
			}
			if vop.f != nil {
				return doSet(ev, f.names, vop.f(ev))
			}
			return ""
		}
	} else {
		if f.values == nil {
			cp.errorf(args, "set form lacks equal sign")
		}
		vop := cp.compileTerms(f.values)
		checkSetType(cp, args, f, vop)
		return func(ev *Evaluator) string {
			return doSet(ev, f.names, vop.f(ev))
		}
	}
}

func compileVar(cp *Compiler, fn *parse.FormNode) strOp {
	return compileVarSet(cp, fn.Args, true)
}

func compileSet(cp *Compiler, fn *parse.FormNode) strOp {
	return compileVarSet(cp, fn.Args, false)
}

func doSet(ev *Evaluator, names []string, values []Value) string {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(names) != len(values) {
		return "arity mismatch"
	}

	for i, name := range names {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		*ev.scope[name] = values[i]
	}

	return ""
}

func compileDel(cp *Compiler, fn *parse.FormNode) strOp {
	// Do conventional compiling of all terms, including ensuring that
	// variables can be resolved
	f := &delForm{}
	for _, n := range fn.Args.Nodes {
		termReq := "must be a varible"
		if len(n.Nodes) != 1 {
			cp.errorf(n, "%s", termReq)
		}
		nf := n.Nodes[0]
		if nf.Typ != parse.VariableFactor {
			cp.errorf(n, "%s", termReq)
		}
		name := nf.Node.(*parse.StringNode).Text
		cp.resolveVar(name, nf)
		if !cp.hasVarOnThisScope(name) {
			cp.errorf(n, "can only delete variable on current scope")
		}
		cp.popVar(name)
		f.names = append(f.names, name)
	}
	return func(ev *Evaluator) string {
		for _, name := range f.names {
			delete(ev.scope, name)
		}
		return ""
	}
}
