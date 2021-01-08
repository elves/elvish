package eval

import (
	"fmt"
	"os"
	"sync"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/errs"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/fsutil"
	"github.com/elves/elvish/pkg/parse"
)

// An operation with some side effects.
type effectOp interface{ exec(*Frame) error }

// An effectOp that creates all variables in a scope before executing the body.
type scopeOp struct {
	inner  effectOp
	locals []string
}

func wrapScopeOp(op effectOp, locals []string) effectOp {
	return scopeOp{op, locals}
}

func (op scopeOp) Range() diag.Ranging { return op.inner.(diag.Ranger).Range() }

func (op scopeOp) exec(fm *Frame) error {
	if len(op.locals) == 0 {
		return op.inner.exec(fm)
	}
	fm.local.names = append(fm.local.names, op.locals...)
	for _, name := range op.locals {
		fm.local.slots = append(fm.local.slots, makeVarFromName(name))
	}
	return op.inner.exec(fm)
}

func (cp *compiler) chunkOp(n *parse.Chunk) effectOp {
	return chunkOp{n.Range(), cp.pipelineOps(n.Pipelines)}
}

type chunkOp struct {
	diag.Ranging
	subops []effectOp
}

func (op chunkOp) exec(fm *Frame) error {
	for _, subop := range op.subops {
		err := subop.exec(fm)
		if err != nil {
			return err
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

func (op *pipelineOp) exec(fm *Frame) error {
	if fm.IsInterrupted() {
		return fm.errorp(op, ErrInterrupted)
	}

	if op.bg {
		fm = fm.fork("background job" + op.source)
		fm.intCh = nil
		fm.background = true
		fm.Evaler.addNumBgJobs(1)

		if fm.Evaler.Editor() != nil {
			// TODO: Redirect output in interactive mode so that the line
			// editor does not get messed up.
		}
	}

	nforms := len(op.subops)

	var wg sync.WaitGroup
	wg.Add(nforms)
	errors := make([]Exception, nforms)

	var nextIn *Port

	// For each form, create a dedicated evalCtx and run asynchronously
	for i, formOp := range op.subops {
		hasChanInput := i > 0
		newFm := fm.fork("[form op]")
		if i > 0 {
			newFm.ports[0] = nextIn
		}
		if i < nforms-1 {
			// Each internal port pair consists of a (byte) pipe pair and a
			// channel.
			// os.Pipe sets O_CLOEXEC, which is what we want.
			reader, writer, e := os.Pipe()
			if e != nil {
				return fm.errorpf(op, "failed to create pipe: %s", e)
			}
			ch := make(chan interface{}, pipelineChanBufferSize)
			newFm.ports[1] = &Port{
				File: writer, Chan: ch, closeFile: true, closeChan: true}
			nextIn = &Port{
				File: reader, Chan: ch, closeFile: true, closeChan: false}
		}
		thisOp := formOp
		thisError := &errors[i]
		go func() {
			err := thisOp.exec(newFm)
			newFm.Close()
			if err != nil {
				switch err := err.(type) {
				case *exception:
					*thisError = err
				default:
					*thisError = &exception{reason: err, stackTrace: nil}
				}
			}
			wg.Done()
			if hasChanInput {
				// If the command has channel input, drain it. This
				// mitigates the effect of erroneous pipelines like
				// "range 100 | cat"; without draining the pipeline will
				// lock up.
				for range newFm.InputChan() {
				}
			}
		}()
	}

	if op.bg {
		// Background job, wait for form termination asynchronously.
		go func() {
			wg.Wait()
			fm.Evaler.addNumBgJobs(-1)
			msg := "job " + op.source + " finished"
			err := MakePipelineError(errors)
			if err != nil {
				msg += ", errors = " + err.Error()
			}
			if fm.Evaler.getNotifyBgJobSuccess() || err != nil {
				editor := fm.Evaler.Editor()
				if editor != nil {
					editor.Notify("%s", msg)
				} else {
					fm.ErrorFile().WriteString(msg + "\n")
				}
			}
		}()
		return nil
	}
	wg.Wait()
	return fm.errorp(op, MakePipelineError(errors))
}

func (cp *compiler) formOp(n *parse.Form) effectOp {
	var tempLValues []lvalue
	var assignmentOps []effectOp
	if len(n.Assignments) > 0 {
		assignmentOps = cp.assignmentOps(n.Assignments)
		if n.Head == nil && n.Vars == nil {
			// Permanent assignment.
			return seqOp{assignmentOps}
		}
		for _, a := range n.Assignments {
			lvalues := cp.parseIndexingLValue(a.Left)
			tempLValues = append(tempLValues, lvalues.lvalues...)
		}
		logger.Println("temporary assignment of", len(n.Assignments), "pairs")
	}

	// Depending on the type of the form, exactly one of the three below will be
	// set.
	var (
		specialOp      effectOp
		headOp         valuesOp
		spaceyAssignOp effectOp
	)

	// Forward declaration; needed when compiling assignment forms.
	var argOps []valuesOp

	if n.Head != nil {
		headStr, ok := oneString(n.Head)
		if ok {
			special, fnRef := resolveCmdHeadInternally(cp, headStr, n.Head)
			switch {
			case special != nil:
				specialOp = special(cp, n)
			case fnRef != nil:
				headOp = variableOp{n.Head.Range(), false, headStr + FnSuffix, fnRef}
			default:
				headOp = literalValues(n.Head, NewExternalCmd(headStr))
			}
		} else {
			// Head exists and is not a literal string. Evaluate as a normal
			// expression.
			headOp = cp.compoundOp(n.Head)
		}
		argOps = cp.compoundOps(n.Args)
	} else {
		// Assignment form.
		lhs := cp.parseCompoundLValues(n.Vars)
		argOps = cp.compoundOps(n.Args)
		var rhsRanging diag.Ranging
		if len(argOps) > 0 {
			rhsRanging = diag.MixedRanging(argOps[0], argOps[len(argOps)-1])
		} else {
			rhsRanging = diag.PointRanging(n.Range().To)
		}
		rhs := seqValuesOp{rhsRanging, argOps}
		spaceyAssignOp = &assignOp{n.Range(), lhs, rhs}
	}

	optsOp := cp.mapPairs(n.Opts)
	redirOps := cp.redirOps(n.Redirs)
	// TODO: n.ErrorRedir

	return &formOp{n.Range(), tempLValues, assignmentOps, redirOps, specialOp, headOp, argOps, optsOp, spaceyAssignOp}
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
	tempLValues    []lvalue
	assignmentOps  []effectOp
	redirOps       []effectOp
	specialOp      effectOp
	headOp         valuesOp
	argOps         []valuesOp
	optsOp         *mapPairsOp
	spaceyAssignOp effectOp
}

func (op *formOp) exec(fm *Frame) (errRet error) {
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
		for _, subop := range op.assignmentOps {
			err := subop.exec(fm)
			if err != nil {
				return err
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
					errRet = err
				}
				logger.Printf("restored %s = %s", v, val)
			}
		}()
	}

	// redirs
	for _, redirOp := range op.redirOps {
		err := redirOp.exec(fm)
		if err != nil {
			return err
		}
	}

	if op.specialOp != nil {
		return op.specialOp.exec(fm)
	}
	var headFn Callable
	var args []interface{}
	if op.headOp != nil {
		var err error
		// head
		headFn, err = evalForCommand(fm, op.headOp, "command")
		if err != nil {
			return fm.errorp(op.headOp, err)
		}

		// args
		for _, argOp := range op.argOps {
			moreArgs, err := argOp.exec(fm)
			if err != nil {
				return err
			}
			args = append(args, moreArgs...)
		}
	}

	// opts
	// TODO(xiaq): This conversion should be avoided.

	convertedOpts := make(map[string]interface{})
	err := op.optsOp.exec(fm, func(k, v interface{}) error {
		if ks, ok := k.(string); ok {
			convertedOpts[ks] = v
			return nil
		}
		// TODO(xiaq): Point to the particular key.
		return fm.errorp(op, errs.BadValue{
			What: "option key", Valid: "string", Actual: vals.Kind(k)})
	})
	if err != nil {
		return fm.errorp(op, err)
	}

	if headFn != nil {
		fm.traceback = fm.addTraceback(op)
		err := headFn.Call(fm, args, convertedOpts)
		if _, ok := err.(*exception); ok {
			return err
		}
		return &exception{err, fm.traceback}
	}
	return op.spaceyAssignOp.exec(fm)
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
		Actual: vals.Kind(value)})
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
	lhs := cp.parseIndexingLValue(n.Left)
	rhs := cp.compoundOp(n.Right)
	return &assignOp{n.Range(), lhs, rhs}
}

