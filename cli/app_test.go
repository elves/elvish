package cli

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/sys"
)

const (
	testTTYHeight = 24
	testTTYWidth  = 60
)

// Lifecycle aspects.

func TestSetup_ErrorAbortsReadCode(t *testing.T) {
	a, tty := setup()
	setupErr := errors.New("a fake error")
	tty.SetSetup(func() {}, setupErr)

	_, err := a.ReadCode()

	if err != setupErr {
		t.Errorf("ReadCode returns error %v, want %v", err, setupErr)
	}
}

func TestSetup_RestoreIsCalled(t *testing.T) {
	a, tty := setup()
	restoreCalled := 0
	tty.SetSetup(func() { restoreCalled++ }, nil)

	tty.Inject(term.K('\n'))
	a.ReadCode()

	if restoreCalled != 1 {
		t.Errorf("Restore callback called %d times, want once", restoreCalled)
	}
}

func TestState_IsResetBeforeReadCodeReturns(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		CodeAreaState: codearea.State{
			Buffer: codearea.Buffer{Content: "some code"}}})

	tty.Inject(term.K('\n'))
	a.ReadCode()

	if code := a.CodeArea().CopyState().Buffer.Content; code != "" {
		t.Errorf("Editor state has code %q, want empty", code)
	}
}

func TestBeforeReadline(t *testing.T) {
	called := 0
	a, tty := setupWithSpec(AppSpec{
		BeforeReadline: []func(){func() { called++ }},
	})

	tty.Inject(term.K('\n'))
	a.ReadCode()

	if called != 1 {
		t.Errorf("BeforeReadline hook called %d times, want 1", called)
	}
}

func TestAfterReadline(t *testing.T) {
	calledWith := []string{}
	a, tty := setupWithSpec(AppSpec{
		AfterReadline: []func(string){func(s string) {
			calledWith = append(calledWith, s)
		}},
	})

	feedInput(tty, "abc\n")

	a.ReadCode()

	wantCalledWith := []string{"abc"}
	if !reflect.DeepEqual(calledWith, wantCalledWith) {
		t.Errorf("AfterReadline hook called with %v, want %v", calledWith, wantCalledWith)
	}
}

func TestFinalRedraw(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		CodeAreaState: codearea.State{
			Buffer: codearea.Buffer{Content: "code"}},
		State: State{
			Addon: layout.Label{Content: styled.Plain("addon")}}})
	codeCh, _ := ReadCodeAsync(a)

	// Wait until the stable state.
	wantBuf := bb().
		Write("code").
		Newline().SetDotHere().Write("addon").Buffer()
	tty.TestBuffer(t, wantBuf)

	cleanup(a, codeCh)

	// Final redraw hides the addon, and puts the cursor on a new line.
	wantFinalBuf := bb().
		Write("code").Newline().SetDotHere().Buffer()
	tty.TestBuffer(t, wantFinalBuf)
}

// Signals.

func TestSIGHUP_ReturnsEOF(t *testing.T) {
	a, tty := setup()

	tty.Inject(term.K('a'))

	_, errCh := ReadCodeAsync(a)
	// Wait until the initial redraw.
	tty.TestBuffer(t, bb().Write("a").SetDotHere().Buffer())

	tty.InjectSignal(syscall.SIGHUP)

	select {
	case err := <-errCh:
		if err != io.EOF {
			t.Errorf("want ReadCode to return io.EOF on SIGHUP, got %v", err)
		}
	case <-time.After(time.Second):
		t.Errorf("SIGHUP did not cause ReadCode to return")
	}
}

func TestSIGINT_ResetsState(t *testing.T) {
	a, tty := setup()

	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)
	// Ensure that the terminal shows an non-empty state.
	feedInput(tty, "code")
	tty.TestBuffer(t, bb().Write("code").SetDotHere().Buffer())

	tty.InjectSignal(syscall.SIGINT)

	// Verify that the state has now reset.
	tty.TestBuffer(t, bb().Buffer())
}

