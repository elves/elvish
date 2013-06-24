package main

import (
    "os"
    "encoding/json"
)

var TubeFile *os.File
var ResDecoder *json.Decoder

func InitTube(fd uintptr) {
    TubeFile = os.NewFile(fd, "<das tube>")
    ResDecoder = json.NewDecoder(TubeFile)
}
