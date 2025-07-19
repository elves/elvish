package eval

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
)

// An operation with some side effects.
type effectOp interface{ exec(*Frame) Exception }

type chunkOp struct {
	diag.Ranging
	pipelines []*pipelineOp
}

func (cp *compiler) chunkOp(n *parse.Chunk) *chunkOp {
	return &chunkOp{n.Range(), cp.pipelineOps(n.Pipelines)}
}

func (op *chunkOp) exec(fm *Frame) Exception {
	for _, subop := range op.pipelines {
		exc := subop.exec(fm)
		if exc != nil {
			return exc
		}
	}
	// Check for interrupts after the chunk.
	// We also check for interrupts before each pipeline, so there is no
	// need to check it before the chunk or after each pipeline.
	if fm.Canceled() {
		return fm.errorp(op, ErrInterrupted)
	}
	return nil
}

type pipelineOp struct {
	diag.Ranging
	bg     bool
	source string
	forms  []*formOp
}

func (cp *compiler) pipelineOps(ns []*parse.Pipeline) []*pipelineOp {
	ops := make([]*pipelineOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.pipelineOp(n)
	}
	return ops
}

func (cp *compiler) pipelineOp(n *parse.Pipeline) *pipelineOp {
	formOps := cp.formOps(n.Forms)

	return &pipelineOp{n.Range(), n.Background, parse.SourceText(n), formOps}
}

const pipelineChanBufferSize = 32

// Keeps track of whether the File and Port parts of a port are owned by a form
// and should be closed when the form finishes execution.
type formOwnedPort struct {
	File bool
	Chan bool
}

func (fop formOwnedPort) close(p *Port) {
	if fop.File {
		p.File.Close()
	}
	if fop.Chan {
		close(p.Chan)
	}
}

