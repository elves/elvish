package cliedit

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/codearea"
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

//elvdoc:fn binding-map
//
// Converts a normal map into a binding map.

//elvdoc:fn commit-code
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read. Internally, this works by raising a special exception.

//elvdoc:fn commit-eof
//
// Causes the Elvish REPL to terminate. Internally, this works by raising a
// special exception.

func initMiscBuiltins(app *cli.App, ns eval.Ns) {
	ns.AddGoFns("<edit>", map[string]interface{}{
		"binding-map": eddefs.MakeBindingMap,
		"commit-code": app.CommitCode,
		"commit-eof":  app.CommitEOF,
	})
}

var bufferBuiltinsData = map[string]func(*codearea.CodeBuffer){
	"move-left":             makeMove(moveDotLeft),
	"move-right":            makeMove(moveDotRight),
	"move-left-word":        makeMove(moveDotLeftWord),
	"move-right-word":       makeMove(moveDotRightWord),
	"move-left-small-word":  makeMove(moveDotLeftWord),
	"move-right-small-word": makeMove(moveDotRightSmallWord),
	"move-left-alnum-word":  makeMove(moveDotLeftAlnumWord),
	"move-right-alnum-word": makeMove(moveDotRightAlnumWord),
	"move-sol":              makeMove(moveDotSOL),
	"move-eol":              makeMove(moveDotEOL),

	"move-up":   makeMove(moveDotUp),
	"move-down": makeMove(moveDotDown),

	"kill-left":             makeKill(moveDotLeft),
	"kill-right":            makeKill(moveDotRight),
	"kill-left-word":        makeKill(moveDotLeftWord),
	"kill-right-word":       makeKill(moveDotRightWord),
	"kill-left-small-word":  makeKill(moveDotLeftWord),
	"kill-right-small-word": makeKill(moveDotRightSmallWord),
	"kill-left-alnum-word":  makeKill(moveDotLeftAlnumWord),
	"kill-right-alnum-word": makeKill(moveDotRightAlnumWord),
	"kill-sol":              makeKill(moveDotSOL),
	"kill-eol":              makeKill(moveDotEOL),
}

func initBufferBuiltins(app *cli.App, ns eval.Ns) {
	ns.AddGoFns("<edit>", bufferBuiltins(app))
}

func bufferBuiltins(app *cli.App) map[string]interface{} {
	m := make(map[string]interface{})
	for name, fn := range bufferBuiltinsData {
		// Make a lexically scoped copy of fn.
		fn2 := fn
		m[name] = func() {
			app.CodeArea.MutateCodeAreaState(func(s *codearea.State) {
				fn2(&s.CodeBuffer)
			})
		}
	}
	return m
}

// A pure function that takes the current buffer and dot, and returns a new
// value for the dot. Used to derive move- and kill- functions that operate on
// the editor state.
type pureMover func(buffer string, dot int) int

func makeMove(m pureMover) func(*codearea.CodeBuffer) {
	return func(buf *codearea.CodeBuffer) {
		buf.Dot = m(buf.Content, buf.Dot)
	}
}

func makeKill(m pureMover) func(*codearea.CodeBuffer) {
	return func(buf *codearea.CodeBuffer) {
		newDot := m(buf.Content, buf.Dot)
		if newDot < buf.Dot {
			// Dot moved to the left: remove text between new dot and old dot,
			// and move the dot itself
			buf.Content = buf.Content[:newDot] + buf.Content[buf.Dot:]
			buf.Dot = newDot
		} else if newDot > buf.Dot {
			// Dot moved to the right: remove text between old dot and new dot.
			buf.Content = buf.Content[:buf.Dot] + buf.Content[newDot:]
		}
	}
}

// Implementation of pure movers.

//elvdoc:fn move-left
//
// Moves the dot left one rune. Does nothing if the dot is at the beginning of
// the buffer.

//elvdoc:fn kill-left
//
// Kills one rune left of the dot. Does nothing if the dot is at the beginning of
// the buffer.

func moveDotLeft(buffer string, dot int) int {
	_, w := utf8.DecodeLastRuneInString(buffer[:dot])
	return dot - w
}

