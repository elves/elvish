package cli_test

import (
	"errors"
	"io"
	"strings"
	"syscall"
	"testing"
	"time"

	. "github.com/elves/elvish/pkg/cli"
	. "github.com/elves/elvish/pkg/cli/apptest"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/sys"
	"github.com/elves/elvish/pkg/ui"
)

// Lifecycle aspects.

func TestReadCode_AbortsWhenTTYSetupReturnsError(t *testing.T) {
	ttySetupErr := errors.New("a fake error")
	f := Setup(WithTTY(func(tty TTYCtrl) {
		tty.SetSetup(func() {}, ttySetupErr)
	}))

	_, err := f.Wait()

	if err != ttySetupErr {
		t.Errorf("ReadCode returns error %v, want %v", err, ttySetupErr)
	}
}

func TestReadCode_RestoresTTYBeforeReturning(t *testing.T) {
	restoreCalled := 0
	f := Setup(WithTTY(func(tty TTYCtrl) {
		tty.SetSetup(func() { restoreCalled++ }, nil)
	}))

	f.Stop()

	if restoreCalled != 1 {
		t.Errorf("Restore callback called %d times, want once", restoreCalled)
	}
}

func TestReadCode_ResetsStateBeforeReturning(t *testing.T) {
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.CodeAreaState.Buffer.Content = "some code"
	}))

	f.Stop()

	if code := GetCodeBuffer(f.App); code != (CodeBuffer{}) {
		t.Errorf("Editor state has CodeBuffer %v, want empty", code)
	}
}

func TestReadCode_CallsBeforeReadline(t *testing.T) {
	callCh := make(chan bool, 1)
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.BeforeReadline = []func(){func() { callCh <- true }}
	}))
	defer f.Stop()

	select {
	case <-callCh:
		// OK, do nothing.
	case <-time.After(time.Second):
		t.Errorf("BeforeReadline not called")
	}
}

func TestReadCode_CallsBeforeReadlineBeforePromptTrigger(t *testing.T) {
	callCh := make(chan string, 2)
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.BeforeReadline = []func(){func() { callCh <- "hook" }}
		spec.Prompt = testPrompt{trigger: func(bool) { callCh <- "prompt" }}
	}))
	defer f.Stop()

	if first := <-callCh; first != "hook" {
		t.Errorf("BeforeReadline hook not called before prompt trigger")
	}
}

func TestReadCode_CallsAfterReadline(t *testing.T) {
	callCh := make(chan string, 1)
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.AfterReadline = []func(string){func(s string) { callCh <- s }}
	}))

	feedInput(f.TTY, "abc\n")
	f.Wait()

	select {
	case calledWith := <-callCh:
		wantCalledWith := "abc"
		if calledWith != wantCalledWith {
			t.Errorf("AfterReadline hook called with %v, want %v",
				calledWith, wantCalledWith)
		}
	case <-time.After(time.Second):
		t.Errorf("AfterReadline not called")
	}
}

func TestReadCode_FinalRedraw(t *testing.T) {
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.CodeAreaState.Buffer.Content = "code"
		spec.State.Addon = Label{Content: ui.T("addon")}
	}))

	// Wait until the stable state.
	wantBuf := bb().
		Write("code").
		Newline().SetDotHere().Write("addon").Buffer()
	f.TTY.TestBuffer(t, wantBuf)

	f.Stop()

	// Final redraw hides the addon, and puts the cursor on a new line.
	wantFinalBuf := bb().
		Write("code").Newline().SetDotHere().Buffer()
	f.TTY.TestBuffer(t, wantFinalBuf)
}

// Signals.

func TestReadCode_ReturnsEOFOnSIGHUP(t *testing.T) {
	f := Setup()

	f.TTY.Inject(term.K('a'))
	// Wait until the initial redraw.
	f.TTY.TestBuffer(t, bb().Write("a").SetDotHere().Buffer())

	f.TTY.InjectSignal(syscall.SIGHUP)

	_, err := f.Wait()
	if err != io.EOF {
		t.Errorf("want ReadCode to return io.EOF on SIGHUP, got %v", err)
	}
}

