package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/el/codearea"
)

func TestDot(t *testing.T) {
	ed, _, ev, cleanup := setupStarted()
	defer cleanup()

	ed.app.CodeArea().MutateState(func(s *codearea.State) {
		s.CodeBuffer = codearea.CodeBuffer{Content: "code", Dot: 4}
	})
	evalf(ev, `edit:-dot = 0`)
	wantBuf := codearea.CodeBuffer{Content: "code", Dot: 0}
	if buf := ed.app.CodeArea().CopyState().CodeBuffer; buf != wantBuf {
		t.Errorf("content = %v, want %v", buf, wantBuf)
	}
}

func TestCurrentCommand(t *testing.T) {
	ed, _, ev, cleanup := setupStarted()
	defer cleanup()

	evalf(ev, `edit:current-command = code`)
	wantBuf := codearea.CodeBuffer{Content: "code", Dot: 4}
	if buf := ed.app.CodeArea().CopyState().CodeBuffer; buf != wantBuf {
		t.Errorf("content = %v, want %v", buf, wantBuf)
	}
}
