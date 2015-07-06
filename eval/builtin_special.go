package eval

// Builtin special forms.

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/elves/elvish/parse"
)

type exitusOp func(*evalCtx) exitus
type builtinSpecialCompile func(*compileCtx, *parse.Form) exitusOp

type builtinSpecial struct {
	compile builtinSpecialCompile
}

var builtinSpecials map[string]builtinSpecial
var BuiltinSpecialNames []string

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
	for k, _ := range builtinSpecials {
		BuiltinSpecialNames = append(BuiltinSpecialNames, k)
	}
}

func mayAssign(tvar, tval Type) bool {
	if isAny(tval) || isAny(tvar) {
		return true
	}
	// BUG(xiaq): mayAssign uses a wrong way to check the equality of two
	// interfaces, which happens to work when all the Type instances are empty
	// structs.
	return tval == tvar
}

func checkSetType(cc *compileCtx, names []string, values []*parse.Compound, vop valuesOp, p parse.Pos) {
	n, more := vop.tr.count()
	if more {
		if n > len(names) {
			cc.errorf(p, "number of variables (%d) can never match that of values (%d or more)", len(names), n)
		}
		// Only check the variables before the "more" part.
		names = names[:n]
	} else if n != len(names) {
		cc.errorf(p, "number of variables (%d) doesn't match that of values (%d or more)", len(names), n)
	}

	for i, name := range names {
		tval := vop.tr[i].t
		tvar := cc.ResolveVar(splitQualifiedName(name))
		if !mayAssign(tvar, tval) {
			cc.errorf(values[i].Pos, "type mismatch: assigning %#v value to %#v variable", tval, tvar)
		}
	}
}

// ensure that a CompoundNode contains exactly one PrimaryNode.
func ensurePrimary(cc *compileCtx, cn *parse.Compound, msg string) *parse.Primary {
	if len(cn.Nodes) != 1 || cn.Nodes[0].Right != nil {
		cc.errorf(cn.Pos, msg)
	}
	return cn.Nodes[0].Left
}

// ensureVariableOrStringPrimary ensures that a CompoundNode contains exactly
// one PrimaryNode of type VariablePrimary or StringPrimary.
func ensureVariableOrStringPrimary(cc *compileCtx, cn *parse.Compound, msg string) (*parse.Primary, string) {
	pn := ensurePrimary(cc, cn, msg)
	switch pn.Typ {
	case parse.VariablePrimary, parse.StringPrimary:
		return pn, pn.Node.(*parse.String).Text
	default:
		cc.errorf(cn.Pos, msg)
		return nil, ""
	}
}

// ensureVariablePrimary ensures that a CompoundNode contains exactly one
// PrimaryNode of type VariablePrimary.
func ensureVariablePrimary(cc *compileCtx, cn *parse.Compound, msg string) (*parse.Primary, string) {
	pn, text := ensureVariableOrStringPrimary(cc, cn, msg)
	if pn.Typ != parse.VariablePrimary {
		cc.errorf(pn.Pos, msg)
	}
	return pn, text
}

// ensureStringPrimary ensures that a CompoundNode contains exactly one
// PrimaryNode of type VariablePrimary.
func ensureStringPrimary(cc *compileCtx, cn *parse.Compound, msg string) (*parse.Primary, string) {
	pn, text := ensureVariableOrStringPrimary(cc, cn, msg)
	if pn.Typ != parse.StringPrimary {
		cc.errorf(pn.Pos, msg)
	}
	return pn, text
}

