package eval

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hashmap"
)

func (cp *compiler) chunk(n *parse.Chunk) effectOpBody {
	return chunkOp{cp.pipelineOps(n.Pipelines)}
}

type chunkOp struct {
	subops []effectOp
}

func (op chunkOp) invoke(fm *Frame) error {
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
		return ErrInterrupted
	}
	return nil
}

func (cp *compiler) pipeline(n *parse.Pipeline) effectOpBody {
	return &pipelineOp{n.Background, n.SourceText(), cp.formOps(n.Forms)}
}

type pipelineOp struct {
	bg     bool
	source string
	subops []effectOp
}

const pipelineChanBufferSize = 32

func (op *pipelineOp) invoke(fm *Frame) error {
	if fm.IsInterrupted() {
		return ErrInterrupted
	}

	if op.bg {
		fm = fm.fork("background job" + op.source)
		fm.intCh = nil
		fm.background = true
		fm.Evaler.state.addNumBgJobs(1)

		if fm.Editor != nil {
			// TODO: Redirect output in interactive mode so that the line
			// editor does not get messed up.
		}
	}

	nforms := len(op.subops)

	var wg sync.WaitGroup
	wg.Add(nforms)
	errors := make([]*Exception, nforms)

	var nextIn *Port

	// For each form, create a dedicated evalCtx and run asynchronously
	for i, op := range op.subops {
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
				return fmt.Errorf("failed to create pipe: %s", e)
			}
			ch := make(chan interface{}, pipelineChanBufferSize)
			newFm.ports[1] = &Port{
				File: writer, Chan: ch, CloseFile: true, CloseChan: true}
			nextIn = &Port{
				File: reader, Chan: ch, CloseFile: true, CloseChan: false}
		}
		thisOp := op
		thisError := &errors[i]
		go func() {
			err := newFm.eval(thisOp)
			newFm.Close()
			if err != nil {
				*thisError = err.(*Exception)
			}
			wg.Done()
			if hasChanInput {
				// If the command has channel input, drain it. This
				// mitigates the effect of erroneous pipelines like
				// "range 100 | cat"; without draining the pipeline will
				// lock up.
				for range newFm.ports[0].Chan {
				}
			}
		}()
	}

	if op.bg {
		// Background job, wait for form termination asynchronously.
		go func() {
			wg.Wait()
			fm.Evaler.state.addNumBgJobs(-1)
			msg := "job " + op.source + " finished"
			err := ComposeExceptionsFromPipeline(errors)
			if err != nil {
				msg += ", errors = " + err.Error()
			}
			if fm.Evaler.state.getNotifyBgJobSuccess() || err != nil {
				if fm.Editor != nil {
					fm.Editor.Notify("%s", msg)
				} else {
					fm.ports[2].File.WriteString(msg + "\n")
				}
			}
		}()
		return nil
	} else {
		wg.Wait()
		return ComposeExceptionsFromPipeline(errors)
	}
}

