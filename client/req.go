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

type Req struct {
    Cmd *ReqCmd `json:",omitempty"`
}

var reqFile *os.File

func InitReq(fd uintptr) {
    reqFile = os.NewFile(fd, "<request pipe>")
}

func SendReq(req Req) {
    json, err := json.Marshal(req)
    if err != nil {
        panic("failed to marshal request")
    }
    reqFile.Write(json)
    reqFile.WriteString("\n")
}
