package eval

// Builtin special forms. Special forms behave mostly like ordinary commands -
// they are valid commands syntactically, and can take part in pipelines - but
// they have special rules for the evaluation of their arguments and can affect
// the compilation phase (whereas ordinary commands can only affect the
// evaluation phase).
//
// For example, the "and" special form evaluates its arguments from left to
// right, and stops as soon as one booleanly false value is obtained: the
// command "and $false (fail haha)" does not produce an exception.
//
// As another example, the "del" special form removes a variable, affecting the
// compiler.
//
// Flow control structures are also implemented as special forms in elvish, with
// closures functioning as code blocks.

import (
	"os"
	"path/filepath"
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
)

type compileBuiltin func(*compiler, *parse.Form) effectOp

var builtinSpecials map[string]compileBuiltin

// IsBuiltinSpecial is the set of all names of builtin special forms. It is
// intended for external consumption, e.g. the syntax highlighter.
var IsBuiltinSpecial = map[string]bool{}

// NoSuchModule encodes an error where a module spec cannot be resolved.
type NoSuchModule struct{ spec string }

// Error implements the error interface.
func (err NoSuchModule) Error() string { return "no such module: " + err.spec }

func init() {
	// Needed to avoid initialization loop
	builtinSpecials = map[string]compileBuiltin{
		"var": compileVar,
		"set": compileSet,
		"tmp": compileTmp,
		"del": compileDel,
		"fn":  compileFn,

		"use": compileUse,

		"and":      compileAnd,
		"or":       compileOr,
		"coalesce": compileCoalesce,

		"if":    compileIf,
		"while": compileWhile,
		"for":   compileFor,
		"try":   compileTry,

		"pragma": compilePragma,
	}
	for name := range builtinSpecials {
		IsBuiltinSpecial[name] = true
	}
}

// VarForm = 'var' { VariablePrimary } [ '=' { Compound } ]
func compileVar(cp *compiler, fn *parse.Form) effectOp {
	lhsArgs, rhs := compileLHSRHS(cp, fn)
	lhs := cp.parseCompoundLValues(lhsArgs, newLValue)
	if rhs == nil {
		// Just create new variables, nothing extra to do at runtime.
		return nopOp{}
	}
	return &assignOp{fn.Range(), lhs, rhs, false}
}

// SetForm = 'set' { LHS } '=' { Compound }
func compileSet(cp *compiler, fn *parse.Form) effectOp {
	lhs, rhs := compileSetArgs(cp, fn)
	return &assignOp{fn.Range(), lhs, rhs, false}
}

// TmpForm = 'tmp' { LHS } '=' { Compound }
func compileTmp(cp *compiler, fn *parse.Form) effectOp {
	if len(cp.scopes) <= 1 {
		cp.errorpf(fn, "tmp may only be used inside a function")
	}
	lhs, rhs := compileSetArgs(cp, fn)
	return &assignOp{fn.Range(), lhs, rhs, true}
}

func compileSetArgs(cp *compiler, fn *parse.Form) (lvaluesGroup, valuesOp) {
	lhsArgs, rhs := compileLHSRHS(cp, fn)
	if rhs == nil {
		cp.errorpf(diag.PointRanging(fn.Range().To), "need = and right-hand-side")
	}
	lhs := cp.parseCompoundLValues(lhsArgs, setLValue)
	return lhs, rhs
}

func compileLHSRHS(cp *compiler, fn *parse.Form) ([]*parse.Compound, valuesOp) {
	for i, cn := range fn.Args {
		if parse.SourceText(cn) == "=" {
			lhs := fn.Args[:i]
			if i == len(fn.Args)-1 {
				return lhs, nopValuesOp{diag.PointRanging(fn.Range().To)}
			}
			return lhs, seqValuesOp{
				diag.MixedRanging(fn.Args[i+1], fn.Args[len(fn.Args)-1]),
				cp.compoundOps(fn.Args[i+1:])}
		}
	}
	return fn.Args, nil
}

const delArgMsg = "arguments to del must be variable or variable elements"

