package eval

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elves/elvish/parse"
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

		if !allok(errors) {
			if len(errors) == 1 {
				throw(errors[0].Inner)
			} else {
				throw(MultiError{errors})
			}
		}
	}
}

func (cp *compiler) form(n *parse.Form) OpFunc {
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
			// Temporary assignment.
			cp.errorpf(n.Assignments[0].Begin(), n.Assignments[len(n.Assignments)-1].End(), "temporary assignments not yet supported")
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
	// TODO: n.NamedArgs
	redirOps := cp.redirOps(n.Redirs)
	// TODO: n.ErrorRedir

	begin, end := n.Begin(), n.End()
	// ec here is always a subevaler created in compiler.pipeline, so it can
	// be safely modified.
	return func(ec *EvalCtx) {
		// head
		headValues := headOp.Exec(ec)
		ec.must(headValues, "head of command", headOp.Begin, headOp.End).mustLen(1)
		headCaller := mustCaller(headValues[0])

		// args
		var args []Value
		for _, argOp := range argOps {
			args = append(args, argOp.Exec(ec)...)
		}

		// redirs
		for _, redirOp := range redirOps {
			redirOp.Exec(ec)
		}

		ec.begin, ec.end = begin, end
		headCaller.Call(ec, args)
	}
}

func (cp *compiler) control(n *parse.Control) OpFunc {
	switch n.Kind {
	case parse.IfControl:
		condOps := cp.errorCaptureOps(n.Conditions)
		bodyOps := cp.chunkOps(n.Bodies)
		var elseOp Op
		if n.ElseBody != nil {
			elseOp = cp.chunkOp(n.ElseBody)
		}
		return func(ec *EvalCtx) {
			for i, condOp := range condOps {
				if condOp.Exec(ec)[0].(Error).Inner == nil {
					bodyOps[i].Exec(ec)
					return
				}
			}
			if elseOp.Func != nil {
				elseOp.Exec(ec)
			}
		}
	case parse.WhileControl:
		condOp := cp.errorCaptureOp(n.Condition)
		bodyOp := cp.chunkOp(n.Body)
		return func(ec *EvalCtx) {
			for condOp.Exec(ec)[0].(Error).Inner == nil {
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
		iteratorOp := cp.singleVariableOp(n.Iterator, "must be a single variable")
		valuesOp := cp.arrayOp(n.Array)
		bodyOp := cp.chunkOp(n.Body)
		return func(ec *EvalCtx) {
			iterator := iteratorOp.Exec(ec)
			values := valuesOp.Exec(ec)
			for _, v := range values {
				doSet(ec, iterator, []Value{v})
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
	variablesOp := cp.multiVariableOp(n.Dst)
	valuesOp := cp.compoundOp(n.Src)

	return func(ec *EvalCtx) {
		doSet(ec, variablesOp.Exec(ec), valuesOp.Exec(ec))
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
				ec.errorf("failed to open file %q: %s", src, err)
			}
			ec.ports[dst] = &Port{
				File: f, Chan: make(chan Value), CloseFile: true, CloseChan: true,
			}
		}
	}
}
