package histutil

import (
	"src.elv.sh/pkg/store/storedefs"
)

// NewDBStore returns a Store backed by a database with the view of all
// commands frozen at creation.
func NewDBStore(db DB) (Store, error) {
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

func (s dbStore) AllCmds() ([]storedefs.Cmd, error) {
	return s.db.CmdsWithSeq(0, s.upper)
}

func (s dbStore) AddCmd(cmd storedefs.Cmd) (int, error) {
	return s.db.AddCmd(cmd.Text)
}

func (s dbStore) Cursor(prefix string) Cursor {
	return &dbStoreCursor{
		s.db, prefix, s.upper, storedefs.Cmd{Seq: s.upper}, ErrEndOfHistory}
}

type dbStoreCursor struct {
	db     DB
	prefix string
	upper  int
	cmd    storedefs.Cmd
	err    error
}

func (c *dbStoreCursor) Prev() {
	if c.cmd.Seq < 0 {
		return
	}
	cmd, err := c.db.PrevCmd(c.cmd.Seq, c.prefix)
	c.set(cmd, err, -1)
}

func (c *dbStoreCursor) Next() {
	if c.cmd.Seq >= c.upper {
		return
	}
	cmd, err := c.db.NextCmd(c.cmd.Seq+1, c.prefix)
	if cmd.Seq < c.upper {
		c.set(cmd, err, c.upper)
	}
	if cmd.Seq >= c.upper {
		c.cmd = storedefs.Cmd{Seq: c.upper}
		c.err = ErrEndOfHistory
	}
}

func (c *dbStoreCursor) set(cmd storedefs.Cmd, err error, endSeq int) {
	if err == nil {
		c.cmd = cmd
		c.err = nil
	} else if err.Error() == storedefs.ErrNoMatchingCmd.Error() {
		c.cmd = storedefs.Cmd{Seq: endSeq}
		c.err = ErrEndOfHistory
	} else {
		// Don't change c.cmd
		c.err = err
	}
}

func (c *dbStoreCursor) Get() (storedefs.Cmd, error) {
	return c.cmd, c.err
}
