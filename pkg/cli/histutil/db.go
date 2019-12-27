package histutil

import (
	"strings"

	"github.com/elves/elvish/pkg/store"
)

// DB is the interface of the storage database.
type DB interface {
	NextCmdSeq() (int, error)
	AddCmd(cmd string) (int, error)
	CmdsWithSeq(from, upto int) ([]store.Cmd, error)
	PrevCmd(upto int, prefix string) (store.Cmd, error)
}

// TestDB is an implementation of the DB interface that can be used for testing.
type TestDB struct {
	AllCmds []string

	OneOffError error
}

func (s *TestDB) error() error {
	err := s.OneOffError
	s.OneOffError = nil
	return err
}

func (s *TestDB) NextCmdSeq() (int, error) {
	return len(s.AllCmds), s.error()
}

func (s *TestDB) AddCmd(cmd string) (int, error) {
	if s.OneOffError != nil {
		return -1, s.error()
	}
	s.AllCmds = append(s.AllCmds, cmd)
	return len(s.AllCmds) - 1, nil
}

func (s *TestDB) CmdsWithSeq(from, upto int) ([]store.Cmd, error) {
	var cmds []store.Cmd
	for i := from; i < upto; i++ {
		cmds = append(cmds, store.Cmd{Text: s.AllCmds[i], Seq: i})
	}
	return cmds, s.error()
}

func (s *TestDB) PrevCmd(upto int, prefix string) (store.Cmd, error) {
	if s.OneOffError != nil {
		return store.Cmd{}, s.error()
	}
	if upto < 0 || upto > len(s.AllCmds) {
		upto = len(s.AllCmds)
	}
	for i := upto - 1; i >= 0; i-- {
		if strings.HasPrefix(s.AllCmds[i], prefix) {
			return store.Cmd{Text: s.AllCmds[i], Seq: i}, nil
		}
	}
	return store.Cmd{}, ErrEndOfHistory
}
