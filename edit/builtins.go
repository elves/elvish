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
	"kill-line-b": leKillLineB,
	"kill-line-f": leKillLineF,
	"kill-rune-b": leKillRuneB,
	"move-dot-b": leMoveDotB,
	"move-dot-f": leMoveDotF,
}

func leKillLineB(ed *Editor) {
	ed.line = ed.line[ed.dot:]
	ed.dot = 0
}

func leKillLineF(ed *Editor) {
	ed.line = ed.line[:ed.dot]
}

func leKillRuneB(ed *Editor) {
	if ed.dot > 0 {
		_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
		ed.line = ed.line[:ed.dot-w] + ed.line[ed.dot:]
		ed.dot -= w
	} else {
		ed.beep()
	}
}

func leMoveDotB(ed *Editor) {
	_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
	ed.dot -= w
}

func leMoveDotF(ed *Editor) {
	_, w := utf8.DecodeRuneInString(ed.line[ed.dot:])
	ed.dot += w
}
