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
	// Full argument list. args[0] is always some form of command name.
	args []Value
	ios [3]*io // IOs for in, out and err.
	// A pointer to the builtin function, if the command is builtin.
	f builtinFunc
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
			return "", fmt.Errorf("not executable")
		}
	}
	for _, p := range ev.searchPaths {
		full := p + "/" + exe
		if isExecutable(full) {
			return full, nil
		}
	}
	return "", fmt.Errorf("not found")
}

func envAsSlice(env map[string]string) (s []string) {
	s = make([]string, 0, len(env))
	for k, v := range env {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

func (ev *Evaluator) evalCommand(n *parse.CommandNode) (cmd *command, ioTypes [3]ioType) {
	var e error
	if len(n.Nodes) == 0 {
		ev.errorf(n, "command is emtpy")
	}

	// Build argument list. This is universal for all command types.
	args := ev.evalTermList(&n.ListNode)

	if _, ok := args[0].(*Scalar); !ok {
		// XXX
		ev.errorf(n, "first word is not scalar: %s", args[0])
	}

	// Save unresolved args[0] as name.
	name := args[0].String()

	// Resolve command name.
	bi, isBuiltin := builtins[name]
	if isBuiltin {
		ioTypes = bi.ioTypes
	} else {
		// Try external command
		args[0].(*Scalar).str, e = ev.search(name)
		if e != nil {
			ev.errorf(n, "can't resolve: %s", e)
		}
		// Use zero value (fileIO) for ioTypes
	}

	// IO list.
	ios := [3]*io{}
	defaultErrIO := &io{f: os.Stderr}
	if defaultErrIO.compatible(ioTypes[2]) {
		ios[2] = defaultErrIO
	}

	// Check IO redirections, turn all FilenameRedir to FdRedir.
	for _, r := range n.Redirs {
		fd := r.Fd()
		if fd > 2 {
			// TODO locate redir node
			ev.errorf(n, "redir on fd > 2 not yet supported")
		}

		switch r := r.(type) {
		case *parse.CloseRedir:
			ios[fd] = &io{}
		case *parse.FdRedir:
			if ioTypes[fd] == chanIO {
				// TODO locate redir node
				ev.errorf(n, "fd redir on channel IO")
			}
			if r.OldFd > 2 {
				// TODO locate redir node
				ev.errorf(n, "fd redir from fd > 2 not yet supported")
			}
			ios[fd] = ios[r.OldFd]
		case *parse.FilenameRedir:
			if ioTypes[fd] == chanIO {
				// TODO locate redir node
				ev.errorf(n, "filename redir on channel IO")
			}
			fname := ev.evalTerm(r.Filename)
			if len(fname) != 1 {
				ev.errorf(r.Filename, "File name term must evaluate to exactly one word, got %d", len(fname))
			}
			if _, ok := fname[0].(*Scalar); !ok {
				ev.errorf(r.Filename, "file name not scalar: %s", fname[0])
				return
			}
			// TODO haz hardcoded permbits now
			f, e := os.OpenFile(fname[0].String(), r.Flag, 0644)
			if e != nil {
				ev.errorf(r.Filename, "failed to open file %q: %s", fname[0], e)
			}
			ios[fd] = &io{f: f}
			// XXX Files opened in redirections of builtins shouldn't be
			// closed.
			ev.filesToClose = append(ev.filesToClose, f)
		}
	}

	cmd = &command{name, args, ios, nil}
	if isBuiltin {
		cmd.f = bi.f
	}
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
func (ev *Evaluator) ExecPipeline(pl *parse.ListNode) (updates []<-chan *StateUpdate, err *Error) {
	defer ev.recover(&err)
	defer func() {
		for _, f := range ev.filesToClose {
			f.Close()
		}
	}()

	ncmds := len(pl.Nodes)
	if ncmds == 0 {
		return []<-chan *StateUpdate{}, nil
	}

	cmds := make([]*command, 0, ncmds)

	var nextIn *io
	for i, n := range pl.Nodes {
		cmd, ioTypes := ev.evalCommand(n.(*parse.CommandNode))

		// Create and connect pipes.
		if i == 0 {
			// First command. Only connect input when no input redirection is
			// present.
			if cmd.ios[0] == nil {
				if ioTypes[0] == chanIO {
					ev.errorf(n, "channel input from user not yet supported")
				}
				cmd.ios[0] = &io{f: os.Stdin}
			}
		} else {
			if cmd.ios[0] != nil {
				ev.errorf(n, "command #%d has both pipe input and input redirection")
			} else if !nextIn.compatible(ioTypes[0]) {
				ev.errorf(n, "command #%d has incompatible input pipe")
			}
			cmd.ios[0] = nextIn
		}
		if i == ncmds - 1 {
			if cmd.ios[1] == nil {
				if ioTypes[1] == chanIO {
					ev.errorf(n, "channel output to user not yet supported")
				}
				cmd.ios[1] = &io{f: os.Stdout}
			}
		} else {
			if cmd.ios[1] != nil {
				ev.errorf(n, "command #%d has both pipe output and output redirection", i)
			}
			switch ioTypes[1] {
			case unusedIO:
				ev.errorf(n, "command #%d has unused output connected in pipeline", i)
			case fileIO:
				// os.Pipe sets O_CLOEXEC, which is what we want.
				reader, writer, e := os.Pipe()
				if e != nil {
					ev.errorf(n, "failed to create pipe: %s", e)
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

	updates = make([]<-chan *StateUpdate, ncmds)
	for i, cmd := range cmds {
		updates[i] = ev.execCommand(cmd)
	}
	return updates, nil
}

// execCommand executes a command.
func (ev *Evaluator) execCommand(cmd *command) <-chan *StateUpdate {
	if cmd.f != nil {
		return ev.execBuiltin(cmd)
	} else {
		return ev.execExternal(cmd)
	}
}

// execBuiltin executes a builtin command.
func (ev *Evaluator) execBuiltin(cmd *command) <-chan *StateUpdate {
	update := make(chan *StateUpdate)
	go func() {
		msg := cmd.f(cmd.args, cmd.ios)
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

	args := make([]string, len(cmd.args))
	for i, a := range cmd.args {
		// XXX Silently coerce all args into string
		args[i] = a.String()
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: envAsSlice(ev.env), Files: files[:], Sys: &sys}
	pid, err := syscall.ForkExec(args[0], args, &attr)

	update := make(chan *StateUpdate)
	if err != nil {
		update <- &StateUpdate{Terminated: true, Msg: err.Error()}
		close(update)
	} else {
		go waitStateUpdate(pid, update)
	}

	return update
}
