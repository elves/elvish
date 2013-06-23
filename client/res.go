package main

import (
    "os"
    "encoding/json"
)

type Res struct {
    Cmd *ResCmd
    ProcState *ResProcState
}

type ResCmd struct {
    Pid int
}

type ResProcState struct {
    Pid int
    Exited bool
    ExitStatus int
    Signaled bool
    TermSig int
    CoreDump bool
    Stopped bool
    StopSig int
    Continued bool
}

var resDecoder *json.Decoder

func InitRes(fd uintptr) {
    resDecoder = json.NewDecoder(os.NewFile(fd, "<response pipe>"))
}

func RecvRes() (r Res, err error) {
    err = resDecoder.Decode(&r)
    return
}
