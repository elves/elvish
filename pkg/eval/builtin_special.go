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
	"os"
	"path/filepath"
	"strings"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
)

type compileBuiltin func(*compiler, *parse.Form) effectOpBody

var builtinSpecials map[string]compileBuiltin

// IsBuiltinSpecial is the set of all names of builtin special forms. It is
// intended for external consumption, e.g. the syntax highlighter.
var IsBuiltinSpecial = map[string]bool{}

type noSuchModule struct{ spec string }

func (err noSuchModule) Error() string { return "no such module: " + err.spec }

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

const delArgMsg = "arguments to del must be variable or variable elements"

// DelForm = 'del' { VariablePrimary }
func compileDel(cp *compiler, fn *parse.Form) effectOpBody {
	var ops []effectOp
	for _, cn := range fn.Args {
		if len(cn.Indexings) != 1 {
			cp.errorpf(cn, delArgMsg)
			continue
		}
		head, indicies := cn.Indexings[0].Head, cn.Indexings[0].Indicies
		if head.Type != parse.Bareword {
			if head.Type == parse.Variable {
				cp.errorpf(cn, "arguments to del must drop $")
			} else {
				cp.errorpf(cn, delArgMsg)
			}
			continue
		}

		sigil, qname := SplitVariableRef(head.Value)
		if sigil != "" {
			cp.errorpf(cn, "arguments to del may not have a sigils, got %q", sigil)
			continue
		}
		var f effectOpBody
		if len(indicies) == 0 {
			ns, name := SplitQNameNsFirst(qname)
			switch ns {
			case "", ":", "local:":
				if !cp.thisScope().has(name) {
					cp.errorpf(cn, "no variable $%s in local scope", name)
					continue
				}
				cp.thisScope().del(name)
				f = delLocalVarOp{name}
			case "E:":
				f = delEnvVarOp{name}
			default:
				cp.errorpf(cn, "only variables in local: or E: can be deleted")
				continue
			}
		} else {
			if !cp.registerVariableGet(qname) {
				cp.errorpf(cn, "no variable $%s", head.Value)
				continue
			}
			f = newDelElementOp(qname, head.Range().From, head.Range().To, cp.arrayOps(indicies))
		}
		ops = append(ops, effectOp{f, cn.Range()})
	}
	return seqOp{ops}
}

type delLocalVarOp struct{ name string }

func (op delLocalVarOp) invoke(fm *Frame) error {
	delete(fm.local, op.name)
	return nil
}

type delEnvVarOp struct{ name string }

func (op delEnvVarOp) invoke(*Frame) error {
	return os.Unsetenv(op.name)
}

func newDelElementOp(qname string, begin, headEnd int, indexOps []valuesOp) effectOpBody {
	ends := make([]int, len(indexOps)+1)
	ends[0] = headEnd
	for i, op := range indexOps {
		ends[i+1] = op.To
	}
	return &delElemOp{qname, indexOps, begin, ends}
}

type delElemOp struct {
	qname    string
	indexOps []valuesOp
	begin    int
	ends     []int
}

func (op *delElemOp) invoke(fm *Frame) error {
	var indicies []interface{}
	for _, indexOp := range op.indexOps {
		indexValues, err := indexOp.exec(fm)
		if err != nil {
			return err
		}
		if len(indexValues) != 1 {
			return fm.errorpf(indexOp.From, indexOp.To, "index must evaluate to a single value in argument to del")
		}
		indicies = append(indicies, indexValues[0])
	}
	err := vars.DelElement(fm.ResolveVar(op.qname), indicies)
	if err != nil {
		if level := vars.ElementErrorLevel(err); level >= 0 {
			return fm.errorpf(op.begin, op.ends[level], "%s", err.Error())
		}
		return err
	}
	return nil
}

// FnForm = 'fn' StringPrimary LambdaPrimary
//
// fn f []{foobar} is a shorthand for set '&'f = []{foobar}.
func compileFn(cp *compiler, fn *parse.Form) effectOpBody {
	args := cp.walkArgs(fn)
	nameNode := args.next()
	varName := mustString(cp, nameNode, "must be a literal string") + FnSuffix
	bodyNode := args.nextMustLambda()
	args.mustEnd()

	cp.registerVariableSet(":" + varName)
	op := cp.lambda(bodyNode)

	return fnOp{varName, op}
}

type fnOp struct {
	varName  string
	lambdaOp valuesOpBody
}

