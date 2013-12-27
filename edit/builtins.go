package edit

import (
	"fmt"
	"unicode/utf8"
)

// Line editor builtins.
// These are not exposed to the user in anyway yet. Ideally, they should
// reside in a dedicated namespace and callable by users, e.g.
// le/kill-line-forward.

type leBuiltin func (ed *Editor)

var leBuiltins = map[string]leBuiltin{
	"kill-line-b": killLineB,
	"kill-line-f": killLineF,
	"kill-rune-b": killRuneB,
	"move-dot-b": moveDotB,
	"move-dot-f": moveDotF,
	"complete": complete,
	"cancel-completion": cancelCompletion,
	"select-cand-b": selectCandB,
	"select-cand-f": selectCandF,
	"cycle-cand-f": cycleCandF,
}

func killLineB(ed *Editor) {
	ed.line = ed.line[ed.dot:]
	ed.dot = 0
}

func killLineF(ed *Editor) {
	ed.line = ed.line[:ed.dot]
}

func killRuneB(ed *Editor) {
	if ed.dot > 0 {
		_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
		ed.line = ed.line[:ed.dot-w] + ed.line[ed.dot:]
		ed.dot -= w
	} else {
		ed.beep()
	}
}

func moveDotB(ed *Editor) {
	_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
	ed.dot -= w
}

func moveDotF(ed *Editor) {
	_, w := utf8.DecodeRuneInString(ed.line[ed.dot:])
	ed.dot += w
}

func complete(ed *Editor) {
	startCompletion(ed)
	c := ed.completion
	if c == nil {
		return
	}
	if len(c.candidates) == 0 {
		ed.pushTip(fmt.Sprintf("No completion for %s", ed.line[c.start:c.end]))
		ed.completion = nil
	} else {
		c.current = 0
	}
}

func selectCandB(ed *Editor) {
	ed.completion.prev(false)
}

func selectCandF(ed *Editor) {
	ed.completion.next(false)
}

func cycleCandF(ed *Editor) {
	ed.completion.next(true)
}

func cancelCompletion(ed *Editor) {
	ed.completion = nil
}
