package main

import (
	"os"
	"fmt"
	"strings"
	"syscall"
	"../libdasc/parse"
	"../libdasc/editor"
)

const (
	FILE_CLOSE uintptr = ^uintptr(0)
)

var env map[string]string
var search_paths []string

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: dasc <text fd> <control fd>\n");
}

func lackeol() {
	fmt.Println("\033[7m%\033[m")
}

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

func main() {
	env = make(map[string]string)
	for _, e := range os.Environ() {
		arr := strings.SplitN(e, "=", 2)
		if len(arr) == 2 {
			env[arr[0]] = arr[1]
		}
	}

	path_var, ok := env["PATH"]
	if ok {
		search_paths = strings.Split(path_var, ":")
		// fmt.Printf("Search paths are %v\n", search_paths)
	} else {
		search_paths = []string{"/bin"}
	}

	cmd_no := 0

repl:
	for {
		cmd_no++
		name := fmt.Sprintf("<interactive code %d>", cmd_no)

		ed, err := editor.Init(os.Stdin)
		if err != nil {
			panic(err)
		}
		lr := ed.ReadLine("das> ")
		err = ed.Cleanup()
		if err != nil {
			panic(err)
		}

		if lr.Eof {
			lackeol()
			break
		} else if lr.Err != nil {
			panic(lr.Err)
		}
		line := lr.Line
		tree, err := parse.Do(name, line, false)
		if err != nil {
			fmt.Println("Parser error:", err)
			continue
		}
		args := evalCommandArgs(tree.Root)
		if len(args) == 0 {
			continue
		}
		full := search(args[0])
		if len(full) == 0 {
			fmt.Println("command not found:", args[0])
			continue
		}
		args[0] = full

		files := []uintptr{0, 1, 2}

		for _, r := range tree.Root.Redirs {
			fd := r.Fd()

			if fd > 2 {
				fmt.Fprintln(os.Stderr, "Redir on fd > 2 not yet supported")
				continue repl
			}

			switch r := r.(type) {
			case *parse.FdRedir:
				oldFd := r.OldFd
				if oldFd > 2 {
					fmt.Fprintln(os.Stderr, "FD redir from fd > 2 not yet supported")
					continue repl
				}
				files[fd] = files[oldFd]
			case *parse.CloseRedir:
				files[fd] = FILE_CLOSE
			case *parse.FilenameRedir:
				// TODO haz hardcoded permbits now
				fname := evalTerm(r.Filename)
				oldFd, err := syscall.Open(fname, r.Flag, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to open file %q: %s\n",
					            r.Filename, err)
					continue repl
				} else {
					files[fd] = uintptr(oldFd)
				}
			default:
				panic("unreachable")
			}
		}

		sys := syscall.SysProcAttr{}
		attr := syscall.ProcAttr{Env: envAsSlice(env), Files: files, Sys: &sys}
		pid, err := syscall.ForkExec(full, args, &attr)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fork/exec: %s", err)
			continue repl
		}

		var ws syscall.WaitStatus
		var ru syscall.Rusage

		// TODO Should check ws
		syscall.Wait4(pid, &ws, 0, &ru)
	}
}
