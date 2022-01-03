package eval

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
)

// An operation with some side effects.
type effectOp interface{ exec(*Frame) Exception }

func (cp *compiler) chunkOp(n *parse.Chunk) effectOp {
	return chunkOp{n.Range(), cp.pipelineOps(n.Pipelines)}
}

type chunkOp struct {
	diag.Ranging
	subops []effectOp
}

func (op chunkOp) exec(fm *Frame) Exception {
	for _, subop := range op.subops {
		exc := subop.exec(fm)
		if exc != nil {
			return exc
		}
	}
	// Check for interrupts after the chunk.
	// We also check for interrupts before each pipeline, so there is no
	// need to check it before the chunk or after each pipeline.
	if fm.IsInterrupted() {
		return fm.errorp(op, ErrInterrupted)
	}
	return nil
}

func (cp *compiler) pipelineOp(n *parse.Pipeline) effectOp {
	formOps := cp.formOps(n.Forms)

	return &pipelineOp{n.Range(), n.Background, parse.SourceText(n), formOps}
}

func (cp *compiler) pipelineOps(ns []*parse.Pipeline) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.pipelineOp(n)
	}
	return ops
}

type pipelineOp struct {
	diag.Ranging
	bg     bool
	source string
	subops []effectOp
}

const pipelineChanBufferSize = 32

func (op *pipelineOp) exec(fm *Frame) Exception {
	if fm.IsInterrupted() {
		return fm.errorp(op, ErrInterrupted)
	}

	if op.bg {
		fm = fm.Fork("background job" + op.source)
		fm.intCh = nil
		fm.background = true
		fm.Evaler.addNumBgJobs(1)
	}

	nforms := len(op.subops)

	var wg sync.WaitGroup
	wg.Add(nforms)
	excs := make([]Exception, nforms)

	var nextIn *Port

	// For each form, create a dedicated evalCtx and run asynchronously
	for i, formOp := range op.subops {
		newFm := fm.Fork("[form op]")
		inputIsPipe := i > 0
		outputIsPipe := i < nforms-1
		if inputIsPipe {
			newFm.ports[0] = nextIn
		}
		if outputIsPipe {
			// Each internal port pair consists of a (byte) pipe pair and a
			// channel.
			// os.Pipe sets O_CLOEXEC, which is what we want.
			reader, writer, e := os.Pipe()
			if e != nil {
				return fm.errorpf(op, "failed to create pipe: %s", e)
			}
			ch := make(chan interface{}, pipelineChanBufferSize)
			sendStop := make(chan struct{})
			sendError := new(error)
			readerGone := new(int32)
			newFm.ports[1] = &Port{
				File: writer, Chan: ch,
				closeFile: true, closeChan: true,
				sendStop: sendStop, sendError: sendError, readerGone: readerGone}
			nextIn = &Port{
				File: reader, Chan: ch,
				closeFile: true, closeChan: false,
				// Store in input port for ease of retrieval later
				sendStop: sendStop, sendError: sendError, readerGone: readerGone}
		}
		f := func(formOp effectOp, pexc *Exception) {
			exc := formOp.exec(newFm)
			if exc != nil && !(outputIsPipe && isReaderGone(exc)) {
				*pexc = exc
			}
			if inputIsPipe {
				input := newFm.ports[0]
				*input.sendError = errs.ReaderGone{}
				close(input.sendStop)
				atomic.StoreInt32(input.readerGone, 1)
			}
			newFm.Close()
			wg.Done()
		}
		if i == nforms-1 && !op.bg {
			f(formOp, &excs[i])
		} else {
			go f(formOp, &excs[i])
		}
	}

	if op.bg {
		// Background job, wait for form termination asynchronously.
		go func() {
			wg.Wait()
			fm.Evaler.addNumBgJobs(-1)
			if notify := fm.Evaler.BgJobNotify; notify != nil {
				msg := "job " + op.source + " finished"
				err := MakePipelineError(excs)
				if err != nil {
					msg += ", errors = " + err.Error()
				}
				if fm.Evaler.getNotifyBgJobSuccess() || err != nil {
					notify(msg)
				}
			}
		}()
		return nil
	}
	wg.Wait()
	return fm.errorp(op, MakePipelineError(excs))
}

func isReaderGone(exc Exception) bool {
	_, ok := exc.Reason().(errs.ReaderGone)
	return ok
}

