package edit

import (
	"fmt"
	"unicode/utf8"
)

// Line editor builtins.
// These are not exposed to the user in anyway yet. Ideally, they should
// reside in a dedicated namespace and callable by users, e.g.
// le/kill-line-forward.

type editorAction int

const (
	noAction editorAction = iota
	changeMode
	changeModeAndReprocess
	exitReadLine
)

type leReturn struct {
	action editorAction
	newMode bufferMode
	readLineReturn LineRead
}

type leBuiltin func (ed *Editor) *leReturn

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

func killLineB(ed *Editor) *leReturn {
	ed.line = ed.line[ed.dot:]
	ed.dot = 0
	return nil
}

func killLineF(ed *Editor) *leReturn {
	ed.line = ed.line[:ed.dot]
	return nil
}

func killRuneB(ed *Editor) *leReturn {
	if ed.dot > 0 {
		_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
		ed.line = ed.line[:ed.dot-w] + ed.line[ed.dot:]
		ed.dot -= w
	} else {
		ed.beep()
	}
	return nil
}

func moveDotB(ed *Editor) *leReturn {
	_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
	ed.dot -= w
	return nil
}

func moveDotF(ed *Editor) *leReturn {
	_, w := utf8.DecodeRuneInString(ed.line[ed.dot:])
	ed.dot += w
	return nil
}

func complete(ed *Editor) *leReturn {
	startCompletion(ed)
	c := ed.completion
	if c == nil {
		return nil
	}
	if len(c.candidates) == 0 {
		ed.pushTip(fmt.Sprintf("No completion for %s", ed.line[c.start:c.end]))
		ed.completion = nil
	} else {
		c.current = 0
	}
	return nil
}

func selectCandB(ed *Editor) *leReturn {
	ed.completion.prev(false)
	return nil
}

func selectCandF(ed *Editor) *leReturn {
	ed.completion.next(false)
	return nil
}

func cycleCandF(ed *Editor) *leReturn {
	ed.completion.next(true)
	return nil
}

func cancelCompletion(ed *Editor) *leReturn {
	ed.completion = nil
	ed.mode = ModeInsert
	return nil
}
