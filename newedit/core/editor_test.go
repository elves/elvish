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

func TestReadCode_PassesInputEventsToMode(t *testing.T) {
	ed := NewEditor(newFakeTTY(eventsABCEnter), nil)
	m := &fakeMode{maxKeys: len(eventsABCEnter)}
	ed.state.Mode = m

	ed.ReadCode()

	if !reflect.DeepEqual(m.keysHandled, keysABCEnter) {
		t.Errorf("Mode gets keys %v, want %v", m.keysHandled, keysABCEnter)
	}
}

func TestReadCode_CallsBeforeReadlineOnce(t *testing.T) {
	ed := NewEditor(newFakeTTY(eventsABCEnter), nil)

	called := 0
	ed.config.BeforeReadline = []func(){func() { called++ }}

	ed.ReadCode()

	if called != 1 {
		t.Errorf("BeforeReadline hook called %d times, want 1", called)
	}
}

func TestReadCode_CallsAfterReadlineOnceWithCode(t *testing.T) {
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

func TestReadCode_RespectsMaxHeight(t *testing.T) {
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

func TestReadCode_RendersHighlightedCode(t *testing.T) {
	terminal := newFakeTTY(eventsABC)
	ed := NewEditor(terminal, nil)
	ed.config.RenderConfig.Highlighter = func(code string) (styled.Text, []error) {
		return styled.Text{
			styled.Segment{styled.Style{Foreground: "red"}, code}}, nil
	}

	go ed.ReadCode()

	wantBuf := ui.NewBufferBuilder(80).
		WriteString("abc", "31" /* SGR for red foreground */).
		SetDotToCursor().Buffer()
	if !checkBuffer(terminal.bufCh, wantBuf) {
		t.Errorf("Did not see buffer containing highlighted code")
	}
	terminal.eventCh <- tty.KeyEvent(kEnter)
}

func TestReadCode_RendersErrorFromHighlighter(t *testing.T) {
	// TODO
}

func TestReadCode_RendersPrompt(t *testing.T) {
	// TODO
}

func TestReadCode_RendersRprompt(t *testing.T) {
	// TODO
}

func TestReadCode_SupportsPersistentRprompt(t *testing.T) {
	// TODO
}

var checkBufferTimeout = time.Second

func checkBuffer(ch <-chan *ui.Buffer, want *ui.Buffer) bool {
	for {
		select {
		case buf := <-ch:
			if reflect.DeepEqual(buf, want) {
				return true
			}
		case <-time.After(checkBufferTimeout):
			return false
		}
	}
}
