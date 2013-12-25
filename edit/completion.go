package edit

type completion struct {
	candidates []string
	current int
}

func (c *completion) prev() {
	if c.current > 0 {
		c.current--
	}
}

func (c *completion) next() {
	if c.current < len(c.candidates) - 1 {
		c.current++
	}
}