func (cp *compiler) assignmentOps(ns []*parse.Assignment) []effectOp {
	ops := make([]effectOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.assignmentOp(n)
	}
	return ops
}

func (cp *compiler) literal(n *parse.Primary, msg string) string {
	switch n.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return n.Value
	default:
		cp.errorpf(n, msg)
		return ""
	}
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

type invalidFD struct{ fd int }

func (err invalidFD) Error() string { return fmt.Sprintf("invalid fd: %d", err.fd) }

// Returns a suitable dummy value for the channel part of the port when
// redirecting from or to a file, so that the read and write attempts fail
// silently (instead of blocking or panicking).
//
// TODO: Instead of letting read and write attempts fail silently, consider
// raising an exception instead.
func chanForFileRedir(mode parse.RedirMode) chan interface{} {
	if mode == parse.Read {
		// ClosedChan produces no values when reading.
		return ClosedChan
	}
	// BlackholeChan discards all values written to it.
	return BlackholeChan
}

func (op *redirOp) exec(fm *Frame) error {
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
			fm.ports[dst] = &Port{}
		case src >= len(fm.ports) || fm.ports[src] == nil:
			return fm.errorp(op, invalidFD{src})
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
			return fm.errorpf(op, "failed to open file %s: %s", vals.Repr(src, vals.NoPretty), err)
		}
		fm.ports[dst] = &Port{File: f, closeFile: true, Chan: chanForFileRedir(op.mode)}
	case vals.File:
		fm.ports[dst] = &Port{File: src, closeFile: false, Chan: chanForFileRedir(op.mode)}
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
		fm.ports[dst] = &Port{File: f, closeFile: false, Chan: chanForFileRedir(op.mode)}
	default:
		return fm.errorp(op.srcOp, errs.BadValue{
			What:  "redirection source",
			Valid: "string, file or pipe", Actual: vals.Kind(src)})
	}
	return nil
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
		What: what, Valid: valid, Actual: vals.Repr(value, vals.NoPretty)})
}

type seqOp struct{ subops []effectOp }

func (op seqOp) exec(fm *Frame) error {
	for _, subop := range op.subops {
		err := subop.exec(fm)
		if err != nil {
			return err
		}
	}
	return nil
}
