package edit

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/strutil"
)

// Line editor builtins.

type action struct {
	actionType
	returnValue LineRead
}

type actionType int

const (
	noAction actionType = iota
	reprocessKey
	exitReadLine
)

type builtin func(ed *Editor)

var builtins = map[string]builtin{
	// Command and insert mode
	"start-insert":        startInsert,
	"start-command":       startCommand,
	"kill-line-left":      killLineLeft,
	"kill-line-right":     killLineRight,
	"kill-word-left":      killWordLeft,
	"kill-rune-left":      killRuneLeft,
	"kill-rune-right":     killRuneRight,
	"move-dot-left":       moveDotLeft,
	"move-dot-right":      moveDotRight,
	"move-dot-left-word":  moveDotLeftWord,
	"move-dot-right-word": moveDotRightWord,
	"move-dot-up":         moveDotUp,
	"move-dot-down":       moveDotDown,
	"insert-key":          insertKey,
	"return-line":         returnLine,
	"return-eof":          returnEOF,
	"default-command":     defaultCommand,
	"default-insert":      defaultInsert,

	// Completion mode
	"complete-prefix-or-start-completion": completePrefixOrStartCompletion,

	"start-completion":   startCompletion,
	"cancel-completion":  cancelCompletion,
	"select-cand-up":     selectCandUp,
	"select-cand-down":   selectCandDown,
	"select-cand-left":   selectCandLeft,
	"select-cand-right":  selectCandRight,
	"cycle-cand-right":   cycleCandRight,
	"default-completion": defaultCompletion,

	// Navigation mode
	"start-navigation":   startNavigation,
	"select-nav-up":      selectNavUp,
	"select-nav-down":    selectNavDown,
	"ascend-nav":         ascendNav,
	"descend-nav":        descendNav,
	"default-navigation": defaultNavigation,

	// History mode
	"start-history":               startHistory,
	"select-history-prev":         selectHistoryPrev,
	"select-history-next":         selectHistoryNext,
	"select-history-next-or-quit": selectHistoryNextOrQuit,
	"default-history":             defaultHistory,
}

var (
	takeNoArg     = errors.New("editor builtins take no arguments")
	noSuchBuiltin = errors.New("no such editor builtin")
	inactive      = errors.New("editor inactive")
)

// Call satisfies the eval.Foreign interface.
func (ed *Editor) Call(name string, args []eval.Value) error {
	if len(args) > 0 {
		return takeNoArg
	}
	f, ok := builtins[name]
	if !ok {
		return noSuchBuiltin
	}
	if !ed.active {
		return inactive
	}
	f(ed)
	return nil
}

func startInsert(ed *Editor) {
	ed.mode = modeInsert
}

func defaultCommand(ed *Editor) {
	k := ed.lastKey
	ed.pushTip(fmt.Sprintf("Unbound: %s", k))
}

func startCommand(ed *Editor) {
	ed.mode = modeCommand
}

func killLineLeft(ed *Editor) {
	sol := strutil.FindLastSOL(ed.line[:ed.dot])
	ed.line = ed.line[:sol] + ed.line[ed.dot:]
	ed.dot = sol
}

func killLineRight(ed *Editor) {
	eol := strutil.FindFirstEOL(ed.line[ed.dot:]) + ed.dot
	ed.line = ed.line[:ed.dot] + ed.line[eol:]
}

// NOTE(xiaq): A word is now defined as a series of non-whitespace chars.
func killWordLeft(ed *Editor) {
	if ed.dot == 0 {
		return
	}
	space := strings.LastIndexFunc(
		strings.TrimRightFunc(ed.line[:ed.dot], unicode.IsSpace),
		unicode.IsSpace) + 1
	ed.line = ed.line[:space] + ed.line[ed.dot:]
	ed.dot = space
}

func killRuneLeft(ed *Editor) {
	if ed.dot > 0 {
		_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
		ed.line = ed.line[:ed.dot-w] + ed.line[ed.dot:]
		ed.dot -= w
	} else {
		ed.flash()
	}
}

func killRuneRight(ed *Editor) {
	if ed.dot < len(ed.line) {
		_, w := utf8.DecodeRuneInString(ed.line[ed.dot:])
		ed.line = ed.line[:ed.dot] + ed.line[ed.dot+w:]
	} else {
		ed.flash()
	}
}

func moveDotLeft(ed *Editor) {
	_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
	ed.dot -= w
}

func moveDotRight(ed *Editor) {
	_, w := utf8.DecodeRuneInString(ed.line[ed.dot:])
	ed.dot += w
}

