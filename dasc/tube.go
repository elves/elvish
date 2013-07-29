package main

import (
	"os"
	"encoding/json"
)

var TextTube *os.File
var ResDecoder *json.Decoder
var FdTube int

func InitTube(textTube int, fdTube int) {
	TextTube = os.NewFile(uintptr(textTube), "<das tube>")
	ResDecoder = json.NewDecoder(TextTube)
	FdTube = fdTube
}
