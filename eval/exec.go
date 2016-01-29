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

var (
	arityMismatch = NewFailure("arity mismatch")
	cdNoArg       = NewFailure("implicit cd accepts no arguments")
	evalFailure   = NewFailure("generic eval failure")
)

func (ec *evalCtx) exec(op exitusOp) Exitus {
	ex := op(ec)
	ec.closePorts()
	return ex
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

// Call calls a builtin function.
func (b *builtinFn) Call(ec *evalCtx, args []Value) Exitus {
	return b.Impl(ec, args)
}

// Call calls a closure.
func (c *closure) Call(ec *evalCtx, args []Value) Exitus {
	// TODO Support optional/rest argument
	if len(args) != len(c.ArgNames) {
		return arityMismatch
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

	ex, err := ec.eval(c.Op)
	if err != nil {
		fmt.Print(err.(*errutil.ContextualError).Pprint())
		return evalFailure
	}

	return ex
}

// waitStatusToExitus converts syscall.WaitStatus to an exitus.
func waitStatusToExitus(ws syscall.WaitStatus) Exitus {
	switch {
	case ws.Exited():
		es := ws.ExitStatus()
		if es == 0 {
			return OK
		}
		return NewFailure(fmt.Sprint(es))
	case ws.Signaled():
		msg := fmt.Sprintf("signaled %v", ws.Signal())
		if ws.CoreDump() {
			msg += " (core dumped)"
		}
		return NewFailure(msg)
	case ws.Stopped():
		msg := fmt.Sprintf("stopped %v", ws.StopSignal())
		trap := ws.TrapCause()
		if trap != -1 {
			msg += fmt.Sprintf(" (trapped %v)", trap)
		}
		return NewFailure(msg)
	/*
		case ws.Continued():
			return newUnexitedStateUpdate("continued")
	*/
	default:
		return NewFailure(fmt.Sprint("unknown WaitStatus", ws))
	}
}

// Call calls an external command.
func (e externalCmd) Call(ec *evalCtx, argVals []Value) Exitus {
	if DontSearch(e.Name) {
		stat, err := os.Stat(e.Name)
		if err == nil && stat.IsDir() {
			// implicit cd
			return cdInner(e.Name, ec)
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
	if err != nil {
		return NewFailure("search: " + err.Error())
	}

	args[0] = path
	pid, err := syscall.ForkExec(path, args, &attr)
	if err != nil {
		return NewFailure("forkExec: " + err.Error())
	}

	var ws syscall.WaitStatus
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		return NewFailure(fmt.Sprintf("wait:", err.Error()))
	} else {
		return waitStatusToExitus(ws)
	}
}

func (t *list) Call(ec *evalCtx, argVals []Value) Exitus {
	var v Value = t
	for _, idx := range argVals {
		// XXX the positions are obviously wrong.
		v = evalSubscript(ec, v, idx, 0, 0)
	}
	ec.ports[1].ch <- v
	return OK
}

// XXX duplicate
func (t map_) Call(ec *evalCtx, argVals []Value) Exitus {
	var v Value = t
	for _, idx := range argVals {
		// XXX the positions are obviously wrong.
		v = evalSubscript(ec, v, idx, 0, 0)
	}
	ec.ports[1].ch <- v
	return OK
}
