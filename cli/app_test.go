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
	a, tty := setup()

	setupErr := errors.New("a fake error")
	tty.SetSetup(func() {}, setupErr)

	_, err := a.ReadCode()

	if err != setupErr {
		t.Errorf("ReadCode returns error %v, want %v", err, setupErr)
	}
}

func TestReadCode_CallsRestore(t *testing.T) {
	a, tty := setup()

	restoreCalled := 0
	tty.SetSetup(func() { restoreCalled++ }, nil)
	tty.Inject(term.KeyEvent{Rune: '\n'})

	a.ReadCode()

	if restoreCalled != 1 {
		t.Errorf("Restore callback called %d times, want once", restoreCalled)
	}
}

func TestReadCode_ResetsStateBeforeReturn(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		CodeAreaState: codearea.State{
			Buffer: codearea.Buffer{Content: "some code"}}})

	tty.Inject(term.KeyEvent{Rune: '\n'})

	a.ReadCode()

	if code := a.CodeArea().CopyState().Buffer.Content; code != "" {
		t.Errorf("Editor state has code %q, want empty", code)
	}
}

func TestReadCode_PassesInputEventsToPopup(t *testing.T) {
	/*
		a, tty := setup()

		m := &fakeMode{maxKeys: 3}
		a.state.Raw.Mode = m
		tty.Inject(term.KeyEvent{Rune: 'a'})
		tty.Inject(term.KeyEvent{Rune: 'b'})
		tty.Inject(term.KeyEvent{Rune: 'c'})

		a.ReadCode()

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
	called := 0
	a, tty := setupWithSpec(AppSpec{BeforeReadline: func() { called++ }})

	// Causes BasicMode to quit
	tty.Inject(term.KeyEvent{Rune: '\n'})

	a.ReadCode()

	if called != 1 {
		t.Errorf("BeforeReadline hook called %d times, want 1", called)
	}
}

func TestReadCode_CallsAfterReadlineOnceWithCode(t *testing.T) {
	called := 0
	code := ""
	a, tty := setupWithSpec(AppSpec{
		AfterReadline: func(s string) {
			called++
			code = s
		}})

	// Causes BasicMode to write state.Code and then quit
	tty.Inject(term.KeyEvent{Rune: 'a'})
	tty.Inject(term.KeyEvent{Rune: 'b'})
	tty.Inject(term.KeyEvent{Rune: 'c'})
	tty.Inject(term.KeyEvent{Rune: '\n'})

	a.ReadCode()

	if called != 1 {
		t.Errorf("AfterReadline hook called %d times, want 1", called)
	}
	if code != "abc" {
		t.Errorf("AfterReadline hook called with %q, want %q", code, "abc")
	}
}

func TestReadCode_RespectsMaxHeight(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		MaxHeight: func() int { return 2 },
		CodeAreaState: codearea.State{
			// The code needs 3 lines to completely show.
			Buffer: codearea.Buffer{Content: strings.Repeat("a", 15)}}})

	tty.SetSize(10, 5) // Width = 5 to make it easy to test

	codeCh, _ := ReadCodeAsync(a)

	wantBuf := ui.NewBufferBuilder(5).
		WritePlain(strings.Repeat("a", 10)).Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(a, codeCh)
}

var bufChTimeout = 1 * time.Second

func TestReadCode_RendersHighlightedCode(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		Highlighter: testHighlighter{
			get: func(code string) (styled.Text, []error) {
				return styled.MakeText(code, "red"), nil
			},
		}})

	tty.Inject(term.KeyEvent{Rune: 'a'})
	tty.Inject(term.KeyEvent{Rune: 'b'})
	tty.Inject(term.KeyEvent{Rune: 'c'})

	codeCh, _ := ReadCodeAsync(a)

	wantBuf := ui.NewBufferBuilder(80).
		WriteStringSGR("abc", "31" /* SGR for red foreground */).
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(a, codeCh)
}

func TestReadCode_RendersErrorFromHighlighter(t *testing.T) {
	// TODO
}

func TestReadCode_RedrawsOnHighlighterLateUpdate(t *testing.T) {
	// TODO
}

func TestReadCode_RendersPrompt(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		Prompt: constPrompt{styled.Plain("> ")}})

	tty.Inject(term.KeyEvent{Rune: 'a'})

	codeCh, _ := ReadCodeAsync(a)

	wantBuf := ui.NewBufferBuilder(80).
		WritePlain("> a").
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(a, codeCh)
}

func TestReadCode_RendersRPrompt(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		RPrompt: constPrompt{styled.Plain("R")}})
	tty.SetSize(80, 4) // Set a width of 4 for easier testing.

	tty.Inject(term.KeyEvent{Rune: 'a'})

	codeCh, _ := ReadCodeAsync(a)

	wantBuf := ui.NewBufferBuilder(4).
		WritePlain("a").SetDotToCursor().WritePlain("  R").Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(a, codeCh)
}

