// +build ignore

package eval

// Builtin special forms.

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/elves/elvish/parse"
)

type compileBuiltin func(*compiler, *parse.Form) exitusOp

var builtinSpecials map[string]compileBuiltin
var BuiltinSpecialNames []string

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]compileBuiltin{
		"var": compileVar,
		"set": compileSet,
		"del": compileDel,

		"use": compileUse,

		"fn": compileFn,
		// "if": builtinSpecial{compileIf},
	}
	for k, _ := range builtinSpecials {
		BuiltinSpecialNames = append(BuiltinSpecialNames, k)
	}
}

// ensureStartWithVariabl ensures the first compound of the form is a
// VariablePrimary. This is merely for better error messages; No actual
// processing is done.
func ensureStartWithVariable(cp *compiler, fn *parse.Form, form string) {
	if len(fn.Args.Nodes) == 0 {
		cp.errorf(fn.Pos, "expect variable after %s", form)
	}
	ensureVariablePrimary(cp, fn.Args.Nodes[0], "expect variable")
}

// VarForm    = 'var' { StringPrimary } [ '=' Compound ]
func compileVar(cp *compiler, fn *parse.Form) exitusOp {
	var (
		names  []string
		types  []Type
		values []*parse.Compound
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
		types = append(types, anyType{})
	}

	for i, name := range names {
		cp.pushVar(name, types[i])
	}

	var vop valuesOp
	if values != nil {
		vop = cp.compounds(values)
		checkSetType(cp, names, values, vop, fn.Pos)
	}
	return func(ec *evalCtx) exitus {
		for i, name := range names {
			ec.local[name] = newInternalVariable(types[i].Default(), types[i])
		}
		if vop.f != nil {
			return doSet(ec, names, vop.f(ec))
		}
		return ok
	}
}

