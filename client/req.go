package main

import (
    "encoding/json"
)

type ReqCmd struct {
    Path string
    Args []string
    Env map[string]string
    RedirInput bool
    Input int `json:"-"`
    RedirOutput bool
    Output int `json:"-"`
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
