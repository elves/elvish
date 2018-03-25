package daemon

import (
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/elves/elvish/store"
	"github.com/elves/elvish/store/storedefs"
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

	service := &Service{st, err}
	rpc.RegisterName(ServiceName, service)

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

// Service provides the daemon RPC service. It is suitable as a service for
// net/rpc.
type Service struct {
	store storedefs.Store
	err   error
}

// Implementations of RPC methods.

// Version returns the API version number.
func (s *Service) Version(req *VersionRequest, res *VersionResponse) error {
	if s.err != nil {
		return s.err
	}
	res.Version = Version
	return nil
}

// Pid returns the process ID of the daemon.
func (s *Service) Pid(req *PidRequest, res *PidResponse) error {
	res.Pid = syscall.Getpid()
	return nil
}

func (s *Service) NextCmdSeq(req *NextCmdSeqRequest, res *NextCmdSeqResponse) error {
	if s.err != nil {
		return s.err
	}
	seq, err := s.store.NextCmdSeq()
	res.Seq = seq
	return err
}

func (s *Service) AddCmd(req *AddCmdRequest, res *AddCmdResponse) error {
	if s.err != nil {
		return s.err
	}
	seq, err := s.store.AddCmd(req.Text)
	res.Seq = seq
	return err
}

func (s *Service) DelCmd(req *DelCmdRequest, res *DelCmdResponse) error {
	if s.err != nil {
		return s.err
	}
	err := s.store.DelCmd(req.Seq)
	return err
}

func (s *Service) Cmd(req *CmdRequest, res *CmdResponse) error {
	if s.err != nil {
		return s.err
	}
	text, err := s.store.Cmd(req.Seq)
	res.Text = text
	return err
}

func (s *Service) Cmds(req *CmdsRequest, res *CmdsResponse) error {
	if s.err != nil {
		return s.err
	}
	cmds, err := s.store.Cmds(req.From, req.Upto)
	res.Cmds = cmds
	return err
}

func (s *Service) NextCmd(req *NextCmdRequest, res *NextCmdResponse) error {
	if s.err != nil {
		return s.err
	}
	seq, text, err := s.store.NextCmd(req.From, req.Prefix)
	res.Seq, res.Text = seq, text
	return err
}

func (s *Service) PrevCmd(req *PrevCmdRequest, res *PrevCmdResponse) error {
	if s.err != nil {
		return s.err
	}
	seq, text, err := s.store.PrevCmd(req.Upto, req.Prefix)
	res.Seq, res.Text = seq, text
	return err
}

func (s *Service) AddDir(req *AddDirRequest, res *AddDirResponse) error {
	if s.err != nil {
		return s.err
	}
	return s.store.AddDir(req.Dir, req.IncFactor)
}

func (s *Service) DelDir(req *DelDirRequest, res *DelDirResponse) error {
	if s.err != nil {
		return s.err
	}
	return s.store.DelDir(req.Dir)
}

func (s *Service) Dirs(req *DirsRequest, res *DirsResponse) error {
	if s.err != nil {
		return s.err
	}
	dirs, err := s.store.Dirs(req.Blacklist)
	res.Dirs = dirs
	return err
}

func (s *Service) SharedVar(req *SharedVarRequest, res *SharedVarResponse) error {
	if s.err != nil {
		return s.err
	}
	value, err := s.store.SharedVar(req.Name)
	res.Value = value
	return err
}

func (s *Service) SetSharedVar(req *SetSharedVarRequest, res *SetSharedVarResponse) error {
	if s.err != nil {
		return s.err
	}
	return s.store.SetSharedVar(req.Name, req.Value)
}

func (s *Service) DelSharedVar(req *DelSharedVarRequest, res *DelSharedVarResponse) error {
	if s.err != nil {
		return s.err
	}
	return s.store.DelSharedVar(req.Name)
}
