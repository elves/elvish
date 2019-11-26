package cliedit

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/stub"
	"github.com/elves/elvish/cli/cliutil"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/parse/parseutil"
	"github.com/elves/elvish/ui"
	"github.com/elves/elvish/util"
)

//elvdoc:fn binding-table
//
// Converts a normal map into a binding map.

//elvdoc:fn -dump-buf
//
// Dumps the current UI buffer as HTML.

func dumpBuf(tty cli.TTY) string {
	return bufToHTML(tty.Buffer())
}

//elvdoc:fn close-listing
//
// Closes any active listing.

func closeListing(app cli.App) {
	app.MutateState(func(s *cli.State) { s.Addon = nil })
}

//elvdoc:fn end-of-history
//
// Adds a notification saying "End of history".

func endOfHistory(app cli.App) {
	app.Notify("End of history")
}

//elvdoc:fn insert-raw
//
// Requests the next terminal input to be inserted uninterpreted.

func insertRaw(app cli.App, tty cli.TTY) {
	tty.SetRawInput(1)
	stub.Start(app, stub.Config{
		Binding: el.FuncHandler(func(event term.Event) bool {
			switch event := event.(type) {
			case term.KeyEvent:
				app.CodeArea().MutateState(func(s *codearea.State) {
					s.Buffer.InsertAtDot(string(event.Rune))
				})
				closeListing(app)
				return true
			default:
				return false
			}
		}),
		Name:  " RAW ",
		Focus: false,
	})
}

//elvdoc:fn key
//
// ```elvish
// edit:key $string
// ```
//
// Parses a string into a key.

//elvdoc:fn redraw
//
// Triggers a redraw.

//elvdoc:fn return-line
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read.

//elvdoc:fn return-eof
//
// Causes the Elvish REPL to terminate. Internally, this works by raising a
// special exception.

//elvdoc:fn smart-enter
//
// Inserts a literal newline if the code before the dot is not syntactically
// complete Elvish code. Accepts the current line otherwise.

func smartEnter(app cli.App) {
	// TODO(xiaq): Fix the race condition.
	buf := cliutil.GetCodeBuffer(app)
	if isSyntaxComplete(buf.Content[:buf.Dot]) {
		app.CommitCode()
	} else {
		app.CodeArea().MutateState(func(s *codearea.State) {
			s.Buffer.InsertAtDot("\n")
		})
	}
}

func isSyntaxComplete(code string) bool {
	_, err := parse.AsChunk("[interactive]", code)
	if err != nil {
		for _, e := range err.(parse.MultiError).Entries {
			if e.Context.Begin == len(code) {
				return false
			}
		}
	}
	return true
}

//elvdoc:fn wordify
//
//
// ```elvish
// edit:wordify $code
// ```
// Breaks Elvish code into words.

func wordify(fm *eval.Frame, code string) {
	out := fm.OutputChan()
	for _, s := range parseutil.Wordify(code) {
		out <- s
	}
}

func initTTYBuiltins(app cli.App, tty cli.TTY, ns eval.Ns) {
	ns.AddGoFns("<edit>", map[string]interface{}{
		"-dump-buf":  func() string { return dumpBuf(tty) },
		"insert-raw": func() { insertRaw(app, tty) },
	})
}

func initMiscBuiltins(app cli.App, ns eval.Ns) {
	ns.AddGoFns("<edit>", map[string]interface{}{
		"binding-table":  MakeBindingMap,
		"close-listing":  func() { closeListing(app) },
		"end-of-history": func() { endOfHistory(app) },
		"key":            ui.ToKey,
		"redraw":         app.Redraw,
		"return-line":    app.CommitCode,
		"return-eof":     app.CommitEOF,
		"smart-enter":    func() { smartEnter(app) },
		"wordify":        wordify,
	})
}

var bufferBuiltinsData = map[string]func(*codearea.Buffer){
	"move-dot-left":             makeMove(moveDotLeft),
	"move-dot-right":            makeMove(moveDotRight),
	"move-dot-left-word":        makeMove(moveDotLeftWord),
	"move-dot-right-word":       makeMove(moveDotRightWord),
	"move-dot-left-small-word":  makeMove(moveDotLeftWord),
	"move-dot-right-small-word": makeMove(moveDotRightSmallWord),
	"move-dot-left-alnum-word":  makeMove(moveDotLeftAlnumWord),
	"move-dot-right-alnum-word": makeMove(moveDotRightAlnumWord),
	"move-dot-sol":              makeMove(moveDotSOL),
	"move-dot-eol":              makeMove(moveDotEOL),

	"move-dot-up":   makeMove(moveDotUp),
	"move-dot-down": makeMove(moveDotDown),

	"kill-rune-left":        makeKill(moveDotLeft),
	"kill-rune-right":       makeKill(moveDotRight),
	"kill-word-left":        makeKill(moveDotLeftWord),
	"kill-word-right":       makeKill(moveDotRightWord),
	"kill-small-word-left":  makeKill(moveDotLeftWord),
	"kill-small-word-right": makeKill(moveDotRightSmallWord),
	"kill-left-alnum-word":  makeKill(moveDotLeftAlnumWord),
	"kill-right-alnum-word": makeKill(moveDotRightAlnumWord),
	"kill-line-left":        makeKill(moveDotSOL),
	"kill-line-right":       makeKill(moveDotEOL),
}

