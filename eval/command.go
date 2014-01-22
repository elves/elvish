package eval

import (
	"../parse"
	"../util"
	"fmt"
	"os"
	"strings"
	"syscall"
)

const (
	// A special impossible fd value. Used for "close fd" in
	// syscall.ProcAttr.Files.
	FD_NIL uintptr = ^uintptr(0)
)

// A port conveys data stream. It may be a Unix fd (wrapped by os.File), where
// f is not nil, or a channel, where ch is not nil. When both are nil, the port
// is closed and may not be used.
type port struct {
	f  *os.File
	ch chan Value
}

type StreamType byte

const (
	fdStream   StreamType = iota // Default stream type. Corresponds to port.f.
	chanStream                   // Corresponds to port.ch.
	unusedStream
)

func (i *port) compatible(typ StreamType) bool {
	if i == nil {
		return false
	}
	if typ == unusedStream {
		return true
	}
	switch {
	case i.f != nil:
		return typ == fdStream
	case i.ch != nil:
		return typ == chanStream
	default:
		return true
	}
}

// The "head" of a command is either a function, the path of an external
// command or a closure.
type CommandHead struct {
	Func    BuiltinFunc // A builtin function, if the command is builtin.
	Path    string      // Command full path, if the command is external.
	Closure *Closure    // The closure value, if the command is a closure.
}

// command packs runtime states of a fully constructured command.
type command struct {
	name  string   // Command name, used in error messages.
	args  []Value  // Argument list, minus command name.
	ports [3]*port // Ports for in, out and err.
	CommandHead
}

