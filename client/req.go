package main

import (
    "os"
    "encoding/json"
)

type ReqCmd struct {
    Path string
    Args []string
    Env map[string]string
}

func (r ReqCmd) Type() string {
    return "cmd"
}

type Req struct {
    Type string
    Data interface{}
}

var reqFile *os.File

func InitReq(fd uintptr) {
    reqFile = os.NewFile(fd, "<request pipe>")
}

func SendReq(req Typer) {
    json, err := json.Marshal(Req{req.Type(), req})
    if err != nil {
        panic("failed to marshal request")
    }
    reqFile.Write(json)
    reqFile.WriteString("\n")
}