// ensureStartWithVariabl ensures the first compound of the form is a
// VariablePrimary. This is merely for better error messages; No actual
// processing is done.
func ensureStartWithVariable(cc *compileCtx, fn *parse.Form, form string) {
	if len(fn.Args.Nodes) == 0 {
		cc.errorf(fn.Pos, "expect variable after %s", form)
	}
	ensureVariablePrimary(cc, fn.Args.Nodes[0], "expect variable")
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
func compileVar(cc *compileCtx, fn *parse.Form) exitusOp {
	var (
		names  []string
		types  []Type
		values []*parse.Compound
	)

	ensureStartWithVariable(cc, fn, "var")

	for i, cn := range fn.Args.Nodes {
		expect := "expect variable, type or equal sign"
		pn, text := ensureVariableOrStringPrimary(cc, cn, expect)
		if pn.Typ == parse.VariablePrimary {
			names = append(names, text)
		} else {
			if text == "=" {
				values = fn.Args.Nodes[i+1:]
				break
			} else {
				if t, ok := typenames[text]; !ok {
					cc.errorf(pn.Pos, "%v is not a valid type name", text)
				} else {
					if len(names) == len(types) {
						cc.errorf(pn.Pos, "duplicate type")
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
		cc.pushVar(name, types[i])
	}

	var vop valuesOp
	if values != nil {
		vop = cc.compounds(values)
		checkSetType(cc, names, values, vop, fn.Pos)
	}
	return func(ec *evalCtx) exitus {
		for i, name := range names {
			ec.local[name] = newInternalVariable(types[i].Default(), types[i])
		}
		if vop.f != nil {
			return doSet(ec, names, vop.f(ec))
		}
		return success
	}
}

// SetForm = 'set' { VariablePrimary } '=' { Compound }
func compileSet(cc *compileCtx, fn *parse.Form) exitusOp {
	var (
		names  []string
		values []*parse.Compound
	)

	ensureStartWithVariable(cc, fn, "set")

	for i, cn := range fn.Args.Nodes {
		expect := "expect variable or equal sign"
		pn, text := ensureVariableOrStringPrimary(cc, cn, expect)
		if pn.Typ == parse.VariablePrimary {
			ns, name := splitQualifiedName(text)
			cc.mustResolveVar(ns, name, cn.Pos)
			names = append(names, text)
		} else {
			if text != "=" {
				cc.errorf(pn.Pos, expect)
			}
			values = fn.Args.Nodes[i+1:]
			break
		}
	}

	var vop valuesOp
	vop = cc.compounds(values)
	checkSetType(cc, names, values, vop, fn.Pos)

	return func(ec *evalCtx) exitus {
		return doSet(ec, names, vop.f(ec))
	}
}

var (
	arityMismatch = newFailure("arity mismatch")
	typeMismatch  = newFailure("type mismatch")
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

	return success
}

// DelForm = 'del' { VariablePrimary }
func compileDel(cc *compileCtx, fn *parse.Form) exitusOp {
	// Do conventional compiling of all compound expressions, including
	// ensuring that variables can be resolved
	var names, envNames []string
	for _, cn := range fn.Args.Nodes {
		_, qname := ensureVariablePrimary(cc, cn, "expect variable")
		ns, name := splitQualifiedName(qname)
		switch ns {
		case "", "local":
			if cc.resolveVarOnThisScope(name) == nil {
				cc.errorf(cn.Pos, "variable $%s not found on current local scope", name)
			}
			cc.popVar(name)
			names = append(names, name)
		case "env":
			envNames = append(envNames, name)
		default:
			cc.errorf(cn.Pos, "can only delete a variable in local: or env: namespace")
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
func compileUse(cc *compileCtx, fn *parse.Form) exitusOp {
	var fnameNode *parse.Compound
	var fname, modname string

	switch len(fn.Args.Nodes) {
	case 0:
		cc.errorf(fn.Args.Pos, "expect module name or file name")
	case 1, 2:
		fnameNode = fn.Args.Nodes[0]
		_, fname = ensureStringPrimary(cc, fnameNode, "expect string literal")
		if len(fn.Args.Nodes) == 2 {
			modnameNode := fn.Args.Nodes[1]
			_, modname = ensureStringPrimary(
				cc, modnameNode, "expect string literal")
			if modname == "" {
				cc.errorf(modnameNode.Pos, "module name is empty")
			}
		} else {
			modname = stem(fname)
			if modname == "" {
				cc.errorf(fnameNode.Pos, "stem of file name is empty")
			}
		}
	default:
		cc.errorf(fn.Args.Nodes[2].Pos, "superfluous argument")
	}
	switch {
	case strings.HasPrefix(fname, "/"):
		// Absolute file name, do nothing
	case strings.HasPrefix(fname, "./") || strings.HasPrefix(fname, "../"):
		// File name relative to current source
		fname = path.Clean(path.Join(cc.dir, fname))
	default:
		// File name relative to data dir
		fname = path.Clean(path.Join(cc.dataDir, fname))
	}
	src, err := readFileUTF8(fname)
	if err != nil {
		cc.errorf(fnameNode.Pos, "cannot read module: %s", err.Error())
	}

	cn, err := parse.Parse(fname, src)
	if err != nil {
		// TODO(xiaq): Pretty print
		cc.errorf(fnameNode.Pos, "cannot parse module: %s", err.Error())
	}

	newCc := &compileCtx{
		cc.Compiler,
		fname, src, path.Dir(fname),
		[]staticNS{staticNS{}}, staticNS{},
	}

	op, err := newCc.compile(cn)
	if err != nil {
		// TODO(xiaq): Pretty print
		cc.errorf(fnameNode.Pos, "cannot compile module: %s", err.Error())
	}

	cc.mod[modname] = newCc.scopes[0]

	return func(ec *evalCtx) exitus {
		// TODO(xiaq): Should install a failHandler that fails the use call
		newEc := &evalCtx{
			ec.Evaler,
			fname, src, "module " + modname,
			ns{}, ns{},
			ec.ports, nil,
		}
		op(newEc)
		ec.mod[modname] = newEc.local
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
func compileFn(cc *compileCtx, fn *parse.Form) exitusOp {
	if len(fn.Args.Nodes) == 0 {
		cc.errorf(fn.Pos, "expect function name after fn")
	}
	_, fnName := ensureStringPrimary(cc, fn.Args.Nodes[0], "expect string literal")
	varName := fnPrefix + fnName

	var closureNode *parse.Closure
	var argNames []*parse.Compound

	for i, cn := range fn.Args.Nodes[1:] {
		expect := "expect variable or closure"
		pn := ensurePrimary(cc, cn, expect)
		switch pn.Typ {
		case parse.ClosurePrimary:
			if i+2 != len(fn.Args.Nodes) {
				cc.errorf(fn.Args.Nodes[i+2].Pos, "garbage after closure literal")
			}
			closureNode = pn.Node.(*parse.Closure)
			break
		case parse.VariablePrimary:
			argNames = append(argNames, cn)
		default:
			cc.errorf(pn.Pos, expect)
		}
	}

	if len(argNames) > 0 {
		closureNode = &parse.Closure{
			closureNode.Pos,
			&parse.Spaced{argNames[0].Pos, argNames},
			closureNode.Chunk,
		}
	}

	op := cc.closure(closureNode)

	cc.pushVar(varName, callableType{})

	return func(ec *evalCtx) exitus {
		ec.local[varName] = newInternalVariable(op.f(ec)[0], callableType{})
		return success
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

type ifBranch struct {
	condition valuesOp
	body      valuesOp
}

// IfForm = 'if' Branch { 'else' 'if' Branch } [ 'else' Branch ]
// Branch = SpacedNode.condition ClosurePrimary.body
//
// The condition part of a Branch ends as soon as a Compound of a single
// ClosurePrimary is encountered.
func compileIf(cc *compileCtx, fn *parse.Form) exitusOp {
	compounds := fn.Args.Nodes
	var branches []*ifBranch

	nextBranch := func() {
		var conds []*parse.Compound
		for i, cn := range compounds {
			if closure, ok := maybeClosurePrimary(cn); ok {
				if i == 0 {
					cc.errorf(cn.Pos, "expect condition")
				}
				condition := cc.compounds(conds)
				if closure.ArgNames != nil && len(closure.ArgNames.Nodes) > 0 {
					cc.errorf(closure.ArgNames.Pos, "unexpected arguments")
				}
				body := cc.closure(closure)
				branches = append(branches, &ifBranch{condition, body})
				compounds = compounds[i+1:]
				return
			}
			conds = append(conds, cn)
		}
		cc.errorf(compounds[len(compounds)-1].Pos, "expect body after this")
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
				cc.errorf(compounds[0].Pos, "expect body after this")
			} else if len(compounds) > 2 {
				cc.errorf(compounds[2].Pos, "trailing garbage")
			}
			body, ok := maybeClosurePrimary(compounds[1])
			if !ok {
				cc.errorf(compounds[1].Pos, "expect body")
			}
			branches = append(branches, &ifBranch{
				literalValue(boolean(true)), cc.closure(body)})
		} else {
			cc.errorf(compounds[0].Pos, "trailing garbage")
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
				return success
			}
		}
		return success
	}
}

func compileStaticTypeof(cc *compileCtx, fn *parse.Form) exitusOp {
	// Do conventional compiling of all compounds, only keeping the static type
	// information
	var trs []typeRun
	for _, cn := range fn.Args.Nodes {
		trs = append(trs, cc.compound(cn).tr)
	}
	return func(ec *evalCtx) exitus {
		out := ec.ports[1].ch
		for _, tr := range trs {
			out <- str(tr.String())
		}
		return success
	}
}
