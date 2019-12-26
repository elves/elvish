// Package codearea implements a widget for showing and editing code in CLI.
package codearea

import (
	"bytes"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/ui"
)

// CodeArea is a Widget for displaying and editing code.
type CodeArea interface {
	el.Widget
	// CopyState returns a copy of the state.
	CopyState() CodeAreaState
	// MutateState calls the given the function while locking StateMutex.
	MutateState(f func(*CodeAreaState))
	// Submit triggers the OnSubmit callback.
	Submit()
}

// CodeAreaSpec specifies the configuration and initial state for CodeArea.
type CodeAreaSpec struct {
	// A Handler that takes precedence over the default handling of events.
	OverlayHandler el.Handler
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
	Abbreviations func(f func(abbr, full string))
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
	if spec.OverlayHandler == nil {
		spec.OverlayHandler = el.DummyHandler{}
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
	if w.OverlayHandler.Handle(event) {
		return true
	}

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
	// We only implement essential keybindings here. Other keybindings can be
	// added via handler overlays.
	switch key {
	case ui.K('\n'):
		w.resetInserts()
		w.Submit()
		return true
	case ui.K(ui.Backspace):
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
		var abbr, full string
		// Try to expand an abbreviation, preferring the longest one
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
		return true
	}
}
