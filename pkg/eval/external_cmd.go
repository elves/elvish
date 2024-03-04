package eval

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/persistent/hash"
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

func (e externalCmd) Equal(a any) bool {
	return e == a
}

func (e externalCmd) Hash() uint32 {
	return hash.String(e.Name)
}

func (e externalCmd) Repr(int) string {
	return "<external " + parse.Quote(e.Name) + ">"
}

// Call calls an external command.
func (e externalCmd) Call(fm *Frame, argVals []any, opts map[string]any) error {
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
			fm.Deprecate("implicit cd is deprecated; use cd or location mode instead", fm.traceback.Head, 21)
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

	if runtime.GOOS == "windows" && !filepath.IsAbs(path) {
		// For some reason, Windows's CreateProcess API doesn't like forward
		// slashes in relative paths: ".\foo.bat" works but "./foo.bat" results
		// in an error message that "'.' is not recognized as an internal or
		// external command, operable program or batch file."
		//
		// There seems to be no good reason for this behavior, so we work around
		// it by replacing forward slashes with backslashes. PowerShell seems to
		// be something similar to support "./foo.bat".
		path = strings.ReplaceAll(path, "/", "\\")
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
	ws := state.Sys().(syscall.WaitStatus)
	if ws.Signaled() && isSIGPIPE(ws.Signal()) {
		readerGone := fm.ports[1].readerGone
		if readerGone != nil && readerGone.Load() {
			return errs.ReaderGone{}
		}
	}
	return NewExternalCmdExit(e.Name, state.Sys().(syscall.WaitStatus), proc.Pid)
}