func (op *pipelineOp) exec(fm *Frame) Exception {
	if fm.Canceled() {
		return fm.errorp(op, ErrInterrupted)
	}

	if op.bg {
		fm = fm.Fork()
		fm.ctx = context.Background()
		fm.background = true
		fm.Evaler.addNumBgJobs(1)
	}

	nforms := len(op.forms)

	var wg sync.WaitGroup
	wg.Add(nforms)
	excs := make([]Exception, nforms)

	var nextIn *Port

	// For each form, create a dedicated evalCtx and run asynchronously
	for i, form := range op.forms {
		newFm := fm.Fork()
		var fops []formOwnedPort
		inputIsPipe := i > 0
		outputIsPipe := i < nforms-1
		if inputIsPipe {
			newFm.ports[0] = nextIn
			growAccess(&fops, 0).File = true
		}
		if outputIsPipe {
			// Each internal port pair consists of a (byte) pipe pair and a
			// channel.
			// os.Pipe sets O_CLOEXEC, which is what we want.
			reader, writer, e := os.Pipe()
			if e != nil {
				return fm.errorpf(op, "failed to create pipe: %s", e)
			}
			ch := make(chan any, pipelineChanBufferSize)
			sendStop := make(chan struct{})
			sendError := new(error)
			readerGone := new(atomic.Bool)
			newFm.ports[1] = &Port{
				File: writer, Chan: ch,
				sendStop: sendStop, sendError: sendError, readerGone: readerGone}
			*growAccess(&fops, 1) = formOwnedPort{File: true, Chan: true}
			nextIn = &Port{
				File: reader, Chan: ch,
				// Store in input port for ease of retrieval later
				sendStop: sendStop, sendError: sendError, readerGone: readerGone}
		}
		f := func(form *formOp, fops []formOwnedPort, pexc *Exception) {
			exc := form.exec(newFm, &fops)
			if exc != nil && !(outputIsPipe && isReaderGone(exc)) && exc.Reason() != nil {
				*pexc = exc
			}
			if inputIsPipe {
				input := newFm.ports[0]
				*input.sendError = errs.ReaderGone{}
				close(input.sendStop)
				input.readerGone.Store(true)
			}
			for i, fop := range fops {
				fop.close(newFm.ports[i])
			}
			wg.Done()
		}
		if i == nforms-1 && !op.bg {
			f(form, fops, &excs[i])
		} else {
			go f(form, fops, &excs[i])
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

type formOp struct {
	diag.Ranging
	redirs []*redirOp
	body   formBody
}

func (cp *compiler) formOps(ns []*parse.Form) []*formOp {
	ops := make([]*formOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.formOp(n)
	}
	return ops
}

func (cp *compiler) formOp(n *parse.Form) *formOp {
	redirOps := cp.redirOps(n.Redirs)
	body := cp.formBody(n)

	return &formOp{n.Range(), redirOps, body}
}

type formBody struct {
	// Exactly one field will be populated.
	specialOp   effectOp
	ordinaryCmd ordinaryCmd
}

type ordinaryCmd struct {
	headOp valuesOp
	argOps []valuesOp
	optsOp *mapPairsOp
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
		} else {
			cp.autofixUnresolvedVar(head + FnSuffix)
			if cp.currentPragma().unknownCommandIsExternal || fsutil.DontSearch(head) {
				headOp = literalValues(n.Head, NewExternalCmd(head))
			} else {
				cp.errorpfPartial(n.Head, "unknown command disallowed by current pragma")
			}
		}
	} else {
		// Head is not a literal string: evaluate as a normal expression.
		headOp = cp.compoundOp(n.Head)
	}

	argOps := cp.compoundOps(n.Args)
	optsOp := cp.mapPairs(n.Opts)
	return formBody{ordinaryCmd: ordinaryCmd{headOp, argOps, optsOp}}
}

func (op *formOp) exec(fm *Frame, fops *[]formOwnedPort) (errRet Exception) {
	// fm here is always a sub-frame created in compiler.pipeline, so it can
	// be safely modified.

	// Redirections.
	for _, redirOp := range op.redirs {
		exc := redirOp.exec(fm, fops)
		if exc != nil {
			return exc
		}
	}

	if op.body.specialOp != nil {
		return op.body.specialOp.exec(fm)
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

	var args []any
	for _, argOp := range cmd.argOps {
		moreArgs, exc := argOp.exec(fm)
		if exc != nil {
			return exc
		}
		args = append(args, moreArgs...)
	}

	// TODO(xiaq): This conversion should be avoided.
	convertedOpts := make(map[string]any)
	exc := cmd.optsOp.exec(fm, func(k, v any) Exception {
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

func allTrue(vs []any) bool {
	for _, v := range vs {
		if !vals.Bool(v) {
			return false
		}
	}
	return true
}

type redirOp struct {
	diag.Ranging
	dstOp   valuesOp
	srcOp   valuesOp
	srcIsFd bool
	mode    parse.RedirMode
	flag    int
}

func (cp *compiler) redirOps(ns []*parse.Redir) []*redirOp {
	ops := make([]*redirOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.redirOp(n)
	}
	return ops
}

const defaultFileRedirPerm = 0644

// redir compiles a Redir into a op.
func (cp *compiler) redirOp(n *parse.Redir) *redirOp {
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

type InvalidFD struct{ FD int }

func (err InvalidFD) Error() string { return fmt.Sprintf("invalid fd: %d", err.FD) }

func (op *redirOp) exec(fm *Frame, fops *[]formOwnedPort) Exception {
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

	dstPort := growAccess(&fm.ports, dst)
	dstFop := growAccess(fops, dst)
	if *dstPort != nil {
		dstFop.close(*dstPort)
		*dstFop = formOwnedPort{File: false, Chan: false}
	}

	if op.srcIsFd {
		src, err := evalForFd(fm, op.srcOp, true, "redirection source")
		if err != nil {
			return fm.errorp(op, err)
		}
		switch {
		case src == -1:
			// close
			*dstPort = &Port{
				// Ensure that writing to value output throws an exception
				sendStop: closedSendStop, sendError: &ErrPortDoesNotSupportValueOutput}
		case src >= len(fm.ports) || fm.ports[src] == nil:
			return fm.errorp(op, InvalidFD{FD: src})
		default:
			*dstPort = fm.ports[src]
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
		*dstPort = fileRedirPort(op.mode, f)
		dstFop.File = true
	case vals.File:
		*dstPort = fileRedirPort(op.mode, src)
	default:
		if _, isMap := src.(vals.Map); !isMap && !vals.IsFieldMap(src) {
			return fm.errorp(op.srcOp, errs.BadValue{
				What:  "redirection source",
				Valid: "string, file or map", Actual: vals.Kind(src)})
		}
		var srcFile *os.File
		switch op.mode {
		case parse.Read:
			v, err := vals.Index(src, "r")
			f, ok := v.(*os.File)
			if err != nil || !ok {
				return fm.errorp(op.srcOp, errs.BadValue{
					What:   "map for input redirection",
					Valid:  "map with file in the 'r' field",
					Actual: vals.ReprPlain(src)})
			}
			srcFile = f
		case parse.Write:
			v, err := vals.Index(src, "w")
			f, ok := v.(*os.File)
			if err != nil || !ok {
				return fm.errorp(op.srcOp, errs.BadValue{
					What:   "map for output redirection",
					Valid:  "map with file in the 'w' field",
					Actual: vals.ReprPlain(src)})
			}
			srcFile = f
		default:
			return fm.errorpf(op, "can only use < or > with maps")
		}
		*dstPort = fileRedirPort(op.mode, srcFile)
	}
	return nil
}

// Creates a port that only have a file component, populating the
// channel-related fields with suitable values depending on the redirection
// mode.
func fileRedirPort(mode parse.RedirMode, f *os.File) *Port {
	if mode == parse.Read {
		return &Port{
			File: f,
			// ClosedChan produces no values when reading.
			Chan: ClosedChan,
		}
	}
	return &Port{
		File: f,
		// Throws errValueOutputIsClosed when writing.
		Chan: nil, sendStop: closedSendStop, sendError: &ErrPortDoesNotSupportValueOutput,
	}
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

// Accesses s[i], growing the slice with zero values if necessary.
func growAccess[T any](s *[]T, i int) *T {
	if i >= len(*s) {
		old := *s
		*s = make([]T, i+1)
		copy(*s, old)
	}
	return &(*s)[i]
}
