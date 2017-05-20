package edit

import (
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/edit/uitypes"
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
	registerBindings(modeInsert, "", map[uitypes.Key]string{
		// Moving.
		uitypes.Key{uitypes.Left, 0}:             "move-dot-left",
		uitypes.Key{uitypes.Right, 0}:            "move-dot-right",
		uitypes.Key{uitypes.Up, uitypes.Alt}:     "move-dot-up",
		uitypes.Key{uitypes.Down, uitypes.Alt}:   "move-dot-down",
		uitypes.Key{uitypes.Left, uitypes.Ctrl}:  "move-dot-left-word",
		uitypes.Key{uitypes.Right, uitypes.Ctrl}: "move-dot-right-word",
		uitypes.Key{uitypes.Home, 0}:             "move-dot-sol",
		uitypes.Key{uitypes.End, 0}:              "move-dot-eol",
		// Killing.
		uitypes.Key{'U', uitypes.Ctrl}:    "kill-line-left",
		uitypes.Key{'K', uitypes.Ctrl}:    "kill-line-right",
		uitypes.Key{'W', uitypes.Ctrl}:    "kill-word-left",
		uitypes.Key{uitypes.Backspace, 0}: "kill-rune-left",
		// Some terminal send ^H on backspace
		// uitypes.Key{'H', uitypes.Ctrl}: "kill-rune-left",
		uitypes.Key{uitypes.Delete, 0}: "kill-rune-right",
		// Inserting.
		uitypes.Key{'.', uitypes.Alt}:           "insert-last-word",
		uitypes.Key{uitypes.Enter, uitypes.Alt}: "insert-key",
		// Controls.
		uitypes.Key{uitypes.Enter, 0}:  "smart-enter",
		uitypes.Key{'D', uitypes.Ctrl}: "return-eof",
		uitypes.Key{uitypes.F2, 0}:     "toggle-quote-paste",

		// Other modes.
		// uitypes.Key{'[', uitypes.Ctrl}: "command-start",
		uitypes.Key{uitypes.Tab, 0}:    "compl:smart-start",
		uitypes.Key{uitypes.Up, 0}:     "history:start",
		uitypes.Key{uitypes.Down, 0}:   "end-of-history",
		uitypes.Key{'N', uitypes.Ctrl}: "nav:start",
		uitypes.Key{'R', uitypes.Ctrl}: "histlist:start",
		uitypes.Key{',', uitypes.Alt}:  "bang:start",
		uitypes.Key{'L', uitypes.Ctrl}: "loc:start",
		uitypes.Key{'V', uitypes.Ctrl}: "insert-raw",

		uitypes.Default: "insert:default",
	})
	registerBindings(modeCommand, "", map[uitypes.Key]string{
		// Moving.
		uitypes.Key{'h', 0}: "move-dot-left",
		uitypes.Key{'l', 0}: "move-dot-right",
		uitypes.Key{'k', 0}: "move-dot-up",
		uitypes.Key{'j', 0}: "move-dot-down",
		uitypes.Key{'b', 0}: "move-dot-left-word",
		uitypes.Key{'w', 0}: "move-dot-right-word",
		uitypes.Key{'0', 0}: "move-dot-sol",
		uitypes.Key{'$', 0}: "move-dot-eol",
		// Killing.
		uitypes.Key{'x', 0}: "kill-rune-right",
		uitypes.Key{'D', 0}: "kill-line-right",
		// Controls.
		uitypes.Key{'i', 0}: "insert:start",
		uitypes.Default:     "command:default",
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

func (*insert) Mode() ModeType {
	return modeInsert
}

// uitypes.Insert mode is the default mode and has an empty mode.
func (ins *insert) ModeLine() renderer {
	if ins.quotePaste {
		return modeLineRenderer{" INSERT (quote paste) ", ""}
	}
	return nil
}

type command struct{}

func (*command) Mode() ModeType {
	return modeCommand
}

func (*command) ModeLine() renderer {
	return modeLineRenderer{" COMMAND ", ""}
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
	if ed.store == nil {
		ed.addTip("store offline")
		return
	}
	cmd, err := ed.store.GetLastCmd(-1, "")
	if err == nil {
		ed.insertAtDot(lastWord(cmd.Text))
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
		// There is a parsing error at the end. uitypes.Insert a newline and copy
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
		literals := ed.line[ed.dot-ed.insert.literalInserts-1 : ed.dot]
		for abbr, full := range ed.abbreviations {
			if strings.HasSuffix(literals, abbr) {
				ed.line = ed.line[:ed.dot-len(abbr)] + full + ed.line[ed.dot:]
				ed.dot += len(full) - len(abbr)
				return
			}
		}
		// No match.
		ed.insert.insertedLiteral = true
	} else {
		ed.Notify("Unbound: %s", k)
	}
}

// likeChar returns if a key looks like a character meant to be input (as
// opposed to a function key).
func likeChar(k uitypes.Key) bool {
	return k.Mod == 0 && k.Rune > 0 && unicode.IsGraphic(k.Rune)
}

func commandDefault(ed *Editor) {
	k := ed.lastKey
	ed.Notify("Unbound: %s", k)
}
