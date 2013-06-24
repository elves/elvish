package main

import (
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

func SendReq(req Req) {
    json, err := json.Marshal(req)
    if err != nil {
        panic("failed to marshal request")
    }
    TubeFile.Write(json)
    TubeFile.WriteString("\n")
}
