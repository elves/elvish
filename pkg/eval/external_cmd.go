package eval

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/util"
	"github.com/xiaq/persistent/hash"
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
func (e ExternalCmd) Call(fm *Frame, argVals []interface{}, opts map[string]interface{}) error {
	if len(opts) > 0 {
		return ErrExternalCmdOpts
	}
	if util.DontSearch(e.Name) {
		stat, err := os.Stat(e.Name)
		if err == nil && stat.IsDir() {
			// implicit cd
			if len(argVals) > 0 {
				return ErrImplicitCdNoArg
			}
			return fm.Chdir(e.Name)
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
		// NOTE Maybe we should enfore string arguments instead of coercing all
		// args into string
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
		return err
	}
	return NewExternalCmdExit(e.Name, state.Sys().(syscall.WaitStatus), proc.Pid)
}

// EachExternal calls f for each name that can resolve to an external
// command.
// TODO(xiaq): Windows support
func EachExternal(f func(string)) {
	for _, dir := range searchPaths() {
		// XXX Ignore error
		infos, _ := ioutil.ReadDir(dir)
		for _, info := range infos {
			if !info.IsDir() && util.IsExecutableFile(info) {
				f(info.Name())
			}
		}
	}
}