func (op fnOp) invoke(fm *Frame) error {
	// Initialize the function variable with the builtin nop function. This step
	// allows the definition of recursive functions; the actual function will
	// never be called.
	fm.local[op.varName] = vars.FromInit(NewGoFn("<shouldn't be called>", nop))
	values, err := op.lambdaOp.invoke(fm)
	if err != nil {
		return err
	}
	closure := values[0].(*Closure)
	closure.Op = wrapFn(closure.Op)
	return fm.local[op.varName].Set(closure)
}

func wrapFn(op effectOp) effectOp {
	return effectOp{fnWrap{op}, op.Ranging}
}

type fnWrap struct{ wrapped effectOp }

func (op fnWrap) invoke(fm *Frame) error {
	err := fm.eval(op.wrapped)
	if err != nil && Cause(err) != Return {
		// rethrow
		return err
	}
	return nil
}

// UseForm = 'use' StringPrimary
func compileUse(cp *compiler, fn *parse.Form) effectOpBody {
	var name, spec string

	switch len(fn.Args) {
	case 0:
		end := fn.Head.Range().To
		cp.errorpf(diag.PointRanging(end), "lack module name")
	case 1:
		spec = mustString(cp, fn.Args[0],
			"module spec should be a literal string")
		// Use the last path component as the name; for instance, if path =
		// "a/b/c/d", name is "d". If path doesn't have slashes, name = path.
		name = spec[strings.LastIndexByte(spec, '/')+1:]
	case 2:
		// TODO(xiaq): Allow using variable as module path
		spec = mustString(cp, fn.Args[0],
			"module spec should be a literal string")
		name = mustString(cp, fn.Args[1],
			"module name should be a literal string")
	default: // > 2
		cp.errorpf(diag.MixedRanging(fn.Args[2], fn.Args[len(fn.Args)-1]),
			"superfluous argument(s)")
	}

	cp.thisScope().set(name + NsSuffix)

	return useOp{name, spec}
}

type useOp struct{ name, spec string }

func (op useOp) invoke(fm *Frame) error {
	ns, err := loadModule(fm, op.spec)
	if err != nil {
		return err
	}
	fm.local.AddNs(op.name, ns)
	return nil
}

func loadModule(fm *Frame, spec string) (Ns, error) {
	if strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../") {
		var dir string
		if fm.srcMeta.Type == FileSource {
			dir = filepath.Dir(fm.srcMeta.Name)
		} else {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return nil, err
			}
		}
		path := filepath.Clean(dir + "/" + spec + ".elv")
		return loadModuleFile(fm, spec, path)
	}
	if ns, ok := fm.Evaler.modules[spec]; ok {
		return ns, nil
	}
	if code, ok := fm.bundled[spec]; ok {
		return evalModule(fm, spec, NewInternalElvishSource(false, spec, code))
	}
	if fm.libDir == "" {
		return nil, noSuchModule{spec}
	}
	return loadModuleFile(fm, spec, fm.libDir+"/"+spec+".elv")
}

func loadModuleFile(fm *Frame, spec, path string) (Ns, error) {
	if ns, ok := fm.modules[path]; ok {
		return ns, nil
	}
	code, err := readFileUTF8(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, noSuchModule{spec}
		}
		return nil, err
	}
	return evalModule(fm, path, NewModuleSource(path, code))
}

func evalModule(fm *Frame, key string, src *Source) (Ns, error) {
	n, err := parse.AsChunk(src.Name, src.Code)
	if err != nil {
		return nil, err
	}

	// Make an empty scope to evaluate the module in.
	modGlobal := Ns{}

	newFm := &Frame{
		fm.Evaler, src,
		diag.Ranging{From: 0, To: len(src.Code)},
		modGlobal, make(Ns),
		fm.intCh, fm.ports,
		fm.addTraceback(), false,
	}

	op, err := compile(newFm.Builtin.static(), modGlobal.static(), n, src)
	if err != nil {
		return nil, err
	}

	// Load the namespace before executing. This prevent circular "use"es from
	// resulting in an infinite recursion.
	fm.Evaler.modules[key] = modGlobal
	err = newFm.Eval(op)
	if err != nil {
		// Unload the namespace.
		delete(fm.modules, key)
		return nil, err
	}
	return modGlobal, nil
}

