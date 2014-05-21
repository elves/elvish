package eval

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/xiaq/elvish/util"
)

const (
	// FdNil is a special impossible fd value. Used for "close fd" in
	// syscall.ProcAttr.Files.
	FdNil uintptr = ^uintptr(0)
)

// A port conveys data stream. It may be a Unix fd (wrapped by os.File), where
// f is not nil, or a channel, where ch is not nil. When both are nil, the port
// is closed and may not be used.
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

func (i *port) compatible(typ StreamType) bool {
	switch typ {
	case fdStream:
		return i != nil && i.f != nil
	case chanStream:
		return i != nil && i.ch != nil
	default: // Actually case unusedStream:
		return true
	}
}

// A Command is either a builtin function, a builtin special form, an external
// command or a closure.
type Command struct {
	Func    builtinFuncImpl // A builtin function
	Special strOp           // A builtin special form
	Path    string          // External command full path
	Closure *Closure        // The closure value
}

// form packs runtime states of a fully constructured form.
type form struct {
	name       string  // Command name, used in error messages.
	args       []Value // Evaluated argument list
	annotation *formAnnotation
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

// Search for executable `exe`.
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

// execCommand executes a command.
func (ev *Evaluator) execForm(fm *form) <-chan *StateUpdate {
	switch {
	case fm.Func != nil:
		return ev.execBuiltinFunc(fm)
	case fm.Special != nil:
		return ev.execBuiltinSpecial(fm)
	case fm.Path != "":
		return ev.execExternal(fm)
	case fm.Closure != nil:
		return ev.execClosure(fm)
	default:
		panic("Bad eval.form struct")
	}
}

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
	newEv := ev.copy()
	newEv.scope = make(map[string]*Value)
	for name, pvalue := range fm.Closure.Enclosed {
		newEv.scope[name] = pvalue
	}
	newEv.statusCb = nil
	go func() {
		// TODO Support calling closure originated in another source.
		err := newEv.eval(ev.name, ev.text, fm.Closure.Op)
		if err != nil {
			fmt.Print(err.(*util.ContextualError).Pprint())
		}
		// Ports are closed after executaion of closure is complete.
		newEv.closePorts()
		// TODO Support returning value.
		update <- &StateUpdate{Terminated: true}
		close(update)
	}()
	return update
}

// execBuiltinSpecial executes a builtin special form.
func (ev *Evaluator) execBuiltinSpecial(fm *form) <-chan *StateUpdate {
	update := make(chan *StateUpdate)
	go func() {
		msg := fm.Special(ev)
		// Ports are closed after executaion of builtin is complete.
		ev.closePorts()
		update <- &StateUpdate{Terminated: true, Msg: msg}
		close(update)
	}()
	return update
}

// execBuiltinFunc executes a builtin function.
// XXX(xiaq): Duplicate with execBuiltinSpecial.
func (ev *Evaluator) execBuiltinFunc(fm *form) <-chan *StateUpdate {
	update := make(chan *StateUpdate)
	go func() {
		msg := fm.Func(ev, fm.args)
		// Ports are closed after executaion of builtin is complete.
		ev.closePorts()
		update <- &StateUpdate{Terminated: true, Msg: msg}
		close(update)
	}()
	return update
}

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
			Terminated: ws.Exited(), Msg: fmt.Sprintf("%v", ws)}
	}
	close(update)
}

// execExternal executes an external command.
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
