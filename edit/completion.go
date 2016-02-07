package edit

import (
	"fmt"
	"unicode/utf8"
)

type styled struct {
	text  string
	style string
}

type candidate struct {
	source, menu styled
}

type completion struct {
	completer  string
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

func startCompletion(ed *Editor) {
	startCompletionInner(ed, false)
}

func completePrefixOrStartCompletion(ed *Editor) {
	startCompletionInner(ed, true)
}

func startCompletionInner(ed *Editor, completePrefix bool) {
	token := tokenAtDot(ed)
	node := token.Node
	if node == nil {
		return
	}

	c := &completion{}
	for _, compl := range completers {
		candidates := compl.completer(node, ed)
		if candidates != nil {
			c.completer = compl.name
			c.candidates = candidates
			break
		}
	}

	if c.candidates == nil {
		ed.pushTip("unsupported completion :(")
	} else if len(c.candidates) == 0 {
		ed.pushTip(fmt.Sprintf("no candidate for %s", c.completer))
	} else {
		if completePrefix {
			// If there is a non-empty longest common prefix, insert it and
			// don't start completion mode.
			// As a special case, when there is exactly one candidate, it is
			// immeidately accepted.
			prefix := c.candidates[0].source.text
			for _, cand := range c.candidates[1:] {
				prefix = commonPrefix(prefix, cand.source.text)
			}
			if prefix != "" {
				ed.insertAtDot(prefix)
				return
			}
		}
		ed.completion = c
		ed.mode = modeCompletion
	}
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

func commonPrefix(s, t string) string {
	for i, r := range s {
		if i >= len(t) {
			return s[:i]
		}
		r2, _ := utf8.DecodeRuneInString(t[i:])
		if r2 != r {
			return s[:i]
		}
	}
	return s
}
