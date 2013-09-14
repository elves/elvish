package main

import (
	"os"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"../libdasc/parse"
	"../libdasc/editor"
)

var env map[string]string
var search_paths []string

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: dasc <text fd> <control fd>\n");
}

func getIntArg(i int) int {
	if i < len(os.Args) {
		a, err := strconv.Atoi(os.Args[i])
		if err == nil {
			return a
		}
	}
	usage()
	os.Exit(1)
	return -1
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
	InitTube(getIntArg(1), getIntArg(2))

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

		nredirs := len(tree.Root.Redirs)

		cmd := ReqCmd{
			Path: args[0],
			Args: args,
			Env: env,
			Redirs: make([][2]int, nredirs),
			FdsToSend: make([]int, nredirs),
		}

		for i, r := range tree.Root.Redirs {
			cmd.Redirs[i][0] = r.Oldfd()

			switch r := r.(type) {
			case *parse.FdRedir:
				cmd.Redirs[i][1] = r.Newfd
			case *parse.CloseRedir:
				cmd.Redirs[i][1] = FdClose
			case *parse.FilenameRedir:
				// TODO haz hardcoded permbits now
				fname := evalTerm(r.Filename)
				newfd, err := syscall.Open(fname, r.Flag, 0644)
				if err != nil {
					// TODO Should abort command execution
					fmt.Fprintf(os.Stderr, "Failed to open file %q: %s\n",
					            r.Filename, err)
					cmd.Redirs[i][1] = FdClose
				} else {
					cmd.Redirs[i][1] = FdSend
					cmd.FdsToSend[i] = newfd
				}
			default:
				panic("unreachable")
			}
		}

		SendReq(Req{Cmd: &cmd})

		for i, r := range cmd.Redirs {
			if r[1] == FdSend {
				syscall.Close(cmd.FdsToSend[i])
			}
		}

		for {
			res, err := RecvRes()
			if err != nil {
				fmt.Printf("broken response pipe, quitting")
				os.Exit(1)
			}

			if res.ProcState != nil {
				break
			} else if br := res.BadRequest; br != nil {
				fmt.Printf("server complained bad request: %s\n", br.Err)
				break
			} else if c := res.Cmd; c != nil {
				// fmt.Printf("forked: %d\n", c.Pid)
			}
		}
	}
	SendReq(Req{Exit: &ReqExit{}})
}
