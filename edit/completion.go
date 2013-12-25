package edit

import (
	"../parse"
)

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

func findCandidates(text string) (candidates []string) {
	// Find last token
	l := parse.Lex("<completion>", text)
	var lastToken parse.Item
	for token := range l.Chan() {
		if token.Typ != parse.ItemEOF {
			lastToken = token
		}
	}
	candidates = append(candidates, lastToken.String())
	return
}