func (cp *compiler) form(n *parse.Form) effectOpBody {
	var saveVarsOps []lvaluesOp
	var assignmentOps []effectOp
	if len(n.Assignments) > 0 {
		assignmentOps = cp.assignmentOps(n.Assignments)
		if n.Head == nil && n.Vars == nil {
			// Permanent assignment.
			return seqOp{assignmentOps}
		}
		for _, a := range n.Assignments {
			v, r := cp.lvaluesOp(a.Left)
			saveVarsOps = append(saveVarsOps, v, r)
		}
		logger.Println("temporary assignment of", len(n.Assignments), "pairs")
	}

	// Depending on the type of the form, exactly one of the three below will be
	// set.
	var (
		specialOpFunc  effectOpBody
		headOp         valuesOp
		spaceyAssignOp effectOp
	)

	// Forward declaration; needed when compiling assignment forms.
	var argOps []valuesOp

	if n.Head != nil {
		headStr, ok := oneString(n.Head)
		if ok {
			compileForm, ok := builtinSpecials[headStr]
			if ok {
				// Special form.
				specialOpFunc = compileForm(cp, n)
			} else {
				var headOpFunc valuesOpBody
				explode, ns, name := ParseVariableRef(headStr)
				if !explode && cp.registerVariableGet(ns, name+FnSuffix) {
					// $head~ resolves.
					headOpFunc = variableOp{false, ns, name + FnSuffix}
				} else {
					// Fall back to $e:head~.
					headOpFunc = literalValues(ExternalCmd{headStr})
				}
				headOp = valuesOp{headOpFunc, n.Head.Range().From, n.Head.Range().To}
			}
		} else {
			// Head exists and is not a literal string. Evaluate as a normal
			// expression.
			headOp = cp.compoundOp(n.Head)
		}
	} else {
		// Assignment form.
		varsOp, restOp := cp.lvaluesMulti(n.Vars)
		// This cannot be replaced with newSeqValuesOp as it depends on the fact
		// that argOps will be changed later.
		argsOp := valuesOp{
			funcValuesOp(func(fm *Frame) ([]interface{}, error) {
				var values []interface{}
				for _, op := range argOps {
					moreValues, err := op.exec(fm)
					if err != nil {
						return nil, err
					}
					values = append(values, moreValues...)
				}
				return values, nil
			}), -1, -1}
		if len(argOps) > 0 {
			argsOp.begin = argOps[0].begin
			argsOp.end = argOps[len(argOps)-1].end
		}
		spaceyAssignOp = effectOp{
			&assignmentOp{varsOp, restOp, argsOp},
			n.Range().From, argsOp.end,
		}
	}

	argOps = cp.compoundOps(n.Args)
	optsOp := cp.mapPairs(n.Opts)
	redirOps := cp.redirOps(n.Redirs)
	// TODO: n.ErrorRedir

	return &formOp{saveVarsOps, assignmentOps, redirOps, specialOpFunc, headOp, argOps, optsOp, spaceyAssignOp, n.Range().From, n.Range().To}
}

type formOp struct {
	saveVarsOps    []lvaluesOp
	assignmentOps  []effectOp
	redirOps       []effectOp
	specialOpBody  effectOpBody
	headOp         valuesOp
	argOps         []valuesOp
	optsOp         valuesOpBody
	spaceyAssignOp effectOp
	begin, end     int
}

