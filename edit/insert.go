package edit

import (
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/util"
)

// Builtins related to insert and command mode.

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

// Insert mode is the default mode and has an empty mode.
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

func startInsert(ed *Editor) {
	ed.mode = &ed.insert
}

func startCommand(ed *Editor) {
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
	_, lastLine, err := ed.store.LastCmd(-1, "")
	if err == nil {
		ed.insertAtDot(lastWord(lastLine))
	} else {
		ed.addTip("db error: %s", err.Error())
	}
}

func lastWord(s string) string {
	s = strings.TrimRightFunc(s, unicode.IsSpace)
	i := strings.LastIndexFunc(s, unicode.IsSpace) + 1
	return s[i:]
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
		// There is a parsing error at the end. Insert a newline and copy
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

func defaultInsert(ed *Editor) {
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

func defaultCommand(ed *Editor) {
	k := ed.lastKey
	ed.Notify("Unbound: %s", k)
}

// likeChar returns if a key looks like a character meant to be input (as
// opposed to a function key).
func likeChar(k Key) bool {
	return k.Mod == 0 && k.Rune > 0 && unicode.IsGraphic(k.Rune)
}
