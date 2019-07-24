package codearea

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
)

// Widget supports code-editing functions. It implements the clitypes.Widget
// interface. An empty Widget is directly usable.
type Widget struct {
	// Publically accessible state.
	State State
	// A function that highlights the given code and returns any errors it has
	// found when highlighting. If this function is not given, the Widget does
	// not highlight the code nor show any errors.
	Highlighter func(code string) (styled.Text, []error)
	// A function that calls the callback with string pairs for abbreviations
	// and their expansions. If this function is not given, the Widget does not
	// expand any abbreviations.
	Abbreviations func(f func(abbr, full string))
	// A function that returns whether pasted texts (from bracketed pastes)
	// should be quoted. If this function is not given, the Widget defaults to
	// not quoting pasted texts.
	QuotePaste func() bool
	// A function that is called on the submit event.
	OnSubmit func(code string)

	// Consecutively inserted text. Used for expanding abbreviations.
	inserts string
	// Value of State.CodeBuffer when handleKeyEvent was last called. Used for
	// detecting whether insertion has been interrupted.
	lastCodeBuffer CodeBuffer
	// Whether the widget is in the middle of bracketed pasting.
	pasting bool
	// Buffer for keeping Pasted text during bracketed pasting.
	pasteBuffer bytes.Buffer
}

func dummyHighlighter(code string) (styled.Text, []error) {
	return styled.Plain(code), nil
}

func dummyAbbreviations(func(a, f string)) {}

func dummyQuotePaste() bool { return false }

func dummyOnSubmit(string) {}

// Initializes nil members to sensible default values. This method is called
// at the beginning of all public methods.
func (w *Widget) init() {
	if w.Highlighter == nil {
		w.Highlighter = dummyHighlighter
	}
	if w.Abbreviations == nil {
		w.Abbreviations = dummyAbbreviations
	}
	if w.QuotePaste == nil {
		w.QuotePaste = dummyQuotePaste
	}
	if w.OnSubmit == nil {
		w.OnSubmit = dummyOnSubmit
	}
}

// Submit emits a submit event with the current code content.
func (w *Widget) Submit() {
	w.init()
	w.State.Mutex.RLock()
	defer w.State.Mutex.RUnlock()
	w.OnSubmit(w.State.CodeBuffer.Content)
}

// Show shows the code area, including the prompt and rprompt, highlighted code,
// the cursor, and compilation errors in the code content.
func (w *Widget) Render(bb *ui.BufferBuilder) {
	w.init()
	s := w.State
	s.Mutex.RLock()
	styledCode, errors := w.Highlighter(s.CodeBuffer.Content)
	v := &view{
		s.Prompt, s.RPrompt, styledCode, s.CodeBuffer.Dot, errors,
	}
	s.Mutex.RUnlock()

	render(v, bb)
}

// Handle handles KeyEvent's of non-function keys, as well as PasteSetting
// events.
func (w *Widget) Handle(event term.Event) bool {
	w.init()

	switch event := event.(type) {
	case term.PasteSetting:
		return w.handlePasteSetting(bool(event))
	case term.KeyEvent:
		return w.handleKeyEvent(ui.Key(event))
	}
	return false
}

func (w *Widget) resetInserts() {
	w.inserts = ""
	w.lastCodeBuffer = CodeBuffer{}
}

func (w *Widget) handlePasteSetting(start bool) bool {
	w.resetInserts()
	if start {
		w.pasting = true
	} else {
		text := w.pasteBuffer.String()
		if w.QuotePaste() {
			text = parse.Quote(text)
		}
		w.State.MutateCodeBuffer(func(c *CodeBuffer) { c.InsertAtDot(text) })

		w.pasting = false
		w.pasteBuffer = bytes.Buffer{}
	}
	return true
}

func (w *Widget) handleKeyEvent(key ui.Key) bool {
	isFuncKey := key.Mod != 0 || key.Rune < 0
	if w.pasting {
		if isFuncKey {
			// TODO: Notify the user of the error, or insert the original
			// character as is.
		} else {
			w.pasteBuffer.WriteRune(key.Rune)
		}
		return true
	}
	// We only implement essential keybindings here. Other keybindings can be
	// added via handler overlays.
	switch key {
	case ui.K('\n'):
		w.resetInserts()
		w.Submit()
		return true
	case ui.K(ui.Backspace):
		w.resetInserts()
		w.State.MutateCodeBuffer(func(c *CodeBuffer) {
			// Remove the last rune.
			_, chop := utf8.DecodeLastRuneInString(c.Content[:c.Dot])
			*c = CodeBuffer{
				Content: c.Content[:c.Dot-chop] + c.Content[c.Dot:],
				Dot:     c.Dot - chop,
			}
		})
		return true
	default:
		if isFuncKey || !unicode.IsGraphic(key.Rune) {
			w.resetInserts()
			return false
		}
		w.State.Mutex.Lock()
		defer w.State.Mutex.Unlock()
		if w.lastCodeBuffer != w.State.CodeBuffer {
			// Something has happened between the last insert and this one;
			// reset the state.
			w.resetInserts()
		}
		s := string(key.Rune)
		w.State.CodeBuffer.InsertAtDot(s)
		w.inserts += s
		w.lastCodeBuffer = w.State.CodeBuffer
		var abbr, full string
		// Try to expand an abbreviation, preferring the longest one
		w.Abbreviations(func(a, f string) {
			if strings.HasSuffix(w.inserts, a) && len(a) > len(abbr) {
				abbr, full = a, f
			}
		})
		if len(abbr) > 0 {
			c := &w.State.CodeBuffer
			*c = CodeBuffer{
				Content: c.Content[:c.Dot-len(abbr)] + full + c.Content[c.Dot:],
				Dot:     c.Dot - len(abbr) + len(full),
			}
			w.resetInserts()
		}
		return true
	}
}
