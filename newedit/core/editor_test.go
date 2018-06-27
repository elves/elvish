package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var (
	ka     = ui.Key{Rune: 'a'}
	kb     = ui.Key{Rune: 'b'}
	kc     = ui.Key{Rune: 'c'}
	kEnter = ui.Key{Rune: ui.Enter}

	keysABCEnter = []ui.Key{ka, kb, kc, kEnter}
	eventsABC    = []tty.Event{
		tty.KeyEvent(ka), tty.KeyEvent(kb), tty.KeyEvent(kc)}
	eventsABCEnter = []tty.Event{
		tty.KeyEvent(ka), tty.KeyEvent(kb),
		tty.KeyEvent(kc), tty.KeyEvent(kEnter)}
)

func TestRead_PassesInputEventsToMode(t *testing.T) {
	ed := NewEditor(newFakeTTY(eventsABCEnter))
	m := newFakeMode(len(eventsABCEnter))
	ed.state.Mode = m

	ed.Read()

	if !reflect.DeepEqual(m.keys, keysABCEnter) {
		t.Errorf("Mode gets keys %v, want %v", m.keys, keysABCEnter)
	}
}

func TestRead_CallsBeforeReadlineOnce(t *testing.T) {
	ed := NewEditor(newFakeTTY(eventsABCEnter))

	called := 0
	ed.config.BeforeReadline = []func(){func() { called++ }}

	ed.Read()

	if called != 1 {
		t.Errorf("BeforeReadline hook called %d times, want 1", called)
	}
}

func TestRead_CallsAfterReadlineOnceWithCode(t *testing.T) {
	ed := NewEditor(newFakeTTY(eventsABCEnter))

	called := 0
	code := ""
	ed.config.AfterReadline = []func(string){func(s string) {
		called++
		code = s
	}}

	ed.Read()

	if called != 1 {
		t.Errorf("AfterReadline hook called %d times, want 1", called)
	}
	if code != "abc" {
		t.Errorf("AfterReadline hook called with %q, want %q", code, "abc")
	}
}

func TestRead_RespectsMaxHeight(t *testing.T) {
	// TODO
}

var bufChTimeout = 1 * time.Second

func TestRead_RendersHighlightedCode(t *testing.T) {
	terminal := newFakeTTY(eventsABC)
	ed := NewEditor(terminal)
	ed.config.Render.Highlighter = func(code string) (styled.Text, []error) {
		return styled.Text{
			styled.Segment{styled.Style{Foreground: "red"}, code}}, nil
	}

	go ed.Read()

	wantBuf := ui.NewBuffer(80)
	wantBuf.WriteString("abc", "31" /* SGR for red foreground */)
checkBuffer:
	for {
		select {
		case buf := <-terminal.bufCh:
			// Check if the buffer matches out expectation.
			if reflect.DeepEqual(buf.Lines, wantBuf.Lines) {
				break checkBuffer
			}
		case <-time.After(time.Second):
			t.Errorf("Timeout waiting for matching buffer")
			break checkBuffer
		}
	}
	terminal.eventCh <- tty.KeyEvent(kEnter)
}

func TestRead_RendersErrorFromHighlighter(t *testing.T) {
	// TODO
}

func TestRead_RendersPrompt(t *testing.T) {
	// TODO
}

func TestRead_RendersRprompt(t *testing.T) {
	// TODO
}

func TestRead_SupportsPersistentRprompt(t *testing.T) {
	// TODO
}
