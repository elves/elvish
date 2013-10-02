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
	// syscall.ProcAttr.Files and various other things internally.
	FD_NIL uintptr = ^uintptr(0)
)

// Represents an IO for commands. At most one of f and ch is non-nil. When
// both are nil, the IO is closed.
type io struct {
	f *os.File
	ch chan string
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
	args []string
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
func search(exe string) (string, error) {
	for _, p := range []string{"/", "./", "../"} {
		if strings.HasPrefix(exe, p) {
			if isExecutable(exe) {
				return exe, nil
			}
			return "", fmt.Errorf("not executable")
		}
	}
	for _, p := range search_paths {
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

func evalTerm(n parse.Node) (string, error) {
	return n.(*parse.StringNode).Text, nil
}

func evalTermList(ln *parse.ListNode) ([]string, error) {
	ss := make([]string, len(ln.Nodes))
	for i, n := range ln.Nodes {
		var e error
		ss[i], e = evalTerm(n)
		if e != nil {
			return nil, e
		}
	}
	return ss, nil
}

func evalCommand(n *parse.CommandNode) (cmd *command, ioTypes [3]ioType, files []*os.File, err error) {
	if len(n.Nodes) == 0 {
		err = fmt.Errorf("command is emtpy")
		return
	}

	// Build argument list. This is universal for all command types.
	args, e := evalTermList(&n.ListNode)
	if e != nil {
		err = fmt.Errorf("error evaluating arguments: %s", e)
		return
	}

	// Save unresolved args[0] as name.
	name := args[0]

	// Resolve command name.
	bi, isBuiltin := builtins[name]
	if isBuiltin {
		ioTypes = bi.ioTypes
	} else {
		// Try external command
		args[0], e = search(n.Nodes[0].(*parse.StringNode).Text)
		if e != nil {
			err = fmt.Errorf("can't resolve: %s", e)
			return
		}
		// Use zero value (fileIO) for ioTypes
	}

	// IO list.
	ios := [3]*io{}
	defaultErrIO := &io{f: os.Stderr}
	if defaultErrIO.compatible(ioTypes[2]) {
		ios[2] = defaultErrIO
	}

	files = make([]*os.File, 0)
	// Check IO redirections, turn all FilenameRedir to FdRedir.
	for _, r := range n.Redirs {
		fd := r.Fd()
		if fd > 2 {
			err = fmt.Errorf("redir on fd > 2 not yet supported")
			return
		}

		switch r := r.(type) {
		case *parse.CloseRedir:
			ios[fd] = &io{}
		case *parse.FdRedir:
			if ioTypes[fd] == chanIO {
				err = fmt.Errorf("fd redir on channel IO")
				return
			}
			if r.OldFd > 2 {
				err = fmt.Errorf("fd redir from fd > 2 not yet supported")
				return
			}
			ios[fd] = ios[r.OldFd]
		case *parse.FilenameRedir:
			if ioTypes[fd] == chanIO {
				err = fmt.Errorf("filename redir on channel IO")
				return
			}
			fname, e := evalTerm(r.Filename)
			if e != nil {
				err = fmt.Errorf("failed to evaluate filename: %q: %s",
				                 r.Filename, e)
				return
			}
			// TODO haz hardcoded permbits now
			f, e := os.OpenFile(fname, r.Flag, 0644)
			if e != nil {
				err = fmt.Errorf("failed to open file %q: %s", r.Filename, e)
				return
			}
			files = append(files, f)
			ios[fd] = &io{f: f}
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
func ExecPipeline(pl *parse.ListNode) (updates []<-chan *StateUpdate, err error) {
	ncmds := len(pl.Nodes)
	if ncmds == 0 {
		return []<-chan *StateUpdate{}, nil
	}

	cmds := make([]*command, 0, ncmds)

	var filesToClose []*os.File
	var chansToClose []chan string
	defer func() {
		for _, f := range filesToClose {
			f.Close()
		}
		for _, ch := range chansToClose {
			close(ch)
		}
	}()

	var nextIn *io
	for i, n := range pl.Nodes {
		cmd, ioTypes, files, err := evalCommand(n.(*parse.CommandNode))
		filesToClose = append(filesToClose, files...)
		if err != nil {
			return nil, fmt.Errorf("error with command #%d: %s", i, err)
		}

		// Create and connect pipes.
		if i == 0 {
			// First command. Only connect input when no input redirection is
			// present.
			if cmd.ios[0] == nil {
				if ioTypes[0] == chanIO {
					return nil, fmt.Errorf("channel input from user not yet supported")
				}
				cmd.ios[0] = &io{f: os.Stdin}
			}
		} else {
			if cmd.ios[0] != nil {
				return nil, fmt.Errorf("command #%d has both pipe input and input redirection")
			} else if !nextIn.compatible(ioTypes[0]) {
				return nil, fmt.Errorf("command #%d has incompatible input pipe")
			}
			cmd.ios[0] = nextIn
		}
		if i == ncmds - 1 {
			if cmd.ios[1] == nil {
				if ioTypes[1] == chanIO {
					return nil, fmt.Errorf("channel output to user not yet supported")
				}
				cmd.ios[1] = &io{f: os.Stdout}
			}
		} else {
			if cmd.ios[1] != nil {
				return nil, fmt.Errorf("command #%d has both pipe output and output redirection", i)
			}
			switch ioTypes[1] {
			case unusedIO:
				return nil, fmt.Errorf("command #%d has unused output connected in pipeline", i)
			case fileIO:
				// os.Pipe sets O_CLOEXEC, which is what we want.
				reader, writer, e := os.Pipe()
				if e != nil {
					return nil, fmt.Errorf("failed to create pipe: %s", e)
				}
				filesToClose = append(filesToClose, reader, writer)
				nextIn = &io{f: reader}
				cmd.ios[1] = &io{f: writer}
			case chanIO:
				// TODO Buffered channel?
				ch := make(chan string)
				chansToClose = append(chansToClose, ch)
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
		updates[i] = execCommand(cmd)
	}
	return updates, nil
}

// execCommand executes a command.
func execCommand(cmd *command) <-chan *StateUpdate {
	if cmd.f != nil {
		return execBuiltin(cmd)
	} else {
		return execExternal(cmd)
	}
}

// execBuiltin executes a builtin command.
func execBuiltin(cmd *command) <-chan *StateUpdate {
	update := make(chan *StateUpdate)
	go func() {
		// XXX builtins should return an exit code
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
func execExternal(cmd *command) <-chan *StateUpdate {
	files := make([]uintptr, len(cmd.ios))
	for i, io := range cmd.ios {
		if io == nil || io.f == nil {
			files[i] = FD_NIL
		} else {
			files[i] = io.f.Fd()
		}
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: envAsSlice(env), Files: files[:], Sys: &sys}
	pid, err := syscall.ForkExec(cmd.args[0], cmd.args, &attr)

	update := make(chan *StateUpdate)
	if err != nil {
		update <- &StateUpdate{Terminated: true, Msg: err.Error()}
		close(update)
	} else {
		go waitStateUpdate(pid, update)
	}

	return update
}