// DelForm = 'del' { LHS }
func compileDel(cp *compiler, fn *parse.Form) effectOp {
	var ops []effectOp
	for _, cn := range fn.Args {
		if len(cn.Indexings) != 1 {
			cp.errorpf(cn, delArgMsg)
			continue
		}
		head, indices := cn.Indexings[0].Head, cn.Indexings[0].Indices
		if head.Type == parse.Variable {
			cp.errorpf(cn, "arguments to del must drop $")
		} else if !parse.ValidLHSVariable(head, false) {
			cp.errorpf(cn, delArgMsg)
		}

		qname := head.Value
		var f effectOp
		ref := resolveVarRef(cp, qname, nil)
		if ref == nil {
			cp.errorpf(cn, "no variable $%s", head.Value)
			continue
		}
		if len(indices) == 0 {
			if ref.scope == envScope {
				f = delEnvVarOp{fn.Range(), ref.subNames[0]}
			} else if ref.scope == localScope && len(ref.subNames) == 0 {
				f = delLocalVarOp{ref.index}
				cp.thisScope().infos[ref.index].deleted = true
			} else {
				cp.errorpf(cn, "only variables in the local scope or E: can be deleted")
				continue
			}
		} else {
			f = newDelElementOp(ref, head.Range().From, head.Range().To, cp.arrayOps(indices))
		}
		ops = append(ops, f)
	}
	return seqOp{ops}
}

type delLocalVarOp struct{ index int }

func (op delLocalVarOp) exec(fm *Frame) Exception {
	fm.local.slots[op.index] = nil
	return nil
}

type delEnvVarOp struct {
	diag.Ranging
	name string
}

func (op delEnvVarOp) exec(fm *Frame) Exception {
	return fm.errorp(op, os.Unsetenv(op.name))
}

func newDelElementOp(ref *varRef, begin, headEnd int, indexOps []valuesOp) effectOp {
	ends := make([]int, len(indexOps)+1)
	ends[0] = headEnd
	for i, op := range indexOps {
		ends[i+1] = op.Range().To
	}
	return &delElemOp{ref, indexOps, begin, ends}
}

type delElemOp struct {
	ref      *varRef
	indexOps []valuesOp
	begin    int
	ends     []int
}

func (op *delElemOp) Range() diag.Ranging {
	return diag.Ranging{From: op.begin, To: op.ends[0]}
}

func (op *delElemOp) exec(fm *Frame) Exception {
	var indices []interface{}
	for _, indexOp := range op.indexOps {
		indexValues, exc := indexOp.exec(fm)
		if exc != nil {
			return exc
		}
		if len(indexValues) != 1 {
			return fm.errorpf(indexOp, "index must evaluate to a single value in argument to del")
		}
		indices = append(indices, indexValues[0])
	}
	err := vars.DelElement(deref(fm, op.ref), indices)
	if err != nil {
		if level := vars.ElementErrorLevel(err); level >= 0 {
			return fm.errorp(diag.Ranging{From: op.begin, To: op.ends[level]}, err)
		}
		return fm.errorp(op, err)
	}
	return nil
}

// FnForm = 'fn' StringPrimary LambdaPrimary
//
// fn f { foobar } is a shorthand for set '&'f = { foobar }.
func compileFn(cp *compiler, fn *parse.Form) effectOp {
	args := cp.walkArgs(fn)
	nameNode := args.next()
	name := stringLiteralOrError(cp, nameNode, "function name")
	bodyNode := args.nextMustLambda("function body")
	args.mustEnd()

	// Define the variable before compiling the body, so that the body may refer
	// to the function itself.
	index := cp.thisScope().add(name + FnSuffix)
	op := cp.lambda(bodyNode)

	return fnOp{nameNode.Range(), index, op}
}

type fnOp struct {
	keywordRange diag.Ranging
	varIndex     int
	lambdaOp     valuesOp
}

func (op fnOp) exec(fm *Frame) Exception {
	// Initialize the function variable with the builtin nop function. This step
	// allows the definition of recursive functions; the actual function will
	// never be called.
	fm.local.slots[op.varIndex].Set(NewGoFn("<shouldn't be called>", nop))
	values, exc := op.lambdaOp.exec(fm)
	if exc != nil {
		return exc
	}
	c := values[0].(*Closure)
	c.op = fnWrap{c.op}
	return fm.errorp(op.keywordRange, fm.local.slots[op.varIndex].Set(c))
}

type fnWrap struct{ effectOp }

func (op fnWrap) Range() diag.Ranging { return op.effectOp.(diag.Ranger).Range() }

func (op fnWrap) exec(fm *Frame) Exception {
	exc := op.effectOp.exec(fm)
	if exc != nil && exc.Reason() != Return {
		// rethrow
		return exc
	}
	return nil
}

