package edit

import (
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

// Builtins related to insert and command mode.

var (
	_ = registerBuiltins("", map[string]func(*Editor){
		"kill-line-left":       killLineLeft,
		"kill-line-right":      killLineRight,
		"kill-word-left":       killWordLeft,
		"kill-small-word-left": killSmallWordLeft,
		"kill-rune-left":       killRuneLeft,
		"kill-rune-right":      killRuneRight,

		"move-dot-left":       moveDotLeft,
		"move-dot-right":      moveDotRight,
		"move-dot-left-word":  moveDotLeftWord,
		"move-dot-right-word": moveDotRightWord,
		"move-dot-sol":        moveDotSOL,
		"move-dot-eol":        moveDotEOL,
		"move-dot-up":         moveDotUp,
		"move-dot-down":       moveDotDown,

		"insert-last-word": insertLastWord,
		"insert-key":       insertKey,

		"return-line": returnLine,
		"smart-enter": smartEnter,
		"return-eof":  returnEOF,

		"toggle-quote-paste": toggleQuotePaste,
		"insert-raw":         startInsertRaw,

		"end-of-history": endOfHistory,
		"redraw":         redraw,
	})
	_ = registerBuiltins("insert", map[string]func(*Editor){
		"start":   insertStart,
		"default": insertDefault,
	})
	_ = registerBuiltins("command", map[string]func(*Editor){
		"start":   commandStart,
		"default": commandDefault,
	})
)

func init() {
	registerBindings(modeInsert, "", map[ui.Key]string{
		// Moving.
		{ui.Left, 0}:        "move-dot-left",
		{ui.Right, 0}:       "move-dot-right",
		{ui.Up, ui.Alt}:     "move-dot-up",
		{ui.Down, ui.Alt}:   "move-dot-down",
		{ui.Left, ui.Ctrl}:  "move-dot-left-word",
		{ui.Right, ui.Ctrl}: "move-dot-right-word",
		{'b', ui.Alt}:       "move-dot-left-word",
		{'f', ui.Alt}:       "move-dot-right-word",
		{ui.Home, 0}:        "move-dot-sol",
		{ui.End, 0}:         "move-dot-eol",
		// Killing.
		{'U', ui.Ctrl}:    "kill-line-left",
		{'K', ui.Ctrl}:    "kill-line-right",
		{'W', ui.Ctrl}:    "kill-word-left",
		{ui.Backspace, 0}: "kill-rune-left",
		// Some terminal send ^H on backspace
		// ui.Key{'H', ui.Ctrl}: "kill-rune-left",
		{ui.Delete, 0}: "kill-rune-right",
		// Inserting.
		{'.', ui.Alt}:      "insert-last-word",
		{ui.Enter, ui.Alt}: "insert-key",
		// Controls.
		{ui.Enter, 0}:  "smart-enter",
		{'D', ui.Ctrl}: "return-eof",
		{ui.F2, 0}:     "toggle-quote-paste",

		// Other modes.
		// ui.Key{'[', ui.Ctrl}: "command-start",
		{ui.Tab, 0}:    "completion:smart-start",
		{ui.Up, 0}:     "history:start",
		{ui.Down, 0}:   "end-of-history",
		{'N', ui.Ctrl}: "navigation:start",
		{'R', ui.Ctrl}: "histlist:start",
		{'1', ui.Alt}:  "lastcmd:start",
		{'/', ui.Ctrl}: "location:start",
		{'V', ui.Ctrl}: "insert-raw",

		ui.Default: "insert:default",
	})
	registerBindings(modeCommand, "", map[ui.Key]string{
		// Moving.
		{'h', 0}: "move-dot-left",
		{'l', 0}: "move-dot-right",
		{'k', 0}: "move-dot-up",
		{'j', 0}: "move-dot-down",
		{'b', 0}: "move-dot-left-word",
		{'w', 0}: "move-dot-right-word",
		{'0', 0}: "move-dot-sol",
		{'$', 0}: "move-dot-eol",
		// Killing.
		{'x', 0}: "kill-rune-right",
		{'D', 0}: "kill-line-right",
		// Controls.
		{'i', 0}:   "insert:start",
		ui.Default: "command:default",
	})
}

type insert struct {
	quotePaste bool
	// The number of consecutive key inserts. Used for abbreviation expansion.
	literalInserts int
	// Indicates whether a key was inserted (via insert-default). A hack for
	// maintaining the inserts field.
	insertedLiteral bool
}

// ui.Insert mode is the default mode and has an empty mode.
func (ins *insert) ModeLine() renderer {
	if ins.quotePaste {
		return modeLineRenderer{" INSERT (quote paste) ", ""}
	}
	return nil
}

func (*insert) Binding(k ui.Key) eval.CallableValue {
	return getBinding(modeInsert, k)
}

type command struct{}

func (*command) ModeLine() renderer {
	return modeLineRenderer{" COMMAND ", ""}
}

func (*command) Binding(k ui.Key) eval.CallableValue {
	return getBinding(modeCommand, k)
}

func insertStart(ed *Editor) {
	ed.mode = &ed.insert
}

func commandStart(ed *Editor) {
	ed.mode = &ed.command
}

func killLineLeft(ed *Editor) {
	sol := util.FindLastSOL(ed.line[:ed.dot])
	ed.line = ed.line[:sol] + ed.line[ed.dot:]
	ed.dot = sol
}

func killLineRight(ed *Editor) {
	eol := util.FindFirstEOL(ed.line[ed.dot:]) + ed.dot
	ed.line = ed.line[:ed.dot] + ed.line[eol:]
}

// NOTE(xiaq): A word is a run of non-space runes. When killing a word,
// trimming spaces are removed as well. Examples:
// "abc  xyz" -> "abc  ", "abc xyz " -> "abc  ".

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

// NOTE(xiaq): A small word is either a run of alphanumeric (Unicode category L
// or N) runes or a run of non-alphanumeric runes. This is consistent with vi's
// definition of word, except that "_" is not considered alphanumeric. When
// killing a small word, trimming spaces are removed as well. Examples:
// "abc/~" -> "abc", "~/abc" -> "~/", "abc* " -> "abc"

func killSmallWordLeft(ed *Editor) {
	left := strings.TrimRightFunc(ed.line[:ed.dot], unicode.IsSpace)
	// The case of left == "" is handled as well.
	r, _ := utf8.DecodeLastRuneInString(left)
	if isAlnum(r) {
		left = strings.TrimRightFunc(left, isAlnum)
	} else {
		left = strings.TrimRightFunc(
			left, func(r rune) bool { return !isAlnum(r) })
	}
	ed.line = left + ed.line[ed.dot:]
	ed.dot = len(left)
}

func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
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

func moveDotSOL(ed *Editor) {
	sol := util.FindLastSOL(ed.line[:ed.dot])
	ed.dot = sol
}

func moveDotEOL(ed *Editor) {
	eol := util.FindFirstEOL(ed.line[ed.dot:]) + ed.dot
	ed.dot = eol
}

func moveDotUp(ed *Editor) {
	sol := util.FindLastSOL(ed.line[:ed.dot])
	if sol == 0 {
		ed.flash()
		return
	}
	prevEOL := sol - 1
	prevSOL := util.FindLastSOL(ed.line[:prevEOL])
	width := util.Wcswidth(ed.line[sol:ed.dot])
	ed.dot = prevSOL + len(util.TrimWcwidth(ed.line[prevSOL:prevEOL], width))
}

func moveDotDown(ed *Editor) {
	eol := util.FindFirstEOL(ed.line[ed.dot:]) + ed.dot
	if eol == len(ed.line) {
		ed.flash()
		return
	}
	nextSOL := eol + 1
	nextEOL := util.FindFirstEOL(ed.line[nextSOL:]) + nextSOL
	sol := util.FindLastSOL(ed.line[:ed.dot])
	width := util.Wcswidth(ed.line[sol:ed.dot])
	ed.dot = nextSOL + len(util.TrimWcwidth(ed.line[nextSOL:nextEOL], width))
}

func insertLastWord(ed *Editor) {
	if ed.daemon == nil {
		ed.addTip("daemon offline")
		return
	}
	_, cmd, err := ed.daemon.PrevCmd(-1, "")
	if err == nil {
		ed.insertAtDot(lastWord(cmd))
	} else {
		ed.addTip("db error: %s", err.Error())
	}
}

func lastWord(s string) string {
	words := wordify(s)
	if len(words) == 0 {
		return ""
	}
	return words[len(words)-1]
}

func insertKey(ed *Editor) {
	k := ed.lastKey
	ed.insertAtDot(string(k.Rune))
}

func returnLine(ed *Editor) {
	ed.nextAction = action{typ: exitReadLine, returnLine: ed.line}
}

func smartEnter(ed *Editor) {
	if ed.parseErrorAtEnd {
		// There is a parsing error at the end. ui.Insert a newline and copy
		// indents from previous line.
		indent := findLastIndent(ed.line[:ed.dot])
		ed.insertAtDot("\n" + indent)
	} else {
		returnLine(ed)
	}
}

func findLastIndent(s string) string {
	line := s[util.FindLastSOL(s):]
	trimmed := strings.TrimLeft(line, " \t")
	return line[:len(line)-len(trimmed)]
}

func returnEOF(ed *Editor) {
	if len(ed.line) == 0 {
		ed.nextAction = action{typ: exitReadLine, returnErr: io.EOF}
	}
}

func toggleQuotePaste(ed *Editor) {
	ed.insert.quotePaste = !ed.insert.quotePaste
}

func endOfHistory(ed *Editor) {
	ed.Notify("End of history")
}

func redraw(ed *Editor) {
	ed.refresh(true, true)
}

func insertDefault(ed *Editor) {
	k := ed.lastKey
	if likeChar(k) {
		insertKey(ed)
		// Match abbreviations.
		expanded := false
		literals := ed.line[ed.dot-ed.insert.literalInserts-1 : ed.dot]
		ed.abbrIterate(func(abbr, full string) bool {
			if strings.HasSuffix(literals, abbr) {
				ed.line = ed.line[:ed.dot-len(abbr)] + full + ed.line[ed.dot:]
				ed.dot += len(full) - len(abbr)
				expanded = true
				return false
			}
			return true
		})
		// No match.
		if !expanded {
			ed.insert.insertedLiteral = true
		}
	} else {
		ed.Notify("Unbound: %s", k)
	}
}

// likeChar returns if a key looks like a character meant to be input (as
// opposed to a function key).
func likeChar(k ui.Key) bool {
	return k.Mod == 0 && k.Rune > 0 && unicode.IsGraphic(k.Rune)
}

func commandDefault(ed *Editor) {
	k := ed.lastKey
	ed.Notify("Unbound: %s", k)
}
