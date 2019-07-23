package codearea

import (
	"sync"

	"github.com/elves/elvish/styled"
)

// State keeps the state of the widget. Its access must be synchronized through
// the mutex.
type State struct {
	Mutex       sync.RWMutex
	CodeBuffer  CodeBuffer
	PendingCode PendingCode
	Prompt      styled.Text
	RPrompt     styled.Text
}

// CodeBuffer represents the state of the buffer.
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
	Begin int
	// End index of the text area that the pending code replaces, as a byte
	// index into RawState.Code.
	End int
	// The content of the pending code.
	Text string
}

func (s *State) MutateCodeBuffer(f func(*CodeBuffer)) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	f(&s.CodeBuffer)
}

func (c *CodeBuffer) InsertAtDot(text string) {
	*c = CodeBuffer{
		Content: c.Content[:c.Dot] + text + c.Content[c.Dot:],
		Dot:     c.Dot + len(text),
	}
}

/*
func (w *Widget) CodeBuffer() CodeBuffer {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state.code
}

func (w *Widget) MutateCodeBuffer(f func(*CodeBuffer)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	f(&w.state.code)
}

func (w *Widget) PendingCode() PendingCode {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state.pending
}

func (w *Widget) MutatePendingCode(f func(*PendingCode)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	f(&w.state.pending)
}

func (w *Widget) Prompt() styled.Text {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state.prompt
}

func (w *Widget) SetPrompt() styled.Text {
}

func (w *Widget) RPrompt() styled.Text {
}

func (w *Widget) SetRPrompt() styled.Text {
}
*/
