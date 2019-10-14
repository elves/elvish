package cli

import (
	"errors"
	"io"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/sys"
)

func TestReadCode_AbortsOnSetupError(t *testing.T) {
	ed, tty := setup()

	setupErr := errors.New("a fake error")
	tty.SetSetup(func() {}, setupErr)

	_, err := ed.ReadCode()

	if err != setupErr {
		t.Errorf("ReadCode returns error %v, want %v", err, setupErr)
	}
}

func TestReadCode_CallsRestore(t *testing.T) {
	ed, tty := setup()

	restoreCalled := 0
	tty.SetSetup(func() { restoreCalled++ }, nil)
	tty.Inject(term.KeyEvent{Rune: '\n'})

	ed.ReadCode()

	if restoreCalled != 1 {
		t.Errorf("Restore callback called %d times, want once", restoreCalled)
	}
}

func TestReadCode_ResetsStateBeforeReturn(t *testing.T) {
	ed, tty := setup()

	tty.Inject(term.KeyEvent{Rune: '\n'})
	ed.CodeArea.State.CodeBuffer.Content = "some code"

	ed.ReadCode()

	if code := ed.CodeArea.State.CodeBuffer.Content; code != "" {
		t.Errorf("Editor state has code %q, want empty", code)
	}
}

func TestReadCode_PassesInputEventsToPopup(t *testing.T) {
	/*
		ed, tty := setup()

		m := &fakeMode{maxKeys: 3}
		ed.state.Raw.Mode = m
		tty.Inject(term.KeyEvent{Rune: 'a'})
		tty.Inject(term.KeyEvent{Rune: 'b'})
		tty.Inject(term.KeyEvent{Rune: 'c'})

		ed.ReadCode()

		wantKeysHandled := []ui.Key{
			{Rune: 'a'}, {Rune: 'b'}, {Rune: 'c'},
		}
		if !reflect.DeepEqual(m.keysHandled, wantKeysHandled) {
			t.Errorf("Mode gets keys %v, want %v", m.keysHandled, wantKeysHandled)
		}
	*/
}

func TestReadCode_PassesInputEventsToCodeArea(t *testing.T) {
}

func TestReadCode_CallsBeforeReadlineOnce(t *testing.T) {
	ed, tty := setup()

	called := 0
	ed.Config.BeforeReadline = func() { called++ }
	// Causes BasicMode to quit
	tty.Inject(term.KeyEvent{Rune: '\n'})

	ed.ReadCode()

	if called != 1 {
		t.Errorf("BeforeReadline hook called %d times, want 1", called)
	}
}

func TestReadCode_CallsAfterReadlineOnceWithCode(t *testing.T) {
	ed, tty := setup()

	called := 0
	code := ""
	ed.Config.AfterReadline = func(s string) {
		called++
		code = s
	}
	// Causes BasicMode to write state.Code and then quit
	tty.Inject(term.KeyEvent{Rune: 'a'})
	tty.Inject(term.KeyEvent{Rune: 'b'})
	tty.Inject(term.KeyEvent{Rune: 'c'})
	tty.Inject(term.KeyEvent{Rune: '\n'})

	ed.ReadCode()

	if called != 1 {
		t.Errorf("AfterReadline hook called %d times, want 1", called)
	}
	if code != "abc" {
		t.Errorf("AfterReadline hook called with %q, want %q", code, "abc")
	}
}

func TestReadCode_RespectsMaxHeight(t *testing.T) {
	ed, tty := setup()

	ed.Config.MaxHeight = func() int { return 2 }
	tty.SetSize(10, 5) // Width = 5 to make it easy to test

	// The code needs 3 lines to completely show.
	ed.CodeArea.State.CodeBuffer.Content = strings.Repeat("a", 15)

	codeCh, _ := ed.ReadCodeAsync()

	wantBuf := ui.NewBufferBuilder(5).
		WritePlain(strings.Repeat("a", 10)).Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(ed, codeCh)
}

var bufChTimeout = 1 * time.Second

func TestReadCode_RendersHighlightedCode(t *testing.T) {
	ed, tty := setup()

	ed.Config.Highlighter = testHighlighter{
		get: func(code string) (styled.Text, []error) {
			return styled.Text{
				&styled.Segment{styled.Style{Foreground: "red"}, code}}, nil
		},
	}
	tty.Inject(term.KeyEvent{Rune: 'a'})
	tty.Inject(term.KeyEvent{Rune: 'b'})
	tty.Inject(term.KeyEvent{Rune: 'c'})

	codeCh, _ := ed.ReadCodeAsync()

	wantBuf := ui.NewBufferBuilder(80).
		WriteStringSGR("abc", "31" /* SGR for red foreground */).
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(ed, codeCh)
}

func TestReadCode_RendersErrorFromHighlighter(t *testing.T) {
	// TODO
}

func TestReadCode_RedrawsOnHighlighterLateUpdate(t *testing.T) {
	// TODO
}

func TestReadCode_RendersPrompt(t *testing.T) {
	ed, tty := setup()

	ed.Config.Prompt = constPrompt{styled.Plain("> ")}
	tty.Inject(term.KeyEvent{Rune: 'a'})

	codeCh, _ := ed.ReadCodeAsync()

	wantBuf := ui.NewBufferBuilder(80).
		WritePlain("> a").
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(ed, codeCh)
}

