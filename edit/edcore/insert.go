package edcore

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse/parseutil"
	"github.com/elves/elvish/util"
)

// Builtins related to insert and command mode.

func init() {
	atEditorInit(initCoreFns)
	atEditorInit(initInsert)
	atEditorInit(initCommand)
}

func initCoreFns(ed *editor, ns eval.Ns) {
	ns.AddBuiltinFns("edit:", map[string]interface{}{
		"kill-line-left":       ed.killLineLeft,
		"kill-line-right":      ed.killLineRight,
		"kill-word-left":       ed.killWordLeft,
		"kill-small-word-left": ed.killSmallWordLeft,
		"kill-rune-left":       ed.killRuneLeft,
		"kill-rune-right":      ed.killRuneRight,

		"move-dot-left":       ed.moveDotLeft,
		"move-dot-right":      ed.moveDotRight,
		"move-dot-left-word":  ed.moveDotLeftWord,
		"move-dot-right-word": ed.moveDotRightWord,
		"move-dot-sol":        ed.moveDotSOL,
		"move-dot-eol":        ed.moveDotEOL,
		"move-dot-up":         ed.moveDotUp,
		"move-dot-down":       ed.moveDotDown,

		"insert-last-word": ed.insertLastWord,
		"insert-key":       ed.insertKey,

		"return-line": ed.returnLine,
		"smart-enter": ed.smartEnter,
		"return-eof":  ed.returnEOF,

		"toggle-quote-paste": ed.toggleQuotePaste,
		"insert-raw":         ed.startInsertRaw,

		"end-of-history": ed.endOfHistory,
		"redraw":         ed.redraw,
	})
}

type insert struct {
	binding eddefs.BindingMap
	insertState
}

type insertState struct {
	quotePaste bool
	// The number of consecutive key inserts. Used for abbreviation expansion.
	literalInserts int
	// Indicates whether a key was inserted (via insert-default). A hack for
	// maintaining the inserts field.
	insertedLiteral bool
}

func initInsert(ed *editor, ns eval.Ns) {
	insert := &insert{binding: emptyBindingMap}
	ed.insert = insert

	insertNs := eval.Ns{
		"binding": vars.FromPtr(&insert.binding),
	}
	insertNs.AddBuiltinFns("edit:insert:", map[string]interface{}{
		"start":   ed.SetModeInsert,
		"default": ed.insertDefault,
	})
	ns.AddNs("insert", insertNs)
}

func (ins *insert) Teardown() {
	ins.insertState = insertState{}
}

// Insert mode is the default mode and has an empty mode, unless quotePaste is
// true.
func (ins *insert) ModeLine() ui.Renderer {
	if ins.quotePaste {
		return ui.NewModeLineRenderer(" INSERT (quote paste) ", "")
	}
	return nil
}

func (ins *insert) Binding(k ui.Key) eval.Callable {
	return ins.binding.GetOrDefault(k)
}

type command struct {
	binding eddefs.BindingMap
}

func initCommand(ed *editor, ns eval.Ns) {
	command := &command{binding: emptyBindingMap}
	ed.command = command

	commandNs := eval.Ns{
		"binding": vars.FromPtr(&command.binding),
	}
	commandNs.AddBuiltinFns("edit:command:", map[string]interface{}{
		"start":   ed.commandStart,
		"default": ed.commandDefault,
	})
	ns.AddNs("command", commandNs)
}

func (*command) Teardown() {}

func (*command) ModeLine() ui.Renderer {
	return ui.NewModeLineRenderer(" COMMAND ", "")
}

func (cmd *command) Binding(k ui.Key) eval.Callable {
	return cmd.binding.GetOrDefault(k)
}

func (ed *editor) commandStart() {
	ed.SetMode(ed.command)
}

func (ed *editor) killLineLeft() {
	sol := util.FindLastSOL(ed.buffer[:ed.dot])
	ed.buffer = ed.buffer[:sol] + ed.buffer[ed.dot:]
	ed.dot = sol
}

func (ed *editor) killLineRight() {
	eol := util.FindFirstEOL(ed.buffer[ed.dot:]) + ed.dot
	ed.buffer = ed.buffer[:ed.dot] + ed.buffer[eol:]
}

// NOTE(xiaq): A word is a run of non-space runes. When killing a word,
// trimming spaces are removed as well. Examples:
// "abc  xyz" -> "abc  ", "abc xyz " -> "abc  ".

func (ed *editor) killWordLeft() {
	if ed.dot == 0 {
		return
	}
	space := strings.LastIndexFunc(
		strings.TrimRightFunc(ed.buffer[:ed.dot], unicode.IsSpace),
		unicode.IsSpace) + 1
	ed.buffer = ed.buffer[:space] + ed.buffer[ed.dot:]
	ed.dot = space
}

// NOTE(xiaq): A small word is either a run of alphanumeric (Unicode category L
// or N) runes or a run of non-alphanumeric runes. This is consistent with vi's
// definition of word, except that "_" is not considered alphanumeric. When
// killing a small word, trimming spaces are removed as well. Examples:
// "abc/~" -> "abc", "~/abc" -> "~/", "abc* " -> "abc"

func (ed *editor) killSmallWordLeft() {
	left := strings.TrimRightFunc(ed.buffer[:ed.dot], unicode.IsSpace)
	// The case of left == "" is handled as well.
	r, _ := utf8.DecodeLastRuneInString(left)
	if isAlnum(r) {
		left = strings.TrimRightFunc(left, isAlnum)
	} else {
		left = strings.TrimRightFunc(
			left, func(r rune) bool { return !isAlnum(r) })
	}
	ed.buffer = left + ed.buffer[ed.dot:]
	ed.dot = len(left)
}

