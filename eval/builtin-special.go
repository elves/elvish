package eval

// Builtin special forms.

import "github.com/elves/elvish/parse"

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

func checkSetType(cp *Compiler, names []string, values []*parse.CompoundNode, vop valuesOp, p parse.Pos) {
	if !vop.tr.mayCountTo(len(names)) {
		cp.errorf(p, "number of variables doesn't match that of values")
	}
	_, more := vop.tr.count()
	if more {
		// TODO Try to check soundness to some extent
		return
	}
	for i, name := range names {
		t := vop.tr[i].t
		if _, ok := t.(AnyType); ok {
			// TODO Check type soundness at runtime
			continue
		}
		if cp.ResolveVar(name) != t {
			cp.errorf(values[i].Pos, "type mismatch")
		}
	}
}

// ensure that a CompoundNode contains exactly one PrimaryNode.
func ensurePrimary(cp *Compiler, cn *parse.CompoundNode, msg string) *parse.PrimaryNode {
	if len(cn.Nodes) != 1 || cn.Nodes[0].Right != nil {
		cp.errorf(cn.Pos, msg)
	}
	return cn.Nodes[0].Left
}

// ensure that a CompoundNode contains exactly one PrimaryNode of type
// VariablePrimary or StringPrimary
func ensureVariableOrStringPrimary(cp *Compiler, cn *parse.CompoundNode, msg string) (*parse.PrimaryNode, string) {
	pn := ensurePrimary(cp, cn, msg)
	switch pn.Typ {
	case parse.VariablePrimary, parse.StringPrimary:
		return pn, pn.Node.(*parse.StringNode).Text
	default:
		cp.errorf(cn.Pos, msg)
		return nil, ""
	}
}

// ensure the first compound of the form is a VariablePrimary. This is merely
// for better error messages; No actual processing is done.
func ensureStartWithVariable(cp *Compiler, fn *parse.FormNode, form string) {
	if len(fn.Args.Nodes) == 0 {
		cp.errorf(fn.Pos, "expect variable after %s", form)
	}
	expect := "expect variable"
	pn := ensurePrimary(cp, fn.Args.Nodes[0], expect)
	if pn.Typ != parse.VariablePrimary {
		cp.errorf(pn.Pos, expect)
	}
}

const (
	varArg0Req          = "must be either a variable or a table"
	varArg0ReqMultiElem = "must be either a variable or a string referring to a type"
	varArg1ReqMulti     = "must be a table with no dict part"
	varArg1ReqSingle    = "must be a string referring to a type"

	setArg0Req          = varArg0Req
	setArg0ReqMultiElem = "must be a variable"
	setArg1ReqMulti     = varArg1ReqMulti
)

// An invocation of the var special form looks like:
//
// VarForm    = 'var' { VarGroup } [ { VariablePrimary } ] [ Assignment ]
// VarGroup   = { VariablePrimary } StringPrimary
// Assignment = '=' { Compound }
//
// Variables in the same VarGroup has the type specified by the StringPrimary.
// Trailing variables have type Any. For instance,
//
// var $u $v Type1 $x $y Type2 $z = a b c d e
//
// gives $u and $v type Type1, $x $y type Type2 and $z type Any and
// assigns them the values a, b, c, d, e respectively.
func compileVar(cp *Compiler, fn *parse.FormNode) strOp {
	var (
		names  []string
		types  []Type
		values []*parse.CompoundNode
	)

	ensureStartWithVariable(cp, fn, "var")

	for i, cn := range fn.Args.Nodes {
		expect := "expect variable, type or equal sign"
		pn, text := ensureVariableOrStringPrimary(cp, cn, expect)
		if pn.Typ == parse.VariablePrimary {
			names = append(names, text)
		} else {
			if text == "=" {
				values = fn.Args.Nodes[i+1:]
				break
			} else {
				if t, ok := typenames[text]; !ok {
					cp.errorf(pn.Pos, "%v is not a valid type name", text)
				} else {
					if len(names) == len(types) {
						cp.errorf(pn.Pos, "duplicate type")
					}
					for i := len(types); i < len(names); i++ {
						types = append(types, t)
					}
				}
			}
		}
	}

	for i := len(types); i < len(names); i++ {
		types = append(types, AnyType{})
	}

	for i, name := range names {
		cp.pushVar(name, types[i])
	}

	var vop valuesOp
	if values != nil {
		vop = cp.compileCompounds(values)
		checkSetType(cp, names, values, vop, fn.Pos)
	}
	return func(ev *Evaluator) string {
		for i, name := range names {
			ev.scope[name] = valuePtr(types[i].Default())
		}
		if vop.f != nil {
			return doSet(ev, names, vop.f(ev))
		}
		return ""
	}
}

// An invocation of the set special form looks like:
//
// SetForm = 'set' { VariablePrimary } '=' { Compound }
func compileSet(cp *Compiler, fn *parse.FormNode) strOp {
	var (
		names  []string
		values []*parse.CompoundNode
	)

	ensureStartWithVariable(cp, fn, "set")

	for i, cn := range fn.Args.Nodes {
		expect := "expect variable or equal sign"
		pn, text := ensureVariableOrStringPrimary(cp, cn, expect)
		if pn.Typ == parse.VariablePrimary {
			names = append(names, text)
		} else {
			if text != "=" {
				cp.errorf(pn.Pos, expect)
			}
			values = fn.Args.Nodes[i+1:]
			break
		}
	}

	var vop valuesOp
	vop = cp.compileCompounds(values)
	checkSetType(cp, names, values, vop, fn.Pos)

	return func(ev *Evaluator) string {
		return doSet(ev, names, vop.f(ev))
	}
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
	var names []string
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
		names = append(names, name)
	}
	return func(ev *Evaluator) string {
		for _, name := range names {
			delete(ev.scope, name)
		}
		return ""
	}
}