func (op *formOp) invoke(fm *Frame) (errRet error) {
	// ec here is always a subevaler created in compiler.pipeline, so it can
	// be safely modified.

	// Temporary assignment.
	if len(op.saveVarsOps) > 0 {
		// There is a temporary assignment.
		// Save variables.
		var saveVars []vars.Var
		var saveVals []interface{}
		for _, op := range op.saveVarsOps {
			moreSaveVars, err := op.exec(fm)
			if err != nil {
				return err
			}
			saveVars = append(saveVars, moreSaveVars...)
		}
		for i, v := range saveVars {
			// XXX(xiaq): If the variable to save is a elemVariable, save
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
					// XXX Old value is nonexistent. We should delete the
					// variable. However, since the compiler now doesn't delete
					// it, we don't delete it in the evaler either.
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

	if op.specialOpBody != nil {
		return op.specialOpBody.invoke(fm)
	}
	var headFn Callable
	var args []interface{}
	if op.headOp.body != nil {
		// head
		headFn, errRet = fm.ExecAndUnwrap("head of command", op.headOp).One().Callable()
		if errRet != nil {
			return errRet
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
	// XXX This conversion should be avoided.
	optValues, err := op.optsOp.invoke(fm)
	if err != nil {
		return err
	}
	opts := optValues[0].(hashmap.Map)
	convertedOpts := make(map[string]interface{})
	for it := opts.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		if ks, ok := k.(string); ok {
			convertedOpts[ks] = v
		} else {
			return fmt.Errorf("Option key must be string, got %s", vals.Kind(k))
		}
	}

	fm.begin, fm.end = op.begin, op.end

	if headFn != nil {
		return headFn.Call(fm, args, convertedOpts)
	} else {
		return op.spaceyAssignOp.exec(fm)
	}
}

func allTrue(vs []interface{}) bool {
	for _, v := range vs {
		if !vals.Bool(v) {
			return false
		}
	}
	return true
}

func (cp *compiler) assignment(n *parse.Assignment) effectOpBody {
	variablesOp, restOp := cp.lvaluesOp(n.Left)
	valuesOp := cp.compoundOp(n.Right)
	return &assignmentOp{variablesOp, restOp, valuesOp}
}

// ErrMoreThanOneRest is returned when the LHS of an assignment contains more
// than one rest variables.
var ErrMoreThanOneRest = errors.New("more than one @ lvalue")

type assignmentOp struct {
	variablesOp lvaluesOp
	restOp      lvaluesOp
	valuesOp    valuesOp
}

func (op *assignmentOp) invoke(fm *Frame) (errRet error) {
	variables, err := op.variablesOp.exec(fm)
	if err != nil {
		return err
	}
	rest, err := op.restOp.exec(fm)
	if err != nil {
		return err
	}

	values, err := op.valuesOp.exec(fm)
	if err != nil {
		return err
	}

	if len(rest) > 1 {
		return ErrMoreThanOneRest
	}
	if len(rest) == 1 {
		if len(variables) > len(values) {
			return ErrArityMismatch
		}
	} else {
		if len(variables) != len(values) {
			return ErrArityMismatch
		}
	}

	for i, variable := range variables {
		err := variable.Set(values[i])
		if err != nil {
			return err
		}
	}

	if len(rest) == 1 {
		err := rest[0].Set(vals.MakeList(values[len(variables):]...))
		if err != nil {
			return err
		}
	}
	return nil
}

func fixNilVariables(vs []vars.Var, perr *error) {
	for _, v := range vs {
		if vars.IsBlackhole(v) {
			continue
		}
		if v.Get() == nil {
			err := v.Set("")
			*perr = util.Errors(*perr, err)
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
func (cp *compiler) redir(n *parse.Redir) effectOpBody {
	var dstOp valuesOp
	if n.Left != nil {
		dstOp = cp.compoundOp(n.Left)
	}
	flag := makeFlag(n.Mode)
	if flag == -1 {
		// TODO: Record and get redirection sign position
		cp.errorf("bad redirection sign")
	}
	return &redirOp{dstOp, cp.compoundOp(n.Right), n.RightIsFd, n.Mode, flag}
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
	dstOp   valuesOp
	srcOp   valuesOp
	srcIsFd bool
	mode    parse.RedirMode
	flag    int
}

func (op *redirOp) invoke(fm *Frame) error {
	var dst int
	if op.dstOp.body == nil {
		// use default dst fd
		switch op.mode {
		case parse.Read:
			dst = 0
		case parse.Write, parse.ReadWrite, parse.Append:
			dst = 1
		default:
			return fmt.Errorf("bad RedirMode; parser bug")
		}
	} else {
		var err error
		// dst must be a valid fd
		dst, err = fm.ExecAndUnwrap("Fd", op.dstOp).One().NonNegativeInt()
		if err != nil {
			return err
		}
	}

	fm.growPorts(dst + 1)
	// Logger.Printf("closing old port %d of %s", dst, ec.context)
	fm.ports[dst].Close()

	srcUnwrap := fm.ExecAndUnwrap("redirection source", op.srcOp).One()
	if op.srcIsFd {
		src, err := srcUnwrap.FdOrClose()
		if err != nil {
			return err
		}
		if src == -1 {
			// close
			fm.ports[dst] = &Port{}
		} else {
			fm.ports[dst] = fm.ports[src].Fork()
		}
	} else {
		src, err := srcUnwrap.Any()
		if err != nil {
			return err
		}
		switch src := src.(type) {
		case string:
			f, err := os.OpenFile(src, op.flag, defaultFileRedirPerm)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %s", vals.Repr(src, vals.NoPretty), err)
			}
			fm.ports[dst] = &Port{
				File: f, Chan: BlackholeChan,
				CloseFile: true,
			}
		case vals.File:
			fm.ports[dst] = &Port{
				File: src.Inner, Chan: BlackholeChan,
				CloseFile: false,
			}
		case vals.Pipe:
			var f *os.File
			switch op.mode {
			case parse.Read:
				f = src.ReadEnd
			case parse.Write:
				f = src.WriteEnd
			default:
				return errors.New("can only use < or > with pipes")
			}
			fm.ports[dst] = &Port{
				File: f, Chan: BlackholeChan,
				CloseFile: false,
			}
		default:
			srcUnwrap.error("string or file", "%s", vals.Kind(src))
		}
	}
	return nil
}

type seqOp struct{ subops []effectOp }

func (op seqOp) invoke(fm *Frame) error {
	for _, subop := range op.subops {
		err := subop.exec(fm)
		if err != nil {
			return err
		}
	}
	return nil
}

type funcOp func(*Frame) error

func (op funcOp) invoke(fm *Frame) error {
	return op(fm)
}
