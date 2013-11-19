package eval

import (
	"os"
	"fmt"
	"strings"
	"syscall"
	"../parse"
)

const (
	// A special impossible fd value. Used for "close fd" in
	// syscall.ProcAttr.Files.
	FD_NIL uintptr = ^uintptr(0)
)

// Represents an IO for commands. At most one of f and ch is non-nil. When
// both are nil, the IO is closed.
type io struct {
	f *os.File
	ch chan Value
}

func (i *io) compatible(typ ioType) bool {
	if i == nil {
		return false
	}
	if typ == unusedIO {
		return true
	}
	switch {
	case i.f != nil:
		return typ == fileIO
	case i.ch != nil:
		return typ == chanIO
	default:
		return true
	}
}

type command struct {
	name string // Command name, used in error messages.
	args []Value // Argument list, minus command name.
	ios [3]*io // IOs for in, out and err.
	fn builtinFunc // A builtin function, if the command is builtin.
	path string // Command full path, if the command is external.
	closure *Closure // The closure value, if the command is a closure.
}

type StateUpdate struct {
	Terminated bool
	Msg string
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
	return !fm.IsDir() && (fm & 0111 != 0)
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

func (ev *Evaluator) evalRedir(r parse.Redir, ios []*io, ioTypes []ioType) {
	ev.push(r)
	defer ev.pop()

	fd := r.Fd()
	if fd > 2 {
		ev.errorf("redir on fd > 2 not yet supported")
	}

	switch r := r.(type) {
	case *parse.CloseRedir:
		ios[fd] = &io{}
	case *parse.FdRedir:
		if ioTypes[fd] == chanIO {
			ev.errorf("fd redir on channel IO")
		}
		if r.OldFd > 2 {
			ev.errorf("fd redir from fd > 2 not yet supported")
		}
		ios[fd] = ios[r.OldFd]
	case *parse.FilenameRedir:
		if ioTypes[fd] == chanIO {
			ev.errorf("filename redir on channel IO")
		}
		fname := ev.evalTermSingleScalar(r.Filename, "filename").str
		// TODO haz hardcoded permbits now
		f, e := os.OpenFile(fname, r.Flag, 0644)
		if e != nil {
			ev.errorf("failed to open file %q: %s", fname[0], e)
		}
		ios[fd] = &io{f: f}
		// XXX Files opened in redirections of builtins shouldn't be
		// closed.
		ev.filesToClose = append(ev.filesToClose, f)
	}
}

func (ev *Evaluator) preevalCommand(n *parse.CommandNode) (cmd *command, ioTypes [3]ioType) {
	// Evaluate name.
	nameValues := ev.evalTerm(n.Name)
	if len(nameValues) != 1 {
		ev.errorfNode(n.Name, "command name must be a single value")
	}
	name := nameValues[0]

	// Start building command.
	cmd = &command{name: name.String(ev)}

	// Resolve command. Assign one of cmd.{fn path closure} and ioTypes.
	switch name := name.(type) {
	case *Scalar:
		nameStr := name.str
		// Try function
		if v, ok := ev.getVar("fn-" + nameStr); ok {
			if fn, ok := v.(*Closure); ok {
				cmd.closure = fn
				// XXX Use zero value (fileIO) for ioTypes now
				break
			}
		}

		// Try builtin
		if bi, ok := builtins[nameStr]; ok {
			cmd.fn = bi.fn
			ioTypes = bi.ioTypes
			break
		}

		// Try external command
		path, e := ev.search(nameStr)
		if e != nil {
			ev.errorfNode(n.Name, "%s", e)
		}
		cmd.path = path
		// Use zero value (fileIO) for ioTypes
	case *Closure:
		cmd.closure = name
		// XXX Use zero value (fileIO) for ioTypes now
	default:
		ev.errorfNode(n.Name, "Command name must be either scalar or closure")
	}

	// IO list.
	defaultErrIO := &io{f: os.Stderr}
	// XXX Should we allow chan IO stderr at all?
	if defaultErrIO.compatible(ioTypes[2]) {
		cmd.ios[2] = defaultErrIO
	}

	// Evaluate IO redirections.
	for _, r := range n.Redirs {
		ev.evalRedir(r, cmd.ios[:], ioTypes[:])
	}

	// Evaluate arguments after everything else.
	cmd.args = ev.evalTermList(n.Args)
	return
}

// ExecPipeline executes a pipeline.
//
// As many things as possible are done before any command actually gets
// executed, to avoid leaving the pipeline broken - resolving command names,
// opening files, and in future, evaluating shell constructs. If any error is
// encountered, pids is nil and err contains the error.
//
// However, if error is encountered when executing individual commands, the
// rest of the pipeline will still be executed. In that case, the
// corresponding elements in pids is -1 and err is typed *CommandErrors. For
// each pids[i] == -1, err.(*CommandErrors)Errors[i] contains the
// corresponding error.
func (ev *Evaluator) evalPipeline(pl *parse.ListNode) []<-chan *StateUpdate {
	ev.push(pl)
	defer ev.pop()

	defer func() {
		for _, f := range ev.filesToClose {
			f.Close()
		}
		ev.filesToClose = ev.filesToClose[0:]
	}()

	ncmds := len(pl.Nodes)
	if ncmds == 0 {
		return []<-chan *StateUpdate{}
	}

	cmds := make([]*command, 0, ncmds)

	var nextIn *io
	for i, n := range pl.Nodes {
		cmd, ioTypes := ev.preevalCommand(n.(*parse.CommandNode))

		// Create and connect pipes.
		if i == 0 {
			// First command. Only connect input when no input redirection is
			// present.
			if cmd.ios[0] == nil {
				if ioTypes[0] == chanIO {
					// TODO locate command
					ev.errorf("channel input from user not yet supported")
				}
				cmd.ios[0] = &io{f: os.Stdin}
			}
		} else {
			if cmd.ios[0] != nil {
				ev.errorf("command #%d has both pipe input and input redirection")
			} else if !nextIn.compatible(ioTypes[0]) {
				ev.errorf("command #%d has incompatible input pipe")
			}
			cmd.ios[0] = nextIn
		}
		if i == ncmds - 1 {
			if cmd.ios[1] == nil {
				if ioTypes[1] == chanIO {
					ev.errorf("channel output to user not yet supported")
				}
				cmd.ios[1] = &io{f: os.Stdout}
			}
		} else {
			if cmd.ios[1] != nil {
				ev.errorf("command #%d has both pipe output and output redirection", i)
			}
			switch ioTypes[1] {
			case unusedIO:
				ev.errorf("command #%d has unused output connected in pipeline", i)
			case fileIO:
				// os.Pipe sets O_CLOEXEC, which is what we want.
				reader, writer, e := os.Pipe()
				if e != nil {
					ev.errorf("failed to create pipe: %s", e)
				}
				// XXX The pipe end for builtins shouldn't be closed.
				ev.filesToClose = append(ev.filesToClose, reader, writer)
				nextIn = &io{f: reader}
				cmd.ios[1] = &io{f: writer}
			case chanIO:
				// TODO Buffered channel?
				// XXX Builtins are relied on to close channels.
				ch := make(chan Value)
				nextIn = &io{ch: ch}
				cmd.ios[1] = &io{ch: ch}
			default:
				panic("unreachable")
			}
		}

		cmds = append(cmds, cmd)
	}

	updates := make([]<-chan *StateUpdate, ncmds)
	for i, cmd := range cmds {
		updates[i] = ev.execCommand(cmd)
	}
	return updates
}

// execCommand executes a command.
func (ev *Evaluator) execCommand(cmd *command) <-chan *StateUpdate {
	switch {
	case cmd.fn != nil:
		return ev.execBuiltin(cmd)
	case cmd.path != "":
		return ev.execExternal(cmd)
	case cmd.closure != nil:
		return ev.execClosure(cmd)
	default:
		panic("Bad eval.command struct")
	}
}

func (ev *Evaluator) execClosure(cmd *command) <-chan *StateUpdate {
	update := make(chan *StateUpdate, 1)

	locals := make(map[string]Value)
	// TODO Support optional/rest argument
	if len(cmd.args) != len(cmd.closure.ArgNames) {
		// TODO Check arity before exec'ing
		update <- &StateUpdate{Terminated: true, Msg: "arity mismatch"}
		close(update)
		return update
	}
	// Pass argument by populating locals.
	for i, name := range cmd.closure.ArgNames {
		locals[name] = cmd.args[i]
	}

	// Make a subevaluator.
	// TODO Guard against concurrent writes to globals.
	newEv := Evaluator{
		globals: ev.globals,
		locals: locals,
		env: ev.env,
		searchPaths: ev.searchPaths,
	}
	go func() {
		// TODO Support calling closure originated in another source.
		newEv.Eval(ev.name, ev.text, cmd.closure.Chunk)
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
		msg := cmd.fn(ev, cmd.args, cmd.ios)
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
	files := make([]uintptr, len(cmd.ios))
	for i, io := range cmd.ios {
		if io == nil || io.f == nil {
			files[i] = FD_NIL
		} else {
			files[i] = io.f.Fd()
		}
	}

	args := make([]string, len(cmd.args) + 1)
	args[0] = cmd.path
	for i, a := range cmd.args {
		// XXX Silently coerce all args into string
		args[i+1] = a.String(ev)
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: ev.env.Export(), Files: files[:], Sys: &sys}
	pid, err := syscall.ForkExec(cmd.path, args, &attr)

	update := make(chan *StateUpdate)
	if err != nil {
		update <- &StateUpdate{Terminated: true, Msg: err.Error()}
		close(update)
	} else {
		go waitStateUpdate(pid, update)
	}

	return update
}
