package testutil

type cleanuper struct{ fns []func() }

func (c *cleanuper) Cleanup(fn func()) { c.fns = append(c.fns, fn) }

func (c *cleanuper) runCleanups() {
	for i := len(c.fns) - 1; i >= 0; i-- {
		c.fns[i]()
	}
}
