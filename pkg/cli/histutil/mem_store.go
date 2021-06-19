package histutil

import (
	"strings"

	"src.elv.sh/pkg/store/storedefs"
)

// NewMemStore returns a Store that stores command history in memory.
func NewMemStore(texts ...string) Store {
	cmds := make([]storedefs.Cmd, len(texts))
	for i, text := range texts {
		cmds[i] = storedefs.Cmd{Text: text, Seq: i}
	}
	return &memStore{cmds}
}

type memStore struct{ cmds []storedefs.Cmd }

func (s *memStore) AllCmds() ([]storedefs.Cmd, error) {
	return s.cmds, nil
}

func (s *memStore) AddCmd(cmd storedefs.Cmd) (int, error) {
	if cmd.Seq < 0 {
		cmd.Seq = len(s.cmds) + 1
	}
	s.cmds = append(s.cmds, cmd)
	return cmd.Seq, nil
}

func (s *memStore) Cursor(prefix string) Cursor {
	return &memStoreCursor{s.cmds, prefix, len(s.cmds)}
}

type memStoreCursor struct {
	cmds   []storedefs.Cmd
	prefix string
	index  int
}

func (c *memStoreCursor) Prev() {
	if c.index < 0 {
		return
	}
	for c.index--; c.index >= 0; c.index-- {
		if strings.HasPrefix(c.cmds[c.index].Text, c.prefix) {
			return
		}
	}
}

func (c *memStoreCursor) Next() {
	if c.index >= len(c.cmds) {
		return
	}
	for c.index++; c.index < len(c.cmds); c.index++ {
		if strings.HasPrefix(c.cmds[c.index].Text, c.prefix) {
			return
		}
	}
}

func (c *memStoreCursor) Get() (storedefs.Cmd, error) {
	if c.index < 0 || c.index >= len(c.cmds) {
		return storedefs.Cmd{}, ErrEndOfHistory
	}
	return c.cmds[c.index], nil
}
