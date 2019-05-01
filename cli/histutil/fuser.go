package histutil

import (
	"sync"
)

// Fuser provides a view of command history that is fused from the shared
// storage-backed command history and per-session history.
type Fuser struct {
	mutex sync.RWMutex

	shared  Store
	session Store

	// Only used in FastForward.
	db DB
}

// NewFuser returns a new Fuser from a database.
func NewFuser(db DB) (*Fuser, error) {
	shared, session, err := initStores(db)
	if err != nil {
		return nil, err
	}
	return &Fuser{shared: shared, session: session, db: db}, nil
}

func initStores(db DB) (shared, session Store, err error) {
	shared, err = NewDBStoreFrozen(db)
	if err != nil {
		return nil, nil, err
	}
	return shared, NewMemoryStore(), nil
}

// FastForward fast-forwards the view of command history, so that commands added
// by other sessions since the start of the current session are available.
func (f *Fuser) FastForward() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	shared, session, err := initStores(f.db)
	if err != nil {
		return err
	}
	f.shared, f.session = shared, session
	return nil
}

// AddCmd adds a command to both the database and the per-session history.
func (f *Fuser) AddCmd(cmd string) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	seq, err := f.shared.AddCmd(Entry{Text: cmd})
	if err != nil {
		return -1, err
	}
	f.session.AddCmd(Entry{Text: cmd, Seq: seq})
	return seq, nil
}

// AllCmds returns all visible commands, consisting of commands that were
// already in the database at startup, plus the per-session history.
func (f *Fuser) AllCmds() ([]Entry, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	sharedCmds, err := f.shared.AllCmds()
	if err != nil {
		return nil, err
	}
	sessionCmds, _ := f.session.AllCmds()
	return append(sharedCmds, sessionCmds...), nil
}

// SessionCmds returns the per-session history.
func (f *Fuser) SessionCmds() []Entry {
	f.session.AllCmds()
	cmds, _ := f.session.AllCmds()
	return cmds
}

// Walker returns a walker for the fused command history.
func (f *Fuser) Walker(prefix string) Walker {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	sessionCmds, _ := f.session.AllCmds()
	// TODO: Avoid the type cast.
	return NewWalker(f.db, f.shared.(dbStore).upper, sessionCmds, prefix)
}
