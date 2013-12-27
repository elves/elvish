package edit

import (
	"io/ioutil"
	"../parse"
)

type tokenPart struct {
	text string
	completed bool
}

type candidate struct {
	text string
	parts []tokenPart
}

func newCandidate() *candidate {
	return &candidate{}
}

func (c *candidate) push(tp tokenPart) {
	c.text += tp.text
	c.parts = append(c.parts, tp)
}

type completion struct {
	start, end int // The text to complete is Editor.line[start:end]
	candidates []*candidate
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

func startCompletion(ed *Editor) {
	c := &completion{current: -1}
	// Find last token
	l := parse.Lex("<completion>", ed.line[:ed.dot])
	var lastToken parse.Item
	for token := range l.Chan() {
		if token.Typ != parse.ItemEOF {
			lastToken = token
		}
	}
	prefix := lastToken.Val
	c.start = ed.dot - len(prefix)
	c.end = ed.dot

	infos, err := ioutil.ReadDir(".")
	if err != nil {
		return
	}
	for _, info := range infos {
		name := info.Name()
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			cand := newCandidate()
			cand.push(tokenPart{prefix, false})
			cand.push(tokenPart{name[len(prefix):], true})
			c.candidates = append(c.candidates, cand)
		}
	}
	ed.completion = c
}
