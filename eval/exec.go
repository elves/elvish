package eval

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/elves/elvish/util"
)

const (
	// FdNil is a special impossible fd value. Used for "close fd" in
	// syscall.ProcAttr.Files.
	FdNil uintptr = ^uintptr(0)
)

// A port conveys data stream. When f is not nil, it may convey fdStream. When
// ch is not nil, it may convey chanStream. When both are nil, it is always
// closed and may not convey any stream (unusedStream).
type port struct {
	f       *os.File
	ch      chan Value
	closeF  bool
	closeCh bool
}

// StreamType represents what form of data stream a command expects on each
// port.
type StreamType byte

// Possible values of StreamType.
const (
	unusedStream StreamType = 0
	fdStream                = 1 << iota // Corresponds to port.f.
	chanStream                          // Corresponds to port.ch.
	hybridStream = fdStream | chanStream
)

// closePorts closes the suitable components of all ports in ev.ports that were
// marked marked for closing.
func (ev *Evaluator) closePorts() {
	for _, port := range ev.ports {
		if port.closeF {
			port.f.Close()
		}
		if port.closeCh {
			close(port.ch)
		}
	}
}

// StateUpdate represents a change of state of a command.
type StateUpdate struct {
	Exited bool
	Exitus Exitus
	Update string
}

func newExitedStateUpdate(e Exitus) *StateUpdate {
	return &StateUpdate{Exited: true, Exitus: e}
}

func newUnexitedStateUpdate(u string) *StateUpdate {
	return &StateUpdate{Exited: false, Update: u}
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
func (ev *Evaluator) execSpecial(op exitusOp) <-chan *StateUpdate {
	update := make(chan *StateUpdate)
	go func() {
		ex := op(ev)
		// Ports are closed after executaion of builtin is complete.
		ev.closePorts()
		update <- newExitedStateUpdate(ex)
		close(update)
	}()
	return update
}

// execNonSpecial executes a form that is not a special form.
func (ev *Evaluator) execNonSpecial(cmd Value, args []Value) <-chan *StateUpdate {
	// Closure
	if closure, ok := cmd.(*Closure); ok {
		return ev.execClosure(closure, args)
	}

	cmdStr := cmd.String()

	// Defined function
	if v, ok := ev.scope["fn-"+cmdStr]; ok {
		if closure, ok := (*v.valuePtr).(*Closure); ok {
			return ev.execClosure(closure, args)
		}
	}

	// Builtin function
	if bi, ok := builtinFuncs[cmdStr]; ok {
		return ev.execBuiltinFunc(bi.fn, args)
	}

	// External command
	return ev.execExternal(cmdStr, args)
}

// execBuiltinFunc executes a builtin function.
func (ev *Evaluator) execBuiltinFunc(fn builtinFuncImpl, args []Value) <-chan *StateUpdate {
	update := make(chan *StateUpdate)
	go func() {
		ex := fn(ev, args)
		// Ports are closed after executaion of builtin is complete.
		ev.closePorts()
		update <- newExitedStateUpdate(ex)
		close(update)
	}()
	return update
}

// execClosure executes a closure form.
func (ev *Evaluator) execClosure(closure *Closure, args []Value) <-chan *StateUpdate {
	update := make(chan *StateUpdate, 1)

	// TODO Support optional/rest argument
	if len(args) != len(closure.ArgNames) {
		// TODO Check arity before exec'ing
		update <- newExitedStateUpdate(arityMismatch)
		close(update)
		return update
	}

	// Make a subevaluator.
	// BUG(xiaq): When evaluating closures, async access to global variables
	// and ports can be problematic.
	ev.scope = make(map[string]Variable)
	for name, variable := range closure.Captured {
		ev.scope[name] = variable
	}
	// Pass arguments.
	for i, name := range closure.ArgNames {
		// TODO(xiaq): support static type of arguments
		ev.scope[name] = newVariable(args[i], AnyType{})
	}

	ev.statusCb = nil

	go func() {
		// TODO(xiaq): Support calling closure originated in another source.
		err := ev.eval(ev.name, ev.text, closure.Op)
		if err != nil {
			fmt.Print(err.(*util.ContextualError).Pprint())
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
func waitStatusToStateUpdate(ws syscall.WaitStatus) *StateUpdate {
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
func waitStateUpdate(pid int, update chan<- *StateUpdate) {
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

// execExternal executes an external command form.
func (ev *Evaluator) execExternal(cmd string, argVals []Value) <-chan *StateUpdate {
	files := make([]uintptr, len(ev.ports))
	for i, port := range ev.ports {
		if port == nil || port.f == nil {
			files[i] = FdNil
		} else {
			files[i] = port.f.Fd()
		}
	}

	args := make([]string, len(argVals)+1)
	for i, a := range argVals {
		// NOTE Maybe we should enfore string arguments instead of coercing all
		// args into string
		args[i+1] = a.String()
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: ev.env.Export(), Files: files[:], Sys: &sys}

	path, err := ev.search(cmd)
	var pid int

	if err == nil {
		args[0] = path
		pid, err = syscall.ForkExec(path, args, &attr)
	}
	// Ports are closed after fork-exec of external is complete.
	ev.closePorts()

	update := make(chan *StateUpdate)
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
