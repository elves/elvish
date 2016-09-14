package eval

import (
	"fmt"
	"os"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/stub"
)

// Op is an operation on an EvalCtx.
type Op struct {
	Func       OpFunc
	Begin, End int
}

type OpFunc func(*EvalCtx)

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
	}
}

const pipelineChanBufferSize = 32

func (cp *compiler) pipeline(n *parse.Pipeline) OpFunc {
	ops := cp.formOps(n.Forms)

	return func(ec *EvalCtx) {
		bg := n.Background
		if bg {
			ec = ec.fork("background job " + n.SourceText())

			// Set up a new stub.
			st, err := stub.NewStub(os.Stderr)
			if err != nil {
				ec.errorf("failed to spawn stub: %v", err)
			}
			st.SetTitle(n.SourceText())
			ec.Stub = st

			if ec.Editor != nil {
				// TODO: Redirect output in interactive mode so that the line
				// editor does not get messed up.
			}
		}

		var nextIn *Port

		errorChans := make([]chan Error, len(ops))
		predReturnChans := make([]chan bool, len(ops))

		// For each form, create a dedicated evalCtx and run asynchronously
		for i, op := range ops {
			newEc := ec.fork(fmt.Sprintf("form op %v", op))
			if i > 0 {
				newEc.ports[0] = nextIn
			}
			if i < len(ops)-1 {
				// Each internal port pair consists of a (byte) pipe pair and a
				// channel.
				// os.Pipe sets O_CLOEXEC, which is what we want.
				reader, writer, e := os.Pipe()
				if e != nil {
					ec.errorf("failed to create pipe: %s", e)
				}
				ch := make(chan Value, pipelineChanBufferSize)
				newEc.ports[1] = &Port{
					File: writer, Chan: ch, CloseFile: true, CloseChan: true}
				nextIn = &Port{
					File: reader, Chan: ch, CloseFile: true, CloseChan: false}
			}
			thisOp := op
			errorChans[i] = make(chan Error)
			predReturnChans[i] = make(chan bool)
			thisErrorChan := errorChans[i]
			thisPredReturnChan := predReturnChans[i]
			go func() {
				err := newEc.PEval(thisOp)
				// Logger.Printf("closing ports of %s", newEc.context)
				ClosePorts(newEc.ports)
				thisErrorChan <- Error{err}
				thisPredReturnChan <- newEc.predReturn
			}()
		}

		collectErrorsAndPredReturns := func() ([]Error, bool) {
			// Wait for all forms to finish and collect error returns.
			errors := make([]Error, len(ops))
			for i, errorChan := range errorChans {
				errors[i] = <-errorChan
			}

			// Evaluate predReturn as the conjunction of all predReturn's.
			predReturn := true
			for _, predReturnChan := range predReturnChans {
				if !<-predReturnChan {
					predReturn = false
				}
			}

			return errors, predReturn
		}

		if bg {
			// Background job, wait for form termination asynchronously.
			go func() {
				errors, predReturn := collectErrorsAndPredReturns()
				ec.Stub.Terminate()
				msg := "job " + n.SourceText() + " finished"
				if !allok(errors) {
					msg += ", errors = " + makeCompositeError(errors).Error()
				}
				if !predReturn {
					msg += ", pred = false"
				}
				if ec.Editor != nil {
					m := ec.Editor.ActiveMutex()
					m.Lock()
					defer m.Unlock()

					if ec.Editor.Active() {
						ec.Editor.Notify("%s", msg)
					} else {
						ec.ports[2].File.WriteString(msg)
					}
				} else {
					ec.ports[2].File.WriteString(msg)
				}
			}()
		} else {
			errors, predReturn := collectErrorsAndPredReturns()
			maybeThrow(makeCompositeError(errors))
			ec.predReturn = predReturn
		}
	}
}

func makeCompositeError(errors []Error) error {
	if allok(errors) {
		return nil
	}
	if len(errors) == 1 {
		return errors[0].Inner
	} else {
		return MultiError{errors}
	}
}

func throwCompositeError(errors []Error) {
	if !allok(errors) {
		if len(errors) == 1 {
			throw(errors[0].Inner)
		} else {
			throw(MultiError{errors})
		}
	}
}

func (cp *compiler) form(n *parse.Form) OpFunc {
	var saveVarsOps []LValuesOp
	var assignmentOps []Op
	if len(n.Assignments) > 0 {
		assignmentOps = cp.assignmentOps(n.Assignments)
		if n.Head == nil {
			// Permanent assignment.
			return func(ec *EvalCtx) {
				for _, op := range assignmentOps {
					op.Exec(ec)
				}
			}
		} else {
			for _, a := range n.Assignments {
				v, r := cp.lvaluesOp(a.Dst)
				saveVarsOps = append(saveVarsOps, v, r)
			}
		}
	}

	if n.Control != nil {
		if len(n.Args) > 0 {
			cp.errorpf(n.Args[0].Begin(), n.Args[len(n.Args)-1].End(), "control structure takes no arguments")
		}
		redirOps := cp.redirOps(n.Redirs)
		controlOp := cp.controlOp(n.Control)
		return func(ec *EvalCtx) {
			for _, redirOp := range redirOps {
				redirOp.Exec(ec)
			}
			controlOp.Exec(ec)
		}
	}

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
	headOp := cp.compoundOp(n.Head)
	argOps := cp.compoundOps(n.Args)
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
				Logger.Printf("saved %s = %s", v, val)
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
					Logger.Printf("restored %s = %s", v, val)
				}
			}()
		}

		// head
		headValues := headOp.Exec(ec)
		ec.must(headValues, "head of command", headOp.Begin, headOp.End).mustLen(1)
		headFn := mustFn(headValues[0])

		// args
		var args []Value
		for _, argOp := range argOps {
			args = append(args, argOp.Exec(ec)...)
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
		headFn.Call(ec, args, convertedOpts)
	}
}

