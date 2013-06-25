package main

import (
    "os"
    "fmt"
    "bufio"
    "strconv"
    "strings"
    "syscall"
)

func usage() {
    fmt.Fprintf(os.Stderr, "Usage: dasc <req fd> <res fd>\n");
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

func search(path string) string {
    if path[0] == '/' {
        return path
    }
    return "/bin/" + path
}

func readline(stdin *bufio.Reader) (line string, err error) {
    line, err = stdin.ReadString('\n')
    if err == nil {
        line = line[:len(line)-1]
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

    env := make(map[string]string)
    for _, e := range os.Environ() {
        arr := strings.SplitN(e, "=", 2)
        if len(arr) == 2 {
            env[arr[0]] = arr[1]
        }
    }

    for {
        prompt()
        line, err := readline(stdin)
        if err != nil {
            lackeol()
            break
        }
        words := strings.Split(line, " ")
        if len(words) == 0 {
            continue
        }
        words[0] = search(words[0])
        cmd := ReqCmd{
            Path: words[0],
            Args: words,
            Env: env,
            RedirOutput: true,
            Output: devnull,
        }

        SendReq(Req{Cmd: &cmd})

        for {
            res, err := RecvRes()
            if err != nil {
                fmt.Printf("broken response pipe, quitting")
                os.Exit(1)
            } else {
                fmt.Printf("response: %v\n", res)
            }

            if res.ProcState != nil {
                break
            }
        }
    }
}
