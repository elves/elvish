package main

import (
    "os"
    "fmt"
    "bufio"
    "strconv"
    "strings"
    "encoding/json"
)

type ReqCmd struct {
    Path string
    Args []string
    Env map[string]string
}

type Req struct {
    Type string
    Data interface{}
}

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
    reqfd := uintptr(getIntArg(1))
    resfd := uintptr(getIntArg(2))
    req := os.NewFile(reqfd, "<request pipe>")
    res := bufio.NewReader(os.NewFile(resfd, "<response pipe>"))

    stdin := bufio.NewReader(os.Stdin)

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
            words[0], words, env,
        }

        payload := Req{"cmd", cmd}

        json, err := json.Marshal(payload)
        if err != nil {
            panic("failed to marshal request")
        }
        req.Write(json)
        req.WriteString("\n")

        for {
            msg, err := res.ReadString('\n')
            if err != nil {
                fmt.Printf("broken response pipe, quitting")
                os.Exit(1)
            } else {
                fmt.Printf("response: %s", msg)
            }
            if strings.Contains(msg, "procState") {
                break
            }
        }
    }
}