func (cp *compiler) formOp(n *parse.Form) effectOp {
	var tempLValues []lvalue
	var assignmentOps []effectOp
	if len(n.Assignments) > 0 {
		if n.Head == nil {
			cp.errorpf(n, `using the syntax of temporary assignment for non-temporary assignment is no longer supported; use "var" or "set" instead`)
			return nopOp{}
		} else {
			as := n.Assignments
			cp.deprecate(diag.MixedRanging(as[0], as[len(as)-1]),
				`the legacy temporary assignment syntax is deprecated; use "tmp" instead`, 18)
		}
		assignmentOps = cp.assignmentOps(n.Assignments)
		for _, a := range n.Assignments {
			lvalues := cp.parseIndexingLValue(a.Left, setLValue|newLValue)
			tempLValues = append(tempLValues, lvalues.lvalues...)
		}
		logger.Println("temporary assignment of", len(n.Assignments), "pairs")
	}

	redirOps := cp.redirOps(n.Redirs)
	body := cp.formBody(n)

	return &formOp{n.Range(), tempLValues, assignmentOps, redirOps, body}
}

func (cp *compiler) formBody(n *parse.Form) formBody {
	if n.Head == nil {
		// Compiling an incomplete form node, return an empty body.
		return formBody{}
	}

	// Determine if this form is a special command.
	if head, ok := cmpd.StringLiteral(n.Head); ok {
		special, _ := resolveCmdHeadInternally(cp, head, n.Head)
		if special != nil {
			specialOp := special(cp, n)
			return formBody{specialOp: specialOp}
		}
	}

	var headOp valuesOp
	if head, ok := cmpd.StringLiteral(n.Head); ok {
		// Head is a literal string: resolve to function or external (special
		// commands are already handled above).
		if _, fnRef := resolveCmdHeadInternally(cp, head, n.Head); fnRef != nil {
			headOp = variableOp{n.Head.Range(), false, head + FnSuffix, fnRef}
		} else if cp.currentPragma().unknownCommandIsExternal || fsutil.DontSearch(head) {
			headOp = literalValues(n.Head, NewExternalCmd(head))
		} else {
			cp.errorpf(n.Head, "unknown command disallowed by current pragma")
		}
	} else {
		// Head is not a literal string: evaluate as a normal expression.
		headOp = cp.compoundOp(n.Head)
	}

	argOps := cp.compoundOps(n.Args)
	optsOp := cp.mapPairs(n.Opts)
	return formBody{ordinaryCmd: ordinaryCmd{headOp, argOps, optsOp}}
}

func (cp *compiler) formOps(ns []*parse.Form) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.formOp(n)
	}
	return ops
}

type formOp struct {
	diag.Ranging
	tempLValues   []lvalue
	tempAssignOps []effectOp
	redirOps      []effectOp
	body          formBody
}

type formBody struct {
	// Exactly one field will be populated.
	specialOp   effectOp
	assignOp    effectOp
	ordinaryCmd ordinaryCmd
}

type ordinaryCmd struct {
	headOp valuesOp
	argOps []valuesOp
	optsOp *mapPairsOp
}