//elvdoc:fn move-right
//
// Moves the dot right one rune. Does nothing if the dot is at the end of the
// buffer.

//elvdoc:fn kill-left
//
// Kills one rune right of the dot. Does nothing if the dot is at the end of the
// buffer.

func moveDotRight(buffer string, dot int) int {
	_, w := utf8.DecodeRuneInString(buffer[dot:])
	return dot + w
}

//elvdoc:fn move-sol
//
// Moves the dot to the start of the current line.

//elvdoc:fn kill-sol
//
// Deletes the text between the dot and the start of the current line.

func moveDotSOL(buffer string, dot int) int {
	return util.FindLastSOL(buffer[:dot])
}

//elvdoc:fn move-eol
//
// Moves the dot to the end of the current line.

//elvdoc:fn kill-eol
//
// Deletes the text between the dot and the end of the current line.

func moveDotEOL(buffer string, dot int) int {
	return util.FindFirstEOL(buffer[dot:]) + dot
}

//elvdoc:fn move-up
//
// Moves the dot up one line, trying to preserve the visual horizontal position.
// Does nothing if dot is already on the first line of the buffer.

func moveDotUp(buffer string, dot int) int {
	sol := util.FindLastSOL(buffer[:dot])
	if sol == 0 {
		// Already in the first line.
		return dot
	}
	prevEOL := sol - 1
	prevSOL := util.FindLastSOL(buffer[:prevEOL])
	width := util.Wcswidth(buffer[sol:dot])
	return prevSOL + len(util.TrimWcwidth(buffer[prevSOL:prevEOL], width))
}

//elvdoc:fn move-down
//
// Moves the dot down one line, trying to preserve the visual horizontal
// position. Does nothing if dot is already on the last line of the buffer.

func moveDotDown(buffer string, dot int) int {
	eol := util.FindFirstEOL(buffer[dot:]) + dot
	if eol == len(buffer) {
		// Already in the last line.
		return dot
	}
	nextSOL := eol + 1
	nextEOL := util.FindFirstEOL(buffer[nextSOL:]) + nextSOL
	sol := util.FindLastSOL(buffer[:dot])
	width := util.Wcswidth(buffer[sol:dot])
	return nextSOL + len(util.TrimWcwidth(buffer[nextSOL:nextEOL], width))
}

// TODO(xiaq): Document the concepts of words, small words and alnum words.

//elvdoc:fn move-left-word
//
// Moves the dot to the beginning of the last word to the left of the dot.

//elvdoc:fn kill-left-word
//
// Deletes the the last word to the left of the dot.

func moveDotLeftWord(buffer string, dot int) int {
	return moveDotLeftGeneralWord(categorizeWord, buffer, dot)
}

//elvdoc:fn move-right-word
//
// Moves the dot to the beginning of the first word to the right of the dot.

//elvdoc:fn kill-right-word
//
// Deletes the the first word to the right of the dot.

func moveDotRightWord(buffer string, dot int) int {
	return moveDotRightGeneralWord(categorizeWord, buffer, dot)
}

func categorizeWord(r rune) int {
	switch {
	case unicode.IsSpace(r):
		return 0
	default:
		return 1
	}
}

//elvdoc:fn move-left-small-word
//
// Moves the dot to the beginning of the last small word to the left of the dot.

//elvdoc:fn kill-left-small-word
//
// Deletes the the last small word to the left of the dot.

func moveDotLeftSmallWord(buffer string, dot int) int {
	return moveDotLeftGeneralWord(categorizeSmallWord, buffer, dot)
}

//elvdoc:fn move-right-small-word
//
// Moves the dot to the beginning of the first small word to the right of the dot.

//elvdoc:fn kill-right-small-word
//
// Deletes the the first small word to the right of the dot.

