// elvishd is an agent for sharing variables and command-line history among
// multiple elvish processes.
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/xiaq/elvish/service"
	"github.com/xiaq/elvish/util"
)

const (
	SignalBufferSize = 32
)

func main() {
	laddr, err := util.SocketName()
	if err != nil {
		log.Fatalln("get socket name:", err)
	}

	// Listen to socket
	listener, err := net.Listen("unix", laddr)
	if err != nil {
		log.Fatalln("listen to socket:", err)
	}

	// Set up Unix signal handler
	sigch := make(chan os.Signal, SignalBufferSize)
	signal.Notify(sigch)
	go func() {
		for sig := range sigch {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				// TODO(xiaq): Notify current clients of termination
				os.Remove(laddr)
				os.Exit(0)
			default:
				// Ignore all other signals
			}
		}
	}()

	service.Serve(listener)
}
