package eval

import (
	"fmt"
	"os"
	"sync"

	"github.com/elves/elvish/parse"
)

// Op is an operation on an EvalCtx.
type Op struct {
	Func       OpFunc
	Begin, End int
}

// OpFunc is the body of an Op.
type OpFunc func(*EvalCtx)

// Exec executes an Op.
func (op Op) Exec(ec *EvalCtx) {
	ec.begin, ec.end = op.Begin, op.End
	op.Func(ec)
}

func (cp *compiler) chunk(n *parse.Chunk) OpFunc {
	ops := cp.pipelineOps(n.Pipelines)

	return func(ec *EvalCtx) {
		for _, op := range ops {
			op.Exec(ec)
		}
		// Check for interrupts after the chunk.
		// We also check for interrupts before each pipeline, so there is no
		// need to check it before the chunk or after each pipeline.
		ec.CheckInterrupts()
	}
}

const pipelineChanBufferSize = 32

func (cp *compiler) pipeline(n *parse.Pipeline) OpFunc {
	ops := cp.formOps(n.Forms)

	return func(ec *EvalCtx) {
		ec.CheckInterrupts()

		bg := n.Background
		if bg {
			ec = ec.fork("background job " + n.SourceText())
			ec.intCh = nil
			ec.background = true

			if ec.Editor != nil {
				// TODO: Redirect output in interactive mode so that the line
				// editor does not get messed up.
			}
		}

		nforms := len(ops)

		var wg sync.WaitGroup
		wg.Add(nforms)
		errors := make([]*Exception, nforms)

		var nextIn *Port

		// For each form, create a dedicated evalCtx and run asynchronously
		for i, op := range ops {
			newEc := ec.fork(fmt.Sprintf("form op %v", op))
			if i > 0 {
				newEc.ports[0] = nextIn
			}
			if i < nforms-1 {
				// Each internal port pair consists of a (byte) pipe pair and a
				// channel.
				// os.Pipe sets O_CLOEXEC, which is what we want.
				reader, writer, e := os.Pipe()
				if e != nil {
					throwf("failed to create pipe: %s", e)
				}
				ch := make(chan Value, pipelineChanBufferSize)
				newEc.ports[1] = &Port{
					File: writer, Chan: ch, CloseFile: true, CloseChan: true}
				nextIn = &Port{
					File: reader, Chan: ch, CloseFile: true, CloseChan: false}
			}
			thisOp := op
			thisError := &errors[i]
			go func() {
				err := newEc.PEval(thisOp)
				// Logger.Printf("closing ports of %s", newEc.context)
				ClosePorts(newEc.ports)
				if err != nil {
					*thisError = err.(*Exception)
				}
				wg.Done()
			}()
		}

		if bg {
			// Background job, wait for form termination asynchronously.
			go func() {
				wg.Wait()
				msg := "job " + n.SourceText() + " finished"
				err := ComposeExceptionsFromPipeline(errors)
				if err != nil {
					msg += ", errors = " + err.Error()
				}
				if ec.Editor != nil {
					m := ec.Editor.ActiveMutex()
					m.Lock()
					defer m.Unlock()

					if ec.Editor.Active() {
						ec.Editor.Notify("%s", msg)
					} else {
						ec.ports[2].File.WriteString(msg + "\n")
					}
				} else {
					ec.ports[2].File.WriteString(msg + "\n")
				}
			}()
		} else {
			wg.Wait()
			maybeThrow(ComposeExceptionsFromPipeline(errors))
		}
	}
}

