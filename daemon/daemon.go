// Package daemon implements a daemon for mediating access to the storage
// backend of elvish.
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

// Daemon is a daemon.
type Daemon struct {
	sockpath string
	dbpath   string
}

// New creates a new daemon.
func New(sockpath, dbpath string) *Daemon {
	return &Daemon{sockpath, dbpath}
}

// Main runs the daemon. It does not take care of forking and stuff; it assumes
// that it is already running in the correct process.
func (d *Daemon) Main() int {
	logger.Println("pid is", syscall.Getpid())

	st, err := store.NewStore(d.dbpath)
	if err != nil {
		logger.Printf("failed to create storage: %v", err)
		logger.Println("aborting")
		return 2
	}

	logger.Println("going to listen", d.sockpath)
	listener, err := net.Listen("unix", d.sockpath)
	if err != nil {
		logger.Printf("failed to listen on %s: %v", d.sockpath, err)
		logger.Println("aborting")
		return 2
	}

	quitSignals := make(chan os.Signal)
	signal.Notify(quitSignals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-quitSignals
		logger.Printf("received signal %s, shutting down", sig)
		err := os.Remove(d.sockpath)
		if err != nil {
			logger.Println("failed to remove socket %s: %v", d.sockpath, err)
		}
		logger.Println("exiting")
		os.Exit(0)
	}()

	service := &Service{st}
	rpc.RegisterName(api.ServiceName, service)

	logger.Println("starting to serve RPC calls")
	rpc.Accept(listener)
	return 0
}
