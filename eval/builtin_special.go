package eval

//go:generate ./builtin_modules.bash

// Builtin special forms.

import (
	"fmt"
	"os"
	"strings"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
)

type compileBuiltin func(*compiler, *parse.Form) OpFunc

var builtinSpecials map[string]compileBuiltin

// BuiltinSpecialNames contains all names of builtin special forms. It is
// useful for the syntax highlighter.
var BuiltinSpecialNames []string

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]compileBuiltin{
		"del": compileDel,
		"fn":  compileFn,
		"use": compileUse,
	}
	for k := range builtinSpecials {
		BuiltinSpecialNames = append(BuiltinSpecialNames, k)
	}
}

// DelForm = 'del' { VariablePrimary }
func compileDel(cp *compiler, fn *parse.Form) OpFunc {
	// Do conventional compiling of all compound expressions, including
	// ensuring that variables can be resolved
	var names, envNames []string
	for _, cn := range fn.Args {
		cp.compiling(cn)
		qname := mustString(cp, cn, "should be a literal variable name")
		explode, ns, name := ParseAndFixVariable(qname)
		if explode {
			cp.errorf("removing exploded variable makes no sense")
		}
		switch ns {
		case "", "local":
			if !cp.thisScope()[name] {
				cp.errorf("variable $%s not found on current local scope", name)
			}
			delete(cp.thisScope(), name)
			names = append(names, name)
		case "E":
			envNames = append(envNames, name)
		default:
			cp.errorf("can only delete a variable in local: or E:")
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

// makeFnOp wraps an op such that a return is converted to an ok.
func makeFnOp(op Op) Op {
	return Op{func(ec *EvalCtx) {
		err := ec.PEval(op)
		if err != nil && err != Return {
			// rethrow
			throw(err)
		}
	}, op.Begin, op.End}
}

// FnForm = 'fn' StringPrimary LambdaPrimary
//
// fn f []{foobar} is a shorthand for set '&'f = []{foobar}.
func compileFn(cp *compiler, fn *parse.Form) OpFunc {
	if len(fn.Args) == 0 {
		end := fn.End()
		cp.errorpf(end, end, "should be followed by function name")
	}
	fnName := mustString(cp, fn.Args[0], "must be a literal string")
	varName := FnPrefix + fnName

	if len(fn.Args) == 1 {
		end := fn.Args[0].End()
		cp.errorpf(end, end, "should be followed by a lambda")
	}
	pn := mustPrimary(cp, fn.Args[1], "should be a lambda")
	if pn.Type != parse.Lambda {
		cp.compiling(pn)
		cp.errorf("should be a lambda")
	}
	if len(fn.Args) > 2 {
		cp.errorpf(fn.Args[2].Begin(), fn.Args[len(fn.Args)-1].End(), "superfluous argument(s)")
	}

	cp.registerVariableSet(":" + varName)
	op := cp.lambda(pn)

	return func(ec *EvalCtx) {
		// Initialize the function variable with the builtin nop
		// function. This step allows the definition of recursive
		// functions; the actual function will never be called.
		ec.local[varName] = NewPtrVariable(&BuiltinFn{"<shouldn't be called>", nop})
		closure := op(ec)[0].(*Closure)
		closure.Op = makeFnOp(closure.Op)
		ec.local[varName].Set(closure)
	}
}

// UseForm = 'use' StringPrimary [ Compound ]
func compileUse(cp *compiler, fn *parse.Form) OpFunc {
	var modname string
	var filenameOp ValuesOp
	var filenameBegin, filenameEnd int

	switch len(fn.Args) {
	case 0:
		end := fn.Head.End()
		cp.errorpf(end, end, "lack module name")
	case 2:
		filenameOp = cp.compoundOp(fn.Args[1])
		filenameBegin = fn.Args[1].Begin()
		filenameEnd = fn.Args[1].End()
		fallthrough
	case 1:
		modname = mustString(cp, fn.Args[0], "should be a literal module name")
	default:
		cp.errorpf(fn.Args[2].Begin(), fn.Args[len(fn.Args)-1].End(), "superfluous argument(s)")
	}

	return func(ec *EvalCtx) {
		if filenameOp.Func != nil {
			values := filenameOp.Exec(ec)
			valuesMust := &muster{ec, "module filename", filenameBegin, filenameEnd, values}
			filename := string(valuesMust.mustOneStr())
			use(ec, modname, &filename)
		} else {
			use(ec, modname, nil)
		}
	}
}

func use(ec *EvalCtx, modname string, pfilename *string) {
	if _, ok := ec.Evaler.Modules[modname]; ok {
		// Module already loaded.
		return
	}

	// Load the source.
	var filename, source string

	if pfilename != nil {
		filename = *pfilename
		var err error
		source, err = readFileUTF8(filename)
		maybeThrow(err)
	} else {
		// No filename; defaulting to $datadir/$modname.elv.
		dataDir, err := store.DataDir()
		maybeThrow(err)
		filename = dataDir + "/" + strings.Replace(modname, ":", "/", -1) + ".elv"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// File does not exist. Try loading from the table of builtin
			// modules.
			var ok bool
			if source, ok = builtinModules[modname]; ok {
				// Source is loaded. Do nothing more.
				filename = "<builtin module>"
			} else {
				throw(fmt.Errorf("cannot load %s: %s does not exist", modname, filename))
			}
		} else {
			// File exists. Load it.
			source, err = readFileUTF8(filename)
			maybeThrow(err)
		}
	}

	n, err := parse.Parse(filename, source)
	maybeThrow(err)

	// Make an empty namespace.
	local := Namespace{}

	// TODO(xiaq): Should handle failures when evaluting the module
	newEc := &EvalCtx{
		ec.Evaler, "module " + modname,
		filename, source,
		local, Namespace{},
		ec.ports, nil, true,
		0, len(source), ec.addTraceback(), false,
	}

	op, err := newEc.Compile(n, filename, source)
	// TODO the err originates in another source, should add appropriate information.
	maybeThrow(err)

	// Load the namespace before executing. This avoids mutual and self use's to
	// result in an infinite recursion.
	ec.Evaler.Modules[modname] = local
	err = newEc.PEval(op)
	if err != nil {
		// Unload the namespace.
		delete(ec.Modules, modname)
		throw(err)
	}
}
