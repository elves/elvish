// Package daemon implements a service for mediating access to the data store,
// and its client.
//
// Most RPCs exposed by the service correspond to the methods of Store in the
// store package and are not documented here.
package daemon

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"src.elv.sh/pkg/daemon/internal/api"
	"src.elv.sh/pkg/logutil"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/store"
)

var logger = logutil.GetLogger("[daemon] ")

// Program is the daemon subprogram.
var Program prog.Program = program{}

type program struct {
	ServeChans ServeChans
}

func (p program) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	if !f.Daemon {
		return prog.ErrNotSuitable
	}
	if len(args) > 0 {
		return prog.BadUsage("arguments are not allowed with -daemon")
	}
	setUmaskForDaemon()
	exit := Serve(f.Sock, f.DB, p.ServeChans)
	return prog.Exit(exit)
}

// ServeChans keeps channels that can be passed to Serve.
type ServeChans struct {
	// If not nil, will be closed when the daemon is ready to serve requests.
	Ready chan<- struct{}
	// Causes the daemon to abort if closed or sent any date. If nil, Serve will
	// set up its own signal channel by listening to SIGINT and SIGTERM.
	Signal <-chan os.Signal
}

// Serve runs the daemon service, listening on the socket specified by sockpath
// and serving data from dbpath until all clients have exited. See doc for
// ServeChans for additional options.
func Serve(sockpath, dbpath string, chans ServeChans) int {
	logger.Println("pid is", syscall.Getpid())
	logger.Println("going to listen", sockpath)
	listener, err := net.Listen("unix", sockpath)
	if err != nil {
		logger.Printf("failed to listen on %s: %v", sockpath, err)
		logger.Println("aborting")
		return 2
	}

	st, err := store.NewStore(dbpath)
	if err != nil {
		logger.Printf("failed to create storage: %v", err)
		logger.Printf("serving anyway")
	}

	server := rpc.NewServer()
	server.RegisterName(api.ServiceName, &service{st, err})

	connCh := make(chan net.Conn, 10)
	listenErrCh := make(chan error, 1)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				listenErrCh <- err
				close(listenErrCh)
				return
			}
			connCh <- conn
		}
	}()

	sigCh := chans.Signal
	if sigCh == nil {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		sigCh = ch
	}

	conns := make(map[net.Conn]struct{})
	connDoneCh := make(chan net.Conn, 10)

	interrupt := func() bool {
		if len(conns) == 0 {
			logger.Println("exiting since there are no clients")
			return true
		}
		logger.Printf("going to close %v active connections", len(conns))
		for conn := range conns {
			err := conn.Close()
			if err != nil {
				logger.Println("failed to close connection:", err)
			}
		}
		return false
	}

	if chans.Ready != nil {
		close(chans.Ready)
	}

loop:
	for {
		select {
		case sig := <-sigCh:
			logger.Printf("received signal %s", sig)
			if interrupt() {
				break loop
			}
		case err := <-listenErrCh:
			logger.Println("could not listen:", err)
			if len(conns) == 0 {
				logger.Println("exiting since there are no clients")
				break loop
			}
			logger.Println("continuing to serve until all existing clients exit")
		case conn := <-connCh:
			conns[conn] = struct{}{}
			go func() {
				server.ServeConn(conn)
				connDoneCh <- conn
			}()
		case conn := <-connDoneCh:
			delete(conns, conn)
			if len(conns) == 0 {
				logger.Println("all clients disconnected, exiting")
				break loop
			}
		}
	}

	err = os.Remove(sockpath)
	if err != nil {
		logger.Printf("failed to remove socket %s: %v", sockpath, err)
	}
	if st != nil {
		err = st.Close()
		if err != nil {
			logger.Printf("failed to close storage: %v", err)
		}
	}
	err = listener.Close()
	if err != nil {
		logger.Printf("failed to close listener: %v", err)
	}
	// Ensure that the listener goroutine has exited before returning
	<-listenErrCh
	return 0
}
