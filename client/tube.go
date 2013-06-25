package main

import (
    "os"
    "encoding/json"
)

var TubeFile *os.File
var ResDecoder *json.Decoder
var FdTube uintptr

func InitTube(textTube uintptr, fdTube uintptr) {
    TubeFile = os.NewFile(textTube, "<das tube>")
    ResDecoder = json.NewDecoder(TubeFile)
    FdTube = fdTube
}
