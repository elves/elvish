package daemon

import (
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/elves/elvish/store"
)

// Serve runs the daemon service, listening on the socket specified by sockpath
// and serving data from dbpath. It quits upon receiving SIGTERM, SIGINT or when
// all active clients have disconnected.
func Serve(sockpath, dbpath string) {
	logger.Println("pid is", syscall.Getpid())
	logger.Println("going to listen", sockpath)
	listener, err := listen(sockpath)
	if err != nil {
		logger.Printf("failed to listen on %s: %v", sockpath, err)
		logger.Println("aborting")
		os.Exit(2)
	}

	st, err := store.NewStore(dbpath)
	if err != nil {
		logger.Printf("failed to create storage: %v", err)
		logger.Printf("serving anyway")
	}

	quitSignals := make(chan os.Signal)
	quitChan := make(chan struct{})
	signal.Notify(quitSignals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		select {
		case sig := <-quitSignals:
			logger.Printf("received signal %s", sig)
		case <-quitChan:
			logger.Printf("No active client, daemon exit")
		}
		err := os.Remove(sockpath)
		if err != nil {
			logger.Printf("failed to remove socket %s: %v", sockpath, err)
		}
		err = st.Close()
		if err != nil {
			logger.Printf("failed to close storage: %v", err)
		}
		err = listener.Close()
		if err != nil {
			logger.Printf("failed to close listener: %v", err)
		}
		logger.Println("listener closed, waiting to exit")
	}()

	svc := &service{st, err}
	rpc.RegisterName(ServiceName, svc)

	logger.Println("starting to serve RPC calls")

	firstClient := true
	activeClient := sync.WaitGroup{}
	// prevent daemon exit before serving first client
	activeClient.Add(1)
	go func() {
		activeClient.Wait()
		close(quitChan)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Printf("Failed to accept: %#v", err)
			break
		}

		if firstClient {
			firstClient = false
		} else {
			activeClient.Add(1)
		}
		go func() {
			rpc.DefaultServer.ServeConn(conn)
			activeClient.Done()
		}()
	}

	logger.Println("exiting")
}
