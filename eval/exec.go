package eval

import (
	"fmt"
	"os"
	"syscall"

	"github.com/elves/elvish/errutil"
)

const (
	// FdNil is a special impossible fd value. Used for "close fd" in
	// syscall.ProcAttr.Files.
	fdNil uintptr = ^uintptr(0)
)

// execSpecial executes a builtin special form.
//
// NOTE(xiaq): execSpecial and execNonSpecial are always called on an
// intermediate "form redir" where only the form-local ports are marked
// shouldClose. ec.closePorts should be called at appropriate moments.
func (ec *evalCtx) execSpecial(op exitusOp) <-chan exitus {
	update := make(chan exitus)
	go func() {
		ex := op(ec)
		// Ports are closed after executaion of builtin is complete.
		ec.closePorts()
		update <- ex
		close(update)
	}()
	return update
}

func (ec *evalCtx) resolveNonSpecial(cmd Value) callable {
	// Closure
	if cl, ok := cmd.(callable); ok {
		return cl
	}

	cmdStr := toString(cmd)

	// Defined callable
	ns, name := splitQualifiedName(cmdStr)
	if v := ec.ResolveVar(ns, FnPrefix+name); v != nil {
		if clb, ok := v.Get().(callable); ok {
			return clb
		}
	}

	// External command
	return externalCmd{cmdStr}
}

// execNonSpecial executes a form that is not a special form.
func (ec *evalCtx) execNonSpecial(cmd Value, args []Value) <-chan exitus {
	return ec.resolveNonSpecial(cmd).Exec(ec, args)
}

// Exec executes a builtin function.
func (b *builtinFn) Exec(ec *evalCtx, args []Value) <-chan exitus {
	update := make(chan exitus)
	go func() {
		ex := b.Impl(ec, args)
		// Ports are closed after executaion of builtin is complete.
		ec.closePorts()
		update <- ex
		close(update)
	}()
	return update
}

var (
	arityMismatch = newFailure("arity mismatch")
)

// Exec executes a closure.
func (c *closure) Exec(ec *evalCtx, args []Value) <-chan exitus {
	update := make(chan exitus, 1)

	// TODO Support optional/rest argument
	if len(args) != len(c.ArgNames) {
		// TODO Check arity before exec'ing
		update <- arityMismatch
		close(update)
		return update
	}

	// This evalCtx is dedicated to the current form, so we modify it in place.
	// BUG(xiaq): When evaluating closures, async access to global variables
	// and ports can be problematic.

	// Make upvalue namespace and capture variables.
	ec.up = make(map[string]Variable)
	for name, variable := range c.Captured {
		ec.up[name] = variable
	}
	// Make local namespace and pass arguments.
	ec.local = make(map[string]Variable)
	for i, name := range c.ArgNames {
		ec.local[name] = newInternalVariable(args[i])
	}

	// TODO(xiaq): Also change ec.name and ec.text since the closure being
	// called can come from another source.

	go func() {
		vs, err := ec.eval(c.Op)
		if err != nil {
			fmt.Print(err.(*errutil.ContextualError).Pprint())
		}
		// Ports are closed after executaion of closure is complete.
		ec.closePorts()
		if HasFailure(vs) {
			var flow exitusSort
			es := make([]exitus, len(vs))
			// NOTE(xiaq): If there is a flow exitus, the last one is
			// re-returned. Maybe we could use a more elegant semantics.
			for i, v := range vs {
				es[i] = v.(exitus)
				if es[i].Sort >= FlowSortLower {
					flow = es[i].Sort
				}
			}
			if flow != 0 {
				update <- newFlowExitus(flow)
			} else {
				update <- newTraceback(es)
			}
		} else {
			update <- ok
		}
		close(update)
	}()
	return update
}

// waitStatusToExitus converts syscall.WaitStatus to an exitus.
func waitStatusToExitus(ws syscall.WaitStatus) exitus {
	switch {
	case ws.Exited():
		es := ws.ExitStatus()
		if es == 0 {
			return ok
		}
		return newFailure(fmt.Sprint(es))
	case ws.Signaled():
		msg := fmt.Sprintf("signaled %v", ws.Signal())
		if ws.CoreDump() {
			msg += " (core dumped)"
		}
		return newFailure(msg)
	case ws.Stopped():
		msg := fmt.Sprintf("stopped %v", ws.StopSignal())
		trap := ws.TrapCause()
		if trap != -1 {
			msg += fmt.Sprintf(" (trapped %v)", trap)
		}
		return newFailure(msg)
	/*
		case ws.Continued():
			return newUnexitedStateUpdate("continued")
	*/
	default:
		return newFailure(fmt.Sprint("unknown WaitStatus", ws))
	}
}

// waitStateUpdate wait(2)s for pid, and feeds the WaitStatus into an
// exitus channel after proper conversion.
func waitStateUpdate(pid int, update chan<- exitus) {
	var ws syscall.WaitStatus
	_, err := syscall.Wait4(pid, &ws, 0, nil)

	if err != nil {
		update <- newFailure(fmt.Sprintf("wait:", err.Error()))
	} else {
		update <- waitStatusToExitus(ws)
	}

	close(update)
}

var (
	cdNoArg = newFailure("implicit cd accepts no arguments")
)

// Exec executes an external command.
func (e externalCmd) Exec(ec *evalCtx, argVals []Value) <-chan exitus {
	update := make(chan exitus, 1)

	if DontSearch(e.Name) {
		stat, err := os.Stat(e.Name)
		if err == nil && stat.IsDir() {
			// implicit cd
			ex := cdInner(e.Name, ec)
			ec.closePorts()
			update <- ex
			close(update)
			return update
		}
	}

	files := make([]uintptr, len(ec.ports))
	for i, port := range ec.ports {
		if port == nil || port.f == nil {
			files[i] = fdNil
		} else {
			files[i] = port.f.Fd()
		}
	}

	args := make([]string, len(argVals)+1)
	for i, a := range argVals {
		// NOTE Maybe we should enfore string arguments instead of coercing all
		// args into string
		args[i+1] = toString(a)
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: os.Environ(), Files: files[:], Sys: &sys}

	path, err := ec.Search(e.Name)
	var pid int

	if err == nil {
		args[0] = path
		pid, err = syscall.ForkExec(path, args, &attr)
	}
	// Ports are closed after fork-exec of external is complete.
	ec.closePorts()

	if err != nil {
		update <- newFailure(err.Error())
		close(update)
	} else {
		go waitStateUpdate(pid, update)
	}

	return update
}

func (t *list) Exec(ec *evalCtx, argVals []Value) <-chan exitus {
	update := make(chan exitus)
	go func() {
		var v Value = t
		for _, idx := range argVals {
			// XXX the positions are obviously wrong.
			v = evalSubscript(ec, v, idx, 0, 0)
		}
		ec.ports[1].ch <- v
		ec.closePorts()
		update <- ok
		close(update)
	}()
	return update
}

// XXX duplicate
func (t map_) Exec(ec *evalCtx, argVals []Value) <-chan exitus {
	update := make(chan exitus)
	go func() {
		var v Value = t
		for _, idx := range argVals {
			// XXX the positions are obviously wrong.
			v = evalSubscript(ec, v, idx, 0, 0)
		}
		ec.ports[1].ch <- v
		ec.closePorts()
		update <- ok
		close(update)
	}()
	return update
}
