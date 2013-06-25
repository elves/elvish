package main

import (
    "os"
    "encoding/json"
)

var TextTube *os.File
var ResDecoder *json.Decoder
var FdTube uintptr

func InitTube(textTube uintptr, fdTube uintptr) {
    TextTube = os.NewFile(textTube, "<das tube>")
    ResDecoder = json.NewDecoder(TextTube)
    FdTube = fdTube
}
