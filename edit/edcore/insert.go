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
		"kill-line-left":        ed.applyKill(moveDotSOL),
		"kill-line-right":       ed.applyKill(moveDotEOL),
		"kill-word-left":        ed.applyKill(moveDotLeftWord),
		"kill-word-right":       ed.applyKill(moveDotRightWord),
		"kill-small-word-left":  ed.applyKill(moveDotLeftSmallWord),
		"kill-small-word-right": ed.applyKill(moveDotRightSmallWord),
		"kill-rune-left":        ed.applyKill(moveDotLeft),
		"kill-rune-right":       ed.applyKill(moveDotRight),

		"move-dot-left":             ed.applyMove(moveDotLeft),
		"move-dot-right":            ed.applyMove(moveDotRight),
		"move-dot-left-word":        ed.applyMove(moveDotLeftWord),
		"move-dot-right-word":       ed.applyMove(moveDotRightWord),
		"move-dot-left-small-word":  ed.applyMove(moveDotLeftSmallWord),
		"move-dot-right-small-word": ed.applyMove(moveDotRightSmallWord),
		"move-dot-sol":              ed.applyMove(moveDotSOL),
		"move-dot-eol":              ed.applyMove(moveDotEOL),

		"move-dot-up":   ed.moveDotUp,
		"move-dot-down": ed.moveDotDown,

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

func (ed *editor) applyMove(move func(string, int) int) func() {
	return func() {
		ed.dot = move(ed.buffer, ed.dot)
	}
}

func (ed *editor) applyKill(move func(string, int) int) func() {
	return func() {
		index := move(ed.buffer, ed.dot)

		if index < ed.dot { // delete left
			ed.buffer = ed.buffer[:index] + ed.buffer[ed.dot:]
			ed.dot = index
		} else { // delete right
			ed.buffer = ed.buffer[:ed.dot] + ed.buffer[index:]
		}
	}
}

func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

func moveDotLeft(buffer string, dot int) int {
	_, w := utf8.DecodeLastRuneInString(buffer[:dot])
	return dot - w
}

func moveDotRight(buffer string, dot int) int {
	_, w := utf8.DecodeRuneInString(buffer[dot:])
	return dot + w
}

// Returns a move- function that moves the dot to the beginning of the
// last run of characters of the same non-whitespace category to the
// left of the dot..
//
// The categorization of characters is derived from the categorizer
// function, which returns for each character an integer representing
// the category of the integer.  The return value `0` of the
// categorizer function is used as a special category representing
// non-word, or "always-skip" characters.  All other return values are
// equivalent, used only to distinguish between different character
// classes.
//
// See `categorizeWord`, `categorizeSmallWord`, and
// `categorizeAlphanumeric` for examples.
func moveDotLeftCategoryFunc(categorize func(rune) int, buffer string, dot int) int {
	// skip trailing whitespaces left of cursor
	left := strings.TrimRightFunc(buffer[:dot], func(r rune) bool {
		return categorize(r) == 0
	})

	// get category of last character
	r, _ := utf8.DecodeLastRuneInString(left)
	cat := categorize(r)

	// skip characters of same category
	left = strings.TrimRightFunc(left, func(r rune) bool {
		return categorize(r) == cat
	})

	return len(left)
}

// Returns a move-function that moves the cursor to the beginning of
// the next run of characters of the same non-whitespace category to
// the right of the dot.
//
// See `moveDotLeftCategoryFunc`.
func moveDotRightCategoryFunc(categorize func(rune) int, buffer string, dot int) int {

	// skip leading whitespaces right of dot
	right := strings.TrimLeftFunc(buffer[dot:], func(r rune) bool {
		return categorize(r) == 0
	})

	// check whether any whitespace was skipped; if whitespace was
	// skipped, then dot is already successfully moved to next
	// non-whitespace run
	if dot < len(buffer)-len(right) {
		return len(buffer) - len(right)
	}

	// no whitespace was skipped, so we still have to skip to the next
	// non-whitespace run

	// get category of first non-whitespace character
	r, _ := utf8.DecodeRuneInString(right)
	cat := categorize(r)

	// skip characters of same category
	right = strings.TrimLeftFunc(right, func(r rune) bool {
		return categorize(r) == cat
	})

	// skip remaining whitespace
	right = strings.TrimLeftFunc(right, func(r rune) bool {
		return categorize(r) == 0
	})

	return len(buffer) - len(right)
}

func categorizeWord(r rune) int {
	switch {
	case unicode.IsSpace(r):
		return 0
	default:
		return 1
	}
}

func categorizeSmallWord(r rune) int {
	switch {
	case unicode.IsSpace(r):
		return 0
	case isAlnum(r):
		return 1
	default:
		return 2
	}
}

func categorizeAlphanumeric(r rune) int {
	switch {
	case isAlnum(r):
		return 1
	default:
		return 0
	}
}

func moveDotLeftWord(buffer string, dot int) int {
	return moveDotLeftCategoryFunc(categorizeWord, buffer, dot)
}
func moveDotRightWord(buffer string, dot int) int {
	return moveDotRightCategoryFunc(categorizeWord, buffer, dot)
}
func moveDotLeftSmallWord(buffer string, dot int) int {
	return moveDotLeftCategoryFunc(categorizeSmallWord, buffer, dot)
}
func moveDotRightSmallWord(buffer string, dot int) int {
	return moveDotRightCategoryFunc(categorizeSmallWord, buffer, dot)
}
func moveDotLeftAlphanumeric(buffer string, dot int) int {
	return moveDotLeftCategoryFunc(categorizeAlphanumeric, buffer, dot)
}
func moveDotRightAlphanumeric(buffer string, dot int) int {
	return moveDotRightCategoryFunc(categorizeAlphanumeric, buffer, dot)
}

func moveDotSOL(buffer string, dot int) int {
	return util.FindLastSOL(buffer[:dot])
}

func moveDotEOL(buffer string, dot int) int {
	return util.FindFirstEOL(buffer[dot:]) + dot
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
