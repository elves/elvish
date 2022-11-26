// Package daemon implements a service for mediating access to the data store,
// and its client.
//
// Most RPCs exposed by the service correspond to the methods of Store in the
// store package and are not documented here.
package daemon

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"src.elv.sh/pkg/daemon/internal/api"
	"src.elv.sh/pkg/logutil"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/rpc"
	"src.elv.sh/pkg/store"
)

var logger = logutil.GetLogger("[daemon] ")

// Program is the daemon subprogram.
type Program struct {
	run   bool
	paths *prog.DaemonPaths
	// Used in tests.
	serveOpts ServeOpts
}

func (p *Program) RegisterFlags(fs *prog.FlagSet) {
	fs.BoolVar(&p.run, "daemon", false,
		"[internal flag] Run the storage daemon instead of an Elvish shell")
	p.paths = fs.DaemonPaths()
}

func (p *Program) Run(fds [3]*os.File, args []string) error {
	if !p.run {
		return prog.NextProgram()
	}
	if len(args) > 0 {
		return prog.BadUsage("arguments are not allowed with -daemon")
	}

	// The stdout is redirected to a unique log file (see the spawn function),
	// so just use it for logging.
	logutil.SetOutput(fds[1])
	setUmaskForDaemon()
	exit := Serve(p.paths.Sock, p.paths.DB, p.serveOpts)
	return prog.Exit(exit)
}

// ServeOpts keeps options that can be passed to Serve.
type ServeOpts struct {
	// If not nil, will be closed when the daemon is ready to serve requests.
	Ready chan<- struct{}
	// Causes the daemon to abort if closed or sent any date. If nil, Serve will
	// set up its own signal channel by listening to SIGINT and SIGTERM.
	Signals <-chan os.Signal
	// If not nil, overrides the response of the Version RPC.
	Version *int
}

// Serve runs the daemon service, listening on the socket specified by sockpath
// and serving data from dbpath until all clients have exited. See doc for
// ServeOpts for additional options.
func Serve(sockpath, dbpath string, opts ServeOpts) int {
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
	version := api.Version
	if opts.Version != nil {
		version = *opts.Version
	}
	server.RegisterName(api.ServiceName, &service{version, st, err})

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

	sigCh := opts.Signals
	if sigCh == nil {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		sigCh = ch
	}

	conns := make(map[net.Conn]struct{})
	connDoneCh := make(chan net.Conn, 10)

	interrupt := func() {
		if len(conns) == 0 {
			logger.Println("exiting since there are no clients")
		}
		logger.Printf("going to close %v active connections", len(conns))
		for conn := range conns {
			// Ignore the error - if we can't close the connection it's because
			// the client has closed it. There is nothing we can do anyway.
			conn.Close()
		}
	}

	if opts.Ready != nil {
		close(opts.Ready)
	}

loop:
	for {
		select {
		case sig := <-sigCh:
			logger.Printf("received signal %v", sig)
			interrupt()
			break loop
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