// UseForm = 'use' StringPrimary
func compileUse(cp *compiler, fn *parse.Form) effectOp {
	var name, spec string

	switch len(fn.Args) {
	case 0:
		end := fn.Head.Range().To
		cp.errorpf(diag.PointRanging(end), "lack module name")
	case 1:
		spec = stringLiteralOrError(cp, fn.Args[0], "module spec")
		// Use the last path component as the name; for instance, if path =
		// "a/b/c/d", name is "d". If path doesn't have slashes, name = path.
		name = spec[strings.LastIndexByte(spec, '/')+1:]
	case 2:
		// TODO(xiaq): Allow using variable as module path
		spec = stringLiteralOrError(cp, fn.Args[0], "module spec")
		name = stringLiteralOrError(cp, fn.Args[1], "module name")
	default: // > 2
		cp.errorpf(diag.MixedRanging(fn.Args[2], fn.Args[len(fn.Args)-1]),
			"superfluous argument(s)")
	}

	return useOp{fn.Range(), cp.thisScope().add(name + NsSuffix), spec}
}

type useOp struct {
	diag.Ranging
	varIndex int
	spec     string
}

func (op useOp) exec(fm *Frame) Exception {
	ns, err := use(fm, op.spec, op)
	if err != nil {
		return fm.errorp(op, err)
	}
	fm.local.slots[op.varIndex].Set(ns)
	return nil
}

// TODO: Add support for module specs relative to a package/workspace.
// See https://github.com/elves/elvish/issues/1421.
func use(fm *Frame, spec string, r diag.Ranger) (*Ns, error) {
	// Handle relative imports. Note that this deliberately does not support Windows backslash as a
	// path separator because module specs are meant to be platform independent. If necessary, we
	// translate a module spec to an appropriate path for the platform.
	if strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../") {
		var dir string
		if fm.srcMeta.IsFile {
			dir = filepath.Dir(fm.srcMeta.Name)
		} else {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return nil, err
			}
		}
		path := filepath.Clean(dir + "/" + spec)
		return useFromFile(fm, spec, path, r)
	}

	// Handle imports of pre-defined modules like `builtin` and `str`.
	if ns, ok := fm.Evaler.modules[spec]; ok {
		return ns, nil
	}
	if code, ok := fm.Evaler.BundledModules[spec]; ok {
		return evalModule(fm, spec,
			parse.Source{Name: "[bundled " + spec + "]", Code: code}, r)
	}

	// Handle imports relative to the Elvish module search directories.
	//
	// TODO: For non-relative imports, use the spec (instead of the full path)
	// as the module key instead to avoid searching every time.
	for _, dir := range fm.Evaler.LibDirs {
		ns, err := useFromFile(fm, spec, filepath.Join(dir, spec), r)
		if _, isNoSuchModule := err.(NoSuchModule); isNoSuchModule {
			continue
		}
		return ns, err
	}

	// Sadly, we couldn't resolve the module spec.
	return nil, NoSuchModule{spec}
}

// TODO: Make access to fm.Evaler.modules concurrency-safe.
func useFromFile(fm *Frame, spec, path string, r diag.Ranger) (*Ns, error) {
	if ns, ok := fm.Evaler.modules[path]; ok {
		return ns, nil
	}
	_, err := os.Stat(path + ".so")
	if err != nil {
		code, err := readFileUTF8(path + ".elv")
		if err != nil {
			if os.IsNotExist(err) {
				return nil, NoSuchModule{spec}
			}
			return nil, err
		}
		src := parse.Source{Name: path + ".elv", Code: code, IsFile: true}
		return evalModule(fm, path, src, r)
	}

	plug, err := pluginOpen(path + ".so")
	if err != nil {
		return nil, NoSuchModule{spec}
	}
	sym, err := plug.Lookup("Ns")
	if err != nil {
		return nil, err
	}
	ns, ok := sym.(**Ns)
	if !ok {
		return nil, NoSuchModule{spec}
	}
	fm.Evaler.modules[path] = *ns
	return *ns, nil
}

// TODO: Make access to fm.Evaler.modules concurrency-safe.
func evalModule(fm *Frame, key string, src parse.Source, r diag.Ranger) (*Ns, error) {
	ns, exec, err := fm.PrepareEval(src, r, new(Ns))
	if err != nil {
		return nil, err
	}
	// Installs the namespace before executing. This prevent circular use'es
	// from resulting in an infinite recursion.
	fm.Evaler.modules[key] = ns
	err = exec()
	if err != nil {
		// Unload the namespace.
		delete(fm.Evaler.modules, key)
		return nil, err
	}
	return ns, nil
}

