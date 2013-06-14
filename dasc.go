package main

import "net"
//import "time"

func main() {
    c, err := net.Dial("unix", "", "/tmp/das")
    if err != nil {
        panic(err.String())
    }
    defer c.close()
}
