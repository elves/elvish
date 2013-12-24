package edit

import (
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
