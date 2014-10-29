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
	f           *os.File
	ch          chan Value
	shouldClose bool
}

// StreamType represents what form of data stream a command expects on each
// port.
type StreamType byte

// Possible values of StreamType.
const (
	unusedStream StreamType = iota
	fdStream                // Corresponds to port.f.
	chanStream              // Corresponds to port.ch.
)

// commonType returns a StreamType compatible with both StreamType's. A
// StreamType is compatible with itself or unusedStream.
func (typ StreamType) commonType(typ2 StreamType) (StreamType, bool) {
	switch {
	case typ == unusedStream, typ == typ2:
		return typ2, true
	case typ2 == unusedStream:
		return typ, true
	default:
		return 0, false
	}
}

// mayConvey returns whether port may convey a stream of typ.
func (i *port) mayConvey(typ StreamType) bool {
	switch typ {
	case fdStream:
		return i != nil && i.f != nil
	case chanStream:
		return i != nil && i.ch != nil
	default: // Actually case unusedStream:
		return true
	}
}

// Command is exactly one of a builtin function, a builtin special form, an
// external command or a closure.
type Command struct {
	Func    builtinFuncImpl // A builtin function
	Special strOp           // A builtin special form
	Path    string          // External command full path
	Closure *Closure        // The closure value
}

// form packs runtime states of a fully constructured form.
type form struct {
	name string  // Command name, used in error messages.
	args []Value // Evaluated argument list
	Command
}

// closePorts closes all ports in ev.ports that were marked shouldClose.
func (ev *Evaluator) closePorts() {
	for _, port := range ev.ports {
		if !port.shouldClose {
			continue
		}
		if port.f != nil {
			port.f.Close()
		}
		if port.ch != nil {
			close(port.ch)
		}
	}
}

// StateUpdate represents a change of state of a command.
type StateUpdate struct {
	Terminated bool
	Msg        string
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

// execForm executes a form.
//
// NOTE(xiaq): execForm is always called on an intermediate "form redir" where
// only the form-local ports are marked shouldClose. ev.closePorts should be
// called at appropriate moments.
func (ev *Evaluator) execForm(fm *form) <-chan *StateUpdate {
	switch {
	case fm.Func != nil, fm.Special != nil:
		return ev.execBuiltin(fm)
	case fm.Path != "":
		return ev.execExternal(fm)
	case fm.Closure != nil:
		return ev.execClosure(fm)
	default:
		panic("Bad eval.form struct")
	}
}

// execClosure executes a closure form.
func (ev *Evaluator) execClosure(fm *form) <-chan *StateUpdate {
	update := make(chan *StateUpdate, 1)

	locals := make(map[string]Value)
	// TODO Support optional/rest argument
	if len(fm.args) != len(fm.Closure.ArgNames) {
		// TODO Check arity before exec'ing
		update <- &StateUpdate{Terminated: true, Msg: "arity mismatch"}
		close(update)
		return update
	}
	// Pass argument by populating locals.
	for i, name := range fm.Closure.ArgNames {
		locals[name] = fm.args[i]
	}

	// Make a subevaluator.
	// BUG(xiaq): When evaluating closures, async access to globals, in and out can be problematic.
	ev.scope = make(map[string]*Value)
	for name, pvalue := range fm.Closure.Enclosed {
		ev.scope[name] = pvalue
	}
	ev.statusCb = nil
	go func() {
		// TODO Support calling closure originated in another source.
		err := ev.eval(ev.name, ev.text, fm.Closure.Op)
		if err != nil {
			fmt.Print(err.(*util.ContextualError).Pprint())
		}
		// Ports are closed after executaion of closure is complete.
		ev.closePorts()
		// TODO Support returning value.
		update <- &StateUpdate{Terminated: true}
		close(update)
	}()
	return update
}

// execBuiltin executes a builtin special form or builtin function.
func (ev *Evaluator) execBuiltin(fm *form) <-chan *StateUpdate {
	update := make(chan *StateUpdate)
	go func() {
		var msg string
		if fm.Special != nil {
			msg = fm.Special(ev)
		} else {
			msg = fm.Func(ev, fm.args)
		}
		// Ports are closed after executaion of builtin is complete.
		ev.closePorts()
		update <- &StateUpdate{Terminated: true, Msg: msg}
		close(update)
	}()
	return update
}

// sprintStatus returns a human-readable representation of a
// syscall.WaitStatus.
func sprintStatus(ws syscall.WaitStatus) string {
	switch {
	case ws.Exited():
		es := ws.ExitStatus()
		if es == 0 {
			return ""
		}
		return fmt.Sprintf("exited %v", es)
	case ws.Signaled():
		msg := fmt.Sprintf("signaled %v", ws.Signal())
		if ws.CoreDump() {
			msg += " (core dumped)"
		}
		return msg
	case ws.Stopped():
		msg := fmt.Sprintf("stopped %v", ws.StopSignal())
		trap := ws.TrapCause()
		if trap != -1 {
			msg += fmt.Sprintf(" (trapped %v)", trap)
		}
		return msg
	case ws.Continued():
		return "continued"
	default:
		return fmt.Sprintf("unknown status %v", ws)
	}
}

// waitStateUpdate wait(2)s for pid, and feeds the WaitStatus's into a
// *StateUpdate channel after proper conversion.
func waitStateUpdate(pid int, update chan<- *StateUpdate) {
	for {
		var ws syscall.WaitStatus
		_, err := syscall.Wait4(pid, &ws, 0, nil)

		if err != nil {
			if err != syscall.ECHILD {
				update <- &StateUpdate{Msg: err.Error()}
			}
			break
		}
		update <- &StateUpdate{
			Terminated: ws.Exited(), Msg: sprintStatus(ws)}
	}
	close(update)
}

// execExternal executes an external command form.
func (ev *Evaluator) execExternal(fm *form) <-chan *StateUpdate {
	files := make([]uintptr, len(ev.ports))
	for i, port := range ev.ports {
		if port == nil || port.f == nil {
			files[i] = FdNil
		} else {
			files[i] = port.f.Fd()
		}
	}

	args := make([]string, len(fm.args)+1)
	args[0] = fm.Path
	for i, a := range fm.args {
		// NOTE Maybe we should enfore string arguments instead of coercing all
		// args into string
		args[i+1] = a.String()
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: ev.env.Export(), Files: files[:], Sys: &sys}
	pid, err := syscall.ForkExec(fm.Path, args, &attr)
	// Ports are closed after fork-exec of external is complete.
	ev.closePorts()

	update := make(chan *StateUpdate)
	if err != nil {
		go func() {
			update <- &StateUpdate{Terminated: true, Msg: err.Error()}
			close(update)
		}()
	} else {
		go waitStateUpdate(pid, update)
	}

	return update
}
