package eval

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/elves/elvish/errutil"
)

const (
	// FdNil is a special impossible fd value. Used for "close fd" in
	// syscall.ProcAttr.Files.
	fdNil uintptr = ^uintptr(0)
)

// StateUpdate represents a change of state of a command.
type stateUpdate struct {
	Exited bool
	Exitus exitus
	Update string
}

func newExitedStateUpdate(e exitus) *stateUpdate {
	return &stateUpdate{Exited: true, Exitus: e}
}

func newUnexitedStateUpdate(u string) *stateUpdate {
	return &stateUpdate{Exited: false, Update: u}
}

// isExecutable determines whether path refers to an executable file.
func isExecutable(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false
	}
	fm := fi.Mode()
	return !fm.IsDir() && (fm&0111 != 0)
}

// search tries to resolve an external command and return the full (possibly
// relative) path.
func (ev *Evaler) search(exe string) (string, error) {
	for _, p := range []string{"/", "./", "../"} {
		if strings.HasPrefix(exe, p) {
			if isExecutable(exe) {
				return exe, nil
			}
			return "", fmt.Errorf("external command not executable")
		}
	}
	for _, p := range ev.searchPaths {
		full := p + "/" + exe
		if isExecutable(full) {
			return full, nil
		}
	}
	return "", fmt.Errorf("external command not found")
}

// execSpecial executes a builtin special form.
//
// NOTE(xiaq): execSpecial and execNonSpecial are always called on an
// intermediate "form redir" where only the form-local ports are marked
// shouldClose. ec.closePorts should be called at appropriate moments.
func (ec *evalCtx) execSpecial(op exitusOp) <-chan *stateUpdate {
	update := make(chan *stateUpdate)
	go func() {
		ex := op(ec)
		// Ports are closed after executaion of builtin is complete.
		ec.closePorts()
		update <- newExitedStateUpdate(ex)
		close(update)
	}()
	return update
}

func (ec *evalCtx) resolveNonSpecial(cmd Value) callable {
	// Closure
	if cl, ok := cmd.(*closure); ok {
		return cl
	}

	cmdStr := toString(cmd)

	// Defined callable
	ns, name := splitQualifiedName(cmdStr)
	if v := ec.ResolveVar(ns, fnPrefix+name); v != nil {
		if clb, ok := v.Get().(callable); ok {
			return clb
		}
	}

	// External command
	return externalCmd{cmdStr}
}

// execNonSpecial executes a form that is not a special form.
func (ec *evalCtx) execNonSpecial(cmd Value, args []Value) <-chan *stateUpdate {
	return ec.resolveNonSpecial(cmd).Exec(ec, args)
}

// Exec executes a builtin function.
func (b *builtinFn) Exec(ec *evalCtx, args []Value) <-chan *stateUpdate {
	update := make(chan *stateUpdate)
	go func() {
		ex := b.Impl(ec, args)
		// Ports are closed after executaion of builtin is complete.
		ec.closePorts()
		update <- newExitedStateUpdate(ex)
		close(update)
	}()
	return update
}

// Exec executes a closure.
func (c *closure) Exec(ec *evalCtx, args []Value) <-chan *stateUpdate {
	update := make(chan *stateUpdate, 1)

	// TODO Support optional/rest argument
	if len(args) != len(c.ArgNames) {
		// TODO Check arity before exec'ing
		update <- newExitedStateUpdate(arityMismatch)
		close(update)
		return update
	}

	// Make a subecaler.
	// BUG(xiaq): When ecaluating closures, async access to global variables
	// and ports can be problematic.

	// Make up namespace and capture variables.
	ec.up = make(map[string]Variable)
	for name, variable := range c.Captured {
		ec.up[name] = variable
	}
	// Make local namespace and pass arguments.
	ec.local = make(map[string]Variable)
	for i, name := range c.ArgNames {
		// TODO(xiaq): support static type of arguments
		ec.local[name] = newInternalVariable(args[i], anyType{})
	}

	// TODO(xiaq): The failure handler should let the whole closure fail.
	ec.failHandler = nil

	go func() {
		// TODO(xiaq): Support calling closure originated in another source.
		err := ec.eval(c.Op)
		if err != nil {
			fmt.Print(err.(*errutil.ContextualError).Pprint())
		}
		// Ports are closed after executaion of closure is complete.
		ec.closePorts()
		// TODO Support returning value.
		update <- newExitedStateUpdate(success)
		close(update)
	}()
	return update
}

// waitStatusToStateUpdate converts syscall.WaitStatus to a StateUpdate.
func waitStatusToStateUpdate(ws syscall.WaitStatus) *stateUpdate {
	switch {
	case ws.Exited():
		es := ws.ExitStatus()
		if es == 0 {
			return newExitedStateUpdate(success)
		}
		return newExitedStateUpdate(newFailure(fmt.Sprint(es)))
	case ws.Signaled():
		msg := fmt.Sprintf("signaled %v", ws.Signal())
		if ws.CoreDump() {
			msg += " (core dumped)"
		}
		return newUnexitedStateUpdate(msg)
	case ws.Stopped():
		msg := fmt.Sprintf("stopped %v", ws.StopSignal())
		trap := ws.TrapCause()
		if trap != -1 {
			msg += fmt.Sprintf(" (trapped %v)", trap)
		}
		return newUnexitedStateUpdate(msg)
	case ws.Continued():
		return newUnexitedStateUpdate("continued")
	default:
		return newUnexitedStateUpdate(fmt.Sprint("unknown status", ws))
	}
}

// waitStateUpdate wait(2)s for pid, and feeds the WaitStatus's into a
// *StateUpdate channel after proper conversion.
func waitStateUpdate(pid int, update chan<- *stateUpdate) {
	for {
		var ws syscall.WaitStatus
		_, err := syscall.Wait4(pid, &ws, 0, nil)

		if err != nil {
			update <- newExitedStateUpdate(newFailure(err.Error()))
			break
		}
		update <- waitStatusToStateUpdate(ws)
		if ws.Exited() {
			break
		}
	}
	close(update)
}

// Exec executes an external command.
func (e externalCmd) Exec(ec *evalCtx, argVals []Value) <-chan *stateUpdate {
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

	path, err := ec.search(e.Name)
	var pid int

	if err == nil {
		args[0] = path
		pid, err = syscall.ForkExec(path, args, &attr)
	}
	// Ports are closed after fork-exec of external is complete.
	ec.closePorts()

	update := make(chan *stateUpdate)
	if err != nil {
		go func() {
			update <- newExitedStateUpdate(newFailure(err.Error()))
			close(update)
		}()
	} else {
		go waitStateUpdate(pid, update)
	}

	return update
}