func (cp *compiler) form(n *parse.Form) OpFunc {
	var saveVarsOps []LValuesOp
	var assignmentOps []Op
	if len(n.Assignments) > 0 {
		assignmentOps = cp.assignmentOps(n.Assignments)
		if n.Head == nil && n.Vars == nil {
			// Permanent assignment.
			return func(ec *EvalCtx) {
				for _, op := range assignmentOps {
					op.Exec(ec)
				}
			}
		}
		for _, a := range n.Assignments {
			v, r := cp.lvaluesOp(a.Left)
			saveVarsOps = append(saveVarsOps, v, r)
		}
		logger.Println("temporary assignment of", len(n.Assignments), "pairs")
	}

	if n.Head != nil {
		headStr, ok := oneString(n.Head)
		if ok {
			compileForm, ok := builtinSpecials[headStr]
			if ok {
				// special form
				return compileForm(cp, n)
			}
			// Ignore the output. If a matching function exists it will be
			// captured and eventually the Evaler executes it. If not, nothing
			// happens here and the Evaler executes an external command.
			cp.registerVariableGet(FnPrefix + headStr)
			// XXX Dynamic head names should always refer to external commands
		}
	}

	argOps := cp.compoundOps(n.Args)

	var headOp ValuesOp
	var spaceyAssignOp Op
	if n.Head != nil {
		headOp = cp.compoundOp(n.Head)
	} else {
		varsOp, restOp := cp.lvaluesMulti(n.Vars)
		argsOp := ValuesOp{
			func(ec *EvalCtx) []Value {
				var vs []Value
				for _, op := range argOps {
					vs = append(vs, op.Exec(ec)...)
				}
				return vs
			},
			-1, -1,
		}
		if len(argOps) > 0 {
			argsOp.Begin = argOps[0].Begin
			argsOp.End = argOps[len(argOps)-1].End
		}
		spaceyAssignOp = Op{
			makeAssignmentOpFunc(varsOp, restOp, argsOp),
			n.Begin(), argsOp.End,
		}
	}

	optsOp := cp.mapPairs(n.Opts)
	redirOps := cp.redirOps(n.Redirs)
	// TODO: n.ErrorRedir

	begin, end := n.Begin(), n.End()
	// ec here is always a subevaler created in compiler.pipeline, so it can
	// be safely modified.
	return func(ec *EvalCtx) {
		// Temporary assignment.
		if len(saveVarsOps) > 0 {
			// There is a temporary assignment.
			// Save variables.
			var saveVars []Variable
			var saveVals []Value
			for _, op := range saveVarsOps {
				saveVars = append(saveVars, op.Exec(ec)...)
			}
			for _, v := range saveVars {
				val := v.Get()
				saveVals = append(saveVals, val)
				logger.Printf("saved %s = %s", v, val)
			}
			// Do assignment.
			for _, op := range assignmentOps {
				op.Exec(ec)
			}
			// Defer variable restoration. Will be executed even if an error
			// occurs when evaling other part of the form.
			defer func() {
				for i, v := range saveVars {
					val := saveVals[i]
					if val == nil {
						// XXX Old value is nonexistent. We should delete the
						// variable. However, since the compiler now doesn't delete
						// it, we don't delete it in the evaler either.
						val = String("")
					}
					v.Set(val)
					logger.Printf("restored %s = %s", v, val)
				}
			}()
		}

		var headFn Callable
		var args []Value
		if headOp.Func != nil {
			// head
			headValues := headOp.Exec(ec)
			ec.must(headValues, "head of command", headOp.Begin, headOp.End).mustLen(1)
			headFn = mustFn(headValues[0])

			// args
			for _, argOp := range argOps {
				args = append(args, argOp.Exec(ec)...)
			}
		}

		// opts
		// XXX This conversion should be avoided.
		opts := optsOp(ec)[0].(Map)
		convertedOpts := make(map[string]Value)
		for k, v := range *opts.inner {
			if ks, ok := k.(String); ok {
				convertedOpts[string(ks)] = v
			} else {
				throwf("Option key must be string, got %s", k.Kind())
			}
		}

		// redirs
		for _, redirOp := range redirOps {
			redirOp.Exec(ec)
		}

		ec.begin, ec.end = begin, end

		if headFn != nil {
			headFn.Call(ec, args, convertedOpts)
		} else {
			spaceyAssignOp.Exec(ec)
		}
	}
}

func allTrue(vs []Value) bool {
	for _, v := range vs {
		if !ToBool(v) {
			return false
		}
	}
	return true
}

