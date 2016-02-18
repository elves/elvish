package eval

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// FdNil is a special impossible fd value used for "close fd" in
// syscall.ProcAttr.Files.
const fdNil uintptr = ^uintptr(0)

var ErrCdNoArg = errors.New("implicit cd accepts no arguments")

// ExternalCmd is an external command.
type ExternalCmd struct {
	Name string
}

func (ExternalCmd) Kind() string {
	return "fn"
}

func (e ExternalCmd) Repr() string {
	return "<external " + e.Name + " >"
}

// Call calls an external command.
func (e ExternalCmd) Call(ec *EvalCtx, argVals []Value) {
	if DontSearch(e.Name) {
		stat, err := os.Stat(e.Name)
		if err == nil && stat.IsDir() {
			// implicit cd
			if len(argVals) > 0 {
				throw(ErrCdNoArg)
			}
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
