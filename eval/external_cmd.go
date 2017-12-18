package eval

import (
	"errors"
	"os"
	"os/exec"
	"syscall"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hash"
)

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

func (e ExternalCmd) Equal(a interface{}) bool {
	return e == a
}

func (e ExternalCmd) Hash() uint32 {
	return hash.String(e.Name)
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

	files := make([]*os.File, len(ec.ports))
	for i, port := range ec.ports {
		files[i] = port.File
	}

	args := make([]string, len(argVals)+1)
	for i, a := range argVals {
		// NOTE Maybe we should enfore string arguments instead of coercing all
		// args into string
		args[i+1] = ToString(a)
	}

	path, err := exec.LookPath(e.Name)
	if err != nil {
		throw(err)
	}

	args[0] = path

	sys := makeSysProcAttr(ec.background)
	proc, err := os.StartProcess(path, args, &os.ProcAttr{Files: files, Sys: sys})

	if err != nil {
		throw(err)
	}

	state, err := proc.Wait()

	if err != nil {
		throw(err)
	} else {
		maybeThrow(NewExternalCmdExit(e.Name, state.Sys().(syscall.WaitStatus), proc.Pid))
	}
}
