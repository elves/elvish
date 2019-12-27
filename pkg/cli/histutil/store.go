package histutil

import (
	"errors"

	"github.com/elves/elvish/pkg/store"
)

var errStoreIsEmpty = errors.New("store is empty")

// Store is an abstract interface for history store.
type Store interface {
	// AddCmd adds a new command history entry and returns its sequence number.
	// Depending on the implementation, the Store might respect cmd.Seq and
	// return it as is, or allocate another sequence number.
	AddCmd(cmd store.Cmd) (int, error)
	// AllCmds returns all commands kept in the store.
	AllCmds() ([]store.Cmd, error)
	// LastCmd returns the last command in the store.
	LastCmd() (store.Cmd, error)
}

// NewMemoryStore returns a Store that stores command history in memory.
func NewMemoryStore() Store {
	return &memoryStore{}
}

type memoryStore struct{ cmds []store.Cmd }

func (s *memoryStore) AllCmds() ([]store.Cmd, error) {
	return s.cmds, nil
}

func (s *memoryStore) AddCmd(cmd store.Cmd) (int, error) {
	s.cmds = append(s.cmds, cmd)
	return cmd.Seq, nil
}

func (s *memoryStore) LastCmd() (store.Cmd, error) {
	if len(s.cmds) == 0 {
		return store.Cmd{}, errStoreIsEmpty
	}
	return s.cmds[len(s.cmds)-1], nil
}

// NewDBStore returns a Store backed by a database.
func NewDBStore(db DB) Store {
	return dbStore{db, -1}
}

// NewDBStoreFrozen returns a Store backed by a database, with the view of all
// commands frozen at creation.
func NewDBStoreFrozen(db DB) (Store, error) {
	upper, err := db.NextCmdSeq()
	if err != nil {
		return nil, err
	}
	return dbStore{db, upper}, nil
}

type dbStore struct {
	db    DB
	upper int
}

func (s dbStore) AllCmds() ([]store.Cmd, error) {
	// TODO: Return the actual command sequence in the DB. The DB currently
	// doesn't have an RPC method for that.
	cmds, err := s.db.Cmds(0, s.upper)
	if err != nil {
		return nil, err
	}
	entries := make([]store.Cmd, len(cmds))
	for i, cmd := range cmds {
		entries[i] = store.Cmd{cmd, i}
	}
	return entries, nil
}

func (s dbStore) AddCmd(cmd store.Cmd) (int, error) {
	return s.db.AddCmd(cmd.Text)
}

func (s dbStore) LastCmd() (store.Cmd, error) {
	return s.db.PrevCmd(s.upper, "")
}
