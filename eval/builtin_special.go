package eval

// Builtin special forms. Special forms behave mostly like ordinary commands -
// they are valid commands syntactically, and can take part in pipelines - but
// they have special rules for the evaluation of their arguments and can affect
// the compilation phase (whereas ordinary commands can only affect the
// evaluation phase).
//
// For instance, the "and" special form evaluates its arguments from left to
// right, and stops as soon as one booleanly false value is obtained: the
// command "and $false (fail haha)" does not produce an exception.
//
// As another instance, the "del" special form removes a variable, affecting the
// compiler.
//
// Flow control structures are also implemented as special forms in elvish, with
// closures functioning as code blocks.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/parse"
)

type compileBuiltin func(*compiler, *parse.Form) OpFunc

var (
	// ErrNoLibDir is thrown by "use" when the Evaler does not have a library
	// directory.
	ErrNoLibDir = errors.New("Evaler does not have a lib directory")
	// ErrRelativeUseNotFromMod is thrown by "use" when relative use is used
	// not from a module
	ErrRelativeUseNotFromMod = errors.New("Relative use not from module")
	// ErrRelativeUseGoesOutsideLib is thrown when a relative use goes out of
	// the library directory.
	ErrRelativeUseGoesOutsideLib = errors.New("Module outside library directory")
)

var builtinSpecials map[string]compileBuiltin

