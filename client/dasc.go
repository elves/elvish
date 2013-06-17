package main

import (
    "os"
    "fmt"
    "bufio"
    "strconv"
    "strings"
    "encoding/json"
)

type command struct {
    Path string `json:"path"`
    Args []string `json:"args"`
    Env map[string]string `json:"env"`
}

type request struct {
    Type string `json:"type"`
    Data interface{} `json:"data"`
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
        cmd := command{}
        cmd.Path = words[0]
        cmd.Args = words
        cmd.Env = map[string]string{}

        payload := request{"command", cmd}

        json, err := json.Marshal(payload)
        if err != nil {
            panic("failed to marshal request")
        }
        req.Write(json)
        req.WriteString("\n")

        msg, err := res.ReadBytes('\n')
        if err == nil {
            fmt.Printf("response: %s", msg)
        }
    }
}
