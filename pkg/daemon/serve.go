package daemon

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"src.elv.sh/pkg/daemon/internal/api"
	"src.elv.sh/pkg/rpc"
	"src.elv.sh/pkg/store"
	"src.elv.sh/pkg/trace"
)

// Serve runs the daemon service, listening on the socket specified by sockpath
// and serving data from dbpath. It quits upon receiving SIGTERM, SIGINT or when
// all active clients have disconnected.
func Serve(sockpath, dbpath string) {
	trace.Printf(trace.Daemon, 0, "pid is %d", syscall.Getpid())
	trace.Printf(trace.Daemon, 0, "listening on %v", sockpath)
	listener, err := listen(sockpath)
	if err != nil {
		trace.Printf(trace.Daemon, 0, "failed to listen on %s: %v\naborting", sockpath, err)
		os.Exit(2)
	}

	st, err := store.NewStore(dbpath)
	if err != nil {
		trace.Printf(trace.Daemon, 0, "failed to create storage: %v\nserving anyway", err)
	}

	quitSignals := make(chan os.Signal)
	quitChan := make(chan struct{})
	signal.Notify(quitSignals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		select {
		case sig := <-quitSignals:
			trace.Printf(trace.Daemon, 0, "received signal %s", sig)
		case <-quitChan:
			trace.Printf(trace.Daemon, 0, "no active clients")
		}
		err := os.Remove(sockpath)
		if err != nil {
			trace.Printf(trace.Daemon, 0, "failed to remove socket %s: %v", sockpath, err)
		}
		err = st.Close()
		if err != nil {
			trace.Printf(trace.Daemon, 0, "failed to close storage: %v", err)
		}
		err = listener.Close()
		if err != nil {
			trace.Printf(trace.Daemon, 0, "failed to close listener: %v", err)
		}
		trace.Printf(trace.Daemon, 0, "listener closed, waiting to exit")
	}()

	svc := &service{st, err}
	rpc.RegisterName(api.ServiceName, svc)
	trace.Printf(trace.Daemon, 0, "starting to serve RPC calls")

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
			trace.Printf(trace.Daemon, 0, "failed to accept connection: %v", err)
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

	trace.Printf(trace.Daemon, 0, "exiting")
}
