package history

import (
	"strings"
)

// Store is the interface of the storage backend.
type Store interface {
	PrevCmd(upto int, prefix string) (int, string, error)
}

// mockStore is an implementation of the Store interface that can be used for
// testing.
type mockStore struct {
	cmds  []string
	errAt int
	err   error
}

func (s *mockStore) PrevCmd(upto int, prefix string) (int, string, error) {
	if upto < 0 {
		upto = len(s.cmds)
	}
	for i := upto - 1; i >= 0; i-- {
		if s.err != nil && i == s.errAt {
			return -1, "", s.err
		}
		if strings.HasPrefix(s.cmds[i], prefix) {
			return i, s.cmds[i], nil
		}
	}
	return -1, "", ErrEndOfHistory
}
