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

func mustSinglePrimary(cp *Compiler, cn *parse.CompoundNode, msg string) *parse.PrimaryNode {
	if len(cn.Nodes) != 1 || cn.Nodes[0].Right != nil {
		cp.errorf(cn.Pos, msg)
	}
	return cn.Nodes[0].Left
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

// The var special form can take any of the following forms:
// var [$u $v type1 $x type2 ...]
// var [$u $v type1 $x type2 ...] [value1 value2 ...]
// var $v type       (short for var [$v type])
// var $v type value (short for var [$v type] [value])
func compileVar(cp *Compiler, fn *parse.FormNode) strOp {
	var (
		names  []string
		types  []Type
		values []*parse.CompoundNode
	)

	args := fn.Args
	if len(args.Nodes) == 0 {
		cp.errorf(fn.Pos, "empty var form")
	}

	p0 := mustSinglePrimary(cp, args.Nodes[0], varArg0Req)

	switch p0.Typ {
	case parse.VariablePrimary:
		if len(args.Nodes) < 2 {
			// TODO Identify the end of args.Nodes[0]
			cp.errorf(args.Nodes[0].Pos, "must be followed by type")
		}
		if len(args.Nodes) > 3 {
			cp.errorf(args.Nodes[3].Pos, "too many arguments")
		}

		names = []string{p0.Node.(*parse.StringNode).Text}

		p1 := mustSinglePrimary(cp, args.Nodes[1], varArg1ReqSingle)
		if p1.Typ != parse.StringPrimary {
			cp.errorf(p1.Pos, varArg1ReqSingle)
		}
		p1s := p1.Node.(*parse.StringNode).Text
		if t, ok := typenames[p1s]; !ok {
			cp.errorf(p1.Pos, varArg1ReqSingle)
		} else {
			types = []Type{t}
		}

		if len(args.Nodes) == 3 {
			values = []*parse.CompoundNode{args.Nodes[2]}
		}
	case parse.TablePrimary:
		if len(args.Nodes) > 2 {
			cp.errorf(args.Nodes[3].Pos, "too many arguments")
		} else if len(args.Nodes) == 2 {
			p1 := mustSinglePrimary(cp, args.Nodes[1], varArg1ReqMulti)
			if p1.Typ != parse.TablePrimary {
				cp.errorf(p1.Pos, varArg1ReqMulti)
			}
			t1 := p1.Node.(*parse.TableNode)
			if len(t1.Dict) > 0 {
				cp.errorf(t1.Pos, varArg1ReqMulti)
			}
			values = append(values, t1.List...)
		}

		t0 := p0.Node.(*parse.TableNode)
		if len(t0.Dict) > 0 {
			cp.errorf(t0.Pos, "must not contain dict part")
		}
		var firstUntyped *parse.PrimaryNode
		for _, cn := range t0.List {
			p := mustSinglePrimary(cp, cn, varArg0ReqMultiElem)
			switch p.Typ {
			case parse.StringPrimary:
				ps := p.Node.(*parse.StringNode).Text
				if t, ok := typenames[ps]; !ok {
					cp.errorf(p.Pos, varArg0ReqMultiElem)
				} else {
					if len(names) == 0 {
						cp.errorf(p.Pos, "first element must be variable")
					} else if len(names) == len(types) {
						cp.errorf(p.Pos, "duplicate type")
					}
					for i := len(types); i < len(names); i++ {
						types = append(types, t)
					}
					firstUntyped = nil
				}
			case parse.VariablePrimary:
				if firstUntyped == nil {
					firstUntyped = p
				}
				names = append(names, p.Node.(*parse.StringNode).Text)
			default:
				cp.errorf(p.Pos, varArg0ReqMultiElem)
			}
		}
		if len(types) < len(names) {
			cp.errorf(firstUntyped.Pos, "variables from here lack type")
		}
	default:
		cp.errorf(p0.Pos, varArg0Req)
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

// The set special form can take any of the following forms:
// set [$u $v ...] [value1 value2 ...]
// var $v value (short for set [$v] [value])
func compileSet(cp *Compiler, fn *parse.FormNode) strOp {
	var (
		names  []string
		values []*parse.CompoundNode
	)

	args := fn.Args
	if len(args.Nodes) == 0 {
		cp.errorf(fn.Pos, "empty var form")
	} else if len(args.Nodes) == 1 {
		// TODO Identify the end of args.Nodes[0]
		cp.errorf(args.Nodes[0].Pos, "must be followed by value argument")
	} else if len(args.Nodes) > 2 {
		cp.errorf(args.Nodes[2].Pos, "too many arguments")
	}

	p0 := mustSinglePrimary(cp, args.Nodes[0], setArg0Req)

	switch p0.Typ {
	case parse.VariablePrimary:
		names = []string{p0.Node.(*parse.StringNode).Text}
		values = []*parse.CompoundNode{args.Nodes[1]}
	case parse.TablePrimary:
		t0 := p0.Node.(*parse.TableNode)
		if len(t0.Dict) > 0 {
			cp.errorf(t0.Pos, "must not contain dict part")
		}
		for _, cn := range t0.List {
			p := mustSinglePrimary(cp, cn, setArg0ReqMultiElem)
			switch p.Typ {
			case parse.VariablePrimary:
				names = append(names, p.Node.(*parse.StringNode).Text)
			default:
				cp.errorf(p.Pos, setArg0ReqMultiElem)
			}
		}

		p1 := mustSinglePrimary(cp, args.Nodes[1], setArg1ReqMulti)
		if p1.Typ != parse.TablePrimary {
			cp.errorf(p1.Pos, setArg1ReqMulti)
		}
		t1 := p1.Node.(*parse.TableNode)
		if len(t1.Dict) > 0 {
			cp.errorf(t1.Pos, setArg1ReqMulti)
		}
		values = append(values, t1.List...)
	default:
		cp.errorf(p0.Pos, setArg0Req)
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
