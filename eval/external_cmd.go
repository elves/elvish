package eval

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// FdNil is a special impossible fd value used for "close fd" in
// syscall.ProcAttr.Files.
const fdNil uintptr = ^uintptr(0)

var (
	ErrExternalCmdOpts = errors.New("external commands don't accept elvish options")
	ErrCdNoArg         = errors.New("implicit cd accepts no arguments")
)

// ExternalCmd is an external command.
type ExternalCmd struct {
	Name string
}

func (ExternalCmd) Kind() string {
	return "fn"
}

func (e ExternalCmd) Repr(int) string {
	return "<external " + parse.Quote(e.Name) + ">"
}

// Call calls an external command.
func (e ExternalCmd) Call(ec *EvalCtx, argVals []Value, opts map[string]Value) {
	if len(opts) > 0 {
		throw(ErrExternalCmdOpts)
	}
	if util.DontSearch(e.Name) {
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

	sys := syscall.SysProcAttr{Setpgid: ec.background}
	attr := syscall.ProcAttr{Env: os.Environ(), Files: files[:], Sys: &sys}

	path, err := ec.Search(e.Name)
	if err != nil {
		throw(err)
	}

	args[0] = path
	pid, err := syscall.ForkExec(path, args, &attr)
	if err != nil {
		throw(errors.New("forkExec: " + err.Error()))
	}

	var ws syscall.WaitStatus
	_, err = syscall.Wait4(pid, &ws, syscall.WUNTRACED, nil)

	if err != nil {
		throw(fmt.Errorf("wait: %s", err.Error()))
	} else {
		maybeThrow(NewExternalCmdExit(e.Name, ws, pid))
	}
}