func moveDotLeftWord(ed *Editor) {
	if ed.dot == 0 {
		return
	}
	space := strings.LastIndexFunc(
		strings.TrimRightFunc(ed.line[:ed.dot], unicode.IsSpace),
		unicode.IsSpace) + 1
	ed.dot = space
}

func moveDotRightWord(ed *Editor) {
	// Move to first space
	p := strings.IndexFunc(ed.line[ed.dot:], unicode.IsSpace)
	if p == -1 {
		ed.dot = len(ed.line)
		return
	}
	ed.dot += p
	// Move to first nonspace
	p = strings.IndexFunc(ed.line[ed.dot:], notSpace)
	if p == -1 {
		ed.dot = len(ed.line)
		return
	}
	ed.dot += p
}

func notSpace(r rune) bool {
	return !unicode.IsSpace(r)
}

func moveDotUp(ed *Editor) {
	sol := strutil.FindLastSOL(ed.line[:ed.dot])
	if sol == 0 {
		ed.flash()
		return
	}
	prevEOL := sol - 1
	prevSOL := strutil.FindLastSOL(ed.line[:prevEOL])
	width := WcWidths(ed.line[sol:ed.dot])
	ed.dot = prevSOL + len(TrimWcWidth(ed.line[prevSOL:prevEOL], width))
}

func moveDotDown(ed *Editor) {
	eol := strutil.FindFirstEOL(ed.line[ed.dot:]) + ed.dot
	if eol == len(ed.line) {
		ed.flash()
		return
	}
	nextSOL := eol + 1
	nextEOL := strutil.FindFirstEOL(ed.line[nextSOL:]) + nextSOL
	sol := strutil.FindLastSOL(ed.line[:ed.dot])
	width := WcWidths(ed.line[sol:ed.dot])
	ed.dot = nextSOL + len(TrimWcWidth(ed.line[nextSOL:nextEOL], width))
}

func insertKey(ed *Editor) {
	k := ed.lastKey
	ed.line = ed.line[:ed.dot] + string(k.Rune) + ed.line[ed.dot:]
	ed.dot += utf8.RuneLen(k.Rune)
}

func returnLine(ed *Editor) {
	ed.nextAction = action{exitReadLine, LineRead{Line: ed.line}}
}

func returnEOF(ed *Editor) {
	if len(ed.line) == 0 {
		ed.nextAction = action{exitReadLine, LineRead{EOF: true}}
	}
}

func selectCandUp(ed *Editor) {
	ed.completion.prev(false)
}

func selectCandDown(ed *Editor) {
	ed.completion.next(false)
}

func selectCandLeft(ed *Editor) {
	if c := ed.completion.current - ed.completionLines; c >= 0 {
		ed.completion.current = c
	}
}

func selectCandRight(ed *Editor) {
	if c := ed.completion.current + ed.completionLines; c < len(ed.completion.candidates) {
		ed.completion.current = c
	}
}

func cycleCandRight(ed *Editor) {
	ed.completion.next(true)
}

func cancelCompletion(ed *Editor) {
	ed.completion = nil
	ed.mode = modeInsert
}

func defaultInsert(ed *Editor) {
	k := ed.lastKey
	if k.Mod == 0 && k.Rune > 0 && unicode.IsGraphic(k.Rune) {
		insertKey(ed)
	} else {
		ed.pushTip(fmt.Sprintf("Unbound: %s", k))
	}
}

func defaultCompletion(ed *Editor) {
	ed.acceptCompletion()
	ed.mode = modeInsert
	ed.nextAction = action{actionType: reprocessKey}
}

func startNavigation(ed *Editor) {
	ed.mode = modeNavigation
	ed.navigation = newNavigation()
}

func selectNavUp(ed *Editor) {
	ed.navigation.prev()
}

func selectNavDown(ed *Editor) {
	ed.navigation.next()
}

func ascendNav(ed *Editor) {
	ed.navigation.ascend()
}

func descendNav(ed *Editor) {
	ed.navigation.descend()
}

func defaultNavigation(ed *Editor) {
	ed.mode = modeInsert
	ed.navigation = nil
}

func startHistory(ed *Editor) {
	ed.history.prefix = ed.line[:ed.dot]
	ed.history.current = len(ed.histories)
	if ed.prevHistory() {
		ed.mode = modeHistory
	} else {
		ed.pushTip("no matching history item")
	}
}

func selectHistoryPrev(ed *Editor) {
	ed.prevHistory()
}

func selectHistoryNext(ed *Editor) {
	ed.nextHistory()
}

func selectHistoryNextOrQuit(ed *Editor) {
	if !ed.nextHistory() {
		ed.mode = modeInsert
	}
}

func defaultHistory(ed *Editor) {
	ed.acceptHistory()
	ed.mode = modeInsert
	ed.nextAction = action{actionType: reprocessKey}
}