// compileAnd compiles the "and" special form.
//
// The and special form evaluates arguments until a false-ish values is found
// and outputs it; the remaining arguments are not evaluated. If there are no
// false-ish values, the last value is output. If there are no arguments, it
// outputs $true, as if there is a hidden $true before actual arguments.
func compileAnd(cp *compiler, fn *parse.Form) effectOpBody {
	return &andOrOp{cp.compoundOps(fn.Args), true, false}
}

// compileOr compiles the "or" special form.
//
// The or special form evaluates arguments until a true-ish values is found and
// outputs it; the remaining arguments are not evaluated. If there are no
// true-ish values, the last value is output. If there are no arguments, it
// outputs $false, as if there is a hidden $false before actual arguments.
func compileOr(cp *compiler, fn *parse.Form) effectOpBody {
	return &andOrOp{cp.compoundOps(fn.Args), false, true}
}

type andOrOp struct {
	argOps []valuesOp
	init   bool
	stopAt bool
}

func (op *andOrOp) invoke(fm *Frame) error {
	var lastValue interface{} = vals.Bool(op.init)
	for _, argOp := range op.argOps {
		values, err := argOp.exec(fm)
		if err != nil {
			return err
		}
		for _, value := range values {
			if vals.Bool(value) == op.stopAt {
				fm.OutputChan() <- value
				return nil
			}
			lastValue = value
		}
	}
	fm.OutputChan() <- lastValue
	return nil
}

func compileIf(cp *compiler, fn *parse.Form) effectOpBody {
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
	var elseOp valuesOp
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}

	return &ifOp{condOps, bodyOps, elseOp}
}

type ifOp struct {
	condOps []valuesOp
	bodyOps []valuesOp
	elseOp  valuesOp
}

func (op *ifOp) invoke(fm *Frame) error {
	bodies := make([]Callable, len(op.bodyOps))
	for i, bodyOp := range op.bodyOps {
		bodies[i] = bodyOp.execlambdaOp(fm)
	}
	elseFn := op.elseOp.execlambdaOp(fm)
	for i, condOp := range op.condOps {
		condValues, err := condOp.exec(fm.fork("if cond"))
		if err != nil {
			return err
		}
		if allTrue(condValues) {
			return bodies[i].Call(fm.fork("if body"), NoArgs, NoOpts)
		}
	}
	if op.elseOp.body != nil {
		return elseFn.Call(fm.fork("if else"), NoArgs, NoOpts)
	}
	return nil
}

func compileWhile(cp *compiler, fn *parse.Form) effectOpBody {
	args := cp.walkArgs(fn)
	condNode := args.next()
	bodyNode := args.nextMustLambda()
	elseNode := args.nextMustLambdaIfAfter("else")
	args.mustEnd()

	condOp := cp.compoundOp(condNode)
	bodyOp := cp.primaryOp(bodyNode)
	var elseOp valuesOp
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}

	return &whileOp{condOp, bodyOp, elseOp}
}

type whileOp struct {
	condOp, bodyOp, elseOp valuesOp
}

func (op *whileOp) invoke(fm *Frame) error {
	body := op.bodyOp.execlambdaOp(fm)
	elseBody := op.elseOp.execlambdaOp(fm)

	iterated := false
	for {
		condValues, err := op.condOp.exec(fm.fork("while cond"))
		if err != nil {
			return err
		}
		if !allTrue(condValues) {
			break
		}
		iterated = true
		err = fm.fork("while").Call(body, NoArgs, NoOpts)
		if err != nil {
			exc := err.(*Exception)
			if exc.Cause == Continue {
				// do nothing
			} else if exc.Cause == Break {
				break
			} else {
				return err
			}
		}
	}

	if op.elseOp.body != nil && !iterated {
		return elseBody.Call(fm.fork("while else"), NoArgs, NoOpts)
	}
	return nil
}

func compileFor(cp *compiler, fn *parse.Form) effectOpBody {
	args := cp.walkArgs(fn)
	varNode := args.next()
	iterNode := args.next()
	bodyNode := args.nextMustLambda()
	elseNode := args.nextMustLambdaIfAfter("else")
	args.mustEnd()

	varOp, restOp := cp.lvaluesOp(varNode.Indexings[0])
	if restOp.body != nil {
		cp.errorpf(restOp, "rest not allowed")
	}

	iterOp := cp.compoundOp(iterNode)
	bodyOp := cp.primaryOp(bodyNode)
	var elseOp valuesOp
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}

	return &forOp{varOp, iterOp, bodyOp, elseOp}
}