func (cp *compiler) assignment(n *parse.Assignment) OpFunc {
	variablesOp, restOp := cp.lvaluesOp(n.Left)
	valuesOp := cp.compoundOp(n.Right)
	return makeAssignmentOpFunc(variablesOp, restOp, valuesOp)
}

func makeAssignmentOpFunc(variablesOp, restOp LValuesOp, valuesOp ValuesOp) OpFunc {
	return func(ec *EvalCtx) {
		variables := variablesOp.Exec(ec)
		rest := restOp.Exec(ec)

		// If any LHS ends up being nil, assign an empty string to all of them.
		//
		// This is to fix #176, which only happens in the top level of REPL; in
		// other cases, a failure in the evaluation of the RHS causes this
		// level to fail, making the variables unaccessible.
		defer fixNilVariables(variables)
		defer fixNilVariables(rest)

		values := valuesOp.Exec(ec)

		if len(rest) > 1 {
			throw(ErrMoreThanOneRest)
		}
		if len(rest) == 1 {
			if len(variables) > len(values) {
				throw(ErrArityMismatch)
			}
		} else {
			if len(variables) != len(values) {
				throw(ErrArityMismatch)
			}
		}

		for i, variable := range variables {
			variable.Set(values[i])
		}

		if len(rest) == 1 {
			rest[0].Set(NewList(values[len(variables):]...))
		}
	}
}

func fixNilVariables(vs []Variable) {
	for _, v := range vs {
		if v.Get() == nil {
			v.Set(String(""))
		}
	}
}

func (cp *compiler) literal(n *parse.Primary, msg string) string {
	switch n.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return n.Value
	default:
		cp.compiling(n)
		cp.errorf(msg)
		return "" // not reached
	}
}

const defaultFileRedirPerm = 0644

// redir compiles a Redir into a op.
func (cp *compiler) redir(n *parse.Redir) OpFunc {
	var dstOp ValuesOp
	if n.Left != nil {
		dstOp = cp.compoundOp(n.Left)
	}
	srcOp := cp.compoundOp(n.Right)
	sourceIsFd := n.RightIsFd
	mode := n.Mode
	flag := makeFlag(mode)

	return func(ec *EvalCtx) {
		var dst int
		if dstOp.Func == nil {
			// use default dst fd
			switch mode {
			case parse.Read:
				dst = 0
			case parse.Write, parse.ReadWrite, parse.Append:
				dst = 1
			default:
				// XXX should report parser bug
				panic("bad RedirMode; parser bug")
			}
		} else {
			// dst must be a valid fd
			dst = ec.must(dstOp.Exec(ec), "FD", dstOp.Begin, dstOp.End).mustOneNonNegativeInt()
		}

		ec.growPorts(dst + 1)
		// Logger.Printf("closing old port %d of %s", dst, ec.context)
		ec.ports[dst].Close()

		srcMust := ec.must(srcOp.Exec(ec), "redirection source", srcOp.Begin, srcOp.End)
		if sourceIsFd {
			src := string(srcMust.mustOneStr())
			if src == "-" {
				// close
				ec.ports[dst] = &Port{}
			} else {
				fd := srcMust.zerothMustNonNegativeInt()
				ec.ports[dst] = ec.ports[fd].Fork()
			}
		} else {
			switch src := srcMust.mustOne().(type) {
			case String:
				f, err := os.OpenFile(string(src), flag, defaultFileRedirPerm)
				if err != nil {
					throwf("failed to open file %s: %s", src.Repr(NoPretty), err)
				}
				ec.ports[dst] = &Port{
					File: f, Chan: BlackholeChan,
					CloseFile: true,
				}
			case File:
				ec.ports[dst] = &Port{
					File: src.inner, Chan: BlackholeChan,
					CloseFile: false,
				}
			case Pipe:
				var f *os.File
				switch mode {
				case parse.Read:
					f = src.r
				case parse.Write:
					f = src.w
				default:
					cp.errorf("can only use < or > with pipes")
				}
				ec.ports[dst] = &Port{
					File: f, Chan: BlackholeChan,
					CloseFile: false,
				}
			default:
				srcMust.error("string or file", "%s", src.Kind())
			}
		}
	}
}