func TestReadCode_RendersRPrompt(t *testing.T) {
	ed, tty := setup()
	tty.SetSize(80, 4) // Set a width of 4 for easier testing.

	ed.Config.RPrompt = constPrompt{styled.Plain("R")}
	tty.Inject(term.KeyEvent{Rune: 'a'})

	codeCh, _ := ed.ReadCodeAsync()

	wantBuf := ui.NewBufferBuilder(4).
		WritePlain("a").SetDotToCursor().WritePlain("  R").Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(ed, codeCh)
}

func TestReadCode_TriggersPrompt(t *testing.T) {
	ed, _ := setup()

	called := 0
	ed.Config.Prompt = testPrompt{trigger: func(bool) { called++ }}

	codeCh, _ := ed.ReadCodeAsync()
	cleanup(ed, codeCh)

	if called != 1 {
		t.Errorf("Prompt.Trigger called %d times, want once", called)
	}
}

func TestReadCode_RedrawsOnPromptLateUpdate(t *testing.T) {
	ed, tty := setup()

	promptContent := "old"
	prompt := testPrompt{
		get:         func() styled.Text { return styled.Plain(promptContent) },
		lateUpdates: make(chan styled.Text),
	}
	ed.Config.Prompt = prompt

	codeCh, _ := ed.ReadCodeAsync()
	bufOldPrompt := ui.NewBufferBuilder(80).
		WritePlain("old").SetDotToCursor().Buffer()
	// Wait until old prompt is rendered
	tty.TestBuffer(t, bufOldPrompt)

	promptContent = "new"
	prompt.lateUpdates <- nil
	bufNewPrompt := ui.NewBufferBuilder(80).
		WritePlain("new").SetDotToCursor().Buffer()
	tty.TestBuffer(t, bufNewPrompt)

	cleanup(ed, codeCh)
}

func TestReadCode_SupportsPersistentRPrompt(t *testing.T) {
	// TODO
}

func TestReadCode_DrawsAndFlushesNotes(t *testing.T) {
	ed, tty := setup()

	codeCh, _ := ed.ReadCodeAsync()

	// Sanity-check initial state.
	initBuf := ui.NewBufferBuilder(80).Buffer()
	tty.TestBuffer(t, initBuf)

	ed.Notify("note")

	wantNotesBuf := ui.NewBufferBuilder(80).WritePlain("note").Buffer()
	tty.TestNotesBuffer(t, wantNotesBuf)

	if n := len(ed.State.Notes); n > 0 {
		t.Errorf("State.Notes has %d elements after redrawing, want 0", n)
	}

	cleanup(ed, codeCh)
}

func TestReadCode_PutCursorBelowCodeAreaInFinalRedraw(t *testing.T) {
	app, tty := setup()
	app.CodeArea.State.CodeBuffer.Content = "some code"
	app.State.Listing = layout.Label{Content: styled.Plain("listing")}

	codeCh, _ := app.ReadCodeAsync()

	// Wait until the initial draw, to ensure that we are indeed observing a
	// different state later.
	wantBuf := ui.NewBufferBuilder(80).
		WritePlain("some code").
		Newline().SetDotToCursor().WritePlain("listing").Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(app, codeCh)

	wantFinalBuf := ui.NewBufferBuilder(80).
		WritePlain("some code").Newline().SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantFinalBuf)
}

func TestReadCode_QuitsOnSIGHUP(t *testing.T) {
	ed, tty := setup()

	tty.Inject(term.KeyEvent{Rune: 'a'})

	codeCh, errCh := ed.ReadCodeAsync()

	wantBuf := ui.NewBufferBuilder(80).WritePlain("a").
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	tty.InjectSignal(syscall.SIGHUP)

	select {
	case <-codeCh:
		err := <-errCh
		if err != io.EOF {
			t.Errorf("want ReadCode to return io.EOF on SIGHUP, got %v", err)
		}
	case <-time.After(time.Second):
		t.Errorf("SIGHUP did not cause ReadCode to return")
	}
}

func TestReadCode_ResetsOnSIGINT(t *testing.T) {
	ed, tty := setup()

	tty.Inject(term.KeyEvent{Rune: 'a'})

	codeCh, _ := ed.ReadCodeAsync()
	wantBuf := ui.NewBufferBuilder(80).WritePlain("a").
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	tty.InjectSignal(syscall.SIGINT)

	wantBuf = ui.NewBufferBuilder(80).Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(ed, codeCh)
}

func TestReadCode_RedrawsOnSIGWINCH(t *testing.T) {
	ed, tty := setup()

	content := "1234567890"
	ed.CodeArea.State.CodeBuffer = codearea.CodeBuffer{
		Content: content, Dot: len(content),
	}

	codeCh, _ := ed.ReadCodeAsync()

	wantBuf := ui.NewBufferBuilder(80).WritePlain("1234567890").
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	tty.SetSize(24, 4)
	tty.InjectSignal(sys.SIGWINCH)

	wantBuf = ui.NewBufferBuilder(4).WritePlain("1234567890").
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(ed, codeCh)
}

func setup() (*App, TTYCtrl) {
	tty, ttyControl := NewFakeTTY()
	app := NewApp(tty)
	return app, ttyControl
}

func cleanup(app *App, codeCh <-chan string) {
	app.CommitEOF()
	// Make sure that ReadCode has exited
	<-codeCh
}