func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

func (ed *editor) killRuneLeft() {
	if ed.dot > 0 {
		_, w := utf8.DecodeLastRuneInString(ed.buffer[:ed.dot])
		ed.buffer = ed.buffer[:ed.dot-w] + ed.buffer[ed.dot:]
		ed.dot -= w
	} else {
		ed.flash()
	}
}

func (ed *editor) killRuneRight() {
	if ed.dot < len(ed.buffer) {
		_, w := utf8.DecodeRuneInString(ed.buffer[ed.dot:])
		ed.buffer = ed.buffer[:ed.dot] + ed.buffer[ed.dot+w:]
	} else {
		ed.flash()
	}
}

func (ed *editor) moveDotLeft() {
	_, w := utf8.DecodeLastRuneInString(ed.buffer[:ed.dot])
	ed.dot -= w
}

func (ed *editor) moveDotRight() {
	_, w := utf8.DecodeRuneInString(ed.buffer[ed.dot:])
	ed.dot += w
}

func (ed *editor) moveDotLeftWord() {
	if ed.dot == 0 {
		return
	}
	space := strings.LastIndexFunc(
		strings.TrimRightFunc(ed.buffer[:ed.dot], unicode.IsSpace),
		unicode.IsSpace) + 1
	ed.dot = space
}

func (ed *editor) moveDotRightWord() {
	// Move to first space
	p := strings.IndexFunc(ed.buffer[ed.dot:], unicode.IsSpace)
	if p == -1 {
		ed.dot = len(ed.buffer)
		return
	}
	ed.dot += p
	// Move to first nonspace
	p = strings.IndexFunc(ed.buffer[ed.dot:], notSpace)
	if p == -1 {
		ed.dot = len(ed.buffer)
		return
	}
	ed.dot += p
}

func notSpace(r rune) bool {
	return !unicode.IsSpace(r)
}

func (ed *editor) moveDotSOL() {
	sol := util.FindLastSOL(ed.buffer[:ed.dot])
	ed.dot = sol
}

func (ed *editor) moveDotEOL() {
	eol := util.FindFirstEOL(ed.buffer[ed.dot:]) + ed.dot
	ed.dot = eol
}

func (ed *editor) moveDotUp() {
	sol := util.FindLastSOL(ed.buffer[:ed.dot])
	if sol == 0 {
		ed.flash()
		return
	}
	prevEOL := sol - 1
	prevSOL := util.FindLastSOL(ed.buffer[:prevEOL])
	width := util.Wcswidth(ed.buffer[sol:ed.dot])
	ed.dot = prevSOL + len(util.TrimWcwidth(ed.buffer[prevSOL:prevEOL], width))
}

func (ed *editor) moveDotDown() {
	eol := util.FindFirstEOL(ed.buffer[ed.dot:]) + ed.dot
	if eol == len(ed.buffer) {
		ed.flash()
		return
	}
	nextSOL := eol + 1
	nextEOL := util.FindFirstEOL(ed.buffer[nextSOL:]) + nextSOL
	sol := util.FindLastSOL(ed.buffer[:ed.dot])
	width := util.Wcswidth(ed.buffer[sol:ed.dot])
	ed.dot = nextSOL + len(util.TrimWcwidth(ed.buffer[nextSOL:nextEOL], width))
}

func (ed *editor) insertLastWord() {
	if ed.daemon == nil {
		ed.AddTip("daemon offline")
		return
	}
	_, cmd, err := ed.daemon.PrevCmd(-1, "")
	if err == nil {
		ed.InsertAtDot(lastWord(cmd))
	} else {
		ed.AddTip("db error: %s", err.Error())
	}
}

func lastWord(s string) string {
	words := parseutil.Wordify(s)
	if len(words) == 0 {
		return ""
	}
	return words[len(words)-1]
}

func (ed *editor) insertKey() {
	k := ed.lastKey
	ed.InsertAtDot(string(k.Rune))
}

func (ed *editor) returnLine() {
	ed.SetAction(commitLine)
}

func (ed *editor) smartEnter() {
	if ed.parseErrorAtEnd {
		// There is a parsing error at the end. ui.Insert a newline and copy
		// indents from previous line.
		indent := findLastIndent(ed.buffer[:ed.dot])
		ed.InsertAtDot("\n" + indent)
	} else {
		ed.returnLine()
	}
}

func findLastIndent(s string) string {
	line := s[util.FindLastSOL(s):]
	trimmed := strings.TrimLeft(line, " \t")
	return line[:len(line)-len(trimmed)]
}

func (ed *editor) returnEOF() {
	if len(ed.buffer) == 0 {
		ed.SetAction(commitEOF)
	}
}

func (ed *editor) toggleQuotePaste() {
	ed.insert.quotePaste = !ed.insert.quotePaste
}

func (ed *editor) endOfHistory() {
	ed.Notify("End of history")
}

func (ed *editor) redraw() {
	ed.refresh(true, true)
}

func (ed *editor) insertDefault() {
	k := ed.lastKey
	if likeChar(k) {
		ed.insertKey()
		// Match abbreviations.
		expanded := false
		literals := ed.buffer[ed.dot-ed.insert.literalInserts-1 : ed.dot]
		abbrIterate(ed.abbr, func(abbr, full string) bool {
			if strings.HasSuffix(literals, abbr) {
				ed.buffer = ed.buffer[:ed.dot-len(abbr)] + full + ed.buffer[ed.dot:]
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

func (ed *editor) commandDefault() {
	k := ed.lastKey
	ed.Notify("Unbound: %s", k)
}
