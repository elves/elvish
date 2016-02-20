package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// TODO Need to learn the semantics handle of SIGHUP and see if it deserves
// special treatment.
func main() {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan)
	fmt.Println("ok")
	os.Stdout.Sync()
	for sig := range signalChan {
		fmt.Println(int(sig.(syscall.Signal)))
		os.Stdout.Sync()
	}
}
