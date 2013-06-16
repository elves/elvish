package main

import (
    "os"
    "fmt"
    "bufio"
    "strconv"
    "encoding/json"
)

type command struct {
    Path string `json:"path"`
    Args []string `json:"args"`
    Env map[string]string `json:"env"`
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

func main() {
    reqfd := uintptr(getIntArg(1))
    resfd := uintptr(getIntArg(2))
    req := os.NewFile(reqfd, "<request pipe>")
    res := bufio.NewReader(os.NewFile(resfd, "<response pipe>"))

    cmd := command{
        "/usr/bin/env",
        []string{"/usr/bin/env"},
        map[string]string{"some_key": "some_value"},
    }
    json, err := json.Marshal(cmd)
    if err != nil {
        panic("failed to marshal command")
    }
    req.Write(json)
    req.WriteString("\n")

    msg, err := res.ReadBytes('\n')
    if err == nil {
        fmt.Printf("response: %s", msg)
    }
}
