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
	typ parse.ItemType
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

func findCandidates(p string, all []string) (cands []*candidate) {
	// Prefix match
	for _, s := range all {
		if len(s) >= len(p) && s[:len(p)] == p {
			cand := newCandidate()
			cand.push(tokenPart{p, false})
			cand.push(tokenPart{s[len(p):], true})
			cands = append(cands, cand)
		}
	}
	return
}

func fileNames(dir string) (names []string, err error) {
	infos, e := ioutil.ReadDir(".")
	if e != nil {
		err = e
		return
	}
	for _, info := range infos {
		names = append(names, info.Name())
	}
	return
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
	pattern := lastToken.Val
	c.start = ed.dot - len(pattern)
	c.end = ed.dot
	c.typ = lastToken.Typ

	names, err := fileNames(".")
	if err != nil {
		ed.pushTip(err.Error())
		return
	}
	c.candidates = findCandidates(pattern, names)
	ed.completion = c
}
