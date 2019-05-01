package histutil

import "strings"

// An implementation of the Store interface that can be used for testing.
type testDB struct {
	cmds []string

	oneOffError error
}

func (s *testDB) error() error {
	err := s.oneOffError
	s.oneOffError = nil
	return err
}

func (s *testDB) NextCmdSeq() (int, error) {
	return len(s.cmds), s.error()
}

func (s *testDB) AddCmd(cmd string) (int, error) {
	if s.oneOffError != nil {
		return -1, s.error()
	}
	s.cmds = append(s.cmds, cmd)
	return len(s.cmds) - 1, nil
}

func (s *testDB) Cmds(from, upto int) ([]string, error) {
	return s.cmds[from:upto], s.error()
}

func (s *testDB) PrevCmd(upto int, prefix string) (int, string, error) {
	if s.oneOffError != nil {
		return -1, "", s.error()
	}
	if upto < 0 || upto > len(s.cmds) {
		upto = len(s.cmds)
	}
	for i := upto - 1; i >= 0; i-- {
		if strings.HasPrefix(s.cmds[i], prefix) {
			return i, s.cmds[i], nil
		}
	}
	return -1, "", ErrEndOfHistory
}
