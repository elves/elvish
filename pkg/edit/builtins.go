package edit

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/parseutil"
	"src.elv.sh/pkg/strutil"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/wcwidth"
)

//elvdoc:fn binding-table
//
// Converts a normal map into a binding map.

//elvdoc:fn -dump-buf
//
// Dumps the current UI buffer as HTML. This command is used to generate
// "ttyshots" on the [website](https://elv.sh).
//
// Example:
//
// ```elvish
// ttyshot = ~/a.html
// edit:insert:binding[Ctrl-X] = { edit:-dump-buf > $tty }
// ```

func dumpBuf(tty cli.TTY) string {
	return bufToHTML(tty.Buffer())
}

//elvdoc:fn close-mode
//
// Closes the current active mode.

func closeMode(app cli.App) {
	app.PopAddon()
}

//elvdoc:fn end-of-history
//
// Adds a notification saying "End of history".

func endOfHistory(app cli.App) {
	app.Notify("End of history")
}

//elvdoc:fn redraw
//
// ```elvish
// edit:redraw &full=$false
// ```
//
// Triggers a redraw.
//
// The `&full` option controls whether to do a full redraw. By default, all
// redraws performed by the line editor are incremental redraws, updating only
// the part of the screen that has changed from the last redraw. A full redraw
// updates the entire command line.

type redrawOpts struct{ Full bool }

func (redrawOpts) SetDefaultOptions() {}

func redraw(app cli.App, opts redrawOpts) {
	if opts.Full {
		app.RedrawFull()
	} else {
		app.Redraw()
	}
}

//elvdoc:fn clear
//
// ```elvish
// edit:clear
// ```
//
// Clears the screen.
//
// This command should be used in place of the external `clear` command to clear
// the screen.

func clear(app cli.App, tty cli.TTY) {
	tty.HideCursor()
	tty.ClearScreen()
	app.RedrawFull()
	tty.ShowCursor()
}

//elvdoc:fn insert-raw
//
// Requests the next terminal input to be inserted uninterpreted.

func insertRaw(app cli.App, tty cli.TTY) {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return
	}
	tty.SetRawInput(1)
	w := modes.NewStub(modes.StubSpec{
		Bindings: tk.FuncBindings(func(w tk.Widget, event term.Event) bool {
			switch event := event.(type) {
			case term.KeyEvent:
				codeArea.MutateState(func(s *tk.CodeAreaState) {
					s.Buffer.InsertAtDot(string(event.Rune))
				})
				app.PopAddon()
				return true
			default:
				return false
			}
		}),
		Name: " RAW ",
	})
	app.PushAddon(w)
}

//elvdoc:fn key
//
// ```elvish
// edit:key $string
// ```
//
// Parses a string into a key.

var errMustBeKeyOrString = errors.New("must be key or string")

func toKey(v interface{}) (ui.Key, error) {
	switch v := v.(type) {
	case ui.Key:
		return v, nil
	case string:
		return ui.ParseKey(v)
	default:
		return ui.Key{}, errMustBeKeyOrString
	}
}

//elvdoc:fn notify
//
// ```elvish
// edit:notify $message
// ```
//
// Prints a notification message.
//
// If called while the editor is active, this will print the message above the
// editor, and redraw the editor.
//
// If called while the editor is inactive, the message will be queued, and shown
// once the editor becomes active.

//elvdoc:fn return-line
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read. If called from a key binding, takes effect after the key
// binding returns.

//elvdoc:fn return-eof
//
// Causes the Elvish REPL to terminate. If called from a key binding, takes
// effect after the key binding returns.

//elvdoc:fn smart-enter
//
// Inserts a literal newline if the current code is not syntactically complete
// Elvish code. Accepts the current line otherwise.

func smartEnter(app cli.App) {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return
	}
	commit := false
	codeArea.MutateState(func(s *tk.CodeAreaState) {
		buf := &s.Buffer
		if isSyntaxComplete(buf.Content) {
			commit = true
		} else {
			buf.InsertAtDot("\n")
		}
	})
	if commit {
		app.CommitCode()
	}
}

