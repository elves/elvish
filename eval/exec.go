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

type io struct {
	f *os.File
	ch chan string
}

func (i *io) compatible(typ ioType) bool {
	if i == nil || typ == unusedIO {
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

func evalCommand(n *parse.CommandNode, in, out *io) (cmd *command, files []*os.File, err error) {
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

	// IO compatibility list, defaulting to all fileIO.
	var ioTypes [3]ioType

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
	}

	if !in.compatible(ioTypes[0]) {
		err = fmt.Errorf("Incompatible input pipe")
		return
	}
	if !out.compatible(ioTypes[1]) {
		err = fmt.Errorf("Incompatible output pipe")
		return
	}

	// IO list.
	ios := [3]*io{in, out}
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
			ios[fd] = nil
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
		// Create pipes.
		var in, out *io
		if i == 0 {
			in = &io{f: os.Stdin}
		} else {
			in = nextIn
		}
		if i == ncmds - 1 {
			out = &io{f: os.Stdout}
		} else {
			// os.Pipe sets O_CLOEXEC, which is what we want.
			// XXX Assumes fileIO now
			reader, writer, e := os.Pipe()
			if e != nil {
				return nil, fmt.Errorf("failed to create pipe: %s", e)
			}
			filesToClose = append(filesToClose, reader, writer)
			nextIn = &io{f: reader}
			out = &io{f: writer}
		}

		cmd, files, err := evalCommand(n.(*parse.CommandNode), in, out)
		filesToClose = append(filesToClose, files...)

		if err != nil {
			return nil, fmt.Errorf("error with command #%d: %s", i, err)
		}

		if cmd.f != nil {
			return nil, fmt.Errorf("Only external command is supported now")
		}
		cmds = append(cmds, cmd)
	}

	updates = make([]<-chan *StateUpdate, ncmds)
	for i, cmd := range cmds {
		updates[i] = ExecCommand(cmd)
	}
	return updates, nil
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

// ExecCommand executes a command.
func ExecCommand(cmd *command) <-chan *StateUpdate {
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