func moveDotRightSmallWord(buffer string, dot int) int {
	return moveDotRightGeneralWord(categorizeSmallWord, buffer, dot)
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

//elvdoc:fn move-left-alnum-word
//
// Moves the dot to the beginning of the last alnum word to the left of the dot.

//elvdoc:fn kill-left-alnum-word
//
// Deletes the the last alnum word to the left of the dot.

func moveDotLeftAlnumWord(buffer string, dot int) int {
	return moveDotLeftGeneralWord(categorizeAlnum, buffer, dot)
}

//elvdoc:fn move-right-alnum-word
//
// Moves the dot to the beginning of the first alnum word to the right of the dot.

//elvdoc:fn kill-right-alnum-word
//
// Deletes the the first alnum word to the right of the dot.

func moveDotRightAlnumWord(buffer string, dot int) int {
	return moveDotRightGeneralWord(categorizeAlnum, buffer, dot)
}

func categorizeAlnum(r rune) int {
	switch {
	case isAlnum(r):
		return 1
	default:
		return 0
	}
}

// Word movements are are more complex than one may expect. There are also
// several flavors of word movements supported by Elvish.
//
// To understand word movements, we first need to categorize runes into several
// categories: a whitespace category, plus one or more word category. The
// flavors of word movements are described by their different categorization:
//
// * Plain word: two categories: whitespace, and non-whitespace. This flavor
//   corresponds to WORD in vi.
//
// * Small word: whitespace, alphanumeric, and everything else. This flavor
//   corresponds to word in vi.
//
// * Alphanumeric word: non-alphanumeric (all treated as whitespace) and
//   alphanumeric. This flavor corresponds to word in readline and zsh (when
//   moving left; see below for the difference in behavior when moving right).
//
// After fixing the flavor, a "word" is a run of runes in the same
// non-whitespace category. For instance, the text "cd ~/tmp" has:
//
// * Two plain words: "cd" and "~/tmp".
//
// * Three small words: "cd", "~/" and "tmp".
//
// * Two alphanumeric words: "cd" and "tmp".
//
// To move left one word, we always move to the beginning of the last word to
// the left of the dot (excluding the dot). That is:
//
// * If we are in the middle of a word, we will move to its beginning.
//
// * If we are already at the beginning of a word, we will move to the beginning
//   of the word before that.
//
// * If we are in a run of whitespaces, we will move to the beginning of the
//   word before the run of whitespaces.
//
// Moving right one word works similarly: we move to the beginning of the first
// word to the right of the dot (excluding the dot). This behavior is the same
// as vi and zsh, but differs from GNU readline (used by bash) and fish, which
// moves the dot to one point after the end of the first word to the right of
// the dot.
//
// See the test case for a real-world example of how the different flavors of
// word movements work.
//
// A remark: This definition of "word movement" is general enough to include
// single-rune movements as a special case, where each rune is in its own word
// category (even whitespace runes). Single-rune movements are not implemented
// as such though, to avoid making things unnecessarily complex.

// A function that describes a word flavor by categorizing runes. The return
// value of 0 represents the whitespace category while other values represent
// different word categories.
type categorizer func(rune) int

// Move the dot left one word, using the word flavor described by the
// categorizer.
func moveDotLeftGeneralWord(categorize categorizer, buffer string, dot int) int {
	left := buffer[:dot]
	skipCat := func(cat int) {
		left = strings.TrimRightFunc(left, func(r rune) bool {
			return categorize(r) == cat
		})
	}

	// skip trailing whitespaces left of dot
	skipCat(0)

	// get category of last rune
	r, _ := utf8.DecodeLastRuneInString(left)
	cat := categorize(r)

	// skip this word
	skipCat(cat)

	return len(left)
}

// Move the dot right one word, using the word flavor described by the
// categorizer.
func moveDotRightGeneralWord(categorize categorizer, buffer string, dot int) int {
	right := buffer[dot:]
	skipCat := func(cat int) {
		right = strings.TrimLeftFunc(right, func(r rune) bool {
			return categorize(r) == cat
		})
	}

	// skip leading whitespaces right of dot
	skipCat(0)

	// check whether any whitespace was skipped; if whitespace was
	// skipped, then dot is already successfully moved to next
	// non-whitespace run
	if dot < len(buffer)-len(right) {
		return len(buffer) - len(right)
	}

	// no whitespace was skipped, so we still have to skip to the next word

	// get category of first rune
	r, _ := utf8.DecodeRuneInString(right)
	cat := categorize(r)
	// skip this word
	skipCat(cat)
	// skip remaining whitespace
	skipCat(0)

	return len(buffer) - len(right)
}

func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}