func (op *formOp) exec(fm *Frame) (errRet Exception) {
	// fm here is always a sub-frame created in compiler.pipeline, so it can
	// be safely modified.

	// Temporary assignment.
	if len(op.tempLValues) > 0 {
		// There is a temporary assignment.
		// Save variables.
		var saveVars []vars.Var
		var saveVals []interface{}
		for _, lv := range op.tempLValues {
			variable, err := derefLValue(fm, lv)
			if err != nil {
				return fm.errorp(op, err)
			}
			saveVars = append(saveVars, variable)
		}
		for i, v := range saveVars {
			// TODO(xiaq): If the variable to save is a elemVariable, save
			// the outermost variable instead.
			if u := vars.HeadOfElement(v); u != nil {
				v = u
				saveVars[i] = v
			}
			val := v.Get()
			saveVals = append(saveVals, val)
			logger.Printf("saved %s = %s", v, val)
		}
		// Do assignment.
		for _, subop := range op.tempAssignOps {
			exc := subop.exec(fm)
			if exc != nil {
				return exc
			}
		}
		// Defer variable restoration. Will be executed even if an error
		// occurs when evaling other part of the form.
		defer func() {
			for i, v := range saveVars {
				val := saveVals[i]
				if val == nil {
					// TODO(xiaq): Old value is nonexistent. We should delete
					// the variable. However, since the compiler now doesn't
					// delete it, we don't delete it in the evaler either.
					val = ""
				}
				err := v.Set(val)
				if err != nil {
					errRet = fm.errorp(op, err)
				}
				logger.Printf("restored %s = %s", v, val)
			}
		}()
	}

	// Redirections.
	for _, redirOp := range op.redirOps {
		exc := redirOp.exec(fm)
		if exc != nil {
			return exc
		}
	}

	if op.body.specialOp != nil {
		return op.body.specialOp.exec(fm)
	}
	if op.body.assignOp != nil {
		return op.body.assignOp.exec(fm)
	}

	// Ordinary command: evaluate head, arguments and options.
	cmd := op.body.ordinaryCmd

	// Special case: evaluating an incomplete form node. Return directly.
	if cmd.headOp == nil {
		return nil
	}

	headFn, err := evalForCommand(fm, cmd.headOp, "command")
	if err != nil {
		return fm.errorp(cmd.headOp, err)
	}

	var args []interface{}
	for _, argOp := range cmd.argOps {
		moreArgs, exc := argOp.exec(fm)
		if exc != nil {
			return exc
		}
		args = append(args, moreArgs...)
	}

	// TODO(xiaq): This conversion should be avoided.
	convertedOpts := make(map[string]interface{})
	exc := cmd.optsOp.exec(fm, func(k, v interface{}) Exception {
		if ks, ok := k.(string); ok {
			convertedOpts[ks] = v
			return nil
		}
		// TODO(xiaq): Point to the particular key.
		return fm.errorp(op, errs.BadValue{
			What: "option key", Valid: "string", Actual: vals.Kind(k)})
	})
	if exc != nil {
		return exc
	}

	fm.traceback = fm.addTraceback(op)
	err = headFn.Call(fm, args, convertedOpts)
	if exc, ok := err.(Exception); ok {
		return exc
	}
	return &exception{err, fm.traceback}
}

func evalForCommand(fm *Frame, op valuesOp, what string) (Callable, error) {
	value, err := evalForValue(fm, op, what)
	if err != nil {
		return nil, err
	}
	switch value := value.(type) {
	case Callable:
		return value, nil
	case string:
		if fsutil.DontSearch(value) {
			return NewExternalCmd(value), nil
		}
	}
	return nil, fm.errorp(op, errs.BadValue{
		What:   what,
		Valid:  "callable or string containing slash",
		Actual: vals.ReprPlain(value)})
}

func allTrue(vs []interface{}) bool {
	for _, v := range vs {
		if !vals.Bool(v) {
			return false
		}
	}
	return true
}

func (cp *compiler) assignmentOp(n *parse.Assignment) effectOp {
	lhs := cp.parseIndexingLValue(n.Left, setLValue|newLValue)
	rhs := cp.compoundOp(n.Right)
	return &assignOp{n.Range(), lhs, rhs, false}
}

func (cp *compiler) assignmentOps(ns []*parse.Assignment) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.assignmentOp(n)
	}
	return ops
}

const defaultFileRedirPerm = 0644

// redir compiles a Redir into a op.
func (cp *compiler) redirOp(n *parse.Redir) effectOp {
	var dstOp valuesOp
	if n.Left != nil {
		dstOp = cp.compoundOp(n.Left)
	}
	flag := makeFlag(n.Mode)
	if flag == -1 {
		// TODO: Record and get redirection sign position
		cp.errorpf(n, "bad redirection sign")
	}
	return &redirOp{n.Range(), dstOp, cp.compoundOp(n.Right), n.RightIsFd, n.Mode, flag}
}

func (cp *compiler) redirOps(ns []*parse.Redir) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.redirOp(n)
	}
	return ops
}

func makeFlag(m parse.RedirMode) int {
	switch m {
	case parse.Read:
		return os.O_RDONLY
	case parse.Write:
		return os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case parse.ReadWrite:
		return os.O_RDWR | os.O_CREATE
	case parse.Append:
		return os.O_WRONLY | os.O_CREATE | os.O_APPEND
	default:
		return -1
	}
}

type redirOp struct {
	diag.Ranging
	dstOp   valuesOp
	srcOp   valuesOp
	srcIsFd bool
	mode    parse.RedirMode
	flag    int
}

type InvalidFD struct{ FD int }

