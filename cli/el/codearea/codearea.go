// Package codearea implements a widget for showing and editing code in CLI.
package codearea

import (
	"bytes"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/styled"
)

// Widget supports code-editing functions. It implements the clitypes.Widget
// interface. An empty Widget is directly usable.
type Widget interface {
	el.Widget
	// CopyState returns a copy of the state.
	CopyState() State
	// MutateState calls the given the function while locking StateMutex.
	MutateState(f func(*State))
}

// Spec specifies the configuration and initial state for Widget.
type Spec struct {
	// A Handler that takes precedence over the default handling of events.
	OverlayHandler el.Handler
	// A function that highlights the given code and returns any errors it has
	// found when highlighting. If this function is not given, the Widget does
	// not highlight the code nor show any errors.
	Highlighter func(code string) (styled.Text, []error)
	// Prompt callback.
	Prompt func() styled.Text
	// Right-prompt callback.
	RPrompt func() styled.Text
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
	State State
}

type widget struct {
	// Mutex for synchronizing access to State.
	StateMutex sync.RWMutex
	// Configuration and state.
	Spec

	// Consecutively inserted text. Used for expanding abbreviations.
	inserts string
	// Value of State.CodeBuffer when handleKeyEvent was last called. Used for
	// detecting whether insertion has been interrupted.
	lastCodeBuffer Buffer
	// Whether the widget is in the middle of bracketed pasting.
	pasting bool
	// Buffer for keeping Pasted text during bracketed pasting.
	pasteBuffer bytes.Buffer
}

// New creates a new codearea widget from the given Spec.
func New(spec Spec) Widget {
	if spec.OverlayHandler == nil {
		spec.OverlayHandler = el.DummyHandler{}
	}
	if spec.Highlighter == nil {
		spec.Highlighter = dummyHighlighter
	}
	if spec.Prompt == nil {
		spec.Prompt = dummyPrompt
	}
	if spec.RPrompt == nil {
		spec.RPrompt = dummyPrompt
	}
	if spec.Abbreviations == nil {
		spec.Abbreviations = dummyAbbreviations
	}
	if spec.QuotePaste == nil {
		spec.QuotePaste = dummyQuotePaste
	}
	if spec.OnSubmit == nil {
		spec.OnSubmit = dummyOnSubmit
	}
	return &widget{Spec: spec}
}

// ConstPrompt returns a prompt callback that always writes the same styled
// text.
func ConstPrompt(content styled.Text) func() styled.Text {
	return func() styled.Text { return content }
}

func dummyHighlighter(code string) (styled.Text, []error) {
	return styled.Plain(code), nil
}

func dummyPrompt() styled.Text { return nil }

func dummyAbbreviations(func(a, f string)) {}

func dummyQuotePaste() bool { return false }

func dummyOnSubmit() {}

// Submit emits a submit event with the current code content.
func (w *widget) Submit() {
	w.OnSubmit()
}

// Render renders the code area, including the prompt and rprompt, highlighted
// code, the cursor, and compilation errors in the code content.
func (w *widget) Render(width, height int) *ui.Buffer {
	view := getView(w)
	bb := ui.NewBufferBuilder(width)
	renderView(view, bb)
	b := bb.Buffer()
	truncateToHeight(b, height)
	return b
}

// Handle handles KeyEvent's of non-function keys, as well as PasteSetting
// events.
func (w *widget) Handle(event term.Event) bool {
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

func (w *widget) MutateState(f func(*State)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}

func (w *widget) CopyState() State {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}

func (w *widget) resetInserts() {
	w.inserts = ""
	w.lastCodeBuffer = Buffer{}
}

func (w *widget) handlePasteSetting(start bool) bool {
	w.resetInserts()
	if start {
		w.pasting = true
	} else {
		text := w.pasteBuffer.String()
		if w.QuotePaste() {
			text = parse.Quote(text)
		}
		w.MutateState(func(s *State) { s.Buffer.InsertAtDot(text) })

		w.pasting = false
		w.pasteBuffer = bytes.Buffer{}
	}
	return true
}

func (w *widget) handleKeyEvent(key ui.Key) bool {
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
		w.MutateState(func(s *State) {
			c := &s.Buffer
			// Remove the last rune.
			_, chop := utf8.DecodeLastRuneInString(c.Content[:c.Dot])
			*c = Buffer{
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
			*c = Buffer{
				Content: c.Content[:c.Dot-len(abbr)] + full + c.Content[c.Dot:],
				Dot:     c.Dot - len(abbr) + len(full),
			}
			w.resetInserts()
		}
		return true
	}
}
