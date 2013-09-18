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

// TODO return a separate error
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
// TODO return a separate error
func search(exe string) string {
	for _, p := range []string{"/", "./", "../"} {
		if strings.HasPrefix(exe, p) {
			return exe
		}
	}
	for _, p := range search_paths {
		full := p + "/" + exe
		if isExecutable(full) {
			return full
		}
	}
	return ""
}

func envAsSlice(env map[string]string) (s []string) {
	s = make([]string, len(env))
	for k, v := range env {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

func evalTerm(n parse.Node) string {
	return n.(*parse.StringNode).Text
}

func evalCommandArgs(n *parse.CommandNode) (args []string) {
	args = make([]string, 0, len(n.Nodes))
	for _, w := range n.Nodes {
		args = append(args, evalTerm(w))
	}
	return
}

// ExecCommand executes a command.
func ExecCommand(cmd *parse.CommandNode) (pid int, err error) {
	args := evalCommandArgs(cmd)
	if len(args) == 0 {
		return 0, fmt.Errorf("empty command")
	}
	full := search(args[0])
	if len(full) == 0 {
		return 0, fmt.Errorf("command not found: %s", args[0])
	}
	args[0] = full

	files := []uintptr{0, 1, 2}

	for _, r := range cmd.Redirs {
		fd := r.Fd()

		if fd > 2 {
			return 0, fmt.Errorf("redir on fd > 2 not yet supported")
		}

		switch r := r.(type) {
		case *parse.FdRedir:
			oldFd := r.OldFd
			if oldFd > 2 {
				return 0, fmt.Errorf("fd redir from fd > 2 not yet supported")
			}
			files[fd] = files[oldFd]
		case *parse.CloseRedir:
			files[fd] = FILE_CLOSE
		case *parse.FilenameRedir:
			// TODO haz hardcoded permbits now
			fname := evalTerm(r.Filename)
			oldFd, err := syscall.Open(fname, r.Flag, 0644)
			if err != nil {
				return 0, fmt.Errorf("failed to open file %q: %s",
				                     r.Filename, err)
			}
			files[fd] = uintptr(oldFd)
			defer syscall.Close(oldFd)
		default:
			panic("unreachable")
		}
	}

	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: envAsSlice(env), Files: files, Sys: &sys}
	return syscall.ForkExec(full, args, &attr)
}