type forOp struct {
	varOp  lvaluesOp
	iterOp valuesOp
	bodyOp valuesOp
	elseOp valuesOp
}

func (op *forOp) invoke(fm *Frame) error {
	variables, err := op.varOp.exec(fm)
	if err != nil {
		return err
	}
	if len(variables) != 1 {
		return fm.errorpf(op.varOp.From, op.varOp.To, "only one variable allowed")
	}
	variable := variables[0]
	iterable, err := fm.ExecAndUnwrap("value being iterated", op.iterOp).One().Any()
	if err != nil {
		return err
	}

	body := op.bodyOp.execlambdaOp(fm)
	elseBody := op.elseOp.execlambdaOp(fm)

	iterated := false
	var errElement error
	errIterate := vals.Iterate(iterable, func(v interface{}) bool {
		iterated = true
		err := variable.Set(v)
		if err != nil {
			errElement = err
			return false
		}
		err = fm.fork("for").Call(body, NoArgs, NoOpts)
		if err != nil {
			exc := err.(*Exception)
			if exc.Cause == Continue {
				// do nothing
			} else if exc.Cause == Break {
				return false
			} else {
				errElement = err
				return false
			}
		}
		return true
	})
	if errIterate != nil {
		return errIterate
	}
	if errElement != nil {
		return errElement
	}

	if !iterated && elseBody != nil {
		return elseBody.Call(fm.fork("for else"), NoArgs, NoOpts)
	}
	return nil
}

func compileTry(cp *compiler, fn *parse.Form) effectOpBody {
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

	var exceptVarOp lvaluesOp
	var bodyOp, exceptOp, elseOp, finallyOp valuesOp
	bodyOp = cp.primaryOp(bodyNode)
	if exceptVarNode != nil {
		var restOp lvaluesOp
		exceptVarOp, restOp = cp.lvaluesOp(exceptVarNode)
		if restOp.body != nil {
			cp.errorpf(restOp, "may not use @rest in except variable")
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

	return &tryOp{bodyOp, exceptVarOp, exceptOp, elseOp, finallyOp}
}

type tryOp struct {
	bodyOp      valuesOp
	exceptVarOp lvaluesOp
	exceptOp    valuesOp
	elseOp      valuesOp
	finallyOp   valuesOp
}

func (op *tryOp) invoke(fm *Frame) error {
	body := op.bodyOp.execlambdaOp(fm)
	exceptVar, err := op.exceptVarOp.execMustOne(fm)
	if err != nil {
		return err
	}
	except := op.exceptOp.execlambdaOp(fm)
	elseFn := op.elseOp.execlambdaOp(fm)
	finally := op.finallyOp.execlambdaOp(fm)

	err = fm.fork("try body").Call(body, NoArgs, NoOpts)
	if err != nil {
		if except != nil {
			if exceptVar != nil {
				err := exceptVar.Set(err.(*Exception))
				if err != nil {
					return err
				}
			}
			err = fm.fork("try except").Call(except, NoArgs, NoOpts)
		}
	} else {
		if elseFn != nil {
			err = fm.fork("try else").Call(elseFn, NoArgs, NoOpts)
		}
	}
	if finally != nil {
		errFinally := finally.Call(fm.fork("try finally"), NoArgs, NoOpts)
		if errFinally != nil {
			// TODO: If err is not nil, this discards err. Use something similar
			// to pipeline exception to expose both.
			return errFinally
		}
	}
	return err
}

// execLambdaOp executes a ValuesOp that is known to yield a lambda and returns
// the lambda. If the ValuesOp is empty, it returns a nil.
func (op valuesOp) execlambdaOp(fm *Frame) Callable {
	if op.body == nil {
		return nil
	}

	values, err := op.exec(fm)
	if err != nil {
		panic("must not be erroneous")
	}
	return values[0].(Callable)
}

// execMustOne executes the LValuesOp and returns an error if it does not
// evaluate to exactly one Variable. If the given LValuesOp is empty, it returns
// nil.
func (op lvaluesOp) execMustOne(fm *Frame) (vars.Var, error) {
	if op.body == nil {
		return nil, nil
	}
	variables, err := op.exec(fm)
	if err != nil {
		return nil, err
	}
	if len(variables) != 1 {
		return nil, fm.errorpf(op.From, op.To, "should be one variable")
	}
	return variables[0], nil
}
