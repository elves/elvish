package history

import (
	"sync"
)

// Fuser provides a unified view into a shared storage-backed command history
// and per-session history.
type Fuser struct {
	store      Store
	storeUpper int

	*sync.RWMutex

	// Per-session history.
	cmds []string
	seqs []int
}

func NewFuser(store Store) (*Fuser, error) {
	upper, err := store.NextCmdSeq()
	if err != nil {
		return nil, err
	}
	return &Fuser{
		store:      store,
		storeUpper: upper,
		RWMutex:    &sync.RWMutex{},
	}, nil
}

func (f *Fuser) FastForward() error {
	f.Lock()
	defer f.Unlock()

	upper, err := f.store.NextCmdSeq()
	if err != nil {
		return err
	}
	f.storeUpper = upper
	f.cmds = nil
	f.seqs = nil
	return nil
}

func (f *Fuser) AddCmd(cmd string) error {
	f.Lock()
	defer f.Unlock()
	seq, err := f.store.AddCmd(cmd)
	if err != nil {
		return err
	}
	f.cmds = append(f.cmds, cmd)
	f.seqs = append(f.seqs, seq)
	return nil
}

func (f *Fuser) AllCmds() ([]string, error) {
	f.RLock()
	defer f.RUnlock()
	cmds, err := f.store.Cmds(0, f.storeUpper)
	if err != nil {
		return nil, err
	}
	return append(cmds, f.cmds...), nil
}

func (f *Fuser) SessionCmds() []string {
	return f.cmds
}

func (f *Fuser) Walker(prefix string) *Walker {
	f.RLock()
	defer f.RUnlock()
	return NewWalker(f.store, f.storeUpper, f.cmds, f.seqs, prefix)
}
