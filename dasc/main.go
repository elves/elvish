package main

import (
	"os"
	"fmt"
	"bufio"
	"strconv"
	"strings"
	"syscall"
	"./parse"
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

func prompt() {
	fmt.Print("> ")
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

func readline(stdin *bufio.Reader) (line string, err error) {
	line, err = stdin.ReadString('\n')
	if err == nil {
		line = line[:len(line)-1]
	}
	return
}

func evalList(n *parse.ListNode) (words []string) {
	words = make([]string, 0, len(n.Nodes))
	for _, w := range n.Nodes {
		words = append(words, w.(*parse.StringNode).Text)
	}
	return
}

func main() {
	InitTube(getIntArg(1), getIntArg(2))

	stdin := bufio.NewReader(os.Stdin)
	devnull, err := syscall.Open("/dev/null", syscall.O_WRONLY, 0)

	if err != nil {
		panic("Failed to open /dev/null")
	}

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

		prompt()
		line, err := readline(stdin)
		if err != nil {
			lackeol()
			break
		}
		tree, err := parse.Do(name, line)
		if err != nil {
			fmt.Println("Parser error:", err)
			continue
		}
		words := evalList(tree.Root)
		if len(words) == 0 {
			continue
		}
		full := search(words[0])
		if len(full) == 0 {
			fmt.Println("command not found:", words[0])
			continue
		}
		words[0] = full
		cmd := ReqCmd{
			Path: words[0],
			Args: words,
			Env: env,
			// RedirOutput: true,
			Output: devnull,
		}

		SendReq(Req{Cmd: &cmd})

		for {
			res, err := RecvRes()
			if err != nil {
				fmt.Printf("broken response pipe, quitting")
				os.Exit(1)
			} else {
				// fmt.Printf("response: %s\n", res)
			}

			if res.ProcState != nil {
				break
			}
		}
	}
	SendReq(Req{Exit: &ReqExit{}})
}