func (cp *compiler) control(n *parse.Control) OpFunc {
	switch n.Kind {
	case parse.IfControl:
		condOps := cp.chunkOps(n.Conditions)
		bodyOps := cp.chunkOps(n.Bodies)
		var elseOp Op
		if n.ElseBody != nil {
			elseOp = cp.chunkOp(n.ElseBody)
		}
		return func(ec *EvalCtx) {
			for i, condOp := range condOps {
				condOp.Exec(ec)
				if ec.predReturn {
					bodyOps[i].Exec(ec)
					return
				}
			}
			if elseOp.Func != nil {
				elseOp.Exec(ec)
			}
		}
	case parse.TryControl:
		Logger.Println("compiling a try control")
		bodyOp := cp.errorCaptureOp(n.Body)
		var exceptOp, elseOp ValuesOp
		var finallyOp Op
		var exceptVarOp LValuesOp
		if n.ExceptBody != nil {
			if n.ExceptVar != nil {
				var restOp LValuesOp
				exceptVarOp, restOp = cp.lvaluesOp(n.ExceptVar)
				if restOp.Func != nil {
					cp.errorpf(restOp.Begin, restOp.End, "may not use @rest in except variable")
				}
			}
			exceptOp = cp.errorCaptureOp(n.ExceptBody)
		}
		if n.ElseBody != nil {
			elseOp = cp.errorCaptureOp(n.ElseBody)
		}
		if n.FinallyBody != nil {
			finallyOp = cp.chunkOp(n.FinallyBody)
		}
		return func(ec *EvalCtx) {
			e := bodyOp.Exec(ec)[0].(Error).Inner
			Logger.Println("e is now", e)
			if e != nil {
				if exceptOp.Func != nil {
					if exceptVarOp.Func != nil {
						exceptVars := exceptVarOp.Exec(ec)
						if len(exceptVars) != 1 {
							throw(ErrArityMismatch)
						}
						exceptVars[0].Set(Error{e})
					}
					e = exceptOp.Exec(ec)[0].(Error).Inner
					Logger.Println("e is now", e)
				}
			} else {
				if elseOp.Func != nil {
					e = elseOp.Exec(ec)[0].(Error).Inner
					Logger.Println("e is now", e)
				}
			}
			if finallyOp.Func != nil {
				finallyOp.Exec(ec)
			}
			if e != nil {
				throw(e)
			}
		}
	case parse.WhileControl:
		condOp := cp.chunkOp(n.Condition)
		bodyOp := cp.chunkOp(n.Body)
		return func(ec *EvalCtx) {
			for {
				condOp.Exec(ec)
				if !ec.predReturn {
					break
				}
				//for condOp.Exec(ec)[0].(Error).Inner == nil {
				ex := ec.PEval(bodyOp)
				if ex == Continue {
					// do nothing
				} else if ex == Break {
					break
				} else if ex != nil {
					throw(ex)
				}
			}
		}
	case parse.ForControl:
		iteratorOp, restOp := cp.lvaluesOp(n.Iterator)
		if restOp.Func != nil {
			cp.errorpf(restOp.Begin, restOp.End, "may not use @rest in iterator")
		}
		valuesOp := cp.arrayOp(n.Array)
		bodyOp := cp.chunkOp(n.Body)
		return func(ec *EvalCtx) {
			iterators := iteratorOp.Exec(ec)
			if len(iterators) != 1 {
				throw(ErrArityMismatch)
			}
			iterator := iterators[0]

			values := valuesOp.Exec(ec)
			for _, v := range values {
				iterator.Set(v)
				ex := ec.PEval(bodyOp)
				if ex == Continue {
					// do nothing
				} else if ex == Break {
					break
				} else if ex != nil {
					throw(ex)
				}
			}
		}
	case parse.BeginControl:
		return cp.chunk(n.Body)
	default:
		cp.errorpf(n.Begin(), n.End(), "unknown ControlKind %s, compiler bug", n.Kind)
		panic("unreachable")
	}
}

func (cp *compiler) assignment(n *parse.Assignment) OpFunc {
	variablesOp, restOp := cp.lvaluesOp(n.Dst)
	valuesOp := cp.compoundOp(n.Src)

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
	if n.Dest != nil {
		dstOp = cp.compoundOp(n.Dest)
	}
	srcOp := cp.compoundOp(n.Source)
	sourceIsFd := n.SourceIsFd
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
					ec.errorf("failed to open file %s: %s", src.Repr(NoPretty), err)
				}
				ec.ports[dst] = &Port{
					File: f, Chan: make(chan Value),
					CloseFile: true, CloseChan: true,
				}
			case File:
				ec.ports[dst] = &Port{
					File: src.inner, Chan: make(chan Value),
					CloseFile: false, CloseChan: true,
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
					File: f, Chan: make(chan Value),
					CloseFile: false, CloseChan: true,
				}
			default:
				srcMust.error("string or file", "%s", src.Kind())
			}
		}
	}
}
