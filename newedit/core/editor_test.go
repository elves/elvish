package core

import (
	"reflect"
	"strings"
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
	ed := NewEditor(newFakeTTY(eventsABCEnter), nil)
	m := &fakeMode{maxKeys: len(eventsABCEnter)}
	ed.state.Mode = m

	ed.ReadCode()

	if !reflect.DeepEqual(m.keysHandled, keysABCEnter) {
		t.Errorf("Mode gets keys %v, want %v", m.keysHandled, keysABCEnter)
	}
}

func TestRead_CallsBeforeReadlineOnce(t *testing.T) {
	ed := NewEditor(newFakeTTY(eventsABCEnter), nil)

	called := 0
	ed.config.BeforeReadline = []func(){func() { called++ }}

	ed.ReadCode()

	if called != 1 {
		t.Errorf("BeforeReadline hook called %d times, want 1", called)
	}
}

func TestRead_CallsAfterReadlineOnceWithCode(t *testing.T) {
	ed := NewEditor(newFakeTTY(eventsABCEnter), nil)

	called := 0
	code := ""
	ed.config.AfterReadline = []func(string){func(s string) {
		called++
		code = s
	}}

	ed.ReadCode()

	if called != 1 {
		t.Errorf("AfterReadline hook called %d times, want 1", called)
	}
	if code != "abc" {
		t.Errorf("AfterReadline hook called with %q, want %q", code, "abc")
	}
}

func TestRead_RespectsMaxHeight(t *testing.T) {
	maxHeight := 5

	terminal := newFakeTTY(nil)
	ed := NewEditor(terminal, nil)
	// Will fill more than maxHeight but less than terminal height
	ed.state.Code = strings.Repeat("a", 80*10)
	ed.state.Dot = len(ed.state.Code)

	go ed.ReadCode()

	buf1 := <-terminal.bufCh
	// Make sure that normally the height does exceed maxHeight.
	if h := len(buf1.Lines); h <= maxHeight {
		t.Errorf("Buffer height is %d, should > %d", h, maxHeight)
	}

	ed.config.RenderConfig.MaxHeight = maxHeight
	ed.loop.Redraw(false)
	buf2 := <-terminal.bufCh
	if h := len(buf2.Lines); h > maxHeight {
		t.Errorf("Buffer height is %d, should <= %d", h, maxHeight)
	}

	terminal.eventCh <- tty.KeyEvent(kEnter)
}

var bufChTimeout = 1 * time.Second

func TestRead_RendersHighlightedCode(t *testing.T) {
	terminal := newFakeTTY(eventsABC)
	ed := NewEditor(terminal, nil)
	ed.config.RenderConfig.Highlighter = func(code string) (styled.Text, []error) {
		return styled.Text{
			styled.Segment{styled.Style{Foreground: "red"}, code}}, nil
	}

	go ed.ReadCode()

	wantBuf := ui.NewBufferBuilder(80).
		WriteString("abc", "31" /* SGR for red foreground */).
		Buffer()
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