func TestReadCode_TriggersPrompt(t *testing.T) {
	called := 0
	a, _ := setupWithSpec(AppSpec{
		Prompt: testPrompt{trigger: func(bool) { called++ }}})

	codeCh, _ := ReadCodeAsync(a)
	cleanup(a, codeCh)

	if called != 1 {
		t.Errorf("Prompt.Trigger called %d times, want once", called)
	}
}

func TestReadCode_RedrawsOnPromptLateUpdate(t *testing.T) {
	promptContent := "old"
	prompt := testPrompt{
		get:         func() styled.Text { return styled.Plain(promptContent) },
		lateUpdates: make(chan styled.Text),
	}

	a, tty := setupWithSpec(AppSpec{Prompt: prompt})

	codeCh, _ := ReadCodeAsync(a)
	bufOldPrompt := ui.NewBufferBuilder(80).
		WritePlain("old").SetDotToCursor().Buffer()
	// Wait until old prompt is rendered
	tty.TestBuffer(t, bufOldPrompt)

	promptContent = "new"
	prompt.lateUpdates <- nil
	bufNewPrompt := ui.NewBufferBuilder(80).
		WritePlain("new").SetDotToCursor().Buffer()
	tty.TestBuffer(t, bufNewPrompt)

	cleanup(a, codeCh)
}

func TestReadCode_SupportsPersistentRPrompt(t *testing.T) {
	// TODO
}

func TestReadCode_DrawsAndFlushesNotes(t *testing.T) {
	a, tty := setup()

	codeCh, _ := ReadCodeAsync(a)

	// Sanity-check initial state.
	initBuf := ui.NewBufferBuilder(80).Buffer()
	tty.TestBuffer(t, initBuf)

	a.Notify("note")

	wantNotesBuf := ui.NewBufferBuilder(80).WritePlain("note").Buffer()
	tty.TestNotesBuffer(t, wantNotesBuf)

	if n := len(a.CopyState().Notes); n > 0 {
		t.Errorf("State.Notes has %d elements after redrawing, want 0", n)
	}

	cleanup(a, codeCh)
}

func TestReadCode_PutCursorBelowCodeAreaInFinalRedraw(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		CodeAreaState: codearea.State{
			Buffer: codearea.Buffer{Content: "some code"}},
		State: State{
			Listing: layout.Label{Content: styled.Plain("listing")}}})

	codeCh, _ := ReadCodeAsync(a)

	// Wait until the initial draw, to ensure that we are indeed observing a
	// different state later.
	wantBuf := ui.NewBufferBuilder(80).
		WritePlain("some code").
		Newline().SetDotToCursor().WritePlain("listing").Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(a, codeCh)

	wantFinalBuf := ui.NewBufferBuilder(80).
		WritePlain("some code").Newline().SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantFinalBuf)
}

func TestReadCode_QuitsOnSIGHUP(t *testing.T) {
	a, tty := setup()

	tty.Inject(term.KeyEvent{Rune: 'a'})

	codeCh, errCh := ReadCodeAsync(a)

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
	a, tty := setup()

	tty.Inject(term.KeyEvent{Rune: 'a'})

	codeCh, _ := ReadCodeAsync(a)
	wantBuf := ui.NewBufferBuilder(80).WritePlain("a").
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	tty.InjectSignal(syscall.SIGINT)

	wantBuf = ui.NewBufferBuilder(80).Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(a, codeCh)
}

func TestReadCode_RedrawsOnSIGWINCH(t *testing.T) {
	content := "1234567890"
	a, tty := setupWithSpec(AppSpec{
		CodeAreaState: codearea.State{
			Buffer: codearea.Buffer{Content: content, Dot: len(content)}}})

	codeCh, _ := ReadCodeAsync(a)

	wantBuf := ui.NewBufferBuilder(80).WritePlain("1234567890").
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	tty.SetSize(24, 4)
	tty.InjectSignal(sys.SIGWINCH)

	wantBuf = ui.NewBufferBuilder(4).WritePlain("1234567890").
		SetDotToCursor().Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(a, codeCh)
}

func setup() (App, TTYCtrl) {
	return setupWithSpec(AppSpec{})
}

func setupWithSpec(spec AppSpec) (App, TTYCtrl) {
	tty, ttyControl := NewFakeTTY()
	spec.TTY = tty
	a := NewApp(spec)
	return a, ttyControl
}

func cleanup(a App, codeCh <-chan string) {
	a.CommitEOF()
	// Make sure that ReadCode has exited
	<-codeCh
}
