package main

import (
    "os"
    "bufio"
    "errors"
    "encoding/json"
)

type TypePicker struct {
    Type string
}

type DataPicker struct {
    Data interface{}
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

func RecvRes() (interface{}, error) {
    pkt, err := resFile.ReadBytes('\n')
    if err != nil {
        return nil, err
    }

    var tp TypePicker
    err = json.Unmarshal(pkt, &tp)
    if err != nil {
        return nil, err
    }
    var dp DataPicker
    switch (tp.Type) {
    case "cmd":
        dp.Data = &ResCmd{}
    case "procState":
        dp.Data = &ResProcState{}
    default:
        return nil, errors.New("Unknown response type")
    }
    err = json.Unmarshal(pkt, &dp)
    if err != nil {
        return nil, err
    }
    return dp.Data, err
}
