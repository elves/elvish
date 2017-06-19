// Package daemon implements the daemon service for mediating access to the
// storage backend. It does not take care daemonization; for that part, see
// daemon/exec.
package daemon

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[daemon] ")

// Serve runs the daemon service. It does not return.
func Serve(sockpath, dbpath string) {
	logger.Println("pid is", syscall.Getpid())

	st, err := store.NewStore(dbpath)
	if err != nil {
		logger.Printf("failed to create storage: %v", err)
		logger.Println("aborting")
		os.Exit(2)
	}

	logger.Println("going to listen", sockpath)
	listener, err := net.Listen("unix", sockpath)
	if err != nil {
		logger.Printf("failed to listen on %s: %v", sockpath, err)
		logger.Println("aborting")
		os.Exit(2)
	}

	quitSignals := make(chan os.Signal)
	signal.Notify(quitSignals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-quitSignals
		logger.Printf("received signal %s, shutting down", sig)
		err := os.Remove(sockpath)
		if err != nil {
			logger.Println("failed to remove socket %s: %v", sockpath, err)
		}
		logger.Println("exiting")
		os.Exit(0)
	}()

	service := &Service{st}
	rpc.RegisterName(api.ServiceName, service)

	logger.Println("starting to serve RPC calls")
	rpc.Accept(listener)
	os.Exit(0)
}
