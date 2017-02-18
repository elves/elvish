package eval

// Builtin special forms.

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/elves/elvish/parse"
)

type compileBuiltin func(*compiler, *parse.Form) OpFunc

var ErrNoDataDir = errors.New("There is no data directory")

var builtinSpecials map[string]compileBuiltin

// BuiltinSpecialNames contains all names of builtin special forms. It is
// useful for the syntax highlighter.
var BuiltinSpecialNames []string

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]compileBuiltin{
		"del":   compileDel,
		"fn":    compileFn,
		"use":   compileUse,
		"if":    compileIf,
		"while": compileWhile,
		"for":   compileFor,
		"try":   compileTry,
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
		if err != nil && err.(*Exception).Cause != Return {
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
		// No filename; defaulting to $datadir/lib/$modname.elv.
		if ec.DataDir == "" {
			throw(ErrNoDataDir)
		}
		filename = ec.DataDir + "/lib/" + strings.Replace(modname, ":", "/", -1) + ".elv"
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
		ec.ports, nil,
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

func compileIf(cp *compiler, fn *parse.Form) OpFunc {
	args := cp.walkArgs(fn)
	var condNodes, bodyNodes []*parse.Compound
	for {
		condNodes = append(condNodes, args.next())
		bodyNodes = append(bodyNodes, args.next())
		if !args.nextIs("elif") {
			break
		}
	}
	elseNode := args.nextLedBy("else")
	args.mustEnd()

	condOps := cp.compoundOps(condNodes)
	bodyOps := cp.compoundOps(bodyNodes)
	var elseOp ValuesOp
	if elseNode != nil {
		elseOp = cp.compoundOp(elseNode)
	}

	return func(ec *EvalCtx) {
		bodies := make([]Fn, len(bodyOps))
		for i, bodyOp := range bodyOps {
			bodies[i] = bodyOp.execMustOneFn(ec)
		}
		else_ := elseOp.execMustOneFn(ec)
		for i, condOp := range condOps {
			if allTrue(condOp.Exec(ec)) {
				bodies[i].Call(ec, NoArgs, NoOpts)
				return
			}
		}
		if elseOp.Func != nil {
			else_.Call(ec, NoArgs, NoOpts)
		}
	}
}

func compileWhile(cp *compiler, fn *parse.Form) OpFunc {
	args := cp.walkArgs(fn)
	condNode := args.next()
	bodyNode := args.next()
	args.mustEnd()

	condOp := cp.compoundOp(condNode)
	bodyOp := cp.compoundOp(bodyNode)

	return func(ec *EvalCtx) {
		body := bodyOp.execMustOneFn(ec)

		for {
			cond := condOp.Exec(ec)
			if !allTrue(cond) {
				break
			}
			err := ec.PCall(body, NoArgs, NoOpts)
			if err != nil {
				exc := err.(*Exception)
				if exc.Cause == Continue {
					// do nothing
				} else if exc.Cause == Break {
					continue
				} else {
					throw(err)
				}
			}
		}
	}
}

func compileFor(cp *compiler, fn *parse.Form) OpFunc {
	args := cp.walkArgs(fn)
	varNode := args.next()
	iterNode := args.next()
	bodyNode := args.next()
	elseNode := args.nextLedBy("else")
	args.mustEnd()

	varOp, restOp := cp.lvaluesOp(varNode.Indexings[0])
	if restOp.Func != nil {
		cp.errorpf(restOp.Begin, restOp.End, "rest not allowed")
	}

	iterOp := cp.compoundOp(iterNode)
	bodyOp := cp.compoundOp(bodyNode)
	var elseOp ValuesOp
	if elseNode != nil {
		elseOp = cp.compoundOp(elseNode)
	}

	return func(ec *EvalCtx) {
		variables := varOp.Exec(ec)
		if len(variables) != 1 {
			ec.errorpf(varOp.Begin, varOp.End, "only one variable allowed")
		}
		variable := variables[0]

		iterables := iterOp.Exec(ec)
		if len(iterables) != 1 {
			ec.errorpf(iterOp.Begin, iterOp.End, "should be one iterable")
		}
		iterable, ok := iterables[0].(Iterator)
		if !ok {
			ec.errorpf(iterOp.Begin, iterOp.End, "should be one iterable")
		}

		body := bodyOp.execMustOneFn(ec)
		elseBody := elseOp.execMustOneFn(ec)

		iterated := false
		iterable.Iterate(func(v Value) bool {
			iterated = true
			variable.Set(v)
			err := ec.PCall(body, NoArgs, NoOpts)
			if err != nil {
				exc := err.(*Exception)
				if exc.Cause == Continue {
					// do nothing
				} else if exc.Cause == Break {
					return false
				} else {
					throw(err)
				}
			}
			return true
		})

		if !iterated && elseBody != nil {
			elseBody.Call(ec, NoArgs, NoOpts)
		}
	}
}

func compileTry(cp *compiler, fn *parse.Form) OpFunc {
	Logger.Println("compiling try")
	args := cp.walkArgs(fn)
	bodyNode := args.next()
	Logger.Printf("body is %q", bodyNode.SourceText())
	var exceptVarNode *parse.Indexing
	var exceptNode *parse.Compound
	if args.nextIs("except") {
		Logger.Println("except-ing")
		n := args.next()
		if len(n.Indexings) != 1 {
			cp.errorpf(n.Begin(), n.End(), "should be one variable")
		}
		exceptVarNode = n.Indexings[0]
		exceptNode = args.next()
		Logger.Printf("except-var = %q, except = %q", exceptVarNode.SourceText(), exceptNode.SourceText())
	}
	elseNode := args.nextLedBy("else")
	finallyNode := args.nextLedBy("finally")
	args.mustEnd()

	var exceptVarOp LValuesOp
	var bodyOp, exceptOp, elseOp, finallyOp ValuesOp
	bodyOp = cp.compoundOp(bodyNode)
	if exceptVarNode != nil {
		var restOp LValuesOp
		exceptVarOp, restOp = cp.lvaluesOp(exceptVarNode)
		if restOp.Func != nil {
			cp.errorpf(restOp.Begin, restOp.End, "may not use @rest in except variable")
		}
		exceptOp = cp.compoundOp(exceptNode)
	}
	if elseNode != nil {
		elseOp = cp.compoundOp(elseNode)
	}
	if finallyNode != nil {
		finallyOp = cp.compoundOp(finallyNode)
	}

	return func(ec *EvalCtx) {
		body := bodyOp.execMustOneFn(ec)
		exceptVar := exceptVarOp.execMustOne(ec)
		except := exceptOp.execMustOneFn(ec)
		else_ := elseOp.execMustOneFn(ec)
		finally := finallyOp.execMustOneFn(ec)

		err := ec.PCall(body, NoArgs, NoOpts)
		if err != nil {
			if except != nil {
				exceptVar.Set(err.(*Exception))
				err = ec.PCall(except, NoArgs, NoOpts)
			}
		} else {
			if else_ != nil {
				err = ec.PCall(else_, NoArgs, NoOpts)
			}
		}
		if finally != nil {
			finally.Call(ec, NoArgs, NoOpts)
		}
		if err != nil {
			throw(err)
		}
	}
}

// execMustOneFn executes the ValuesOp and raises an exception if it does not
// evaluate to exactly one Fn. If the given ValuesOp is empty, it returns nil.
func (op ValuesOp) execMustOneFn(ec *EvalCtx) Fn {
	if op.Func == nil {
		return nil
	}

	values := op.Exec(ec)
	if len(values) != 1 {
		ec.errorpf(op.Begin, op.End, "should be one fn")
	}
	fn, ok := values[0].(Fn)
	if !ok {
		ec.errorpf(op.Begin, op.End, "should be one fn")
	}
	return fn
}

// execMustOne executes the LValuesOp and raises an exception if it does not
// evaluate to exactly one Variable. If the given LValuesOp is empty, it returns
// nil.
func (op LValuesOp) execMustOne(ec *EvalCtx) Variable {
	if op.Func == nil {
		return nil
	}
	variables := op.Exec(ec)
	if len(variables) != 1 {
		ec.errorpf(op.Begin, op.End, "should be one variable")
	}
	return variables[0]
}
