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
func (ev *Evaluator) search(exe string) (string, error) {
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
// shouldClose. ev.closePorts should be called at appropriate moments.
func (ev *Evaluator) execSpecial(op exitusOp) <-chan *stateUpdate {
	update := make(chan *stateUpdate)
	go func() {
		ex := op(ev)
		// Ports are closed after executaion of builtin is complete.
		ev.closePorts()
		update <- newExitedStateUpdate(ex)
		close(update)
	}()
	return update
}

func (ev *Evaluator) resolveNonSpecial(cmd Value) callable {
	// Closure
	if cl, ok := cmd.(*closure); ok {
		return cl
	}

	cmdStr := toString(cmd)

	// Defined callable
	ns, name := splitQualifiedName(cmdStr)
	if v := ev.ResolveVar(ns, "fn-"+name); v != nil {
		if clb, ok := v.Get().(callable); ok {
			return clb
		}
	}

	// External command
	return externalCmd{cmdStr}
}

// execNonSpecial executes a form that is not a special form.
func (ev *Evaluator) execNonSpecial(cmd Value, args []Value) <-chan *stateUpdate {
	return ev.resolveNonSpecial(cmd).Exec(ev, args)
}

// Exec executes a builtin function.
func (b *builtinFn) Exec(ev *Evaluator, args []Value) <-chan *stateUpdate {
	update := make(chan *stateUpdate)
	go func() {
		ex := b.Impl(ev, args)
		// Ports are closed after executaion of builtin is complete.
		ev.closePorts()
		update <- newExitedStateUpdate(ex)
		close(update)
	}()
	return update
}

// Exec executes a closure.
func (c *closure) Exec(ev *Evaluator, args []Value) <-chan *stateUpdate {
	update := make(chan *stateUpdate, 1)

	// TODO Support optional/rest argument
	if len(args) != len(c.ArgNames) {
		// TODO Check arity before exec'ing
		update <- newExitedStateUpdate(arityMismatch)
		close(update)
		return update
	}

	// Make a subevaluator.
	// BUG(xiaq): When evaluating closures, async access to global variables
	// and ports can be problematic.

	// Make captured namespace and capture variables.
	ev.captured = make(map[string]Variable)
	for name, variable := range c.Captured {
		ev.captured[name] = variable
	}
	// Make local namespace and pass arguments.
	ev.local = make(map[string]Variable)
	for i, name := range c.ArgNames {
		// TODO(xiaq): support static type of arguments
		ev.local[name] = newInternalVariable(args[i], anyType{})
	}

	ev.statusCb = nil

	go func() {
		// TODO(xiaq): Support calling closure originated in another source.
		err := ev.eval(ev.name, ev.text, c.Op)
		if err != nil {
			fmt.Print(err.(*errutil.ContextualError).Pprint())
		}
		// Ports are closed after executaion of closure is complete.
		ev.closePorts()
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
func (e externalCmd) Exec(ev *Evaluator, argVals []Value) <-chan *stateUpdate {
	files := make([]uintptr, len(ev.ports))
	for i, port := range ev.ports {
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

	path, err := ev.search(e.Name)
	var pid int

	if err == nil {
		args[0] = path
		pid, err = syscall.ForkExec(path, args, &attr)
	}
	// Ports are closed after fork-exec of external is complete.
	ev.closePorts()

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
