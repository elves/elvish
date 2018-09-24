package core

import (
	"errors"
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

func TestReadCode_AbortsOnSetupError(t *testing.T) {
	ed, terminal, _ := setup()

	terminal.SetupErr = errors.New("a fake error")

	_, err := ed.ReadCode()

	if err != terminal.SetupErr {
		t.Errorf("ReadCode returns error %v, want %v", err, terminal.SetupErr)
	}
}

func TestReadCode_CallsRestore(t *testing.T) {
	ed, terminal, _ := setup()

	restoreCalled := 0
	terminal.RestoreFunc = func() { restoreCalled++ }
	terminal.EventCh <- tty.KeyEvent{Rune: '\n'}

	ed.ReadCode()

	if restoreCalled != 1 {
		t.Errorf("Restore callback called %d times, want once", restoreCalled)
	}
}

func TestReadCode_ResetsStateBeforeReturn(t *testing.T) {
	ed, terminal, _ := setup()

	terminal.EventCh <- tty.KeyEvent{Rune: '\n'}
	ed.State.Raw.Code = "some code"

	ed.ReadCode()

	if code := ed.State.Raw.Code; code != "" {
		t.Errorf("Editor state has code %q, want empty", code)
	}
}

func TestReadCode_PassesInputEventsToMode(t *testing.T) {
	ed, terminal, _ := setup()

	m := &fakeMode{maxKeys: 3}
	ed.State.Raw.Mode = m
	terminal.EventCh <- tty.KeyEvent{Rune: 'a'}
	terminal.EventCh <- tty.KeyEvent{Rune: 'b'}
	terminal.EventCh <- tty.KeyEvent{Rune: 'c'}

	ed.ReadCode()

	wantKeysHandled := []ui.Key{
		ui.Key{Rune: 'a'}, ui.Key{Rune: 'b'}, ui.Key{Rune: 'c'},
	}
	if !reflect.DeepEqual(m.keysHandled, wantKeysHandled) {
		t.Errorf("Mode gets keys %v, want %v", m.keysHandled, wantKeysHandled)
	}
}

func TestReadCode_CallsBeforeReadlineOnce(t *testing.T) {
	ed, terminal, _ := setup()

	called := 0
	ed.BeforeReadline = func() { called++ }
	// Causes BasicMode to quit
	terminal.EventCh <- tty.KeyEvent{Rune: '\n'}

	ed.ReadCode()

	if called != 1 {
		t.Errorf("BeforeReadline hook called %d times, want 1", called)
	}
}

func TestReadCode_CallsAfterReadlineOnceWithCode(t *testing.T) {
	ed, terminal, _ := setup()

	called := 0
	code := ""
	ed.AfterReadline = func(s string) {
		called++
		code = s
	}
	// Causes BasicMode to write state.Code and then quit
	terminal.EventCh <- tty.KeyEvent{Rune: 'a'}
	terminal.EventCh <- tty.KeyEvent{Rune: 'b'}
	terminal.EventCh <- tty.KeyEvent{Rune: 'c'}
	terminal.EventCh <- tty.KeyEvent{Rune: '\n'}

	ed.ReadCode()

	if called != 1 {
		t.Errorf("AfterReadline hook called %d times, want 1", called)
	}
	if code != "abc" {
		t.Errorf("AfterReadline hook called with %q, want %q", code, "abc")
	}
}

func TestReadCode_RespectsMaxHeight(t *testing.T) {
	ed, terminal, _ := setup()

	maxHeight := 5
	// Will fill more than maxHeight but less than terminal height
	ed.State.Raw.Code = strings.Repeat("a", 80*10)
	ed.State.Raw.Dot = len(ed.State.Raw.Code)

	codeCh, _ := ed.readCodeAsync()

	buf1 := <-terminal.BufCh
	// Make sure that normally the height does exceed maxHeight.
	if h := len(buf1.Lines); h <= maxHeight {
		t.Errorf("Buffer height is %d, should > %d", h, maxHeight)
	}

	ed.Config.Mutex.Lock()
	ed.Config.Raw.MaxHeight = maxHeight
	ed.Config.Mutex.Unlock()

	ed.Redraw(false)

	buf2 := <-terminal.BufCh
	if h := len(buf2.Lines); h > maxHeight {
		t.Errorf("Buffer height is %d, should <= %d", h, maxHeight)
	}

	cleanup(terminal, codeCh)
}

var bufChTimeout = 1 * time.Second

func TestReadCode_RendersHighlightedCode(t *testing.T) {
	ed, terminal, _ := setup()

	ed.Highlighter = func(code string) (styled.Text, []error) {
		return styled.Text{
			&styled.Segment{styled.Style{Foreground: "red"}, code}}, nil
	}
	terminal.EventCh <- tty.KeyEvent{Rune: 'a'}
	terminal.EventCh <- tty.KeyEvent{Rune: 'b'}
	terminal.EventCh <- tty.KeyEvent{Rune: 'c'}

	codeCh, _ := ed.readCodeAsync()

	wantBuf := ui.NewBufferBuilder(80).
		WriteString("abc", "31" /* SGR for red foreground */).
		SetDotToCursor().Buffer()
	if !checkBuffer(wantBuf, terminal.BufCh) {
		t.Errorf("Did not see buffer containing highlighted code")
	}

	cleanup(terminal, codeCh)
}

func TestReadCode_RendersErrorFromHighlighter(t *testing.T) {
	// TODO
}

func TestReadCode_RendersPrompt(t *testing.T) {
	ed, terminal, _ := setup()

	ed.Prompt = constPrompt{styled.Unstyled("> ")}
	terminal.EventCh <- tty.KeyEvent{Rune: 'a'}

	codeCh, _ := ed.readCodeAsync()

	wantBuf := ui.NewBufferBuilder(80).
		WriteUnstyled("> a").
		SetDotToCursor().Buffer()
	if !checkBuffer(wantBuf, terminal.BufCh) {
		t.Errorf("Did not see buffer containing prompt")
	}

	cleanup(terminal, codeCh)
}

func TestReadCode_RendersRPrompt(t *testing.T) {
	ed, terminal, _ := setup()

	terminal.width = 4
	ed.RPrompt = constPrompt{styled.Unstyled("R")}
	terminal.EventCh <- tty.KeyEvent{Rune: 'a'}

	codeCh, _ := ed.readCodeAsync()

	wantBuf := ui.NewBufferBuilder(4).
		WriteUnstyled("a").SetDotToCursor().WriteUnstyled("  R").Buffer()
	if !checkBuffer(wantBuf, terminal.BufCh) {
		t.Errorf("Did not see buffer containing rprompt")
	}

	cleanup(terminal, codeCh)
}

func TestReadCode_SupportsPersistentRPrompt(t *testing.T) {
	// TODO
}

func TestReadCode_DrawsAndFlushesNotes(t *testing.T) {
	ed, terminal, _ := setup()

	codeCh, _ := ed.readCodeAsync()

	// Sanity-check initial state.
	initBuf := ui.NewBufferBuilder(80).Buffer()
	if !checkBuffer(initBuf, terminal.BufCh) {
		t.Errorf("did not get initial state")
	}

	ed.Notify("note")

	wantNotesBuf := ui.NewBufferBuilder(80).WriteUnstyled("note").Buffer()
	if !checkBuffer(wantNotesBuf, terminal.NotesBufCh) {
		t.Errorf("did not render notes")
	}

	if n := len(ed.State.Raw.Notes); n > 0 {
		t.Errorf("State.Raw.Notes has %d elements after redrawing, want 0", n)
	}

	cleanup(terminal, codeCh)
}

func TestReadCode_UsesFinalStateInFinalRedraw(t *testing.T) {
	ed, terminal, _ := setup()

	ed.State.Raw.Code = "some code"
	// We use the dot as a signal for distinguishing non-final and final state.
	// In the final state, the dot will be set to the length of the code (9).
	ed.State.Raw.Dot = 1

	codeCh, _ := ed.readCodeAsync()

	// Wait until a non-final state is drawn.
	wantBuf := ui.NewBufferBuilder(80).WriteUnstyled("s").SetDotToCursor().
		WriteUnstyled("ome code").Buffer()
	if !checkBuffer(wantBuf, terminal.BufCh) {
		t.Errorf("did not get expected buffer before sending Enter")
	}

	cleanup(terminal, codeCh)

	// Last element in bufs is nil
	finalBuf := terminal.Bufs[len(terminal.Bufs)-2]
	wantFinalBuf := ui.NewBufferBuilder(80).WriteUnstyled("some code").
		SetDotToCursor().Buffer()
	if !reflect.DeepEqual(finalBuf, wantFinalBuf) {
		t.Errorf("final buffer is %v, want %v", finalBuf, wantFinalBuf)
	}
}

func TestReadCode_QuitsOnSIGHUP(t *testing.T) {
	ed, terminal, sigs := setup()

	terminal.EventCh <- tty.KeyEvent{Rune: 'a'}

	codeCh, _ := ed.readCodeAsync()

	wantBuf := ui.NewBufferBuilder(80).WriteUnstyled("a").
		SetDotToCursor().Buffer()
	if !checkBuffer(wantBuf, terminal.BufCh) {
		t.Errorf("did not get expected buffer before sending SIGHUP")
	}

	sigs.Ch <- syscall.SIGHUP

	select {
	case <-codeCh:
		// TODO: Test that ReadCode returns with io.EOF
	case <-time.After(time.Second):
		t.Errorf("SIGHUP did not cause ReadCode to return")
	}
}

func TestReadCode_ResetsOnSIGHUP(t *testing.T) {
	ed, terminal, sigs := setup()

	terminal.EventCh <- tty.KeyEvent{Rune: 'a'}

	codeCh, _ := ed.readCodeAsync()
	wantBuf := ui.NewBufferBuilder(80).WriteUnstyled("a").
		SetDotToCursor().Buffer()
	if !checkBuffer(wantBuf, terminal.BufCh) {
		t.Errorf("did not get expected buffer before sending SIGINT")
	}

	sigs.Ch <- syscall.SIGINT

	wantBuf = ui.NewBufferBuilder(80).Buffer()
	if !checkBuffer(wantBuf, terminal.BufCh) {
		t.Errorf("Terminal state is not reset after SIGINT")
	}

	cleanup(terminal, codeCh)
}

func TestReadCode_RedrawsOnSIGWINCH(t *testing.T) {
	ed, terminal, sigs := setup()

	ed.State.Raw.Code = "1234567890"
	ed.State.Raw.Dot = len(ed.State.Raw.Code)

	codeCh, _ := ed.readCodeAsync()

	wantBuf := ui.NewBufferBuilder(80).WriteUnstyled("1234567890").
		SetDotToCursor().Buffer()
	if !checkBuffer(wantBuf, terminal.BufCh) {
		t.Errorf("did not get expected buffer before sending SIGWINCH")
	}

	terminal.SetSize(24, 4)
	sigs.Ch <- sys.SIGWINCH

	wantBuf = ui.NewBufferBuilder(4).WriteUnstyled("1234567890").
		SetDotToCursor().Buffer()
	if !checkBuffer(wantBuf, terminal.BufCh) {
		t.Errorf("Terminal is not redrawn after SIGWINCH")
	}

	cleanup(terminal, codeCh)
}

func setup() (*Editor, *FakeTTY, *FakeSignalSource) {
	terminal := NewFakeTTY()
	sigsrc := NewFakeSignalSource()
	ed := NewEditor(terminal, sigsrc)
	return ed, terminal, sigsrc
}

func cleanup(t *FakeTTY, codeCh <-chan string) {
	// Causes BasicMode to quit
	t.EventCh <- tty.KeyEvent{Rune: '\n'}
	// Wait until ReadCode has finished execution
	<-codeCh
}
