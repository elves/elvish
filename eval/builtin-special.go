package eval

// Builtin special forms.

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/elves/elvish/parse"
)

type exitusOp func(*Evaluator) exitus
type builtinSpecialCompile func(*Compiler, *parse.FormNode) exitusOp

type builtinSpecial struct {
	compile builtinSpecialCompile
}

var builtinSpecials map[string]builtinSpecial

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]builtinSpecial{
		"var": builtinSpecial{compileVar},
		"set": builtinSpecial{compileSet},
		"del": builtinSpecial{compileDel},

		"use": builtinSpecial{compileUse},

		"fn": builtinSpecial{compileFn},
		"if": builtinSpecial{compileIf},

		"static-typeof": builtinSpecial{compileStaticTypeof},
	}
}

func mayAssign(tvar, tval Type) bool {
	if isAny(tval) || isAny(tvar) {
		return true
	}
	// XXX(xiaq) This is not how you check the equality of two interfaces. But
	// it happens to work when all the Type instances we have are empty
	// structs.
	return tval == tvar
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
		tval := vop.tr[i].t
		tvar := cp.ResolveVar(splitQualifiedName(name))
		if !mayAssign(tvar, tval) {
			cp.errorf(values[i].Pos, "type mismatch: assigning %#v value to %#v variable", tval, tvar)
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

// ensureVariableOrStringPrimary ensures that a CompoundNode contains exactly
// one PrimaryNode of type VariablePrimary or StringPrimary.
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

// ensureVariablePrimary ensures that a CompoundNode contains exactly one
// PrimaryNode of type VariablePrimary.
func ensureVariablePrimary(cp *Compiler, cn *parse.CompoundNode, msg string) (*parse.PrimaryNode, string) {
	pn, text := ensureVariableOrStringPrimary(cp, cn, msg)
	if pn.Typ != parse.VariablePrimary {
		cp.errorf(pn.Pos, msg)
	}
	return pn, text
}

// ensureStringPrimary ensures that a CompoundNode contains exactly one
// PrimaryNode of type VariablePrimary.
func ensureStringPrimary(cp *Compiler, cn *parse.CompoundNode, msg string) (*parse.PrimaryNode, string) {
	pn, text := ensureVariableOrStringPrimary(cp, cn, msg)
	if pn.Typ != parse.StringPrimary {
		cp.errorf(pn.Pos, msg)
	}
	return pn, text
}

// ensureStartWithVariabl ensures the first compound of the form is a
// VariablePrimary. This is merely for better error messages; No actual
// processing is done.
func ensureStartWithVariable(cp *Compiler, fn *parse.FormNode, form string) {
	if len(fn.Args.Nodes) == 0 {
		cp.errorf(fn.Pos, "expect variable after %s", form)
	}
	ensureVariablePrimary(cp, fn.Args.Nodes[0], "expect variable")
}

// VarForm    = 'var' { VarGroup } [ '=' Compound ]
// VarGroup   = { VariablePrimary } [ StringPrimary ]
//
// Variables in the same VarGroup has the type specified by the StringPrimary.
// Only in the last VarGroup the StringPrimary may be omitted, in which case it
// defaults to "any". For instance,
//
// var $u $v Type1 $x $y Type2 $z = a b c d e
//
// gives $u and $v type Type1, $x $y type Type2 and $z type Any and
// assigns them the values a, b, c, d, e respectively.
func compileVar(cp *Compiler, fn *parse.FormNode) exitusOp {
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
	return func(ev *Evaluator) exitus {
		for i, name := range names {
			ev.local[name] = newInternalVariable(types[i].Default(), types[i])
		}
		if vop.f != nil {
			return doSet(ev, names, vop.f(ev))
		}
		return success
	}
}

// SetForm = 'set' { VariablePrimary } '=' { Compound }
func compileSet(cp *Compiler, fn *parse.FormNode) exitusOp {
	var (
		names  []string
		values []*parse.CompoundNode
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

	return func(ev *Evaluator) exitus {
		return doSet(ev, names, vop.f(ev))
	}
}

var (
	arityMismatch = newFailure("arity mismatch")
	typeMismatch  = newFailure("type mismatch")
)

func doSet(ev *Evaluator, names []string, values []Value) exitus {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(names) != len(values) {
		return arityMismatch
	}

	for i, name := range names {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		variable := ev.ResolveVar(splitQualifiedName(name))
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

	return success
}

// DelForm = 'del' { VariablePrimary }
func compileDel(cp *Compiler, fn *parse.FormNode) exitusOp {
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
	return func(ev *Evaluator) exitus {
		for _, name := range names {
			delete(ev.local, name)
		}
		for _, name := range envNames {
			// TODO(xiaq): Signify possible error
			os.Unsetenv(name)
		}
		return success
	}
}

func stem(fname string) string {
	base := path.Base(fname)
	ext := path.Ext(base)
	return base[0 : len(base)-len(ext)]
}

// UseForm = 'use' StringPrimary.modname Primary.fname
//         = 'use' StringPrimary.fname
func compileUse(cp *Compiler, fn *parse.FormNode) exitusOp {
	var fnameNode *parse.CompoundNode
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

	newCp := NewCompiler(cp.builtin, cp.dataDir)
	op, err := newCp.Compile(fname, src, path.Dir(fname), cn)
	if err != nil {
		// TODO(xiaq): Pretty print
		cp.errorf(fnameNode.Pos, "cannot compile module: %s", err.Error())
	}

	cp.mod[modname] = newCp.scopes[0]

	return func(ev *Evaluator) exitus {
		// XXX(xiaq): Should use some part of ev
		newEv := NewEvaluator(ev.store, cp.dataDir)
		op(newEv)
		ev.mod[modname] = newEv.local
		return success
	}
}

// FnForm = 'fn' StringPrimary { VariablePrimary } ClosurePrimary
//
// fn defines a function. This isn't strictly needed, since user-defined
// functions are just variables. The following two lines should be exactly
// equivalent:
//
// fn f $a $b { put (* $a $b) (/ $a *b) }
// var $fn-f = { |$a $b| put (* $a $b) (/ $a $b) }
func compileFn(cp *Compiler, fn *parse.FormNode) exitusOp {
	if len(fn.Args.Nodes) == 0 {
		cp.errorf(fn.Pos, "expect function name after fn")
	}
	_, fnName := ensureStringPrimary(cp, fn.Args.Nodes[0], "expect string literal")
	varName := "fn-" + fnName

	var closureNode *parse.ClosureNode
	var argNames []*parse.CompoundNode

	for i, cn := range fn.Args.Nodes[1:] {
		expect := "expect variable or closure"
		pn := ensurePrimary(cp, cn, expect)
		switch pn.Typ {
		case parse.ClosurePrimary:
			if i+2 != len(fn.Args.Nodes) {
				cp.errorf(fn.Args.Nodes[i+2].Pos, "garbage after closure literal")
			}
			closureNode = pn.Node.(*parse.ClosureNode)
			break
		case parse.VariablePrimary:
			argNames = append(argNames, cn)
		default:
			cp.errorf(pn.Pos, expect)
		}
	}

	if len(argNames) > 0 {
		closureNode = &parse.ClosureNode{
			closureNode.Pos,
			&parse.SpacedNode{argNames[0].Pos, argNames},
			closureNode.Chunk,
		}
	}

	op := cp.closure(closureNode)

	cp.pushVar(varName, callableType{})

	return func(ev *Evaluator) exitus {
		ev.local[varName] = newInternalVariable(op.f(ev)[0], callableType{})
		return success
	}
}

func maybeClosurePrimary(cn *parse.CompoundNode) (*parse.ClosureNode, bool) {
	if len(cn.Nodes) == 1 && cn.Nodes[0].Right == nil && cn.Nodes[0].Left.Typ == parse.ClosurePrimary {
		return cn.Nodes[0].Left.Node.(*parse.ClosureNode), true
	}
	return nil, false
}

func maybeStringPrimary(cn *parse.CompoundNode) (string, bool) {
	if len(cn.Nodes) == 1 && cn.Nodes[0].Right == nil && cn.Nodes[0].Left.Typ == parse.StringPrimary {
		return cn.Nodes[0].Left.Node.(*parse.StringNode).Text, true
	}
	return "", false
}

type ifBranch struct {
	condition valuesOp
	body      valuesOp
}

// IfForm = 'if' Branch { 'else' 'if' Branch } [ 'else' Branch ]
// Branch = SpacedNode.condition ClosurePrimary.body
//
// The condition part of a Branch ends as soon as a Compound of a single
// ClosurePrimary is encountered.
func compileIf(cp *Compiler, fn *parse.FormNode) exitusOp {
	compounds := fn.Args.Nodes
	var branches []*ifBranch

	nextBranch := func() {
		var conds []*parse.CompoundNode
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
	return func(ev *Evaluator) exitus {
		for _, ib := range branches {
			if allTrue(ib.condition.f(ev)) {
				f := ib.body.f(ev)[0].(*closure)
				su := f.Exec(ev.copy("closure of if"), []Value{})
				for _ = range su {
				}
				// TODO(xiaq): Return the exitus of the body
				return success
			}
		}
		return success
	}
}

func compileStaticTypeof(cp *Compiler, fn *parse.FormNode) exitusOp {
	// Do conventional compiling of all compounds, only keeping the static type
	// information
	var trs []typeRun
	for _, cn := range fn.Args.Nodes {
		trs = append(trs, cp.compound(cn).tr)
	}
	return func(ev *Evaluator) exitus {
		out := ev.ports[1].ch
		for _, tr := range trs {
			out <- str(tr.String())
		}
		return success
	}
}