// compileAnd compiles the "and" special form.
//
// The and special form evaluates arguments until a false-ish values is found
// and outputs it; the remaining arguments are not evaluated. If there are no
// false-ish values, the last value is output. If there are no arguments, it
// outputs $true, as if there is a hidden $true before actual arguments.
func compileAnd(cp *compiler, fn *parse.Form) effectOp {
	return &andOrOp{fn.Range(), cp.compoundOps(fn.Args), true, false}
}

// compileOr compiles the "or" special form.
//
// The or special form evaluates arguments until a true-ish values is found and
// outputs it; the remaining arguments are not evaluated. If there are no
// true-ish values, the last value is output. If there are no arguments, it
// outputs $false, as if there is a hidden $false before actual arguments.
func compileOr(cp *compiler, fn *parse.Form) effectOp {
	return &andOrOp{fn.Range(), cp.compoundOps(fn.Args), false, true}
}

type andOrOp struct {
	diag.Ranging
	argOps []valuesOp
	init   bool
	stopAt bool
}

func (op *andOrOp) exec(fm *Frame) Exception {
	var lastValue interface{} = vals.Bool(op.init)
	out := fm.ValueOutput()
	for _, argOp := range op.argOps {
		values, exc := argOp.exec(fm)
		if exc != nil {
			return exc
		}
		for _, value := range values {
			if vals.Bool(value) == op.stopAt {
				return fm.errorp(op, out.Put(value))
			}
			lastValue = value
		}
	}
	return fm.errorp(op, out.Put(lastValue))
}

// Compiles the "coalesce" special form, which is like "or", but evaluates until
// a non-nil value is found.
func compileCoalesce(cp *compiler, fn *parse.Form) effectOp {
	return &coalesceOp{fn.Range(), cp.compoundOps(fn.Args)}
}

type coalesceOp struct {
	diag.Ranging
	argOps []valuesOp
}

func (op *coalesceOp) exec(fm *Frame) Exception {
	out := fm.ValueOutput()
	for _, argOp := range op.argOps {
		values, exc := argOp.exec(fm)
		if exc != nil {
			return exc
		}
		for _, value := range values {
			if value != nil {
				return fm.errorp(op, out.Put(value))
			}
		}
	}
	return fm.errorp(op, out.Put(nil))
}

func compileIf(cp *compiler, fn *parse.Form) effectOp {
	args := cp.walkArgs(fn)
	var condNodes []*parse.Compound
	var bodyNodes []*parse.Primary
	condLeader := "if"
	for {
		condNodes = append(condNodes, args.next())
		bodyNodes = append(bodyNodes, args.nextMustThunk(condLeader))
		if !args.nextIs("elif") {
			break
		}
		condLeader = "elif"
	}
	elseNode := args.nextMustThunkIfAfter("else")
	args.mustEnd()

	condOps := cp.compoundOps(condNodes)
	bodyOps := cp.primaryOps(bodyNodes)
	var elseOp valuesOp
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}

	return &ifOp{fn.Range(), condOps, bodyOps, elseOp}
}

type ifOp struct {
	diag.Ranging
	condOps []valuesOp
	bodyOps []valuesOp
	elseOp  valuesOp
}

func (op *ifOp) exec(fm *Frame) Exception {
	bodies := make([]Callable, len(op.bodyOps))
	for i, bodyOp := range op.bodyOps {
		bodies[i] = execLambdaOp(fm, bodyOp)
	}
	elseFn := execLambdaOp(fm, op.elseOp)
	for i, condOp := range op.condOps {
		condValues, exc := condOp.exec(fm.Fork("if cond"))
		if exc != nil {
			return exc
		}
		if allTrue(condValues) {
			return fm.errorp(op, bodies[i].Call(fm.Fork("if body"), NoArgs, NoOpts))
		}
	}
	if op.elseOp != nil {
		return fm.errorp(op, elseFn.Call(fm.Fork("if else"), NoArgs, NoOpts))
	}
	return nil
}

