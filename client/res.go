package main

import (
    "os"
    "bufio"
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

var resFile *bufio.Reader

func InitRes(fd uintptr) {
    resFile = bufio.NewReader(os.NewFile(fd, "<response pipe>"))
}

func ReadRes() (string, error) {
    return resFile.ReadString('\n')
}

func RecvRes() (Res, error) {
    pkt, err := resFile.ReadBytes('\n')
    if err != nil {
        return Res{}, err
    }

    var r Res
    err = json.Unmarshal(pkt, &r)
    return r, err
}
