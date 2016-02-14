package edit

import "github.com/elves/elvish/store"

// Command history listing subosytem.

type historyListing struct {
	all []string
}

func newHistoryListing(s *store.Store) (*historyListing, error) {
	seq, err := s.NextCmdSeq()
	if err != nil {
		return nil, err
	}
	cmds, err := s.Cmds(seq-100, seq)
	if err != nil {
		return nil, err
	}
	return &historyListing{cmds}, nil
}
