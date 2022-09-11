package edit

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/strutil"
	"src.elv.sh/pkg/wcwidth"
)

func initBufferBuiltins(app cli.App, nb eval.NsBuilder) {
	m := make(map[string]any)
	for name, fn := range bufferBuiltinsData {
		// Make a lexically scoped copy of fn.
		fn := fn
		m[name] = func() {
			codeArea, ok := focusedCodeArea(app)
			if !ok {
				return
			}
			codeArea.MutateState(func(s *tk.CodeAreaState) {
				fn(&s.Buffer)
			})
		}
	}
	nb.AddGoFns(m)
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

	"transpose-rune":       makeTransform(transposeRunes),
	"transpose-word":       makeTransform(transposeWord),
	"transpose-small-word": makeTransform(transposeSmallWord),
	"transpose-alnum-word": makeTransform(transposeAlnumWord),
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

// A pure function that takes the current buffer and dot, and returns a new
// value for the buffer and dot.
type pureTransformer func(buffer string, dot int) (string, int)

func makeTransform(t pureTransformer) func(*tk.CodeBuffer) {
	return func(buf *tk.CodeBuffer) {
		buf.Content, buf.Dot = t(buf.Content, buf.Dot)
	}
}

// Implementation of pure movers.

//elvdoc:fn move-dot-left
//
// ```elvish
// edit:move-dot-left
// ```
//
// Moves the dot left one rune. Does nothing if the dot is at the beginning of
// the buffer.

//elvdoc:fn kill-rune-left
//
// ```elvish
// edit:kill-rune-left
// ```
//
// Kills one rune left of the dot. Does nothing if the dot is at the beginning of
// the buffer.

func moveDotLeft(buffer string, dot int) int {
	_, w := utf8.DecodeLastRuneInString(buffer[:dot])
	return dot - w
}

//elvdoc:fn move-dot-right
//
// ```elvish
// edit:move-dot-right
// ```
//
// Moves the dot right one rune. Does nothing if the dot is at the end of the
// buffer.

//elvdoc:fn kill-rune-left
//
// ```elvish
// edit:kill-rune-left
// ```
//
// Kills one rune right of the dot. Does nothing if the dot is at the end of the
// buffer.

func moveDotRight(buffer string, dot int) int {
	_, w := utf8.DecodeRuneInString(buffer[dot:])
	return dot + w
}

//elvdoc:fn move-dot-sol
//
// ```elvish
// edit:move-dot-sol
// ```
//
// Moves the dot to the start of the current line.

//elvdoc:fn kill-line-left
//
// ```elvish
// edit:kill-line-left
// ```
//
// Deletes the text between the dot and the start of the current line.

func moveDotSOL(buffer string, dot int) int {
	return strutil.FindLastSOL(buffer[:dot])
}

//elvdoc:fn move-dot-eol
//
// ```elvish
// edit:move-dot-eol
// ```
//
// Moves the dot to the end of the current line.

//elvdoc:fn kill-line-right
//
// ```elvish
// edit:kill-line-right
// ```
//
// Deletes the text between the dot and the end of the current line.

func moveDotEOL(buffer string, dot int) int {
	return strutil.FindFirstEOL(buffer[dot:]) + dot
}

//elvdoc:fn move-dot-up
//
// ```elvish
// edit:move-dot-up
// ```
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
// ```elvish
// edit:move-dot-down
// ```
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

//elvdoc:fn transpose-rune
//
// ```elvish
// edit:transpose-rune
// ```
//
// Swaps the runes to the left and right of the dot. If the dot is at the
// beginning of the buffer, swaps the first two runes, and if the dot is at the
// end, it swaps the last two.

func transposeRunes(buffer string, dot int) (string, int) {
	if len(buffer) == 0 {
		return buffer, dot
	}

	var newBuffer string
	var newDot int
	// transpose at the beginning of the buffer transposes the first two
	// characters, and at the end the last two
	if dot == 0 {
		first, firstLen := utf8.DecodeRuneInString(buffer)
		if firstLen == len(buffer) {
			return buffer, dot
		}
		second, secondLen := utf8.DecodeRuneInString(buffer[firstLen:])
		newBuffer = string(second) + string(first) + buffer[firstLen+secondLen:]
		newDot = firstLen + secondLen
	} else if dot == len(buffer) {
		second, secondLen := utf8.DecodeLastRuneInString(buffer)
		if secondLen == len(buffer) {
			return buffer, dot
		}
		first, firstLen := utf8.DecodeLastRuneInString(buffer[:len(buffer)-secondLen])
		newBuffer = buffer[:len(buffer)-firstLen-secondLen] + string(second) + string(first)
		newDot = len(newBuffer)
	} else {
		first, firstLen := utf8.DecodeLastRuneInString(buffer[:dot])
		second, secondLen := utf8.DecodeRuneInString(buffer[dot:])
		newBuffer = buffer[:dot-firstLen] + string(second) + string(first) + buffer[dot+secondLen:]
		newDot = dot + secondLen
	}

	return newBuffer, newDot
}

//elvdoc:fn move-dot-left-word
//
// ```elvish
// edit:move-dot-left-word
// ```
//
// Moves the dot to the beginning of the last word to the left of the dot.

//elvdoc:fn kill-word-left
//
// ```elvish
// edit:kill-word-left
// ```
//
// Deletes the last word to the left of the dot.

func moveDotLeftWord(buffer string, dot int) int {
	return moveDotLeftGeneralWord(categorizeWord, buffer, dot)
}

//elvdoc:fn move-dot-right-word
//
// ```elvish
// edit:move-dot-right-word
// ```
//
// Moves the dot to the beginning of the first word to the right of the dot.

//elvdoc:fn kill-word-right
//
// ```elvish
// edit:kill-word-right
// ```
//
// Deletes the first word to the right of the dot.

func moveDotRightWord(buffer string, dot int) int {
	return moveDotRightGeneralWord(categorizeWord, buffer, dot)
}

//elvdoc:fn transpose-word
//
// ```elvish
// edit:transpose-word
// ```
//
// Swaps the words to the left and right of the dot. If the dot is at the
// beginning of the buffer, swaps the first two words, and the dot is at the
// end, it swaps the last two.

func transposeWord(buffer string, dot int) (string, int) {
	return transposeGeneralWord(categorizeWord, buffer, dot)
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
// ```elvish
// edit:move-dot-left-small-word
// ```
//
// Moves the dot to the beginning of the last small word to the left of the dot.

//elvdoc:fn kill-small-word-left
//
// ```elvish
// edit:kill-small-word-left
// ```
//
// Deletes the last small word to the left of the dot.

func moveDotLeftSmallWord(buffer string, dot int) int {
	return moveDotLeftGeneralWord(tk.CategorizeSmallWord, buffer, dot)
}

//elvdoc:fn move-dot-right-small-word
//
// ```elvish
// edit:move-dot-right-small-word
// ```
//
// Moves the dot to the beginning of the first small word to the right of the dot.

//elvdoc:fn kill-small-word-right
//
// ```elvish
// edit:kill-small-word-right
// ```
//
// Deletes the first small word to the right of the dot.

func moveDotRightSmallWord(buffer string, dot int) int {
	return moveDotRightGeneralWord(tk.CategorizeSmallWord, buffer, dot)
}

//elvdoc:fn transpose-small-word
//
// ```elvish
// edit:transpose-small-word
// ```
//
// Swaps the small words to the left and right of the dot. If the dot is at the
// beginning of the buffer, it swaps the first two small words, and if the dot
// is at the end, it swaps the last two.

func transposeSmallWord(buffer string, dot int) (string, int) {
	return transposeGeneralWord(tk.CategorizeSmallWord, buffer, dot)
}

//elvdoc:fn move-dot-left-alnum-word
//
// ```elvish
// edit:move-dot-left-alnum-word
// ```
//
// Moves the dot to the beginning of the last alnum word to the left of the dot.

//elvdoc:fn kill-alnum-word-left
//
// ```elvish
// edit:kill-alnum-word-left
// ```
//
// Deletes the last alnum word to the left of the dot.

func moveDotLeftAlnumWord(buffer string, dot int) int {
	return moveDotLeftGeneralWord(categorizeAlnum, buffer, dot)
}

//elvdoc:fn move-dot-right-alnum-word
//
// ```elvish
// edit:move-dot-right-alnum-word
// ```
//
// Moves the dot to the beginning of the first alnum word to the right of the dot.

//elvdoc:fn kill-alnum-word-right
//
// ```elvish
// edit:kill-alnum-word-right
// ```
//
// Deletes the first alnum word to the right of the dot.

func moveDotRightAlnumWord(buffer string, dot int) int {
	return moveDotRightGeneralWord(categorizeAlnum, buffer, dot)
}

//elvdoc:fn transpose-alnum-word
//
// ```elvish
// edit:transpose-alnum-word
// ```
//
// Swaps the alnum words to the left and right of the dot. If the dot is at the
// beginning of the buffer, it swaps the first two alnum words, and if the dot
// is at the end, it swaps the last two.

func transposeAlnumWord(buffer string, dot int) (string, int) {
	return transposeGeneralWord(categorizeAlnum, buffer, dot)
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

// Move the dot left one word, using the word flavor described by the
// categorizer.
func moveDotLeftGeneralWord(categorize categorizer, buffer string, dot int) int {
	// skip trailing whitespaces left of dot
	pos := skipWsLeft(categorize, buffer, dot)

	// skip this word
	pos = skipSameCatLeft(categorize, buffer, pos)

	return pos
}

// Move the dot right one word, using the word flavor described by the
// categorizer.
func moveDotRightGeneralWord(categorize categorizer, buffer string, dot int) int {
	// skip leading whitespaces right of dot
	pos := skipWsRight(categorize, buffer, dot)

	if pos > dot {
		// Dot was within whitespaces, and we have now moved to the start of the
		// next word.
		return pos
	}

	// Dot was within a word; skip both the word and whitespaces

	// skip this word
	pos = skipSameCatRight(categorize, buffer, pos)
	// skip remaining whitespace
	pos = skipWsRight(categorize, buffer, pos)

	return pos
}

// Transposes the words around the cursor, using the word flavor described
// by the categorizer.
func transposeGeneralWord(categorize categorizer, buffer string, dot int) (string, int) {
	if strings.TrimFunc(buffer, func(r rune) bool { return categorize(r) == 0 }) == "" {
		// buffer contains only whitespace
		return buffer, dot
	}

	// after skipping whitespace, find the end of the right word
	pos := skipWsRight(categorize, buffer, dot)
	var rightEnd int
	if pos == len(buffer) {
		// there is only whitespace to the right of the dot
		rightEnd = skipWsLeft(categorize, buffer, pos)
	} else {
		rightEnd = skipSameCatRight(categorize, buffer, pos)
	}
	// if the dot started in the middle of a word, 'pos' is the same as dot,
	// so we should skip word characters to the left to find the start of the
	// word
	rightStart := skipSameCatLeft(categorize, buffer, rightEnd)

	leftEnd := skipWsLeft(categorize, buffer, rightStart)
	var leftStart int
	if leftEnd == 0 {
		// right word is the first word, use it as the left word and find a
		// new right word
		leftStart = rightStart
		leftEnd = rightEnd

		rightStart = skipWsRight(categorize, buffer, leftEnd)
		if rightStart == len(buffer) {
			// there is only one word in the buffer
			return buffer, dot
		}

		rightEnd = skipSameCatRight(categorize, buffer, rightStart)
	} else {
		leftStart = skipSameCatLeft(categorize, buffer, leftEnd)
	}

	return buffer[:leftStart] + buffer[rightStart:rightEnd] + buffer[leftEnd:rightStart] + buffer[leftStart:leftEnd] + buffer[rightEnd:], rightEnd
}

// Skips all runes to the left of the dot that belongs to the same category.
func skipSameCatLeft(categorize categorizer, buffer string, pos int) int {
	if pos == 0 {
		return pos
	}

	r, _ := utf8.DecodeLastRuneInString(buffer[:pos])
	cat := categorize(r)
	return skipCatLeft(categorize, cat, buffer, pos)
}

// Skips whitespaces to the left of the dot.
func skipWsLeft(categorize categorizer, buffer string, pos int) int {
	return skipCatLeft(categorize, 0, buffer, pos)
}

func skipCatLeft(categorize categorizer, cat int, buffer string, pos int) int {
	left := strings.TrimRightFunc(buffer[:pos], func(r rune) bool {
		return categorize(r) == cat
	})

	return len(left)
}

// Skips all runes to the right of the dot that belongs to the same
// category.
func skipSameCatRight(categorize categorizer, buffer string, pos int) int {
	if pos == len(buffer) {
		return pos
	}

	r, _ := utf8.DecodeRuneInString(buffer[pos:])
	cat := categorize(r)
	return skipCatRight(categorize, cat, buffer, pos)
}

// Skips whitespaces to the right of the dot.
func skipWsRight(categorize categorizer, buffer string, pos int) int {
	return skipCatRight(categorize, 0, buffer, pos)
}

func skipCatRight(categorize categorizer, cat int, buffer string, pos int) int {
	right := strings.TrimLeftFunc(buffer[pos:], func(r rune) bool {
		return categorize(r) == cat
	})

	return len(buffer) - len(right)
}