func (cmd *command) closePorts(ev *Evaluator) {
	for i, port := range cmd.ports {
		if port == nil {
			continue
		}
		switch port.f {
		case nil, ev.in.f, ev.out.f, os.Stderr:
			// XXX Is the heuristics correct?
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
		fname := ev.evalTermSingleScalar(r.Filename, "filename").str
		// TODO haz hardcoded permbits now
		f, e := os.OpenFile(fname, r.Flag, 0644)
		if e != nil {
			ev.errorf("failed to open file %q: %s", fname[0], e)
		}
		ports[fd] = &port{f: f}
	}
}

func (ev *Evaluator) ResolveCommand(name string) (head CommandHead, streamTypes [3]StreamType, err error) {
	defer util.Recover(&err)
	head, streamTypes = ev.resolveCommand(name, nil)
	return head, streamTypes, nil
}

func (ev *Evaluator) resolveCommand(name string, n parse.Node) (head CommandHead, streamTypes [3]StreamType) {
	if n != nil {
		ev.push(n)
		defer ev.pop()
	}

	// Try function
	if v, err := ev.ResolveVar("fn-" + name); err == nil {
		if fn, ok := v.(*Closure); ok {
			head.Closure = fn
			// XXX Use zero value (fileStream) for streamTypes now
			return
		}
	}

	// Try builtin
	if bi, ok := builtins[name]; ok {
		head.Func = bi.fn
		copy(streamTypes[:2], bi.streamTypes[:])
		streamTypes[2] = fdStream
		return
	}

	// Try external command
	path, e := ev.search(name)
	if e != nil {
		ev.errorf("%s", e)
	}
	head.Path = path
	// Use zero value (fileStream) for streamTypes
	return
}

func (ev *Evaluator) preevalCommand(n *parse.CommandNode) (cmd *command, streamTypes [3]StreamType) {
	// Evaluate name.
	nameValues := ev.evalTerm(n.Name)
	if len(nameValues) != 1 {
		ev.errorfNode(n.Name, "command name must be a single value")
	}
	name := nameValues[0]

	// Start building command.
	nameStr := name.String(ev)
	cmd = &command{name: nameStr}

	// Resolve command. Assign one of cmd.{fn path closure} and streamTypes.
	switch name := name.(type) {
	case *Scalar:
		cmd.CommandHead, streamTypes = ev.resolveCommand(nameStr, n.Name)
	case *Closure:
		cmd.CommandHead.Closure = name
		// XXX Use zero value (fileStream) for streamTypes now
	default:
		ev.errorfNode(n.Name, "Command name must be either scalar or closure")
	}

	// Port list.
	defaultErrPort := &port{f: os.Stderr}
	// XXX Should we allow chanStream stderr at all?
	if defaultErrPort.compatible(streamTypes[2]) {
		cmd.ports[2] = defaultErrPort
	}

	// Evaluate stream redirections.
	for _, r := range n.Redirs {
		ev.evalRedir(r, cmd.ports[:], streamTypes[:])
	}

	// Evaluate arguments after everything else.
	cmd.args = ev.evalTermList(n.Args)
	return
}

// execCommand executes a command.
func (ev *Evaluator) execCommand(cmd *command) <-chan *StateUpdate {
	switch {
	case cmd.Func != nil:
		return ev.execBuiltin(cmd)
	case cmd.Path != "":
		return ev.execExternal(cmd)
	case cmd.Closure != nil:
		return ev.execClosure(cmd)
	default:
		panic("Bad eval.command struct")
	}
}

func (ev *Evaluator) execClosure(cmd *command) <-chan *StateUpdate {
	update := make(chan *StateUpdate, 1)

	locals := make(map[string]Value)
	// TODO Support optional/rest argument
	if len(cmd.args) != len(cmd.Closure.ArgNames) {
		// TODO Check arity before exec'ing
		update <- &StateUpdate{Terminated: true, Msg: "arity mismatch"}
		close(update)
		return update
	}
	// Pass argument by populating locals.
	for i, name := range cmd.Closure.ArgNames {
		locals[name] = cmd.args[i]
	}

	// Make a subevaluator.
	// XXX Concurrent access to globals, in and out can be problematic.
	newEv := ev.copy()
	newEv.locals = locals
	go func() {
		// TODO Support calling closure originated in another source.
		newEv.Eval(ev.name, ev.text, cmd.Closure.Chunk)
		// Streams are closed after executaion of closure is complete.
		cmd.closePorts(ev)
		// TODO Support returning value.
		update <- &StateUpdate{Terminated: true}
		close(update)
	}()
	return update
}

// execBuiltin executes a builtin command.
func (ev *Evaluator) execBuiltin(cmd *command) <-chan *StateUpdate {
	update := make(chan *StateUpdate)
	go func() {
		var ports [2]*port
		copy(ports[:], cmd.ports[:2])
		msg := cmd.Func(ev, cmd.args, ports)
		// Streams are closed after executaion of builtin is complete.
		cmd.closePorts(ev)
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
func (ev *Evaluator) execExternal(cmd *command) <-chan *StateUpdate {
	files := make([]uintptr, len(cmd.ports))
	for i, port := range cmd.ports {
		if port == nil || port.f == nil {
			files[i] = FD_NIL
		} else {
			files[i] = port.f.Fd()
		}
	}

	args := make([]string, len(cmd.args)+1)
	args[0] = cmd.Path
	for i, a := range cmd.args {
		// NOTE Maybe we should enfore scalar arguments instead of coercing all
		// args into string
		args[i+1] = a.String(ev)
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: ev.env.Export(), Files: files[:], Sys: &sys}
	pid, err := syscall.ForkExec(cmd.Path, args, &attr)
	// Streams are closed after fork-exec of external is complete.
	cmd.closePorts(ev)

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
// These are done before commands are actually executed.
func (ev *Evaluator) preevalPipeline(pl *parse.ListNode) (cmds []*command, pipelineStreamTypes [3]StreamType) {
	ev.push(pl)
	defer ev.pop()

	ncmds := len(pl.Nodes)
	if ncmds == 0 {
		return
	}
	cmds = make([]*command, 0, ncmds)

	var nextIn *port
	for i, n := range pl.Nodes {
		cmd, streamTypes := ev.preevalCommand(n.(*parse.CommandNode))

		var prependCmd, appendCmd *command

		// Create and connect pipes.
		if i == 0 {
			// First command. Only connect input when no input redirection is
			// present.
			pipelineStreamTypes[0] = streamTypes[0]
			if cmd.ports[0] == nil {
				if streamTypes[0] == chanStream {
					// Prepend an implicit feedchan
					ch := make(chan Value)
					prependCmd = &command{
						name: "(implicit feedchan)",
						args: nil,
						ports: [3]*port{
							ev.in,
							&port{ch: ch},
							&port{f: os.Stderr},
						},
						CommandHead: CommandHead{Func: feedchan},
					}
					cmd.ports[0] = &port{ch: ch}
				} else {
					cmd.ports[0] = ev.in
				}
			}
		} else {
			if cmd.ports[0] != nil {
				ev.errorf("command #%d has both pipe input and input redirection")
			} else if !nextIn.compatible(streamTypes[0]) {
				ev.errorf("command #%d has incompatible input pipe")
			}
			cmd.ports[0] = nextIn
		}
		if i == ncmds-1 {
			pipelineStreamTypes[1] = streamTypes[1]
			if cmd.ports[1] == nil {
				if streamTypes[1] == chanStream {
					// Append an implicit printchan
					ch := make(chan Value)
					cmd.ports[1] = &port{ch: ch}
					appendCmd = &command{
						name: "(implicit printchan)",
						args: nil,
						ports: [3]*port{
							&port{ch: ch},
							ev.out,
							&port{f: os.Stderr},
						},
						CommandHead: CommandHead{Func: printchan},
					}
				} else {
					cmd.ports[1] = ev.out
				}
			}
		} else {
			if cmd.ports[1] != nil {
				ev.errorf("command #%d has both pipe output and output redirection", i)
			}
			switch streamTypes[1] {
			case unusedStream:
				ev.errorf("command #%d has unused output connected in pipeline", i)
			case fdStream:
				// os.Pipe sets O_CLOEXEC, which is what we want.
				reader, writer, e := os.Pipe()
				if e != nil {
					ev.errorf("failed to create pipe: %s", e)
				}
				nextIn = &port{f: reader}
				cmd.ports[1] = &port{f: writer}
			case chanStream:
				// TODO Buffered channel?
				ch := make(chan Value)
				nextIn = &port{ch: ch}
				cmd.ports[1] = &port{ch: ch}
			default:
				panic("unreachable")
			}
		}

		if prependCmd != nil {
			cmds = append(cmds, prependCmd)
		}
		cmds = append(cmds, cmd)
		if appendCmd != nil {
			cmds = append(cmds, appendCmd)
		}
	}
	return
}

// execPipeline executes a pipeline set up by preevalPipeline.
//
// TODO Should return a slice of exit statuses.
func (ev *Evaluator) execPipeline(cmds []*command) {
	updates := make([]<-chan *StateUpdate, len(cmds))
	for i, cmd := range cmds {
		updates[i] = ev.execCommand(cmd)
	}

	for i, update := range updates {
		for up := range update {
			switch up.Msg {
			case "0", "":
			default:
				// XXX Update of commands in subevaluators should not be printed.
				fmt.Printf("Command #%d update: %s\n", i, up.Msg)
			}
		}
	}
}

// evalPipeline combines preevalPipeline and execPipeline.
func (ev *Evaluator) evalPipeline(pl *parse.ListNode) {
	cmds, _ := ev.preevalPipeline(pl)
	ev.execPipeline(cmds)
}
