package histutil

import "src.elv.sh/pkg/store/storedefs"

// NewHybridStore returns a store that provides a view of all the commands that
// exists in the database, plus a in-memory session history.
func NewHybridStore(db DB) (Store, error) {
	if db == nil {
		return NewMemStore(), nil
	}
	dbStore, err := NewDBStore(db)
	if err != nil {
		return NewMemStore(), err
	}
	return hybridStore{dbStore, NewMemStore()}, nil
}

type hybridStore struct {
	shared, session Store
}

func (s hybridStore) AddCmd(cmd storedefs.Cmd) (int, error) {
	seq, err := s.shared.AddCmd(cmd)
	s.session.AddCmd(storedefs.Cmd{Text: cmd.Text, Seq: seq})
	return seq, err
}

func (s hybridStore) AllCmds() ([]storedefs.Cmd, error) {
	shared, err := s.shared.AllCmds()
	session, err2 := s.session.AllCmds()
	if err == nil {
		err = err2
	}
	if len(shared) == 0 {
		return session, err
	}
	return append(shared, session...), err
}

func (s hybridStore) Cursor(prefix string) Cursor {
	return &hybridStoreCursor{
		s.shared.Cursor(prefix), s.session.Cursor(prefix), false}
}

type hybridStoreCursor struct {
	shared    Cursor
	session   Cursor
	useShared bool
}

func (c *hybridStoreCursor) Prev() {
	if c.useShared {
		c.shared.Prev()
		return
	}
	c.session.Prev()
	if _, err := c.session.Get(); err == ErrEndOfHistory {
		c.useShared = true
		c.shared.Prev()
	}
}

func (c *hybridStoreCursor) Next() {
	if !c.useShared {
		c.session.Next()
		return
	}
	c.shared.Next()
	if _, err := c.shared.Get(); err == ErrEndOfHistory {
		c.useShared = false
		c.session.Next()
	}
}

func (c *hybridStoreCursor) Get() (storedefs.Cmd, error) {
	if c.useShared {
		return c.shared.Get()
	}
	return c.session.Get()
}