// IsBuiltinSpecial is the set of all names of builtin special forms. It is
// intended for external consumption, e.g. the syntax highlighter.
var IsBuiltinSpecial = map[string]bool{}

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]compileBuiltin{
		"del":   compileDel,
		"fn":    compileFn,
		"use":   compileUse,
		"and":   compileAnd,
		"or":    compileOr,
		"if":    compileIf,
		"while": compileWhile,
		"for":   compileFor,
		"try":   compileTry,
	}
	for name := range builtinSpecials {
		IsBuiltinSpecial[name] = true
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
		explode, ns, name := ParseVariable(qname)
		if explode {
			cp.errorf("removing exploded variable makes no sense")
		}
		switch ns {
		case "", "local":
			if cp.thisScope().has(name) {
				cp.errorf("variable $%s not found on current local scope", name)
			}
			cp.thisScope().del(name)
			names = append(names, name)
		case "E":
			envNames = append(envNames, name)
		default:
			cp.errorf("can only delete a variable in local: or E:")
		}

	}
	return func(ec *Frame) {
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
	return Op{func(ec *Frame) {
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
	args := cp.walkArgs(fn)
	nameNode := args.next()
	varName := mustString(cp, nameNode, "must be a literal string") + FnSuffix
	bodyNode := args.nextMustLambda()
	args.mustEnd()

	cp.registerVariableSetQname(":" + varName)
	op := cp.lambda(bodyNode)

	return func(ec *Frame) {
		// Initialize the function variable with the builtin nop
		// function. This step allows the definition of recursive
		// functions; the actual function will never be called.
		ec.local[varName] = vartypes.NewPtrVariable(&BuiltinFn{"<shouldn't be called>", nop})
		closure := op(ec)[0].(*Closure)
		closure.Op = makeFnOp(closure.Op)
		err := ec.local[varName].Set(closure)
		maybeThrow(err)
	}
}

// UseForm = 'use' StringPrimary
func compileUse(cp *compiler, fn *parse.Form) OpFunc {
	if len(fn.Args) == 0 {
		end := fn.Head.End()
		cp.errorpf(end, end, "lack module name")
	} else if len(fn.Args) >= 2 {
		cp.errorpf(fn.Args[1].Begin(), fn.Args[len(fn.Args)-1].End(), "superfluous argument(s)")
	}

	spec := mustString(cp, fn.Args[0], "should be a literal string")

	// When modspec = "a/b/c:d", modname is c:d, and modpath is a/b/c/d
	modname := spec[strings.LastIndexByte(spec, '/')+1:]
	modpath := strings.Replace(spec, ":", "/", -1)
	cp.thisScope().set(modname + NsSuffix)

	return func(ec *Frame) {
		use(ec, modname, modpath)
	}
}

func use(ec *Frame, modname, modpath string) {
	resolvedPath := ""
	if strings.HasPrefix(modpath, "./") || strings.HasPrefix(modpath, "../") {
		if ec.modPath == "" {
			throw(ErrRelativeUseNotFromMod)
		}
		// Resolve relative modpath.
		resolvedPath = filepath.Clean(filepath.Dir(ec.modPath) + "/" + modpath)
	} else {
		resolvedPath = filepath.Clean(modpath)
	}
	if strings.HasPrefix(resolvedPath, "../") {
		throw(ErrRelativeUseGoesOutsideLib)
	}
	modpath = resolvedPath

	// Put the just loaded module into local scope.
	ec.local[modname+NsSuffix] = vartypes.NewPtrVariable(loadModule(ec, modpath))
}

func loadModule(ec *Frame, modpath string) Ns {
	if ns, ok := ec.Evaler.modules[modpath]; ok {
		// Module already loaded.
		return ns
	}

	// Load the source.
	var filename, source string

	// No filename; defaulting to $datadir/lib/$modpath.elv.
	if ec.libDir == "" {
		throw(ErrNoLibDir)
	}

	filename = filepath.Join(ec.libDir, modpath+".elv")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File does not exist. Try loading from the table of builtin
		// modules.
		var ok bool
		if source, ok = ec.bundled[modpath]; ok {
			// Source is loaded. Do nothing more.
			filename = "<builtin module>"
		} else {
			throw(fmt.Errorf("cannot load %s: %s does not exist", modpath, filename))
		}
	} else {
		// File exists. Load it.
		source, err = readFileUTF8(filename)
		maybeThrow(err)
	}

	n, err := parse.Parse(filename, source)
	maybeThrow(err)

	// Make an empty scope to evaluate the module in.
	local := make(Ns)

	newEc := &Frame{
		ec.Evaler, "module " + modpath,
		filename, source, modpath,
		local, make(Ns),
		ec.ports,
		0, len(source), ec.addTraceback(), false,
	}

	op, err := newEc.Compile(n, filename, source)
	maybeThrow(err)

	// Load the namespace before executing. This avoids mutual and self use's to
	// result in an infinite recursion.
	ec.Evaler.modules[modpath] = local
	err = newEc.PEval(op)
	if err != nil {
		// Unload the namespace.
		delete(ec.modules, modpath)
		throw(err)
	}
	return local
}

// compileAnd compiles the "and" special form.
// The and special form evaluates arguments until a false-ish values is found
// and outputs it; the remaining arguments are not evaluated. If there are no
// false-ish values, the last value is output. If there are no arguments, it
// outputs $true, as if there is a hidden $true before actual arguments.
func compileAnd(cp *compiler, fn *parse.Form) OpFunc {
	return compileAndOr(cp, fn, true, false)
}

// compileOr compiles the "or" special form.
// The or special form evaluates arguments until a true-ish values is found and
// outputs it; the remaining arguments are not evaluated. If there are no
// true-ish values, the last value is output. If there are no arguments, it
// outputs $false, as if there is a hidden $false before actual arguments.
func compileOr(cp *compiler, fn *parse.Form) OpFunc {
	return compileAndOr(cp, fn, false, true)
}

func compileAndOr(cp *compiler, fn *parse.Form, init, stopAt bool) OpFunc {
	argOps := cp.compoundOps(fn.Args)
	return func(ec *Frame) {
		var lastValue types.Value = types.Bool(init)
		for _, op := range argOps {
			values := op.Exec(ec)
			for _, value := range values {
				if types.ToBool(value) == stopAt {
					ec.OutputChan() <- value
					return
				}
				lastValue = value
			}
		}
		ec.OutputChan() <- lastValue
	}
}

func compileIf(cp *compiler, fn *parse.Form) OpFunc {
	args := cp.walkArgs(fn)
	var condNodes []*parse.Compound
	var bodyNodes []*parse.Primary
	for {
		condNodes = append(condNodes, args.next())
		bodyNodes = append(bodyNodes, args.nextMustLambda())
		if !args.nextIs("elif") {
			break
		}
	}
	elseNode := args.nextMustLambdaIfAfter("else")
	args.mustEnd()

	condOps := cp.compoundOps(condNodes)
	bodyOps := cp.primaryOps(bodyNodes)
	var elseOp ValuesOp
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}

	return func(ec *Frame) {
		bodies := make([]Callable, len(bodyOps))
		for i, bodyOp := range bodyOps {
			bodies[i] = bodyOp.execlambdaOp(ec)
		}
		else_ := elseOp.execlambdaOp(ec)
		for i, condOp := range condOps {
			if allTrue(condOp.Exec(ec.fork("if cond"))) {
				bodies[i].Call(ec.fork("if body"), NoArgs, NoOpts)
				return
			}
		}
		if elseOp.Func != nil {
			else_.Call(ec.fork("if else"), NoArgs, NoOpts)
		}
	}
}