func TestReadCode_ResetsStateOnSIGINT(t *testing.T) {
	f := Setup()
	defer f.Stop()

	// Ensure that the terminal shows an non-empty state.
	feedInput(f.TTY, "code")
	f.TTY.TestBuffer(t, bb().Write("code").SetDotHere().Buffer())

	f.TTY.InjectSignal(syscall.SIGINT)

	// Verify that the state has now reset.
	f.TTY.TestBuffer(t, bb().Buffer())
}

func TestReadCode_RedrawsOnSIGWINCH(t *testing.T) {
	f := Setup()
	defer f.Stop()

	// Ensure that the terminal shows the input with the intial width.
	feedInput(f.TTY, "1234567890")
	f.TTY.TestBuffer(t, bb().Write("1234567890").SetDotHere().Buffer())

	// Emulate a window size change.
	f.TTY.SetSize(24, 4)
	f.TTY.InjectSignal(sys.SIGWINCH)

	// Test that the editor has redrawn using the new width.
	f.TTY.TestBuffer(t, term.NewBufferBuilder(4).
		Write("1234567890").SetDotHere().Buffer())
}

// Code area.

func TestReadCode_LetsCodeAreaHandleEvents(t *testing.T) {
	f := Setup()
	defer f.Stop()

	feedInput(f.TTY, "code")
	f.TTY.TestBuffer(t, bb().Write("code").SetDotHere().Buffer())
}

func TestReadCode_ShowsHighlightedCode(t *testing.T) {
	f := Setup(withHighlighter(
		testHighlighter{
			get: func(code string) (ui.Text, []error) {
				return ui.T(code, ui.FgRed), nil
			},
		}))
	defer f.Stop()

	feedInput(f.TTY, "code")
	wantBuf := bb().Write("code", ui.FgRed).SetDotHere().Buffer()
	f.TTY.TestBuffer(t, wantBuf)
}

func TestReadCode_ShowsErrorsFromHighlighter(t *testing.T) {
	f := Setup(withHighlighter(
		testHighlighter{
			get: func(code string) (ui.Text, []error) {
				errors := []error{errors.New("ERR 1"), errors.New("ERR 2")}
				return ui.T(code), errors
			},
		}))
	defer f.Stop()

	feedInput(f.TTY, "code")

	wantBuf := bb().
		Write("code").SetDotHere().Newline().
		Write("ERR 1").Newline().
		Write("ERR 2").Buffer()
	f.TTY.TestBuffer(t, wantBuf)
}

func TestReadCode_RedrawsOnLateUpdateFromHighlighter(t *testing.T) {
	var styling ui.Styling
	hl := testHighlighter{
		get: func(code string) (ui.Text, []error) {
			return ui.T(code, styling), nil
		},
		lateUpdates: make(chan struct{}),
	}
	f := Setup(withHighlighter(hl))
	defer f.Stop()

	feedInput(f.TTY, "code")

	f.TTY.TestBuffer(t, bb().Write("code").SetDotHere().Buffer())

	styling = ui.FgRed
	hl.lateUpdates <- struct{}{}
	f.TTY.TestBuffer(t, bb().Write("code", ui.FgRed).SetDotHere().Buffer())
}

func withHighlighter(hl Highlighter) func(*AppSpec, TTYCtrl) {
	return WithSpec(func(spec *AppSpec) { spec.Highlighter = hl })
}

func TestReadCode_ShowsPrompt(t *testing.T) {
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.Prompt = NewConstPrompt(ui.T("> "))
	}))
	defer f.Stop()

	f.TTY.Inject(term.K('a'))
	f.TTY.TestBuffer(t, bb().Write("> a").SetDotHere().Buffer())
}

func TestReadCode_CallsPromptTrigger(t *testing.T) {
	triggerCh := make(chan bool, 1)
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.Prompt = testPrompt{trigger: func(bool) { triggerCh <- true }}
	}))
	defer f.Stop()

	select {
	case <-triggerCh:
	// Good, test passes
	case <-time.After(time.Second):
		t.Errorf("Trigger not called within 1s")
	}
}