func TestSIGWINCH_TriggersRedraw(t *testing.T) {
	a, tty := setup()
	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)

	// Ensure that the terminal shows the input with the intial width.
	feedInput(tty, "1234567890")
	tty.TestBuffer(t, bb().Write("1234567890").SetDotHere().Buffer())

	// Emulate a window size change.
	tty.SetSize(24, 4)
	tty.InjectSignal(sys.SIGWINCH)

	// Test that the editor has redrawn using the new width.
	tty.TestBuffer(t, term.NewBufferBuilder(4).
		Write("1234567890").SetDotHere().Buffer())
}

// Code area.

func TestCodeArea_HandlesEvents(t *testing.T) {
	a, tty := setup()
	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)

	feedInput(tty, "code")
	tty.TestBuffer(t, bb().Write("code").SetDotHere().Buffer())
}

func TestHighlighter(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		Highlighter: testHighlighter{
			get: func(code string) (styled.Text, []error) {
				return styled.MakeText(code, "red"), nil
			},
		}})

	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)
	feedInput(tty, "code")

	wantBuf := bb().
		Write("code", "red").
		SetDotHere().Buffer()
	tty.TestBuffer(t, wantBuf)
}

func TestHighlighter_Errors(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		Highlighter: testHighlighter{
			get: func(code string) (styled.Text, []error) {
				errors := []error{errors.New("ERR 1"), errors.New("ERR 2")}
				return styled.Plain(code), errors
			},
		}})

	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)
	feedInput(tty, "code")

	wantBuf := bb().
		Write("code").SetDotHere().Newline().
		Write("ERR 1").Newline().
		Write("ERR 2").Buffer()
	tty.TestBuffer(t, wantBuf)
}

func TestHighlighter_LateUpdate(t *testing.T) {
	style := ""
	hl := testHighlighter{
		get: func(code string) (styled.Text, []error) {
			return styled.MakeText(code, style), nil
		},
		lateUpdates: make(chan styled.Text),
	}
	a, tty := setupWithSpec(AppSpec{Highlighter: hl})

	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)
	feedInput(tty, "code")

	tty.TestBuffer(t, bb().Write("code").SetDotHere().Buffer())

	style = "red"
	hl.lateUpdates <- nil
	tty.TestBuffer(t, bb().WriteStyled(
		styled.MakeText("code", "red")).SetDotHere().Buffer())
}

func TestPrompt(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		Prompt: constPrompt{styled.Plain("> ")}})

	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)

	tty.Inject(term.K('a'))

	tty.TestBuffer(t, bb().Write("> a").SetDotHere().Buffer())
}

func TestPrompt_Trigger(t *testing.T) {
	called := 0
	a, _ := setupWithSpec(AppSpec{
		Prompt: testPrompt{trigger: func(bool) { called++ }}})

	codeCh, _ := ReadCodeAsync(a)
	cleanup(a, codeCh)

	if called != 1 {
		t.Errorf("Prompt.Trigger called %d times, want once", called)
	}
}

func TestPrompt_LateUpdate(t *testing.T) {
	promptContent := "old"
	prompt := testPrompt{
		get:         func() styled.Text { return styled.Plain(promptContent) },
		lateUpdates: make(chan styled.Text),
	}
	a, tty := setupWithSpec(AppSpec{Prompt: prompt})

	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)

	// Wait until old prompt is rendered
	tty.TestBuffer(t, bb().Write("old").SetDotHere().Buffer())

	promptContent = "new"
	prompt.lateUpdates <- nil
	tty.TestBuffer(t, bb().Write("new").SetDotHere().Buffer())
}

func TestRPrompt(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		RPrompt: constPrompt{styled.Plain("R")}})

	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)

	tty.Inject(term.K('a'))

	wantBuf := bb().
		Write("a").SetDotHere().
		Write(strings.Repeat(" ", testTTYWidth-2)).
		Write("R").Buffer()
	tty.TestBuffer(t, wantBuf)
}

