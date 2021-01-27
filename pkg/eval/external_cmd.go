package eval

import (
	"errors"
	"os"
	"os/exec"
	"syscall"

	"github.com/xiaq/persistent/hash"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
)

var (
	// ErrExternalCmdOpts is thrown when an external command is passed Elvish
	// options.
	//
	// TODO: Catch this kind of errors at compilation time.
	ErrExternalCmdOpts = errors.New("external commands don't accept elvish options")
	// ErrImplicitCdNoArg is thrown when an implicit cd form is passed arguments.
	ErrImplicitCdNoArg = errors.New("implicit cd accepts no arguments")
)

// externalCmd is an external command.
type externalCmd struct {
	Name string
}

// NewExternalCmd returns a callable that executes the named external command.
//
// An external command converts all arguments to strings, and does not accept
// any option.
func NewExternalCmd(name string) Callable {
	return externalCmd{name}
}

func (e externalCmd) Kind() string {
	return "fn"
}

func (e externalCmd) Equal(a interface{}) bool {
	return e == a
}

func (e externalCmd) Hash() uint32 {
	return hash.String(e.Name)
}

func (e externalCmd) Repr(int) string {
	return "<external " + parse.Quote(e.Name) + ">"
}

// Call calls an external command.
func (e externalCmd) Call(fm *Frame, argVals []interface{}, opts map[string]interface{}) error {
	if len(opts) > 0 {
		return ErrExternalCmdOpts
	}
	if fsutil.DontSearch(e.Name) {
		stat, err := os.Stat(e.Name)
		if err == nil && stat.IsDir() {
			// implicit cd
			if len(argVals) > 0 {
				return ErrImplicitCdNoArg
			}
			return fm.Evaler.Chdir(e.Name)
		}
	}

	files := make([]*os.File, len(fm.ports))
	for i, port := range fm.ports {
		if port != nil {
			files[i] = port.File
		}
	}

	args := make([]string, len(argVals)+1)
	for i, a := range argVals {
		// TODO: Maybe we should enforce string arguments instead of coercing
		// all args to strings.
		args[i+1] = vals.ToString(a)
	}

	path, err := exec.LookPath(e.Name)
	if err != nil {
		return err
	}

	args[0] = path

	sys := makeSysProcAttr(fm.background)
	proc, err := os.StartProcess(path, args, &os.ProcAttr{Files: files, Sys: sys})
	if err != nil {
		return err
	}

	state, err := proc.Wait()
	if err != nil {
		// This should be a can't happen situation. Nonetheless, treat it as a
		// soft error rather than panicking since the Go documentation is not
		// explicit that this can only happen if we make a mistake. Such as
		// calling `Wait` twice on a particular process object.
		return err
	}
	return NewExternalCmdExit(e.Name, state.Sys().(syscall.WaitStatus), proc.Pid)
}