func TestReadCode_RedrawsOnLateUpdateFromPrompt(t *testing.T) {
	promptContent := "old"
	prompt := testPrompt{
		get:         func() ui.Text { return ui.T(promptContent) },
		lateUpdates: make(chan struct{}),
	}
	f := Setup(WithSpec(func(spec *AppSpec) { spec.Prompt = prompt }))
	defer f.Stop()

	// Wait until old prompt is rendered
	f.TTY.TestBuffer(t, bb().Write("old").SetDotHere().Buffer())

	promptContent = "new"
	prompt.lateUpdates <- struct{}{}
	f.TTY.TestBuffer(t, bb().Write("new").SetDotHere().Buffer())
}

func TestReadCode_ShowsRPrompt(t *testing.T) {
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.RPrompt = NewConstPrompt(ui.T("R"))
	}))
	defer f.Stop()

	f.TTY.Inject(term.K('a'))

	wantBuf := bb().
		Write("a").SetDotHere().
		Write(strings.Repeat(" ", FakeTTYWidth-2)).
		Write("R").Buffer()
	f.TTY.TestBuffer(t, wantBuf)
}

func TestReadCode_ShowsRPromptInFinalRedrawIfPersistent(t *testing.T) {
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.CodeAreaState.Buffer.Content = "code"
		spec.RPrompt = NewConstPrompt(ui.T("R"))
		spec.RPromptPersistent = func() bool { return true }
	}))
	defer f.Stop()

	f.TTY.Inject(term.K('\n'))

	wantBuf := bb().
		Write("code" + strings.Repeat(" ", FakeTTYWidth-5) + "R").
		Newline().SetDotHere(). // cursor on newline in final redraw
		Buffer()
	f.TTY.TestBuffer(t, wantBuf)
}

func TestReadCode_HidesRPromptInFinalRedrawIfNotPersistent(t *testing.T) {
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.CodeAreaState.Buffer.Content = "code"
		spec.RPrompt = NewConstPrompt(ui.T("R"))
		spec.RPromptPersistent = func() bool { return false }
	}))
	defer f.Stop()

	f.TTY.Inject(term.K('\n'))

	wantBuf := bb().
		Write("code").          // no rprompt
		Newline().SetDotHere(). // cursor on newline in final redraw
		Buffer()
	f.TTY.TestBuffer(t, wantBuf)
}

// Addon.

func TestReadCode_LetsAddonHandleEvents(t *testing.T) {
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.State.Addon = NewCodeArea(CodeAreaSpec{
			Prompt: func() ui.Text { return ui.T("addon> ") },
		})
	}))
	defer f.Stop()

	feedInput(f.TTY, "input")

	wantBuf := bb().Newline(). // empty main code area
					Write("addon> input").SetDotHere(). // addon
					Buffer()
	f.TTY.TestBuffer(t, wantBuf)
}

type testAddon struct {
	Empty
	focus bool
}

func (a testAddon) Focus() bool { return a.focus }

func TestReadCode_RespectsAddonFocusMethod(t *testing.T) {
	addon := testAddon{}
	f := Setup(WithSpec(func(spec *AppSpec) { spec.State.Addon = &addon }))
	defer f.Stop()

	wantBuf := bb().
		SetDotHere(). // main code area has focus
		Newline().Buffer()
	f.TTY.TestBuffer(t, wantBuf)

	addon.focus = true
	f.App.Redraw()

	wantBuf = bb().
		Newline().SetDotHere(). // addon has focus
		Buffer()
	f.TTY.TestBuffer(t, wantBuf)
}

// Misc features.

