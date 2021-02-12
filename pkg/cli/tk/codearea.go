package tk

import (
	"bytes"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

// CodeArea is a Widget for displaying and editing code.
type CodeArea interface {
	Widget
	// CopyState returns a copy of the state.
	CopyState() CodeAreaState
	// MutateState calls the given the function while locking StateMutex.
	MutateState(f func(*CodeAreaState))
	// Submit triggers the OnSubmit callback.
	Submit()
}

// CodeAreaSpec specifies the configuration and initial state for CodeArea.
type CodeAreaSpec struct {
	// Key bindings.
	Bindings Bindings
	// A function that highlights the given code and returns any errors it has
	// found when highlighting. If this function is not given, the Widget does
	// not highlight the code nor show any errors.
	Highlighter func(code string) (ui.Text, []error)
	// Prompt callback.
	Prompt func() ui.Text
	// Right-prompt callback.
	RPrompt func() ui.Text
	// A function that calls the callback with string pairs for abbreviations
	// and their expansions. If this function is not given, the Widget does not
	// expand any abbreviations.
	Abbreviations          func(f func(abbr, full string))
	SmallWordAbbreviations func(f func(abbr, full string))
	// A function that returns whether pasted texts (from bracketed pastes)
	// should be quoted. If this function is not given, the Widget defaults to
	// not quoting pasted texts.
	QuotePaste func() bool
	// A function that is called on the submit event.
	OnSubmit func()

	// State. When used in New, this field specifies the initial state.
	State CodeAreaState
}

// CodeAreaState keeps the mutable state of the CodeArea widget.
type CodeAreaState struct {
	Buffer      CodeBuffer
	Pending     PendingCode
	HideRPrompt bool
}

// CodeBuffer represents the buffer of the CodeArea widget.
type CodeBuffer struct {
	// Content of the buffer.
	Content string
	// Position of the dot (more commonly known as the cursor), as a byte index
	// into Content.
	Dot int
}

// PendingCode represents pending code, such as during completion.
type PendingCode struct {
	// Beginning index of the text area that the pending code replaces, as a
	// byte index into RawState.Code.
	From int
	// End index of the text area that the pending code replaces, as a byte
	// index into RawState.Code.
	To int
	// The content of the pending code.
	Content string
}

// ApplyPending applies pending code to the code buffer, and resets pending code.
func (s *CodeAreaState) ApplyPending() {
	s.Buffer, _, _ = patchPending(s.Buffer, s.Pending)
	s.Pending = PendingCode{}
}

func (c *CodeBuffer) InsertAtDot(text string) {
	*c = CodeBuffer{
		Content: c.Content[:c.Dot] + text + c.Content[c.Dot:],
		Dot:     c.Dot + len(text),
	}
}

type codeArea struct {
	// Mutex for synchronizing access to State.
	StateMutex sync.RWMutex
	// Configuration and state.
	CodeAreaSpec

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

// NewCodeArea creates a new CodeArea from the given spec.
func NewCodeArea(spec CodeAreaSpec) CodeArea {
	if spec.Bindings == nil {
		spec.Bindings = DummyBindings{}
	}
	if spec.Highlighter == nil {
		spec.Highlighter = func(s string) (ui.Text, []error) { return ui.T(s), nil }
	}
	if spec.Prompt == nil {
		spec.Prompt = func() ui.Text { return nil }
	}
	if spec.RPrompt == nil {
		spec.RPrompt = func() ui.Text { return nil }
	}
	if spec.Abbreviations == nil {
		spec.Abbreviations = func(func(a, f string)) {}
	}
	if spec.SmallWordAbbreviations == nil {
		spec.SmallWordAbbreviations = func(func(a, f string)) {}
	}
	if spec.QuotePaste == nil {
		spec.QuotePaste = func() bool { return false }
	}
	if spec.OnSubmit == nil {
		spec.OnSubmit = func() {}
	}
	return &codeArea{CodeAreaSpec: spec}
}

// Submit emits a submit event with the current code content.
func (w *codeArea) Submit() {
	w.OnSubmit()
}

// Render renders the code area, including the prompt and rprompt, highlighted
// code, the cursor, and compilation errors in the code content.
func (w *codeArea) Render(width, height int) *term.Buffer {
	view := getView(w)
	bb := term.NewBufferBuilder(width)
	renderView(view, bb)
	b := bb.Buffer()
	truncateToHeight(b, height)
	return b
}

// Handle handles KeyEvent's of non-function keys, as well as PasteSetting
// events.
func (w *codeArea) Handle(event term.Event) bool {
	switch event := event.(type) {
	case term.PasteSetting:
		return w.handlePasteSetting(bool(event))
	case term.KeyEvent:
		return w.handleKeyEvent(ui.Key(event))
	}
	return false
}

func (w *codeArea) MutateState(f func(*CodeAreaState)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}

func (w *codeArea) CopyState() CodeAreaState {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}

func (w *codeArea) resetInserts() {
	w.inserts = ""
	w.lastCodeBuffer = CodeBuffer{}
}

func (w *codeArea) handlePasteSetting(start bool) bool {
	w.resetInserts()
	if start {
		w.pasting = true
	} else {
		text := w.pasteBuffer.String()
		if w.QuotePaste() {
			text = parse.Quote(text)
		}
		w.MutateState(func(s *CodeAreaState) { s.Buffer.InsertAtDot(text) })

		w.pasting = false
		w.pasteBuffer = bytes.Buffer{}
	}
	return true
}

// Tries to expand a simple abbreviation. This function assumes that the state
// mutex is already being held.
func (w *codeArea) expandSimpleAbbr() {
	var abbr, full string
	// Find the longest matching abbreviation.
	w.Abbreviations(func(a, f string) {
		if strings.HasSuffix(w.inserts, a) && len(a) > len(abbr) {
			abbr, full = a, f
		}
	})
	if len(abbr) > 0 {
		c := &w.State.Buffer
		*c = CodeBuffer{
			Content: c.Content[:c.Dot-len(abbr)] + full + c.Content[c.Dot:],
			Dot:     c.Dot - len(abbr) + len(full),
		}
		w.resetInserts()
	}
}

// Tries to expand a word abbreviation. This function assumes that the state
// mutex is already being held.
func (w *codeArea) expandWordAbbr(trigger rune, categorizer func(rune) int) {
	c := &w.State.Buffer
	if c.Dot < len(c.Content) {
		// Word abbreviations are only expanded at the end of the buffer.
		return
	}
	triggerLen := len(string(trigger))
	if triggerLen >= len(w.inserts) {
		// Only the trigger has been inserted, or a simple abbreviation was just
		// expanded. In either case, there is nothing to expand.
		return
	}
	// The trigger is only used to determine word boundary; when considering
	// what to expand, we only consider the part that was inserted before it.
	inserts := w.inserts[:len(w.inserts)-triggerLen]

	var abbr, full string
	// Find the longest matching abbreviation.
	w.SmallWordAbbreviations(func(a, f string) {
		if len(a) <= len(abbr) {
			// This abbreviation can't be the longest.
			return
		}
		if !strings.HasSuffix(inserts, a) {
			// This abbreviation was not inserted.
			return
		}
		// Verify the trigger rune creates a word boundary.
		r, _ := utf8.DecodeLastRuneInString(a)
		if categorizer(trigger) == categorizer(r) {
			return
		}
		// Verify the rune preceding the abbreviation, if any, creates a word
		// boundary.
		if len(c.Content) > len(a)+triggerLen {
			r1, _ := utf8.DecodeLastRuneInString(c.Content[:len(c.Content)-len(a)-triggerLen])
			r2, _ := utf8.DecodeRuneInString(a)
			if categorizer(r1) == categorizer(r2) {
				return
			}
		}
		abbr, full = a, f
	})
	if len(abbr) > 0 {
		*c = CodeBuffer{
			Content: c.Content[:c.Dot-len(abbr)-triggerLen] + full + string(trigger),
			Dot:     c.Dot - len(abbr) + len(full),
		}
		w.resetInserts()
	}
}

func (w *codeArea) handleKeyEvent(key ui.Key) bool {
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

	if w.Bindings.Handle(w, term.KeyEvent(key)) {
		return true
	}

	// We only implement essential keybindings here. Other keybindings can be
	// added via handler overlays.
	switch key {
	case ui.K('\n'):
		w.resetInserts()
		w.Submit()
		return true
	case ui.K(ui.Backspace), ui.K('H', ui.Ctrl):
		w.resetInserts()
		w.MutateState(func(s *CodeAreaState) {
			c := &s.Buffer
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
		w.StateMutex.Lock()
		defer w.StateMutex.Unlock()
		if w.lastCodeBuffer != w.State.Buffer {
			// Something has happened between the last insert and this one;
			// reset the state.
			w.resetInserts()
		}
		s := string(key.Rune)
		w.State.Buffer.InsertAtDot(s)
		w.inserts += s
		w.lastCodeBuffer = w.State.Buffer
		w.expandSimpleAbbr()
		w.expandWordAbbr(key.Rune, CategorizeSmallWord)
		return true
	}
}

// IsAlnum determines if the rune is an alphanumeric character.
func IsAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// CategorizeSmallWord determines if the rune is whitespace, alphanum, or
// something else.
func CategorizeSmallWord(r rune) int {
	switch {
	case unicode.IsSpace(r):
		return 0
	case IsAlnum(r):
		return 1
	default:
		return 2
	}
}
