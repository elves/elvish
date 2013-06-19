package main

import (
    "os"
    "bufio"
)

var resFile *bufio.Reader

func InitRes(fd uintptr) {
    resFile = bufio.NewReader(os.NewFile(fd, "<response pipe>"))
}

func ReadRes() (string, error) {
    return resFile.ReadString('\n')
}
