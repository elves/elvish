package histutil

// Store is an abstract interface for history store.
type Store interface {
	// AddCmd adds a new command history entry and returns its sequence number.
	// Depending on the implementation, the Store might respect cmd.Seq and
	// return it as is, or allocate another sequence number.
	AddCmd(cmd Entry) (int, error)
	// AllCmds returns a commands kept in the store.
	AllCmds() ([]Entry, error)
}

// Entry represents a command history item.
type Entry struct {
	Text string
	Seq  int
}

// NewMemoryStore returns a Store that stores command history in memory.
func NewMemoryStore() Store {
	return &memoryStore{}
}

type memoryStore struct{ cmds []Entry }

func (s *memoryStore) AllCmds() ([]Entry, error) {
	return s.cmds, nil
}

func (s *memoryStore) AddCmd(cmd Entry) (int, error) {
	s.cmds = append(s.cmds, cmd)
	return cmd.Seq, nil
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

func (s dbStore) AllCmds() ([]Entry, error) {
	// TODO: Return the actual command sequence in the DB. The DB currently
	// doesn't have an RPC method for that.
	cmds, err := s.db.Cmds(-1, s.upper)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, len(cmds))
	for i, cmd := range cmds {
		entries[i] = Entry{cmd, i}
	}
	return entries, nil
}

func (s dbStore) AddCmd(cmd Entry) (int, error) {
	return s.db.AddCmd(cmd.Text)
}
