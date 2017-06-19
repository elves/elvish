// Package service implements the daemon service for mediating access to the
// storage backend.
package service

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

// Service provides the daemon RPC service.
type Service struct {
	store *store.Store
}

func (s *Service) Version(req *api.VersionRequest, res *api.VersionResponse) error {
	res.Version = api.Version
	return nil
}

func (s *Service) Pid(req *api.PidRequest, res *api.PidResponse) error {
	res.Pid = syscall.Getpid()
	return nil
}

func (s *Service) NextCmdSeq(req *api.NextCmdSeqRequest, res *api.NextCmdSeqResponse) error {
	seq, err := s.store.NextCmdSeq()
	res.Seq = seq
	return err
}

func (s *Service) AddCmd(req *api.AddCmdRequest, res *api.AddCmdResponse) error {
	seq, err := s.store.AddCmd(req.Text)
	res.Seq = seq
	return err
}

func (s *Service) Cmd(req *api.CmdRequest, res *api.CmdResponse) error {
	text, err := s.store.Cmd(req.Seq)
	res.Text = text
	return err
}

func (s *Service) Cmds(req *api.CmdsRequest, res *api.CmdsResponse) error {
	cmds, err := s.store.Cmds(req.From, req.Upto)
	res.Cmds = cmds
	return err
}

func (s *Service) NextCmd(req *api.NextCmdRequest, res *api.NextCmdResponse) error {
	seq, text, err := s.store.NextCmd(req.From, req.Prefix)
	res.Seq, res.Text = seq, text
	return err
}

func (s *Service) PrevCmd(req *api.PrevCmdRequest, res *api.PrevCmdResponse) error {
	seq, text, err := s.store.PrevCmd(req.Upto, req.Prefix)
	res.Seq, res.Text = seq, text
	return err
}

func (s *Service) Dirs(req *api.DirsRequest, res *api.DirsResponse) error {
	dirs, err := s.store.GetDirs(req.Blacklist)
	res.Dirs = dirs
	return err
}