func TestReadCode_TrimsBufferToMaxHeight(t *testing.T) {
	f := Setup(func(spec *AppSpec, tty TTYCtrl) {
		spec.MaxHeight = func() int { return 2 }
		// The code needs 3 lines to completely show.
		spec.CodeAreaState.Buffer.Content = strings.Repeat("a", 15)
		tty.SetSize(10, 5) // Width = 5 to make it easy to test
	})
	defer f.Stop()

	wantBuf := term.NewBufferBuilder(5).
		Write(strings.Repeat("a", 10)). // Only show 2 lines due to MaxHeight.
		Buffer()
	f.TTY.TestBuffer(t, wantBuf)
}

func TestReadCode_ShowNotes(t *testing.T) {
	// Set up with a binding where 'a' can block indefinitely. This is useful
	// for testing the behavior of writing multiple notes.
	inHandler := make(chan struct{})
	unblock := make(chan struct{})
	f := Setup(WithSpec(func(spec *AppSpec) {
		spec.OverlayHandler = MapHandler{
			term.K('a'): func() {
				inHandler <- struct{}{}
				<-unblock
			},
		}
	}))
	defer f.Stop()

	// Wait until initial draw.
	f.TTY.TestBuffer(t, bb().Buffer())

	// Make sure that the app is blocked within an event handler.
	f.TTY.Inject(term.K('a'))
	<-inHandler

	// Write two notes, and unblock the event handler
	f.App.Notify("note")
	f.App.Notify("note 2")
	unblock <- struct{}{}

	// Test that the note is rendered onto the notes buffer.
	wantNotesBuf := bb().Write("note").Newline().Write("note 2").Buffer()
	f.TTY.TestNotesBuffer(t, wantNotesBuf)

	// Test that notes are flushed after being rendered.
	if n := len(f.App.CopyState().Notes); n > 0 {
		t.Errorf("State.Notes has %d elements after redrawing, want 0", n)
	}
}

func TestReadCode_DoesNotCrashWithNilTTY(t *testing.T) {
	f := Setup(WithSpec(func(spec *AppSpec) { spec.TTY = nil }))
	defer f.Stop()
}

// Other properties.

func TestReadCode_DoesNotLockWithALotOfInputsWithNewlines(t *testing.T) {
	// Regression test for #887
	f := Setup(WithTTY(func(tty TTYCtrl) {
		for i := 0; i < 1000; i++ {
			tty.Inject(term.K('#'), term.K('\n'))
		}
	}))
	terminated := make(chan struct{})
	go func() {
		f.Wait()
		close(terminated)
	}()
	select {
	case <-terminated:
	// OK
	case <-time.After(time.Second):
		t.Errorf("ReadCode did not terminate within 1s")
	}
}

func TestReadCode_DoesNotReadMoreEventsThanNeeded(t *testing.T) {
	f := Setup()
	defer f.Stop()
	f.TTY.Inject(term.K('a'), term.K('\n'), term.K('b'))
	f.Wait()
	if event := <-f.TTY.EventCh(); event != term.K('b') {
		t.Errorf("got event %v, want %v", event, term.K('b'))
	}
}

// Test utilities.

func bb() *term.BufferBuilder {
	return term.NewBufferBuilder(FakeTTYWidth)
}

func feedInput(ttyCtrl TTYCtrl, input string) {
	for _, r := range input {
		ttyCtrl.Inject(term.K(r))
	}
}

// A Highlighter implementation useful for testing.
type testHighlighter struct {
	get         func(code string) (ui.Text, []error)
	lateUpdates chan struct{}
}

func (hl testHighlighter) Get(code string) (ui.Text, []error) {
	return hl.get(code)
}

func (hl testHighlighter) LateUpdates() <-chan struct{} {
	return hl.lateUpdates
}

// A Prompt implementation useful for testing.
type testPrompt struct {
	trigger     func(force bool)
	get         func() ui.Text
	lateUpdates chan struct{}
}

func (p testPrompt) Trigger(force bool) {
	if p.trigger != nil {
		p.trigger(force)
	}
}

func (p testPrompt) Get() ui.Text {
	if p.get != nil {
		return p.get()
	}
	return nil
}

func (p testPrompt) LateUpdates() <-chan struct{} {
	return p.lateUpdates
}
