package eval

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/sys"
)

var (
	// InterruptDeadline is the amount of time elvish waits for foreground
	// tasks to finish after receiving a SIGINT. If a task didn't actually exit
	// in time, its exit status takes the special "still running" value.
	InterruptDeadline = 50 * time.Millisecond
	// PutInForeground determines whether elvish should attempt to put itself in
	// foreground after each pipeline execution.
	PutInForeground         = true
	outputCaptureBufferSize = 16
)

var ErrStillRunning = errors.New("still running")

func (cp *compiler) chunk(n *parse.Chunk) Op {
	ops := cp.pipelines(n.Pipelines)

	return func(ec *EvalCtx) {
		for _, op := range ops {
			op(ec)
		}
	}
}

const pipelineChanBufferSize = 32

func (cp *compiler) pipeline(n *parse.Pipeline) Op {
	ops := cp.forms(n.Forms)
	p := n.Begin()

	return func(ec *EvalCtx) {
		var nextIn *Port

		errorChans := make([]chan Error, len(ops))

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
					ec.errorf(p, "failed to create pipe: %s", e)
				}
				ch := make(chan Value, pipelineChanBufferSize)
				newEc.ports[1] = &Port{
					File: writer, Chan: ch, CloseFile: true, CloseChan: true}
				nextIn = &Port{
					File: reader, Chan: ch, CloseFile: true, CloseChan: false}
			}
			thisOp := op
			errorChans[i] = make(chan Error)
			thisErrorChan := errorChans[i]
			go func() {
				err := newEc.PEval(thisOp)
				// Logger.Printf("closing ports of %s", newEc.context)
				ClosePorts(newEc.ports)
				thisErrorChan <- Error{err}
			}()
		}

		intCh := make(chan os.Signal)
		signal.Notify(intCh, syscall.SIGINT)
		interrupted := make(chan struct{})
		cancel := make(chan struct{}, 1)
		go func() {
			// When SIGINT is received, sleep for InterruptDeadline before the
			// closing interrupted channel.
			select {
			case <-intCh:
			case <-cancel:
				return
			}
			select {
			case <-time.After(InterruptDeadline):
			case <-cancel:
				return
			}
			close(interrupted)
		}()

		// Wait for all forms to finish and collect error returns, unless an
		// interrupt was received and the form didn't quit within
		// InterruptDeadline.
		errors := make([]Error, len(ops))
		for i, errorChan := range errorChans {
			select {
			case errors[i] = <-errorChan:
			case <-interrupted:
				errors[i] = Error{ErrStillRunning}
			}
		}

		// Make sure the SIGINT listener exits.
		close(cancel)
		signal.Stop(intCh)

		// Make sure I am in foreground.
		if PutInForeground && sys.IsATTY(0) {
			err := sys.Tcsetpgrp(0, syscall.Getpgrp())
			if err != nil {
				throw(err)
			}
		}

		if !allok(errors) {
			if len(errors) == 1 {
				throw(errors[0].inner)
			} else {
				throw(multiError{errors})
			}
		}
	}
}