func compileWhile(cp *compiler, fn *parse.Form) effectOp {
	args := cp.walkArgs(fn)
	condNode := args.next()
	bodyNode := args.nextMustThunk("while body")
	elseNode := args.nextMustThunkIfAfter("else")
	args.mustEnd()

	condOp := cp.compoundOp(condNode)
	bodyOp := cp.primaryOp(bodyNode)
	var elseOp valuesOp
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}

	return &whileOp{fn.Range(), condOp, bodyOp, elseOp}
}

type whileOp struct {
	diag.Ranging
	condOp, bodyOp, elseOp valuesOp
}

func (op *whileOp) exec(fm *Frame) Exception {
	body := execLambdaOp(fm, op.bodyOp)
	elseBody := execLambdaOp(fm, op.elseOp)

	iterated := false
	for {
		condValues, exc := op.condOp.exec(fm.Fork("while cond"))
		if exc != nil {
			return exc
		}
		if !allTrue(condValues) {
			break
		}
		iterated = true
		err := body.Call(fm.Fork("while"), NoArgs, NoOpts)
		if err != nil {
			exc := err.(Exception)
			if exc.Reason() == Continue {
				// Do nothing
			} else if exc.Reason() == Break {
				break
			} else {
				return exc
			}
		}
	}

	if op.elseOp != nil && !iterated {
		return fm.errorp(op, elseBody.Call(fm.Fork("while else"), NoArgs, NoOpts))
	}
	return nil
}

func compileFor(cp *compiler, fn *parse.Form) effectOp {
	args := cp.walkArgs(fn)
	varNode := args.next()
	iterNode := args.next()
	bodyNode := args.nextMustThunk("for body")
	elseNode := args.nextMustThunkIfAfter("else")
	args.mustEnd()

	lvalue := cp.compileOneLValue(varNode, setLValue|newLValue)

	iterOp := cp.compoundOp(iterNode)
	bodyOp := cp.primaryOp(bodyNode)
	var elseOp valuesOp
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}

	return &forOp{fn.Range(), lvalue, iterOp, bodyOp, elseOp}
}

type forOp struct {
	diag.Ranging
	lvalue lvalue
	iterOp valuesOp
	bodyOp valuesOp
	elseOp valuesOp
}

func (op *forOp) exec(fm *Frame) Exception {
	variable, err := derefLValue(fm, op.lvalue)
	if err != nil {
		return fm.errorp(op.lvalue, err)
	}
	iterable, err := evalForValue(fm, op.iterOp, "value being iterated")
	if err != nil {
		return fm.errorp(op, err)
	}

	body := execLambdaOp(fm, op.bodyOp)
	elseBody := execLambdaOp(fm, op.elseOp)

	iterated := false
	var errElement error
	errIterate := vals.Iterate(iterable, func(v interface{}) bool {
		iterated = true
		err := variable.Set(v)
		if err != nil {
			errElement = err
			return false
		}
		err = body.Call(fm.Fork("for"), NoArgs, NoOpts)
		if err != nil {
			exc := err.(Exception)
			if exc.Reason() == Continue {
				// do nothing
			} else if exc.Reason() == Break {
				return false
			} else {
				errElement = err
				return false
			}
		}
		return true
	})
	if errIterate != nil {
		return fm.errorp(op, errIterate)
	}
	if errElement != nil {
		return fm.errorp(op, errElement)
	}

	if !iterated && elseBody != nil {
		return fm.errorp(op, elseBody.Call(fm.Fork("for else"), NoArgs, NoOpts))
	}
	return nil
}

func compileTry(cp *compiler, fn *parse.Form) effectOp {
	logger.Println("compiling try")
	args := cp.walkArgs(fn)
	bodyNode := args.nextMustThunk("try body")
	logger.Printf("body is %q", parse.SourceText(bodyNode))
	var catchVarNode *parse.Compound
	var catchNode *parse.Primary
	if args.peekIs("except") {
		cp.deprecate(args.peek(),
			`"except" is deprecated; use "catch" instead`, 18)
	}
	if args.nextIs("except") || args.nextIs("catch") {
		// Parse an optional lvalue into exceptVarNode.
		n := args.peek()
		if _, ok := cmpd.StringLiteral(n); ok {
			catchVarNode = n
			args.next()
		}
		catchNode = args.nextMustThunk("catch body")
	}
	elseNode := args.nextMustThunkIfAfter("else")
	finallyNode := args.nextMustThunkIfAfter("finally")
	args.mustEnd()

	if catchNode == nil && finallyNode == nil {
		cp.errorpf(fn, "try must be followed by a catch block or a finally block")
	}

	var catchVar lvalue
	var bodyOp, catchOp, elseOp, finallyOp valuesOp
	bodyOp = cp.primaryOp(bodyNode)
	if catchVarNode != nil {
		catchVar = cp.compileOneLValue(catchVarNode, setLValue|newLValue)
	}
	if catchNode != nil {
		catchOp = cp.primaryOp(catchNode)
	}
	if elseNode != nil {
		elseOp = cp.primaryOp(elseNode)
	}
	if finallyNode != nil {
		finallyOp = cp.primaryOp(finallyNode)
	}

	return &tryOp{fn.Range(), bodyOp, catchVar, catchOp, elseOp, finallyOp}
}

