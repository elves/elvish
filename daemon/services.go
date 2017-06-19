package daemon

import (
	"syscall"

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/store"
)

// Service provides the daemon RPC service.
type Service struct {
	store *store.Store
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