func TestRPrompt_Persistent(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		CodeAreaState: codearea.State{
			Buffer: codearea.Buffer{Content: "code"}},
		RPrompt:           constPrompt{styled.Plain("R")},
		RPromptPersistent: func() bool { return true },
	})

	tty.Inject(term.K('\n'))
	a.ReadCode()

	wantBuf := bb().
		Write("code" + strings.Repeat(" ", testTTYWidth-5) + "R").
		Newline().SetDotHere(). // cursor on newline in final redraw
		Buffer()
	tty.TestBuffer(t, wantBuf)
}

func TestRPrompt_NotPersistent(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		CodeAreaState: codearea.State{
			Buffer: codearea.Buffer{Content: "code"}},
		RPrompt:           constPrompt{styled.Plain("R")},
		RPromptPersistent: func() bool { return false },
	})

	tty.Inject(term.K('\n'))
	a.ReadCode()

	wantBuf := bb().
		Write("code").          // no rprompt
		Newline().SetDotHere(). // cursor on newline in final redraw
		Buffer()
	tty.TestBuffer(t, wantBuf)
}

// Addon.

func TestAddon_HandlesEvents(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		State: State{
			Addon: codearea.New(codearea.Spec{
				Prompt: func() styled.Text { return styled.Plain("addon> ") },
			})}})

	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)
	feedInput(tty, "input")

	wantBuf := bb().Newline(). // empty main code area
					Write("addon> input").SetDotHere(). // addon
					Buffer()
	tty.TestBuffer(t, wantBuf)
}

type testAddon struct {
	layout.Empty
	focus bool
}

func (a testAddon) Focus() bool { return a.focus }

func TestAddon_ControlsFocus(t *testing.T) {
	addon := testAddon{}
	a, tty := setupWithSpec(AppSpec{State: State{Addon: &addon}})

	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)

	wantBuf := bb().
		SetDotHere(). // main code area has focus
		Newline().Buffer()
	tty.TestBuffer(t, wantBuf)

	addon.focus = true
	a.Redraw()

	wantBuf = bb().
		Newline().SetDotHere(). // addon has focus
		Buffer()
	tty.TestBuffer(t, wantBuf)
}

// Misc features.

func TestMaxHeight(t *testing.T) {
	a, tty := setupWithSpec(AppSpec{
		MaxHeight: func() int { return 2 },
		CodeAreaState: codearea.State{
			// The code needs 3 lines to completely show.
			Buffer: codearea.Buffer{Content: strings.Repeat("a", 15)}}})
	tty.SetSize(10, 5) // Width = 5 to make it easy to test
	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)

	wantBuf := term.NewBufferBuilder(5).
		Write(strings.Repeat("a", 10)). // Only show 2 lines due to MaxHeight.
		Buffer()
	tty.TestBuffer(t, wantBuf)
}

func TestNotes(t *testing.T) {
	a, tty := setup()
	codeCh, _ := ReadCodeAsync(a)
	defer cleanup(a, codeCh)

	// Wait until initial draw.
	tty.TestBuffer(t, bb().Buffer())

	a.Notify("note")

	// Test that the note is rendered onto the notes buffer.
	wantNotesBuf := bb().
		Write("note").Buffer()
	tty.TestNotesBuffer(t, wantNotesBuf)

	// Test that notes are flushed after being rendered.
	if n := len(a.CopyState().Notes); n > 0 {
		t.Errorf("State.Notes has %d elements after redrawing, want 0", n)
	}
}

// Test utilities.

func setup() (App, TTYCtrl) {
	return setupWithSpec(AppSpec{})
}

func setupWithSpec(spec AppSpec) (App, TTYCtrl) {
	tty, ttyControl := NewFakeTTY()
	ttyControl.SetSize(testTTYHeight, testTTYWidth)
	spec.TTY = tty
	a := NewApp(spec)
	return a, ttyControl
}

func bb() *term.BufferBuilder {
	return term.NewBufferBuilder(testTTYWidth)
}

func cleanup(a App, codeCh <-chan string) {
	a.CommitEOF()
	// Make sure that ReadCode has exited
	<-codeCh
}

func feedInput(ttyCtrl TTYCtrl, input string) {
	for _, r := range input {
		ttyCtrl.Inject(term.K(r))
	}
}
