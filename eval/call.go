package eval

import (
	"errors"
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
	ErrArityMismatch = errors.New("arity mismatch")
	ErrCdNoArg       = errors.New("implicit cd accepts no arguments")
)

// Caller is anything may be called on an evalCtx with a list of Value's.
type Caller interface {
	Value
	Call(ec *EvalCtx, args []Value)
}

func mustCaller(v Value) Caller {
	caller, ok := getCaller(v)
	if !ok {
		throw(fmt.Errorf("a %s is not callable", v.Kind()))
	}
	return caller
}

func getCaller(v Value) (Caller, bool) {
	if caller, ok := v.(Caller); ok {
		return caller, true
	}
	if indexer, ok := getIndexer(v); ok {
		return IndexerCaller{indexer}, true
	}
	return nil, false
}

func (ec *EvalCtx) PCall(f Caller, args []Value) (ex error) {
	defer errutil.Catch(&ex)
	f.Call(ec, args)
	return nil
}

// Call calls a builtin function.
func (b *BuiltinFn) Call(ec *EvalCtx, args []Value) {
	b.Impl(ec, args)
}

// Call calls a closure.
func (c *Closure) Call(ec *EvalCtx, args []Value) {
	// TODO Support optional/rest argument
	// TODO Support keyword arguments
	if !c.Variadic && len(args) != len(c.ArgNames) {
		throw(ErrArityMismatch)
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
	if !c.Variadic {
		for i, name := range c.ArgNames {
			ec.local[name] = NewPtrVariable(args[i])
		}
	}
	ec.local["args"] = NewPtrVariable(List{&args})
	ec.local["kwargs"] = NewPtrVariable(Map{&map[Value]Value{}})

	// TODO(xiaq): Also change ec.name and ec.text since the closure being
	// called can come from another source.

	c.Op(ec)
}

// waitStatusToError converts syscall.WaitStatus to an Error.
func waitStatusToError(ws syscall.WaitStatus) error {
	switch {
	case ws.Exited():
		es := ws.ExitStatus()
		if es == 0 {
			return nil
		}
		return errors.New(fmt.Sprint(es))
	case ws.Signaled():
		msg := fmt.Sprintf("signaled %v", ws.Signal())
		if ws.CoreDump() {
			msg += " (core dumped)"
		}
		return errors.New(msg)
	case ws.Stopped():
		msg := fmt.Sprintf("stopped %v", ws.StopSignal())
		trap := ws.TrapCause()
		if trap != -1 {
			msg += fmt.Sprintf(" (trapped %v)", trap)
		}
		return errors.New(msg)
	/*
		case ws.Continued():
			return newUnexitedStateUpdate("continued")
	*/
	default:
		return fmt.Errorf("unknown WaitStatus", ws)
	}
}

// Call calls an external command.
func (e ExternalCmd) Call(ec *EvalCtx, argVals []Value) {
	if DontSearch(e.Name) {
		stat, err := os.Stat(e.Name)
		if err == nil && stat.IsDir() {
			// implicit cd
			cdInner(e.Name, ec)
			return
		}
	}

	files := make([]uintptr, len(ec.ports))
	for i, port := range ec.ports {
		if port == nil || port.File == nil {
			files[i] = fdNil
		} else {
			files[i] = port.File.Fd()
		}
	}

	args := make([]string, len(argVals)+1)
	for i, a := range argVals {
		// NOTE Maybe we should enfore string arguments instead of coercing all
		// args into string
		args[i+1] = ToString(a)
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: os.Environ(), Files: files[:], Sys: &sys}

	path, err := ec.Search(e.Name)
	if err != nil {
		throw(errors.New("search: " + err.Error()))
	}

	args[0] = path
	pid, err := syscall.ForkExec(path, args, &attr)
	if err != nil {
		throw(errors.New("forkExec: " + err.Error()))
	}

	var ws syscall.WaitStatus
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		throw(fmt.Errorf("wait: %s", err.Error()))
	} else {
		maybeThrow(waitStatusToError(ws))
	}
}

// Call resolves a command name to either a Caller variable or external command
// and calls it.
func (s String) Call(ec *EvalCtx, args []Value) {
	resolve(string(s), ec).Call(ec, args)
}

func resolve(s string, ec *EvalCtx) Caller {
	// Try variable
	splice, ns, name := parseVariable(string(s))
	if !splice {
		if v := ec.ResolveVar(ns, FnPrefix+name); v != nil {
			if caller, ok := v.Get().(Caller); ok {
				return caller
			}
		}
	}

	// External command
	return ExternalCmd{string(s)}
}

// IndexerCaller is an adapter that makes it possible to use a Indexer as a
// Caller.
type IndexerCaller struct {
	Indexer
}

func (ic IndexerCaller) Call(ec *EvalCtx, args []Value) {
	results := ic.Indexer.Index(args)
	for _, v := range results {
		ec.ports[1].Chan <- v
	}
}