// SetForm = 'set' { VariablePrimary } '=' { Compound }
func compileSet(cp *compiler, fn *parse.Form) exitusOp {
	var (
		names  []string
		values []*parse.Compound
	)

	ensureStartWithVariable(cp, fn, "set")

	for i, cn := range fn.Args.Nodes {
		expect := "expect variable or equal sign"
		pn, text := ensureVariableOrStringPrimary(cp, cn, expect)
		if pn.Typ == parse.VariablePrimary {
			ns, name := splitQualifiedName(text)
			cp.mustResolveVar(ns, name, cn.Pos)
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
	vop = cp.compounds(values)
	checkSetType(cp, names, values, vop, fn.Pos)

	return func(ec *evalCtx) exitus {
		return doSet(ec, names, vop.f(ec))
	}
}

var (
	arityMismatch = newFailure("arity mismatch")
)

func doSet(ec *evalCtx, names []string, values []Value) exitus {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(names) != len(values) {
		return arityMismatch
	}

	for i, name := range names {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		variable := ec.ResolveVar(splitQualifiedName(name))
		if variable == nil {
			return newFailure(fmt.Sprintf("variable $%s not found; the compiler has a bug", name))
		}
		tvar := variable.StaticType()
		tval := values[i].Type()
		if !mayAssign(tvar, tval) {
			return typeMismatch
		}
		variable.Set(values[i])
	}

	return ok
}

// DelForm = 'del' { VariablePrimary }
func compileDel(cp *compiler, fn *parse.Form) exitusOp {
	// Do conventional compiling of all compound expressions, including
	// ensuring that variables can be resolved
	var names, envNames []string
	for _, cn := range fn.Args.Nodes {
		_, qname := ensureVariablePrimary(cp, cn, "expect variable")
		ns, name := splitQualifiedName(qname)
		switch ns {
		case "", "local":
			if cp.resolveVarOnThisScope(name) == nil {
				cp.errorf(cn.Pos, "variable $%s not found on current local scope", name)
			}
			cp.popVar(name)
			names = append(names, name)
		case "env":
			envNames = append(envNames, name)
		default:
			cp.errorf(cn.Pos, "can only delete a variable in local: or env: namespace")
		}

	}
	return func(ec *evalCtx) exitus {
		for _, name := range names {
			delete(ec.local, name)
		}
		for _, name := range envNames {
			// BUG(xiaq): We rely on the fact that os.Unsetenv always returns
			// nil.
			os.Unsetenv(name)
		}
		return ok
	}
}

func stem(fname string) string {
	base := path.Base(fname)
	ext := path.Ext(base)
	return base[0 : len(base)-len(ext)]
}

// UseForm = 'use' StringPrimary.modname Primary.fname
//         = 'use' StringPrimary.fname
func compileUse(cp *compiler, fn *parse.Form) exitusOp {
	var fnameNode *parse.Compound
	var fname, modname string

	switch len(fn.Args.Nodes) {
	case 0:
		cp.errorf(fn.Args.Pos, "expect module name or file name")
	case 1, 2:
		fnameNode = fn.Args.Nodes[0]
		_, fname = ensureStringPrimary(cp, fnameNode, "expect string literal")
		if len(fn.Args.Nodes) == 2 {
			modnameNode := fn.Args.Nodes[1]
			_, modname = ensureStringPrimary(
				cp, modnameNode, "expect string literal")
			if modname == "" {
				cp.errorf(modnameNode.Pos, "module name is empty")
			}
		} else {
			modname = stem(fname)
			if modname == "" {
				cp.errorf(fnameNode.Pos, "stem of file name is empty")
			}
		}
	default:
		cp.errorf(fn.Args.Nodes[2].Pos, "superfluous argument")
	}
	switch {
	case strings.HasPrefix(fname, "/"):
		// Absolute file name, do nothing
	case strings.HasPrefix(fname, "./") || strings.HasPrefix(fname, "../"):
		// File name relative to current source
		fname = path.Clean(path.Join(cp.dir, fname))
	default:
		// File name relative to data dir
		fname = path.Clean(path.Join(cp.dataDir, fname))
	}
	src, err := readFileUTF8(fname)
	if err != nil {
		cp.errorf(fnameNode.Pos, "cannot read module: %s", err.Error())
	}

	cn, err := parse.Parse(fname, src)
	if err != nil {
		// TODO(xiaq): Pretty print
		cp.errorf(fnameNode.Pos, "cannot parse module: %s", err.Error())
	}

	newCc := &compiler{
		cp.Compiler,
		fname, src, path.Dir(fname),
		[]staticNS{staticNS{}}, staticNS{},
	}

	op, err := newCc.compile(cn)
	if err != nil {
		// TODO(xiaq): Pretty print
		cp.errorf(fnameNode.Pos, "cannot compile module: %s", err.Error())
	}

	cp.mod[modname] = newCc.scopes[0]

	return func(ec *evalCtx) exitus {
		// TODO(xiaq): Should handle failures when evaluting the module
		newEc := &evalCtx{
			ec.Evaler,
			fname, src, "module " + modname,
			ns{}, ns{},
			ec.ports,
		}
		op.f(newEc)
		ec.mod[modname] = newEc.local
		return ok
	}
}

// makeFnOp wraps a valuesOp such that a return is converted to an ok.
func makeFnOp(op valuesOp) valuesOp {
	f := func(ec *evalCtx) []Value {
		vs := op.f(ec)
		if len(vs) == 1 {
			if e, ok := vs[0].(exitus); ok {
				if e.Sort == Return {
					return []Value{newFlowExitus(Ok)}
				}
			}
		}
		return vs
	}
	return valuesOp{op.tr, f}
}

// FnForm = 'fn' StringPrimary { VariablePrimary } ClosurePrimary
//
// fn defines a function. This isn't strictly needed, since user-defined
// functions are just variables. The following two lines should be exactly
// equivalent:
//
// fn f $a $b { put (* $a $b) (/ $a *b) }
// var $fn-f = { |$a $b| put (* $a $b) (/ $a $b) }
func compileFn(cp *compiler, fn *parse.Form) exitusOp {
	if len(fn.Args.Nodes) == 0 {
		cp.errorf(fn.Pos, "expect function name after fn")
	}
	_, fnName := ensureStringPrimary(cp, fn.Args.Nodes[0], "expect string literal")
	varName := fnPrefix + fnName

	var closureNode *parse.Closure
	var argNames []*parse.Compound

	for i, cn := range fn.Args.Nodes[1:] {
		expect := "expect variable or closure"
		pn := ensurePrimary(cp, cn, expect)
		switch pn.Typ {
		case parse.ClosurePrimary:
			if i+2 != len(fn.Args.Nodes) {
				cp.errorf(fn.Args.Nodes[i+2].Pos, "garbage after closure literal")
			}
			closureNode = pn.Node.(*parse.Closure)
			break
		case parse.VariablePrimary:
			argNames = append(argNames, cn)
		default:
			cp.errorf(pn.Pos, expect)
		}
	}

	if len(argNames) > 0 {
		closureNode = &parse.Closure{
			closureNode.Pos,
			&parse.Spaced{argNames[0].Pos, argNames},
			closureNode.Chunk,
		}
	}

	op := cp.closure(closureNode)

	cp.pushVar(varName, callableType{})

	return func(ec *evalCtx) exitus {
		closure := op.f(ec)[0].(*closure)
		closure.Op = makeFnOp(closure.Op)
		ec.local[varName] = newInternalVariable(closure, callableType{})
		return ok
	}
}

func maybeClosurePrimary(cn *parse.Compound) (*parse.Closure, bool) {
	if len(cn.Nodes) == 1 && cn.Nodes[0].Right == nil && cn.Nodes[0].Left.Typ == parse.ClosurePrimary {
		return cn.Nodes[0].Left.Node.(*parse.Closure), true
	}
	return nil, false
}

func maybeStringPrimary(cn *parse.Compound) (string, bool) {
	if len(cn.Nodes) == 1 && cn.Nodes[0].Right == nil && cn.Nodes[0].Left.Typ == parse.StringPrimary {
		return cn.Nodes[0].Left.Node.(*parse.String).Text, true
	}
	return "", false
}

/*
type ifBranch struct {
	condition valuesOp
	body      valuesOp
}

// IfForm = 'if' Branch { 'else' 'if' Branch } [ 'else' Branch ]
// Branch = SpacedNode.condition ClosurePrimary.body
//
// The condition part of a Branch ends as soon as a Compound of a single
// ClosurePrimary is encountered.
func compileIf(cp *compiler, fn *parse.Form) exitusOp {
	compounds := fn.Args.Nodes
	var branches []*ifBranch

	nextBranch := func() {
		var conds []*parse.Compound
		for i, cn := range compounds {
			if closure, ok := maybeClosurePrimary(cn); ok {
				if i == 0 {
					cp.errorf(cn.Pos, "expect condition")
				}
				condition := cp.compounds(conds)
				if closure.ArgNames != nil && len(closure.ArgNames.Nodes) > 0 {
					cp.errorf(closure.ArgNames.Pos, "unexpected arguments")
				}
				body := cp.closure(closure)
				branches = append(branches, &ifBranch{condition, body})
				compounds = compounds[i+1:]
				return
			}
			conds = append(conds, cn)
		}
		cp.errorf(compounds[len(compounds)-1].Pos, "expect body after this")
	}
	// if branch
	nextBranch()
	// else-if branches
	for len(compounds) >= 2 {
		s1, _ := maybeStringPrimary(compounds[0])
		s2, _ := maybeStringPrimary(compounds[1])
		if s1 == "else" && s2 == "if" {
			compounds = compounds[2:]
			nextBranch()
		} else {
			break
		}
	}
	// else branch
	if len(compounds) > 0 {
		s, _ := maybeStringPrimary(compounds[0])
		if s == "else" {
			if len(compounds) == 1 {
				cp.errorf(compounds[0].Pos, "expect body after this")
			} else if len(compounds) > 2 {
				cp.errorf(compounds[2].Pos, "trailing garbage")
			}
			body, ok := maybeClosurePrimary(compounds[1])
			if !ok {
				cp.errorf(compounds[1].Pos, "expect body")
			}
			branches = append(branches, &ifBranch{
				literalValue(boolean(true)), cp.closure(body)})
		} else {
			cp.errorf(compounds[0].Pos, "trailing garbage")
		}
	}
	return func(ec *evalCtx) exitus {
		for _, ib := range branches {
			if allTrue(ib.condition.f(ec)) {
				f := ib.body.f(ec)[0].(*closure)
				su := f.Exec(ec.copy("closure of if"), []Value{})
				for _ = range su {
				}
				// TODO(xiaq): Return the exitus of the body
				return ok
			}
		}
		return ok
	}
}
*/
