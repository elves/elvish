package edit

import "fmt"

type tokenPart struct {
	text      string
	completed bool
}

type candidate struct {
	text  string
	parts []tokenPart
	attr  string // Attribute used for preview
}

func newCandidate(tps ...tokenPart) *candidate {
	c := &candidate{}
	for _, tp := range tps {
		c.push(tp)
	}
	return c
}

func (c *candidate) push(tp tokenPart) {
	c.text += tp.text
	c.parts = append(c.parts, tp)
}

type completion struct {
	start, end int // The text to complete is Editor.line[start:end]
	typ        TokenType
	candidates []*candidate
	current    int
}

func (c *completion) prev(cycle bool) {
	c.current--
	if c.current == -1 {
		if cycle {
			c.current = len(c.candidates) - 1
		} else {
			c.current++
		}
	}
}

func (c *completion) next(cycle bool) {
	c.current++
	if c.current == len(c.candidates) {
		if cycle {
			c.current = 0
		} else {
			c.current--
		}
	}
}

func startCompletion(ed *Editor, k Key) *leReturn {
	token := tokenAtDot(ed)
	node := token.Node

	c := &completion{
		start: node.Begin(),
		end:   ed.dot,
		typ:   Bareword, // TODO set the actual type
	}
	for _, compl := range completers {
		candidates := compl.completer(node)
		if candidates != nil {
			c.candidates = candidates
			break
		}
	}

	if c.candidates == nil {
		ed.pushTip("unsupported completion :(")
	} else if len(c.candidates) == 0 {
		ed.pushTip(fmt.Sprintf("no completion for %s", token.Text))
	} else {
		ed.completion = c
		ed.mode = modeCompletion
	}
	return nil
}

var BadToken = Token{}

func tokenAtDot(ed *Editor) Token {
	if len(ed.tokens) == 0 || ed.dot > len(ed.line) {
		return BadToken
	}
	if ed.dot == len(ed.line) {
		return ed.tokens[len(ed.tokens)-1]
	}
	for _, token := range ed.tokens {
		if ed.dot < token.Node.End() {
			return token
		}
	}
	return BadToken
}
