package eval

// Builtin special forms.

import (
	"os"

	"github.com/elves/elvish/parse"
)

type compileBuiltin func(*compiler, *parse.Form) Op

var builtinSpecials map[string]compileBuiltin

// BuiltinSpecialNames contains all names of builtin special forms. It is
// useful for the syntax highlighter.
var BuiltinSpecialNames []string

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]compileBuiltin{
		"del": compileDel,
		"fn":  compileFn,
		//"use": compileUse,
	}
	for k := range builtinSpecials {
		BuiltinSpecialNames = append(BuiltinSpecialNames, k)
	}
}

func doSet(ec *EvalCtx, variables []Variable, values []Value) {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(variables) != len(values) {
		throw(ErrArityMismatch)
	}

	for i, variable := range variables {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		variable.Set(values[i])
	}
}

// DelForm = 'del' { VariablePrimary }
func compileDel(cp *compiler, fn *parse.Form) Op {
	// Do conventional compiling of all compound expressions, including
	// ensuring that variables can be resolved
	var names, envNames []string
	for _, cn := range fn.Args {
		qname := mustString(cp, cn, "should be a literal variable name")
		splice, ns, name := parseVariable(qname)
		if splice {
			cp.errorf(cn.Begin(), "removing spliced variable makes no sense")
		}
		switch ns {
		case "", "local":
			if !cp.thisScope()[name] {
				cp.errorf(cn.Begin(), "variable $%s not found on current local scope", name)
			}
			delete(cp.thisScope(), name)
			names = append(names, name)
		case "env":
			envNames = append(envNames, name)
		default:
			cp.errorf(cn.Begin(), "can only delete a variable in local: or env:")
		}

	}
	return func(ec *EvalCtx) {
		for _, name := range names {
			delete(ec.local, name)
		}
		for _, name := range envNames {
			// BUG(xiaq): We rely on the fact that os.Unsetenv always returns
			// nil.
			os.Unsetenv(name)
		}
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
func compileUse(cp *compiler, fn *parse.Form) op {
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

// makeFnOp wraps an op such that a return is converted to an ok.
func makeFnOp(op Op) Op {
	return func(ec *EvalCtx) {
		ex := ec.PEval(op)
		if ex != Return {
			// rethrow
			throw(ex)
		}
	}
}

// FnForm = 'fn' StringPrimary LambdaPrimary
//
// fn f []{foobar} is a shorthand for set '&'f = []{foobar}.
func compileFn(cp *compiler, fn *parse.Form) Op {
	if len(fn.Args) == 0 {
		cp.errorf(fn.End(), "should be followed by function name")
	}
	fnName := mustString(cp, fn.Args[0], "must be a literal string")
	varName := FnPrefix + fnName

	if len(fn.Args) == 1 {
		cp.errorf(fn.Args[0].End(), "should be followed by a lambda")
	}
	pn := mustPrimary(cp, fn.Args[1], "should be a lambda")
	if pn.Type != parse.Lambda {
		cp.errorf(pn.Begin(), "should be a lambda")
	}
	if len(fn.Args) > 2 {
		cp.errorf(fn.Args[2].Begin(), "superfluous argument")
	}

	cp.registerVariableSet(":" + varName)
	op := cp.lambda(pn)

	return func(ec *EvalCtx) {
		closure := op(ec)[0].(*Closure)
		closure.Op = makeFnOp(closure.Op)
		ec.local[varName] = NewPtrVariable(closure)
	}
}
