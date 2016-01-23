package eval

// Builtin special forms.

import (
	"os"

	"github.com/elves/elvish/parse-ng"
)

type compileBuiltin func(*compiler, *parse.Form) exitusOp

var builtinSpecials map[string]compileBuiltin
var BuiltinSpecialNames []string

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]compileBuiltin{
		"set": compileSet,
		"del": compileDel,
		"fn":  compileFn,
		//"use": compileUse,
	}
	for k, _ := range builtinSpecials {
		BuiltinSpecialNames = append(BuiltinSpecialNames, k)
	}
}

// SetForm = 'set' { StringPrimary } '=' { Compound }
func compileSet(cp *compiler, fn *parse.Form) exitusOp {
	var (
		names  []string
		values []*parse.Compound
	)

	if len(fn.Args) == 0 {
		cp.errorf(fn.Begin, "empty set")
	}
	mustString(cp, fn.Args[0], "should be a literal variable name")

	for i, cn := range fn.Args {
		name := mustString(cp, cn, "should be a literal variable name or equal sign")
		if name == "=" {
			values = fn.Args[i+1:]
			break
		}
		cp.registerVariableSet(name)
		names = append(names, name)
	}

	valueOps := cp.compounds(values)
	valuesOp := catOps(valueOps)

	return func(ec *evalCtx) exitus {
		return doSet(ec, names, valuesOp(ec))
	}
}

func doSet(ec *evalCtx, names []string, values []Value) exitus {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(names) != len(values) {
		return arityMismatch
	}

	for i, qname := range names {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		ns, name := splitQualifiedName(qname)
		variable := ec.ResolveVar(ns, name)
		if variable == nil {
			// New variable
			variable = newInternalVariable(values[i])
			ec.local[name] = variable
		} else {
			variable.Set(values[i])
		}
	}

	return ok
}

// DelForm = 'del' { VariablePrimary }
func compileDel(cp *compiler, fn *parse.Form) exitusOp {
	// Do conventional compiling of all compound expressions, including
	// ensuring that variables can be resolved
	var names, envNames []string
	for _, cn := range fn.Args {
		qname := mustString(cp, cn, "should be a literal variable name")
		ns, name := splitQualifiedName(qname)
		switch ns {
		case "", "local":
			if !cp.thisScope()[name] {
				cp.errorf(cn.Begin, "variable $%s not found on current local scope", name)
			}
			delete(cp.thisScope(), name)
			names = append(names, name)
		case "env":
			envNames = append(envNames, name)
		default:
			cp.errorf(cn.Begin, "can only delete a variable in local: or env:")
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

/*
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
*/

// makeFnOp wraps a valuesOp such that a return is converted to an ok.
func makeFnOp(op valuesOp) valuesOp {
	return func(ec *evalCtx) []Value {
		vs := op(ec)
		if len(vs) == 1 {
			if e, ok_ := vs[0].(exitus); ok_ {
				if e.Sort == Return {
					return []Value{ok}
				}
			}
		}
		return vs
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
func compileFn(cp *compiler, fn *parse.Form) exitusOp {
	if len(fn.Args) == 0 {
		cp.errorf(fn.End, "should be followed by function name")
	}
	fnName := mustString(cp, fn.Args[0], "must be a literal string")
	varName := fnPrefix + fnName

	if len(fn.Args) == 1 {
		cp.errorf(fn.Args[0].End, "should be followed by a lambda")
	}
	pn := mustPrimary(cp, fn.Args[1], "should be a lambda")
	if pn.Type != parse.Lambda {
		cp.errorf(pn.Begin, "should be a lambda")
	}
	if len(fn.Args) > 2 {
		cp.errorf(fn.Args[2].Begin, "superfluous argument")
	}

	cp.registerVariableSet(varName)
	op := cp.lambda(pn)

	return func(ec *evalCtx) exitus {
		closure := op(ec)[0].(*closure)
		closure.Op = makeFnOp(closure.Op)
		ec.local[varName] = newInternalVariable(closure)
		return ok
	}
}
