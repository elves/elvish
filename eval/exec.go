package eval

import (
	"os"
	"fmt"
	"strings"
	"syscall"
	"../parse"
)

const (
	FILE_CLOSE uintptr = ^uintptr(0)
)

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

func evalTerm(n parse.Node) error {
	_ = n.(*parse.StringNode)
	return nil
}

func evalTermList(ln *parse.ListNode) error {
	for _, n := range ln.Nodes {
		evalTerm(n)
	}
	return nil
}

func resolveExternal(n *parse.StringNode) error {
	s, err := search(n.Text)
	if err != nil {
		return err
	}
	n.Text = s
	return nil
}

// CommandErrors holds multiple errors.
type CommandErrors struct {
	Errors []error
}

func (ce CommandErrors) Error() string {
	return fmt.Sprintf("%v", ce.Errors)
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
func ExecPipeline(pl *parse.ListNode) (pids []int, err error) {
	ncmds := len(pl.Nodes)
	if ncmds == 0 {
		return []int{}, nil
	}

	nextReadPipe := -1

	for i, cmd := range pl.Nodes {
		cmd := cmd.(*parse.CommandNode)

		if len(cmd.Nodes) == 0 {
			return nil, fmt.Errorf("command #%d is emtpy", i)
		}

		err = evalTermList(&cmd.ListNode)
		if err != nil {
			return nil, fmt.Errorf("error evaluating command #%d: %s", err)
		}

		err = resolveExternal(cmd.Nodes[0].(*parse.StringNode))
		if err != nil {
			return nil, fmt.Errorf("can't resolve command #%d: %s", i, err)
		}

		// Create pipes.
		var readPipe, writePipe int
		readPipe = nextReadPipe
		writePipe = -1
		if i != ncmds - 1 {
			// os.Pipe sets O_CLOEXEC, which is what we want.
			reader, writer, e := os.Pipe()
			if e != nil {
				return nil, fmt.Errorf("failed to create pipe: %s", e)
			}
			defer reader.Close()
			defer writer.Close()
			nextReadPipe = int(reader.Fd())
			writePipe = int(writer.Fd())
		}

		// Check IO redirections, turn all FilenameRedir to FdRedir.
		// XXX pipes are not yet connected.
		for j, r := range cmd.Redirs {
			fd := r.Fd()
			if fd > 2 {
				return nil, fmt.Errorf("redir on fd > 2 not yet supported")
			} else if fd == 0 && readPipe != -1 {
				return nil, fmt.Errorf("input already connected to pipe")
			} else if fd == 1 && writePipe != -1 {
				return nil, fmt.Errorf("output already connected to pipe")
			}
			switch r := r.(type) {
			case *parse.FdRedir:
				if r.OldFd > 2 {
					return nil, fmt.Errorf("fd redir from fd > 2 not yet supported")
				}
			case *parse.FilenameRedir:
				evalTerm(r.Filename)
				fname := r.Filename.(*parse.StringNode).Text
				// TODO haz hardcoded permbits now
				f, err := os.OpenFile(fname, r.Flag, 0644)
				if err != nil {
					return nil, fmt.Errorf("failed to open file %q: %s",
					                       r.Filename, err)
				}
				oldFd := int(f.Fd())
				cmd.Redirs[j] = parse.NewFdRedir(fd, oldFd)
				defer syscall.Close(oldFd)
			}
		}

		// Connect pipes.
		if readPipe != -1 {
			readRedir := parse.NewFdRedir(0, readPipe)
			cmd.Redirs = append(cmd.Redirs, readRedir)
		}
		if writePipe != -1 {
			writeRedir := parse.NewFdRedir(1, writePipe)
			cmd.Redirs = append(cmd.Redirs, writeRedir)
		}

	}

	pids = make([]int, ncmds)
	cmderr := CommandErrors{Errors: make([]error, ncmds)}
	haserr := false

	for i, cmd := range pl.Nodes {
		pid, err := ExecCommand(cmd.(*parse.CommandNode))

		if err != nil {
			pids[i] = -1
			cmderr.Errors[i] = err
			haserr = true
		} else {
			pids[i] = pid
		}
	}

	if haserr {
		return pids, cmderr
	}
	return pids, nil
}

func extractTexts(ln *parse.ListNode) (texts []string) {
	texts = make([]string, 0, len(ln.Nodes))
	for _, n := range ln.Nodes {
		texts = append(texts, n.(*parse.StringNode).Text)
	}
	return
}

// ExecCommand executes a command.
func ExecCommand(cmd *parse.CommandNode) (pid int, err error) {
	args := extractTexts(&cmd.ListNode)

	files := []uintptr{0, 1, 2}
	for _, r := range cmd.Redirs {
		fd := r.Fd()

		switch r := r.(type) {
		case *parse.FdRedir:
			oldFd := r.OldFd
			if oldFd < 3 {
				files[fd] = files[r.OldFd]
			} else {
				files[fd] = uintptr(oldFd)
			}
		case *parse.CloseRedir:
			files[fd] = FILE_CLOSE
		case *parse.FilenameRedir:
			panic("can't haz FilenameRedir here")
		default:
			panic("unreachable")
		}
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: envAsSlice(env), Files: files, Sys: &sys}

	return syscall.ForkExec(args[0], args, &attr)
}