func isSyntaxComplete(code string) bool {
	_, err := parse.Parse(parse.Source{Code: code}, parse.Config{})
	if err != nil {
		for _, e := range err.(*parse.Error).Entries {
			if e.Context.From == len(code) {
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

func wordify(fm *eval.Frame, code string) error {
	out := fm.ValueOutput()
	for _, s := range parseutil.Wordify(code) {
		err := out.Put(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func initTTYBuiltins(app cli.App, tty cli.TTY, nb eval.NsBuilder) {
	nb.AddGoFns("<edit>", map[string]interface{}{
		"-dump-buf":  func() string { return dumpBuf(tty) },
		"insert-raw": func() { insertRaw(app, tty) },
		"clear":      func() { clear(app, tty) },
	})
}

func initMiscBuiltins(app cli.App, nb eval.NsBuilder) {
	nb.AddGoFns("<edit>", map[string]interface{}{
		"binding-table":  makeBindingMap,
		"close-mode":     func() { closeMode(app) },
		"end-of-history": func() { endOfHistory(app) },
		"key":            toKey,
		"notify":         app.Notify,
		"redraw":         func(opts redrawOpts) { redraw(app, opts) },
		"return-line":    app.CommitCode,
		"return-eof":     app.CommitEOF,
		"smart-enter":    func() { smartEnter(app) },
		"wordify":        wordify,
	})
}

var bufferBuiltinsData = map[string]func(*tk.CodeBuffer){
	"move-dot-left":             makeMove(moveDotLeft),
	"move-dot-right":            makeMove(moveDotRight),
	"move-dot-left-word":        makeMove(moveDotLeftWord),
	"move-dot-right-word":       makeMove(moveDotRightWord),
	"move-dot-left-small-word":  makeMove(moveDotLeftSmallWord),
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
	"kill-small-word-left":  makeKill(moveDotLeftSmallWord),
	"kill-small-word-right": makeKill(moveDotRightSmallWord),
	"kill-left-alnum-word":  makeKill(moveDotLeftAlnumWord),
	"kill-right-alnum-word": makeKill(moveDotRightAlnumWord),
	"kill-line-left":        makeKill(moveDotSOL),
	"kill-line-right":       makeKill(moveDotEOL),
}

func initBufferBuiltins(app cli.App, nb eval.NsBuilder) {
	nb.AddGoFns("<edit>", bufferBuiltins(app))
}

func bufferBuiltins(app cli.App) map[string]interface{} {
	m := make(map[string]interface{})
	for name, fn := range bufferBuiltinsData {
		// Make a lexically scoped copy of fn.
		fn2 := fn
		m[name] = func() {
			codeArea, ok := focusedCodeArea(app)
			if !ok {
				return
			}
			codeArea.MutateState(func(s *tk.CodeAreaState) {
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

func makeMove(m pureMover) func(*tk.CodeBuffer) {
	return func(buf *tk.CodeBuffer) {
		buf.Dot = m(buf.Content, buf.Dot)
	}
}

func makeKill(m pureMover) func(*tk.CodeBuffer) {
	return func(buf *tk.CodeBuffer) {
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
	return strutil.FindLastSOL(buffer[:dot])
}

//elvdoc:fn move-dot-eol
//
// Moves the dot to the end of the current line.

//elvdoc:fn kill-line-right
//
// Deletes the text between the dot and the end of the current line.

func moveDotEOL(buffer string, dot int) int {
	return strutil.FindFirstEOL(buffer[dot:]) + dot
}

//elvdoc:fn move-dot-up
//
// Moves the dot up one line, trying to preserve the visual horizontal position.
// Does nothing if dot is already on the first line of the buffer.

func moveDotUp(buffer string, dot int) int {
	sol := strutil.FindLastSOL(buffer[:dot])
	if sol == 0 {
		// Already in the first line.
		return dot
	}
	prevEOL := sol - 1
	prevSOL := strutil.FindLastSOL(buffer[:prevEOL])
	width := wcwidth.Of(buffer[sol:dot])
	return prevSOL + len(wcwidth.Trim(buffer[prevSOL:prevEOL], width))
}

//elvdoc:fn move-dot-down
//
// Moves the dot down one line, trying to preserve the visual horizontal
// position. Does nothing if dot is already on the last line of the buffer.

func moveDotDown(buffer string, dot int) int {
	eol := strutil.FindFirstEOL(buffer[dot:]) + dot
	if eol == len(buffer) {
		// Already in the last line.
		return dot
	}
	nextSOL := eol + 1
	nextEOL := strutil.FindFirstEOL(buffer[nextSOL:]) + nextSOL
	sol := strutil.FindLastSOL(buffer[:dot])
	width := wcwidth.Of(buffer[sol:dot])
	return nextSOL + len(wcwidth.Trim(buffer[nextSOL:nextEOL], width))
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
	return moveDotLeftGeneralWord(tk.CategorizeSmallWord, buffer, dot)
}

//elvdoc:fn move-dot-right-small-word
//
// Moves the dot to the beginning of the first small word to the right of the dot.

//elvdoc:fn kill-small-word-right
//
// Deletes the the first small word to the right of the dot.

func moveDotRightSmallWord(buffer string, dot int) int {
	return moveDotRightGeneralWord(tk.CategorizeSmallWord, buffer, dot)
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
	case tk.IsAlnum(r):
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

// Skip only word characters left from pos in buffer and return that position.
//
// Note that a word character must be to the left of pos.
func (categorize categorizer) skipWordLeft(buffer string, pos int) int {
	if pos == 0 {
		return pos
	}

	r, _ := utf8.DecodeLastRuneInString(buffer[:pos])
	cat := categorize(r)
	// assert cat is not 0
	return categorize.skipCatLeft(cat, buffer, pos)
}

// Skip whitespace left from pos in buffer and return that position.
func (categorize categorizer) skipWsLeft(buffer string, pos int) int {
	return categorize.skipCatLeft(0, buffer, pos)
}

func (categorize categorizer) skipCatLeft(cat int, buffer string, pos int) int {
	left := strings.TrimRightFunc(buffer[:pos], func(r rune) bool {
		return categorize(r) == cat
	})

	return len(left)
}

// Skip only word characters right from pos in buffer and return that position.
//
// Note that a word character must be to the right of pos.
func (categorize categorizer) skipWordRight(buffer string, pos int) int {
	if pos == len(buffer) {
		return pos
	}

	r, _ := utf8.DecodeRuneInString(buffer[pos:])
	cat := categorize(r)
	// assert cat is not 0
	return categorize.skipCatRight(cat, buffer, pos)
}

// Skip whitespace right from pos in buffer and return that position.
func (categorize categorizer) skipWsRight(buffer string, pos int) int {
	return categorize.skipCatRight(0, buffer, pos)
}

func (categorize categorizer) skipCatRight(cat int, buffer string, pos int) int {
	right := strings.TrimLeftFunc(buffer[pos:], func(r rune) bool {
		return categorize(r) == cat
	})

	return len(buffer) - len(right)
}

// Move the dot left one word, using the word flavor described by the
// categorizer.
func moveDotLeftGeneralWord(categorize categorizer, buffer string, dot int) int {
	// skip trailing whitespaces left of dot
	pos := categorize.skipWsLeft(buffer, dot)

	// skip this word
	pos = categorize.skipWordLeft(buffer, pos)

	return pos
}

// Move the dot right one word, using the word flavor described by the
// categorizer.
func moveDotRightGeneralWord(categorize categorizer, buffer string, dot int) int {
	// skip leading whitespaces right of dot
	pos := categorize.skipWsRight(buffer, dot)

	// check whether any whitespace was skipped; if whitespace was
	// skipped, then dot is already successfully moved to next
	// non-whitespace run
	if dot < pos {
		return pos
	}

	// no whitespace was skipped, so we still have to skip to the next word

	// skip this word
	pos = categorize.skipWordRight(buffer, pos)
	// skip remaining whitespace
	pos = categorize.skipWsRight(buffer, pos)

	return pos
}

// Like mode.FocusedCodeArea, but handles the error by writing a notification.
func focusedCodeArea(app cli.App) (tk.CodeArea, bool) {
	codeArea, err := modes.FocusedCodeArea(app)
	if err != nil {
		app.Notify(err.Error())
		return nil, false
	}
	return codeArea, true
}
