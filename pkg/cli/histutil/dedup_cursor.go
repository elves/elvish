package histutil

import "src.elv.sh/pkg/store/storedefs"

// NewDedupCursor returns a cursor that skips over all duplicate entries.
func NewDedupCursor(c Cursor) Cursor {
	return &dedupCursor{c, 0, nil, make(map[string]bool)}
}

type dedupCursor struct {
	c       Cursor
	current int
	stack   []storedefs.Cmd
	occ     map[string]bool
}

func (c *dedupCursor) Prev() {
	if c.current < len(c.stack)-1 {
		c.current++
		return
	}
	for {
		c.c.Prev()
		cmd, err := c.c.Get()
		if err != nil {
			c.current = len(c.stack)
			break
		}
		if !c.occ[cmd.Text] {
			c.current = len(c.stack)
			c.stack = append(c.stack, cmd)
			c.occ[cmd.Text] = true
			break
		}
	}
}

func (c *dedupCursor) Next() {
	if c.current >= 0 {
		c.current--
	}
}

func (c *dedupCursor) Get() (storedefs.Cmd, error) {
	switch {
	case c.current < 0:
		return storedefs.Cmd{}, ErrEndOfHistory
	case c.current < len(c.stack):
		return c.stack[c.current], nil
	default:
		return c.c.Get()
	}
}
