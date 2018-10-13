package daemon

import (
	"syscall"

	"github.com/elves/elvish/store/storedefs"
)

// A net/rpc service for the daemon.
type service struct {
	store storedefs.Store
	err   error
}

// Implementations of RPC methods.

// Version returns the API version number.
func (s *service) Version(req *VersionRequest, res *VersionResponse) error {
	if s.err != nil {
		return s.err
	}
	res.Version = Version
	return nil
}

// Pid returns the process ID of the daemon.
func (s *service) Pid(req *PidRequest, res *PidResponse) error {
	res.Pid = syscall.Getpid()
	return nil
}

func (s *service) NextCmdSeq(req *NextCmdSeqRequest, res *NextCmdSeqResponse) error {
	if s.err != nil {
		return s.err
	}
	seq, err := s.store.NextCmdSeq()
	res.Seq = seq
	return err
}

func (s *service) AddCmd(req *AddCmdRequest, res *AddCmdResponse) error {
	if s.err != nil {
		return s.err
	}
	seq, err := s.store.AddCmd(req.Text)
	res.Seq = seq
	return err
}

func (s *service) DelCmd(req *DelCmdRequest, res *DelCmdResponse) error {
	if s.err != nil {
		return s.err
	}
	err := s.store.DelCmd(req.Seq)
	return err
}

func (s *service) Cmd(req *CmdRequest, res *CmdResponse) error {
	if s.err != nil {
		return s.err
	}
	text, err := s.store.Cmd(req.Seq)
	res.Text = text
	return err
}

func (s *service) Cmds(req *CmdsRequest, res *CmdsResponse) error {
	if s.err != nil {
		return s.err
	}
	cmds, err := s.store.Cmds(req.From, req.Upto)
	res.Cmds = cmds
	return err
}

func (s *service) NextCmd(req *NextCmdRequest, res *NextCmdResponse) error {
	if s.err != nil {
		return s.err
	}
	seq, text, err := s.store.NextCmd(req.From, req.Prefix)
	res.Seq, res.Text = seq, text
	return err
}

func (s *service) PrevCmd(req *PrevCmdRequest, res *PrevCmdResponse) error {
	if s.err != nil {
		return s.err
	}
	seq, text, err := s.store.PrevCmd(req.Upto, req.Prefix)
	res.Seq, res.Text = seq, text
	return err
}

func (s *service) AddDir(req *AddDirRequest, res *AddDirResponse) error {
	if s.err != nil {
		return s.err
	}
	return s.store.AddDir(req.Dir, req.IncFactor)
}

func (s *service) DelDir(req *DelDirRequest, res *DelDirResponse) error {
	if s.err != nil {
		return s.err
	}
	return s.store.DelDir(req.Dir)
}

func (s *service) Dirs(req *DirsRequest, res *DirsResponse) error {
	if s.err != nil {
		return s.err
	}
	dirs, err := s.store.Dirs(req.Blacklist)
	res.Dirs = dirs
	return err
}

func (s *service) SharedVar(req *SharedVarRequest, res *SharedVarResponse) error {
	if s.err != nil {
		return s.err
	}
	value, err := s.store.SharedVar(req.Name)
	res.Value = value
	return err
}

func (s *service) SetSharedVar(req *SetSharedVarRequest, res *SetSharedVarResponse) error {
	if s.err != nil {
		return s.err
	}
	return s.store.SetSharedVar(req.Name, req.Value)
}

func (s *service) DelSharedVar(req *DelSharedVarRequest, res *DelSharedVarResponse) error {
	if s.err != nil {
		return s.err
	}
	return s.store.DelSharedVar(req.Name)
}
