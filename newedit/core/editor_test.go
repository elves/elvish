package core

import (
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/sys"
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
	terminal := newFakeTTY()
	ed := NewEditor(terminal, nil)
	m := &fakeMode{maxKeys: 3}
	ed.state.Mode = m

	terminal.eventCh <- tty.KeyEvent{Rune: 'a'}
	terminal.eventCh <- tty.KeyEvent{Rune: 'b'}
	terminal.eventCh <- tty.KeyEvent{Rune: 'c'}

	ed.ReadCode()

	wantKeysHandled := []ui.Key{
		ui.Key{Rune: 'a'}, ui.Key{Rune: 'b'}, ui.Key{Rune: 'c'},
	}
	if !reflect.DeepEqual(m.keysHandled, wantKeysHandled) {
		t.Errorf("Mode gets keys %v, want %v", m.keysHandled, wantKeysHandled)
	}
}

func TestReadCode_CallsBeforeReadlineOnce(t *testing.T) {
	terminal := newFakeTTY()
	ed := NewEditor(terminal, nil)

	called := 0
	ed.config.BeforeReadline = []func(){func() { called++ }}

	// Causes basicMode to quit
	terminal.eventCh <- tty.KeyEvent{Rune: '\n'}

	ed.ReadCode()

	if called != 1 {
		t.Errorf("BeforeReadline hook called %d times, want 1", called)
	}
}

func TestReadCode_CallsAfterReadlineOnceWithCode(t *testing.T) {
	terminal := newFakeTTY()
	ed := NewEditor(terminal, nil)

	called := 0
	code := ""
	ed.config.AfterReadline = []func(string){func(s string) {
		called++
		code = s
	}}

	// Causes basicMode to write state.Code and then quit
	terminal.eventCh <- tty.KeyEvent{Rune: 'a'}
	terminal.eventCh <- tty.KeyEvent{Rune: 'b'}
	terminal.eventCh <- tty.KeyEvent{Rune: 'c'}
	terminal.eventCh <- tty.KeyEvent{Rune: '\n'}

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

	terminal := newFakeTTY()
	ed := NewEditor(terminal, nil)
	// Will fill more than maxHeight but less than terminal height
	ed.state.Code = strings.Repeat("a", 80*10)
	ed.state.Dot = len(ed.state.Code)

	codeCh, _ := readCodeAsync(ed)

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

	terminal.eventCh <- tty.KeyEvent{Rune: '\n'}
	<-codeCh
}

var bufChTimeout = 1 * time.Second

func TestReadCode_RendersHighlightedCode(t *testing.T) {
	terminal := newFakeTTY()
	ed := NewEditor(terminal, nil)
	ed.config.RenderConfig.Highlighter = func(code string) (styled.Text, []error) {
		return styled.Text{
			styled.Segment{styled.Style{Foreground: "red"}, code}}, nil
	}

	terminal.eventCh <- tty.KeyEvent{Rune: 'a'}
	terminal.eventCh <- tty.KeyEvent{Rune: 'b'}
	terminal.eventCh <- tty.KeyEvent{Rune: 'c'}
	codeCh, _ := readCodeAsync(ed)

	wantBuf := ui.NewBufferBuilder(80).
		WriteString("abc", "31" /* SGR for red foreground */).
		SetDotToCursor().Buffer()
	if !checkBuffer(terminal.bufCh, wantBuf) {
		t.Errorf("Did not see buffer containing highlighted code")
	}

	terminal.eventCh <- tty.KeyEvent{Rune: '\n'}
	<-codeCh
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

func TestReadCode_QuitsOnSIGHUP(t *testing.T) {
	terminal := newFakeTTY()
	sigs := newFakeSignalSource()
	ed := NewEditor(terminal, sigs)

	codeCh, _ := readCodeAsync(ed)
	terminal.eventCh <- tty.KeyEvent{Rune: 'a'}
	wantBuf := ui.NewBufferBuilder(80).WriteUnstyled("a").
		SetDotToCursor().Buffer()
	if !checkBuffer(terminal.bufCh, wantBuf) {
		t.Errorf("did not get expected buffer before sending SIGHUP")
	}

	sigs.ch <- syscall.SIGHUP

	select {
	case <-codeCh:
		// TODO: Test that ReadCode returns with io.EOF
	case <-time.After(time.Second):
		t.Errorf("SIGHUP did not cause ReadCode to return")
	}
}

func TestReadCode_ResetsOnSIGHUP(t *testing.T) {
	terminal := newFakeTTY()
	sigs := newFakeSignalSource()
	ed := NewEditor(terminal, sigs)

	codeCh, _ := readCodeAsync(ed)
	terminal.eventCh <- tty.KeyEvent{Rune: 'a'}
	wantBuf := ui.NewBufferBuilder(80).WriteUnstyled("a").
		SetDotToCursor().Buffer()
	if !checkBuffer(terminal.bufCh, wantBuf) {
		t.Errorf("did not get expected buffer before sending SIGINT")
	}

	sigs.ch <- syscall.SIGINT

	wantBuf = ui.NewBufferBuilder(80).Buffer()
	if !checkBuffer(terminal.bufCh, wantBuf) {
		t.Errorf("Terminal state is not reset after SIGINT")
	}

	terminal.eventCh <- tty.KeyEvent{Rune: '\n'}
	<-codeCh
}

func TestReadCode_RedrawsOnSIGWINCH(t *testing.T) {
	terminal := newFakeTTY()
	sigs := newFakeSignalSource()
	ed := NewEditor(terminal, sigs)

	ed.state.Code = "1234567890"
	ed.state.Dot = len(ed.state.Code)

	codeCh, _ := readCodeAsync(ed)
	wantBuf := ui.NewBufferBuilder(80).WriteUnstyled("1234567890").
		SetDotToCursor().Buffer()
	if !checkBuffer(terminal.bufCh, wantBuf) {
		t.Errorf("did not get expected buffer before sending SIGWINCH")
	}

	terminal.w = 4
	sigs.ch <- sys.SIGWINCH

	wantBuf = ui.NewBufferBuilder(4).WriteUnstyled("1234567890").
		SetDotToCursor().Buffer()
	if !checkBuffer(terminal.bufCh, wantBuf) {
		t.Errorf("Terminal is not redrawn after SIGWINCH")
	}

	terminal.eventCh <- tty.KeyEvent{Rune: '\n'}
	<-codeCh
}

func readCodeAsync(ed *Editor) (<-chan string, <-chan error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := ed.ReadCode()
		codeCh <- code
		errCh <- err
	}()
	return codeCh, errCh
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
