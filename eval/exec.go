package eval

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/xiaq/elvish/parse"
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
	f  *os.File
	ch chan Value
}

// StreamType represents what form of data stream a command expects on each
// port.
type StreamType byte

// Possible values of StreamType.
const (
	fdStream   StreamType = iota // Default stream type. Corresponds to port.f.
	chanStream                   // Corresponds to port.ch.
	unusedStream
)

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
	Func    builtinFuncImpl    // A builtin function
	Special builtinSpecialImpl // A builtin special form
	Path    string             // External command full path
	Closure *Closure           // The closure value
}

// form packs runtime states of a fully constructured form.
type form struct {
	name  string              // Command name, used in error messages.
	nodes *parse.TermListNode // Unevaluated argument list, does not include command
	args  []Value             // Evaluated argument list
	ports [3]*port            // Ports for in, out and err.
	Command
}

func (fm *form) closePorts(ev *Evaluator) {
	for i, port := range fm.ports {
		if port == nil {
			continue
		}
		switch port.f {
		case nil, ev.in.f, ev.out.f, os.Stderr:
			// XXX(xiaq) Is the heuristics correct?
		default:
			port.f.Close()
		}
		if port.ch != nil {
			// Only close output channels
			if i == 1 {
				close(port.ch)
			}
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

func (ev *Evaluator) evalRedir(r parse.Redir, ports []*port, streamTypes []StreamType) {
	ev.push(r)
	defer ev.pop()

	fd := r.Fd()
	if fd > 2 {
		ev.errorf("redir on fd > 2 not yet supported")
	}

	switch r := r.(type) {
	case *parse.CloseRedir:
		ports[fd] = &port{}
	case *parse.FdRedir:
		if streamTypes[fd] == chanStream {
			ev.errorf("fd redir on channel port")
		}
		if r.OldFd > 2 {
			ev.errorf("fd redir from fd > 2 not yet supported")
		}
		ports[fd] = ports[r.OldFd]
	case *parse.FilenameRedir:
		if streamTypes[fd] == chanStream {
			ev.errorf("filename redir on channel port")
		}
		fname := string(*ev.evalTermSingleString(r.Filename, "filename"))
		// TODO haz hardcoded permbits now
		f, e := os.OpenFile(fname, r.Flag, 0644)
		if e != nil {
			ev.errorf("failed to open file %q: %s", fname[0], e)
		}
		ports[fd] = &port{f: f}
	}
}

// ResolveCommand tries to find a command with the given name the stream types
// it expects for three standard ports. If a command with that name doesn't
// exists, err is non-nil.
func (ev *Evaluator) ResolveCommand(name string) (cmd Command, streamTypes [3]StreamType, err error) {
	defer util.Recover(&err)
	cmd, streamTypes = ev.resolveCommand(name, nil)
	return cmd, streamTypes, nil
}

func (ev *Evaluator) resolveCommand(name string, n parse.Node) (cmd Command, streamTypes [3]StreamType) {
	if n != nil {
		ev.push(n)
		defer ev.pop()
	}

	// Try function
	if v, err := ev.ResolveVar("fn-" + name); err == nil {
		if fn, ok := v.(*Closure); ok {
			cmd.Closure = fn
			// BUG(xiaq): Functions are assumed to have zero streamTypes (fileStream)
			return
		}
	}

	// Try builtin special
	if bi, ok := builtinSpecials[name]; ok {
		cmd.Special = bi.fn
		copy(streamTypes[:2], bi.streamTypes[:])
		streamTypes[2] = fdStream
		return
	}

	// Try builtin func
	// XXX(xiaq): has duplicate code with builtin special
	if bi, ok := builtinFuncs[name]; ok {
		cmd.Func = bi.fn
		copy(streamTypes[:2], bi.streamTypes[:])
		streamTypes[2] = fdStream
		return
	}

	// Try external command
	path, e := ev.search(name)
	if e != nil {
		ev.errorf("%s", e)
	}
	cmd.Path = path
	// Use zero value (fileStream) for streamTypes
	return
}

func (ev *Evaluator) preevalForm(n *parse.FormNode) (fm *form, streamTypes [3]StreamType) {
	// Evaluate name.
	cmdValues := ev.evalTerm(n.Command)
	if len(cmdValues) != 1 {
		ev.errorfNode(n.Command, "command must be a single value")
	}
	cmd := cmdValues[0]

	// Start building form.
	cmdStr := cmd.String(ev)
	fm = &form{name: cmdStr}

	// Resolve command. Assign one of fm.Command.{fn path closure} and streamTypes.
	switch cmd := cmd.(type) {
	case *String:
		fm.Command, streamTypes = ev.resolveCommand(cmdStr, n.Command)
	case *Closure:
		fm.Command.Closure = cmd
		// BUG(xiaq): Closures are assumed to have zero streamTypes (fileStream)
	default:
		ev.errorfNode(n.Command, "Command must be either string or closure")
	}

	// Port list.
	defaultErrPort := &port{f: os.Stderr}
	// XXX(xiaq) Should we allow chanStream stderr at all?
	if defaultErrPort.compatible(streamTypes[2]) {
		fm.ports[2] = defaultErrPort
	}

	// Evaluate stream redirections.
	for _, r := range n.Redirs {
		ev.evalRedir(r, fm.ports[:], streamTypes[:])
	}

	// Evaluate arguments after everything else.
	if fm.Command.Special != nil {
		fm.nodes = n.Args
	} else {
		fm.args = ev.evalTermList(n.Args)
	}
	return
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
	newEv.locals = locals
	newEv.in = fm.ports[0]
	newEv.out = fm.ports[1]
	go func() {
		// TODO Support calling closure originated in another source.
		newEv.Eval(ev.name, ev.text, fm.Closure.Chunk)
		// Streams are closed after executaion of closure is complete.
		fm.closePorts(ev)
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
		var ports [2]*port
		copy(ports[:], fm.ports[:2])
		msg := fm.Special(ev, fm.nodes, ports)
		// Streams are closed after executaion of builtin is complete.
		fm.closePorts(ev)
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
		var ports [2]*port
		copy(ports[:], fm.ports[:2])
		msg := fm.Func(ev, fm.args, ports)
		// Streams are closed after executaion of builtin is complete.
		fm.closePorts(ev)
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
	files := make([]uintptr, len(fm.ports))
	for i, port := range fm.ports {
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
		args[i+1] = a.String(ev)
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: ev.env.Export(), Files: files[:], Sys: &sys}
	pid, err := syscall.ForkExec(fm.Path, args, &attr)
	// Streams are closed after fork-exec of external is complete.
	fm.closePorts(ev)

	update := make(chan *StateUpdate)
	if err != nil {
		update <- &StateUpdate{Terminated: true, Msg: err.Error()}
		close(update)
	} else {
		go waitStateUpdate(pid, update)
	}

	return update
}

// preevalPipeline resolves commands, sets up pipes and applies redirections.
// These are done before commands are actually executed. Input of first form
// and output of last form are NOT connected to ev.{in out}.
func (ev *Evaluator) preevalPipeline(pl *parse.ListNode) (fms []*form, pipelineStreamTypes [3]StreamType) {
	ev.push(pl)
	defer ev.pop()

	nfms := len(pl.Nodes)
	if nfms == 0 {
		return
	}
	fms = make([]*form, 0, nfms)

	var nextIn *port
	for i, n := range pl.Nodes {
		fm, streamTypes := ev.preevalForm(n.(*parse.FormNode))

		var prependCmd, appendCmd *form

		// Connect input pipe.
		if i == 0 {
			pipelineStreamTypes[0] = streamTypes[0]
		} else {
			if fm.ports[0] != nil {
				ev.errorf("form #%d has both pipe input and input redirection")
			} else if !nextIn.compatible(streamTypes[0]) {
				ev.errorf("form #%d has incompatible input pipe")
			}
			fm.ports[0] = nextIn
		}
		// Create and connect output pipe.
		if i == nfms-1 {
			pipelineStreamTypes[1] = streamTypes[1]
		} else {
			if fm.ports[1] != nil {
				ev.errorf("form #%d has both pipe output and output redirection", i)
			}
			switch streamTypes[1] {
			case unusedStream:
				ev.errorf("form #%d has unused output connected in pipeline", i)
			case fdStream:
				// os.Pipe sets O_CLOEXEC, which is what we want.
				reader, writer, e := os.Pipe()
				if e != nil {
					ev.errorf("failed to create pipe: %s", e)
				}
				nextIn = &port{f: reader}
				fm.ports[1] = &port{f: writer}
			case chanStream:
				// TODO Buffered channel?
				ch := make(chan Value)
				nextIn = &port{ch: ch}
				fm.ports[1] = &port{ch: ch}
			default:
				panic("unreachable")
			}
		}

		if prependCmd != nil {
			fms = append(fms, prependCmd)
		}
		fms = append(fms, fm)
		if appendCmd != nil {
			fms = append(fms, appendCmd)
		}
	}
	return
}

func makePrintchan(in chan Value, out *os.File) *form {
	return &form{
		name: "(implicit printchan)",
		args: nil,
		ports: [3]*port{
			&port{ch: in},
			&port{f: out},
			&port{f: os.Stderr},
		},
		Command: Command{Func: printchan},
	}
}

func makeFeedchan(in *os.File, out chan Value) *form {
	return &form{
		name: "(implicit feedchan)",
		args: nil,
		ports: [3]*port{
			&port{f: in},
			&port{ch: out},
			&port{f: os.Stderr},
		},
		Command: Command{Func: feedchan},
	}
}

// execPipeline executes a pipeline set up by preevalPipeline.
// It fullfils the input of fms[0] and output of fms[len(fms)-1], inserting
// adaptors if needed.
//
// TODO Should return a slice of exit statuses.
func (ev *Evaluator) execPipeline(fms []*form, types [3]StreamType) []<-chan *StateUpdate {
	var implicits [2]*form

	// Pipeline input.
	if fms[0].ports[0] == nil {
		if ev.in.compatible(types[0]) {
			fms[0].ports[0] = ev.in
		} else {
			// Prepend an adapter.
			// XXX(xiaq): This now assumes at least one of ev.in.{f ch} is not nil
			switch types[0] {
			case fdStream:
				// chan -> fd adapter: printchan
				reader, writer, e := os.Pipe()
				if e != nil {
					ev.errorf("failed to create pipe: %s", e)
				}
				implicits[0] = makePrintchan(ev.in.ch, writer)
				fms[0].ports[0] = &port{f: reader}
			case chanStream:
				// fd -> chan adapter: feedchan
				ch := make(chan Value)
				implicits[0] = makeFeedchan(ev.in.f, ch)
				fms[0].ports[0] = &port{ch: ch}
			default:
				panic("unreachable")
			}
		}
	}

	// Pipeline output
	if fms[len(fms)-1].ports[1] == nil {
		if ev.out.compatible(types[1]) {
			fms[len(fms)-1].ports[1] = ev.out
		} else {
			// Append an adapter.
			// XXX(xiaq): This now assumes at least one of ev.out.{f ch} is not nil
			switch types[1] {
			case fdStream:
				// fd -> chan adapter: feedchan
				reader, writer, e := os.Pipe()
				if e != nil {
					ev.errorf("failed to create pipe: %s", e)
				}
				implicits[1] = makeFeedchan(reader, ev.out.ch)
				fms[len(fms)-1].ports[1] = &port{f: writer}
			case chanStream:
				// chan -> fd adapter: printchan
				ch := make(chan Value)
				implicits[1] = makePrintchan(ch, ev.out.f)
				fms[len(fms)-1].ports[1] = &port{ch: ch}
			default:
				panic("unreachable")
			}
		}
	}

	for _, imp := range implicits {
		if imp != nil {
			fms = append(fms, imp)
		}
	}

	updates := make([]<-chan *StateUpdate, len(fms))
	for i, fm := range fms {
		updates[i] = ev.execForm(fm)
	}
	return updates
}

func (ev *Evaluator) waitPipeline(updates []<-chan *StateUpdate) {
	for i, update := range updates {
		for up := range update {
			switch up.Msg {
			case "0", "":
			default:
				// BUG(xiaq): Command update of commands in subevaluators are
				// always printed.
				fmt.Printf("Command #%d update: %s\n", i, up.Msg)
			}
		}
	}
}

func (ev *Evaluator) evalPipelineAsync(pl *parse.ListNode) []<-chan *StateUpdate {
	fms, types := ev.preevalPipeline(pl)
	return ev.execPipeline(fms, types)
}

// evalPipeline combines preevalPipeline and execPipeline.
func (ev *Evaluator) evalPipeline(pl *parse.ListNode) {
	ev.waitPipeline(ev.evalPipelineAsync(pl))
}