type tryOp struct {
	diag.Ranging
	bodyOp    valuesOp
	catchVar  lvalue
	catchOp   valuesOp
	elseOp    valuesOp
	finallyOp valuesOp
}

func (op *tryOp) exec(fm *Frame) Exception {
	body := execLambdaOp(fm, op.bodyOp)
	var exceptVar vars.Var
	if op.catchVar.ref != nil {
		var err error
		exceptVar, err = derefLValue(fm, op.catchVar)
		if err != nil {
			return fm.errorp(op, err)
		}
	}
	catch := execLambdaOp(fm, op.catchOp)
	elseFn := execLambdaOp(fm, op.elseOp)
	finally := execLambdaOp(fm, op.finallyOp)

	err := body.Call(fm.Fork("try body"), NoArgs, NoOpts)
	if err != nil {
		if catch != nil {
			if exceptVar != nil {
				err := exceptVar.Set(err.(Exception))
				if err != nil {
					return fm.errorp(op.catchVar, err)
				}
			}
			err = catch.Call(fm.Fork("try catch"), NoArgs, NoOpts)
		}
	} else {
		if elseFn != nil {
			err = elseFn.Call(fm.Fork("try else"), NoArgs, NoOpts)
		}
	}
	if finally != nil {
		errFinally := finally.Call(fm.Fork("try finally"), NoArgs, NoOpts)
		if errFinally != nil {
			// TODO: If err is not nil, this discards err. Use something similar
			// to pipeline exception to expose both.
			return fm.errorp(op, errFinally)
		}
	}
	return fm.errorp(op, err)
}

// PragmaForm = 'pragma' 'fallback-resolver' '=' { Compound }
func compilePragma(cp *compiler, fn *parse.Form) effectOp {
	args := cp.walkArgs(fn)
	nameNode := args.next()
	name := stringLiteralOrError(cp, nameNode, "pragma name")
	eqNode := args.next()
	eq := stringLiteralOrError(cp, eqNode, "literal =")
	if eq != "=" {
		cp.errorpf(eqNode, "must be literal =")
	}
	valueNode := args.next()
	args.mustEnd()

	switch name {
	case "unknown-command":
		value := stringLiteralOrError(cp, valueNode, "value for unknown-command")
		switch value {
		case "disallow":
			cp.currentPragma().unknownCommandIsExternal = false
		case "external":
			cp.currentPragma().unknownCommandIsExternal = true
		default:
			cp.errorpf(valueNode,
				"invalid value for unknown-command: %s", parse.Quote(value))
		}
	default:
		cp.errorpf(nameNode, "unknown pragma %s", parse.Quote(name))
	}
	return nopOp{}
}

func (cp *compiler) compileOneLValue(n *parse.Compound, f lvalueFlag) lvalue {
	if len(n.Indexings) != 1 {
		cp.errorpf(n, "must be valid lvalue")
	}
	lvalues := cp.parseIndexingLValue(n.Indexings[0], f)
	if lvalues.rest != -1 {
		cp.errorpf(lvalues.lvalues[lvalues.rest], "rest variable not allowed")
	}
	if len(lvalues.lvalues) != 1 {
		cp.errorpf(n, "must be exactly one lvalue")
	}
	return lvalues.lvalues[0]
}

// Executes a valuesOp that is known to yield a lambda and returns the lambda.
// Returns nil if op is nil.
func execLambdaOp(fm *Frame, op valuesOp) Callable {
	if op == nil {
		return nil
	}
	values, exc := op.exec(fm)
	if exc != nil {
		panic("must not be erroneous")
	}
	return values[0].(Callable)
}