func (cp *compiler) form(n *parse.Form) Op {
	if len(n.Assignments) > 0 {
		if n.Head != nil {
			cp.errorf(n.Begin(), "temporary assignments not yet supported")
		}
		ops := cp.assignments(n.Assignments)
		return func(ec *EvalCtx) {
			for _, op := range ops {
				op(ec)
			}
		}
	}

	if n.Control != nil {
		if len(n.Args) > 0 {
			cp.errorf(n.Args[0].Begin(), "control structure takes no arguments")
		}
		redirOps := cp.redirs(n.Redirs)
		controlOp := cp.control(n.Control)
		return func(ec *EvalCtx) {
			for _, redirOp := range redirOps {
				redirOp(ec)
			}
			controlOp(ec)
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
	headOp := cp.compound(n.Head)
	argOps := cp.compounds(n.Args)
	// TODO: n.NamedArgs
	redirOps := cp.redirs(n.Redirs)
	// TODO: n.ErrorRedir

	p := n.Begin()
	// ec here is always a subevaler created in compiler.pipeline, so it can
	// be safely modified.
	return func(ec *EvalCtx) {
		// head
		headValues := headOp(ec)
		headMust := ec.must(headValues, "the head of command", p)
		headMust.mustLen(1)
		headCaller := mustCaller(headValues[0])

		// args
		var args []Value
		for _, argOp := range argOps {
			args = append(args, argOp(ec)...)
		}

		// redirs
		for _, redirOp := range redirOps {
			redirOp(ec)
		}

		headCaller.Call(ec, args)
	}
}

func (cp *compiler) control(n *parse.Control) Op {
	switch n.Kind {
	case parse.IfControl:
		condOps := cp.errorCaptures(n.Conditions)
		bodyOps := cp.chunks(n.Bodies)
		var elseOp Op
		if n.ElseBody != nil {
			elseOp = cp.chunk(n.ElseBody)
		}
		return func(ec *EvalCtx) {
			for i, condOp := range condOps {
				if condOp(ec)[0].(Error).inner == nil {
					bodyOps[i](ec)
					return
				}
			}
			if elseOp != nil {
				elseOp(ec)
			}
		}
	case parse.WhileControl:
		condOp := cp.errorCapture(n.Condition)
		bodyOp := cp.chunk(n.Body)
		return func(ec *EvalCtx) {
			for condOp(ec)[0].(Error).inner == nil {
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
		iteratorOp := cp.singleVariable(n.Iterator, "must be a single variable")
		valuesOp := cp.array(n.Array)
		bodyOp := cp.chunk(n.Body)
		return func(ec *EvalCtx) {
			iterator := iteratorOp(ec)
			values := valuesOp(ec)
			for _, v := range values {
				doSet(ec, []Variable{iterator}, []Value{v})
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
		cp.errorf(n.Begin(), "unknown ControlKind %s, compiler bug", n.Kind)
		panic("unreachable")
	}
}

func (cp *compiler) assignment(n *parse.Assignment) Op {
	variableOps := cp.multiVariable(n.Dst)
	valuesOp := cp.compound(n.Src)

	return func(ec *EvalCtx) {
		variables := make([]Variable, len(variableOps))
		for i, variableOp := range variableOps {
			variables[i] = variableOp(ec)
		}
		doSet(ec, variables, valuesOp(ec))
	}
}

func (cp *compiler) multiVariable(n *parse.Indexing) []VariableOp {
	var variableOps []VariableOp
	if n.Head.Type == parse.Braced {
		// XXX ignore n.Indicies.
		compounds := n.Head.Braced
		indexings := make([]*parse.Indexing, len(compounds))
		for i, cn := range compounds {
			if len(cn.Indexings) != 1 {
				cp.errorf(cn.Begin(), "must be a variable spec")
			}
			indexings[i] = cn.Indexings[0]
		}
		variableOps = cp.singleVariables(indexings, "must be a variable spc")
	} else {
		variableOps = []VariableOp{cp.singleVariable(n, "must be a variable spec or a braced list of those")}
	}
	return variableOps
}

func (cp *compiler) singleVariable(n *parse.Indexing, msg string) VariableOp {
	// XXX will we be using this for purposes other than setting?
	varname := cp.literal(n.Head, msg)
	p := n.Begin()

	if len(n.Indicies) == 0 {
		cp.registerVariableSet(varname)

		return func(ec *EvalCtx) Variable {
			splice, ns, barename := parseVariable(varname)
			if splice {
				// XXX
				ec.errorf(p, "not yet supported")
			}
			variable := ec.ResolveVar(ns, barename)
			if variable == nil {
				if ns == "" || ns == "local" {
					// New variable.
					// XXX We depend on the fact that this variable will
					// immeidately be set.
					variable = NewPtrVariable(nil)
					ec.local[barename] = variable
				} else if mod, ok := ec.modules[ns]; ok {
					variable = NewPtrVariable(nil)
					mod[barename] = variable
				} else {
					ec.errorf(p, "cannot set $%s", varname)
				}
			}
			return variable
		}
	}
	cp.registerVariableGet(varname)
	indexOps := cp.arrays(n.Indicies)
	indexBegins := make([]int, len(n.Indicies))
	indexEnds := make([]int, len(n.Indicies))
	for i, in := range n.Indicies {
		indexBegins[i] = in.Begin()
		indexEnds[i] = in.End()
	}

	return func(ec *EvalCtx) Variable {
		splice, ns, name := parseVariable(varname)
		if splice {
			// XXX
			ec.errorf(p, "not yet supported")
		}
		variable := ec.ResolveVar(ns, name)
		if variable == nil {
			ec.errorf(p, "variable $%s does not exisit, compiler bug", varname)
		}
		if len(indexOps) == 0 {
			// Just a variable, return directly.
			return variable
		}

		// Indexing. Do Index up to the last but one index.
		value := variable.Get()
		n := len(indexOps)
		for _, op := range indexOps[:n-1] {
			indexer := mustIndexer(value, ec)

			indicies := op(ec)
			values := indexer.Index(indicies)
			if len(values) != 1 {
				throw(errors.New("multi indexing not implemented"))
			}
			value = values[0]
		}
		// Now this must be an IndexSetter.
		indexSetter, ok := value.(IndexSetter)
		if !ok {
			ec.errorf( /* from p to */ indexBegins[n-1], "cannot be indexed for setting (value is %s, type %s)", value.Repr(NoPretty), value.Kind())
		}
		// XXX Duplicate code.
		indicies := indexOps[n-1](ec)
		if len(indicies) != 1 {
			ec.errorf(indexBegins[n-1], "index must eval to a single Value (got %v)", indicies)
		}
		return elemVariable{indexSetter, indicies[0]}
	}
}

func (cp *compiler) literal(n *parse.Primary, msg string) string {
	switch n.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return n.Value
	default:
		cp.errorf(n.Begin(), msg)
		return "" // not reached
	}
}

const defaultFileRedirPerm = 0644

// redir compiles a Redir into a op.
func (cp *compiler) redir(n *parse.Redir) Op {
	var dstOp ValuesOp
	if n.Dest != nil {
		dstOp = cp.compound(n.Dest)
	}
	p := n.Begin()
	srcOp := cp.compound(n.Source)
	sourceIsFd := n.SourceIsFd
	pSrc := n.Source.Begin()
	mode := n.Mode
	flag := makeFlag(mode)

	return func(ec *EvalCtx) {
		var dst int
		if dstOp == nil {
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
			dst = ec.must(dstOp(ec), "FD", p).mustOneNonNegativeInt()
		}

		ec.growPorts(dst + 1)
		// Logger.Printf("closing old port %d of %s", dst, ec.context)
		ec.ports[dst].Close()

		srcMust := ec.must(srcOp(ec), "redirection source", pSrc)
		src := string(srcMust.mustOneStr())
		if sourceIsFd {
			if src == "-" {
				// close
				ec.ports[dst] = &Port{}
			} else {
				fd := srcMust.zerothMustNonNegativeInt()
				ec.ports[dst] = ec.ports[fd].Fork()
			}
		} else {
			f, err := os.OpenFile(src, flag, defaultFileRedirPerm)
			if err != nil {
				ec.errorf(p, "failed to open file %q: %s", src, err)
			}
			ec.ports[dst] = &Port{
				File: f, Chan: make(chan Value), CloseFile: true, CloseChan: true,
			}
		}
	}
}
