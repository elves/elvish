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
	values []*parse.CompoundNode
}

type delForm struct {
	names []string
}

func checkSetType(cp *Compiler, args *parse.SpacedNode, f *varSetForm, vop valuesOp) {
	if !vop.tr.mayCountTo(len(f.names)) {
		cp.errorf(args.Pos, "number of variables doesn't match that of values")
	}
	_, more := vop.tr.count()
	if more {
		// TODO Try to check soundness to some extent
		return
	}
	for i, name := range f.names {
		t := vop.tr[i].t
		if _, ok := t.(AnyType); ok {
			// TODO Check type soundness at runtime
			continue
		}
		if cp.ResolveVar(name) != t {
			cp.errorf(f.values[i].Pos, "type mismatch")
		}
	}
}

// compileVarSet compiles a var or set special form. If v is true, a var special
// form is being compiled.
//
// The arguments in the var/set special form must consist of zero or more
// variables followed by `=` and then zero or more compound expressions. The
// number of values the compound expressions evaluate to must be equal to the
// number of variables, but compileVarSet does not attempt to check this.
func compileVarSet(cp *Compiler, args *parse.SpacedNode, v bool) strOp {
	f := &varSetForm{}
	lastTyped := 0
	for i, n := range args.Nodes {
		compoundReq := ""
		if v {
			compoundReq = "must be a variable, literal type name or literal `=`"
		} else {
			compoundReq = "must be a variable or literal `=`"
		}
		if len(n.Nodes) != 1 || n.Nodes[0].Right != nil {
			cp.errorf(n.Pos, "%s", compoundReq)
		}
		nf := n.Nodes[0].Left

		var text string
		if m, ok := nf.Node.(*parse.StringNode); ok {
			text = m.Text
		} else {
			cp.errorf(n.Pos, "%s", compoundReq)
		}

		if nf.Typ == parse.StringPrimary {
			if text == "=" {
				f.values = args.Nodes[i+1:]
				break
			} else if t := typenames[text]; v && t != nil {
				if i == 0 {
					cp.errorf(n.Pos, "type name must follow variables")
				}
				for j := lastTyped; j < i; j++ {
					f.types = append(f.types, t)
				}
				lastTyped = i
			} else {
				cp.errorf(n.Pos, "%s", compoundReq)
			}
		} else if nf.Typ == parse.VariablePrimary {
			if !v {
				// For set, ensure that the variable can be resolved
				cp.mustResolveVar(text, nf.Pos)
			}
			f.names = append(f.names, text)
		} else {
			cp.errorf(n.Pos, "%s", compoundReq)
		}
	}
	if v {
		if len(f.types) != len(f.names) {
			cp.errorf(args.Pos, "Some variables lack type")
		}
		for i, name := range f.names {
			cp.pushVar(name, f.types[i])
		}
		var vop valuesOp
		if f.values != nil {
			vop = cp.compileCompounds(f.values)
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
			cp.errorf(args.Pos, "set form lacks equal sign")
		}
		vop := cp.compileCompounds(f.values)
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
	// Do conventional compiling of all compound expressions, including
	// ensuring that variables can be resolved
	f := &delForm{}
	for _, n := range fn.Args.Nodes {
		compoundReq := "must be a varible"
		if len(n.Nodes) != 1 || n.Nodes[0].Right != nil {
			cp.errorf(n.Pos, "%s", compoundReq)
		}
		nf := n.Nodes[0].Left
		if nf.Typ != parse.VariablePrimary {
			cp.errorf(n.Pos, "%s", compoundReq)
		}
		name := nf.Node.(*parse.StringNode).Text
		cp.mustResolveVar(name, nf.Pos)
		if !cp.hasVarOnThisScope(name) {
			cp.errorf(n.Pos, "can only delete variable on current scope")
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