func (err InvalidFD) Error() string { return fmt.Sprintf("invalid fd: %d", err.FD) }

func (op *redirOp) exec(fm *Frame) Exception {
	var dst int
	if op.dstOp == nil {
		// No explicit FD destination specified; use default destinations
		switch op.mode {
		case parse.Read:
			dst = 0
		case parse.Write, parse.ReadWrite, parse.Append:
			dst = 1
		default:
			return fm.errorpf(op, "bad RedirMode; parser bug")
		}
	} else {
		// An explicit FD destination specified, evaluate it.
		var err error
		dst, err = evalForFd(fm, op.dstOp, false, "redirection destination")
		if err != nil {
			return fm.errorp(op, err)
		}
	}

	growPorts(&fm.ports, dst+1)
	fm.ports[dst].close()

	if op.srcIsFd {
		src, err := evalForFd(fm, op.srcOp, true, "redirection source")
		if err != nil {
			return fm.errorp(op, err)
		}
		switch {
		case src == -1:
			// close
			fm.ports[dst] = &Port{
				// Ensure that writing to value output throws an exception
				sendStop: closedSendStop, sendError: &ErrNoValueOutput}
		case src >= len(fm.ports) || fm.ports[src] == nil:
			return fm.errorp(op, InvalidFD{FD: src})
		default:
			fm.ports[dst] = fm.ports[src].fork()
		}
		return nil
	}
	src, err := evalForValue(fm, op.srcOp, "redirection source")
	if err != nil {
		return fm.errorp(op, err)
	}
	switch src := src.(type) {
	case string:
		f, err := os.OpenFile(src, op.flag, defaultFileRedirPerm)
		if err != nil {
			return fm.errorpf(op, "failed to open file %s: %s", vals.ReprPlain(src), err)
		}
		fm.ports[dst] = fileRedirPort(op.mode, f, true)
	case vals.File:
		fm.ports[dst] = fileRedirPort(op.mode, src, false)
	case vals.Pipe:
		var f *os.File
		switch op.mode {
		case parse.Read:
			f = src.ReadEnd
		case parse.Write:
			f = src.WriteEnd
		default:
			return fm.errorpf(op, "can only use < or > with pipes")
		}
		fm.ports[dst] = fileRedirPort(op.mode, f, false)
	default:
		return fm.errorp(op.srcOp, errs.BadValue{
			What:  "redirection source",
			Valid: "string, file or pipe", Actual: vals.Kind(src)})
	}
	return nil
}

// Creates a port that only have a file component, populating the
// channel-related fields with suitable values depending on the redirection
// mode.
func fileRedirPort(mode parse.RedirMode, f *os.File, closeFile bool) *Port {
	if mode == parse.Read {
		return &Port{
			File: f, closeFile: closeFile,
			// ClosedChan produces no values when reading.
			Chan: ClosedChan,
		}
	}
	return &Port{
		File: f, closeFile: closeFile,
		// Throws errValueOutputIsClosed when writing.
		Chan: nil, sendStop: closedSendStop, sendError: &ErrNoValueOutput,
	}
}

// Makes the size of *ports at least n, adding nil's if necessary.
func growPorts(ports *[]*Port, n int) {
	if len(*ports) >= n {
		return
	}
	oldPorts := *ports
	*ports = make([]*Port, n)
	copy(*ports, oldPorts)
}

func evalForFd(fm *Frame, op valuesOp, closeOK bool, what string) (int, error) {
	value, err := evalForValue(fm, op, what)
	if err != nil {
		return -1, err
	}
	switch value {
	case "stdin":
		return 0, nil
	case "stdout":
		return 1, nil
	case "stderr":
		return 2, nil
	}
	var fd int
	if vals.ScanToGo(value, &fd) == nil {
		return fd, nil
	} else if value == "-" && closeOK {
		return -1, nil
	}
	valid := "fd name or number"
	if closeOK {
		valid = "fd name or number or '-'"
	}
	return -1, fm.errorp(op, errs.BadValue{
		What: what, Valid: valid, Actual: vals.ReprPlain(value)})
}

type seqOp struct{ subops []effectOp }

func (op seqOp) exec(fm *Frame) Exception {
	for _, subop := range op.subops {
		exc := subop.exec(fm)
		if exc != nil {
			return exc
		}
	}
	return nil
}

type nopOp struct{}

func (nopOp) exec(fm *Frame) Exception { return nil }