func compileWhile(cp *compiler, fn *parse.Form) OpFunc {
	args := cp.walkArgs(fn)
	condNode := args.next()
	bodyNode := args.nextMustLambda()
	args.mustEnd()

	condOp := cp.compoundOp(condNode)
	bodyOp := cp.primaryOp(bodyNode)

	return func(ec *Frame) {
		body := bodyOp.execlambdaOp(ec)

		for {
			cond := condOp.Exec(ec.fork("while cond"))
			if !allTrue(cond) {
				break
			}
			err := ec.fork("while").PCall(body, NoArgs, NoOpts)
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
	bodyNode := args.nextMustLambda()
	elseNode := args.nextMustLambdaIfAfter("else")
	args.mustEnd()

	varOp, restOp := cp.lvaluesOp(varNode.Indexings[0])
	if restOp.Func != nil {
		cp.errorpf(restOp.Begin, restOp.End, "rest not allowed")
	}

	iterOp := cp.compoundOp(iterNode)
	bodyOp := cp.primaryOp(bodyNode)
	var elseOp ValuesOp
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}

	return func(ec *Frame) {
		variables := varOp.Exec(ec)
		if len(variables) != 1 {
			ec.errorpf(varOp.Begin, varOp.End, "only one variable allowed")
		}
		variable := variables[0]

		iterable := ec.ExecAndUnwrap("value being iterated", iterOp).One().Iterable()

		body := bodyOp.execlambdaOp(ec)
		elseBody := elseOp.execlambdaOp(ec)

		iterated := false
		iterable.Iterate(func(v types.Value) bool {
			iterated = true
			err := variable.Set(v)
			maybeThrow(err)
			err = ec.fork("for").PCall(body, NoArgs, NoOpts)
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
			elseBody.Call(ec.fork("for else"), NoArgs, NoOpts)
		}
	}
}

func compileTry(cp *compiler, fn *parse.Form) OpFunc {
	logger.Println("compiling try")
	args := cp.walkArgs(fn)
	bodyNode := args.nextMustLambda()
	logger.Printf("body is %q", bodyNode.SourceText())
	var exceptVarNode *parse.Indexing
	var exceptNode *parse.Primary
	if args.nextIs("except") {
		logger.Println("except-ing")
		n := args.peek()
		// Is this a variable?
		if len(n.Indexings) == 1 && n.Indexings[0].Head.Type == parse.Bareword {
			exceptVarNode = n.Indexings[0]
			args.next()
		}
		exceptNode = args.nextMustLambda()
	}
	elseNode := args.nextMustLambdaIfAfter("else")
	finallyNode := args.nextMustLambdaIfAfter("finally")
	args.mustEnd()

	var exceptVarOp LValuesOp
	var bodyOp, exceptOp, elseOp, finallyOp ValuesOp
	bodyOp = cp.primaryOp(bodyNode)
	if exceptVarNode != nil {
		var restOp LValuesOp
		exceptVarOp, restOp = cp.lvaluesOp(exceptVarNode)
		if restOp.Func != nil {
			cp.errorpf(restOp.Begin, restOp.End, "may not use @rest in except variable")
		}
	}
	if exceptNode != nil {
		exceptOp = cp.primaryOp(exceptNode)
	}
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}
	if finallyNode != nil {
		finallyOp = cp.primaryOp(finallyNode)
	}

	return func(ec *Frame) {
		body := bodyOp.execlambdaOp(ec)
		exceptVar := exceptVarOp.execMustOne(ec)
		except := exceptOp.execlambdaOp(ec)
		else_ := elseOp.execlambdaOp(ec)
		finally := finallyOp.execlambdaOp(ec)

		err := ec.fork("try body").PCall(body, NoArgs, NoOpts)
		if err != nil {
			if except != nil {
				if exceptVar != nil {
					err := exceptVar.Set(err.(*Exception))
					maybeThrow(err)
				}
				err = ec.fork("try except").PCall(except, NoArgs, NoOpts)
			}
		} else {
			if else_ != nil {
				err = ec.fork("try else").PCall(else_, NoArgs, NoOpts)
			}
		}
		if finally != nil {
			finally.Call(ec.fork("try finally"), NoArgs, NoOpts)
		}
		if err != nil {
			throw(err)
		}
	}
}

// execLambdaOp executes a ValuesOp that is known to yield a lambda and returns
// the lambda. If the ValuesOp is empty, it returns a nil.
func (op ValuesOp) execlambdaOp(ec *Frame) Callable {
	if op.Func == nil {
		return nil
	}

	return op.Exec(ec)[0].(Callable)
}

// execMustOne executes the LValuesOp and raises an exception if it does not
// evaluate to exactly one Variable. If the given LValuesOp is empty, it returns
// nil.
func (op LValuesOp) execMustOne(ec *Frame) vartypes.Variable {
	if op.Func == nil {
		return nil
	}
	variables := op.Exec(ec)
	if len(variables) != 1 {
		ec.errorpf(op.Begin, op.End, "should be one variable")
	}
	return variables[0]
}