func initBufferBuiltins(app cli.App, ns eval.Ns) {
	ns.AddGoFns("<edit>", bufferBuiltins(app))
}

func bufferBuiltins(app cli.App) map[string]interface{} {
	m := make(map[string]interface{})
	for name, fn := range bufferBuiltinsData {
		// Make a lexically scoped copy of fn.
		fn2 := fn
		m[name] = func() {
			app.CodeArea().MutateState(func(s *codearea.State) {
				fn2(&s.Buffer)
			})
		}
	}
	return m
}

// A pure function that takes the current buffer and dot, and returns a new
// value for the dot. Used to derive move- and kill- functions that operate on
// the editor state.
type pureMover func(buffer string, dot int) int

func makeMove(m pureMover) func(*codearea.Buffer) {
	return func(buf *codearea.Buffer) {
		buf.Dot = m(buf.Content, buf.Dot)
	}
}

func makeKill(m pureMover) func(*codearea.Buffer) {
	return func(buf *codearea.Buffer) {
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

//elvdoc:fn move-dot-left
//
// Moves the dot left one rune. Does nothing if the dot is at the beginning of
// the buffer.

//elvdoc:fn kill-rune-left
//
// Kills one rune left of the dot. Does nothing if the dot is at the beginning of
// the buffer.

func moveDotLeft(buffer string, dot int) int {
	_, w := utf8.DecodeLastRuneInString(buffer[:dot])
	return dot - w
}

//elvdoc:fn move-dot-right
//
// Moves the dot right one rune. Does nothing if the dot is at the end of the
// buffer.

//elvdoc:fn kill-rune-left
//
// Kills one rune right of the dot. Does nothing if the dot is at the end of the
// buffer.

func moveDotRight(buffer string, dot int) int {
	_, w := utf8.DecodeRuneInString(buffer[dot:])
	return dot + w
}

//elvdoc:fn move-dot-sol
//
// Moves the dot to the start of the current line.

//elvdoc:fn kill-line-left
//
// Deletes the text between the dot and the start of the current line.

func moveDotSOL(buffer string, dot int) int {
	return util.FindLastSOL(buffer[:dot])
}

//elvdoc:fn move-dot-eol
//
// Moves the dot to the end of the current line.

//elvdoc:fn kill-line-right
//
// Deletes the text between the dot and the end of the current line.

func moveDotEOL(buffer string, dot int) int {
	return util.FindFirstEOL(buffer[dot:]) + dot
}

//elvdoc:fn move-dot-up
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

//elvdoc:fn move-dot-down
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

//elvdoc:fn move-dot-left-word
//
// Moves the dot to the beginning of the last word to the left of the dot.

//elvdoc:fn kill-word-left
//
// Deletes the the last word to the left of the dot.

func moveDotLeftWord(buffer string, dot int) int {
	return moveDotLeftGeneralWord(categorizeWord, buffer, dot)
}

//elvdoc:fn move-dot-right-word
//
// Moves the dot to the beginning of the first word to the right of the dot.

//elvdoc:fn kill-word-right
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

//elvdoc:fn move-dot-left-small-word
//
// Moves the dot to the beginning of the last small word to the left of the dot.

//elvdoc:fn kill-small-word-left
//
// Deletes the the last small word to the left of the dot.

func moveDotLeftSmallWord(buffer string, dot int) int {
	return moveDotLeftGeneralWord(categorizeSmallWord, buffer, dot)
}

//elvdoc:fn move-dot-right-small-word
//
// Moves the dot to the beginning of the first small word to the right of the dot.

//elvdoc:fn kill-small-word-right
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

//elvdoc:fn move-dot-left-alnum-word
//
// Moves the dot to the beginning of the last alnum word to the left of the dot.

//elvdoc:fn kill-alnum-word-left
//
// Deletes the the last alnum word to the left of the dot.

func moveDotLeftAlnumWord(buffer string, dot int) int {
	return moveDotLeftGeneralWord(categorizeAlnum, buffer, dot)
}

//elvdoc:fn move-dot-right-alnum-word
//
// Moves the dot to the beginning of the first alnum word to the right of the dot.

//elvdoc:fn kill-alnum-word-right
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
